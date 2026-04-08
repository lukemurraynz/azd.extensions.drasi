package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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

// TestValue_UnmarshalYAML verifies that config.Value accepts both the plain scalar
// format ("true") and the legacy {value: "..."} struct format used in older YAML files.
// This is critical: drasi apply requires plain scalars on the wire, but the Go loader
// must still parse files that were generated with the {value: ...} wrapper.
func TestValue_UnmarshalYAML_PlainScalar(t *testing.T) {
	t.Parallel()
	input := `inCluster: "true"`
	var m map[string]config.Value
	require.NoError(t, yaml.Unmarshal([]byte(input), &m))
	assert.Equal(t, "true", m["inCluster"].StringValue)
	assert.Nil(t, m["inCluster"].SecretRef)
}

func TestValue_UnmarshalYAML_StructForm(t *testing.T) {
	t.Parallel()
	input := "inCluster:\n  value: \"true\"\n"
	var m map[string]config.Value
	require.NoError(t, yaml.Unmarshal([]byte(input), &m))
	assert.Equal(t, "true", m["inCluster"].StringValue)
}

func TestValue_UnmarshalYAML_SecretRef(t *testing.T) {
	t.Parallel()
	input := "endpoint:\n  secretRef:\n    vaultName: my-vault\n    secretName: my-secret\n"
	var m map[string]config.Value
	require.NoError(t, yaml.Unmarshal([]byte(input), &m))
	require.NotNil(t, m["endpoint"].SecretRef)
	assert.Equal(t, "my-vault", m["endpoint"].SecretRef.VaultName)
	assert.Equal(t, "my-secret", m["endpoint"].SecretRef.SecretName)
}

func TestValue_UnmarshalYAML_Sequence(t *testing.T) {
	t.Parallel()
	input := "tables:\n  - public.orders\n  - public.items\n"
	var m map[string]config.Value
	require.NoError(t, yaml.Unmarshal([]byte(input), &m))
	// Sequence values are opaque passthrough; StringValue stays empty.
	assert.Empty(t, m["tables"].StringValue)
	assert.Nil(t, m["tables"].SecretRef)
}

func TestValue_UnmarshalYAML_DrasiAPIMappingPassthrough(t *testing.T) {
	t.Parallel()
	input := "connectionString:\n  kind: Secret\n  name: pg-source-secrets\n  key: connectionString\n"
	var m map[string]config.Value
	require.NoError(t, yaml.Unmarshal([]byte(input), &m))
	// Drasi API discriminated union mappings are opaque; no extension fields populated.
	assert.Empty(t, m["connectionString"].StringValue)
	assert.Nil(t, m["connectionString"].SecretRef)
	assert.Nil(t, m["connectionString"].EnvRef)
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
