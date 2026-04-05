---
applyTo: "**/*.bicep,**/*.bicepparam,**/*.tf,**/*.tfvars,**/k8s/**/*.yml,**/k8s/**/*.yaml,**/kubernetes/**/*.yml,**/kubernetes/**/*.yaml"
description: "Data sovereignty, Azure region residency, GDPR data boundary, and geographic compliance guardrails for IaC and Kubernetes resources"
---

# Data Sovereignty and Regional Compliance

Use this instruction when designing, reviewing, or creating infrastructure code that stores or processes data subject to regional compliance requirements (GDPR, UK GDPR, sovereignty mandates).

**IMPORTANT**: Use `microsoft.learn.mcp` to verify current Azure region availability and compliance documentation. Do not assume regional support — verify.

## Core Principle

Data locality decisions are **permanent** (or very expensive to reverse). Establish region strategy before creating resources and codify it as code. Region selection is classified as ❌ **Permanent (cluster/resource rebuild required)** — see the Kubernetes instructions for the decision permanence rule.

---

## Azure Region Selection

### Classify data before choosing a region

| Classification | Requirement | Recommended Azure Regions |
|---|---|---|
| EU personal data (GDPR) | Must not leave EEA without an adequacy decision or SCCs | northeurope, westeurope, swedencentral, germanywestcentral |
| UK personal data (UK GDPR) | Prefer UK regions; EEA currently adequate | uksouth, ukwest |
| Sovereign / restricted | Must stay in-country; use dedicated Azure offerings | Azure Government (US), Azure operated by 21Vianet (China) |
| Non-personal operational | No geographic restriction | Any paired region closest to workload |

### Bicep / ARM guardrails

```bicep
// ✅ Parameterise and validate location before resources are created
@description('Azure region. Must satisfy data residency requirements for this workload.')
@allowed(['uksouth', 'ukwest', 'northeurope', 'westeurope', 'swedencentral'])
param location string = 'uksouth'

// ✅ Declare paired region explicitly for geo-redundant storage
// Document why — the choice affects GDPR boundary
var pairedLocation = location == 'uksouth' ? 'ukwest' : (location == 'northeurope' ? 'westeurope' : 'northeurope')

// ❌ Never hard-code to a region that could violate data residency
// param location string = 'eastus'  // Breaks EU/UK data residency
```

### Terraform guardrails

```hcl
variable "location" {
  type        = string
  description = "Azure region. Must satisfy data residency requirements for this workload."
  validation {
    condition     = contains(["uksouth", "ukwest", "northeurope", "westeurope", "swedencentral"], var.location)
    error_message = "Location must be within the approved EU/UK residency boundary."
  }
}
```

---

## Cross-Region Replication

Always verify that the paired or replica region satisfies the **same** regulatory classification as the primary.

| Replication type | Risk | Action |
|---|---|---|
| Geo-redundant storage (GRS/GZRS) | Data replicated to paired region | Verify paired region is in same regulatory boundary; use ZRS or LRS if not |
| Azure Cosmos DB multi-region writes | Each write region must meet the same classification | Audit all configured write regions |
| Azure SQL geo-replication | Secondary region must be within the same boundary | Check secondary region before enabling |
| ACR geo-replication | Images replicated to each configured region | Limit replications to boundary-compliant regions |

If replication would cross a regulatory boundary, use **zone-redundant storage (ZRS)** or **locally redundant storage (LRS)** and accept the availability trade-off. Document the decision as an ADR.

---

## Azure Policy Enforcement

Enforce region constraints with Azure Policy; do not rely on template parameters alone — a misconfigured deployment pipeline can bypass `@allowed` decorators.

