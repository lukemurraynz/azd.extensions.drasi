---
name: creating-azd-extensions
description: >-
  Author, build, and publish Azure Developer CLI (azd) extensions in Go. USE FOR: creating new azd extensions, implementing lifecycle hooks, exposing custom CLI commands, building MCP server capabilities, writing extension.yaml manifests, cross-platform build scripts, distributing via registry or local sources.
version: 1.2.0
lastUpdated: 2026-04-05
---

# Creating Azure Developer CLI Extensions

Build Go binaries that extend `azd` with custom commands, lifecycle hooks, and MCP tools. Extensions communicate with the azd host via gRPC and are distributed through registries.

> ⚠️ **Beta feature**: The azd extensions framework is in beta. APIs may change between azd releases. Verify your target azd version supports the capabilities you need: `azd version` (requires 1.10.0+).

## What's New in v1.2.0 (2026-04-05)

- ✅ **Correct lifecycle API** — `azdext.NewEventManager` (replaces incorrect `NewExtensionHost`)
- ✅ **`azdext.WithAccessToken(ctx)` requirement** — must be called before any gRPC service call
- ✅ **Environment state API** — `GetEnvironmentValue`/`SetEnvironmentValue` gRPC patterns
- ✅ **Available gRPC services** reference table
- ✅ **`listen` subcommand** — required convention when `lifecycle-events` capability is declared

## What's New in v1.1.0 (2026-04-07)

- ✅ **Known SDK gotchas** — correct `azsecrets` import path, `log/slog` requirement
- ✅ **Go standards checklist** — error naming, `t.Parallel()`, table-driven tests, structured logging
- ✅ **Bicep / IaC guidance** — AVM-first mandate, API version verification

## What's New in v1.0.0 (2026-03-28)

- ✅ **Initial skill** covering azd extension authoring from scratch
- ✅ **Full extension.yaml schema** with all capability declarations
- ✅ **Lifecycle event patterns** reverse-engineered from `microsoft.azd.demo` and `azure.ai.agents` extension sources
- ✅ **Cross-platform build scripts** (build.ps1 + build.sh)
- ✅ **Registry distribution** patterns (dev + official + custom sources)

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
├── extension.yaml              # Manifest — required
├── version.txt                 # Semantic version string — required
├── main.go                     # Entry point: azdext.Run(cmd.NewRootCommand())
├── internal/
│   └── cmd/
│       ├── root.go             # Cobra command tree
│       └── listen.go           # Lifecycle event subscriptions
├── build.ps1                   # Windows cross-compile build script
├── build.sh                    # Linux/macOS cross-compile build script
└── go.mod / go.sum
```

## extension.yaml Manifest

```yaml
name: my-extension
namespace: myext # CLI prefix: azd myext <command>
version: 1.0.0
description: "Short description of what this extension does."
usage: "azd myext <command> [flags]"

capabilities:
  - name: custom-commands # Expose CLI commands under namespace
  - name: lifecycle-events # Subscribe to azd workflow events
  - name: metadata # Command tree + config schema discovery

displayName: "My Extension"
tags:
  - category: "Custom"

executablePath:
  windows: ./dist/my-extension-windows-amd64.exe
  linux: ./dist/my-extension-linux-amd64
  darwin: ./dist/my-extension-darwin-amd64
```

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

`azdext.Run` handles gRPC host registration, stdin/stdout communication with the azd host process, and error propagation. Never call `os.Exit` directly — return errors and let `azdext.Run` propagate them.

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

When an extension declares the `lifecycle-events` capability, azd calls the extension's
`listen` subcommand to start event subscription. The command **blocks** until azd closes the
connection — `eventManager.Receive(ctx)` is the blocking call.

```go
package cmd

import (
    "context"
    "fmt"

    "github.com/azure/azure-dev/cli/azd/pkg/azdext"
    "github.com/spf13/cobra"
)

func newListenCommand() *cobra.Command {
    return &cobra.Command{
        Use:   "listen",
        Short: "Subscribe to azd lifecycle events (invoked by azd host)",
        RunE:  runListen,
    }
}

