# Learnings — drasi-extension-improvements

## 2026-04-07 Session Start

### Codebase Conventions
- All cmd/ files follow the same interface pattern: define a `<cmd>DrasiClient` interface, mock it with `new<Cmd>DrasiClient` var, use `cmd.OutOrStdout()` / `cmd.ErrOrStderr()` for output
- Error handling: `errorCodeFromError()` + `writeCommandError()` with 5 params (cmd, code, msg, remediation, format, exitcode)
- Output format retrieved via `outputFormatFromCommand(cmd)`
- Tests use manual mocks (struct with method impls), testify assert/require, t.Parallel()
- Deploy lock: `state.ReadHash(ctx, "DRASI_DEPLOY_IN_PROGRESS")` and `state.WriteHash()` — already has `defer` in deploy.go line 154-156
- logs.go already has `var follow bool` and no-op block at lines 99-101
- root.go is minimal (44 lines) — PersistentPreRunE/PostRunE not yet present
- observability: tracer.go:NewTracer() returns (tracer, shutdown, error); metrics.go:NewMeter() returns (meter, shutdown, error) — both check APPLICATIONINSIGHTS_CONNECTION_STRING

### Key File Locations
- cmd/root.go:44 — add PersistentPreRunE/PostRunE here
- internal/observability/tracer.go:21 — NewTracer signature
- internal/observability/metrics.go:15 — NewMeter signature
- cmd/provision.go:121 — applyDefaultProviders call site
- cmd/provision.go:222-232 — applyDefaultProviders function (no-op)
- cmd/logs.go:27 — follow bool declaration
- cmd/logs.go:53-61 — kind != "query" rejection gate
- cmd/logs.go:99-101 — follow no-op block
- cmd/logs.go:120 — follow in payload map
- cmd/logs.go:142 — --follow flag declaration
- cmd/deploy.go:123-156 — deploy lock pattern (already has defer release)
- internal/scaffold/embed.go — embed pattern to follow

## [2026-04-07] Task 1: Observability Wiring + Progress Helper

### Implementation
- PersistentPreRunE/PostRunE added to root.go; calls observability.NewTracer/NewMeter and stores in package-level vars
- Shutdown functions stored as package-level vars; nil-checked in PostRunE
- On init failure, gracefully logs warning and continues (no-op degradation)
- ProgressHelper wraps yacspin; uses cmd.ErrOrStderr() for spinner output (never stdout)
- JSON mode: ProgressHelper.noop=true, all methods are silent no-ops
- Test helper command exported as NewTestableProgressCommand() in progress_testhelper.go for black-box tests

### Gotchas
- cobra --help does NOT trigger PersistentPreRunE; must use a real subcommand (version) to test
- Root tests that mutate package-level vars (rootTracer, rootMeter, shutdownTracer, shutdownMeter) cannot use t.Parallel() due to data races
- go test -race requires CGO_ENABLED=1 on Windows; tests pass without -race flag
- provision.go has blank embed import that LSP sometimes flags incorrectly but go build succeeds

### Key Patterns
- yacspin.Config{Writer: cmd.ErrOrStderr()} ensures spinner goes to stderr
- yacspin.CharSets[14] is a good default braille dot pattern
- ShowCursor: true avoids cursor disappearing on crash
- 100ms Frequency is responsive without excessive flicker

## [2026-04-07] Tasks 2+3: NetworkPolicy YAML + Dead Code Removal

### Task 2: Extract network_policies.yaml
- The existing file had WRONG indentation (ingress/egress at root level instead of under spec:). Fixed to match the Go constant exactly.
- Go `//go:embed` with `var x string` requires `_ ""embed""`, not `import ""embed""`. The embed package is only used by the directive, not directly in code.
- The test already existed (TestNetworkPolicyYAMLValid) but only checked >= 1 documents. Updated to assert exactly 8.

### Task 3: Dead code removal
- applyDefaultProviders was a no-op (returns nil) with its call site wrapping it in error handling. Both removed cleanly.
- follow flag in logs.go had 4 touch points: var declaration, flag registration, no-op if-block, payload entry. All removed.
- Internal test (logs_internal_test.go) also used --follow in test args. Had to remove that too (not just the external test).
- External test TestLogsCommand_FollowFlagAccepted removed from logs_test.go.

