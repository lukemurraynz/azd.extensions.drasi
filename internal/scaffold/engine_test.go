package scaffold_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/scaffold"
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

	// Second call without --force — must fail with a conflict error on a
	// non-azure.yaml file (azure.yaml is silently skipped).
	_, err = scaffold.Scaffold("blank", dir, false)
	require.Error(t, err, "expected conflict error on re-run without --force")
	assert.Contains(t, err.Error(), "already exists")
	assert.NotContains(t, err.Error(), "azure.yaml",
		"azure.yaml must not trigger the conflict check")
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

func TestScaffold_PreExistingAzureYAML_SkippedWithoutForce(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Simulate `azd init` creating an azure.yaml before the extension runs.
	azureYAML := filepath.Join(dir, "azure.yaml")
	originalContent := "name: from-azd-init\n"
	require.NoError(t, os.WriteFile(azureYAML, []byte(originalContent), 0600))

	// `azd drasi init` (without --force) must succeed despite the existing azure.yaml.
	files, err := scaffold.Scaffold("blank", dir, false)
	require.NoError(t, err)
	assert.NotEmpty(t, files)

	// azure.yaml must NOT appear in the returned file list (it was skipped).
	for _, f := range files {
		assert.NotEqual(t, "azure.yaml", f,
			"azure.yaml must be skipped, not included in the output list")
	}

	// The original azure.yaml content from `azd init` must be preserved.
	content, readErr := os.ReadFile(azureYAML)
	require.NoError(t, readErr)
	assert.Equal(t, originalContent, string(content),
		"azure.yaml must be left untouched when it already exists")
}

func TestScaffold_InvalidTemplate_ReturnsError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	_, err := scaffold.Scaffold("does-not-exist", dir, false)
	require.Error(t, err)
}

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

// T104: cosmos-change-feed reaction scaffold tests

func TestScaffold_CosmosFeed_CreatesDebugReaction(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	files, err := scaffold.Scaffold("cosmos-change-feed", dir, false)
	require.NoError(t, err)
	require.NotEmpty(t, files)

	// A Debug reaction YAML must exist under drasi/reactions/.
	reactionPath := filepath.Join(dir, "drasi", "reactions", "log-changes.yaml")
	_, statErr := os.Stat(reactionPath)
	assert.NoError(t, statErr, "drasi/reactions/log-changes.yaml must exist")
}

func TestScaffold_CosmosFeed_ReactionIsDebugKind(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	_, err := scaffold.Scaffold("cosmos-change-feed", dir, false)
	require.NoError(t, err)

	reactionPath := filepath.Join(dir, "drasi", "reactions", "log-changes.yaml")
	content, readErr := os.ReadFile(reactionPath)
	require.NoError(t, readErr)
	assert.Contains(t, string(content), "kind: Debug", "reaction must be Debug kind")
}

func TestScaffold_CosmosFeed_HasCosmosGremlinBicepModule(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	files, err := scaffold.Scaffold("cosmos-change-feed", dir, false)
	require.NoError(t, err)

	hasCosmosModule := false
	for _, f := range files {
		if strings.HasSuffix(filepath.ToSlash(f), "infra/modules/cosmos-gremlin.bicep") {
			hasCosmosModule = true
			break
		}
	}
	assert.True(t, hasCosmosModule, "cosmos-change-feed template must emit infra/modules/cosmos-gremlin.bicep")
}

func TestScaffold_AllTemplates_ProduceFiles(t *testing.T) {
	t.Parallel()

	templates := []string{
		"blank",
		"blank-terraform",
		"cosmos-change-feed",
		"event-hub-routing",
		"postgresql-source",
		"query-subscription",
	}

	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			files, err := scaffold.Scaffold(tmpl, dir, false)
			require.NoError(t, err, "template %q must scaffold without error", tmpl)
			require.NotEmpty(t, files, "template %q must produce at least one file", tmpl)

			for _, rel := range files {
				fullPath := filepath.Join(dir, rel)
				info, statErr := os.Stat(fullPath)
				require.NoError(t, statErr, "file must exist on disk: %s", rel)
				assert.True(t, info.Size() > 0, "file must be non-empty: %s", rel)
			}
		})
	}
}

func TestScaffold_AllTemplates_ForceOverwrite(t *testing.T) {
	t.Parallel()

	templates := []string{
		"blank-terraform",
		"event-hub-routing",
		"query-subscription",
	}

	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()

			// First scaffold
			_, err := scaffold.Scaffold(tmpl, dir, false)
			require.NoError(t, err)

			// Second without force should fail
			_, err = scaffold.Scaffold(tmpl, dir, false)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "already exists")

			// Third with force should succeed
			files, err := scaffold.Scaffold(tmpl, dir, true)
			require.NoError(t, err)
			assert.NotEmpty(t, files)
		})
	}
}
