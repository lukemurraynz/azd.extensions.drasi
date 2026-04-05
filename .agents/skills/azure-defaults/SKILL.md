---
name: azure-defaults
description: >-
  Provide Azure infrastructure defaults for naming conventions, regions, tags,
  security baselines, WAF criteria, and pricing guidance.
  USE FOR: looking up Azure naming standards, choosing regions, setting resource
  tags, applying security baselines, or verifying governance defaults before deployment.
compatibility: Works with Claude Code, GitHub Copilot, VS Code, and any Agent Skills compatible tool.
license: MIT
metadata:
  author: jonathan-vella
  version: "1.0"
  category: azure-infrastructure
---

# Azure Defaults Skill

> **This skill is MANDATORY for every Azure deployment.** All agents MUST load this
> skill before generating infrastructure code. Non-compliant outputs MUST be rejected.

Single source of truth for all Azure infrastructure configuration used across agents.
Replaces individual `_shared/` file lookups with one consolidated reference.

---

## Quick Reference (Load First)

### Default Regions

| Service             | Default Region       | Reason                                     |
| ------------------- | -------------------- | ------------------------------------------ |
| **All resources**   | `australiaeast`      | Primary region — closest to NZ/AU users    |
| **Static Web Apps** | `eastasia`           | Nearest SWA-supported region to NZ/AU      |
| **Azure OpenAI**    | `australiaeast`      | Limited availability — verify region first |
| **Failover**        | `australiasoutheast` | AU geo-paired region                       |

### Required Tags (Azure Policy Enforced)

> [!IMPORTANT]
> These 3 tags MUST appear on every deployed resource. Azure Policy will reject
> deployments missing any of them. Always defer to `04-governance-constraints.md`
> for additional subscription-specific tag requirements.

| Tag                  | Required | Example Value             | Description                         |
| -------------------- | -------- | ------------------------- | ----------------------------------- |
| `displayName`        | Yes      | `My Application`          | Human-readable application name     |
| `locationIdentifier` | Yes      | `az.public.australiaeast` | Cloud provider, cloud, and region   |
| `cloud`              | Yes      | `public`                  | Cloud environment (`public`, `gov`) |

Bicep pattern:

```bicep
tags: {
  displayName: displayName
  locationIdentifier: 'az.${cloud}.${location}'
  cloud: cloud
}
```

### Unique Suffix Pattern

Generate ONCE in `main.bicep`, pass to ALL modules:

```bicep
// main.bicep
var uniqueSuffix = uniqueString(resourceGroup().id)

module keyVault 'modules/key-vault.bicep' = {
  params: { uniqueSuffix: uniqueSuffix }
}
```

### Security Baseline

> [!NOTE]
> Security baselines in this skill apply defaults suitable for development/testing environments. For production hardening (private endpoints, disabled public access, diagnostic logging), consult the relevant service-specific skills (e.g., `private-networking`, `secret-management`, `observability-monitoring`).

| Setting                    | Value               | Applies To                        |
| -------------------------- | ------------------- | --------------------------------- |
| `supportsHttpsTrafficOnly` | `true`              | Storage accounts                  |
| `minimumTlsVersion`        | `'TLS1_2'`          | All services                      |
| `allowBlobPublicAccess`    | `false`             | Storage accounts                  |
| `publicNetworkAccess`      | `'Disabled'` (prod) | Data services                     |
| Authentication             | Managed Identity    | Prefer over keys/strings          |
| SQL Auth                   | Azure AD-only       | `azureADOnlyAuthentication: true` |

---

### AKS Baseline Defaults (The AKS Book)

| Decision Area      | Default for Future Projects                                                                 |
| ------------------ | -------------------------------------------------------------------------------------------- |
| AKS Tier           | **Automatic** when its opinionated config fits; **Standard** for Windows or custom networking |
| Networking         | **Azure CNI Overlay**; avoid kubenet for new clusters                                        |
| Identity           | **System-assigned MI** for cluster; **Workload Identity** for pods                           |
| Availability Zones | **Zones 1/2/3** for control plane (including dev/test)                                       |
| Node Pools         | **Separate system + user pools**; system pool ~3 nodes, user pool autoscale/NAP             |
| Node OS            | **Azure Linux 3.0**; Ubuntu 24.04 LTS when Azure Linux compatibility is a concern           |
| Versioning         | **N-1** Kubernetes; **SecurityPatch** OS channel for prod; planned maintenance windows      |
| Monitoring         | **Managed Prometheus + Container Insights**; pre-create workspaces; **ContainerLogV2**      |
| Resource Groups    | **One cluster per RG**; enable **node resource group lockdown**                             |

