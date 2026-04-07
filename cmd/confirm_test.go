package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// saveConfirmVars saves and restores confirmFunc and isTTYFunc package-level vars.
func saveConfirmVars(t *testing.T) {
	t.Helper()
	origConfirmFunc := confirmFunc
	origIsTTYFunc := isTTYFunc
	t.Cleanup(func() {
		confirmFunc = origConfirmFunc
		isTTYFunc = origIsTTYFunc
	})
}

func TestConfirmDestructive_Force(t *testing.T) {
	saveConfirmVars(t)

	called := false
	confirmFunc = func(_ string, _ *bool) error {
		called = true
		return nil
	}

	result, err := ConfirmDestructive("Delete everything?", true)

	require.NoError(t, err)
	assert.True(t, result)
	assert.False(t, called, "confirmFunc must not be called when force=true")
}

func TestConfirmDestructive_NonInteractive(t *testing.T) {
	saveConfirmVars(t)

	isTTYFunc = func() bool { return false }
	confirmFunc = func(_ string, _ *bool) error {
		t.Fatal("confirmFunc must not be called in non-interactive mode")
		return nil
	}

	result, err := ConfirmDestructive("Delete everything?", false)

	require.Error(t, err)
	assert.False(t, result)
	assert.Contains(t, err.Error(), "--force")
}

func TestConfirmDestructive_ConfirmYes(t *testing.T) {
	saveConfirmVars(t)

	isTTYFunc = func() bool { return true }
	confirmFunc = func(_ string, result *bool) error {
		*result = true
		return nil
	}

	result, err := ConfirmDestructive("Delete everything?", false)

	require.NoError(t, err)
	assert.True(t, result)
}

func TestConfirmDestructive_ConfirmNo(t *testing.T) {
	saveConfirmVars(t)

	isTTYFunc = func() bool { return true }
	confirmFunc = func(_ string, result *bool) error {
		*result = false
		return nil
	}

	result, err := ConfirmDestructive("Delete everything?", false)

	require.NoError(t, err)
	assert.False(t, result)
}
