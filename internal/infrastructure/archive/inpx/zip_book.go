package inpx

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"git.derfenix.pro/derfenix/archiveopds/internal/application/port/outbound"
	"git.derfenix.pro/derfenix/archiveopds/internal/domain/book"
	domerr "git.derfenix.pro/derfenix/archiveopds/internal/domain/errors"
)

const (
	// maxBookInMemoryBytes — до этого размера распакованный файл держим в RAM (bytes.Reader).
	maxBookInMemoryBytes = 8 << 20
	// maxBookLoadBytes — верхняя граница выдачи одной книги (RAM или временный файл на диске).
	maxBookLoadBytes = 256 << 20
)

func isNotFoundOpenErr(err error) bool {
	return err != nil && os.IsNotExist(err)
}

// innerPathUnsafe reports paths that must never be resolved inside a zip (zip-slip style abuse).
func innerPathUnsafe(inner string) bool {
	trim := strings.TrimSpace(inner)
	if trim == "" {
		return true
	}
	if filepath.IsAbs(trim) {
		return true
	}
	s := filepath.ToSlash(trim)
	return strings.Contains(s, "..")
}

// errZipEntryNotFound is returned when findAndOpenZipEntry cannot match inner to a file in the volume.
var errZipEntryNotFound = errors.New("inpx: no zip entry for inner path")

// findAndOpenZipEntry locks only for lookup and zip.File.Open; the caller must close rc. Reads run without the volume lock.
func (n *Navigator) findAndOpenZipEntry(zpath, inner string) (*zip.File, io.ReadCloser, error) {
	v, err := n.getOrOpenVolume(zpath)
	if err != nil {
		return nil, nil, err
	}
	v.mu.Lock()
	f := findZipEntry(v.zr, inner)
	if f == nil {
		v.mu.Unlock()
		v.releaseLease() // getOrOpenVolume lease, no entry opened
		return nil, nil, errZipEntryNotFound
	}
	rc, err := f.Open()
	v.mu.Unlock()
	if err != nil {
		v.releaseLease() // getOrOpenVolume lease, Open failed
		return f, nil, err
	}
	// getOrOpenVolume already incremented v.refs; onClose runs after inner rc.Close
	rc = &refCountedReadCloser{
		rc: rc,
		onClose: func() {
			v.releaseLease()
		},
	}
	return f, rc, nil
}

// refCountedReadCloser closes the zip entry body then releases the volume lease (idempotent Close).
type refCountedReadCloser struct {
	rc      io.ReadCloser
	onClose func()
	once    sync.Once
	err     error
}

func (r *refCountedReadCloser) Read(p []byte) (int, error) {
	return r.rc.Read(p)
}

func (r *refCountedReadCloser) Close() error {
	r.once.Do(func() {
		r.err = r.rc.Close()
		if r.onClose != nil {
			r.onClose()
		}
	})
	return r.err
}

func (n *Navigator) openFromZip(ctx context.Context, id book.ID) (outbound.ReadSeekCloser, int64, string, error) {
	zipStem, inner, ok := decodeBookRef(id)
	if !ok {
		return nil, 0, "", domerr.ErrNotFound
	}
	if innerPathUnsafe(inner) {
		return nil, 0, "", domerr.ErrNotFound
	}
	zpath := filepath.Join(n.root, zipStem+".zip")

	f, rc, err := n.findAndOpenZipEntry(zpath, inner)
	if err != nil {
		if errors.Is(err, errZipEntryNotFound) {
			return nil, 0, "", domerr.ErrNotFound
		}
		if isNotFoundOpenErr(err) {
			return nil, 0, "", domerr.ErrNotFound
		}
		if f == nil {
			return nil, 0, "", fmt.Errorf("%w: %w", domerr.ErrArchiveOpen, err)
		}
		return nil, 0, "", err
	}
	defer func() { _ = rc.Close() }()

	usize := f.UncompressedSize64
	ctype := mimeFromName(f.Name)

	if usize > 0 && usize > maxBookLoadBytes {
		return nil, 0, "", fmt.Errorf("%w (%d МиБ)", domerr.ErrBookTooLarge, maxBookLoadBytes>>20)
	}

	if usize > 0 && usize <= maxBookInMemoryBytes {
		data, err := readExact(ctx, rc, int64(usize))
		if err != nil {
			return nil, 0, "", err
		}
		br := bytes.NewReader(data)
		return &readSeekNopCloser{br}, int64(len(data)), ctype, nil
	}

	trc, sz, err := materializeZipEntryToTemp(ctx, rc, int64(usize))
	if err != nil {
		return nil, 0, "", err
	}
	return trc, sz, ctype, nil
}

func readExact(ctx context.Context, r io.Reader, n int64) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	data := make([]byte, n)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}
	return data, ctx.Err()
}

