package domain_test

import (
	"testing"
	"time"
	"url-shortener/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestURL_Validate(t *testing.T) {
	tests := []struct {
		name    string
		url     *domain.URL
		wantErr bool
	}{
		{
			name: "valid URL",
			url: &domain.URL{
				ShortCode: "abc123",
				LongURL:   "https://example.com",
			},
			wantErr: false,
		},
		{
			name: "empty long URL",
			url: &domain.URL{
				ShortCode: "abc123",
				LongURL:   "",
			},
			wantErr: true,
		},
		{
			name: "empty short code",
			url: &domain.URL{
				ShortCode: "",
				LongURL:   "https://example.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.url.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestURL_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		url      *domain.URL
		expected bool
	}{
		{
			name: "not expired - no expiry time",
			url: &domain.URL{
				ExpiresAt: nil,
			},
			expected: false,
		},
		{
			name: "not expired - future expiry",
			url: &domain.URL{
				ExpiresAt: func() *time.Time {
					t := time.Now().Add(1 * time.Hour)
					return &t
				}(),
			},
			expected: false,
		},
		{
			name: "expired - past expiry",
			url: &domain.URL{
				ExpiresAt: func() *time.Time {
					t := time.Now().Add(-1 * time.Hour)
					return &t
				}(),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.url.IsExpired()
			assert.Equal(t, tt.expected, result)
		})
	}
}
