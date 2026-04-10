package deployment

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDrasiRunner is a test double for drasiRunner.
type mockDrasiRunner struct {
	checkVersionFunc func(ctx context.Context) error
	runCommandFunc   func(ctx context.Context, args ...string) error
	commandsCalled   [][]string
}

func (m *mockDrasiRunner) CheckVersion(ctx context.Context) error {
	if m.checkVersionFunc != nil {
		return m.checkVersionFunc(ctx)
	}
	return nil
}

func (m *mockDrasiRunner) RunCommand(ctx context.Context, args ...string) error {
	argsCopy := make([]string, len(args))
	copy(argsCopy, args)
	m.commandsCalled = append(m.commandsCalled, argsCopy)
	if m.runCommandFunc != nil {
		return m.runCommandFunc(ctx, args...)
	}
	return nil
}

type engineTestHarness struct {
	mockState *mockEnvClient
	mockDrasi *mockDrasiRunner
	engine    *Engine
}

func newEngineHarness() *engineTestHarness {
	mc := &mockEnvClient{store: make(map[string]string)}
	sm := NewStateManager(mc, "test-env")
	md := &mockDrasiRunner{}
	return &engineTestHarness{
		mockState: mc,
		mockDrasi: md,
		engine:    NewEngine(sm, md, nil),
	}
}

func TestEngine_Deploy_HappyPath_AllCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		manifest *config.ResolvedManifest
		opts     DeployOptions
	}{
		{
			name: "empty state with source and query deploys",
			manifest: &config.ResolvedManifest{
				Sources: []config.Source{{ID: "alerts-source"}},
				Queries: []config.ContinuousQuery{{ID: "severity-escalation"}},
			},
			opts: DeployOptions{Environment: "test-env"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := newEngineHarness()
			err := h.engine.Deploy(context.Background(), tt.manifest, tt.opts)
			require.NoError(t, err)

			// Both components should have been applied; source also waited.
			// apply source, wait source, apply query → 3 commands
			assert.Len(t, h.mockDrasi.commandsCalled, 3)
		})
	}
}

func TestEngine_Deploy_DryRun_NoSideEffects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		manifest *config.ResolvedManifest
		opts     DeployOptions
	}{
		{
			name: "dry run deploy",
			manifest: &config.ResolvedManifest{
				Sources: []config.Source{{ID: "alerts-source"}},
				Queries: []config.ContinuousQuery{{ID: "severity-escalation"}},
			},
			opts: DeployOptions{DryRun: true, Environment: "test-env"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := newEngineHarness()
			err := h.engine.Deploy(context.Background(), tt.manifest, tt.opts)
			require.NoError(t, err)

			// Dry run: no drasi commands called, no state written
			assert.Empty(t, h.mockDrasi.commandsCalled)
			assert.Empty(t, h.mockState.store)
		})
	}
}

func TestEngine_Deploy_NoOp_WhenHashUnchanged(t *testing.T) {
	t.Parallel()

	h := newEngineHarness()
	manifest := &config.ResolvedManifest{
		Sources: []config.Source{{ID: "alerts-source"}},
	}
	computed := manifestToHashes(manifest)
	require.Len(t, computed, 1)
	h.mockState.store[computed[0].StateKey()] = computed[0].Hash

	err := h.engine.Deploy(context.Background(), manifest, DeployOptions{Environment: "test-env"})
	require.NoError(t, err)

	// No-op: no drasi commands called
	assert.Empty(t, h.mockDrasi.commandsCalled)
}

func TestEngine_Deploy_EnvVarChange_TriggersRedeploy(t *testing.T) {
	t.Parallel()

	h := newEngineHarness()
	manifest := &config.ResolvedManifest{
		Sources: []config.Source{{ID: "alerts-source"}},
	}

	// First deploy with env vars A.
	envA := map[string]string{"VAULT": "kv-old"}
	err := h.engine.Deploy(context.Background(), manifest, DeployOptions{
		Environment: "test-env",
		EnvVars:     envA,
	})
	require.NoError(t, err)
	// apply + wait = 2 commands
	require.Len(t, h.mockDrasi.commandsCalled, 2)

	// Reset commands log.
	h.mockDrasi.commandsCalled = nil

	// Second deploy with same env vars → no-op.
	err = h.engine.Deploy(context.Background(), manifest, DeployOptions{
		Environment: "test-env",
		EnvVars:     envA,
	})
	require.NoError(t, err)
	assert.Empty(t, h.mockDrasi.commandsCalled)

	// Third deploy with changed env vars → triggers delete+apply+wait.
	envB := map[string]string{"VAULT": "kv-new"}
	err = h.engine.Deploy(context.Background(), manifest, DeployOptions{
		Environment: "test-env",
		EnvVars:     envB,
	})
	require.NoError(t, err)
	// delete + apply + wait = 3 commands
	require.Len(t, h.mockDrasi.commandsCalled, 3)
	assert.Equal(t, "delete", h.mockDrasi.commandsCalled[0][0])
	assert.Equal(t, "apply", h.mockDrasi.commandsCalled[1][0])
	assert.Equal(t, "wait", h.mockDrasi.commandsCalled[2][0])
}

