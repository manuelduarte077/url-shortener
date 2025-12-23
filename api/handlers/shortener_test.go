package handlers_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
	"url-shortener/api/handlers"
	"url-shortener/internal/application"
	"url-shortener/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockURLRepository struct {
	mock.Mock
}

func (m *MockURLRepository) Save(ctx context.Context, url *domain.URL) error {
	args := m.Called(ctx, url)
	return args.Error(0)
}

func (m *MockURLRepository) FindByShortCode(ctx context.Context, shortCode string) (*domain.URL, error) {
	args := m.Called(ctx, shortCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.URL), args.Error(1)
}

func (m *MockURLRepository) Exists(ctx context.Context, shortCode string) (bool, error) {
	args := m.Called(ctx, shortCode)
	return args.Bool(0), args.Error(1)
}

type MockShortCodeGenerator struct {
	mock.Mock
}

func (m *MockShortCodeGenerator) Generate() string {
	args := m.Called()
	return args.String(0)
}

func TestNewShortenerHandler(t *testing.T) {
	tmpl := template.Must(template.New("test").Parse("test"))
	repo := new(MockURLRepository)
	gen := new(MockShortCodeGenerator)
	service := application.NewShortenerService(repo, gen)

	handler := handlers.NewShortenerHandler(service, tmpl)

	assert.NotNil(t, handler)
}

