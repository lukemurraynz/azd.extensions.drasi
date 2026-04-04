# Tasks: azd-drasi Extension

**Feature**: `001-azd-drasi-extension` | **Branch**: `001-azd-drasi-extension`  
**Input**: [plan.md](./plan.md) · [spec.md](./spec.md) · [data-model.md](./data-model.md) · [contracts/cli-contract.md](./contracts/cli-contract.md)

> TDD is **MANDATORY** (Constitution Principle VIII). Tests marked `[TEST]` MUST be written to fail before the implementation task they precede.

## Format: `- [ ] TXXX [P?] [USN?] Description — file path`

- **[P]**: Parallelisable — different files, no dependency on incomplete tasks
- **[USN]**: User story label (US1–US5); not used in Setup / Foundational / Polish phases
- **[TEST]**: Write this test first; it must fail before the impl task is started

---

## Phase 1: Setup

**Purpose**: Initialise the Go module, extension manifest, build scripts, and lint config.
No user story tasks can begin until Phase 1 is complete.

- [ ] T001 Create go.mod and go.sum with all dependencies pinned (azdext SDK, cobra v1.8+, yaml.v3, jsonschema/v6, azsecrets, azidentity, testify, testcontainers-go) — `go.mod`
  > ⚠️ **CRITICAL**: Use `github.com/azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets` — the `sdk/keyvault/azsecrets` module path is **deprecated** and must NOT appear in go.mod
- [ ] T002 Create `main.go` entry point — calls `azdext.Run(cmd.NewRootCommand())`; `azdext.Run` handles cobra plumbing and provides the context that gRPC-calling commands consume via `azdext.WithAccessToken(cmd.Context())`
  > ⚠️ **gRPC auth requirement**: Every command `RunE` that uses `azdext.NewAzdClient()` MUST begin with `ctx := azdext.WithAccessToken(cmd.Context())` — omitting this causes all gRPC service calls to fail with authentication errors — `main.go`
- [ ] T003 [P] Create extension.yaml (namespace: drasi, capabilities: [custom-commands, lifecycle-events], executablePath for 5 targets, minAzdVersion: "1.10.0") — `extension.yaml`
- [ ] T004 [P] Create version.txt with initial semver string "1.0.0" — `version.txt`
- [ ] T005 [P] Create build.ps1 cross-compile script for windows/amd64 — `build.ps1`
- [ ] T006 [P] Create build.sh cross-compile script for linux/amd64, darwin/amd64, darwin/arm64 — `build.sh`
- [ ] T007 [P] Create .golangci.yml with required linters (errcheck, govet, staticcheck, gosec, gofmt) — `.golangci.yml`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Output formatting, error types, root command, config entity model, and the Drasi CLI
client shell. All user stories depend on this phase completing first.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

### Tests (write before implementation)

- [ ] T008 [TEST] Write formatter test covering table mode, JSON mode, nil/empty data — `internal/output/formatter_test.go`
- [ ] T009 [TEST] [P] Write errors test covering all ERR\_\* constants, exit code mapping, FormatError JSON shape — `internal/output/errors_test.go`
- [ ] T010 [TEST] Write root command test covering global flag registration, --output propagation, --debug flag, version string — `cmd/root_test.go`

### Implementation

- [ ] T011 Create `internal/output/formatter.go` — `Format(data any, fmt OutputFormat) string`; table renderer with aligned columns; JSON marshaller; `OutputFormat` type (table/json) — `internal/output/formatter.go`
- [ ] T012 [P] Create `internal/output/errors.go` — `FormatError(code, msg, remediation string) string`; all ERR\_\* string constants (ERR_NO_AUTH, ERR_DRASI_CLI_NOT_FOUND, ERR_DRASI_CLI_VERSION, ERR_COMPONENT_TIMEOUT, ERR_TOTAL_TIMEOUT, ERR_VALIDATION_FAILED, ERR_MISSING_REFERENCE, ERR_CIRCULAR_DEPENDENCY, ERR_MISSING_QUERY_LANGUAGE, ERR_KV_AUTH_FAILED, ERR_AKS_CONTEXT_NOT_FOUND, ERR_FORCE_REQUIRED, ERR_NO_MANIFEST, ERR_NOT_IMPLEMENTED, ERR_DEPLOY_IN_PROGRESS, ERR_DAPR_NOT_READY); exit code map — `internal/output/errors.go` > **ERR_NOT_IMPLEMENTED note**: This constant is intentional (not a TODO) — it supports the FR-010 `azd drasi upgrade` stub command. Do NOT remove it during lint/dead-code cleanup. Each call site should include `// intentional stub — FR-010` to prevent future automated removal.
- [ ] T013 Create `cmd/root.go` and `cmd/listen.go` — `NewRootCommand()` returns `*cobra.Command`; registers ALL subcommands including `newListenCommand()` from `cmd/listen.go` (**required** — azd calls `<ext> listen` when the extension declares the `lifecycle-events` capability); persistent `--output` flag (table|json, default table); persistent `--debug` flag; injects version from version.txt at build time.
      `newListenCommand()` in `cmd/listen.go`: `RunE` calls `ctx := azdext.WithAccessToken(cmd.Context())`, then `azdext.NewAzdClient()` + `defer azdClient.Close()`, then `azdext.NewEventManager(azdClient)` + `defer eventManager.Close()`; subscribes to `postprovision` and `predeploy` project events via `eventManager.AddProjectEventHandler`; calls `eventManager.Receive(ctx)` (blocks until azd closes connection); all diagnostic output goes to `os.Stderr` — NEVER `os.Stdout` (stdout is the gRPC channel) — `cmd/root.go`, `cmd/listen.go`