---

## CAF Naming Conventions

### Standard Abbreviations

| Resource         | Abbreviation | Name Pattern                | Max Length |
| ---------------- | ------------ | --------------------------- | ---------- |
| Resource Group   | `rg`         | `rg-{project}-{env}`        | 90         |
| Virtual Network  | `vnet`       | `vnet-{project}-{env}`      | 64         |
| Subnet           | `snet`       | `snet-{purpose}-{env}`      | 80         |
| NSG              | `nsg`        | `nsg-{purpose}-{env}`       | 80         |
| Key Vault        | `kv`         | `kv-{short}-{env}-{suffix}` | **24**     |
| Storage Account  | `st`         | `st{short}{env}{suffix}`    | **24**     |
| App Service Plan | `asp`        | `asp-{project}-{env}`       | 40         |
| App Service      | `app`        | `app-{project}-{env}`       | 60         |
| SQL Server       | `sql`        | `sql-{project}-{env}`       | 63         |
| SQL Database     | `sqldb`      | `sqldb-{project}-{env}`     | 128        |
| Static Web App   | `stapp`      | `stapp-{project}-{env}`     | 40         |
| CDN / Front Door | `fd`         | `fd-{project}-{env}`        | 64         |
| Log Analytics    | `log`        | `log-{project}-{env}`       | 63         |
| App Insights     | `appi`       | `appi-{project}-{env}`      | 255        |
| Container App    | `ca`         | `ca-{project}-{env}`        | 32         |
| Container Env    | `cae`        | `cae-{project}-{env}`       | 60         |
| Cosmos DB        | `cosmos`     | `cosmos-{project}-{env}`    | 44         |
| Service Bus      | `sb`         | `sb-{project}-{env}`        | 50         |

### Length-Constrained Resources

Key Vault and Storage Account have 24-char limits. Always include `uniqueSuffix`:

```bicep
// Key Vault: kv-{8chars}-{3chars}-{6chars} = 21 chars max
var kvName = 'kv-${take(projectName, 8)}-${take(environment, 3)}-${take(uniqueSuffix, 6)}'

// Storage: st{8chars}{3chars}{6chars} = 19 chars max (no hyphens!)
var stName = 'st${take(replace(projectName, '-', ''), 8)}${take(environment, 3)}${take(uniqueSuffix, 6)}'
```

### Naming Rules

- **DO**: Use lowercase with hyphens (`kv-myapp-dev-abc123`)
- **DO**: Include `uniqueSuffix` in globally unique names (Key Vault, Storage, SQL Server)
- **DO**: Use `take()` to truncate long names within limits
- **DON'T**: Use hyphens in Storage Account names (only lowercase + numbers)
- **DON'T**: Hardcode unique values — always derive from `uniqueString(resourceGroup().id)`
- **DON'T**: Exceed max length — Bicep won't warn, deployment will fail

---

## Azure Verified Modules (AVM)

### AVM-First Policy

1. **ALWAYS** check AVM availability first via `mcp_bicep_list_avm_metadata`
2. Use AVM module defaults for SKUs when available
3. If custom SKU needed, require live deprecation research
4. **NEVER** hardcode SKUs without validation
5. **NEVER** write raw Bicep for a resource that has an AVM module

### Common AVM Modules

| Resource           | Module Path                                        | Min Version |
| ------------------ | -------------------------------------------------- | ----------- |
| Key Vault          | `br/public:avm/res/key-vault/vault`                | `0.13`      |
| Virtual Network    | `br/public:avm/res/network/virtual-network`        | `0.7`       |
| Storage Account    | `br/public:avm/res/storage/storage-account`        | `0.32`      |
| App Service Plan   | `br/public:avm/res/web/serverfarm`                 | `0.7`       |
| App Service        | `br/public:avm/res/web/site`                       | `0.22`      |
| SQL Server         | `br/public:avm/res/sql/server`                     | `0.21`      |
| Log Analytics      | `br/public:avm/res/operational-insights/workspace` | `0.15`      |
| App Insights       | `br/public:avm/res/insights/component`             | `0.7`       |
| NSG                | `br/public:avm/res/network/network-security-group` | `0.5.0`     |
| Static Web App     | `br/public:avm/res/web/static-site`                | `0.9`       |
| Container App      | `br/public:avm/res/app/container-app`              | `0.21`      |
| Container Env      | `br/public:avm/res/app/managed-environment`        | `0.13`      |
| Cosmos DB          | `br/public:avm/res/document-db/database-account`   | `0.19`      |
| Front Door         | `br/public:avm/res/cdn/profile`                    | `0.19`      |
| Service Bus        | `br/public:avm/res/service-bus/namespace`          | `0.16`      |
| Container Registry | `br/public:avm/res/container-registry/registry`    | `0.12`      |

