# Implementation Plan: azd-drasi Extension

**Branch**: `001-azd-drasi-extension` | **Date**: 2026-04-04 | **Spec**: [spec.md](./spec.md)  
**Research**: [research.md](./research.md) | **Data Model**: [data-model.md](./data-model.md)  
**CLI Contract**: [contracts/cli-contract.md](./contracts/cli-contract.md) | **Quick Start**: [quickstart.md](./quickstart.md)

## Summary

A Go-based Azure Developer CLI (`azd`) extension that registers the `azd drasi` command group,
enabling developers to scaffold, provision, deploy, monitor, and tear down Drasi reactive
data-pipeline workloads on AKS from a single, idiomatic CLI entrypoint.

The extension integrates with the `azdext` gRPC SDK, reads a multi-file declarative YAML
configuration model (`drasi/drasi.yaml` + glob-resolved sources/queries/reactions), performs
offline validation with DAG cycle detection, executes deployment via the Drasi CLI subprocess
wrapper (positional-arg syntax, min v0.10.0), manages deployment state with content-hash
diffing in the azd environment file, and provisions AKS infrastructure + Workload Identity
via Bicep IaC. All Azure identity patterns use OIDC + FederatedIdentityCredential — no
connection strings or service principal secrets.

---

## Technical Context

**Language/Version**: Go 1.22+  
**Primary Dependencies**:

- `github.com/azure/azure-dev/cli/azd/pkg/azdext` — azd extension gRPC SDK
- `github.com/spf13/cobra` v1.8+ — CLI command framework
- `gopkg.in/yaml.v3` — YAML loading/serialisation
- `github.com/santhosh-tekuri/jsonschema/v6` — JSON Schema validation
- `github.com/stretchr/testify` — unit test assertions
- `github.com/testcontainers/testcontainers-go` — integration test containers

**Storage**: azd environment state file (`.azure/<env>/`) for content-hash persistence;
Azure Key Vault for secrets (read-only at deploy time); no database

**Testing**: `go test ./...`; `golangci-lint`; `bicep build`; `yamllint`; `az deployment what-if`

**Target Platform**: azd extension binary (windows/amd64, linux/amd64, darwin/amd64, darwin/arm64)

- AKS cluster (AKS 1.28+, `drasi-system` namespace)

**Project Type**: CLI extension (binary, distributed via azd extension registry)

**Performance Goals**:

- `azd drasi validate` (offline): < 2 seconds for projects up to 200 component YAML files (sources + queries + reactions + middleware) on a standard CI runner
- `azd drasi deploy` end-to-end: < 15 minutes total (5 min per-component timeout)
- `azd drasi provision` (new cluster + Drasi install): 8–12 minutes typical

**Constraints**:

- AKS is NON-NEGOTIABLE for Drasi hosting (Constitution Principle VI)
- No connection-string auth anywhere (Constitution Principle IV, NON-NEGOTIABLE)
- 80% test coverage gate (Constitution Principle VIII)
- No `os.Exit` in extension code (azdext SDK contract — gRPC channel corruption)
- `stdout` write FORBIDDEN in lifecycle event handlers (gRPC channel corruption)
- Drasi CLI min version: `0.10.0`; fail-fast with `ERR_DRASI_CLI_VERSION` if older

**Scale/Scope**: Single-developer local use to CI/CD pipelines; extension distributable via
public azd registry; targets workloads with up to ~50 Drasi components per environment

## Constitution Check

_GATE: Must pass before Phase 0 research. Re-checked after Phase 1 design — PASS._

