---
name: creating-azd-extensions
description: >-
  Author, build, and publish Azure Developer CLI (azd) extensions in Go. USE FOR: creating new azd extensions, implementing lifecycle hooks, exposing custom CLI commands, building MCP server capabilities, writing extension.yaml manifests, cross-platform build scripts, distributing via registry or local sources.
version: 1.7.0
lastUpdated: 2026-04-06
---

# Creating Azure Developer CLI Extensions

Build Go binaries that extend `azd` with custom commands, lifecycle hooks, and MCP tools. Extensions communicate with the azd host via gRPC and are distributed through registries.

> âš ï¸ **Beta feature**: The azd extensions framework is in beta. APIs may change between azd releases. Verify your target azd version supports the capabilities you need: `azd version` (requires 1.10.0+).

## What's New in v1.6.0 (2026-04-07)

- âœ… **Metadata capability is now a default expectation** â€” production azd extensions should declare `metadata` and ship a hidden `metadata` command (or packaged `metadata.json`) so help/IntelliSense stay accurate.
- âœ… **Prefer SDK helpers over hand-rolled listener commands** â€” use `azdext.NewListenCommand(...)` and `azdext.NewMetadataCommand(...)` with the pinned azd SDK instead of manually wiring event managers unless you have a proven reason not to.
- âœ… **ExtensionHost is the recommended lifecycle registration surface** â€” register lifecycle handlers with `NewExtensionHost(...).WithProjectEventHandler(...)` / `WithServiceEventHandler(...)`; do not treat hidden `listen` as a user-facing command.
- âœ… **Docs must match the public command surface** â€” never list hidden `listen`/`metadata` commands in README command tables or user quickstarts.

## What's New in v1.7.0 (2026-04-06)

This update captures significant patterns from azd v1.23.7â€“v1.23.13 (March 2026) with no breaking changes to extension SDK APIs.

- âœ… **NewVersionCommand helper with JSON support** â€” use zdext.NewVersionCommand(versionString) instead of hand-rolling version display; supports --output json for automation.
- âœ… **MCP Security Policy is production-mandatory** â€” extensions exposing MCP tools must use NewMCPSecurityPolicy() (blocks cloud metadata endpoints, RFC 1918 private networks, enforces HTTPS, redacts sensitive headers, validates symlinks). Not optional for production.
- âœ… **Service event filtering with language/host targeting** â€” use ServiceEventOptions{Host: "aks", Language: "dotnet"} in lifecycle handlers for prepackage/postpackage hooks that only apply to specific deployment targets.
- âœ… **Copilot Service gRPC integration** â€” new zdClient.Copilot().<Method>() service available; optional for AI-assisted diagnostics and extension recommendations (future pattern, not yet widely adopted).
- âœ… **Extension source validation** â€” zd extension source validate command; validate extension.yaml and binary authenticity before installation (operator-level feature).
- âœ… **Better startup error diagnostics** â€” extension startup failures now categorized by layer (binary exec, gRPC bind, metadata parse, handler init); clearer error messages in azd output.
- âœ… **Extension registry `website` field** â€” optionally add website: "https://your-docs" to extension.yaml for discoverable documentation links in registry UI.

**Impact on Drasi extension**: MCP security is now a requirement if Drasi exposes AI tools to agents. Service event filtering is useful for environment-specific (AKS vs Container Apps) packaging logic.

**Evidence**: https://github.com/Azure/azure-dev/blob/main/cli/azd/docs/extensions/extension-sdk-reference.md (v1.7.0 APIs), March 2026 release blog: https://devblogs.microsoft.com/azure-sdk/azure-developer-cli-azd-march-2026/

## What's New in v1.3.0 (2026-04-06)

- âœ… **Environment state API correction** â€” correct method names are `GetValue`/`SetValue` (NOT `GetEnvironmentValue`/`SetEnvironmentValue`)
- âœ… **`azdext.AzdClient` is a concrete struct** â€” define consumer-side interfaces in packages that need testability
- âœ… **White-box injection pattern** â€” package-level `var runXxxFunc = defaultRunXxx` enables testing without live gRPC
- âœ… **`t.Parallel()` restriction for mutating command tests** â€” tests that mutate package-level vars must NOT call `t.Parallel()`
- âœ… **`//go:embed all:templates` requirement** â€” the `all:` prefix is mandatory to include dotfiles in embedded FS
- âœ… **`commandError` pattern** â€” error code embedded in `.Error()` string; `writeCommandError` vs `notImplemented` distinction

## What's New in v1.4.0 (2026-04-06)

