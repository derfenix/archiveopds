package inpx

import (
	"context"

	"git.derfenix.pro/derfenix/archiveopds/internal/application/port/outbound"
	"git.derfenix.pro/derfenix/archiveopds/internal/domain/book"
	"git.derfenix.pro/derfenix/archiveopds/internal/domain/catalog"
)

var (
	_ outbound.ArchiveNavigator = (*Navigator)(nil)
	_ outbound.NavigatorStats   = (*Navigator)(nil)
)

func (n *Navigator) RootSections(_ context.Context) ([]catalog.Section, error) {
	return n.sections, nil
}

func (n *Navigator) ListBooks(ctx context.Context, sectionID string, page catalog.Page) ([]book.Book, int, error) {
	var full []book.Book
	switch {
	case sectionID == FlatCatalogSectionID:
		full = n.allBooks
	default:
		if b, ok := n.bySection[sectionID]; ok {
			full = b
		} else {
			return nil, 0, nil
		}
	}
	total := len(full)
	lim := page.Limit
	if lim <= 0 {
		lim = catalog.DefaultPageSize
	}
	if lim > catalog.MaxPageSize {
		lim = catalog.MaxPageSize
	}
	off := page.Offset
	if off < 0 {
		off = 0
	}
	if off > total {
		off = total
	}
	end := off + lim
	if end > total {
		end = total
	}
	pageBooks := full[off:end]
	filled, err := n.attachFB2Annotations(ctx, pageBooks)
	if err != nil {
		return nil, 0, err
	}
	return filled, total, nil
}

// IndexedBookCount — число записей в объединённом индексе (для /readyz).
func (n *Navigator) IndexedBookCount() int { return len(n.allBooks) }

func (n *Navigator) OpenReadSeeker(ctx context.Context, id book.ID) (outbound.ReadSeekCloser, int64, string, string, error) {
	rc, sz, ct, err := openFromZip(ctx, n.root, id)
	if err != nil {
		return nil, 0, "", "", err
	}
	fn := ""
	if b, ok := n.byID[id]; ok {
		fn = downloadFilename(b)
	}
	return rc, sz, ct, fn, nil
}
