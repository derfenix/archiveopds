package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientIPFromRequest(t *testing.T) {
	t.Parallel()
	r := httptest.NewRequest(http.MethodGet, "http://example.com/x", nil)
	r.RemoteAddr = "192.168.1.2:12345"
	if got := clientIPFromRequest(r, false); got != "192.168.1.2" {
		t.Fatalf("RemoteAddr: got %q", got)
	}
	r2 := r.Clone(t.Context())
	r2.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")
	if got := clientIPFromRequest(r2, false); got != "192.168.1.2" {
		t.Fatalf("without trust: got %q", got)
	}
	r3 := r.Clone(t.Context())
	r3.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")
	if got := clientIPFromRequest(r3, true); got != "10.0.0.1" {
		t.Fatalf("with trust: got %q", got)
	}
}

func TestFirstXForwardedFor(t *testing.T) {
	t.Parallel()
	if s := firstXForwardedFor(" 10.0.0.1 , 9.9.9.9"); s != "10.0.0.1" {
		t.Fatalf("got %q", s)
	}
}

func TestRateLimit_perIPMapCapResets(t *testing.T) {
	t.Parallel()
	var hits int
	h := WithRateLimit(RateLimitConfig{
		RPS: 1000, Burst: 10, PerIP: true, MaxTrackedIPs: 2,
	}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hits++ }))
	ips := []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}
	for i, ip := range ips {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ip + ":1234"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("i=%d ip=%s code %d", i, ip, rec.Code)
		}
	}
	if hits != 3 {
		t.Fatalf("hits=%d", hits)
	}
}