- âœ… **Environment-aware operational commands** â€” for status/logs/diagnose, root `--environment` now resolves `AZURE_AKS_CONTEXT` from azd env state and routes Drasi/Kubectl calls to the correct cluster context.
- âœ… **Context-capable Drasi wrappers** â€” add `ListComponentsInContext` / `DescribeComponentInContext` so CLI wrappers can prepend `--context` deterministically instead of relying on ambient kube context.
- âœ… **Force-gated destructive/runtime operations** â€” `teardown` and `upgrade` require `--force`; return `ERR_FORCE_REQUIRED` when omitted.
- âœ… **No-stub production rule enforcement** â€” remove `ERR_NOT_IMPLEMENTED` runtime paths from shipping command surface; keep placeholders out of non-test code.
- âœ… **Command testability seam pattern (recommended)** â€” use command-local injectable factories (e.g., `newLogsDrasiClient`, `newStatusDrasiClient`) to test success/error/output branches without live azd or Drasi.
- âœ… **Toolchain preflight for CI/dev** â€” verify `go`, `gopls`, and `golangci-lint` are installed before claiming build/test validation; fail closed with explicit blocker reporting when missing.

## What's New in v1.5.0 (2026-04-07)

- âœ… **Windows PATH persistence pitfall** â€” Go may be installed (`C:\Program Files\Go\bin\go.exe`) but absent from the active shell PATH. Always verify command resolution, then patch session PATH before build/test.
- âœ… **Persistent user PATH remediation** â€” when needed, set user PATH via `[Environment]::SetEnvironmentVariable("Path", ..., "User")` and include `C:\Program Files\Go\bin`; re-open shells to inherit.
- âœ… **Preflight commands (mandatory before reporting failures):** `Get-Command go`, `go version`, `go env GOPATH GOROOT`, and PATH inspection. Do not conclude â€œGo not installedâ€ from `where go` alone.
- âœ… **Runtime-verified compatibility checks** â€” after toolchain repair, run both `go test ./...` and `go build ./...` plus extension `build.ps1` to validate cross-arch artifact generation.
- âœ… **Drasi CLI version output normalization** â€” parse `drasi version` outputs that include labels/prefixes (e.g., `Drasi CLI version: 0.10.0`, `v0.10.0`) before semver comparison.

## What's New in v1.2.0 (2026-04-05)

- âœ… **`azdext.WithAccessToken(ctx)` requirement** â€” must be called before any gRPC service call
- âœ… **Available gRPC services** reference table
- âœ… **`listen` subcommand** â€” required hidden convention when `lifecycle-events` capability is declared

## What's New in v1.1.0 (2026-04-07)

- âœ… **Known SDK gotchas** â€” correct `azsecrets` import path, `log/slog` requirement
- âœ… **Go standards checklist** â€” error naming, `t.Parallel()`, table-driven tests, structured logging
- âœ… **Bicep / IaC guidance** â€” AVM-first mandate, API version verification

## What's New in v1.0.0 (2026-03-28)

- âœ… **Initial skill** covering azd extension authoring from scratch
- âœ… **Full extension.yaml schema** with all capability declarations
- âœ… **Lifecycle event patterns** reverse-engineered from `microsoft.azd.demo` and `azure.ai.agents` extension sources
- âœ… **Cross-platform build scripts** (build.ps1 + build.sh)
- âœ… **Registry distribution** patterns (dev + official + custom sources)

## Quick Start

```bash
# Verify azd supports extensions
azd version                              # Requires 1.10.0+

# Explore available extensions
azd extension list --source dev

# Bootstrap a new extension module
mkdir my-extension && cd my-extension
go mod init github.com/org/my-extension
go get github.com/azure/azure-dev/cli/azd/pkg/azdext@latest
go get github.com/spf13/cobra@latest
```

## Extension File Structure

```
my-extension/
â”œâ”€â”€ extension.yaml              # Manifest â€” required
â”œâ”€â”€ version.txt                 # Semantic version string â€” required
â”œâ”€â”€ main.go                     # Entry point: azdext.Run(cmd.NewRootCommand())
â”œâ”€â”€ internal/
â”‚   â””â”€â”€ cmd/
â”‚       â”œâ”€â”€ root.go             # Cobra command tree
â”‚       â””â”€â”€ listen.go           # Lifecycle event subscriptions
â”œâ”€â”€ build.ps1                   # Windows cross-compile build script
â”œâ”€â”€ build.sh                    # Linux/macOS cross-compile build script
â””â”€â”€ go.mod / go.sum
```

## extension.yaml Manifest

```yaml
name: my-extension
namespace: myext # CLI prefix: azd myext <command>
version: 1.0.0
description: "Short description of what this extension does."
usage: "azd myext <command> [flags]"

capabilities:
  - custom-commands
  - lifecycle-events
  - metadata

displayName: "My Extension"
tags:
  - category: "Custom"

executablePath:
  windows: ./dist/my-extension-windows-amd64.exe
  linux: ./dist/my-extension-linux-amd64
  darwin: ./dist/my-extension-darwin-amd64
```

> **Default rule:** if you are shipping a real extension with user-facing commands, include `metadata` unless you have a documented compatibility reason not to. Hidden/internal commands (`listen`, `metadata`) must never appear in user-facing README command tables.

### Capabilities Reference

