---
applyTo: "**/*.go"
description: "Go development best practices for building production-quality Go services, tools, and libraries. Covers module management, idiomatic naming, error handling, concurrency, interfaces, testing, and tooling aligned with ISE Engineering Playbook guidelines."
---

# Go Instructions

Conventions for Go development targeting Go 1.21+ (the current stable baseline with `log/slog` and `slices`/`maps` packages). These conventions apply to services, CLI tools, and libraries unless noted otherwise.

## Project Structure

Go modules follow a standard layout:

```text
go.mod
go.sum
main.go               # Entry point for simple single-binary projects
cmd/
  myapp/
    main.go           # Entry point for multi-binary projects
internal/
  config/             # Internal packages — not importable by other modules
  handlers/
  models/
  services/
pkg/
  shared/             # Public packages — importable by external consumers
```

- `go.mod` and `go.sum` at module root. Always commit both.
- Use `internal/` for packages that must not be imported by other modules.
- Use `cmd/<appname>/main.go` for multi-binary projects.
- Keep `main.go` thin: wire dependencies and start the process only.
- Organize by domain responsibility: `config`, `handlers`, `models`, `services`.
- Keep all files at root-level modules when fewer than 10 source files exist; add folders only when complexity justifies it.

## go.mod Conventions

```gomod
module github.com/org/repo

go 1.21

require (
    github.com/spf13/cobra v1.8.0
    go.uber.org/zap v1.27.0
    golang.org/x/sync v0.6.0
)
```

- Pin minimum Go version with the `go` directive.
- Always commit `go.sum` — it provides supply-chain integrity verification.
- Run `go mod tidy` before committing to remove unused dependencies.
- Use `replace` directives only for local development; remove before committing.
- One module per repository unless there is a strong reason for a multi-module layout.

## Naming

| Element              | Convention                     | Example                        |
| -------------------- | ------------------------------ | ------------------------------ |
| Packages             | lowercase, single word         | `handler`, `config`            |
| Exported types       | PascalCase                     | `UserService`, `Config`        |
| Unexported types     | camelCase                      | `userStore`, `parseResult`     |
| Interfaces           | PascalCase, describe behaviour | `Reader`, `EventHandler`       |
| Exported constants   | PascalCase                     | `MaxRetries`, `DefaultTimeout` |
| Unexported constants | camelCase                      | `defaultTimeout`, `maxRetries` |
| Sentinel errors      | PascalCase, `Err` prefix       | `ErrNotFound`, `ErrTimeout`    |
| Test functions       | `Test<Unit>_<Scenario>`        | `TestUserService_Create`       |

Additional rules:

- Interface names often end with `-er`: `Writer`, `Closer`, `Stringer`.
- Avoid stutter: `user.UserService` → prefer `user.Service`.
- Prefer short names for short-lived variables (`i`, `v`, `err`) and descriptive names for package-level identifiers.
- Acronyms are all-caps when exported: `HTTPClient`, `APIKey`, `URL`.
- Avoid `_` in package names and file names (use `camelCase` or `lower` for package names).

## Error Handling

Go errors are values — handle them explicitly at the call site. Never discard errors silently.

### Sentinel Errors

```go
var ErrNotFound = errors.New("not found")
var ErrTimeout  = errors.New("operation timed out")

// Check with errors.Is (respects wrapping)
if errors.Is(err, ErrNotFound) {
    // handle not found
}
```

### Custom Error Types

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// Check with errors.As
var ve *ValidationError
if errors.As(err, &ve) {
    log.Printf("field %s: %s", ve.Field, ve.Message)
}
```

### Error Wrapping

```go
// Wrap with %w to preserve the chain for errors.Is / errors.As
if err := db.Query(ctx, q); err != nil {
    return fmt.Errorf("querying users: %w", err)
}
```

- Always wrap with `%w` (not `%v`) so callers can inspect the cause.
- Add context that helps diagnose the failure without repeating the full call stack.
- Return `nil` explicitly on success.
- Never use `panic` for expected error paths; reserve `panic` for programmer-error invariant violations.
- Empty `catch`-style blocks (discarding `err`) are forbidden — handle, log safely, or return.

## Concurrency

### Context Propagation

```go
func fetchUser(ctx context.Context, id string) (*User, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, fmt.Errorf("creating request: %w", err)
    }
    // ...
}
```

- Pass `context.Context` as the **first parameter** in all I/O-bound functions.
- Never store contexts in structs; pass them through the call chain.
- Use `context.WithTimeout` or `context.WithDeadline` for all network and database calls.

### WaitGroups

```go
var wg sync.WaitGroup

