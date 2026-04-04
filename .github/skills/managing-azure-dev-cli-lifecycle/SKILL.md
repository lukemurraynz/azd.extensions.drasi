---
name: managing-azure-dev-cli-lifecycle
description: >-
  Manages complete Azure Developer CLI workflows safely, covering provisioning, deployment, monitoring, and cleanup. USE FOR: azd up, azd provision, azd deploy, azd down --purge, troubleshoot azd failures, manage multi-environment configurations, cleanup orphaned Azure resources, switch azd environments.
license: MIT
---

# Managing Azure Developer CLI Lifecycle

## Overview

This skill provides structured workflows for the complete Azure Developer CLI (azd) lifecycle: from safe provisioning through efficient deployment to careful cleanup. It includes validation steps, common trap detection, and troubleshooting patterns for complex stacks (Kubernetes, App Service, databases, container registries, and hybrid setups).

Designed for teams using azd in CI/CD pipelines and local development with infrastructure-as-code (Bicep/Terraform), Kubernetes clusters, and managed services. **Reusable across all project types.**
Examples in linked action files intentionally use placeholders like `<env-name>`, `<namespace>`, and `<api-url>` so guidance can be applied to future projects without renaming assumptions.

---

## Capabilities

| Capability                | Action                                                    | Description                                                                                                                                                                                        |
| ------------------------- | --------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Provision Safely**      | [provision-safely](actions/provision-safely.md)           | Validate prerequisites, check Azure credentials, review what-if analysis, and provision infrastructure with rollback awareness. Includes environment variable setup and infrastructure validation. |
| **Deploy Efficiently**    | [deploy-efficiently](actions/deploy-efficiently.md)       | Build container images, push to registry, trigger deployment, and verify post-deployment health checks in proper order. Covers SPA build-time variables and configuration injection.               |
| **Cleanup Completely**    | [cleanup-completely](actions/cleanup-completely.md)       | Understand `azd down --purge` semantics (recommended default), verify cleanup ordering (K8s → databases → infrastructure), detect stranded resources, and purge soft-deleted Azure services.       |
| **Troubleshoot Failures** | [troubleshoot-failures](actions/troubleshoot-failures.md) | Diagnose azd exit codes, recover from stuck deployments, reset environment state, and resolve common patterns (image pull errors, firewall blocks, soft-delete conflicts, etc.).                   |
| **Manage Environments**   | [manage-environments](actions/manage-environments.md)     | Set up multi-environment configurations, switch between dev/staging/prod, configure secrets, and validate environment isolation.                                                                   |

---

## Standards

| Standard                               | File                                                           | Description                                                                                                                                                                                                                                                                                         |
| -------------------------------------- | -------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Pre-flight & Deployment Checklists** | [checklist.md](standards/checklist.md)                         | Go/no-go criteria before each phase, deployment readiness checks, post-deployment validation, and template authoring checklist (versioning, Renovate, README requirements).                                                                                                                         |
| **Common Traps**                       | [common-traps.md](standards/common-traps.md)                   | 15 documented pitfalls: environment caching, state leaks, K8s orphans, ACR image cleanup, event-system persistence, SPA build-time variables, CORS mismatches, database firewall issues, subscription mismatches, soft-delete persistence, Entra ID cleanup, multi-region tag collisions, and more. |
| **Cleanup Ordering**                   | [cleanup-ordering.md](standards/cleanup-ordering.md)           | Dependency graph for safe resource cleanup (what must be deleted first to avoid orphans). Generic for any infrastructure.                                                                                                                                                                           |
| **Environment Variables**              | [environment-variables.md](standards/environment-variables.md) | Key azd environment variables, default parameter value syntax (`${VAR=default}`), optional resources pattern, and loading `.azure/<env>/.env` for local integration tests.                                                                                                                          |
| **Authoring Hooks**                    | [authoring-hooks.md](standards/authoring-hooks.md)             | Best practices for writing azd lifecycle hooks: PowerShell parameter defaults for testability, Azure CLI auth sync, `predown` hook patterns for Entra ID cleanup, and Entra ID eventual consistency retry logic.                                                                                    |