| Capability                   | Purpose                                                                       |
| ---------------------------- | ----------------------------------------------------------------------------- |
| `custom-commands`            | Expose new `azd <namespace> <command>` subcommands                            |
| `lifecycle-events`           | Subscribe to preprovision, postprovision, predeploy, postdeploy, etc.         |
| `mcp-server`                 | Provide MCP tools for AI agents with azd project context                      |
| `service-target-provider`    | Custom deployment targets (replaces built-in container app, function targets) |
| `framework-service-provider` | Custom language/build framework support                                       |
| `metadata`                   | Command tree and configuration schema discovery                               |

## Entry Point (main.go)

```go
package main

import (
    "github.com/azure/azure-dev/cli/azd/pkg/azdext"
    "github.com/org/my-extension/internal/cmd"
)

func main() {
    azdext.Run(cmd.NewRootCommand())
}
```

`azdext.Run` handles gRPC host registration, stdin/stdout communication with the azd host process, and error propagation. Never call `os.Exit` directly â€” return errors and let `azdext.Run` propagate them.

## Root Command (internal/cmd/root.go)

```go
package cmd

import (
    "github.com/azure/azure-dev/cli/azd/pkg/azdext"
    "github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
    rootCmd := &cobra.Command{
        Use:           "azd myext <command> [options]",
        SilenceUsage:  true,
        SilenceErrors: true,
    }

    // listen is REQUIRED when the extension declares lifecycle-events capability.
    // azd calls this subcommand to subscribe the extension to workflow events.
    rootCmd.AddCommand(newListenCommand())
    rootCmd.AddCommand(newQueryCommand())
    return rootCmd
}

// IMPORTANT: Never construct azdext.NewAzdClient() at the root command level.
// The access token context (azdext.WithAccessToken) is only valid inside a RunE
// function, where cmd.Context() is populated by the azd host.
```

## Lifecycle Events (internal/cmd/listen.go)

Subscribe to azd workflow events to run logic at key points in the provision/deploy cycle.

When an extension declares the `lifecycle-events` capability, azd calls the extension's hidden
`listen` subcommand to start event subscription. Prefer the SDK helper `azdext.NewListenCommand`
plus `ExtensionHost` registration methods over directly constructing `NewEventManager`.

```go
package cmd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/spf13/cobra"
)

func newListenCommand() *cobra.Command {
	return azdext.NewListenCommand(func(host *azdext.ExtensionHost) {
		host.
			WithProjectEventHandler("postprovision", handlePostProvision).
			WithProjectEventHandler("predeploy", handlePreDeploy)
	})
}

func handlePostProvision(ctx context.Context, args *azdext.ProjectEventArgs) error {
	projectName := ""
	if args != nil && args.Project != nil {
		projectName = args.Project.Name
	}
	slog.InfoContext(ctx, "post-provision", slog.String("project", projectName))
	return nil
}

func handlePreDeploy(ctx context.Context, args *azdext.ProjectEventArgs) error {
	projectName := ""
	if args != nil && args.Project != nil {
		projectName = args.Project.Name
	}
	slog.InfoContext(ctx, "pre-deploy", slog.String("project", projectName))
	return nil
}
```

> **Important:** `listen` must remain hidden/internal. Do not document `azd <ext> listen` as a user command. Write diagnostics to `os.Stderr` (or use `slog` with a `TextHandler` targeting `os.Stderr`), never to `os.Stdout`. The azd gRPC channel uses stdout â€” writing there corrupts the communication channel.

## Metadata Capability (internal/cmd/metadata.go)

Ship metadata for every production extension command tree unless you have a documented compatibility reason not to. This enables azd help integration, IntelliSense, and schema-aware tooling.

```go
package cmd

import (
	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/spf13/cobra"
)

const (
	extensionID            = "azure.example"
	metadataSchemaVersion = "1.0"
)

func newMetadataCommand() *cobra.Command {
	return azdext.NewMetadataCommand(metadataSchemaVersion, extensionID, NewRootCommand)
}
```

Register it from the root command:

```go
rootCmd.AddCommand(newListenCommand())
rootCmd.AddCommand(newMetadataCommand())
```

If you need richer configuration metadata, you can replace the helper with a custom hidden `metadata` command that calls `azdext.GenerateExtensionMetadata(...)` and then adds configuration schemas before marshaling JSON.

### Available Lifecycle Events

| Event           | Trigger                        |
| --------------- | ------------------------------ |
| `preprovision`  | Before `azd provision` starts  |
| `postprovision` | After `azd provision` succeeds |
| `predeploy`     | Before `azd deploy` starts     |
| `postdeploy`    | After `azd deploy` succeeds    |
| `prepackage`    | Before packaging a service     |
| `postpackage`   | After packaging a service      |
| `predown`       | Before `azd down` starts       |
| `postdown`      | After `azd down` succeeds      |
| `prerestore`    | Before `azd restore` starts    |
| `postrestore`   | After `azd restore` succeeds   |

## Custom Commands

Register custom commands under your extension namespace. Each `RunE` that calls an azd gRPC service
must call `azdext.WithAccessToken(cmd.Context())` and construct its own `AzdClient`. Do NOT
construct `AzdClient` at the root command level â€” the access token is only valid inside `RunE`.

