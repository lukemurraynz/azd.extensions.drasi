package cmd_test

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCommand_ValidFixture_ExitsZero(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	configPath := filepath.Join("..", "internal", "config", "testdata", "valid-manifest", "drasi.yaml")
	root.SetArgs([]string{"validate", "--config", configPath})
	err := root.Execute()
	assert.NoError(t, err)
}

func TestValidateCommand_MissingConfig_ExitsTwo(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"validate", "--config", "does-not-exist/drasi.yaml"})
	err := root.Execute()
	require.Error(t, err)
}

func TestValidateCommand_ErrorFixture_ExitsOne(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	configPath := filepath.Join("..", "internal", "config", "testdata", "missing-reference", "drasi.yaml")
	root.SetArgs([]string{"validate", "--config", configPath})
	err := root.Execute()
	require.Error(t, err)
}

func TestValidateCommand_JSONOutput_ValidSchema(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	configPath := filepath.Join("..", "internal", "config", "testdata", "missing-reference", "drasi.yaml")
	root.SetArgs([]string{"validate", "--config", configPath, "--output", "json"})
	_ = root.Execute()
	var out map[string]any
	err := json.Unmarshal(buf.Bytes(), &out)
	require.NoError(t, err, "output should be valid JSON")
	assert.Contains(t, out, "status")
	assert.Contains(t, out, "issues")
}

func TestValidateCommand_StrictMode_WarningsAreErrors(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	configPath := filepath.Join("..", "internal", "config", "testdata", "overlay-warning", "drasi.yaml")
	root.SetArgs([]string{"validate", "--config", configPath, "--environment", "dev", "--strict"})
	err := root.Execute()
	require.Error(t, err, "strict mode should promote warnings to errors")
}

func TestValidateCommand_EnvironmentFlag_Accepted(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	configPath := filepath.Join("..", "internal", "config", "testdata", "valid-manifest", "drasi.yaml")
	root.SetArgs([]string{"validate", "--config", configPath, "--environment", "dev"})
	err := root.Execute()
	assert.NoError(t, err)
}