- [ ] T014 [P] Create `internal/config/model.go` — all entity structs from data-model.md: `DrasiManifest`, `IncludeSpec`, `Source`, `ContinuousQuery`, `SourceRef`, `JoinSpec`, `JoinKey`, `Reaction`, `Middleware`, `Environment`, `Value`, `SecretRef`, `EnvRef`, `ResolvedManifest`, `ComponentHash` with `StateKey()` — `internal/config/model.go`
- [ ] T015 [P] Create `internal/validation/errors.go` — `ValidationLevel` type; `ValidationIssue` struct (Level, File, Line, Code, Message, Remediation); `ValidationResult` struct; `HasErrors()` method — `internal/validation/errors.go`

**Checkpoint**: `go build ./...` succeeds; `golangci-lint run` clean; `go test ./internal/output/... ./cmd/ -run TestRoot` all pass.

---

## Phase 3: User Story 5 — Validate Configuration (Priority: P5)

**Goal**: `azd drasi validate` detects schema errors, broken cross-references, and circular
dependencies entirely offline, with file path + line number on every error.

**Independent Test**: Drop a query file that references `source: non-existent-id`. Run
`azd drasi validate`. Confirm exit 1, error identifies the file, line, and broken reference.
Fix and re-run; confirm exit 0.

### Tests (write before implementation)

- [ ] T016 [TEST] [US5] Write config loader test: single file load, multi-file glob expansion, missing file error, malformed YAML error — `internal/config/loader_test.go`
- [ ] T017 [TEST] [P] [US5] Write config resolver test: dev overlay merges into base, prod param overrides base, deterministic sort produces identical output across runs — `internal/config/resolver_test.go`
- [ ] T018 [TEST] [P] [US5] Write schema validation test: valid Source passes, unknown field fails, all 5 entity types validated — `internal/config/schema_test.go`
- [ ] T019 [TEST] [P] [US5] Write cross-reference validation test: valid source ref, unknown source ref = ERR_MISSING_REFERENCE, unknown reaction ref = ERR_MISSING_REFERENCE, multiple errors accumulated — `internal/validation/references_test.go`
- [ ] T020 [TEST] [P] [US5] Write DAG cycle detection test: linear chain passes, direct cycle A→B→A fails ERR_CIRCULAR_DEPENDENCY, disconnected graph passes — `internal/validation/graph_test.go`
- [ ] T021 [TEST] [P] [US5] Write queryLanguage enforcement test: explicit Cypher passes, explicit GQL passes, missing queryLanguage = ERR_MISSING_QUERY_LANGUAGE, all errors accumulated not first-fail — `internal/validation/querylang_test.go`
- [ ] T022 [TEST] [US5] Write validate command test: exit 0 on valid fixture, exit 1 on error fixture, exit 2 on missing drasi.yaml, --strict promotes warnings to errors, --output json valid schema — `cmd/validate_test.go`

### Implementation

- [ ] T023 [US5] Create config loader: glob resolution from `drasi.yaml` includes patterns, per-file YAML decode, accumulate into raw slices by kind — `internal/config/loader.go`
- [ ] T024 [P] [US5] Create config resolver: merge environment overlay parameters into base config, deterministic sort by ID on all slices, produce `ResolvedManifest` — `internal/config/resolver.go`
  > **Guard**: An overlay MUST only patch fields of component IDs that already exist in the base manifest. If an overlay references a component ID not present in base, emit `ERR_VALIDATION_FAILED` ("overlay references unknown component id: %s") rather than silently adding it — unanticipated component creation via overlay is almost always a config mistake.
- [ ] T025 [P] [US5] Create schema validation wrapper: embed JSON Schema files via `//go:embed`, validate each entity against its schema using jsonschema/v6, emit `ValidationIssue` with file + line — `internal/config/schema.go`
  > **[VERIFY]** Confirm the `santhosh-tekuri/jsonschema/v6` programmatic validation API before implementing. Expected pattern: `compiler := jsonschema.NewCompiler(); compiler.AddResource(schemaURI, schemaReader); schema, err := compiler.Compile(schemaURI); err = schema.Validate(instance)`. The v6 API changed significantly from v5 — do NOT assume v5 patterns (`jsonschema.Validate(instance, schemaPath)`) still work. Check the current README at `https://github.com/santhosh-tekuri/jsonschema`.