```bicep
// Assign built-in policy: "Allowed locations" (e56962a6-4747-49cd-b67b-bf8b01975c4f)
resource locationPolicy 'Microsoft.Authorization/policyAssignments@2024-04-01' = {
  name: 'allowed-locations-${resourceGroup().name}'
  scope: resourceGroup()
  properties: {
    displayName: 'Restrict deployments to approved residency regions'
    policyDefinitionId: '/providers/Microsoft.Authorization/policyDefinitions/e56962a6-4747-49cd-b67b-bf8b01975c4f'
    parameters: {
      listOfAllowedLocations: {
        value: [location, pairedLocation]
      }
    }
  }
}
```

---

## GDPR Data Boundary Checklist

Before deploying any resource that stores or processes personal data:

- [ ] Region is within the designated regulatory boundary (EEA / UK / in-country)
- [ ] Geo-replication (if enabled) does not cross the boundary
- [ ] Log Analytics workspace is in the same boundary as the resource it monitors
- [ ] Backup vault and recovery services vault are co-located or boundary-compliant
- [ ] Azure Cognitive Services / OpenAI: confirm the endpoint region; data is processed in the endpoint region
- [ ] AI training or fine-tuning: confirm training data does not leave the boundary
- [ ] Third-party SaaS integrated via API: confirm data processing location; Data Processing Agreements in place
- [ ] Diagnostic settings: destination storage account or Event Hub is in the same boundary

---

## Kubernetes and AKS

- AKS clusters must be created in boundary-compliant Azure regions — this is ❌ **Permanent**; node pool regions cannot be changed after creation.
- Do NOT route traffic through Azure Front Door global POPs if data cannot leave the boundary — use regional endpoints or Azure Traffic Manager with geographic routing policies.
- Azure Key Vault referenced by Kubernetes secrets must be in the same boundary region.
- Container images pulled from ACR with geo-replication enabled: verify every configured replica region is boundary-compliant.
- Persistent volumes backed by Azure Disk or Azure Files: storage account region follows the AKS cluster region; confirm before enabling geo-redundancy.

---

## GitHub Actions and CI/CD Pipelines

- Pipeline runners process code and build artifacts — not personal data in most cases — so runner location is generally not restricted.
- Secrets stored in GitHub Secrets are encrypted and hosted in GitHub's US-based infrastructure. Do **not** store personal data or regulated payloads in Actions secrets.
- If pipelines read or transform regulated data (for example, database migration scripts, data seeding), run on **self-hosted runners** within the boundary.
- Deployment pipeline outputs (logs, artifacts) should not contain personal data. Sanitize before uploading as GitHub Actions artifacts.

---

## Common Anti-Patterns

| Anti-pattern | Risk | Fix |
|---|---|---|
| `location = 'eastus'` hard-coded in Bicep | Data lands outside EU/UK | Parameterise and validate with `@allowed` |
| Enabling GRS without checking paired region | Replication to non-compliant region | Use ZRS if paired region is out of boundary |
| Log Analytics workspace in different region from resource | Logs (potentially containing PII) leave boundary | Co-locate workspace and resource |
| Azure OpenAI endpoint in non-boundary region | Prompts and responses processed outside boundary | Choose endpoint in northeurope or swedencentral |
| Storing PII in pipeline environment variables | Appears in logs and GitHub UI | Use a secrets manager; redact logs |
| ACR geo-replication to unchecked regions | Image metadata (not code) replicated out of boundary | Restrict replica locations in ACR configuration |
| Omitting location from resource group template | Resource group defaults to deployer's current region | Always pass `location` explicitly |

---

## References

- [Azure data residency overview](https://azure.microsoft.com/explore/global-infrastructure/data-residency/)
- [GDPR guidance for Azure](https://learn.microsoft.com/compliance/regulatory/gdpr)
- [Azure regions and availability zones](https://learn.microsoft.com/azure/reliability/regions-overview)
- [Azure paired regions](https://learn.microsoft.com/azure/reliability/cross-region-replication-azure)
- [Azure Policy built-in: Allowed locations](https://www.azadvertizer.net/azpolicyadvertizer/e56962a6-4747-49cd-b67b-bf8b01975c4f.html)
