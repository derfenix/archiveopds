package inpx

import (
	"context"
	"strings"
	"sync"

	"git.derfenix.pro/derfenix/archiveopds/internal/domain/book"
)

// annSlot — результат одной попытки прочитать аннотацию из FB2.
type annSlot struct {
	done bool   // уже обращались к архиву
	text string // непусто, если из FB2 извлекли текст (перекрывает поле из .inp)
}

func isLikelyFB2(b book.Book) bool {
	if strings.Contains(strings.ToLower(b.MIMEType), "fb2") {
		return true
	}
	return strings.HasSuffix(strings.ToLower(strings.TrimSpace(b.RelPath)), ".fb2")
}

// attachFB2Annotations подставляет оригинальную <annotation> из .fb2 (приоритет над кратким текстом из .inp).
func (n *Navigator) attachFB2Annotations(ctx context.Context, books []book.Book) ([]book.Book, error) {
	if len(books) == 0 {
		return books, nil
	}
	out := append([]book.Book(nil), books...)

	if n.annWorkers <= 1 {
		for i := range out {
			if err := ctx.Err(); err != nil {
				return out, err
			}
			n.attachOneFB2Annotation(ctx, &out[i])
		}
		return out, nil
	}

	type job struct{ idx int }
	var jobs []job
	for i := range out {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		b := &out[i]
		if !isLikelyFB2(*b) {
			continue
		}
		if v, ok := n.annCache.Load(b.ID); ok {
			slot := v.(annSlot)
			if slot.done {
				if slot.text != "" {
					b.Annotation = slot.text
				}
				continue
			}
		}
		jobs = append(jobs, job{idx: i})
	}

	sem := make(chan struct{}, n.annWorkers)
	var wg sync.WaitGroup
	for _, j := range jobs {
		j := j
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if err := ctx.Err(); err != nil {
				return
			}
			n.attachOneFB2Annotation(ctx, &out[j.idx])
		}()
	}
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return out, err
	}
	return out, nil
}

func (n *Navigator) attachOneFB2Annotation(ctx context.Context, b *book.Book) {
	if !isLikelyFB2(*b) {
		return
	}
	if v, ok := n.annCache.Load(b.ID); ok {
		slot := v.(annSlot)
		if slot.done {
			if slot.text != "" {
				b.Annotation = slot.text
			}
			return
		}
	}
	stem, inner, ok := decodeBookRef(b.ID)
	if !ok {
		n.annCache.Store(b.ID, annSlot{done: true})
		return
	}
	ann, err := readFB2AnnotationFromZip(ctx, n.root, stem, inner)
	if err != nil {
		n.annCache.Store(b.ID, annSlot{done: true})
		return
	}
	if strings.TrimSpace(ann) != "" {
		n.annCache.Store(b.ID, annSlot{done: true, text: ann})
		b.Annotation = ann
		return
	}
	n.annCache.Store(b.ID, annSlot{done: true})
}