- [ ] T026 [P] [US5] Create embedded JSON Schema files for all 5 entity types — `internal/config/schema/manifest.schema.json`, `source.schema.json`, `continuousquery.schema.json`, `reaction.schema.json`, `middleware.schema.json`
- [ ] T027 [P] [US5] Create cross-reference validator: collect all source IDs and reaction IDs, validate every query's `sources` and `reactions` fields, emit ERR_MISSING_REFERENCE with file+line — `internal/validation/references.go`
- [ ] T028 [P] [US5] Create DAG validator: build adjacency list query→sources + query→reactions, Tarjan DFS cycle detection, emit ERR_CIRCULAR_DEPENDENCY listing the cycle path — `internal/validation/graph.go`
- [ ] T029 [P] [US5] Create queryLanguage validator: iterate all ContinuousQuery entities, reject any missing or empty `queryLanguage`, emit ERR_MISSING_QUERY_LANGUAGE — `internal/validation/querylang.go`
- [ ] T030 [US5] Create top-level validator pipeline: loader → resolver → schema → references → graph → querylang; accumulate all issues, never stop at first error — `internal/validation/validator.go`
- [ ] T031 [US5] Implement `cmd/validate.go`: `--config` flag (default `drasi/drasi.yaml`), `--strict` flag, call loader+validator, format results via output package, exit 0/1/2 — `cmd/validate.go`

**Checkpoint**: `azd drasi validate` exits 1 with specific file:line errors on all 5 failure types; exits 0 on all clean fixtures; `--output json` matches schema from contracts/cli-contract.md.

---

## Phase 4: User Story 1 — Scaffold a Drasi Project (Priority: P1) 🎯 MVP

**Goal**: `azd drasi init` scaffolds a complete project in an empty directory. Re-running is
idempotent. `--template cosmos-change-feed` pre-populates a working configuration.

**Independent Test**: Run `azd drasi init` in a temp dir. Confirm all declared files exist and
are valid YAML. Run again — no files modified (idempotent). Run `azd drasi validate` on the
output; exit 0 with no errors. No Azure subscription needed.

### Tests (write before implementation)

- [ ] T032 [TEST] [US1] Write scaffold engine test: blank template creates expected 6 files, conflict on re-run without --force, --force overwrites, returned file list matches actual FS — `internal/scaffold/engine_test.go`
- [ ] T033 [TEST] [P] [US1] Write init command test: --template flag accepted (blank/cosmos-change-feed/event-hub-routing/query-subscription), --force flag, --output json emits file list only, idempotent re-run exits 0 with no-changes message — `cmd/init_test.go`

### Implementation

- [ ] T034 [US1] Create blank template files — `internal/scaffold/templates/blank/`:
  - `drasi/drasi.yaml` with `includes`, `environments`, and `featureFlags` stubs (all values empty/false by default so validate exits 0 on a blank scaffold)
  - `drasi/sources/example-source.yaml`, `drasi/queries/example-query.yaml`, `drasi/reactions/example-reaction.yaml`
  - `drasi/environments/dev.yaml` (empty overlay showing the merge format, with inline comments)
  - `docker-compose.yml` (FR-036): binds a local Drasi server (`ghcr.io/drasi-project/drasi:${DRASI_VERSION:-0.10.0}`) + Dapr sidecar service; **[VERIFY]** exact image reference against `https://github.com/orgs/drasi-project/packages` before committing — the image name and tag format must be confirmed from the published packages list
  - `infra/README.md` placeholder file: contains one sentence pointing to `../../infra/main.bicep` and instructs user to run `azd drasi provision` rather than deploying directly; NOT a Bicep module itself
  - `.vscode/launch.json` stub with a Go extension debug configuration targeting `main.go`; inline comment marking it as a development convenience stub (not production-required) so contributors can F5-debug the extension binary without confusion about whether it does anything deployment-related
- [ ] T035 [P] [US1] Create cosmos-change-feed template: Cosmos Gremlin Source yaml (KV refs for connection), Cypher ContinuousQuery yaml with explicit `queryLanguage: Cypher`, dapr-pubsub Reaction yaml, all secrets as `{kind: secret, vaultName: ..., secretName: ...}` refs, inline comments explaining each field — `internal/scaffold/templates/cosmos-change-feed/`
- [ ] T036 [P] [US1] Create event-hub-routing template: Event Hub Source yaml, query yaml, reaction yaml with inline comments — `internal/scaffold/templates/event-hub-routing/`
- [ ] T037 [P] [US1] Create query-subscription template: generic PostgreSQL Source yaml, parametric Cypher query yaml, dapr-pubsub Reaction yaml — `internal/scaffold/templates/query-subscription/`
- [ ] T038 [US1] Create scaffold engine: embed templates via `//go:embed`, `Scaffold(templateName, targetDir string, force bool) ([]string, error)`, conflict detection (fail if file exists + no force), copy+render template files, return created file paths — `internal/scaffold/engine.go`
- [ ] T039 [US1] Implement `cmd/init.go`: `--template` flag defaulting to `blank`, `--force` flag, call scaffold engine, format created files list via output package — `cmd/init.go`
- [ ] T040 [P] [US1] Create Dev Container definition: features block installing azd (≥1.10.0), drasi CLI (≥0.10.0), dapr, go 1.22, kubectl, bicep, azure-cli; `postCreateCommand: go mod download` — `.devcontainer/devcontainer.json`

