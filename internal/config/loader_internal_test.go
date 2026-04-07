package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponentHash_StateKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		hash ComponentHash
		want string
	}{
		{
			name: "uppercases kind and replaces hyphens",
			hash: ComponentHash{Kind: "source", ID: "my-source"},
			want: "DRASI_HASH_SOURCE_my_source",
		},
		{
			name: "preserves underscores in id",
			hash: ComponentHash{Kind: "continuousquery", ID: "already_snake"},
			want: "DRASI_HASH_CONTINUOUSQUERY_already_snake",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, tc.hash.StateKey())
		})
	}
}

func TestLoadMiddleware_LoadsDecodedMiddleware(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "middleware.yaml")
	require.NoError(t, os.WriteFile(path, []byte(`apiVersion: v1
kind: Middleware
name: enrich-orders
spec:
  kind: Map
  config:
    endpoint:
      value: https://example.test
`), 0o600))

	middleware, err := loadMiddleware(path, filepath.Join("middlewares", "middleware.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "Middleware", middleware.Kind)
	assert.Equal(t, "enrich-orders", middleware.ID)
	assert.Equal(t, "Map", middleware.Spec.Kind)
	require.Contains(t, middleware.Spec.Config, "endpoint")
	assert.Equal(t, "https://example.test", middleware.Spec.Config["endpoint"].StringValue)
	assert.Equal(t, "middlewares/middleware.yaml", middleware.FilePath)
	assert.Greater(t, middleware.Line, 0)
}

func TestDecodeNode_ValidYAML_ReturnsRootNode(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "source.yaml")
	require.NoError(t, os.WriteFile(path, []byte("kind: Source\nname: sample\n"), 0o600))

	node, err := decodeNode(path)
	require.NoError(t, err)
	assert.Equal(t, 1, node.Line)
	assert.NotEmpty(t, node.Content)
	assert.Equal(t, "kind", node.Content[0].Value)
	assert.Equal(t, "Source", node.Content[1].Value)
}

func TestDecodeNode_EmptyFile_ReturnsDocumentNode(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "empty.yaml")
	require.NoError(t, os.WriteFile(path, []byte(""), 0o600))

	node, err := decodeNode(path)
	require.NoError(t, err)
	assert.Empty(t, node.Content)
	assert.Zero(t, node.Kind)
}

func TestDecodeNode_InvalidYAML_ReturnsError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.yaml")
	require.NoError(t, os.WriteFile(path, []byte("kind: [broken\n"), 0o600))

	_, err := decodeNode(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing")
	assert.Contains(t, err.Error(), path)
}

func TestExpandIncludePattern_GlobAndDoubleStar(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createConfigFile(t, dir, filepath.Join("queries", "root.yaml"), "kind: ContinuousQuery\n")
	createConfigFile(t, dir, filepath.Join("queries", "nested", "child.yaml"), "kind: ContinuousQuery\n")
	createConfigFile(t, dir, filepath.Join("queries", "nested", "ignore.txt"), "ignored\n")

	globMatches, err := expandIncludePattern(dir, "queries/*.yaml")
	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(dir, "queries", "root.yaml")}, globMatches)

	doubleStarMatches, err := expandIncludePattern(dir, "queries/**/*.yaml")
	require.NoError(t, err)
	assert.Equal(t, []string{
		filepath.Join(dir, "queries", "nested", "child.yaml"),
		filepath.Join(dir, "queries", "root.yaml"),
	}, doubleStarMatches)
}

func TestSplitPattern_RemovesEmptySegments(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{"queries", "nested", "child.yaml"}, splitPattern("/queries//nested/child.yaml/"))
}

func TestMatchDoubleStarPattern_MultipleDoubleStars(t *testing.T) {
	t.Parallel()

	assert.True(t, matchDoubleStarPattern("queries/**/nested/**/child.yaml", "queries/a/nested/b/c/child.yaml"))
	assert.False(t, matchDoubleStarPattern("queries/**/nested/**/child.yaml", "queries/a/other/b/c/child.yaml"))
}

func TestMatchPatternParts_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		patternParts []string
		valueParts   []string
		want         bool
	}{
		{
			name:         "empty pattern matches empty value",
			patternParts: nil,
			valueParts:   nil,
			want:         true,
		},
		{
			name:         "empty pattern does not match non empty value",
			patternParts: nil,
			valueParts:   []string{"queries"},
			want:         false,
		},
		{
			name:         "pattern longer than value fails",
			patternParts: []string{"queries", "child.yaml"},
			valueParts:   []string{"queries"},
			want:         false,
		},
		{
			name:         "double star matches zero segments",
			patternParts: []string{"queries", "**", "child.yaml"},
			valueParts:   []string{"queries", "child.yaml"},
			want:         true,
		},
		{
			name:         "double star matches multiple segments",
			patternParts: []string{"queries", "**", "child.yaml"},
			valueParts:   []string{"queries", "nested", "deep", "child.yaml"},
			want:         true,
		},
		{
			name:         "multiple double stars can backtrack",
			patternParts: []string{"**", "nested", "**", "child.yaml"},
			valueParts:   []string{"queries", "a", "nested", "b", "child.yaml"},
			want:         true,
		},
		{
			name:         "wildcard segment respects filepath match",
			patternParts: []string{"queries", "*.yaml"},
			valueParts:   []string{"queries", "child.yaml"},
			want:         true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, matchPatternParts(tc.patternParts, tc.valueParts))
		})
	}
}

func createConfigFile(t *testing.T, dir, relPath, content string) {
	t.Helper()

	path := filepath.Join(dir, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
}