```go
func newQueryCommand() *cobra.Command {
    return &cobra.Command{
        Use:   "query",
        Short: "Query deployed resources",
        RunE: func(cmd *cobra.Command, args []string) error {
            // WithAccessToken is required before any gRPC service call.
            ctx := azdext.WithAccessToken(cmd.Context())

            azdClient, err := azdext.NewAzdClient()
            if err != nil {
                return fmt.Errorf("creating azd client: %w", err)
            }
            defer azdClient.Close()

            resp, err := azdClient.Environment().GetCurrent(ctx, &azdext.EmptyRequest{})
            if err != nil {
                return fmt.Errorf("getting current environment: %w", err)
            }

            slog.Info("current environment", "name", resp.Environment.Name)
            return nil
        },
    }
}
```

Consumers run this as: `azd myext query`

## Environment State API

Persist extension state in the azd environment â€” these values appear in `.azure/<env>/.env` and are
managed exclusively through the gRPC API (never via direct file I/O).

```go
// Read a value (returns empty string if key is absent â€” not an error)
resp, err := azdClient.Environment().GetValue(ctx, &azdext.GetEnvRequest{
    EnvName: currentEnv, // from azdClient.Environment().GetCurrent(...)
    Key:     "DRASI_PROVISIONED",
})
val := resp.Value // "" when not set

// Write a value
_, err = azdClient.Environment().SetValue(ctx, &azdext.SetEnvRequest{
    EnvName: currentEnv,
    Key:     "DRASI_PROVISIONED",
    Value:   "true",
})
```

### Available gRPC Services

All services are accessed via `azdClient.<Service>().<Method>(ctx, &azdext.<Request>{})`. The
`ctx` MUST be a context enriched by `azdext.WithAccessToken`.

| Service         | Key Methods                                              |
| --------------- | -------------------------------------------------------- |
| `Project()`     | `Get` â€” current project config (name, services, infra)   |
| `Environment()` | `GetCurrent`, `List`, `GetValue`, `SetValue`             |
| `Deployment()`  | Deployment status and resource queries                   |
| `Prompt()`      | Prompt the user for input during command execution       |
| `Workflow()`    | Trigger azd workflow steps programmatically              |
| `Copilot()`     | NEW (v1.23.9+) — AI recommendations, diagnostics         | Optional; for AI-assisted workflows            |
| `Extension()`   | Extension metadata and health checks                     | NEW (v1.23.7+)                                 |

### MCP Security Policy (Production Requirement)

If your extension exposes MCP tools via the `mcp-server` capability, you MUST configure and apply a security policy. This is not optional for production extensions.

**Why it matters**: MCP tools run with the azd process's permissions and can access local files, environment variables, and cloud credentials. An unsecured MCP server can leak metadata endpoints, intercept network traffic, or read sensitive files.

Use `azdext.NewMCPSecurityPolicy()` to create a default policy that:

- Blocks access to cloud metadata endpoints (169.254.169.254, imds.azurestack)
- Denies requests to RFC 1918 private networks unless explicitly allowed
- Enforces HTTPS for all external calls
- Redacts sensitive headers (Authorization, X-API-Key, etc.) from logs
- Validates symlinks to prevent directory traversal attacks

```go
// internal/mcp/security.go

package mcp

import (
	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
)

// SecurityPolicy returns a configured MCP security policy for production use.
func SecurityPolicy() *azdext.MCPSecurityPolicy {
	policy := azdext.NewMCPSecurityPolicy()
	
	// Allow internal corporate APIs (optional, if your extension needs them)
	policy.AddAllowedHost("internal-api.corp.com")
	
	// The policy now:
	// - Blocks 169.254.169.254 (cloud metadata)
	// - Blocks RFC 1918 ranges by default (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16)
	// - Requires HTTPS for external calls
	// - Redacts sensitive headers from logs
	
	return policy
}
```

Then apply it when constructing the extension host:

```go
// main.go

func main() {
	// ...
	host := azdext.NewExtensionHost()
	host.WithMCPSecurityPolicy(mcp.SecurityPolicy())
	// Register commands, handlers, tools...
	azdext.Run(host)
}
```

**Example: Custom allow-list for internal APIs**

If your extension legitimately needs to call internal services on private networks:

```go
// Create a custom policy with explicit allow-list
policy := azdext.NewMCPSecurityPolicy()
policy.AddAllowedHost("internal-db.private.cloud")
policy.AddAllowedHost("config-server:8080")

// All other hosts still respect the default restrictions
```

**Validation**: Run `azd extension validate <binary>` (v1.24.0+) to check that your extension declares the security policy correctly.

## MCP Server Capability

Declare `mcp-server` in extension.yaml to expose AI-accessible tools:

```yaml
capabilities:
  - name: mcp-server
    config:
      tools:
        - name: get_deployment_status
          description: "Returns the current deployment status for all services"
        - name: list_environments
          description: "Lists all azd environments in the current project"
```

