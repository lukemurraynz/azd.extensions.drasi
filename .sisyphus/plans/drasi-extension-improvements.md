# Drasi Azure Developer CLI Extension — Comprehensive Improvements

## TL;DR

> **Quick Summary**: Improve the Drasi azd extension across 17 identified areas covering functional gaps (diagnose stubs, deploy rollback, logs for all kinds), developer experience (describe command, progress indicators, interactive confirmations), code quality (externalize YAML, wire observability, remove dead code), and feature expansion (new templates, telemetry, environment overlays).
> 
> **Deliverables**:
> - 6 new/enhanced commands (diagnose, describe, status, logs, teardown, upgrade)
> - Deploy engine hardened with rollback and crash-safe locking
> - Observability wired into all commands via OpenTelemetry
> - NetworkPolicy YAML externalized to embedded files
> - PostgreSQL scaffold template added
> - Environment overlay working example in all templates
> - All changes covered by tests maintaining 80% coverage gate
> 
> **Estimated Effort**: Large
> **Parallel Execution**: YES — 5 waves
> **Critical Path**: Task 1 (shared helpers) → Tasks 2-8 (core improvements) → Tasks 9-15 (DX + features) → Tasks 16-17 (integration) → Final Verification

---

## Context

### Original Request
Analyze the Drasi azd extension product, features, blueprint skills, and identify improvements or capabilities that could be enhanced. Then plan for all 17 identified improvements.

### Interview Summary
**Key Discussions**:
- User confirmed all 17 improvements should be included in a single plan
- Extension is Go-based, uses azdext gRPC SDK, Cobra commands, testify for testing
- CI enforces 80% coverage, race detector, golangci-lint, Bicep validation

**Research Findings**:
- `yacspin` (spinners) and `survey/v2` (prompts) are already in go.mod as transitive dependencies
- OpenTelemetry packages (otel, metric, trace, OTLP exporter) are direct dependencies
- Test patterns use manual mocks (mockDrasiRunner, mockEnvClient), not gomock
- Deploy lock uses key `DRASI_DEPLOY_IN_PROGRESS` with values `"true"`/`""`
- `drasi watch` only supports query kind; other kinds need `kubectl logs` approach
- `DescribeComponent`/`DescribeComponentInContext` already exist in drasi client

### Metis Review
**Identified Gaps** (addressed):
- Deploy lock key is `DRASI_DEPLOY_IN_PROGRESS` (not `DRASI_DEPLOY_LOCK` as initially stated) — corrected
- `survey/v2` is archived but acceptable as existing transitive dependency
- Diagnose checks show "skipped" not "ok" — plan addresses with real implementations
- Observability requires OpenTelemetry wiring at PersistentPreRun level, not per-command

---

## Work Objectives

### Core Objective
Harden the Drasi azd extension by fixing functional gaps, improving developer experience, cleaning up code quality issues, and expanding the feature set to cover common Drasi deployment scenarios.

### Concrete Deliverables
- Enhanced `diagnose` command with real Key Vault and Log Analytics checks
- Deploy engine with rollback on failure and crash-safe locking
- `describe` command for single-component detail view
- `status` command showing all kinds by default
- `logs` command supporting all component kinds
- Progress indicators (yacspin) on long-running operations
- Real lifecycle hook implementations (postprovision health check, predeploy validation)
- Interactive confirmation prompts on destructive operations
- Externalized NetworkPolicy YAML via Go embed
- Wired OpenTelemetry observability across all commands
- PostgreSQL scaffold template
- Environment overlay working example
- Key Vault translator integration tests

### Definition of Done
- [ ] `go test ./... -race -count=1` passes with 80%+ coverage
- [ ] `golangci-lint run` passes
- [ ] All existing tests continue to pass (no regressions)
- [ ] Each new feature has unit tests covering happy path + error path

### Must Have
- Backward compatibility with existing `drasi.yaml` manifests
- All commands continue to work in `--output json` and `--output table` modes
- No new external dependencies (use existing go.mod dependencies only)
- Tests use manual mocks consistent with existing patterns (no gomock)
- Error codes use existing `output.ERR_*` constants

### Must NOT Have (Guardrails)
- Do NOT add new CLI flags to existing commands without documenting in README
- Do NOT modify the deploy ordering (sources→queries→middlewares→reactions)
- Do NOT change the hash algorithm or state key format (`DRASI_HASH_<KIND>_<ID>`)
- Do NOT break the `drasi watch` wrapper — only ADD alternatives for non-query kinds
- Do NOT add real Azure API calls to diagnose (use existing CLI wrappers only)
- Do NOT create full Bicep modules for the PostgreSQL template (reuse existing module patterns)
- Do NOT over-engineer rollback — simple "delete components applied in this run" is sufficient
- Do NOT add telemetry that phones home without user consent (respect existing APPLICATIONINSIGHTS_CONNECTION_STRING opt-in)
- Do NOT log secrets, Key Vault values, or PII in any new code paths

---

## Verification Strategy

> **ZERO HUMAN INTERVENTION** — ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: YES (Go test + testify)
- **Automated tests**: YES (tests-after — maintain 80% coverage)
- **Framework**: Go standard testing + github.com/stretchr/testify (assert, require)
- **Pattern**: Manual test doubles matching existing mockDrasiRunner/mockEnvClient style

### QA Policy
Every task MUST include agent-executed QA scenarios.
Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **CLI commands**: Use Bash — Run command with test args, assert stdout/stderr/exit code
- **Go packages**: Use Bash — `go test ./path/to/package -v -run TestName`
- **Formatting**: Use Bash — `go test ./... -race -count=1` for full suite

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Foundation — shared utilities and refactors):
├── Task 1: Shared progress helper + observability wiring [quick]
├── Task 2: Externalize NetworkPolicy YAML [quick]
├── Task 3: Remove applyDefaultProviders no-op + clean --follow flag [quick]
├── Task 4: Environment overlay example in templates [quick]
├── Task 5: Crash-safe deploy lock [quick]

Wave 2 (Core command improvements — MAX PARALLEL):
├── Task 6: Diagnose real checks (depends: none) [unspecified-high]
├── Task 7: Deploy rollback on failure (depends: 5) [deep]
├── Task 8: Describe command (depends: none) [unspecified-high]
├── Task 9: Status show all kinds (depends: none) [quick]
├── Task 10: Logs for all component kinds (depends: none) [unspecified-high]

Wave 3 (Developer experience):
├── Task 11: Interactive confirmations (depends: none) [quick]
├── Task 12: Progress indicators on commands (depends: 1) [unspecified-high]
├── Task 13: Lifecycle hook implementations (depends: 6) [unspecified-high]

Wave 4 (Feature expansion):
├── Task 14: PostgreSQL scaffold template (depends: 4) [unspecified-high]
├── Task 15: Telemetry via observability (depends: 1, 12) [unspecified-high]
├── Task 16: Key Vault translator integration tests (depends: none) [unspecified-high]

Wave FINAL (After ALL tasks — 4 parallel reviews, then user okay):
├── Task F1: Plan compliance audit (oracle)
├── Task F2: Code quality review (unspecified-high)
├── Task F3: Real manual QA (unspecified-high)
└── Task F4: Scope fidelity check (deep)
-> Present results -> Get explicit user okay