func runListen(cmd *cobra.Command, args []string) error {
    // WithAccessToken enriches the context with the azd gRPC access token.
    // MUST be called before constructing AzdClient or EventManager.
    ctx := azdext.WithAccessToken(cmd.Context())

    azdClient, err := azdext.NewAzdClient()
    if err != nil {
        return fmt.Errorf("creating azd client: %w", err)
    }
    defer azdClient.Close()

    eventManager := azdext.NewEventManager(azdClient)
    defer eventManager.Close()

    if err := eventManager.AddProjectEventHandler(
        ctx, "postprovision", handlePostProvision,
    ); err != nil {
        return fmt.Errorf("subscribing to postprovision: %w", err)
    }

    if err := eventManager.AddProjectEventHandler(
        ctx, "predeploy", handlePreDeploy,
    ); err != nil {
        return fmt.Errorf("subscribing to predeploy: %w", err)
    }

    // Optional: filter service events by host type
    if err := eventManager.AddServiceEventHandler(
        ctx, "prepackage", handlePrePackage,
        &azdext.ServerEventOptions{Host: "aks"},
    ); err != nil {
        return fmt.Errorf("subscribing to prepackage: %w", err)
    }

    // Receive blocks until the azd host closes the connection.
    return eventManager.Receive(ctx)
}

func handlePostProvision(ctx context.Context, args *azdext.ProjectEventArgs) error {
    slog.Info("post-provision", "env", args.Environment.Name)
    return nil
}

func handlePreDeploy(ctx context.Context, args *azdext.ProjectEventArgs) error {
    slog.Info("pre-deploy", "project", args.Project.Name)
    return nil
}

func handlePrePackage(ctx context.Context, args *azdext.ServiceEventArgs) error {
    slog.Info("pre-package", "service", args.Service.Name)
    return nil
}
```

> **Important:** Write diagnostics to `os.Stderr` (or use `slog` with a `TextHandler` targeting `os.Stderr`), never to `os.Stdout`. The azd gRPC channel uses stdout — writing there corrupts the communication channel.

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
construct `AzdClient` at the root command level — the access token is only valid inside `RunE`.

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

Persist extension state in the azd environment — these values appear in `.azure/<env>/.env` and are
managed exclusively through the gRPC API (never via direct file I/O).

```go
// Read a value (returns empty string if key is absent — not an error)
resp, err := azdClient.Environment().GetEnvironmentValue(ctx, &azdext.GetEnvRequest{
    EnvName: currentEnv, // from azdClient.Environment().GetCurrent(...)
    Key:     "DRASI_PROVISIONED",
})
val := resp.Value // "" when not set

// Write a value
_, err = azdClient.Environment().SetEnvironmentValue(ctx, &azdext.SetEnvRequest{
    EnvName: currentEnv,
    Key:     "DRASI_PROVISIONED",
    Value:   "true",
})
```

### Available gRPC Services

All services are accessed via `azdClient.<Service>().<Method>(ctx, &azdext.<Request>{})`. The
`ctx` MUST be a context enriched by `azdext.WithAccessToken`.

| Service         | Key Methods                                                        |
| --------------- | ------------------------------------------------------------------ |
| `Project()`     | `Get` — current project config (name, services, infra)             |
| `Environment()` | `GetCurrent`, `List`, `GetEnvironmentValue`, `SetEnvironmentValue` |
| `Deployment()`  | Deployment status and resource queries                             |
| `Prompt()`      | Prompt the user for input during command execution                 |
| `Workflow()`    | Trigger azd workflow steps programmatically                        |

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

The `azdext` SDK provides an MCP-compatible server host. The extension binary handles both CLI commands and MCP tool calls through the same entry point — azd routes requests to the appropriate handler at runtime.

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

### Dev Registry (Unsigned — Local and CI)

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

### Structured Logging — `log/slog` (Go 1.21+)

Always use `log/slog` for operational output. `fmt.Printf` to stdout corrupts the azd gRPC channel; `fmt.Fprintf(os.Stderr, ...)` is acceptable only for one-line debugging.

```go
// main.go — initialise before azdext.Run
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
// CORRECT — exported sentinel error as package-level var
var ErrSourceNotFound = errors.New("drasi source not found")
var ErrAuthFailed     = errors.New("authentication failed")

