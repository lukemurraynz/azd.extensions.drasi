package cmd_test

import (
	"bytes"
	"testing"

	"github.com/azure/azd.extensions.drasi/cmd"
	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStatusCommand_NoDrasiCLI verifies status fails with drasi CLI error when absent.
func TestStatusCommand_NoDrasiCLI(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"status"})

	err := root.Execute()

	require.Error(t, err)
	assert.True(t,
		bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_NOT_FOUND)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_VERSION)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_ERROR)),
		"status must fail with known drasi CLI code; got: %s", err.Error(),
	)
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
	assert.True(t,
		bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_NOT_FOUND)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_VERSION)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_ERROR)),
		"status must fail with known drasi CLI code; got: %s", err.Error(),
	)
}

func TestStatusCommand_RootEnvironmentFlagAccepted(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--environment", "dev", "status"})

	err := root.Execute()

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unknown flag")
	assert.True(t,
		bytes.Contains([]byte(err.Error()), []byte(output.ERR_NO_AUTH)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_AKS_CONTEXT_NOT_FOUND)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_NOT_FOUND)),
		"status must fail with known auth/context/cli code; got: %s", err.Error(),
	)
}
