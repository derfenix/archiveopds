package httpapi

import (
	"container/list"
	"net"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

// RateLimitConfig configures token-bucket rate limiting for HTTP.
type RateLimitConfig struct {
	RPS            float64
	Burst          int
	PerIP          bool
	TrustForwarded bool
	// MaxTrackedIPs caps distinct per-IP limiters; when full, the least-recently-used entry is evicted. 0 = no cap.
	MaxTrackedIPs int
}

// WithRateLimit enforces a token-bucket limit. When rps <= 0, next is returned unchanged.
// When PerIP is set, each client address has its own bucket; otherwise a single process-wide bucket is used.
func WithRateLimit(cfg RateLimitConfig, next http.Handler) http.Handler {
	if cfg.RPS <= 0 {
		return next
	}
	if cfg.Burst < 1 {
		cfg.Burst = 10
	}
	global := rate.NewLimiter(rate.Limit(cfg.RPS), cfg.Burst)
	var (
		mu   sync.Mutex
		byk  map[string]*rate.Limiter
		lru  *list.List
		lidx map[string]*list.Element
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lim := global
		if cfg.PerIP {
			ip := clientIPFromRequest(r, cfg.TrustForwarded)
			if ip != "" {
				mu.Lock()
				if byk == nil {
					byk = make(map[string]*rate.Limiter)
					lru = list.New()
					lidx = make(map[string]*list.Element)
				}
				l := byk[ip]
				if l != nil {
					if el := lidx[ip]; el != nil {
						lru.MoveToBack(el)
					}
				} else {
					if cfg.MaxTrackedIPs > 0 && len(byk) >= cfg.MaxTrackedIPs {
						if fr := lru.Front(); fr != nil {
							old := fr.Value.(string)
							lru.Remove(fr)
							delete(lidx, old)
							delete(byk, old)
						}
					}
					l = rate.NewLimiter(rate.Limit(cfg.RPS), cfg.Burst)
					byk[ip] = l
					el := lru.PushBack(ip)
					lidx[ip] = el
				}
				lim = l
				mu.Unlock()
			}
		}
		if !lim.Allow() {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func firstXForwardedFor(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if i := strings.Index(s, ","); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	return s
}

func clientIPFromRequest(r *http.Request, trustXFF bool) string {
	if trustXFF {
		if xff := firstXForwardedFor(r.Header.Get("X-Forwarded-For")); xff != "" {
			xff = strings.Trim(xff, "[]")
			if host, _, err := net.SplitHostPort(xff); err == nil {
				if net.ParseIP(host) != nil {
					return host
				}
			}
			if net.ParseIP(xff) != nil {
				return xff
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
