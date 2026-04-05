---
name: drasi-queries
description: >-
  End-to-end skill for building Drasi solutions (Sources, ContinuousQueries, Reactions pipeline), including automation, troubleshooting, and safe extension patterns. USE FOR: create Drasi continuous query, configure Drasi source, set up Drasi reaction, deploy Drasi queries, troubleshoot Drasi errors, write Cypher queries for Drasi.
---

# Drasi End-to-End Skill: Sources -> ContinuousQueries -> Reactions

## Currency and Verification

- **Date checked:** 2026-04-03 (verified via Drasi official docs and API reference)
- **Compatibility:** Drasi Server v1, REST API v1, Drasi CLI (latest stable)
- **Sources:** [Drasi REST API](https://drasi.io/reference/rest-api/), [Drasi Query Language](https://drasi.io/reference/query-language/), [Drasi Custom Functions](https://drasi.io/reference/query-language/drasi-custom-functions/)
- **Verification steps:**
  1. All REST endpoint paths, methods, and payloads checked against latest docs
  2. CLI commands (`drasi apply`, `drasi delete`, etc.) confirmed in CLI reference
  3. Cypher function support and case-sensitivity verified in Drasi docs
  4. No references to non-existent Drasi SDKs or namespaces
  5. Resource lifecycle and dependency order validated
  6. All YAML resource examples use `apiVersion: v1`
  7. Webhook and log reaction patterns match documented structure

## When to Use

Trigger this skill for any of the following:

- Drasi resources: Source, SourceProvider, ContinuousQuery, QueryContainer, Reaction, ReactionProvider
- Drasi CLI operations: `drasi apply`, `drasi delete`, `drasi wait`, `drasi describe`, `drasi list`, `drasi init`
- Repo workflows: `scripts/drasi-refresh.ps1`, `infrastructure/drasi/*`
- Query language or provider questions specific to Drasi (Cypher/GQL subset, Drasi custom functions)

**Apply this skill for any Drasi resource work:**

- Source / SourceProvider
- ContinuousQuery / QueryContainer
- Reaction / ReactionProvider
- Drasi CLI workflows and CI/CD automation

> Only use this skill for Drasi-related tasks.

---

## 0. Quick Start (Portable + Repo Examples)

- In any project, define a clear resource layout (sources/queries/reactions/secrets) and keep all Drasi YAML in version control.
- Default namespace is typically `drasi-system` (use `-n <namespace>` if not set in your Drasi context).
- Prefer an automation script that renders templates and performs delete → apply (for this repo: `scripts/drasi-refresh.ps1`).
- For future projects: choose a runtime early using the decision matrix below.
- Drasi Platform evolves quickly (early release); pin CLI/control-plane versions and keep `[VERIFY]` blocks in PRs.
- For patterns and samples: see upstream `drasi-project/learning` and `drasi-project/docs`.
- This repo's operational reference is `infrastructure/drasi/README.md`.

### Deployment Mode Decision Matrix

| Mode                                        | Runtime                | Best For                                               | Deploy Model                         |
| ------------------------------------------- | ---------------------- | ------------------------------------------------------ | ------------------------------------ |
| **Drasi for Kubernetes** (`drasi-platform`) | C# / CNCF Sandbox      | Production K8s workloads, full source/reaction catalog | `drasi init` + `drasi apply`         |
| **Drasi Server** (`drasi-server`)           | Standalone Rust binary | Docker Compose, edge, non-K8s environments             | Binary/Docker, REST API, YAML config |
| **drasi-lib** (`drasi-core`)                | Embedded Rust crate    | In-process continuous queries, custom runtimes         | `cargo add drasi-core`               |
| **drasi-core-python**                       | Embedded Python (PyO3) | Python services, data pipelines, notebooks             | `pip install drasi-core-python`      |

Choose **Drasi for Kubernetes** unless you have a specific need for standalone or embedded deployment.

## 0.1 Repo-Specific Gotchas (Emergency Alerts Demo)

- **Cypher subset**: Avoid unsupported functions like `collect()` (common failure: `UnknownFunction("collect")` → query `TerminalError`).
- **Delete-before-apply**: Drasi resources generally don’t update in place; use `scripts/drasi-refresh.ps1` when iterating.
- **Queries are split across files**: This repo keeps multiple query YAMLs under `infrastructure/drasi/queries/` (not a single “query pack”). `scripts/drasi-refresh.ps1` supports applying a directory.
- **Frontend endpoints are build-time**: The SPA uses Vite build-time injection; ensure `VITE_API_URL` and `VITE_SIGNALR_URL` are passed at image build time (not pod env vars).
- **HTTP reaction base URL must include backend reaction route prefix**: Use a URL like `http://<api-service>.<namespace>.svc.cluster.local/<reaction-route-prefix>`. If the prefix is missing (or service port is wrong), reaction pods can show repeated 30s callback timeouts with no downstream effects.
- **Reaction auth token must match backend expectation**: `X-Reaction-Token` (or your chosen auth header) in the Reaction must equal the backend’s configured token. Mismatches produce `401` on reaction endpoints.
- **“All queries active” is not enough**: After running a scenario, validate reaction flow at downstream endpoints (events feed, operator inbox, notifications, or equivalent).
- **CLI/apiVersion compatibility**: Query YAML authored as `apiVersion: query.reactive.drasi.io/v1` can fail with some CLI versions. Normalize to `apiVersion: v1` at apply time in your deployment script.

### Known Pitfalls (2026-04-03)

| Area              | Pitfall                                                                 | Mitigation                                                                       |
| ----------------- | ----------------------------------------------------------------------- | -------------------------------------------------------------------------------- |
| Cypher functions  | Uppercase aggregation (`COUNT()`) or unsupported (`collect()`)          | Use only lowercase (`count()`, etc.) and verify function exists in Drasi docs    |
| Query arithmetic  | `datetime.realtime()` arithmetic, division by zero, raw timestamp usage | Use `drasi.trueLater()`, always guard division, use `drasi.changeDateTime(node)` |
| Resource updates  | Drasi resources cannot be updated in place                              | Always delete before apply; automate with scripts                                |
| API version       | Using `apiVersion: query.reactive.drasi.io/v1` in YAML                  | Normalize to `apiVersion: v1` for compatibility                                  |
| Webhook reactions | Missing route prefix or wrong service port                              | Ensure base URL includes backend reaction route prefix and correct port          |
| Reaction auth     | Token mismatch between Reaction and backend                             | Set `X-Reaction-Token` header to match backend expectation                       |
| Resource deletion | Deleting sources before dependent queries/reactions                     | Always delete in order: reactions → queries → sources                            |
| CLI/REST drift    | CLI or REST API changes not reflected in skill                          | [VERIFY] Check Drasi docs for breaking changes before each update                |
| Security/auth     | No authentication/authorization in REST API by default                  | [VERIFY] Add security controls if required; confirm with Drasi docs              |

**These patterns are NOT supported in Drasi and will cause queries to deploy with no error, then appear Inactive/TerminalError:**

| ❌ Pattern (FAILS)                                             | ✅ Correct Pattern                                                                                                               | Issue                                                                                                                                                    |
| -------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `datetime.realtime() - duration({ hours: 1 })` in WHERE        | Use aggregation: `max(drasi.changeDateTime(n))` or reactive trigger `drasi.trueLater(cond, timestamp + duration({...}))`         | Cannot perform arithmetic on `datetime.realtime()` in WHERE clauses; use reactive triggers for time-based logic                                          |
| `duration.between(time1, time2).seconds`                       | Use `drasi.trueLater(condition, time1 + duration({seconds: 60}))`                                                                | `duration.between()` function does not exist in Drasi                                                                                                    |
| `MATCH (a:t1), (b:t2) WHERE a.id = b.id` (implicit join)       | Add explicit `joins:` config: `joins: [{id: "T1_HAS_T2", keys: [{label: "t1", property: "id"}, {label: "t2", property: "id"}]}]` | Drasi requires explicit join definitions; implicit Cartesian products become invalid                                                                     |
| `collect()`, `list()` (Neo4j APOC)                             | Use `count()`, `max()`, `min()`, `avg()`, `sum()` or compute in Postgres                                                         | Drasi Cypher is a subset; doesn't include full Neo4j APOC library                                                                                        |
| `COUNT(x)`, `SUM(x)`, `MAX(x)`, `MIN(x)`, `AVG(x)` (uppercase) | `count(x)`, `sum(x)`, `max(x)`, `min(x)`, `avg(x)` (lowercase)                                                                   | **Drasi function names are case-sensitive.** Uppercase variants are treated as unknown user-defined functions → `UnknownFunction("COUNT")` TerminalError |

> [!WARNING]
> **Aggregation functions are case-sensitive.** `COUNT()`, `SUM()`, `MAX()`, `MIN()`, `AVG()` (uppercase) cause `UnknownFunction` errors. Always use lowercase: `count()`, `sum()`, `max()`, `min()`, `avg()`. Add a CI validation step: `grep -E '(COUNT|SUM|MAX|MIN|AVG)\(' queries/ && echo "ERROR: Use lowercase aggregation functions" && exit 1`

| `max(node.timestamp_column)` (raw timestamp/datetime column) | `max(drasi.changeDateTime(node))` | `max()` does not accept raw timestamp or datetime column values — causes `FunctionError: Max InvalidArgument(0)`. Use `drasi.changeDateTime(node)` to get a temporal value that max() can compare |
| `toFloat(a) / toFloat(b)` in WITH/RETURN without a zero guard | `CASE WHEN b > 0 THEN toFloat(a) / toFloat(b) ELSE 0.0 END` | WHERE guards (e.g. `WHERE b >= 5`) **do not prevent DivideByZero** during reactive re-evaluation against all elements in the source. Always guard divisions with a CASE expression |

> [!WARNING]
> **WHERE guards do not prevent reactive division by zero.** Drasi re-evaluates expressions reactively without re-applying WHERE filters first. All division expressions MUST use an inline `CASE WHEN denominator > 0 THEN numerator / denominator ELSE 0 END` guard, regardless of WHERE clause filters.

**Real Incident (2026-02-22):**

- `sla-countdown`: Invalid `duration.between(drasi.changeDateTime(a), datetime.realtime()).seconds` → Fixed with `drasi.trueLater()` reactive trigger
- `delivery-retry-storm`: Implicit MATCH + WHERE join → Added `joins:` section with explicit relationship definition
- `approver-workload-monitor`: WHERE with datetime arithmetic → Removed time filter, use `max(drasi.changeDateTime(ar))` for tracking
- `delivery-success-rate`: Same datetime arithmetic error → Removed time filter, use aggregation
- `landslide-hotspot-risk`: 4 tables in `nodes:` but only 1 `joins:` definition → Simplified to 2-table join

**Real Incident (2026-02-23):**

- `delivery-retry-storm`, `approver-workload-monitor`, `delivery-success-rate`: All had uppercase `COUNT()`, `SUM()`, `MAX()`, `COLLECT()` → All lowercase → `delivery-retry-storm` then hit `UnknownFunction("collect")` because `collect()` is not supported even in lowercase; removed `collect()` from WITH entirely
- `approver-workload-monitor`: `max(ar.decided_at)` (raw timestamp column) → `FunctionError: Max InvalidArgument(0)` → Fixed with `max(drasi.changeDateTime(ar))`
- `delivery-success-rate`: Same `max()` on raw timestamp → same fix; then hit `DivideByZero` despite `WHERE totalAttempts >= 5` guard → Fixed by wrapping division in `CASE WHEN totalAttempts > 0 THEN ... ELSE 0.0 END`

**Prevention Checklist When Authoring Queries:**

1. ✅ Every function call checked against https://drasi.io/reference/query-language/drasi-custom-functions/
2. ✅ No `datetime.realtime()` arithmetic in WHERE (use `drasi.trueLater()` instead)
3. ✅ All multi-table patterns have explicit `joins:` section with `keys:` mappings
4. ✅ No unsupported APOC functions (`collect`, `list`, `reduce`, etc.)
5. ✅ All aggregation function names are **lowercase**: `count`, `sum`, `max`, `min`, `avg` — uppercase causes `UnknownFunction` TerminalError
6. ✅ No `max()`/`min()` on raw timestamp or datetime columns — use `max(drasi.changeDateTime(node))` instead
7. ✅ Any division expression is guarded: `CASE WHEN divisor > 0 THEN ... / divisor ELSE 0.0 END` (WHERE guards do not prevent reactive DivideByZero)
8. ✅ Run `drasi apply -f query.yaml -n drasi-system` and verify: `drasi list query -n drasi-system | grep <name>` shows `Running` (not `TerminalError` or missing)
9. ✅ Wait ~10s post-apply before assuming success (async processing)

---

## 1. Non-negotiable Rules

### 1.1 Resource Format & Deployment

- All Drasi resources **must be YAML**.
- Always apply resources using the **Drasi CLI** (not `kubectl apply`): `drasi apply -f <file>.yaml -n <namespace>`.
- After apply, gate on readiness when possible: `drasi wait -f <file>.yaml -n <namespace> -t <seconds>`.
- Store YAML in version control; never apply ad-hoc changes.

### 1.2 Prefer Native Drasi Features

- Use Drasi query semantics and **Drasi custom functions** (e.g., temporal/change detection).
- Use built-in Sources/Reactions unless a provider is missing.

### 1.3 Version-Dependent Features: Verify First

- If a feature may depend on Drasi version (query language, functions, payloads), document a verification plan:
  ```text
  [VERIFY]
  EvidenceType = Docs | ReleaseNotes | Issue | Repro
  WhereToCheck = <URL, repo, command, or repro steps>
  ```
- Always set `queryLanguage` explicitly (`Cypher` or `GQL`) in every ContinuousQuery; do not rely on defaults.

### 1.4 Function Verification Requirement (Critical)

**Before using ANY Cypher function in a Drasi query, verify it exists in official Drasi documentation:**

- https://drasi.io/reference/query-language/cypher/
- https://drasi.io/reference/query-language/drasi-custom-functions/

Also verify the **function signature and return type** (some functions return `drasi.awaiting`), and any **prerequisites** (e.g., temporal functions require a temporal Element Index).

**If the function is NOT listed in Drasi docs, do NOT use it.** Drasi's Cypher dialect is a subset of Neo4j Cypher and does NOT support:

- **PostGIS/geospatial functions**: `ST_INTERSECTS`, `ST_GeomFromText`, `ST_Within`, `ST_Overlaps`, `ST_Contains`, etc.
- **Full APOC library**
- **Neo4j-specific functions** not explicitly documented by Drasi

**Workaround patterns for unsupported features:**
| Unsupported Feature | Workaround |
|---------------------|------------|
| Geospatial intersection (`ST_INTERSECTS`) | Pre-compute `region_code` during ingestion, query by region equality |
| Complex spatial queries | Use Postgres materialized views with PostGIS, expose results as Drasi source labels |
| Advanced aggregations | Compute in Postgres, expose as view/table for Drasi to query |

**Capabilities change frequently** — always re-verify against current docs before assuming a function works.

### 1.5 Provider Schema Check (Required)

- Before authoring Sources/Reactions, confirm provider schema via CLI:
  - `drasi describe sourceprovider <name>`
  - `drasi describe reactionprovider <name>`
- Do **not** assume property names (`timeout` vs `timeoutSeconds`, `token` vs `headers`) without verification.

### 1.6 CDC Prerequisites (PostgreSQL)

- CDC sources require target tables to exist **with primary keys** before applying queries.
- **Required property**: PostgreSQL sources must specify `tables` array listing tables to track:
  ```yaml
  properties:
    tables:
      - public.alerts
      - public.areas
  ```
- If tables are created via EF migrations, ensure migrations run **before** `drasi apply`.
  - Common failure: `No primary key found for <table>` → Source unavailable → Queries TerminalError.

### 1.7 Delete-Before-Update Requirement (Critical)

- Drasi resources **cannot be updated in place**. It's fine to test a query with a different name, but to modify an existing Source, ContinuousQuery, or Reaction with the same name:
  1. Delete the existing resource: `drasi delete <kind> <name> -n <namespace>`
  2. Apply the updated resource: `drasi apply -f <file>.yaml -n <namespace>`
- **Failure mode:** `drasi apply` on an existing resource fails with "already configured" or similar error.
- **CI/CD implication:** Pipelines must implement delete-then-apply logic for updates, not just `drasi apply`.

### 1.8 Out of Scope

This skill does **not** apply to:

- General Cypher or Neo4j queries outside Drasi
- Event streaming or broker configuration not consumed by Drasi
- Downstream application logic beyond Reaction boundaries
- Schema design decisions not represented as Drasi labels/properties

---

## 2. Authoritative Vocabulary

- **Source**: Connects to a system, emits change events.
- **ContinuousQuery**: Runs continuously, maintains live result set, emits changes to Reactions.
- **Reaction**: Consumes query result changes, performs actions (push to queue/webhook/db/etc).
- **Providers**: Plug-in types (SourceProvider, ReactionProvider).
- **Bootstrap Providers** (where supported): Deliver initial data to a ContinuousQuery before streaming begins (e.g., PostgreSQL COPY, ScriptFile JSONL, Platform, Noop).
- **drasi-server**: Standalone Rust binary for non-K8s deployments (REST API, YAML config, Docker Compose).
- **drasi-lib / drasi-core**: Embeddable Rust library for in-process continuous queries.
- **drasi-core-python**: Python bindings (PyO3) for embedding Drasi in Python services.
- **Dynamic Plugin**: Source/reaction loaded at runtime from OCI artifacts or shared libraries (.so/.dylib/.dll).
  > Use these terms consistently in docs and comments.

---

## 3. Recommended Repository Layout

**Prefer one YAML file per resource** for better version control, independent deployments, and cleaner delete-before-apply workflows.

```
infrastructure/
  drasi/
    sources/
      postgres-cdc.yaml               # one source per file (preferred)
    queries/
      emergency-alerts.yaml           # query pack (multi-doc YAML) OR one query per file
      sla-countdown.yaml
      operational-analytics.yaml
    reactions/
      emergency-alerts-http.yaml
    secrets/
      drasi-reaction-auth.yaml
    README.md
```

### Why one file per resource?

| Benefit                 | Explanation                                                                                     |
| ----------------------- | ----------------------------------------------------------------------------------------------- |
| **Git history**         | `git log infrastructure/drasi/queries/severity-escalation.yaml` shows only that query's changes |
| **Code review**         | PRs show exactly which resource changed, not a diff buried in a large file                      |
| **Independent CI/CD**   | Deploy only the changed query; skip unchanged resources                                         |
| **Delete-before-apply** | `drasi delete query severity-escalation` matches file name exactly                              |
| **Ownership**           | Assign CODEOWNERS per query/reaction for domain-specific review                                 |

### Anti-pattern: Unintentional bundled multi-document YAML

Avoid bundling unrelated resources into one file **by accident**. If you intentionally ship a "query pack" (multi-document YAML), treat it as a unit and ensure automation deletes _all_ contained resources before re-applying (see `scripts/drasi-refresh.ps1` for an example approach).

Prefer one resource per file when:

- Queries change independently
- You want granular CI/CD (deploy only what changed)

If you choose a query pack, keep it cohesive (same domain) and name it explicitly (e.g., `emergency-alerts.yaml`).

Example of a query pack:

```yaml
# ✅ Query pack: multiple ContinuousQuery docs deployed together
---
name: query-1
...
---
name: query-2
...
```

This makes git diffs noisy and forces full-file deployment even for single-query changes.

### Apply order script example

```powershell
$ns = 'drasi-system'
Get-ChildItem infrastructure/drasi/sources/*.yaml | ForEach-Object { drasi apply -f $_.FullName -n $ns }
Get-ChildItem infrastructure/drasi/queries/*.yaml | ForEach-Object { drasi apply -f $_.FullName -n $ns }
Get-ChildItem infrastructure/drasi/reactions/*.yaml | ForEach-Object { drasi apply -f $_.FullName -n $ns }
```

---

## 4. Resource Definitions & Best Practices

### 4.1 Drasi Custom Functions (Use These!)

Drasi provides custom functions for change tracking and temporal logic. **Prefer these over raw Cypher.**

| Function                                                                   | Purpose                                            | Example                                                                                                    |
| -------------------------------------------------------------------------- | -------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| `drasi.changeDateTime(element)`                                            | When element was last changed                      | `WHERE drasi.changeDateTime(a) > datetime.realtime() - duration({ hours: 1 })`                             |
| `drasi.previousValue(expression[, default])`                               | Previous value (incl. unchanged-field changes)     | `WHERE t.state = 'active' AND drasi.previousValue(t.state) = 'pending'`                                    |
| `drasi.previousDistinctValue(expression[, default])`                       | Previous _distinct_ value (useful for transitions) | `WHERE t.state = 'active' AND drasi.previousDistinctValue(t.state) = 'pending'`                            |
| `drasi.trueLater(expr, timestamp)`                                         | Evaluate expression at future time                 | `drasi.trueLater(a.status = 'PendingApproval', a.created_at + duration({ minutes: 5 }))`                   |
| `drasi.trueUntil(expr, timestamp)`                                         | Ensure expr stays true until time                  | `drasi.trueUntil(a.status = 'PendingApproval', a.created_at + duration({ minutes: 5 }))`                   |
| `drasi.trueFor(expr, duration)`                                            | Ensure expr stays true for duration                | `drasi.trueFor(a.status = 'PendingApproval', duration({ minutes: 5 }))`                                    |
| `drasi.slidingWindow(duration, aggregation_expression)`                    | Windowed aggregations                              | `drasi.slidingWindow(duration({ seconds: 10 }), avg(p.Value))`                                             |
| `drasi.linearGradient(x, y)`                                               | Slope of best-fit line                             | `drasi.linearGradient(p.x, p.y)`                                                                           |
| `drasi.getVersionByTimestamp(element, timestamp)`                          | Historical element snapshot (temporal index)       | `drasi.getVersionByTimestamp(a, datetime.realtime() - duration({ minutes: 30 }))`                          |
| `drasi.getVersionsByTimeRange(element, from, to, include_initial_version)` | All versions in time range (temporal index)        | `drasi.getVersionsByTimeRange(a, datetime.realtime() - duration({ hours: 1 }), datetime.realtime(), true)` |
| `drasi.listMin(list)`                                                      | Minimum value in list                              | `drasi.listMin([45, 33, 66])` → `33`                                                                       |
| `drasi.listMax(list)`                                                      | Maximum value in list                              | `drasi.listMax([drasi.changeDateTime(a), drasi.changeDateTime(b)])`                                        |

**Notes:**

- `drasi.trueLater` / `drasi.trueUntil` / `drasi.trueFor` may return `drasi.awaiting` while Drasi schedules a re-evaluation.
- TEMPORAL functions require a temporal Element Index.

#### Function Selection by Intent

Use this quick map when choosing Drasi custom functions:

| Intent                                     | Preferred Function(s)            | Why                                                            |
| ------------------------------------------ | -------------------------------- | -------------------------------------------------------------- |
| Detect state transitions                   | `drasi.previousDistinctValue`    | Tracks real state changes, not repeated same-value updates     |
| Detect any previous value                  | `drasi.previousValue`            | Includes unchanged-field change events                         |
| Trigger after deadline                     | `drasi.trueLater`                | Schedules future re-evaluation at an exact timestamp           |
| Ensure condition held for duration         | `drasi.trueFor`                  | Expresses SLA-style dwell-time conditions                      |
| Ensure condition remains true until cutoff | `drasi.trueUntil`                | Guards "must remain true until X" checks                       |
| Time-window aggregation                    | `drasi.slidingWindow`            | Avoids manual window state management                          |
| Trend slope                                | `drasi.linearGradient`           | Computes trend direction/strength for numeric signals          |
| Point-in-time snapshot                     | `drasi.getVersionByTimestamp`    | Retrieves historical element version (temporal index required) |
| History over interval                      | `drasi.getVersionsByTimeRange`   | Supports audit/debug over bounded time windows                 |
| Reduce list values                         | `drasi.listMin`, `drasi.listMax` | Useful for projected list comparisons                          |

#### Function Behavior Guardrails

- Prefer `drasi.changeDateTime(node)` for temporal aggregation with `max()`/`min()` instead of raw timestamp fields.
- Treat `drasi.awaiting` as a third state for FUTURE functions; do not assume immediate boolean output.
- Use explicit zero guards in arithmetic (`CASE WHEN divisor > 0 THEN ...`) because reactive re-evaluation can still hit divide-by-zero.
- Re-verify function signatures in docs before merge; Drasi function availability is version-sensitive.

**Example: Detect state transition**

```cypher
MATCH (o:Order)
WHERE o.status = 'Approved'
  AND drasi.previousDistinctValue(o.status) = 'Pending'
RETURN o.id AS orderId, o.status AS newStatus
```

**Example: SLA countdown with future evaluation**

```cypher
MATCH (a:Alert)
WHERE a.status = 'PendingApproval'
  AND drasi.trueLater(a.status = 'PendingApproval', drasi.changeDateTime(a) + duration({ minutes: 30 }))
RETURN a.alert_id AS alertId, a.headline AS headline
```

### 4.2 Cross-Source Joins (Synthetic Relationships)

Query across multiple sources using synthetic `joins`:

```yaml
apiVersion: v1
kind: ContinuousQuery
name: order-delivery
spec:
  sources:
    subscriptions:
      - id: retail-db
        nodes:
          - sourceLabel: orders
      - id: logistics-db
        nodes:
          - sourceLabel: vehicles
    joins:
      - id: PICKUP_BY
        keys:
          - label: orders
            property: plate
          - label: vehicles
            property: plate
  query: |
    MATCH (o:orders)-[:PICKUP_BY]->(v:vehicles)
    WHERE o.status = 'ready' AND v.location = 'Curbside'
    RETURN o.id AS orderId, v.make AS vehicleMake
```

**Key points:**

- `joins` create synthetic relationships between sources with no natural connection
- Keys must match property values exactly (case-sensitive)
- Each source must be subscribed with its labels

#### Supported Graph Model (Operational Baseline)

In this skill, queries operate on a property graph composed from Drasi source subscriptions:

- **Nodes** come from subscribed `sourceLabel` entries.
- **Edges** come from explicit `joins` definitions (synthetic relationships).
- **Relationship names** are the `joins.id` values used in `MATCH` patterns.
- **Cross-source traversal** requires each hop to be declared in `joins`; no implicit joins.

Reliable pattern examples:

```cypher
// Single-label pattern
MATCH (a:alerts)
WHERE a.status = 'Open'
RETURN a.alert_id
```

```cypher
// Explicit synthetic relationship from join id
MATCH (o:orders)-[:PICKUP_BY]->(v:vehicles)
RETURN o.id, v.make
```

```cypher
// Multi-hop requires each edge to be defined in joins
MATCH (o:orders)-[:PICKUP_BY]->(v:vehicles)-[:LOCATED_IN]->(d:depots)
RETURN o.id, d.region
```

Patterns to avoid in this baseline:

- Implicit joins (`MATCH (a:t1), (b:t2) WHERE a.id = b.id`)
- Variable-length path traversal (`-[*1..3]->`)
- Undeclared relationship types not backed by `joins`

### 4.3 Source Catalog

| Source Kind    | Runtime     | Description                                           | Key Properties                                           |
| -------------- | ----------- | ----------------------------------------------------- | -------------------------------------------------------- |
| **PostgreSQL** | K8s, Server | CDC via logical replication (WAL)                     | `host`, `port`, `user`, `password`, `database`, `tables` |
| **Cosmos DB**  | K8s         | Azure Cosmos DB change feed                           | Connection string, database, container                   |
| **Dataverse**  | K8s         | Microsoft Dataverse change tracking                   | Environment URL, entities                                |
| **Event Hub**  | K8s         | Azure Event Hub consumer                              | Connection string, consumer group                        |
| **Kubernetes** | K8s         | K8s resource watch (Pods, Deployments, etc.)          | Kubeconfig, resource types                               |
| **Relational** | K8s         | MySQL / SQL Server CDC                                | Connection string, tables                                |
| **HTTP**       | Server      | Webhook receiver with Handlebars mapping, HMAC auth   | Endpoint config, mapping template                        |
| **gRPC**       | Server      | gRPC streaming source                                 | Service address, proto definition                        |
| **Mock**       | Server      | Synthetic test data (counter, sensorReading, generic) | Generator type, interval                                 |
| **SDK**        | K8s         | Custom source via Source SDK                          | Implementation-dependent                                 |

### 4.3a Source Example (PostgreSQL)

```yaml
apiVersion: v1
kind: Source
name: alerts-source
spec:
  kind: PostgreSQL
  properties:
    host: ${POSTGRES_HOST}
    port: 5432
    user: ${POSTGRES_USER}
    database: alerts-db
    ssl: false
    tables: # Required: list tables to track
      - public.alerts
      - public.areas
    password: ${POSTGRES_PASSWORD} # injected at deploy time (do not commit real values)
    # OR (if your provider supports it):
    # password:
    #   kind: Secret
    #   name: pg-creds
    #   key: password
```

**Key points:**

- `tables` property is **required** - lists specific tables for CDC
- Labels in queries match table names (case-sensitive): `public.alerts` → label `alerts`
- Primary keys required on all tracked tables

### 4.4 ContinuousQuery Example

```yaml
apiVersion: v1
kind: ContinuousQuery
name: my-query
spec:
  mode: query
  sources:
    subscriptions:
      - id: my-source
        nodes:
          - sourceLabel: items
  queryLanguage: Cypher
  query: |
    MATCH (n:items)
    WHERE n.is_active = true
    RETURN n.id AS id
```

**Best Practice:** Prefer Drasi change semantics:

```cypher
MATCH (o:Order)
WHERE o.status = "Approved"
  AND drasi.previousDistinctValue(o.status) = "Pending"
RETURN o.id
```

### 4.5 Debug Reaction (Development/Testing)

Use Debug Reaction for testing queries - provides a web UI showing live results:

```yaml
apiVersion: v1
kind: Reaction
name: alerts-debug
spec:
  kind: Debug
  queries:
    pending-approval:
    severity-escalation:
    geographic-correlation:
```

**Access:** After deploying, port-forward to view the UI:

```bash
kubectl port-forward svc/alerts-debug 8080:8080 -n drasi-system
# Open http://localhost:8080
```

### 4.6 HTTP Reaction (Production Webhooks)

Http Reactions support Handlebars templating with `{{before.*}}` and `{{after.*}}`:

```yaml
apiVersion: v1
kind: Reaction
name: alerts-webhook
spec:
  kind: Http
  properties:
    baseUrl: ${DRASI_HTTP_REACTION_BASE_URL}
    timeout: 30000
  queries:
    severity-escalation: >
      added:
        url: /alerts/{{after.alertId}}/escalation
        method: POST
        headers:
          Content-Type: application/json
          X-Reaction-Token: ${DRASI_REACTION_TOKEN}
        body: |
          {
            "alertId": "{{after.alertId}}",
            "previousSeverity": "{{after.previousSeverity}}",
            "newSeverity": "{{after.escalatedSeverity}}"
          }
      deleted:
        url: /alerts/{{before.alertId}}/de-escalation
        method: DELETE
```

**Event types:**

- `added`: Row entered query results
- `updated`: Row changed while in results (use `{{before.*}}` and `{{after.*}}`)
- `deleted`: Row left query results

**Output formats** (for queue/event reactions):

- `packed`: Single message with `addedResults`, `updatedResults`, `deletedResults` arrays
- `unpacked`: Individual messages per change with `op: "i"/"u"/"d"`

**Idempotency:** Reactions are at-least-once. Design handlers for duplicate delivery.

### 4.7 Reaction Catalog

| Reaction Kind        | Runtime     | Description                                                                                       |
| -------------------- | ----------- | ------------------------------------------------------------------------------------------------- |
| **Debug**            | K8s         | Web UI showing live query results (development/testing)                                           |
| **Http**             | K8s, Server | Webhook with Handlebars templating (`added`/`updated`/`deleted`)                                  |
| **SignalR**          | K8s         | Real-time push to SignalR clients                                                                 |
| **MCP**              | K8s         | **Exposes query results via Model Context Protocol** — enables AI agents to consume reactive data |
| **sync-vectorstore** | K8s         | **Synchronizes query results to vector stores** — enables RAG/AI search over reactive data        |
| **Azure**            | K8s         | Azure Event Grid, Storage Queue, and other Azure service integrations                             |
| **AWS**              | K8s         | AWS service integrations                                                                          |
| **Dapr**             | K8s         | Dapr pub/sub integration                                                                          |
| **Debezium**         | K8s         | Debezium CDC format output                                                                        |
| **Gremlin**          | K8s         | Graph database write-back (Gremlin API)                                                           |
| **Power Platform**   | K8s         | Microsoft Power Platform / Dataverse integration                                                  |
| **SQL**              | K8s, Server | Stored procedure execution (PostgreSQL, MySQL, SQL Server)                                        |
| **SSE**              | Server      | Server-Sent Events streaming                                                                      |
| **gRPC**             | Server      | gRPC streaming with adaptive batching                                                             |
| **Log**              | Server      | Handlebars template logging                                                                       |
| **SDK**              | K8s         | Custom reaction via Reaction SDK                                                                  |

> The **MCP** and **sync-vectorstore** reactions are significant for AI/agent workloads. See §4.10 AI/Agent Integration below.

### 4.8 Reaction Reliability Contract (Required)

For production HTTP reactions, enforce these downstream handler rules:

- **Timeout budget:** keep end-to-end reaction handling under the configured reaction timeout.
- **Retry policy:** retry transient failures (`429`, `502`, `503`, `504`) with bounded exponential backoff + jitter.
- **Circuit breaker:** stop immediate retries after repeated failures and expose degraded mode status.
- **Idempotency key:** derive a stable key from query + operation + business identifier.
- **Dead-letter path:** when retries are exhausted, persist payload + reason for replay and audit.
- **ACK discipline:** only return success after side effects are committed (DB/message publish complete).

Example idempotency key pattern:

```text
<query-name>:<op>:<business-id>:<event-time-bucket>
```

### 4.9 Reaction Token Rotation (Zero-Downtime)

If you use `X-Reaction-Token` (or equivalent auth header), rotate with overlap:

1. Add `token_next` in backend config while retaining `token_current`.
2. Update Drasi reaction secret to `token_next`; re-apply reaction.
3. Verify callbacks succeed with `token_next` in logs/metrics.
4. Remove `token_current` after a stability window.

Never perform single-step token replacement in both systems simultaneously.

### 4.10 Logging and Data Minimization

- Never log full reaction headers or raw auth tokens.
- Redact high-risk fields (email, phone, address, secrets) before writing logs.
- Log correlation IDs, query name, operation, and status code for troubleshooting.
- Keep payload-body logging behind short-lived debug flags and disable by default in production.

### 4.11 AI and Agent Integration

Drasi's ecosystem includes first-class support for AI agent workflows:

| Component            | Package/Reaction                     | Purpose                                                                                                      |
| -------------------- | ------------------------------------ | ------------------------------------------------------------------------------------------------------------ |
| **MCP Reaction**     | `reaction.platform/MCP`              | Exposes continuous query results via Model Context Protocol; AI agents consume reactive data without polling |
| **sync-vectorstore** | `reaction.platform/sync-vectorstore` | Keeps vector stores (Azure AI Search, etc.) in sync with query results for RAG patterns                      |
| **langchain-drasi**  | `pip install langchain-drasi`        | Python bridge for LangChain/LangGraph agents; 6 built-in notification handlers                               |

**When to use each:**

- **MCP Reaction**: Agent needs live query results as tool context (pairs with Microsoft Agent Framework / `foundry-mcp`).
- **sync-vectorstore**: RAG pipeline where embeddings must track changing source data automatically.
- **langchain-drasi**: Python agent orchestration with LangChain/LangGraph that needs Drasi change notifications.

```yaml
# MCP Reaction example
kind: Reaction
apiVersion: v1
name: agent-context
spec:
  kind: MCP
  queries:
    active-alerts:
      query: active-severity-alerts
```

### 4.12 Custom Sources and Reactions

Build custom connectors when the built-in catalogs (4.3/4.7) don't cover your data source or downstream system.

**SDK Availability:**

| Role     | Language   | Package               | Install                           |
| -------- | ---------- | --------------------- | --------------------------------- |
| Source   | Rust       | `drasi-source-sdk`    | `cargo add drasi-source-sdk`      |
| Source   | Java       | `io.drasi:source.sdk` | Maven Central                     |
| Source   | .NET       | `Drasi.Source.SDK`    | NuGet                             |
| Reaction | TypeScript | `@drasi/reaction-sdk` | `npm install @drasi/reaction-sdk` |
| Reaction | Python     | `drasi_reaction_sdk`  | `pip install drasi-reaction-sdk`  |
| Reaction | .NET       | `Drasi.Reaction.SDK`  | NuGet                             |

**Event contracts (all SDKs share this shape):**

- **ChangeEvent**: `sequence`, `queryId`, `addedResults[]`, `deletedResults[]`, `updatedResults[]{before, after}`
- **ControlEvent**: `sequence`, `queryId`, `controlSignal` (`started` | `stopped` | `paused`)

**Minimal custom reaction (TypeScript):**

```typescript
import { DrasiReaction } from "@drasi/reaction-sdk";

const reaction = new DrasiReaction((event) => {
  for (const added of event.addedResults) {
    console.log("New result:", added);
  }
});
reaction.start();
```

**Build and deploy pattern:**

1. Implement handler using the SDK for your language
2. Package as OCI container image (Dockerfile)
3. Push to your container registry (ACR recommended)
4. Register via `kind: Reaction` / `kind: Source` YAML with `spec.kind: SDK` and your image reference
5. Requires **Dapr v1.14+** sidecar for pub/sub communication between sources, query containers, and reactions

---

## 5. CLI Workflow

**Apply order (dependency-driven):**

1. Sources (must exist before queries reference them)
2. Queries (must exist before reactions subscribe to them)
3. Reactions

**Per-file apply (recommended):**

```bash
# Apply a single resource
drasi apply -f infrastructure/drasi/sources/postgres-cdc.yaml -n <namespace>

# Apply all resources in order
for f in infrastructure/drasi/sources/*.yaml; do drasi apply -f "$f" -n <namespace>; done
for f in infrastructure/drasi/queries/*.yaml; do drasi apply -f "$f" -n <namespace>; done
for f in infrastructure/drasi/reactions/*.yaml; do drasi apply -f "$f" -n <namespace>; done
```

**Delete-before-apply (for updates):**

```bash
# Update an existing query (query pack example)
drasi delete query severity-escalation -n <namespace>
drasi apply -f infrastructure/drasi/queries/emergency-alerts.yaml -n <namespace>
```

**Readiness gate:** Before applying, ensure the control plane is ready:

- Namespace exists, deployments are available, and `drasi list source -n <namespace>` succeeds.
- If not, run `drasi init -n <namespace>` and wait for deployments.

**Avoid folder apply:** `drasi apply -f infrastructure/drasi/queries/` may fail on some platforms. Prefer explicit file paths.

### 5.1 CLI / Control Plane Compatibility Matrix (Mandatory)

Track and validate compatibility at deploy time:

| Layer               | Pinning Guidance                                  | Verify Command                                                                    | Failure Signal                                                           |
| ------------------- | ------------------------------------------------- | --------------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
| Drasi CLI           | Pin explicit CLI version in CI image/tool cache   | `drasi version`                                                                   | CLI command shape differs or apply/list fails unexpectedly               |
| Drasi control plane | Track platform release per environment            | `kubectl get deploy -n <namespace>` + platform release notes                      | `404`, schema validation drift, or inconsistent status reporting         |
| Query schema        | Pin `apiVersion: v1` and explicit `queryLanguage` | `drasi apply -f <query>.yaml -n <namespace>`                                      | Query accepted but enters `TerminalError` quickly                        |
| Providers           | Pin provider image/version where self-hosted      | `drasi describe sourceprovider <name>` / `drasi describe reactionprovider <name>` | Property mismatch (`timeout` vs `timeoutSeconds`, header key mismatches) |

Record these values in pipeline logs for every deploy.

---

## 6. CI/CD Automation Patterns

### Per-file change detection

```yaml
# GitHub Actions example - only deploy changed queries
- name: Get changed Drasi files
  id: changed
  run: |
    echo "queries=$(git diff --name-only ${{ github.event.before }} HEAD -- 'infrastructure/drasi/queries/*.yaml' | tr '\n' ' ')" >> $GITHUB_OUTPUT

- name: Deploy changed queries
  if: steps.changed.outputs.queries != ''
  run: |
    for file in ${{ steps.changed.outputs.queries }}; do
      query_name=$(yq '.name' "$file")
      drasi delete query "$query_name" -n <namespace> || true  # ignore if doesn't exist
      drasi apply -f "$file" -n <namespace>
    done
```

### Best practices

- **Skip unchanged:** Only apply resources with actual YAML diffs
- **Delete-then-apply:** Always delete before apply for updates (see 1.7)
- **Capture version:** Log `drasi version` at start of pipeline for debugging
- **Readiness check:** Wait for `drasi list source -n <namespace>` to succeed before applying queries
- **Idempotent deletes:** Use `|| true` on delete commands (resource may not exist on first deploy)

### 6.1 Static Query Guardrails (Fail Fast)

Before `drasi apply`, run static checks on changed query YAML files:

```bash
#!/usr/bin/env bash
set -euo pipefail

changed_queries=$(git diff --name-only "${BASE_SHA:-HEAD~1}" HEAD -- 'infrastructure/drasi/queries/*.yaml')
[ -z "$changed_queries" ] && exit 0

for file in $changed_queries; do
  # Require explicit query language and apiVersion.
  rg -q '^apiVersion:\\s*v1\\s*$' "$file" || { echo "Missing apiVersion: v1 in $file"; exit 1; }
  rg -q 'queryLanguage:\\s*(Cypher|GQL)' "$file" || { echo "Missing queryLanguage in $file"; exit 1; }

  # Block unsupported/known-failing function patterns.
  rg -q 'collect\\s*\\(' "$file" && { echo "collect() unsupported in $file"; exit 1; }
  rg -q 'duration\\.between\\s*\\(' "$file" && { echo "duration.between() unsupported in $file"; exit 1; }

  # Drasi function names are case-sensitive; uppercase aggregates are usually invalid.
  rg -q '\\b(COUNT|SUM|MAX|MIN|AVG)\\s*\\(' "$file" && { echo "Uppercase aggregate in $file"; exit 1; }
done
```

Add this as a required CI stage before deployment.

### Environment-specific configuration

For multi-environment deployments (dev/staging/prod), use one of these patterns:

**Option A: Environment substitution (recommended)**
Keep one set of YAML files with placeholders, substitute at deploy time:

```yaml
# infrastructure/drasi/sources/postgres-cdc.yaml
name: postgres-alerts
spec:
  kind: PostgreSQL
  properties:
    host: ${POSTGRES_HOST}
    password:
      kind: Secret
      name: ${ENV}-pg-creds
      key: password
```

```bash
envsubst < infrastructure/drasi/sources/postgres-cdc.yaml | drasi apply -f - -n <namespace>
```

**Option B: Environment overlays**
Separate folders per environment (more files, but explicit):

```
infrastructure/drasi/
  base/queries/pending-approval.yaml
  overlays/
    dev/sources/postgres-alerts.yaml
    prod/sources/postgres-alerts.yaml
```

### Naming conventions

| Resource Type | Naming Pattern       | Example                                         |
| ------------- | -------------------- | ----------------------------------------------- |
| Source        | `<system>-<purpose>` | `postgres-alerts`, `cosmos-users`               |
| Query         | `<domain>-<pattern>` | `severity-escalation`, `geographic-correlation` |
| Reaction      | `<query>-<action>`   | `alerts-http-webhook`, `correlation-signalr`    |

Use lowercase kebab-case. Resource names should match filenames exactly for traceability.

---

## 7. Troubleshooting Checklist

### Query won't start / TerminalError

1. Check source is ready: `drasi list source -n <namespace>` → `AVAILABLE` should be `true`
2. Verify labels exist in source data (case-sensitive match)
3. Check for unsupported Cypher (see 1.4)
4. View query status/logs: `drasi describe query <name> -n <namespace>`

### TerminalError mentions Dapr publish / `127.0.0.1:3500` (common)

If you see errors like:
`Error publishing the subscription event ... http://127.0.0.1:3500/v1.0/publish/... Connection refused`

Notes:

- `127.0.0.1:3500` is the **Dapr sidecar HTTP API** (default port).
- Drasi query errors can be **sticky** (status/errorMessage may reflect a previous failure) until you **delete + re-apply** the query (see 1.7).
- Do not change `DAPR_HTTP_PORT` unless you have verified the Dapr sidecar is listening on a non-default port.

Quick checks:

1. Confirm the query-host pod is healthy (example): `kubectl get pods -n drasi-system | grep default-query-host`
2. Confirm the Dapr sidecar port from logs:
   - `kubectl logs -n drasi-system <query-host-pod> -c daprd --tail=200`
   - Look for `HTTP server is running on port 3500`
3. Fix: restart query-host + refresh queries (delete → apply)
   - `kubectl rollout restart deployment/default-query-host -n drasi-system`
   - `drasi delete query <name> -n drasi-system` then `drasi apply -f <file>.yaml -n drasi-system`

### "Unsupported Cypher" or parse errors

- Verify function exists in [Drasi Cypher docs](https://drasi.io/reference/query-language/cypher/)
- Check for Neo4j/PostGIS functions that Drasi doesn't support (see 1.4)
- Validate WITH/RETURN column scoping (columns must be explicitly passed through WITH)
- Some Cypher operators/predicates may be version-dependent (e.g., string predicates like `CONTAINS`). Prefer exact matches or precomputed flags/tags when in doubt, and verify against Drasi docs for your platform version.

### Label mismatch / no results

- Source labels are derived from table names (case-sensitive)
- Verify: `drasi describe source <name> -n <namespace>` shows expected labels
- PostgreSQL: table `alerts` → label `alerts` (not `Alert`)

### Queries active, but no downstream reaction effects

1. Inspect reaction target: `drasi describe reaction <name> -n <namespace>` and confirm:
   - `baseUrl` includes your backend reaction route prefix
   - in-cluster DNS points to API service (`*.svc.cluster.local`)
2. Inspect reaction pod logs for callback failures/timeouts:
   - `kubectl logs -n <drasi-namespace> deploy/<reaction-deployment> --tail=200`
3. Verify token alignment:
   - Reaction auth header value equals backend-configured expected token (ConfigMap/Secret/app settings)
4. Inspect API logs for auth failures:
   - `kubectl logs -n <app-namespace> deploy/<api-deployment> --tail=200 | rg "Unauthorized|drasi/reactions|reaction"`
5. Run a scenario and verify downstream endpoints return data:
   - Example: `GET /api/v1/drasi/events/recent`
   - Example: `GET /api/v1/operator/inbox`

### Apply fails with "already configured"

- Drasi resources can't be updated in place (see 1.7)
- Fix: `drasi delete <kind> <name> -n <namespace>` then `drasi apply -f <file>.yaml -n <namespace>`

### Source not ready / CDC errors

- Ensure target tables exist **with primary keys** before applying source
- For EF Core: run migrations before `drasi apply`
- Check: `drasi describe source <name> -n <namespace>` for detailed error

### CLI returns 404 or version mismatch

- Verify Drasi CLI version matches platform: `drasi version`
- Reinstall CLI matching your cluster's Drasi version

---

## 8. Copilot Guardrails

**DO:**

- Use Drasi custom functions (DELTA/FUTURE/WINDOW/TEMPORAL) and handle `drasi.awaiting` for FUTURE functions
- Verify function support in docs before using (see 1.4)
- Enable/verify temporal Element Index before using temporal functions (`drasi.getVersionByTimestamp`, `drasi.getVersionsByTimeRange`)
- Use Debug reactions during development to validate query output
- Project stable IDs in RETURN (for idempotent reaction handling)
- Always set `queryLanguage` explicitly (`Cypher` or `GQL`) in every ContinuousQuery
- Include `tables` property for PostgreSQL sources
- Check the Source Catalog (4.3) and Reaction Catalog (4.7) before assuming a connector does not exist
- Use the MCP Reaction for AI agent integration instead of polling query APIs (see 4.11)
- Use `context7` MCP to verify Drasi function support and API surface (not `microsoft.learn.mcp`)
- Use the appropriate SDK for custom sources (Rust/Java/.NET) and reactions (TS/Python/.NET) — check 4.12
- Package custom sources/reactions as OCI images and confirm Dapr sidecar is available before deploying

**DON'T:**

- Assume full Neo4j Cypher or PostGIS support
- Use functions not documented in Drasi docs (ST_INTERSECTS, APOC, etc.)
- Inline secrets (use Secret references)
- Reapply unchanged queries blindly
- **Assume `drasi apply` updates existing resources** — delete first (see 1.7)
- Use variable-length paths (`-[*1..3]->`) — not supported
- Implement custom sources/reactions without the official SDK for your language (see 4.12)

---

## 9. Quality Gates

**Query checklist:**

- [ ] All functions + signatures verified in Drasi docs
- [ ] FUTURE functions handle `drasi.awaiting` (no silent "false" assumptions)
- [ ] Temporal Element Index enabled if using temporal functions
- [ ] Uses Drasi custom functions for change detection where applicable
- [ ] Early filtering (WHERE clauses before expensive operations)
- [ ] Stable identifiers in RETURN
- [ ] Labels match source table names (case-sensitive)

**Source checklist:**

- [ ] Secrets referenced via `kind: Secret`
- [ ] `tables` property lists all tracked tables (PostgreSQL)
- [ ] Labels documented for query authors
- [ ] Source type exists in the Source Catalog (4.3); verify runtime compatibility (K8s vs Server)

**Reaction checklist:**

- [ ] Idempotency strategy (handlers tolerate duplicates)
- [ ] Retry policy documented for transient callback failures (429/502/503/504)
- [ ] Timeout budget defined and aligned with reaction `timeout` setting
- [ ] Circuit-breaker/degraded-mode behavior declared for repeated callback failure
- [ ] Error handling for downstream failures
- [ ] Observability (logs, metrics for reaction processing)
- [ ] Reaction `baseUrl` includes the backend reaction route prefix and is reachable from the Drasi namespace
- [ ] Reaction auth header token matches backend expected token/config
- [ ] Token rotation runbook tested (dual-token overlap, no downtime)
- [ ] Logs redact auth headers and PII fields by default
- [ ] ADAC: reaction delivery status and lag are declared (health endpoint or metrics)
- [ ] ADAC: UI/API surfaces degraded mode when reaction delivery fails or data is stale

**CI checklist:**

- [ ] Static query guardrails run before deployment (see 6.1)
- [ ] Deploy logs include CLI and control-plane compatibility evidence (see 5.1)

**Custom source/reaction checklist (when building SDK components):**

- [ ] Using the official Drasi SDK for the target language (see 4.12)
- [ ] ChangeEvent handler processes `addedResults`, `deletedResults`, and `updatedResults`
- [ ] ControlEvent handler reacts to `started`, `stopped`, `paused` signals
- [ ] Packaged as OCI container image with health endpoint
- [ ] Dapr v1.14+ sidecar confirmed available in the target namespace
- [ ] Image pushed to ACR and referenced in `spec.properties.container` of the YAML manifest

---

## 10. Known Issues for Automation

- ContinuousQuery re-apply may fail (Query already configured)
- CLI drasi apply may return 404 on version mismatch
- Default query language may change; queries without `queryLanguage` may break (set it explicitly; see issue #385)
- Not all reactions support Handlebars templating (verify provider capabilities; see issue #345)
- Temporal/index configuration is evolving (verify current docs before relying on temporal functions; see issue #377)
- Custom providers must guard against injection
- Folder apply may fail on some platforms: prefer file path (`drasi apply -f <file>.yaml`)

---

## 11. References

- https://drasi.io/
- https://drasi.io/reference/query-language/cypher/
- https://drasi.io/reference/query-language/gql/
- https://drasi.io/reference/query-language/drasi-custom-functions/
- https://drasi.io/concepts/sources/
- https://drasi.io/concepts/reactions/
- https://drasi.io/how-to-guides/configure-reactions/configure-drasi-debug-reaction/
- https://drasi.io/how-to-guides/configure-reactions/configure-http-reaction/
- https://github.com/drasi-project
- https://github.com/drasi-project/drasi-platform
- https://github.com/drasi-project/drasi-core
- https://github.com/drasi-project/drasi-server
- https://github.com/drasi-project/drasi-core-python
- https://github.com/drasi-project/langchain-drasi
- https://github.com/drasi-project/grafana-signalr
- https://github.com/drasi-project/design-documents
- https://github.com/drasi-project/docs
- https://github.com/drasi-project/learning
- https://github.com/drasi-project/drasi-platform/issues/308
- https://github.com/drasi-project/drasi-platform/issues/345
- https://github.com/drasi-project/drasi-platform/issues/377
- https://github.com/drasi-project/drasi-platform/issues/376
- https://github.com/drasi-project/drasi-platform/issues/385
- https://github.com/drasi-project/drasi-core/issues/206
- https://www.npmjs.com/package/@drasi/reaction-sdk
- https://pypi.org/project/drasi-reaction-sdk/
- https://crates.io/crates/drasi-source-sdk

---

## 12. Currency and Verification

- **Date checked:** 2026-06-23
- **Status:** Drasi is in early release (CNCF Sandbox) with an expanding ecosystem; APIs, Cypher subset, and CLI behavior may change between releases without deprecation notices.
- **Ecosystem breadth:** 10 source types, 17 reaction types (including MCP and sync-vectorstore for AI), 4 deployment modes (K8s platform, Server binary, Rust embedded, Python embedded), 6 SDKs for custom source/reaction development (see 4.12).
- **Notes:** Drasi docs remain split by product (`drasi-lib`, `Drasi Server`, `Drasi for Kubernetes`); keep query/function assumptions tied to the runtime you are deploying.
- **Sources:** [Drasi documentation](https://drasi.io/), [Drasi custom functions](https://drasi.io/reference/query-language/drasi-custom-functions/), [Drasi platform releases](https://github.com/drasi-project/drasi-platform/releases)
- **MCP verification:** Use **`context7`** to query the latest Drasi documentation, API surface, and Cypher function support. Drasi is not indexed in `microsoft.learn.mcp`.
- **Verification steps:**
  1. Run `drasi version` and confirm CLI/platform compatibility with current releases.
  2. Run `drasi list` to validate API connectivity and schema compatibility in the target namespace.
  3. Validate Cypher against Drasi docs (not Neo4j assumptions) before merging query changes.
  4. Re-check release notes for `apiVersion` or query language changes before upgrades.
  5. Use `context7` to cross-check Cypher function availability and parameter signatures.

### Known Pitfalls

| Area                         | Pitfall                                                                                                | Mitigation                                                            |
| ---------------------------- | ------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------- |
| Cypher subset                | No `collect()`, no datetime arithmetic in WHERE, case-sensitive function names — not full Neo4j Cypher | Test queries against Drasi directly; never assume Neo4j compatibility |
| `apiVersion` format          | Format has changed between releases; YAML may silently fail validation                                 | Check release notes for schema changes when upgrading `drasi` CLI     |
| Multi-resource YAML          | Bundled YAML with `---` separators causes silent failures                                              | Use one resource per file (see Section 3)                             |
| Delete-before-update         | `drasi apply` does not support in-place updates for queries and reactions                              | Always `drasi delete` then `drasi apply` when changing resources      |
| Source connection validation | Drasi doesn't validate source connection strings at apply time                                         | Test source connectivity separately before applying dependent queries |
| Reaction auth rotation       | One-step token swaps can break callbacks mid-deploy                                                    | Use dual-token overlap rotation (Section 4.9)                         |
| Reaction logging             | Full payload/header logging can leak secrets or PII                                                    | Redact sensitive fields and never log raw auth headers                |

---

## 13. Next Steps (Optional)

- Track new source and reaction types as they are released; update catalogs in 4.3 and 4.7.
- Add version-gated sections (e.g., `>= 0.10.x`) when specific features require a minimum Drasi version.
- Expand AI/Agent Integration patterns as MCP Reaction and langchain-drasi mature.
- Add performance tuning guidance for high-volume queries when Drasi publishes benchmarks.
- Evaluate Grafana SignalR data source plugin for real-time operational dashboards.
