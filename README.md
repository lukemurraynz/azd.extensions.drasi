# Drasi azd extension

[![Open in GitHub Codespaces](https://github.com/codespaces/badge.svg)](https://codespaces.new/lukemurraynz/azd.extensions.drasi?quickstart=1)

> An unofficial [Azure Developer CLI](https://aka.ms/azd) extension for [Drasi](https://drasi.io) that lets you scaffold, provision, deploy, and operate Drasi reactive data pipeline workloads using native azd workflows.

## Why this exists

Drasi requires coordinating several Azure resources (AKS, Key Vault, managed identity, Workload Identity federation) and a specific component deployment order (sources, queries, reactions). This extension wraps that complexity in familiar azd commands so you can treat a Drasi application the same way you treat any other azd project.

## Features

- Scaffold new Drasi projects from built-in templates (blank, blank-terraform, cosmos-change-feed, event-hub-routing, query-subscription, postgresql-source)
- Offline validation of sources, queries, reactions, and middleware before any deployment
- Provision AKS with OIDC and Workload Identity, Key Vault, Log Analytics, and the Drasi runtime in a single command
- Deploy components in dependency order with per-component health checks
- Real-time status, streaming logs, and five-point health diagnostics
- Teardown individual components or full infrastructure (force-gated)
- Runtime upgrade command for existing clusters (force-gated)
- Key Vault secret reference translation (no secrets stored in configuration files)
- Environment overlays for staging and production parameter overrides

## Prerequisites

### Use the prebuilt extension from Releases (most users)

If you install `azure.drasi` from this repository's GitHub Releases, you do **not** need Go or build tooling.

| Tool                        | Version   | Required for                                                  | Install                                   |
| --------------------------- | --------- | ------------------------------------------------------------- | ----------------------------------------- |
| Azure Developer CLI (`azd`) | >= 1.10.0 | Installing and running the extension                          | <https://aka.ms/azd>                      |
| Drasi CLI                   | >= 0.10.0 | Drasi component operations (`deploy`, `status`, `logs`, etc.) | <https://drasi.io/drasi-server/getting-started/>   |
| Azure CLI                   | >= 2.60.0 | Azure authentication and infrastructure operations            | <https://aka.ms/azcli>                    |
| kubectl                     | >= 1.28   | Cluster connectivity for health/status/log operations         | <https://kubernetes.io/docs/tasks/tools/> |

### Build this repository from source (contributors)

| Tool               | Version | Required for                     | Install                      |
| ------------------ | ------- | -------------------------------- | ---------------------------- |
| Go                 | >= 1.22 | `go test ./...`, local build     | <https://go.dev/dl/>         |
| Bash or PowerShell | recent  | Running `build.sh` / `build.ps1` | Included on most dev systems |

If you also want to run end-to-end commands against Azure/AKS from a local source build, install the runtime tools from the section above (`azd`, Drasi CLI, Azure CLI, `kubectl`).

> [!TIP]
> The fastest way to get all prerequisites is to open this repository in the included Dev Container. In VS Code, run **Dev Containers: Reopen in Container** or click the Codespace badge above. All tools are pre-installed.

## Installation

### Install from GitHub Releases

You can also install from this repository's GitHub Releases by adding it as a custom extension source:

```bash
azd extension source add -n drasi-lukemurray-azdext -t url -l "https://github.com/lukemurraynz/azd.extensions.drasi/releases/latest/download/registry.json"
```

Then install from that source:

```bash
azd extension install azure.drasi -s drasi-lukemurray-azdext
```

Verify:

```bash
azd drasi --help
azd drasi version
```

## Quick start

1. Install the runtime prerequisites from **Use the prebuilt extension from Releases** and then install the extension.
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
| `--config` | string | `drasi/drasi.yaml` | Path to the Drasi manifest file. |
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
| `--config`          | string | `drasi/drasi.yaml` | Path to the Drasi manifest file.                      |
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
| `--config`          | string | `drasi/drasi.yaml` | Path to the Drasi manifest file.                             |
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

```bash
azd drasi version
```

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

The extension binary communicates with the azd host over gRPC and shells out to the Drasi CLI for cluster operations. See the [architecture overview](docs/architecture.md) for details.

## Troubleshooting

Common errors and remediation steps are in [docs/troubleshooting.md](docs/troubleshooting.md). You can also run:

```bash
azd drasi diagnose
```

to check AKS connectivity, Drasi API health, Dapr runtime, Key Vault auth, and Log Analytics in one pass.

## Contributing

1. Clone the repository.
2. Open in the Dev Container (recommended) or install local build prerequisites (Go 1.22+ and Bash/PowerShell). Install runtime prerequisites (`azd`, Drasi CLI, Azure CLI, `kubectl`) only if you plan to run live Azure/AKS workflows.
3. Run tests: `go test ./...`
4. Build: `./build.sh` (Linux/macOS) or `./build.ps1` (Windows)

Pull requests should include tests for new behavior and should pass `go vet ./...` and `golangci-lint run`.

### Releasing

From a clean working tree, push a version tag to trigger the release workflow:

```bash
git tag v1.0.1
git push origin v1.0.1
```

The `.github/workflows/release.yml` workflow builds cross-platform binaries, creates a GitHub Release, and uploads the platform archives along with `registry.json`.

## License

MIT
