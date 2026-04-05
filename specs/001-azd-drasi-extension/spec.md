# Feature Specification: azd-drasi Extension

**Feature Branch**: `001-azd-drasi-extension`
**Created**: 2026-04-04
**Status**: Draft
**Input**: Build an azd extension ("azd-drasi") that enables developers and platform engineers to scaffold, provision, deploy, and operate Drasi-based change-driven data processing applications using native azd workflows.

## User Scenarios & Testing _(mandatory)_

### User Story 1 — Scaffold a Drasi Project (Priority: P1)

A developer starts a new change-driven application. They run `azd drasi init` and receive a complete, runnable project scaffold: an azd project structure, a `drasi/` configuration directory with a root manifest (`drasi.yaml`) and placeholder YAML files for sources, continuous queries, and reactions, plus Bicep infrastructure templates targeting AKS. The developer has no prior Drasi knowledge; they understand what to fill in from inline comments and the README quickstart.

**Why this priority**: Without scaffolding, no other story is reachable. This is the entry point that converts a blank directory into a valid Drasi project. All subsequent stories depend on the structure it produces.

**Independent Test**: Run `azd drasi init` in an empty directory, inspect the output tree, confirm all declared files exist and are schema-valid, confirm the README quickstart instructions are complete and accurate. No Azure subscription needed.

**Acceptance Scenarios**:

1. **Given** an empty directory with no azd project, **When** the user runs `azd drasi init`, **Then** a complete directory structure is created including `azure.yaml`, `drasi/drasi.yaml`, `drasi/sources/`, `drasi/queries/`, `drasi/reactions/`, `drasi/environments/dev.yaml`, and `infra/` Bicep templates; the command exits with code 0 and a summary of created files.
2. **Given** an existing azd project, **When** the user runs `azd drasi init`, **Then** only missing Drasi-specific files are added; existing files are not overwritten; the command produces a diff summary of additions.
3. **Given** the user passes `--template cosmos-change-feed`, **When** `azd drasi init --template cosmos-change-feed` runs, **Then** the scaffold is pre-populated with a working Cosmos DB source, a sample continuous query, and a Dapr pub/sub reaction with inline comments explaining each field.
4. **Given** the user passes `--output json`, **When** `azd drasi init` completes, **Then** all file paths created are emitted as a JSON array to stdout; no human-readable text is mixed in.
5. **Given** `azd drasi init` is re-run on an already-initialized project, **When** no `--force` flag is passed, **Then** no files are modified and the command exits with code 0 indicating nothing changed (idempotent).

---

### User Story 2 — Provision Azure Infrastructure for Drasi (Priority: P2)

A platform engineer runs `azd drasi provision` to deploy all Azure resources required to host Drasi: an AKS cluster, a Log Analytics workspace, Azure Key Vault, managed identity assignments, and optional Cosmos DB and Event Hub resources. The command completes without any manual portal intervention. Re-running it on an existing environment updates resources to desired state rather than duplicating them.

**Why this priority**: Infrastructure must exist before Drasi can be deployed. This story is the second blocking prerequisite. It delivers standalone value: a provisioned, secured AKS environment ready for any Drasi workload.

**Independent Test**: Run `azd drasi provision` against a fresh Azure subscription environment using a validated `drasi/environments/dev.yaml`. Confirm all resources exist in the resource group, all managed identity role assignments are present, and no secrets appear in any template or deployment output.

**Acceptance Scenarios**:

1. **Given** a scaffolded Drasi project and valid Azure credentials, **When** `azd drasi provision` runs, **Then** an AKS cluster, Log Analytics workspace, and Key Vault are deployed in the target resource group; all resources carry mandatory tags (`environment`, `project`, `component`, `managed-by=azd`); command exits code 0.
2. **Given** an already-provisioned environment, **When** `azd drasi provision` is re-run without changes, **Then** no new resources are created, no existing resources are modified, and the command exits code 0 (idempotent convergence).
3. **Given** the `dev.yaml` environment overlay specifies `aks.nodeCount: 1`, **When** provisioning completes, **Then** the AKS cluster has exactly one node.
4. **Given** provisioning fails mid-way due to a quota error, **When** the error occurs, **Then** the command exits with a non-zero code, emits a structured error message identifying the quota limit, and provides a link to the Azure quota increase documentation.
5. **Given** the user has not logged in via `azd auth login`, **When** `azd drasi provision` runs, **Then** the command fails fast with exit code 1, an error code `ERR_NO_AUTH`, and a remediation instruction "Run `azd auth login` to authenticate."
6. **Given** `--output json` is passed, **When** provisioning completes, **Then** all deployed resource IDs and endpoints are emitted as structured JSON to stdout.

---

### User Story 3 — Deploy Drasi Sources, Continuous Queries, and Reactions (Priority: P3)

A developer runs `azd drasi deploy` to push their Drasi configuration (sources, continuous queries, reactions) onto the provisioned AKS cluster. The command validates the configuration files, resolves cross-file references, and applies them to the Drasi runtime. Re-deploying after a change updates only the modified components. The developer can confirm each query is running and each reaction is bound before the command exits.