### Finding Latest AVM Version

```text
// Use Bicep MCP tool:
mcp_bicep_list_avm_metadata → filter by resource type → use latest version

// Or check: https://aka.ms/avm/index
```

### AVM Usage Pattern

```bicep
module keyVault 'br/public:avm/res/key-vault/vault:0.13' = {
  name: '${kvName}-deploy'
  params: {
    name: kvName
    location: location
    tags: tags
    enableRbacAuthorization: true
    enablePurgeProtection: true
  }
}
```

---

## AVM Known Pitfalls

### Region Limitations

| Service         | Limitation                                                                  | Workaround                                |
| --------------- | --------------------------------------------------------------------------- | ----------------------------------------- |
| Static Web Apps | Only 5 regions: `westus2`, `centralus`, `eastus2`, `westeurope`, `eastasia` | Use `eastasia` for NZ/AU proximity        |
| Azure OpenAI    | Limited regions per model                                                   | Check availability before planning        |
| Container Apps  | Most regions but not all                                                    | Verify `cae` environment in target region |

### Parameter Type Mismatches

Known issues when using AVM modules — verify before coding:

**Log Analytics Workspace** (`operational-insights/workspace`):

- `dailyQuotaGb` is `int` in AVM, not `string`
- **DO**: `dailyQuotaGb: 5`
- **DON'T**: `dailyQuotaGb: '5'`

**Container Apps Managed Environment** (`app/managed-environment`):

- `appLogsConfiguration` deprecated in newer versions
- **DO**: Use `logsConfiguration` with destination object
- **DON'T**: Use `appLogsConfiguration.destination: 'log-analytics'`

**Container Apps** (`app/container-app`):

- `scaleSettings` is an object, not array of rules
- **DO**: Check AVM schema for exact object shape
- **DON'T**: Assume `scaleRules: [...]` array format

**SQL Server** (`sql/server`):

- `sku` parameter is a typed object `{name, tier, capacity}`
- **DO**: Pass full SKU object matching schema
- **DON'T**: Pass just string `'S0'`
- `availabilityZone` requires specific format per region

**App Service** (`web/site`):

- `APPINSIGHTS_INSTRUMENTATIONKEY` deprecated
- **DO**: Use `APPLICATIONINSIGHTS_CONNECTION_STRING` instead
- **DON'T**: Set instrumentation key directly

**Key Vault** (`key-vault/vault`):

- `softDeleteRetentionInDays` is immutable after creation
- **DO**: Set correctly on first deploy (default: 90)
- **DON'T**: Try to change after vault exists

**Static Web App** (`web/static-site`):

- Free SKU may not be deployable via ARM in all regions
- **DO**: Use `Standard` SKU for reliable ARM deployment
- **DON'T**: Assume Free tier works everywhere via Bicep

---

## Terraform Conventions

### AVM-TF Registry Lookup

Find the latest AVM-TF module version before generating code:

```text
// Use Terraform MCP tool:
mcp_terraform_get_latest_module_version → registry.terraform.io/modules/Azure/{module}/azurerm

// Or browse: https://registry.terraform.io/modules/Azure
```

### Tag Syntax (HCL)

```hcl
# locals.tf — merge baseline tags with caller-supplied extras
locals {
  tags = merge(var.tags, {
    displayName        = var.display_name
    locationIdentifier = "az.${var.cloud}.${var.location}"
    cloud              = var.cloud
  })
}
```

### Required Commands

```bash
# Format all .tf files before committing
terraform fmt -recursive

# Validate syntax and provider schema
terraform validate

# Preview changes before applying
terraform plan -out=plan.tfplan
```

### State Backend

Use Azure Storage Account for all remote state. **Never** use HCP Terraform Cloud:

