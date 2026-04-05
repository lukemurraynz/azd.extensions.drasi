# Research: azd-drasi Extension

**Phase**: 0 — Research & Unknown Resolution  
**Date**: 2026-04-04  
**Feature**: 001-azd-drasi-extension  
**Spec**: [spec.md](./spec.md)

---

## Summary

All NEEDS CLARIFICATION items from the Technical Context have been resolved. This document records
each decision, its rationale, and the alternatives considered. Sources are authoritative (official
docs, GitHub source, or skill files with verified dates).

---

## 1. azd Extension Framework — Runtime Contract

**Decision**: Use the `azdext` SDK gRPC model (binary + extension.yaml + cobra).

**Rationale**: The `azdext.Run()` entrypoint handles all IPC with the azd host via gRPC. Custom
commands registered under the `drasi` namespace are exposed as `azd drasi <command>`. Writing to
stdout from lifecycle event handlers is FORBIDDEN (corrupts gRPC channel) — use the provided
`azdext` output API instead.

**Key facts verified** (source: `creating-azd-extensions` SKILL.md v1.0.0, 2026-03-28):

- Minimum azd version: **1.10.0+** (extensions beta gate)
- SDK import: `github.com/azure/azure-dev/cli/azd/pkg/azdext`
- Command framework: `github.com/spf13/cobra`
- Lifecycle capabilities: `custom-commands`, `lifecycle-events`, `metadata`
- Binary communication: gRPC (managed by azdext.Run — never call os.Exit directly)
- Cross-platform build: `build.ps1` (Windows) + `build.sh` (Linux/macOS) targeting windows/amd64, linux/amd64, darwin/amd64, darwin/arm64

**Alternatives considered**:

- Python CLI wrapper: Rejected — constitution Technology Constraints table locks Go for extension language. Non-negotiable.
- Shell script hook: Rejected — requires binary for gRPC lifecycle event support (lifecycle-events capability).

---

## 2. Drasi Runtime Hosting Platform

**Decision**: **AKS** is the required and non-negotiable runtime hosting platform for Drasi
components. The user's plan input (Section 3.5) specified "Azure Container Apps" — this is a
**Constitution Principle VI VIOLATION** and has been overridden.

**Rationale**: Constitution Principle VI (NON-NEGOTIABLE): "AKS is the default and required
hosting model for all Drasi components." Drasi for Kubernetes deploys into a Kubernetes namespace
(`drasi-system` default) and relies on Kubernetes CRDs, Dapr sidecar injection, and `drasi init`
which targets a kubeconfig context. Container Apps does not support these primitives.

**Key facts verified** (source: `drasi-queries` SKILL.md, drasi.io AKS installation guide):

- Drasi installs into namespace `drasi-system` via `drasi init`
- `drasi init` requires a valid kubectl context (AKS or compatible Kubernetes)
- Dapr is installed by `drasi init` automatically (version flags: `--dapr-runtime-version`, `--dapr-sidecar-version`)
- Default registry: `ghcr.io` (Drasi container images)
- Redis + MongoDB installed as Drasi dependencies

**Alternatives considered**:

- Azure Container Apps: Rejected — not supported by Drasi for Kubernetes platform; violates Principle VI.
- k3s/kind (local only): Supported for local dev compose (FR-036) but not for Azure deployment.

---

## 3. Drasi CLI Command Signatures (Verified)

**Decision**: All Drasi CLI invocations use **positional** kind + name syntax (not named flags).

**Rationale**: Verified against `github.com/drasi-project/drasi-platform/cli/cmd/` source files
(delete.go: `Use: "delete [kind name]"`, wait.go: `Use: "wait [...]"`). Named flags for `--kind`
and `--id` do not exist in the Drasi CLI; the CLI uses positional args consistent with kubectl.

**Verified command signatures** (source: drasi-platform GitHub source, CLI cmd/ directory):

```
drasi apply -f <manifest-file>
drasi wait <kind> <name> --timeout <seconds>
drasi delete <kind> <name>
drasi list <kind>
drasi describe <kind> <name>
```

Kind string for ContinuousQuery: `continuousquery` (canonical; `query` alias exists but MUST NOT
be used in extension code per Clarifications Session 2026-04-08).

**Minimum CLI version**: `0.10.0` (verified from drasi-platform releases; latest stable Nov 2025).

**Alternatives considered**:

- Named flags (`--kind`, `--id`): Rejected — not in Drasi CLI source; would cause runtime failures.

---

## 4. AKS Workload Identity Setup for Drasi

**Decision**: Use AKS OIDC Issuer + Workload Identity addon + FederatedIdentityCredential pattern
to give Drasi pods passwordless access to Azure Key Vault.

**Rationale**: Principle IV (Secure by Default, NON-NEGOTIABLE) forbids connection-string auth.
Workload Identity is the recommended pattern for AKS pod-to-Azure-service authentication per the
`identity-managed-identity` skill. The FederatedIdentityCredential binds a UAMI to the Drasi
Kubernetes ServiceAccount.

