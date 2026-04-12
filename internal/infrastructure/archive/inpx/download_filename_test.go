package inpx

import (
	"strings"
	"testing"

	"git.derfenix.pro/derfenix/archiveopds/internal/domain/book"
)

func TestNormalizeAuthorField_commas(t *testing.T) {
	t.Parallel()
	got := normalizeAuthorField("Громов,Александр,Николаевич")
	want := "Громов, Александр, Николаевич"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestNormalizeAuthorField_multipleAuthors(t *testing.T) {
	t.Parallel()
	got := normalizeAuthorField("Иванов И. : Петров П.П.")
	if !strings.Contains(got, "Иванов") || !strings.Contains(got, "Петров") {
		t.Fatalf("got %q", got)
	}
	if !strings.Contains(got, ", ") {
		t.Fatalf("expected comma-separated: %q", got)
	}
}

func TestDownloadFilename_format(t *testing.T) {
	t.Parallel()
	b := book.Book{
		Author:    "Громов,Александр,Николаевич",
		BookTitle: "Первый из могикан",
		Title:     "Громов,Александр,Николаевич — Первый из могикан",
		RelPath:   "110119",
		MIMEType:  "application/fb2+xml",
	}
	name := downloadFilename(b)
	if !strings.HasSuffix(name, ".fb2") {
		t.Fatalf("want .fb2 suffix: %q", name)
	}
	if !strings.Contains(name, "Первый из могикан") {
		t.Fatalf("missing title: %q", name)
	}
	if !strings.Contains(name, "Громов") {
		t.Fatalf("missing author: %q", name)
	}
	if !strings.Contains(name, ". ") {
		t.Fatalf("want dot between author block and title: %q", name)
	}
}
