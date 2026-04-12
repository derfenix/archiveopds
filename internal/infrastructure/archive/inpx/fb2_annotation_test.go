package inpx

import (
	"strings"
	"testing"
)

func TestExtractFB2Annotation(t *testing.T) {
	t.Parallel()
	const xml = `<?xml version="1.0" encoding="UTF-8"?>
<FictionBook xmlns="http://www.gribuser.ru/xml/fictionbook/2.0">
<description>
<title-info>
<annotation>
<p>Первый абзац.</p>
<empty-line/>
<p>Второй <emphasis>важный</emphasis> фрагмент.</p>
</annotation>
</title-info>
</description>
</FictionBook>`
	got := extractFB2Annotation([]byte(xml))
	if !strings.Contains(got, "Первый") || !strings.Contains(got, "важный") {
		t.Fatalf("unexpected annotation: %q", got)
	}
}

func TestFinalizeAnnotation_truncate(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("а", maxAnnotationRunes+100)
	out := finalizeAnnotation(long)
	if utf8Count(out) > maxAnnotationRunes+3 { // …
		t.Fatalf("too long: %d", utf8Count(out))
	}
}

func utf8Count(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}