| #    | Principle                           | Gate                                                               | Status             | Notes                                                                                         |
| ---- | ----------------------------------- | ------------------------------------------------------------------ | ------------------ | --------------------------------------------------------------------------------------------- |
| I    | Declarative IaC First               | All Azure resources in Bicep; no post-deploy portal config         | ✅ PASS            | IaC in `infra/modules/`; 2-phase KV lockdown is Bicep-only                                    |
| II   | Idempotency                         | All `azd drasi` commands re-entrant; missing state = create        | ✅ PASS            | Content-hash change detection ensures safe re-runs                                            |
| III  | Composable with azd                 | Registers as azd extension; supports `lifecycle-events` capability | ✅ PASS            | `extension.yaml` declares `custom-commands` + `lifecycle-events`                              |
| IV   | Secure by Default (NON-NEGOTIABLE)  | No connection strings; Workload Identity; Key Vault secrets only   | ✅ PASS            | OIDC + FederatedIdentityCredential; KV→K8s Secret translation                                 |
| V    | Developer Experience                | `azd drasi` namespace; `--output json`; `--dry-run`; < 2s validate | ✅ PASS            | Core 8 commands plus `upgrade` stub follow a consistent flag/output contract                  |
| VI   | AKS Required (NON-NEGOTIABLE)       | Drasi components MUST run on AKS                                   | ✅ PASS (OVERRIDE) | User plan input specified Container Apps — **overridden**. AKS with OIDC + Workload Identity. |
| VII  | Observable by Default               | OTel → Log Analytics workspace; structured structured error codes  | ✅ PASS            | `observability/` package; ContainerLogV2; Managed Prometheus                                  |
| VIII | Test-First Quality (NON-NEGOTIABLE) | TDD; 80% coverage; `golangci-lint`; `bicep lint`                   | ✅ PASS            | Tests written before implementation in each phase                                             |
| IX   | Semver Distribution                 | `extension.yaml` version ≥ 1.0.0; GitHub Releases + registry.yaml  | ✅ PASS            | `version.txt` is single source of truth                                                       |
| X    | Documentation                       | README quickstart; `docs/` directory; troubleshooting guide        | ✅ PASS            | `quickstart.md` generated; `docs/` in project structure                                       |

### Violation Override Record

| Violation                                               | Principle                     | Resolution                                                                                                                                            |
| ------------------------------------------------------- | ----------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| User plan input: Azure Container Apps for Drasi runtime | Principle VI (NON-NEGOTIABLE) | Overridden to AKS. Container Apps does not support Kubernetes CRDs, Dapr sidecar injection, or `drasi init` kubeconfig semantics. See research.md §2. |

## Project Structure

### Specification Artifacts (this feature)

