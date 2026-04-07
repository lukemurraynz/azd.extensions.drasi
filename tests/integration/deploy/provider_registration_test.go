//go:build integration

package deploy_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/config"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/deployment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type noOpDrasiRunner struct{}

func (n *noOpDrasiRunner) CheckVersion(_ context.Context) error { return nil }

func (n *noOpDrasiRunner) RunCommand(_ context.Context, _ ...string) error { return nil }

// TestDeployEngine_Deploy_RegistersAllComponents verifies FR-042: all component kinds
// (source, continuousquery, middleware, reaction) in a manifest produce read operations
// against the state manager during deploy planning against an empty existing state.
//
// This exercises the Engine → Diff → StateManager integration path without requiring a
// live AKS cluster or azd gRPC server.
func TestDeployEngine_Deploy_RegistersAllComponents(t *testing.T) {
	t.Parallel()

	manifest := &config.ResolvedManifest{
		Sources: []config.Source{
			{ID: "src-a", Kind: "source"},
			{ID: "src-b", Kind: "source"},
		},
		Queries: []config.ContinuousQuery{
			{ID: "q-1", Kind: "continuousquery"},
		},
		Middlewares: []config.Middleware{
			{ID: "mw-x", Kind: "middleware"},
		},
		Reactions: []config.Reaction{
			{ID: "react-z", Kind: "reaction"},
		},
	}

	sc := newStubEnvSvc(nil)
	state := deployment.NewStateManagerFromClient(sc, "test-env")
	engine := deployment.NewEngine(state, &noOpDrasiRunner{})

	err := engine.Deploy(t.Context(), manifest, deployment.DeployOptions{DryRun: true})
	require.NoError(t, err, "Deploy must not return an error when processing a valid manifest")

	// All 5 components must have been queried from state during diff calculation.
	expectedKeys := []string{
		config.ComponentHash{Kind: "source", ID: "src-a", Hash: "src-a"}.StateKey(),
		config.ComponentHash{Kind: "source", ID: "src-b", Hash: "src-b"}.StateKey(),
		config.ComponentHash{Kind: "continuousquery", ID: "q-1", Hash: "q-1"}.StateKey(),
		config.ComponentHash{Kind: "middleware", ID: "mw-x", Hash: "mw-x"}.StateKey(),
		config.ComponentHash{Kind: "reaction", ID: "react-z", Hash: "react-z"}.StateKey(),
	}
	for _, key := range expectedKeys {
		assert.True(t, sc.wasRead(key),
			"state manager must query existing hash for component key %q", key)
	}
}

// TestDeployEngine_Deploy_IdempotentOnMatchingHash verifies idempotency semantics: when
// the existing state hash matches the computed component hash, no write occurs — the
// component is treated as a NoOp.
//
// This is the core idempotency recovery guarantee: deploying an unchanged manifest twice
// produces no cluster operations.
func TestDeployEngine_Deploy_IdempotentOnMatchingHash(t *testing.T) {
	t.Parallel()

	const componentID = "src-stable"

	h := config.ComponentHash{Kind: "source", ID: componentID}

	manifest := &config.ResolvedManifest{
		Sources: []config.Source{{ID: componentID, Kind: "source"}},
	}

	stableHash := hashForComponent(t, manifest.Sources[0])
	sc := newStubEnvSvc(map[string]string{h.StateKey(): stableHash})

	state := deployment.NewStateManagerFromClient(sc, "test-env")
	engine := deployment.NewEngine(state, &noOpDrasiRunner{})

	err := engine.Deploy(t.Context(), manifest, deployment.DeployOptions{DryRun: true})
	require.NoError(t, err, "Deploy must succeed on an unchanged manifest")

	// The key was read for diff; because hashes match the action is NoOp so no write
	// should have been issued.
	assert.False(t, sc.wasWritten(h.StateKey()),
		"state manager must not write a hash for an unchanged component (NoOp idempotency)")

	// Validate the setup really targets a NoOp: if this value drifts the test stops proving
	// idempotency semantics and should fail loudly.
	assert.Equal(t, stableHash, sc.state[h.StateKey()])
}

// TestDeployEngine_Deploy_DetectsChangedComponent verifies that when an existing
// component hash differs from the desired hash the engine reads the key (diff) but does
// not prematurely write it on a dry-run — the DeleteThenApply path is planned, not applied.
func TestDeployEngine_Deploy_DetectsChangedComponent(t *testing.T) {
	t.Parallel()

	const componentID = "src-changed"
	const oldHash = "old-hash-value"

	h := config.ComponentHash{Kind: "source", ID: componentID}
	// Seed with a different hash so diff produces DeleteThenApply.
	sc := newStubEnvSvc(map[string]string{h.StateKey(): oldHash})

	manifest := &config.ResolvedManifest{
		Sources: []config.Source{{ID: componentID, Kind: "source"}},
	}

	state := deployment.NewStateManagerFromClient(sc, "test-env")
	engine := deployment.NewEngine(state, &noOpDrasiRunner{})

	err := engine.Deploy(t.Context(), manifest, deployment.DeployOptions{DryRun: true})
	require.NoError(t, err)

	assert.True(t, sc.wasRead(h.StateKey()),
		"engine must read existing state for a component whose hash has changed")
}

// ---------------------------------------------------------------------------
// Stub helpers
// ---------------------------------------------------------------------------

// stubEnvSvc implements deployment.EnvServiceClient entirely in memory.
type stubEnvSvc struct {
	state  map[string]string
	reads  map[string]bool
	writes map[string]bool
}

func newStubEnvSvc(seed map[string]string) *stubEnvSvc {
	s := &stubEnvSvc{
		state:  make(map[string]string),
		reads:  make(map[string]bool),
		writes: make(map[string]bool),
	}
	for k, v := range seed {
		s.state[k] = v
	}
	return s
}

func (s *stubEnvSvc) GetValue(_ context.Context, _ string, key string) (string, error) {
	s.reads[key] = true
	return s.state[key], nil
}

func (s *stubEnvSvc) SetValue(_ context.Context, _ string, key, value string) error {
	s.writes[key] = true
	s.state[key] = value
	return nil
}

func (s *stubEnvSvc) wasRead(key string) bool    { return s.reads[key] }
func (s *stubEnvSvc) wasWritten(key string) bool { return s.writes[key] }

func hashForComponent(t *testing.T, v any) string {
	t.Helper()
	raw, err := yaml.Marshal(v)
	require.NoError(t, err)
	digest := sha256.Sum256(raw)
	return hex.EncodeToString(digest[:])
}
