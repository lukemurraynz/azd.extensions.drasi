package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/scaffold"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type deployEnvSetCall struct {
	EnvName string
	Key     string
	Value   string
}

type deployTestEnvironmentState struct {
	mu              sync.Mutex
	currentEnv      string
	values          map[string]string
	setCalls        []deployEnvSetCall
	getCurrentCalls int
}

func newDeployTestEnvironmentState(envName string, initial map[string]string) *deployTestEnvironmentState {
	values := make(map[string]string, len(initial))
	maps.Copy(values, initial)

	return &deployTestEnvironmentState{
		currentEnv: envName,
		values:     values,
	}
}

func (s *deployTestEnvironmentState) service() *testEnvironmentService {
	return &testEnvironmentService{
		getCurrentFunc: func(_ context.Context, _ *azdext.EmptyRequest) (*azdext.EnvironmentResponse, error) {
			s.mu.Lock()
			defer s.mu.Unlock()

			s.getCurrentCalls++
			return &azdext.EnvironmentResponse{Environment: &azdext.Environment{Name: s.currentEnv}}, nil
		},
		getValueFunc: func(_ context.Context, req *azdext.GetEnvRequest) (*azdext.KeyValueResponse, error) {
			s.mu.Lock()
			defer s.mu.Unlock()

			return &azdext.KeyValueResponse{Value: s.values[req.Key]}, nil
		},
		setValueFunc: func(_ context.Context, req *azdext.SetEnvRequest) (*azdext.EmptyResponse, error) {
			s.mu.Lock()
			defer s.mu.Unlock()

			s.values[req.Key] = req.Value
			s.setCalls = append(s.setCalls, deployEnvSetCall{
				EnvName: req.EnvName,
				Key:     req.Key,
				Value:   req.Value,
			})

			return &azdext.EmptyResponse{}, nil
		},
	}
}

func (s *deployTestEnvironmentState) lockValues() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	values := make([]string, 0, len(s.setCalls))
	for _, call := range s.setCalls {
		if call.Key == "DRASI_DEPLOY_IN_PROGRESS" {
			values = append(values, call.Value)
		}
	}

	return values
}

func (s *deployTestEnvironmentState) currentLockValue() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.values["DRASI_DEPLOY_IN_PROGRESS"]
}

func (s *deployTestEnvironmentState) getCurrentCallCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.getCurrentCalls
}

func scaffoldDeployManifest(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	_, err := scaffold.Scaffold("blank", dir, true)
	require.NoError(t, err)

	return filepath.Join(dir, "drasi", "drasi.yaml")
}

func breakDeployQueryManifest(t *testing.T, configPath string) {
	t.Helper()

	queryPath := filepath.Join(filepath.Dir(configPath), "queries", "example-query.yaml")
	brokenQuery := `apiVersion: v1
kind: ContinuousQuery
name: example-query
spec:
  mode: query
  sources:
    subscriptions:
      - id: example-source
`
	require.NoError(t, os.WriteFile(queryPath, []byte(brokenQuery), 0o600))
}

func installFakeDrasiVersionCommand(t *testing.T) {
	t.Helper()

	installFakeCommands(t, map[string]string{
		"drasi": "echo Drasi CLI version: v0.10.2",
	})
}

func configureDeployTestServer(t *testing.T, state *deployTestEnvironmentState) {
	t.Helper()

	addr := startTestEnvironmentServer(t, state.service())
	t.Setenv("AZD_SERVER", addr)
}

func executeDeployCommand(t *testing.T, args ...string) (*bytes.Buffer, *bytes.Buffer, error) {
	t.Helper()

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs(args)

	return stdout, stderr, root.Execute()
}

func TestDeployCommand_InvalidTimeout_InternalValidationPath(t *testing.T) {
	stdout, stderr, err := executeDeployCommand(t, "deploy", "--timeout", "not-a-duration")

	require.Error(t, err)
	assert.Contains(t, err.Error(), output.ERR_VALIDATION_FAILED)
	assert.Contains(t, err.Error(), "invalid --timeout value")
	assert.Empty(t, stdout.String())
	assert.Contains(t, stderr.String(), output.ERR_VALIDATION_FAILED)
	assert.Contains(t, stderr.String(), "Use Go duration format")
}

