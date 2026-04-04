---
applyTo: "**/*_test.go"
description: "Go testing best practices for table-driven tests, testify, testcontainers, benchmarks, and deterministic test design"
---

# Go Test Instructions

All conventions from [go.instructions.md](go.instructions.md) apply, including naming, error handling, and context propagation.

**IMPORTANT**: Use the `iseplaybook` MCP server for ISE testing best practices. Use `context7` MCP server (`/golang/go`) for Go testing package API verification.

## Test Pyramid and Scope

| Level       | Scope                                                 | Speed   | Owns                                        |
| ----------- | ----------------------------------------------------- | ------- | ------------------------------------------- |
| Unit        | Single function/method, no I/O                        | < 10 ms | Business logic, validation, transformations |
| Integration | Multiple packages, real or containerized dependencies | < 10 s  | Database queries, HTTP handlers, middleware |
| E2E         | Full binary or HTTP server                            | < 30 s  | Critical API journeys, smoke tests          |

Unit tests run on every `go test ./...`. Integration tests are gated by build tags. Run both in CI.

## Test File Conventions

- Test files live alongside source files with `_test.go` suffix
- Same package for white-box tests (access unexported symbols)
- `_test` package suffix for black-box tests (test public API only)

```text
internal/
  services/
    user.go
    user_test.go              # white-box unit tests
  handlers/
    user_handler.go
    user_handler_test.go      # black-box integration tests (package handlers_test)
tests/
  integration/
    db_test.go                # integration tests with //go:build integration
```

## Test Naming

Format: `Test<Unit>_<Scenario>` or `Test<Unit>_<Scenario>_<Expected>`:

```go
func TestUserService_Create(t *testing.T) { ... }
func TestValidateEmail_EmptyInput_ReturnsError(t *testing.T) { ... }
func TestParseConfig_MissingField_UsesDefault(t *testing.T) { ... }
```

## Unit Tests

```go
func TestUserService_Create(t *testing.T) {
    t.Parallel()

    repo := &mockUserRepo{}
    svc := NewUserService(repo)

    user := &User{Name: "Alice", Email: "alice@example.com"}
    err := svc.Create(context.Background(), user)

    require.NoError(t, err)
    assert.NotEmpty(t, user.ID)
}
```

Rules:

- Call `t.Parallel()` on every test that doesn't share mutable state
- Use `require` for preconditions that must hold (stops the test immediately)
- Use `assert` for expectations (allows test to continue and report all failures)
- Use `context.Background()` in unit tests; use `context.WithTimeout` in integration tests

## Table-Driven Tests

The standard Go pattern for testing multiple cases without duplication:

```go
func TestValidateEmail(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {name: "valid email", input: "user@example.com", wantErr: false},
        {name: "missing @", input: "notanemail", wantErr: true},
        {name: "empty string", input: "", wantErr: true},
        {name: "missing domain", input: "user@", wantErr: true},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()

            err := validateEmail(tc.input)
            if tc.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

Rules:

- Always include a `name` field for readable subtest output
- Call `t.Parallel()` on each subtest when tests are independent
- Go 1.22+ captures loop variables correctly; `tc := tc` is no longer needed
- Keep the test struct focused; avoid embedding complex assertions in the struct

## Test Helpers and Fixtures

Use helper functions with `t.Helper()` for shared setup:

```go
func newTestUser(t *testing.T, overrides ...func(*User)) *User {
    t.Helper()
    user := &User{
        ID:    "test-id",
        Name:  "Test User",
        Email: "test@example.com",
    }
    for _, fn := range overrides {
        fn(user)
    }
    return user
}

func TestUserService_Update(t *testing.T) {
    t.Parallel()

    user := newTestUser(t, func(u *User) {
        u.Name = "Updated Name"
    })

    // ... test logic
}
```

Rules:

- Always call `t.Helper()` so failures report the caller's line, not the helper's
- Register cleanup with `t.Cleanup()` for resources that must be released
- Prefer function overrides or option functions over many helper variants

## Mocking at Interface Boundaries

Define small interfaces at the consumer and mock them in tests:

```go
// Defined in the service package (consumer)
type userRepository interface {
    Get(ctx context.Context, id string) (*User, error)
    Create(ctx context.Context, user *User) error
}

// Manual mock in test file
type mockUserRepo struct {
    getFunc    func(ctx context.Context, id string) (*User, error)
    createFunc func(ctx context.Context, user *User) error
}

func (m *mockUserRepo) Get(ctx context.Context, id string) (*User, error) {
    if m.getFunc != nil {
        return m.getFunc(ctx, id)
    }
    return nil, ErrNotFound
}

func (m *mockUserRepo) Create(ctx context.Context, user *User) error {
    if m.createFunc != nil {
        return m.createFunc(ctx, user)
    }
    return nil
}
```

Rules:

- Mock at interface boundaries; never mock concrete types
- Prefer manual mocks for small interfaces (1-3 methods)
- Use `mockall`/`gomock`/`moq` for interfaces with many methods
- Define interfaces at the point of use, not the point of implementation

## Integration Tests

Gate integration tests with build tags so `go test ./...` stays fast:

```go
//go:build integration

