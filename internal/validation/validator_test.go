package validation_test

import (
	"path/filepath"
	"testing"

	"github.com/azure/azd.extensions.drasi/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_ValidManifest(t *testing.T) {
	t.Parallel()
	dir := filepath.Join("..", "config", "testdata", "valid-manifest")
	result, err := validation.Validate(dir, "drasi.yaml", "")
	require.NoError(t, err)
	assert.False(t, result.HasErrors())
	assert.False(t, result.HasWarnings())
}

func TestValidator_MissingQueryLanguage(t *testing.T) {
	t.Parallel()
	dir := filepath.Join("..", "config", "testdata", "missing-query-language")
	result, err := validation.Validate(dir, "drasi.yaml", "")
	require.NoError(t, err)
	assert.True(t, result.HasErrors())
	found := false
	for _, issue := range result.Issues {
		if issue.Code == "ERR_MISSING_QUERY_LANGUAGE" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestValidator_MissingReference(t *testing.T) {
	t.Parallel()
	dir := filepath.Join("..", "config", "testdata", "missing-reference")
	result, err := validation.Validate(dir, "drasi.yaml", "")
	require.NoError(t, err)
	assert.True(t, result.HasErrors())
	found := false
	for _, issue := range result.Issues {
		if issue.Code == "ERR_MISSING_REFERENCE" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestValidator_CircularDependency(t *testing.T) {
	t.Parallel()
	dir := filepath.Join("..", "config", "testdata", "circular-dep")
	result, err := validation.Validate(dir, "drasi.yaml", "")
	require.NoError(t, err)
	assert.True(t, result.HasErrors())
	found := false
	for _, issue := range result.Issues {
		if issue.Code == "ERR_CIRCULAR_DEPENDENCY" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestValidator_OverlayWarning(t *testing.T) {
	t.Parallel()
	dir := filepath.Join("..", "config", "testdata", "overlay-warning")
	result, err := validation.Validate(dir, "drasi.yaml", "dev")
	require.NoError(t, err)
	assert.False(t, result.HasErrors(), "overlay warning must not be an error")
	assert.True(t, result.HasWarnings(), "expected warning for undeclared overlay parameter")
}