func TestEngine_Deploy_DeleteThenApply_WhenHashChanged(t *testing.T) {
	t.Parallel()

	h := newEngineHarness()
	h.mockState.store["DRASI_HASH_SOURCE_alerts_source"] = "old-hash"

	manifest := &config.ResolvedManifest{
		Sources: []config.Source{{ID: "alerts-source"}},
	}
	err := h.engine.Deploy(context.Background(), manifest, DeployOptions{Environment: "test-env"})
	require.NoError(t, err)

	// delete + apply + wait → 3 commands
	assert.Len(t, h.mockDrasi.commandsCalled, 3)
	assert.Equal(t, "delete", h.mockDrasi.commandsCalled[0][0])
	assert.Equal(t, "apply", h.mockDrasi.commandsCalled[1][0])
	assert.Equal(t, "wait", h.mockDrasi.commandsCalled[2][0])
}

func TestEngine_Deploy_DeployOrder_SourcesBeforeQueries(t *testing.T) {
	t.Parallel()

	h := newEngineHarness()
	manifest := &config.ResolvedManifest{
		Queries: []config.ContinuousQuery{{ID: "my-query"}},
		Sources: []config.Source{{ID: "my-source"}},
	}
	err := h.engine.Deploy(context.Background(), manifest, DeployOptions{Environment: "test-env"})
	require.NoError(t, err)

	// 3 commands: apply source, wait source, apply query
	// drasi wait only supports source and reaction; queries are not waited on.
	require.Len(t, h.mockDrasi.commandsCalled, 3)

	// First command: apply source
	assert.Equal(t, "apply", h.mockDrasi.commandsCalled[0][0])
	// Second command: wait for source
	assert.Equal(t, "wait", h.mockDrasi.commandsCalled[1][0])
	assert.Equal(t, "source", h.mockDrasi.commandsCalled[1][1])
	// Third command: apply query (no wait — drasi wait does not support continuousquery)
	assert.Equal(t, "apply", h.mockDrasi.commandsCalled[2][0])
}