```hcl
# backend.tf
terraform {
  backend "azurerm" {
    resource_group_name  = "rg-tfstate-prod"
    storage_account_name = "sttfstate{suffix}"
    container_name       = "tfstate"
    key                  = "{project}.terraform.tfstate"
  }
}
```

### Unique Suffix

Generate once per root module, pass to all child modules:

```hcl
resource "random_string" "suffix" {
  length  = 4
  lower   = true
  numeric = true
  special = false
}
```

---

## Common AVM-TF Modules

| Resource               | Bicep AVM                                                   | Terraform AVM                                                |
| ---------------------- | ----------------------------------------------------------- | ------------------------------------------------------------ |
| Key Vault              | `br/public:avm/res/key-vault/vault`                         | `Azure/avm-res-keyvault-vault/azurerm`                       |
| Storage Account        | `br/public:avm/res/storage/storage-account`                 | `Azure/avm-res-storage-storageaccount/azurerm`               |
| Virtual Network        | `br/public:avm/res/network/virtual-network`                 | `Azure/avm-res-network-virtualnetwork/azurerm`               |
| App Service Plan       | `br/public:avm/res/web/serverfarm`                          | `Azure/avm-res-web-serverfarm/azurerm`                       |
| Web App                | `br/public:avm/res/web/site`                                | `Azure/avm-res-web-site/azurerm`                             |
| Container Registry     | `br/public:avm/res/container-registry/registry`             | `Azure/avm-res-containerregistry-registry/azurerm`           |
| AKS                    | `br/public:avm/res/container-service/managed-cluster`       | `Azure/avm-res-containerservice-managedcluster/azurerm`      |
| SQL Database           | `br/public:avm/res/sql/server`                              | `Azure/avm-res-sql-server/azurerm`                           |
| Cosmos DB              | `br/public:avm/res/document-db/database-account`            | `Azure/avm-res-documentdb-databaseaccount/azurerm`           |
| Service Bus            | `br/public:avm/res/service-bus/namespace`                   | `Azure/avm-res-servicebus-namespace/azurerm`                 |
| Event Hub              | `br/public:avm/res/event-hub/namespace`                     | `Azure/avm-res-eventhub-namespace/azurerm`                   |
| Log Analytics          | `br/public:avm/res/operational-insights/workspace`          | `Azure/avm-res-operationalinsights-workspace/azurerm`        |
| App Insights           | `br/public:avm/res/insights/component`                      | `Azure/avm-res-insights-component/azurerm`                   |
| Private DNS Zone       | `br/public:avm/res/network/private-dns-zone`                | `Azure/avm-res-network-privatednszones/azurerm`              |
| User-Assigned Identity | `br/public:avm/res/managed-identity/user-assigned-identity` | `Azure/avm-res-managedidentity-userassignedidentity/azurerm` |
| API Management         | `br/public:avm/res/api-management/service`                  | `Azure/avm-res-apimanagement-service/azurerm`                |

---

## WAF Assessment Criteria

### Scoring Scale

| Score | Definition                                  |
| ----- | ------------------------------------------- |
| 9-10  | Exceeds best practices, production-ready    |
| 7-8   | Meets best practices with minor gaps        |
| 5-6   | Adequate but improvements needed            |
| 3-4   | Significant gaps, address before production |
| 1-2   | Critical deficiencies, not production-ready |

### Pillar Definitions

| Pillar      | Icon | Focus Areas                                              |
| ----------- | ---- | -------------------------------------------------------- |
| Security    | 🔒   | Identity, network, data protection, threat detection     |
| Reliability | 🔄   | SLA, redundancy, disaster recovery, health monitoring    |
| Performance | ⚡   | Response time, scalability, caching, load testing        |
| Cost        | 💰   | Right-sizing, reserved instances, monitoring spend       |
| Operations  | 🔧   | IaC, CI/CD, monitoring, incident response, documentation |

### Assessment Rules

- **DO**: Score each pillar 1-10 with confidence level (High/Medium/Low)
- **DO**: Identify specific gaps with remediation recommendations
- **DO**: Calculate composite WAF score as average of all pillars
- **DON'T**: Give perfect 10/10 scores without exceptional justification
- **DON'T**: Skip any pillar even if requirements seem light
- **DON'T**: Provide generic recommendations — be specific to the workload

---

