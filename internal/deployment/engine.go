package deployment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/config"
)

// drasiRunner is the consumer-side interface for drasi CLI execution.
// *drasi.Client satisfies this interface implicitly.
type drasiRunner interface {
	CheckVersion(ctx context.Context) error
	RunCommand(ctx context.Context, args ...string) error
}

// DeployOptions configures a deploy run.
type DeployOptions struct {
	DryRun              bool
	Environment         string
	NoRollback          bool
	TotalTimeout        time.Duration
	PerComponentTimeout time.Duration
	// EnvVars holds azd environment values used to resolve $(VARNAME) patterns
	// in component YAML files before applying them to the cluster.
	EnvVars map[string]string
	// KubeContext is the kubectl context name for the target AKS cluster.
	// Used by secret sync to target the correct cluster.
	KubeContext string
}

// Engine orchestrates the full deploy lifecycle.
type Engine struct {
	state       *StateManager
	drasiClient drasiRunner
	runner      cmdRunner
}

// NewEngine creates an Engine. If runner is nil, a default execCmdRunner is used.
func NewEngine(state *StateManager, drasi drasiRunner, runner cmdRunner) *Engine {
	if runner == nil {
		runner = &execCmdRunner{}
	}
	return &Engine{state: state, drasiClient: drasi, runner: runner}
}

// Deploy applies a resolved manifest in dependency order:
// sources → queries → middleware → reactions.
func (e *Engine) Deploy(ctx context.Context, manifest *config.ResolvedManifest, opts DeployOptions) error {
	ctx, cancel := WithTotalDeployTimeout(ctx, opts.TotalTimeout)
	defer cancel()

	hashes := manifestToHashes(manifest)

	// Include env vars in the hash so changed values trigger a redeploy
	// even when the YAML template itself is unchanged.
	envSuffix := hashEnvVars(opts.EnvVars)
	for i := range hashes {
		hashes[i].Hash += envSuffix
	}

	existingState := make(map[string]string, len(hashes))
	for _, h := range hashes {
		val, err := e.state.ReadHash(ctx, h.StateKey())
		if err != nil {
			return fmt.Errorf("reading state for %s/%s: %w", h.Kind, h.ID, err)
		}
		existingState[h.StateKey()] = val
	}

	actions := SortForDeploy(Diff(hashes, existingState), manifest)
	appliedComponents := make([]ComponentAction, 0, len(actions))

	// Sync Key Vault secrets to Kubernetes Secrets before applying components.
	// Skip in dry-run mode since it would have cluster side effects.
	if !opts.DryRun {
		if err := syncSecrets(ctx, manifest.SecretMappings, opts.EnvVars, opts.KubeContext, e.runner); err != nil {
			return fmt.Errorf("syncing secrets from Key Vault to Kubernetes: %w", err)
		}
	}

	for _, action := range actions {
		if action.Action == ActionNoOp {
			continue
		}
		if opts.DryRun {
			continue
		}

		compCtx, compCancel := WithPerComponentTimeout(ctx, opts.PerComponentTimeout)

		var applyErr error
		if action.Action == ActionDeleteThenApply {
			applyErr = e.drasiClient.RunCommand(compCtx, "delete", action.Kind, action.ID)
		}

		if applyErr == nil {
			applyErr = e.applyComponent(compCtx, action, manifest, opts.EnvVars)
		}

		// drasi wait only supports source and reaction kinds.
		// Continuous queries and middleware become ready implicitly once their
		// dependencies are healthy; there is no separate readiness probe for them.
		if applyErr == nil && (action.Kind == "source" || action.Kind == "reaction") {
			applyErr = e.drasiClient.RunCommand(compCtx, "wait", action.Kind, action.ID, "--timeout", "300")
		}

		compCancel()

		if applyErr != nil {
			if !opts.NoRollback {
				for i := len(appliedComponents) - 1; i >= 0; i-- {
					comp := appliedComponents[i]
					if rbErr := e.drasiClient.RunCommand(ctx, "delete", comp.Kind, comp.ID); rbErr != nil {
						slog.WarnContext(ctx, "rollback delete failed",
							slog.String("kind", comp.Kind),
							slog.String("id", comp.ID),
							slog.Any("error", rbErr))
					}
				}
			}
			return applyErr
		}

		// Persist the new hash so subsequent runs can diff correctly.
		stateKey := config.ComponentHash{Kind: action.Kind, ID: action.ID}.StateKey()
		if writeErr := e.state.WriteHash(ctx, stateKey, action.Hash); writeErr != nil {
			return fmt.Errorf("writing state for %s/%s: %w", action.Kind, action.ID, writeErr)
		}

		appliedComponents = append(appliedComponents, action)
	}

	return nil
}

