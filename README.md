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
- Teardown individual components or full infrastructure
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
| `azd drasi status` | Show component health and state |
| `azd drasi logs` | Stream or tail logs from Drasi pods |
| `azd drasi diagnose` | Run five health checks against a live cluster |
| `azd drasi teardown` | Remove components or full infrastructure |
| `azd drasi upgrade` | Upgrade the Drasi runtime on an existing cluster |
| `azd drasi listen` | Listen for Drasi reaction events |

Run `azd drasi <command> --help` for flags and examples on any command.

## Configuration

The extension reads `drasi/drasi.yaml` in your project root. Key concepts:

- **Secret references**: use `kind: secret` with `vaultName` and `secretName` to pull values from Key Vault at deploy time. No secrets in source control.
- **Environment overlays**: place parameter overrides in `drasi/environments/<name>.yaml` and pass `--environment <name>` to `azd drasi deploy`.
- **Feature flags**: controlled in the `featureFlags` section of `drasi.yaml`.

See the [configuration reference](docs/configuration-reference.md) for the full YAML schema.

## Architecture

The extension binary communicates with the azd host over gRPC and shells out to the Drasi CLI for cluster operations. See [architecture overview](docs/architecture.md) for a component diagram and flow descriptions.

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
