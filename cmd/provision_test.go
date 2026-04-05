package cmd_test

import (
	"bytes"
	"testing"

	"github.com/azure/azd.extensions.drasi/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T041: provision command CLI flag-registration tests.
//
// These black-box tests confirm that flags are declared correctly on the provision
// command. They do not require a live Azure connection and pass regardless of the
// underlying provision implementation.
//
// Success-path and injection tests live in provision_internal_test.go (package cmd)
// because they override the package-level runProvisionFunc variable.

// TestProvisionCommand_EnvironmentFlag_Accepted verifies that --environment is registered
// on the provision command so callers do not get a cobra "unknown flag" parse error.
func TestProvisionCommand_EnvironmentFlag_Accepted(t *testing.T) {
	// NOTE: Not parallel — black-box tests run in the same process as white-box tests
	// that mutate the package-level runProvisionFunc; parallelism causes races.
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"provision", "--environment", "dev"})
	err := root.Execute()

	// The command will fail (no live Azure) but must not fail with "unknown flag".
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unknown flag", "--environment must be a registered flag on provision")
}

// TestProvisionCommand_OutputJSONFlag_Accepted verifies that --output json is accepted
// by the provision command without a flag-parse error.
func TestProvisionCommand_OutputJSONFlag_Accepted(t *testing.T) {
	// NOTE: Not parallel — see TestProvisionCommand_EnvironmentFlag_Accepted for rationale.
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"provision", "--output", "json"})
	err := root.Execute()

	// Will fail (no live Azure) but the flag must parse cleanly.
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unknown flag", "--output must be accepted by provision")
}

// TestProvisionCommand_NoAuth_ExitsTwo verifies that when Azure credentials are
// missing the command returns an error with ERR_NO_AUTH.
func TestProvisionCommand_NoAuth_ExitsTwo(t *testing.T) {
	// NOTE: Not parallel — see TestProvisionCommand_EnvironmentFlag_Accepted for rationale.
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"provision"})
	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ERR_NO_AUTH",
		"error must be ERR_NO_AUTH when Azure credentials are absent")
}
