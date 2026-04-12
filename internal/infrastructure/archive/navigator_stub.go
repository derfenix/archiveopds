package archive

import (
	"context"

	"git.derfenix.pro/derfenix/archiveopds/internal/application/port/outbound"
	"git.derfenix.pro/derfenix/archiveopds/internal/domain/book"
	"git.derfenix.pro/derfenix/archiveopds/internal/domain/catalog"
	domerr "git.derfenix.pro/derfenix/archiveopds/internal/domain/errors"
)

// NavigatorStub — заглушка: позже заменить на потоковое чтение zip/tar/7z и индексацию.
type NavigatorStub struct{}

var (
	_ outbound.ArchiveNavigator = (*NavigatorStub)(nil)
	_ outbound.NavigatorStats   = (*NavigatorStub)(nil)
)

func NewNavigatorStub() *NavigatorStub {
	return &NavigatorStub{}
}

func (n *NavigatorStub) RootSections(_ context.Context) ([]catalog.Section, error) {
	return []catalog.Section{{ID: "root", Title: "Каталог"}}, nil
}

func (n *NavigatorStub) ListBooks(_ context.Context, sectionID string, _ catalog.Page) ([]book.Book, int, error) {
	if sectionID != "" && sectionID != "root" {
		return nil, 0, nil
	}
	return nil, 0, nil
}

func (n *NavigatorStub) SearchBooks(_ context.Context, _ catalog.SearchCriteria) ([]book.Book, int, error) {
	return nil, 0, nil
}

func (n *NavigatorStub) OpenReadSeeker(_ context.Context, _ book.ID) (outbound.ReadSeekCloser, int64, string, string, error) {
	return nil, 0, "", "", domerr.ErrNotFound
}

// IndexedBookCount всегда 0 — заглушка без индекса.
func (n *NavigatorStub) IndexedBookCount() int { return 0 }
