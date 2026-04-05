package observability

import (
	"context"
	"os"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
)

// NewMeter returns an OpenTelemetry meter and a shutdown function.
//
// When APPLICATIONINSIGHTS_CONNECTION_STRING is absent the meter is a no-op.
// When the env var is present a real SDK MeterProvider is configured.
// The caller must invoke shutdown before process exit to flush pending metrics.
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
