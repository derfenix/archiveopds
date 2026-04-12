package opds

import (
	"regexp"
	"strconv"
	"strings"

	"git.derfenix.pro/derfenix/archiveopds/internal/domain/book"
)

var stripTagsRE = regexp.MustCompile(`<[^>]*>`)

// publicationTitle — atom:title по OPDS: название произведения без дублирования автора.
func publicationTitle(b book.Book) string {
	if t := strings.TrimSpace(b.BookTitle); t != "" {
		return t
	}
	auth := strings.TrimSpace(b.Author)
	base := strings.TrimSpace(b.Title)
	if auth != "" {
		prefix := auth + " — "
		if strings.HasPrefix(base, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(base, prefix))
		}
	}
	if base != "" {
		return base
	}
	return strings.TrimSpace(b.RelPath)
}

func splitGenreTerms(genre string) []string {
	var out []string
	for _, p := range strings.Split(genre, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func plainAnnotation(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if strings.Contains(s, "<") {
		s = stripTagsRE.ReplaceAllString(s, " ")
	}
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}

func bookSummaryText(b book.Book) string {
	if a := plainAnnotation(b.Annotation); a != "" {
		return a
	}
	var lines []string
	if s := strings.TrimSpace(b.Series); s != "" {
		line := s
		if n := strings.TrimSpace(b.SeriesIndex); n != "" {
			line += " · № " + n
		}
		lines = append(lines, line)
	}
	if g := strings.TrimSpace(b.Genre); g != "" {
		lines = append(lines, g)
	}
	if b.Year > 0 {
		lines = append(lines, strconv.Itoa(b.Year))
	}
	if b.Size > 0 {
		lines = append(lines, formatByteSize(b.Size))
	}
	if lang := strings.TrimSpace(b.Language); lang != "" {
		lines = append(lines, lang)
	}
	return strings.Join(lines, " · ")
}

func formatByteSize(n int64) string {
	if n < 1024 {
		return strconv.FormatInt(n, 10) + " B"
	}
	if n < 1024*1024 {
		return strconv.FormatInt(n/1024, 10) + " KB"
	}
	return strconv.FormatInt(n/(1024*1024), 10) + " MB"
}

func acquisitionEntry(b book.Book, now, acqURL string) atomEntry {
	e := atomEntry{
		ID:      acqURL,
		Title:   publicationTitle(b),
		Updated: now,
		Links: []atomLink{
			{Rel: "http://opds-spec.org/acquisition", Href: acqURL, Type: b.MIMEType},
		},
	}
	if a := strings.TrimSpace(b.Author); a != "" {
		e.Authors = []atomPerson{{Name: a}}
	}
	for _, term := range splitGenreTerms(b.Genre) {
		e.Categories = append(e.Categories, atomCategory{Term: term, Label: term})
	}
	if lang := strings.TrimSpace(b.Language); lang != "" {
		e.DCLanguage = lang
	}
	if b.Year > 0 {
		e.DCIssued = strconv.Itoa(b.Year)
	}
	if mt := strings.TrimSpace(b.MIMEType); mt != "" {
		e.DCFormat = mt
	}
	if id := strings.TrimSpace(b.LibraryID); id != "" {
		e.DCIdentifier = id
	}
	sum := bookSummaryText(b)
	if sum != "" {
		e.Summary = &atomSummary{Type: "text", Text: sum}
	}
	return e
}
