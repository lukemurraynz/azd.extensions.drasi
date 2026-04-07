package cmd

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAzdEnvServiceAdapter_GetValue(t *testing.T) {
	t.Parallel()

	client := newTestAzdClient(t, &testEnvironmentService{
		getValueFunc: func(_ context.Context, req *azdext.GetEnvRequest) (*azdext.KeyValueResponse, error) {
			assert.Equal(t, "dev", req.EnvName)
			assert.Equal(t, "AZURE_AKS_CONTEXT", req.Key)
			return &azdext.KeyValueResponse{Value: "aks-dev"}, nil
		},
	})

	adapter := &azdEnvServiceAdapter{client: client}
	value, err := adapter.GetValue(context.Background(), "dev", "AZURE_AKS_CONTEXT")

	require.NoError(t, err)
	assert.Equal(t, "aks-dev", value)
}

func TestAzdEnvServiceAdapter_SetValue(t *testing.T) {
	t.Parallel()

	client := newTestAzdClient(t, &testEnvironmentService{
		setValueFunc: func(_ context.Context, req *azdext.SetEnvRequest) (*azdext.EmptyResponse, error) {
			assert.Equal(t, "dev", req.EnvName)
			assert.Equal(t, "DRASI_PROVISIONED", req.Key)
			assert.Equal(t, "true", req.Value)
			return &azdext.EmptyResponse{}, nil
		},
	})

	adapter := &azdEnvServiceAdapter{client: client}
	require.NoError(t, adapter.SetValue(context.Background(), "dev", "DRASI_PROVISIONED", "true"))
}

func TestResolveEnvironmentName(t *testing.T) {
	t.Parallel()

	t.Run("explicit returns directly", func(t *testing.T) {
		t.Parallel()

		root := NewRootCommand()
		cmd, _, err := root.Find([]string{"deploy"})
		require.NoError(t, err)

		resolved, err := resolveEnvironmentName(context.Background(), cmd, nil, "explicit-env")
		require.NoError(t, err)
		assert.Equal(t, "explicit-env", resolved)
	})

	t.Run("root flag fallback", func(t *testing.T) {
		t.Parallel()

		root := NewRootCommand()
		require.NoError(t, root.PersistentFlags().Set("environment", "root-env"))
		cmd, _, err := root.Find([]string{"deploy"})
		require.NoError(t, err)

		resolved, err := resolveEnvironmentName(context.Background(), cmd, nil, "")
		require.NoError(t, err)
		assert.Equal(t, "root-env", resolved)
	})

	t.Run("current environment from azd", func(t *testing.T) {
		t.Parallel()

		client := newTestAzdClient(t, &testEnvironmentService{
			getCurrentFunc: func(_ context.Context, _ *azdext.EmptyRequest) (*azdext.EnvironmentResponse, error) {
				return &azdext.EnvironmentResponse{Environment: &azdext.Environment{Name: "current-env"}}, nil
			},
		})

		root := NewRootCommand()
		cmd, _, err := root.Find([]string{"deploy"})
		require.NoError(t, err)

		resolved, err := resolveEnvironmentName(context.Background(), cmd, client, "")
		require.NoError(t, err)
		assert.Equal(t, "current-env", resolved)
	})

	t.Run("missing current environment returns error", func(t *testing.T) {
		t.Parallel()

		client := newTestAzdClient(t, &testEnvironmentService{})

		root := NewRootCommand()
		cmd, _, err := root.Find([]string{"deploy"})
		require.NoError(t, err)

		resolved, err := resolveEnvironmentName(context.Background(), cmd, client, "")
		require.Error(t, err)
		assert.Empty(t, resolved)
		assert.Contains(t, err.Error(), "current azd environment is not set")
	})
}

func TestDeployCommand_EnvironmentSelectionBypassesGetCurrent(t *testing.T) {
	missingConfig := filepath.Join(t.TempDir(), "missing-drasi.yaml")
	serverAddr := startTestEnvironmentServer(t, &testEnvironmentService{
		getCurrentFunc: func(_ context.Context, _ *azdext.EmptyRequest) (*azdext.EnvironmentResponse, error) {
			return nil, errors.New("GetCurrent must not be called")
		},
	})
	t.Setenv("AZD_SERVER", serverAddr)

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "command environment flag",
			args: []string{"deploy", "--environment", "explicit-env", "--config", missingConfig},
		},
		{
			name: "root environment flag",
			args: []string{"--environment", "root-env", "deploy", "--config", missingConfig},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := NewRootCommand()
			root.SetOut(new(strings.Builder))
			root.SetErr(new(strings.Builder))
			root.SetArgs(tt.args)

			err := root.Execute()
			require.Error(t, err)
			assert.Contains(t, err.Error(), output.ERR_NO_MANIFEST)
			assert.NotContains(t, err.Error(), output.ERR_NO_AUTH)
		})
	}
}
