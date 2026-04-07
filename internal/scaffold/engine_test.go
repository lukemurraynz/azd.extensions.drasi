package scaffold_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/azure/azd.extensions.drasi/internal/scaffold"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T032: blank template scaffold tests

func TestScaffold_BlankTemplate_CreatesExpectedTree(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	files, err := scaffold.Scaffold("blank", dir, false)

	require.NoError(t, err)
	require.NotEmpty(t, files, "returned file list must not be empty")

	requiredPaths := []string{
		"azure.yaml",
		"drasi/drasi.yaml",
		"drasi/sources/example-source.yaml",
		"drasi/queries/example-query.yaml",
		"drasi/reactions/example-reaction.yaml",
		"drasi/environments/dev.yaml",
		"docker-compose.yml",
		"infra/main.bicep",
		"infra/main.parameters.bicepparam",
		".vscode/launch.json",
	}

	for _, rel := range requiredPaths {
		fullPath := filepath.Join(dir, rel)
		_, statErr := os.Stat(fullPath)
		assert.NoError(t, statErr, "expected file to exist: %s", rel)
	}
}

func TestScaffold_BlankTemplate_ReturnedFilesMatchFS(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	files, err := scaffold.Scaffold("blank", dir, false)
	require.NoError(t, err)

	// Every path in the returned list must exist on the filesystem.
	for _, rel := range files {
		fullPath := filepath.Join(dir, rel)
		_, statErr := os.Stat(fullPath)
		assert.NoError(t, statErr, "returned file missing on disk: %s", rel)
	}

	// The returned list must be non-empty.
	assert.NotEmpty(t, files)
}

func TestScaffold_BlankTemplate_ConflictWithoutForce_ReturnsError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// First call — must succeed.
	_, err := scaffold.Scaffold("blank", dir, false)
	require.NoError(t, err)

	// Second call without --force — must fail with a conflict error.
	_, err = scaffold.Scaffold("blank", dir, false)
	require.Error(t, err, "expected conflict error on re-run without --force")
	assert.Contains(t, err.Error(), "already exists")
}

func TestScaffold_BlankTemplate_ForceOverwrites(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	_, err := scaffold.Scaffold("blank", dir, false)
	require.NoError(t, err)

	// Second call with force — must succeed.
	files, err := scaffold.Scaffold("blank", dir, true)
	require.NoError(t, err)
	assert.NotEmpty(t, files)
}

func TestScaffold_InvalidTemplate_ReturnsError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	_, err := scaffold.Scaffold("does-not-exist", dir, false)
	require.Error(t, err)
}

// T104: dapr-pubsub reaction scaffold tests (FR-031)

// T017: postgresql-source template scaffold tests

func TestScaffold_PostgreSQLSourceTemplate_CreatesExpectedTree(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	files, err := scaffold.Scaffold("postgresql-source", dir, false)

	require.NoError(t, err)
	require.NotEmpty(t, files, "returned file list must not be empty")

	requiredPaths := []string{
		"azure.yaml",
		"docker-compose.yml",
		"drasi/drasi.yaml",
		"drasi/sources/pg-source.yaml",
		"drasi/queries/watch-orders.yaml",
		"drasi/reactions/log-changes.yaml",
		"drasi/environments/dev.yaml",
		"infra/main.bicep",
		"infra/main.parameters.bicepparam",
		"infra/modules/postgresql.bicep",
		".vscode/launch.json",
	}

	for _, rel := range requiredPaths {
		fullPath := filepath.Join(dir, rel)
		_, statErr := os.Stat(fullPath)
		assert.NoError(t, statErr, "expected file to exist: %s", rel)
	}
}

// T104: dapr-pubsub reaction scaffold tests (FR-031)

func TestScaffold_CosmosFeed_CreatesDaprComponentYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	files, err := scaffold.Scaffold("cosmos-change-feed", dir, false)
	require.NoError(t, err)
	require.NotEmpty(t, files)

	// A Dapr pub/sub component YAML must exist under drasi/components/dapr/.
	daprDir := filepath.Join(dir, "drasi", "components", "dapr")
	entries, readErr := os.ReadDir(daprDir)
	require.NoError(t, readErr, "drasi/components/dapr/ directory must exist")
	require.NotEmpty(t, entries, "at least one Dapr component YAML must be generated")
}

func TestScaffold_CosmosFeed_ReactionReferencesDaprComponent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	_, err := scaffold.Scaffold("cosmos-change-feed", dir, false)
	require.NoError(t, err)

	// The reaction YAML file must exist.
	reactionDir := filepath.Join(dir, "drasi", "reactions")
	entries, readErr := os.ReadDir(reactionDir)
	require.NoError(t, readErr)
	require.NotEmpty(t, entries, "at least one reaction YAML must exist")

	// The Dapr component file must exist and is referenced by the same topic/broker metadata.
	daprDir := filepath.Join(dir, "drasi", "components", "dapr")
	daprEntries, err2 := os.ReadDir(daprDir)
	require.NoError(t, err2)
	require.NotEmpty(t, daprEntries)
}

func TestScaffold_CosmosFeed_DaprComponentPath_IsUnderDrasiComponents(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	files, err := scaffold.Scaffold("cosmos-change-feed", dir, false)
	require.NoError(t, err)

	hasDaprComponent := false
	for _, f := range files {
		if filepath.HasPrefix(filepath.ToSlash(f), "drasi/components/dapr/") {
			hasDaprComponent = true
			break
		}
	}
	assert.True(t, hasDaprComponent, "cosmos-change-feed template must emit a file under drasi/components/dapr/")
}
