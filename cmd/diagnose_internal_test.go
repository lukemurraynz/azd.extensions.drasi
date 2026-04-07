package cmd

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/drasi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeDiagnoseClient struct {
	checkVersionErr error
	listErr         error
	lastKind        string
	lastContext     string
}

func (f *fakeDiagnoseClient) CheckVersion(_ context.Context) error {
	return f.checkVersionErr
}

func (f *fakeDiagnoseClient) ListComponents(_ context.Context, kind string) ([]drasi.ComponentSummary, error) {
	f.lastKind = kind
	f.lastContext = ""
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []drasi.ComponentSummary{}, nil
}

func (f *fakeDiagnoseClient) ListComponentsInContext(_ context.Context, kind, kubeContext string) ([]drasi.ComponentSummary, error) {
	f.lastKind = kind
	f.lastContext = kubeContext
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []drasi.ComponentSummary{}, nil
}

// saveDiagnoseVars saves all package-level var functions and returns a restore function.
func saveDiagnoseVars(t *testing.T) {
	t.Helper()
	origClientFactory := newDiagnoseDrasiClient
	origKubectlVersion := kubectlClientVersionCheck
	origKubectlPath := kubectlOnPathCheck
	origIsDaprReady := isDaprReady
	origAzKeyVaultCheck := azKeyVaultCheck
	origAzLogAnalyticsCheck := azLogAnalyticsCheck
	t.Cleanup(func() {
		newDiagnoseDrasiClient = origClientFactory
		kubectlClientVersionCheck = origKubectlVersion
		kubectlOnPathCheck = origKubectlPath
		isDaprReady = origIsDaprReady
		azKeyVaultCheck = origAzKeyVaultCheck
		azLogAnalyticsCheck = origAzLogAnalyticsCheck
	})
}

// stubAllChecksPass mocks all prerequisite checks to pass so the command
// reaches the Key Vault / Log Analytics checks.
func stubAllChecksPass(client diagnoseDrasiClient) {
	newDiagnoseDrasiClient = func() diagnoseDrasiClient { return client }
	kubectlClientVersionCheck = func(context.Context, string) error { return nil }
	kubectlOnPathCheck = func() error { return nil }
	isDaprReady = func(context.Context, string) (bool, string, error) { return true, "Dapr operator pod is present", nil }
}

func TestDiagnoseCommand_JSONSuccess_EmitsChecks(t *testing.T) {
	saveDiagnoseVars(t)

	client := &fakeDiagnoseClient{}
	stubAllChecksPass(client)

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"--output", "json", "diagnose"})

	err := root.Execute()
	require.NoError(t, err)
	assert.Equal(t, "source", client.lastKind)
	assert.Contains(t, stdout.String(), `"status": "ok"`)
	assert.Contains(t, stdout.String(), `"checks"`)
	assert.Contains(t, stdout.String(), `"aks-connectivity"`)
	assert.Empty(t, stderr.String())
}

func TestDiagnoseCommand_DaprNotReady_ReturnsError(t *testing.T) {
	saveDiagnoseVars(t)

	client := &fakeDiagnoseClient{}
	stubAllChecksPass(client)
	isDaprReady = func(context.Context, string) (bool, string, error) { return false, "no Dapr operator pod found", nil }

	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"diagnose"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ERR_DAPR_NOT_READY")
}

func TestDiagnoseCommand_DrasiListFailure_ReturnsError(t *testing.T) {
	saveDiagnoseVars(t)

	client := &fakeDiagnoseClient{listErr: errors.New("ERR_DRASI_CLI_ERROR: list failed")}
	stubAllChecksPass(client)

	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"diagnose"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ERR_DRASI_CLI_ERROR")
}

