package catalog

// Постраничная выдача фидов (Atom/RFC 5005: rel first, previous, next, last).
const (
	DefaultPageSize = 100
	MaxPageSize     = 500
)

// Page задаёт срез выдачи: Offset и Limit (элементов на страницу).
type Page struct {
	Offset int
	Limit  int
}