func TestEngine_Deploy_PropagatesRunCommandError(t *testing.T) {
	t.Parallel()

	h := newEngineHarness()
	h.mockDrasi.runCommandFunc = func(ctx context.Context, args ...string) error {
		return fmt.Errorf("drasi apply failed")
	}
	manifest := &config.ResolvedManifest{
		Sources: []config.Source{{ID: "alerts-source"}},
	}
	err := h.engine.Deploy(context.Background(), manifest, DeployOptions{Environment: "test-env"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "drasi apply failed")
}

func TestDeployRollbackOnFailure(t *testing.T) {
	t.Parallel()

	h := newEngineHarness()
	manifest := &config.ResolvedManifest{
		Sources:   []config.Source{{ID: "source-one"}},
		Queries:   []config.ContinuousQuery{{ID: "query-two"}},
		Reactions: []config.Reaction{{ID: "reaction-three"}},
	}

	// Fail on wait for the reaction (drasi wait supports source and reaction).
	// Queries and middleware are NOT waited on, so we trigger failure on reaction wait.
	h.mockDrasi.runCommandFunc = func(ctx context.Context, args ...string) error {
		if len(args) > 0 && args[0] == "wait" && len(args) > 2 && args[1] == "reaction" && args[2] == "reaction-three" {
			return fmt.Errorf("reaction wait failed")
		}
		return nil
	}

	err := h.engine.Deploy(context.Background(), manifest, DeployOptions{Environment: "test-env"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reaction wait failed")

	// Commands: apply source, wait source, apply query, apply reaction, wait reaction (fail),
	// then rollback: delete query, delete source → 7 total
	require.Len(t, h.mockDrasi.commandsCalled, 7)
	assert.Equal(t, []string{"wait", "source", "source-one", "--timeout", "300"}, h.mockDrasi.commandsCalled[1])
	assert.Equal(t, []string{"wait", "reaction", "reaction-three", "--timeout", "300"}, h.mockDrasi.commandsCalled[4])
	// Rollback deletes in reverse order: query first (last applied), then source
	assert.Equal(t, []string{"delete", "continuousquery", "query-two"}, h.mockDrasi.commandsCalled[5])
	assert.Equal(t, []string{"delete", "source", "source-one"}, h.mockDrasi.commandsCalled[6])
	// Reaction state should not be persisted since it failed
	assert.Equal(t, "", h.mockState.store["DRASI_HASH_REACTION_reaction_three"])
}

func TestDeployNoRollback(t *testing.T) {
	t.Parallel()

	h := newEngineHarness()
	manifest := &config.ResolvedManifest{
		Sources:   []config.Source{{ID: "source-one"}},
		Reactions: []config.Reaction{{ID: "reaction-two"}},
	}

	// Fail on wait for the reaction (drasi wait supports source and reaction).
	h.mockDrasi.runCommandFunc = func(ctx context.Context, args ...string) error {
		if len(args) > 0 && args[0] == "wait" && len(args) > 2 && args[1] == "reaction" && args[2] == "reaction-two" {
			return fmt.Errorf("reaction wait failed")
		}
		return nil
	}

	err := h.engine.Deploy(context.Background(), manifest, DeployOptions{Environment: "test-env", NoRollback: true})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reaction wait failed")

	// With NoRollback, no delete commands should be issued.
	for _, cmd := range h.mockDrasi.commandsCalled {
		assert.NotEqual(t, "delete", cmd[0], "no rollback deletes should occur with NoRollback: true")
	}
	// Commands: apply source, wait source, apply reaction, wait reaction (fail) → 4
	assert.Len(t, h.mockDrasi.commandsCalled, 4)
}

func TestDeployRollbackFailure(t *testing.T) {
	t.Parallel()

	h := newEngineHarness()
	manifest := &config.ResolvedManifest{
		Sources:   []config.Source{{ID: "source-one"}},
		Reactions: []config.Reaction{{ID: "reaction-two"}},
	}

	h.mockDrasi.runCommandFunc = func(ctx context.Context, args ...string) error {
		// Rollback delete of source-one fails
		if len(args) == 3 && args[0] == "delete" && args[1] == "source" && args[2] == "source-one" {
			return fmt.Errorf("rollback delete failed")
		}
		// Reaction wait fails, triggering rollback
		if len(args) > 0 && args[0] == "wait" && len(args) > 2 && args[1] == "reaction" && args[2] == "reaction-two" {
			return fmt.Errorf("reaction wait failed")
		}
		return nil
	}

	err := h.engine.Deploy(context.Background(), manifest, DeployOptions{Environment: "test-env"})
	require.Error(t, err)
	// The original deploy error is returned, not the rollback failure.
	assert.Contains(t, err.Error(), "reaction wait failed")
	assert.NotContains(t, err.Error(), "rollback delete failed")
	// Commands: apply source, wait source, apply reaction, wait reaction (fail),
	// rollback: delete source (fails) → 5 total
	require.Len(t, h.mockDrasi.commandsCalled, 5)
	assert.Equal(t, []string{"delete", "source", "source-one"}, h.mockDrasi.commandsCalled[4])
}

func TestEngine_Teardown_HappyPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		manifest *config.ResolvedManifest
		opts     DeployOptions
	}{
		{
			name: "teardown manifest",
			manifest: &config.ResolvedManifest{
				Sources:   []config.Source{{ID: "alerts-source"}},
				Queries:   []config.ContinuousQuery{{ID: "severity-escalation"}},
				Reactions: []config.Reaction{{ID: "alerts-http"}},
			},
			opts: DeployOptions{Environment: "test-env"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := newEngineHarness()
			err := h.engine.Teardown(context.Background(), tt.manifest, tt.opts)
			require.NoError(t, err)

			// 3 components → 3 delete commands
			assert.Len(t, h.mockDrasi.commandsCalled, 3)
			for _, cmd := range h.mockDrasi.commandsCalled {
				assert.Equal(t, "delete", cmd[0])
			}
		})
	}
}

func TestEngine_Teardown_ReverseOrder_ReactionsFirst(t *testing.T) {
	t.Parallel()

	h := newEngineHarness()
	manifest := &config.ResolvedManifest{
		Sources:   []config.Source{{ID: "my-source"}},
		Queries:   []config.ContinuousQuery{{ID: "my-query"}},
		Reactions: []config.Reaction{{ID: "my-reaction"}},
	}
	err := h.engine.Teardown(context.Background(), manifest, DeployOptions{Environment: "test-env"})
	require.NoError(t, err)

	require.Len(t, h.mockDrasi.commandsCalled, 3)
	// Reverse order: reaction → query → source
	assert.Equal(t, "reaction", h.mockDrasi.commandsCalled[0][1])
	assert.Equal(t, "continuousquery", h.mockDrasi.commandsCalled[1][1])
	assert.Equal(t, "source", h.mockDrasi.commandsCalled[2][1])
}

func TestEngine_Teardown_PropagatesError(t *testing.T) {
	t.Parallel()

	h := newEngineHarness()
	h.mockDrasi.runCommandFunc = func(ctx context.Context, args ...string) error {
		return fmt.Errorf("delete failed")
	}
	manifest := &config.ResolvedManifest{
		Sources: []config.Source{{ID: "alerts-source"}},
	}
	err := h.engine.Teardown(context.Background(), manifest, DeployOptions{Environment: "test-env"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete failed")
}

func TestEngine_Teardown_ClearsStateOnSuccess(t *testing.T) {
	t.Parallel()

	h := newEngineHarness()
	h.mockState.store["DRASI_HASH_SOURCE_my_source"] = "some-hash"

	manifest := &config.ResolvedManifest{
		Sources: []config.Source{{ID: "my-source"}},
	}
	err := h.engine.Teardown(context.Background(), manifest, DeployOptions{Environment: "test-env"})
	require.NoError(t, err)

	// State should be cleared
	assert.Equal(t, "", h.mockState.store["DRASI_HASH_SOURCE_my_source"])
}

func TestEngine_ApplyComponent_TempFileHasRestrictedPermissions(t *testing.T) {
	t.Parallel()

	h := newEngineHarness()

	// Capture the temp file path from the drasi apply call so we can stat it.
	var capturedPath string
	h.mockDrasi.runCommandFunc = func(ctx context.Context, args ...string) error {
		if len(args) >= 3 && args[0] == "apply" && args[1] == "-f" {
			capturedPath = args[2]
			// Stat the file while it still exists (before defer os.Remove runs).
			info, err := os.Stat(capturedPath)
			if err != nil {
				return fmt.Errorf("stat temp file: %w", err)
			}
			// On Unix-like systems, os.Chmod(0600) restricts to owner read/write.
			// On Windows, os.Chmod only affects the read-only bit; permission
			// masking is not supported, so we skip the permission assertion.
			if runtime.GOOS != "windows" {
				mode := info.Mode().Perm()
				if mode&0077 != 0 {
					return fmt.Errorf("temp file has insecure permissions %04o; expected 0600", mode)
				}
			} else {
				// Windows: just verify the file exists and is writable.
				_ = info.Mode()
			}
		}
		return nil
	}

	manifest := &config.ResolvedManifest{
		Sources: []config.Source{{ID: "perm-test-source"}},
	}
	err := h.engine.Deploy(context.Background(), manifest, DeployOptions{Environment: "test-env"})
	require.NoError(t, err)
	assert.NotEmpty(t, capturedPath, "drasi apply should have been called with a temp file path")
}

func TestMarshalComponent_ReadsOriginalFileForSource(t *testing.T) {
	t.Parallel()

	manifestDir := t.TempDir()
	relPath := filepath.Join("components", "source.yaml")
	absPath := filepath.Join(manifestDir, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(absPath), 0o755))

	want := []byte("apiVersion: v1\nkind: Source\nname: alerts\nspec:\n  kind: CosmosDB\n")
	require.NoError(t, os.WriteFile(absPath, want, 0o600))

	manifest := &config.ResolvedManifest{
		ManifestDir: manifestDir,
		Sources: []config.Source{{
			ID:       "alerts",
			FilePath: filepath.ToSlash(relPath),
			Spec:     config.SourceSpec{Kind: "ignored-by-file-read"},
		}},
	}

	got, err := marshalComponent(ComponentAction{Kind: "source", ID: "alerts"}, manifest, nil)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestMarshalComponent_FallbackMarshalForKindsWithoutManifestDir(t *testing.T) {
	t.Parallel()

	manifest := &config.ResolvedManifest{
		Queries: []config.ContinuousQuery{{
			APIVersion: "v1",
			Kind:       "ContinuousQuery",
			ID:         "alerts-query",
			Spec:       config.QuerySpec{Query: "MATCH (n) RETURN n"},
		}},
		Middlewares: []config.Middleware{{
			APIVersion: "v1",
			Kind:       "Middleware",
			ID:         "enricher",
			Spec:       config.MiddlewareSpec{Kind: "Identity"},
		}},
		Reactions: []config.Reaction{{
			APIVersion: "v1",
			Kind:       "Reaction",
			ID:         "sink",
			Spec:       config.ReactionSpec{Kind: "EventGrid"},
		}},
	}

	tests := []struct {
		name   string
		action ComponentAction
		want   string
	}{
		{name: "query marshals from struct", action: ComponentAction{Kind: "continuousquery", ID: "alerts-query"}, want: "name: alerts-query"},
		{name: "middleware marshals from struct", action: ComponentAction{Kind: "middleware", ID: "enricher"}, want: "name: enricher"},
		{name: "reaction marshals from struct", action: ComponentAction{Kind: "reaction", ID: "sink"}, want: "name: sink"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := marshalComponent(tt.action, manifest, nil)
			require.NoError(t, err)
			assert.Contains(t, string(got), tt.want)
		})
	}
}

func TestMarshalComponent_ComponentNotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	manifest := &config.ResolvedManifest{}

	got, err := marshalComponent(ComponentAction{Kind: "source", ID: "missing"}, manifest, nil)

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "component source/missing not found")
}

