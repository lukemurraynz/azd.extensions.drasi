---
description: "Azure Verified Modules (AVM) and Bicep"
applyTo: "**/*.bicep, **/*.bicepparam"
---

# Azure Verified Modules (AVM) Bicep

## Overview

Azure Verified Modules (AVM) are pre-built, tested, and validated Bicep modules that follow Azure best practices. Use these modules to create, update, or review Azure Infrastructure as Code (IaC) with confidence.

## Discover Modules

### Bicep Public Registry

- Search for modules: `br/public:avm/res/{service}/{resource}:{version}`
- Browse available modules: `https://github.com/Azure/bicep-registry-modules/tree/main/avm/res`
- Example: `br/public:avm/res/storage/storage-account:0.30.0`

### Official AVM Index

- **Bicep Resource Modules**: `https://raw.githubusercontent.com/Azure/Azure-Verified-Modules/refs/heads/main/docs/static/module-indexes/BicepResourceModules.csv`
- **Bicep Pattern Modules**: `https://raw.githubusercontent.com/Azure/Azure-Verified-Modules/refs/heads/main/docs/static/module-indexes/BicepPatternModules.csv`

### Module Documentation

- **GitHub Repository**: `https://github.com/Azure/bicep-registry-modules/tree/main/avm/res/{service}/{resource}`
- **README**: Each module contains comprehensive documentation with examples

### Resource vs Pattern Modules

- Prefer **resource modules** for composing custom architectures.
- Use **pattern modules** for common end-to-end scenarios.

## Use Modules

### From Examples

1. Review module README in `https://github.com/Azure/bicep-registry-modules/tree/main/avm/res/{service}/{resource}`
2. Copy example code from module documentation
3. Verify the README badge/module index for the latest stable version (examples can lag)
4. Reference module using `br/public:avm/res/{service}/{resource}:{version}`
5. Configure required and optional parameters

### Example Usage

```bicep
module storageAccount 'br/public:avm/res/storage/storage-account:0.30.0' = {
  name: 'storage-account-deployment'
  scope: resourceGroup()
  params: {
    name: storageAccountName
    location: location
    skuName: 'Standard_LRS'
    tags: tags
  }
}
```

### When AVM Module Not Available

If no AVM module exists for a resource type:

- Check the AVM Bicep Resource Index to confirm the module does not exist.
- **CRITICAL: Verify the Azure service is not deprecated or end-of-support** before proceeding (see Service Deprecation Validation below)
- **CRITICAL: Verify the API version exists** using `az provider show --namespace Microsoft.YourProvider` or check [learn.microsoft.com/azure/templates](https://learn.microsoft.com/azure/templates) before using
- Use native Bicep resource declarations with the **latest stable (non-preview) API version that actually exists**.
- Consider internal, AVM-inspired modules for repeated patterns so you can migrate later with minimal change.

#### API Version Validation Patterns (CRITICAL)

**ALWAYS verify API versions before use**—deployment will fail if the version doesn't exist.

```bicep
// ❌ WRONG: Missing API version entirely
resource storage 'Microsoft.Storage/storageAccounts' = {
  name: 'mystorageacct'
  // BCP029: Resource type must include valid API version
}

// ❌ WRONG: Using invalid/non-existent API version
resource storage 'Microsoft.Storage/storageAccounts@2025-01-02' = {
  name: 'mystorageacct'
  // BCP081: Invalid API version '2025-01-02' for type 'Microsoft.Storage/storageAccounts'
}

// ✅ CORRECT: Valid API version (verify first!)
resource storage 'Microsoft.Storage/storageAccounts@2025-08-01' = {
  name: 'mystorageacct'
  location: location
  sku: {
    name: 'Standard_LRS'
  }
  kind: 'StorageV2'
}
```

**Verification steps:**

```bash
# Method 1: Check available API versions for a resource type
az provider show \
  --namespace Microsoft.Storage \
  --query "resourceTypes[?resourceType=='storageAccounts'].apiVersions"

# Method 2: Check ARM template reference
# Visit: https://learn.microsoft.com/azure/templates/microsoft.storage/storageaccounts
```

#### Resource Property Access Patterns

Use symbolic names and property accessors—avoid legacy `reference()` function:

```bicep
// ❌ WRONG: Using reference() function (legacy JSON pattern)
output blobEndpoint string = reference(resourceId('Microsoft.Storage/storageAccounts', 'myStorage'), '2024-01-01').primaryEndpoints.blob

// ✅ CORRECT: Use symbolic name with property accessor
resource storage 'Microsoft.Storage/storageAccounts@2025-08-01' = {
  name: 'myStorage'
  location: location
  sku: { name: 'Standard_LRS' }
  kind: 'StorageV2'
}

output blobEndpoint string = storage.properties.primaryEndpoints.blob

// ❌ WRONG: Referencing non-existent property
output foo string = storage.bar  // BCP053: Property 'bar' does not exist

// ✅ CORRECT: Reference valid properties only
output storageName string = storage.name
output storageId string = storage.id

// ✅ CORRECT: Access nested properties via dot notation
resource publicIp 'Microsoft.Network/publicIPAddresses@2025-01-01' = {
  name: 'myPublicIp'
  location: location
  properties: {
    publicIPAllocationMethod: 'Static'
    dnsSettings: {
      domainNameLabel: 'myapp'
    }
  }
}

output fqdn string = publicIp.properties.dnsSettings.fqdn

// ✅ CORRECT: Reference existing resources (not deployed by this template)
resource existingStorage 'Microsoft.Storage/storageAccounts@2025-08-01' existing = {
  name: storageAccountName
}

output existingBlobEndpoint string = existingStorage.properties.primaryEndpoints.blob
```

**Rules:**

- Always include `@YYYY-MM-DD` API version on **every** resource declaration
- Verify API version exists before use (deployment will fail otherwise)
- Use symbolic names (`storage.properties.X`) not `reference()` function
- Use `existing` keyword for resources outside current deployment

## Author Code

### Module References

- **Resource Modules**: `br/public:avm/res/{service}/{resource}:{version}`
- **Pattern Modules**: `br/public:avm/ptn/{pattern}:{version}`
- Example: `br/public:avm/res/network/virtual-network:0.7.2`

### Symbolic Names

- Use lowerCamelCase for all names (variables, parameters, resources, modules)
- Use resource type descriptive names (e.g., `storageAccount` not `storageAccountName`)
- Avoid 'name' suffix in symbolic names as they represent the resource, not the resource's name
- Avoid distinguishing variables and parameters by suffixes

```bicep
module storageAccount 'br/public:avm/res/storage/storage-account:0.30.0' = {
  name: 'storage-account-deployment'
  params: {
    name: storageAccountName
  }
}

resource storageAccountFirewall 'Microsoft.Storage/storageAccounts/networkRuleSets@2025-01-01' = {
  name: '${storageAccountName}/default'
  properties: {
    defaultAction: 'Deny'
  }
}
```

### Module Versioning

- Always pin to specific module versions: `:{version}`
- Use semantic versioning (e.g., `:0.30.0`)
- Review module changelog before upgrading
- Test version upgrades in non-production environments first

### Code Structure

- ✅ **Declare** parameters at top of file with `@description()` decorators
- ✅ **Use** `@sys.description()` only when a parameter named `description` causes a naming collision
- ✅ **Specify** `@minLength()` and `@maxLength()` for naming parameters
- ✅ **Use** `@allowed()` decorator sparingly to avoid blocking valid deployments
- ✅ **Use** union types for environment selectors (e.g., `'dev' | 'staging' | 'prod'`) when supported
- ✅ **Set** default values safe for test environments (low-cost SKUs)
- ✅ **Use** variables for complex expressions instead of embedding in resource properties
- ✅ **Leverage** `loadJsonContent()` for external configuration files

### Resource References

- ✅ **Use** symbolic names for references (e.g., `storageAccount.id`) not `reference()` or `resourceId()`
- ✅ **Create** dependencies through symbolic names, not explicit `dependsOn`
- ✅ **Use** `existing` keyword for accessing properties from other resources
- ✅ **Access** module outputs via dot notation (e.g., `storageAccount.outputs.resourceId`)

### Resource Naming

- ✅ **Use** `uniqueString()` with meaningful prefixes for unique names
- ✅ **Add** prefixes since some resources don't allow names starting with numbers
- ✅ **Respect** resource-specific naming constraints (length, characters)

### Child Resources

- ✅ **Avoid** excessive nesting of child resources
- ✅ **Use** `parent` property or nesting instead of constructing names manually

### Security

- ❌ **Never** include secrets or keys in outputs
- ❌ **Never** propose post-deployment manual steps for configuration that exists in ARM template properties
- ✅ **Use** resource properties directly in outputs (e.g., `storageAccount.outputs.primaryBlobEndpoint`)
- ✅ **Prefer** managed identities and Entra ID over Key Vault for service-to-service auth (fewer secrets to manage)
- ✅ **Prefer** built-in resource properties (e.g., `authConfig`, `identity`) over post-deployment scripting
- ✅ **Mask** secrets in parameter files and CI variables
- ✅ **Enable** managed identities where possible
- ✅ **Tier diagnostics**: Baseline (minimal required categories) vs Investigation (time-boxed deep dive) — never enable all categories by default
- ✅ **Disable** public access when network isolation is enabled
- ✅ **Check ARM template references** before suggesting manual configuration (often settable in Bicep via `properties`)
- ✅ **Document post-deployment steps only if truly unavoidable** (e.g., data plane operations that can't be expressed in ARM)
- ❌ **Never** use deterministic or guessable passwords (including `uniqueString(...)`) for admin credentials. Use secure parameters, random secrets, and store in Key Vault.
- ✅ **Parameterize security hardening** (public network access, local auth, private endpoints). Default to secure settings for `prod`.
- ✅ **For production**: disable local auth where supported, prefer private endpoints, and avoid public access for ACR/App Configuration/AKS unless explicitly justified.
- ✅ **Use clear identity naming**: `*ResourceId`, `*PrincipalId`, `*ClientId` to avoid mixing identity types.

### Types

- ✅ **Import** types from modules when available: `import { deploymentType } from './module.bicep'`
- ✅ **Use** user-defined types for complex parameter structures
- ✅ **Leverage** type inference for variables

```bicep
import { storageAccountParams } from './types.bicep'

type alertRule = {
  name: string
  severity: 'Sev0' | 'Sev1' | 'Sev2'
  enabled: bool
}

param rules array<alertRule>
param storage storageAccountParams
```

### Documentation

- ✅ **Include** helpful `//` comments for complex logic
- ✅ **Use** `@description()` on all parameters with clear explanations
- ✅ **Document** non-obvious design decisions

## Validate & Integrate

### API Version Validation (CRITICAL)

Before using any non-AVM resource type, verify the API version is supported:

**Option 1: Check ARM Template Reference (Recommended)**

```bash
# Navigate to: https://learn.microsoft.com/azure/templates/
# Search for resource type and verify the API version exists in your region
# Example: Microsoft.DBforPostgreSQL/serverGroupsv2
# Supported versions: 2023-03-02-preview, 2022-11-08, etc.
```

**Option 2: Query Azure Providers**

```bash
# List supported API versions for a provider
az provider show --namespace Microsoft.DBforPostgreSQL \
  --query "resourceTypes[?resourceType=='serverGroupsv2'].apiVersions" -o tsv

# Output example:
# 2023-03-02-preview
# 2022-11-08
```

**Common Mistakes to Avoid:**

- ❌ Using future API versions (e.g., `2024-08-01` when only `2023-03-02-preview` exists)
- ❌ Using non-existent preview versions
- ❌ Forgetting to verify region support (not all API versions available everywhere)

**Rule: If `az bicep build` succeeds but deployment fails with `NoRegisteredProviderFound`, the API version does not exist.**

After any changes to Bicep files, run the following commands to ensure all files build successfully:

```shell
# Ensure Bicep CLI is up to date
az bicep upgrade

# Build and validate all changed Bicep files (not just main.bicep)
az bicep build --file main.bicep
```

### Bicep Parameter Files

- ✅ **Always** update accompanying `*.bicepparam` files when modifying `*.bicep` files
- ✅ **Validate** parameter files match current parameter definitions
- ✅ **Test** deployments with parameter files before committing

### What-If Validation

- Run `az deployment what-if` at the appropriate scope (subscription or resource group) as part of pre-merge validation.

### CI/CD Validation & Deployment Pattern

> **Note:** Action `uses:` references below use tag-only format (`@v4`) for readability. In real workflows, pin to full SHA per `cicd-security.instructions.md` (e.g., `actions/checkout@<sha> # v4.x.y`).

**Pull Request: Syntax + What-If Analysis (no deployment)**

```yaml
validate-infrastructure:
  runs-on: ubuntu-latest
  if: github.event_name == 'pull_request'
  steps:
    - uses: actions/checkout@v4
    - uses: azure/setup-azure-cli@v3

    - name: Check for Preview APIs (fail if found in production code)
      run: |
        if grep -r "@[0-9]\{4\}-[0-9]\{2\}-[0-9]\{2\}-preview" infrastructure/bicep/*.bicep; then
          echo "❌ FAIL: Preview APIs detected. Use stable API versions only."
          exit 1
        fi

    - name: Validate Bicep Syntax & Linting
      run: |
        az bicep build --file infrastructure/bicep/main.bicep --outdir /tmp
        echo "✅ Bicep syntax valid"
```

    - name: What-If Analysis (shows proposed changes)
      run: |
        az deployment subscription what-if \
          --location uksouth \
          --template-file infrastructure/bicep/main.bicep \
          --parameters environment=dev projectName=myapp

    - name: Check What-If Output for Deployment Failures
      run: |
        # Fail if what-if indicates service is unavailable in region
        if grep -i "provisioning state.*failed\|resource.*not.*support\|invalid.*deployment" what-if-output.txt; then
          echo "❌ FAIL: Potential deployment failure detected. Service may be unavailable in region."
          exit 1
        fi

````

**Main Branch: Full Deployment with Verification**

```yaml
deploy-infrastructure:
  runs-on: ubuntu-latest
  if: github.event_name == 'push' && github.ref == 'refs/heads/main'
  needs: [validate-infrastructure]  # Explicit dependency - must pass validation first

  steps:
    - uses: actions/checkout@v4

    - uses: azure/login@v2
      with:
        client-id: ${{ secrets.AZURE_CLIENT_ID }}
        tenant-id: ${{ secrets.AZURE_TENANT_ID }}
        subscription-id: ${{ secrets.AZURE_SUBSCRIPTION_ID }}

    - name: Deploy Infrastructure
      run: |
        az deployment subscription create \
          --name "infra-$(date +%s)" \
          --location uksouth \
          --template-file infrastructure/bicep/main.bicep \
          --parameters environment=prod projectName=myapp

    - name: Verify Deployment
      run: az resource list --resource-group "myapp-prod-rg" --output table

    - name: Wait for Resource Health
      run: |
        # Give newly deployed resources time to stabilize
        echo "⏳ Waiting 2 minutes for resources to stabilize..."
        sleep 120

        # Check critical resources are healthy (fail if any resource shows provisioning error)
        az resource list --resource-group "myapp-prod-rg" \
          --query "[?provisioningState != 'Succeeded']" \
          --output table
````

**Rules:**

- Infrastructure validation runs on every PR (early error detection)
- **Preview API check fails PR** - no preview APIs in production paths
- **What-If gating** - deployment only proceeds if what-if succeeds
- Infrastructure deployment happens **before** application jobs
- Use explicit `needs:` dependencies to enforce order
- Include verification steps (resource list, pod status, health checks, etc.)
- Deploy only to main; validate on all branches
- **Fail fast** on what-if errors (service unavailability, incompatible parameters) to catch issues before actual deployment

## Tool Integration

### Use Available Tools

- **Schema Information**: Use resource schema lookup tools when available
- **Deployment Guidance**: Use official Azure/Bicep documentation for service-specific guidance

### GitHub Copilot Integration

When working with Bicep:

1. Prompt: "Use AVM module `br/public:avm/res/{service}/{resource}:{version}` where available; otherwise use native Bicep with latest stable API."
2. Prompt: "Cross-check README badge/module index before pinning versions (examples can lag)."
3. Prompt: "Generate validation commands (az bicep build + az deployment what-if at correct scope)."
4. Update accompanying `.bicepparam` files.
5. Document customizations or deviations from examples.

## Service Deprecation Validation (CRITICAL)

**Before using any new Azure service in infrastructure (production or non-prod), validate:**

### Service Support Status

1. **Check Microsoft Learn** for official status:
   - Use `microsoft.learn.mcp` MCP to search: `"<service> deprecated end of support EOL"`
   - Example query: `"Azure Cosmos DB for PostgreSQL status"`
   - Look for **Important** or **Deprecated** warnings at service documentation top

2. **Verify no "end-of-support" or "no longer supported"** messaging in official docs
   - Services marked "end-of-support" or "deprecated for new projects" **must not** be used
   - If unavoidable (legacy system), document in PR with migration plan and timeline

3. **Check regional availability**:
   ```bash
   # Verify service is available in target region(s)
   az provider show --namespace Microsoft.DBforPostgreSQL \
     --query "resourceTypes[?resourceType=='flexibleServers'].locations"
   ```

### Validate API Version Is Stable (Not Preview)

- **Never use preview APIs** (`@YYYY-MM-DD-preview`) in production code without explicit justification and PR appr approval
- Preview APIs may be removed or change incompatibly
- Use only APIs marked as **stable** (e.g., `@2024-08-01`, not `@2023-03-02-preview`)
- If preview is necessary, document the rationale and add a reminder to upgrade when stable version is available

### Checklist Before Committing

- [ ] Service is **not marked deprecated** or "end-of-support" in Microsoft Learn
- [ ] Service is **available in target regions** (verify with `az provider show` or portal)
- [ ] API version used is **stable** (not preview or beta)
- [ ] **What-if deployment** succeeds without errors indicating service unavailability
- [ ] **Deployed successfully** in dev/staging environment first (catches regional/subscription issues early)

### Red Flags (Immediate Escalation)

If you encounter any of these, stop and escalate to architect/lead:

- ⚠️ Service docs say "**no longer supported for new projects**"
- ⚠️ Service docs say "**end-of-support**" or "**deprecated**"
- ⚠️ API version contains "preview" or "beta"
- ⚠️ `az deployment what-if` returns **"InvalidTemplateDeployment"** or **"ResourceTypeNotSupported"** (service may not be in region)
- ⚠️ Deployment fails with "**provisioning state 'Failed'**" (likely service availability issue)

### Example: Avoiding the Cosmos DB for PostgreSQL Trap

❌ **What went wrong:**

```bicep
resource cosmosCluster 'Microsoft.DBforPostgreSQL/serverGroupsv2@2023-03-02-preview' = {
  // Used preview API + deprecated service
  // No validation of regional availability or EOS status
  // Result: "provisioning state 'Failed'" in prod
}
```

✅ **What should have happened:**

1. Search `microsoft.learn.mcp` for "Azure Cosmos DB for PostgreSQL"
2. **Find**: "Azure Cosmos DB for PostgreSQL is no longer supported for new projects"
3. **Escalate**: Propose using `Azure Database for PostgreSQL Flexible Server` instead
4. **Validate**: Confirm new service is available in uksouth and uses stable API (`@2024-08-01`)
5. **Deploy**: Fresh `postgres-flexible.bicep` module with validation in CI

## Tool Integration

## API and Schema Validation

### API Version Verification (CRITICAL)

Before using any API version in Bicep:

1. **Verify region support**: Not all API versions are available in all regions immediately
2. **Check ARM template reference**: Use microsoft.learn.mcp MCP to confirm the resource type supports your target API version
3. **Document the verification**: Include proof in PR comments

**Validation steps:**

```bash
# Query available API versions for the resource type in your region
az provider show --namespace Microsoft.ContainerService \
  --query "resourceTypes[?resourceType=='managedClusters'].apiVersions" --output table

# Build and validate locally before deployment
az bicep build --file main.bicep
```

**Common pitfall**: Using API versions from documentation examples without verifying region support. The latest examples may reference newer versions not yet available globally.

### Role Definition Validation

Azure built-in role IDs and names change over time:

1. **Never hard-code role IDs** without verification
2. **Use microsoft.learn.mcp MCP** to confirm current role definition IDs
3. **Reference**: https://learn.microsoft.com/azure/role-based-access-control/built-in-roles

**Example validation:**

```bash
# Before deployment, verify role exists
az role definition list --query "[?contains(roleName, 'Cosmos DB')]"
```

**Common pitfall**: Role IDs may be incorrect or the role may not exist globally. Examples from documentation may be outdated.

### Required Parameter Discovery

Some Azure APIs have required parameters not immediately obvious from module examples:

1. **Always check the ARM template schema** for your resource type
2. **Use microsoft.learn.mcp MCP** to search resource schema requirements
3. **Example**: `administratorLoginPassword` is required for Cosmos DB PostgreSQL even when `passwordAuth` is disabled

**Validation steps:**

```bash
# Build template to catch schema mismatches early
az bicep build --file main.bicep

# Validate with sample parameters before deployment
az deployment subscription what-if \
  --template-file main.bicep \
  --parameters environment=dev projectName=test cosmosAdminPassword='Test!P@ss123'
```

## Troubleshooting

### Common Issues

1. **Module Version**: Always specify exact version in module reference
2. **Missing Dependencies**: Ensure resources are created before dependent modules
3. **Validation Failures**: Run `az bicep build` to identify syntax/type errors
4. **Parameter Files**: Ensure `.bicepparam` files are updated when parameters change

### Support Resources

- **AVM Documentation**: `https://azure.github.io/Azure-Verified-Modules/`
- **Bicep Registry**: `https://github.com/Azure/bicep-registry-modules`
- **Bicep Documentation**: `https://learn.microsoft.com/azure/azure-resource-manager/bicep/`
- **Best Practices**: `https://learn.microsoft.com/azure/azure-resource-manager/bicep/best-practices`

## Anti-Patterns to Avoid

❌ **Using a deprecated or end-of-support Azure service without migration plan**

- Always verify service is not marked "no longer supported for new projects" in Microsoft Learn
- Query: `microsoft.learn.mcp` MCP with service name + "deprecated" or "end of support"
- **Example (Real)**: Azure Cosmos DB for PostgreSQL is EOL; should use Azure Database for PostgreSQL Flexible Server instead
- **Impact**: Deployment failures with "provisioning state 'Failed'", regional unavailability, no support from Microsoft
- **Prevention**: Check `microsoft.learn.mcp` before architectural decisions; escalate if service shows deprecation warnings

❌ **Using preview or beta APIs in production code**

- Preview APIs (`@YYYY-MM-DD-preview`) may be removed or break incompatibly
- Always use stable API versions (e.g., `@2024-08-01` not `@2023-03-02-preview`)
- Document and justify preview API usage; add calendar reminder to upgrade
- **Example**: `Microsoft.DBforPostgreSQL/serverGroupsv2@2023-03-02-preview` should have been `@2024-08-01` stable
- **Impact**: Silent breaking changes, deployment failures after service GA updates, compatibility issues
- **Prevention**: Enforce stable APIs in code review; flag preview APIs in CI/CD

❌ **Skipping service regional availability validation before deployment**

- Not all services are available in all regions (even if they're in public preview or GA globally)
- Always run `az deployment what-if` in target region before production deployment
- Validate with `az provider show --namespace <service> --query locations`
- **Example**: Cosmos DB for PostgreSQL not available in uksouth for new deployments
- **Impact**: Deployment failures only caught in production environment
- **Prevention**: Add regional validation step to CI/CD pipeline before deploying to prod

❌ **Using outdated or unsupported API versions without verification**

- Always review [ARM template version history](https://learn.microsoft.com/azure/templates/) for your resource type
- Older versions may be deprecated, unsupported, or unavailable in your region
- Use `microsoft.learn.mcp` MCP to verify before hard-coding API versions
- **Example**: `2023-10-02` for AKS managedClusters not available in uksouth (correct: `2025-01-01`)
- **Prevention**: Add API version validation check to PR template; use `az provider show` to verify

❌ **Hard-coding role definition IDs without current verification**

- Role IDs are generally stable but can change and may vary by region
- Always reference [Azure built-in roles](https://learn.microsoft.com/azure/role-based-access-control/built-in-roles)
- Use microsoft.learn.mcp MCP to confirm IDs before deployment
- **Example**: `028f4ed7-e6c9-4bd5-b3dc-51d3d976a1b6` does not exist (correct: verify first)
- **Prevention**: Use role names where possible; validate IDs in CI/CD before deployment

❌ **Assuming a parameter is optional without checking the ARM schema**

- Different API versions may have different required fields
- Always validate against the [ARM template reference](https://learn.microsoft.com/azure/templates/)
- Run `az deployment what-if` with all required parameters to validate early
- **Example**: `administratorLoginPassword` required for Cosmos DB PostgreSQL even with Entra ID auth
- **Prevention**: Add schema validation to Bicep linting; require what-if before merge

❌ **Skipping `az deployment what-if` or only testing locally with `az bicep build`**

- `az bicep build` catches syntax errors but not all schema/parameter issues
- Always run what-if at the correct scope before deployment
- What-if reveals parameter mismatches, property incompatibilities, and service unavailability
- **Example**: Local build succeeds (`az bicep build`) but what-if fails with "service not available"
- **Prevention**: Make what-if mandatory in CI/CD gating; block merge if what-if fails

❌ **Deploying resources that collide with soft-deleted Azure resources**

- Key Vault, Cognitive Services, App Configuration, and API Management support soft-delete
- A deleted resource retains its globally unique name for the purge-protection period (7-90 days)
- Redeploying with the same name produces `ConflictError` or `NameUnavailable`
- **Prevention**: Use `uniqueString()` to generate names; document soft-delete purge commands; verify with `az keyvault list-deleted`, `az cognitiveservices account list-deleted`
- **Recovery**: Purge the soft-deleted resource (`az keyvault purge --name <name>`) or use a different project name prefix

❌ **Using `Standard` SKU for AI model deployments without regional availability check**

- Azure OpenAI deployment SKUs (`Standard`, `GlobalStandard`, `ProvisionedManaged`) vary by model and region
- `Standard` is not available in all regions for all models; `GlobalStandard` has broader availability
- **Prevention**: Use `GlobalStandard` as default for model deployments; check [model availability matrix](https://learn.microsoft.com/azure/ai-services/openai/concepts/models) before committing
- **Verification**: `az cognitiveservices account list-skus --resource-group <rg> --name <account>` or Azure Portal → Model Deployments → Available SKUs

❌ **Setting `networkAcls.defaultAction: 'Deny'` without configuring allowed callers**

- AI Services, Storage, Key Vault, and Cosmos DB support network ACLs
- Setting `Deny` without adding virtual network rules or IP exceptions blocks all callers including Container Apps
- With managed identity (`disableLocalAuth: true`), auth is already enforced — `Allow` with managed identity is a valid secure pattern
- **Prevention**: Use `Allow` for public-access-with-auth patterns; use `Deny` only with VNet integration + private endpoints fully configured
- **Symptom**: 403 Forbidden on API calls from Container Apps or local dev after deployment

❌ **Deploying resources that require unregistered Azure subscription preview features**

- Some Azure services (NSP, certain AI capabilities) require subscription feature flags (e.g., `AllowNetworkSecurityPerimeter`)
- Deployment fails with opaque errors if the feature is not registered or still Pending
- **Prevention**: Check feature status before deploying: `az feature show --namespace <provider> --name <feature> --query properties.state`
- **Pattern**: Make feature-gated resources conditional with `enableXxx bool = false` params; require explicit opt-in
- **Registration**: `az feature register --namespace <provider> --name <feature>` (can take minutes to hours to propagate)

❌ **Not making preview or potentially-failing resources conditional with safe defaults**

- AI Foundry capability hosts, model deployments, and NSP resources can cause opaque validation errors
- Hard-coding these as always-deployed blocks the entire `azd up` when the feature isn't available
- **Pattern**: Use `param enableXxx bool = false` with `resource r '...' = if (enableXxx) { ... }` and conditional outputs
- **Prevention**: Default preview features to `false`; wire toggles through `.bicepparam` via env vars; document which toggles exist

## Bicep Linting & Code Quality (MANDATORY)

Enforce linting in CI/CD to prevent common issues:

### Linting Check (CI/CD Requirement)

```bash
# Build and lint Bicep files
az bicep build --file infrastructure/bicep/main.bicep --outdir /tmp

# Check for linter warnings/errors
# BicepLint rules enforced:
# - no-unused-params: Flag unused parameters (must be removed or used)
# - use-recent-api-versions: Disallow deprecated/old API versions
# - secure-parameter-default: Disallow secrets with defaults
# - use-stable-vm-image: Prevent unknown/unstable image references
```

### Common Linting Violations & Fixes

**no-unused-params**

```bicep
// ❌ WRONG: Parameter unused
param aksClusterName string  // Never referenced in code
param environment string     // Only used in resource name

resource example 'Microsoft.Compute/virtualMachines@2025-03-01' = {
  name: 'vm-${environment}'
  // aksClusterName not used anywhere
}

// ✅ CORRECT: Remove unused parameter
// (aksClusterName removed entirely)
param environment string

resource example 'Microsoft.Compute/virtualMachines@2025-03-01' = {
  name: 'vm-${environment}'
}
```

**use-recent-api-versions**

```bicep
// ❌ WRONG: Old API version
resource exampleOld 'Microsoft.Storage/storageAccounts@2021-02-01' = { ... }

// ✅ CORRECT: Use latest stable version
resource exampleNew 'Microsoft.Storage/storageAccounts@2025-08-01' = { ... }
```

**secure-parameter-default**

```bicep
// ❌ WRONG: Secret with default value exposed
@secure()
param adminPassword string = 'DefaultPassword123'

// ✅ CORRECT: No default for secure parameters
@secure()
param adminPassword string
```

### CI/CD Linting Enforcement

Add this step to `.github/workflows/ci-cd.yml`:

```yaml
- name: Bicep Linting & Validation
  run: |
    echo "🔍 Running Bicep linting..."

    # Build and lint
    az bicep build --file infrastructure/bicep/main.bicep --outdir /tmp 2>&1 | tee /tmp/bicep-lint.txt

    # Fail on warnings (treat warnings as errors)
    if grep -iE "warning|error" /tmp/bicep-lint.txt | grep -v "Tips:"; then
      echo "❌ Bicep linting failed (warnings/errors detected)"
      exit 1
    fi

    echo "✅ Bicep linting passed"

    # Validate what-if
    echo "📋 Running deployment what-if..."
    az deployment sub what-if \
      -l uksouth \
      -f infrastructure/bicep/main.json \
      -p environment=dev \
      --query "changes[].[type, after.type]" || exit 1

    echo "✅ Deployment validation passed"
```

## Validation & Compliance

### Compliance Checklist

Before submitting any Bicep code:

- [ ] Code builds successfully (`az bicep build`)
- [ ] Linting passes (no warnings/errors)
- [ ] No unused parameters
- [ ] API versions verified to exist (not preview unless justified)
- [ ] `az deployment what-if` run at the correct scope with sample parameters
- [ ] Accompanying `.bicepparam` files updated
- [ ] Module versions pinned
- [ ] No secrets in outputs
- [ ] Regional availability validated (if targeting specific Azure region)
- [ ] **API versions verified** in target region(s) using ARM template reference or `az provider show`
- [ ] **Role definition IDs verified** with microsoft.learn.mcp MCP or Azure RBAC documentation
- [ ] **Required parameters documented** with `@description()` - all mandatory fields per ARM schema included
- [ ] **Parameter defaults validated** - not production secrets, safe for test environments
- [ ] **API documentation checked** - unusual required parameters documented (e.g., administratorLoginPassword)
