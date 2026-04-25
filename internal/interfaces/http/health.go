package httpapi

import (
	"net/http"
	"strings"

	"git.derfenix.pro/derfenix/archiveopds/internal/application/port/outbound"
)

// RegisterHealth регистрирует /healthz (живость) и /readyz (индекс при заданном archive).
func RegisterHealth(mux *http.ServeMux, archivePath string, nav outbound.ArchiveNavigator) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if _, err := w.Write([]byte("ok\n")); err != nil {
			return
		}
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if !catalogReady(archivePath, nav) {
			http.Error(w, "catalog not ready", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if _, err := w.Write([]byte("ok\n")); err != nil {
			return
		}
	})
}

func catalogReady(archivePath string, nav outbound.ArchiveNavigator) bool {
	if strings.TrimSpace(archivePath) == "" {
		return true
	}
	st, ok := nav.(outbound.NavigatorStats)
	if !ok {
		return false
	}
	return st.IndexedBookCount() > 0
}