The `azdext` SDK provides an MCP-compatible server host. The extension binary handles both CLI commands and MCP tool calls through the same entry point â€” azd routes requests to the appropriate handler at runtime.

## Build Scripts

### build.sh (Linux/macOS)

```bash
#!/usr/bin/env bash
set -euo pipefail

VERSION=$(cat version.txt)
LDFLAGS="-X main.version=${VERSION} -s -w"

mkdir -p dist

echo "Building linux/amd64..."
GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o dist/my-extension-linux-amd64 .

echo "Building darwin/amd64..."
GOOS=darwin GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o dist/my-extension-darwin-amd64 .

echo "Building darwin/arm64..."
GOOS=darwin GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o dist/my-extension-darwin-arm64 .

echo "Done. Artifacts in dist/"
```

### build.ps1 (Windows)

```powershell
$ErrorActionPreference = 'Stop'

$version = Get-Content version.txt
$ldflags = "-X main.version=$version -s -w"

New-Item -ItemType Directory -Force -Path dist | Out-Null

Write-Host "Building windows/amd64..."
$env:GOOS = 'windows'; $env:GOARCH = 'amd64'
go build -ldflags $ldflags -o dist/my-extension-windows-amd64.exe .

Write-Host "Building linux/amd64..."
$env:GOOS = 'linux'; $env:GOARCH = 'amd64'
go build -ldflags $ldflags -o dist/my-extension-linux-amd64 .

Write-Host "Done. Artifacts in dist/"
```

## Distribution

### Dev Registry (Unsigned â€” Local and CI)

```bash
# Install from the dev registry (unsigned extensions are allowed)
azd extension add my-extension --source dev

# Configure a custom file-based source
azd config set extension.sources[0].key custom
azd config set extension.sources[0].path ./dist/extension.yaml
azd extension add my-extension --source custom
```

### Official Registry

Submit a PR to the azd extension registry repository. Requirements:

- Binary signed with a code-signing certificate
- Passes automated security scanning
- Public source repository with open-source license

### Consumer Experience

```bash
# Discover
azd extension list
azd extension list --source dev

# Install
azd extension add my-extension
azd extension add github.com/org/my-extension@1.0.0

# Use
azd myext query
azd myext listen

# Remove
azd extension remove my-extension
```

## Go Standards Checklist

These rules apply to every azd extension written in Go and align with `go.instructions.md` and `go-tests.instructions.md`.

### Structured Logging â€” `log/slog` (Go 1.21+)

Always use `log/slog` for operational output. `fmt.Printf` to stdout corrupts the azd gRPC channel; `fmt.Fprintf(os.Stderr, ...)` is acceptable only for one-line debugging.

```go
// main.go â€” initialise before azdext.Run
func main() {
    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
    slog.SetDefault(logger)

    if err := run(); err != nil {
        slog.Error("extension failed", "error", err)
        os.Exit(1)
    }
}

// In handlers
slog.Info("drasi source installed", "name", sourceName)
slog.Error("could not connect to Key Vault", "error", err)
```

### Sentinel Error Naming

```go
// CORRECT â€” exported sentinel error as package-level var
var ErrSourceNotFound = errors.New("drasi source not found")
var ErrAuthFailed     = errors.New("authentication failed")

// WRONG â€” string constants (violates Go conventions, can't be compared via errors.Is)
// const ERR_SOURCE_NOT_FOUND = "drasi source not found"
```

Use `errors.Is(err, ErrXxx)` for equality checks in callers; never compare error strings.

### Test Requirements

Every test file MUST call `t.Parallel()` as the first statement in each `TestXxx` function and in each table-driven sub-test.

```go
func TestInstallSource(t *testing.T) {
    t.Parallel()

    cases := []struct {
        name    string
        input   string
        wantErr error
    }{
        {"valid source", "drasi-default", nil},
        {"empty name", "", ErrSourceNotFound},
    }

    for _, tc := range cases {
        tc := tc // loop capture â€” required for Go < 1.22
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            // ...
        })
    }
}
```

Tests that require a real azd process or Key Vault connection must use a build tag:

```go
//go:build integration

package cmd_test
```

Run unit tests with `go test ./...`; integration tests with `go test -tags=integration ./...`.

### Known Azure SDK Gotchas

#### `azsecrets` Import Path

The `sdk/keyvault/azsecrets` module path is **deprecated and will not receive updates**. Always use:

```go
// CORRECT
import "github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"

// WRONG â€” deprecated, may return incorrect NewClient arity
// import "github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
```

> **[VERIFY]** Before using `azsecrets.NewClient`: check the current signature at
> `https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets`
> â€” the README and MIGRATION.md have historically disagreed on the return arity.
> The constructor currently returns `(*Client, error)`; confirm this matches the
> version in your `go.sum` before writing call sites.

### Bicep / IaC (when extension provisions Azure resources)

