<!--SYNC IMPACT REPORT
Version change: (none) → 1.0.0 | 1.0.0 → 1.0.1 | 1.0.1 → 1.0.2

--- v1.0.0 (2026-04-04) ---
Added sections:
  - Core Principles (10 principles: I–X)
  - Technology Constraints
  - Non-Goals
  - Governance
Removed sections: n/a (initial creation from template)
Modified principles: n/a (initial creation)
Templates requiring updates:
  - .specify/templates/plan-template.md — Constitution Check gates reference principles I–X ✅ compatible (generic format)
  - .specify/templates/spec-template.md — FR numbered IDs and acceptance scenarios align with principles II, VIII ✅ compatible
  - .specify/templates/tasks-template.md — Test-first phase structure aligns with principle VIII; layered task organization aligns with principle VI ✅ compatible
Deferred items: none

--- v1.0.1 (2026-04-04) ---
Added sections:
  - Required Skills (new section before Governance)
Modified sections: none
Rationale: Skill loading is not automatic in AI agents; explicit constitution-level mandate
  ensures speckit.plan, speckit.tasks, and speckit.implement reliably load domain skills
  before generating plans, task lists, or implementation code.
Templates requiring updates: none (new section is agent-execution guidance only)
Deferred items: none

--- v1.0.2 (2026-04-04) ---
Added sections: none
Modified sections:
  - Required Skills — expanded from 3 skills to 14 skills (7 mandatory + 7 conditional)
Rationale: Discovery of additional skills in .github/skills/ covering AKS architecture,
  Key Vault integration, Managed Identity/Workload Identity, observability/OpenTelemetry,
  Dev Containers, private networking, RBAC role selection, GitHub Actions CI/CD, azd
  lifecycle management, threat modelling, and event-driven messaging. Skills are split
  into Mandatory (load for every phase) and Conditional (load when subsystem is in scope).
  Mandatory skills added: aks-cluster-architecture, secret-management, identity-managed-identity,
  observability-monitoring.
  Conditional skills added: creating-devcontainers, private-networking, azure-role-selector,
  github-actions-ci-cd, managing-azure-dev-cli-lifecycle, threat-modelling, event-driven-messaging.
Templates requiring updates: none (section is agent-execution guidance only)
Deferred items: none
-->

# Drasi AZD Extensions Constitution

## Core Principles

### I. Declarative Infrastructure (NON-NEGOTIABLE)

All infrastructure, configuration, and pipeline definitions MUST be expressed as code.
Bicep is the preferred IaC language; Terraform is permitted for cross-cloud scenarios.
Imperative post-deployment scripts are FORBIDDEN for control-plane operations;
they are permitted ONLY for data-plane initialization (e.g., seed data, Drasi source registration).
Every resource MUST be environment-scoped and carry mandatory tags:
`environment`, `project`, `component`, `managed-by=azd`.
Manual portal changes to provisioned resources are NOT permitted after initial deployment.

**Rationale**: Reproducibility across dev/staging/production requires a single declarative source of truth.
Drift between environments caused by out-of-band changes invalidates the extension's core promise.

### II. Idempotency (NON-NEGOTIABLE)

Every `azd drasi` command MUST be safe to execute multiple times without producing unintended
side effects. Re-running `azd drasi provision` or `azd drasi deploy` on an already-provisioned
environment MUST converge to the desired state, not duplicate or corrupt it.
Commands MUST detect existing state before creating resources and perform upsert semantics.
Destructive operations (delete, purge) MUST require explicit opt-in flags (e.g., `--force`).

**Rationale**: CI/CD pipelines and operator retries are first-class use cases.
Non-idempotent commands create operational risk and block automation.

### III. Composability

The extension MUST integrate into the azd lifecycle as an additive layer, not a replacement.
Commands MUST follow azd hook conventions (`preProvision`, `postProvision`, `preDeploy`, `postDeploy`).
Extension commands MUST be invokable standalone (`azd drasi deploy`) and as azd lifecycle hooks.
The extension MUST NOT override or shadow built-in azd commands.
The CLI layer, orchestration layer, and infrastructure layer MUST remain independently replaceable.
Third-party dependencies MUST be justified; prefer Azure SDKs and native CLI tooling.

**Rationale**: Developers adopt this extension within existing azd workflows.
Tight coupling or replacement patterns create adoption friction and maintenance burden.

### IV. Secure by Default (NON-NEGOTIABLE)

