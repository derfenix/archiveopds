package opds

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"
	"time"

	"git.derfenix.pro/derfenix/archiveopds/internal/domain/book"
	"git.derfenix.pro/derfenix/archiveopds/internal/domain/catalog"
)

// FeedBuilder — адаптер представления: домен → Atom OPDS 1.2.
type FeedBuilder struct {
	BaseURL string
}

const xmlnsDC = "http://purl.org/dc/elements/1.1/"

type atomFeed struct {
	XMLName   xml.Name    `xml:"http://www.w3.org/2005/Atom feed"`
	XmlnsOPDS string      `xml:"xmlns:opds,attr"`
	XmlnsDC   string      `xml:"xmlns:dc,attr"`
	ID        string      `xml:"id"`
	Title     string      `xml:"title"`
	Updated   string      `xml:"updated"`
	Links     []atomLink  `xml:"link"`
	Entries   []atomEntry `xml:"entry"`
}

type atomLink struct {
	Rel  string `xml:"rel,attr,omitempty"`
	Href string `xml:"href,attr"`
	Type string `xml:"type,attr,omitempty"`
}

type atomPerson struct {
	Name string `xml:"name"`
}

type atomCategory struct {
	Term  string `xml:"term,attr"`
	Label string `xml:"label,attr,omitempty"`
}

type atomSummary struct {
	Type string `xml:"type,attr"`
	Text string `xml:",chardata"`
}

type atomEntry struct {
	ID           string         `xml:"id"`
	Title        string         `xml:"title"`
	Updated      string         `xml:"updated"`
	Authors      []atomPerson   `xml:"author,omitempty"`
	Categories   []atomCategory `xml:"category,omitempty"`
	DCLanguage   string         `xml:"http://purl.org/dc/elements/1.1/ language,omitempty"`
	DCIssued     string         `xml:"http://purl.org/dc/elements/1.1/ issued,omitempty"`
	DCFormat     string         `xml:"http://purl.org/dc/elements/1.1/ format,omitempty"`
	DCIdentifier string         `xml:"http://purl.org/dc/elements/1.1/ identifier,omitempty"`
	Summary      *atomSummary   `xml:"summary,omitempty"`
	Links        []atomLink     `xml:"link"`
}

// CatalogRootPath — канонический вход OPDS (Calibre OPDS-reader и др. ждут …/opds).
const CatalogRootPath = "/opds"

func (f *FeedBuilder) NavigationFeed(sections []catalog.Section) ([]byte, error) {
	return f.navigationCatalog(sections, "/opds/nav", "Navigation")
}

// RootCatalogFeed — тот же навигационный каталог, но self/start на /opds (без /nav).
func (f *FeedBuilder) RootCatalogFeed(sections []catalog.Section) ([]byte, error) {
	return f.navigationCatalog(sections, CatalogRootPath, "Каталог")
}

func (f *FeedBuilder) navigationCatalog(sections []catalog.Section, selfPath, title string) ([]byte, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	feed := atomFeed{
		XmlnsOPDS: "http://opds-spec.org/2010/catalog",
		XmlnsDC:   xmlnsDC,
		ID:        f.join(selfPath),
		Title:     title,
		Updated:   now,
		Links: []atomLink{
			{Rel: "self", Href: f.join(selfPath), Type: NavigationCatalogMediaType},
			{Rel: "start", Href: f.join(CatalogRootPath), Type: NavigationCatalogMediaType},
			{Rel: "search", Href: f.join("/opds/opensearch.xml"), Type: "application/opensearchdescription+xml"},
		},
	}
	for _, s := range sections {
		sub := "/opds/section/" + url.PathEscape(s.ID) + "?" + SectionEntryQuery()
		feed.Entries = append(feed.Entries, atomEntry{
			ID:      f.join(sub),
			Title:   s.Title,
			Updated: now,
			Links: []atomLink{
				{Rel: "subsection", Href: f.join(sub), Type: AcquisitionFeedMediaType},
			},
		})
	}
	return marshalFeed(feed)
}

// AcquisitionFeed — фид раздела с постраничностью (RFC 5005: first, previous, next, last).
func (f *FeedBuilder) AcquisitionFeed(books []book.Book, sectionID string, total, page, limit int, rawQuery string) ([]byte, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	path := "/opds/section/" + url.PathEscape(sectionID)
	pageLinks := f.paginationLinks(path, total, page, limit, rawQuery)
	var links []atomLink
	links = append(links, pageLinks[0])
	links = append(links, atomLink{Rel: "start", Href: f.join(CatalogRootPath), Type: NavigationCatalogMediaType})
	links = append(links, atomLink{Rel: "up", Href: f.join(CatalogRootPath), Type: NavigationCatalogMediaType})
	// Иначе часть клиентов (в т.ч. Android) не показывают поиск, пока открыт список книг раздела.
	links = append(links, atomLink{Rel: "search", Href: f.join("/opds/opensearch.xml"), Type: "application/opensearchdescription+xml"})
	links = append(links, pageLinks[1:]...)

	feed := atomFeed{
		XmlnsOPDS: "http://opds-spec.org/2010/catalog",
		XmlnsDC:   xmlnsDC,
		ID:        pageLinks[0].Href,
		Title:     "Books",
		Updated:   now,
		Links:     links,
	}
	for _, b := range books {
		acq := f.join("/opds/acquire/" + url.PathEscape(string(b.ID)))
		feed.Entries = append(feed.Entries, acquisitionEntry(b, now, acq))
	}
	return marshalFeed(feed)
}

// SearchAcquisitionFeed — результаты поиска с той же моделью страниц.
func (f *FeedBuilder) SearchAcquisitionFeed(books []book.Book, total, page, limit int, rawQuery string) ([]byte, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	path := "/opds/search"
	pageLinks := f.paginationLinks(path, total, page, limit, rawQuery)
	var links []atomLink
	links = append(links, pageLinks[0])
	links = append(links, atomLink{Rel: "start", Href: f.join(CatalogRootPath), Type: NavigationCatalogMediaType})
	links = append(links, atomLink{Rel: "up", Href: f.join(CatalogRootPath), Type: NavigationCatalogMediaType})
	links = append(links, atomLink{Rel: "search", Href: f.join("/opds/opensearch.xml"), Type: "application/opensearchdescription+xml"})
	links = append(links, pageLinks[1:]...)

	feed := atomFeed{
		XmlnsOPDS: "http://opds-spec.org/2010/catalog",
		XmlnsDC:   xmlnsDC,
		ID:        pageLinks[0].Href,
		Title:     "Поиск",
		Updated:   now,
		Links:     links,
	}
	for _, b := range books {
		acq := f.join("/opds/acquire/" + url.PathEscape(string(b.ID)))
		feed.Entries = append(feed.Entries, acquisitionEntry(b, now, acq))
	}
	return marshalFeed(feed)
}

func (f *FeedBuilder) join(p string) string {
	base := strings.TrimRight(f.BaseURL, "/")
	return base + p
}

func marshalFeed(feed atomFeed) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")
	if err := enc.Encode(feed); err != nil {
		return nil, fmt.Errorf("encode atom: %w", err)
	}
	return buf.Bytes(), nil
}
