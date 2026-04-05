---
applyTo: "**/*.bicep,**/*.bicepparam,**/*.tf,**/*.tfvars"
description: Azure infrastructure code requirements for Bicep/Terraform including API version verification, role definition IDs, resource limitations, container registry enforcement, and service deprecation/EOL detection
---

# Azure Infrastructure Code Requirements

These rules apply when authoring or reviewing Bicep, ARM, or Terraform templates that deploy to Azure.

## Container Registry Image Reference Enforcement

- **NEVER hardcode Azure Container Registry (ACR) hostnames (e.g., \*.azurecr.io) in Kubernetes manifests or deployment files.**
- **ALWAYS use pipeline-injected variables (e.g., `${ACR_NAME}`) for image references.**
- **ALWAYS apply manifests via CI/CD with variable substitution.**
- If Copilot detects ErrImagePull or ImagePullBackOff, it MUST check for hardcoded .azurecr.io references and escalate if found, recommending variable substitution or manifest correction.
- Copilot should recommend a pipeline lint step to block hardcoded ACR hostnames in manifests.

## Global-Unique Names & DNS Labels (AKS)

- Global namespaces (Key Vault, App Config, Azure Maps, Public IP DNS labels) must be unique.
- Prefer deterministic suffixes (`uniqueString(rg.id)` or subscription-based short suffix) in names and AKS service DNS labels to avoid `NameUnavailable` / `DnsRecordInUse`.

## Verify Currency Before Assuming

### 1. API Versions (CRITICAL)

- **MUST verify** the API version exists: https://learn.microsoft.com/azure/templates/
- **MUST check** supported API versions: `az provider show --namespace Microsoft.YourService --query "resourceTypes[?resourceType=='yourResource'].apiVersions"`
- **MUST validate** the version is NOT a future date that doesn't actually exist (e.g., `2024-08-01` that only supports `2023-03-02-preview`)
- **MUST test** with `az bicep build` to catch non-existent API versions before deployment
- Document: Include verification evidence (exact command output or learn.microsoft.com reference) in PR descriptions
- If deployment fails with `NoRegisteredProviderFound`, the API version doesn't exist — check `az provider show` output
- Pitfall: `az bicep build` succeeds but Azure has no registered provider for that API version/region combination

### 2. Role Definition IDs

- Always look up current values; do not assume hard-coded IDs are correct
- Check: https://learn.microsoft.com/azure/role-based-access-control/built-in-roles
- Validate: `az role definition list --query "[?contains(roleName, 'YourRole')]"`
- Pitfall: Role IDs may not exist globally or may differ from examples

### 3. Required Parameters

- Always check the ARM template schema; do not assume a parameter is optional
- Check: Full ARM template schema for the resource type at learn.microsoft.com/azure/templates
- Validate: Run `az deployment what-if` with all likely parameters first
- Document: Mark unusual required parameters in comments (e.g., administratorLoginPassword)
- Pitfall: Different API versions have different required fields; examples may be incomplete

### 4. Resource Limitations (CRITICAL)

- **MUST verify** allowed values for SKUs, sizes, and quotas (e.g., storage sizes, vCores)
- Common pitfall: Not all storage sizes are valid (e.g., PostgreSQL Elastic Clusters: 128 GB works, 256 GB fails for worker nodes)
- Test with smallest/safest values first in dev environments
- If deployment fails with "not allowed" error, reduce to next smaller standard size (32GB, 64GB, 128GB, 512GB, 1TB)

**Mark uncertainties** with `[VERIFY]` tags pointing to the resource schema or documentation source.
**If API version cannot be verified, block implementation** and request human researcher confirmation.

## Azure Service Deprecation & EOL Detection (Critical)

Copilot MUST escalate and block deployment if any of these conditions are detected.

### Detection Triggers

- Service documentation contains "**no longer supported for new projects**"
- Service documentation contains "**end-of-support**", "**deprecated**", or "**reaches end of life**"
- API version is preview/beta (`@YYYY-MM-DD-preview` or `@YYYY-MM-dd-beta`)
- `az deployment what-if` fails with "**InvalidTemplateDeployment**", "**ResourceTypeNotSupported**", or "**provisioning state Failed**" (indicates service unavailability)

### Response Protocol

If any detection trigger fires:

1. **Do NOT proceed** with deployment
2. **Tag response** as `[ESCALATION REQUIRED: Deprecated Service]` or `[ESCALATION REQUIRED: Preview API]`
3. **State the finding**: Quote the official documentation showing deprecation or preview status
4. **Propose alternative**: Research and recommend the Microsoft-supported replacement service
5. **Provide migration path**: Include link to official migration guide or create migration spec
6. **Defer decision**: Mark as requiring architect/tech lead approval before proceeding

### Copilot Responsibilities Under This Policy

When authoring infrastructure code:

1. **Always query** `microsoft.learn.mcp` MCP for service name + "deprecated, end-of-life, end-of-support" before finalizing
2. **Flag preview APIs** immediately with explicit justification requirement
3. **Run `az deployment what-if`** locally before returning code (catch service unavailability early)
4. **Block merges** if escalation tags are present (requires human sign-off to proceed)
5. **Document alternatives** in PR/architecture review with links to official migration paths

When reviewing existing code:

1. **Audit** any Bicep using preview APIs or older resource types
2. **Query deprecation status** for services in use
3. **Report findings** to team with urgency: flag if production is using end-of-life services
4. **Recommend migration timeline** based on support end date (from Microsoft docs)
