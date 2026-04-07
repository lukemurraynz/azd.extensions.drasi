package cmd

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeUpgradeClient struct {
	checkVersionErr error
	version         string
	getVersionErr   error
	checkCalls      int
	getCalls        int
}

func (f *fakeUpgradeClient) CheckVersion(context.Context) error {
	f.checkCalls++
	return f.checkVersionErr
}

func (f *fakeUpgradeClient) GetVersion(context.Context) (string, error) {
	f.getCalls++
	if f.getVersionErr != nil {
		return "", f.getVersionErr
	}
	return f.version, nil
}

func saveUpgradeVars(t *testing.T) {
	t.Helper()
	origClientFactory := newUpgradeDrasiClient
	origResolveKubeContext := resolveUpgradeKubeContext
	origSwitchKubectlContext := switchUpgradeKubectlContext
	origRunDrasiCommand := runUpgradeDrasiCommand
	t.Cleanup(func() {
		newUpgradeDrasiClient = origClientFactory
		resolveUpgradeKubeContext = origResolveKubeContext
		switchUpgradeKubectlContext = origSwitchKubectlContext
		runUpgradeDrasiCommand = origRunDrasiCommand
	})
}

func TestUpgradeCommand_ForceFlag_SkipsPrompt(t *testing.T) {
	saveConfirmVars(t)

	called := false
	confirmFunc = func(_ string, _ *bool) error {
		called = true
		return nil
	}
	// Force=true should skip the prompt entirely.
	// The command will fail later (no drasi CLI), but that's expected.
	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"upgrade", "--force"})

	_ = root.Execute()

	assert.False(t, called, "confirmFunc must not be called when --force is set")
}

func TestUpgradeCommand_NoForce_NonInteractive_Error(t *testing.T) {
	saveConfirmVars(t)

	isTTYFunc = func() bool { return false }
	confirmFunc = func(_ string, _ *bool) error {
		t.Fatal("confirmFunc must not be called in non-interactive mode")
		return nil
	}

	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"upgrade"})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), output.ERR_FORCE_REQUIRED)
	assert.Contains(t, err.Error(), "--force")
}

func TestUpgradeCommand_DryRun_PrintsPlannedAction(t *testing.T) {
	saveConfirmVars(t)
	saveUpgradeVars(t)

	client := &fakeUpgradeClient{version: "0.10.1"}
	newUpgradeDrasiClient = func() upgradeDrasiClient { return client }
	resolveUpgradeKubeContext = func(context.Context, *cobra.Command, string) (string, error) { return "", nil }
	runUpgradeDrasiCommand = func(context.Context, ...string) error {
		t.Fatal("runUpgradeDrasiCommand must not be called during dry-run")
		return nil
	}

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"upgrade", "--force", "--dry-run"})

	err := root.Execute()

	require.NoError(t, err)
	assert.Equal(t, 1, client.checkCalls)
	assert.Equal(t, 1, client.getCalls)
	assert.Contains(t, stdout.String(), "Current Drasi runtime version: 0.10.1")
	assert.Contains(t, stdout.String(), "Dry-run: upgrade would reinstall the Drasi runtime using the installed CLI version.")
}

func TestUpgradeCommand_DryRun_JSONOutput_UsesUnknownVersionOnError(t *testing.T) {
	saveConfirmVars(t)
	saveUpgradeVars(t)

	client := &fakeUpgradeClient{getVersionErr: errors.New("version unavailable")}
	newUpgradeDrasiClient = func() upgradeDrasiClient { return client }
	resolveUpgradeKubeContext = func(context.Context, *cobra.Command, string) (string, error) { return "", nil }
	runUpgradeDrasiCommand = func(context.Context, ...string) error {
		t.Fatal("runUpgradeDrasiCommand must not be called during dry-run")
		return nil
	}

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--output", "json", "upgrade", "--force", "--dry-run"})

	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, stdout.String(), `"status": "dry-run"`)
	assert.Contains(t, stdout.String(), `"currentVersion": "unknown"`)
}
