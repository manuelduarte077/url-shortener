package middleware

import (
	"net/http"
	"url-shortener/pkg/observability"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

func TracingMiddleware(next http.Handler) http.Handler {
	tracer := observability.GetTracer()
	propagator := otel.GetTextMapPropagator()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx, span := tracer.Start(ctx, r.Method+" "+r.URL.Path,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPRequestMethodOriginal(r.Method),
				semconv.URLFull(r.URL.String()),
				semconv.URLScheme(r.URL.Scheme),
				semconv.ServerAddress(r.Host),
				semconv.UserAgentOriginal(r.UserAgent()),
			),
		)
		defer span.End()

		propagator.Inject(ctx, propagation.HeaderCarrier(r.Header))
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