```text
specs/001-azd-drasi-extension/
├── plan.md                      # This file
├── research.md                  # Phase 0 research findings
├── data-model.md                # Phase 1 entity model + Go types
├── quickstart.md                # Phase 1 getting-started guide
├── contracts/
│   └── cli-contract.md          # Phase 1 CLI surface contract
└── tasks.md                     # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
azd.extensions.drasi/
│
│  # Extension binary entrypoint
├── main.go                          # azdext.Run(cmd.NewRootCommand())
├── extension.yaml                   # Manifest: namespace=drasi, capabilities
├── version.txt                      # Semver string (e.g. "1.0.0") — single source of truth
├── go.mod
├── go.sum
├── build.ps1                        # Cross-compile: windows/amd64
├── build.sh                         # Cross-compile: linux/amd64, darwin/amd64, darwin/arm64
│
│  # Cobra command layer
├── cmd/
│   ├── root.go                      # NewRootCommand(); --output --debug flags; version string
│   ├── init.go                      # azd drasi init  [--template] [--force] [--environment]
│   ├── provision.go                 # azd drasi provision [--environment]
│   ├── deploy.go                    # azd drasi deploy [--config] [--environment] [--dry-run]
│   ├── status.go                    # azd drasi status [--environment]
│   ├── logs.go                      # azd drasi logs [--component] [--kind] [--tail] [--follow] [--environment]
│   ├── diagnose.go                  # azd drasi diagnose [--environment]
│   ├── validate.go                  # azd drasi validate [--config] [--strict] [--environment]
│   ├── teardown.go                  # azd drasi teardown --force [--infrastructure] [--environment]
│   └── upgrade.go                   # azd drasi upgrade (FR-010 stub)
│
│  # Business logic (no cobra, no azdext, no os.Exit)
├── internal/
│   │
│   │  # Configuration engine
│   ├── config/
│   │   ├── model.go                 # DrasiManifest, Source, ContinuousQuery, Reaction, Value, SecretRef
│   │   ├── loader.go                # Multi-file YAML loader; glob resolution; manifest merge
│   │   ├── resolver.go              # Environment overlay merge; deterministic sort
│   │   ├── schema.go                # Embedded JSON Schema files + validation via jsonschema/v6
│   │   └── schema/                  # Embedded JSON Schema assets
│   │       ├── source.schema.json
│   │       ├── continuousquery.schema.json
│   │       ├── reaction.schema.json
│   │       ├── middleware.schema.json
│   │       └── manifest.schema.json
│   │
│   │  # Offline validation
│   ├── validation/
│   │   ├── validator.go             # Top-level Validate(manifest) → ValidationResult
│   │   ├── references.go            # Cross-reference resolution: queries→sources, queries→reactions
│   │   ├── graph.go                 # DAG adjacency list + DFS cycle detection (Tarjan's)
│   │   ├── querylang.go             # Validates queryLanguage presence (never default)
│   │   └── errors.go                # ValidationIssue, ValidationResult, error codes
│   │
│   │  # Drasi CLI subprocess wrapper
│   ├── drasi/
│   │   ├── client.go                # Client struct; version check; subprocess exec wrapper
│   │   ├── apply.go                 # ApplyFile(ctx, path) via `drasi apply -f <path>`
│   │   ├── wait.go                  # WaitOnline(ctx, kind, id, timeout) via `drasi wait <kind> <id>`
│   │   ├── delete.go                # DeleteComponent(ctx, kind, id) via `drasi delete <kind> <id>`
│   │   ├── list.go                  # ListComponents(ctx, kind) via `drasi list <kind>`
│   │   └── describe.go              # DescribeComponent(ctx, kind, id) via `drasi describe <kind> <id>`
│   │
│   │  # Deployment engine
│   ├── deployment/
│   │   ├── engine.go                # Deploy(ctx, plan) → DeploymentResult; orchestrate create/delete-then-apply/noop
│   │   ├── diff.go                  # ComputeHash(component) → string; BuildPlan(manifest, state) → DeploymentPlan
│   │   ├── state.go                 # ReadState(env, key) / WriteState(env, key, val) via azd env file
│   │   ├── order.go                 # SortForDeploy(plan) → ordered; SortForDelete(plan) → reversed
│   │   └── timeout.go               # PerComponentTimeout=5min, TotalTimeout=15min; ERR_COMPONENT_TIMEOUT
│   │
│   │  # Key Vault → K8s Secret translation
│   ├── keyvault/
│   │   ├── spc.go                   # SecretProviderClass + synced Secret manifest generator
│   │   └── translator.go            # TranslateRefs(manifest) → replaces SecretRef with K8s Secret ref
│   │
│   │  # Project scaffolding
│   ├── scaffold/
│   │   ├── engine.go                # Scaffold(template, target, force) → []CreatedFile
│   │   └── templates/               # Embedded template FS
│   │       ├── blank/               # Minimal drasi.yaml + empty dirs
│   │       ├── cosmos-change-feed/  # Cosmos DB Gremlin source starter
│   │       ├── event-hub-routing/   # Event Hub source starter
│   │       └── query-subscription/  # Generic query + dapr-pubsub reaction starter
│   │
│   │  # Observability
│   ├── observability/
│   │   ├── tracer.go                # OTel trace provider → Azure Monitor OTLP exporter
│   │   └── metrics.go               # OTel meter provider → Managed Prometheus exporter
│   │
│   │  # Output formatting
│   └── output/
│       ├── formatter.go             # Format(data, OutputFormat) → string; table/json modes
│       └── errors.go                # FormatError(code, msg) → structured error; exit code mapping
│
│  # Bicep IaC
├── infra/
│   ├── main.bicep                   # Root module; wires all sub-modules
│   ├── main.parameters.bicepparam   # Parameter file; env-specific values
│   └── modules/
│       ├── aks.bicep                # AKS cluster (OIDC issuer + Workload Identity addon enabled)
│       ├── keyvault.bicep           # Key Vault; RBAC enabled; 2-phase public→locked pattern
│       ├── uami.bicep               # User-Assigned MI + 3 role assignments (KV, Monitoring, optional AcrPull)
│       ├── loganalytics.bicep       # Log Analytics workspace; ContainerLogV2; OTel config
│       ├── fedcred.bicep            # FederatedIdentityCredential; subject=system:serviceaccount:drasi-system:drasi-resource-provider
│       ├── acr.bicep                # ACR (conditional; enabled by usePrivateAcr Bicep param)
│       ├── cosmos.bicep             # Optional Cosmos DB (conditional)
│       ├── eventhub.bicep           # Optional Event Hubs namespace (conditional)
│       └── servicebus.bicep         # Optional Service Bus (conditional)
│
│  # Dev Container
├── .devcontainer/
│   └── devcontainer.json            # Tools: azd≥1.10.0, drasi≥0.10.0, dapr, go1.22, kubectl, bicep, azure-cli
│
│  # CI/CD
├── .github/
│   └── workflows/
│       ├── ci.yml                   # PR gate: build, test (go test ./...), lint, bicep build
│       └── release.yml              # Tag gate: cross-compile 4 targets, GitHub Release, registry.yaml update
│
│  # Documentation
├── README.md                        # Quick start, prerequisites, commands overview
└── docs/
    ├── architecture.md              # Component diagram, data flow, AKS/Drasi integration
    ├── configuration-reference.md   # Full drasi.yaml schema, all fields, SecretRef syntax
    └── troubleshooting.md           # Error codes, common failures, diagnostic steps
```

