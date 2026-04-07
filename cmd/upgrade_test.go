package cmd_test

import (
	"bytes"
	"testing"

	"github.com/azure/azd.extensions.drasi/cmd"
	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpgrade_ForceRequired(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"upgrade"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), output.ERR_FORCE_REQUIRED)
}

func TestUpgrade_EnvironmentFlagAccepted(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	// --environment is a root persistent flag, so it comes before the subcommand.
	root.SetArgs([]string{"--environment", "dev", "upgrade", "--force"})
	err := root.Execute()

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unknown flag",
		"--environment must be a registered root persistent flag")
	assert.True(t,
		bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_NOT_FOUND)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_VERSION)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_ERROR)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_NO_AUTH)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_AKS_CONTEXT_NOT_FOUND)),
		"upgrade must fail with known error code; got: %s", err.Error(),
	)
}

func TestUpgrade_JSONOutput_ErrorCode(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	var errBuf bytes.Buffer
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&errBuf)

	root.SetArgs([]string{"--output", "json", "upgrade", "--force"})
	err := root.Execute()

	require.Error(t, err)
	assert.True(t,
		bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_NOT_FOUND)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_VERSION)) ||
			bytes.Contains([]byte(err.Error()), []byte(output.ERR_DRASI_CLI_ERROR)),
		"upgrade must fail with known drasi CLI code; got: %s", err.Error(),
	)
	assert.True(t,
		bytes.Contains(errBuf.Bytes(), []byte(output.ERR_DRASI_CLI_NOT_FOUND)) ||
			bytes.Contains(errBuf.Bytes(), []byte(output.ERR_DRASI_CLI_VERSION)) ||
			bytes.Contains(errBuf.Bytes(), []byte(output.ERR_DRASI_CLI_ERROR)),
		"stderr must include known drasi CLI error code; got: %s", errBuf.String(),
	)
}
