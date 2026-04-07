# Extension CLI Contract

**Feature**: 001-azd-drasi-extension  
**Phase**: 1 — Design  
**Date**: 2026-04-04

This document defines the public CLI interface of the `azd drasi` command group.
It is the contract between the extension binary and its consumers (developers, CI/CD pipelines).

---

## CLI Grammar

```bash
azd drasi <command> [flags]
```

All commands are registered under the `drasi` namespace (matching `extension.yaml`
`namespace: drasi`). All commands accept `--output [table|json]` (default `table`).

---

## Commands

### `azd drasi init`

Scaffold a new Drasi project in the current directory.

```bash
azd drasi init [--template <name>] [--force] [--output <format>]
```

| Flag         | Type   | Default | Description                                                                                |
| ------------ | ------ | ------- | ------------------------------------------------------------------------------------------ |
| `--template` | string | `""`    | Starter template: `cosmos-change-feed`, `event-hub-routing`, `query-subscription`, `blank` |
| `--force`    | bool   | false   | Overwrite existing files                                                                   |
| `--output`   | string | `table` | Output format: `table` or `json`                                                           |

**Exit codes:**

| Code | Meaning                                                                  |
| ---- | ------------------------------------------------------------------------ |
| 0    | Success: files created                                                   |
| 1    | General error (invalid template name, duplicate files without `--force`) |

**Success output (table):**

```text
Initialized Drasi project in drasi/
  Created drasi/drasi.yaml
  Created drasi/sources/my-source.yaml
  Created drasi/queries/my-query.yaml
  Created drasi/reactions/my-reaction.yaml
```

**Success output (JSON):**

```json
{
  "status": "ok",
  "files": ["drasi/drasi.yaml", "drasi/sources/my-source.yaml"]
}
```

---

### `azd drasi validate`

Run offline schema and cross-reference validation against all Drasi YAML files.
No network or cluster access. Returns structured findings.

```bash
azd drasi validate [--config <path>] [--strict] [--output <format>]
```

| Flag       | Type   | Default            | Description              |
| ---------- | ------ | ------------------ | ------------------------ |
| `--config` | string | `drasi/drasi.yaml` | Path to root manifest    |
| `--strict` | bool   | false              | Treat warnings as errors |
| `--output` | string | `table`            | Output format            |

**Exit codes:**

| Code | Meaning                                                              |
| ---- | -------------------------------------------------------------------- |
| 0    | All validations passed                                               |
| 1    | One or more validation errors found (or warnings in `--strict` mode) |
| 2    | Tool precondition failure (config file not found, parse error)       |

**Success output:**

```text
✓ drasi/drasi.yaml    — valid
✓ drasi/sources/postgres.yaml — valid
✗ drasi/queries/sales-report.yaml:14 ERR_MISSING_QUERY_LANGUAGE
    ContinuousQuery 'sales-report' is missing required field 'queryLanguage'.
    Fix: add `queryLanguage: Cypher` or `queryLanguage: GQL`
```

**JSON output (on error):**

```json
{
  "status": "error",
  "issues": [
    {
      "level": "error",
      "file": "drasi/queries/sales-report.yaml",
      "line": 14,
      "code": "ERR_MISSING_QUERY_LANGUAGE",
      "message": "ContinuousQuery 'sales-report' is missing required field 'queryLanguage'.",
      "remediation": "Add `queryLanguage: Cypher` or `queryLanguage: GQL`"
    }
  ]
}
```

---

### `azd drasi provision`

Provision Azure infrastructure (AKS, Key Vault, UAMI) and install the Drasi runtime via `drasi init`.

```bash
azd drasi provision [--environment <name>] [--output <format>]
```

| Flag            | Type   | Default                | Description             |
| --------------- | ------ | ---------------------- | ----------------------- |
| `--environment` | string | Active azd environment | Target environment name |
| `--output`      | string | `table`                | Output format           |

**Exit codes:**

| Code | Meaning                                                              |
| ---- | -------------------------------------------------------------------- |
| 0    | Infrastructure provisioned and Drasi runtime online                  |
| 1    | Provisioning failed (Bicep deployment error or `drasi init` error)   |
| 2    | Precondition failure (azd not configured, missing Azure credentials) |

**Side effects:**

1. ARM deployment via Bicep (creates AKS cluster, Key Vault, UAMI, Log Analytics workspace)
2. `drasi init --context <aks-context>` installs Drasi runtime on AKS
3. Writes `DRASI_PROVISIONED=true` to azd environment

