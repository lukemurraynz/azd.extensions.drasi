# Cleanup Completely

## Purpose

Safely delete Azure resources and local state with understanding of `azd down --purge` semantics, cleanup ordering, and stranded resource detection.

---

## Flow

### Step 1: Understand `azd down` vs `azd down --purge` 📚

Know what gets deleted at each level.

| Scope | `azd down` | `azd down --purge` |
|-------|-----------|------------------|
| Azure Infrastructure | ✅ DELETED | ✅ DELETED |
| Azure Storage (DBs, etc.) | ✅ DELETED | ✅ DELETED |
| Kubernetes Manifests | ✅ DELETED | ✅ DELETED |
| `.azd/` local environment state | ✅ KEPT | ❌ DELETED |
| Configuration secrets | ✅ KEPT | ❌ DELETED |

**When to use each:**

- **`azd down`** (default): Clean up Azure resources but keep local development environment for re-deployment
  - Use when: Tearing down a dev/test environment but may redeploy to same environment later
  
- **`azd down --purge`** (full reset): Delete everything, including local configuration
  - Use when: Switching subscriptions, disposing environment completely, or resetting state after failures

---

### Step 2: Diagnose Current Cleanup Status 🔍

Understand what resources exist before cleanup.

**Commands:**
```powershell
# Get resource group name from infrastructure configuration
# Resource group is derived from AZURE_ENV_NAME and projectName in azure.yaml
$azdEnv = @{}
azd env get-values | ForEach-Object {
  if ($_ -match '(.+)=(.*)') {
    $azdEnv[$matches[1].Trim()] = $matches[2].Trim('"')
  }
}

# The resource group name follows pattern: {projectName}-{environment}-rg
# (where environment = AZURE_ENV_NAME from azure.yaml parameters)
$resourceGroup = $azdEnv['resourceGroupName'] ?? "<your-project>-$($azdEnv['AZURE_ENV_NAME'])-rg"
Write-Host "Target Resource Group: $resourceGroup"

az resource list --resource-group "$resourceGroup" --output table

# Count resources by type
az resource list --resource-group "$resourceGroup" `
  --query "groupBy(type) | map({type: @[0].type, count: length(@)})" `
  --output table

# Check for Kubernetes resources (outside AKS cluster)
kubectl get all -A --output wide

# Check for persistent volumes that might hold data
kubectl get pvc -A
```

**🛑 STOP**: Review resource count and note:
- Total number of resources
- Any databases with data you need to backup
- Any persistent volumes with important data

**Common Resources in Your Stack:**
- Azure Kubernetes Service (AKS)
- Azure Container Registry (ACR)
- Azure Database for PostgreSQL
- Key Vault
- Application Insights
- Network resources (VNet, NSG, LoadBalancer)
- Storage accounts
- Managed identities

---

### Step 3: Back Up Critical Data (If Needed) 💾

Save important data before deletion.

**Commands:**
```powershell
# Backup PostgreSQL database
$dbHost = (azd env get-values | Select-String "POSTGRES_HOST").Line -replace ".*=", ""
$dbUser = (azd env get-values | Select-String "POSTGRES_USER").Line -replace ".*=", ""
$dbName = (azd env get-values | Select-String "POSTGRES_DB").Line -replace ".*=", ""

Write-Host "Backing up PostgreSQL database: $dbName"
pg_dump --host "$dbHost" --username "$dbUser" --database "$dbName" > "backup-$dbName-$(Get-Date -Format 'yyyyMMdd-HHmmss').sql"

# Verify backup
if (Test-Path "backup-*.sql") {
  Write-Host "Backup created successfully" -ForegroundColor Green
  Get-Item "backup-*.sql" | Format-Table LastWriteTime, Length
}

# Backup application configuration from Key Vault (if needed)
$kvName = (azd env get-values | Select-String "KEY_VAULT_NAME|keyVaultName").Line -replace ".*=", ""
az keyvault secret list --vault-name "$kvName" --output table
```

**🛑 STOP**: If you have critical data, ensure backup completes before proceeding.

**Success Criteria:**
- [ ] Database backup file created and non-empty
- [ ] Key Vault secrets documented (or exported)
- [ ] Backup files stored in version control or backup storage

---

### Step 4: Prepare for Cleanup ⚠️

Create a cleanup manifest to track what will be deleted.

