package observability

import (
	"context"
	"net/url"
	"os"
	"strings"
	"time"
	"url-shortener/configs"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer trace.Tracer
)

func InitTracing(cfg *configs.Config) (func(), error) {
	otlpEndpointStr := getEnv("OTLP_ENDPOINT", "localhost:4318")
	ctx := context.Background()

	endpoint := otlpEndpointStr
	isInsecure := true

	// If endpoint contains http:// or https://, parse it
	if strings.HasPrefix(otlpEndpointStr, "http://") || strings.HasPrefix(otlpEndpointStr, "https://") {
		parsedURL, err := url.Parse(otlpEndpointStr)
		if err != nil {
			return nil, err
		}
		endpoint = parsedURL.Host
		isInsecure = parsedURL.Scheme == "http"
	}

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpoint),
	}
	if isInsecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	exp, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("url-shortener"),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer = otel.Tracer("url-shortener")

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			// log error but don't fail - this is cleanup
		}
	}, nil
}

func GetTracer() trace.Tracer {
	if tracer == nil {
		return otel.Tracer("url-shortener")
	}
	return tracer
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
