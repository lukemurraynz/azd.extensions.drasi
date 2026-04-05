# Troubleshoot Failures

## Purpose

Diagnose and recover from azd command failures, stuck deployments, and common error patterns.

---

## Common azd Exit Codes and Recovery

| Exit Code | Meaning | Common Cause | Recovery |
|-----------|---------|--------------|----------|
| 0 | Success | N/A | Proceed to next step |
| 1 | Generic failure | Many possible causes | Check logs; see below |
| 2 | Invalid parameters | Missing or incorrect args | Verify `azure.yaml`, env vars |
| 127 | Command not found | Missing CLI tool | Install missing dependency |
| 130 | Interrupted (Ctrl+C) | User cancelled | Safe to retry |

---

## Step 1: Increase Logging and Retry 🔍

Get more diagnostic information.

**Commands:**
```powershell
# Re-run the failed command with debug output
azd provision --debug 2>&1 | Tee-Object -FilePath "azd-debug-$(Get-Date -Format 'yyyyMMdd-HHmmss').log"

# OR for deployment
azd deploy --debug 2>&1 | Tee-Object -FilePath "azd-deploy-debug-$(Get-Date -Format 'yyyyMMdd-HHmmss').log"

# Check Azure CLI logs (if azd delegates to az)
# Logs may be in ~/.azure/azure-cli.log (Linux/Mac) or %APPDATA%\.azure\azure-cli.log (Windows)
```

**What to look for in logs:**
- Red text or `Error:` markers
- Stack traces (Python tracebacks)
- Azure API error codes (e.g., `ResourceTypeNotSupported`)
- Timeout messages or hanging operations

**🛑 STOP**: Review the logged error. If unclear, proceed to Step 2.

---

## Step 2: Check Prerequisites and Permissions ✅

Verify system and Azure state.

**Commands:**
```powershell
# Verify Azure login and subscription
Write-Host "=== Azure Authentication ===" -ForegroundColor Cyan
az account show --output table
az account get-access-token --query @.expiresOn
# If token expires soon, run: az login

# Verify permissions for resource group
Write-Host "=== Azure Permissions ===" -ForegroundColor Cyan
$rg = (azd env get-values | Select-String "RESOURCE_GROUP|resourceGroupName" | Select-Object -First 1) -replace ".*=", ""
az role assignment list --resource-group "$rg" --output table

# Check your role
$myId = az account show --query "user.name"
az role assignment list --assignee "$myId" --resource-group "$rg" --output table

# Verify tools are up-to-date and compatible
Write-Host "=== Tool Versions ===" -ForegroundColor Cyan
azd version
az version
az bicep version
kubectl version --client
docker version
docker info

# Check network connectivity to Azure endpoints
Write-Host "=== Network Connectivity ===" -ForegroundColor Cyan
Test-Connection -ComputerName "wss.azure.com" -Count 1  # Azure default endpoint
```

**Expected Output:**
```
Current subscription: <your-sub>
User: <your-email>
Role: Contributor (or Owner)

azd: <installed-version>
az: 2.50.0
```

**🛑 STOP**: If any of these fail:
- **No login**: Run `azd auth login`
- **No Contributor role**: Request IAM permissions from subscription owner
- **Old tool versions**: Update tools (azd, az, kubectl)
- **Network blocked**: Check firewall/VPN rules

**Success Criteria:**
- [ ] `az account show` displays your account
- [ ] You have `Contributor` or higher role on target resource group
- [ ] Tooling is current/stable (`azd version`, `az version`) and no upgrade warning blocks deployment
- [ ] Docker daemon is reachable (`docker info` succeeds)
- [ ] Network connectivity to Azure is working

---

## Step 3: Check azure.yaml Configuration 📋

Validate project configuration.

**Commands:**
```powershell
# Validate azure.yaml syntax
Write-Host "Validating azure.yaml..." -ForegroundColor Cyan
$yamlPath = "azure.yaml"
if (!(Test-Path $yamlPath)) {
  Write-Error "azure.yaml not found in $PWD"
  exit 1
}

# Check for required fields
$content = Get-Content $yamlPath -Raw
$required = @('name', 'metadata', 'services')
$required | ForEach-Object {
  if ($content -match $_) {
    Write-Host "✓ Found: $_" -ForegroundColor Green
  } else {
    Write-Error "Missing: $_"
  }
}

# List services
Write-Host "Services in azure.yaml:" -ForegroundColor Cyan
Select-String "services:" -A 20 $yamlPath

# Validate Bicep infrastructure
Write-Host "Validating Bicep infrastructure..." -ForegroundColor Cyan
az bicep build --file "<path-to-main.bicep>"

if ($LASTEXITCODE -ne 0) {
  Write-Error "Bicep validation failed"
  exit 1
}
```

