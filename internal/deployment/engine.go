package deployment

import (
	"context"

	"github.com/azure/azd.extensions.drasi/internal/config"
)

// DeployOptions configures a deploy run.
type DeployOptions struct {
	DryRun      bool
	Environment string
}

// Engine orchestrates the full deploy lifecycle.
type Engine struct {
	state *StateManager
	// drasiClient will be injected during implementation; stub has nil
}

// NewEngine creates an Engine.
func NewEngine(state *StateManager) *Engine {
	return &Engine{state: state}
}

// Deploy applies a resolved manifest to the cluster in dependency order.
func (e *Engine) Deploy(ctx context.Context, manifest *config.ResolvedManifest, opts DeployOptions) error {
	panic("not implemented")
}

// Teardown deletes all components in reverse dependency order.
func (e *Engine) Teardown(ctx context.Context, manifest *config.ResolvedManifest, opts DeployOptions) error {
	panic("not implemented")
}