## Azure Pricing MCP Service Names

Exact names for the Azure Pricing MCP tool. Using wrong names returns 0 results.

| Azure Service       | Correct `service_name`          | Common SKUs                                |
| ------------------- | ------------------------------- | ------------------------------------------ |
| AKS                 | `Azure Kubernetes Service`      | `Free`, `Standard`, `Premium`              |
| API Management      | `API Management`                | `Consumption`, `Developer`, `Standard`     |
| App Insights        | `Application Insights`          | `Enterprise`, `Basic`                      |
| App Service         | `Azure App Service`             | `B1`, `S1`, `P1v3`, `P1v4`                 |
| Application Gateway | `Application Gateway`           | `Standard_v2`, `WAF_v2`                    |
| Azure Bastion       | `Azure Bastion`                 | `Basic`, `Standard`                        |
| Azure DNS           | `Azure DNS`                     | `Public`, `Private`                        |
| Azure Firewall      | `Azure Firewall`                | `Standard`, `Premium`                      |
| Azure Functions     | `Functions`                     | `Consumption`, `Premium`                   |
| Azure Monitor       | `Azure Monitor`                 | `Logs`, `Metrics`                          |
| Container Apps      | `Azure Container Apps`          | `Consumption`                              |
| Container Instances | `Container Instances`           | `Standard`                                 |
| Container Registry  | `Container Registry`            | `Basic`, `Standard`, `Premium`             |
| Cosmos DB           | `Azure Cosmos DB`               | `Serverless`, `Provisioned`                |
| Data Factory        | `Azure Data Factory v2`         | `Data Flow`, `Pipeline`                    |
| Event Grid          | `Event Grid`                    | `Basic`                                    |
| Event Hubs          | `Event Hubs`                    | `Basic`, `Standard`, `Premium`             |
| Front Door          | `Azure Front Door`              | `Standard`, `Premium`                      |
| Key Vault           | `Key Vault`                     | `Standard`                                 |
| Load Balancer       | `Load Balancer`                 | `Basic`, `Standard`                        |
| Log Analytics       | `Log Analytics`                 | `Per GB`, `Commitment Tier`                |
| Logic Apps          | `Logic Apps`                    | `Consumption`, `Standard`                  |
| MySQL Flexible      | `Azure Database for MySQL`      | `B1ms`, `D2ds_v4`, `E2ds_v4`               |
| NAT Gateway         | `NAT Gateway`                   | `Standard`                                 |
| PostgreSQL Flexible | `Azure Database for PostgreSQL` | `B1ms`, `D2ds_v4`, `E2ds_v4`               |
| Redis Cache         | `Azure Cache for Redis`         | `Basic`, `Standard`, `Premium`             |
| SQL Database        | `SQL Database`                  | `Basic`, `Standard`, `S0`, `S1`, `Premium` |
| Service Bus         | `Service Bus`                   | `Basic`, `Standard`, `Premium`             |
| Static Web Apps     | `Azure Static Web Apps`         | `Free`, `Standard`                         |
| Storage             | `Storage`                       | `Standard`, `Premium`, `LRS`, `GRS`        |
| VPN Gateway         | `VPN Gateway`                   | `Basic`, `VpnGw1`, `VpnGw2`                |
| Virtual Machines    | `Virtual Machines`              | `D4s_v5`, `B2s`, `E4s_v5`                  |

- **DO**: Use exact names from the table above
- **DON'T**: Use "Azure SQL" (returns 0 results) — use "SQL Database"
- **DON'T**: Use "Web App" — use "Azure App Service"

### Bulk Estimates

For multi-resource cost estimates, prefer `azure_bulk_estimate` over calling `azure_cost_estimate`
per resource. It accepts a `resources` array and returns aggregated totals.

Each resource supports a `quantity` parameter (default: 1) for multi-instance scenarios.
Use `output_format: "compact"` to reduce response size when detailed metadata is not needed.

---

## Service Recommendation Matrix

### Workload Patterns