**Key facts verified** (source: `aks-cluster-architecture` SKILL.md, AKS docs, `identity-managed-identity` SKILL.md):

- AKS cluster flags: `--enable-oidc-issuer --enable-workload-identity` (❌ permanent — cluster rebuild to add)
- FederatedIdentityCredential subject: `system:serviceaccount:drasi-system:drasi-api`
  - Namespace: `drasi-system` (Drasi default from `drasi init`)
  - ServiceAccount name: `drasi-api` [VERIFY at implementation: check drasi-platform installer source]
- Audience: `api://AzureADTokenExchange`
- Pod annotation: `azure.workload.identity/client-id: <UAMI_CLIENT_ID>`
- Pod label: `azure.workload.identity/use: "true"`
- All steps MUST be in Bicep (no manual `az identity federated-credential create`)
- Bicep parameter `drasiNamespace` (default `drasi-system`) for custom namespace support

**RBAC roles required** (principle of least privilege per Principle IV):
| Role | Scope | Purpose |
|------|-------|---------|
| `Key Vault Secrets User` (`4633458b-17de-408a-b874-0445c86b69e6`) | Key Vault resource | Read secrets for KV→K8s Secret translation |
| `Monitoring Metrics Publisher` (`3913510d-42f4-4e42-8a64-420c390055eb`) | Log Analytics workspace | Export OpenTelemetry metrics |
| `AcrPull` (`7f951dda-4ed3-4680-a7ca-43fe172d538d`) | ACR resource | Pull Drasi images (conditional on `usePrivateAcr=true`) |

**Alternatives considered**:

- AAD Pod Identity: Deprecated — use Workload Identity per AKS deprecation table.
- Connection-string auth to Key Vault: Rejected — Principle IV violation, NON-NEGOTIABLE.
- Service principal secret: Rejected — Principle IV violation.

---

## 5. Content-Hash State Management (FR-026)

**Decision**: Hash each component's resolved YAML with SHA-256; persist in the azd environment
state file under key `DRASI_HASH_<KIND>_<ID>` (uppercase).

**Rationale**: The azd state file (`.azure/<env>/`) is managed by azd and survives across CLI
sessions. Using it avoids a separate state backend, stays aligned with azd idioms, and supports
multi-environment state isolation natively.

**Key facts verified**:

- azd state file path: `.azure/<env>/.env` (key-value pairs) or `.azure/<env>/config.json`
  - [VERIFY at implementation]: determine whether azd environment state R/W is exposed through
    `azdext` SDK APIs or requires direct file access to `.azure/<env>/.env`. Prefer SDK API.
- Hash algorithm: SHA-256 of canonical YAML serialization (keys sorted, normalized)
- Missing key → create semantics (FR-026: absent key = never deployed)
- Key format: `DRASI_HASH_CONTINUOUSQUERY_my-query-id` (kind uppercased, ID verbatim)

**Alternatives considered**:

- Separate JSON state file: Rejected — creates state drift risk; azd env file is already co-located and tracked.
- etag-based comparison: Rejected — Drasi API does not expose etags for CRD resources.

---

## 6. Key Vault → Kubernetes Secret Translation (FR-043)

**Decision**: At deploy time, for each KV reference (`{kind: secret, vaultName: X, secretName: Y}`),
fetch the secret value using the UAMI Workload Identity credential, write a K8s Secret in the
`drasi-system` namespace, and substitute the reference with `{kind: Secret, name: <k8s-name>, key: Y}`.

**Rationale**: Drasi YAML resources must contain K8s Secret references, not raw KV URIs. The
translation is deployment-time-only (not at init or provision time), ensuring secrets are never
persisted to the config file.

**Key facts verified** (source: `secret-management` SKILL.md):

- Key Vault RBAC role needed: `Key Vault Secrets User` (role ID `4633458b-17de-408a-b874-0445c86b69e6`)
- Authentication: `DefaultAzureCredential` (picks up azd auth login for local dev; Workload Identity in CI)
- K8s Secret name convention: `drasi-secret-<vaultName>-<secretName>` (lowercased, slugified)
- The IaC lockout trap: KV with `publicNetworkAccess: Disabled` requires 2-phase deployment
  (Phase 1: deploy with Allow + populate secrets; Phase 2: lock down). Apply this to infra/modules/keyvault.bicep.

**Alternatives considered**:

- CSI Secret Store driver: More complex (requires CSI add-on, SecretProviderClass); overkill for
  Drasi's deployment-time-only pattern. Rejected for v1.
- Hardcoded K8s Secrets in IaC: Rejected — Principle IV violation (secrets in IaC).

---

## 7. Configuration Engine — Multi-File YAML Model

**Decision**: Build a custom Go YAML loader using `gopkg.in/yaml.v3` with JSON Schema validation
via `github.com/santhosh-tekuri/jsonschema/v6`.