for _, item := range items {
    wg.Add(1)
    go func(item Item) {
        defer wg.Done()
        process(ctx, item)
    }(item)
}
wg.Wait()
```

- Use `defer wg.Done()` immediately after spawning the goroutine.
- Every goroutine must have a defined exit condition: context cancellation, channel close, or WaitGroup signal.

### errgroup for Concurrent Work

```go
import "golang.org/x/sync/errgroup"

g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { return fetchA(ctx) })
g.Go(func() error { return fetchB(ctx) })
if err := g.Wait(); err != nil {
    return fmt.Errorf("concurrent fetch: %w", err)
}
```

Prefer `errgroup` when any error should cancel remaining goroutines.

### Channel Patterns

- Use buffered channels to prevent goroutine leaks.
- Always close channels on the producer side; range over them on the consumer side.
- Prefer `sync.Mutex` for protecting shared state when there is no message-passing intent.
- Use `select` with a `case <-ctx.Done(): return` to avoid blocking indefinitely.

## Interfaces

Define interfaces at the point of use, not at the point of implementation.

```go
// Defined in the package that consumes it (service)
type userRepository interface {
    Get(ctx context.Context, id string) (*User, error)
    Create(ctx context.Context, user *User) error
}

type UserService struct {
    repo userRepository
}

// Implemented in a separate package (db) — no explicit declaration needed
type postgresUserRepo struct{ db *sql.DB }

func (r *postgresUserRepo) Get(ctx context.Context, id string) (*User, error) { ... }
```

- Keep interfaces small — single-method interfaces are idiomatic.
- Use interface composition for larger contracts.
- Accept interfaces in function parameters; return concrete types from constructors.
- Export an interface only when other packages need to satisfy it.

## HTTP Handlers

```go
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    id := r.PathValue("id") // Go 1.22+ net/http routing

    user, err := h.repo.Get(ctx, id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            http.Error(w, "not found", http.StatusNotFound)
            return
        }
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(user); err != nil {
        // Response may be partially written; log but do not write again
        slog.ErrorContext(ctx, "encoding response", slog.Any("error", err))
    }
}
```

- Always check and handle the error from `json.NewEncoder(w).Encode(...)`.
- Set `Content-Type` before writing the body.
- Do not expose internal error details to clients.

## Testing

See [go-tests.instructions.md](go-tests.instructions.md) for full testing conventions (table-driven tests, testify, testcontainers, benchmarks, HTTP handler tests, anti-patterns).

## Logging

Use `log/slog` (Go 1.21+) for structured logging. Avoid `fmt.Printf` for operational output.

```go
import "log/slog"

slog.InfoContext(ctx, "processing request",
    slog.String("user_id", userID),
    slog.String("operation", "create"),
)

slog.ErrorContext(ctx, "database query failed",
    slog.String("query", "get_user"),
    slog.Any("error", err),
)
```

- Use `slog.InfoContext`/`slog.ErrorContext` to propagate context (trace IDs, request IDs).
- Never log secrets, tokens, or PII.
- Use structured key-value pairs, not format strings.
- Configure `slog.SetDefault` at startup with the appropriate handler (JSON for production, text for development).

## Tooling

| Tool                  | Purpose                   | Command                         |
| --------------------- | ------------------------- | ------------------------------- |
| `go vet`              | Built-in static analysis  | `go vet ./...`                  |
| `staticcheck`         | Advanced static analysis  | `staticcheck ./...`             |
| `golangci-lint`       | Meta-linter (50+ linters) | `golangci-lint run`             |
| `gofmt` / `goimports` | Formatting                | `gofmt -w .` / `goimports -w .` |
| `govulncheck`         | Vulnerability scanner     | `govulncheck ./...`             |
| `go test -race`       | Data race detector        | `go test -race ./...`           |

### `.golangci.yml` (v2 format)

Use golangci-lint v2 config format. The `default: standard` preset enables the linters the team considers safe and useful for most projects (`govet`, `errcheck`, `staticcheck`, `unused`, `gosimple`, `ineffassign`, and others). The `common-false-positives` exclusion preset quiets issues that are almost never real bugs (e.g. unchecked errors on `Close()`).

```yaml
# golangci-lint v2 configuration
# Docs: https://golangci-lint.run/docs/configuration/file/
version: "2"

