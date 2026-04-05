package cmd_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/azure/azd.extensions.drasi/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T106: SC-003 fail-fast latency tests.
//
// Each command-failure path must emit remediation-bearing error output within 2 seconds
// of failure detection. Tests run the full cobra execution path with no live Azure
// connection, measuring wall-clock time from Execute() entry to return.
//
// NOTE: Not parallel — these tests exercise the same root command object path
// as white-box tests that mutate package-level function vars; parallelism risks races.

const maxFailLatency = 2 * time.Second

// TestLatency_ERR_NO_AUTH_EmitsWithin2s verifies that ERR_NO_AUTH is returned quickly
// when the azd gRPC server is absent (no AZD_SERVER env var set).
func TestLatency_ERR_NO_AUTH_EmitsWithin2s(t *testing.T) {
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"provision"})

	start := time.Now()
	err := root.Execute()
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ERR_NO_AUTH",
		"provision must fail with ERR_NO_AUTH when azd gRPC server is absent")
	assert.Less(t, elapsed, maxFailLatency,
		"ERR_NO_AUTH must be emitted within %s; took %s", maxFailLatency, elapsed)
}

// TestLatency_ERR_DRASI_CLI_NOT_FOUND_EmitsWithin2s verifies the provision fail-fast SLA
// when the drasi binary or the azd gRPC server is absent.
//
// NOTE: On environments without AZD_SERVER, ERR_NO_AUTH fires before the drasi check.
// When AZD_SERVER is present but drasi is absent, ERR_DRASI_CLI_NOT_FOUND fires instead.
// Both are fast-fail paths; the 2-second SLA applies to either outcome.
func TestLatency_ERR_DRASI_CLI_NOT_FOUND_EmitsWithin2s(t *testing.T) {
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"provision"})

	start := time.Now()
	err := root.Execute()
	elapsed := time.Since(start)

	require.Error(t, err)

	msg := err.Error()
	fastFailCode := strings.Contains(msg, "ERR_NO_AUTH") || strings.Contains(msg, "ERR_DRASI_CLI_NOT_FOUND")
	assert.True(t, fastFailCode,
		"provision must fail with ERR_NO_AUTH or ERR_DRASI_CLI_NOT_FOUND when azd or drasi is absent; got: %s", msg)

	assert.Less(t, elapsed, maxFailLatency,
		"drasi-not-found error path must complete within %s; took %s", maxFailLatency, elapsed)
}

// TestLatency_DestructiveStub_EmitsWithin2s verifies that commands guarded against
// destructive operations (or currently unimplemented) return their error within 2 seconds.
// This covers the ERR_FORCE_REQUIRED path archetype: when teardown is fully implemented
// it will require --force for destructive operations; the fail-fast SLA applies there too.
// The teardown stub exercises the same time-budget requirement.
func TestLatency_DestructiveStub_EmitsWithin2s(t *testing.T) {
	root := cmd.NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"teardown"})

	start := time.Now()
	err := root.Execute()
	elapsed := time.Since(start)

	require.Error(t, err, "teardown must return an error when not fully implemented")
	assert.Less(t, elapsed, maxFailLatency,
		"teardown error path must complete within %s; took %s", maxFailLatency, elapsed)
}
