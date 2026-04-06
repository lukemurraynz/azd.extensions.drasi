//go:build integration

package parity_test

// T108: FR-037 — smoke command matrix parity tests.
//
// Exercises the full cobra command tree via NewRootCommand() to assert:
//   - All documented commands are registered and reachable.
//   - --help exits with code 0 and outputs a usage block.
//   - Absent-dependency commands fail with a known error code
//     (ERR_NO_AUTH, ERR_FORCE_REQUIRED, ERR_NO_MANIFEST, etc.).
//   - JSON output schema is consistent for commands that support --output json.
//
// These tests do NOT require a live Azure connection, AKS cluster, or Drasi runtime.
// They validate the CLI surface contract so regressions in command registration,
// flag wiring, or error-code emission are caught at CI time.

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/azure/azd.extensions.drasi/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// commandCase describes a single smoke scenario for one command invocation.
type commandCase struct {
	// name is the subtest label.
	name string
	// args are the cobra arguments (command name + flags).
	args []string
	// wantErrCode is a substring expected in err.Error() when the command fails.
	// Empty string means the command must succeed (err == nil).
	wantErrCode string
	// wantOutputContains is a substring expected in stdout when wantErrCode is empty.
	wantOutputContains string
}

// TestCommandMatrix_HelpExitsClean verifies that --help on every registered top-level
// command exits without error and includes the command name in the usage output.
func TestCommandMatrix_HelpExitsClean(t *testing.T) {
	t.Parallel()

	commands := []string{
		"validate",
		"init",
		"provision",
		"deploy",
		"status",
		"logs",
		"diagnose",
		"teardown",
		"upgrade",
		"listen",
	}

	for _, cmdName := range commands {
		cmdName := cmdName
		t.Run(cmdName+"_help", func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			root := cmd.NewRootCommand()
			root.SetOut(&stdout)
			root.SetErr(&stderr)
			root.SetArgs([]string{cmdName, "--help"})

			err := root.Execute()
			// cobra exits help with nil, not an error.
			assert.NoError(t, err, "%s --help must exit cleanly", cmdName)
			assert.Contains(t, stdout.String(), cmdName,
				"%s --help must include the command name in usage output", cmdName)
		})
	}
}

// TestCommandMatrix_ValidateNoManifest verifies that "validate" fails with ERR_NO_MANIFEST
// when the default drasi/drasi.yaml file does not exist in the working directory.
func TestCommandMatrix_ValidateNoManifest(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	root := cmd.NewRootCommand()
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	// Use a non-existent config path to trigger ERR_NO_MANIFEST.
	root.SetArgs([]string{"validate", "--config", "nonexistent/drasi.yaml"})

	err := root.Execute()
	require.Error(t, err, "validate with a missing manifest must return an error")
	assert.Contains(t, err.Error(), "ERR_NO_MANIFEST",
		"validate must emit ERR_NO_MANIFEST for a missing manifest file")
}

// TestCommandMatrix_ValidateNoManifest_JSONSchema verifies that "validate --output json"
// emits a JSON object with a "status" key when the manifest is missing.
func TestCommandMatrix_ValidateNoManifest_JSONSchema(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	root := cmd.NewRootCommand()
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"--output", "json", "validate", "--config", "nonexistent/drasi.yaml"})

	_ = root.Execute() // error expected; we only care about JSON shape in stderr

	// The error payload is written to stderr; check it parses as JSON with an "error" key.
	errOutput := stderr.String()
	require.NotEmpty(t, errOutput, "validate --output json must emit JSON to stderr on error")

	var payload map[string]any
	err := json.Unmarshal([]byte(strings.TrimSpace(errOutput)), &payload)
	require.NoError(t, err, "validate error output must be valid JSON; got: %q", errOutput)
	// ErrorResponse shape: {"status":"error","code":"ERR_...","message":"...","remediation":"..."}
	assert.Equal(t, "error", payload["status"],
		"validate JSON error payload must have status=error")
	assert.Contains(t, payload, "code",
		"validate JSON error payload must have a 'code' key")
}