**Commands:**
```powershell
# Create cleanup manifest
$cleanupLog = "cleanup-manifest-$(Get-Date -Format 'yyyyMMdd-HHmmss').txt"
Write-Host "Creating cleanup manifest: $cleanupLog"

# Capture current state
Write-Host "=== Pre-Cleanup Inventory ===" | Tee-Object -FilePath $cleanupLog -Append
az resource list --resource-group "$resourceGroup" --output table | Tee-Object -FilePath $cleanupLog -Append
kubectl get all -A | Tee-Object -FilePath $cleanupLog -Append

# Capture environment variables (without secrets)
Write-Host "=== Environment Configuration ===" | Tee-Object -FilePath $cleanupLog -Append
azd env get-values | Tee-Object -FilePath $cleanupLog -Append

# Calculate resource costs (approximate)
Write-Host "=== Resource Estimates ===" | Tee-Object -FilePath $cleanupLog -Append
Write-Host "Use Azure Cost Analysis for actual charges" | Tee-Object -FilePath $cleanupLog -Append
```

**Expected Output:**
```
=== Pre-Cleanup Inventory ===
NAME                          TYPE                                 LOCATION
aks-cluster                   Microsoft.ContainerService/...       <region>
acr-registry                  Microsoft.ContainerRegistry/...      <region>
postgres-server               Microsoft.DBforPostgreSQL/...        <region>
...

=== Environment Configuration ===
AZURE_SUBSCRIPTION_ID=...
AZURE_LOCATION=<region>
AZURE_ENV_NAME=dev
...
```

**Success Criteria:**
- [ ] Cleanup manifest created
- [ ] All resources listed and reviewed
- [ ] Environment variables captured

---

### Step 5: Stop Kubernetes Workloads 🛑

Gracefully shut down Kubernetes deployments before deleting infrastructure.

**Commands:**
```powershell
# Scale deployments to 0 (graceful shutdown)
Write-Host "Scaling down Kubernetes deployments..." -ForegroundColor Cyan
kubectl scale deployment <api-deployment-name> -n <your-namespace> --replicas=0
kubectl scale deployment <frontend-deployment-name> -n <your-namespace> --replicas=0

# Wait for graceful shutdown
Write-Host "Waiting for pods to terminate..." -ForegroundColor Gray
kubectl wait --for=delete pod -l app=<api-app-label> -n <your-namespace> --timeout=30s 2>/dev/null || $true
kubectl wait --for=delete pod -l app=<frontend-app-label> -n <your-namespace> --timeout=30s 2>/dev/null || $true

# Verify all pods are gone
kubectl get pods -n <your-namespace>
```

**Expected Output:**
```
No resources found in <your-namespace> namespace.
```

**🛑 STOP**: Ensure all pods are terminated before proceeding. This prevents orphaned processes.

**Success Criteria:**
- [ ] All pods in `<your-namespace>` namespace are gone (or only system pods remain)
- [ ] Exit code = 0

---

### Step 6: Clean Up Entra ID Resources (If Applicable) 🔐

`azd down` and `azd down --purge` do **not** delete Entra ID resources (app registrations, service principals, federated credentials). These must be cleaned up via a `predown` hook or manually.

**Check if your template creates Entra ID resources:**
```powershell
# List app registrations tagged with your environment
$envId = azd env get-values | Select-String "AZD_ENV_ID" | ForEach-Object { $_ -replace ".*=", "" }
az ad app list --filter "tags/any(t:t eq 'azd-env-id: $envId')" | ConvertFrom-Json | Select-Object displayName, appId
```

**If `predown` hook is configured:**
```powershell
# The hook runs automatically before deletion — verify it ran successfully
# (check output for "Predown cleanup complete")
# Then continue to Step 7
```

**If no `predown` hook exists (manual cleanup):**
```powershell
# Delete app registrations by display name pattern
$envName = azd env get-values | Select-String "AZURE_ENV_NAME" | ForEach-Object { $_ -replace ".*=", "" }
az ad app list --filter "startswith(displayName, '$envName')" | ConvertFrom-Json | ForEach-Object {
    Write-Host "Deleting: $($_.displayName)"
    az ad app delete --id $_.appId
}

# Permanently remove soft-deleted Entra ID resources
# (Required before redeploying with the same names)
az ad app list --filter "startswith(displayName, '$envName')" --show-deleted | ConvertFrom-Json | ForEach-Object {
    az ad app delete --id $_.appId --permanent
}
```

> If your template creates Entra ID resources and you don't have a `predown` hook, add one — see [authoring-hooks.md](../standards/authoring-hooks.md) and [common-traps.md](../standards/common-traps.md) Trap #14.

**Success Criteria:**
- [ ] No app registrations remain with environment-specific display names or `azd-env-id` tags
- [ ] Soft-deleted Entra ID resources permanently purged (prevents name conflicts on redeploy)

---

### Step 7: Delete Kubernetes Namespace (Optional) 🗑️

Clean up K8s resources before Azure infrastructure.

