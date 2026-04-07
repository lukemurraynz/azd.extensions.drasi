# Data Model: azd-drasi Extension

**Phase**: 1 — Design  
**Date**: 2026-04-04  
**Feature**: 001-azd-drasi-extension

---

## Entity Overview

Eight domain entities drive the extension. The table below maps each to its YAML file location,
Go struct path, and key spec references.

| Entity             | YAML Location                             | Go Struct                  | Spec FR(s)     |
| ------------------ | ----------------------------------------- | -------------------------- | -------------- |
| `DrasiManifest`    | `drasi/drasi.yaml`                        | `internal/config/model.go` | FR-011, FR-012 |
| `Source`           | `drasi/sources/*.yaml`                    | `internal/config/model.go` | FR-013         |
| `ContinuousQuery`  | `drasi/queries/*.yaml`                    | `internal/config/model.go` | FR-014         |
| `Reaction`         | `drasi/reactions/*.yaml`                  | `internal/config/model.go` | FR-015         |
| `Middleware`       | `drasi/middleware/*.yaml`                 | `internal/config/model.go` | FR-016         |
| `Environment`      | `drasi/environments/<name>.yaml`          | `internal/config/model.go` | FR-017         |
| `SourceProvider`   | Runtime-only (registered by `drasi init`) | —                          | FR-025         |
| `ReactionProvider` | Runtime-only (registered by `drasi init`) | —                          | FR-025         |

---

## Go Type Definitions

### `internal/config/model.go`

```go
// DrasiManifest is the root configuration file: drasi/drasi.yaml
type DrasiManifest struct {
    APIVersion   string            `yaml:"apiVersion"`     // Must be "v1"
    Includes     []IncludeSpec     `yaml:"includes"`       // Glob patterns to source/query/reaction YAMLs
    Environments map[string]string `yaml:"environments"`   // name → overlay file path
    FeatureFlags map[string]bool   `yaml:"featureFlags"`   // name → enabled
}

type IncludeSpec struct {
    Kind    string `yaml:"kind"`    // "sources" | "queries" | "reactions" | "middleware"
    Pattern string `yaml:"pattern"` // Glob pattern relative to drasi/
}

// Source represents a binding to an external data system.
type Source struct {
    APIVersion string            `yaml:"apiVersion"`  // "v1"
    Kind       string            `yaml:"kind"`        // "Source"
    ID         string            `yaml:"id"`
    SourceKind string            `yaml:"sourceKind"`  // "cosmosGremlin" | "eventHub" | "postgresql" | "sqlserver"
    Properties map[string]Value  `yaml:"properties"`  // Connection details; values may be SecretRef
}

// ContinuousQuery represents a declarative Cypher or GQL query.
type ContinuousQuery struct {
    APIVersion    string      `yaml:"apiVersion"` // "v1"
    Kind          string      `yaml:"kind"`       // "ContinuousQuery"
    ID            string      `yaml:"id"`
    QueryLanguage string      `yaml:"queryLanguage"` // MUST be "Cypher" or "GQL" — never defaulted
    Sources       []SourceRef `yaml:"sources"`
    Query         string      `yaml:"query"`
    Joins         []JoinSpec  `yaml:"joins,omitempty"`
    Reactions     []string    `yaml:"reactions"` // Reaction IDs
    AutoStart     bool        `yaml:"autoStart"`
}

type SourceRef struct {
    ID     string   `yaml:"id"`
    Nodes  []string `yaml:"nodes,omitempty"`  // Optional node/label filters
}

// JoinSpec defines an explicit cross-source join (implicit Cartesian joins are NOT supported).
type JoinSpec struct {
    ID   string    `yaml:"id"`
    Keys []JoinKey `yaml:"keys"`
}

type JoinKey struct {
    Label    string `yaml:"label"`
    Property string `yaml:"property"`
}

// Reaction represents an action triggered on ContinuousQuery result changes.
type Reaction struct {
    APIVersion   string           `yaml:"apiVersion"` // "v1"
    Kind         string           `yaml:"kind"`       // "Reaction"
    ID           string           `yaml:"id"`
    ReactionKind string           `yaml:"reactionKind"` // "dapr-pubsub" | "http" | "signalr" | "eventgrid" | "storedproc" | "debug"
    Config       map[string]Value `yaml:"config"`
}

// Middleware is an optional enrichment component between query and reaction.
type Middleware struct {
    APIVersion    string           `yaml:"apiVersion"` // "v1"
    Kind          string           `yaml:"kind"`       // "Middleware"
    ID            string           `yaml:"id"`
    MiddlewareKind string          `yaml:"middlewareKind"`
    Config        map[string]Value `yaml:"config"`
}

// Environment holds parameter overrides for a named deployment context.
type Environment struct {
    APIVersion     string           `yaml:"apiVersion"` // "v1"
    Parameters     map[string]Value `yaml:"parameters"`
    Scaling        map[string]Value `yaml:"scaling,omitempty"`
    FeatureFlagOverrides map[string]bool `yaml:"featureFlags,omitempty"`
}

// Value is a union type: either a plain string or a SecretRef.
// Only one of StringValue or SecretRef is populated.
type Value struct {
    StringValue string     // Set when the value is a plain string
    SecretRef   *SecretRef // Set when the value is a Key Vault reference
    EnvRef      *EnvRef    // Set when the value is an environment variable binding
}

// SecretRef represents a Key Vault secret reference.
// YAML form: { kind: secret, vaultName: <name>, secretName: <key> }
type SecretRef struct {
    VaultName  string `yaml:"vaultName"`
    SecretName string `yaml:"secretName"`
}

// EnvRef represents an environment variable binding.
type EnvRef struct {
    Name string `yaml:"name"`
}
```