**Common Issues:**
| Issue | Fix |
|-------|-----|
| `azure.yaml not found` | Create from template: `azd init --template <template>` |
| `Invalid YAML syntax` | Use YAML linter; check indentation (spaces, not tabs) |
| `Service path doesn't exist` | Update service `project` paths to match actual directories |
| `Bicep file invalid` | Run `az bicep build <file>` to see specific errors |

**Success Criteria:**
- [ ] `azure.yaml` exists and is valid YAML
- [ ] All services have valid `project` paths
- [ ] `az bicep build` succeeds without errors

---

## Step 4: Validate Environment Variables 🔧

Check azd environment configuration.

**Commands:**
```powershell
# List all environment variables
Write-Host "=== azd Environment Variables ===" -ForegroundColor Cyan
azd env get-values | Sort-Object

# Check critical variables
Write-Host "=== Critical Variables ===" -ForegroundColor Cyan
$critical = @('AZURE_SUBSCRIPTION_ID', 'AZURE_ENV_NAME', 'AZURE_LOCATION')
azd env get-values | Where-Object { $_ -match ($critical -join '|') }

# Validate values
$env = @{}
azd env get-values | ForEach-Object {
  if ($_ -match '(.+)=(.*)') {
    $env[$matches[1].Trim()] = $matches[2].Trim('"')
  }
}

if ([string]::IsNullOrEmpty($env['AZURE_SUBSCRIPTION_ID'])) {
  Write-Error "AZURE_SUBSCRIPTION_ID is not set"
}

if ([string]::IsNullOrEmpty($env['AZURE_LOCATION'])) {
  Write-Error "AZURE_LOCATION is not set"
}

# Verify location is valid
Write-Host "Valid Azure locations:" -ForegroundColor Cyan
az account list-locations --query "[].name" | Select-Object -First 5
```

**Common Issues:**
| Issue | Fix |
|-------|-----|
| `AZURE_SUBSCRIPTION_ID empty` | Run `azd env set AZURE_SUBSCRIPTION_ID <sub-id>` |
| `AZURE_LOCATION invalid` | Run `az account list-locations --query "[].name"` to find valid region |
| `Environment not selected` | Run `azd env select <env-name>` or `azd env new <env-name>` |

**Success Criteria:**
- [ ] `AZURE_SUBSCRIPTION_ID` is set to a valid UUID
- [ ] `AZURE_LOCATION` is a valid Azure region (e.g., `eastus`)
- [ ] `AZURE_ENV_NAME` is set (e.g., `dev`)

---

## Step 5: Diagnose Specific Failures 🔍

Match your error to a known pattern.

### Provisioning Failures

**Error: `ResourceTypeNotSupported`**
```
Error: The template deployment failed with code 'InvalidTemplateDeployment'.
Details: The template defines an invalid value for the resource provisioned.
```
**Cause**: API version doesn't support the resource type in this region.
**Fix**:
```powershell
az provider show --namespace "Microsoft.YourService" --query "resourceTypes[?resourceType=='YourResource'].apiVersions"
```

**Error: `Forbidden` or `Authorization failed`**
```
Error: The client '...' with object id '...' does not have authorization to perform action
```
**Cause**: Insufficient IAM permissions.
**Fix**:
```powershell
# Request role assignment from subscription owner:
# - Minimum: Contributor on resource group
# - Better: Owner on subscription
az role assignment list --scope "/subscriptions/$subscriptionId" --assignee "$yourEmail"
```

**Error: `Timeout waiting for resource`**
```
Timed out waiting for resource Microsoft.Compute/virtualMachines/...
```
**Cause**: Resource creation is slow or quota exceeded.
**Fix**:
```powershell
# Check Azure quotas
az account show --query "subscriptionId" # Verify quota in Portal
# Higher-priority alternative: Use smaller SKU or different region
```

