---
name: creating-azd-extensions
description: >-
  Author, build, and publish Azure Developer CLI (azd) extensions in Go. USE FOR: creating new azd extensions, implementing lifecycle hooks, exposing custom CLI commands, writing extension.yaml manifests, adding metadata/version commands, hardening cross-platform build scripts, release workflows, registry distribution, scaffold templates with embed.FS, wrapping external CLIs, and kube context resolution from azd environment state.
version: 1.9.0
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

## Version management

Keep version in exactly two places and validate consistency at release time.

| File | Format | Example |
|------|--------|---------|
| `extension.yaml` | YAML `version:` field (may be quoted) | `version: "1.0.0"` |
| `version.txt` | Plain text, no quotes | `1.0.0` |

The build script reads `version.txt` and passes it via ldflags. The release workflow validates both files match the git tag.

**Never** hardcode the version in `main.go` beyond the `"dev"` default. The actual version is injected at build time:

```go
// main.go
var version = "dev"
```

```bash
# build.sh
VERSION=$(tr -d '[:space:]' < version.txt)
go build -ldflags "-X main.version=${VERSION} -s -w" -o dist/my-extension .
```

### Version bump checklist

When bumping a version:

1. Update `version.txt`
2. Update `extension.yaml` version field
3. Tag the commit: `git tag v1.0.1 && git push origin v1.0.1`

## Release workflow

A release workflow should build cross-platform binaries, generate `registry.json`, create a GitHub Release, and optionally open a PR against the azd extensions registry.

### Version validation (release step)

Validate that the git tag, `version.txt`, and `extension.yaml` all agree before building:

```bash
VERSION="${GITHUB_REF_NAME#v}"

# Check version.txt
test "$(tr -d '[:space:]' < version.txt)" = "${VERSION}"

# Check extension.yaml (handles quoted and unquoted values)
MANIFEST_VERSION=$(grep -E '^version:' extension.yaml | head -1 | \
  sed 's/^version:[[:space:]]*//' | tr -d '"'"'"'[:space:]')
if [ "${MANIFEST_VERSION}" != "${VERSION}" ]; then
  echo "extension.yaml version '${MANIFEST_VERSION}' does not match tag '${VERSION}'"
  exit 1
fi
```

The `tr -d '"'"'"'[:space:]'` pattern uses bash quote concatenation to strip double quotes, single quotes, and whitespace from the extracted value. Do not use Python for this (avoids PyYAML dependency on GitHub runners).

### Registry.json generation

