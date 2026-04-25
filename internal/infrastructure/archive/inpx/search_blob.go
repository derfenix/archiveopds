package inpx

import (
	"strings"

	"git.derfenix.pro/derfenix/archiveopds/internal/domain/book"
)

func buildSearchBlob(b book.Book) string {
	return strings.ToLower(strings.Join([]string{
		b.Author, b.BookTitle, b.Genre, b.Title,
		b.Series, b.SeriesIndex, b.Annotation, b.LibraryID, b.Language,
	}, " "))
}