**Error: `NameUnavailable` (App Configuration / Key Vault)**
```
Error: The resource name '<name>' is already in use.
```
**Cause**: Global name collision, commonly from soft-deleted resources retained by Azure.
**Fixes**:
```powershell
# 1. Check currently selected environment and candidate names
azd env list
azd env get-values | Select-String "APPCONFIG_NAME_SUFFIX|KEYVAULT_NAME_SUFFIX|AZURE_ENV_NAME"

# 2. Prefer full cleanup if this environment was recently torn down
azd down --purge

# 3. If reusing env names, rotate suffixes and retry preview
azd env set APPCONFIG_NAME_SUFFIX "<new-unique-suffix>"
azd env set KEYVAULT_NAME_SUFFIX "<new-unique-suffix>"
azd provision --preview
```

**Error: `ServerIsBusy` (PostgreSQL Flexible Server)**
```
Code: ServerIsBusy
Message: The server is busy. Please retry the operation later.
```
**Cause**: Transient Azure control-plane saturation; operation is throttled temporarily.
**Fix (bounded exponential backoff)**:
```powershell
$maxAttempts = 5
for ($attempt = 1; $attempt -le $maxAttempts; $attempt++) {
  Write-Host "Provision attempt $attempt/$maxAttempts..." -ForegroundColor Cyan
  azd provision
  if ($LASTEXITCODE -eq 0) { break }

  $delaySeconds = [Math]::Min(300, [Math]::Pow(2, $attempt) * 15)
  Write-Host "Provision failed. Waiting $delaySeconds seconds before retry..." -ForegroundColor Yellow
  Start-Sleep -Seconds $delaySeconds
}
```
**Do not retry blindly** if errors are authorization/configuration related (`Forbidden`, `InvalidTemplate`, missing parameters). Fix root cause first.

### Deployment Failures

**Error: `ImagePullBackOff`**
```
kubectl get pods -n <your-namespace>
# Pods stuck in ImagePullBackOff state
```
**Cause**: AKS cannot pull image from ACR.
**Fixes**:
```powershell
# 1. Verify image exists in ACR
az acr repository list --name "<acr-name>" --output table

# 2. Verify ACR is accessible from AKS
# AKS managed identity must have AcrPull role on ACR

# 3. Manually pull to test
az acr login --name "<acr-name>"
docker pull "<acr-name>.azurecr.io/api:<tag>"

# 4. Restart pods to retry pull
kubectl rollout restart deployment/api-deployment -n <your-namespace>
```

**Error: `denied: requested access to the resource is denied` (ACR push/pull)**
```
docker push <acr>.azurecr.io/api:<tag>
# denied: requested access to the resource is denied
```
**Cause**: Missing/expired ACR auth or insufficient role assignment (`AcrPush`/`AcrPull`).
**Fixes**:
```powershell
# 1. Refresh login
az acr login --name "<acr-name>"

# 2. Validate current principal and role assignment on ACR scope
$principal = az account show --query "user.name" -o tsv
$acrId = az acr show --name "<acr-name>" --query "id" -o tsv
az role assignment list --assignee $principal --scope $acrId --output table

# 3. Re-push and stop deployment if push fails
docker push "<acr-name>.azurecr.io/api:<tag>"
if ($LASTEXITCODE -ne 0) { throw "ACR push failed; stop deployment." }
```

**Error: `Services stuck with <pending> ExternalIP`**
```
kubectl get svc -n <your-namespace>
# SERVICE             TYPE           EXTERNAL-IP   PORT(S)
# <api-service-name>  LoadBalancer   <pending>     8080:30123/TCP
```
**Cause**: Azure LoadBalancer IP allocation slow or quota reached.
**Fix**:
```powershell
# Wait 2-5 minutes (normal for first deployment)
kubectl get svc -n <your-namespace> --watch

# If still <pending> after 5 mins:
# 1. Check Azure quotas (e.g., public IP limits)
# 2. Try smaller cluster or delete unused resources
# 3. Manually allocate: az network public-ip create ...
```

