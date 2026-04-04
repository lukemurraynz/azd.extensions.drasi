---
name: terraform-security-scan
description: >-
  Scan Terraform configurations for security vulnerabilities and compliance
  against CIS and Azure Security Benchmarks.
  USE FOR: running Trivy or Checkov scans, checking CIS compliance, auditing
  IaC for vulnerabilities, implementing security gates in CI/CD pipelines,
  or reviewing Terraform security posture.
metadata:
  author: github-copilot-skills-terraform
  version: "1.0.0"
  category: terraform-security
---

# Terraform Security Scan Skill

This skill helps you perform comprehensive security scanning and compliance checking of Terraform configurations for Azure infrastructure.

## When to Use This Skill

- Reviewing Terraform code for security vulnerabilities
- Checking compliance with security frameworks
- Pre-deployment security gates
- Security audits and assessments
- Pull request security reviews

## Security Check Categories

### Authentication and Secrets

#### Check: No Hardcoded Credentials

**Bad:**

```hcl
output "storage_key" {
  value = azurerm_storage_account.example.primary_access_key
}
```

**Good:**

```hcl
data "azurerm_key_vault_secret" "storage_connection" {
  name         = "storage-connection-string"
  key_vault_id = data.azurerm_key_vault.main.id
}
```

### Encryption

#### Check: Storage Encryption

```hcl
resource "azurerm_storage_account" "secure" {
  name                     = "stsecuredata"
  resource_group_name      = azurerm_resource_group.main.name
  location                 = azurerm_resource_group.main.location
  account_tier             = "Standard"
  account_replication_type = "GRS"

  min_tls_version                 = "TLS1_2"
  public_network_access_enabled   = false
  allow_nested_items_to_be_public = false
}
```

### Network Security

#### Check: NSG Rules

```hcl
resource "azurerm_network_security_group" "web" {
  name                = "nsg-web-tier"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name

  security_rule {
    name                       = "AllowHTTPS"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "443"
    source_address_prefix      = "Internet"
    destination_address_prefix = "*"
  }
}
```

### RBAC and Access Control

#### Check: Least Privilege

```hcl
resource "azurerm_role_assignment" "storage_reader" {
  scope                = azurerm_storage_account.main.id
  role_definition_name = "Storage Blob Data Reader"
  principal_id         = azurerm_user_assigned_identity.app.principal_id
}
```

## Security Scanning Commands

### Static Analysis with Trivy (Recommended)

> **NOTE:** `tfsec` has been deprecated by Aqua Security and merged into Trivy. Use Trivy for all new projects.

> [!WARNING]
> **tfsec is deprecated** and has been merged into Trivy. Existing tfsec configurations must migrate to `trivy config` by Q4 2025. New projects must use Trivy directly. Migration: replace `tfsec --format sarif` with `trivy config --format sarif --scanners misconfig`.

```bash
brew install trivy
trivy config .
trivy config . --format json --output security-report.json
```

### Checkov Scanning

```bash
pip install checkov
checkov -d .
checkov -d . --framework terraform --check CKV_AZURE
```

## Compliance Frameworks

### Azure Security Benchmark

Key controls to verify:

- Network security controls
- Identity management
- Data protection
- Asset management
- Logging and threat detection

### CIS Azure Foundations

Check these sections:

- 1.x - Identity and Access Management
- 3.x - Storage Accounts
- 4.x - Database Services
- 5.x - Logging and Monitoring
- 6.x - Networking

## Integration with CI/CD

### GitHub Actions Security Gate

```yaml
name: Security Scan

on: [pull_request]

jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run Trivy IaC Scan
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: "config"
          scan-ref: "."
          exit-code: "1"
          severity: "CRITICAL,HIGH"

      - name: Run Checkov
        uses: bridgecrewio/checkov-action@master
        with:
          directory: .
          framework: terraform
          soft_fail: false
```

## Additional Resources

For detailed compliance checklists, security patterns, and scanning tool configurations, see the [reference guide](references/REFERENCE.md).

---

## Currency and Verification

- **Date checked:** 2026-03-31
- **Key change:** tfsec has a migration path to Trivy; use `trivy config .` as the primary scanner for new pipelines.
- **Tools verified:** Trivy (replaces tfsec), Checkov, `terraform validate`, `az bicep build`
- **Sources:** [tfsec deprecation notice](https://github.com/aquasecurity/tfsec), [Trivy docs](https://aquasecurity.github.io/trivy/), [Checkov docs](https://www.checkov.io/)
- **Verification steps:**
  1. Verify Trivy installed: `trivy --version`
  2. Run Terraform security scan: `trivy config --tf-vars terraform.tfvars .`
  3. Run Checkov scan: `checkov -d . --framework terraform`

### Known Pitfalls

| Area                    | Pitfall                                                                                           | Mitigation                                                                                     |
| ----------------------- | ------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| tfsec → Trivy migration | tfsec maintenance is reduced and Trivy is the preferred path; avoid introducing new tfsec-only pipelines | Replace `tfsec .` with `trivy config .` in all CI pipelines                                    |
| Trivy false positives   | Trivy may flag valid configurations as insecure (e.g., public subnets in DMZ patterns)            | Use `.trivyignore` for documented exceptions; suppress with inline `#trivy:ignore` comments    |
| Checkov custom policies | Default Checkov policies may not cover organization-specific security requirements                | Create custom Checkov policies for org-specific rules; store in repo under `checkov-policies/` |
| Scan timing in CI       | Running scans after `terraform apply` misses issues already deployed                              | Place scan steps before `terraform plan`; fail the pipeline on HIGH/CRITICAL findings          |
| Variable file coverage  | `trivy config .` doesn't automatically load `.tfvars` files for variable evaluation               | Use `--tf-vars terraform.tfvars` flag to include variable definitions in scan context          |

---

## Related Skills

- [Terraform Patterns](../terraform-patterns/SKILL.md) — IaC structure and module patterns
- [GitHub Actions Terraform](../github-actions-terraform/SKILL.md) — Integrating scans into CI/CD
- [API Security Review](../api-security-review/SKILL.md) — Application-layer security patterns