**Checkpoint**: `azd drasi init` creates all expected files; `azd drasi validate` on the output exits 0; re-run exits 0 with "nothing to change"; `--template cosmos-change-feed` creates all 3 component YAMLs with no schema errors.

---

## Phase 5: User Story 2 — Provision Azure Infrastructure (Priority: P2)

**Goal**: `azd drasi provision` deploys AKS + Key Vault + UAMI + Workload Identity via Bicep
and installs the Drasi runtime. Re-running converges without duplicating resources.

**Independent Test**: Run `azd drasi provision` against a fresh Azure subscription environment
from a validated `drasi/environments/dev.yaml`. Confirm all resources exist with required tags,
all role assignments are present, no secrets in any deployment output, exit 0.

### Tests (write before implementation)

- [ ] T041 [TEST] [US2] Write provision command test: --environment flag wires to azd lifecycle, DRASI_PROVISIONED written on success, ERR_NO_AUTH on missing credentials, --output json emits resource IDs — `cmd/provision_test.go`

### Implementation (Bicep modules — T042–T050 parallelisable)

- [ ] T042 [US2] Create Log Analytics workspace module: workspace resource, ContainerLogV2 data collection rule, Azure Monitor Workspace for Managed Prometheus — `infra/modules/loganalytics.bicep`
- [ ] T043 [P] [US2] Create Key Vault module: enableRbacAuthorization: true, softDelete enabled, purgeProtection enabled (90d retention), publicNetworkAccess: Enabled (Phase 1; Phase 2 lockdown via separate deployer step) — `infra/modules/keyvault.bicep`
- [ ] T044 [P] [US2] Create UAMI module: User-Assigned Managed Identity resource; role assignments — Key Vault Secrets User (`4633458b-17de-408a-b874-0445c86b69e6`) on KV, Monitoring Metrics Publisher (`3913510d-42f4-4e42-8a64-420c390055eb`) on Log Analytics, conditional AcrPull (`7f951dda-4ed3-4680-a7ca-43fe172d538d`) when `usePrivateAcr == true` — `infra/modules/uami.bicep`
- [ ] T045 [P] [US2] Create AKS cluster module: AKS 1.28+, `enableOidcIssuer: true`, `enableWorkloadIdentity: true`, omsAgent DCR association for ContainerLogV2, Standard_D4s_v5 system node pool, `nodeCount` param — `infra/modules/aks.bicep`
- [ ] T046 [P] [US2] Create FederatedIdentityCredential module: resource `FederatedIdentityCredential` on UAMI, subject `system:serviceaccount:${drasiNamespace}:drasi-api`, audience `api://AzureADTokenExchange`, OIDC issuer from AKS output — `infra/modules/fedcred.bicep`
- [ ] T047 [P] [US2] Create conditional ACR module: Premium SKU ACR, enabled by `usePrivateAcr` Bicep param, outputs loginServer — `infra/modules/acr.bicep`
- [ ] T048 [P] [US2] Create conditional Cosmos DB module: Gremlin API account, enabled by `enableCosmosDb` param — `infra/modules/cosmos.bicep`
- [ ] T049 [P] [US2] Create conditional Event Hubs module: Standard namespace, enabled by `enableEventHub` param — `infra/modules/eventhub.bicep`
- [ ] T050 [P] [US2] Create conditional Service Bus module: Standard namespace, enabled by `enableServiceBus` param — `infra/modules/servicebus.bicep`
- [ ] T051 [US2] Create root Bicep module wiring all sub-modules; params: location, environmentName, aksClusterName, drasiNamespace (default drasi-system), keyVaultName, uamiName, logAnalyticsWorkspaceName, usePrivateAcr (bool, default false), acrName, enableCosmosDb, enableEventHub, enableServiceBus — `infra/main.bicep`
- [ ] T052 [P] [US2] Create parameter file with env-specific defaults and placeholder values — `infra/main.parameters.bicepparam`
- [ ] T053 [US2] Implement `cmd/provision.go`: invoke Bicep via azd lifecycle hook, run `drasi init --context <aks-context>` (with `--registry <acr>` when usePrivateAcr), write `DRASI_PROVISIONED=true` to azd env state, emit resource IDs on success — `cmd/provision.go`
  > **[VERIFY] FR-025**: Confirm whether `drasi init` automatically registers all default SourceProviders and ReactionProviders (PostgreSQL, CosmosGremlin, SQLServer, EventHub, SignalR, EventGrid, StorageQueue, StoredProc, Debug) or whether each provider requires a separate `drasi apply -f <provider-manifest>.yaml` call. Check the Drasi CLI `drasi init --help` output and the drasi-project/drasi GitHub docs. If registration is not automatic, T053 must add an explicit loop applying each default provider manifest after `drasi init`. The current task assumes `drasi init` handles this; if that assumption is wrong, an additional task T053b will be needed.

**Checkpoint**: `bicep build infra/main.bicep` exits 0 with zero warnings; `azd drasi provision` completes without portal intervention; all resources tagged; re-run is no-op.

---

## Phase 6: User Story 3 — Deploy Drasi Components (Priority: P3)

**Goal**: `azd drasi deploy` validates, translates KV refs to K8s Secrets, and idempotently
applies sources → queries → reactions to the Drasi runtime on AKS.

