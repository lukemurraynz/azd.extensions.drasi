# Drasi azd extension

An [Azure Developer CLI](https://aka.ms/azd) extension for [Drasi](https://drasi.io) that lets you scaffold, provision, deploy, and operate Drasi reactive data pipeline workloads using native azd workflows.

## Why this exists

Drasi requires coordinating several Azure resources (AKS, Key Vault, managed identity, Workload Identity federation) and a specific component deployment order (sources, queries, reactions). This extension wraps that complexity in familiar azd commands so you can treat a Drasi application the same way you treat any other azd project.

## Features

- Scaffold new Drasi projects from built-in templates (blank, cosmos-change-feed)
- Offline validation of sources, queries, reactions, and middleware before any deployment
- Provision AKS with OIDC and Workload Identity, Key Vault, Log Analytics, and the Drasi runtime in a single command
- Deploy components in dependency order with per-component health checks
- Real-time status, streaming logs, and five-point health diagnostics
- Teardown individual components or full infrastructure (force-gated)
- Runtime upgrade command for existing clusters (force-gated)
- Key Vault secret reference translation (no secrets stored in configuration files)
- Environment overlays for staging and production parameter overrides

## Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Azure Developer CLI (`azd`) | >= 1.10.0 | https://aka.ms/azd |
| Drasi CLI | >= 0.10.0 | https://drasi.io/docs/getting-started |
| Azure CLI | >= 2.60.0 | https://aka.ms/azcli |
| Go | >= 1.22 | https://go.dev |
| Docker | >= 24.0 | https://www.docker.com |
| kubectl | >= 1.28 | https://kubernetes.io/docs/tasks/tools/ |

## Installation

```bash
azd extension install azd-drasi
```

Verify:

```bash
azd drasi --help
```

## Publish and consume through GitHub Releases

Yes — this repository can be used as a GitHub-hosted azd extension source for other systems.

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
  -l "https://github.com/Azure/azd.extensions.drasi/releases/latest/download/registry.json"
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

1. Install the extension (see above).
2. Scaffold a project: `azd drasi init --template cosmos-change-feed`
3. Validate: `azd drasi validate`
4. Authenticate: `azd auth login`
5. Provision infrastructure: `azd drasi provision`
6. Deploy components: `azd drasi deploy`

## Command reference

| Command | Description |
|---------|-------------|
| `azd drasi init` | Scaffold a new project from a template |
| `azd drasi validate` | Validate configuration offline |
| `azd drasi provision` | Provision Azure infrastructure and Drasi runtime |
| `azd drasi deploy` | Deploy Drasi sources, queries, and reactions |
| `azd drasi status` | Show component health and state (uses active kube context or `--environment` root flag) |
| `azd drasi logs` | Watch continuous query output via `drasi watch` (uses active kube context or `--environment` root flag) |
| `azd drasi diagnose` | Run five health checks against a live cluster (uses active kube context or `--environment` root flag) |
| `azd drasi teardown --force` | Remove components or full infrastructure |
| `azd drasi upgrade --force` | Upgrade the Drasi runtime on an existing cluster |

Run `azd drasi <command> --help` for flags and examples on any command.

The extension also contains hidden host-invoked commands (`listen` for lifecycle events and
`metadata` for azd command discovery). These are internal integration points and are not part of
the user-facing command surface.

## Configuration

The extension reads `drasi/drasi.yaml` in your project root. Key concepts:

- **Secret references**: use `kind: secret` with `vaultName` and `secretName` to pull values from Key Vault at deploy time. No secrets in source control.
- **Environment overlays**: place parameter overrides in `drasi/environments/<name>.yaml` and pass `--environment <name>` to `azd drasi deploy`.
- **Environment context routing**: for `status`, `logs`, and `diagnose`, pass root `--environment <name>` to resolve `AZURE_AKS_CONTEXT` from azd env state and run against that cluster context.
- **Feature flags**: controlled in the `featureFlags` section of `drasi.yaml`.

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
