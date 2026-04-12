package inpx

import (
	"path"
	"regexp"
	"strings"
	"unicode/utf8"

	"git.derfenix.pro/derfenix/archiveopds/internal/domain/book"
)

var (
	commaNoSpaceRE = regexp.MustCompile(`,([^\s,])`)
	multiCommaRE   = regexp.MustCompile(`,\s*,+`)
)

// downloadFilename — «Автор1, Автор2. Название.fb2» для заголовка Content-Disposition.
func downloadFilename(b book.Book) string {
	ext := fileExtForBook(b)
	auth := normalizeAuthorField(b.Author)
	title := strings.TrimSpace(b.BookTitle)
	if title == "" {
		title = titleFromDisplay(b.Title, b.Author)
	}
	title = strings.TrimSpace(title)
	if title == "" {
		title = "book"
	}

	var base string
	switch {
	case auth != "":
		base = auth + ". " + title
	default:
		base = title
	}
	base = sanitizeFileBase(base)
	if base == "" {
		base = "book"
	}
	return base + ext
}

func normalizeAuthorField(author string) string {
	author = strings.TrimSpace(author)
	if author == "" {
		return ""
	}
	author = strings.ReplaceAll(author, " : ", ", ")
	author = strings.ReplaceAll(author, "; ", ", ")
	author = strings.ReplaceAll(author, ";", ", ")
	for commaNoSpaceRE.MatchString(author) {
		author = commaNoSpaceRE.ReplaceAllString(author, `, $1`)
	}
	author = strings.Trim(author, ", ")
	author = multiCommaRE.ReplaceAllString(author, ", ")
	return strings.TrimSpace(author)
}

func titleFromDisplay(display, author string) string {
	display = strings.TrimSpace(display)
	if display == "" {
		return ""
	}
	prefix := strings.TrimSpace(author) + " — "
	if author != "" && strings.HasPrefix(display, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(display, prefix))
	}
	prefix2 := strings.TrimSpace(author) + " - "
	if author != "" && strings.HasPrefix(display, prefix2) {
		return strings.TrimSpace(strings.TrimPrefix(display, prefix2))
	}
	return display
}

func fileExtForBook(b book.Book) string {
	if e := path.Ext(b.RelPath); e != "" {
		return strings.ToLower(e)
	}
	switch {
	case strings.Contains(b.MIMEType, "fb2"):
		return ".fb2"
	case strings.Contains(b.MIMEType, "epub"):
		return ".epub"
	case strings.Contains(b.MIMEType, "pdf"):
		return ".pdf"
	case strings.Contains(b.MIMEType, "mobi"):
		return ".mobi"
	default:
		return ".bin"
	}
}

var invalidFileRunes = regexp.MustCompile(`[\x00-\x1f\\/:*?"<>|]+`)

func sanitizeFileBase(s string) string {
	s = invalidFileRunes.ReplaceAllString(s, " ")
	s = strings.Join(strings.Fields(s), " ")
	s = strings.TrimRight(s, ". ")
	if utf8.RuneCountInString(s) > 200 {
		r := []rune(s)
		s = string(r[:200])
		s = strings.TrimRight(s, ". ")
	}
	return s
}
