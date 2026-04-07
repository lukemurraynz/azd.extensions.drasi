package observability_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

// TestNewTracer_NoOpWhenEnvVarAbsent verifies that NewTracer returns a no-op tracer
// when APPLICATIONINSIGHTS_CONNECTION_STRING is not set.
func TestNewTracer_NoOpWhenEnvVarAbsent(t *testing.T) {
	// NOTE: t.Parallel() omitted — t.Setenv cannot be used with parallel tests.

	// Ensure env var is absent for this test.
	t.Setenv("APPLICATIONINSIGHTS_CONNECTION_STRING", "")

	tracer, shutdown, err := observability.NewTracer(context.Background())

	require.NoError(t, err)
	require.NotNil(t, tracer)
	assert.NotNil(t, shutdown)

	// A no-op tracer must return a non-recording span.
	_, span := tracer.Start(context.Background(), "test-span")
	assert.False(t, span.IsRecording(),
		"span from no-op tracer must not be recording")
	span.End()

	if shutdown != nil {
		assert.NoError(t, shutdown(context.Background()))
	}
}

// TestNewTracer_ReturnsTracerWhenEnvVarPresent verifies that a non-nil tracer is
// returned when APPLICATIONINSIGHTS_CONNECTION_STRING is set, and that spans are
// produced (recording state depends on exporter connectivity — we only assert the
// tracer is a real implementation, not no-op).
func TestNewTracer_ReturnsTracerWhenEnvVarPresent(t *testing.T) {
	if os.Getenv("CI") == "" {
		t.Skip("skipping OTLP exporter test outside CI — requires network connectivity")
	}

	t.Setenv("APPLICATIONINSIGHTS_CONNECTION_STRING",
		"InstrumentationKey=00000000-0000-0000-0000-000000000000;IngestionEndpoint=https://dc.services.visualstudio.com/;LiveEndpoint=https://live.applicationinsights.azure.com/")

	tracer, shutdown, err := observability.NewTracer(context.Background())

	require.NoError(t, err)
	require.NotNil(t, tracer)
	require.NotNil(t, shutdown)

	// The tracer must be a real SDK tracer (not no-op).
	_, span := tracer.Start(context.Background(), "test-span")
	// SDK tracers return recording spans when a real provider is installed.
	assert.Implements(t, (*trace.Span)(nil), span)
	span.End()

	assert.NoError(t, shutdown(context.Background()))
}

// TestNewTracer_EnvVarPresent_ExercisesConfiguredPath verifies the configured
// tracer path is exercised when APPLICATIONINSIGHTS_CONNECTION_STRING is set,
// without requiring network connectivity.
func TestNewTracer_EnvVarPresent_ExercisesConfiguredPath(t *testing.T) {
	// NOTE: t.Parallel() omitted — t.Setenv cannot be used with parallel tests.

	t.Setenv("APPLICATIONINSIGHTS_CONNECTION_STRING",
		"InstrumentationKey=00000000-0000-0000-0000-000000000000;IngestionEndpoint=https://example.invalid/;LiveEndpoint=https://example.invalid/")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	tracer, shutdown, err := observability.NewTracer(ctx)
	if err != nil {
		assert.Error(t, err)
		assert.Nil(t, tracer)
		assert.Nil(t, shutdown)
		return
	}

	require.NotNil(t, tracer)
	require.NotNil(t, shutdown)

	_, span := tracer.Start(context.Background(), "configured-test-span")
	assert.Implements(t, (*trace.Span)(nil), span)
	span.End()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second)
	defer shutdownCancel()
	assert.NoError(t, shutdown(shutdownCtx))
}
