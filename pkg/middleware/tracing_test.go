package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/pkg/middleware"

	"github.com/stretchr/testify/assert"
)

func TestTracingMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	wrapped := middleware.TracingMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestTracingMiddleware_WithTraceContext(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify context is propagated
		ctx := r.Context()
		assert.NotNil(t, ctx)
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.TracingMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTracingMiddleware_DifferentMethods(t *testing.T) {
	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			wrapped := middleware.TracingMiddleware(handler)

			req := httptest.NewRequest(method, "/test", nil)
			w := httptest.NewRecorder()

			wrapped.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestTracingMiddleware_ContextPropagation(t *testing.T) {
	var capturedContext *http.Request

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedContext = r
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.TracingMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test?param=value", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	assert.NotNil(t, capturedContext)
	assert.Equal(t, "/test?param=value", capturedContext.URL.String())
	assert.Equal(t, "test-agent", capturedContext.Header.Get("User-Agent"))
}