| Pattern           | Cost-Optimized Tier        | Balanced Tier                    | Enterprise Tier                         |
| ----------------- | -------------------------- | -------------------------------- | --------------------------------------- |
| **Static Site**   | SWA Free + Blob            | SWA Std + CDN + KV               | SWA Std + FD + KV + Monitor             |
| **API-First**     | App Svc B1 + SQL Basic     | App Svc S1 + SQL S1 + KV         | App Svc P1v3 + SQL Premium + APIM       |
| **N-Tier Web**    | App Svc B1 + SQL Basic     | App Svc S1 + SQL S1 + Redis + KV | App Svc P1v4 + SQL Premium + Redis + FD |
| **Serverless**    | Functions Consumption      | Functions Premium + CosmosDB     | Functions Premium + CosmosDB + APIM     |
| **Container**     | Container Apps Consumption | Container Apps + ACR + KV        | AKS + ACR + KV + Monitor                |
| **Data Platform** | SQL Basic + Blob           | Synapse Serverless + ADLS        | Synapse Dedicated + ADLS + Purview      |

### Detection Signals

Map user language to workload pattern:

| User Says                              | Likely Pattern |
| -------------------------------------- | -------------- |
| "website", "landing page", "blog"      | Static Site    |
| "REST API", "microservices", "backend" | API-First      |
| "web app", "portal", "dashboard"       | N-Tier Web     |
| "event-driven", "triggers", "webhooks" | Serverless     |
| "Docker", "Kubernetes", "containers"   | Container      |
| "analytics", "data warehouse", "ETL"   | Data Platform  |

### Business Domain Signals

| Industry          | Common Compliance | Default Security                      |
| ----------------- | ----------------- | ------------------------------------- |
| Healthcare        | HIPAA             | Private endpoints, encryption at rest |
| Financial         | PCI-DSS, SOC 2    | WAF, private endpoints, audit logging |
| Government        | FedRAMP, IL4/5    | Azure Gov, private endpoints          |
| Retail/E-commerce | PCI-DSS           | WAF, DDoS protection                  |
| Education         | FERPA             | Data residency, access controls       |

### Company Size Heuristics

| Size                | Budget Signal  | Default Tier   | Security Posture       |
| ------------------- | -------------- | -------------- | ---------------------- |
| Startup (<50)       | "$50-200/mo"   | Cost-Optimized | Basic managed identity |
| Mid-Market (50-500) | "$500-2000/mo" | Balanced       | Private endpoints, KV  |
| Enterprise (500+)   | "$2000+/mo"    | Enterprise     | Full WAF compliance    |

### Industry Compliance Pre-Selection

| Industry   | Auto-Select                       |
| ---------- | --------------------------------- |
| Healthcare | HIPAA checkbox, private endpoints |
| Finance    | PCI-DSS + SOC 2, WAF required     |
| Government | Data residency, enhanced audit    |
| Retail     | PCI-DSS if payments, DDoS         |

---

## Governance Discovery

### MANDATORY Gate

Governance discovery is a **hard gate**. If Azure connectivity is unavailable or policies cannot
be fully retrieved (including management group-inherited), STOP and inform the user.
Do NOT proceed to implementation planning with incomplete policy data.

### Discovery Commands (Ordered by Completeness)

**1. REST API (MANDATORY — includes management group-inherited policies)**:

```bash
SUB_ID=$(az account show --query id -o tsv)
az rest --method GET \
  --url "https://management.azure.com/subscriptions/\
${SUB_ID}/providers/Microsoft.Authorization/\
policyAssignments?api-version=2025-11-01" \
  --query "value[].{name:name, \
displayName:properties.displayName, \
scope:properties.scope, \
enforcementMode:properties.enforcementMode, \
policyDefinitionId:properties.policyDefinitionId}" \
  -o json
```

> [!CAUTION]
> `az policy assignment list` only returns subscription-scoped assignments.
> Management group policies (often Deny/tag enforcement) are invisible to it.
> **ALWAYS use the REST API above as the primary discovery method.**

**2. Policy Definition Drill-Down (for each Deny/DeployIfNotExists)**:

```bash
# For built-in or subscription-scoped policies
az policy definition show --name "{guid}" \
  --query "{displayName:displayName, \
effect:policyRule.then.effect, \
conditions:policyRule.if}" -o json

# For management-group-scoped custom policies
az policy definition show --name "{guid}" \
  --management-group "{mgId}" \
  --query "{displayName:displayName, \
effect:policyRule.then.effect}" -o json

# For policy set definitions (initiatives)
az policy set-definition show --name "{guid}" \
  --query "{displayName:displayName, \
policyCount:policyDefinitions | length(@)}" -o json
```

**3. ARG KQL (supplemental — subscription-scoped only)**:

```kusto
PolicyResources
| where type == 'microsoft.authorization/policyassignments'
| where properties.enforcementMode == 'Default'
| project name, displayName=properties.displayName,
  effect=properties.parameters.effect.value,
  scope=properties.scope
| order by name asc
```

### Azure Policy Discovery Workflow

Before creating implementation plans, discover active policies:

```text
1. Verify Azure connectivity: az account show
2. REST API: Get ALL effective policy assignments (subscription + MG inherited)
3. Compare count with Azure Portal (Policy > Assignments) — must match
4. For each Deny/DeployIfNotExists: drill into policy definition JSON
5. Check tag enforcement policies (names containing 'tag' or 'Tag')
6. Check allowed resource types and locations
7. Document ALL findings in 04-governance-constraints.md
```

### Common Policy Constraints

> [!NOTE]
> The governance constraints JSON output schema must include `bicepPropertyPath`,
> `azurePropertyPath`, and `requiredValue` fields for each Deny policy to enable
> downstream programmatic consumption by the Code Generator and review subagent.
> `azurePropertyPath` follows the Azure REST API resource property path (dot-separated,
> resource type camelCase first) and enables IaC-tool-agnostic enforcement.

| Policy             | Impact                          | Solution                              |
| ------------------ | ------------------------------- | ------------------------------------- |
| Required tags      | Deployment fails without tags   | Include all 3 required tags           |
| Allowed locations  | Resources rejected outside list | Use `australiaeast` default           |
| SQL AAD-only auth  | SQL password auth blocked       | Use `azureADOnlyAuthentication: true` |
| Storage shared key | Shared key access denied        | Use managed identity RBAC             |
| Zone redundancy    | Non-zonal SKUs rejected         | Use P1v4+ for App Service Plans       |

---

## Research Workflow (All Agents)

### Standard 4-Step Pattern

1. **Validate Prerequisites** — Confirm previous artifact exists. If missing, STOP.
2. **Read Agent Context** — Read previous artifact for context. Read template for H2 structure.
3. **Domain-Specific Research** — Query ONLY for NEW information not in artifacts.
4. **Confidence Gate (80% Rule)** — Proceed at 80%+ confidence. Below 80%, ASK user.

### Confidence Levels

| Level           | Indicators                  | Action                                      |
| --------------- | --------------------------- | ------------------------------------------- |
| High (80-100%)  | All critical info available | Proceed                                     |
| Medium (60-79%) | Some assumptions needed     | Document assumptions, ask for critical gaps |
| Low (0-59%)     | Major gaps                  | STOP — request clarification                |

### Context Reuse Rules

- **DO**: Read previous agent's artifact for context
- **DO**: Cache shared defaults (read once per session)
- **DO**: Query external sources only for NEW information
- **DON'T**: Re-query Azure docs for resources already in artifacts
- **DON'T**: Search workspace repeatedly (context flows via artifacts)
- **DON'T**: Re-validate previous agent's work (trust artifact chain)

### Agent-Specific Research Focus

| Agent        | Primary Research                      | Skip (Already in Artifacts)                                                                                                                                                                                   |
| ------------ | ------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Requirements | User needs, business context          | —                                                                                                                                                                                                             |
| Architect    | WAF gaps, SKU comparisons, pricing    | Service list (from 01)                                                                                                                                                                                        |
| Bicep Plan   | AVM availability, governance policies | Architecture decisions (from 02)                                                                                                                                                                              |
| Bicep Code   | AVM schemas, parameter types          | Resource list (from 04). NOTE: Governance constraints from `04-governance-constraints.md` MUST still be read and enforced — "trust artifact chain" means accepting decisions, not skipping compliance checks. |
| Deploy       | Azure state (what-if), credentials    | Template structure (from 05)                                                                                                                                                                                  |

---

## Service Lifecycle Validation

### AVM Default Trust

When using AVM modules with default SKU parameters:

- Trust the AVM default — Microsoft maintains these
- No additional deprecation research needed for defaults
- If overriding SKU parameter, run deprecation research

### Deprecation Research (For Non-AVM or Custom SKUs)

| Source            | Query Pattern                                              | Reliability |
| ----------------- | ---------------------------------------------------------- | ----------- |
| Azure Updates     | `azure.microsoft.com/updates/?query={service}+deprecated`  | High        |
| Microsoft Learn   | Check "Important" / "Note" callouts on service pages       | High        |
| Azure CLI         | `az provider show --namespace {provider}` for API versions | Medium      |
| Resource Provider | Check available SKUs in target region                      | High        |

