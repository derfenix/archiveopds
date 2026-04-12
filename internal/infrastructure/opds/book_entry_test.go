package opds

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"

	"git.derfenix.pro/derfenix/archiveopds/internal/domain/book"
)

func xmlMarshalEntry(e atomEntry) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")
	w := struct {
		XMLName xml.Name  `xml:"http://www.w3.org/2005/Atom feed"`
		XmlnsDC string    `xml:"xmlns:dc,attr"`
		Entry   atomEntry `xml:"entry"`
	}{
		XmlnsDC: xmlnsDC,
		Entry:   e,
	}
	if err := enc.Encode(w); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func TestPublicationTitle(t *testing.T) {
	t.Parallel()
	b := book.Book{Author: "Иванов", BookTitle: "Рассказ", Title: "Иванов — Рассказ"}
	if got := publicationTitle(b); got != "Рассказ" {
		t.Fatalf("got %q", got)
	}
	b2 := book.Book{Title: "Только заголовок"}
	if got := publicationTitle(b2); got != "Только заголовок" {
		t.Fatalf("got %q", got)
	}
}

func TestAcquisitionEntry_containsMetadata(t *testing.T) {
	t.Parallel()
	b := book.Book{
		ID:          "x",
		Title:       "А. Пушкин — Евгений Онегин",
		Author:      "А. Пушкин",
		BookTitle:   "Евгений Онегин",
		Genre:       "antique, poetry",
		Year:        1833,
		MIMEType:    "application/fb2+xml",
		Size:        500000,
		Language:    "ru",
		LibraryID:   "fb-1",
		Annotation:  "Роман в стихах.",
		Series:      "Классика",
		SeriesIndex: "1",
	}
	e := acquisitionEntry(b, "2020-01-01T00:00:00Z", "http://h/opds/acquire/x")
	raw, err := xmlMarshalEntry(e)
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	for _, needle := range []string{
		"<title>Евгений Онегин</title>",
		"<name>А. Пушкин</name>",
		`http://purl.org/dc/elements/1.1/">ru</language>`,
		`http://purl.org/dc/elements/1.1/">1833</issued>`,
		`http://purl.org/dc/elements/1.1/">application/fb2+xml</format>`,
		`http://purl.org/dc/elements/1.1/">fb-1</identifier>`,
		`term="antique"`,
		`term="poetry"`,
		`<summary type="text">`,
		"Роман в стихах",
	} {
		if !strings.Contains(s, needle) {
			t.Fatalf("missing %q in\n%s", needle, s)
		}
	}
}
