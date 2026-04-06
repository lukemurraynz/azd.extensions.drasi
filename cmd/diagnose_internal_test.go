package cmd

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/azure/azd.extensions.drasi/internal/drasi"
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

func TestDiagnoseCommand_JSONSuccess_EmitsChecks(t *testing.T) {
	origClientFactory := newDiagnoseDrasiClient
	origKubectlVersion := kubectlClientVersionCheck
	origKubectlPath := kubectlOnPathCheck
	origIsDaprReady := isDaprReady
	t.Cleanup(func() {
		newDiagnoseDrasiClient = origClientFactory
		kubectlClientVersionCheck = origKubectlVersion
		kubectlOnPathCheck = origKubectlPath
		isDaprReady = origIsDaprReady
	})

	client := &fakeDiagnoseClient{}
	newDiagnoseDrasiClient = func() diagnoseDrasiClient { return client }
	kubectlClientVersionCheck = func(context.Context, string) error { return nil }
	kubectlOnPathCheck = func() error { return nil }
	isDaprReady = func(context.Context, string) (bool, string, error) { return true, "Dapr operator pod is present", nil }

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
	origClientFactory := newDiagnoseDrasiClient
	origKubectlVersion := kubectlClientVersionCheck
	origKubectlPath := kubectlOnPathCheck
	origIsDaprReady := isDaprReady
	t.Cleanup(func() {
		newDiagnoseDrasiClient = origClientFactory
		kubectlClientVersionCheck = origKubectlVersion
		kubectlOnPathCheck = origKubectlPath
		isDaprReady = origIsDaprReady
	})

	client := &fakeDiagnoseClient{}
	newDiagnoseDrasiClient = func() diagnoseDrasiClient { return client }
	kubectlClientVersionCheck = func(context.Context, string) error { return nil }
	kubectlOnPathCheck = func() error { return nil }
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
	origClientFactory := newDiagnoseDrasiClient
	origKubectlVersion := kubectlClientVersionCheck
	origKubectlPath := kubectlOnPathCheck
	origIsDaprReady := isDaprReady
	t.Cleanup(func() {
		newDiagnoseDrasiClient = origClientFactory
		kubectlClientVersionCheck = origKubectlVersion
		kubectlOnPathCheck = origKubectlPath
		isDaprReady = origIsDaprReady
	})

	client := &fakeDiagnoseClient{listErr: errors.New("ERR_DRASI_CLI_ERROR: list failed")}
	newDiagnoseDrasiClient = func() diagnoseDrasiClient { return client }
	kubectlClientVersionCheck = func(context.Context, string) error { return nil }
	kubectlOnPathCheck = func() error { return nil }
	isDaprReady = func(context.Context, string) (bool, string, error) { return true, "Dapr operator pod is present", nil }

	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"diagnose"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ERR_DRASI_CLI_ERROR")
}