**Error: `CORS policy: No Access-Control-Allow-Origin header`**
```
Browser console: Access to XMLHttpRequest blocked by CORS policy
```
**Cause**: API missing CORS configuration or not restarted after ConfigMap update.
**Fix**:
```powershell
# 1. Verify ConfigMap has correct CORS origins
kubectl get configmap app-config -n <your-namespace> -o yaml | grep -A 5 "Cors__AllowedOrigins"

# 2. Verify API sees the config
kubectl logs -l app=<api-app-label> -n <your-namespace> | grep -i cors

# 3. Restart pods to pick up ConfigMap changes
kubectl rollout restart deployment/<api-deployment-name> -n <your-namespace>

# 4. Wait for new pods and test preflight
$apiUrl = "http://<api-service-name>.<your-namespace>.svc.cluster.local:8080"
curl -i -X OPTIONS "$apiUrl/api/v1/alerts" -H "Origin: http://localhost:3000" | grep Access-Control
```

**Error: `secrets \"...\" not found` or Key Vault secret lookup failure**
```
Error: Failed to retrieve secret from Key Vault ...
```
**Cause**: Missing RBAC on Key Vault, secret missing, or fallback not configured.
**Fixes**:
```powershell
# 1. Check Key Vault secret read path
az keyvault secret show --vault-name "<kv-name>" --name "<secret-name>"

# 2. Validate fallback path exists in azd environment
azd env get-values | Select-String "POSTGRES_ADMIN_PASSWORD|DB_PASSWORD"

# 3. Fail fast if both paths are empty (do not apply unresolved placeholders)
if ([string]::IsNullOrEmpty($env:POSTGRES_ADMIN_PASSWORD) -and [string]::IsNullOrEmpty($env:DB_PASSWORD)) {
  throw "No database secret resolved from Key Vault or azd environment."
}
```

**Error: `namespaces \"<name>\" not found` during `kubectl apply`**
```
Error from server (NotFound): namespaces "<your-namespace>" not found
```
**Cause**: Namespaced resources applied before namespace creation.
**Fix**:
```powershell
kubectl create namespace <your-namespace> --dry-run=client -o yaml | kubectl apply -f -
kubectl get namespace <your-namespace>
# Retry manifest apply only after namespace exists
kubectl apply -f k8s/rbac.yaml
```

**Error: `Executable doesn't exist` (Playwright smoke tests)**
```
browserType.launch: Executable doesn't exist at ...
```
**Cause**: Playwright browser runtime not installed on current machine/agent.
**Fix**:
```powershell
npx playwright install chromium-headless-shell
npx playwright test
```

### Cleanup Failures

**Error: `azd down Exit Code: 1`**
```
Deleting resource group...
Error: Failed to delete some resources.
```
**Causes & Fixes**:

1. **K8s resources not cleaned up first**:
   ```powershell
   kubectl delete namespace <your-namespace> --force --grace-period=0
   azd down --no-prompt  # Retry
   ```

2. **Resource locks prevent deletion**:
   ```powershell
   az lock list --resource-group "<rg>" --output table
   az lock delete --ids "<lock-id>"
   azd down --no-prompt  # Retry
   ```

3. **Database deletion timeout**:
   ```powershell
   # Wait for database to release locks, then retry
   Start-Sleep -Seconds 30
   azd down --no-prompt
   
   # OR manual deletion
   az group delete --name "<rg>" --yes
   ```

4. **Permissions removal in progress**:
   ```powershell
   # If you just removed yourself from the role, re-add temporarily
   az role assignment create --assignee "<your-email>" --role "Owner" --scope "/subscriptions/<sub-id>"
   azd down --no-prompt
   ```

---

## Step 6: Check Logs and Events 📊

Deep dive into what actually happened.

**Commands:**
```powershell
# Check Kubernetes events
Write-Host "=== Kubernetes Events ===" -ForegroundColor Cyan
kubectl get events -n <your-namespace> --sort-by='.lastTimestamp'

# Check pod logs
Write-Host "=== Pod Logs ===" -ForegroundColor Cyan
kubectl logs -l app=<api-app-label> -n <your-namespace> --tail=50

# Check deployment status
kubectl describe deployment <api-deployment-name> -n <your-namespace>

# Check service status
kubectl describe service <api-service-name> -n <your-namespace>

# Azure resource deployment logs (if provisioning failed)
$rg = (azd env get-values | Select-String "RESOURCE_GROUP|resourceGroupName" | Select-Object -First 1) -replace ".*=", ""
az deployment group list --resource-group "$rg" --query "[0].{name:name, state:properties.provisioningState}" --output table

# Get detailed error from last deployment
az deployment group show --resource-group "$rg" --name "<deployment-name>" --query "properties.error"
```

