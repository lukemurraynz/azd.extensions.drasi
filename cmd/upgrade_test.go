package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/azure/azd.extensions.drasi/cmd"
	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpgrade_ERRNotImplementedInError verifies the error returned by upgrade
// contains ERR_NOT_IMPLEMENTED so callers can detect it.
func TestUpgrade_ERRNotImplementedInError(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	root.SetArgs([]string{"upgrade"})
	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), string(output.ERR_NOT_IMPLEMENTED),
		"error must expose ERR_NOT_IMPLEMENTED so main() can set correct exit code")
}

// TestUpgrade_SpecificMessageInStderr verifies the spec-required message appears in stderr.
func TestUpgrade_SpecificMessageInStderr(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	var errBuf bytes.Buffer
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&errBuf)

	root.SetArgs([]string{"upgrade"})
	_ = root.Execute()

	errOut := errBuf.String()
	assert.Contains(t, errOut, "upgrade is planned for a future release",
		"stderr must contain the spec-required roadmap message")
	assert.Contains(t, errOut, "https://github.com/azure/azd.extensions.drasi/issues",
		"stderr must contain the roadmap URL")
}

// TestUpgrade_EnvironmentFlagAccepted verifies that --environment is a registered flag
// so callers do not receive an "unknown flag" cobra parse error.
func TestUpgrade_EnvironmentFlagAccepted(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	root.SetArgs([]string{"upgrade", "--environment", "dev"})
	err := root.Execute()

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unknown flag",
		"--environment must be a registered flag on upgrade")
	assert.Contains(t, err.Error(), string(output.ERR_NOT_IMPLEMENTED),
		"error must still be ERR_NOT_IMPLEMENTED even when --environment is provided")
}

// TestUpgrade_JSONOutput_ContainsErrorCode verifies that --output json mode surfaces the
// error code in the stderr payload.
func TestUpgrade_JSONOutput_ContainsErrorCode(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	var errBuf bytes.Buffer
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&errBuf)

	root.SetArgs([]string{"--output", "json", "upgrade"})
	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), string(output.ERR_NOT_IMPLEMENTED))

	// In JSON mode the FormatError call still writes to stderr.
	errOut := errBuf.String()
	assert.True(t,
		strings.Contains(errOut, string(output.ERR_NOT_IMPLEMENTED)),
		"stderr must contain ERR_NOT_IMPLEMENTED in JSON output mode, got: %s", errOut,
	)
}
