package cmd_test

import (
	"bytes"
	"testing"

	"github.com/azure/azd.extensions.drasi/cmd"
	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStatusCommand_NotImplemented verifies the current stub returns ERR_NOT_IMPLEMENTED.
func TestStatusCommand_NotImplemented(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"status"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), output.ERR_NOT_IMPLEMENTED)
}

// TestStatusCommand_Help verifies the command registers and shows usage.
func TestStatusCommand_Help(t *testing.T) {
	t.Parallel()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	root := cmd.NewRootCommand()
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"status", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Usage:")
	assert.Empty(t, stderr.String())
}

// TestStatusCommand_EnvironmentFlagAccepted verifies --environment is a registered flag.
func TestStatusCommand_EnvironmentFlagAccepted(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"status", "--environment", "dev"})

	err := root.Execute()

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unknown flag",
		"--environment must be accepted on status command")
	assert.Contains(t, err.Error(), output.ERR_NOT_IMPLEMENTED)
}

// TestStatusCommand_OutputJSONFlagAccepted verifies --output json does not cause a flag parse error.
func TestStatusCommand_OutputJSONFlagAccepted(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--output", "json", "status"})

	err := root.Execute()

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unknown flag")
	assert.Contains(t, err.Error(), output.ERR_NOT_IMPLEMENTED)
}