**Rationale**: Drasi resource schemas are simple and project-specific. A custom loader allows
tight integration with the DAG cycle detection, cross-reference validation, and incremental hash
computation. Standard third-party config frameworks (Viper, etc.) add indirection without benefit.

**Key design choices**:

- Glob resolution via `filepath.Glob` from `drasi/drasi.yaml` `includes:` paths
- Schema per resource type (source, query, reaction, middleware, environment) as embedded JSON Schema files
- Cross-reference resolution: collect all IDs, then validate references; single-pass accumulation
- DAG: adjacency list from query→sources + query→reactions; DFS cycle detection (Tarjan's or Kahn's)
- Deterministic model: sort all slices by ID before processing; ensures identical hash output

**Alternatives considered**:

- CUE / jsonnet: More expressive but heavy dependency; not standard in Go CLI ecosystem.
- Viper: Designed for app config, not for loading typed YAML resources. Rejected.

---

## 8. CI/CD and Release Pipeline

**Decision**: GitHub Actions with `ci.yml` (build/test/lint on every PR) and `release.yml`
(cross-compile + GitHub Release + registry.yaml update on semver tags).

**Rationale**: SC-007 requires a CI/CD pipeline completing in under 20 minutes. GitHub Actions is
the project default per constitution. Cross-compilation produces binaries for windows/amd64,
linux/amd64, darwin/amd64, darwin/arm64.

**Key facts verified** (source: `github-actions-ci-cd` SKILL.md):

- golangci-lint for Go linting
- `bicep build` + `az deployment what-if` for IaC validation
- Matrix build strategy for 4 OS/arch combinations
- Artifact upload + GitHub Release creation via `gh release create`
- registry.yaml updated in the same commit as the release (or via PR to extension registry repo)

**Alternatives considered**:

- Azure DevOps Pipelines: Permitted by constitution but not the default. Will document as an
  extension point in Phase 8.

---

## 9. Deployment Order and Rollback Strategy

**Decision**: Deploy in order sources → queries → reactions. Delete in reverse order. Apply
`delete-then-apply` for changed components. Detect orphaned-delete state on next run.

**Rationale**: Drasi resources reference each other directionally (queries reference sources by ID,
reactions reference queries). Deploying out-of-order causes unresolvable-reference runtime errors.
Verified in drasi-queries SKILL.md known pitfalls table: "Always delete in order: reactions →
queries → sources."

**Partial-failure recovery** (Edge Case from spec):

- After a failed `delete-then-apply`, the component hash IS removed from state (delete succeeded)
  but the component does not exist on the runtime (apply failed).
- On next `azd drasi deploy`: missing hash + missing component → create semantics → apply retry.
- This ensures recovery without operator intervention.

**Alternatives considered**:

- Rolling update (update-in-place): Rejected — Drasi CRD resources do not support in-place updates.
- Saga pattern with compensating transactions: Overkill for v1; component-level retry is sufficient.

---

## 10. OpenTelemetry Observability Strategy

**Decision**: Export OTel traces and metrics from all Drasi components and from the extension CLI
itself to Azure Monitor (Log Analytics workspace) via the OpenTelemetry Collector.

**Rationale**: Principle VII mandates OTel + Azure Monitor. The Log Analytics workspace is
provisioned by `azd drasi provision`. OTel Collector is deployed as a DaemonSet (or sidecar) on
AKS and configured with the Azure Monitor exporter.

**Key facts verified** (source: `observability-monitoring` SKILL.md):

- AKS Managed Prometheus: Use for cluster-level metrics; avoid duplicate ingestion with Container Insights
- ContainerLogV2: Migrate to before September 30, 2026 deadline
- Golden signals (latency, traffic, errors, saturation) via histogram metrics
- Drasi component status exposed via `drasi list` and `drasi describe` — use these for `azd drasi status`

**Alternatives considered**:

- Application Insights SDK (legacy): Permitted per constitution as fallback; rejected for new
  implementation in favour of OTel-first approach.

---

## Unresolved Items (Implementation-Time VERIFY Blocks)

| ID  | Item                                                                                                 | Verification Steps                                                                                           |
| --- | ---------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------ |
| V1  | ServiceAccount name (`drasi-api`) in Drasi's `drasi-system` namespace                                | Check `installer/` source in drasi-platform GitHub before writing Bicep FederatedIdentityCredential          |
| V2  | azd env state R/W API in `azdext` SDK                                                                | Check `pkg/azdext` package docs for environment state accessor before writing `internal/deployment/state.go` |
| V3  | Exact container image name for local dev compose (`ghcr.io/drasi-project/drasi`)                     | Check `https://github.com/drasi-project/drasi-platform/pkgs/container` before generating compose template    |
| V4  | azd min version for `lifecycle-events` capability (confirmed 1.10.0 but verify any patch constraint) | Check `https://github.com/Azure/azure-dev/releases` changelog for extensions beta history                    |
