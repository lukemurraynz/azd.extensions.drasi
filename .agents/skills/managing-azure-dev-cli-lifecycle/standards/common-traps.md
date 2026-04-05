# Common Traps & Pitfalls

Documented pitfalls in azd workflows and how to avoid them.

---

## 1. Soft-Delete Persistence After `azd down` (Without `--purge`)

**The Trap**: Azure services like App Configuration and Key Vault use soft-delete retention. Running `azd down` (without `--purge`) leaves soft-deleted resources that still incur costs.

**Example**:
```powershell
azd down --no-prompt
# App Configuration deleted, but soft-delete retained for 7-30 days
# Key Vault deleted, but soft-delete retained for 7-30 days
# Resources still consume storage; still block resource name reuse

azd up  # Try to re-deploy
# ERROR: Resource name already exists (soft-deleted copy blocking name reuse)
```

**Impact**: Cost leaks; resource name conflicts; unexpected failures on re-deployment; manual purge required.

**Prevention**:
- **Always use `azd down --purge`** (default recommendation)
- `--purge` fully deletes soft-deleted resources immediately
- Never use bare `azd down` unless re-deploying within seconds

**Recovery**:
```powershell
# Purge soft-deleted App Configuration
az appconfig delete --name "appconfig-$env" --resource-group "$rg" --yes

# Purge soft-deleted Key Vault
az keyvault purge --name "kv-$env"  # NOTE: No resource group; soft-deleted at subscription level

# Only then retry
azd down --purge --no-prompt
```

---

## 2. Environment Variable Caching

**The Trap**: Environment variables set by `azd env set` persist across commands and can leak between environments.

**Example**:
```powershell
azd env select dev
azd env set POSTGRES_PASSWORD "dev-secret"

azd env select staging
# Bug: staging now has POSTGRES_PASSWORD from dev!
# (it will be overwritten if you set it, but reads from cache)
```

**Impact**: Wrong secret used in staging; potential credential leaks; data corruption.

**Prevention**:
- Always run `azd env select <name>` before `azd env set-secret` or deployment
- Verify: `azd env get-values | grep "ENV_NAME|POSTGRES"` matches intent
- Never assume environment persists across scripts

**Recovery**:
```powershell
# Clear and reset
azd env select staging
azd env set POSTGRES_PASSWORD "staging-correct-secret"
```

---

## 3. State Leaks Between `azd up` Cycles

**The Trap**: Running `azd up` twice in succession can cause partial state if the first run failed partway.

**Example**:
```powershell
# First attempt (partially fails)
azd up                          # Provision succeeds, deploy fails

# Second attempt (uses stale state)
azd up                          # Provision skipped, deploy retried with stale config
# Result: Old pod images still running; new code not deployed
```

**Impact**: Inconsistent state; code doesn't update; debugging confusion.

**Prevention**:
- Run `azd provision --preview` before `azd up`
- After failures, run `azd down` then `azd up` (clean reset)
- Avoid `azd up` for production; use separate `azd provision` + `azd deploy` + tests

**Recovery**:
```powershell
# If deployment state is unclear
azd down --no-prompt            # Clean up
azd up                          # Fresh deployment
```

---

## 4. Kubernetes Resources Not Deleted with `azd down`

**The Trap**: `azd down` deletes Azure infrastructure but may leave K8s resources behind if the postdeploy hook or manifest cleanup fails.

**Example**:
```powershell
azd down --no-prompt
# K8s resources remain (pods, services, PVCs)
az resource list --resource-group "$rg"         # Shows AKS cluster still exists!
```

**Impact**: Resource group deletion fails; cost from stranded resources; cleanup requires manual intervention.

**Prevention**:
- Always run `kubectl delete namespace <ns> --force --grace-period=0` before `azd down`
- Verify K8s resources deleted: `kubectl get all -n <your-namespace>` (should be empty)
- In [cleanup-completely](../actions/cleanup-completely.md), Step 5 handles this

**Recovery**:
```powershell
# If K8s resources remain after azd down
kubectl delete namespace <your-namespace> --force --grace-period=0
# Then retry cleanup
az group delete --name "$rg" --yes
```

---

## 5. ACR Images Not Cleaned Up

