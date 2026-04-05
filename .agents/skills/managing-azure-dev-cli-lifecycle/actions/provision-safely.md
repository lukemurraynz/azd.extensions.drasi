# Provision Safely

## Purpose

Select or initialize an AZD template, validate prerequisites, review infrastructure changes, and provision Azure resources with confidence. Includes what-if analysis, environment setup, and rollback awareness.

**Supported Infrastructure Types:**
- Kubernetes (AKS)
- App Service (Web Apps, API Apps)
- Azure Container Instances
- Azure Static Web Apps
- Service Fabric
- Or any custom infrastructure defined in Bicep/Terraform

---

## Flow

### Step 0: Select or Initialize Azure Developer CLI Template 🎯

Start with an existing AZD template or initialize a new one.

**Commands:**
```powershell
# Option 1: Initialize from Awesome AZD gallery (recommended for first-time)
Write-Host "Choose a template from https://github.com/Azure/awesome-azd" -ForegroundColor Cyan
azd init -t hello-azd                # Simple Node.js example
# OR
azd init -t todo-nodejs-coredata    # Todo app with Cosmos DB
# OR
azd init -t <template-name>         # Any template from Awesome AZD

# Option 2: Initialize an existing project as AZD-compatible
Write-Host "Converting existing project..." -ForegroundColor Cyan
azd init                            # Interactive mode; asks for project name and location
```

**Success Criteria:**
- [ ] Template selected or project initialized
- [ ] `azure.yaml` file created in project root
- [ ] Infrastructure folder exists (`./infra` or `./infrastructure`)
- [ ] Application code in expected location

---

### Step 1: Validate Prerequisites ✅

Check that all tools are installed and accessible.

**Commands:**
```powershell
# Check azd
azd version

# Check Azure CLI
az version

# Check Bicep (required for bicep-based infrastructure)
az bicep version

# Check kubectl (if using Kubernetes)
kubectl version --client

# Check docker (if using container images)
docker version
docker info
```

**🛑 STOP**: If any command is missing or out-of-date, install before proceeding.

**Success Criteria:**
- [ ] `azd version` returns a supported/stable release (prefer latest stable)
- [ ] `az version` returns a current Azure CLI release
- [ ] `az bicep version` confirms Bicep is installed (usually bundled with Azure CLI)
- [ ] `kubectl version` is compatible with target AKS cluster
- [ ] `docker version` is installed (only required if building container images)
- [ ] `docker info` succeeds (daemon is running)

---

### Step 2: Authenticate with Azure ✅

Set up Azure credentials for provisioning.

**Commands:**
```powershell
# Log in to Azure (one-time, or refresh if expired)
azd auth login

# Verify logged-in account
az account show

# List available subscriptions
az account list --output table

# Set correct subscription (if needed)
az account set --subscription "<subscription-id-or-name>"

# Verify azd can access Azure resources
azd env list
```

**🛑 STOP**: Verify the subscription shown is the correct target.

**Success Criteria:**
- [ ] `az account show` displays your account
- [ ] `az account show --query "{ id: id, name: name }"` shows the correct subscription
- [ ] `az account show --query subscriptionId` returns a valid UUID
- [ ] `azd env list` succeeds without errors

---

### Step 3: Set Up Environment Variables ✅

Configure azd environment for the target deployment.

**Commands:**
```powershell
# Create new or select existing environment
azd env new <env-name>                    # Create (e.g., azd env new dev)
# OR
azd env select <env-name>                 # Select existing

# List all configured environments
azd env list

# Set critical environment variables
azd env set AZURE_LOCATION "<region>"      # e.g., eastus, westeurope, southeastasia
azd env set AZURE_ENV_NAME "<env-name>"    # e.g., dev, staging, prod

# Optional but recommended for globally-unique resources in iterative environments
azd env set APPCONFIG_NAME_SUFFIX "<unique-suffix>"   # e.g., yk9wdya9y65v
azd env set KEYVAULT_NAME_SUFFIX "<unique-suffix>"

# Verify environment
azd env get-values | grep -E "AZURE_LOCATION|AZURE_ENV_NAME|AZURE_SUBSCRIPTION_ID"
```

**Expected Output:**
```
AZURE_ENV_NAME=dev
AZURE_LOCATION=<region>
AZURE_SUBSCRIPTION_ID=<your-subscription-id>
```

**🛑 STOP**: Verify the environment matches your intent (dev/staging/prod, correct region).

**Success Criteria:**
- [ ] `azd env list` shows your target environment
- [ ] `azd env get-values` includes AZURE_LOCATION and AZURE_ENV_NAME
- [ ] Subscription ID is correct (matches step 2)
- [ ] Location is a valid Azure region (az account list-locations --query "[].name")
- [ ] Optional uniqueness suffixes are set if reusing environment names frequently

---

### Step 4: Review Infrastructure (What-If Analysis) 🔍

Dry-run the infrastructure deployment to preview changes.

**Commands:**
```powershell
# Run what-if analysis (no resources created)
azd provision --preview

# OR for more verbose output with timestamps
azd provision --preview 2>&1 | Tee-Object -FilePath "provision-preview-$(Get-Date -Format 'yyyyMMdd-HHmmss').log"

# Optional: if prior runs failed with App Configuration name collisions, check soft-deleted resources
az appconfig list-deleted --output table
```

**Expected Output:**
```
Provisioning infrastructure...
Deployment preview (what-if):
  Create: Microsoft.Compute/virtualMachines
  Create: Microsoft.Database/servers
  Modify: Microsoft.Network/virtualNetworks
  ...
Preview complete. No resource changes made.
```

**🛑 STOP**: Review the changes. Ensure:
- No unintended deletions or modifications
- All expected resources are listed
- No permission errors (would show `Forbidden` or similar)

