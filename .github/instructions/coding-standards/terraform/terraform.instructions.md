---
applyTo: "**/*.tf,**/*.tfvars,**/*.terraform,**/*.tfstate,**/*.tflint.hcl,**/*.tf.json,**/*.tfvars.json"
description: "Terraform Infrastructure as Code best practices following ISE Engineering Playbook guidelines, Azure Verified Modules (AVM), and Azure security standards"
---

# Terraform Infrastructure Instructions

Follow HashiCorp Terraform best practices, Azure security guidance, and ISE Engineering Playbook standards.

**VERIFY-FIRST (MANDATORY):**

- Always validate assumptions using the `iseplaybook` and `microsoft.learn.mcp` MCP servers.
- Terraform, provider, and Azure behaviors are version-dependent — never assume defaults.
- If context is unclear (greenfield vs existing), ask before restructuring.

---

## Decision Gate: Greenfield vs Existing Infrastructure

Before generating or refactoring Terraform:

- **Greenfield**: New infrastructure with no prior state
- **Existing / Legacy**: Live infrastructure or established repository

Rules:

- Do **not** restructure existing repositories unless explicitly requested.
- For existing code, enhance incrementally and document deviations.
- For greenfield projects, follow recommended structure and defaults.

If unclear, ask the user.

---

## Azure Verified Modules (AVM) — Usage Policy

### Platform / Landing Zones / Shared Services

- **Prefer consuming Azure Verified Modules (AVM) directly**

### Productized Modules / Internal IP

- **Use AVM as reference patterns only**
- **Do NOT wrap or depend on AVM modules**

### Application Stacks

- Prefer AVM when available
- Otherwise implement resources directly using AVM patterns

---

## AVM Reference Usage Rules

When learning from AVM implementations, extract:

- Security defaults (TLS, encryption, network isolation)
- Variable validation patterns
- Dynamic blocks for optional features
- Output naming conventions
- Module testing and example structure

Do NOT blindly copy-paste; adapt intentionally and document decisions.

---

## Non-Negotiables

1. Never edit `.tfstate` or `.terraform/`
2. Never commit secrets or credentials
3. Always commit `.terraform.lock.hcl`
4. Authentication priority:
   1. Managed Identity
   2. OIDC / Federated Credentials
   3. Service Principal with certificate
   4. Service Principal with secret (last resort)
5. Make small, reversible changes only
6. `terraform fmt` and `terraform validate` are mandatory
7. Use `moved` blocks for renames
8. Use `import` blocks for brownfield adoption

---

## Security Defaults (REQUIRED)

Unless explicitly justified and documented:

- TLS 1.2 or higher
- HTTPS-only traffic
- `public_network_access_enabled = false`
- Encryption at rest and in transit
- Private Endpoints for PaaS services
- Least-privilege RBAC
- Diagnostic settings enabled with tiered approach

---

## Custom Module Requirements

When creating custom Terraform modules:

- Implement **actual Azure resources**, not module wrappers
- Follow AVM structural and security patterns
- Include **all required files**:

```text
modules/<module-name>/
├── main.tf
├── variables.tf
├── outputs.tf
├── versions.tf
├── README.md
└── examples/
    └── basic/
        ├── main.tf
        ├── variables.tf
        ├── outputs.tf
        ├── terraform.tfvars.example
        └── README.md
```

### Module Example Requirements

- `examples/` is **REQUIRED**
- Examples must be fully deployable
- Document example intent and tradeoffs

### Provider & Version Management

- Always verify latest provider versions before scaffolding
- Use pessimistic constraints (`~>`)
- Pin exact versions for production when required
- Perform provider/module upgrades in isolated PRs

#### Example:

```hcl
terraform {
  required_version = ">= 1.9.0"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.0"
    }
  }
}
```

## State & Environment Isolation

- One state file per environment and blast-radius boundary
- Never mix environments in the same state or workspace
- Prefer CI identities (OIDC) for backend access
- Enforce locking and least-privilege access to state storage

## Refactoring Safety (Terraform ≥ 1.5)

Resource Renames

Use moved blocks to prevent destroy/recreate:

```hcl
moved {
  from = azurerm_resource_group.old
  to   = azurerm_resource_group.main
}
```

Brownfield Adoption

Use import blocks for existing resources:

```hcl
import {
  to = azurerm_resource_group.main
  id = "/subscriptions/<sub>/resourceGroups/<rg>"
}
```

Removing Resources from State Without Destroying (Terraform ≥ 1.7)

Use removed blocks to stop managing a resource while keeping it in Azure:

```hcl
removed {
  from = azurerm_resource_group.legacy

  lifecycle {
    destroy = false
  }
}
```

**Rule:** Use `removed` blocks when transferring ownership of resources to another state or team. Complement to `moved` blocks.

Invariants

Use preconditions and postconditions to enforce safety:

```hcl
lifecycle {
  precondition {
    condition     = length(var.project) >= 3
    error_message = "Project name must be at least 3 characters."
  }
}
```

---

## Validation Pipeline (Local + CI)

**Minimum:**

- `terraform fmt -recursive`
- `terraform validate`

**Recommended:**

- `tflint`
- `terraform test`
- `terraform plan -out=tfplan`
- `terraform show -json tfplan > tfplan.json`

**Rules:**

- Plans must be reviewed before apply
- Apply must be gated for production environments
- Store plan artifacts for traceability

---

## Security Scanning (Strongly Recommended)

Use at least one static analysis tool:

- Trivy (`trivy config .`) — recommended; supersedes deprecated tfsec
- Checkov
- CIS Azure Foundations Benchmark
- Azure Security Benchmark

> **NOTE:** `tfsec` is deprecated and merged into Trivy. Migrate existing pipelines.

**Rules:**

- Fail builds on critical findings
- Document accepted risks explicitly

---

## Native Testing Framework (Terraform ≥ 1.6)

See [terraform-tests.instructions.md](terraform-tests.instructions.md) for full testing conventions (mock providers, plan/apply assertions, variable validation tests, security default tests, integration tests, parallel execution, CI pipeline integration).

---

## Ephemeral Resources and Write-Only Attributes (Terraform ≥ 1.10)

Ephemeral resources exist only during plan/apply and are **never persisted in state**. Write-only attributes (Terraform ≥ 1.11) prevent specific values from appearing in state.

### Ephemeral Variables

```hcl
variable "db_password" {
  type      = string
  ephemeral = true  # Not stored in state or plan files (Terraform ≥ 1.10)
}
```

### Sensitivity Comparison

| Attribute          | In State | In Logs  | Use Case                        |
| ------------------ | -------- | -------- | ------------------------------- |
| (none)             | Yes      | Yes      | Normal values                   |
| `sensitive = true` | Yes      | Redacted | Secrets (pre-1.10, still valid) |
| `ephemeral = true` | **No**   | **No**   | Secrets (1.10+, preferred)      |

### Write-Only Attributes (Terraform ≥ 1.11)

Some providers support write-only attributes (e.g., `password_wo`) that accept values but never store them in state:

```hcl
ephemeral "random_password" "db" {
  length = 16
}

resource "azurerm_postgresql_flexible_server" "main" {
  # ...
  administrator_password_wo         = ephemeral.random_password.db.result
  administrator_password_wo_version = 1
}
```

**Rule:** For Terraform ≥ 1.10, prefer `ephemeral = true` variables over `sensitive = true` for secrets. Use write-only attributes when the provider supports them to eliminate secrets from state entirely.

---

## Check Blocks — Continuous Validation (Terraform ≥ 1.5)

`check` blocks declare assertions that run on every plan/apply to validate infrastructure state:

```hcl
check "api_health" {
  data "http" "health" {
    url = "https://${azurerm_linux_web_app.main.default_hostname}/health"
  }

  assert {
    condition     = data.http.health.status_code == 200
    error_message = "API health endpoint is not responding."
  }
}
```

**Rule:** Use `check` blocks for post-deployment validation (health endpoints, DNS resolution, certificate expiry). Failures produce warnings, not errors — they don't block apply.

---

## Provider-Defined Functions (Terraform ≥ 1.8)

Providers can expose custom functions callable as `provider::provider_name::function_name()`. The AzureRM provider exposes functions for common operations:

```hcl
# Use provider functions instead of workaround patterns
locals {
  parsed = provider::azurerm::parse_resource_id(azurerm_storage_account.main.id)
}
```

**Rule:** Check provider documentation for available functions before implementing complex `regex()` or `split()` workarounds. Provider functions are type-safe and maintained by the provider team.

---

