package book

// ID стабильный идентификатор записи в каталоге (например, хэш пути или UUID из индекса).
type ID string

// Book — сущность: одна книга внутри архивного хранилища.
type Book struct {
	ID        ID
	Title     string // отображаемое: обычно «автор — название»
	Author    string
	BookTitle string // название без автора (для поиска и метаданных)
	Genre     string
	Year      int // 0 если не удалось извлечь из индекса
	RelPath   string
	MIMEType  string
	Size      int64

	// Из расширенной строки .inp (если есть в дампе).
	Series      string
	SeriesIndex string
	Language    string // код или название, как в индексе
	Annotation  string
	LibraryID   string // внутренний id каталога (например Flibusta)
}