**Independent Test**: Run `azd drasi deploy` on a provisioned environment with a sample config.
Confirm all components Online. Re-run without changes — no-op (all hashes match). Change one
query file, re-run — only that query is deleted and re-applied.

### Tests (write before implementation)

- [ ] T054 [TEST] [US3] Write Drasi CLI client test: version check passes 0.10.0, fails on 0.9.2 with ERR_DRASI_CLI_VERSION, ERR_DRASI_CLI_NOT_FOUND when binary absent — `internal/drasi/client_test.go`
- [ ] T055 [TEST] [P] [US3] Write apply test: success stdout captured, failed apply propagates stderr as error, context cancellation on timeout — `internal/drasi/apply_test.go`
- [ ] T056 [TEST] [P] [US3] Write wait test: Online immediately returns nil, Online after 2 polls returns nil, exceeds timeout returns ERR_COMPONENT_TIMEOUT — `internal/drasi/wait_test.go`
- [ ] T057 [TEST] [P] [US3] Write delete test: successful delete returns nil, not-found case is non-fatal, other errors propagate — `internal/drasi/delete_test.go`
- [ ] T058 [TEST] [P] [US3] Write list test: parse tabular output to []ComponentSummary, empty list returns empty slice, parse error surfaced — `internal/drasi/list_test.go`
- [ ] T059 [TEST] [P] [US3] Write describe test: parse component detail struct including status and error reason, component not found returns typed error — `internal/drasi/describe_test.go`
- [ ] T060 [TEST] [US3] Write deployment diff test: unchanged hash → NoOp; changed hash → DeleteThenApply; missing key in state → Create; key format DRASI_HASH_CONTINUOUSQUERY_my-id — `internal/deployment/diff_test.go`
- [ ] T061 [TEST] [P] [US3] Write deployment order test: SortForDeploy produces sources→queries→reactions order; SortForDelete produces exact reverse — `internal/deployment/order_test.go`
- [ ] T062 [TEST] [P] [US3] Write state test: ReadHash returns empty string for missing key; WriteHash persists; round-trip preserves hash — `internal/deployment/state_test.go`
- [ ] T063 [TEST] [P] [US3] Write timeout test: per-component 5min deadline, total 15min deadline, ERR_COMPONENT_TIMEOUT includes component identity — `internal/deployment/timeout_test.go`
- [ ] T064 [TEST] [US3] Write engine test: happy-path all Create, hash written after each success; partial failure leaves already-written hashes intact; dry-run makes no FS or subprocess calls — `internal/deployment/engine_test.go`
- [ ] T065 [TEST] [P] [US3] Write KV translator test: SecretRef replaced with K8s Secret ref; plain string value passes through; missing KV secret returns typed error — `internal/keyvault/translator_test.go`
- [ ] T066 [TEST] [US3] Write deploy command test: validation gate blocks deploy on invalid config; --dry-run emits plan without side effects; --environment wires correct env state; ERR_DAPR_NOT_READY when Dapr absent — `cmd/deploy_test.go`
- [ ] T067 [TEST] [P] [US3] Write teardown command test: missing --force exits 2 with ERR_FORCE_REQUIRED; --infrastructure flag accepted; components deleted in reverse order reactions→queries→sources — `cmd/teardown_test.go`

### Implementation

- [ ] T068 [US3] Create Drasi CLI client: `Client` struct; `CheckVersion()` using semver parse, ERR_DRASI_CLI_NOT_FOUND when binary not on PATH; `RunCommand(ctx, args...)` subprocess wrapper capturing stdout/stderr — `internal/drasi/client.go`
- [ ] T069 [P] [US3] Create apply wrapper: `ApplyFile(ctx context.Context, path string) error` → `drasi apply -f <path>` — `internal/drasi/apply.go`
- [ ] T070 [P] [US3] Create wait wrapper: `WaitOnline(ctx context.Context, kind, id string, timeoutSec int) error` → `drasi wait <kind> <id> --timeout <n>`; kind is always canonical lowercase (e.g. `continuousquery`) — `internal/drasi/wait.go`
- [ ] T071 [P] [US3] Create delete wrapper: `DeleteComponent(ctx context.Context, kind, id string) error` → `drasi delete <kind> <id>` — `internal/drasi/delete.go`
- [ ] T072 [P] [US3] Create list wrapper: `ListComponents(ctx context.Context, kind string) ([]ComponentSummary, error)` → `drasi list <kind>`, parse tabular stdout — `internal/drasi/list.go`
- [ ] T073 [P] [US3] Create describe wrapper: `DescribeComponent(ctx context.Context, kind, id string) (*ComponentDetail, error)` → `drasi describe <kind> <id>`, parse stdout to struct with Status and ErrorReason — `internal/drasi/describe.go`
- [ ] T074 [US3] Create deployment state store: `ReadHash(ctx context.Context, azdClient AzdClient, envName, kind, id string) (string, error)`; `WriteHash(ctx, azdClient, envName, kind, id, hash string) error`; key format `DRASI_HASH_<KIND>_<ID>` (uppercase) — `internal/deployment/state.go`
  > **API**: Use the `azdext` gRPC Environment service — NOT direct file I/O on `.azure/<env>/.env`. The correct calls are:
  >
  > - `azdClient.Environment().GetEnvironmentValue(ctx, &azdext.GetEnvRequest{EnvName: envName, Key: key})` — returns `KeyValueResponse`; treat a not-found response as an empty string (hash absent = component not yet deployed)
  > - `azdClient.Environment().SetEnvironmentValue(ctx, &azdext.SetEnvRequest{EnvName: envName, Key: key, Value: hash})`
  >   The `AzdClient` must be injected as an interface dependency (not constructed inside `state.go`) to keep the package testable without a live gRPC connection.