## Actionable Patterns

### Pattern 1: Provider Version Constraints (Version Pinning)

**❌ WRONG: No version constraint (unpredictable upgrades)**

```hcl
terraform {
  required_providers {
    azurerm = {
      source = "hashicorp/azurerm"  # ⚠️ No version = latest (breaking changes!)
    }
  }
}
```

**❌ WRONG: Overly broad constraint (allows major version changes)**

```hcl
terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">= 3.0"  # ⚠️ Could install v5.0 (breaking changes)
    }
  }
}
```

**✅ CORRECT: Pessimistic constraint (minor/patch updates only)**

```hcl
terraform {
  required_version = ">= 1.5.0"  # ✅ Minimum Terraform version

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.0"  # ✅ Allows 4.x (not 5.0)
    }
  }
}
```

**Rule:** Use `~> 4.0` for minor version updates. Pin exact version (`= 4.12.0`) for critical production. Commit `.terraform.lock.hcl` to lock transitive dependencies.

---

### Pattern 2: Remote State Management (Backend Configuration)

**❌ WRONG: Local state (no collaboration, no locking)**

```hcl
# No backend block - defaults to local file
terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.0"
    }
  }
}
# ⚠️ State stored in ./terraform.tfstate (single-user, no locking)
```

**✅ CORRECT: Remote backend with locking (collaborative)**

```hcl
terraform {
  backend "azurerm" {
    resource_group_name  = "terraform-state-rg"
    storage_account_name = "tfstatestorage"
    container_name       = "tfstate"
    key                  = "prod.terraform.tfstate"  # ✅ Isolated per environment
  }

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.0"
    }
  }
}
```

**Rule:** Always use remote backends (Azure Storage Account, Terraform Cloud). Enable state locking. Isolate state per environment (dev/staging/prod).

---

### Pattern 3: Variable Validation (Input Safety)

**❌ WRONG: No validation (runtime errors)**

```hcl
variable "environment" {
  type = string  # ⚠️ Any string accepted (typos, invalid values)
}

variable "project_name" {
  type = string  # ⚠️ No min length (could be empty)
}
```

**✅ CORRECT: Declarative validation (early failure)**

```hcl
variable "environment" {
  type = string

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod."
  }
}

variable "project_name" {
  type = string

  validation {
    condition     = length(var.project_name) >= 3 && length(var.project_name) <= 24
    error_message = "Project name must be 3-24 characters."
  }
}
```

**Rule:** Add `validation` blocks for enums, length constraints, regex patterns. Fail early in `terraform plan` before API calls.

---

### Pattern 4: Sensitive Data Handling (Secrets Management)

**❌ WRONG: Hardcoded secrets in code (Git history leak)**

```hcl
resource "azurerm_key_vault_secret" "db_password" {
  name         = "db-password"
  value        = "MySecretPassword123!"  # ⚠️ Visible in Git, state file, logs
  key_vault_id = azurerm_key_vault.main.id
}
```

**✅ CORRECT: Use sensitive variables with external sources**

```hcl
variable "db_password" {
  type      = string
  sensitive = true  # ✅ Redacted in logs
}

resource "azurerm_key_vault_secret" "db_password" {
  name         = "db-password"
  value        = var.db_password  # ✅ Passed via environment variable
  key_vault_id = azurerm_key_vault.main.id
}
```

**Usage:**

```bash
export TF_VAR_db_password="$(az keyvault secret show --name db-password --vault-name vault --query value -o tsv)"
terraform apply
```

**Rule:** Never hardcode secrets. Use `sensitive = true` for variables. Source from Azure Key Vault, environment variables, or Managed Identity.

---

### Pattern 5: Moved Blocks (Resource Renames Without Destroy)

**❌ WRONG: Renaming resource without moved block (destroys existing resource)**

```hcl
# Old code (v1)
resource "azurerm_resource_group" "old_name" {
  name     = "my-rg"
  location = "uksouth"
}

# Modified code (v2) - WITHOUT moved block
resource "azurerm_resource_group" "new_name" {  # ⚠️ Destroys old_name, creates new_name!
  name     = "my-rg"
  location = "uksouth"
}
```

**✅ CORRECT: Use moved block to preserve existing resource**

