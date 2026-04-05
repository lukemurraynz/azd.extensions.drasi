package deployment

import (
	"context"
	"fmt"
	"testing"

	"github.com/azure/azd.extensions.drasi/internal/config"
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
	mockState  *mockEnvClient
	mockDrasi  *mockDrasiRunner
	engine     *Engine
}

func newEngineHarness() *engineTestHarness {
	mc := &mockEnvClient{store: make(map[string]string)}
	sm := NewStateManager(mc, "test-env")
	md := &mockDrasiRunner{}
	return &engineTestHarness{
		mockState: mc,
		mockDrasi: md,
		engine:    NewEngine(sm, md),
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

			// Both components should have been applied + waited
			// apply source, wait source, apply query, wait query → 4 commands
			assert.Len(t, h.mockDrasi.commandsCalled, 4)
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
	// Pre-populate state with the same hash the engine would compute
	// manifestToHashes uses ID as hash placeholder, so hash == ID
	h.mockState.store["DRASI_HASH_SOURCE_alerts-source"] = "alerts-source"

	manifest := &config.ResolvedManifest{
		Sources: []config.Source{{ID: "alerts-source"}},
	}
	err := h.engine.Deploy(context.Background(), manifest, DeployOptions{Environment: "test-env"})
	require.NoError(t, err)

	// No-op: no drasi commands called
	assert.Empty(t, h.mockDrasi.commandsCalled)
}

func TestEngine_Deploy_DeleteThenApply_WhenHashChanged(t *testing.T) {
	t.Parallel()

	h := newEngineHarness()
	// Pre-populate state with a different hash
	h.mockState.store["DRASI_HASH_SOURCE_alerts-source"] = "old-hash"

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

	// 4 commands: apply source, wait source, apply query, wait query
	require.Len(t, h.mockDrasi.commandsCalled, 4)

	// First two commands are for source (apply -f <file>)
	assert.Equal(t, "apply", h.mockDrasi.commandsCalled[0][0])
	// Third command is wait for source
	assert.Equal(t, "wait", h.mockDrasi.commandsCalled[1][0])
	assert.Equal(t, "source", h.mockDrasi.commandsCalled[1][1])
	// Then apply query
	assert.Equal(t, "apply", h.mockDrasi.commandsCalled[2][0])
	// Wait for query
	assert.Equal(t, "wait", h.mockDrasi.commandsCalled[3][0])
	assert.Equal(t, "continuousquery", h.mockDrasi.commandsCalled[3][1])
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
	h.mockState.store["DRASI_HASH_SOURCE_my-source"] = "some-hash"

	manifest := &config.ResolvedManifest{
		Sources: []config.Source{{ID: "my-source"}},
	}
	err := h.engine.Teardown(context.Background(), manifest, DeployOptions{Environment: "test-env"})
	require.NoError(t, err)

	// State should be cleared
	assert.Equal(t, "", h.mockState.store["DRASI_HASH_SOURCE_my-source"])
}
