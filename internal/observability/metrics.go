package observability

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// NewMeter returns an OpenTelemetry meter and a shutdown function.
// If APPLICATIONINSIGHTS_CONNECTION_STRING is absent, a no-op meter is used.
// When present, a real MeterProvider is configured. Call shutdown to flush metrics.
func NewMeter(ctx context.Context) (metric.Meter, func(context.Context) error, error) {
	connStr := os.Getenv(appInsightsEnvVar)
	if connStr == "" {
		m := metricnoop.NewMeterProvider().Meter("drasi")
		return m, noopShutdown, nil
	}

	provider := sdkmetric.NewMeterProvider()
	shutdown := func(ctx context.Context) error {
		return provider.Shutdown(ctx)
	}
	return provider.Meter("drasi"), shutdown, nil
}