// TestCommandMatrix_InitBlankTemplate verifies that "init --template blank" succeeds in a
// temp directory and emits the expected output.
func TestCommandMatrix_InitBlankTemplate(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	var stdout, stderr bytes.Buffer
	root := cmd.NewRootCommand()
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"init", "--template", "blank", "--output-dir", dir})

	err := root.Execute()
	require.NoError(t, err, "init --template blank must succeed in a temp directory")
	assert.Contains(t, stdout.String(), "Initialized",
		"init must confirm project initialization in stdout")
}

// TestCommandMatrix_InitBlankTemplate_JSONSchema verifies that "init --output json" emits
// a JSON object with "status" and "files" keys on success.
func TestCommandMatrix_InitBlankTemplate_JSONSchema(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	var stdout, stderr bytes.Buffer
	root := cmd.NewRootCommand()
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"--output", "json", "init", "--template", "blank", "--output-dir", dir})

	err := root.Execute()
	require.NoError(t, err, "init --output json must succeed")

	var payload map[string]any
	decodeErr := json.Unmarshal([]byte(strings.TrimSpace(stdout.String())), &payload)
	require.NoError(t, decodeErr, "init JSON output must be valid JSON; got: %q", stdout.String())
	assert.Equal(t, "ok", payload["status"], "init JSON output must have status=ok")
	assert.Contains(t, payload, "files", "init JSON output must have a 'files' key")
}

// TestCommandMatrix_ProvisionFailsFast verifies that "provision" fails with a known
// fast-fail error code when no azd gRPC server is available.
func TestCommandMatrix_ProvisionFailsFast(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	root := cmd.NewRootCommand()
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"provision"})

	err := root.Execute()
	require.Error(t, err, "provision must fail when azd gRPC server is absent")

	msg := err.Error()
	isFastFail := strings.Contains(msg, "ERR_NO_AUTH") || strings.Contains(msg, "ERR_DRASI_CLI_NOT_FOUND")
	assert.True(t, isFastFail,
		"provision must fail with ERR_NO_AUTH or ERR_DRASI_CLI_NOT_FOUND; got: %s", msg)
}

// TestCommandMatrix_CommandsFailWithKnownCodes verifies command error surfaces are stable.
func TestCommandMatrix_CommandsFailWithKnownCodes(t *testing.T) {
	t.Parallel()

	cases := []commandCase{
		{name: "status", args: []string{"status"}, wantErrCode: "ERR_DRASI_CLI_NOT_FOUND"},
		{name: "logs", args: []string{"logs"}, wantErrCode: "ERR_VALIDATION_FAILED"},
		{name: "diagnose", args: []string{"diagnose"}, wantErrCode: "ERR_DRASI_CLI_NOT_FOUND"},
		{name: "teardown", args: []string{"teardown"}, wantErrCode: "ERR_FORCE_REQUIRED"},
		{name: "upgrade", args: []string{"upgrade"}, wantErrCode: "ERR_FORCE_REQUIRED"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			root := cmd.NewRootCommand()
			root.SetOut(&stdout)
			root.SetErr(&stderr)
			root.SetArgs(tc.args)

			err := root.Execute()
			require.Error(t, err, "%s must return an error in dependency-absent test context", tc.name)
			assert.Contains(t, err.Error(), tc.wantErrCode,
				"%s must return %s; got: %s", tc.name, tc.wantErrCode, err.Error())
		})
	}
}

// TestCommandMatrix_RootHelp verifies that the root command --help exit is clean
// and includes the top-level usage description.
func TestCommandMatrix_RootHelp(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	root := cmd.NewRootCommand()
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"--help"})

	err := root.Execute()
	assert.NoError(t, err, "root --help must exit cleanly")
	// Root Short is "Manage Drasi reactive data pipeline workloads" — always present.
	assert.Contains(t, stdout.String(), "Manage Drasi",
		"root --help must include the Drasi description in the usage output")
}
