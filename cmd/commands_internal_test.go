package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetVersion_SetsExtensionVersion(t *testing.T) {
	originalVersion := extensionVersion
	t.Cleanup(func() {
		extensionVersion = originalVersion
	})

	SetVersion("1.2.3-test")

	assert.Equal(t, "1.2.3-test", extensionVersion)
}

func TestNewVersionCommand_CreatesVersionCommand(t *testing.T) {
	t.Parallel()

	outputFormat := "table"
	command := newVersionCommand(&outputFormat)

	require.NotNil(t, command)
	assert.Equal(t, "version", command.Use)
}

func TestNewMetadataCommand_CreatesMetadataCommand(t *testing.T) {
	t.Parallel()

	command := newMetadataCommand()

	require.NotNil(t, command)
	assert.Equal(t, "metadata", command.Use)
}

func TestNewInitCommand_DefinesExpectedFlags(t *testing.T) {
	t.Parallel()

	command := newInitCommand()

	require.NotNil(t, command)
	assert.Equal(t, "init", command.Use)
	require.NotNil(t, command.Flags().Lookup("template"))
	require.NotNil(t, command.Flags().Lookup("force"))
	require.NotNil(t, command.Flags().Lookup("output-dir"))
	require.NotNil(t, command.Flags().Lookup("environment"))
}

func TestNewInitCommand_ExecuteScaffoldsBlankTemplate(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()
	root := &cobra.Command{Use: "root"}
	root.PersistentFlags().String("output", "table", "Output format")

	command := newInitCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	command.SetOut(&stdout)
	root.AddCommand(command)
	root.SetArgs([]string{"init", "--template", "blank", "--output-dir", outputDir})

	err := root.Execute()
	require.NoError(t, err)

	assert.Contains(t, stdout.String(), "Initialized Drasi project")
	assert.FileExists(t, filepath.Join(outputDir, "azure.yaml"))
	assert.FileExists(t, filepath.Join(outputDir, "drasi", "drasi.yaml"))
	assert.FileExists(t, filepath.Join(outputDir, ".vscode", "launch.json"))
	assert.FileExists(t, filepath.Join(outputDir, "infra", "main.bicep"))
	assert.FileExists(t, filepath.Join(outputDir, "docker-compose.yml"))
}

func TestSyncWriter_WriteWritesToUnderlyingWriter(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	writer := &syncWriter{w: buffer}

	n, err := writer.Write([]byte("hello progress"))
	require.NoError(t, err)

	assert.Equal(t, len("hello progress"), n)
	assert.Equal(t, "hello progress", buffer.String())
}

func TestNewProgressHelper_JSONOutputReturnsNoopHelper(t *testing.T) {
	t.Parallel()

	root := NewRootCommand()
	child := &cobra.Command{Use: "test-child"}
	root.AddCommand(child)
	require.NoError(t, root.PersistentFlags().Set("output", "json"))

	helper, err := NewProgressHelper(child)
	require.NoError(t, err)
	require.NotNil(t, helper)

	assert.True(t, helper.noop)
	assert.NoError(t, helper.Start())
	helper.Message("ignored")
	assert.NoError(t, helper.Stop())
}

func TestProgressHelper_NoopMethodsDoNothing(t *testing.T) {
	t.Parallel()

	helper := &ProgressHelper{noop: true}

	assert.NoError(t, helper.Start())
	helper.Message("still ignored")
	assert.NoError(t, helper.Stop())
}

func TestResolveEnvironmentName_ExplicitValueSkipsGRPC(t *testing.T) {
	t.Parallel()

	called := false
	azdClient := newTestAzdClient(t, &testEnvironmentService{
		getCurrentFunc: func(context.Context, *azdext.EmptyRequest) (*azdext.EnvironmentResponse, error) {
			called = true
			return nil, assert.AnError
		},
	})

	command := &cobra.Command{Use: "test"}
	resolved, err := resolveEnvironmentName(context.Background(), command, azdClient, "explicit-env")
	require.NoError(t, err)

	assert.Equal(t, "explicit-env", resolved)
	assert.False(t, called)
}

func TestEnvServiceAdapter_SetValueCallsServer(t *testing.T) {
	t.Parallel()

	var received *azdext.SetEnvRequest
	azdClient := newTestAzdClient(t, &testEnvironmentService{
		setValueFunc: func(_ context.Context, req *azdext.SetEnvRequest) (*azdext.EmptyResponse, error) {
			received = req
			return &azdext.EmptyResponse{}, nil
		},
	})

	adapter := &azdEnvServiceAdapter{client: azdClient}
	err := adapter.SetValue(context.Background(), "dev", "EXAMPLE_KEY", "example-value")
	require.NoError(t, err)
	require.NotNil(t, received)

	assert.Equal(t, "dev", received.EnvName)
	assert.Equal(t, "EXAMPLE_KEY", received.Key)
	assert.Equal(t, "example-value", received.Value)
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