- [ ] T075 [P] [US3] Create deployment diff: `ComputeHash(component any) string` using SHA-256 of canonical YAML (keys sorted); `BuildPlan(manifest ResolvedManifest, state StateReader) DeploymentPlan` comparing hashes — `internal/deployment/diff.go`
- [ ] T076 [P] [US3] Create deployment order: `SortForDeploy(plan DeploymentPlan)` → sources, queries, reactions, middleware; `SortForDelete(plan)` → reverse — `internal/deployment/order.go`
- [ ] T077 [P] [US3] Create deployment timeout: `PerComponentTimeout = 5 * time.Minute`; `TotalDeployTimeout = 15 * time.Minute`; `WithComponentDeadline(ctx, componentID)` → derived context + ERR_COMPONENT_TIMEOUT helper — `internal/deployment/timeout.go`
- [ ] T078 [US3] Create deployment engine: `Deploy(ctx, plan, drasiClient, stateStore) DeploymentResult`; ordered execution of Create/DeleteThenApply/NoOp actions; write hash after each successful apply+wait; preserve partial-failure state on next run by not writing hash for failed components — `internal/deployment/engine.go`
- [ ] T079 [US3] Create Key Vault client: `GetSecret(ctx context.Context, vaultName, secretName string) (string, error)` using `azsecrets.NewClient` + `DefaultAzureCredential` — `internal/keyvault/client.go`
  > **[VERIFY]** Before coding: confirm current `azsecrets.NewClient` signature at `https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets`. Expected: `NewClient(vaultURL string, credential azcore.TokenCredential, options *ClientOptions) (*Client, error)`. The vault URL is `https://<vaultName>.vault.azure.net/` — `vaultName` is NOT a GetSecret argument. `GetSecret` signature: `(ctx, name, version string, options *GetSecretOptions) (GetSecretResponse, error)` where **version = "" for latest**. Confirm both signatures match the pinned module version before writing any call sites.
  > **[VERIFY]** README.md and MIGRATION.md in the azsecrets module have historically shown contradictory constructor signatures. Trust only `pkg.go.dev` for the version in go.mod.
- [ ] T080 [P] [US3] Create KV→K8s Secret translator: `TranslateRefs(ctx, manifest ResolvedManifest, kvClient KVClient, k8sClient K8sClient, namespace string) (ResolvedManifest, error)`; walk all `Value` fields; for each `SecretRef` fetch KV secret, write K8s Secret (`drasi-secret-<vaultName>-<secretName>`), replace with K8s Secret ref — `internal/keyvault/translator.go`
- [ ] T081 [US3] Implement `cmd/deploy.go`: validate gate (exit 1 if fails) → KV translation → BuildPlan → dry-run branch (print plan, exit 0) or execute → format DeploymentResult with per-component status — `cmd/deploy.go`
  > **featureFlags gate (spec FR-007 / spec line 60)**: After BuildPlan and before execution, filter out any component marked `experimental: true` when `manifest.FeatureFlags.EnableExperimentalQueries == false`. Emit a `WARN` log for each skipped component (`slog.Warn("skipping experimental component", "id", id, "kind", kind)`). This filter must apply in both dry-run and live-execute paths.
- [ ] T082 [P] [US3] Implement `cmd/teardown.go`: require `--force` (exit 2 with ERR_FORCE_REQUIRED if absent); call drasi delete in reverse order reactions→queries→sources; `--infrastructure` flag triggers Bicep teardown — `cmd/teardown.go`

**Checkpoint**: `azd drasi deploy` idempotently deploys 3 components; re-deploy with unchanged YAML is all-NoOp; change one query file and re-deploy — only that query is DeleteThenApply; partial failure on component 2 leaves component 1 hash written for recovery.

---

## Phase 7: User Story 4 — Operate and Monitor (Priority: P4)

**Goal**: `status`, `logs`, `diagnose` give complete operational visibility of live Drasi
components — per-component health, streaming logs, and 5-check diagnostics.

**Independent Test**: Deploy a working config, artificially degrade one source by removing its
connection string KV secret. Run `azd drasi status` — confirm the query depending on that source
shows Failed with error reason; exit code 1. Run `azd drasi diagnose` — Key Vault check FAIL.

### Tests (write before implementation)