---

## Principles

1. **Validation First**: Always run `azd provision --preview` (or `what-if`) before provisioning; catch errors early.
2. **Purge by Default**: Use `azd down --purge` as the standard cleanup (avoids soft-delete issues, gives fresh state).
3. **Environment Isolation**: Environment variables leak between commands; verify correct environment with `azd env list` before critical ops.
4. **Ordered Teardown**: Delete workload resources before dependent data services before shared infrastructure to avoid orphans and cost leaks.
5. **Observability**: Log azd exit codes, command timings, and resource state at each phase for troubleshooting.
6. **ADAC (Auto-Detect, Auto-Declare, Auto-Communicate)**: Runtime surfaces must detect live data health, declare it in UI/API, and communicate degraded modes with actionable reasons.

---

## Incident-Proven Guardrails (Mandatory)

These are mandatory execution rules to prevent common deployment regressions:

1. **Hard preflight gates before `azd up`/`azd deploy`**
   - Docker daemon must be reachable (`docker info` succeeds, not just `docker version` client).
   - Correct environment must be selected explicitly (`azd env select <env>` + verify with `azd env list`).
   - `kubectl` context must target the expected cluster/namespace before postdeploy hooks.

2. **Fail-fast postdeploy behavior**
   - Treat any failed `docker push`, `kubectl apply`, secret lookup, or script step as fatal.
   - Do not continue deployment after a failed image push or failed secret retrieval.

3. **Soft-delete and global-name collision handling**
   - Detect `NameUnavailable` for App Configuration/Key Vault.
   - Either purge soft-deleted resources or rotate suffix variables (`APPCONFIG_NAME_SUFFIX`, `KEYVAULT_NAME_SUFFIX`) before retry.

4. **Transient error policy**
   - Retry transient Azure control-plane errors (e.g., PostgreSQL `ServerIsBusy`) with bounded exponential backoff.
   - Do not retry non-transient auth/configuration errors without a state change.

5. **Deployment Definition of Done**
   - All workload pods are `Running` and ready.
   - API and frontend endpoints return successful health/status responses.
   - Backend tests, frontend unit tests, and smoke E2E checks complete successfully.
   - ADAC verified: health endpoints and UI declare data freshness, connection state, and degraded mode reason when applicable.

---

## Usage

### Quick Reference

**Provision:**

```bash
azd provision --preview                # Validate first (no changes)
azd provision                           # Provision infrastructure
```

**Deploy:**

```bash
azd deploy                              # Build, push, and deploy apps
azd monitor                             # Check health post-deploy
```

**Cleanup (Default: Use `--purge`):**

```bash
azd down --purge                        # Delete infra AND .azd state (recommended; avoids soft-delete issues)
# OR (rare)
azd down                                # Delete infra only; keep .azd state (only if re-deploying immediately)
```

### When to Use Each Action

- **Provision Safely**: Before running `azd up` or `azd provision` in a new environment
- **Deploy Efficiently**: After provisioning succeeds, when updating application code or configuration
- **Cleanup Completely**: When tearing down environments or debugging failed deployments
- **Troubleshoot Failures**: When a step fails with a non-zero exit code or stuck resources
- **Manage Environments**: When setting up dev/staging/prod or switching between multiple deployments

---

## Key Concepts

### `azd up` vs `azd provision` vs `azd deploy`

| Command         | Scope                | Use Case                              |
| --------------- | -------------------- | ------------------------------------- |
| `azd up`        | Provision + deploy   | First-time setup or full refresh      |
| `azd provision` | Infrastructure only  | Change Bicep, update infrastructure   |
| `azd deploy`    | Application only     | Update code, rebuild images           |
| `azd package`   | Package only         | Build artifacts without deploying     |
| `azd restore`   | Restore dependencies | Install app dependencies before build |

### `azd down` vs `azd down --purge` (Recommended: Use `--purge` by Default)