**Why this priority**: This is the core value delivery of the extension — the moment a continuous query begins tracking real data changes. It is the primary measurable outcome of the 30-minute end-to-end success criterion.

**Independent Test**: From a provisioned environment, run `azd drasi deploy` with a sample configuration (Cosmos DB source, high-value-orders query, Dapr pub/sub reaction). Confirm Drasi reports all components as `Online`, the query is reachable via the Drasi API, and re-running deploy produces no duplicate registrations.

**Acceptance Scenarios**:

1. **Given** a provisioned AKS cluster with Drasi runtime and a valid `drasi/` configuration, **When** `azd drasi deploy` runs, **Then** all sources, continuous queries, and reactions defined in `drasi.yaml` are applied to the runtime; command exits code 0 with a per-component status summary.
2. **Given** a configuration that includes a query referencing a source ID that does not exist in `sources/*.yaml`, **When** `azd drasi deploy` runs, **Then** the command fails before applying any resources; exit code 1; error message identifies the unresolved reference and the file/line it appears in.
3. **Given** an already-deployed configuration where only one query file is changed, **When** `azd drasi deploy` runs, **Then** only the changed query is updated on the runtime; unchanged sources and reactions are not reapplied.
4. **Given** `featureFlags.enableExperimentalQueries: false` in `drasi.yaml`, **When** `azd drasi deploy` runs, **Then** any query marked `experimental: true` is skipped; a warning is emitted but the command succeeds.
5. **Given** `--dry-run` is passed, **When** `azd drasi deploy --dry-run` runs, **Then** the command resolves all references and validates schemas but makes no changes to the runtime; it emits a structured plan of what would be applied.
6. **Given** a reaction of type `dapr-pubsub` deploying to an environment where Dapr is not installed, **When** `azd drasi deploy` runs, **Then** the command fails with error code `ERR_DAPR_NOT_READY`, a human-readable description, and the remediation step "Ensure Dapr is installed on the AKS cluster. Run `azd drasi provision` to install it automatically."

---

### User Story 4 — Operate and Monitor Running Drasi Components (Priority: P4)

An SRE team member uses `azd drasi status`, `azd drasi logs`, and `azd drasi diagnose` to monitor the health of deployed Drasi components in production. `status` shows a per-component health summary with actionable hints for degraded components. `logs` streams structured log output from the Drasi control plane and selected query or reaction components. `diagnose` runs a pre-defined set of health checks and emits a structured report.

**Why this priority**: Operational visibility is required before users can trust the extension in production. Without it, failures are opaque and recovery is manual. This story focuses on day-2 operations.

**Independent Test**: Deploy a working Drasi configuration, then artificially degrade one component (e.g., remove a source connection string). Confirm `azd drasi status` reports the component as degraded with a specific error and remediation hint. Confirm `azd drasi logs --query high-value-orders` streams only logs for that query component.

**Acceptance Scenarios**:

1. **Given** all Drasi components are healthy, **When** `azd drasi status` runs, **Then** a table is displayed showing each source, query, and reaction with status `Online`; command exits code 0.
2. **Given** one continuous query is in a `Failed` state, **When** `azd drasi status` runs, **Then** the failed component is highlighted; a human-readable error reason is shown; a remediation hint ("Check source connectivity for `<sourceId>`") is displayed; command exits code 1.
3. **Given** `azd drasi logs --component <queryId>` is run, **When** executing, **Then** structured log lines from the named query component stream to stdout in real time until Ctrl-C; each line includes timestamp, component ID, severity, and message.
4. **Given** `azd drasi diagnose` runs, **When** all checks pass, **Then** a structured report is emitted listing each check (AKS connectivity, Drasi API reachability, Dapr sidecar status, Secrets Store CSI sync health for generated `SecretProviderClass` objects) with `PASS`/`FAIL` per check; exit code 0 on all pass, 1 if any fail.
5. **Given** `--output json` is passed to any operational command, **When** the command completes, **Then** the output is valid JSON; no human-readable formatting is mixed in.

---

### User Story 5 — Validate Drasi Configuration Before Deployment (Priority: P5)

A developer runs `azd drasi validate` as a pre-deployment gate in CI/CD or locally before committing. The command schema-validates all files under `drasi/`, resolves all cross-file ID references, detects circular dependencies between queries, and reports all errors with file paths and line numbers. It requires no Azure connectivity; it runs entirely against the local configuration files.

**Why this priority**: Validation without deployment is an independent, offline-safe capability that gates both local developer workflow and CI/CD pipelines. It enables the test-first quality principle from the constitution (Principle VIII) at the configuration layer.

**Independent Test**: Create a `drasi/queries/` file that references a non-existent source ID. Run `azd drasi validate`. Confirm exit code 1, confirm the error message identifies the file, line, and unresolved reference. Fix the reference and re-run; confirm exit code 0.

**Acceptance Scenarios**:

1. **Given** a fully valid `drasi/` configuration directory, **When** `azd drasi validate` runs, **Then** all files are parsed and validated against their schemas; all cross-file references are resolved; exit code 0; a summary lists all validated components with a count per type.
2. **Given** a query file references `source: non-existent-id`, **When** `azd drasi validate` runs, **Then** exit code 1; error output identifies the query file, the invalid source reference, and lists all available source IDs as a suggestion.
3. **Given** a YAML file contains an invalid field (e.g., `type` instead of `kind` for a source), **When** `azd drasi validate` runs, **Then** exit code 1; error output includes the file path, line number, field name, expected values, and a link to the schema documentation.
4. **Given** `--output json` is passed, **When** `azd drasi validate` completes (pass or fail), **Then** the result is a JSON object with `valid: true/false`, `errors: []`, `warnings: []`, and a `components` summary; exit code reflects validity.
5. **Given** `azd drasi validate` is invoked with no `drasi.yaml` present, **When** executing, **Then** exit code 1; error `ERR_NO_MANIFEST`; remediation step "Run `azd drasi init` to scaffold a Drasi project."

---

### Edge Cases

- Re-running `azd drasi deploy` while a previous deployment is still in progress: the command must detect the in-progress state and either wait (with timeout) or fail with `ERR_DEPLOY_IN_PROGRESS` and a remediation step. It must not corrupt the running deployment.
- Circular dependency in query joins (Query A joins Query B which joins Query A): `azd drasi validate` must detect and report this as a fatal error before any deployment attempt.
- Environment overlay file (e.g., `environments/prod.yaml`) defines a value for a parameter not declared in `drasi.yaml`: the command must emit a warning but not fail; undeclared override parameters must be surfaced in `--output json` for auditability.
- Partial failure during `azd drasi deploy` (first three components apply, fourth fails): the command must report which components were applied and which failed; on next run, already-applied components must not be duplicated (idempotent recovery).
- `azd drasi provision` runs against a subscription where the target resource group already contains resources not tagged `managed-by=azd`: the command must not modify or delete unmanaged resources; it must emit a warning listing unmanaged resources found.
- A source YAML file references a Key Vault secret that does not exist: `azd drasi validate` must flag this as a warning (cannot verify online in offline mode), and `azd drasi deploy` must fail with a specific error identifying the missing secret reference.
- User runs `azd drasi teardown` without `--force`: the command must prompt for confirmation (interactive) or fail with `ERR_FORCE_REQUIRED` (non-interactive / CI), listing all resources that would be deleted.- Modifying an existing ContinuousQuery (e.g., changing the Cypher query string) then running `azd drasi deploy`: because Drasi resources cannot be updated in place, the deploy must delete the existing query then apply the new version. If delete succeeds but apply fails, the system is left with no query for that ID — the extension must detect this partial state on the next deploy run and attempt the apply again without duplicating the delete.

---

## Requirements _(mandatory)_

### Functional Requirements

#### CLI Command Group

- **FR-001**: The extension MUST register a top-level `azd drasi` command group that does not conflict with any built-in azd commands.
- **FR-002**: The extension MUST implement the following command surface in v1.0: core commands `init`, `provision`, `deploy`, `status`, `logs`, `diagnose`, `validate`, `teardown`, plus an `upgrade` command stub as defined in FR-010.
- **FR-003**: Every command MUST accept `--output json` to emit machine-readable structured output to stdout; human-readable formatting MUST be suppressed when this flag is set.
- **FR-004**: Every command MUST accept `--environment <name>` to specify which environment overlay to activate; if omitted, the active azd environment is used.
- **FR-005**: Every command MUST exit with code 0 on success, code 1 on handled error, and code 2 on unexpected error.
- **FR-006**: Every error exit MUST emit to stderr: an error code string (e.g., `ERR_NO_AUTH`), a human-readable description, and a remediation step or documentation URL.
- **FR-007**: `azd drasi init` MUST create the full project scaffold described in the configuration model (FR-018 to FR-031) in an idempotent manner.
- **FR-008**: `azd drasi init` MUST support a `--template <name>` flag; available templates at minimum: `cosmos-change-feed`, `event-hub-routing`, `query-subscription`; each template MUST be pre-populated with working placeholder configuration and inline comments.
- **FR-009**: All commands MUST be re-entrant and safe to run in parallel CI/CD jobs without corrupting shared state. Mutating commands (`provision`, `deploy`, `teardown`) MUST use explicit concurrency control when they can target shared environment state; read-only commands (`status`, `logs`, `diagnose`, `validate`) MUST remain safe under concurrent invocation without requiring global locks.
- **FR-010**: `azd drasi upgrade` MUST be a valid command stub returning `ERR_NOT_IMPLEMENTED` with a planned-feature message in v1.0.

#### Configuration Model