func TestDiagnoseKeyVaultOk(t *testing.T) {
	saveDiagnoseVars(t)
	t.Setenv("AZURE_KEYVAULT_NAME", "test-vault")

	client := &fakeDiagnoseClient{}
	stubAllChecksPass(client)
	azKeyVaultCheck = func(_ context.Context, vaultName string) (string, string, error) {
		return "ok", "Key Vault " + vaultName + " is accessible", nil
	}

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--output", "json", "diagnose"})

	err := root.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), `"key-vault-auth"`)
	assert.Contains(t, stdout.String(), `"ok"`)
	assert.Contains(t, stdout.String(), "Key Vault test-vault is accessible")
}

func TestDiagnoseKeyVaultFailed(t *testing.T) {
	saveDiagnoseVars(t)
	t.Setenv("AZURE_KEYVAULT_NAME", "missing-vault")

	client := &fakeDiagnoseClient{}
	stubAllChecksPass(client)
	azKeyVaultCheck = func(_ context.Context, _ string) (string, string, error) {
		return "failed", "vault not found", errors.New("exit status 1")
	}

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--output", "json", "diagnose"})

	err := root.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), `"key-vault-auth"`)
	assert.Contains(t, stdout.String(), `"failed"`)
	assert.Contains(t, stdout.String(), "Key Vault Secrets User")
}

func TestDiagnoseKeyVaultSkipped(t *testing.T) {
	saveDiagnoseVars(t)
	// AZURE_KEYVAULT_NAME is not set — triggers "skipped"

	client := &fakeDiagnoseClient{}
	stubAllChecksPass(client)

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--output", "json", "diagnose"})

	err := root.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), `"key-vault-auth"`)
	assert.Contains(t, stdout.String(), `"skipped"`)
	assert.Contains(t, stdout.String(), "AZURE_KEYVAULT_NAME not set")
}

func TestDiagnoseLogAnalyticsOk(t *testing.T) {
	saveDiagnoseVars(t)
	t.Setenv("AZURE_LOG_ANALYTICS_WORKSPACE_NAME", "test-workspace")
	t.Setenv("AZURE_RESOURCE_GROUP", "test-rg")

	client := &fakeDiagnoseClient{}
	stubAllChecksPass(client)
	azLogAnalyticsCheck = func(_ context.Context, _ string, wsName string) (string, string, error) {
		return "ok", "Log Analytics workspace " + wsName + " is accessible", nil
	}

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--output", "json", "diagnose"})

	err := root.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), `"log-analytics"`)
	assert.Contains(t, stdout.String(), `"ok"`)
	assert.Contains(t, stdout.String(), "Log Analytics workspace test-workspace is accessible")
}

func TestDiagnoseLogAnalyticsFailed(t *testing.T) {
	saveDiagnoseVars(t)
	t.Setenv("AZURE_LOG_ANALYTICS_WORKSPACE_NAME", "missing-ws")
	t.Setenv("AZURE_RESOURCE_GROUP", "missing-rg")

	client := &fakeDiagnoseClient{}
	stubAllChecksPass(client)
	azLogAnalyticsCheck = func(_ context.Context, _ string, _ string) (string, string, error) {
		return "failed", "workspace not found", errors.New("exit status 1")
	}

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--output", "json", "diagnose"})

	err := root.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), `"log-analytics"`)
	assert.Contains(t, stdout.String(), `"failed"`)
	assert.Contains(t, stdout.String(), "Log Analytics workspace exists")
}

func TestDiagnoseLogAnalyticsSkipped(t *testing.T) {
	saveDiagnoseVars(t)
	// Neither AZURE_LOG_ANALYTICS_WORKSPACE_NAME nor AZURE_RESOURCE_GROUP set — triggers "skipped"

	client := &fakeDiagnoseClient{}
	stubAllChecksPass(client)

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--output", "json", "diagnose"})

	err := root.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), `"log-analytics"`)
	assert.Contains(t, stdout.String(), `"skipped"`)
	assert.Contains(t, stdout.String(), "AZURE_LOG_ANALYTICS_WORKSPACE_NAME or AZURE_RESOURCE_GROUP not set")
}