| Command            | Keeps                         | Deletes                                                | Use Case                                                  | Recommended                             |
| ------------------ | ----------------------------- | ------------------------------------------------------ | --------------------------------------------------------- | --------------------------------------- |
| `azd down`         | `.azd/` environment directory | Azure resources only                                   | Rare: re-deploy to same environment immediately           | ⚠️ Not recommended (soft-delete issues) |
| `azd down --purge` | Nothing                       | Azure resources + local state + soft-deleted resources | Standard cleanup, reset, environment teardown, full purge | ✅ **Default choice**                   |

### `azd pipeline config` — CI/CD Setup

Automates CI/CD pipeline creation for GitHub Actions or Azure Pipelines:

```bash
azd pipeline config                    # Interactive setup
azd pipeline config --provider github  # GitHub Actions with OIDC (default auth)
azd pipeline config --provider azdo    # Azure DevOps Pipelines (requires PAT)
```

What it does:

1. Authenticates with Azure and creates a service principal
2. Configures OIDC (GitHub) or client credentials (Azure DevOps) authentication
3. Copies pipeline definition files from template (e.g., `.github/workflows/azure-dev.yml`)
4. Sets pipeline variables and secrets
5. Commits, pushes, and triggers initial pipeline run

### `azd config` — Global Defaults

Set persistent defaults that apply across all environments:

```bash
azd config set defaults.subscription <subscription-id>  # Default subscription
azd config set defaults.location australiaeast           # Default region
azd config show                                          # View all config
azd config unset defaults.location                       # Remove a default
azd config reset                                         # Reset all config
azd config list-alpha                                    # List alpha features
```

### Common Exit Codes

| Code | Meaning            | Action                                                                    |
| ---- | ------------------ | ------------------------------------------------------------------------- |
| 0    | Success            | Proceed to next step                                                      |
| 1    | Generic error      | Check logs; see [troubleshoot-failures](actions/troubleshoot-failures.md) |
| 2    | Invalid parameters | Verify environment variables and azure.yaml                               |
| 127  | Command not found  | Install missing tool (kubectl, docker, bicep CLI)                         |

---

## Hosting Options (`host` in `azure.yaml`)

azd supports multiple hosting targets. Set `host` per service in `azure.yaml`:

| Host Value     | Target                   | Notes                                                                                           |
| -------------- | ------------------------ | ----------------------------------------------------------------------------------------------- |
| `containerapp` | Azure Container Apps     | Preferred for containerized workloads. Supports `docker.remoteBuild: true` for ACR-based builds |
| `appservice`   | Azure App Service        | Web apps, APIs. Supports code and container deployment                                          |
| `function`     | Azure Functions          | Event-driven compute. Supports all trigger types                                                |
| `aks`          | Azure Kubernetes Service | Full Kubernetes orchestration                                                                   |
| `staticwebapp` | Azure Static Web Apps    | Frontend SPAs, static sites with optional API                                                   |
| `ai.endpoint`  | Azure AI endpoints       | AI model hosting                                                                                |
| `springapp`    | Azure Spring Apps        | Spring Boot applications                                                                        |

### Service Configuration Example

```yaml
name: my-app
metadata:
  template: my-app@1.0.0
services:
  api:
    project: ./src/api
    language: csharp
    host: containerapp
    docker:
      path: ./Dockerfile
      context: ../
      remoteBuild: true # Build in ACR (no local Docker required)
  web:
    project: ./src/web
    language: js
    host: staticwebapp
  worker:
    project: ./src/worker
    language: python
    host: function
```

### `azd compose` — Resource Declaration

Declare dependent resources directly in `azure.yaml` using the `resources` block:

```yaml
resources:
  api:
    type: host.containerapp
    port: 8080
    uses:
      - mydb
      - mycache
  mydb:
    type: db.postgres
  mycache:
    type: db.redis
  chat:
    type: ai.openai.model
    model:
      name: gpt-4o
      version: "2024-08-06"
```

