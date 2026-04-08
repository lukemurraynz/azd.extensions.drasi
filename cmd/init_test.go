package cmd_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitCommand_BlankTemplate_ExitsZero(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"init", "--template", "blank", "--output-dir", dir})
	err := root.Execute()
	assert.NoError(t, err)
}

func TestInitCommand_TemplateFlag_AcceptsAllTemplates(t *testing.T) {
	t.Parallel()

	templates := []string{
		"blank",
		"event-hub-routing",
		"query-subscription",
	}

	for _, tmpl := range templates {
		tmpl := tmpl
		t.Run(tmpl, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			root := cmd.NewRootCommand()
			root.SetOut(&bytes.Buffer{})
			root.SetErr(&bytes.Buffer{})
			root.SetArgs([]string{"init", "--template", tmpl, "--output-dir", dir})
			err := root.Execute()
			assert.NoError(t, err, "template %s should be accepted", tmpl)
		})
	}
}

func TestInitCommand_DefaultTemplate_IsBlank(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	// No --template flag — should default to "blank".
	root.SetArgs([]string{"init", "--output-dir", dir})
	err := root.Execute()
	assert.NoError(t, err)

	// azure.yaml must exist as proof that the blank template was used.
	_, statErr := os.Stat(filepath.Join(dir, "azure.yaml"))
	assert.NoError(t, statErr, "azure.yaml must exist after blank scaffold")
}

func TestInitCommand_ForceFlag_Accepted(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// First run.
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"init", "--template", "blank", "--output-dir", dir})
	require.NoError(t, root.Execute())

	// Second run with --force — must succeed.
	root2 := cmd.NewRootCommand()
	root2.SetOut(&bytes.Buffer{})
	root2.SetErr(&bytes.Buffer{})
	root2.SetArgs([]string{"init", "--template", "blank", "--force", "--output-dir", dir})
	err := root2.Execute()
	assert.NoError(t, err)
}

func TestInitCommand_ConflictWithoutForce_ExitsOne(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"init", "--template", "blank", "--output-dir", dir})
	require.NoError(t, root.Execute())

	root2 := cmd.NewRootCommand()
	root2.SetOut(&bytes.Buffer{})
	root2.SetErr(&bytes.Buffer{})
	root2.SetArgs([]string{"init", "--template", "blank", "--output-dir", dir})
	err := root2.Execute()
	require.Error(t, err, "re-run without --force must return an error")
}

func TestInitCommand_JSONOutput_EmitsFileList(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := cmd.NewRootCommand()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"init", "--template", "blank", "--output-dir", dir, "--output", "json"})
	err := root.Execute()
	require.NoError(t, err)

	var out map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out), "stdout must be valid JSON")
	assert.Equal(t, "ok", out["status"], "status must be 'ok'")
	files, ok := out["files"]
	assert.True(t, ok, "JSON output must contain 'files' key")
	assert.NotEmpty(t, files, "files list must not be empty")
}

func TestInitCommand_EnvironmentFlag_Accepted(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"init", "--template", "blank", "--output-dir", dir, "--environment", "dev"})
	err := root.Execute()
	assert.NoError(t, err)
}

func TestInitCommand_InvalidTemplate_ExitsOne(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"init", "--template", "nonexistent-template", "--output-dir", dir})
	err := root.Execute()
	require.Error(t, err)
}