func TestEngine_ApplyComponent_MarshalError_ReturnsContext(t *testing.T) {
	t.Parallel()

	h := newEngineHarness()
	manifest := &config.ResolvedManifest{}

	err := h.engine.applyComponent(context.Background(), ComponentAction{Kind: "source", ID: "missing"}, manifest, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "marshalling source/missing")
	assert.Empty(t, h.mockDrasi.commandsCalled)
}

func TestEngine_ApplyComponent_RunCommandError_Propagates(t *testing.T) {
	t.Parallel()

	h := newEngineHarness()
	h.mockDrasi.runCommandFunc = func(ctx context.Context, args ...string) error {
		return fmt.Errorf("apply failed")
	}
	manifest := &config.ResolvedManifest{
		Sources: []config.Source{{
			APIVersion: "v1",
			Kind:       "Source",
			ID:         "alerts",
			Spec:       config.SourceSpec{Kind: "CosmosDB"},
		}},
	}

	err := h.engine.applyComponent(context.Background(), ComponentAction{Kind: "source", ID: "alerts"}, manifest, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "apply failed")
	require.Len(t, h.mockDrasi.commandsCalled, 1)
	assert.Equal(t, "apply", h.mockDrasi.commandsCalled[0][0])
	assert.Equal(t, "-f", h.mockDrasi.commandsCalled[0][1])
}

