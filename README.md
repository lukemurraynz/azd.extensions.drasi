# Drasi azd extension

[![Open in GitHub Codespaces](https://github.com/codespaces/badge.svg)](https://codespaces.new/lukemurraynz/azd.extensions.drasi?quickstart=1)

An [Azure Developer CLI](https://aka.ms/azd) extension for [Drasi](https://drasi.io) that lets you scaffold, provision, deploy, and operate Drasi reactive data pipeline workloads using native azd workflows.

## Why this exists

Drasi requires coordinating several Azure resources (AKS, Key Vault, managed identity, Workload Identity federation) and a specific component deployment order (sources, queries, reactions). This extension wraps that complexity in familiar azd commands so you can treat a Drasi application the same way you treat any other azd project.

## Features

- Scaffold new Drasi projects from built-in templates (blank, blank-terraform, cosmos-change-feed, event-hub-routing, query-subscription)
- Offline validation of sources, queries, reactions, and middleware before any deployment
- Provision AKS with OIDC and Workload Identity, Key Vault, Log Analytics, and the Drasi runtime in a single command
- Deploy components in dependency order with per-component health checks
- Real-time status, streaming logs, and five-point health diagnostics
- Teardown individual components or full infrastructure (force-gated)
- Runtime upgrade command for existing clusters (force-gated)
- Key Vault secret reference translation (no secrets stored in configuration files)
- Environment overlays for staging and production parameter overrides

## Prerequisites

These tools are needed to use the extension after installation.

| Tool | Version | Install |
| ------ | --------- | --------- |
| Azure Developer CLI (`azd`) | >= 1.10.0 | <https://aka.ms/azd> |
| Drasi CLI | >= 0.10.0 | <https://drasi.io/docs/getting-started> |
| Azure CLI | >= 2.60.0 | <https://aka.ms/azcli> |
| Docker | >= 24.0 | <https://www.docker.com> |
| kubectl | >= 1.28 | <https://kubernetes.io/docs/tasks/tools/> |

Go is only needed if you are building the extension from source. See [Contributing](#contributing).

## Installation

```bash
azd extension install azd-drasi
```

Verify:

```bash
azd drasi --help
```

## Publish and consume through GitHub Releases

This repository can be used as a GitHub-hosted azd extension source for other systems.

The intended flow is:

1. push a version tag like `v1.0.1`
2. let `.github/workflows/release.yml` build cross-platform binaries
3. let that workflow create a GitHub Release
4. let that workflow upload:
   - platform archives
   - `registry.json`

After that, other systems can add this repository as a custom azd extension source.

### Maintainer flow

From a clean working tree:

```bash
git tag v1.0.1
git push origin v1.0.1
```

That triggers the release workflow, which publishes release assets and a `registry.json` file that points at those assets.

### Consumer flow on another machine

Add the GitHub Release registry as an azd extension source:

```bash
azd extension source add \
  -n drasi-github \
  -t url \
  -l "https://github.com/lukemurraynz/azd.extensions.drasi/releases/latest/download/registry.json"
```

Then install from that source:

```bash
azd extension install azure.drasi -s drasi-github
```

Verify:

```bash
azd drasi --help
azd drasi version
```

### Notes

- `registry.json` is the important part for azd consumption; GitHub Releases alone are not enough.
- azd installs from the registry entry, which then points to the correct release asset for the current platform.
- If you later publish through the official azd extensions registry, consumers can install without adding a custom source first.

## Quick start

See the [full quick start guide](specs/001-azd-drasi-extension/quickstart.md) for detailed steps. The short version:

> [!TIP]
> The fastest way to get all prerequisites is to open this repository in the included Dev Container. In VS Code, run **Dev Containers: Reopen in Container** or click the Codespace badge above. All tools (azd, Drasi CLI, Go, Docker, kubectl, Azure CLI) are pre-installed.

1. Install the extension (see above).
2. Scaffold a project: `azd drasi init --template cosmos-change-feed`
3. Validate: `azd drasi validate`
4. Authenticate: `azd auth login`
5. Provision infrastructure: `azd drasi provision`
6. Deploy components: `azd drasi deploy`

## Command reference

### Global flags

These flags are available on every command.

| Flag                | Type   | Default | Description                                                                         |
| ------------------- | ------ | ------- | ----------------------------------------------------------------------------------- |
| `--debug`           | bool   | `false` | Enable verbose debug logging.                                                       |
| `-e, --environment` | string |         | Name of the azd environment to use. Controls which AKS context the command targets. |
| `--output`          | string | `table` | Output format. Accepted values: `table`, `json`.                                    |

### init

Scaffold a new Drasi project from a built-in template.

| Flag           | Type   | Default | Description                                                                                                                              |
| -------------- | ------ | ------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| `--template`   | string | `blank` | Template name. One of: `blank`, `blank-terraform`, `cosmos-change-feed`, `event-hub-routing`, `query-subscription`, `postgresql-source`. |
| `--output-dir` | string | `.`     | Directory to write scaffolded files into.                                                                                                |
| `--force`      | bool   | `false` | Overwrite existing files without prompting.                                                                                              |

```bash
azd drasi init --template cosmos-change-feed
azd drasi init --template blank-terraform --output-dir ./my-project
```

### validate

Validate the Drasi configuration offline (no cluster or network access required).

| Flag       | Type   | Default            | Description                      |
| ---------- | ------ | ------------------ | -------------------------------- |
| `--config` | string | `drasi\drasi.yaml` | Path to the Drasi manifest file. |
| `--strict` | bool   | `false`            | Treat warnings as errors.        |

```bash
azd drasi validate
azd drasi validate --strict
azd drasi validate --config path/to/drasi.yaml
```

### provision

Provision Azure infrastructure (AKS, Key Vault, UAMI, Log Analytics) and install the Drasi runtime.

| Flag                | Type   | Default | Description                  |
| ------------------- | ------ | ------- | ---------------------------- |
| `-e, --environment` | string |         | Target azd environment name. |

```bash
azd drasi provision
azd drasi provision --environment staging
```

### deploy

Deploy Drasi sources, queries, middleware, and reactions in dependency order.

| Flag                | Type   | Default            | Description                                           |
| ------------------- | ------ | ------------------ | ----------------------------------------------------- |
| `--config`          | string | `drasi\drasi.yaml` | Path to the Drasi manifest file.                      |
| `--dry-run`         | bool   | `false`            | Preview what would be applied without making changes. |
| `-e, --environment` | string |                    | Target azd environment for overlay resolution.        |

```bash
azd drasi deploy
azd drasi deploy --dry-run
azd drasi deploy --environment prod
azd drasi deploy --output json
```

### status

Show component health and state for the active cluster.

| Flag     | Type   | Default | Description                                                                              |
| -------- | ------ | ------- | ---------------------------------------------------------------------------------------- |
| `--kind` | string |         | Filter by component kind. One of: `source`, `continuousquery`, `middleware`, `reaction`. |

Uses the active kube context, or resolves `AZURE_AKS_CONTEXT` from azd environment state when `--environment` is passed as a root flag.

```bash
azd drasi status
azd drasi status --kind source
azd drasi --environment staging status
azd drasi status --output json
```

### logs

Stream logs from Drasi components. When `--kind` and `--component` are provided for a continuous query, runs in watch mode via `drasi watch`.

| Flag          | Type   | Default | Description                                                                    |
| ------------- | ------ | ------- | ------------------------------------------------------------------------------ |
| `--kind`      | string |         | Component kind. One of: `source`, `continuousquery`, `middleware`, `reaction`. |
| `--component` | string |         | Component ID to stream logs for.                                               |

```bash
azd drasi logs --kind continuousquery --component order-changes
azd drasi logs --kind source --component my-source
azd drasi --environment dev logs --kind continuousquery --component my-query
```

### diagnose

Run five health checks against a live cluster: AKS connectivity, Drasi API, Dapr runtime, Key Vault auth, and Log Analytics.

No command-specific flags. Uses global flags only.

```bash
azd drasi diagnose
azd drasi --environment prod diagnose
azd drasi diagnose --output json
```

### teardown

Remove deployed Drasi components. With `--infrastructure`, also deletes the Azure resource group.

| Flag                | Type   | Default            | Description                                                  |
| ------------------- | ------ | ------------------ | ------------------------------------------------------------ |
| `--config`          | string | `drasi\drasi.yaml` | Path to the Drasi manifest file.                             |
| `--force`           | bool   | `false`            | Required. Confirms the destructive action.                   |
| `--infrastructure`  | bool   | `false`            | Also delete the Azure resource group and all infrastructure. |
| `-e, --environment` | string |                    | Target azd environment.                                      |

```bash
# Remove components only (keep infrastructure)
azd drasi teardown --force

# Remove components and infrastructure
azd drasi teardown --force --infrastructure

# Target a specific environment
azd drasi teardown --force --environment staging
```

### upgrade

Upgrade the Drasi runtime on an existing cluster.

| Flag      | Type | Default | Description                            |
| --------- | ---- | ------- | -------------------------------------- |
| `--force` | bool | `false` | Required. Confirms the upgrade action. |

Uses the active kube context, or resolves `AZURE_AKS_CONTEXT` from azd environment state when `--environment` is passed as a root flag.

```bash
azd drasi upgrade --force
azd drasi --environment prod upgrade --force
```

### version

Print the extension version.

No command-specific flags.

```bash
azd drasi version
```

### Hidden commands

The extension also contains host-invoked commands (`listen` for lifecycle events and
`metadata` for azd command discovery). These are internal integration points and are not part of
the user-facing command surface.

## Example scenarios

### Scaffold, validate, and deploy

A typical first-run workflow from an empty directory:

```bash
mkdir my-drasi-app && cd my-drasi-app
azd init
azd drasi init --template cosmos-change-feed
azd drasi validate --strict
azd auth login
azd drasi provision
azd drasi deploy
azd drasi status
```

### Dry-run before deploying

Preview what components would be applied without touching the cluster:

```bash
azd drasi deploy --dry-run
```

### Multi-environment workflow

Use environment overlays to target different clusters with different parameter values:

```bash
# Provision and deploy to dev
azd drasi provision --environment dev
azd drasi deploy --environment dev

# Provision and deploy to prod with production secrets
azd drasi provision --environment prod
azd drasi deploy --environment prod
```

### Monitor a running deployment

Check component health, stream query output, and run diagnostics:

```bash
azd drasi status
azd drasi status --kind continuousquery --output json
azd drasi logs --kind continuousquery --component order-changes
azd drasi diagnose
```

### JSON output for scripting

All commands support `--output json` for machine-readable output:

```bash
azd drasi status --output json
azd drasi diagnose --output json
azd drasi deploy --output json
```

### Teardown and cleanup

Remove components when done, or tear down the full environment:

```bash
# Remove Drasi components only
azd drasi teardown --force

# Remove everything including Azure infrastructure
azd drasi teardown --force --infrastructure

# Upgrade the Drasi runtime on an existing cluster
azd drasi upgrade --force
```

### Using a non-default config path

Point to a manifest in a different location:

```bash
azd drasi validate --config custom/path/drasi.yaml
azd drasi deploy --config custom/path/drasi.yaml
```

## Configuration

The extension reads `drasi/drasi.yaml` in your project root. Key concepts:

- Secret references use `kind: secret` with `vaultName` and `secretName` to pull values from Key Vault at deploy time. No secrets in source control.
- Environment overlays let you place parameter overrides in `drasi/environments/<name>.yaml` and pass `--environment <name>` to `azd drasi deploy`.
- Environment context routing applies to `status`, `logs`, and `diagnose`. Pass root `--environment <name>` to resolve `AZURE_AKS_CONTEXT` from azd env state and run against that cluster context.
- Feature flags are controlled in the `featureFlags` section of `drasi.yaml`.

See the [configuration reference](docs/configuration-reference.md) for the full YAML schema.

## Architecture

The extension binary communicates with the azd host over gRPC and shells out to the Drasi CLI for cluster operations. See [architecture overview](docs/architecture.md) for flow details and [solution diagram](docs/diagrams/azd-drasi-solution.drawio) for the draw.io architecture artifact.

## Build and run through the devcontainer

The repository includes a ready-to-use devcontainer in `.devcontainer/devcontainer.json`. It installs:

- `azd` 1.23+
- Go 1.22+
- Azure CLI + Bicep
- PowerShell 7.5
- kubectl
- Docker-in-Docker
- Drasi CLI 0.10.0

It also runs `go mod download` after container creation, so the workspace is ready for builds as soon as the devcontainer finishes starting.

### Windows host

1. Install Docker Desktop and VS Code.
2. Install the **Dev Containers** extension in VS Code.
3. Open this repository in VS Code.
4. Run **Dev Containers: Reopen in Container**.
5. Wait for the post-create steps to finish.

Inside the devcontainer, you can build and test with either shell:

```bash
go test ./...
go build ./...
bash ./build.sh
```

or from PowerShell inside the container:

```powershell
go test ./...
go build ./...
pwsh -File ./build.ps1
```

You do **not** need to run both build scripts for a normal development workflow. Pick one:

- use `bash ./build.sh` if you are working in a Bash shell inside the devcontainer
- use `pwsh -File ./build.ps1` if you are working in a PowerShell shell inside the devcontainer

Both scripts run the same verification steps and produce the same `bin/` outputs.

To run the extension locally inside the devcontainer without installing it globally:

```bash
go run . --help
go run . version
```

To build a release-style binary from inside the devcontainer and inspect the outputs:

```bash
bash ./build.sh
ls -R bin
```

### Linux host

1. Install Docker Engine or Docker Desktop.
2. Install VS Code.
3. Install the **Dev Containers** extension in VS Code.
4. Open this repository in VS Code.
5. Run **Dev Containers: Reopen in Container**.
6. Wait for the post-create steps to finish.

Once inside the devcontainer, use the same commands as on Windows because the build happens inside the Linux container image:

```bash
go test ./...
go build ./...
bash ./build.sh
```

If you prefer PowerShell inside the container:

```powershell
pwsh -File ./build.ps1
```

Again, you only need **one** of the two build scripts unless you are explicitly verifying both shell entry points.

To run the extension entry point directly in the container:

```bash
go run . --help
go run . version
```

### Notes about using the devcontainer

- The devcontainer gives Windows and Linux contributors the same toolchain and versions.
- `build.sh` and `build.ps1` both run verification first (`go test ./...` and `go build ./...`) before producing binaries.
- Built artifacts are written under `bin/`.
- If you want to test the azd command surface itself after building, start with:

```bash
go run . --help
go run . version
```

and then install the built binary into a local azd extension source only if you specifically need installation-level testing.

## Repository hygiene

This repository intentionally ignores local build outputs and generated IaC JSON artifacts to keep pull requests readable.

- Local extension binaries under `.local/azd-extension-registry/artifacts/` are ignored
- Generated Bicep JSON outputs under `infra/**/*.json` are ignored
- Generated scaffold template JSON outputs under `internal/scaffold/templates/*/infra/**/*.json` are ignored

If generated artifacts are already present locally, remove them before opening a PR.

Windows PowerShell cleanup example:

```powershell
Remove-Item -LiteralPath ".local/azd-extension-registry/artifacts/azd-drasi.exe" -Force -ErrorAction SilentlyContinue
Remove-Item -LiteralPath "internal/scaffold/templates/blank/infra/main.json" -Force -ErrorAction SilentlyContinue
```

## Troubleshooting

Common errors and remediation steps are in [docs/troubleshooting.md](docs/troubleshooting.md). You can also run:

```bash
azd drasi diagnose
```

to check AKS connectivity, Drasi API health, Dapr runtime, Key Vault auth, and Log Analytics in one pass.

## Contributing

1. Clone the repository.
2. Install prerequisites (Go 1.22+, Azure CLI, Drasi CLI).
3. Run tests: `go test ./...`
4. Build: `./build.sh` (Linux/macOS) or `./build.ps1` (Windows)

Pull requests should include tests for new behavior and should pass `go vet ./...` and `golangci-lint run`.

## License

MIT