- Check AVM module index **first**: `https://github.com/Azure/bicep-registry-modules/tree/main/avm/res`
- Prefer `br/public:avm/res/<provider>/<resource>` over authoring native Bicep modules; document any fallback with a justification comment.
- Every resource declaration needs a stable, non-preview API version. Verify at `learn.microsoft.com/azure/templates/` before committing.

---

## Advanced Patterns for Command-Heavy Extensions

These patterns apply when building extensions with multiple commands that call gRPC services and need to be testable without a live azd host.

### Consumer-Side Interface for `azdext.AzdClient`

`azdext.AzdClient` is a **concrete struct**, not an interface. Packages that need to mock it for testing must define their own narrow interface:

```go
// internal/deployment/state.go

// envStateClient is a consumer-side interface for the subset of AzdClient used here.
// Define it in the consuming package, not in the azdext package.
type envStateClient interface {
    Environment() environmentServiceClient
}

type environmentServiceClient interface {
    GetValue(ctx context.Context, req *azdext.GetEnvRequest) (*azdext.KeyValueResponse, error)
    SetValue(ctx context.Context, req *azdext.SetEnvRequest) (*azdext.EmptyResponse, error)
}
```

This follows the Go consumer-side interface pattern: define interfaces where they're used, not where they're implemented.

### White-Box Injection Pattern (Package-Level Function Vars)

Commands that call gRPC or external systems use package-level function variables for injection. This enables testing without spawning a live azd host or external process:

```go
// cmd/deploy.go

// runDeployFunc is overridden in tests to avoid live gRPC calls.
var runDeployFunc = defaultRunDeploy

func newDeployCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:  "deploy",
        RunE: func(cmd *cobra.Command, args []string) error {
            envName, _ := cmd.Root().PersistentFlags().GetString("environment")
            dryRun, _ := cmd.Flags().GetBool("dry-run")
            return runDeployFunc(cmd, envName, dryRun)
        },
    }
    cmd.Flags().Bool("dry-run", false, "Simulate deploy without applying changes")
    return cmd
}

func defaultRunDeploy(cmd *cobra.Command, envName string, dryRun bool) error {
    ctx := azdext.WithAccessToken(cmd.Context())
    // real implementation
    return nil
}
```

### White-Box Test Files

Tests that need to override package-level function vars must live in the **same package** (not `_test`). Use a separate `_internal_test.go` file to keep them clearly separated:

```
cmd/
â”œâ”€â”€ deploy.go                    # package cmd
â”œâ”€â”€ deploy_test.go               # package cmd_test  â€” black-box tests
â””â”€â”€ deploy_internal_test.go      # package cmd       â€” white-box tests (override runDeployFunc)
```

```go
// cmd/deploy_internal_test.go
package cmd // NOT package cmd_test

import (
    "testing"
    "github.com/spf13/cobra"
)

func TestDeployDryRun(t *testing.T) {
    // No t.Parallel() â€” this test mutates runDeployFunc
    called := false
    runDeployFunc = func(cmd *cobra.Command, envName string, dryRun bool) error {
        called = true
        return nil
    }
    t.Cleanup(func() { runDeployFunc = defaultRunDeploy })

    // ... execute command and assert
}
```

### `t.Parallel()` Restriction

Tests that mutate **package-level variables** (injection vars, global state) MUST NOT call `t.Parallel()`. Parallel execution with shared mutable state causes race conditions.

```go
// WRONG â€” races on runDeployFunc
func TestDeployMutation(t *testing.T) {
    t.Parallel()           // FORBIDDEN when mutating package-level vars
    runDeployFunc = ...
}

// CORRECT â€” sequential mutation tests
func TestDeployMutation(t *testing.T) {
    // no t.Parallel()
    runDeployFunc = ...
    t.Cleanup(func() { runDeployFunc = defaultRunDeploy })
}
```

Standard pure-logic tests (no package-level mutation) should still call `t.Parallel()` as the first statement.

### Reading Persistent Root Flags

Root-level persistent flags (e.g., `--environment`, `--output`) are not on the subcommand itself. Read them via the root:

```go
// CORRECT â€” read from root
envName, _ := cmd.Root().PersistentFlags().GetString("environment")
output, _ := cmd.Root().PersistentFlags().GetString("output")

// WRONG â€” will return zero value silently
envName, _ := cmd.Flags().GetString("environment")
```

### `SilenceErrors: true` on Root Command

Setting `SilenceErrors: true` on the root command prevents Cobra from printing errors to stderr automatically. Your command's `RunE` must print errors explicitly before returning them:

```go
// cmd/root.go
rootCmd := &cobra.Command{
    SilenceUsage:  true,
    SilenceErrors: true, // Extension controls all error output to stderr
}
```

```go
// cmd/errors.go â€” write structured errors to stderr before returning
func writeCommandError(cmd *cobra.Command, code string, message string) error {
    ce := &commandError{code: code, message: fmt.Sprintf("%s: %s", code, message)}
    fmt.Fprintln(os.Stderr, ce.Error())
    return ce
}
```

The `commandError.Error()` string embeds the error code, so callers can check with `strings.Contains(err.Error(), "ERR_NO_AUTH")`.