### Known Deprecation Patterns

| Pattern                    | Status            | Replacement           |
| -------------------------- | ----------------- | --------------------- |
| "Classic" anything         | DEPRECATED        | ARM equivalents       |
| CDN `Standard_Microsoft`   | DEPRECATED 2027   | Azure Front Door      |
| App Gateway v1             | DEPRECATED        | App Gateway v2        |
| "v1" suffix services       | Likely deprecated | Check for v2          |
| Old API versions (2020-xx) | Outdated          | Use latest stable API |

### What-If Deprecation Signals

Deploy agent should scan what-if output for:
`deprecated|sunset|end.of.life|no.longer.supported|classic.*not.*supported|retiring`

If detected, STOP and report before deployment.

---

## Template-First Output Rules

### Mandatory Compliance

| Rule         | Requirement                                            |
| ------------ | ------------------------------------------------------ |
| Exact text   | Use template H2 text verbatim                          |
| Exact order  | Required H2s appear in template-defined order          |
| Anchor rule  | Extra sections allowed only AFTER last required H2     |
| No omissions | All template H2s must appear in output                 |
| Attribution  | Include `> Generated by {agent} agent \| {YYYY-MM-DD}` |

### Output Location

All agent outputs go to `agent-output/{project}/`:

| Step | Output File                      | Agent                   |
| ---- | -------------------------------- | ----------------------- |
| 1    | `01-requirements.md`             | Requirements            |
| 2    | `02-architecture-assessment.md`  | Architect               |
| 3    | `03-des-*.{py,md}`               | Design                  |
| 4    | `04-implementation-plan.md`      | Bicep Plan              |
| 4    | `04-governance-constraints.md`   | Bicep Plan              |
| 4    | `04-preflight-check.md`          | Bicep Code (pre-flight) |
| 5    | `05-implementation-reference.md` | Bicep Code              |
| 6    | `06-deployment-summary.md`       | Deploy                  |
| 7    | `07-*.md` (7 documents)          | azure-artifacts skill   |

### Header Format

```markdown
# Step {N}: {Title} - {project-name}

> Generated by {agent} agent | {YYYY-MM-DD}
```

---

## Validation Checklist

Before completing any agent task, verify:

- [ ] Output file saved to `agent-output/{project}/`
- [ ] All required H2 headings from template are present
- [ ] H2 headings match template text exactly
- [ ] All 3 required tags included (`displayName`, `locationIdentifier`, `cloud`)
- [ ] Unique suffix used for globally unique names
- [ ] Security baseline settings applied
- [ ] Region defaults correct (`australiaeast`, or exception documented)
- [ ] Attribution header included with agent name and date

---

## Related Skills

- [Azure Role Selector](../azure-role-selector/SKILL.md) — RBAC role assignments and managed identity
- [Azure Deployment Preflight](../azure-deployment-preflight/SKILL.md) — Pre-deployment validation
- [Cost Optimization](../cost-optimization/SKILL.md) — SKU sizing and spend controls
- [Azure Troubleshooting](../azure-troubleshooting/SKILL.md) — Diagnosing deployed resource issues
- [AKS Cluster Architecture](../aks-cluster-architecture/SKILL.md) — AKS-specific defaults and architecture decisions

### MCP Tooling

- **`drawio`** — Generate architecture reference diagrams via the diagram-smith agent

---

## Currency and Verification

- **Date checked:** 2026-03-31
- **Compatibility:** Azure CLI, Bicep, ARM templates
- **Sources:** [Azure naming conventions](https://learn.microsoft.com/azure/cloud-adoption-framework/ready/azure-best-practices/resource-naming), [Azure regions](https://learn.microsoft.com/azure/reliability/availability-zones-overview), [Azure resource tags](https://learn.microsoft.com/azure/azure-resource-manager/management/tag-resources)
- **Verification steps:**
  1. Verify region availability: `az account list-locations --query "[].name" -o tsv`
  2. Check naming rules per resource type: `az provider show --namespace Microsoft.Storage --query "resourceTypes[?resourceType=='storageAccounts'].{name:resourceType}"` (repeat per resource type)
  3. Verify tag policies: `az policy assignment list --query "[?contains(displayName, 'tag')]"`
