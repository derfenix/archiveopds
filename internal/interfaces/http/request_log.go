package httpapi

import (
	"log/slog"
	"net/http"
	"time"
)

// WithRequestLogger пишет сводку каждого запроса (info) и детали (debug).
func WithRequestLogger(log *slog.Logger, h http.Handler) http.Handler {
	if log == nil {
		return h
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := &responseWriterLog{ResponseWriter: w, status: http.StatusOK}
		h.ServeHTTP(lw, r)
		ms := time.Since(start).Milliseconds()
		log.Info("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", lw.status,
			"duration_ms", ms,
			"bytes", lw.n,
			"remote", r.RemoteAddr,
		)
		log.Debug("http_detail",
			"query", r.URL.RawQuery,
			"user_agent", r.UserAgent(),
			"referer", r.Referer(),
		)
	})
}

type responseWriterLog struct {
	http.ResponseWriter
	status int
	n      int64
}

func (w *responseWriterLog) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriterLog) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.n += int64(n)
	return n, err
}