### `commandError` vs `notImplemented`

Two distinct error patterns serve different purposes:

| Function | When to Use | Output |
|----------|-------------|--------|
| `writeCommandError(cmd, code, msg)` | Business/validation errors with a known error code | Writes to stderr + returns `commandError` |
| `notImplemented(name string)` | Stub commands not yet implemented | Returns `fmt.Errorf(...)` wrapping `errNotYetImplemented` â€” does NOT write to stderr |

```go
// Use writeCommandError for real command errors
if envName == "" {
    return writeCommandError(cmd, output.ERR_NO_AUTH, "environment name is required")
}

// Use notImplemented for stubs
func runTeardown(cmd *cobra.Command, args []string) error {
    return notImplemented("teardown")
}
```

### `//go:embed all:templates`

When embedding a directory that may contain dotfiles (`.gitignore`, `.env.example`, etc.), the `all:` prefix is mandatory. Without it, Go's embed silently omits files and directories whose names begin with `.` or `_`:

```go
//go:embed all:templates   // CORRECT â€” includes dotfiles
var templateFS embed.FS

//go:embed templates       // WRONG â€” silently omits .gitignore, .env.example, etc.
var templateFS embed.FS
```

---

## Anti-Patterns

| Anti-Pattern                                    | Problem                                                 | Fix                                                     |
| ----------------------------------------------- | ------------------------------------------------------- | ------------------------------------------------------- |
| Using `context.Background()` in event handlers  | Loses azd cancellation and deadline                     | Accept `ctx` from the event handler signature           |
| Writing to `os.Stdout` in lifecycle handlers    | Corrupts azd gRPC communication channel                 | Write diagnostics to `os.Stderr` only                   |
| Embedding secrets in the extension binary       | Security risk â€” binary is distributable                 | Read from azd environment via `azdClient.Environment()` |
| Hardcoding a single OS path in `executablePath` | Only works on one platform                              | Provide `windows`, `linux`, and `darwin` entries        |
| Calling `panic` on azd client errors            | Crashes the azd host process                            | Return errors; `azdext.Run` propagates them cleanly     |
| Using `azd hooks` for complex multi-step logic  | Hook scripts lack structured context and error handling | Use a full extension with lifecycle-events instead      |
| Using `GetEnvironmentValue`/`SetEnvironmentValue` | These method names do not exist in `azdext`           | Use `GetValue`/`SetValue` on `azdClient.Environment()`  |
| Constructing `azdext.NewAzdClient()` at root command level | Access token context not yet populated        | Construct `AzdClient` inside each `RunE` function       |
| Defining `azdext.AzdClient` interface in a shared package | Couples test infrastructure to SDK internals  | Define narrow consumer-side interfaces in each consuming package |
| Calling `t.Parallel()` in tests that mutate package-level vars | Race condition on injection vars          | Omit `t.Parallel()` for any test that writes to a package-level var |
| Using `cmd.Flags().GetString("environment")` for persistent flags | Returns zero value â€” flag is on root   | Use `cmd.Root().PersistentFlags().GetString("environment")` |
| Embedding templates without `all:` prefix      | Silently omits dotfiles from embedded FS                | Use `//go:embed all:templates`                          |

## Troubleshooting

### Extension Fails to Start

**Symptoms**: `azd <ext> <command>` errors with "extension failed to initialize" or similar.

**Diagnosis steps**:

1. Check extension binary exists and is executable:
   ```bash
   ls -la ./my-extension-linux-amd64
   file ./my-extension-linux-amd64
   ```

2. Verify extension binary can run standalone:
   ```bash
   ./my-extension-linux-amd64 --help
   ./my-extension-linux-amd64 version
   ```

3. Check for dependency issues (missing libraries, wrong architecture):
   ```bash
   ldd ./my-extension-linux-amd64  # on Linux
   otool -L ./my-extension-linux-amd64  # on macOS
   ```

4. Review azd diagnostics:
   ```bash
   azd version  # Confirm azd >= 1.10.0
   azd extension list  # Check if extension appears
   ```

**Common root causes**:

| Symptom | Cause | Fix |
|---------|-------|-----|
| `binary exec failed` | Extension binary is not executable or missing | Check file permissions: `chmod +x my-extension-*` |
| `gRPC bind failed` | Port conflict or permission denied | Verify no other process uses the same gRPC port; run without sudo unless needed |
| `metadata parse error` | Metadata command returns invalid JSON | Validate output: `./my-extension metadata | jq .` |
| `handler init failed` | Lifecycle handler panicked or returned early | Check slog output; add debug logging to handler functions |
| `ErrSourceNotFound` or similar | Extension binary running but command not found | Verify root command registers your custom commands in `main()` |

### Lifecycle Handlers Fail Silently

**Symptoms**: Event handlers (postprovision, predeploy, etc.) register but do not execute.

**Diagnosis**:

1. Verify handler is registered:
   ```go
   host.WithProjectEventHandler("postprovision", handlePostProvision)
   ```