**Common Issues:**
| Issue | Cause | Resolution |
|-------|-------|-----------|
| `ResourceTypeNotSupported` | API version doesn't exist for region | Check `az provider show --namespace <provider>` |
| `Unauthorized` or `Forbidden` | Missing IAM permissions | Verify role assignment (Owner, Contributor, or scoped rights) |
| `InvalidTemplateDeployment` | Bicep syntax error | Run `az bicep build <file>` locally to validate |
| `ParameterNotFound` | Missing required parameter in azure.yaml | Check azure.yaml against bicep module parameters |
| `NameUnavailable` | Global resource name blocked by soft-deleted service | Purge deleted resource or rotate suffix env vars, then retry |

**Success Criteria:**
- [ ] What-if completes without `Forbidden` or permission errors
- [ ] All expected resources are in "Create" or "Modify" (no unintended deletes)
- [ ] No API version or syntax errors
- [ ] Deployment would fit within subscription quotas (if visible)

---

### Step 5: Understand Rollback Behavior ⚠️

Important: azd does **not** guarantee automatic rollback on failure.

**Key Points:**
- If `azd provision` fails partway through, some resources may be created and others missing
- You are responsible for cleanup via `azd down` or manual deletion
- Test in a dev environment first if provisioning is critical

**Prevention:**
- Use `--preview` before every `provision` to catch errors early
- Monitor the Azure Portal during `provisioning...` to see real-time progress
- Have `azd down` ready in case of failure

---

### Step 6: Provision Infrastructure 🚀

Apply the infrastructure deployment.

**Commands:**
```powershell
# Provision (this will create Azure resources)
azd provision

# Output will show progress
# Deployment provisioned successfully
# .azd/environment directory updated
```

**Expected Output:**
```
Provisioning infrastructure...
Infrastructure provisioning complete.
```

**🛑 STOP**: If provisioning fails, see [troubleshoot-failures](troubleshoot-failures.md). Do **not** proceed to deployment.

**Success Criteria:**
- [ ] Exit code = 0 (success)
- [ ] `Infrastructure provisioning complete` message
- [ ] All resources appear in target resource group (check Azure Portal)
- [ ] `.azd/` directory has updated DOTENV environment state

---

### Step 7: Validate Provisioned Resources ✅

Verify infrastructure is ready for deployment.

**Commands:**
```powershell
# Get provisioned resource group
$rg = (azd env get-values | Select-String "RESOURCE_GROUP|resourceGroupName" | Select-Object -First 1) -replace ".*=", ""
Write-Host "Target Resource Group: $rg"

# List all resources in the group
az resource list --resource-group "$rg" --output table

# Check for key resources (examples for your stack)
az container registry list --resource-group "$rg" --output table       # If using ACR
az aks list --resource-group "$rg" --output table                      # If using AKS
az postgres server list --resource-group "$rg" --output table          # If using PostgreSQL single-server
```

**Expected State:**
- All infrastructure resources (AKS, ACR, databases, Key Vault, etc.) in "Succeeded" state
- AKS cluster has active node pool
- ACR can be accessed from deployed pods

**🛑 STOP**: If any resource is missing or in error state, troubleshoot before continuing to deployment.

**Success Criteria:**
- [ ] Resource group exists and all resources listed
- [ ] AKS cluster is in "Succeeded" state
- [ ] ACR exists and is ready (status = "Enabled")
- [ ] Database (PostgreSQL, Cosmos DB, etc.) is ready
- [ ] Key Vault exists and is accessible

---

### Step 7: Validate azure.yaml Schema (Optional but Recommended)

Ensure your `azure.yaml` matches the official AZD schema.

**Commands:**
```powershell
# Validate your azure.yaml against official schema
Write-Host "Validating azure.yaml schema..." -ForegroundColor Cyan

# Verify Bicep files compile (validates azure.yaml infra.parameters section)
az bicep build --file ./infra/main.bicep --outdir ./infra/
if ($LASTEXITCODE -ne 0) {
  throw "Bicep validation failed. Check azure.yaml infra.parameters section."
}

Write-Host "✓ azure.yaml and Bicep are valid" -ForegroundColor Green
```

**Reference:** https://github.com/Azure/azure-dev/blob/main/schemas/v1.0/azure.yaml.json

**Success Criteria:**
- [ ] `az bicep build` completes without errors
- [ ] All parameters in `azure.yaml` match Bicep module definitions
- [ ] No schema validation errors from IDE or tools

---

## Common Patterns

### Multi-environment Setup

```powershell
# Create dev environment
azd env new dev
azd env set AZURE_LOCATION <dev-region>
azd provision

# Later, create staging environment
azd env new staging
azd env set AZURE_LOCATION <staging-region>
azd provision

# Switch between them
azd env select dev      # Deploy to dev
azd env select staging  # Deploy to staging
```

### Validate Permissions Before Provisioning

```powershell
# Check if you have permission to create resources
az deployment group validate \
  --resource-group "<your-resource-group>" \
  --template-file "infrastructure/bicep/main.bicep" \
  --parameters "infrastructure/bicep/main.bicepparam"
```

### Automate Provisioning in CI/CD

```powershell
# Set all env vars from GitHub Secrets / Azure DevOps
$env:AZURE_ENV_NAME = $EnvName
$env:AZURE_LOCATION = $Location
$env:AZURE_SUBSCRIPTION_ID = $SubscriptionId

# Authenticate with service principal
azd auth login --client-id "$ClientId" --client-secret "$ClientSecret" --tenant-id "$TenantId"

# Provision
azd provision
```

---

## Next Steps

✅ After successful provisioning, proceed to [deploy-efficiently](deploy-efficiently.md) to build and deploy your application.

If provisioning fails, see [troubleshoot-failures](troubleshoot-failures.md).