### Scaffolded Project Output (created by `azd drasi init`)

`azd drasi init` MUST generate a self-contained azd project on disk. In particular, IaC lives **in the user project** (as azd expects), and the extension copies the canonical embedded templates into the scaffold output.

```text
<new project>/
├── azure.yaml
├── drasi/
│   ├── drasi.yaml
│   ├── environments/
│   │   └── dev.yaml
│   ├── sources/
│   ├── queries/
│   └── reactions/
├── infra/
│   ├── main.bicep
│   ├── main.parameters.bicepparam
│   └── modules/
└── docker-compose.yml
```

---

## Implementation Phases

> **Note**: tests are written BEFORE implementation code in each phase (TDD — Constitution Principle VIII).
> Each phase ends with `go test ./...` green + `golangci-lint` clean before moving on.

> **Phase numbering**: plan.md uses Phase 0 (research/design) as pre-work. The implementation phases in `tasks.md` begin at **Phase 1** and run through **Phase 8**. The mapping is:
> `plan.md Phase 0 (Research)` → no tasks.md equivalent (complete)
> `tasks.md Phase 1` → Setup | `Phase 2` → Foundational | `Phases 3–7` → User Stories (US5, US1, US2, US3, US4 in dependency order) | `Phase 8` → Polish

---

### Phase 0: Research — COMPLETE ✅

_See [research.md](./research.md) for all findings._

Verified items:

- azd `azdext` SDK gRPC model and extension.yaml manifest contract
- Drasi CLI positional-arg syntax for extension command set (`init`, `apply`, `wait`, `delete`, `list`, `describe`)
- AKS-only Drasi hosting model (Container Apps override documented)
- Workload Identity FederatedIdentityCredential subject (`system:serviceaccount:drasi-system:drasi-resource-provider`)
- Content-hash state management via azd environment state file
- Key Vault RBAC roles (GUIDs verified from `secret-management` skill)
- Multi-file YAML config engine design decisions
- CI/CD pipeline approach (GitHub Actions + golangci-lint + bicep lint)

---

### Phase 1: Design Artifacts — COMPLETE ✅

Generated:

- [data-model.md](./data-model.md) — Go type definitions for all 8 entities
- [contracts/cli-contract.md](./contracts/cli-contract.md) — CLI surface (8 core commands + `upgrade` stub, flags, exit codes, JSON shapes)
- [quickstart.md](./quickstart.md) — 10-step getting-started guide

---

### Phase 2: Extension Scaffold + Command Framework

**Goal**: Runnable `azd drasi --help` with the full command inventory from `contracts/cli-contract.md` registered and validated.

**Deliverables**:

