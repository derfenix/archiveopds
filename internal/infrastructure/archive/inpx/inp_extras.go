package inpx

import "strings"

// inpMetadata — дополнительные поля строки .inp (Flibusta / LibRusEc, разделитель 0x04 или таб).
// Базовые индексы: 0 автор, 1 жанр, 2 название, 3–4 серия/номер, 5 путь, 6 размер, 7 id, 8 архив;
// при расширенном формате часто 12 — язык, 15 — аннотация (см. 16+ полей).
func inpMetadata(fields []string) (series, seriesIndex, libraryID, language, annotation string) {
	if len(fields) > 3 {
		series = strings.TrimSpace(fields[3])
	}
	if len(fields) > 4 {
		seriesIndex = strings.TrimSpace(fields[4])
	}
	if len(fields) > 7 {
		libraryID = strings.TrimSpace(fields[7])
	}
	if len(fields) > 12 {
		language = strings.TrimSpace(fields[12])
	}
	if len(fields) > 15 {
		annotation = strings.TrimSpace(fields[15])
	}
	return
}