**The Trap**: `azd down` deletes ACR resource but pushes to the registry can leave untagged or old image layers.

**Example**:
```powershell
# Deploy 5 times; each push leaves old layers
docker push "$acr.azurecr.io/api:latest"        # 1st
docker push "$acr.azurecr.io/api:latest"        # 2nd (replaces tag, but old layers stay)
docker push "$acr.azurecr.io/api:latest"        # 3rd
# ...
# ACR now has 5 image layers; costs continue even after azd down
```

**Impact**: Unexpected ACR storage costs; quota exhaustion; slow pulls.

**Prevention**:
- Tag images with commit SHA: `docker tag api:latest $acr.azurecr.io/api:$commit_sha`
- Set Acr lifecycle policies to auto-delete untagged images
- Regularly: `az acr purge --filter "api:.*" --ago 30d --untagged`

**Recovery**:
```powershell
# Clean up old images
az acr repository delete --name "$acr" --repository api
# Or use: az acr purge --registry "$acr" --filter "*:*" --ago 7d --untagged
```

---

## 6. Database Persistence After `azd down`

**The Trap**: Database backups or persistent data remain in Azure Storage or managed snapshots.

**Example**:
```powershell
azd down --no-prompt
# Database deleted, but backup in Azure Storage remains
az backup vault list --resource-group "$rg" | Select-String "DroppedFromSchedule"
# Backup storage costs still accumulate
```

**Impact**: Unexpected cost; data privacy (old backups not deleted); quota issues on next deployment.

**Prevention**:
- Check for backups before cleanup: `az backup vault list --resource-group $rg`
- Delete retention policies: `az backup container unregister --vault-name "$kv" --container-name "$container"`
- Use [cleanup-completely](../actions/cleanup-completely.md), Step 3 for pre-cleanup backup

**Recovery**:
```powershell
# Find and delete orphaned backups
az backup vault list --resource-group "$rg"
az backup backup-properties set --vault "name" --soft-delete-feature-state Disable
```

---

## 7. Firewall Rules Block Connectivity

**The Trap**: Azure SQL, PostgreSQL, or Cosmos DB firewall rules don't include AKS outbound IP or local client IP.

**Example**:
```powershell
# AKS tries to connect to database
# Error: "Host x.x.x.x is not allowed to connect to this MySQL server"

# Local dev try to migrate
# Error: Connection timeout (firewall is blocking)
```

**Impact**: Deployments fail with cryptic "connection refused" errors; E2E tests can't run; database unreachable.

**Prevention**:
- Bicep should set firewall rules: `allowAllAzureServices: true` (for AKS)
- Add local IP to allow dev migration: `az postgres server firewall-rule create --name AllowLocalDev --ip-address <your-ip>`
- Verify with: `az postgres server firewall-rule list --resource-group $rg --server-name <server>`

**Recovery**:
```powershell
# Allow AKS:
az postgres server firewall-rule create \
  --name AllowAKS \
  --resource-group "$rg" \
  --server-name "postgres-$env" \
  --start-ip-address 0.0.0.0 \
  --end-ip-address 255.255.255.255

# Allow local IP:
az postgres server firewall-rule create \
  --name AllowLocalDev \
  --resource-group "$rg" \
  --server-name "postgres-$env" \
  --start-ip-address <your-public-ip> \
  --end-ip-address <your-public-ip>
```

---

## 8. Event-Driven Resources Not Cleaned Up (e.g., Drasi, KEDA, Event Grid)

**The Trap**: Event-driven components (Drasi continuous queries, KEDA scalers, Event Grid subscriptions) persist in the cluster even after app deletion.

**Example**:
```powershell
azd down --no-prompt
# Event sources remain in the cluster
kubectl get crd | grep -E '<event-crd-pattern>'     # Still exist!
kubectl get <custom-event-resources> -A             # Still running!
# Unexpected cost; interference with next deployment
```

**Impact**: Event-system conflicts on re-deployment; unexpected resource consumption; state entanglement.

**Prevention**:
- Use cleanup logic in postdeploy hook or CI/CD pipeline before `azd down`
- Verify cleanup: `kubectl get crd | grep -E '<event-crd-pattern>'` (should be empty)
- Document custom cleanup scripts in infrastructure README