**Commands:**
```powershell
# Delete K8s namespace (cascades to all resources within, including PVCs, ConfigMaps, Secrets)
Write-Host "Deleting Kubernetes namespace..." -ForegroundColor Cyan
kubectl delete namespace <your-namespace> --ignore-not-found=true --grace-period=30

# Wait for namespace deletion
Write-Host "Waiting for namespace deletion..." -ForegroundColor Gray
kubectl wait --for=delete namespace/<your-namespace> --timeout=60s 2>/dev/null || $true

# Verify deletion
kubectl get namespace <your-namespace> 2>/dev/null
if ($?) {
  Write-Warning "Namespace still exists; may retry manually"
} else {
  Write-Host "Namespace deleted successfully" -ForegroundColor Green
}
```

**Expected Output:**
```
namespace "<your-namespace>" deleted
```

**🛑 STOP**: If namespace deletion hangs (>60s), it may have stuck finalizers. Manually check:
```powershell
kubectl api-resources --verbs=list --namespaced=true -n <your-namespace>
```

**Success Criteria:**
- [ ] Namespace deleted (no errors)
- [ ] `kubectl get namespace <your-namespace>` returns not-found error

---

### Step 8: Run Azure Cleanup with `azd down --purge` 🗑️

Delete Azure infrastructure, local state, and soft-deleted resources.

**Command (Recommended Default):**
```powershell
# This deletes Azure resources AND .azd/ local state AND soft-deleted App Config/Key Vault
Write-Host "Running cleanup: azd down --purge (recommended)..." -ForegroundColor Cyan
azd down --purge --no-prompt

# Expected output:
# Deleting resource group...
# Resource group deleted successfully.
# Removing environment configuration...
# Environment purged.
```

**Why `--purge` is default:**
- ✅ Avoids soft-delete issues (App Configuration, Key Vault retention)
- ✅ Clears all local state (fresh start)
- ✅ Cleanest full cleanup
- ✅ Prevents cost leaks from retained resources

**Alternative (rare):**
Only use `azd down` (without `--purge`) if you plan to re-deploy to the same environment immediately:
```powershell
# Deletes Azure resources but keeps .azd/environment for quick re-deploy
azd down --no-prompt
```

**🛑 STOP**: Monitor the deletion process. If it fails, see Step 9.

**Expected Output:**
```
Deleting resource group...
Resource group deleted successfully.

Removing resources...
✓ Deleted: Microsoft.Compute/virtualMachines
✓ Deleted: Microsoft.Network/virtualNetworks
...
All resources cleaned up.
```

**Common Issues:**
| Issue | Cause | Solution |
|-------|-------|----------|
| `Exit code: 1` | Partial delete failure (K8s resources not deleted first) | See Step 9: Recover from Partial Failure |
| `Timeout waiting for resource delete` | Resource busy or has locks | Check for resource locks: `az lock list --resource-group "$rg"` |
| `Forbidden: insufficient permissions` | Using wrong credentials | Verify `az account show` matches subscription owner |

**Success Criteria:**
- [ ] Exit code = 0
- [ ] `Resource group deleted successfully` message
- [ ] Azure Portal shows resource group no longer exists (may take 1-2 mins)

---

### Step 9: Verify Complete Cleanup ✅

Confirm all resources are deleted.

**Commands:**
```powershell
# Verify resource group no longer exists
$resourceGroup = (azd env get-values | Select-String "resourceGroupName").Line -replace ".*=", ""
$rg = az group show --name "$resourceGroup" 2>/dev/null
if ($rg) {
  Write-Warning "Resource group still exists: $resourceGroup"
} else {
  Write-Host "✓ Resource group deleted" -ForegroundColor Green
}

# Check for orphaned resources (should be empty)
az resource list --query "[?resourceGroup=='$resourceGroup']" --output table

# Verify ACR images are deleted (if using shared registry)
Write-Host "Checking for orphaned ACR images..." -ForegroundColor Cyan
$acrName = (azd env get-values | Select-String "acrName").Line -replace ".*=", ""
az acr repository list --name $acrName --output table 2>/dev/null || Write-Host "ACR already deleted or not accessible"

# Clean up local Docker images (optional)
Write-Host "Cleaning up local Docker images..." -ForegroundColor Cyan
docker image rm "$acrName.azurecr.io/<api-repository>:<tag-to-remove>" 2>/dev/null || $true
docker image rm "$acrName.azurecr.io/<frontend-repository>:<tag-to-remove>" 2>/dev/null || $true
```

**Expected Output:**
```
✓ Resource group deleted
<no resources found>
```

**🛑 STOP**: If resources still exist, escalate to Step 9.

