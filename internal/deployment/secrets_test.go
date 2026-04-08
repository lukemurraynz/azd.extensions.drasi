package deployment

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCmdRunner struct {
	calls   []mockCmdCall
	runFunc func(ctx context.Context, stdin io.Reader, name string, args ...string) ([]byte, error)
}

type mockCmdCall struct {
	Name  string
	Args  []string
	Stdin string
}

func (m *mockCmdRunner) RunCmd(ctx context.Context, stdin io.Reader, name string, args ...string) ([]byte, error) {
	var stdinData string
	if stdin != nil {
		data, err := io.ReadAll(stdin)
		if err != nil {
			return nil, err
		}
		stdinData = string(data)
	}

	argsCopy := append([]string(nil), args...)
	m.calls = append(m.calls, mockCmdCall{
		Name:  name,
		Args:  argsCopy,
		Stdin: stdinData,
	})

	if m.runFunc != nil {
		return m.runFunc(ctx, strings.NewReader(stdinData), name, args...)
	}

	return nil, nil
}

func TestSyncSecrets_EmptyMappings_Noop(t *testing.T) {
	t.Parallel()

	runner := &mockCmdRunner{}
	err := syncSecrets(context.Background(), nil, nil, "", runner)
	require.NoError(t, err)
	assert.Empty(t, runner.calls)
}

func TestSyncSecrets_HappyPath_SingleEntry(t *testing.T) {
	t.Parallel()

	runner := &mockCmdRunner{
		runFunc: func(ctx context.Context, stdin io.Reader, name string, args ...string) ([]byte, error) {
			if name == "az" {
				return []byte("super-secret\n"), nil
			}
			return nil, nil
		},
	}

	err := syncSecrets(context.Background(), []config.SecretMapping{{
		VaultName:  "kv-main",
		SecretName: "db-password",
		K8sSecret:  "app-secrets",
		K8sKey:     "password",
	}}, nil, "", runner)
	require.NoError(t, err)
	require.Len(t, runner.calls, 2)

	assert.Equal(t, "az", runner.calls[0].Name)
	assert.Equal(t, []string{"keyvault", "secret", "show", "--vault-name", "kv-main", "--name", "db-password", "--query", "value", "-o", "tsv"}, runner.calls[0].Args)

	assert.Equal(t, "kubectl", runner.calls[1].Name)
	assert.Equal(t, []string{"apply", "-f", "-"}, runner.calls[1].Args)
	assert.Contains(t, runner.calls[1].Stdin, "name: app-secrets")
	assert.Contains(t, runner.calls[1].Stdin, "namespace: drasi-system")
	assert.Contains(t, runner.calls[1].Stdin, "password: \"super-secret\"")
}

func TestSyncSecrets_HappyPath_MultipleEntriesSameK8sSecret(t *testing.T) {
	t.Parallel()

	runner := &mockCmdRunner{
		runFunc: func(ctx context.Context, stdin io.Reader, name string, args ...string) ([]byte, error) {
			if name != "az" {
				return nil, nil
			}
			require.Len(t, args, 11)
			switch args[6] {
			case "first-secret":
				return []byte("value-one\n"), nil
			case "second-secret":
				return []byte("value-two\n"), nil
			default:
				return nil, fmt.Errorf("unexpected secret name: %s", args[6])
			}
		},
	}

	err := syncSecrets(context.Background(), []config.SecretMapping{
		{VaultName: "kv-main", SecretName: "first-secret", K8sSecret: "shared-secret", K8sKey: "username"},
		{VaultName: "kv-main", SecretName: "second-secret", K8sSecret: "shared-secret", K8sKey: "password"},
	}, nil, "", runner)
	require.NoError(t, err)

	kubectlCalls := filterCallsByName(runner.calls, "kubectl")
	require.Len(t, kubectlCalls, 1)
	assert.Contains(t, kubectlCalls[0].Stdin, "username: \"value-one\"")
	assert.Contains(t, kubectlCalls[0].Stdin, "password: \"value-two\"")
}

func TestSyncSecrets_HappyPath_DifferentK8sSecrets(t *testing.T) {
	t.Parallel()

	runner := &mockCmdRunner{
		runFunc: func(ctx context.Context, stdin io.Reader, name string, args ...string) ([]byte, error) {
			if name == "az" {
				return []byte("value\n"), nil
			}
			return nil, nil
		},
	}

	err := syncSecrets(context.Background(), []config.SecretMapping{
		{VaultName: "kv-main", SecretName: "secret-one", K8sSecret: "secret-a", K8sKey: "key1"},
		{VaultName: "kv-main", SecretName: "secret-two", K8sSecret: "secret-b", K8sKey: "key2"},
	}, nil, "", runner)
	require.NoError(t, err)

	kubectlCalls := filterCallsByName(runner.calls, "kubectl")
	assert.Len(t, kubectlCalls, 2)
}

func TestSyncSecrets_ExpandsEnvVars(t *testing.T) {
	t.Parallel()

	runner := &mockCmdRunner{
		runFunc: func(ctx context.Context, stdin io.Reader, name string, args ...string) ([]byte, error) {
			if name == "az" {
				return []byte("value\n"), nil
			}
			return nil, nil
		},
	}

	err := syncSecrets(context.Background(), []config.SecretMapping{{
		VaultName:  "$(VAULT_NAME)",
		SecretName: "$(SECRET_NAME)",
		K8sSecret:  "app-secret",
		K8sKey:     "password",
	}}, map[string]string{"VAULT_NAME": "expanded-kv", "SECRET_NAME": "expanded-secret"}, "", runner)
	require.NoError(t, err)

	require.NotEmpty(t, runner.calls)
	assert.Equal(t, []string{"keyvault", "secret", "show", "--vault-name", "expanded-kv", "--name", "expanded-secret", "--query", "value", "-o", "tsv"}, runner.calls[0].Args)
}

