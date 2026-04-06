---
name: creating-azd-extensions
description: >-
  Author, build, and publish Azure Developer CLI (azd) extensions in Go. USE FOR: creating new azd extensions, implementing lifecycle hooks, exposing custom CLI commands, writing extension.yaml manifests, adding metadata/version commands, hardening cross-platform build scripts, and distributing via registry or local sources.
version: 1.8.0
lastUpdated: 2026-04-07
---

# Creating Azure Developer CLI Extensions

Build Go binaries that extend `azd` with custom commands and lifecycle hooks over the `azdext` gRPC SDK.

> **Beta feature:** azd extensions remain beta. Match your implementation to the azd version you actually target, and prefer pinned SDK behavior over memory or older examples.

## Current default rules

- Ship **user-facing commands** plus hidden integration commands only when required by capabilities.
- If you declare `lifecycle-events`, add a hidden `listen` command.
- If you ship a real command surface, add `metadata` and a hidden `metadata` command.
- Expose a visible `version` command.
- Never document hidden commands like `listen` or `metadata` as public UX.
- Never write operational output to stdout from gRPC-driven integration paths.

## Recommended file structure

```text
my-extension/
├── extension.yaml
├── version.txt
├── main.go
├── cmd/
│   ├── root.go
│   ├── listen.go
│   ├── metadata.go
│   └── version.go
├── build.ps1
├── build.sh
└── go.mod
```

## extension.yaml manifest

```yaml
id: azure.example
namespace: example
displayName: "Example azd Extension"
description: "Short description of what this extension does."
usage: azd example <command> [options]
version: "1.0.0"
requiredAzdVersion: ">= 1.10.0"
language: go

capabilities:
  - custom-commands
  - lifecycle-events
  - metadata
```

### Manifest guidance

- `metadata` should be the default for production extensions.
- Add `mcp-server`, `service-target-provider`, or `framework-service-provider` only for a concrete use case.
- Keep examples aligned with the real user-facing command surface.
- Do not include hidden integration commands in README command tables.

## Entry point (`main.go`)

```go
package main

import (
	"log/slog"
	"os"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/org/my-extension/cmd"
)

var version = "dev"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	cmd.SetVersion(version)
	azdext.Run(cmd.NewRootCommand())
}
```

## Root command

```go
package cmd

import "github.com/spf13/cobra"

const (
	extensionID           = "azure.example"
	metadataSchemaVersion = "1.0"
)

var extensionVersion = "dev"

func SetVersion(version string) {
	extensionVersion = version
}

func NewRootCommand() *cobra.Command {
	var outputFormat string

	rootCmd := &cobra.Command{
		Use:           "azd example <command> [options]",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().StringVar(&outputFormat, "output", "table", "Output format: table or json")

	rootCmd.AddCommand(newListenCommand())
	rootCmd.AddCommand(newMetadataCommand())
	rootCmd.AddCommand(newVersionCommand(&outputFormat))
	rootCmd.AddCommand(newQueryCommand())

	return rootCmd
}
```

## Lifecycle events

Prefer SDK helpers over manually wiring event managers.

```go
package cmd

import (
	"context"
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

### Lifecycle rules

- `listen` is hidden and host-invoked.
- Do not tell users to run `azd <ext> listen`.
- Use `WithProjectEventHandler(...)` / `WithServiceEventHandler(...)` on `ExtensionHost`.
- `azdext.NewListenCommand(...)` is the preferred helper for the hidden listener command.

## Metadata command

Prefer generated metadata from the live Cobra tree.

```go
package cmd

import (
	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/spf13/cobra"
)

func newMetadataCommand() *cobra.Command {
	return azdext.NewMetadataCommand(metadataSchemaVersion, extensionID, NewRootCommand)
}
```

### Metadata guidance

- Hidden `metadata` command is the default guidance.
- If you need richer configuration metadata, build a custom hidden command around `azdext.GenerateExtensionMetadata(...)`.
- Do not treat static `metadata.json` packaging as the default recommendation unless you have separately verified that path for your target azd version.

## Version command

Expose a visible version command for supportability and parity with official extensions.

```go
package cmd

import (
	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/spf13/cobra"
)

func newVersionCommand(outputFormat *string) *cobra.Command {
	return azdext.NewVersionCommand(extensionID, extensionVersion, outputFormat)
}
```

## Custom commands

Every `RunE` that calls azd gRPC services must construct its own client inside the command and use the access-token-enriched context.

```go
func newQueryCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "query",
		Short: "Query deployed resources",
		RunE: func(cmd *cobra.Command, args []string) error {
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

## Environment state API

Persist extension state through azd environment gRPC APIs, not direct file writes.

```go
resp, err := azdClient.Environment().GetValue(ctx, &azdext.GetEnvRequest{
	EnvName: currentEnv,
	Key:     "EXAMPLE_STATE",
})

_, err = azdClient.Environment().SetValue(ctx, &azdext.SetEnvRequest{
	EnvName: currentEnv,
	Key:     "EXAMPLE_STATE",
	Value:   "true",
})
```

## Build scripts

Build scripts should do four things:

1. verify Go is available
2. print `go version` and `go env`
3. run `go test ./...` and `go build ./...`
4. build release artifacts and verify they exist

### `build.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail

if ! command -v go >/dev/null 2>&1; then
  echo "go not found on PATH" >&2
  exit 1
fi

go version
go env GOPATH GOROOT

VERSION=$(tr -d '[:space:]' < version.txt)
LDFLAGS="-X main.version=${VERSION} -s -w"

go test ./...
go build ./...

mkdir -p dist
GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o dist/my-extension-linux-amd64 .
test -f dist/my-extension-linux-amd64
```

### `build.ps1`

```powershell
$ErrorActionPreference = 'Stop'

Get-Command go | Out-Null
go version
go env GOPATH GOROOT

$version = (Get-Content version.txt).Trim()
$ldflags = "-X main.version=$version -s -w"

go test ./...
go build ./...

New-Item -ItemType Directory -Force -Path dist | Out-Null
go build -ldflags $ldflags -o dist/my-extension-windows-amd64.exe .

if (-not (Test-Path 'dist/my-extension-windows-amd64.exe')) {
    throw 'Missing windows artifact'
}
```

## Distribution and consumer experience

```bash
# Discover
azd extension list
azd extension list --source dev

# Install
azd extension add my-extension
azd extension add github.com/org/my-extension@1.0.0

# Use
azd myext query
azd myext version

# Remove
azd extension remove my-extension
```

## Avoid these mistakes

- documenting `listen` or `metadata` as user-facing commands
- constructing `AzdClient` at root-command creation time
- writing stdout logs from gRPC-driven integration paths
- telling users to add MCP/service-target/framework-service capabilities without a real use case
- relying on stale examples instead of the pinned SDK you actually build against

## Validation checklist

- `go test ./...`
- `go build ./...`
- run build scripts successfully
- metadata capability declared when appropriate
- hidden integration commands are not documented publicly
- visible `version` command is present