// WRONG — string constants (violates Go conventions, can't be compared via errors.Is)
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
        tc := tc // loop capture — required for Go < 1.22
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

// WRONG — deprecated, may return incorrect NewClient arity
// import "github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
```

> **[VERIFY]** Before using `azsecrets.NewClient`: check the current signature at
> `https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets`
> — the README and MIGRATION.md have historically disagreed on the return arity.
> The constructor currently returns `(*Client, error)`; confirm this matches the
> version in your `go.sum` before writing call sites.

### Bicep / IaC (when extension provisions Azure resources)

- Check AVM module index **first**: `https://github.com/Azure/bicep-registry-modules/tree/main/avm/res`
- Prefer `br/public:avm/res/<provider>/<resource>` over authoring native Bicep modules; document any fallback with a justification comment.
- Every resource declaration needs a stable, non-preview API version. Verify at `learn.microsoft.com/azure/templates/` before committing.

---

## Anti-Patterns

| Anti-Pattern                                    | Problem                                                 | Fix                                                     |
| ----------------------------------------------- | ------------------------------------------------------- | ------------------------------------------------------- |
| Using `context.Background()` in event handlers  | Loses azd cancellation and deadline                     | Accept `ctx` from the event handler signature           |
| Writing to `os.Stdout` in lifecycle handlers    | Corrupts azd gRPC communication channel                 | Write diagnostics to `os.Stderr` only                   |
| Embedding secrets in the extension binary       | Security risk — binary is distributable                 | Read from azd environment via `azdClient.Environment()` |
| Hardcoding a single OS path in `executablePath` | Only works on one platform                              | Provide `windows`, `linux`, and `darwin` entries        |
| Calling `panic` on azd client errors            | Crashes the azd host process                            | Return errors; `azdext.Run` propagates them cleanly     |
| Using `azd hooks` for complex multi-step logic  | Hook scripts lack structured context and error handling | Use a full extension with lifecycle-events instead      |

## Scope Boundaries

**USE FOR:** creating and building azd extensions, lifecycle hooks, custom CLI commands, MCP server capabilities, extension.yaml manifests, cross-platform build scripts, registry distribution.

**DO NOT USE FOR:**

- Consuming existing extensions in an azd project → use `azd-deployment` skill
- Writing `hooks:` entries in `azure.yaml` → use `azd-deployment` skill (simpler script-based approach)
- General Go development patterns → see `go.instructions.md`
- Bicep or infrastructure authoring → see `azure-verified-modules` or `azure-deployment-preflight`

## Currency

This skill targets the azd extensions beta framework introduced in azd 1.10.0.

- azd minimum version: **1.10.0**
- SDK: `github.com/azure/azure-dev/cli/azd/pkg/azdext` (beta — verify compatibility with `azd version`)
- Reference extensions: `microsoft.azd.demo`, `azure.ai.agents`
- Last reviewed: **2026-03-28**

## References

- [azd Extensibility Overview](https://learn.microsoft.com/azure/developer/azure-developer-cli/azd-extensibility)
- [azure-dev GitHub Repository](https://github.com/Azure/azure-dev)
- [microsoft.azd.demo source](https://github.com/Azure/azure-dev/tree/main/extensions/microsoft.azd.demo)
- [azure.ai.agents extension source](https://github.com/Azure/azure-dev/tree/main/extensions/azure.ai.agents)
- [Extension Schema (extension.schema.json)](https://github.com/Azure/azure-dev/blob/main/schemas/v1.0/extension.schema.json)
- [Official Extension Registry](https://aka.ms/azd/extensions/registry)
- [Dev Extension Registry](https://aka.ms/azd/extensions/registry/dev)
