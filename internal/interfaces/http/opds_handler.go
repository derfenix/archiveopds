package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	appcatalog "git.derfenix.pro/derfenix/archiveopds/internal/application/catalog"
	"git.derfenix.pro/derfenix/archiveopds/internal/domain/book"
	domaincatalog "git.derfenix.pro/derfenix/archiveopds/internal/domain/catalog"
	domerr "git.derfenix.pro/derfenix/archiveopds/internal/domain/errors"
	"git.derfenix.pro/derfenix/archiveopds/internal/infrastructure/media"
	"git.derfenix.pro/derfenix/archiveopds/internal/infrastructure/opds"
)

// OPDSHandler — входной адаптер HTTP → прикладные сценарии.
type OPDSHandler struct {
	Log *slog.Logger
	// ExposeErrors — отдавать клиенту текст внутренних ошибок (только для отладки).
	ExposeErrors bool

	ListSections *appcatalog.ListSections
	ListBooks    *appcatalog.ListBooks
	SearchBooks  *appcatalog.SearchBooks
	StreamBook   *appcatalog.StreamBook
	Feed         *opds.FeedBuilder
}

func (h *OPDSHandler) Register(mux *http.ServeMux) {
	// Корень каталога — типичный URL для клиентов (Calibre OPDS-reader: …/opds).
	mux.HandleFunc("GET /opds", h.handleCatalogRoot)
	mux.HandleFunc("GET /opds/", h.handleCatalogRoot)
	mux.HandleFunc("GET /opds/nav", h.handleNav)
	mux.HandleFunc("GET /opds/section/", h.handleSectionPrefix)
	mux.HandleFunc("GET /opds/search", h.handleSearch)
	mux.HandleFunc("GET /opds/opensearch.xml", h.handleOpenSearch)
	mux.HandleFunc("GET /opds/acquire/", h.handleAcquirePrefix)
}

func (h *OPDSHandler) handleCatalogRoot(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	if p != "/opds" && p != "/opds/" {
		h.log().Debug("opds catalog root: unexpected path", "path", p)
		http.NotFound(w, r)
		return
	}
	h.log().Debug("opds catalog root")
	sections, err := h.ListSections.Execute(r.Context())
	if err != nil {
		h.writeHTTPError(w, r, http.StatusInternalServerError, "internal error", err)
		return
	}
	body, err := h.Feed.RootCatalogFeed(sections)
	if err != nil {
		h.writeHTTPError(w, r, http.StatusInternalServerError, "internal error", err)
		return
	}
	h.log().Debug("opds catalog root: ok", "sections", len(sections), "body_bytes", len(body))
	w.Header().Set("Content-Type", opds.NavigationCatalogMediaType+";charset=utf-8")
	_, _ = w.Write(body)
}

func (h *OPDSHandler) handleNav(w http.ResponseWriter, r *http.Request) {
	h.log().Debug("opds nav")
	sections, err := h.ListSections.Execute(r.Context())
	if err != nil {
		h.writeHTTPError(w, r, http.StatusInternalServerError, "internal error", err)
		return
	}
	body, err := h.Feed.NavigationFeed(sections)
	if err != nil {
		h.writeHTTPError(w, r, http.StatusInternalServerError, "internal error", err)
		return
	}
	h.log().Debug("opds nav: ok", "sections", len(sections), "body_bytes", len(body))
	w.Header().Set("Content-Type", opds.NavigationCatalogMediaType+";charset=utf-8")
	_, _ = w.Write(body)
}

func (h *OPDSHandler) handleSectionPrefix(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/opds/section/")
	id = strings.Trim(id, "/")
	if id == "" {
		h.log().Debug("opds section: empty id")
		http.NotFound(w, r)
		return
	}
	q := r.URL.Query()
	page, limit := parseFeedPage(q)
	h.log().Debug("opds section", "section_id", id, "page", page, "limit", limit)
	q.Set("page", strconv.Itoa(page))
	q.Set("limit", strconv.Itoa(limit))
	raw := q.Encode()

	books, total, err := h.ListBooks.Execute(r.Context(), id, domaincatalog.Page{
		Offset: pageToOffset(page, limit),
		Limit:  limit,
	})
	if err != nil {
		h.writeHTTPError(w, r, http.StatusInternalServerError, "internal error", err)
		return
	}
	body, err := h.Feed.AcquisitionFeed(books, id, total, page, limit, raw)
	if err != nil {
		h.writeHTTPError(w, r, http.StatusInternalServerError, "internal error", err)
		return
	}
	h.log().Debug("opds section: ok",
		"section_id", id, "books_page", len(books), "total", total, "body_bytes", len(body))
	w.Header().Set("Content-Type", opds.AcquisitionFeedMediaType+";charset=utf-8")
	_, _ = w.Write(body)
}

