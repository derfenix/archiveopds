package inpx

import "testing"

func TestInpMetadata_extendedRow(t *testing.T) {
	t.Parallel()
	fields := make([]string, 16)
	fields[3] = "Миры"
	fields[4] = "12"
	fields[7] = "n123456"
	fields[12] = "ru"
	fields[15] = "Краткое описание произведения."

	s, si, lid, lang, ann := inpMetadata(fields)
	if s != "Миры" || si != "12" || lid != "n123456" || lang != "ru" || ann != "Краткое описание произведения." {
		t.Fatalf("got %q %q %q %q %q", s, si, lid, lang, ann)
	}
}