- `go.mod` with all primary dependencies pinned
- `main.go`: `azdext.Run(cmd.NewRootCommand())` entry point
- `extension.yaml`: namespace `drasi`, capabilities, OS-specific `executablePath`, `minAzdVersion: "1.10.0"`
- `version.txt`: `1.0.0`
- `cmd/root.go`: root command, `--output` flag (table/json), `--debug` flag; registers all subcommands including `newListenCommand()`
- `cmd/listen.go`: `newListenCommand()` \u2014 `RunE` uses `azdext.WithAccessToken(cmd.Context())`, `azdext.NewEventManager(azdClient)`, subscribes to `postProvision`/`preDeploy` events, calls `eventManager.Receive(ctx)` (blocking); required by the `lifecycle-events` capability in extension.yaml
- Command stubs: full command inventory from `contracts/cli-contract.md` in `cmd/`, correct flags, exit codes, placeholder output
- `internal/output/formatter.go`: `Format()` for table + JSON modes
- `internal/output/errors.go`: `FormatError()` for all error codes
- `build.ps1` + `build.sh`: cross-compile scripts for all 4 target platforms

**Tests (TDD — write first)**:

- `cmd/root_test.go`: flag parsing, help text, version string
- `internal/output/formatter_test.go`: table output, JSON output, error shapes

**Quality gates**: `azd drasi --help` runs cleanly; `golangci-lint` clean; 80%+ coverage on `cmd/` + `internal/output/`

---

### Phase 3: Configuration Engine

**Goal**: Load, resolve, validate, and hash any Drasi YAML project directory.

**Deliverables**:

- `internal/config/model.go`: all entity structs (from data-model.md)
- `internal/config/loader.go`: glob resolution from `drasi/drasi.yaml`, multi-file load
- `internal/config/resolver.go`: environment overlay merge, deterministic sort
- `internal/config/schema.go`: embedded JSON Schema validation per entity type
- `internal/config/schema/*.schema.json`: schema files for all entity types
- `internal/validation/validator.go`: top-level validate pipeline
- `internal/validation/references.go`: cross-reference checker (query→source, query→reaction)
- `internal/validation/graph.go`: DAG + Tarjan DFS cycle detection
- `internal/validation/querylang.go`: `queryLanguage` presence enforcement
- `internal/validation/errors.go`: error code constants + ValidationIssue type
- `cmd/validate.go`: implemented (calls loader + validator + formatter)

**Tests (TDD — write first)**:

- `internal/config/loader_test.go`: single file, multi-file glob, missing file, bad YAML
- `internal/config/resolver_test.go`: overlay merge, determinism
- `internal/validation/references_test.go`: valid refs, broken ref, multi-source
- `internal/validation/graph_test.go`: linear chain, cycle, disconnected
- `internal/validation/querylang_test.go`: present, absent, default value trap

**Gate**: `azd drasi validate` returns correct exit codes for all test fixture projects

---

### Phase 4: Drasi CLI Client

**Goal**: Reliable subprocess wrapper for all Drasi CLI commands with version gating.

**Deliverables**:

- `internal/drasi/client.go`: `Client` struct; `CheckVersion()` → `ERR_DRASI_CLI_VERSION` if < 0.10.0; `exec.CommandContext` wrapper
- `internal/drasi/apply.go`: `ApplyFile(ctx, path)` → `drasi apply -f <path>`
- `internal/drasi/wait.go`: `WaitOnline(ctx, kind, id, timeout)` → `drasi wait <kind> <id> --timeout <n>`
- `internal/drasi/delete.go`: `DeleteComponent(ctx, kind, id)` → `drasi delete <kind> <id>`
- `internal/drasi/list.go`: `ListComponents(ctx, kind)` → `drasi list <kind>`, parses tabular output
- `internal/drasi/describe.go`: `DescribeComponent(ctx, kind, id)` → `drasi describe <kind> <id>`, parses output to struct

Note: always use canonical kind string `continuousquery` (not alias `query`).

**Tests (TDD — write first)**:

- `internal/drasi/client_test.go`: version check pass + fail; CLI not found
- `internal/drasi/apply_test.go`: apply succeeds, apply fails, timeout
- `internal/drasi/wait_test.go`: online immediately, online after polling, timeout error
- `internal/drasi/delete_test.go`: delete succeeds, not-found case
- `internal/drasi/list_test.go`: parse tabular output, empty result
- `internal/drasi/describe_test.go`: parse component metadata

Approach: Mock subprocess outputs via `fakeDrasi` test helper (in-process fake binary or file-based mock).

---

### Phase 5: Deployment Engine

**Goal**: Idempotent, hash-diffed, ordered deployment of Drasi components.

