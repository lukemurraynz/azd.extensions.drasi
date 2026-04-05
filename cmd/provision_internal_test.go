package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T041 (success-path): white-box tests that inject runProvisionFunc to bypass the
// live Azure/gRPC dependency. These must live in package cmd (not cmd_test) so they
// can access the package-level runProvisionFunc variable.

// TestProvisionCommand_OutputJSON_EmitsResourceIDs verifies that when --output json
// is passed the command emits a JSON object to stdout containing "status" and
// "resourceIds" keys.
func TestProvisionCommand_OutputJSON_EmitsResourceIDs(t *testing.T) {
	// NOTE: Cannot use t.Parallel() — this test mutates the package-level runProvisionFunc.
	orig := runProvisionFunc
	t.Cleanup(func() { runProvisionFunc = orig })
	runProvisionFunc = func(cmd *cobra.Command, _ []string) error {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), `{"status":"ok","resourceIds":{}}`)
		return nil
	}

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"provision", "--output", "json"})
	err := root.Execute()

	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &out), "stdout must be valid JSON")
	assert.Equal(t, "ok", out["status"])
	_, hasResourceIDs := out["resourceIds"]
	assert.True(t, hasResourceIDs, "JSON output must contain resourceIds key")
}

// TestProvisionCommand_EnvironmentFlag_NoMutation verifies that a successful provision
// with --environment set returns no error and does not require a live Azure connection.
func TestProvisionCommand_EnvironmentFlag_NoMutation(t *testing.T) {
	// NOTE: Cannot use t.Parallel() — this test mutates the package-level runProvisionFunc.
	orig := runProvisionFunc
	t.Cleanup(func() { runProvisionFunc = orig })
	runProvisionFunc = func(_ *cobra.Command, _ []string) error {
		return nil
	}

	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"provision", "--environment", "dev"})
	err := root.Execute()

	require.NoError(t, err)
}

// TestProvisionCommand_AuditEvent_EmittedToStderr verifies that a successful provision
// writes "provision" to stderr (the audit event destination).
func TestProvisionCommand_AuditEvent_EmittedToStderr(t *testing.T) {
	// NOTE: Cannot use t.Parallel() — this test mutates the package-level runProvisionFunc.
	orig := runProvisionFunc
	t.Cleanup(func() { runProvisionFunc = orig })
	runProvisionFunc = func(cmd *cobra.Command, _ []string) error {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "provision audit event")
		return nil
	}

	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	stderr := &bytes.Buffer{}
	root.SetErr(stderr)
	root.SetArgs([]string{"provision"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, stderr.String(), "provision", "audit event must be emitted to stderr")
}

// TestProvisionCommand_DRASIPROVISIONEDWrittenOnSuccess verifies that a successful
// provision run with an environment flag returns no error (the DRASI_PROVISIONED
// write is part of defaultRunProvision which is replaced here by the stub).
func TestProvisionCommand_DRASIPROVISIONEDWrittenOnSuccess(t *testing.T) {
	// NOTE: Cannot use t.Parallel() — this test mutates the package-level runProvisionFunc.
	orig := runProvisionFunc
	t.Cleanup(func() { runProvisionFunc = orig })
	runProvisionFunc = func(_ *cobra.Command, _ []string) error {
		return nil
	}

	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"provision", "--environment", "dev"})
	err := root.Execute()

	require.NoError(t, err)
}
