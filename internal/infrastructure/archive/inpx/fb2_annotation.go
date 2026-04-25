package inpx

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"unicode/utf8"

	domerr "git.derfenix.pro/derfenix/archiveopds/internal/domain/errors"
)

// fb2ReadHead — сколько байт с начала файла читать для разметки description/annotation.
const fb2ReadHead = 768 << 10

// maxAnnotationRunes — ограничение длины summary в OPDS.
const maxAnnotationRunes = 12000

// extractFB2Annotation вытаскивает текст из <annotation> (FictionBook 2.x).
func extractFB2Annotation(data []byte) string {
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = false

	var depth int // >0 — внутри <annotation>
	var out strings.Builder

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			if depth > 0 && out.Len() > 0 {
				return finalizeAnnotation(out.String())
			}
			return ""
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if depth == 0 {
				if t.Name.Local == "annotation" {
					depth = 1
				}
				continue
			}
			depth++
			if t.Name.Local == "empty-line" {
				out.WriteByte('\n')
			}
		case xml.EndElement:
			if depth == 0 {
				continue
			}
			depth--
			if depth == 0 && t.Name.Local == "annotation" {
				return finalizeAnnotation(out.String())
			}
		case xml.CharData:
			if depth > 0 {
				s := strings.TrimSpace(string(t))
				if s == "" {
					continue
				}
				if out.Len() > 0 && !strings.HasSuffix(out.String(), "\n") {
					out.WriteByte(' ')
				}
				out.WriteString(s)
			}
		}
	}

	if depth > 0 && out.Len() > 0 {
		return finalizeAnnotation(out.String())
	}
	return ""
}

func finalizeAnnotation(s string) string {
	lines := strings.Split(s, "\n")
	var b strings.Builder
	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")
		if line == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(line)
	}
	s = strings.TrimSpace(b.String())
	if s == "" {
		return ""
	}
	if utf8.RuneCountInString(s) <= maxAnnotationRunes {
		return s
	}
	r := []rune(s)
	return string(r[:maxAnnotationRunes]) + "…"
}

func readFB2AnnotationFromZip(ctx context.Context, n *Navigator, root, zipStem, inner string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if innerPathUnsafe(inner) {
		return "", nil
	}
	zpath := filepath.Join(root, zipStem+".zip")

	f, rc, err := n.findAndOpenZipEntry(zpath, inner)
	if err != nil {
		if errors.Is(err, errZipEntryNotFound) {
			return "", nil
		}
		if isNotFoundOpenErr(err) {
			return "", err
		}
		if f == nil {
			return "", fmt.Errorf("%w: %w", domerr.ErrArchiveOpen, err)
		}
		return "", err
	}
	defer func() { _ = rc.Close() }()
	data, err := readZipEntryPrefix(ctx, rc, fb2ReadHead)
	if err != nil {
		return "", err
	}
	ann := extractFB2Annotation(data)
	return ann, nil
}

func readZipEntryPrefix(ctx context.Context, r io.Reader, max int) ([]byte, error) {
	lr := io.LimitReader(r, int64(max))
	data, err := io.ReadAll(lr)
	if err != nil {
		return nil, err
	}
	return data, ctx.Err()
}
