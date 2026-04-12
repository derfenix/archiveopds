package httpapi

import "net/http"

// ServerBanner — значение заголовка Server (Calibre OPDS Reader читает feed.headers['server']).
const ServerBanner = "archiveopds"

// WithServerHeader оборачивает handler и всегда выставляет заголовок Server.
func WithServerHeader(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", ServerBanner)
		h.ServeHTTP(w, r)
	})
}