- [ ] T083 [TEST] [US4] Write status command test: all Online returns exit 0 health table; one Failed returns exit 1 with error reason + remediation hint; --output json valid schema — `cmd/status_test.go`
- [ ] T084 [TEST] [P] [US4] Write logs command test: --component filter accepted, --kind filter accepted, --tail n flag wired, --follow flag accepted, pod not found exits 1 — `cmd/logs_test.go`
- [ ] T085 [TEST] [P] [US4] Write diagnose command test: all PASS exits 0; single FAIL exits 1; JSON output lists all 5 checks with PASS/FAIL; --output json valid schema — `cmd/diagnose_test.go`
- [ ] T086 [TEST] [P] [US4] Write tracer and metrics provider tests:
  - `internal/observability/tracer_test.go`: no-op provider initialised when `APPLICATIONINSIGHTS_CONNECTION_STRING` env var absent; OTLP exporter created when present
  - `internal/observability/metrics_test.go`: no-op meter when env var absent; meter with correct instrument names when present; counter names match spec (`drasi.components.deployed`, `drasi.deploy.errors`, `drasi.deploy.duration_seconds`)
    > **[TEST] coverage**: both `tracer.go` and `metrics.go` (T090, T091) must have test coverage before implementation per TDD requirement

### Implementation

- [ ] T087 [US4] Implement `cmd/status.go`: call `drasi list` for each kind (source, continuousquery, reaction), call `drasi describe` for each; format health table (kind, id, status, age, error reason); emit remediation hint for Failed/Pending components; exit 1 if any component non-Online — `cmd/status.go`
- [ ] T088 [P] [US4] Implement `cmd/logs.go`: shell out to `kubectl logs` for Drasi pods in drasiNamespace; filter by `--component` (pod label match) or `--kind`; `--tail n` passes to kubectl; `--follow` passes -f to kubectl — `cmd/logs.go`
- [ ] T089 [P] [US4] Implement `cmd/diagnose.go`: run 5 checks sequentially — (1) AKS API server reachable via kubectl, (2) drasi-api pod running in drasiNamespace, (3) Dapr sidecar injector pod running, (4) Key Vault GET secret via DefaultAzureCredential, (5) Log Analytics last-5-min ingestion present; accumulate PASS/FAIL; exit 1 if any fail — `cmd/diagnose.go`
- [ ] T090 [P] [US4] Create OTel trace provider: `NewTracer(ctx) trace.Tracer`; OTLP exporter → Azure Monitor endpoint from APPLICATIONINSIGHTS_CONNECTION_STRING; no-op provider when env var absent — `internal/observability/tracer.go`
  > **[VERIFY] OTLP → Azure Monitor**: Confirm the correct Go package and endpoint derivation before implementing. The Application Insights connection string encodes an `IngestionEndpoint` field. For OTel → Azure Monitor in Go, two options exist:
  >
  > 1. `github.com/microsoft/ApplicationInsights-Go` + Azure Monitor bridge (legacy, not recommended)
  > 2. `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp` targeting the OTLP endpoint extracted from the connection string (`<IngestionEndpoint>/v2.1/track`) with `x-ms-application-insights-...` headers
  >    Confirm the current recommended approach from Azure Monitor docs before implementing. The auto-detect-via-env-var pattern (`APPLICATIONINSIGHTS_CONNECTION_STRING`) may only work with the OpenTelemetry Azure Monitor Distro (`go.opentelemetry.io/contrib/exporters/autoexport`), not the raw OTLP exporter.
- [ ] T091 [P] [US4] Create OTel metrics provider: `NewMeter(ctx) metric.Meter`; counters: `drasi.components.deployed`, `drasi.deploy.errors`, `drasi.deploy.duration_seconds` — `internal/observability/metrics.go`
  > **[TEST]**: Tests are in T086 (`internal/observability/metrics_test.go`). Implement T091 only after T086 tests exist and fail.

**Checkpoint**: `azd drasi status` reflects live component state within 30s of change; `azd drasi diagnose` reports all 5 checks; `azd drasi logs --component <id>` streams only that component's log lines.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: CI/CD pipelines, release automation, and end-user documentation.

- [ ] T092 [P] Create CI workflow: trigger push + PR; jobs: `go build ./...`, `go test ./... -race -coverprofile=coverage.out`, `golangci-lint run`, coverage gate (fail < 80%), `bicep build infra/main.bicep`, `yamllint drasi/`; matrix ubuntu-latest + windows-latest — `.github/workflows/ci.yml`
- [ ] T093 [P] Create release workflow: trigger `v*` tags; run CI checks; cross-compile 4 targets via build.sh + build.ps1; `gh release create` with binary assets (windows.zip, linux-amd64.tar.gz, darwin-amd64.tar.gz, darwin-arm64.tar.gz); PR to update registry.yaml in extension registry repo — `.github/workflows/release.yml`
- [ ] T094 Create README.md: prerequisites table (azd ≥1.10.0, drasi ≥0.10.0, go 1.22, kubectl, bicep), `azd extension install` command, 6-step quick start referencing quickstart.md, `azd drasi <command> --help` reference table, troubleshooting link — `README.md`
- [ ] T095 [P] Create architecture doc: component diagram (extension binary → azdext gRPC → azd host; extension → drasi CLI subprocess → drasi-platform on AKS); provision flow; deploy flow with KV translation — `docs/architecture.md`
- [ ] T096 [P] Create configuration reference: full YAML schema for all 5 entity types (DrasiManifest, Source, ContinuousQuery, Reaction, Middleware); SecretRef syntax; environment overlay format; feature flags semantics; example for each entity type — `docs/configuration-reference.md`
- [ ] T097 [P] Create troubleshooting guide: table of all ERR\_\* error codes with trigger conditions and remediation steps; common failure scenarios (CLI not found, timeout, KV lockout, partial deploy); diagnostic flow linking to `azd drasi diagnose` — `docs/troubleshooting.md`
- [ ] T098 Run quickstart.md end-to-end validation: follow quickstart.md Steps 1–10 in a clean environment; update any step text or command that has drifted from actual behaviour — `specs/001-azd-drasi-extension/quickstart.md`

