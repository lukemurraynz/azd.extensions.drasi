package deployment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/azure/azd.extensions.drasi/internal/config"
)

// drasiRunner is the consumer-side interface for drasi CLI execution.
// *drasi.Client satisfies this interface implicitly.
type drasiRunner interface {
	CheckVersion(ctx context.Context) error
	RunCommand(ctx context.Context, args ...string) error
}

// DeployOptions configures a deploy run.
type DeployOptions struct {
	DryRun      bool
	Environment string
}

// Engine orchestrates the full deploy lifecycle.
type Engine struct {
	state       *StateManager
	drasiClient drasiRunner
}

// NewEngine creates an Engine.
func NewEngine(state *StateManager, drasi drasiRunner) *Engine {
	return &Engine{state: state, drasiClient: drasi}
}

// Deploy applies a resolved manifest to the cluster in dependency order.
// Sources → queries → middleware → reactions. Skips no-op components.
// Writes state hashes on success. Dry-run mode computes the diff without running commands.
func (e *Engine) Deploy(ctx context.Context, manifest *config.ResolvedManifest, opts DeployOptions) error {
	ctx, cancel := WithTotalDeployTimeout(ctx)
	defer cancel()

	hashes := manifestToHashes(manifest)

	existingState := make(map[string]string, len(hashes))
	for _, h := range hashes {
		val, err := e.state.ReadHash(ctx, h.StateKey())
		if err != nil {
			return fmt.Errorf("reading state for %s/%s: %w", h.Kind, h.ID, err)
		}
		existingState[h.StateKey()] = val
	}

	actions := SortForDeploy(Diff(hashes, existingState), manifest)

	for _, action := range actions {
		if action.Action == ActionNoOp {
			continue
		}
		if opts.DryRun {
			continue
		}

		compCtx, compCancel := WithPerComponentTimeout(ctx)

		var applyErr error
		if action.Action == ActionDeleteThenApply {
			applyErr = e.drasiClient.RunCommand(compCtx, "delete", action.Kind, action.ID)
		}

		if applyErr == nil {
			applyErr = e.applyComponent(compCtx, action, manifest)
		}

		if applyErr == nil {
			applyErr = e.drasiClient.RunCommand(compCtx, "wait", action.Kind, action.ID, "--timeout", "300")
		}

		compCancel()

		if applyErr != nil {
			return applyErr
		}

		// Persist the new hash so subsequent runs can diff correctly.
		stateKey := config.ComponentHash{Kind: action.Kind, ID: action.ID}.StateKey()
		if writeErr := e.state.WriteHash(ctx, stateKey, action.Hash); writeErr != nil {
			return fmt.Errorf("writing state for %s/%s: %w", action.Kind, action.ID, writeErr)
		}
	}

	return nil
}

// Teardown deletes all components in reverse dependency order.
// Clears persisted hashes on success.
func (e *Engine) Teardown(ctx context.Context, manifest *config.ResolvedManifest, opts DeployOptions) error {
	hashes := manifestToHashes(manifest)

	actions := make([]ComponentAction, len(hashes))
	for i, h := range hashes {
		actions[i] = ComponentAction{Kind: h.Kind, ID: h.ID, Hash: h.Hash, Action: ActionDeleteThenApply}
	}
	ordered := SortForDelete(actions, manifest)

	for _, action := range ordered {
		if err := e.drasiClient.RunCommand(ctx, "delete", action.Kind, action.ID); err != nil {
			return err
		}
		stateKey := config.ComponentHash{Kind: action.Kind, ID: action.ID}.StateKey()
		if err := e.state.WriteHash(ctx, stateKey, ""); err != nil {
			return fmt.Errorf("clearing state for %s/%s: %w", action.Kind, action.ID, err)
		}
	}

	return nil
}

// applyComponent writes the component to a temp YAML file and calls drasi apply.
func (e *Engine) applyComponent(ctx context.Context, action ComponentAction, manifest *config.ResolvedManifest) error {
	raw, err := marshalComponent(action, manifest)
	if err != nil {
		return fmt.Errorf("marshalling %s/%s: %w", action.Kind, action.ID, err)
	}

	tmpFile, err := os.CreateTemp("", fmt.Sprintf("drasi-%s-%s-*.yaml", action.Kind, action.ID))
	if err != nil {
		return fmt.Errorf("creating temp file for %s/%s: %w", action.Kind, action.ID, err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err = tmpFile.Write(raw); err != nil {
		tmpFile.Close()
		return fmt.Errorf("writing temp file for %s/%s: %w", action.Kind, action.ID, err)
	}
	tmpFile.Close()

	return e.drasiClient.RunCommand(ctx, "apply", "-f", tmpFile.Name())
}

// marshalComponent serialises a single component from the manifest to YAML bytes.
func marshalComponent(action ComponentAction, manifest *config.ResolvedManifest) ([]byte, error) {
	switch action.Kind {
	case "source":
		for _, s := range manifest.Sources {
			if s.ID == action.ID {
				return yaml.Marshal(s)
			}
		}
	case "continuousquery":
		for _, q := range manifest.Queries {
			if q.ID == action.ID {
				return yaml.Marshal(q)
			}
		}
	case "middleware":
		for _, m := range manifest.Middlewares {
			if m.ID == action.ID {
				return yaml.Marshal(m)
			}
		}
	case "reaction":
		for _, r := range manifest.Reactions {
			if r.ID == action.ID {
				return yaml.Marshal(r)
			}
		}
	}
	return nil, fmt.Errorf("component %s/%s not found in manifest", action.Kind, action.ID)
}

// manifestToHashes converts a ResolvedManifest to a flat slice of ComponentHash values.
// Hashes are computed from canonical YAML content so unchanged components become no-ops.
func manifestToHashes(manifest *config.ResolvedManifest) []config.ComponentHash {
	var hashes []config.ComponentHash
	for _, s := range manifest.Sources {
		hashes = append(hashes, config.ComponentHash{Kind: "source", ID: s.ID, Hash: hashYAML(s)})
	}
	for _, q := range manifest.Queries {
		hashes = append(hashes, config.ComponentHash{Kind: "continuousquery", ID: q.ID, Hash: hashYAML(q)})
	}
	for _, m := range manifest.Middlewares {
		hashes = append(hashes, config.ComponentHash{Kind: "middleware", ID: m.ID, Hash: hashYAML(m)})
	}
	for _, r := range manifest.Reactions {
		hashes = append(hashes, config.ComponentHash{Kind: "reaction", ID: r.ID, Hash: hashYAML(r)})
	}
	return hashes
}

func hashYAML(v any) string {
	raw, err := yaml.Marshal(v)
	if err != nil {
		// yaml.Marshal should be deterministic for config model structs.
		raw = []byte(fmt.Sprintf("%#v", v))
	}
	digest := sha256.Sum256(raw)
	return hex.EncodeToString(digest[:])
}