// Teardown deletes all components in reverse dependency order and clears persisted hashes.
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
func (e *Engine) applyComponent(ctx context.Context, action ComponentAction, manifest *config.ResolvedManifest, envVars map[string]string) error {
	raw, err := marshalComponent(action, manifest, envVars)
	if err != nil {
		return fmt.Errorf("marshalling %s/%s: %w", action.Kind, action.ID, err)
	}

	// Fail-fast: reject any remaining $(VARNAME) references that were not resolved.
	// Sending literal $(VARNAME) to the cluster will cause silent runtime failures.
	if unresolved := envVarPattern.FindAll(raw, -1); len(unresolved) > 0 {
		names := make([]string, len(unresolved))
		for i, m := range unresolved {
			names[i] = string(m)
		}
		return fmt.Errorf("unresolved environment variable references in %s/%s: %v — set these in azd environment state before deploying",
			action.Kind, action.ID, names)
	}

	tmpFile, err := os.CreateTemp("", fmt.Sprintf("drasi-%s-%s-*.yaml", action.Kind, action.ID))
	if err != nil {
		return fmt.Errorf("creating temp file for %s/%s: %w", action.Kind, action.ID, err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	// SECURITY: Restrict temp file to owner-only read/write. Temp files may contain
	// resolved configuration values; prevent other local users from reading them.
	if err := os.Chmod(tmpFile.Name(), 0600); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("securing temp file for %s/%s: %w", action.Kind, action.ID, err)
	}

	if _, err = tmpFile.Write(raw); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("writing temp file for %s/%s: %w", action.Kind, action.ID, err)
	}
	_ = tmpFile.Close()

	return e.drasiClient.RunCommand(ctx, "apply", "-f", tmpFile.Name())
}

// marshalComponent reads the original YAML file for the component from disk.
// We intentionally avoid re-marshalling the internal Go struct because yaml.Marshal
// serialises Value fields as nested objects (e.g. {value: "true"}) instead of the
// plain scalars that the drasi CLI expects (e.g. "true"). Reading the source file
// directly passes the original YAML unmodified to `drasi apply`, which is the only
// correct wire format for drasi 0.10.0.
//
// Fallback: if ManifestDir is empty (e.g. in unit tests that construct manifests
// directly without a loader), we fall back to yaml.Marshal of the in-memory struct.
// This fallback is only exercised by tests; production deployments always have
// ManifestDir populated by ResolveManifest.
func marshalComponent(action ComponentAction, manifest *config.ResolvedManifest, envVars map[string]string) ([]byte, error) {
	var relPath string
	var structVal any
	switch action.Kind {
	case "source":
		for _, s := range manifest.Sources {
			if s.ID == action.ID {
				relPath = s.FilePath
				structVal = s
			}
		}
	case "continuousquery":
		for _, q := range manifest.Queries {
			if q.ID == action.ID {
				relPath = q.FilePath
				structVal = q
			}
		}
	case "middleware":
		for _, m := range manifest.Middlewares {
			if m.ID == action.ID {
				relPath = m.FilePath
				structVal = m
			}
		}
	case "reaction":
		for _, r := range manifest.Reactions {
			if r.ID == action.ID {
				relPath = r.FilePath
				structVal = r
			}
		}
	}
	if structVal == nil {
		return nil, fmt.Errorf("component %s/%s not found in manifest", action.Kind, action.ID)
	}

	// Prefer reading the original file from disk to preserve exact YAML structure.
	if manifest.ManifestDir != "" && relPath != "" {
		absPath := filepath.Join(manifest.ManifestDir, filepath.FromSlash(relPath))
		data, err := os.ReadFile(absPath)
		if err != nil {
			return nil, fmt.Errorf("reading source file for %s/%s: %w", action.Kind, action.ID, err)
		}
		return expandEnvVars(data, envVars), nil
	}

	// Fallback for unit tests that build ResolvedManifest directly without a loader.
	return yaml.Marshal(structVal)
}

var envVarPattern = regexp.MustCompile(`\$\(([A-Za-z_][A-Za-z0-9_]*)\)`)

// expandEnvVars replaces $(VARNAME) patterns in data with values from envVars.
// Patterns whose variable name is not present in envVars are left unchanged.
func expandEnvVars(data []byte, envVars map[string]string) []byte {
	if len(envVars) == 0 {
		return data
	}
	return envVarPattern.ReplaceAllFunc(data, func(match []byte) []byte {
		key := envVarPattern.FindSubmatch(match)[1]
		if val, ok := envVars[string(key)]; ok {
			return []byte(val)
		}
		return match
	})
}

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
		raw = fmt.Appendf(nil, "%#v", v)
	}
	digest := sha256.Sum256(raw)
	return hex.EncodeToString(digest[:])
}

// hashEnvVars returns a deterministic hash suffix derived from the env var map.
// Returns an empty string when the map is empty so hashes are unchanged for
// deployments that do not use env var substitution.
func hashEnvVars(envVars map[string]string) string {
	if len(envVars) == 0 {
		return ""
	}
	// Sort keys for deterministic ordering.
	keys := make([]string, 0, len(envVars))
	for k := range envVars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		_, _ = h.Write([]byte(k))
		_, _ = h.Write([]byte("="))
		_, _ = h.Write([]byte(envVars[k]))
		_, _ = h.Write([]byte("\n"))
	}
	return ":" + hex.EncodeToString(h.Sum(nil))
}