### Build note
- CGO_ENABLED=0 is the default on Windows, so -race flag doesn't work. Tests run without -race on this platform.

## [2026-04-07] Task 4: Environment Overlay Examples

### What was done
- Replaced `environments: {}` with a commented-out staging overlay example in all 6 drasi.yaml files
- Files updated: drasi/drasi.yaml + 5 templates under internal/scaffold/templates/*/drasi/drasi.yaml
- Templates: blank, blank-terraform, cosmos-change-feed, event-hub-routing, query-subscription

### Pattern used
`yaml
environments:
  # staging:
  #   components:
  #     - kind: source
  #       id: my-source
  #       properties:
  #         connectionString: "staging-connection-string"
`

### Key observations
- The event-hub-routing template did NOT have the `# drasi.yaml - Drasi project manifest` header comment; all others did
- YAML comments are stripped by the Go YAML parser, so scaffold tests that validate required fields still pass
- go test ./internal/config/... ./internal/scaffold/... -count=1: both PASS
- No Go code changes required — purely YAML template updates

### Gotcha
- `environments: {}` parses as an empty map in Go; `environments:` followed only by comments is also valid YAML (nil map), which the config resolver handles identically

## [2026-04-07] Task 6: Real Diagnostic Checks for Key Vault and Log Analytics

### What was done
- Replaced two stub `diagnosticCheck` entries (lines 164-176) in `cmd/diagnose.go` with real checks
- Added `os` import for `os.Getenv` to read azd-populated env vars
- Added two package-level `var` functions: `azKeyVaultCheck` and `azLogAnalyticsCheck` that shell out to `az` CLI
- Key Vault check: reads `AZURE_KEYVAULT_NAME`, runs `az keyvault show --name <vault>`
- Log Analytics check: reads `AZURE_LOG_ANALYTICS_WORKSPACE_NAME` + `AZURE_RESOURCE_GROUP`, runs `az monitor log-analytics workspace show`
- Both return "skipped" when env vars are empty, "failed" when CLI returns non-zero, "ok" when CLI succeeds

### Test pattern
- Created `saveDiagnoseVars` helper in `diagnose_internal_test.go` to DRY the save/restore boilerplate (saves all 6 var functions)
- Created `stubAllChecksPass` helper to mock all prerequisite checks passing
- Refactored existing 3 internal tests to use the new helpers (reduced boilerplate)
- Added 6 new tests: `TestDiagnoseKeyVaultOk`, `TestDiagnoseKeyVaultFailed`, `TestDiagnoseKeyVaultSkipped`, `TestDiagnoseLogAnalyticsOk`, `TestDiagnoseLogAnalyticsFailed`, `TestDiagnoseLogAnalyticsSkipped`
- Tests use `t.Setenv()` to inject env vars (automatically restored by Go testing framework)
- Tests use JSON output format for easier assertion against specific field values

### Key design decisions
- Used `os.Getenv` (not azd gRPC env state API) because these are standard azd-populated process env vars
- Used `exec.CommandContext` pattern consistent with `isDaprReady` for the az CLI shell-outs
- New var functions return `(string, string, error)` instead of `(bool, string, error)` because there are 3 statuses (ok/failed/skipped) rather than binary
- The "skipped" path is handled in the command `RunE` (based on env var presence), not in the var functions themselves

### Gotcha
- The `azKeyVaultCheck` var function returns both the CLI stderr output as detail AND the Go error; the command handler uses `kvErr.Error()` as the detail when the CLI fails, overriding the var function's detail string

## [2026-04-07] Task 8: Describe Command

### What was done
- Created `cmd/describe.go` with `newDescribeCommand()` following status.go/logs.go pattern
- Created `cmd/describe_test.go` (internal/white-box) with 5 tests: TestDescribeSuccess, TestDescribeNotFound, TestDescribeJSON, TestDescribeMissingFlags (table-driven, 3 subtests), TestDescribeCheckVersionFailure
- Registered in `cmd/root.go` via `rootCmd.AddCommand(newDescribeCommand())`

### Key design decisions
- Maps `continuousquery` → `query` before calling drasi client (consistent with status.go and logs.go)
- Uses `describeDrasiClient` interface with `CheckVersion`, `DescribeComponent`, `DescribeComponentInContext` (subset of logsDrasiClient, no RunCommandOutput needed)
- JSON output wraps result in `{"status": "ok", "kind": ..., "component": ..., "detail": ...}` matching status.go pattern
- Table output uses `output.Format(detail, output.FormatTable)` which renders ComponentDetail struct fields via reflection
- Not-found error message includes `kind/id` for clear identification

### Patterns reused
- `var newDescribeDrasiClient` package-level function var for test mocking (same as status/logs/diagnose)
- `fakeDescribeClient` struct with field-based behavior control (same pattern as `fakeLogsClient`)
- `t.Cleanup(func() { newDescribeDrasiClient = orig })` for teardown
- `writeCommandError` with `errorCodeFromError` for all error paths
- `resolvedKubeContextForCommand` for kube context resolution

### ComponentDetail struct (internal/drasi/describe.go)
- Fields: ID, Kind, Status, ErrorReason (all string, all exported)
- ComponentNotFoundError type exists for typed not-found errors
- Client method `describeComponent` parses key-value CLI output (ID, Kind, Status, ErrorReason lines)

## [2026-04-07] Task 9: Status All-Kinds Mode

### What was done
- Removed `selectedKind = "source"` default from `cmd/status.go`
- When no `--kind` flag: loops over all 4 kinds (`source`, `continuousquery`, `middleware`, `reaction`)
- Table output: section headers per kind (Sources:, Queries:, Middleware:, Reactions:) + placeholder for empty kinds
- JSON output: `{"status":"ok","sources":[...],"queries":[...],"middleware":[...],"reactions":[...]}` with never-null arrays
- Single-kind mode (`--kind` provided): backward-compatible, unchanged JSON shape with `kind`+`components` keys
- `CheckVersion` called once before the loop/single-kind logic

### Key design decisions
- `allComponentKinds` slice defines canonical order; `kindDisplayNames` and `kindJSONKeys` maps for human-friendly names
- `wireKindForDrasiCLI` helper maps `continuousquery` → `query` (drasi CLI wire format)
- `nonNilResources` helper ensures JSON never has `null` arrays (returns `[]` for nil slices)
- Used `strings.Builder` for table output assembly to avoid interleaved writes

### Test changes
- Updated `fakeStatusClient` with `componentsByKind map[string][]drasi.ComponentSummary` and `calledKinds []string` tracking
- Updated `TestStatusCommand_TableSuccess_DefaultKind` to verify all 4 kinds queried with section headers
- Added: `TestStatusAllKinds`, `TestStatusSingleKind`, `TestStatusAllKindsJSON`, `TestStatusEmpty`, `TestStatusSingleKindJSON`
- All 12 status tests pass (including 4 from status_test.go external tests)

### Gotcha
- Pre-existing build failure in `logs_internal_test.go` (missing `strings` import) prevents `go test ./...` from passing for cmd package; not related to status changes

## [2026-04-07] Task 10: kubectl logs for non-query kinds

### What was done
- Removed `kind != "query"` gate (lines 52-61) in `cmd/logs.go`
- Added `kubectlLogsFunc` package-level var (shells out to `kubectl logs -l <selector> -n drasi-system --tail=100`)
- Non-query kinds (source, middleware, reaction) now use kubectl label selector: `drasi.io/kind=<kind>,drasi.io/component=<id>`
- Query/continuousquery kind still uses `drasi watch` path (unchanged)
- JSON output for kubectl path: `{"status":"ok","kind":...,"component":...,"logs":...}`
- Empty output for kubectl path: `"No logs found for <id> (<kind>)."`
- kubeContext prepended as `--context <ctx>` when present

### Tests
- Fixed 2 broken internal tests: `TestLogsCommand_LogsCallFailure_ReturnsError` renamed to `TestLogsCommand_KubectlFailure_ReturnsError` (now mocks `kubectlLogsFunc`), `TestLogsCommand_DescribeFailure_ReturnsError` changed to use `--kind query` (still tests drasi client path)
- Added 7 new tests: `TestLogsSourceKind`, `TestLogsReactionKind`, `TestLogsMiddlewareKind`, `TestLogsContinuousQueryKind`, `TestLogsSourceKind_JSON`, `TestLogsSourceKind_EmptyOutput`, `TestLogsSourceKind_WithKubeContext`

### Key design decisions
- kubectl branch runs BEFORE `client.CheckVersion()` so non-query kinds don't need a drasi client at all
- Namespace `drasi-system` confirmed by grep across codebase (keyvault translator test, listen.go, provision.go comments)
- `kubectlLogsFunc` follows same pattern as `isDaprReady` in diagnose.go (package-level var + `exec.CommandContext`)
- Label selector `drasi.io/kind=<kind>,drasi.io/component=<id>` is the standard Drasi labeling convention

### Gotchas
- Existing tests `TestLogsCommand_LogsCallFailure_ReturnsError` and `TestLogsCommand_DescribeFailure_ReturnsError` both used `--kind source` but expected `ERR_VALIDATION_FAILED` from the gate. After removing the gate, they needed different mocking strategies.
- External test `TestLogsCommand_RootEnvironmentFlagAccepted` still works because it fails at `resolvedKubeContextForCommand` before reaching the kubectl branch.

## [2026-04-07] Task 7: Deploy rollback on failure

### What was done
- Added `NoRollback bool` to `internal/deployment.DeployOptions`
- Updated `Engine.Deploy()` to track successfully applied components after state writes succeed, then issue rollback deletes in reverse order when a later component fails
- Rollback uses `slog.WarnContext` for delete failures and still returns the original deploy error unchanged
- Added `--no-rollback` flag to `cmd/deploy.go` and wired it through to `deployment.DeployOptions`

### Test pattern
- Reused `mockDrasiRunner.commandsCalled` sequencing to assert rollback order without introducing new mocks
- Failure cases are triggered on `wait` for the second component so the first component is fully applied and eligible for rollback
- Rollback-failure coverage checks that a delete error is attempted and ignored, preserving the original deploy failure

### Key observation
- To avoid rolling back components whose state persistence failed, `appliedComponents` should only be appended after `WriteHash` succeeds, even though the original task sketch suggested appending earlier

## [2026-04-07] Task 11: Interactive Confirmation Prompts

### What was done
- Created `cmd/confirm.go` with `ConfirmDestructive(prompt string, force bool) (bool, error)` function
- Modified `cmd/teardown.go` to replace hardcoded `--force` gate with `ConfirmDestructive` call (context-aware prompt for infrastructure teardown)
- Modified `cmd/upgrade.go` with same pattern
- Created `cmd/confirm_test.go` (white-box, package `cmd`) with 4 unit tests for `ConfirmDestructive`
- Created `cmd/teardown_internal_test.go` with 2 integration tests (force skip + non-interactive error)
- Created `cmd/upgrade_internal_test.go` with 2 integration tests (force skip + non-interactive error)

### Key design decisions
- Used `mattn/go-isatty` for TTY detection (already in go.mod as indirect dep, recommended by Go instructions)
- Used `AlecAivazis/survey/v2` for interactive prompts (already in go.mod as indirect dep from azd SDK)
- `confirmFunc` typed as `func(string, *bool) error` (not `func(string, interface{}) error` from spec) for type safety
- `isTTYFunc` checks both `IsTerminal` and `IsCygwinTerminal` for Windows Cygwin support
- Non-TTY without `--force` returns `commandError` with `ERR_FORCE_REQUIRED` exit code (consistent with existing error pattern)

### Test patterns
- `saveConfirmVars(t *testing.T)` helper in `confirm_test.go` saves and restores both `confirmFunc` and `isTTYFunc` using `t.Cleanup`
- Tests that mutate package-level vars do NOT use `t.Parallel()` (data race prevention)
- Internal test files (`*_internal_test.go`) in package `cmd` for white-box access to `confirmFunc`/`isTTYFunc`
- External tests (`teardown_test.go`, `upgrade_test.go` in package `cmd_test`) still pass unchanged because test stdin is not a TTY

### Gotchas
- Existing black-box tests pass unchanged because in test environment `os.Stdin` is not a TTY, so `isTTYFunc()` returns false, which triggers the same `ERR_FORCE_REQUIRED` error path as the old hardcoded check
- `confirmFunc` signature uses `*bool` (not `interface{}`) to match `survey.AskOne` typed result pattern
- `survey/v2` was already in go.mod as indirect; importing it directly just moves it from indirect to direct require

## [2026-04-07] Task 12: Wire Progress Spinners into Long-Running Commands

### What was done
- Wired `ProgressHelper` from `cmd/progress.go` into 4 commands: `provision.go`, `deploy.go`, `teardown.go`, `upgrade.go`
- Each command creates the spinner after early-return guards (format parsing, --force checks, confirmation prompts)
- Spinner creation uses graceful degradation: if `NewProgressHelper()` fails, falls back to `&ProgressHelper{noop: true}`
- `defer func() { _ = progress.Stop() }()` ensures cleanup on all paths
- Explicit `_ = progress.Stop()` called before final output to avoid spinner interfering with stdout

### Phase messages per command
- **provision.go**: "Resolving environment..." → "Installing Drasi runtime..." → "Applying network policies..." → "Finalizing environment state..."
- **deploy.go**: "Resolving environment..." → "Validating configuration..." → "Deploying components..."
- **teardown.go**: "Resolving environment..." → "Tearing down components..." → "Deleting Azure infrastructure..." (conditional)
- **upgrade.go**: "Resolving cluster context..." → "Upgrading Drasi runtime..."

### Spinner placement rules followed
1. Create after early-return guards (--force, confirmation, format parsing)
2. Start before first slow operation
3. Message updates between phases
4. Stop before final stdout output (explicit call + defer as safety net)
5. Never call Message before Start

### Key observations
- yacspin works fine with `bytes.Buffer` writers in tests (no TTY required for basic operation)
- No test regressions; spinner output goes to stderr which tests don't assert against
- The `noop` field on `ProgressHelper` is unexported but accessible within the `cmd` package, enabling the fallback pattern
- teardown.go and upgrade.go had been modified since the task spec was written (added `ConfirmDestructive` interactive prompts); spinner placement adapted accordingly
- Using `_ =` pattern for ignoring Start/Stop errors is cleaner than `//nolint:errcheck` comments

## [2026-04-07] F2 Final Wave: Code Quality Review

### Verdict
Build PASS | Vet PASS | Files 24 clean / 0 issues | VERDICT: APPROVE

### What was checked
- `go vet ./...` — exit 0, zero issues
- `go build ./...` — exit 0
- Manual line-by-line review of all 24+ changed files (new + modified)
- Automated grep scans for anti-patterns (all clean):
  - `fmt.Println` in production code — 0 hits
  - `TODO` / `FIXME` / `HACK` in production code — 0 hits
  - `REPLACE_ME` / `changeme` / `placeholder` — 0 hits (2 test comment hits, acceptable)
  - Bare type assertions without `ok` — 0 hits
  - Empty error handlers (`if err != nil {}`) — 0 hits
  - `_ = err` suppressed errors — 0 hits (all `_ =` are intentional: progress.Start/Stop, fmt.Fprint write counts)
  - `fmt.Print` / `fmt.Printf` in cmd/ (should use cmd.OutOrStdout()) — 0 hits

### Patterns verified
- All error paths use `writeCommandError()` + `errorCodeFromError()`
- All spinner/progress output goes to `cmd.ErrOrStderr()`
- All JSON output uses `cmd.OutOrStdout()`
- Client mocking uses `var newXxxDrasiClient` pattern consistently
- Tests use manual struct mocks (no gomock), testify assert/require, t.Parallel() where safe
- `applyDefaultProviders` not present (correctly removed in Task 3)
- `--follow` flag not registered in logs.go (correctly removed in Task 3)
- Deploy lock key is `DRASI_DEPLOY_IN_PROGRESS`
- Telemetry nil meter = silent no-op
- No AI slop patterns detected (no excessive narrating comments, no over-abstraction)