func TestExpandEnvVars_ReplacesKnownVars(t *testing.T) {
	t.Parallel()

	input := []byte(`vaultName: "$(AZURE_KEY_VAULT_NAME)"`)
	envVars := map[string]string{"AZURE_KEY_VAULT_NAME": "kv-drasi-abc123"}

	got := expandEnvVars(input, envVars)
	assert.Equal(t, `vaultName: "kv-drasi-abc123"`, string(got))
}

func TestExpandEnvVars_LeavesUnknownVarsUnchanged(t *testing.T) {
	t.Parallel()

	input := []byte(`vaultName: "$(UNKNOWN_VAR)"`)
	envVars := map[string]string{"AZURE_KEY_VAULT_NAME": "kv-drasi-abc123"}

	got := expandEnvVars(input, envVars)
	assert.Equal(t, `vaultName: "$(UNKNOWN_VAR)"`, string(got))
}

func TestExpandEnvVars_MultipleVarsInSameInput(t *testing.T) {
	t.Parallel()

	input := []byte(`vault: "$(VAULT)" db: "$(DATABASE)"`)
	envVars := map[string]string{"VAULT": "my-vault", "DATABASE": "my-db"}

	got := expandEnvVars(input, envVars)
	assert.Equal(t, `vault: "my-vault" db: "my-db"`, string(got))
}

func TestExpandEnvVars_EmptyEnvVars(t *testing.T) {
	t.Parallel()

	input := []byte(`vaultName: "$(AZURE_KEY_VAULT_NAME)"`)

	got := expandEnvVars(input, nil)
	assert.Equal(t, string(input), string(got))

	got = expandEnvVars(input, map[string]string{})
	assert.Equal(t, string(input), string(got))
}