- **FR-011**: The extension MUST use a multi-file configuration model rooted at `drasi/drasi.yaml` (root manifest).
- **FR-012**: The root manifest (`drasi/drasi.yaml`) MUST declare `includes` paths for sources, queries, reactions, and optional middleware; MUST declare named environment overlay paths; MUST declare `featureFlags`.
- **FR-013**: Source definitions MUST reside in `drasi/sources/*.yaml`; each file MUST define exactly one source with fields: `id`, `kind`, `properties`.
  Supported `kind` values in v1: `cosmosGremlin`, `eventHub`, `postgresql`, `sqlserver`.
  NOTE: EventGrid is a **ReactionProvider** in Drasi, not a SourceProvider; it MUST NOT appear as a source `kind`. EventGrid as a data ingestion path is out of scope for v1 sources.
- **FR-014**: Continuous query definitions MUST reside in `drasi/queries/*.yaml`; each file MUST define exactly one query with fields: `id`, `sources`, `queryLanguage` (MUST be explicitly set to `Cypher` or `GQL` — never rely on runtime defaults), `query` (the query string), `reactions`, and optional `joins`.
- **FR-015**: Reaction definitions MUST reside in `drasi/reactions/*.yaml`; each file MUST define exactly one reaction with fields: `id`, `kind`, `config`.
  Supported `kind` values mapped to default Drasi ReactionProviders in v1: `signalr`, `eventgrid`, `storedproc`, `storagequeue`, `debug`.
  Non-default kinds `dapr-pubsub` and `http` MUST be treated as custom reaction types requiring a custom ReactionProvider manifest; `azd drasi provision` MUST install the custom ReactionProvider before `azd drasi deploy` attempts to register reactions of those kinds.
- **FR-016**: Middleware definitions MUST reside in `drasi/middleware/*.yaml`; middleware is optional; each file defines one middleware component with fields: `id`, `kind`, `config`.
- **FR-017**: Environment overlay files MUST reside in `drasi/environments/<name>.yaml`; each overlay file MUST support parameter overrides, scaling configuration overrides, and feature flag overrides; environment overlays MUST NOT introduce new components. Environment overlays MUST NOT directly remove or suppress components — component suppression is achieved exclusively via feature flag overrides (setting a feature flag to `false` in the overlay); any component gated behind that feature flag is then skipped by `azd drasi deploy` for that environment.
- **FR-018**: All cross-file references (source IDs referenced by queries, reaction IDs referenced by queries) MUST use the `id` field as the reference key; no positional or implicit references are permitted.
- **FR-019**: The resolved in-memory model MUST be fully deterministic: given the same input files and environment name, the resolved model MUST be identical across runs.
- **FR-020**: Secrets in configuration files MUST be expressed only as Key Vault references using the pattern `{ kind: secret, vaultName: <name>, secretName: <key> }`; inline secret values are FORBIDDEN.
- **FR-021**: `azd drasi validate` MUST validate schema conformance, resolve all cross-file references, detect circular dependencies in query joins, and verify Key Vault reference syntax (not reachability) entirely offline.
- **FR-022**: `azd drasi validate` MUST report all errors with file path and line number; it MUST NOT stop at first error; it MUST accumulate and report all errors in a single pass.

#### Drasi Runtime Integration