### Docker Configuration

| Option               | Purpose                                        | Default                   |
| -------------------- | ---------------------------------------------- | ------------------------- |
| `docker.path`        | Path to Dockerfile relative to service project | `./Dockerfile`            |
| `docker.context`     | Build context directory                        | Service project directory |
| `docker.remoteBuild` | Build image in ACR instead of locally          | `false`                   |
| `docker.registry`    | Custom registry (overrides default ACR)        | azd-provisioned ACR       |
| `docker.image`       | Pre-built image reference (skip build)         | —                         |

**Remote builds** (`remoteBuild: true`) are recommended for CI/CD pipelines and environments without Docker installed locally. The source is uploaded to ACR and built server-side.

### Hook Scopes

Hooks can be defined at two levels in `azure.yaml`:

```yaml
# Global hooks — run for all services
hooks:
  preprovision:
    shell: pwsh
    run: ./hooks/preprovision.ps1
  postprovision:
    shell: pwsh
    run: ./hooks/postprovision.ps1

# Service-level hooks — run only for this service
services:
  api:
    project: ./src/api
    host: containerapp
    hooks:
      predeploy:
        shell: pwsh
        run: ./hooks/api-predeploy.ps1
      postdeploy:
        shell: pwsh
        run: ./hooks/api-postdeploy.ps1
```

Available hook events: `prerestore`, `postrestore`, `preprovision`, `postprovision`, `prepackage`, `postpackage`, `predeploy`, `postdeploy`, `predown`, `postdown`

---

## Integration with Copilot

This skill activates automatically when you mention:

- `azd up`, `azd provision`, `azd deploy`, `azd down`, `azd down --purge`
- "Provision infrastructure", "deploy to Azure", "cleanup resources", "clean up environment"
- "azd failed", "azd error", "stuck deployment", "orphaned resources"
- "multiple environments", "switch environments", "dev/staging/prod setup"

Copilot will recommend the relevant action file and checklist, validate inputs, and guide you through the workflow with explicit success criteria at each step.

---

## References

- [Azure Developer CLI (azd) Documentation](https://learn.microsoft.com/azure/developer/azure-developer-cli/)
- [azure.yaml JSON Schema (Official)](https://github.com/Azure/azure-dev/blob/main/schemas/v1.0/azure.yaml.json) — Validate your azure.yaml against this schema
- [AZD Awesome Gallery](https://github.com/Azure/awesome-azd) — Pre-built templates for various frameworks and platforms
- [Bicep Best Practices](https://learn.microsoft.com/azure/azure-resource-manager/bicep/best-practices)
- [Kubernetes Deployment Best Practices](../../instructions/kubernetes-deployment-best-practices.instructions.md)
- [App Service Deployment](https://learn.microsoft.com/azure/app-service/deploy-best-practices)
- [Container Instances](https://learn.microsoft.com/azure/container-instances/container-instances-overview)
- [Static Web Apps](https://learn.microsoft.com/azure/static-web-apps/overview)

---

## Currency

- **Date checked:** 2026-03-31
- **Sources:** Microsoft Learn MCP (`microsoft_docs_search`)
- **Authoritative references:** [Azure Developer CLI documentation](https://learn.microsoft.com/azure/developer/azure-developer-cli/), [azure.yaml schema](https://github.com/Azure/azure-dev/blob/main/schemas/v1.0/azure.yaml.json)

### Verification Steps

1. Confirm latest azd CLI version and any new host types or lifecycle hooks
2. Verify `azure.yaml` schema fields against current JSON schema
3. Check for new `azd` commands or deprecated flags in release notes

---

## Related Skills

- [Azure Deployment Preflight](../azure-deployment-preflight/SKILL.md) — Pre-deployment validation for azd projects
- [Azure Defaults](../azure-defaults/SKILL.md) — Naming, tagging, and security baselines
- [GitHub Actions CI/CD](../github-actions-ci-cd/SKILL.md) — CI/CD pipelines for azd deployments
