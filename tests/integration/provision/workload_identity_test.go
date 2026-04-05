//go:build integration

package provision_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWorkloadIdentityBicep_ContainsServiceAccountAnnotation verifies FR-045 step 3:
// the workload identity module defines the ServiceAccount annotation key required by
// the Azure Workload Identity webhook.
func TestWorkloadIdentityBicep_ContainsServiceAccountAnnotation(t *testing.T) {
	t.Parallel()

	bicepPath := filepath.Join("..", "..", "..", "infra", "modules", "drasi-workload-identity.bicep")
	content, err := os.ReadFile(bicepPath)
	require.NoError(t, err, "drasi-workload-identity.bicep must exist")

	assert.Contains(t, string(content), "azure.workload.identity/client-id",
		"bicep file must reference serviceAccount annotation key")
}

// TestWorkloadIdentityBicep_ContainsPodLabel verifies FR-045 step 4:
// the workload identity module defines the pod label required by the Azure Workload
// Identity webhook to inject the token volume into Drasi runtime pods.
func TestWorkloadIdentityBicep_ContainsPodLabel(t *testing.T) {
	t.Parallel()

	bicepPath := filepath.Join("..", "..", "..", "infra", "modules", "drasi-workload-identity.bicep")
	content, err := os.ReadFile(bicepPath)
	require.NoError(t, err, "drasi-workload-identity.bicep must exist")

	assert.Contains(t, string(content), `azure.workload.identity/use: "true"`,
		"bicep file must reference pod label for workload identity injection")
}

// TestWorkloadIdentityBicep_ContainsNamespaceOutput verifies that the workload identity
// module exposes a namespace output so the provision command can target the correct
// Kubernetes namespace when applying ServiceAccount patches.
func TestWorkloadIdentityBicep_ContainsNamespaceOutput(t *testing.T) {
	t.Parallel()

	bicepPath := filepath.Join("..", "..", "..", "infra", "modules", "drasi-workload-identity.bicep")
	content, err := os.ReadFile(bicepPath)
	require.NoError(t, err, "drasi-workload-identity.bicep must exist")

	assert.Contains(t, string(content), "output namespace",
		"bicep file must expose a namespace output")
}
