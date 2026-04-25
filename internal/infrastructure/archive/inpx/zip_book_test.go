package inpx

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestFindZipEntry_numericInnerMatchesFb2(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	zpath := filepath.Join(dir, "d.fb2-009373-367300.zip")
	w, err := os.Create(zpath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(w)
	payload := []byte("<?xml version=\"1.0\"?><FictionBook/>")
	fh, err := zw.Create("110119.fb2")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fh.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	zr, err := zip.OpenReader(zpath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = zr.Close() }()

	f := findZipEntry(zr, "110119")
	if f == nil {
		t.Fatal("expected match for inner 110119 -> 110119.fb2")
	}
	if f.Name != "110119.fb2" {
		t.Fatalf("got %q", f.Name)
	}
}

func TestInnerPathUnsafe(t *testing.T) {
	t.Parallel()
	if !innerPathUnsafe("..") {
		t.Fatal("expected unsafe")
	}
	if !innerPathUnsafe("a/../b") {
		t.Fatal("expected unsafe")
	}
	if !innerPathUnsafe("/abs") {
		t.Fatal("expected unsafe for absolute")
	}
	if innerPathUnsafe("110119.fb2") {
		t.Fatal("expected safe")
	}
}

func TestOpenFromZip_acquireByID(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	zpath := filepath.Join(dir, "d.fb2-009373-367300.zip")
	w, err := os.Create(zpath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(w)
	want := []byte("<FictionBook/>")
	fh, err := zw.Create("110119.fb2")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fh.Write(want); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	id := encodeBookRef("d.fb2-009373-367300", "110119")
	nav := &Navigator{root: dir}
	rc, sizeOut, ctype, err := nav.openFromZip(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rc.Close() }()

	if sizeOut != int64(len(want)) {
		t.Fatalf("size %d want %d", sizeOut, len(want))
	}
	if ctype != "application/fb2+xml" {
		t.Fatalf("ctype %q", ctype)
	}
	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("body %q want %q", got, want)
	}
}

func TestRefCountedReadCloser_idempotentClose(t *testing.T) {
	t.Parallel()
	var n int32
	r := &refCountedReadCloser{
		rc: io.NopCloser(strings.NewReader("x")),
		onClose: func() {
			atomic.AddInt32(&n, 1)
		},
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt32(&n) != 1 {
		t.Fatalf("onClose should run once, got %d", n)
	}
}
