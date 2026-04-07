# Quick Start Guide: azd-drasi Extension

**Feature**: 001-azd-drasi-extension  
**Phase**: 1 — Design (generated reference; update after implementation)

---

## Prerequisites

| Tool                        | Version  | Install                                   |
| --------------------------- | -------- | ----------------------------------------- |
| Azure Developer CLI (`azd`) | ≥ 1.10.0 | <https://aka.ms/azd>                      |
| Drasi CLI                   | ≥ 0.10.0 | <https://drasi.io/docs/getting-started>   |
| Azure CLI                   | ≥ 2.60.0 | <https://aka.ms/azcli>                    |
| Go                          | ≥ 1.22   | <https://go.dev>                          |
| Docker                      | ≥ 24.0   | <https://www.docker.com>                  |
| kubectl                     | ≥ 1.28   | <https://kubernetes.io/docs/tasks/tools/> |

---

## Step 1: Install the Extension

```bash
azd extension install azd-drasi
```

Verify:

```bash
azd drasi --help
```

---

## Step 2: Scaffold a New Project

```bash
mkdir my-drasi-app && cd my-drasi-app
azd init
azd drasi init --template cosmos-change-feed
```

This creates:

```text
my-drasi-app/
├── azure.yaml
├── drasi/
│   ├── drasi.yaml
│   ├── sources/
│   │   └── cosmos-source.yaml
│   ├── queries/
│   │   └── order-changes.yaml
│   └── reactions/
│       └── pubsub-reaction.yaml
└── infra/
    ├── main.bicep
    └── modules/
        ├── aks.bicep
        ├── keyvault.bicep
        └── ...
```

---

## Step 3: Validate Your Configuration

```bash
azd drasi validate
```

All validation runs **offline** — no cluster or network access required.
Fix any reported errors before deploying.

---

## Step 4: Log in to Azure

```bash
azd auth login
```

---

## Step 5: Provision Infrastructure

```bash
azd drasi provision
```

This will:

1. Deploy an AKS cluster with OIDC issuer and Workload Identity enabled
2. Deploy Key Vault, UAMI, Log Analytics workspace (and optional ACR / Cosmos DB / Event Hub)
3. Run `drasi init` to install the Drasi runtime on AKS
4. Configure FederatedIdentityCredential for `system:serviceaccount:drasi-system:drasi-api`

> [!NOTE]
> The provision step creates the FederatedIdentityCredential automatically. If you also follow the standalone Drasi documentation, do not create this credential manually or you will get a conflict error. The extension manages this credential for you.

Expected duration: 8–12 minutes.

---

## Step 6: Deploy Drasi Components

```bash
azd drasi deploy
```

The deployment engine will:

1. Run validation again (pre-deploy gate)
2. Translate Key Vault secret references into Kubernetes Secrets
3. Apply Sources → ContinuousQueries → Reactions in dependency order
4. Wait for each component to reach **Online** state (5 min per component)

---

## Step 7: Check Status

```bash
azd drasi status
```

Expected output:

```text
Component Status — environment: dev

KIND              ID                  STATUS    AGE
Source            cosmos-source       Online    2m
ContinuousQuery   order-changes       Online    2m
Reaction          pubsub-reaction     Online    2m
```

---

## Step 8: View Logs

```bash
# Follow output for a specific continuous query
azd drasi logs --kind continuousquery --component order-changes --follow
```

---

## Step 9: Diagnose Issues

```bash
azd drasi diagnose
```

Runs 5 health checks: AKS connectivity, Drasi API, Dapr runtime, Key Vault auth, Log Analytics.

---

## Step 10: Teardown

```bash
# Remove Drasi components only (keep infrastructure)
azd drasi teardown --force

# Remove components AND Azure infrastructure
azd drasi teardown --force --infrastructure
```

---

## Environment Overrides

Create environment-specific parameter overrides in `drasi/environments/<name>.yaml`:

```yaml
apiVersion: v1
parameters:
  cosmosConnectionString:
    kind: secret
    vaultName: prod-kv
    secretName: cosmos-conn
```

Use with:

```bash
azd drasi deploy --environment prod
```

---

## Dry Run

Preview changes before applying:

```bash
azd drasi deploy --dry-run
```

---

## CI/CD Integration

Add to your `azure.yaml`:

```yaml
hooks:
  preprovision:
    shell: sh
    run: "azd drasi validate"
```

Or use the included GitHub Actions workflow at `.github/workflows/ci.yml`.

---

## Dev Container

Use the included Dev Container for a fully pre-configured environment with all tools installed:

```bash
# Reopen in container (VS Code)
code .
# Dev Containers: Reopen in Container
```

The Dev Container includes: `azd` v1.10.0+, `drasi` v0.10.0+, `dapr`, `go` 1.22, `kubectl`, `bicep`, Azure CLI.