**Deliverables**:

- `internal/deployment/state.go`: `ReadHash(env, kind, id)` / `WriteHash(env, kind, id, hash)` using azdext Environment gRPC service (GetEnvironmentValue/SetEnvironmentValue); no direct `.azure/<env>/.env` file I/O
- `internal/deployment/diff.go`: `ComputeHash(component)` (SHA-256 of canonical YAML); `BuildPlan(manifest, state)` → `DeploymentPlan` with Create/DeleteThenApply/NoOp per component
- `internal/deployment/order.go`: `SortForDeploy()` → sources, queries, middleware, reactions; `SortForDelete()` → reverse
- `internal/deployment/timeout.go`: per-component context with 5min deadline; total deployment context with 15min deadline; emit `ERR_COMPONENT_TIMEOUT`
- `internal/deployment/engine.go`: `Deploy(ctx, plan, drasiClient)` → executes ordered actions; updates hashes on success; preserves partial-failure state for next-run recovery
- Key Vault translation: Before deploy, translate all `SecretRef` values in `ResolvedManifest` to K8s Secret references via `internal/keyvault/translator.go`
- `internal/keyvault/spc.go`: `BuildSecretProviderClass(manifest)` produces Secrets Store CSI manifests for the deploy scope
- `internal/keyvault/translator.go`: `TranslateRefs(ctx, manifest)` walks all properties, emits SecretProviderClass/synced Secret manifests, and replaces `SecretRef` with K8s Secret reference
- `cmd/deploy.go`: implemented (validation gate → KV translation → build plan → dry-run or execute → output result)

**Tests (TDD — write first)**:

- `internal/deployment/diff_test.go`: same config → NoOp; changed config → DeleteThenApply; missing hash → Create
- `internal/deployment/order_test.go`: ordering, reverse ordering
- `internal/deployment/engine_test.go`: full happy-path deploy; partial failure + recovery; timeout handling
- `internal/keyvault/translator_test.go`: SecretRef → K8s ref; plain string passthrough; missing secret error

---

### Phase 6: Bicep IaC — Infrastructure Provision

**Goal**: Fully declarative AKS + Key Vault + UAMI + Workload Identity Bicep that provisions the
complete Drasi hosting environment in one `azd drasi provision` invocation.

**Deliverables**:

- `infra/modules/aks.bicep`:
  - AKS cluster with `enableOidcIssuer: true`, `enableWorkloadIdentity: true`
  - `omsAgent` prerequisite for ContainerLogV2
  - Node pool: Standard_D4s_v5 or smaller (dev), configurable via param
  - API version verification required before coding
- `infra/modules/keyvault.bicep`:
  - `enableRbacAuthorization: true`
  - Soft delete + purge protection (90-day retention)
  - 2-phase pattern: Phase 1 deploy with `publicNetworkAccess: Enabled`; Phase 2 lock down after secrets populated
- `infra/modules/uami.bicep`:
  - User-Assigned MI
  - Role assignments:
    - Key Vault Secrets User (`4633458b-17de-408a-b874-0445c86b69e6`) on Key Vault
    - Monitoring Metrics Publisher (`3913510d-42f4-4e42-8a64-420c390055eb`) on Log Analytics workspace
    - AcrPull (`7f951dda-4ed3-4680-a7ca-43fe172d538d`) on ACR (conditional on `usePrivateAcr`)
- `infra/modules/fedcred.bicep`:
  - `FederatedIdentityCredential` on the UAMI
  - Subject: `system:serviceaccount:${drasiNamespace}:drasi-resource-provider`
  - Audience: `api://AzureADTokenExchange`
  - Service account identity aligned to current upstream installer manifests for the resource provider
- `infra/modules/loganalytics.bicep`:
  - Log Analytics workspace
  - ContainerLogV2 data collection rule
  - Managed Prometheus + Azure Monitor Workspace
- `infra/modules/acr.bicep` (conditional): ACR with `Premium` sku for private networking
- `infra/modules/cosmos.bicep` (conditional): Cosmos DB Gremlin account
- `infra/modules/eventhub.bicep` (conditional): Event Hubs namespace
- `infra/main.bicep`: root module wiring all sub-modules; `drasiNamespace`, `usePrivateAcr`, `enableCosmosDb`, `enableEventHub` params
- `cmd/provision.go`: implemented — calls Bicep via azd lifecycle, then runs `drasi init --context <aks-context>`; writes `DRASI_PROVISIONED=true` to azd env