- **FR-023**: `azd drasi provision` MUST deploy the Drasi runtime onto an AKS cluster using Bicep IaC; Drasi installation MUST include Dapr and the Drasi control plane components.
- **FR-024**: `azd drasi provision` MUST deploy: AKS cluster (with OIDC issuer and Workload Identity addon enabled — see FR-045), Log Analytics workspace, Azure Key Vault, a User-Assigned Managed Identity (UAMI) with the following least-privilege role assignments (per Principle IV, NON-NEGOTIABLE): **`Key Vault Secrets User`** on the Key Vault; **`Monitoring Metrics Publisher`** on the Log Analytics workspace (enables OpenTelemetry metric export per FR-038); **`AcrPull`** on the container registry (only when a private ACR hosts Drasi component images — omitted if public images are used). The conditional is controlled by a Bicep parameter `usePrivateAcr` (boolean, default `false`); when `true`, the role is assigned to the UAMI on the specified ACR and the `--registry <acr-login-server>` flag is passed to `drasi init`. No broader roles (e.g., `Contributor`, `Key Vault Secrets Officer`) are permitted. Optionally: Cosmos DB account, Event Hubs namespace, and Service Bus namespace when declared in the active configuration.
- **FR-045**: `azd drasi provision` MUST configure AKS Workload Identity for the Drasi runtime namespace. This requires: (1) the AKS cluster provisioned with `--enable-oidc-issuer` and `--enable-workload-identity`; (2) a `FederatedIdentityCredential` linking the UAMI to the Drasi Kubernetes ServiceAccount (`system:serviceaccount:drasi-system:drasi-resource-provider`) using the cluster's OIDC issuer URL and audience `api://AzureADTokenExchange`. The namespace `drasi-system` is the Drasi-installed default; a Bicep parameter `drasiNamespace` (default `drasi-system`) MUST be provided to support custom namespace deployments via `drasi init -n <namespace>`. (3) the Drasi Kubernetes ServiceAccount annotated with `azure.workload.identity/client-id: <UAMI_CLIENT_ID>`; (4) Drasi runtime pods labelled with `azure.workload.identity/use: "true"`. All four steps MUST be applied via Bicep; no manual `az identity federated-credential create` steps are permitted.
- **FR-025**: `azd drasi provision` MUST register all default Drasi source providers and reaction providers on the Drasi runtime upon first install (PostgreSQL, CosmosGremlin, SQLServer, EventHub, SignalR, EventGrid, StorageQueue, StoredProc, Debug).
- **FR-026**: `azd drasi deploy` MUST apply sources, continuous queries, reactions, and middleware to the Drasi runtime using the following semantics: **create** if the component does not exist; **delete-then-apply** if the component exists and its configuration has changed (Drasi resources cannot be updated in place); **no-op** if the component exists and configuration is unchanged. The extension MUST determine changed vs. unchanged state by comparing a content hash of the resolved YAML against the last-deployed hash stored in the **azd environment state file** (`.azure/<env>/` directory managed by azd). The hash store key format MUST be `DRASI_HASH_<componentKind>_<componentId>` (uppercase). If the azd state file is absent or the key is missing, the component MUST be treated as not yet deployed (create semantics).
- **FR-027**: `azd drasi deploy` MUST perform a pre-deploy validation pass (equivalent to `azd drasi validate`) and fail before applying any resources if validation fails.
- **FR-028**: `azd drasi deploy` MUST confirm each applied component reaches `Online` state before reporting success; it MUST emit a structured per-component status on completion. The per-component `Online` wait MUST time out after **5 minutes**; the total deploy operation MUST time out after **15 minutes**. On timeout the command MUST exit with code 1 and error code `ERR_COMPONENT_TIMEOUT`, identifying which component(s) failed to reach `Online` within budget.
- **FR-029**: `azd drasi deploy` MUST support `--dry-run` to preview changes without applying them.
- **FR-030**: `azd drasi teardown` MUST require explicit `--force` flag; it MUST delete all Drasi components and optionally delete provisioned Azure infrastructure when `--infrastructure` is also passed.
- **FR-042**: `azd drasi deploy` MUST apply Drasi components in dependency order: sources first, then continuous queries, then middleware, then reactions. This order reflects the Drasi data-flow topology: queries depend on sources; middleware sits between query emission and reaction execution; reactions depend on middleware being present. `azd drasi teardown` (and delete-then-apply within FR-026) MUST delete in reverse order: reactions first, then middleware, then continuous queries, then sources. Violating this order causes Drasi runtime errors due to unresolvable references.
- **FR-043**: At deploy time, the extension MUST translate Key Vault secret references (format: `{ kind: secret, vaultName: <name>, secretName: <key> }` from FR-020) into Kubernetes Secret references (`{ kind: Secret, name: <k8s-secret-name>, key: <key> }`) before applying Drasi resource YAML to the runtime. This translation MUST occur at deploy time, not at init or provision time. The extension MUST generate and apply the required Secrets Store CSI objects (for example, `SecretProviderClass` and synchronized Kubernetes Secret metadata) in the Drasi namespace so secret material is resolved in-cluster by AKS Workload Identity. The extension MUST NOT call Azure Key Vault data-plane APIs directly.
- **FR-044**: The extension manifest (`extension.yaml`) MUST declare: `namespace: drasi` (maps to the `azd drasi` command prefix); `capabilities: [custom-commands, lifecycle-events]`; OS-specific `executablePath` entries for `windows`, `linux`, and `darwin`. The extension binary MUST communicate with the azd host exclusively via gRPC (managed by the `azdext` SDK); writing to stdout inside lifecycle event handlers is FORBIDDEN as it corrupts the gRPC channel.
- **FR-046**: The extension MUST invoke the Drasi CLI (`drasi`) as a subprocess for install/bootstrap and Drasi runtime operations. The extension MUST NOT re-implement the Drasi API client. The only Drasi CLI commands used by the extension are:
  - `drasi init [flags]` — install/bootstrap the Drasi control plane onto the target Kubernetes context during `azd drasi provision`
  - `drasi apply -f <manifest-file>` — apply a single resolved resource YAML to the runtime (used for sources, queries, reactions, providers)
  - `drasi wait <component-kind> <component-id> --timeout <seconds>` — block until the component reaches `Online` state or the timeout elapses
  - `drasi delete <component-kind> <component-id>` — delete a named component from the runtime (used in delete-then-apply and teardown)
  - `drasi list <component-kind>` — enumerate registered components of a given kind (used by `azd drasi status`)
  - `drasi describe <component-kind> <component-id>` — retrieve detailed component status including error reason (used by `azd drasi status` and `azd drasi diagnose`)
    The minimum required Drasi CLI version is **0.10.0** (current latest `drasi-platform` release at spec authoring; update to the current stable release at implementation time if a newer version is available). The kind string used for ContinuousQuery resources in CLI invocations MUST be `continuousquery` (full canonical form); the `query` alias MUST NOT be used in extension code to avoid future CLI ambiguity. The extension MUST fail fast with error code `ERR_DRASI_CLI_NOT_FOUND` and remediation step "Install the Drasi CLI: https://drasi.io/getting-started" if `drasi` is not found on `PATH`. The extension MUST fail with `ERR_DRASI_CLI_VERSION` if the detected CLI version is below the minimum required version. **The extension MUST NOT implement automatic retry for any `drasi` CLI subcommand** (`drasi init`, `drasi apply`, `drasi delete`, `drasi list`, `drasi describe`): a non-zero exit code from any of these commands MUST cause an immediate fail-fast with error code `ERR_DRASI_CLI_ERROR`, wrapping the CLI's original exit code and stderr output. The `drasi wait` command handles internal polling for component readiness autonomously within its own timeout budget (per FR-028); the extension simply blocks on it until it exits and propagates any non-zero exit as `ERR_COMPONENT_TIMEOUT`.

