package cmd_test

import (
	"bytes"
	"testing"

	"github.com/azure/azd.extensions.drasi/cmd"
	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLogsCommand_NotImplemented verifies the current stub returns ERR_NOT_IMPLEMENTED.
func TestLogsCommand_NotImplemented(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"logs"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), output.ERR_NOT_IMPLEMENTED)
}

// TestLogsCommand_Help verifies the command registers and shows usage.
func TestLogsCommand_Help(t *testing.T) {
	t.Parallel()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	root := cmd.NewRootCommand()
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"logs", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Usage:")
	assert.Empty(t, stderr.String())
}

// TestLogsCommand_EnvironmentFlagAccepted verifies --environment does not cause a flag parse error.
func TestLogsCommand_EnvironmentFlagAccepted(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"logs", "--environment", "dev"})

	err := root.Execute()

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unknown flag",
		"--environment must be accepted on logs command")
	assert.Contains(t, err.Error(), output.ERR_NOT_IMPLEMENTED)
}

// TestLogsCommand_ComponentFlagAccepted verifies --component does not cause a flag parse error.
func TestLogsCommand_ComponentFlagAccepted(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"logs", "--component", "my-source"})

	err := root.Execute()

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unknown flag",
		"--component must be a registered flag on logs command")
	assert.Contains(t, err.Error(), output.ERR_NOT_IMPLEMENTED)
}

// TestLogsCommand_KindFlagAccepted verifies --kind does not cause a flag parse error.
func TestLogsCommand_KindFlagAccepted(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"logs", "--kind", "source"})

	err := root.Execute()

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unknown flag",
		"--kind must be a registered flag on logs command")
	assert.Contains(t, err.Error(), output.ERR_NOT_IMPLEMENTED)
}

// TestLogsCommand_TailFlagAccepted verifies --tail does not cause a flag parse error.
func TestLogsCommand_TailFlagAccepted(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"logs", "--tail", "100"})

	err := root.Execute()

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unknown flag",
		"--tail must be a registered flag on logs command")
	assert.Contains(t, err.Error(), output.ERR_NOT_IMPLEMENTED)
}

// TestLogsCommand_FollowFlagAccepted verifies --follow compatibility alias does not cause a flag parse error.
func TestLogsCommand_FollowFlagAccepted(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"logs", "--follow"})

	err := root.Execute()

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unknown flag",
		"--follow must be a registered flag on logs command")
	assert.Contains(t, err.Error(), output.ERR_NOT_IMPLEMENTED)
}

// TestLogsCommand_OutputJSONFlagAccepted verifies --output json does not cause a flag parse error.
func TestLogsCommand_OutputJSONFlagAccepted(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--output", "json", "logs"})

	err := root.Execute()

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unknown flag")
	assert.Contains(t, err.Error(), output.ERR_NOT_IMPLEMENTED)
}
