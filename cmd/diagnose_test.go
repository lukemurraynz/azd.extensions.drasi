package cmd_test

import (
	"bytes"
	"testing"

	"github.com/azure/azd.extensions.drasi/cmd"
	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDiagnoseCommand_NotImplemented verifies the current stub returns ERR_NOT_IMPLEMENTED.
func TestDiagnoseCommand_NotImplemented(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"diagnose"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), output.ERR_NOT_IMPLEMENTED)
}

// TestDiagnoseCommand_Help verifies the command registers and shows usage.
func TestDiagnoseCommand_Help(t *testing.T) {
	t.Parallel()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	root := cmd.NewRootCommand()
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"diagnose", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Usage:")
	assert.Empty(t, stderr.String())
}

// TestDiagnoseCommand_EnvironmentFlagAccepted verifies --environment does not cause a flag parse error.
func TestDiagnoseCommand_EnvironmentFlagAccepted(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"diagnose", "--environment", "dev"})

	err := root.Execute()

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unknown flag",
		"--environment must be accepted on diagnose command")
	assert.Contains(t, err.Error(), output.ERR_NOT_IMPLEMENTED)
}

// TestDiagnoseCommand_OutputJSONFlagAccepted verifies --output json does not cause a flag parse error.
func TestDiagnoseCommand_OutputJSONFlagAccepted(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--output", "json", "diagnose"})

	err := root.Execute()

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unknown flag")
	assert.Contains(t, err.Error(), output.ERR_NOT_IMPLEMENTED)
}