---

### `azd drasi deploy`

Deploy Drasi components (sources, queries, reactions) from the project configuration.

```bash
azd drasi deploy [--config <path>] [--environment <name>] [--dry-run] [--output <format>]
```

| Flag            | Type   | Default                | Description                           |
| --------------- | ------ | ---------------------- | ------------------------------------- |
| `--config`      | string | `drasi/drasi.yaml`     | Root manifest path                    |
| `--environment` | string | Active azd environment | Target environment name               |
| `--dry-run`     | bool   | false                  | Show planned actions without applying |
| `--output`      | string | `table`                | Output format                         |

**Exit codes:**

| Code | Meaning                                                                                    |
| ---- | ------------------------------------------------------------------------------------------ |
| 0    | All components deployed and online                                                         |
| 1    | At least one component failed to deploy                                                    |
| 2    | Precondition failure (validation failed, Drasi CLI not found/too old, cluster unreachable) |

**Dry-run output (table):**

```text
Deployment Plan (--dry-run, no changes will be made):
  SOURCE        postgres-source    create
  CONTINUOUSQUERY  sales-report    create
  REACTION      pubsub-reaction    create
```

**Dry-run JSON:**

```json
{
  "status": "dry-run",
  "actions": [
    {
      "kind": "Source",
      "id": "postgres-source",
      "action": "create",
      "reason": "no prior hash"
    },
    {
      "kind": "ContinuousQuery",
      "id": "sales-report",
      "action": "create",
      "reason": "no prior hash"
    },
    {
      "kind": "Reaction",
      "id": "pubsub-reaction",
      "action": "create",
      "reason": "no prior hash"
    }
  ]
}
```

---

### `azd drasi status`

Show the current health status of all deployed Drasi components.

```bash
azd drasi status [--environment <name>] [--output <format>]
```

| Code | Meaning                                                            |
| ---- | ------------------------------------------------------------------ |
| 0    | All components are Online                                          |
| 1    | One or more components are not Online (Pending, Error, or missing) |
| 2    | Precondition failure                                               |

**Table output:**

```text
Component Status — environment: dev

KIND              ID                  STATUS    AGE
Source            postgres-source     Online    2h
ContinuousQuery   sales-report        Online    2h
Reaction          pubsub-reaction     Pending   5m
```

---

### `azd drasi logs`

Stream or print logs from Drasi runtime pods.

```bash
azd drasi logs [--component <id>] [--kind <kind>] [--tail <n>] [--follow] [--environment <name>]
```

| Flag            | Type   | Default                | Description                                      |
| --------------- | ------ | ---------------------- | ------------------------------------------------ |
| `--component`   | string | `""`                   | Filter by component ID                           |
| `--kind`        | string | `""`                   | Filter by kind: `source`, `query`, or `reaction` |
| `--tail`        | int    | 100                    | Number of trailing lines to print                |
| `--follow`      | bool   | false                  | Stream logs continuously                         |
| `--environment` | string | Active azd environment | Target environment                               |

**Exit codes:**

| Code | Meaning                                    |
| ---- | ------------------------------------------ |
| 0    | Logs retrieved successfully                |
| 1    | Component not found or cluster unreachable |

---

### `azd drasi diagnose`

Run a diagnostic health check across 5 domains and report findings.

```bash
azd drasi diagnose [--environment <name>] [--output <format>]
```

| Check            | What it verifies                                                |
| ---------------- | --------------------------------------------------------------- |
| AKS Connectivity | kubeconfig context, API server reachable                        |
| Drasi API        | `drasi-api` pod running in `drasi-system` (or `drasiNamespace`) |
| Dapr Runtime     | Dapr sidecar injector running                                   |
| Key Vault        | Can authenticate and list secrets                               |
| Log Analytics    | Workspace exists, data flowing                                  |

**Exit codes:**

| Code | Meaning                   |
| ---- | ------------------------- |
| 0    | All checks passed         |
| 1    | One or more checks failed |

---

### `azd drasi teardown`

Remove all deployed Drasi components and optionally the Azure infrastructure.

```bash
azd drasi teardown --force [--infrastructure] [--environment <name>] [--output <format>]
```

