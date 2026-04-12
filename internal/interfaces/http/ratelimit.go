package httpapi

import (
	"net/http"

	"golang.org/x/time/rate"
)

// WithRateLimit — глобальный лимитер на весь процесс (token bucket).
// При rps <= 0 возвращает next без изменений.
func WithRateLimit(rps float64, burst int, next http.Handler) http.Handler {
	if rps <= 0 {
		return next
	}
	if burst < 1 {
		burst = 10
	}
	lim := rate.NewLimiter(rate.Limit(rps), burst)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !lim.Allow() {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
