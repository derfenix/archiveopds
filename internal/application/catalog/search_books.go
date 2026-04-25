package catalog

import (
	"context"

	"git.derfenix.pro/derfenix/archiveopds/internal/application/port/outbound"
	"git.derfenix.pro/derfenix/archiveopds/internal/domain/book"
	domaincatalog "git.derfenix.pro/derfenix/archiveopds/internal/domain/catalog"
)

// SearchBooks — поиск по индексу (автор, название, жанр, год, общий запрос q).
type SearchBooks struct {
	Nav outbound.ArchiveNavigator
}

func (u *SearchBooks) Execute(ctx context.Context, c domaincatalog.SearchCriteria) ([]book.Book, int, error) {
	return u.Nav.SearchBooks(ctx, c)
}