linters:
  default: standard
  exclusions:
    presets:
      - common-false-positives
```

Add explicit linters only when the standard set is insufficient for your use case (e.g. `gosec` for security-sensitive packages).

- Run `go vet ./...` and `golangci-lint run` in every CI pipeline.
- Run `go test -race ./...` to detect data races.
- Run `govulncheck ./...` weekly or on dependency changes.

### Coverage Threshold Enforcement

Use [vladopajic/go-test-coverage](https://github.com/vladopajic/go-test-coverage) or a shell script with `go tool cover` to enforce a minimum coverage threshold in CI. Exclude bootstrap-only `main.go` — it has no testable logic.

```yaml
# .testcoverage.yml
profile: coverage.out
threshold:
  total: 70 # minimum total percentage
exclude:
  paths:
    - cmd/myapp/main\.go # main() bootstrap — no testable logic
```

## Actionable Patterns

### Pattern 1: Error handling — wrapping vs discarding

**❌ WRONG: Discarding or silently swallowing errors**

```go
user, _ := repo.Get(ctx, id)        // ⚠️ error silently discarded

result, err := doSomething()
if err != nil {
    log.Println(err)                // ⚠️ logged but not returned — caller unaware
}
return result
```

**✅ CORRECT: Wrap and return errors with context**

```go
user, err := repo.Get(ctx, id)
if err != nil {
    return nil, fmt.Errorf("get user %s: %w", id, err)
}

result, err := doSomething()
if err != nil {
    return nil, fmt.Errorf("doing something: %w", err)
}
return result, nil
```

**Rule:** Never discard errors with `_`. Always wrap with `%w` to preserve the chain.

---

### Pattern 2: Context propagation

**❌ WRONG: Creating a fresh context in every call**

```go
func fetchUser(id string) (*User, error) {
    ctx := context.Background()     // ⚠️ loses parent cancellation and deadline
    return db.QueryRow(ctx, q, id)
}
```

**✅ CORRECT: Accept and propagate context**

```go
func fetchUser(ctx context.Context, id string) (*User, error) {
    return db.QueryRow(ctx, q, id)  // ✅ respects parent cancellation
}
```

**Rule:** Accept `context.Context` as the first parameter in all I/O-bound functions. Never create `context.Background()` inside a function that could receive one.

---

### Pattern 3: Goroutine leak prevention

**❌ WRONG: Goroutine with no exit condition**

```go
go func() {
    for {
        process(getNextItem())     // ⚠️ leaks forever if nothing stops it
    }
}()
```

**✅ CORRECT: Context-based cancellation**

```go
go func() {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            process(getNextItem())
        }
    }
}()
```

**Rule:** Every goroutine must have a defined exit path via context cancellation, channel close, or WaitGroup.

---

### Pattern 4: Interface definition placement

**❌ WRONG: Interface defined by the implementor**

```go
// in package db
type UserRepository interface {    // ⚠️ forces consumers to import package db
    Get(ctx context.Context, id string) (*User, error)
}
```

**✅ CORRECT: Interface defined by the consumer**

```go
// in package service (where it's used)
type userRepository interface {    // ✅ local, unexported — no import needed
    Get(ctx context.Context, id string) (*User, error)
}
```

**Rule:** Define interfaces in the package that uses them. Satisfy Go's implicit interface implementation without creating import cycles.

---

### Pattern 5: Structured logging

**❌ WRONG: Printf-style log messages**

```go
log.Printf("error processing user %s: %v", userID, err)  // ⚠️ hard to parse
```

**✅ CORRECT: Structured slog with key-value pairs**

```go
slog.ErrorContext(ctx, "processing user failed",
    slog.String("user_id", userID),
    slog.Any("error", err),
)
```

**Rule:** Use `log/slog` (Go 1.21+) with structured key-value pairs. Avoid format strings in log messages.

---

### Pattern 6: Defer for cleanup

**❌ WRONG: Manual cleanup with error-prone branching**

```go
f, err := os.Open(path)
if err != nil {
    return err
}
// ... if an early return is added, the file leaks
f.Close()
```

**✅ CORRECT: defer for guaranteed cleanup**

```go
f, err := os.Open(path)
if err != nil {
    return fmt.Errorf("opening %s: %w", path, err)
}
defer f.Close()  // ✅ always runs, regardless of return path
```

**Rule:** Use `defer` for cleanup of resources opened in the same function: files, database rows, mutexes, HTTP response bodies.

---

### Pattern 7: Table-driven tests

**❌ WRONG: Separate test functions for each case**

```go
func TestValidateEmail_Valid(t *testing.T) { ... }    // ⚠️ duplicated setup
func TestValidateEmail_Missing(t *testing.T) { ... }
func TestValidateEmail_Empty(t *testing.T) { ... }
```

**✅ CORRECT: Single table-driven test**

```go
tests := []struct {
    name    string
    input   string
    wantErr bool
}{
    {"valid", "user@example.com", false},
    {"missing @", "notanemail", true},
    {"empty", "", true},
}
for _, tc := range tests {
    t.Run(tc.name, func(t *testing.T) {
        err := validateEmail(tc.input)
        // assert...
    })
}
```

**Rule:** Prefer table-driven tests for multiple input/output scenarios. They are easier to extend and reduce duplicated setup.

## Best Practices

1. **Simplicity first** — idiomatic Go is explicit and straightforward. Avoid unnecessary abstraction layers.
2. **Explicit over implicit** — Go has no annotations or reflection-based magic. Embrace the explicitness.
3. **Handle errors immediately** — check at the call site; don't defer error handling.
4. **Short functions** — consider splitting functions that exceed 50–60 lines.
5. **Avoid `init()`** — use explicit initialization; `init()` functions are hard to test and trace.
6. **Dependency injection** — pass dependencies through constructors, not global state or `sync.Once`.
7. **Standard library first** — Go's stdlib is rich. Reach for third-party packages only when stdlib is insufficient.
8. **Use `//go:build`** (not the older `// +build`) for conditional compilation.
9. **Embed static assets** — use `//go:embed` rather than reading files at runtime. Place embed directives in a dedicated `embed.go` file at the package root. Pass the `embed.FS` through constructors, never as a global.
10. **Run `-race` in CI** — enable the race detector in all CI test runs, not just locally.

