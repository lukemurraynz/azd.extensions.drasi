package observability_test

import (
	"context"
	"testing"

	"github.com/azure/azd.extensions.drasi/internal/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewMeter_NoOpWhenEnvVarAbsent verifies that NewMeter returns a no-op meter
// when APPLICATIONINSIGHTS_CONNECTION_STRING is not set.
func TestNewMeter_NoOpWhenEnvVarAbsent(t *testing.T) {
	// NOTE: t.Parallel() omitted — t.Setenv cannot be used with parallel tests.

	t.Setenv("APPLICATIONINSIGHTS_CONNECTION_STRING", "")

	meter, shutdown, err := observability.NewMeter(context.Background())

	require.NoError(t, err)
	require.NotNil(t, meter)
	assert.NotNil(t, shutdown)

	if shutdown != nil {
		assert.NoError(t, shutdown(context.Background()))
	}
}

// TestNewMeter_CounterNames verifies that the three required counters can be
// created from the meter without error. The counter names are defined in the spec.
func TestNewMeter_CounterNames(t *testing.T) {
	// NOTE: t.Parallel() omitted — t.Setenv cannot be used with parallel tests.

	t.Setenv("APPLICATIONINSIGHTS_CONNECTION_STRING", "")

	meter, shutdown, err := observability.NewMeter(context.Background())
	require.NoError(t, err)
	require.NotNil(t, meter)

	// drasi.components.deployed
	deployedCounter, err := meter.Int64Counter("drasi.components.deployed",
		// intentionally no unit/description — just verify name is accepted
	)
	require.NoError(t, err, "drasi.components.deployed counter must be creatable")
	assert.NotNil(t, deployedCounter)

	// drasi.deploy.errors
	errorsCounter, err := meter.Int64Counter("drasi.deploy.errors")
	require.NoError(t, err, "drasi.deploy.errors counter must be creatable")
	assert.NotNil(t, errorsCounter)

	// drasi.deploy.duration_seconds — histogram, not counter
	durationHistogram, err := meter.Float64Histogram("drasi.deploy.duration_seconds")
	require.NoError(t, err, "drasi.deploy.duration_seconds histogram must be creatable")
	assert.NotNil(t, durationHistogram)

	if shutdown != nil {
		assert.NoError(t, shutdown(context.Background()))
	}
}