func TestShortenerHandler_ShowForm(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		tmpl           *template.Template
		expectedStatus int
	}{
		{
			name:           "GET request - success",
			method:         http.MethodGet,
			tmpl:           template.Must(template.New("form.html").Parse(`<form>Test Form</form>`)),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST request - method not allowed",
			method:         http.MethodPost,
			tmpl:           template.Must(template.New("form.html").Parse(`<form>Test Form</form>`)),
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "GET request - template error",
			method:         http.MethodGet,
			tmpl:           template.Must(template.New("wrong.html").Parse(`<form>Test Form</form>`)), // Wrong template name
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockURLRepository)
			gen := new(MockShortCodeGenerator)
			service := application.NewShortenerService(repo, gen)
			handler := handlers.NewShortenerHandler(service, tt.tmpl)

			req := httptest.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()

			handler.ShowForm(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestShortenerHandler_CreateShortURL(t *testing.T) {
	tmpl := template.Must(template.New("result.html").Parse(`ShortCode: {{.ShortCode}}, ShortURL: {{.ShortURL}}`))

	tests := []struct {
		name           string
		method         string
		formData       url.Values
		setupMocks     func(*MockURLRepository, *MockShortCodeGenerator)
		expectedStatus int
	}{
		{
			name:   "POST request - success",
			method: http.MethodPost,
			formData: url.Values{
				"url": []string{"https://example.com"},
			},
			setupMocks: func(repo *MockURLRepository, gen *MockShortCodeGenerator) {
				gen.On("Generate").Return("abc123")
				repo.On("Exists", mock.Anything, "abc123").Return(false, nil)
				repo.On("Save", mock.Anything, mock.MatchedBy(func(u *domain.URL) bool {
					return u.ShortCode == "abc123" && u.LongURL == "https://example.com"
				})).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET request - method not allowed",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "POST request - empty URL",
			method: http.MethodPost,
			formData: url.Values{
				"url": []string{""},
			},
			setupMocks:     nil, // No mocks needed - validation fails before service call
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "POST request - invalid URL format",
			method: http.MethodPost,
			formData: url.Values{
				"url": []string{"not-a-url"},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "POST request - service error on save",
			method: http.MethodPost,
			formData: url.Values{
				"url": []string{"https://example.com"},
			},
			setupMocks: func(repo *MockURLRepository, gen *MockShortCodeGenerator) {
				gen.On("Generate").Return("abc123")
				repo.On("Exists", mock.Anything, "abc123").Return(false, nil)
				repo.On("Save", mock.Anything, mock.Anything).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "POST request - template error",
			method: http.MethodPost,
			formData: url.Values{
				"url": []string{"https://example.com"},
			},
			setupMocks: func(repo *MockURLRepository, gen *MockShortCodeGenerator) {
				gen.On("Generate").Return("abc123")
				repo.On("Exists", mock.Anything, "abc123").Return(false, nil)
				repo.On("Save", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "POST request - parse form error",
			method:         http.MethodPost,
			formData:       nil, // Will cause ParseForm to potentially fail with invalid body
			setupMocks:     nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockURLRepository)
			gen := new(MockShortCodeGenerator)
			if tt.setupMocks != nil {
				tt.setupMocks(repo, gen)
			}

			service := application.NewShortenerService(repo, gen)
			testTmpl := tmpl
			if tt.name == "POST request - template error" {
				testTmpl = template.Must(template.New("wrong.html").Parse(`Wrong template`))
			}
			handler := handlers.NewShortenerHandler(service, testTmpl)

			var req *http.Request
			if tt.formData != nil {
				req = httptest.NewRequest(tt.method, "/shorten", bytes.NewBufferString(tt.formData.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else if tt.name == "POST request - parse form error" {
				req = httptest.NewRequest(tt.method, "/shorten", bytes.NewBufferString("%"))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				req = httptest.NewRequest(tt.method, "/shorten", nil)
			}

			w := httptest.NewRecorder()

			handler.CreateShortURL(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			repo.AssertExpectations(t)
			gen.AssertExpectations(t)
		})
	}
}

func TestShortenerHandler_Redirect(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		setupMocks     func(*MockURLRepository, *MockShortCodeGenerator)
		expectedStatus int
		expectedURL    string
	}{
		{
			name:   "GET request - success",
			method: http.MethodGet,
			path:   "/abc123",
			setupMocks: func(repo *MockURLRepository, gen *MockShortCodeGenerator) {
				url := &domain.URL{
					ShortCode: "abc123",
					LongURL:   "https://example.com",
					CreatedAt: time.Now(),
				}
				repo.On("FindByShortCode", mock.Anything, "abc123").Return(url, nil)
			},
			expectedStatus: http.StatusMovedPermanently,
			expectedURL:    "https://example.com",
		},
		{
			name:           "POST request - method not allowed",
			method:         http.MethodPost,
			path:           "/abc123",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "GET request - empty path redirects to root",
			method:         http.MethodGet,
			path:           "/",
			expectedStatus: http.StatusFound,
		},
		{
			name:   "GET request - URL not found",
			method: http.MethodGet,
			path:   "/nonexistent",
			setupMocks: func(repo *MockURLRepository, gen *MockShortCodeGenerator) {
				repo.On("FindByShortCode", mock.Anything, "nonexistent").Return(nil, domain.ErrURLNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "GET request - expired URL",
			method: http.MethodGet,
			path:   "/expired123",
			setupMocks: func(repo *MockURLRepository, gen *MockShortCodeGenerator) {
				expiredTime := time.Now().Add(-1 * time.Hour)
				url := &domain.URL{
					ShortCode: "expired123",
					LongURL:   "https://example.com",
					CreatedAt: time.Now().Add(-2 * time.Hour),
					ExpiresAt: &expiredTime,
				}
				repo.On("FindByShortCode", mock.Anything, "expired123").Return(url, nil)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockURLRepository)
			gen := new(MockShortCodeGenerator)
			if tt.setupMocks != nil {
				tt.setupMocks(repo, gen)
			}

			service := application.NewShortenerService(repo, gen)
			tmpl := template.Must(template.New("test").Parse("test"))
			handler := handlers.NewShortenerHandler(service, tmpl)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.Redirect(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedURL != "" {
				assert.Equal(t, tt.expectedURL, w.Header().Get("Location"))
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestShortenerHandler_buildShortURL(t *testing.T) {
	tests := []struct {
		name     string
		request  *http.Request
		expected string
	}{
		{
			name: "HTTP request",
			request: &http.Request{
				Host: "localhost:8181",
				TLS:  nil,
			},
			expected: "http://localhost:8181/abc123",
		},
		{
			name: "HTTPS request",
			request: &http.Request{
				Host: "localhost:8181",
				TLS:  &tls.ConnectionState{},
			},
			expected: "https://localhost:8181/abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := template.Must(template.New("result.html").Parse(`{{.ShortURL}}`))
			repo := new(MockURLRepository)
			gen := new(MockShortCodeGenerator)
			service := application.NewShortenerService(repo, gen)
			handler := handlers.NewShortenerHandler(service, tmpl)

			gen.On("Generate").Return("abc123")
			repo.On("Exists", mock.Anything, "abc123").Return(false, nil)
			repo.On("Save", mock.Anything, mock.Anything).Return(nil)

			formData := url.Values{"url": []string{"https://example.com"}}
			req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Host = tt.request.Host
			req.TLS = tt.request.TLS

			w := httptest.NewRecorder()
			handler.CreateShortURL(w, req)

			if tt.request.TLS != nil {
				assert.Contains(t, w.Body.String(), "https://")
			} else {
				assert.Contains(t, w.Body.String(), "http://")
			}
		})
	}
}
