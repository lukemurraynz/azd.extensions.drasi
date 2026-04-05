package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/azure/azd.extensions.drasi/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_ValidManifest(t *testing.T) {
	t.Parallel()
	dir := filepath.Join("testdata", "valid-manifest")
	manifest, sources, queries, reactions, middlewares, err := config.LoadManifest(dir, "drasi.yaml")
	require.NoError(t, err)
	assert.Equal(t, "v1", manifest.APIVersion)
	assert.Len(t, sources, 1)
	assert.Equal(t, "postgres-source", sources[0].ID)
	assert.Len(t, queries, 1)
	assert.Equal(t, "order-changes", queries[0].ID)
	assert.Len(t, reactions, 1)
	assert.Equal(t, "pubsub-orders", reactions[0].ID)
	assert.Empty(t, middlewares)
}

func TestLoader_MissingManifest(t *testing.T) {
	t.Parallel()
	_, _, _, _, _, err := config.LoadManifest("testdata/does-not-exist", "drasi.yaml")
	require.Error(t, err)
}

func TestLoader_MalformedYAML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	badManifest := filepath.Join(dir, "drasi.yaml")
	err := writeFile(badManifest, "apiVersion: [not valid yaml: {")
	require.NoError(t, err)
	_, _, _, _, _, err = config.LoadManifest(dir, "drasi.yaml")
	require.Error(t, err)
}

func TestLoader_FilepathAndLinePopulated(t *testing.T) {
	t.Parallel()
	dir := filepath.Join("testdata", "valid-manifest")
	_, sources, _, _, _, err := config.LoadManifest(dir, "drasi.yaml")
	require.NoError(t, err)
	require.Len(t, sources, 1)
	assert.NotEmpty(t, sources[0].FilePath)
	assert.Greater(t, sources[0].Line, 0)
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