func TestExpandEnvVars_NoVarsInInput(t *testing.T) {
	t.Parallel()

	input := []byte(`database: "drasidb"`)
	envVars := map[string]string{"AZURE_KEY_VAULT_NAME": "kv-drasi-abc123"}

	got := expandEnvVars(input, envVars)
	assert.Equal(t, `database: "drasidb"`, string(got))
}

func TestMarshalComponent_ExpandsEnvVarsFromFile(t *testing.T) {
	t.Parallel()

	manifestDir := t.TempDir()
	relPath := filepath.Join("sources", "my-source.yaml")
	absPath := filepath.Join(manifestDir, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(absPath), 0o755))

	content := []byte("vaultName: \"$(AZURE_KEY_VAULT_NAME)\"\ndb: \"$(DB_NAME)\"\n")
	require.NoError(t, os.WriteFile(absPath, content, 0o600))

	manifest := &config.ResolvedManifest{
		ManifestDir: manifestDir,
		Sources: []config.Source{{
			ID:       "my-source",
			FilePath: filepath.ToSlash(relPath),
			Spec:     config.SourceSpec{Kind: "CosmosGremlin"},
		}},
	}
	envVars := map[string]string{"AZURE_KEY_VAULT_NAME": "kv-test", "DB_NAME": "testdb"}

	got, err := marshalComponent(ComponentAction{Kind: "source", ID: "my-source"}, manifest, envVars)
	require.NoError(t, err)
	assert.Equal(t, "vaultName: \"kv-test\"\ndb: \"testdb\"\n", string(got))
}

func TestEngine_ApplyComponent_FailFast_UnresolvedEnvVars(t *testing.T) {
	t.Parallel()

	h := newEngineHarness()

	manifestDir := t.TempDir()
	relPath := filepath.Join("sources", "unresolved.yaml")
	absPath := filepath.Join(manifestDir, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(absPath), 0o755))

	content := []byte("vaultName: \"$(AZURE_KEY_VAULT_NAME)\"\ndb: \"$(MISSING_VAR)\"\n")
	require.NoError(t, os.WriteFile(absPath, content, 0o600))

	manifest := &config.ResolvedManifest{
		ManifestDir: manifestDir,
		Sources: []config.Source{{
			ID:       "unresolved-source",
			FilePath: filepath.ToSlash(relPath),
			Spec:     config.SourceSpec{Kind: "CosmosGremlin"},
		}},
	}
	// Only resolve one of the two vars — the other stays as $(MISSING_VAR).
	envVars := map[string]string{"AZURE_KEY_VAULT_NAME": "kv-test"}

	err := h.engine.applyComponent(
		context.Background(),
		ComponentAction{Kind: "source", ID: "unresolved-source"},
		manifest,
		envVars,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unresolved environment variable references")
	assert.Contains(t, err.Error(), "$(MISSING_VAR)")
	// No drasi commands should have been called.
	assert.Empty(t, h.mockDrasi.commandsCalled)
}

func TestEngine_ApplyComponent_AllVarsResolved_Succeeds(t *testing.T) {
	t.Parallel()

	h := newEngineHarness()

	manifestDir := t.TempDir()
	relPath := filepath.Join("sources", "resolved.yaml")
	absPath := filepath.Join(manifestDir, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(absPath), 0o755))

	content := []byte("vaultName: \"$(AZURE_KEY_VAULT_NAME)\"\ndb: \"$(DB_NAME)\"\n")
	require.NoError(t, os.WriteFile(absPath, content, 0o600))

	manifest := &config.ResolvedManifest{
		ManifestDir: manifestDir,
		Sources: []config.Source{{
			ID:       "resolved-source",
			FilePath: filepath.ToSlash(relPath),
			Spec:     config.SourceSpec{Kind: "CosmosGremlin"},
		}},
	}
	envVars := map[string]string{"AZURE_KEY_VAULT_NAME": "kv-test", "DB_NAME": "testdb"}

	err := h.engine.applyComponent(
		context.Background(),
		ComponentAction{Kind: "source", ID: "resolved-source"},
		manifest,
		envVars,
	)

	require.NoError(t, err)
	require.Len(t, h.mockDrasi.commandsCalled, 1)
	assert.Equal(t, "apply", h.mockDrasi.commandsCalled[0][0])
}
