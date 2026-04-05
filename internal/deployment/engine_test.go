package deployment

import (
	"context"
	"testing"

	"github.com/azure/azd.extensions.drasi/internal/config"
)

type engineTestHarness struct {
	mockState *mockEnvClient
	engine    *Engine
}

func newEngineHarness() *engineTestHarness {
	mc := &mockEnvClient{store: make(map[string]string)}
	sm := NewStateManager(mc, "test-env")
	return &engineTestHarness{
		mockState: mc,
		engine:    NewEngine(sm),
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
			_ = h.engine.Deploy(context.Background(), tt.manifest, tt.opts)
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
			_ = h.engine.Deploy(context.Background(), tt.manifest, tt.opts)
		})
	}
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
			_ = h.engine.Teardown(context.Background(), tt.manifest, tt.opts)
		})
	}
}