No insecure defaults are permitted. Every resource MUST use identity-based access
(Managed Identity or Workload Identity) as the default authentication mechanism.
Connection strings, passwords, and API keys MUST NOT appear in source code, config files,
`azure.yaml`, environment variable defaults, or CI/CD workflow definitions.
All secrets MUST be stored in Azure Key Vault and accessed at runtime via Key Vault references
or the Azure Identity SDK (`DefaultAzureCredential`).
RBAC MUST enforce least-privilege; no wildcard role assignments (Owner/Contributor at subscription
scope) are permitted in generated IaC.
All CLI actions that mutate Azure resources MUST emit an audit log entry to structured output.

**Rationale**: Client-distributed extension code is public. Any hardcoded credential or
over-provisioned role becomes an immediate attack surface across all consumer environments.

### V. Developer Experience (DX-First CLI Design)

Commands MUST use the `azd drasi <verb>` namespace and feel native to the azd CLI convention.
Required commands: `init`, `provision`, `deploy`, `status`, `logs`, `teardown`.
Every command MUST output structured, human-readable results to stdout and machine-readable
JSON when `--output json` is passed.
Error messages MUST include: error code, human-readable description, and a remediation step or
documentation link. Silent failures are FORBIDDEN.
The extension MUST support local development via Dev Containers and GitHub Codespaces with
full parity to cloud environments.
Scaffolding templates for common Drasi scenarios (change feed, event routing, query subscription)
MUST be included and invokable via `azd drasi init --template <name>`.

**Rationale**: Developer adoption is the primary success metric.
Poor CLI ergonomics or cryptic errors cause abandonment regardless of technical correctness.

### VI. Architecture Layering (NON-NEGOTIABLE)

The codebase MUST maintain three distinct, independently testable layers:

- **CLI layer** (`cmd/`): Command parsing, flag validation, output formatting. No business logic.
- **Orchestration layer** (`internal/`): Deployment logic, state management, lifecycle coordination.
  No direct Azure API calls; delegates to the infrastructure layer.
- **Infrastructure layer** (`infra/`): Bicep/Terraform templates, Kubernetes manifests, Dapr
  component definitions. No Go/scripting logic.

Cross-layer dependencies MUST only flow downward (CLI → Orchestration → Infrastructure).
The infrastructure layer MUST be deployable independently of the CLI (e.g., via `az deployment`).
AKS is the default and required hosting model for all Drasi components.
Dapr MUST be used for pub/sub and service invocation where eventing is required.
Pluggable backends (Azure Event Hubs, Azure Service Bus, Cosmos DB change feed) MUST be
supported via replaceable Dapr component YAML; the core extension MUST NOT hardcode a
single event backend.

**Rationale**: Layer violations collapse testability and make the extension impossible to
operate without the CLI binary. Pluggability is essential for enterprise adoption.

### VII. Operability & SRE

The extension MUST provision OpenTelemetry-instrumented workloads by default.
All Drasi components deployed by the extension MUST export logs and traces to Azure Monitor.
`azd drasi status` MUST report health of all deployed Drasi components with actionable
remediation hints for unhealthy states.
`azd drasi logs` MUST stream structured logs from AKS pods to the terminal.
The extension MUST support ring-based deployments and feature flags for Drasi pipeline rollouts.
Fail-fast behavior is REQUIRED: if a pre-condition check fails, the command MUST exit
immediately with a non-zero exit code and a specific remediation message.
Health probes (liveness, readiness, startup) MUST be defined for every Kubernetes workload
generated by the extension.

**Rationale**: Silent or partial failures in event-driven systems cause data loss and cascade
failures. Operators require immediate, actionable diagnostics.

### VIII. Test-First Quality (NON-NEGOTIABLE)

TDD workflow is MANDATORY for all command and orchestration logic:
tests MUST be written and reviewed before implementation begins (Red-Green-Refactor).
Test categories REQUIRED:

- **Unit tests**: All command flag parsing, output formatting, and orchestration logic.
  Coverage gate: 80% line coverage minimum.
- **Integration tests**: All deployment workflows against real or containerized Azure services
  (use `testcontainers` or az CLI mocks as appropriate).
- **IaC validation tests**: `bicep build` and `az deployment what-if` MUST pass in CI before
  any infrastructure PR merges.
  Linting MUST pass for all languages (Go: `golangci-lint`; Bicep: `bicep lint`; YAML: `yamllint`).
  No PR may merge with failing tests, lint errors, or unresolved `TODO`/`FIXME`/`HACK` markers
  in non-test code.

**Rationale**: Event-driven infrastructure bugs are expensive to debug in production.
Tests are the primary mechanism for verifying idempotency (Principle II) and
composability (Principle III) guarantees.

