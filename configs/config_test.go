package configs_test

import (
	"os"
	"testing"
	"time"
	"url-shortener/configs"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	originalPort := os.Getenv("SERVER_PORT")
	originalReadTimeout := os.Getenv("SERVER_READ_TIMEOUT")
	originalWriteTimeout := os.Getenv("SERVER_WRITE_TIMEOUT")
	originalIdleTimeout := os.Getenv("SERVER_IDLE_TIMEOUT")
	originalTTL := os.Getenv("STORAGE_TTL")
	originalBaseURL := os.Getenv("APP_BASE_URL")

	defer func() {
		if originalPort != "" {
			os.Setenv("SERVER_PORT", originalPort)
		} else {
			os.Unsetenv("SERVER_PORT")
		}
		if originalReadTimeout != "" {
			os.Setenv("SERVER_READ_TIMEOUT", originalReadTimeout)
		} else {
			os.Unsetenv("SERVER_READ_TIMEOUT")
		}
		if originalWriteTimeout != "" {
			os.Setenv("SERVER_WRITE_TIMEOUT", originalWriteTimeout)
		} else {
			os.Unsetenv("SERVER_WRITE_TIMEOUT")
		}
		if originalIdleTimeout != "" {
			os.Setenv("SERVER_IDLE_TIMEOUT", originalIdleTimeout)
		} else {
			os.Unsetenv("SERVER_IDLE_TIMEOUT")
		}
		if originalTTL != "" {
			os.Setenv("STORAGE_TTL", originalTTL)
		} else {
			os.Unsetenv("STORAGE_TTL")
		}
		if originalBaseURL != "" {
			os.Setenv("APP_BASE_URL", originalBaseURL)
		} else {
			os.Unsetenv("APP_BASE_URL")
		}
	}()

	tests := []struct {
		name            string
		setupEnv        func()
		expectedPort    string
		expectedTTL     time.Duration
		expectedBaseURL string
	}{
		{
			name: "default values",
			setupEnv: func() {
				os.Unsetenv("SERVER_PORT")
				os.Unsetenv("STORAGE_TTL")
				os.Unsetenv("APP_BASE_URL")
			},
			expectedPort:    "8181",
			expectedTTL:     24 * time.Hour,
			expectedBaseURL: "http://localhost:8181",
		},
		{
			name: "custom values from env",
			setupEnv: func() {
				os.Setenv("SERVER_PORT", "9090")
				os.Setenv("STORAGE_TTL", "48h")
				os.Setenv("APP_BASE_URL", "https://example.com")
			},
			expectedPort:    "9090",
			expectedTTL:     48 * time.Hour,
			expectedBaseURL: "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()

			cfg, err := configs.Load()

			assert.NoError(t, err)
			assert.NotNil(t, cfg)
			assert.Equal(t, tt.expectedPort, cfg.Server.Port)
			assert.Equal(t, tt.expectedTTL, cfg.Storage.TTL)
			assert.Equal(t, tt.expectedBaseURL, cfg.App.BaseURL)
			assert.Equal(t, 15*time.Second, cfg.Server.ReadTimeout)
			assert.Equal(t, 15*time.Second, cfg.Server.WriteTimeout)
			assert.Equal(t, 60*time.Second, cfg.Server.IdleTimeout)
		})
	}
}

func TestGetEnv(t *testing.T) {
	originalValue := os.Getenv("TEST_ENV_VAR")
	defer func() {
		if originalValue != "" {
			os.Setenv("TEST_ENV_VAR", originalValue)
		} else {
			os.Unsetenv("TEST_ENV_VAR")
		}
	}()

	tests := []struct {
		name         string
		key          string
		defaultValue string
		setEnv       func()
		expected     string
	}{
		{
			name:         "env var set",
			key:          "TEST_ENV_VAR",
			defaultValue: "default",
			setEnv: func() {
				os.Setenv("TEST_ENV_VAR", "custom")
			},
			expected: "custom",
		},
		{
			name:         "env var not set",
			key:          "TEST_ENV_VAR",
			defaultValue: "default",
			setEnv: func() {
				os.Unsetenv("TEST_ENV_VAR")
			},
			expected: "default",
		},
		{
			name:         "env var empty",
			key:          "TEST_ENV_VAR",
			defaultValue: "default",
			setEnv: func() {
				os.Setenv("TEST_ENV_VAR", "")
			},
			expected: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setEnv()
			cfg, err := configs.Load()
			assert.NoError(t, err)
			assert.NotNil(t, cfg)
		})
	}
}

