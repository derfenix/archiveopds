package inpx

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
	"sync"

	"git.derfenix.pro/derfenix/archiveopds/internal/domain/book"
	"git.derfenix.pro/derfenix/archiveopds/internal/domain/catalog"
)

// Navigator читает INPX из корня библиотеки и строит навигацию по внешним .zip томам.
type Navigator struct {
	root      string
	sections  []catalog.Section
	allBooks  []book.Book // объединённый каталог (сортировка по Title)
	bySection map[string][]book.Book
	byID      map[book.ID]book.Book

	// annCache — ленивое чтение <annotation> из .fb2 в zip (см. annSlot).
	annCache sync.Map
	// annWorkers — ограничение параллельных обращений к zip при дочитывании аннотаций.
	annWorkers int
}

// Load сканирует все inpxPaths (файлы .inpx) и объединяет индекс.
// annotationWorkers — число параллельных читателей FB2-аннотаций (минимум 1).
func Load(root string, inpxPaths []string, annotationWorkers int) (*Navigator, error) {
	if annotationWorkers < 1 {
		annotationWorkers = 1
	}
	by := make(map[string][]book.Book)
	for _, px := range inpxPaths {
		if err := ingestInpx(root, px, by); err != nil {
			return nil, fmt.Errorf("%s: %w", px, err)
		}
	}
	if len(by) == 0 {
		return nil, fmt.Errorf("в INPX не найдено записей с существующими .zip в %q", root)
	}

	var allBooks []book.Book
	for _, books := range by {
		allBooks = append(allBooks, books...)
	}
	sort.Slice(allBooks, func(i, j int) bool { return allBooks[i].Title < allBooks[j].Title })

	sections := []catalog.Section{{
		ID:    FlatCatalogSectionID,
		Title: "Все книги",
	}}

	byID := make(map[book.ID]book.Book)
	for _, books := range by {
		for _, b := range books {
			byID[b.ID] = b
		}
	}

	return &Navigator{
		root:         root,
		sections:     sections,
		allBooks:     allBooks,
		bySection:    by,
		byID:         byID,
		annWorkers:   annotationWorkers,
	}, nil
}

func ingestInpx(root, inpxPath string, dest map[string][]book.Book) error {
	zr, err := zip.OpenReader(inpxPath)
	if err != nil {
		return fmt.Errorf("открыть inpx: %w", err)
	}
	defer zr.Close()

	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := path.Base(f.Name)
		if !strings.EqualFold(path.Ext(name), ".inp") {
			continue
		}
		inpStem := strings.TrimSuffix(name, path.Ext(name))
		if err := readInpFile(f, inpStem, root, dest); err != nil {
			return fmt.Errorf("%s: %w", f.Name, err)
		}
	}
	return nil
}

func readInpFile(f *zip.File, inpStem, root string, dest map[string][]book.Book) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	sc := bufio.NewScanner(rc)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 4*1024*1024)

	for sc.Scan() {
		rec, ok := parseRecord(sc.Text(), inpStem, root)
		if !ok {
			continue
		}
		b := book.Book{
			ID:          encodeBookRef(rec.SectionStem, rec.InnerPath),
			Title:       rec.DisplayTitle,
			Author:      rec.Author,
			BookTitle:   rec.BookTitle,
			Genre:       rec.Genre,
			Year:        rec.Year,
			RelPath:     rec.InnerPath,
			MIMEType:    MIMEForIndexInner(rec.InnerPath),
			Size:        rec.Size,
			Series:      rec.Series,
			SeriesIndex: rec.SeriesIndex,
			Language:    rec.Language,
			Annotation:  rec.Annotation,
			LibraryID:   rec.LibraryID,
		}
		dest[rec.SectionStem] = append(dest[rec.SectionStem], b)
	}
	if err := sc.Err(); err != nil && err != io.EOF {
		return err
	}
	return nil
}
