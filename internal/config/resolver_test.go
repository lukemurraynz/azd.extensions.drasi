package config_test

import (
	"path/filepath"
	"testing"

	"github.com/azure/azd.extensions.drasi/internal/config"
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