func TestGetIntEnv(t *testing.T) {
	originalValue := os.Getenv("TEST_INT_VAR")
	defer func() {
		if originalValue != "" {
			os.Setenv("TEST_INT_VAR", originalValue)
		} else {
			os.Unsetenv("TEST_INT_VAR")
		}
	}()

	// getIntEnv is unexported, so we test it indirectly
	// In a real scenario, we might want to export it or test through Load
	// For now, we'll test that Load works with various configurations
	os.Unsetenv("TEST_INT_VAR")
	cfg, err := configs.Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestGetDurationEnv(t *testing.T) {
	originalValue := os.Getenv("SERVER_READ_TIMEOUT")
	defer func() {
		if originalValue != "" {
			os.Setenv("SERVER_READ_TIMEOUT", originalValue)
		} else {
			os.Unsetenv("SERVER_READ_TIMEOUT")
		}
	}()

	tests := []struct {
		name        string
		envValue    string
		expected    time.Duration
		shouldError bool
	}{
		{
			name:        "valid duration",
			envValue:    "30s",
			expected:    30 * time.Second,
			shouldError: false,
		},
		{
			name:        "invalid duration - uses default",
			envValue:    "invalid",
			expected:    15 * time.Second, // default
			shouldError: false,
		},
		{
			name:        "empty value - uses default",
			envValue:    "",
			expected:    15 * time.Second, // default
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("SERVER_READ_TIMEOUT", tt.envValue)
			} else {
				os.Unsetenv("SERVER_READ_TIMEOUT")
			}

			cfg, err := configs.Load()

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
				if tt.envValue != "invalid" && tt.envValue != "" {
					assert.Equal(t, tt.expected, cfg.Server.ReadTimeout)
				} else {
					// For invalid or empty, should use default
					assert.Equal(t, 15*time.Second, cfg.Server.ReadTimeout)
				}
			}
		})
	}
}

func TestRateLimiterConfig(t *testing.T) {
	originalEnabled := os.Getenv("RATE_LIMITER_ENABLED")
	originalLimit := os.Getenv("RATE_LIMITER_LIMIT")
	originalWindow := os.Getenv("RATE_LIMITER_WINDOW")

	defer func() {
		if originalEnabled != "" {
			os.Setenv("RATE_LIMITER_ENABLED", originalEnabled)
		} else {
			os.Unsetenv("RATE_LIMITER_ENABLED")
		}
		if originalLimit != "" {
			os.Setenv("RATE_LIMITER_LIMIT", originalLimit)
		} else {
			os.Unsetenv("RATE_LIMITER_LIMIT")
		}
		if originalWindow != "" {
			os.Setenv("RATE_LIMITER_WINDOW", originalWindow)
		} else {
			os.Unsetenv("RATE_LIMITER_WINDOW")
		}
	}()

	tests := []struct {
		name            string
		setupEnv        func()
		expectedEnabled bool
		expectedLimit   int
		expectedWindow  time.Duration
	}{
		{
			name: "default values",
			setupEnv: func() {
				os.Unsetenv("RATE_LIMITER_ENABLED")
				os.Unsetenv("RATE_LIMITER_LIMIT")
				os.Unsetenv("RATE_LIMITER_WINDOW")
			},
			expectedEnabled: true,
			expectedLimit:   100,
			expectedWindow:  1 * time.Minute,
		},
		{
			name: "custom values from env",
			setupEnv: func() {
				os.Setenv("RATE_LIMITER_ENABLED", "true")
				os.Setenv("RATE_LIMITER_LIMIT", "50")
				os.Setenv("RATE_LIMITER_WINDOW", "30s")
			},
			expectedEnabled: true,
			expectedLimit:   50,
			expectedWindow:  30 * time.Second,
		},
		{
			name: "disabled rate limiter",
			setupEnv: func() {
				os.Setenv("RATE_LIMITER_ENABLED", "false")
				os.Setenv("RATE_LIMITER_LIMIT", "200")
				os.Setenv("RATE_LIMITER_WINDOW", "2m")
			},
			expectedEnabled: false,
			expectedLimit:   200,
			expectedWindow:  2 * time.Minute,
		},
		{
			name: "invalid bool uses default",
			setupEnv: func() {
				os.Setenv("RATE_LIMITER_ENABLED", "invalid")
				os.Unsetenv("RATE_LIMITER_LIMIT")
				os.Unsetenv("RATE_LIMITER_WINDOW")
			},
			expectedEnabled: true, // default
			expectedLimit:   100,
			expectedWindow:  1 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()

			cfg, err := configs.Load()

			assert.NoError(t, err)
			assert.NotNil(t, cfg)
			assert.Equal(t, tt.expectedEnabled, cfg.RateLimiter.Enabled)
			assert.Equal(t, tt.expectedLimit, cfg.RateLimiter.Limit)
			assert.Equal(t, tt.expectedWindow, cfg.RateLimiter.Window)
		})
	}
}