Generate `registry.json` from a Python script that reads the `extension.yaml` manifest and produces output matching the [official azd extension registry schema](https://github.com/Azure/azure-dev/blob/main/cli/azd/extensions/registry.schema.json).

Required schema fields:

| Level | Required Fields |
|-------|----------------|
| Extension | `id`, `namespace`, `displayName`, `description`, `versions` |
| Version | `version`, `usage`, `examples`, plus either `artifacts` or `dependencies` |
| Artifact | `url` (required), `entryPoint` (optional), `checksum` (optional) |

There is no `language` or `source` field in the official schema. Do not invent extra fields.

### Cross-platform build matrix

Build for five targets. Use `CGO_ENABLED=0` for static binaries:

```bash
PLATFORMS=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")
for platform in "${PLATFORMS[@]}"; do
  GOOS="${platform%/*}" GOARCH="${platform#*/}" \
    CGO_ENABLED=0 go build -ldflags "${LDFLAGS}" -o "bin/${GOOS}/${GOARCH}/my-extension" .
done
```

### Conditional registry update

If your release workflow opens a PR against an external registry, make the step conditional on the PAT being configured. This prevents the workflow from failing on forks or when the secret is not set:

```yaml
env:
  REGISTRY_PAT_SET: ${{ secrets.REGISTRY_PAT != '' }}
steps:
  - name: Open registry PR
    if: env.REGISTRY_PAT_SET == 'true'
    env:
      GH_TOKEN: ${{ secrets.REGISTRY_PAT }}
    run: |
      # clone registry, update entry, push branch, create PR
```

## Scaffold templates with embed.FS

Extensions that scaffold projects typically embed template directories.

### File structure

```text
internal/
  scaffold/
    embed.go           # //go:embed templates
    scaffold.go        # template copy logic
    scaffold_test.go   # validation tests
    templates/
      blank/
        azure.yaml
        drasi/
          drasi.yaml
          sources/
            example-source.yaml
      cosmos-change-feed/
        ...
```

### Embed directive

```go
// internal/scaffold/embed.go
package scaffold

import "embed"

//go:embed templates
var Content embed.FS
```

### Template validation in tests

Validate embedded templates catch missing fields at test time, not deployment time:

```go
func TestTemplates_RequiredFields(t *testing.T) {
    t.Parallel()

    err := fs.WalkDir(Content, "templates", func(path string, d fs.DirEntry, err error) error {
        if err != nil || d.IsDir() || !strings.HasSuffix(path, ".yaml") {
            return err
        }

        data, err := fs.ReadFile(Content, path)
        require.NoError(t, err)

        var doc map[string]interface{}
        require.NoError(t, yaml.Unmarshal(data, &doc), "parsing %s", path)

        // Validate domain-specific required fields
        if kind, ok := doc["kind"]; ok {
            switch kind {
            case "Source":
                props, _ := doc["properties"].(map[string]interface{})
                assert.Contains(t, props, "kubeConfig",
                    "source template %s missing kubeConfig", path)
            }
        }

        return nil
    })
    require.NoError(t, err)
}
```

### Path handling

Use `fs.WalkDir` (not `filepath.WalkDir`) on `embed.FS`. Embedded paths always use forward slashes, even on Windows.

## Wrapping an external CLI

Extensions that shell out to an external CLI (e.g., `drasi`, `kubectl`, `helm`) need consistent error wrapping and output parsing.

### Execution helper

```go
func runCLI(ctx context.Context, name string, args ...string) (string, error) {
    cmd := exec.CommandContext(ctx, name, args...)
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    if err := cmd.Run(); err != nil {
        return "", fmt.Errorf("%s %s: %w\nstderr: %s",
            name, strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
    }
    return strings.TrimSpace(stdout.String()), nil
}
```

### Output parsing

External CLIs may change output format between versions. Parse defensively:

- Detect the delimiter (tabs vs pipes vs spaces)
- Trim whitespace from every field
- Skip header rows and blank lines
- Validate column count before indexing

```go
func parseListOutput(raw string) ([]Item, error) {
    lines := strings.Split(raw, "\n")
    var items []Item
    for i, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" || i == 0 { // skip blank lines and header
            continue
        }

        cols := splitAndTrim(line, "|")
        if len(cols) < 3 {
            continue // skip malformed lines
        }

        items = append(items, Item{
            Name:   cols[0],
            Kind:   cols[1],
            Status: cols[2],
        })
    }
    return items, nil
}
```

### CLI path resolution

Check that the external CLI exists before attempting to use it. Fail fast with an actionable message:

```go
if _, err := exec.LookPath("drasi"); err != nil {
    return fmt.Errorf("drasi CLI not found on PATH. Install from https://drasi.io/docs/getting-started: %w", err)
}
```

## Kube context resolution from azd environment state

Extensions that operate on Kubernetes need to resolve the correct kube context. The pattern is: read `AZURE_AKS_CONTEXT` from azd environment state, then use it for kubectl/CLI operations.

```go
func resolveKubeContext(ctx context.Context, envName string) (string, error) {
    if envName == "" {
        return "", nil // use current kube context
    }

    azdClient, err := azdext.NewAzdClient()
    if err != nil {
        return "", fmt.Errorf("creating azd client: %w", err)
    }
    defer azdClient.Close()

    resp, err := azdClient.Environment().GetValue(ctx, &azdext.GetEnvRequest{
        EnvName: envName,
        Key:     "AZURE_AKS_CONTEXT",
    })
    if err != nil {
        return "", fmt.Errorf("getting AKS context for environment %s: %w", envName, err)
    }

    return resp.Value, nil
}
```

### Where to resolve

Read `--environment` from the **root** persistent flag, not a local subcommand flag. Resolve kube context once at the top of `RunE`, then pass it to helpers:

```go
RunE: func(cmd *cobra.Command, args []string) error {
    env, _ := cmd.Root().Flags().GetString("environment")
    kubeCtx, err := resolveKubeContext(cmd.Context(), env)
    if err != nil {
        return err
    }

    // Pass kubeCtx to CLI wrappers
    cliArgs := []string{"list"}
    if kubeCtx != "" {
        cliArgs = append(cliArgs, "--context", kubeCtx)
    }
    // ...
}
```

### Common mistake

Do not redeclare `--environment` as a local flag on subcommands when it is already a persistent flag on root. This shadows the root flag and the user's value is silently ignored.

## Health diagnostics

Extensions that manage infrastructure should include a `diagnose` command with structured health checks.

### Check structure

```go
type HealthCheck struct {
    Name    string
    Status  string // "ok", "fail", "skipped"
    Message string
}
```

Report "skipped" (not "ok") when a check cannot run because a dependency is absent. For example, if Key Vault auth is not configured, report `skipped` rather than giving a false positive `ok`.

### Minimum checks for a Kubernetes-based extension

1. AKS connectivity (can kubectl reach the cluster?)
2. Extension API health (is the runtime responding?)
3. Runtime dependencies (Dapr, operators, etc.)
4. Secret store auth (can the service access Key Vault?)
5. Observability (is Log Analytics receiving data?)

## Avoid these mistakes

- documenting `listen` or `metadata` as user-facing commands
- constructing `AzdClient` at root-command creation time
- writing stdout logs from gRPC-driven integration paths
- telling users to add MCP/service-target/framework-service capabilities without a real use case
- relying on stale examples instead of the pinned SDK you actually build against
- declaring a flag that no code reads (dead flags pass tests silently but confuse users)
- redeclaring a root persistent flag as a local flag on a subcommand (shadows the root value)
- assuming external CLI output uses tab delimiters (detect the actual delimiter)
- discarding stderr from `exec.Command` (makes failures undiagnosable)
- reporting "ok" for health checks that were skipped (use "skipped" status instead)
- using Python in GitHub Actions workflows when bash builtins suffice (avoids dependency issues)
- hardcoding version in Go source beyond the `"dev"` default
- forgetting to update both `version.txt` and `extension.yaml` when bumping versions
- shipping Kubernetes source templates without required properties (validate in tests)

## Validation checklist

- `go test ./...`
- `go build ./...`
- run build scripts successfully (both `build.sh` and `build.ps1`)
- metadata capability declared when appropriate
- hidden integration commands are not documented publicly
- visible `version` command is present
- `version.txt` and `extension.yaml` version fields match
- embedded templates validated in tests (required fields, valid YAML)
- every declared flag is read in its command's `RunE`
- no root persistent flag shadowed by local subcommand flags
- external CLI errors include stderr in the wrapped error message
- health checks report "skipped" (not "ok") when a dependency is absent
- release workflow validates version consistency before building
- `registry.json` conforms to the [official azd extension registry schema](https://github.com/Azure/azure-dev/blob/main/cli/azd/extensions/registry.schema.json)
- registry update step is conditional on PAT secret being configured
- cross-platform binaries built with `CGO_ENABLED=0`
