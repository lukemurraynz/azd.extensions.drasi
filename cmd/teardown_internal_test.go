package cmd

import (
	"bytes"
	"testing"

	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeardownCommand_ForceFlag_SkipsPrompt(t *testing.T) {
	saveConfirmVars(t)

	called := false
	confirmFunc = func(_ string, _ *bool) error {
		called = true
		return nil
	}
	// Force=true should skip the prompt entirely.
	// The command will fail later (no azd server), but that's expected.
	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"teardown", "--force"})

	_ = root.Execute()

	assert.False(t, called, "confirmFunc must not be called when --force is set")
}

func TestTeardownCommand_NoForce_NonInteractive_Error(t *testing.T) {
	saveConfirmVars(t)

	isTTYFunc = func() bool { return false }
	confirmFunc = func(_ string, _ *bool) error {
		t.Fatal("confirmFunc must not be called in non-interactive mode")
		return nil
	}

	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"teardown"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), output.ERR_FORCE_REQUIRED)
	assert.Contains(t, err.Error(), "--force")
}