## CLI Tools (cobra-based)

These patterns apply when building Go CLI tools with [cobra](https://github.com/spf13/cobra).

### Ultra-thin `main.go`

```go
package main

import (
    "fmt"
    "os"
    "github.com/org/myapp/internal/cmd"
)

func main() {
    rootCmd := cmd.NewRootCommand()
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)  // errors to stderr, not stdout
        os.Exit(1)
    }
}
```

- `main()` wires dependencies and starts the process only. No business logic.
- Print errors to `os.Stderr` before `os.Exit(1)`. Cobra errors go to stderr by default when `SilenceErrors` is false, but explicit printing ensures consistent formatting.

### Root Command Settings

```go
func NewRootCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:          "myapp",
        Short:        "Short description",
        SilenceUsage:  true,  // suppress usage on errors — avoids noisy output
        SilenceErrors: true,  // errors are printed by main(), not cobra
    }

    cmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
    cmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress output (CI mode)")
    cmd.PersistentFlags().Bool("json", false, "JSON output")

    // Cobra enforces mutual exclusion — avoids manual flag validation
    cmd.MarkFlagsMutuallyExclusive("verbose", "quiet")
    cmd.MarkFlagsMutuallyExclusive("verbose", "json")
    cmd.MarkFlagsMutuallyExclusive("quiet", "json")

    cmd.AddCommand(newVersionCommand())
    return cmd
}
```

Always set `SilenceUsage: true` for tools. Printing usage on every error is noisy and unhelpful for users who already know the syntax.

### Testable Output — `cmd.OutOrStdout()` / `cmd.ErrOrStderr()`

Never write directly to `os.Stdout` or `os.Stderr` in command handlers. Use cobra's writer accessors so tests can capture output:

```go
// In command handler
_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Result: ", value)
_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s\n", msg)

// In tests
cmd.SetOut(&bytes.Buffer{})
cmd.SetErr(&bytes.Buffer{})
```

### Typed Output Mode

Define a typed enum for output modes and read it once at the entry point of each command:

```go
type OutputMode int

const (
    OutputNormal  OutputMode = iota
    OutputVerbose
    OutputQuiet
    OutputJSON
)

func getOutputMode(cmd *cobra.Command) OutputMode {
    if v, _ := cmd.Flags().GetBool("json"); v    { return OutputJSON }
    if v, _ := cmd.Flags().GetBool("verbose"); v { return OutputVerbose }
    if v, _ := cmd.Flags().GetBool("quiet"); v   { return OutputQuiet }
    return OutputNormal
}
```

Pass `OutputMode` through helpers rather than reading flags repeatedly. This makes output behaviour easier to test.

### Version Stamping

Use ldflags for release builds and `debug.ReadBuildInfo` as a fallback for `go install` builds:

```go
// Set by goreleaser: -X github.com/org/myapp/internal/cmd.Version={{.Version}}
var Version string

func getVersion() string {
    if Version != "" {
        return Version
    }
    info, ok := debug.ReadBuildInfo()
    if !ok {
        return "dev"
    }
    if info.Main.Version != "" && info.Main.Version != "(devel)" {
        return info.Main.Version
    }
    return "dev"
}
```

In `.goreleaser.yml`: set `CGO_ENABLED=0` for static cross-compiled binaries and use `-s -w` to strip debug symbols:

```yaml
builds:
  - main: ./cmd/myapp
    env:
      - CGO_ENABLED=0
    ldflags:
      - -s -w -X github.com/org/myapp/internal/cmd.Version={{.Version}}
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
```

### TTY Detection for Interactive Prompts

When a command may prompt the user for confirmation, detect whether stdin and stdout are interactive terminals. Expose the check as a replaceable variable so tests can override it without spawning a TTY:

```go
import "github.com/mattn/go-isatty"

// isInteractiveFunc is a package-level var so tests can override it.
var isInteractiveFunc = func() bool {
    inFd  := os.Stdin.Fd()
    outFd := os.Stdout.Fd()
    // Check both IsTerminal AND IsCygwinTerminal for Windows Cygwin support.
    stdinIsTTY  := isatty.IsTerminal(inFd)  || isatty.IsCygwinTerminal(inFd)
    stdoutIsTTY := isatty.IsTerminal(outFd) || isatty.IsCygwinTerminal(outFd)
    return stdinIsTTY && stdoutIsTTY
}
```

In non-interactive mode (CI, piped input), abort operations that require confirmation unless `--yes` is passed.

### Path Traversal Guard

When writing files to a user-controlled target directory (e.g. install commands), always verify the resolved output path stays within the target root:

```go
absTarget, _ := filepath.Abs(targetDir)
absPath, _   := filepath.Abs(outputPath)
relPath, err := filepath.Rel(absTarget, absPath)
if err != nil || relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
    return fmt.Errorf("path %s escapes target directory", outputPath)
}
```

This prevents zip-slip style attacks where embedded file paths contain `../` traversal sequences.

### Embedded Filesystem

Place embed directives in a dedicated `embed.go` at the package root. Pass `embed.FS` through constructors — never reference it as a global in business logic:

```go
// embed.go (root package)
package myapp

import "embed"

//go:embed templates
var Content embed.FS
```

```go
// cmd/myapp/main.go
rootCmd := cmd.NewRootCommand(myapp.Content)
```

When walking an `embed.FS`, use `fs.WalkDir`, not `filepath.WalkDir`:

```go
fs.WalkDir(content, "templates", func(path string, d fs.DirEntry, err error) error {
    // ...
})
```

### Deterministic Output Helpers

For stable CLI output and tests, avoid direct map iteration (order is non-deterministic):

```go
// sortedKeys returns map keys sorted alphabetically.
func sortedKeys[V any](m map[string]V) []string {
    keys := make([]string, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    return keys
}

// nonNil converts a nil slice to an empty slice for JSON serialisation.
// Prevents null arrays in JSON output.
func nonNil(s []string) []string {
    if s == nil {
        return []string{}
    }
    return s
}
```

## References

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Proverbs](https://go-proverbs.github.io/)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [golangci-lint](https://golangci-lint.run/)
- [ISE Engineering Playbook](https://microsoft.github.io/code-with-engineering-playbook/)
- [willvelida/code-minions](https://github.com/willvelida/code-minions) — reference Go CLI implementation
