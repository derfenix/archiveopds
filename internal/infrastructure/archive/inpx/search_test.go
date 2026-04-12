package inpx

import (
	"context"
	"testing"

	"git.derfenix.pro/derfenix/archiveopds/internal/domain/book"
	"git.derfenix.pro/derfenix/archiveopds/internal/domain/catalog"
)

func TestBookMatches_queryWords(t *testing.T) {
	t.Parallel()
	b := book.Book{
		Author:    "Громов",
		BookTitle: "Первый",
		Genre:     "sf_action",
		Title:     "Громов — Первый",
		Year:      2020,
	}
	words := queryWords("громов первый")
	if !bookMatches(b, catalog.SearchCriteria{}, words) {
		t.Fatal("expected match")
	}
	if bookMatches(b, catalog.SearchCriteria{}, queryWords("несуществующее")) {
		t.Fatal("unexpected match")
	}
}

func TestBookMatches_seriesAndIndex(t *testing.T) {
	t.Parallel()
	b := book.Book{
		Series:      "Попаданцы в тело",
		SeriesIndex: "7",
		BookTitle:   "Часть первая",
		Title:       "Автор — Часть первая",
	}
	if !bookMatches(b, catalog.SearchCriteria{}, queryWords("попаданцы 7")) {
		t.Fatal("q must match series and series index")
	}
	if !bookMatches(b, catalog.SearchCriteria{Series: "тело"}, queryWords("")) {
		t.Fatal("series filter substring")
	}
	if bookMatches(b, catalog.SearchCriteria{Series: "другой цикл"}, queryWords("")) {
		t.Fatal("series filter should reject")
	}
}

func TestNavigator_SearchBooks(t *testing.T) {
	t.Parallel()
	n := &Navigator{
		bySection: map[string][]book.Book{
			"a": {
				{Author: "Автор1", BookTitle: "Книга А", Genre: "жанр1", Year: 2001, Title: "Автор1 — Книга А"},
				{Author: "Автор2", BookTitle: "Книга Б", Genre: "жанр2", Year: 2002, Title: "Автор2 — Книга Б"},
				{Author: "Автор3", BookTitle: "В серии", Genre: "жанр1", Year: 2003, Title: "Автор3 — В серии", Series: "Мир Земли", SeriesIndex: "3"},
			},
		},
	}
	ctx := context.Background()

	out, total, err := n.SearchBooks(ctx, catalog.SearchCriteria{Author: "Автор2"})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(out) != 1 || out[0].BookTitle != "Книга Б" {
		t.Fatalf("got %+v total=%d", out, total)
	}

	out, total, err = n.SearchBooks(ctx, catalog.SearchCriteria{Genre: "жанр1", Year: 2001})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(out) != 1 || out[0].BookTitle != "Книга А" {
		t.Fatalf("got %+v total=%d", out, total)
	}

	out, total, err = n.SearchBooks(ctx, catalog.SearchCriteria{Series: "Мир Земли"})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(out) != 1 || out[0].BookTitle != "В серии" {
		t.Fatalf("series search: got %+v total=%d", out, total)
	}

	out, total, err = n.SearchBooks(ctx, catalog.SearchCriteria{})
	if err != nil || len(out) != 0 || total != 0 {
		t.Fatalf("empty criteria: err=%v len=%d total=%d", err, len(out), total)
	}
}