func TestDeployCommand_EnvironmentResolvedFromServer_ProceedsPastAuth(t *testing.T) {
	state := newDeployTestEnvironmentState("dev", nil)
	configureDeployTestServer(t, state)

	missingConfig := filepath.Join(t.TempDir(), "missing-drasi.yaml")
	stdout, stderr, err := executeDeployCommand(t, "deploy", "--config", missingConfig)

	require.Error(t, err)
	assert.Contains(t, err.Error(), output.ERR_NO_MANIFEST)
	assert.NotContains(t, err.Error(), output.ERR_NO_AUTH)
	assert.Empty(t, stdout.String())
	assert.Contains(t, stderr.String(), output.ERR_NO_MANIFEST)
	assert.Equal(t, 1, state.getCurrentCallCount())
}

func TestDeployCommand_InvalidManifest_ReturnsValidationFailure(t *testing.T) {
	state := newDeployTestEnvironmentState("dev", nil)
	configureDeployTestServer(t, state)

	configPath := scaffoldDeployManifest(t)
	breakDeployQueryManifest(t, configPath)

	stdout, stderr, err := executeDeployCommand(t, "deploy", "--config", configPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), output.ERR_VALIDATION_FAILED)
	assert.Contains(t, stderr.String(), output.ERR_MISSING_QUERY_LANGUAGE)
	assert.Contains(t, stderr.String(), "missing required field 'query'")
	assert.Empty(t, stdout.String())
}

func TestDeployCommand_DryRun_Succeeds(t *testing.T) {
	installFakeDrasiVersionCommand(t)

	state := newDeployTestEnvironmentState("dev", nil)
	configureDeployTestServer(t, state)

	configPath := scaffoldDeployManifest(t)
	stdout, stderr, err := executeDeployCommand(t, "deploy", "--config", configPath, "--dry-run")

	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Dry-run succeeded for environment dev")
	assert.Contains(t, stderr.String(), "deploy on dev")
	assert.Contains(t, stderr.String(), "dry-run")
	assert.Equal(t, "", state.currentLockValue())
}

func TestDeployCommand_StaleDeployLock_ForceReleasesAndSucceeds(t *testing.T) {
	installFakeDrasiVersionCommand(t)

	state := newDeployTestEnvironmentState("dev", map[string]string{
		"DRASI_DEPLOY_IN_PROGRESS": "true",
	})
	configureDeployTestServer(t, state)

	configPath := scaffoldDeployManifest(t)
	stdout, _, err := executeDeployCommand(t, "deploy", "--config", configPath, "--dry-run")

	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Dry-run succeeded for environment dev")

	lockWrites := state.lockValues()
	require.Len(t, lockWrites, 3)
	assert.Equal(t, "", lockWrites[0], "stale lock must be force-released first")
	assert.Contains(t, lockWrites[1], "\"started\"", "acquire must write a JSON payload")
	assert.Equal(t, "", lockWrites[2], "lock must be released after deploy")
	assert.Equal(t, "", state.currentLockValue())
}

func TestDeployCommand_ActiveDeployLock_ReturnsInProgressError(t *testing.T) {
	installFakeDrasiVersionCommand(t)

	freshPayload, err := json.Marshal(map[string]any{
		"pid":     123,
		"started": time.Now().UTC().Format(time.RFC3339),
	})
	require.NoError(t, err)

	state := newDeployTestEnvironmentState("dev", map[string]string{
		"DRASI_DEPLOY_IN_PROGRESS": string(freshPayload),
	})
	configureDeployTestServer(t, state)

	configPath := scaffoldDeployManifest(t)
	stdout, stderr, execErr := executeDeployCommand(t, "deploy", "--config", configPath, "--dry-run")

	require.Error(t, execErr)
	assert.Contains(t, execErr.Error(), output.ERR_DEPLOY_IN_PROGRESS)
	assert.Empty(t, stdout.String())
	assert.Contains(t, stderr.String(), output.ERR_DEPLOY_IN_PROGRESS)
	assert.Empty(t, state.lockValues(), "fresh lock must not be force-released or reacquired")
}

func TestDeployCommand_JSONOutput_EmitsValidPayload(t *testing.T) {
	installFakeDrasiVersionCommand(t)

	state := newDeployTestEnvironmentState("dev", nil)
	configureDeployTestServer(t, state)

	configPath := scaffoldDeployManifest(t)
	stdout, stderr, err := executeDeployCommand(t, "--output", "json", "deploy", "--config", configPath, "--dry-run")

	require.NoError(t, err)
	assert.NotEmpty(t, stderr.String(), "audit event should be emitted to stderr")

	var payload map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &payload), "stdout must be valid JSON")
	assert.Equal(t, "ok", payload["status"])
	assert.Equal(t, "dev", payload["environment"])
	assert.Equal(t, true, payload["dryRun"])
}