**What to look for:**
- Pod `CrashLoopBackOff` or `ImagePullBackOff` → check logs
- Events with `Warning` or `Error` → actionable diagnostics
- Deployment `provisioningState` = `Failed` → check `error` property

---

## Step 7: Reset and Retry 🔄

If diagnosis is unclear, perform a clean reset.

**Commands:**
```powershell
# Option 1: Light reset (keep Azure infrastructure)
Write-Host "Clearing Kubernetes resources..." -ForegroundColor Cyan
kubectl delete namespace <your-namespace> --ignore-not-found=true
kubectl wait --for=delete namespace/<your-namespace> --timeout=30s 2>/dev/null || $true

# Wait, then re-deploy
Start-Sleep -Seconds 5
azd deploy

# Option 2: Full reset (delete and re-provision)
Write-Host "Full reset: cleanup + re-provision..." -ForegroundColor Cyan
azd down --no-prompt

# Wait for deletion
Start-Sleep -Seconds 30

# Re-provision
azd provision
azd deploy
```

**Use light reset when**: Pods are stuck or configuration is stale.
**Use full reset when**: Infrastructure state is corrupted or multiple issues exist.

---

## Step 8: Enable Debug Logging and Continuous Monitoring 🔧

For persistent or unclear issues, enable verbose logging and use AZD monitoring tools.

**Enable Debug Logging:**
```powershell
# Enable AZD debug output
$env:AZD_DEBUG = "true"
$env:AZURE_LOG_LEVEL = "DEBUG"

# Re-run the command that failed with verbose output
azd provision --debug
# OR
azd deploy --debug

# Disable debug when done
$env:AZD_DEBUG = "false"
$env:AZURE_LOG_LEVEL = "INFO"
```

**Use AZD Monitor for Continuous Health Checks:**
```powershell
# Monitor application health in real-time (if configured)
Write-Host "Starting AZD monitor..." -ForegroundColor Cyan
azd monitor

# This opens a dashboard showing:
# - Application Insights metrics (if configured)
# - Log streams
# - Performance data
# - Health status
```

**Retrieve Complete Logs:**
```powershell
# Export logs for detailed analysis
$rg = (azd env get-values | Select-String "RESOURCE_GROUP").Line -replace ".*=", ""

# Get deployment logs
az deployment group list --resource-group "$rg" --query "[0].{name:name, state:properties.provisioningState}" --output table

# Get detailed error from failed deployment
az deployment group show --resource-group "$rg" --name "<deployment-name>" --query "properties.error" -o json > deployment-error.json
Write-Host "Exported deployment error to deployment-error.json"
```

**Success Criteria:**
- [ ] Debug logging shows detailed operation timeline
- [ ] `azd monitor` provides visibility into application health (if configured)
- [ ] Logs exported for external troubleshooting if needed

---

## Common Recovery Patterns

| Symptom | Recovery |
|---------|----------|
| Command hangs (no output for >2 mins) | Press Ctrl+C, retry with `--debug` flag |
| Partial deployment (some pods running, some failed) | `kubectl get pods -A` to diagnose; restart failed pods |
| `<pending>` resources after 10 mins | Check quotas, Azure service status, or delete and redeploy |
| `NameUnavailable` on AppConfig/KeyVault | Run `azd down --purge` or rotate `APPCONFIG_NAME_SUFFIX`/`KEYVAULT_NAME_SUFFIX` |
| `ServerIsBusy` during provisioning | Retry with bounded exponential backoff (max attempts) |
| Playwright executable missing | `npx playwright install chromium-headless-shell` before smoke tests |
| Cleanup fails repeatedly | Delete resource group manually: `az group delete --name "$rg" --yes` |

---

## Escalation Path

If recovery steps don't resolve the issue:

1. **Check Microsoft docs**: https://learn.microsoft.com/azure/developer/azure-developer-cli/
2. **Search Azure issues**: https://github.com/Azure/azure-dev/issues
3. **File a bug**: Include `azd version`, `az version`, `azure.yaml`, and `--debug` logs
4. **Contact Azure support**: For infrastructure-level failures (quotas, permissions, API issues)

---

## Next Steps

✅ After recovery, return to:
- [provision-safely](provision-safely.md) (if provisioning failed)
- [deploy-efficiently](deploy-efficiently.md) (if deployment failed)
- [cleanup-completely](cleanup-completely.md) (if cleanup failed)
