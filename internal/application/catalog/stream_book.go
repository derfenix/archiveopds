package catalog

import (
	"context"

	"git.derfenix.pro/derfenix/archiveopds/internal/application/port/outbound"
	"git.derfenix.pro/derfenix/archiveopds/internal/domain/book"
)

// StreamBook — открытие потока для скачивания (acquisition link).
type StreamBook struct {
	Nav outbound.ArchiveNavigator
}

func (u *StreamBook) Open(ctx context.Context, id book.ID) (outbound.ReadSeekCloser, int64, string, string, error) {
	return u.Nav.OpenReadSeeker(ctx, id)
}
