package cmd_test

import (
	"bytes"
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/cmd"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLogsCommand_ValidationRequiresComponentAndKind verifies required logs selectors.
func TestLogsCommand_ValidationRequiresComponentAndKind(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"logs"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), output.ERR_VALIDATION_FAILED)
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
	assert.Contains(t, err.Error(), output.ERR_VALIDATION_FAILED)
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
	assert.Contains(t, err.Error(), output.ERR_VALIDATION_FAILED)
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
	assert.Contains(t, err.Error(), output.ERR_VALIDATION_FAILED)
}

func TestLogsCommand_RootEnvironmentFlagAccepted(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--environment", "dev", "logs", "--component", "x", "--kind", "source"})

	err := root.Execute()

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unknown flag")
	assert.True(t,
		bytes.Contains([]byte(err.Error()), []byte(output.ERR_NO_AUTH)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_AKS_CONTEXT_NOT_FOUND)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_NOT_FOUND)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_VALIDATION_FAILED)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_VERSION)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_ERROR)),
		"logs must fail with known auth/context/cli code; got: %s", err.Error(),
	)
}