- [ ] T099 Implement `cmd/upgrade.go` — FR-010 requires a valid registered command stub; the command MUST:
  - Register as `azd drasi upgrade` on the root Cobra command
  - Return `ERR_NOT_IMPLEMENTED` with the message `"upgrade is planned for a future release \u2014 see https://github.com/azure/azd.extensions.drasi/issues for the roadmap"`
  - Exit with code 2 (same as other not-implemented paths)
  - Be annotated in code with `// intentional stub \u2014 FR-010; remove when upgrade logic is implemented`
  - Have one [TEST] confirming exit code 2 + ERR_NOT_IMPLEMENTED in output — `cmd/upgrade_test.go`

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 1 (Setup) — no dependencies; start immediately
    │
    ▼
Phase 2 (Foundational) — BLOCKS all user stories
    │
    ├──► Phase 3 (US5 Validate) — can start after Phase 2
    │        │
    │        └──► Phase 6 (US3 Deploy) — validate gate required for deploy
    │
    ├──► Phase 4 (US1 Scaffold) — can start after Phase 2 (independent of US5)
    │
    ├──► Phase 5 (US2 Provision) — can start after Phase 2 (independent of US5/US1)
    │
    └──► Phase 6 (US3 Deploy) — requires Phase 3 (Drasi CLI client + deployment engine)
             │
             └──► Phase 7 (US4 Operate) — reuses Drasi CLI client from Phase 6
                      │
                      └──► Phase 8 (Polish) — all stories complete
```

### User Story Dependencies

| Story              | Depends On                                 | Can Start After  |
| ------------------ | ------------------------------------------ | ---------------- |
| US5 (P5) Validate  | Phase 2 Foundational                       | Phase 2 complete |
| US1 (P1) Scaffold  | Phase 2 Foundational                       | Phase 2 complete |
| US2 (P2) Provision | Phase 2 Foundational                       | Phase 2 complete |
| US3 (P3) Deploy    | Phase 2 + US5 (validate gate)              | Phase 3 complete |
| US4 (P4) Operate   | Phase 2 + US3 (drasi list/describe client) | Phase 6 complete |
| Phase 8 Polish     | All stories complete                       | Phase 7 complete |

### Within Each Phase

1. [TEST] tasks MUST be written first (TDD — Constitution Principle VIII)
2. Tests MUST fail before implementation starts
3. [P] tasks within a phase can run concurrently
4. Leaf implementation tasks (engine.go, cmd/\*.go) depend on their package siblings

---

## Summary

| Phase                  | Story | Tasks                   | [P] Tasks | Test Tasks |
| ---------------------- | ----- | ----------------------- | --------- | ---------- |
| Phase 1: Setup         | —     | T001–T007               | 6         | 0          |
| Phase 2: Foundational  | —     | T008–T015 + T013↳listen | 4         | 3          |
| Phase 3: US5 Validate  | P5    | T016–T031               | 12        | 7          |
| Phase 4: US1 Scaffold  | P1    | T032–T040               | 7         | 2          |
| Phase 5: US2 Provision | P2    | T041–T053               | 11        | 1          |
| Phase 6: US3 Deploy    | P3    | T054–T082               | 22        | 14         |
| Phase 7: US4 Operate   | P4    | T083–T091               | 7         | 4          |
| Phase 8: Polish        | —     | T092–T099               | 6         | 1          |
| **Total**              |       | **100**                 | **75**    | **32**     |

**MVP Scope** (US1 P1 + prerequisites): Phases 1, 2, and 4 = **T001–T015 + T032–T040** (24 tasks).
After MVP: US5 validate (Phase 3) is the next highest-value increment before deploy (Phase 6 requires it).

### Parallel Execution — Phase 6 (US3 Deploy, largest phase)

```bash
# Write all Drasi client tests in parallel (T054–T059)
T054, T055, T056, T057, T058, T059   # 6 parallel test files

# Write all deployment tests in parallel (T060–T067)
T060, T061, T062, T063, T064, T065, T066, T067   # 8 parallel test files

# Client implementation (after T054–T059 all written + failing)
T068  # client.go (sequential — others depend on it)
T069, T070, T071, T072, T073   # parallel after T068

# Deployment implementation (after T060–T064 all failing)
T074  # state.go (sequential — engine depends on it)
T075, T076, T077   # parallel after T074
T078  # engine.go (sequential — depends on T075–T077)
T079, T080   # keyvault parallel after T078
T081, T082   # cmd files parallel after T079+T080
```
