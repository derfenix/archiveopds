package inpx

import (
	"context"
	"strings"
	"unicode"

	"git.derfenix.pro/derfenix/archiveopds/internal/domain/book"
	"git.derfenix.pro/derfenix/archiveopds/internal/domain/catalog"
)

func (n *Navigator) SearchBooks(ctx context.Context, c catalog.SearchCriteria) ([]book.Book, int, error) {
	words := queryWords(c.Query)
	if !hasAnyCriterion(c, words) {
		return nil, 0, nil
	}

	lim := c.Page.Limit
	if lim <= 0 {
		lim = catalog.DefaultPageSize
	}
	if lim > catalog.MaxPageSize {
		lim = catalog.MaxPageSize
	}
	off := c.Page.Offset
	if off < 0 {
		off = 0
	}

	// When allBooks is empty (e.g. unit tests with only bySection, or a minimal stub), merge sections.
	src := n.allBooks
	if len(src) == 0 {
		for _, books := range n.bySection {
			src = append(src, books...)
		}
	}

	var out []book.Book
	matches := 0
	for _, b := range src {
		if err := ctx.Err(); err != nil {
			return out, matches, err
		}
		if !bookMatches(b, c, words) {
			continue
		}
		if matches >= off && len(out) < lim {
			out = append(out, b)
		}
		matches++
	}
	filled, err := n.attachFB2Annotations(ctx, out)
	if err != nil {
		return nil, matches, err
	}
	return filled, matches, nil
}

func hasAnyCriterion(c catalog.SearchCriteria, qwords []string) bool {
	if len(qwords) > 0 {
		return true
	}
	if strings.TrimSpace(c.Author) != "" || strings.TrimSpace(c.Title) != "" || strings.TrimSpace(c.Genre) != "" || strings.TrimSpace(c.Series) != "" {
		return true
	}
	return c.Year != 0
}

func queryWords(q string) []string {
	fields := strings.FieldsFunc(strings.TrimSpace(q), func(r rune) bool {
		return unicode.IsSpace(r) || r == ','
	})
	var w []string
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f != "" {
			w = append(w, strings.ToLower(f))
		}
	}
	return w
}

func bookMatches(b book.Book, c catalog.SearchCriteria, qwords []string) bool {
	if len(qwords) > 0 {
		hay := b.SearchBlob
		if hay == "" {
			hay = buildSearchBlob(b)
		}
		for _, word := range qwords {
			if !strings.Contains(hay, word) {
				return false
			}
		}
	}
	if s := strings.TrimSpace(c.Author); s != "" {
		if !containsFold(b.Author, s) {
			return false
		}
	}
	if s := strings.TrimSpace(c.Title); s != "" {
		if !containsFold(b.BookTitle, s) && !containsFold(b.Title, s) {
			return false
		}
	}
	if s := strings.TrimSpace(c.Genre); s != "" {
		if !containsFold(b.Genre, s) {
			return false
		}
	}
	if s := strings.TrimSpace(c.Series); s != "" {
		if !containsFold(b.Series, s) {
			return false
		}
	}
	if c.Year != 0 {
		if b.Year != c.Year {
			return false
		}
	}
	return true
}

func containsFold(hay, needle string) bool {
	return strings.Contains(strings.ToLower(hay), strings.ToLower(needle))
}
