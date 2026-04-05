---
name: github-actions-terraform
description: >-
  Debug and fix failing Terraform GitHub Actions workflows for Azure
  infrastructure deployments.
  USE FOR: debugging CI/CD pipeline failures, fixing Terraform authentication
  errors, troubleshooting state backend issues, resolving plan/apply errors,
  or setting up new Terraform deployment workflows.
metadata:
  author: github-copilot-skills-terraform
  version: "1.0.0"
  category: terraform-cicd
---

# GitHub Actions Terraform Debugging Skill

This skill helps you debug and fix failing Terraform GitHub Actions workflows for Azure infrastructure deployments.

## When to Use This Skill

- Debugging failing Terraform CI/CD pipelines
- Troubleshooting authentication issues in GitHub Actions
- Fixing plan/apply workflow failures
- Optimizing Terraform workflow performance
- Setting up new Terraform pipelines

## Common Workflow Failures

### 1. Authentication Failures

#### OIDC/Federated Credentials (Recommended)

```yaml
- name: Azure Login
  uses: azure/login@v2
  with:
    client-id: ${{ secrets.AZURE_CLIENT_ID }}
    tenant-id: ${{ secrets.AZURE_TENANT_ID }}
    subscription-id: ${{ secrets.AZURE_SUBSCRIPTION_ID }}
```

**Common Issues:**

- Missing or incorrect federated credential configuration
- Wrong audience setting
- Repository/branch restrictions not matching

**Fix:**

```bash
# Create federated credential
az ad app federated-credential create \
  --id <app-object-id> \
  --parameters '{
    "name": "github-actions",
    "issuer": "https://token.actions.githubusercontent.com",
    "subject": "repo:org/repo:ref:refs/heads/main",
    "audiences": ["api://AzureADTokenExchange"]
  }'
```

### 2. State Backend Errors

#### State Lock Errors

```
Error: Error acquiring the state lock
```

**Fix:**

```bash
terraform force-unlock <LOCK_ID>
```

> [!WARNING]
> **Never run `terraform force-unlock` without verifying no active run holds the lock.** A force-unlock during an active apply can corrupt state. Before unlocking: (1) Check CI/CD for in-progress runs, (2) Verify the lock holder process has terminated, (3) Consider `terraform plan` with `-lock=false` to confirm state consistency before re-locking.

#### State Access Denied

**Fixes:**

- Verify storage account exists
- Check RBAC permissions (Storage Blob Data Contributor)
- Verify container exists
- Check network access (if private endpoint)

### 3. Provider Initialization Failures

```
Error: Failed to query available provider packages
```

**Fixes:**

```yaml
- name: Setup Terraform
  uses: hashicorp/setup-terraform@v3
  with:
    terraform_version: "~> 1.9"  # Latest stable: 1.14.8 as of 2026-03-31; verify at https://github.com/hashicorp/terraform/releases

- name: Terraform Init
  run: terraform init -upgrade
  env:
    ARM_SKIP_PROVIDER_REGISTRATION: "true"
```

### 4. Plan/Apply Failures

#### Resource Already Exists

**Fix:**

```bash
terraform import azurerm_resource_group.main /subscriptions/.../resourceGroups/rg-name
```

## Debugging Steps

### 1. Enable Debug Logging

```yaml
env:
  TF_LOG: DEBUG
  TF_LOG_PATH: terraform.log
```

### 2. Check Azure Context

```yaml
- name: Debug Azure Context
  run: |
    az account show
    az account list-locations -o table
```

## Best Practices

1. **Use OIDC** - Avoid long-lived secrets
2. **Pin versions** - Terraform, providers, actions
3. **Use environments** - For approval gates
4. **Cache providers** - Speed up runs
5. **Artifact plans** - Ensure apply uses exact plan
6. **Minimal permissions** - Least privilege for service principal

## Additional Resources

For complete workflow templates and detailed debugging guides, see the [reference guide](references/REFERENCE.md).

---

## Currency and verification

- **Date checked:** 2026-03-31
- **Terraform latest stable:** 1.14.8 — verify at [GitHub Releases](https://github.com/hashicorp/terraform/releases)
- **AzureRM provider latest:** 4.x — verify at [Terraform Registry](https://registry.terraform.io/providers/hashicorp/azurerm/latest)
- **hashicorp/setup-terraform action:** v3 — verify at [GitHub Marketplace](https://github.com/hashicorp/setup-terraform)
- **Verification steps:** Run `terraform version` in your workflow; check GitHub release notes before pinning a new version.

### Known pitfalls

| Area | Pitfall | Mitigation |
|------|---------|------------|
| Version pinning | Using an exact old version (e.g., `"1.6.0"`) blocks security fixes and AVM module requirements | Use a pessimistic constraint (`"~> 1.9"`) and verify the latest stable at [hashicorp/terraform releases](https://github.com/hashicorp/terraform/releases) |
| AzureRM provider | AVM Bicep/Terraform modules require Terraform `>= 1.9.0`; older versions will fail `terraform init` | Set `terraform_version: "~> 1.9"` (minimum) in `setup-terraform` |
| OIDC audience | Wrong `audience` in federated credential causes `AADSTS700016` login failure | Audience must be exactly `api://AzureADTokenExchange` |
| State lock | Abandoned lock from cancelled run blocks all subsequent operations | Use `terraform force-unlock <LOCK_ID>` after confirming no active run holds the lock |
| `ARM_SKIP_PROVIDER_REGISTRATION` | Setting this globally can hide real registration errors in new subscriptions | Use only when service principal lacks subscription-level `Microsoft.Authorization/register` permission |

---

## Related Skills

- [Terraform Patterns](../terraform-patterns/SKILL.md) — IaC structure and module patterns
- [Terraform Security Scan](../terraform-security-scan/SKILL.md) — Static analysis in pipelines
- [GitHub Actions CI/CD](../github-actions-ci-cd/SKILL.md) — General CI/CD workflow patterns
