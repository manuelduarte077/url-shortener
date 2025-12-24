package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"url-shortener/internal/infrastructure/ratelimiter"
)

func TestRateLimitingMiddleware_Allow(t *testing.T) {
	t.Parallel()

	rl := ratelimiter.NewMemoryRateLimiter(5, 1*time.Second)
	defer rl.(*ratelimiter.MemoryRateLimiter).Close()

	handler := RateLimitingMiddleware(rl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	// Make requests within limit
	for i := 0; i < 5; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Request %d: got status %d, want %d", i+1, rr.Code, http.StatusOK)
		}
	}

	// Next request should be rate limited
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusTooManyRequests)
	}

	expectedBody := `{"error":"Rate limit exceeded","message":"Too many requests. Please try again later."}`
	// http.Error adds a newline, so we trim it
	gotBody := strings.TrimSpace(rr.Body.String())
	if gotBody != expectedBody {
		t.Errorf("got body %q, want %q", gotBody, expectedBody)
	}

	// Check Retry-After header
	if retryAfter := rr.Header().Get("Retry-After"); retryAfter != "60" {
		t.Errorf("got Retry-After %q, want %q", retryAfter, "60")
	}
}

func TestRateLimitingMiddleware_DifferentIPs(t *testing.T) {
	t.Parallel()

	rl := ratelimiter.NewMemoryRateLimiter(2, 1*time.Second)
	defer rl.(*ratelimiter.MemoryRateLimiter).Close()

	handler := RateLimitingMiddleware(rl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Different IPs should have separate limits
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "192.168.1.1:12345"

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "192.168.1.2:12345"

	// Both should be allowed
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr1.Code != http.StatusOK || rr2.Code != http.StatusOK {
		t.Errorf("got statuses %d, %d, want both %d", rr1.Code, rr2.Code, http.StatusOK)
	}
}

func TestRateLimitingMiddleware_XForwardedFor(t *testing.T) {
	t.Parallel()

	rl := ratelimiter.NewMemoryRateLimiter(2, 1*time.Second)
	defer rl.(*ratelimiter.MemoryRateLimiter).Close()

	handler := RateLimitingMiddleware(rl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	req.RemoteAddr = "192.168.1.1:12345"

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusOK)
	}

	// Make another request with same X-Forwarded-For
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("X-Forwarded-For", "10.0.0.1")
	req2.RemoteAddr = "192.168.1.2:12345" // Different RemoteAddr

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", rr2.Code, http.StatusOK)
	}

	// Third request should be rate limited (same X-Forwarded-For)
	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.Header.Set("X-Forwarded-For", "10.0.0.1")
	req3.RemoteAddr = "192.168.1.3:12345"

	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req3)

	if rr3.Code != http.StatusTooManyRequests {
		t.Errorf("got status %d, want %d", rr3.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimitingMiddleware_XRealIP(t *testing.T) {
	t.Parallel()

	rl := ratelimiter.NewMemoryRateLimiter(2, 1*time.Second)
	defer rl.(*ratelimiter.MemoryRateLimiter).Close()

	handler := RateLimitingMiddleware(rl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "10.0.0.1")
	req.RemoteAddr = "192.168.1.1:12345"

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestExtractIdentifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		request  *http.Request
		wantPref string // preferred identifier (X-Forwarded-For or X-Real-IP)
	}{
		{
			name: "X-Forwarded-For takes priority",
			request: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("X-Forwarded-For", "10.0.0.1")
				req.Header.Set("X-Real-IP", "10.0.0.2")
				req.RemoteAddr = "192.168.1.1:12345"
				return req
			}(),
			wantPref: "10.0.0.1",
		},
		{
			name: "X-Real-IP used when X-Forwarded-For not present",
			request: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("X-Real-IP", "10.0.0.2")
				req.RemoteAddr = "192.168.1.1:12345"
				return req
			}(),
			wantPref: "10.0.0.2",
		},
		{
			name: "RemoteAddr used as fallback",
			request: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.RemoteAddr = "192.168.1.1:12345"
				return req
			}(),
			wantPref: "192.168.1.1:12345",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := extractIdentifier(tt.request)
			if got != tt.wantPref {
				t.Errorf("extractIdentifier() = %q, want %q", got, tt.wantPref)
			}
		})
	}
}
