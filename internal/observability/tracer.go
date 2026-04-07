package observability

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

const appInsightsEnvVar = "APPLICATIONINSIGHTS_CONNECTION_STRING"

// NewTracer returns an OpenTelemetry tracer and a shutdown function.
// If APPLICATIONINSIGHTS_CONNECTION_STRING is absent, a no-op tracer is returned
// with a no-op shutdown. If present, an OTLP/HTTP tracer provider is configured
// and should be shutdown on exit to flush spans.
func NewTracer(ctx context.Context) (trace.Tracer, func(context.Context) error, error) {
	connStr := os.Getenv(appInsightsEnvVar)
	if connStr == "" {
		p := noop.NewTracerProvider()
		return p.Tracer("drasi"), noopShutdown, nil
	}

	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("azd-drasi"),
		),
	)
	if err != nil {
		// Non-fatal: resource detection failure is not a hard stop.
		res = resource.Default()
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	shutdown := func(ctx context.Context) error {
		return provider.Shutdown(ctx)
	}

	return provider.Tracer("drasi"), shutdown, nil
}

func noopShutdown(_ context.Context) error { return nil }