package db_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/require"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestPostgresUserRepo_Create(t *testing.T) {
    ctx := context.Background()

    container, err := postgres.Run(ctx,
        "postgres:16-alpine",
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
    )
    require.NoError(t, err)
    t.Cleanup(func() { container.Terminate(ctx) })

    connStr, err := container.ConnectionString(ctx, "sslmode=disable")
    require.NoError(t, err)

    repo, err := NewPostgresUserRepo(connStr)
    require.NoError(t, err)

    user := &User{Name: "Alice", Email: "alice@example.com"}
    err = repo.Create(ctx, user)
    require.NoError(t, err)
    assert.NotEmpty(t, user.ID)

    found, err := repo.Get(ctx, user.ID)
    require.NoError(t, err)
    assert.Equal(t, "Alice", found.Name)
}
```

Run with: `go test -tags=integration ./...`

Rules:

- Use `//go:build integration` to separate from unit tests
- Use `testcontainers-go` for database, queue, and cache dependencies
- Always register cleanup with `t.Cleanup()` for container teardown
- Use `context.WithTimeout` to prevent tests from hanging

## HTTP Handler Tests

Use `httptest` for testing HTTP handlers without starting a real server:

```go
func TestGetUser_ReturnsJSON(t *testing.T) {
    t.Parallel()

    repo := &mockUserRepo{
        getFunc: func(_ context.Context, id string) (*User, error) {
            return &User{ID: id, Name: "Alice"}, nil
        },
    }
    handler := NewUserHandler(repo)

    req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
    req.SetPathValue("id", "42") // Go 1.22+ path values
    rec := httptest.NewRecorder()

    handler.GetUser(rec, req)

    require.Equal(t, http.StatusOK, rec.Code)

    var user User
    err := json.NewDecoder(rec.Body).Decode(&user)
    require.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)
}

func TestGetUser_NotFound_Returns404(t *testing.T) {
    t.Parallel()

    repo := &mockUserRepo{
        getFunc: func(_ context.Context, _ string) (*User, error) {
            return nil, ErrNotFound
        },
    }
    handler := NewUserHandler(repo)

    req := httptest.NewRequest(http.MethodGet, "/users/999", nil)
    req.SetPathValue("id", "999")
    rec := httptest.NewRecorder()

    handler.GetUser(rec, req)

    assert.Equal(t, http.StatusNotFound, rec.Code)
}
```

## Benchmarks

```go
func BenchmarkParseConfig(b *testing.B) {
    data := loadTestConfig(b)

    b.ResetTimer()
    for range b.N {
        _, err := ParseConfig(data)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkParseConfig_Allocs(b *testing.B) {
    data := loadTestConfig(b)

    b.ReportAllocs()
    b.ResetTimer()
    for range b.N {
        ParseConfig(data)
    }
}
```

Run with: `go test -bench=. -benchmem ./...`

Rules:

- Use `b.ResetTimer()` after setup to exclude initialization time
- Use `b.ReportAllocs()` to track memory allocations
- Use `for range b.N` (Go 1.22+) instead of `for i := 0; i < b.N; i++`
- Don't run benchmarks in CI by default; gate them for performance regression checks

## Test Anti-Patterns

| Anti-Pattern                | Why It's Harmful                                  | Fix                                            |
| --------------------------- | ------------------------------------------------- | ---------------------------------------------- |
| Missing `t.Parallel()`      | Tests run sequentially, CI is slow                | Add `t.Parallel()` to independent tests        |
| No `t.Helper()` in helpers  | Failures report the helper line, not the caller   | Add `t.Helper()` to all test helper functions  |
| `time.Sleep` in tests       | Flaky, slow, non-deterministic                    | Use channels, `sync.WaitGroup`, or context     |
| Mocking concrete types      | Tightly couples tests to implementation           | Define interfaces at consumer; mock those      |
| Missing cleanup             | Resource leaks, port conflicts, container orphans | Use `t.Cleanup()` for all resources            |
| Ignored errors              | False positives; test passes when it shouldn't    | Check every error; use `require.NoError`       |
| No build tag on integration | Integration tests break `go test ./...`           | Add `//go:build integration`                   |
| Shared mutable state        | Order-dependent, racy test failures               | Fresh state per test; use `t.Parallel()` guard |

## Tooling and CI Integration

```yaml
# Minimum CI test pipeline
steps:
  - run: go vet ./...
  - run: go test -race -count=1 ./...
  - run: go test -tags=integration -race ./... # separate job with credentials
  - run: go test -bench=. -benchmem ./... # optional: performance gating
```

- Always run with `-race` to detect data races
- Use `-count=1` to defeat test caching when needed
- Run `govulncheck ./...` on dependency changes

## Final Self-Check (Before Proposing Test Changes)

✅ Tests call `t.Parallel()` when independent
✅ Helpers call `t.Helper()`
✅ Resources cleaned up with `t.Cleanup()`
✅ Table-driven tests used for multiple cases
✅ `require` for preconditions, `assert` for expectations
✅ Mocks defined at interface boundaries
✅ Integration tests gated with `//go:build integration`
✅ HTTP handlers tested with `httptest`
✅ No `time.Sleep` for synchronization
✅ Tests run clean with `-race` flag
✅ Errors checked in every test path
