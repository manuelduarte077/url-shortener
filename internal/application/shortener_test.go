package application_test

import (
	"context"
	"testing"
	"time"
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

func TestShortenerService_CreateShortURL(t *testing.T) {
	tests := []struct {
		name          string
		longURL       string
		setupMocks    func(*MockURLRepository, *MockShortCodeGenerator)
		expectedError bool
	}{
		{
			name:    "successful creation",
			longURL: "https://example.com",
			setupMocks: func(repo *MockURLRepository, gen *MockShortCodeGenerator) {
				gen.On("Generate").Return("abc12345")
				repo.On("Exists", mock.Anything, "abc12345").Return(false, nil)
				repo.On("Save", mock.Anything, mock.MatchedBy(func(url *domain.URL) bool {
					return url.ShortCode == "abc12345" && url.LongURL == "https://example.com"
				})).Return(nil)
			},
			expectedError: false,
		},
		{
			name:    "empty URL",
			longURL: "",
			setupMocks: func(repo *MockURLRepository, gen *MockShortCodeGenerator) {
				gen.On("Generate").Return("abc12345")
				repo.On("Exists", mock.Anything, "abc12345").Return(false, nil)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockURLRepository)
			gen := new(MockShortCodeGenerator)
			tt.setupMocks(repo, gen)

			service := application.NewShortenerService(repo, gen)
			ctx := context.Background()

			result, err := service.CreateShortURL(ctx, tt.longURL)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.longURL, result.LongURL)
			}

			repo.AssertExpectations(t)
			gen.AssertExpectations(t)
		})
	}
}

func TestShortenerService_GetLongURL(t *testing.T) {
	tests := []struct {
		name          string
		shortCode     string
		setupMocks    func(*MockURLRepository)
		expectedURL   string
		expectedError bool
	}{
		{
			name:      "successful retrieval",
			shortCode: "abc12345",
			setupMocks: func(repo *MockURLRepository) {
				url := &domain.URL{
					ShortCode: "abc12345",
					LongURL:   "https://example.com",
					CreatedAt: time.Now(),
				}
				repo.On("FindByShortCode", mock.Anything, "abc12345").Return(url, nil)
			},
			expectedURL:   "https://example.com",
			expectedError: false,
		},
		{
			name:      "not found",
			shortCode: "nonexistent",
			setupMocks: func(repo *MockURLRepository) {
				repo.On("FindByShortCode", mock.Anything, "nonexistent").Return(nil, domain.ErrURLNotFound)
			},
			expectedURL:   "",
			expectedError: true,
		},
		{
			name:      "expired URL",
			shortCode: "expired123",
			setupMocks: func(repo *MockURLRepository) {
				expiredTime := time.Now().Add(-1 * time.Hour)
				url := &domain.URL{
					ShortCode: "expired123",
					LongURL:   "https://example.com",
					CreatedAt: time.Now().Add(-2 * time.Hour),
					ExpiresAt: &expiredTime,
				}
				repo.On("FindByShortCode", mock.Anything, "expired123").Return(url, nil)
			},
			expectedURL:   "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockURLRepository)
			gen := new(MockShortCodeGenerator)
			tt.setupMocks(repo)

			service := application.NewShortenerService(repo, gen)
			ctx := context.Background()

			result, err := service.GetLongURL(ctx, tt.shortCode)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURL, result)
			}

			repo.AssertExpectations(t)
		})
	}
}