### IX. Semantic Versioning & Extensibility

The extension MUST follow Semantic Versioning 2.0.0 (semver.org):

- MAJOR: breaking changes to command interfaces or `azure.yaml` schema.
- MINOR: new commands, new template scenarios, backward-compatible behavior additions.
- PATCH: bug fixes, documentation updates, non-semantic refinements.

The extension manifest (`extension.yaml`) MUST declare compatibility constraints
against the minimum azd version required.
The architecture MUST support future extensibility: custom Drasi processors, source connectors,
and reaction handlers MUST be registerable via a documented plugin interface without
modifying core extension code.
The extension MUST be publishable to the official azd extension registry and support
private feed distribution.
Changelog entries are REQUIRED for every release; format: Keep a Changelog (keepachangelog.com).

**Rationale**: azd users deploy extensions into long-lived environments.
Breaking changes without version gates corrupt those environments silently.

### X. Documentation & Usability

Every command MUST have `--help` output documenting: purpose, required flags, optional flags,
and an example invocation.
The repository MUST contain:

- `README.md` with quickstart (provision → deploy → status in under 10 commands).
- `docs/` directory with at least: architecture overview, configuration reference, and
  troubleshooting guide covering the top 5 known failure modes with recovery steps.
- End-to-end scenario examples for: change feed ingestion, event routing, and query subscription.
  Documentation MUST be updated in the same PR as the code change it describes.
  Stale documentation (out of sync with current behavior) constitutes a defect and
  MUST be treated with the same priority as a functional bug.

**Rationale**: Undocumented CLIs are adopted only by the author. Documentation parity
with code is the minimum bar for enterprise-grade tooling.

## Technology Constraints

The following technology choices are binding for all features unless an Architecture Decision
Record (ADR) documents a justified exception:

| Concern              | Required Technology                                  | Alternatives Permitted                  |
| -------------------- | ---------------------------------------------------- | --------------------------------------- |
| IaC                  | Bicep (default)                                      | Terraform (cross-cloud only)            |
| Container hosting    | AKS                                                  | None (AKS is required for Drasi)        |
| Eventing / pub-sub   | Dapr                                                 | None for new eventing surfaces          |
| Event backends       | Azure Event Hubs, Service Bus, Cosmos DB change feed | Must be Dapr-component swappable        |
| Secret storage       | Azure Key Vault                                      | None                                    |
| Observability        | OpenTelemetry → Azure Monitor                        | Application Insights SDK (legacy only)  |
| Auth                 | Managed Identity / Workload Identity                 | None (no connection-string auth)        |
| Extension language   | Go                                                   | None                                    |
| Package distribution | azd extension registry                               | Private NuGet/npm feeds (additive only) |

Azure Policy alignment MUST be verified for all generated resource configurations before release.
Resources MUST NOT be deployed outside approved Azure regions defined in `infra/regions.json`.

## Non-Goals

The following are explicitly out of scope. Any PR that moves toward these goals
without an approved ADR MUST be rejected by code review:

- Tight coupling to a single data source or event system. The extension MUST remain
  backend-agnostic at the Drasi connector layer.
- Proprietary or paid-only dependencies for core functionality. All required runtimes
  and tools MUST have a free tier or open-source license.
- Replacing or wrapping azd's built-in provisioning commands. The extension augments azd;
  it does not substitute it.
- GUI or portal-based management. The extension is CLI-only.
- Multi-cloud infrastructure (non-Azure). Azure is the sole supported cloud target.

## Required Skills

AI agents (Copilot, speckit agents) working on this project MUST load the following skill
files before generating plans, task lists, or implementation code. Skill loading is not
automatic — agents must explicitly read the skill file via the `read_file` tool before
proceeding with domain work. Failure to load these skills is a constitution violation.

### Mandatory Skills

Load these before ANY `/speckit.plan`, `/speckit.tasks`, or `/speckit.implement` phase:

