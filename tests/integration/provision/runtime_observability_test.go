//go:build integration

package provision_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRuntimeObservabilityBicep_ContainsOpenTelemetryExporterSettings verifies that the
// runtime observability module defines OpenTelemetry exporter settings so the provision
// command can configure Drasi runtime workloads to emit telemetry to Log Analytics.
func TestRuntimeObservabilityBicep_ContainsOpenTelemetryExporterSettings(t *testing.T) {
	t.Parallel()

	bicepPath := filepath.Join("..", "..", "..", "infra", "modules", "drasi-runtime-observability.bicep")
	content, err := os.ReadFile(bicepPath)
	require.NoError(t, err, "drasi-runtime-observability.bicep must exist")

	assert.Contains(t, string(content), "otlp",
		"bicep file must reference OpenTelemetry exporter (otlp)")
}

// TestRuntimeObservabilityBicep_ContainsLivenessProbe verifies that the runtime
// observability module defines a liveness probe so the provision command can apply
// health checks to Drasi runtime workloads.
func TestRuntimeObservabilityBicep_ContainsLivenessProbe(t *testing.T) {
	t.Parallel()

	bicepPath := filepath.Join("..", "..", "..", "infra", "modules", "drasi-runtime-observability.bicep")
	content, err := os.ReadFile(bicepPath)
	require.NoError(t, err, "drasi-runtime-observability.bicep must exist")

	assert.Contains(t, string(content), "livenessProbe",
		"bicep file must reference livenessProbe definition")
}

// TestRuntimeObservabilityBicep_ContainsReadinessProbe verifies that the runtime
// observability module defines a readiness probe for Drasi runtime workloads.
func TestRuntimeObservabilityBicep_ContainsReadinessProbe(t *testing.T) {
	t.Parallel()

	bicepPath := filepath.Join("..", "..", "..", "infra", "modules", "drasi-runtime-observability.bicep")
	content, err := os.ReadFile(bicepPath)
	require.NoError(t, err, "drasi-runtime-observability.bicep must exist")

	assert.Contains(t, string(content), "readinessProbe",
		"bicep file must reference readinessProbe definition")
}

// TestRuntimeObservabilityBicep_ContainsStartupProbe verifies that the runtime
// observability module defines a startup probe for Drasi runtime workloads, preventing
// premature liveness/readiness checks during slow container initialization.
func TestRuntimeObservabilityBicep_ContainsStartupProbe(t *testing.T) {
	t.Parallel()

	bicepPath := filepath.Join("..", "..", "..", "infra", "modules", "drasi-runtime-observability.bicep")
	content, err := os.ReadFile(bicepPath)
	require.NoError(t, err, "drasi-runtime-observability.bicep must exist")

	assert.Contains(t, string(content), "startupProbe",
		"bicep file must reference startupProbe definition")
}