#### Reactions and Integration

- **FR-031**: The `dapr-pubsub` reaction kind MUST be the recommended default for event emission; Dapr component YAML for the pub/sub binding MUST be generated by `azd drasi init` when this reaction kind is requested.
- **FR-032**: The `http` reaction kind MUST support configuring a target URL, HTTP method, and optional headers via the reaction `config` block; the URL MAY reference an environment variable or Key Vault reference.
- **FR-033**: All reaction kinds MUST support referencing secrets via Key Vault references in their `config` block (FR-020 pattern).
- **FR-034**: Reaction execution failures MUST be surfaced in `azd drasi status` and `azd drasi logs`; they MUST NOT silently discard failed events.

#### Local Development

- **FR-035**: The repository MUST include a Dev Container definition (`.devcontainer/devcontainer.json`) that installs azd, the Drasi CLI (minimum version **0.10.0**, per FR-046), Dapr CLI, Go toolchain, Bicep CLI, and kubectl.
- **FR-036**: `azd drasi init` MUST generate a local development compose configuration that starts an optional local Drasi topology and Dapr sidecar for offline query testing. The generated compose file MUST parameterize image coordinates using `DRASI_IMAGE_REGISTRY` (default `ghcr.io`) and `DRASI_VERSION` (default `0.10.0`), and it MUST mirror the published upstream Drasi local or Docker topology rather than assuming a single monolithic `drasi` image.
- **FR-037**: All commands MUST function identically when run inside the Dev Container versus the host machine (environment parity requirement).

#### Observability

- **FR-038**: `azd drasi provision` MUST configure all AKS-hosted Drasi components to export OpenTelemetry traces and metrics to the Log Analytics workspace.
- **FR-039**: `azd drasi status` MUST query the Drasi runtime API and present a per-component health table: component name, kind, status, last-seen timestamp, and error reason (if any). If no components are deployed (all `drasi list` results return empty), the command MUST display the message "No components deployed. Run `azd drasi deploy` to deploy Drasi components." and exit with code 0.
- **FR-040**: `azd drasi logs --component <id>` MUST stream structured log lines from the named Drasi component in real time; `--tail <n>` MUST limit output to the last n historical lines on initial connection before resuming the live stream. When `--tail` is omitted, the command MUST stream from the current position forward only (live tail; no historical lines are fetched on initial connection). When `--output json` is set, log output MUST be emitted as JSON objects (NDJSON for streaming mode), with no mixed human-formatted text.
- **FR-041**: `azd drasi diagnose` MUST execute the following checks and report PASS/FAIL per check: AKS cluster reachability, Drasi API reachability, Dapr sidecar status, Secrets Store CSI sync health for generated `SecretProviderClass` objects in the target namespace, and Log Analytics ingestion latency (last 5 minutes).

### Key Entities

- **Source**: Binding to an external data system (Cosmos DB, Event Hub, PostgreSQL, SQL Server). Attributes: `id`, `kind`, `properties` (connection details expressed as Key Vault references or environment variable bindings). A source is the entry point for change events into the Drasi pipeline.
- **ContinuousQuery**: A declarative Cypher or GQL query that evaluates continuously against incoming changes from one or more sources. Attributes: `id`, `queryLanguage` (MUST be `Cypher` or `GQL` — explicit, never default), `sources` (list of source IDs with optional node/label filters), `query` (the query string), `joins` (optional cross-source join definitions using explicit `keys:` mappings — implicit joins are NOT supported by Drasi), `reactions` (list of reaction IDs to trigger on result changes), `autoStart` (boolean). A continuous query maintains stateful result sets and emits deltas when results change.
- **Reaction**: An action executed when a continuous query result changes. Attributes: `id`, `kind` (reaction type: `dapr-pubsub`, `http`, `signalr`, `eventgrid`, `storedproc`, `debug`), `config` (kind-specific configuration). Reactions are reusable and can be referenced by multiple queries.
- **Middleware**: An optional processing/enrichment component placed between query result emission and reaction execution. Attributes: `id`, `kind`, `config`. Middleware is stateless; it transforms or filters change events without altering query semantics.
- **DrasiManifest**: The root configuration file (`drasi/drasi.yaml`) that declares all `includes` glob patterns, named environment overlays, and feature flags. It is the single entry point consumed by all `azd drasi` commands.
- **Environment**: A named deployment context (e.g., `dev`, `staging`, `prod`) with an associated override file (`drasi/environments/<name>.yaml`). Environment files declare parameter overrides, scaling settings, and feature flag overrides; they do not define new components.
- **SourceProvider**: A Drasi runtime registration for a source adapter type (e.g., `PostgreSQL`, `CosmosGremlin`). Source providers MUST be installed on the Drasi runtime before sources of that kind can be registered.
- **ReactionProvider**: A Drasi runtime registration for a reaction handler type. Reaction providers MUST be installed before reactions of that kind can be registered.

