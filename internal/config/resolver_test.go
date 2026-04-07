package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_NoEnvironment(t *testing.T) {
	t.Parallel()
	dir := filepath.Join("testdata", "valid-manifest")
	manifest, sources, queries, reactions, middlewares, err := config.LoadManifest(dir, "drasi.yaml")
	require.NoError(t, err)
	resolved, warnings, err := config.ResolveManifest(manifest, sources, queries, reactions, middlewares, dir, "")
	require.NoError(t, err)
	assert.Empty(t, warnings)
	assert.Len(t, resolved.Sources, 1)
	assert.Len(t, resolved.Queries, 1)
	assert.Len(t, resolved.Reactions, 1)
}

func TestResolver_EnvironmentOverlay(t *testing.T) {
	t.Parallel()
	dir := filepath.Join("testdata", "valid-manifest")
	manifest, sources, queries, reactions, middlewares, err := config.LoadManifest(dir, "drasi.yaml")
	require.NoError(t, err)
	resolved, warnings, err := config.ResolveManifest(manifest, sources, queries, reactions, middlewares, dir, "dev")
	require.NoError(t, err)
	assert.Empty(t, warnings)
	assert.Equal(t, "dev", resolved.Environment.Name)
	assert.Equal(t, "eastus", resolved.Environment.Parameters["region"])
}

func TestResolver_UndeclaredOverlayParameter_EmitsWarning(t *testing.T) {
	t.Parallel()
	dir := filepath.Join("testdata", "overlay-warning")
	manifest, sources, queries, reactions, middlewares, err := config.LoadManifest(dir, "drasi.yaml")
	require.NoError(t, err)
	_, warnings, err := config.ResolveManifest(manifest, sources, queries, reactions, middlewares, dir, "dev")
	require.NoError(t, err)
	assert.NotEmpty(t, warnings, "expected warnings for undeclared overlay parameter")
}

func TestResolver_DeterministicSort(t *testing.T) {
	t.Parallel()
	dir := filepath.Join("testdata", "valid-manifest")
	manifest, sources, queries, reactions, middlewares, err := config.LoadManifest(dir, "drasi.yaml")
	require.NoError(t, err)
	r1, _, err := config.ResolveManifest(manifest, sources, queries, reactions, middlewares, dir, "")
	require.NoError(t, err)
	r2, _, err := config.ResolveManifest(manifest, sources, queries, reactions, middlewares, dir, "")
	require.NoError(t, err)
	assert.Equal(t, r1.Sources, r2.Sources)
	assert.Equal(t, r1.Queries, r2.Queries)
}

func TestResolver_ComponentExclusionRemovesSpecifiedComponent(t *testing.T) {
	t.Parallel()

	manifest := config.DrasiManifest{Environments: map[string]string{"dev": "environments/dev.yaml"}}
	sources := []config.Source{{ID: "postgres-source"}, {ID: "keep-source"}}
	queries := []config.ContinuousQuery{{ID: "order-changes"}}
	reactions := []config.Reaction{{ID: "pubsub-orders"}}
	middlewares := []config.Middleware{{ID: "transformer"}}
	dir := t.TempDir()
	writeEnvironmentOverlay(t, dir, `
name: dev
components:
  exclude:
    - kind: source
      id: postgres-source
`)

	resolved, warnings, err := config.ResolveManifest(manifest, sources, queries, reactions, middlewares, dir, "dev")
	require.NoError(t, err)
	assert.Empty(t, warnings)
	assert.Len(t, resolved.Sources, 1)
	assert.Equal(t, "keep-source", resolved.Sources[0].ID)
	assert.Len(t, resolved.Queries, 1)
	assert.Len(t, resolved.Reactions, 1)
	assert.Len(t, resolved.Middlewares, 1)
}

func TestResolver_NonMatchingExclusionLeavesComponentsIntact(t *testing.T) {
	t.Parallel()

	manifest := config.DrasiManifest{Environments: map[string]string{"dev": "environments/dev.yaml"}}
	sources := []config.Source{{ID: "postgres-source"}}
	queries := []config.ContinuousQuery{{ID: "order-changes"}}
	reactions := []config.Reaction{{ID: "pubsub-orders"}}
	middlewares := []config.Middleware{{ID: "transformer"}}
	dir := t.TempDir()
	writeEnvironmentOverlay(t, dir, `
name: dev
components:
  exclude:
    - kind: source
      id: missing-source
`)

	resolved, warnings, err := config.ResolveManifest(manifest, sources, queries, reactions, middlewares, dir, "dev")
	require.NoError(t, err)
	assert.Len(t, warnings, 1)
	assert.Equal(t, config.WarningMissingComponentExclusion, warnings[0].Code)
	assert.Len(t, resolved.Sources, 1)
	assert.Len(t, resolved.Queries, 1)
	assert.Len(t, resolved.Reactions, 1)
	assert.Len(t, resolved.Middlewares, 1)
}

func TestResolver_ComponentExclusionKindMatchingIsCaseInsensitive(t *testing.T) {
	t.Parallel()

	manifest := config.DrasiManifest{Environments: map[string]string{"dev": "environments/dev.yaml"}}
	sources := []config.Source{{ID: "postgres-source"}}
	queries := []config.ContinuousQuery{{ID: "order-changes"}}
	reactions := []config.Reaction{{ID: "pubsub-orders"}}
	middlewares := []config.Middleware{{ID: "transformer"}}
	dir := t.TempDir()
	writeEnvironmentOverlay(t, dir, `
name: dev
components:
  exclude:
    - kind: ReAcTiOn
      id: pubsub-orders
`)

	resolved, warnings, err := config.ResolveManifest(manifest, sources, queries, reactions, middlewares, dir, "dev")
	require.NoError(t, err)
	assert.Empty(t, warnings)
	assert.Len(t, resolved.Reactions, 0)
	assert.Len(t, resolved.Sources, 1)
	assert.Len(t, resolved.Queries, 1)
	assert.Len(t, resolved.Middlewares, 1)
}

func TestResolver_ComponentExclusionWorksWithParameterOverrides(t *testing.T) {
	t.Parallel()

	manifest := config.DrasiManifest{Environments: map[string]string{"dev": "environments/dev.yaml"}}
	sources := []config.Source{{ID: "postgres-source"}}
	queries := []config.ContinuousQuery{{ID: "order-changes"}}
	reactions := []config.Reaction{{ID: "pubsub-orders"}}
	middlewares := []config.Middleware{{ID: "transformer"}}
	dir := t.TempDir()
	writeEnvironmentOverlay(t, dir, `
name: dev
parameters:
  region: eastus
components:
  exclude:
    - kind: continuousquery
      id: order-changes
`)

	resolved, warnings, err := config.ResolveManifest(manifest, sources, queries, reactions, middlewares, dir, "dev")
	require.NoError(t, err)
	assert.Empty(t, warnings)
	assert.Equal(t, "eastus", resolved.Environment.Parameters["region"])
	assert.Len(t, resolved.Queries, 0)
	assert.Len(t, resolved.Sources, 1)
	assert.Len(t, resolved.Reactions, 1)
	assert.Len(t, resolved.Middlewares, 1)
}

func writeEnvironmentOverlay(t *testing.T, dir, content string) {
	t.Helper()

	envDir := filepath.Join(dir, "environments")
	require.NoError(t, os.MkdirAll(envDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(envDir, "dev.yaml"), []byte(content), 0o600))
}