---

## Resolved Model

After loading and merging all files for a given environment, the extension produces a
`ResolvedManifest`. This is the canonical input to both the validator and the deployment engine.

```go
// internal/config/resolved.go

// ResolvedManifest is the fully merged, environment-overridden model.
// Given the same inputs (files + environment name), the output MUST be identical across runs.
type ResolvedManifest struct {
    Sources     []Source
    Queries     []ContinuousQuery
    Reactions   []Reaction
    Middlewares []Middleware
    Environment Environment
    FeatureFlags map[string]bool
}
```

---

## State Model

The deployment engine uses content hashes stored in the azd environment state file.

```go
// internal/deployment/state.go

// ComponentHash stores a SHA-256 hash of a component's resolved YAML.
type ComponentHash struct {
    Kind string // e.g. "SOURCE", "CONTINUOUSQUERY", "REACTION", "MIDDLEWARE"
    ID   string // component ID
    Hash string // hex-encoded SHA-256
}

// StateKey returns the azd environment state key: DRASI_HASH_<KIND>_<ID>
func (h ComponentHash) StateKey() string {
    return fmt.Sprintf("DRASI_HASH_%s_%s", strings.ToUpper(h.Kind), strings.ToUpper(h.ID))
}
```

---

## Validation Error Model

```go
// internal/validation/errors.go

type ValidationLevel string
const (
    LevelError   ValidationLevel = "error"
    LevelWarning ValidationLevel = "warning"
)

// ValidationIssue represents a single validation finding.
type ValidationIssue struct {
    Level    ValidationLevel
    File     string  // Absolute path to the offending YAML file
    Line     int     // 1-based line number (0 if unknown)
    Code     string  // e.g. "ERR_MISSING_REFERENCE", "ERR_CIRCULAR_DEPENDENCY"
    Message  string  // Human-readable description
    Remediation string // Suggested fix or doc link
}

// ValidationResult accumulates all issues from a single validation pass.
type ValidationResult struct {
    Issues []ValidationIssue
}

func (r *ValidationResult) HasErrors() bool {
    for _, i := range r.Issues {
        if i.Level == LevelError { return true }
    }
    return false
}
```