---

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: A developer with no prior Drasi knowledge can scaffold a project, provision infrastructure, deploy a working continuous query, and observe it reacting to a data change — end to end — in under 30 minutes, following only the README quickstart.
- **SC-002**: Re-running any `azd drasi` command on an environment that is already in the desired state produces no resource changes and exits with code 0; confirmed by running the full cycle twice and comparing resource state before and after the second run.
- **SC-003**: Every command failure produces a human-readable error message and a remediation step within 2 seconds of failure detection; no failure exits silently or with a generic "something went wrong" message.
- **SC-004**: Zero manual Azure portal steps are required at any point in the provision → deploy → operate lifecycle; verified by completing the full lifecycle with no browser or portal access.
- **SC-005**: `azd drasi provision` completes a baseline AKS + Drasi environment (no optional services) in under 15 minutes from first run against a fresh Azure subscription.
- **SC-006**: `azd drasi validate` detects 100% of syntactically invalid files, missing cross-file references, and circular dependencies in a test configuration corpus (defined during the test suite build) and exits with a non-zero code for each.
- **SC-007**: A CI/CD pipeline (GitHub Actions) that runs `azd drasi provision` → `azd drasi deploy` → `azd drasi validate` against a pull request completes in under 20 minutes and requires no human interaction.
- **SC-008**: `azd drasi status` reflects the correct health state of all deployed components within 30 seconds of a component state change in the Drasi runtime.

---

## Assumptions

- **Hosting model**: AKS is the only supported runtime for Drasi components. The user request mentioned Azure Container Apps in one section; this spec follows the constitution (Principle VI, NON-NEGOTIABLE) and the user's primary specification (Section 4.2) which both designate AKS. Container Apps hosting is a future extension requiring an ADR.
- **Drasi runtime stability**: The Drasi project API (`drasi apply`, `drasi wait`, Drasi Kubernetes CRDs) is stable and versioned. The extension targets the **latest stable release of `drasi-platform` at implementation time**; the exact version is pinned in `extension.yaml` and documented in the README as the minimum tested version. Breaking API changes in Drasi itself require a MAJOR version bump in this extension (Principle IX).
- **Drasi CLI coexistence**: The `drasi` CLI (minimum version **0.10.0**) is available alongside `azd` in the Dev Container. The azd extension invokes Drasi CLI commands for bootstrap and runtime operations (`init`, `apply`, `wait`, `delete`, `list`, `describe`) as specified in FR-046; it does not re-implement the Drasi API client from scratch. The CLI binary is resolved from `PATH` at runtime; absence causes a fail-fast `ERR_DRASI_CLI_NOT_FOUND` error. ContinuousQuery resources use the canonical kind string `continuousquery` in all CLI invocations.
- **Dapr pre-installed**: `azd drasi provision` is responsible for installing Dapr on AKS. No assumption is made that Dapr exists before provisioning runs.
- **Cypher/GQL query authoring**: Query syntax validation is out of scope for v1. `azd drasi validate` validates YAML schema and cross-reference integrity, not query language correctness. Query syntax errors surface only at runtime deployment time.
- **Multi-tenant / multi-subscription**: v1 supports a single Azure subscription per environment. Multi-subscription and multi-tenant scenarios are out of scope.
- **Extension language**: The extension is authored in Go per the constitution Technology Constraints table. No alternative language is considered.
- **azd minimum version**: The extension requires azd v1.10.0 or later (the version that introduced the extensions beta framework). The extension manifest (`extension.yaml`) MUST declare this minimum version constraint.
- **Secret values at init time**: `azd drasi init` generates configuration with Key Vault reference placeholders. Actual secret values are provisioned by `azd drasi provision`; the developer does not provide raw credentials at init time.
- **Optional services are opt-in**: Cosmos DB, Event Hub, and Service Bus are not provisioned unless the active environment configuration references them. The default `dev` environment provisions only AKS, Key Vault, and Log Analytics.
- **Extension distribution**: The extension binary is distributed via the azd extension registry (`registry.yaml`) with GitHub Releases as the backing binary host. Users install via `azd extension install drasi`. The CI/CD pipeline (SC-007) MUST publish a versioned GitHub Release and update `registry.yaml` on each tagged release. The `extension.yaml` manifest MUST declare OS-specific download URLs pointing to GitHub Release assets (per FR-044).