**Success Criteria:**
- [ ] Resource group no longer exists
- [ ] `az resource list --resource-group <name>` returns empty or "not found"
- [ ] No orphaned ACR images (or acceptable to keep)

---

### Step 10: Recover from Partial Failure 🔧

If cleanup fails partway through.

**Commands:**
```powershell
# Check what's still in the resource group
$resourceGroup = (azd env get-values | Select-String "resourceGroupName").Line -replace ".*=", ""
$remaining = az resource list --resource-group "$resourceGroup" --output json | ConvertFrom-Json
Write-Host "Remaining resources ($($remaining.Count)):"
$remaining | Select-Object { $_.type }, { $_.name } | Format-Table

# If Kubernetes resources remain:
Write-Host "If K8s resources remain, delete namespace manually:" -ForegroundColor Yellow
kubectl delete namespace <your-namespace> --force --grace-period=0 2>/dev/null || $true

# If database remains due to lock:
Write-Host "Check for resource locks:" -ForegroundColor Yellow
az lock list --resource-group "$resourceGroup" --output table

if ($locks.Count -gt 0) {
  Write-Host "Remove locks manually:" -ForegroundColor Yellow
  az lock delete --ids <lock-id>
}

# Retry cleanup
Write-Host "Retrying azd down..." -ForegroundColor Cyan
azd down --force --no-prompt 2>&1 | Tee-Object -FilePath "cleanup-retry-$(Get-Date -Format 'yyyyMMdd-HHmmss').log"
```

**Common Recovery Patterns:**
| Symptom | Recovery Step |
|---------|---------------|
| `K8s resources not deleted` | Run `kubectl delete namespace <your-namespace> --force --grace-period=0` |
| `Database deletion timeout` | Check for locks: `az lock list --resource-group $rg`; remove if present |
| `ACR push/pull errors after delete` | Images may persist; manually delete: `az acr repository delete --name $acr --repository api` |
| `azd down fails repeatedly` | Delete resource group manually: `az group delete --name $rg --yes` |

---

### Step 11: Clean Up Local State (Optional) 🗑️

If using `azd down` (without `--purge`), manually reset state for fresh start.

**Commands:**
```powershell
# List current environments
azd env list

# Delete environment (if needed for fresh start)
azd env delete <env-name>

# Clear local Azure CLI cache (optional)
# (This removes all cached credentials; do only if needed)
# rm -r ~/.azure (on Linux/Mac)
# Remove-Item $env:APPDATA\.azure -Recurse (on Windows)
```

**When to do this:**
- You plan to avoid re-deploying to the same environment
- You're switching subscriptions and want a clean state
- Local environment files are corrupted or outdated

**Success Criteria:**
- [ ] `azd env list` no longer shows deleted environment
- [ ] Local `.azd/` directory cleaned up (if using `--purge`)

---

## Common Patterns

### Full Cleanup + Fresh Deployment (Recommended for Development)

```powershell
# Full cleanup with purge (clears soft-deletes)
azd down --purge --no-prompt

# Create new environment
$newEnv = "dev-$(Get-Date -Format 'yyyyMMddHHmm')"
azd env new $newEnv --no-prompt

# Provision fresh
azd provision

# Deploy
azd deploy
```

### Quick Redeploy (Same Environment)

Only use if re-deploying immediately to the same environment:

```powershell
# Fast cleanup (keeps .azd state)
azd down --no-prompt

# Wait for deletion
Start-Sleep -Seconds 60

# Provision and deploy
azd provision
azd deploy
```

### Cleanup with Manual Verification

```powershell
# Before cleanup
$resources = az resource list --resource-group "$rg" --output json | ConvertFrom-Json
Write-Host "About to delete $($resources.Count) resources. Continue? (Y/N)"
$confirm = Read-Host
if ($confirm -eq 'Y') {
  azd down --no-prompt
}
```

### Cleanup Failed Deployments

```powershell
# If provision or deploy failed, clean up partial resources
azd down --no-prompt

# Then retry
azd provision
azd deploy
```

---

## Post-Cleanup Checklist

- [ ] Resource group no longer exists in Azure Portal
- [ ] All ACR images deleted (or acceptable for retention)
- [ ] Database backups created (if needed)
- [ ] Local `.azd/` state cleaned or preserved (as intended)
- [ ] Cleanup manifest logged for audit trail
- [ ] All `<pending>` resources resolved

---

## Next Steps

After cleanup:
- ✅ To redeploy to same environment: `azd up`
- ✅ To create new environment: `azd env new <new-name>` → `azd up`
- ❌ To troubleshoot cleanup failure: See [troubleshoot-failures](troubleshoot-failures.md)