**Tests**:

- `bicep build` must succeed (zero warnings)
- `az deployment what-if --template-file infra/main.bicep` in CI pipeline
- Unit test: `cmd/provision_test.go` (flag parsing, env write)
- Integration tests: deployment/provision boundary coverage (containerized and/or ephemeral Azure env), including provider registration and idempotent recovery paths

---

### Phase 7: Scaffold + Init Command

**Goal**: `azd drasi init` creates a fully working project skeleton in < 2 seconds.

**Deliverables**:

- `internal/scaffold/templates/blank/`: minimal `drasi.yaml` + empty source/query/reaction stubs
- `internal/scaffold/templates/cosmos-change-feed/`: Cosmos Gremlin source + Cypher query + dapr-pubsub reaction; all secrets as KV refs
- `internal/scaffold/templates/event-hub-routing/`: Event Hub source + query + reaction
- `internal/scaffold/templates/query-subscription/`: generic source + parametric Cypher query
- `internal/scaffold/engine.go`: `Scaffold(template, dir, force)` → copy/render template files; reject on conflict unless `--force`
- `cmd/init.go`: implemented — calls scaffold engine, reports created files
- `.devcontainer/devcontainer.json`: installs azd ≥ 1.10.0 via azd feature, drasi ≥ 0.10.0, dapr, go 1.22, kubectl, bicep, azure-cli

**Tests (TDD — write first)**:

- `internal/scaffold/engine_test.go`: blank template, cosmos template, conflict rejection, force overwrite

---

### Phase 8: Observability Commands

**Goal**: `status`, `logs`, `diagnose` commands provide complete operational visibility.

**Deliverables**:

- `cmd/status.go`: calls `drasiClient.ListComponents()` per kind → formats health table; exit 1 if any non-Online
- `cmd/logs.go`: shell out to `kubectl logs` for Drasi pods in `drasiNamespace`; supports `--follow` streaming
- `cmd/diagnose.go`: 5 checks (AKS reachable, Drasi API pod running, Dapr injector running, Secrets Store CSI sync health for generated SecretProviderClass objects, Log Analytics data flowing); structured pass/fail output
- `internal/observability/tracer.go`: OTel trace provider → Azure Monitor OTLP endpoint; `APPLICATIONINSIGHTS_CONNECTION_STRING` from azd env
- `internal/observability/metrics.go`: OTel meter provider; emit deployment metrics (components deployed, errors, duration)

**Tests (TDD — write first)**:

- `cmd/status_test.go`: all-online, partial-offline, empty state
- `cmd/diagnose_test.go`: all-pass, individual check failures
- `internal/observability/tracer_test.go`: no-op when telemetry disabled (local dev)

---

### Phase 9: CI/CD Pipeline

**Goal**: Every PR validates the full build + test + lint chain in < 20 minutes.
Every semver tag produces cross-compiled binaries and a GitHub Release.

**Deliverables**:

- `.github/workflows/ci.yml`:
  - Trigger: `push` to any branch, `pull_request`
  - Steps: `go build ./...`, `go test ./... -coverprofile=coverage.out`, `golangci-lint run`, `bicep build infra/main.bicep`, `bicep lint infra/main.bicep`, `yamllint drasi/`, Azure-authenticated `az deployment what-if --template-file infra/main.bicep`
  - Coverage gate: fail if `go tool cover -func=coverage.out` < 80%
  - Matrix: `ubuntu-latest` (primary); `windows-latest` for cross-compile verification
- `.github/workflows/e2e-pr.yml`:
  - Trigger: `pull_request` (targeted paths)
  - Steps: `azd drasi provision` → `azd drasi deploy` → `azd drasi validate` against ephemeral environment using GitHub OIDC; enforce <20 min budget; always teardown/cleanup
- `.github/workflows/release.yml`:
  - Trigger: `push` to `v[0-9]+.*` tags
  - Steps: run CI checks, cross-compile 4 targets (`build.sh` + `build.ps1`), create GitHub Release with binary assets, update `registry.yaml` in extension registry (PR or direct commit)
  - Uses `GITHUB_TOKEN` only; no additional secrets