func (h *OPDSHandler) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, pageLimit := parseFeedPage(q)

	c := domaincatalog.SearchCriteria{
		Query:  q.Get("q"),
		Author: q.Get("author"),
		Title:  q.Get("title"),
		Genre:  q.Get("genre"),
		Series: q.Get("series"),
		Page: domaincatalog.Page{
			Offset: pageToOffset(page, pageLimit),
			Limit:  pageLimit,
		},
	}
	if ys := strings.TrimSpace(q.Get("year")); ys != "" {
		if y, err := strconv.Atoi(ys); err == nil {
			c.Year = y
		}
	}
	h.log().Debug("opds search",
		"page", page, "limit", pageLimit,
		"q", c.Query, "author", c.Author, "title", c.Title, "genre", c.Genre, "series", c.Series, "year", c.Year,
	)

	books, total, err := h.SearchBooks.Execute(r.Context(), c)
	if err != nil {
		h.writeHTTPError(w, r, http.StatusInternalServerError, "internal error", err)
		return
	}
	q.Set("page", strconv.Itoa(page))
	q.Set("limit", strconv.Itoa(pageLimit))
	body, err := h.Feed.SearchAcquisitionFeed(books, total, page, pageLimit, q.Encode())
	if err != nil {
		h.writeHTTPError(w, r, http.StatusInternalServerError, "internal error", err)
		return
	}
	h.log().Debug("opds search: ok", "hits_page", len(books), "total", total, "body_bytes", len(body))
	w.Header().Set("Content-Type", opds.AcquisitionFeedMediaType+";charset=utf-8")
	_, _ = w.Write(body)
}

func (h *OPDSHandler) handleOpenSearch(w http.ResponseWriter, r *http.Request) {
	h.log().Debug("opds opensearch.xml")
	body, err := opds.OpenSearchDescription(h.Feed.BaseURL)
	if err != nil {
		h.writeHTTPError(w, r, http.StatusInternalServerError, "internal error", err)
		return
	}
	h.log().Debug("opds opensearch: ok", "body_bytes", len(body))
	w.Header().Set("Content-Type", "application/opensearchdescription+xml;charset=utf-8")
	_, _ = w.Write(body)
}

func (h *OPDSHandler) handleAcquirePrefix(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimPrefix(r.URL.Path, "/opds/acquire/")
	raw = strings.Trim(raw, "/")
	if raw == "" {
		h.log().Debug("opds acquire: empty id")
		http.NotFound(w, r)
		return
	}
	h.log().Debug("opds acquire", "book_id_prefix", shortID(raw, 48), "book_id_len", len(raw))
	rc, size, contentType, downloadName, err := h.StreamBook.Open(r.Context(), book.ID(raw))
	if err != nil {
		if errors.Is(err, domerr.ErrNotFound) {
			h.log().Debug("opds acquire: not found", "book_id_prefix", shortID(raw, 48))
			http.NotFound(w, r)
			return
		}
		if errors.Is(err, domerr.ErrBookTooLarge) {
			h.log().Warn("opds acquire: book too large", "book_id_prefix", shortID(raw, 48), "err", err)
			h.writeHTTPError(w, r, http.StatusRequestEntityTooLarge, "book too large", err)
			return
		}
		if errors.Is(err, domerr.ErrArchiveOpen) {
			h.writeHTTPError(w, r, http.StatusServiceUnavailable, "archive unavailable", err)
			return
		}
		h.writeHTTPError(w, r, http.StatusInternalServerError, "internal error", err)
		return
	}
	defer rc.Close()

	serveName := downloadName
	if serveName == "" {
		serveName = acquisitionFilename(raw, contentType)
	}
	// Иначе часть клиентов берут имя только из URL /opds/acquire/x1… без расширения → «.unk».
	w.Header().Set("Content-Disposition", contentDispositionAttachment(serveName))

	ctOut := media.RefineContentType(serveName, contentType)
	if ctOut != "" {
		w.Header().Set("Content-Type", ctOut)
	}
	if size > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	}
	w.Header().Set("Accept-Ranges", "bytes")
	h.log().Debug("opds acquire: serving",
		"content_type", ctOut, "size", size, "download_name", downloadName, "serve_name", serveName)
	http.ServeContent(w, r, serveName, time.Unix(0, 0), rc)
}

func shortID(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

// acquisitionFilename — имя для ServeContent (тип по Content-Type уже выставлен выше).
func acquisitionFilename(_ string, contentType string) string {
	ct := strings.ToLower(contentType)
	switch {
	case strings.Contains(ct, "fb2"):
		return "book.fb2"
	case strings.Contains(ct, "epub"):
		return "book.epub"
	case strings.Contains(ct, "pdf"):
		return "book.pdf"
	case strings.Contains(ct, "mobipocket"), strings.Contains(ct, "mobi"):
		return "book.mobi"
	default:
		return "book.bin"
	}
}