func TestSyncSecrets_DefaultNamespace(t *testing.T) {
	t.Parallel()

	runner := &mockCmdRunner{
		runFunc: func(ctx context.Context, stdin io.Reader, name string, args ...string) ([]byte, error) {
			if name == "az" {
				return []byte("value\n"), nil
			}
			return nil, nil
		},
	}

	err := syncSecrets(context.Background(), []config.SecretMapping{{
		VaultName:  "kv-main",
		SecretName: "db-password",
		K8sSecret:  "app-secret",
		K8sKey:     "password",
	}}, nil, "", runner)
	require.NoError(t, err)

	kubectlCalls := filterCallsByName(runner.calls, "kubectl")
	require.Len(t, kubectlCalls, 1)
	assert.Contains(t, kubectlCalls[0].Stdin, "namespace: drasi-system")
}

func TestSyncSecrets_CustomNamespace(t *testing.T) {
	t.Parallel()

	runner := &mockCmdRunner{
		runFunc: func(ctx context.Context, stdin io.Reader, name string, args ...string) ([]byte, error) {
			if name == "az" {
				return []byte("value\n"), nil
			}
			return nil, nil
		},
	}

	err := syncSecrets(context.Background(), []config.SecretMapping{{
		VaultName:  "kv-main",
		SecretName: "db-password",
		K8sSecret:  "app-secret",
		K8sKey:     "password",
		Namespace:  "custom-ns",
	}}, nil, "", runner)
	require.NoError(t, err)

	kubectlCalls := filterCallsByName(runner.calls, "kubectl")
	require.Len(t, kubectlCalls, 1)
	assert.Contains(t, kubectlCalls[0].Stdin, "namespace: custom-ns")
}

func TestSyncSecrets_AzError_PropagatesError(t *testing.T) {
	t.Parallel()

	runner := &mockCmdRunner{
		runFunc: func(ctx context.Context, stdin io.Reader, name string, args ...string) ([]byte, error) {
			if name == "az" {
				return nil, fmt.Errorf("az failed")
			}
			return nil, nil
		},
	}

	err := syncSecrets(context.Background(), []config.SecretMapping{{
		VaultName:  "kv-main",
		SecretName: "db-password",
		K8sSecret:  "app-secret",
		K8sKey:     "password",
	}}, nil, "", runner)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "az failed")
	assert.Len(t, filterCallsByName(runner.calls, "kubectl"), 0)
}

func TestSyncSecrets_KubectlError_PropagatesError(t *testing.T) {
	t.Parallel()

	runner := &mockCmdRunner{
		runFunc: func(ctx context.Context, stdin io.Reader, name string, args ...string) ([]byte, error) {
			if name == "az" {
				return []byte("value\n"), nil
			}
			return nil, fmt.Errorf("kubectl failed")
		},
	}

	err := syncSecrets(context.Background(), []config.SecretMapping{{
		VaultName:  "kv-main",
		SecretName: "db-password",
		K8sSecret:  "app-secret",
		K8sKey:     "password",
	}}, nil, "", runner)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "kubectl failed")
}

func TestSyncSecrets_WithKubeContext(t *testing.T) {
	t.Parallel()

	runner := &mockCmdRunner{
		runFunc: func(ctx context.Context, stdin io.Reader, name string, args ...string) ([]byte, error) {
			if name == "az" {
				return []byte("value\n"), nil
			}
			return nil, nil
		},
	}

	err := syncSecrets(context.Background(), []config.SecretMapping{{
		VaultName:  "kv-main",
		SecretName: "db-password",
		K8sSecret:  "app-secret",
		K8sKey:     "password",
	}}, nil, "my-ctx", runner)
	require.NoError(t, err)

	kubectlCalls := filterCallsByName(runner.calls, "kubectl")
	require.Len(t, kubectlCalls, 1)
	assert.Equal(t, []string{"--context", "my-ctx", "apply", "-f", "-"}, kubectlCalls[0].Args)
}

func TestSyncSecrets_NoKubeContext(t *testing.T) {
	t.Parallel()

	runner := &mockCmdRunner{
		runFunc: func(ctx context.Context, stdin io.Reader, name string, args ...string) ([]byte, error) {
			if name == "az" {
				return []byte("value\n"), nil
			}
			return nil, nil
		},
	}

	err := syncSecrets(context.Background(), []config.SecretMapping{{
		VaultName:  "kv-main",
		SecretName: "db-password",
		K8sSecret:  "app-secret",
		K8sKey:     "password",
	}}, nil, "", runner)
	require.NoError(t, err)

	kubectlCalls := filterCallsByName(runner.calls, "kubectl")
	require.Len(t, kubectlCalls, 1)
	assert.Equal(t, []string{"apply", "-f", "-"}, kubectlCalls[0].Args)
}

func filterCallsByName(calls []mockCmdCall, name string) []mockCmdCall {
	filtered := make([]mockCmdCall, 0)
	for _, call := range calls {
		if call.Name == name {
			filtered = append(filtered, call)
		}
	}
	return filtered
}