**Tests**:

- Workflow YAML lint via `actionlint` in CI
- Manual pre-release validation: `azd extension install --source local ./bin/linux/amd64/azd-drasi`

---

### Phase 10: Documentation

**Goal**: Any developer can go from zero to running in < 30 minutes using only the repo docs.

**Deliverables**:

- `README.md`: prerequisites table, installation, 6-step quick start, commands reference, troubleshooting link
- `docs/architecture.md`: component diagram (extension → azdext gRPC → azd; extension → drasi CLI → drasi-platform on AKS); data flow for provision + deploy
- `docs/configuration-reference.md`: full YAML schema for all entity types; SecretRef syntax; environment overlays; feature flags
- `docs/troubleshooting.md`: error code table (all ERR\_\* codes from contracts/cli-contract.md); common failure scenarios; diagnostic steps linking to `azd drasi diagnose`

---

## Risk Register

| Risk                                                                             | Probability | Impact | Mitigation                                                                                                                                  |
| -------------------------------------------------------------------------------- | ----------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------- |
| Upstream ServiceAccount naming changes in future Drasi releases break WI binding | Medium      | High   | Pin tested Drasi version in release notes; add CI smoke check validating expected ServiceAccount exists before fedcred apply                |
| azd state file API behavior changes in azdext SDK                                | Medium      | Medium | Keep `deployment/state.go` behind an adapter interface; validate against azdext gRPC Environment service in CI and fail fast if unavailable |
| `drasi wait` does not honour `--timeout` in v0.10.0                              | Low         | Medium | Add extension-level context timeout as belt-and-suspenders; tested via integration test                                                     |
| Key Vault purge-protection lockout during teardown testing                       | Low         | Low    | Document in troubleshooting.md; use unique KV names per test environment                                                                    |
| golangci-lint version drift between dev and CI                                   | Low         | Low    | Pin golangci-lint version in `ci.yml`; same version in devcontainer                                                                         |

---

## Dependency Map

```
Phase 2 (scaffold + stubs)
      │
      ▼
Phase 3 (config engine)  ──────────────────┐
      │                                     │
      ▼                                     ▼
Phase 4 (drasi CLI client)          Phase 5 (deployment engine)
      │                                     │
      └──────────────┬──────────────────────┘
                     ▼
             Phase 6 (Bicep IaC)          Phase 7 (scaffold + init)
                     │                           │
                     └────────────┬──────────────┘
                                  ▼
                          Phase 8 (observability commands)
                                  │
                                  ▼
                          Phase 9 (CI/CD pipeline)
                                  │
                                  ▼
                          Phase 10 (documentation)
```

Phases 3+4, Phase 6+7 can proceed concurrently once Phase 2 is complete.

---

## Key Technical Decisions Summary

| Decision            | Choice                                                                                                                                   | Alternative Rejected                                           |
| ------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------- |
| Extension runtime   | Go + azdext gRPC SDK                                                                                                                     | Python/shell — constitution locks Go                           |
| Drasi hosting       | AKS                                                                                                                                      | Azure Container Apps — Principle VI violation                  |
| Identity            | OIDC + Workload Identity + FederatedIdentityCredential                                                                                   | Service principal secret — Principle IV violation              |
| Config format       | Declarative YAML with glob includes                                                                                                      | Single monolithic file — does not scale                        |
| Validation          | Offline JSON Schema + cross-reference + DAG cycle detection                                                                              | Online-only — poor DX for iterate-and-test                     |
| State management    | azd env file + SHA-256 content hash                                                                                                      | Separate state file — creates drift risk                       |
| Secret access       | Secret material resolved in-cluster via Secrets Store CSI + Workload Identity; deploy translates Key Vault refs to K8s Secret references | Secrets in config file — Principle IV violation                |
| CLI validation gate | Pre-deploy pass required; deploy fails if validate fails                                                                                 | Warn-only — production deployments corrupted by invalid config |
| KV lockdown         | 2-phase deployment (public during IaC, locked after secrets)                                                                             | Lock from start — pipeline locked out of its own vault         |