| Skill                       | File                                                | Load When                                                                                                                                   |
| --------------------------- | --------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `creating-azd-extensions`   | `.github/skills/creating-azd-extensions/SKILL.md`   | Any work involving Go extension code, `extension.yaml`, azd lifecycle hooks, `azdext` SDK, or registry distribution                         |
| `drasi-queries`             | `.github/skills/drasi-queries/SKILL.md`             | Any work involving Drasi Sources, ContinuousQueries, Reactions, Cypher/GQL query authoring, Drasi CLI commands, or Drasi YAML resources     |
| `azd-deployment`            | `.github/skills/azd-deployment/SKILL.md`            | Any work involving `azure.yaml`, azd provisioning workflows, Container Apps deployment                                                      |
| `aks-cluster-architecture`  | `.github/skills/aks-cluster-architecture/SKILL.md`  | Any work involving AKS cluster design, node pools, CNI, Workload Identity, or permanent infrastructure decisions (Principle VI)             |
| `secret-management`         | `.github/skills/secret-management/SKILL.md`         | Any work involving Key Vault RBAC, managed identity secret access, or Key Vault → Kubernetes Secret translation (FR-020, FR-043)            |
| `identity-managed-identity` | `.github/skills/identity-managed-identity/SKILL.md` | Any work involving Managed Identity, Workload Identity Federation, RBAC assignments, or passwordless service-to-service auth (Principle IV) |
| `observability-monitoring`  | `.github/skills/observability-monitoring/SKILL.md`  | Any work involving OpenTelemetry instrumentation, Azure Monitor, structured logging, or alerting rules (FR-038–FR-041, Principle VII)       |

### Conditional Skills

Load these when the named subsystem is in scope for the current phase:

| Skill                              | File                                                       | Load When                                                                                                      |
| ---------------------------------- | ---------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| `creating-devcontainers`           | `.github/skills/creating-devcontainers/SKILL.md`           | Implementing local development environment or Dev Container setup (FR-035–FR-037)                              |
| `private-networking`               | `.github/skills/private-networking/SKILL.md`               | Designing AKS private cluster topology, private endpoints, NSGs, or private DNS zones                          |
| `azure-role-selector`              | `.github/skills/azure-role-selector/SKILL.md`              | Selecting or generating RBAC role assignments in Bicep or CLI (least-privilege enforcement per Principle IV)   |
| `github-actions-ci-cd`             | `.github/skills/github-actions-ci-cd/SKILL.md`             | Implementing GitHub Actions workflows for build, test, ACR push, or AKS deployment (SC-007)                    |
| `managing-azure-dev-cli-lifecycle` | `.github/skills/managing-azure-dev-cli-lifecycle/SKILL.md` | Implementing or testing `azd up` / `azd down --purge` lifecycle, multi-environment azd configs, or azd cleanup |
| `threat-modelling`                 | `.github/skills/threat-modelling/SKILL.md`                 | Conducting security architecture review, STRIDE/DREAD analysis, or trust boundary mapping                      |
| `event-driven-messaging`           | `.github/skills/event-driven-messaging/SKILL.md`           | Implementing Service Bus or Event Hub source adapters (FR-013 source kinds)                                    |

**Loading gate**: Before starting any `/speckit.plan`, `/speckit.tasks`, or `/speckit.implement`
phase, the agent MUST confirm all seven mandatory skills have been loaded in the current session.
Conditional skills MUST be loaded when the named subsystem is in scope. If a skill file is not
yet loaded, load it before proceeding — do not rely on training-data knowledge for domain-specific
API shapes, CLI flags, or known pitfalls.

**Rationale**: Domain-specific skills encode verified, version-pinned knowledge (Drasi
function signatures, azd extension gRPC contract, AKS Workload Identity bindings, Key Vault
integration patterns) that diverges from training data. Skipping skill loading has produced
incorrect specs (e.g., wrong source kinds, missing `queryLanguage` field, incorrect upsert
semantics for Drasi resources). The cost of loading a skill file is one tool call; the cost
of a missed constraint is downstream rework across plan and tasks.

## Governance

This constitution supersedes all other practice documents, README guidance, and ad-hoc
conventions for the azd.extensions.drasi project.

**Amendment procedure**:

1. Open a GitHub Issue proposing the amendment with: principle affected, rationale, impact assessment.
2. Obtain approval from at least one maintainer via PR review.
3. Bump `CONSTITUTION_VERSION` according to the versioning rules in Principle IX.
4. Update all dependent templates (plan-template.md, spec-template.md, tasks-template.md)
   in the same PR.
5. Record the amendment in the Sync Impact Report comment at the top of this file.

**Compliance review**: Every PR MUST include a "Constitution Check" section in its plan.md
verifying that all 10 principles are satisfied or explicitly justified as inapplicable.

**Versioning policy**: Version bumps follow the same MAJOR.MINOR.PATCH semantics as
Principle IX applied to governance changes.

**Agent execution note**: AI agents (Copilot, speckit agents) MUST treat MUST/MUST NOT
directives in this document as hard blockers. Suggestions that violate these directives
MUST be rejected. SHOULD directives are strong recommendations requiring documented
justification to deviate.

**Version**: 1.0.2 | **Ratified**: 2026-04-04 | **Last Amended**: 2026-04-04