---

## Deployment Model

```go
// internal/deployment/engine.go

// ComponentAction describes what the deployment engine will do to a component.
type ComponentAction string
const (
    ActionCreate          ComponentAction = "create"
    ActionDeleteThenApply ComponentAction = "delete-then-apply"
    ActionNoOp            ComponentAction = "no-op"
)

// DeploymentPlan is the pre-computed set of actions for a deployment run.
// Built during --dry-run and during actual deploy.
type DeploymentPlan struct {
    Actions []ComponentDeployAction
    // Actions are ordered: sources → queries → reactions → middleware
    // Deletion within delete-then-apply follows reverse order.
}

type ComponentDeployAction struct {
    Kind    string          // "Source" | "ContinuousQuery" | "Reaction" | "Middleware"
    ID      string
    Action  ComponentAction
    Reason  string          // Human-readable reason for the action choice
}

// DeploymentResult captures the outcome of an actual deployment run.
type DeploymentResult struct {
    Succeeded []ComponentDeployAction
    Failed    []ComponentFailure
    Duration  time.Duration
}

type ComponentFailure struct {
    ComponentDeployAction
    ErrorCode string
    Message   string
}
```

---

## Infrastructure Parameter Model

These Bicep parameters are generated by `azd drasi provision` and correspond to the `main.parameters.bicepparam` file.

| Parameter                   | Type   | Default                        | Required                       |
| --------------------------- | ------ | ------------------------------ | ------------------------------ |
| `location`                  | string | `resourceGroup().location`     | yes                            |
| `environmentName`           | string | —                              | yes                            |
| `aksClusterName`            | string | `drasi-${environmentName}`     | yes                            |
| `drasiNamespace`            | string | `drasi-system`                 | no                             |
| `logAnalyticsWorkspaceName` | string | `drasi-law-${environmentName}` | yes                            |
| `keyVaultName`              | string | `drasi-kv-${environmentName}`  | yes                            |
| `uamiName`                  | string | `drasi-id-${environmentName}`  | yes                            |
| `usePrivateAcr`             | bool   | `false`                        | no                             |
| `acrName`                   | string | `""`                           | conditional on `usePrivateAcr` |
| `enableCosmosDb`            | bool   | `false`                        | no                             |
| `enableEventHub`            | bool   | `false`                        | no                             |
| `enableServiceBus`          | bool   | `false`                        | no                             |

---

## Entity Lifecycle State Machine

```text
Source / ContinuousQuery / Reaction (runtime state via Drasi API):

  ┌────────┐  drasi apply   ┌─────────────┐  Drasi runtime OK  ┌────────┐
  │  None  │ ────────────► │  Pending    │ ──────────────────► │ Online │
  └────────┘               └─────────────┘                     └────────┘
                                  │                                  │
                            timeout/error                      config change
                                  ▼                                  ▼
                           ┌─────────────┐               ┌──────────────────┐
                           │TerminalError│               │ delete-then-apply │
                           └─────────────┘               └──────────────────┘

Extension wait semantics: poll drasi describe until Online OR timeout (5 min/component, 15 min total)
```

---

## Key Constraints

- All entity `ID` fields: kebab-case, max 63 chars (Kubernetes name constraint), unique within kind
- `ContinuousQuery.queryLanguage`: MUST be `Cypher` or `GQL` — NOT defaulted at runtime
- `JoinSpec.Keys`: MUST have exactly 2 entries (cross-source join requires 2 entity anchors)
- `Value.SecretRef`: FORBIDDEN in source code, config files, or environment variable defaults
- `Environment` overlays: MAY override parameters and feature flags; MUST NOT add new components
- `ResolvedManifest.Sources/Queries/Reactions`: sorted by ID (determinism guarantee)
