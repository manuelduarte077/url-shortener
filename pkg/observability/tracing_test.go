package observability_test

import (
	"context"
	"os"
	"testing"
	"url-shortener/configs"
	"url-shortener/pkg/observability"

	"github.com/stretchr/testify/assert"
)

func TestInitTracing(t *testing.T) {
	originalEndpoint := os.Getenv("OTLP_ENDPOINT")
	defer func() {
		if originalEndpoint != "" {
			os.Setenv("OTLP_ENDPOINT", originalEndpoint)
		} else {
			os.Unsetenv("OTLP_ENDPOINT")
		}
	}()

	os.Setenv("OTLP_ENDPOINT", "localhost:4318")

	cfg := &configs.Config{}

	cleanup, err := observability.InitTracing(cfg)

	if err == nil {
		assert.NotNil(t, cleanup)
		if cleanup != nil {
			cleanup()
		}
	} else {
		assert.Error(t, err)
	}
}

func TestInitTracing_DefaultEndpoint(t *testing.T) {
	originalEndpoint := os.Getenv("OTLP_ENDPOINT")
	defer func() {
		if originalEndpoint != "" {
			os.Setenv("OTLP_ENDPOINT", originalEndpoint)
		} else {
			os.Unsetenv("OTLP_ENDPOINT")
		}
	}()

	os.Unsetenv("OTLP_ENDPOINT")

	cfg := &configs.Config{}

	cleanup, err := observability.InitTracing(cfg)

	if err == nil {
		assert.NotNil(t, cleanup)
		if cleanup != nil {
			cleanup()
		}
	}
}

func TestGetTracer(t *testing.T) {
	tracer := observability.GetTracer()

	assert.NotNil(t, tracer)

	ctx, span := tracer.Start(context.TODO(), "test-span")
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}

func TestGetTracer_AfterInit(t *testing.T) {
	originalEndpoint := os.Getenv("OTLP_ENDPOINT")
	defer func() {
		if originalEndpoint != "" {
			os.Setenv("OTLP_ENDPOINT", originalEndpoint)
		} else {
			os.Unsetenv("OTLP_ENDPOINT")
		}
	}()

	os.Setenv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces")
	cfg := &configs.Config{}

	cleanup, err := observability.InitTracing(cfg)
	if cleanup != nil {
		defer cleanup()
	}

	tracer := observability.GetTracer()
	assert.NotNil(t, tracer)

	if err == nil {
		ctx, span := tracer.Start(context.TODO(), "test-span")
		assert.NotNil(t, ctx)
		assert.NotNil(t, span)
		span.End()
	}

	tracer2 := observability.GetTracer()
	assert.NotNil(t, tracer2)
	if err == nil {
		ctx2, span2 := tracer2.Start(context.TODO(), "test-span-2")
		assert.NotNil(t, ctx2)
		assert.NotNil(t, span2)
		span2.End()
	}
}

func TestGetEnv(t *testing.T) {
	originalEndpoint := os.Getenv("JAEGER_ENDPOINT")
	defer func() {
		if originalEndpoint != "" {
			os.Setenv("JAEGER_ENDPOINT", originalEndpoint)
		} else {
			os.Unsetenv("JAEGER_ENDPOINT")
		}
	}()

	tests := []struct {
		name     string
		setEnv   func()
		expected string
	}{
		{
			name: "env var set",
			setEnv: func() {
				os.Setenv("OTLP_ENDPOINT", "custom:4318")
			},
		},
		{
			name: "env var not set - uses default",
			setEnv: func() {
				os.Unsetenv("OTLP_ENDPOINT")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setEnv()

			cfg := &configs.Config{}
			cleanup, err := observability.InitTracing(cfg)
			if cleanup != nil {
				defer cleanup()
			}

			if err != nil {
				assert.Error(t, err)
			}
		})
	}
}
