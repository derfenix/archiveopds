package outbound

import (
	"context"

	"git.derfenix.pro/derfenix/archiveopds/internal/domain/book"
	"git.derfenix.pro/derfenix/archiveopds/internal/domain/catalog"
)

// ArchiveNavigator — выходной порт: чтение больших архивов без загрузки всего содержимого в память.
type ArchiveNavigator interface {
	RootSections(ctx context.Context) ([]catalog.Section, error)
	ListBooks(ctx context.Context, sectionID string, page catalog.Page) ([]book.Book, int, error)
	SearchBooks(ctx context.Context, criteria catalog.SearchCriteria) ([]book.Book, int, error)
	// OpenReadSeeker открывает поток книги; downloadName — имя файла для Content-Disposition (может быть пустым).
	OpenReadSeeker(ctx context.Context, bookID book.ID) (ReadSeekCloser, int64, string, string, error)
}

// ReadSeekCloser — поток байт книги для выдачи по HTTP (acquisition).
type ReadSeekCloser interface {
	Read(p []byte) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
	Close() error
}

// NavigatorStats — опционально: размер проиндексированного каталога (для /readyz).
type NavigatorStats interface {
	IndexedBookCount() int
}
