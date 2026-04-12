package httpapi

import (
	"log/slog"
	"net/http"
)

// writeHTTPError пишет ответ с кодом status; текст err не уходит клиенту, если ExposeErrors == false.
func (h *OPDSHandler) writeHTTPError(w http.ResponseWriter, r *http.Request, status int, public string, err error) {
	if err != nil {
		h.log().Error(public, "err", err, "path", r.URL.Path, "status", status)
	} else {
		h.log().Error(public, "path", r.URL.Path, "status", status)
	}
	msg := public
	if h != nil && h.ExposeErrors && err != nil {
		msg = err.Error()
	}
	http.Error(w, msg, status)
}

func (h *OPDSHandler) log() *slog.Logger {
	if h != nil && h.Log != nil {
		return h.Log
	}
	return slog.Default()
}
