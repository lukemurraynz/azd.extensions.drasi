package cmd_test

import (
	"bytes"
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProgressHelper_JSONMode_NoOutput(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	root := cmd.NewRootCommand()
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"--output", "json"})

	// Build a minimal subcommand that exercises ProgressHelper.
	subCmd := cmd.NewTestableProgressCommand()
	root.AddCommand(subCmd)
	root.SetArgs([]string{"--output", "json", "_test-progress"})

	err := root.Execute()

	require.NoError(t, err)
	assert.Empty(t, stdout.String(), "stdout must be empty in JSON mode (no spinner output)")
	assert.Empty(t, stderr.String(), "stderr must be empty in JSON mode (no spinner output)")
}

func TestProgressHelper_TableMode_CreatesSpinner(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	root := cmd.NewRootCommand()
	root.SetOut(stdout)
	root.SetErr(stderr)

	subCmd := cmd.NewTestableProgressCommand()
	root.AddCommand(subCmd)
	root.SetArgs([]string{"--output", "table", "_test-progress"})

	err := root.Execute()

	require.NoError(t, err)
	assert.Empty(t, stdout.String(), "spinner must not write to stdout")
	// In table mode the spinner writes to stderr; some output is expected.
	// yacspin may or may not produce visible output depending on terminal
	// capabilities, so we only assert no error occurred and stdout is clean.
}

func TestProgressHelper_DefaultMode_CreatesSpinner(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	root := cmd.NewRootCommand()
	root.SetOut(stdout)
	root.SetErr(stderr)

	subCmd := cmd.NewTestableProgressCommand()
	root.AddCommand(subCmd)
	root.SetArgs([]string{"_test-progress"})

	err := root.Execute()

	require.NoError(t, err)
	assert.Empty(t, stdout.String(), "spinner must not write to stdout")
}
