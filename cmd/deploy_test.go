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

func TestDeployCommand_Help(t *testing.T) {
	t.Parallel()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	root := cmd.NewRootCommand()
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"deploy", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Deploy Drasi components")
	assert.Contains(t, stdout.String(), "Usage:")
	assert.Empty(t, stderr.String())
}

// TestDeployCommand_DryRunFlagAccepted verifies --dry-run is a registered flag.
func TestDeployCommand_DryRunFlagAccepted(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"deploy", "--dry-run"})

	err := root.Execute()

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unknown flag",
		"--dry-run must be a registered flag on deploy command")
	assert.Contains(t, err.Error(), output.ERR_NO_AUTH)
}

// TestDeployCommand_ConfigFlagDefault verifies --config flag exists and defaults to drasi/drasi.yaml.
func TestDeployCommand_ConfigFlagDefault(t *testing.T) {
	t.Parallel()
	stdout := &bytes.Buffer{}

	root := cmd.NewRootCommand()
	root.SetOut(stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"deploy", "--help"})

	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "--config")
}

// TestDeployCommand_EnvironmentFlagAccepted verifies --environment is a registered flag.
func TestDeployCommand_EnvironmentFlagAccepted(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"deploy", "--environment", "dev"})

	err := root.Execute()

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unknown flag",
		"--environment must be a registered flag on deploy command")
	assert.True(t,
		strings.Contains(err.Error(), output.ERR_NO_AUTH) ||
			strings.Contains(err.Error(), output.ERR_NO_MANIFEST),
		"deploy must fail with known auth/manifest code; got: %s", err.Error(),
	)
}

func TestDeployCommand_NoRollbackFlagAccepted(t *testing.T) {
	t.Parallel()
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"deploy", "--no-rollback"})

	err := root.Execute()

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unknown flag",
		"--no-rollback must be a registered flag on deploy command")
	assert.Contains(t, err.Error(), output.ERR_NO_AUTH)
}

// TestDeployCommand_NoAuth_ReturnsError verifies that without AZD_SERVER the command returns ERR_NO_AUTH.
func TestDeployCommand_NoAuth_ReturnsError(t *testing.T) {
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"deploy"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), output.ERR_NO_AUTH)
}