// materializeZipEntryToTemp копирует содержимое записи zip во временный файл (O(буфер) RAM).
// declared == 0 — размер неизвестен, копируем не более maxBookLoadBytes.
func materializeZipEntryToTemp(ctx context.Context, r io.Reader, declared int64) (*tempReadSeekCloser, int64, error) {
	tmp, err := os.CreateTemp("", "archiveopds-acq-*")
	if err != nil {
		return nil, 0, err
	}
	path := tmp.Name()

	var written int64
	var copyErr error
	if declared > 0 {
		copyErr = copyNCtx(ctx, tmp, r, declared)
		written = declared
	} else {
		written, copyErr = copyLimitedMax(ctx, tmp, r, maxBookLoadBytes)
	}
	if copyErr != nil {
		_ = tmp.Close()
		_ = os.Remove(path)
		return nil, 0, copyErr
	}

	if _, err := tmp.Seek(0, 0); err != nil {
		_ = tmp.Close()
		_ = os.Remove(path)
		return nil, 0, err
	}
	return &tempReadSeekCloser{f: tmp, path: path}, written, nil
}

func copyNCtx(ctx context.Context, dst io.Writer, src io.Reader, n int64) error {
	buf := make([]byte, 64*1024)
	for n > 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		chunk := int64(len(buf))
		if chunk > n {
			chunk = n
		}
		if _, err := io.ReadFull(src, buf[:chunk]); err != nil {
			return err
		}
		if _, err := dst.Write(buf[:chunk]); err != nil {
			return err
		}
		n -= chunk
	}
	return nil
}

func copyLimitedMax(ctx context.Context, dst io.Writer, src io.Reader, max int64) (written int64, err error) {
	buf := make([]byte, 32*1024)
	for {
		if err := ctx.Err(); err != nil {
			return written, err
		}
		n, rerr := src.Read(buf)
		if n > 0 {
			if written+int64(n) > max {
				return written, fmt.Errorf("%w (%d МиБ)", domerr.ErrBookTooLarge, max>>20)
			}
			nw, werr := dst.Write(buf[:n])
			written += int64(nw)
			if werr != nil {
				return written, werr
			}
		}
		if rerr == io.EOF {
			return written, nil
		}
		if rerr != nil {
			return written, rerr
		}
	}
}

type tempReadSeekCloser struct {
	f    *os.File
	path string
}

func (t *tempReadSeekCloser) Read(p []byte) (int, error) { return t.f.Read(p) }

func (t *tempReadSeekCloser) Seek(offset int64, whence int) (int64, error) {
	return t.f.Seek(offset, whence)
}

func (t *tempReadSeekCloser) Close() error {
	err := t.f.Close()
	_ = os.Remove(t.path)
	return err
}

func findZipEntry(zr *zip.ReadCloser, inner string) *zip.File {
	innerNorm := filepath.ToSlash(strings.TrimSpace(inner))
	if innerPathUnsafe(inner) || innerNorm == "" {
		return nil
	}

	if f := matchZipPath(zr, innerNorm, matchExact); f != nil {
		return f
	}
	if f := matchZipPath(zr, innerNorm, matchFold); f != nil {
		return f
	}

	// В индексе часто только числовой id без расширения (110119 → 110119.fb2).
	if path.Ext(innerNorm) == "" {
		for _, ext := range []string{".fb2", ".epub", ".pdf", ".zip", ".mobi", ".azw3", ".djvu", ".txt"} {
			p := innerNorm + ext
			if f := matchZipPath(zr, p, matchExact); f != nil {
				return f
			}
			if f := matchZipPath(zr, p, matchFold); f != nil {
				return f
			}
		}
	}

	// Совпадение по базовому имени без расширения (и подкаталоги).
	wantStem := strings.TrimSuffix(path.Base(innerNorm), path.Ext(path.Base(innerNorm)))
	if wantStem == "" {
		return nil
	}
	var found *zip.File
	for _, cand := range zr.File {
		if cand.FileInfo().IsDir() {
			continue
		}
		base := path.Base(cand.Name)
		stem := strings.TrimSuffix(base, path.Ext(base))
		if stem == wantStem {
			if found != nil {
				// Неоднозначность: оставляем первое совпадение (как в типичных томах).
				continue
			}
			found = cand
		}
	}
	return found
}

func matchZipPath(zr *zip.ReadCloser, want string, eq func(a, b string) bool) *zip.File {
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := filepath.ToSlash(f.Name)
		if eq(name, want) {
			return f
		}
	}
	return nil
}

func matchExact(a, b string) bool { return a == b }

func matchFold(a, b string) bool { return strings.EqualFold(a, b) }

// readSeekNopCloser оборачивает *bytes.Reader для интерфейса ReadSeekCloser.
type readSeekNopCloser struct {
	*bytes.Reader
}

func (r *readSeekNopCloser) Close() error {
	return nil
}