**Recovery**:
```powershell
# Manual cleanup of event systems
kubectl delete -f <path-to-event-system-manifests> 2>/dev/null || $true
kubectl delete <custom-event-resource-kind> -A 2>/dev/null || $true

# Then retry
azd down --purge --no-prompt
```

---

## 9. Frontend Build-Time Variable Injection (SPA Framework Variables)

**The Trap**: SPA build-time variables (Vite `VITE_*`, Next.js `NEXT_PUBLIC_*`, etc.) are baked into the compiled bundle. Changing them requires rebuild.

**Example**:
```powershell
# Built with wrong API URL (Vite example)
docker build -t frontend:latest --build-arg VITE_API_URL="http://localhost:8080" .

# Later, API URL changes to "http://api-prod.azurewebsites.net"
# Frontend still points to localhost → all API calls fail
# Restarting pods doesn't help; the URL is in the compiled JS bundle
```

**Impact**: Frontend unreachable from AKS; ERR_NAME_NOT_RESOLVED or timeout errors; rebuilds required for every config change.

**Prevention**:
- ALWAYS pass explicit build-time variables during docker build (Vite: `--build-arg VITE_*`; Next.js: `--build-arg NEXT_PUBLIC_*`)
- Set Dockerfile defaults to empty: `ARG VITE_API_URL=` (must be provided at build time)
- Verify build output contains correct value: `grep "$apiUrl" dist/assets/*.js` or similar
- Document in project README that SPA rebuild is required when API endpoint changes (unlike backend which uses env vars at runtime)

**Recovery**:
```powershell
# Rebuild SPA with correct API URL
docker build -f frontend/Dockerfile \
  -t "$acr.azurecr.io/frontend:latest" \
  --build-arg VITE_API_URL="$correctApiUrl" \
  frontend/

docker push "$acr.azurecr.io/frontend:latest"
kubectl rollout restart deployment/frontend-deployment -n <your-namespace>
```

---

## 10. CORS Configuration Mismatch

**The Trap**: API CORS origins don't include frontend URL; frontend can't call API.

**Example**:
```powershell
# K8s ConfigMap has mismatched CORS origins
Cors__AllowedOrigins="http://localhost:3000"

# Actual frontend URL: "<frontend-origin>"

# Browser blocks: "CORS policy: No Access-Control-Allow-Origin header"
# Frontend API calls fail; debugging appears to be a network issue
```

**Impact**: Frontend can't fetch from API; all API calls fail due to CORS; appears to be a network error.

**Prevention**:
- Compute frontend URL correctly in postdeploy hook
- Set `CORS_ALLOWED_ORIGINS` environment variable before applying K8s manifests
- Verify ConfigMap has correct origins: `kubectl get configmap app-config -o yaml | grep Cors`
- Restart API pods after ConfigMap changes: `kubectl rollout restart deployment/api-deployment`

**Recovery**:
```powershell
# Check ConfigMap
kubectl get configmap app-config -n <your-namespace> -o yaml | grep -A3 "Cors"

# Update with correct origin
$frontendUrl = "<frontend-origin>"
kubectl set env configmap/app-config -n <your-namespace> "Cors__AllowedOrigins=$frontendUrl"

# Restart API pods to pick up config change
kubectl rollout restart deployment/api-deployment -n <your-namespace>
```

---

## 11. Database Connection String Password Encoding (CRITICAL)

**The Trap**: Database passwords containing special characters (`/`, `+`, `@`, `;`) are not URL-encoded in connection strings, causing silent connection failures.

**Example**:
```powershell
# Password retrieved from Key Vault
$password = "/yoaZxbB8yxPuDpfd2b+sNk66SsFxv34"  # Contains / and +

# ❌ WRONG: Unencoded password in connection string
$connectionString = "Server=myserver.postgres.database.azure.com;Username=admin;Password=$password;Database=mydb"

# Npgsql/ADO.NET treats / and + as special characters
# Result: "Failed to connect to myserver.postgres.database.azure.com:5432"
# Pod shows as READY, but migrations fail silently in background
```

**Impact**: Silent connection failures; pod reaches READY status but database operations fail; migrations never execute; API returns 500 on data endpoints.

**Why Silent:**
- Health probes typically don't test database connectivity
- Startup migration code often wrapped in try-catch with "app will continue" fallback
- Error only detected when data endpoints are first called