2. Check handler signature matches `azdext.ProjectEventHandler`:
   ```go
   func handlePostProvision(ctx context.Context, args *azdext.ProjectEventArgs) error
   ```

3. Verify event is actually firing by adding slog output:
   ```go
   slog.InfoContext(ctx, "postprovision handler called", "project", args.Project.Name)
   ```

4. Run azd with verbose logging:
   ```bash
   azd provision -v  # Enables verbose azd logging
   ```

### MCP Tools Not Visible to Agents

**Symptoms**: Agents cannot find or invoke MCP tools declared in extension.yaml.

**Diagnosis**:

1. Verify `mcp-server` capability is declared:
   ```yaml
   capabilities:
     - name: mcp-server
       config:
         tools:
           - name: my_tool
             description: "..."
   ```

2. Check MCP security policy is applied (if MCP tools are declared):
   ```go
   host.WithMCPSecurityPolicy(mcp.SecurityPolicy())
   ```

3. Validate tool implementation exists in your MCP handler.

4. Run `azd extension list <name>` to see declared capabilities.

### Flag Precedence Issues

**Problem**: Root flags like `--environment` are not visible in subcommand handlers.

**Solution**: Access persistent flags from the root:

```go
func (cmd *MyCommand) RunE(cobraCmd *cobra.Command, args []string) error {
	// WRONG — will return zero value
	// env := cobraCmd.Flags().GetString("environment")
	
	// CORRECT — environment is a persistent flag on root
	env, _ := cobraCmd.Root().PersistentFlags().GetString("environment")
	
	return nil
}
```

## Scope Boundaries

**USE FOR:** creating and building azd extensions, lifecycle hooks, custom CLI commands, MCP server capabilities, extension.yaml manifests, cross-platform build scripts, registry distribution.

**DO NOT USE FOR:**

- Consuming existing extensions in an azd project â†’ use `azd-deployment` skill
- Writing `hooks:` entries in `azure.yaml` â†’ use `azd-deployment` skill (simpler script-based approach)
- General Go development patterns â†’ see `go.instructions.md`
- Bicep or infrastructure authoring â†’ see `azure-verified-modules` or `azure-deployment-preflight`

## Currency

This skill targets the azd extensions beta framework introduced in azd 1.10.0.

- azd minimum version: **1.10.0**
- SDK: `github.com/azure/azure-dev/cli/azd/pkg/azdext` (beta â€” verify compatibility with `azd version`)
- Reference extensions: `microsoft.azd.demo`, `azure.ai.agents`
- Last reviewed: **2026-04-06**

## References

- [azd Extensibility Overview](https://learn.microsoft.com/azure/developer/azure-developer-cli/azd-extensibility)
- [azure-dev GitHub Repository](https://github.com/Azure/azure-dev)
- [microsoft.azd.demo source](https://github.com/Azure/azure-dev/tree/main/extensions/microsoft.azd.demo)
- [azure.ai.agents extension source](https://github.com/Azure/azure-dev/tree/main/extensions/azure.ai.agents)
- [Extension Schema (extension.schema.json)](https://github.com/Azure/azure-dev/blob/main/schemas/v1.0/extension.schema.json)
- [Official Extension Registry](https://aka.ms/azd/extensions/registry)
- [Dev Extension Registry](https://aka.ms/azd/extensions/registry/dev)

## Future Enhancements (Planned, Not Yet Shipping)

These patterns are in development and are not yet available in stable releases. Monitor zd version and the [March 2026 release notes](https://devblogs.microsoft.com/azure-sdk/azure-developer-cli-azd-march-2026/) for availability.

### Copilot Service Integration (v1.24.0+, planned)

Extensions can invoke Copilot to generate recommendations, diagnose issues, or suggest configurations.

`go
// Future API â€” not available yet
func handlePostProvision(ctx context.Context, args *azdext.ProjectEventArgs) error {
    azdClient, _ := azdext.NewAzdClient()
    resp, err := azdClient.Copilot().Recommend(ctx, &azdext.CopilotRecommendationRequest{
        Topic: "aks-security",
        Context: fmt.Sprintf("Project: %s, Services: %d", args.Project.Name, len(args.Project.Services)),
    })
    if err == nil {
        slog.InfoContext(ctx, "copilot recommendation", "suggestion", resp.Recommendation)
    }
    return nil
}
`

### Remote Build Fallback (v1.24.0+, planned)

Extensions can offload build steps to Azure Container Registry (ACR) for cross-arch builds without requiring Docker on the local machine.

### Deployment Timeout Configuration (v1.23.13+, optional use)

Set extension-specific timeouts in xtension.yaml:

`yaml
deployTimeout: 1800 # seconds â€” useful for slow Drasi runtime initialization
`

### Extension Preflight Validation (v1.24.0+, planned)

zd extension validate <extension-binary> will check:
- Binary is valid and executable
- MCP security policy is configured (if MCP tools are declared)
- Metadata command is valid
- Capability declarations match binary capabilities

---