Critical Path: Task 1 → Task 12 → Task 15 → F1-F4 → user okay
Parallel Speedup: ~65% faster than sequential
Max Concurrent: 5 (Wave 2)
```

### Dependency Matrix

| Task | Depends On | Blocks | Wave |
|------|-----------|--------|------|
| 1 | — | 12, 15 | 1 |
| 2 | — | — | 1 |
| 3 | — | — | 1 |
| 4 | — | 14 | 1 |
| 5 | — | 7 | 1 |
| 6 | — | 13 | 2 |
| 7 | 5 | — | 2 |
| 8 | — | — | 2 |
| 9 | — | — | 2 |
| 10 | — | — | 2 |
| 11 | — | — | 3 |
| 12 | 1 | 15 | 3 |
| 13 | 6 | — | 3 |
| 14 | 4 | — | 4 |
| 15 | 1, 12 | — | 4 |
| 16 | — | — | 4 |

### Agent Dispatch Summary

- **Wave 1**: 5 tasks — T1-T5 → `quick`
- **Wave 2**: 5 tasks — T6 → `unspecified-high`, T7 → `deep`, T8 → `unspecified-high`, T9 → `quick`, T10 → `unspecified-high`
- **Wave 3**: 3 tasks — T11 → `quick`, T12 → `unspecified-high`, T13 → `unspecified-high`
- **Wave 4**: 3 tasks — T14-T16 → `unspecified-high`
- **FINAL**: 4 tasks — F1 → `oracle`, F2 → `unspecified-high`, F3 → `unspecified-high`, F4 → `deep`

---

## TODOs

- [ ] 1. Shared Observability Wiring + Progress Helper

  **What to do**:
  - Wire `internal/observability/tracer.go:NewTracer()` and `internal/observability/metrics.go:NewMeter()` into `cmd/root.go` via a `PersistentPreRunE` hook on the root command
  - Store the tracer and meter in a package-level variable or context so child commands can access them
  - Create `cmd/progress.go` with a helper function that wraps `yacspin` (already in go.mod line 87) to provide start/stop/message updates with automatic no-op when `--output json` is set
  - Add tests: verify tracer/meter initialization with and without APPLICATIONINSIGHTS_CONNECTION_STRING, verify progress helper suppresses output in JSON mode
  - Ensure `PersistentPostRunE` calls the shutdown functions returned by NewTracer/NewMeter

  **Must NOT do**:
  - Do NOT add new dependencies — yacspin and otel are already in go.mod
  - Do NOT make observability mandatory — gracefully degrade to no-op when env var is absent (this is already how tracer.go/metrics.go work)
  - Do NOT write output to stdout from the progress helper (use stderr only, per azd extension conventions)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: [`creating-azd-extensions`]
    - `creating-azd-extensions`: Covers root command patterns and PersistentPreRunE conventions

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 2, 3, 4, 5)
  - **Blocks**: Tasks 12, 15
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `cmd/root.go` — Root command setup; add PersistentPreRunE/PostRunE here
  - `internal/observability/tracer.go:21-53` — NewTracer() returns (trace.Tracer, shutdown, error); uses APPLICATIONINSIGHTS_CONNECTION_STRING
  - `internal/observability/metrics.go:15-27` — NewMeter() returns (metric.Meter, shutdown, error); same env var pattern

  **API/Type References**:
  - `go.mod:87` — `github.com/theckman/yacspin v0.13.12` (spinner library)
  - `go.mod:12-17` — OpenTelemetry direct dependencies

  **External References**:
  - yacspin API: https://pkg.go.dev/github.com/theckman/yacspin

  **Acceptance Criteria**:
  - [ ] `go test ./internal/observability/... -v -race` passes with new tests
  - [ ] `go test ./cmd/... -v -race -run TestRoot` passes
  - [ ] No output to stdout from progress helper when --output json

  **QA Scenarios**:

  ```
  Scenario: Observability initializes with no-op when env var absent
    Tool: Bash
    Preconditions: APPLICATIONINSIGHTS_CONNECTION_STRING not set
    Steps:
      1. Run: go test ./internal/observability/... -v -race -run TestNewTracer
      2. Assert: test passes, tracer is non-nil (no-op tracer)
    Expected Result: PASS, 0 failures
    Evidence: .sisyphus/evidence/task-1-otel-noop.txt

  Scenario: Progress helper suppresses in JSON mode
    Tool: Bash
    Preconditions: None
    Steps:
      1. Run: go test ./cmd/... -v -race -run TestProgressHelper
      2. Assert: when format=json, no spinner output written to stderr
    Expected Result: PASS, spinner output only in table mode
    Evidence: .sisyphus/evidence/task-1-progress-json.txt
  ```

  **Commit**: YES
  - Message: `feat(observability): wire OpenTelemetry tracer and meter into root command`
  - Files: `internal/observability/*`, `cmd/root.go`, `cmd/progress.go`, `cmd/progress_test.go`
  - Pre-commit: `go test ./internal/observability/... ./cmd/... -race`

- [ ] 2. Externalize NetworkPolicy YAML to Embedded File

  **What to do**:
  - Extract the `drasiNetworkPoliciesYAML` constant (cmd/provision.go lines 382-572, ~190 lines of YAML) into a new file `cmd/network_policies.yaml`
  - Use Go `embed` directive to embed the file: `//go:embed network_policies.yaml` with `var drasiNetworkPoliciesYAML string`
  - Remove the inline constant from provision.go
  - Update `applyDrasiNetworkPolicies` (line 352-368) to use the embedded variable (no logic change needed — variable name stays the same)
  - Add a test that verifies the embedded YAML is valid Kubernetes YAML (unmarshal into a list of unstructured objects)
  - Verify existing provision tests still pass

  **Must NOT do**:
  - Do NOT change the YAML content itself
  - Do NOT rename the variable — keep `drasiNetworkPoliciesYAML` so all references work

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 3, 4, 5)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `cmd/provision.go:370-572` — The drasiNetworkPoliciesYAML constant to extract
  - `cmd/provision.go:348-368` — applyDrasiNetworkPolicies function that consumes the YAML
  - `internal/scaffold/embed.go` — Example of Go embed usage in this codebase

  **Acceptance Criteria**:
  - [ ] `cmd/network_policies.yaml` exists with exact same YAML content
  - [ ] `cmd/provision.go` no longer contains inline YAML constant
  - [ ] `go test ./cmd/... -race -run TestProvision` passes
  - [ ] New test validates YAML is parseable as Kubernetes manifests

  **QA Scenarios**:

  ```
  Scenario: Embedded YAML is valid Kubernetes YAML
    Tool: Bash
    Preconditions: None
    Steps:
      1. Run: go test ./cmd/... -v -race -run TestNetworkPolicyYAMLValid
      2. Assert: YAML parses without error, contains 8 NetworkPolicy documents
    Expected Result: PASS
    Evidence: .sisyphus/evidence/task-2-yaml-valid.txt

  Scenario: Provision tests still pass after refactor
    Tool: Bash
    Preconditions: None
    Steps:
      1. Run: go test ./cmd/... -v -race -run TestProvision
      2. Assert: all existing provision tests pass unchanged
    Expected Result: PASS, 0 failures
    Evidence: .sisyphus/evidence/task-2-provision-tests.txt
  ```

  **Commit**: YES
  - Message: `refactor(provision): externalize NetworkPolicy YAML to embedded file`
  - Files: `cmd/provision.go`, `cmd/network_policies.yaml`, `cmd/provision_test.go`
  - Pre-commit: `go test ./cmd/... -race -run TestProvision`

- [ ] 3. Remove applyDefaultProviders No-op + Clean --follow Flag

  **What to do**:
  - In `cmd/provision.go`: remove the `applyDefaultProviders` function (lines 222-232) and its call site (lines 121-130). The function is a documented no-op per the inline comment.
  - In `cmd/logs.go`: remove the `--follow` flag declaration (line 142) and the no-op block (lines 99-101). Update the help text to remove references to follow.
  - Remove the `follow` variable from the logs command and from the payload map (line 120)
  - Update tests for both commands to reflect the removed code paths
  - Verify all existing tests still pass

  **Must NOT do**:
  - Do NOT change any other behavior in provision or logs commands
  - Do NOT remove the `follow` variable if it's used in tests — update tests instead

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 4, 5)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `cmd/provision.go:222-232` — applyDefaultProviders function (documented no-op)
  - `cmd/provision.go:119-130` — Call site for applyDefaultProviders
  - `cmd/logs.go:99-101` — follow no-op block
  - `cmd/logs.go:142` — follow flag declaration
  - `cmd/logs.go:120` — follow in payload map

  **Acceptance Criteria**:
  - [ ] `applyDefaultProviders` function no longer exists in provision.go
  - [ ] `--follow` flag no longer appears in `azd drasi logs --help`
  - [ ] `go test ./cmd/... -race` passes

  **QA Scenarios**:

  ```
  Scenario: Provision still works without applyDefaultProviders
    Tool: Bash
    Steps:
      1. Run: go test ./cmd/... -v -race -run TestProvision
      2. Assert: all provision tests pass
    Expected Result: PASS
    Evidence: .sisyphus/evidence/task-3-provision.txt

  Scenario: Logs help no longer shows --follow
    Tool: Bash
    Steps:
      1. Run: go test ./cmd/... -v -race -run TestLogs
      2. Grep logs_test.go or run help capture: verify no "follow" in flag list
    Expected Result: PASS, no --follow in help output
    Evidence: .sisyphus/evidence/task-3-logs-follow.txt
  ```

  **Commit**: YES
  - Message: `chore: remove applyDefaultProviders no-op and --follow compatibility flag`
  - Files: `cmd/provision.go`, `cmd/logs.go`, `cmd/provision_test.go`, `cmd/logs_test.go`
  - Pre-commit: `go test ./cmd/... -race`

- [ ] 4. Environment Overlay Example in Templates

  **What to do**:
  - In every `drasi/drasi.yaml` across `internal/scaffold/templates/*/drasi/drasi.yaml` (5 templates: blank, blank-terraform, cosmos-change-feed, event-hub-routing, query-subscription) AND the root `drasi/drasi.yaml`:
    - Replace the empty `environments: {}` block with a commented-out example showing how to define a `staging` overlay that overrides a parameter value
    - Example structure:
      ```yaml
      environments:
        # staging:
        #   components:
        #     - kind: source
        #       id: my-source
        #       properties:
        #         connectionString: "staging-connection-string"
      ```
  - Verify `internal/config/resolver.go` resolves overlays correctly by reviewing existing test coverage — add a test if none exists that exercises the overlay path with a real environment name
  - Update the root `drasi/drasi.yaml` comment to reference the doc path `docs/configuration-reference.md` for full schema

  **Must NOT do**:
  - Do NOT change the resolver logic — only add documentation/examples to YAML files
  - Do NOT add real secret values or connection strings in examples
  - Do NOT change the YAML schema — only add commented-out example within existing `environments` key

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Simple file edits across templates — no logic changes, just YAML comments
  - **Skills**: [`creating-azd-extensions`]
    - `creating-azd-extensions`: Understands azd extension template structure and conventions

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 3, 5)
  - **Blocks**: Task 14 (PostgreSQL template will include overlay example)
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `drasi/drasi.yaml` — Root manifest with `environments: {}`
  - `internal/scaffold/templates/blank/drasi/drasi.yaml` — Template manifest (all 5 are identical)
  - `internal/config/resolver.go` — Environment overlay resolution logic

  **API/Type References**:
  - `internal/config/resolver.go:ResolveConfig()` — Function that merges environment overlays

  **External References**:
  - `docs/configuration-reference.md` — Existing configuration reference doc

  **WHY Each Reference Matters**:
  - `drasi.yaml` files are the user's first contact with environment overlays — an empty `{}` gives no guidance
  - `resolver.go` confirms the exact YAML structure the overlay parser expects, ensuring the example is syntactically valid
  - The config reference doc is the canonical schema description — linking to it avoids duplicating docs in YAML comments

  **Acceptance Criteria**:
  - [ ] All 6 `drasi.yaml` files contain a commented-out environment overlay example
  - [ ] Example uses `staging` as the environment name
  - [ ] `go test ./internal/config/... -race` passes
  - [ ] `go test ./internal/scaffold/... -race` passes

  **QA Scenarios**:

  ```
  Scenario: Overlay example is syntactically valid YAML
    Tool: Bash
    Steps:
      1. Uncomment the overlay example in drasi/drasi.yaml
      2. Run: go run . validate --config drasi/drasi.yaml
      3. Assert: validation passes (exit code 0) or no YAML parse errors
    Expected Result: Valid YAML — no parse errors
    Failure Indicators: YAML parse error, indentation error, missing key
    Evidence: .sisyphus/evidence/task-4-yaml-valid.txt

  Scenario: Templates contain overlay example
    Tool: Bash
    Steps:
      1. For each template dir in internal/scaffold/templates/*/drasi/drasi.yaml:
         grep -l "staging" <file>
      2. Assert: all 5 templates match
    Expected Result: 5 files found with "staging" example
    Failure Indicators: Any template missing the example
    Evidence: .sisyphus/evidence/task-4-template-check.txt
  ```

  **Commit**: YES
  - Message: `docs(templates): add environment overlay example to drasi.yaml`
  - Files: `internal/scaffold/templates/*/drasi/drasi.yaml`, `drasi/drasi.yaml`
  - Pre-commit: `go test ./internal/scaffold/... -race`

- [ ] 5. Crash-Safe Deploy Lock with Timestamp

  **What to do**:
  - Create `internal/deployment/lock.go` with a `DeployLock` struct that wraps the state manager:
    - `Acquire(ctx)` — writes `DRASI_DEPLOY_IN_PROGRESS` = JSON `{"pid": <pid>, "started": <RFC3339>}` instead of bare `"true"`
    - `Release(ctx)` — writes `DRASI_DEPLOY_IN_PROGRESS` = `""`
    - `IsStale(ctx, maxAge time.Duration)` — reads the lock, parses timestamp, returns true if lock is older than `maxAge` or if the stored PID is no longer running (best-effort, skip PID check on cross-platform issues)
    - `ForceRelease(ctx)` — unconditionally clears the lock (for recovery)
  - Update `cmd/deploy.go`:
    - Replace the lock logic at lines 123, 144, 155 to use the new `DeployLock`
    - Before acquiring: check `IsStale(ctx, 30*time.Minute)` — if stale, log a warning and force-release
    - On deploy error (line 85-87): ensure `Release()` is called in a deferred cleanup (use `defer lock.Release(ctx)` after acquire)
  - The JSON lock value must be backwards-compatible: if the existing value is bare `"true"` (from an old deploy), `IsStale` should treat it as stale (no timestamp to parse)
  - Create `internal/deployment/lock_test.go`:
    - Test Acquire/Release cycle
    - Test stale lock detection (mock time or use a very short maxAge)
    - Test backwards-compatible parse of bare `"true"` value
    - Test concurrent acquire rejection (second acquire fails while first is held)
    - Test deferred release on error path

  **Must NOT do**:
  - Do NOT change the state key name — keep `DRASI_DEPLOY_IN_PROGRESS`
  - Do NOT use file-based locking — use the existing StateManager `ReadHash`/`WriteHash`
  - Do NOT add OS-specific PID checking that would break cross-platform (best-effort only)
  - Do NOT change hash algorithm or state key format (`DRASI_HASH_<KIND>_<ID>`)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Requires careful concurrency handling and backwards-compatible JSON parsing
  - **Skills**: [`creating-azd-extensions`]
    - `creating-azd-extensions`: Understands azd extension state management patterns

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 3, 4)
  - **Blocks**: Task 7 (rollback uses the deploy lock)
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `cmd/deploy.go:123` — Current lock acquire: `state.WriteHash(ctx, "DRASI_DEPLOY_IN_PROGRESS", "true")`
  - `cmd/deploy.go:144` — Lock release on success: `state.WriteHash(ctx, "DRASI_DEPLOY_IN_PROGRESS", "")`
  - `cmd/deploy.go:155` — Lock release on error path
  - `cmd/deploy.go:85-87` — Error handling in deploy loop (stops on first error, no cleanup)
  - `internal/deployment/state.go` — StateManager with `ReadHash(ctx, key)` and `WriteHash(ctx, key, value)` methods

  **API/Type References**:
  - `internal/deployment/state.go:StateManager` — Interface for reading/writing deploy state
  - `internal/deployment/state.go:ReadHash` — Returns `(string, error)`
  - `internal/deployment/state.go:WriteHash` — Returns `error`

  **External References**:
  - Go `time.RFC3339` format for timestamp serialization
  - Go `os.Getpid()` for PID capture

  **WHY Each Reference Matters**:
  - `deploy.go` lock lines show the exact state key and current bare-string protocol — new code must be drop-in compatible
  - `state.go` defines the storage interface — `DeployLock` must compose over it, not replace it
  - The error handling block at line 85-87 is where deferred release must be wired to prevent orphaned locks

  **Acceptance Criteria**:
  - [ ] `internal/deployment/lock.go` exists with `DeployLock` struct
  - [ ] `Acquire`, `Release`, `IsStale`, `ForceRelease` methods implemented
  - [ ] `cmd/deploy.go` uses `DeployLock` instead of bare string writes
  - [ ] Deferred `Release()` on error path prevents orphaned locks
  - [ ] Backwards-compatible: bare `"true"` value parsed as stale
  - [ ] `go test ./internal/deployment/... -race` passes with lock tests
  - [ ] `go test ./cmd/... -race -run TestDeploy` passes

  **QA Scenarios**:

  ```
  Scenario: Lock acquire and release cycle
    Tool: Bash
    Steps:
      1. Run: go test ./internal/deployment/... -v -race -run TestDeployLock
      2. Assert: Acquire succeeds, Release clears state, second Acquire succeeds after Release
    Expected Result: PASS — full lock lifecycle works
    Failure Indicators: Lock not cleared after Release, or second Acquire blocked after Release
    Evidence: .sisyphus/evidence/task-5-lock-cycle.txt

  Scenario: Stale lock detection and force-release
    Tool: Bash
    Steps:
      1. Run: go test ./internal/deployment/... -v -race -run TestStaleLock
      2. Assert: Lock with timestamp older than maxAge returns IsStale=true
      3. Assert: Bare "true" value returns IsStale=true (backwards compat)
    Expected Result: PASS — stale detection works for both new JSON and legacy bare values
    Failure Indicators: IsStale returns false for expired lock or for bare "true"
    Evidence: .sisyphus/evidence/task-5-stale-lock.txt

  Scenario: Deploy error releases lock
    Tool: Bash
    Steps:
      1. Run: go test ./cmd/... -v -race -run TestDeployRollback
      2. Assert: After simulated deploy error, DRASI_DEPLOY_IN_PROGRESS is cleared
    Expected Result: PASS — lock is not orphaned on error
    Failure Indicators: Lock value still set after deploy failure
    Evidence: .sisyphus/evidence/task-5-error-release.txt
  ```

  **Commit**: YES
  - Message: `feat(deploy): implement crash-safe deploy lock with timestamp`
  - Files: `internal/deployment/lock.go`, `internal/deployment/lock_test.go`, `cmd/deploy.go`
  - Pre-commit: `go test ./internal/deployment/... -race`

- [ ] 6. Diagnose Real Key Vault and Log Analytics Health Checks

  **What to do**:
  - In `cmd/diagnose.go`, replace the two stub checks (lines 164-176) that return `"skipped"` status:
    - **Key Vault check**: Use the existing `drasi` CLI wrapper pattern (same as AKS/Dapr checks at lines 60-162) to verify Key Vault connectivity. Approach: shell out to `az keyvault show --name <vault>` or use the existing Drasi client to attempt a secret list/describe. If the Key Vault name is available from environment state (`AZURE_KEYVAULT_NAME`), use it. If unavailable, return `"skipped"` with detail explaining the env var is not set.
    - **Log Analytics check**: Shell out to `az monitor log-analytics workspace show --resource-group <rg> --workspace-name <ws>` using `AZURE_LOG_ANALYTICS_WORKSPACE_NAME` and `AZURE_RESOURCE_GROUP` from env state. If env vars unavailable, return `"skipped"` with detail.
  - Follow the existing `diagnosticCheck` struct pattern at lines 14-19: `Name`, `Status` ("ok"/"failed"/"skipped"), `Detail`, `Remediation`
  - Follow the existing check execution pattern: run a command, parse output, set status based on exit code
  - Create/update `cmd/diagnose_test.go`:
    - Mock the CLI runner for each new check
    - Test happy path (returns "ok")
    - Test failure path (returns "failed" with remediation)
    - Test missing env var path (returns "skipped" with explanation)

  **Must NOT do**:
  - Do NOT add real Azure SDK API calls — use CLI wrappers only (consistent with existing checks)
  - Do NOT import new Azure SDK packages
  - Do NOT change the `diagnosticCheck` struct or existing check patterns
  - Do NOT remove or modify the existing 5 checks (AKS, drasi CLI, kubectl, Dapr, drasi API)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Requires careful error handling and consistent pattern following across multiple check types
  - **Skills**: [`creating-azd-extensions`]
    - `creating-azd-extensions`: Understands azd extension CLI wrapper patterns

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 7, 8, 9, 10)
  - **Blocks**: None
  - **Blocked By**: Task 1 (observability wiring for tracing diagnostic commands)

  **References**:

  **Pattern References**:
  - `cmd/diagnose.go:14-19` — `diagnosticCheck` struct definition
  - `cmd/diagnose.go:60-80` — AKS connectivity check pattern (run command, check exit code, set status)
  - `cmd/diagnose.go:82-100` — Drasi CLI check pattern
  - `cmd/diagnose.go:164-176` — Key Vault and Log Analytics stubs returning "skipped"

  **API/Type References**:
  - `cmd/diagnose.go:diagnosticCheck` — {Name string, Status string, Detail string, Remediation string}
  - Environment state keys: `AZURE_KEYVAULT_NAME`, `AZURE_LOG_ANALYTICS_WORKSPACE_NAME`, `AZURE_RESOURCE_GROUP`

  **External References**:
  - `az keyvault show` CLI reference: https://learn.microsoft.com/cli/azure/keyvault#az-keyvault-show
  - `az monitor log-analytics workspace show` CLI reference: https://learn.microsoft.com/cli/azure/monitor/log-analytics/workspace#az-monitor-log-analytics-workspace-show

  **WHY Each Reference Matters**:
  - The `diagnosticCheck` struct defines the contract — new checks must return the same shape
  - Existing checks (lines 60-100) show the exact pattern: run CLI, check error, populate struct — follow this exactly
  - The stubs at 164-176 are the replacement targets — preserve their position in the check sequence

  **Acceptance Criteria**:
  - [ ] Key Vault check attempts real connectivity when `AZURE_KEYVAULT_NAME` is set
  - [ ] Log Analytics check attempts real connectivity when env vars are set
  - [ ] Both checks gracefully return "skipped" when env vars are missing
  - [ ] Both checks return "failed" with remediation text when CLI command fails
  - [ ] `go test ./cmd/... -race -run TestDiagnose` passes with mock tests for all 3 paths
  - [ ] Existing 5 checks unmodified and still pass

  **QA Scenarios**:

  ```
  Scenario: Diagnose with env vars set (mocked)
    Tool: Bash
    Steps:
      1. Run: go test ./cmd/... -v -race -run TestDiagnoseKeyVaultOk
      2. Assert: Key Vault check returns status "ok"
      3. Run: go test ./cmd/... -v -race -run TestDiagnoseLogAnalyticsOk
      4. Assert: Log Analytics check returns status "ok"
    Expected Result: PASS — both checks report "ok" with mocked successful CLI output
    Failure Indicators: Status is "skipped" or "failed" when mock returns success
    Evidence: .sisyphus/evidence/task-6-diagnose-ok.txt

  Scenario: Diagnose with missing env vars
    Tool: Bash
    Steps:
      1. Run: go test ./cmd/... -v -race -run TestDiagnoseKeyVaultSkipped
      2. Assert: Key Vault check returns status "skipped" with detail mentioning AZURE_KEYVAULT_NAME
    Expected Result: PASS — graceful skip with helpful message
    Failure Indicators: Panic, error, or "failed" status when env var is simply missing
    Evidence: .sisyphus/evidence/task-6-diagnose-skip.txt

  Scenario: Diagnose with CLI failure
    Tool: Bash
    Steps:
      1. Run: go test ./cmd/... -v -race -run TestDiagnoseKeyVaultFailed
      2. Assert: Key Vault check returns status "failed" with remediation text
    Expected Result: PASS — clear failure with actionable remediation
    Failure Indicators: Empty remediation field, or status "ok" despite CLI error
    Evidence: .sisyphus/evidence/task-6-diagnose-fail.txt
  ```

  **Commit**: YES
  - Message: `feat(diagnose): implement real Key Vault and Log Analytics health checks`
  - Files: `cmd/diagnose.go`, `cmd/diagnose_test.go`
  - Pre-commit: `go test ./cmd/... -race -run TestDiagnose`

- [ ] 7. Deploy Rollback on Failure

  **What to do**:
  - In `internal/deployment/engine.go`, modify the `Deploy()` function (lines 43-97):
    - Track successfully-applied components in a `[]AppliedComponent` slice as the deploy loop progresses
    - When an error occurs at line 85-87 (currently just returns error), add rollback logic:
      1. Log a warning: "Deploy failed after applying N components. Rolling back..."
      2. Iterate `appliedComponents` in reverse order
      3. For each, call the equivalent delete/teardown operation (follow the `Teardown()` pattern at lines 99-121)
      4. If rollback itself fails, log the rollback error but still return the original deploy error
    - Add a `--no-rollback` flag to `cmd/deploy.go` that skips rollback (for debugging)
  - Define `AppliedComponent` struct: `{Kind string, ID string, Action string}` — captures what was applied
  - Update `internal/deployment/engine_test.go`:
    - Test: deploy 3 components, second fails → first is rolled back
    - Test: deploy with `--no-rollback` → no rollback on failure
    - Test: rollback itself fails → original error still returned, rollback error logged

  **Must NOT do**:
  - Do NOT change the deploy ordering (sources→queries→middlewares→reactions)
  - Do NOT add retry logic to rollback — single attempt, log and continue
  - Do NOT over-engineer — simple "delete what was applied in this run" is sufficient
  - Do NOT change the `Teardown()` function — use its pattern but don't modify it

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Requires careful error-path logic and understanding of component lifecycle
  - **Skills**: [`creating-azd-extensions`]
    - `creating-azd-extensions`: Understands azd extension deployment patterns

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 6, 8, 9, 10)
  - **Blocks**: None
  - **Blocked By**: Task 5 (crash-safe lock — rollback must release lock correctly)

  **References**:

  **Pattern References**:
  - `internal/deployment/engine.go:43-97` — `Deploy()` function with component iteration loop
  - `internal/deployment/engine.go:85-87` — Error return point (where rollback should trigger)
  - `internal/deployment/engine.go:99-121` — `Teardown()` function as pattern for delete operations
  - `internal/deployment/engine.go:50-60` — Component action iteration (apply/delete per component)

  **API/Type References**:
  - `internal/deployment/engine.go:DeployAction` — Struct with Kind, ID, Action fields
  - `internal/drasi/client.go:ApplyComponent()` — Apply method called during deploy
  - `internal/drasi/client.go:DeleteComponent()` — Delete method for rollback

  **Test References**:
  - `internal/deployment/engine_test.go` — Existing deploy tests with mock runner

  **WHY Each Reference Matters**:
  - `Deploy()` at lines 43-97 is the exact function to modify — the loop structure determines where to insert tracking and rollback
  - `Teardown()` at lines 99-121 shows the canonical delete pattern — rollback should mirror this exactly
  - The error at line 85-87 is the trigger point — rollback runs between catching the error and returning it

  **Acceptance Criteria**:
  - [ ] `Deploy()` tracks applied components as it progresses
  - [ ] On error, previously-applied components are deleted in reverse order
  - [ ] Rollback failure is logged but does not mask original error
  - [ ] `--no-rollback` flag disables rollback behavior
  - [ ] `go test ./internal/deployment/... -race` passes
  - [ ] `go test ./cmd/... -race -run TestDeploy` passes

  **QA Scenarios**:

  ```
  Scenario: Rollback on partial deploy failure
    Tool: Bash
    Steps:
      1. Run: go test ./internal/deployment/... -v -race -run TestDeployRollbackOnFailure
      2. Assert: Mock shows 2 components applied, then 2 delete calls in reverse order
      3. Assert: Original deploy error is returned
    Expected Result: PASS — rollback deletes applied components in reverse order
    Failure Indicators: No delete calls after failure, or delete calls in wrong order
    Evidence: .sisyphus/evidence/task-7-rollback.txt

  Scenario: No rollback with --no-rollback flag
    Tool: Bash
    Steps:
      1. Run: go test ./cmd/... -v -race -run TestDeployNoRollback
      2. Assert: Deploy fails but no delete calls are made
    Expected Result: PASS — error returned, no rollback attempted
    Failure Indicators: Delete calls present despite --no-rollback
    Evidence: .sisyphus/evidence/task-7-no-rollback.txt

  Scenario: Rollback failure doesn't mask original error
    Tool: Bash
    Steps:
      1. Run: go test ./internal/deployment/... -v -race -run TestDeployRollbackFailure
      2. Assert: Both deploy error and rollback error are logged
      3. Assert: Function returns the original deploy error (not the rollback error)
    Expected Result: PASS — original error preserved, rollback error logged
    Failure Indicators: Returned error is the rollback error, or rollback error is silently swallowed
    Evidence: .sisyphus/evidence/task-7-rollback-fail.txt
  ```

  **Commit**: YES
  - Message: `feat(deploy): add rollback on failure for partially-applied components`
  - Files: `internal/deployment/engine.go`, `internal/deployment/engine_test.go`, `cmd/deploy.go`
  - Pre-commit: `go test ./internal/deployment/... -race`

- [ ] 8. Add Describe Command for Single-Component Detail View

  **What to do**:
  - Create `cmd/describe.go` with a new `drasi describe` command:
    - Required flags: `--kind` (source/continuousquery/middleware/reaction), `--component` (component ID)
    - Calls `drasi.DescribeComponent(ctx, kind, id)` which already exists in `internal/drasi/client.go`
    - Output: Render the component detail in table format (default) or JSON (`--output json`)
    - Handle "not found" gracefully — return a clear error message with the component kind and ID
    - Map `continuousquery` → `query` for display (same as logs command)
  - Register the command in `cmd/root.go` alongside existing commands
  - Create `cmd/describe_test.go`:
    - Mock the Drasi client interface (follow `cmd/logs.go:13-18` pattern for interface definition)
    - Test: describe existing component → renders detail
    - Test: describe non-existent component → error message
    - Test: JSON output format → valid JSON
    - Test: missing --kind or --component → error message

  **Must NOT do**:
  - Do NOT modify the Drasi client — `DescribeComponent` already exists
  - Do NOT add new API calls — only use existing client methods
  - Do NOT create a subcommand tree — `describe` is a flat top-level command like `status` and `logs`

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: New command file, needs proper Cobra setup and test coverage
  - **Skills**: [`creating-azd-extensions`]
    - `creating-azd-extensions`: Understands azd extension command patterns

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 6, 7, 9, 10)
  - **Blocks**: None
  - **Blocked By**: Task 1 (observability wiring in root command)

  **References**:

  **Pattern References**:
  - `cmd/status.go` — Best pattern to follow for new command structure (Cobra setup, flags, output formatting)
  - `cmd/logs.go:13-18` — `logsDrasiClient` interface pattern showing how to define a mockable client interface
  - `cmd/root.go` — Command registration point (AddCommand calls)

  **API/Type References**:
  - `internal/drasi/client.go:DescribeComponent(ctx, kind, id)` — Returns component detail (already implemented)
  - `internal/drasi/client.go:DescribeComponentInContext(ctx, kind, id, kubeContext)` — Context-aware variant
  - `internal/output/formatter.go` — Output formatting utilities (table/JSON)

  **Test References**:
  - `cmd/deploy_test.go` — Test pattern with mock client and Cobra command execution

  **WHY Each Reference Matters**:
  - `status.go` is the closest existing command to `describe` — same flags, same output flow, same error handling
  - The client interface at `logs.go:13-18` shows the exact pattern for defining a testable client boundary
  - `DescribeComponent` already exists — this task only creates the CLI surface to expose it

  **Acceptance Criteria**:
  - [ ] `cmd/describe.go` exists with `--kind` and `--component` required flags
  - [ ] `azd drasi describe --help` shows usage with kind and component flags
  - [ ] Table and JSON output formats work
  - [ ] Not-found returns clear error with kind and ID
  - [ ] Command registered in root.go
  - [ ] `go test ./cmd/... -race -run TestDescribe` passes

  **QA Scenarios**:

  ```
  Scenario: Describe existing component (mocked)
    Tool: Bash
    Steps:
      1. Run: go test ./cmd/... -v -race -run TestDescribeSuccess
      2. Assert: Command output contains component detail fields (kind, ID, status)
    Expected Result: PASS — component detail rendered in table format
    Failure Indicators: Empty output, missing fields, or error
    Evidence: .sisyphus/evidence/task-8-describe-ok.txt

  Scenario: Describe non-existent component
    Tool: Bash
    Steps:
      1. Run: go test ./cmd/... -v -race -run TestDescribeNotFound
      2. Assert: Error message contains the kind and component ID
    Expected Result: PASS — clear "not found" error with specific kind and ID
    Failure Indicators: Generic error message, panic, or exit without error
    Evidence: .sisyphus/evidence/task-8-describe-notfound.txt

  Scenario: Describe with JSON output
    Tool: Bash
    Steps:
      1. Run: go test ./cmd/... -v -race -run TestDescribeJSON
      2. Assert: Output is valid JSON parseable by json.Unmarshal
    Expected Result: PASS — valid JSON output
    Failure Indicators: JSON parse error, or table format in JSON mode
    Evidence: .sisyphus/evidence/task-8-describe-json.txt
  ```

  **Commit**: YES
  - Message: `feat(cmd): add describe command for single-component detail view`
  - Files: `cmd/describe.go`, `cmd/describe_test.go`, `cmd/root.go`
  - Pre-commit: `go test ./cmd/... -race -run TestDescribe`

- [ ] 9. Status Show All Component Kinds When No --kind Flag

  **What to do**:
  - In `cmd/status.go`, modify the behavior when `--kind` flag is empty (line 37-38):
    - Currently: `if selectedKind == "" { selectedKind = "source" }` — defaults to showing only sources
    - New behavior: if no `--kind` is specified, iterate all 4 kinds (`source`, `continuousquery`, `middleware`, `reaction`) and aggregate results
    - Display with section headers per kind (e.g., "Sources:", "Queries:", "Middleware:", "Reactions:")
    - For JSON output: wrap in an object with kind keys `{"sources": [...], "queries": [...], "middleware": [...], "reactions": [...]}`
    - Keep the `--kind` flag working as before for filtering to a single kind
  - Update `cmd/status_test.go`:
    - Test: no --kind flag → all 4 kinds displayed
    - Test: --kind source → only sources (existing behavior preserved)
    - Test: JSON output with all kinds → valid JSON with kind keys
    - Test: empty cluster (no components) → clean empty output per kind

  **Must NOT do**:
  - Do NOT change the output format for single-kind queries (backwards compatible)
  - Do NOT add new Drasi client methods — use the existing status method in a loop
  - Do NOT change the component kind strings

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Small change to default behavior in one file — remove default, add loop
  - **Skills**: [`creating-azd-extensions`]
    - `creating-azd-extensions`: Understands azd extension command patterns

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 6, 7, 8, 10)
  - **Blocks**: None
  - **Blocked By**: Task 1 (observability wiring)

  **References**:

  **Pattern References**:
  - `cmd/status.go:37-38` — Current default: `if selectedKind == "" { selectedKind = "source" }`
  - `cmd/status.go` — Full command structure showing how kind is used in the status query
  - `internal/output/formatter.go` — Table and JSON output formatting

  **API/Type References**:
  - Component kinds: `"source"`, `"continuousquery"`, `"middleware"`, `"reaction"`

  **WHY Each Reference Matters**:
  - Line 37-38 is the exact code to change — removing the default and adding a loop over all kinds
  - The status command structure shows how kind flows into the Drasi client call — the loop must call the same method per kind

  **Acceptance Criteria**:
  - [ ] `azd drasi status` (no --kind) shows all 4 component kinds
  - [ ] `azd drasi status --kind source` still shows only sources
  - [ ] JSON output wraps kinds in an object with kind keys
  - [ ] Empty cluster shows clean output with section headers but no rows
  - [ ] `go test ./cmd/... -race -run TestStatus` passes

  **QA Scenarios**:

  ```
  Scenario: Status shows all kinds by default
    Tool: Bash
    Steps:
      1. Run: go test ./cmd/... -v -race -run TestStatusAllKinds
      2. Assert: Output contains section headers for all 4 kinds
    Expected Result: PASS — all 4 kind sections present in output
    Failure Indicators: Only "source" kind shown, or missing kind sections
    Evidence: .sisyphus/evidence/task-9-status-all.txt

  Scenario: Status with --kind flag filters correctly
    Tool: Bash
    Steps:
      1. Run: go test ./cmd/... -v -race -run TestStatusSingleKind
      2. Assert: Only the specified kind is shown (no other kind sections)
    Expected Result: PASS — filtered to single kind
    Failure Indicators: Multiple kinds shown despite --kind flag
    Evidence: .sisyphus/evidence/task-9-status-filter.txt

  Scenario: Status JSON output structure
    Tool: Bash
    Steps:
      1. Run: go test ./cmd/... -v -race -run TestStatusAllKindsJSON
      2. Assert: JSON output has keys "sources", "queries", "middleware", "reactions"
      3. Assert: Each key value is an array
    Expected Result: PASS — valid JSON with kind-keyed structure
    Failure Indicators: Flat array, missing keys, or invalid JSON
    Evidence: .sisyphus/evidence/task-9-status-json.txt
  ```

  **Commit**: YES
  - Message: `feat(status): show all component kinds when no --kind flag provided`
  - Files: `cmd/status.go`, `cmd/status_test.go`
  - Pre-commit: `go test ./cmd/... -race -run TestStatus`

- [ ] 10. Extend Logs Command to Support All Component Kinds

  **What to do**:
  - In `cmd/logs.go`, remove the restriction that rejects non-query kinds (lines 53-61):
    - Currently: only `continuousquery`/`query` kind is accepted; others return an error
    - New behavior: accept all 4 kinds (`source`, `continuousquery`, `middleware`, `reaction`)
  - For `continuousquery` kind: keep existing `drasi watch` behavior (line 95)
  - For `source`, `middleware`, `reaction` kinds: use `kubectl logs` to stream pod logs:
    - Construct pod label selector from kind and component ID (e.g., `drasi.io/kind=source,drasi.io/component=<id>`)
    - Shell out to `kubectl logs -l <selector> -n drasi --follow=<follow-flag> --tail=100`
    - Use the existing `RunCommand` pattern from the Drasi client for exec
  - Keep the `continuousquery` → `query` mapping at line 39-41
  - Update `cmd/logs_test.go`:
    - Test: logs for source kind → kubectl logs called with correct selector
    - Test: logs for reaction kind → kubectl logs called with correct selector
    - Test: logs for continuousquery → drasi watch called (existing behavior)
    - Test: missing --kind or --component → error

  **Must NOT do**:
  - Do NOT break the `drasi watch` wrapper — only ADD `kubectl logs` as alternative for non-query kinds
  - Do NOT change the flag interface (keep `--kind`, `--component`)
  - Do NOT add new dependencies — `kubectl` is already a prerequisite

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Requires branching logic and correct label selector construction
  - **Skills**: [`creating-azd-extensions`]
    - `creating-azd-extensions`: Understands azd extension CLI patterns

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 6, 7, 8, 9)
  - **Blocks**: None
  - **Blocked By**: Task 3 (removes --follow flag first — Task 10 must not re-add it)

  **References**:

  **Pattern References**:
  - `cmd/logs.go:39-41` — `continuousquery` → `query` kind mapping
  - `cmd/logs.go:53-61` — Current rejection of non-query kinds (remove this gate)
  - `cmd/logs.go:95` — `drasi watch` invocation for query kind
  - `cmd/logs.go:13-18` — `logsDrasiClient` interface

  **API/Type References**:
  - Pod label selectors: `drasi.io/kind=<kind>`, `drasi.io/component=<id>` (verify in cluster or Drasi docs)
  - `kubectl logs -l <selector> -n drasi --tail=100`

  **External References**:
  - kubectl logs reference: https://kubernetes.io/docs/reference/kubectl/generated/kubectl_logs/
  - Drasi component labeling: https://drasi.io/drasi-server/reference/cli/

  **WHY Each Reference Matters**:
  - Lines 53-61 are the gate to remove — this is what currently prevents non-query logs
  - Line 95 shows the `drasi watch` path that must be preserved for queries
  - The label selector pattern must match how Drasi labels pods — incorrect selectors mean no logs

  **Acceptance Criteria**:
  - [ ] `logs --kind source --component X` works via kubectl logs
  - [ ] `logs --kind reaction --component X` works via kubectl logs
  - [ ] `logs --kind middleware --component X` works via kubectl logs
  - [ ] `logs --kind continuousquery --component X` still uses `drasi watch`
  - [ ] Label selector uses correct Drasi pod labels
  - [ ] `go test ./cmd/... -race -run TestLogs` passes

  **QA Scenarios**:

  ```
  Scenario: Logs for source kind uses kubectl
    Tool: Bash
    Steps:
      1. Run: go test ./cmd/... -v -race -run TestLogsSourceKind
      2. Assert: Mock captures kubectl logs call with label selector containing "source"
    Expected Result: PASS — kubectl logs invoked with correct selector
    Failure Indicators: Error "unsupported kind", or drasi watch called instead of kubectl
    Evidence: .sisyphus/evidence/task-10-logs-source.txt

  Scenario: Logs for continuousquery still uses drasi watch
    Tool: Bash
    Steps:
      1. Run: go test ./cmd/... -v -race -run TestLogsContinuousQuery
      2. Assert: Mock captures drasi watch call (not kubectl logs)
    Expected Result: PASS — drasi watch used for query kind
    Failure Indicators: kubectl logs called for continuousquery kind
    Evidence: .sisyphus/evidence/task-10-logs-query.txt

  Scenario: Logs for reaction kind
    Tool: Bash
    Steps:
      1. Run: go test ./cmd/... -v -race -run TestLogsReactionKind
      2. Assert: Mock captures kubectl logs call with label selector containing "reaction"
    Expected Result: PASS — kubectl logs invoked for reaction
    Failure Indicators: Error or incorrect selector
    Evidence: .sisyphus/evidence/task-10-logs-reaction.txt
  ```

  **Commit**: YES
  - Message: `feat(logs): extend logs command to support source, reaction, and middleware kinds`
  - Files: `cmd/logs.go`, `cmd/logs_test.go`
  - Pre-commit: `go test ./cmd/... -race -run TestLogs`

- [ ] 11. Interactive Confirmation Prompts for Destructive Operations

  **What to do**:
  - Create `cmd/confirm.go` with a reusable confirmation helper:
    - `ConfirmDestructive(prompt string, force bool) (bool, error)` — if `force` is true, skip prompt and return true; otherwise prompt user via `survey/v2` `Confirm` question
    - Handle non-interactive terminals: if stdin is not a TTY, return error "use --force for non-interactive execution"
  - Update `cmd/teardown.go`:
    - Currently: `--force` is required and the command fails without it
    - New behavior: if `--force` is not set, prompt with `ConfirmDestructive("This will remove all deployed Drasi components. Continue?", force)`
    - If `--infrastructure` is also set, use a stronger prompt: "This will delete all Drasi components AND the Azure resource group. This is irreversible. Continue?"
    - Keep `--force` as a bypass for CI/scripts
  - Update `cmd/upgrade.go`:
    - Same pattern: if `--force` not set, prompt: "This will upgrade the Drasi runtime on the active cluster. Continue?"
  - Update tests for `cmd/teardown_test.go` and `cmd/upgrade_test.go`:
    - Test: with --force → no prompt, proceeds
    - Test: mock survey confirm → proceeds on yes
    - Test: mock survey deny → aborts with "aborted by user" message
    - Test: non-interactive terminal → error message mentioning --force

  **Must NOT do**:
  - Do NOT add new dependencies — `survey/v2` is already in go.mod as transitive dependency
  - Do NOT change the --force flag semantics — it still works as override
  - Do NOT add confirmation to non-destructive commands (deploy, status, etc.)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: TTY detection and survey integration require careful testing
  - **Skills**: [`creating-azd-extensions`]
    - `creating-azd-extensions`: Understands azd extension UX patterns

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 12, 13)
  - **Blocks**: None
  - **Blocked By**: Task 1 (observability), Task 6 (diagnose — avoids merge conflicts in cmd/)

  **References**:

  **Pattern References**:
  - `cmd/teardown.go` — Current `--force` flag usage pattern
  - `cmd/upgrade.go` — Current `--force` flag usage pattern

  **API/Type References**:
  - `go.mod:22` — `github.com/AlecAivazis/survey/v2 v2.3.7` (existing transitive dependency)
  - `survey.AskOne(&survey.Confirm{Message: prompt}, &result)` — survey v2 API

  **External References**:
  - survey/v2 API: https://pkg.go.dev/github.com/AlecAivazis/survey/v2
  - TTY detection in Go: `os.Stdin.Stat()` or `golang.org/x/term` `IsTerminal`

  **WHY Each Reference Matters**:
  - `teardown.go` and `upgrade.go` show the current --force flow — the confirm helper integrates at the same point
  - survey/v2 is the prompting library already available — no new dep needed
  - TTY detection is critical for CI pipelines where stdin is not a terminal

  **Acceptance Criteria**:
  - [ ] `cmd/confirm.go` exists with `ConfirmDestructive` function
  - [ ] `teardown` without --force prompts for confirmation
  - [ ] `teardown --infrastructure` without --force prompts with stronger warning
  - [ ] `upgrade` without --force prompts for confirmation
  - [ ] `--force` bypasses all prompts (backwards compatible)
  - [ ] Non-interactive terminal returns error mentioning --force
  - [ ] `go test ./cmd/... -race -run TestTeardown|TestUpgrade` passes

  **QA Scenarios**:

  ```
  Scenario: Teardown prompts without --force
    Tool: Bash
    Steps:
      1. Run: go test ./cmd/... -v -race -run TestTeardownConfirmYes
      2. Assert: Mock survey returns true → teardown proceeds
    Expected Result: PASS — teardown executes after confirmation
    Failure Indicators: Teardown skips without prompting, or fails even with confirm
    Evidence: .sisyphus/evidence/task-11-teardown-confirm.txt

  Scenario: Teardown aborts on deny
    Tool: Bash
    Steps:
      1. Run: go test ./cmd/... -v -race -run TestTeardownConfirmNo
      2. Assert: Mock survey returns false → "aborted by user" returned
    Expected Result: PASS — clean abort, no teardown operations
    Failure Indicators: Teardown proceeds despite denial
    Evidence: .sisyphus/evidence/task-11-teardown-deny.txt

  Scenario: Force flag bypasses prompt
    Tool: Bash
    Steps:
      1. Run: go test ./cmd/... -v -race -run TestTeardownForce
      2. Assert: No survey call made, teardown proceeds directly
    Expected Result: PASS — --force skips confirmation
    Failure Indicators: Survey called despite --force flag
    Evidence: .sisyphus/evidence/task-11-force-bypass.txt

  Scenario: Non-interactive terminal fails gracefully
    Tool: Bash
    Steps:
      1. Run: go test ./cmd/... -v -race -run TestTeardownNonInteractive
      2. Assert: Error message contains "--force" hint
    Expected Result: PASS — clear error directing user to use --force
    Failure Indicators: Panic, hang, or generic error without --force hint
    Evidence: .sisyphus/evidence/task-11-noninteractive.txt
  ```

  **Commit**: YES
  - Message: `feat(ux): add interactive confirmation prompts for destructive operations`
  - Files: `cmd/confirm.go`, `cmd/confirm_test.go`, `cmd/teardown.go`, `cmd/upgrade.go`, `cmd/teardown_test.go`, `cmd/upgrade_test.go`
  - Pre-commit: `go test ./cmd/... -race -run TestTeardown|TestUpgrade`

---

## Final Verification Wave

> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.

- [ ] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, run command). For each "Must NOT Have": search codebase for forbidden patterns — reject with file:line if found. Check evidence files exist in .sisyphus/evidence/. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [ ] F2. **Code Quality Review** — `unspecified-high`
  Run `go vet ./...` + `golangci-lint run` + `go test ./... -race -count=1`. Review all changed files for: type assertion without ok check, empty error handlers, fmt.Println in production code, commented-out code, unused imports. Check for AI slop: excessive comments, over-abstraction, generic variable names.
  Output: `Build [PASS/FAIL] | Lint [PASS/FAIL] | Tests [N pass/N fail] | Files [N clean/N issues] | VERDICT`

- [ ] F3. **Real Manual QA** — `unspecified-high`
  Execute EVERY QA scenario from EVERY task. Follow exact steps, capture evidence. Test cross-task integration. Save to `.sisyphus/evidence/final-qa/`.
  Output: `Scenarios [N/N pass] | Integration [N/N] | Edge Cases [N tested] | VERDICT`

- [ ] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff. Verify 1:1 — everything in spec was built, nothing beyond spec was built. Check "Must NOT do" compliance. Detect cross-task contamination. Flag unaccounted changes.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

| Task | Commit Message | Files | Pre-commit |
|------|---------------|-------|------------|
| 1 | `feat(observability): wire OpenTelemetry tracer and meter into root command` | internal/observability/*, cmd/root.go | `go test ./internal/observability/... -race` |
| 2 | `refactor(provision): externalize NetworkPolicy YAML to embedded file` | cmd/provision.go, cmd/network_policies.yaml | `go test ./cmd/... -race -run TestProvision` |
| 3 | `chore: remove applyDefaultProviders no-op and --follow compatibility flag` | cmd/provision.go, cmd/logs.go | `go test ./cmd/... -race` |
| 4 | `docs(templates): add environment overlay example to drasi.yaml` | internal/scaffold/templates/*/drasi/drasi.yaml, drasi/drasi.yaml | `go test ./internal/scaffold/... -race` |
| 5 | `feat(deploy): implement crash-safe deploy lock with timestamp` | internal/deployment/lock.go, cmd/deploy.go | `go test ./internal/deployment/... -race` |
| 6 | `feat(diagnose): implement real Key Vault and Log Analytics health checks` | cmd/diagnose.go, cmd/diagnose_test.go | `go test ./cmd/... -race -run TestDiagnose` |
| 7 | `feat(deploy): add rollback on failure for partially-applied components` | internal/deployment/engine.go, internal/deployment/engine_test.go | `go test ./internal/deployment/... -race` |
| 8 | `feat(cmd): add describe command for single-component detail view` | cmd/describe.go, cmd/describe_test.go, cmd/root.go | `go test ./cmd/... -race -run TestDescribe` |
| 9 | `feat(status): show all component kinds when no --kind flag provided` | cmd/status.go, cmd/status_test.go | `go test ./cmd/... -race -run TestStatus` |
| 10 | `feat(logs): extend logs command to support source, reaction, and middleware kinds` | cmd/logs.go, cmd/logs_test.go | `go test ./cmd/... -race -run TestLogs` |
| 11 | `feat(ux): add interactive confirmation prompts for destructive operations` | cmd/teardown.go, cmd/upgrade.go, cmd/confirm.go | `go test ./cmd/... -race -run TestTeardown\|TestUpgrade` |
| 12 | `feat(ux): add progress spinners to long-running commands` | cmd/progress.go, cmd/provision.go, cmd/deploy.go, cmd/teardown.go | `go test ./cmd/... -race` |
| 13 | `feat(hooks): implement postprovision health check and predeploy validation` | cmd/listen.go, cmd/listen_test.go | `go test ./cmd/... -race -run TestListen` |
| 14 | `feat(scaffold): add PostgreSQL source template` | internal/scaffold/templates/postgresql-source/* | `go test ./internal/scaffold/... -race` |
| 15 | `feat(telemetry): track command usage and error rates via OpenTelemetry` | cmd/root.go, internal/observability/telemetry.go | `go test ./internal/observability/... ./cmd/... -race` |
| 16 | `test(keyvault): add integration tests for Key Vault translator in deploy pipeline` | internal/keyvault/translator_test.go | `go test ./internal/keyvault/... -race` |

---

## Success Criteria

### Verification Commands
```bash
go test ./... -race -count=1 -coverprofile=cover.out  # Expected: PASS, coverage >= 80%
go tool cover -func=cover.out | grep total             # Expected: total >= 80.0%
golangci-lint run                                       # Expected: 0 issues
go vet ./...                                            # Expected: 0 errors
go build ./...                                          # Expected: 0 errors
```

### Final Checklist
- [ ] All "Must Have" requirements present
- [ ] All "Must NOT Have" guardrails respected
- [ ] All 16 tasks pass their QA scenarios
- [ ] All tests pass with race detector
- [ ] Coverage >= 80%
- [ ] No lint issues
- [ ] README updated with new commands (describe) and changed behavior (status, logs)
