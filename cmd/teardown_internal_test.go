package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/scaffold"
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

func TestTeardownCommand_Force_WithValidManifest_ReachesDrasiVersionCheck(t *testing.T) {
	manifestPath := scaffoldTeardownTestProject(t)
	addr := startTestEnvironmentServer(t, &testEnvironmentService{
		getCurrentFunc: func(context.Context, *azdext.EmptyRequest) (*azdext.EnvironmentResponse, error) {
			return &azdext.EnvironmentResponse{Environment: &azdext.Environment{Name: "dev"}}, nil
		},
	})
	t.Setenv("AZD_SERVER", addr)
	installFakeCommands(t, map[string]string{
		"drasi": invalidVersionDrasiScript(),
	})

	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"teardown", "--force", "--config", manifestPath})

	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), output.ERR_DRASI_CLI_VERSION)
}

func TestTeardownCommand_ForceWithInfrastructure_ResolvesAzureResourceGroup(t *testing.T) {
	manifestPath := scaffoldTeardownTestProject(t)
	requestedKey := ""
	addr := startTestEnvironmentServer(t, &testEnvironmentService{
		getCurrentFunc: func(context.Context, *azdext.EmptyRequest) (*azdext.EnvironmentResponse, error) {
			return &azdext.EnvironmentResponse{Environment: &azdext.Environment{Name: "dev"}}, nil
		},
		getValueFunc: func(_ context.Context, req *azdext.GetEnvRequest) (*azdext.KeyValueResponse, error) {
			requestedKey = req.Key
			return &azdext.KeyValueResponse{}, nil
		},
	})
	t.Setenv("AZD_SERVER", addr)
	installFakeCommands(t, map[string]string{
		"drasi": successfulDrasiScript(),
	})

	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"teardown", "--force", "--infrastructure", "--config", manifestPath})

	err := root.Execute()

	require.Error(t, err)
	assert.Equal(t, "AZURE_RESOURCE_GROUP", requestedKey)
	assert.Contains(t, err.Error(), output.ERR_NO_AUTH)
	assert.Contains(t, err.Error(), "AZURE_RESOURCE_GROUP")
}

func TestTeardownCommand_JSONOutput_EmitsValidJSON(t *testing.T) {
	manifestPath := scaffoldTeardownTestProject(t)
	addr := startTestEnvironmentServer(t, &testEnvironmentService{
		getCurrentFunc: func(context.Context, *azdext.EmptyRequest) (*azdext.EnvironmentResponse, error) {
			return &azdext.EnvironmentResponse{Environment: &azdext.Environment{Name: "dev"}}, nil
		},
	})
	t.Setenv("AZD_SERVER", addr)
	installFakeCommands(t, map[string]string{
		"drasi": successfulDrasiScript(),
	})

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root := NewRootCommand()
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"--output", "json", "teardown", "--force", "--config", manifestPath})

	err := root.Execute()

	require.NoError(t, err)
	assert.Empty(t, stderr.String())

	var payload map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &payload))
	assert.Equal(t, "ok", payload["status"])
	assert.Equal(t, "dev", payload["environment"])
	assert.Equal(t, false, payload["infrastructure"])
}

func TestTeardownCommand_UserDeclinesConfirmation_PrintsAbortedMessage(t *testing.T) {
	saveConfirmVars(t)

	isTTYFunc = func() bool { return true }
	confirmFunc = func(_ string, result *bool) error {
		*result = false
		return nil
	}

	stdout := &bytes.Buffer{}
	root := NewRootCommand()
	root.SetOut(stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"teardown"})

	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Teardown aborted by user.")
}

func TestTeardownCommand_runAzGroupDelete_Succeeds(t *testing.T) {
	installFakeCommands(t, map[string]string{
		"az": fakeScript("exit /b 0", "exit 0"),
	})

	err := runAzGroupDelete(context.Background(), "rg-test")

	require.NoError(t, err)
}

func TestTeardownCommand_runAzGroupDelete_FailureWrapsCLIOutput(t *testing.T) {
	installFakeCommands(t, map[string]string{
		"az": fakeScript("echo delete failed 1>&2\r\nexit /b 1", "echo \"delete failed\" >&2\nexit 1"),
	})

	err := runAzGroupDelete(context.Background(), "rg-test")

	require.Error(t, err)
	assert.ErrorContains(t, err, "delete failed")
}

func TestTeardownCommand_runAzGroupDelete_AzNotOnPath_ReturnsError(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	err := runAzGroupDelete(context.Background(), "rg-test")

	require.Error(t, err)
	assert.ErrorContains(t, err, "executable file not found")
}

func scaffoldTeardownTestProject(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()
	_, err := scaffold.Scaffold("blank", tempDir, true)
	require.NoError(t, err)

	manifestPath := filepath.Join(tempDir, "drasi", "drasi.yaml")
	_, err = os.Stat(manifestPath)
	require.NoError(t, err)

	return manifestPath
}

func successfulDrasiScript() string {
	return fakeScript(`if "%1"=="version" (
echo v0.10.0
exit /b 0
)
exit /b 0`, `if [ "$1" = "version" ]; then
echo v0.10.0
exit 0
fi
exit 0`)
}

func invalidVersionDrasiScript() string {
	return fakeScript(`if "%1"=="version" (
echo definitely-not-semver
exit /b 0
)
exit /b 0`, `if [ "$1" = "version" ]; then
echo definitely-not-semver
exit 0
fi
exit 0`)
}
