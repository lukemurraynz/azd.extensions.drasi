package cmd_test

import (
	"bytes"
	"testing"

	"github.com/azure/azd.extensions.drasi/cmd"
	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDiagnoseCommand_NoDrasiCLI verifies diagnose fails when drasi CLI is absent.
func TestDiagnoseCommand_NoDrasiCLI(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"diagnose"})

	err := root.Execute()

	require.Error(t, err)
	assert.True(t,
		bytes.Contains([]byte(err.Error()), []byte(output.ERR_AKS_CONTEXT_NOT_FOUND)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_NOT_FOUND)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_VERSION)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_ERROR)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DAPR_NOT_READY)),
		"diagnose must fail with known context/drasi code; got: %s", err.Error(),
	)
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
	assert.True(t,
		bytes.Contains([]byte(err.Error()), []byte(output.ERR_AKS_CONTEXT_NOT_FOUND)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_NO_AUTH)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_NOT_FOUND)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_VERSION)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_ERROR)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DAPR_NOT_READY)),
		"diagnose must fail with known context/auth/drasi code; got: %s", err.Error(),
	)
}

func TestDiagnoseCommand_RootEnvironmentFlagAccepted(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--environment", "dev", "diagnose"})

	err := root.Execute()

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unknown flag")
	assert.True(t,
		bytes.Contains([]byte(err.Error()), []byte(output.ERR_AKS_CONTEXT_NOT_FOUND)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_NO_AUTH)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_NOT_FOUND)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_VERSION)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_ERROR)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DAPR_NOT_READY)),
		"diagnose must fail with known auth/context/drasi code; got: %s", err.Error(),
	)
}
