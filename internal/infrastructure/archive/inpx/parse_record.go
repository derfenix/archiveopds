package inpx

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ParsedINPRecord — результат разбора одной строки .inp после привязки к существующему .zip в root.
type ParsedINPRecord struct {
	SectionStem  string // имя тома без .zip
	Author       string
	Genre        string
	BookTitle    string
	DisplayTitle string // для списков: «автор — название» или путь
	InnerPath    string // путь к файлу внутри архива
	Year         int
	Size         int64

	Series      string
	SeriesIndex string
	LibraryID   string
	Language    string
	Annotation  string
}

func parseRecord(line, inpStem, root string) (ParsedINPRecord, bool) {
	line = strings.ToValidUTF8(line, "")
	if line == "" {
		return ParsedINPRecord{}, false
	}
	fields := splitFields(line)
	if len(fields) < 6 {
		return ParsedINPRecord{}, false
	}

	author := strings.TrimSpace(fields[0])
	genre := strings.TrimSpace(fields[1])
	bookTitle := strings.TrimSpace(fields[2])
	inner := strings.TrimSpace(fields[5])
	inner = strings.Trim(inner, "\r\n")
	inner = filepath.ToSlash(inner)
	if inner == "" {
		return ParsedINPRecord{}, false
	}

	size := parseInt64Field(fields[6])
	year := yearFromINPFields(fields)

	stemFromINP := normalizeZipStem(inpStem)
	var stemFromField string
	if len(fields) >= 9 {
		stemFromField = normalizeZipStem(strings.TrimSpace(fields[8]))
	}

	var sectionStem string
	switch {
	case zipExists(root, stemFromINP):
		sectionStem = stemFromINP
	case zipExists(root, stemFromField):
		sectionStem = stemFromField
	default:
		return ParsedINPRecord{}, false
	}

	displayTitle := bookTitle
	if displayTitle == "" {
		displayTitle = inner
	}
	if author != "" {
		displayTitle = author + " — " + displayTitle
	}

	s, si, lid, lang, ann := inpMetadata(fields)
	return ParsedINPRecord{
		SectionStem:  sectionStem,
		Author:       author,
		Genre:        genre,
		BookTitle:    bookTitle,
		DisplayTitle: displayTitle,
		InnerPath:    inner,
		Year:         year,
		Size:         size,
		Series:       s,
		SeriesIndex:  si,
		LibraryID:    lid,
		Language:     lang,
		Annotation:   ann,
	}, true
}

func normalizeZipStem(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(strings.TrimSuffix(s, ".zip"), ".ZIP")
	return s
}

func zipExists(root, stem string) bool {
	if stem == "" {
		return false
	}
	st, err := os.Stat(filepath.Join(root, stem+".zip"))
	return err == nil && !st.IsDir()
}

func parseInt64Field(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}