| Flag               | Type   | Default                | Description                                                            |
| ------------------ | ------ | ---------------------- | ---------------------------------------------------------------------- |
| `--force`          | bool   | —                      | **Required.** Acknowledges that components will be permanently deleted |
| `--infrastructure` | bool   | false                  | Also delete Azure infrastructure (AKS, Key Vault, UAMI)                |
| `--environment`    | string | Active azd environment | Target environment                                                     |
| `--output`         | string | `table`                | Output format                                                          |

**Exit codes:**

| Code | Meaning                |
| ---- | ---------------------- |
| 0    | Teardown complete      |
| 1    | Teardown failed        |
| 2    | `--force` not provided |

**Delete order:** reactions → queries → sources (reverse of deploy order per FR-042)

---

## Global Flags

All commands inherit:

| Flag       | Type   | Default | Description                      |
| ---------- | ------ | ------- | -------------------------------- |
| `--output` | string | `table` | Output format: `table` or `json` |
| `--debug`  | bool   | false   | Enable verbose debug logging     |

---

## Structured Error Contract

When `--output json` is used and a command fails, the error shape is:

```json
{
  "status": "error",
  "code": "ERR_DRASI_CLI_NOT_FOUND",
  "message": "Drasi CLI not found in PATH. Install with: ...",
  "detail": {}
}
```

### Error Codes

| Code                         | Trigger                                             |
| ---------------------------- | --------------------------------------------------- |
| `ERR_DRASI_CLI_NOT_FOUND`    | `drasi` binary not found in PATH                    |
| `ERR_DRASI_CLI_VERSION`      | Drasi CLI version < 0.10.0                          |
| `ERR_COMPONENT_TIMEOUT`      | Component did not reach Online within 5 min         |
| `ERR_TOTAL_TIMEOUT`          | Deployment exceeded 15 min total                    |
| `ERR_VALIDATION_FAILED`      | `validate` found errors; deploy refused             |
| `ERR_MISSING_REFERENCE`      | A query references an unknown source or reaction ID |
| `ERR_CIRCULAR_DEPENDENCY`    | Dependency graph contains a cycle                   |
| `ERR_MISSING_QUERY_LANGUAGE` | ContinuousQuery is missing `queryLanguage` field    |
| `ERR_KV_AUTH_FAILED`         | Cannot authenticate to Key Vault                    |
| `ERR_AKS_CONTEXT_NOT_FOUND`  | kubeconfig context for AKS not found                |

---

## Extension Manifest Contract

**File**: `extension.yaml` (at repository root)

```yaml
name: azd-drasi
namespace: drasi
displayName: "Drasi for Azure Developer CLI"
description: "Manage Drasi reactive data pipeline workloads with azd."
version: "1.0.0"
minAzdVersion: "1.10.0"
type: extension
capabilities:
  - custom-commands
  - lifecycle-events
executablePath:
  windows/amd64: bin/windows/amd64/azd-drasi.exe
  linux/amd64: bin/linux/amd64/azd-drasi
  linux/arm64: bin/linux/arm64/azd-drasi
  darwin/amd64: bin/darwin/amd64/azd-drasi
  darwin/arm64: bin/darwin/arm64/azd-drasi
```

---

## Drasi YAML Config Contract

**Root manifest**: `drasi/drasi.yaml`

```yaml
apiVersion: v1
includes:
  - kind: sources
    pattern: "sources/**/*.yaml"
  - kind: queries
    pattern: "queries/**/*.yaml"
  - kind: reactions
    pattern: "reactions/**/*.yaml"
environments:
  dev: "environments/dev.yaml"
  prod: "environments/prod.yaml"
```

**Source file** (example):

```yaml
apiVersion: v1
kind: Source
id: postgres-source
sourceKind: PostgreSQL
properties:
  host:
    kind: secret
    vaultName: my-keyvault
    secretName: postgres-host
  port: "5432"
  database: orders
```

**ContinuousQuery file** (example):

```yaml
apiVersion: v1
kind: ContinuousQuery
id: order-changes
queryLanguage: Cypher
sources:
  - id: postgres-source
query: |
  MATCH (o:Order)
  WHERE o.status = 'PENDING'
  RETURN o.id AS orderId, o.amount AS amount
reactions:
  - pubsub-orders
autoStart: true
```

**Reaction file** (example):

```yaml
apiVersion: v1
kind: Reaction
id: pubsub-orders
reactionKind: dapr-pubsub
config:
  topic: order-events
  pubsubName: drasi-pubsub
```