```hcl
# Modified code (v2) - WITH moved block
resource "azurerm_resource_group" "new_name" {
  name     = "my-rg"
  location = "uksouth"
}

moved {
  from = azurerm_resource_group.old_name
  to   = azurerm_resource_group.new_name  # ✅ State migration, no destroy
}
```

**Rule:** When renaming resources in Terraform ≥ 1.5, always add `moved` block. Prevents accidental destroy/recreate of production resources.

---

### Pattern 6: Import Blocks (Brownfield Adoption)

**❌ WRONG: Managing existing resources via `terraform import` CLI (manual, error-prone)**

```bash
# Manual command (not tracked in code)
terraform import azurerm_resource_group.main /subscriptions/<sub>/resourceGroups/<rg>
# ⚠️ Not reproducible, no version control
```

**✅ CORRECT: Use import blocks (declarative, versioned)**

```hcl
import {
  to = azurerm_resource_group.main
  id = "/subscriptions/<sub-id>/resourceGroups/my-rg"  # ✅ Tracked in code
}

resource "azurerm_resource_group" "main" {
  name     = "my-rg"
  location = "uksouth"
}
```

**Rule:** For Terraform ≥ 1.5, use `import` blocks instead of CLI. Enables reproducible brownfield adoption. Run `terraform plan -generate-config-out=generated.tf` to auto-generate resource definitions.

---

### Pattern 7: Dynamic Blocks (Optional Features)

**❌ WRONG: Duplicating resource definitions for variations (code bloat)**

```hcl
resource "azurerm_storage_account" "with_static_website" {
  name                     = "storagewithweb"
  resource_group_name      = azurerm_resource_group.main.name
  location                 = "uksouth"
  account_tier             = "Standard"
  account_replication_type = "LRS"

  static_website {
    index_document = "index.html"
  }
}

resource "azurerm_storage_account" "without_static_website" {  # ⚠️ Duplicated code
  name                     = "storagewithoutweb"
  resource_group_name      = azurerm_resource_group.main.name
  location                 = "uksouth"
  account_tier             = "Standard"
  account_replication_type = "LRS"
  # No static_website block
}
```

**✅ CORRECT: Use dynamic blocks with condition (DRY)**

```hcl
variable "enable_static_website" {
  type    = bool
  default = false
}

resource "azurerm_storage_account" "main" {
  name                     = "storage"
  resource_group_name      = azurerm_resource_group.main.name
  location                 = "uksouth"
  account_tier             = "Standard"
  account_replication_type = "LRS"

  dynamic "static_website" {
    for_each = var.enable_static_website ? [1] : []  # ✅ Conditionally included
    content {
      index_document = "index.html"
    }
  }
}
```

**Rule:** Use `dynamic` blocks with `for_each = var.enabled ? [1] : []` for optional nested configurations. Reduces code duplication.

---

### Pattern 8: Outputs with Sensitivity (Prevent Log Leakage)

**❌ WRONG: Outputting secrets without sensitive flag (logged plaintext)**

```hcl
output "database_password" {
  value = azurerm_key_vault_secret.db_password.value  # ⚠️ Printed in terraform apply
}
```

**✅ CORRECT: Mark outputs as sensitive (redacted in logs)**

```hcl
output "database_password" {
  value     = azurerm_key_vault_secret.db_password.value
  sensitive = true  # ✅ Redacted as "(sensitive value)" in logs
}
```

**Rule:** Always mark outputs containing secrets, passwords, connection strings as `sensitive = true`. Prevents accidental exposure in CI/CD logs.

---

## What NOT to Do

❌ Wrap AVM modules inside custom modules (violates AVM composition patterns)
❌ Hardcode credentials or secrets (use variable with `sensitive = true` or Key Vault)
❌ Enable public access by default (use `public_network_access_enabled = false`)
❌ Perform large refactors without `moved`/`import` blocks (destroys existing resources)
❌ Introduce breaking changes silently (use variable validation and lifecycle preconditions)
❌ Use local state for collaborative projects (always use remote backend with locking)

---

## References

Azure Verified Modules: https://azure.github.io/Azure-Verified-Modules/
Terraform Documentation: https://developer.hashicorp.com/terraform
AzureRM Provider Docs: https://registry.terraform.io/providers/hashicorp/azurerm/latest
GitHub Copilot Instructions: https://docs.github.com/copilot/customizing-copilot/adding-custom-instructions-for-github-copilot