**Prevention**:
- **Always URL-encode passwords** before building connection strings:
  ```powershell
  # ✅ CORRECT: URL-encode password first
  $plainPassword = az keyvault secret show --name db-password --vault-name myvault --query value -o tsv
  $encodedPassword = [System.Net.WebUtility]::UrlEncode($plainPassword)
  $connectionString = "Server=$host;Username=$user;Password=$encodedPassword;Database=$db"
  ```

- **Character encoding map**:
  - `/` → `%2F` (forward slash)
  - `+` → `%2B` (plus sign)
  - `@` → `%40` (at sign)
  - `;` → `%3B` (semicolon)
  - `=` → `%3D` (equals)
  - `&` → `%26` (ampersand)
  - `%` → `%25` (percent - encode first!)
  - Space → `%20`

- **Validation in deployment scripts**:
  ```powershell
  # After building connection string, validate encoding
  if ($connectionString -match 'Password=([^;]+)') {
      $passwordPart = $matches[1]
      # Check for unencoded special chars
      if ($passwordPart -match '[/+@;=&% ]') {
          Write-Error "ERROR: Password contains unencoded special characters"
          Write-Error "Password part: $($passwordPart.Substring(0, [Math]::Min(5, $passwordPart.Length)))... (truncated)"
          exit 1
      }
      Write-Host "✓ Connection string password validation passed" -ForegroundColor Green
  }
  ```

