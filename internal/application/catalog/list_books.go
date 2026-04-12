package catalog

import (
	"context"

	"git.derfenix.pro/derfenix/archiveopds/internal/application/port/outbound"
	"git.derfenix.pro/derfenix/archiveopds/internal/domain/book"
	domaincatalog "git.derfenix.pro/derfenix/archiveopds/internal/domain/catalog"
)

// ListBooks — сценарий прикладного уровня: список книг в разделе.
type ListBooks struct {
	Nav outbound.ArchiveNavigator
}

func (u *ListBooks) Execute(ctx context.Context, sectionID string, page domaincatalog.Page) ([]book.Book, int, error) {
	return u.Nav.ListBooks(ctx, sectionID, page)
}
