package media

import "testing"

func TestRefineContentType(t *testing.T) {
	t.Parallel()
	if got := RefineContentType("a.fb2", "application/octet-stream"); got != "application/fb2+xml" {
		t.Fatalf("fb2: got %q", got)
	}
	if got := RefineContentType("a.fb2", "application/fb2+xml"); got != "application/fb2+xml" {
		t.Fatalf("keep specific: got %q", got)
	}
	if got := RefineContentType("book.bin", "application/octet-stream"); got != "application/octet-stream" {
		t.Fatalf("bin: got %q", got)
	}
}
