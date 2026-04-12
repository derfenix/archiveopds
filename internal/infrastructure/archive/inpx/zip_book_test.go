package inpx

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
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
	defer zr.Close()

	f := findZipEntry(zr, "110119")
	if f == nil {
		t.Fatal("expected match for inner 110119 -> 110119.fb2")
	}
	if f.Name != "110119.fb2" {
		t.Fatalf("got %q", f.Name)
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
	rc, n, ctype, err := openFromZip(context.Background(), dir, id)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()

	if n != int64(len(want)) {
		t.Fatalf("size %d want %d", n, len(want))
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
