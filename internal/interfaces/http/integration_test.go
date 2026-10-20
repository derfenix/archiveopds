package httpapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appcatalog "git.derfenix.pro/derfenix/archiveopds/internal/application/catalog"
	"git.derfenix.pro/derfenix/archiveopds/internal/infrastructure/archive"
	"git.derfenix.pro/derfenix/archiveopds/internal/infrastructure/opds"
)

func TestHealthEndpoints_withStubNavigator(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	nav := archive.NewNavigatorStub()
	RegisterHealth(mux, "", nav)

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	for _, path := range []string{"/healthz", "/readyz"} {
		res, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("%s: %s", path, res.Status)
		}
	}
}

func TestOPDSRootCatalog_containsCatalogTitle(t *testing.T) {
	t.Parallel()
	srv := newOPDSTestServer(t)

	res, err := http.Get(srv.URL + "/opds")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %s", res.Status)
	}
	// тело — Atom; достаточно стабильной подстроки
	buf := make([]byte, 16384)
	n, _ := res.Body.Read(buf)
	body := string(buf[:n])
	if !strings.Contains(body, "<feed") || !strings.Contains(body, "Каталог") {
		t.Fatalf("unexpected body prefix: %q", truncate(body, 400))
	}
}

// Навигационные записи должны нести <content> и title у ссылки: часть клиентов
// (Moon+ Reader) иначе показывает фид как обычный XML-документ.
func TestOPDSRootCatalog_navigationEntriesHaveContent(t *testing.T) {
	t.Parallel()
	res, err := http.Get(newOPDSTestServer(t).URL + "/opds")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`<content type="text">Каталог</content>`,
		`rel="subsection"`,
		`title="Каталог"`,
		`<name>archiveopds</name>`,
		`kind=navigation`,
	} {
		if !strings.Contains(string(body), want) {
			t.Fatalf("feed has no %q:\n%s", want, body)
		}
	}
}

func TestLandingPage_linksToOPDSCatalog(t *testing.T) {
	t.Parallel()
	srv := newOPDSTestServer(t)
	res, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %s", res.Status)
	}
	if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("content-type %q", ct)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`rel="alternate"`,
		opds.NavigationCatalogMediaType,
		`href="http://127.0.0.1/opds"`,
	} {
		if !strings.Contains(string(body), want) {
			t.Fatalf("landing page has no %q:\n%s", want, body)
		}
	}
}

func TestFavicon_served(t *testing.T) {
	t.Parallel()
	res, err := http.Get(newOPDSTestServer(t).URL + "/favicon.ico")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %s", res.Status)
	}
	if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "image/svg+xml") {
		t.Fatalf("content-type %q", ct)
	}
}

func newOPDSTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	nav := archive.NewNavigatorStub()
	mux := http.NewServeMux()
	h := &OPDSHandler{
		ListSections: &appcatalog.ListSections{Nav: nav},
		ListBooks:    &appcatalog.ListBooks{Nav: nav},
		SearchBooks:  &appcatalog.SearchBooks{Nav: nav},
		StreamBook:   &appcatalog.StreamBook{Nav: nav},
		Feed:         &opds.FeedBuilder{BaseURL: "http://127.0.0.1"},
	}
	h.Register(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func TestCatalogReady_archiveSetEmptyIndex(t *testing.T) {
	t.Parallel()
	if catalogReady("/data", archive.NewNavigatorStub()) {
		t.Fatal("expected not ready when archive set but index empty")
	}
	if !catalogReady("", archive.NewNavigatorStub()) {
		t.Fatal("expected ready when archive not configured")
	}
}