- **Application-level encoding (C#)**:
  ```csharp
  // In Program.cs or configuration setup
  var plainPassword = configuration["Database:Password"];
  var encodedPassword = System.Net.WebUtility.UrlEncode(plainPassword);
  
  var connectionStringBuilder = new NpgsqlConnectionStringBuilder {
      Host = configuration["Database:Host"],
      Database = configuration["Database:Name"],
      Username = configuration["Database:User"],
      Password = encodedPassword,  // Use encoded password
      SslMode = SslMode.Require
  };
  
  services.AddDbContext<MyDbContext>(options =>
      options.UseNpgsql(connectionStringBuilder.ConnectionString));
  ```

**Recovery**:
```powershell
# If connection failures detected, encode password and reapply
$plainPassword = az keyvault secret show --name db-password --vault-name "$vaultName" --query value -o tsv
$encodedPassword = [System.Net.WebUtility]::UrlEncode($plainPassword)

Write-Host "Plain password length: $($plainPassword.Length) chars"
Write-Host "Encoded password: $($encodedPassword.Substring(0, 5))... (first 5 chars)"

# Update environment variable before applying K8s manifests
$env:POSTGRES_PASSWORD = $encodedPassword

# Reapply K8s manifests (will regenerate Secret with encoded password)
& <path-to-manifest-apply-script>

# Force pod restart to pick up new Secret
kubectl rollout restart deployment/<api-deployment-name> -n <your-namespace>

# Validate migration now succeeds
kubectl logs -n <your-namespace> -l app=<api-app-label> --tail=50 | Select-String "migration|database|✓"
```

**Related Documentation:**
- C# encoding patterns: `.github/instructions/csharp.instructions.md` → Entity Framework Core section
- Docker deployment patterns: `.github/instructions/docker.instructions.md` → Section 9 (password encoding)
- K8s variable substitution: `.github/guides/kubernetes-variable-substitution-debugging.md`

---

## 12. Subscription or Location Mismatch

**The Trap**: Provisioning to wrong subscription or region; resources created in unexpected places.

**Example**:
```powershell
azd env select dev
azd env set AZURE_SUBSCRIPTION_ID "wrong-subscription-id"

azd provision
# Resources created in wrong subscription
# Bills go to wrong account
# CI/CD pipeline can't find or access resources
# Cleanup issues and cost attribution problems
```

**Impact**: Cost attribution wrong; resources in unexpected location; access denied; CI/CD can't find resources.

**Prevention**:
- Always verify before provision: `az account show --query "{ id: subscriptionId, name: name }"`
- Print environment before critical ops: `azd env get-values | Select-String "SUBSCRIPTION|LOCATION|ENV_NAME"`
- Add explicit confirmation in scripts before `azd provision` or `azd down` on production
- Set role-based access controls (RBAC) to restrict who can deploy to prod
- Use different service principals for dev/staging vs prod

**Recovery**:
```powershell
# Check what was provisioned
az account set --subscription "correct-subscription-id"
az group list --query "[?contains(name, '<your-project>')]"

# Delete from wrong subscription
az account set --subscription "wrong-subscription-id"
az group delete --name "<resource-group>" --yes

# Fix environment and retry
azd env set AZURE_SUBSCRIPTION_ID "<correct-id>"
azd env set AZURE_LOCATION "<correct-region>"
azd provision
```

---

## 13. YAML Placeholder Format Mismatch in K8s Manifests

**The Trap**: K8s manifest placeholders use format `${VARIABLE_NAME}`, but replacement scripts only match partial strings.

**Example**:
```yaml
# K8s Secret with placeholder
apiVersion: v1
kind: Secret
metadata:
  name: db-credentials
type: Opaque
stringData:
  POSTGRES_PASSWORD_B64: ${POSTGRES_PASSWORD_B64}        # Format: ${VAR_NAME}
  DATABASE_CONNECTION: "Server=myhost;Password=${POSTGRES_PASSWORD_B64}"  # Format: ${VAR_NAME}
```

```powershell
# ❌ WRONG: Partial match (missing ${, only matches ...B64})
$yamlContent = $yamlContent.Replace('POSTGRES_PASSWORD_B64}', $encodedPassword)
# Result: Placeholder NOT replaced (regex only matched part of it)
# Secret gets: "Password=\"${POSTGRES_PASSWORD_B64}\"" (literal placeholder remains)

# ✓ CORRECT: Complete placeholder format
$yamlContent = $yamlContent.Replace('${POSTGRES_PASSWORD_B64}', $encodedPassword)
# Result: Placeholder successfully replaced
# Secret gets: "Password=\"abc123xyz\"" (actual value)
```

**Impact**: Unresolved placeholders in manifest → K8s Secret contains literal `${VAR_NAME}` strings → pods fail to connect to database → API returns 500 on all data operations.

**Why Silent**: ConfigMap/Secret apply successfully, but the `kubectl` tool doesn't validate placeholder format; error only surfaces when app tries to use the secret value.

**Prevention**:

1. **Use complete placeholder format in all manifests**:
   ```yaml
   # ✓ CORRECT
   ${VARIABLE_NAME}
   
   # ❌ WRONG (partial)
   VARIABLE_NAME}
   ${VARIABLE
   ```

2. **Pre-validate placeholders before applying manifests**:
   ```powershell
   # List all placeholders in manifest file
   $placeholders = @()
   $yamlContent | Select-String '\$\{[^}]+\}' -AllMatches | ForEach-Object {
       $placeholders += $_.Matches.Value | Select-Object -Unique
   }
   
   Write-Host "Placeholders found: $($placeholders -join ', ')" -ForegroundColor Gray
   
   # Verify each placeholder will be substituted
   $unresolved = @()
   foreach ($placeholder in $placeholders) {
       $value = Invoke-Expression "\$env:${placeholder -replace '[\$\{\}]', ''}"
       if ([string]::IsNullOrEmpty($value)) {
           $unresolved += $placeholder
       }
   }
   
   if ($unresolved.Count -gt 0) {
       Write-Error "ERROR: Unresolved placeholders: $($unresolved -join ', ')"
       Write-Error "Set missing environment variables before applying manifests"
       exit 1
   }
   ```

3. **Post-apply validation**:
   ```powershell
   # After applying Secret, verify no literal placeholders remain
   $secret = kubectl get secret db-credentials -n <namespace> -o jsonpath='{.data.POSTGRES_PASSWORD_B64}' | base64 -d
   
   if ($secret -match '\$\{') {
       Write-Error "ERROR: Secret contains unresolved placeholder: $secret"
       Write-Error "Fix manifest and reapply"
       exit 1
   }
   
   Write-Host "✓ Secret validation passed" -ForegroundColor Green
   ```

4. **Environment variable validation before substitution**:
   ```powershell
   # Before substituting into manifest
   if ([string]::IsNullOrEmpty($env:POSTGRES_PASSWORD_B64)) {
       Write-Error "ERROR: POSTGRES_PASSWORD_B64 not set"
       exit 1
   }
   
   # Substitute with exact placeholder format
   $yamlContent = Get-Content secret.yaml
   $yamlContent = $yamlContent.Replace('${POSTGRES_PASSWORD_B64}', $env:POSTGRES_PASSWORD_B64)
   $yamlContent | Set-Content secret.yaml.resolved
   
   # Verify replacement occurred (check file contains actual value, not placeholder)
   if ((Get-Content secret.yaml.resolved) -match '\$\{POSTGRES_PASSWORD_B64\}') {
       Write-Error "ERROR: Placeholder still present after substitution"
       exit 1
   }
   ```

**Recovery**:
```powershell
# If Secret contains unresolved placeholders:

# 1. Delete the bad Secret
kubectl delete secret db-credentials -n <namespace> --ignore-not-found

# 2. Fix manifest to use complete placeholder format: ${VAR_NAME}
# 3. Ensure environment variables are set
$env:POSTGRES_PASSWORD_B64 = [Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes($plainPassword))

# 4. Substitute placeholders correctly
$yamlContent = Get-Content infrastructure/k8s/secrets.yaml
$yamlContent = $yamlContent.Replace('${POSTGRES_PASSWORD_B64}', $env:POSTGRES_PASSWORD_B64)
$yamlContent | kubectl apply -f -

# 5. Validate
$resolved = kubectl get secret db-credentials -n <namespace> -o jsonpath='{.data.POSTGRES_PASSWORD_B64}' | base64 -d
if ($resolved -match '\$\{') {
    Write-Error "Still contains placeholder: $resolved"
}

# 6. Restart dependent pods
kubectl rollout restart deployment/api-deployment -n <namespace>
```

**Related Documentation:**
- K8s variable substitution patterns: `.github/infrastructure/K8S_MANIFEST_SUBSTITUTION.md`
- PowerShell placeholder validation: `.github/skills/managing-azure-dev-cli-lifecycle/actions/deploy-efficiently.md` → Step 5
- Common substitution error patterns: This document, Trap #12 (Database Connection String Password Encoding)

---

## 14. Entra ID Resources Not Deleted by `azd down`

**The Trap**: `azd down` (and `azd down --purge`) does **not** remove Entra ID (Azure AD) resources—app registrations, service principals, or federated credentials are not tracked by azd's resource group management.

**Example**:
```powershell
azd down --purge --no-prompt
# Azure resource group deleted ✓
# But app registrations and service principals remain in Entra ID!

azd up  # Try to redeploy with same environment
# ERROR: "Cannot create application 'my-app-dev'. App registration with display name 'my-app-dev' already exists."
# (Or worse: silently reuses the old app registration with stale credentials)
```

**Additional complication**: Deleted app registrations are soft-deleted in Entra ID (moved to deleted items). A redeploy cannot create a new one with the same name until the soft-deleted copy is permanently removed. Permanent removal via Azure CLI can be flaky due to **eventual consistency** when your Azure tenant and CLI execution region differ (e.g., tenant in West Europe, CLI running in North Central US).

**Prevention**:
- Add a `predown` hook in `azure.yaml` to clean up Entra ID resources before `azd down` runs:
  ```yaml
  hooks:
    predown:
      windows:
        shell: pwsh
        run: ./hooks/predown-remove-app-registrations.ps1
      posix:
        shell: sh
        run: ./hooks/predown-remove-app-registrations.sh
  ```
- Tag app registrations with a unique `azd-env-id` (not just `azd-env-name`) so the cleanup script can scope deletion precisely. Why unique ID? The same environment name can exist in multiple subscriptions within the same tenant—`azd-env-name` alone can't distinguish them:
  ```bicep
  // Generate unique environment ID from subscription + env + location
  var azdEnvironmentId = '${environmentName}-${uniqueString(subscription().subscriptionId, environmentName, location)}'
  
  // Note: Entra ID resources use string arrays for tags, not key-value pairs
  // Use a helper function to flatten the tags dictionary
  func flattenTags(tags object) string[] => map(items(tags), tag => '${tag.key}: ${tag.value}')
  
  var entraIdTags = flattenTags({ 'azd-env-name': environmentName, 'azd-env-id': azdEnvironmentId })
  ```
- Add retry logic in the cleanup script—Entra ID eventual consistency can cause flaky deletes, especially across regions:
  ```powershell
  # In predown-remove-app-registrations.ps1
  param(
      [string]$EnvironmentId = $env:AZD_ENV_ID,
      [int]$MaxRetries = 5
  )
  
  $apps = az ad app list --filter "tags/any(t:t eq 'azd-env-id: $EnvironmentId')" | ConvertFrom-Json
  
  foreach ($app in $apps) {
      for ($i = 1; $i -le $MaxRetries; $i++) {
          az ad app delete --id $app.appId
          if ($LASTEXITCODE -eq 0) { break }
          Write-Warning "Delete attempt $i failed; retrying in $($i * 5)s..."
          Start-Sleep -Seconds ($i * 5)
      }
      
      # Also permanently remove from soft-deleted items (with retries for eventual consistency)
      for ($i = 1; $i -le $MaxRetries; $i++) {
          az ad app delete --id $app.appId --permanent 2>/dev/null
          if ($LASTEXITCODE -eq 0) { break }
          Start-Sleep -Seconds ($i * 5)
      }
  }
  ```

**Recovery**:
```powershell
# List soft-deleted app registrations
az ad app list --filter "deletedTime ne null" --show-deleted | ConvertFrom-Json | Select-Object displayName, appId

# Permanently delete by app ID
az ad app delete --id "<app-id>" --permanent

# Wait and retry if flaky (eventual consistency)
Start-Sleep -Seconds 30
az ad app delete --id "<app-id>" --permanent
```

---

## 15. `azd-env-name` Tag Collisions on Multi-Region Deployments

**The Trap**: `azd down` identifies resources to delete by the `azd-env-name` resource tag. If you deploy the same environment name to **two different regions** (even in different subscriptions within the same tenant), `azd down` may delete resources from both.

**Example**:
```powershell
# Deploy to East US
azd env set AZURE_LOCATION "eastus"
azd env set AZURE_ENV_NAME "dev"
azd provision                         # Creates resources tagged: azd-env-name=dev

# Deploy same template to West Europe (same subscription, same tenant)
azd env set AZURE_LOCATION "westeurope"
azd env set AZURE_ENV_NAME "dev"      # Same name!
azd provision                         # Creates MORE resources also tagged: azd-env-name=dev

# Now run cleanup for the West Europe instance
azd down --purge                      # ⚠️ Deletes BOTH East US AND West Europe resources!
# (azd-env-name=dev matches both deployments)
```

**Prevention**:
- Include location in environment name to ensure uniqueness:
  ```powershell
  azd env new "dev-eastus"
  azd env set AZURE_ENV_NAME "dev-eastus"
  azd env set AZURE_LOCATION "eastus"
  ```
- For programmatic deployments, use a deterministic unique suffix derived from subscription + environment + location to make resource names globally unique (independent of `azd-env-name` tag scoping):
  ```bicep
  // In main.bicep: generate unique instance ID
  var instanceId = take(uniqueString(subscription().subscriptionId, environmentName, location), 8)
  ```
- Document the multi-region limitation in your template's README so users are aware before deploying to multiple regions with the same name.

**Recovery**:
```powershell
# If wrong resources were deleted, check which environment was affected
az resource list --tag "azd-env-name=dev" --output table

# Reprovision in the correct environment
azd env select dev-eastus
azd provision
```

---

## Prevention Checklist

Before every azd operation:
- [ ] Verify current environment: `azd env list`
- [ ] Verify current subscription: `az account show`
- [ ] Verify target location: `azd env get-values | grep AZURE_LOCATION`
- [ ] For critical ops (provision, down), add confirmation prompt
- [ ] Review `--preview` output before applying changes
- [ ] Check for known traps in this document that apply to your workflow
- [ ] **For K8s manifests: Validate all placeholders use complete `${VAR_NAME}` format** (Trap #13)
- [ ] **For postdeploy scripts: Extract environment name using regex pattern** (Trap #2 + Environment Naming section)
- [ ] **If template creates Entra ID resources: Add `predown` hook** (Trap #14)
- [ ] **If deploying to multiple regions: Use unique environment names per region** (Trap #15)