---

## Clarifications

### Session 2026-04-04

- Q: Which Drasi runtime version should the extension target? → A: Pin to latest stable `drasi-platform` release at implementation time; record exact version in `extension.yaml` and README as minimum tested version.
- Q: How is the compiled extension binary distributed to end users? → A: azd extension registry (`registry.yaml`) backed by GitHub Releases binaries — standard published azd extension distribution path.
- Q: What timeout applies to FR-028 component `Online` readiness wait in `azd drasi deploy`? → A: 5 minutes per component, 15 minutes total deploy timeout.
- Q: Where is the per-component content-hash state persisted for FR-026 change detection? → A: azd environment state file (`.azure/<env>/` directory managed by azd); key format `DRASI_HASH_<componentKind>_<componentId>`.
- Q: Which Managed Identity role assignments must `azd drasi provision` create for FR-024? → A: Exact least-privilege set: `Key Vault Secrets User` on Key Vault; `Monitoring Metrics Publisher` on Log Analytics workspace; `AcrPull` on ACR (conditional on private registry use).
- Note (user-identified gap): Drasi runtime pods require AKS Workload Identity to authenticate to Azure services without stored credentials. Added FR-045: AKS OIDC issuer, Workload Identity addon, `FederatedIdentityCredential`, ServiceAccount annotation (`azure.workload.identity/client-id`), and pod label (`azure.workload.identity/use: "true"`) — all provisioned via Bicep. FR-043 uses in-cluster Secrets Store CSI synchronization and forbids direct Key Vault data-plane calls from extension orchestration code.

### Session 2026-04-08

- Q: What is the minimum Drasi CLI version floor for FR-035 Dev Container and FR-046 fail-fast check? → A: **`0.10.0`** (current latest `drasi-platform` release at spec authoring); update to current stable at implementation time. Removes circular reference between FR-035 and FR-046.
- Q: Which CLI kind string must the extension use for ContinuousQuery resources in `drasi wait` / `drasi delete` invocations? → A: **`continuousquery`** (full canonical form). The `query` alias is permitted by the CLI but MUST NOT be used in extension code. Updated FR-046 and the Drasi CLI coexistence assumption.
- Q: What is the canonical Drasi Kubernetes ServiceAccount identity for the FR-045 FederatedIdentityCredential subject? → A: **`system:serviceaccount:drasi-system:drasi-resource-provider`** (`drasi-system` is the default namespace from `drasi init`; add Bicep parameter `drasiNamespace` for custom namespace scenarios). This matches the service account created by the upstream installer manifest for the control-plane resource provider.
- Q: How does `azd drasi provision` determine when to conditionally assign the `AcrPull` role (FR-024)? → A: Bicep parameter **`usePrivateAcr`** (boolean, default `false`). When `true`, assigns `AcrPull` on the specified ACR and passes `--registry <acr>` to `drasi init`. Updated FR-024.
- Q: What Docker image reference should the FR-036 local compose use for the optional local Drasi server? → A: Do **not** assume a single `ghcr.io/drasi-project/drasi` image. Generate a compose topology that uses the published upstream Drasi component images under a configurable registry/version pair (`DRASI_IMAGE_REGISTRY`, `DRASI_VERSION`) and mirrors the current upstream local or Docker install layout.

### Session 2026-04-05

- Q: Where does middleware fit in the FR-042 deployment/teardown dependency order? → A: **sources → continuous queries → middleware → reactions** (deploy); reverse on teardown. Middleware sits between query emission and reaction execution per spec definition. Updated FR-042.
- Q: Can environment overlay files suppress or disable components defined in the base manifest? → A: Overlays cannot directly remove or suppress components. The only suppression mechanism is **feature flag overrides** in the overlay file — components gated by that flag are skipped by `azd drasi deploy` for that environment. Updated FR-017.
- Q: What retry policy applies when a `drasi` CLI subcommand returns a non-zero exit code? → A: **No automatic retry** for any CLI subcommand (`drasi apply`, `drasi delete`, `drasi list`, `drasi describe`). All non-zero exits are wrapped in `ERR_DRASI_CLI_ERROR` and fail fast. `drasi wait` handles internal polling autonomously; its non-zero exit maps to `ERR_COMPONENT_TIMEOUT`. Updated FR-046.
- Q: What should `azd drasi status` display when no Drasi components are deployed yet? → A: **Exit code 0** with message "No components deployed. Run `azd drasi deploy` to deploy Drasi components." — no error, clear remediation hint. Updated FR-039.
- Q: When `--tail` is omitted from `azd drasi logs`, should historical lines be returned on initial connection? → A: **No historical lines** — stream from current position forward (live tail only). Historical lines are returned only when `--tail <n>` is explicitly specified. Updated FR-040.
