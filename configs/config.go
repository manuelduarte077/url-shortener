package configs

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server      ServerConfig
	Storage     StorageConfig
	App         AppConfig
	RateLimiter RateLimiterConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type StorageConfig struct {
	TTL time.Duration // time to live for stored URLs
}

type AppConfig struct {
	BaseURL string
}

type RateLimiterConfig struct {
	Enabled bool
	Limit   int
	Window  time.Duration
}

func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8181"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Storage: StorageConfig{
			TTL: getDurationEnv("STORAGE_TTL", 24*time.Hour),
		},
		App: AppConfig{
			BaseURL: getEnv("APP_BASE_URL", "http://localhost:8181"),
		},
		RateLimiter: RateLimiterConfig{
			Enabled: getBoolEnv("RATE_LIMITER_ENABLED", true),
			Limit:   getIntEnv("RATE_LIMITER_LIMIT", 100),
			Window:  getDurationEnv("RATE_LIMITER_WINDOW", 1*time.Minute),
		},
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
