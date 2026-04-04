# Environment Variables Reference

Key azd and application environment variables used across the lifecycle.

---

## Critical Variables (Required for All Operations)

Must be set before `azd provision` / `azd deploy`.

| Variable | Purpose | Required | Example | Validation |
|----------|---------|----------|---------|-----------|
| **AZURE_SUBSCRIPTION_ID** | Target Azure subscription | ✅ Yes | `12345678-1234-1234-1234-123456789012` | Valid UUID, must have Contributor+ role |
| **AZURE_LOCATION** | Azure region for resources | ✅ Yes | `eastus`, `westeurope`, `southeastasia` | Must be valid region (check `az account list-locations`) |
| **AZURE_ENV_NAME** | Environment name for resource naming (resource group, DNS labels, etc.) | ✅ Yes | `dev`, `staging`, `prod` | Used in all resource names via `azure.yaml` infrastructure parameters; should match environment directory name |
| **AZURE_TENANT_ID** | Azure AD tenant (if CI/CD) | ⚠️ CI/CD only | `12345678-1234-1234-1234-123456789012` | Required for service principal auth in pipelines |

### Set Critical Variables

```powershell
azd env set AZURE_SUBSCRIPTION_ID "<sub-uuid>"
azd env set AZURE_LOCATION "<region>"
azd env set AZURE_ENV_NAME "<env-name>"
```

### Validate Critical Variables

```powershell
azd env get-values | Select-String "AZURE_SUBSCRIPTION_ID|AZURE_LOCATION|AZURE_ENV_NAME"
```

---

## Environment Name Format & Resource Naming (CRITICAL)

**Problem**: `AZURE_ENV_NAME` format differs across templates and teams, causing naming mismatches if scripts assume one format.
- **Composite format** (common): `<project>-dev`
- **Short format** (also common): `dev`

If scripts always append prefixes/suffixes without normalization, resource lookups can fail.

```powershell
# WRONG: Assume AZURE_ENV_NAME is always short
$AZURE_ENV_NAME = "<project>-dev"
$postgresServer = "postgres-$AZURE_ENV_NAME"      # Might produce unexpected name pattern

# ✓ BETTER: Normalize to a suffix for scripts that require short names
if ($AZURE_ENV_NAME -match '-([^-]+)$') {
    $environmentSuffix = $matches[1]              # Extracts: dev from <project>-dev
} else {
    $environmentSuffix = $AZURE_ENV_NAME          # Keeps: dev
}
```

### Environment Name Extraction Pattern

Use this PowerShell pattern in postdeploy and deployment scripts when both formats are possible:

```powershell
# Get full environment name from azd
$fullEnvironmentName = azd env get-values | Select-String "^AZURE_ENV_NAME=" | ForEach-Object { $_ -replace ".*=", "" }

# Extract short environment suffix if composite; otherwise keep as-is
if ($fullEnvironmentName -match '-([^-]+)$') {
    $environment = $matches[1]
} else {
    $environment = $fullEnvironmentName
}

# Now use $environment for all resource lookups
Write-Host "Normalized environment token: $environment" -ForegroundColor Green
$postgresServer = "postgres-$environment"
$acrName = "acr$environment"
$keyVaultName = "kv-$environment"
```

### Resource Naming Convention

All resources should follow this pattern:

```
<resource-type>-<short-environment>[-suffix]
```

**Examples**:
- PostgreSQL: `postgres-dev`, `postgres-staging`, `postgres-prod`
- Container Registry: `contregdev`, `contregstaging`, `contregprod`
- Key Vault: `kv-dev`, `kv-staging`, `kv-prod`
- DNS Labels: `api-dev-2e8dcf`, `frontend-staging-3f9a1b`, `database-prod-8c4e7d`

### Validation Script

Run this before provisioning to ensure naming consistency:

```powershell
# Validate environment name format
$fullEnv = azd env get-values | Select-String "^AZURE_ENV_NAME=" | ForEach-Object { $_ -replace ".*=", "" }
Write-Host "Full environment name: $fullEnv" -ForegroundColor Gray

# Extract short form
if ($fullEnv -match '-([^-]+)$') {
    $shortEnv = $matches[1]
} else {
    $shortEnv = $fullEnv
}
Write-Host "Normalized environment name: $shortEnv" -ForegroundColor Green

# Validate against actual resources
$rg = "<project>-$shortEnv-rg"
Write-Host "Expected resource group: $rg" -ForegroundColor Gray

$actual = az group show --name $rg --query "name" -o tsv 2>/dev/null
if ($actual) {
    Write-Host "✓ Resource group exists: $actual" -ForegroundColor Green
} else {
    Write-Host "✗ Resource group not found: $rg" -ForegroundColor Red
}
```

---

## Infrastructure Variables (Bicep/Terraform Parameters)

Customize infrastructure deployment via `azure.yaml` `infra.parameters` section. These are **derived from critical variables above**.

| Variable | Purpose | Example | Impact | Source |
|----------|---------|---------|--------|--------|
| **projectName** | Base name for all resources | `<your-project>` | Affects naming: `<projectName>-<env>-<resource>` | Hardcoded in `azure.yaml` |
| **environment** | Environment identifier for resource naming | `dev`, `staging`, `prod` | Used in DNS, storage accounts, resource groups | **Must use `${AZURE_ENV_NAME}`** |
| **location** | Region (usually same as AZURE_LOCATION) | `<region>` | Determines resource location and pricing |
| **allowAllAzureServicesPostgres** | Allow Azure services to access database firewall | `true` | Required for AKS to reach database |
| **createSchema** | Auto-create database schema during provisioning | `false` | Set `true` only if first deployment |
| **keyVaultNameSuffix** | Suffix to make Key Vault name unique | `abc123` | Key Vault names must be globally unique |
| **mapsAadAppId** | Azure Maps service principal ID | `xxxxxxxx-xxxx-...` | Required if using Azure Maps |
| **emailTestRecipients** | Email addresses for test notifications | `user@example.com,admin@example.com` | Used for Email service tests |

### Set in `azure.yaml` (REQUIRED)

**⚠️ CRITICAL**: The resource group name and all infrastructure labels depend on `AZURE_ENV_NAME` and `AZURE_LOCATION`. These **MUST** be variable references, not hardcoded values.

```yaml
infra:
  module: main
  parameters:
    projectName: <your-project>
    environment: ${AZURE_ENV_NAME}              # ✅ Use variable, NOT hardcoded value
    location: ${AZURE_LOCATION}                 # ✅ Use variable, NOT hardcoded value
    allowAllAzureServicesPostgres: true
    createSchema: false
    keyVaultNameSuffix: ${KEYVAULT_NAME_SUFFIX:}
    mapsAadAppId: ${AZURE_MAPS_AAD_APP_ID:}
    emailTestRecipients: ${EMAIL_TEST_RECIPIENTS}
```

**Why?** If `environment` is hardcoded (e.g., `environment: prod`), then:
- All environments provision into the **same resource group** (prod-rg)
- Cleanup affects the wrong environment
- Multi-environment deployments collide
- DNS labels and resource names conflict

### Set Values

```powershell
# Application-specific
azd env set KEYVAULT_NAME_SUFFIX "abc123"
azd env set AZURE_MAPS_AAD_APP_ID "<app-id>"
azd env set EMAIL_TEST_RECIPIENTS "test@example.com"
```

---

## Application Configuration Variables

Set post-provisioning to configure deployed applications.

### ASP.NET Core API Variables

| Variable | Purpose | Injected Via | Example |
|----------|---------|--------------|---------|
| **Auth__AllowAnonymous** | Allow unauthenticated access (dev only!) | K8s ConfigMap | `false` (prod), `true` (dev) |
| **Cors__AllowedOrigins** | CORS allowed frontend origins | K8s ConfigMap | `http://frontend-prod.cloudapp.azure.com` |
| **Database__ConnectionString** | PostgreSQL connection string | K8s Secret | `Server=postgres-dev.postgres.database.azure.com;...` |
| **ASPNETCORE_URLS** | API listen URLs in container | K8s Deployment | `http://+:8080` |
| **ASPNETCORE_ENVIRONMENT** | ASP.NET Core environment | K8s Deployment | `Development`, `Production` |

### Set via K8s ConfigMap (after provisioning)

```bash
kubectl create configmap app-config -n <your-namespace> \
  --from-literal=Auth__AllowAnonymous=false \
  --from-literal=Cors__AllowedOrigins=http://frontend-prod.cloudapp.azure.com \
  --from-literal=ASPNETCORE_Environment=Production
```

---

## Container Build Variables

Used during `docker build` for container images.

### Frontend (Vite) Build-Time Variables

**⚠️ CRITICAL**: These are baked into the compiled JS bundle. Changing them requires rebuild.

| Variable | Purpose | Build Arg | Example |
|----------|---------|-----------|---------|
| **VITE_API_URL** | API endpoint for frontend | `--build-arg VITE_API_URL=<url>` | `http://api-prod.cloudapp.azure.com` |
| **VITE_SIGNALR_URL** | SignalR hub endpoint for real-time | `--build-arg VITE_SIGNALR_URL=<url>` | `http://api-prod.cloudapp.azure.com/hubs/alerts` |

### Build Frontend

```powershell
# Compute API URL dynamically
$apiUrl = "http://api-$env-$suffix.$location.cloudapp.azure.com"
$imageTag = (git rev-parse --short HEAD)

# Build with explicit API URL
docker build -f frontend/Dockerfile \
  -t "$acr.azurecr.io/frontend:$imageTag" \
  --build-arg VITE_API_URL="$apiUrl" \
  frontend/

# Verify bundle contains correct URL
docker run --rm "$acr.azurecr.io/frontend:$imageTag" grep -r "$apiUrl" /app/dist/ || echo "WARNING: URL not in bundle"
```

---

## Database Connection Variables

Computed during provisioning; set in K8s Secrets for pod access.

| Variable | Sourced From | Used By | Example |
|----------|--------------|---------|---------|
| **POSTGRES_HOST** | PostgreSQL FQDN | EF Core, psql client | `postgres-dev.postgres.database.azure.com` |
| **POSTGRES_USER** | Admin username | EF Core | `azureuser@postgres-dev` (note: must include server name for Azure) |
| **POSTGRES_PASSWORD** | Admin password | EF Core | `SecureP@ssw0rd123!` |
| **POSTGRES_DB** | Database name | EF Core | `app_db` |
| **POSTGRES_PORT** | Connection port | EF Core | `5432` |

### Set in K8s Secret

```bash
kubectl create secret generic db-credentials -n <your-namespace> \
  --from-literal=host="postgres-dev.postgres.database.azure.com" \
  --from-literal=user="azureuser@postgres-dev" \
  --from-literal=password="SecureP@ssw0rd123!" \
  --from-literal=database="emergency_alerts"
```

---

## Diagnostic Variables (Optional)

For troubleshooting and debugging.

| Variable | Purpose | Values | Impact |
|----------|---------|--------|--------|
| **AZD_DEBUG** | Enable azd debug output | `true`, `false` | More logs; useful for troubleshooting |
| **AZURE_LOG_LEVEL** | Azure SDK log level | `DEBUG`, `INFO`, `WARNING`, `ERROR` | Controls verbosity of Azure SDK |
| **ASPNETCORE_ENVIRONMENT** | ASP.NET environment | `Development`, `Staging`, `Production` | Affects error pages, logging, security |
| **LOG_LEVEL** | Application log level | `Debug`, `Information`, `Warning`, `Error` | Verbosity of application logs |

### Enable Debugging

```powershell
# For azd
$env:AZD_DEBUG = "true"
azd provision --debug

# For Azure SDK
$env:AZURE_LOG_LEVEL = "DEBUG"

# For API (via K8s ConfigMap)
kubectl set env configmap/app-config -n <your-namespace> LOG_LEVEL=Debug
```

---

## Multi-Environment Variable Examples

### Development Environment

```powershell
azd env select dev
azd env set AZURE_SUBSCRIPTION_ID "dev-sub-uuid"
azd env set AZURE_LOCATION "<dev-region>"
azd env set AZURE_ENV_NAME "dev"
azd env set-secret POSTGRES_PASSWORD "simple-dev-password"  # Less strict in dev
```

### Staging Environment

```powershell
azd env select staging
azd env set AZURE_SUBSCRIPTION_ID "stage-sub-uuid"
azd env set AZURE_LOCATION "eastus"
azd env set AZURE_ENV_NAME "staging"
azd env set-secret POSTGRES_PASSWORD "strong-stage-P@ss123!"
```

### Production Environment

```powershell
azd env select prod
azd env set AZURE_SUBSCRIPTION_ID "prod-sub-uuid"
azd env set AZURE_LOCATION "westeurope"
azd env set AZURE_ENV_NAME "prod"
azd env set-secret POSTGRES_PASSWORD "very-strong-prod-P@ss$RANDOM!"  # Rotate monthly
```

---

## Variable Ordering in `azure.yaml`

Recommended order for clarity:

```yaml
infra:
  parameters:
    # Resource naming
    projectName: <your-project>
    environment: ${AZURE_ENV_NAME}
    location: ${AZURE_LOCATION}
    
    # Feature flags
    allowAllAzureServicesPostgres: true
    createSchema: false
    
    # Integrations
    keyVaultNameSuffix: ${KEYVAULT_NAME_SUFFIX:}
    mapsAadAppId: ${AZURE_MAPS_AAD_APP_ID:}
    
    # Notifications
    emailTestRecipients: ${EMAIL_TEST_RECIPIENTS}
```

---

## Validation Commands

### Check All Critical Variables

```powershell
$critical = @('AZURE_SUBSCRIPTION_ID', 'AZURE_LOCATION', 'AZURE_ENV_NAME')
$critical | ForEach-Object {
  $val = azd env get-values | Select-String $_ | ForEach-Object { $_ -replace ".*=", "" }
  if ([string]::IsNullOrEmpty($val)) {
    Write-Error "MISSING: $_"
  } else {
    Write-Host "✓ $_=$val"
  }
}
```

### Compare Environments

```powershell
Write-Host "Current environment:"
azd env get-values | Sort-Object

# Switch and compare
azd env select other-env
Write-Host "Other environment:"
azd env get-values | Sort-Object
```

### Validate Before Provisioning

```powershell
# Subscription
az account set --subscription (azd env get-values | Select-String "AZURE_SUBSCRIPTION_ID" | ForEach-Object { $_ -replace ".*=", "" })
az account show --query "{ name: name, subscriptionId: subscriptionId }" --output table

# Location
$location = azd env get-values | Select-String "AZURE_LOCATION" | ForEach-Object { $_ -replace ".*=", "" }
az account list-locations --query "[?name=='$location'].displayName" --output table
```

---

## Defaults & Fallbacks

There are **two distinct places** where you can set defaults for template variables. Understand the difference:

### 1. `azure.yaml` infra parameters — empty fallback syntax

```yaml
# azure.yaml: parameters can have empty fallbacks using colon syntax
keyVaultNameSuffix: ${KEYVAULT_NAME_SUFFIX:}          # Empty fallback when unset
mapsAadAppId: ${AZURE_MAPS_AAD_APP_ID:}               # Empty fallback when unset
someVar: ${SOME_VAR:default-value}                    # "default-value" fallback
```

### 2. `main.parameters.json` — default value syntax (recommended for Bicep parameters)

Use `${VAR_NAME=defaultValue}` (equals sign, not colon) in `main.parameters.json` to set sensible defaults users can override without having to `azd env set` anything:

```json
{
  "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentParameters.json#",
  "contentVersion": "1.0.0.0",
  "parameters": {
    "environmentName":        { "value": "${AZURE_ENV_NAME}" },
    "location":               { "value": "${AZURE_LOCATION}" },
    "alertThresholdPercent":  { "value": "${ALERT_THRESHOLD_PERCENT=10}" },
    "retentionDays":          { "value": "${LOG_RETENTION_DAYS=30}" },
    "adminEmail":             { "value": "${ADMIN_EMAIL}" }
  }
}
```

- First deploy: if `ALERT_THRESHOLD_PERCENT` is unset, azd uses `10`
- After deploy: the value is stored in the environment's `.env` file for subsequent runs
- Users override with: `azd env set ALERT_THRESHOLD_PERCENT 25`

If you expose configurable parameters this way, document them in your README.

### 3. Optional Resources Pattern (boolean parameter + output roundtrip)

To make resources optional (deploy or skip based on user choice), use a boolean parameter backed by an environment variable:

**`main.parameters.json`**:
```json
"includeServiceBus": { "value": "${INCLUDE_SERVICE_BUS}" }
```

**`main.bicep`**:
```bicep
@description('Include the Service Bus in the deployment.')
param includeServiceBus bool

// Conditional resource deployment
module serviceBus './modules/service-bus.bicep' = if (includeServiceBus) {
  name: 'serviceBus'
  params: { ... }
}

// Output back to environment: stored in .env and available for subsequent azd runs
output INCLUDE_SERVICE_BUS bool = includeServiceBus
```

On first `azd up`, azd prompts the user for the value. After provisioning, the value is stored in the environment's `.env` file so subsequent deploys do not re-prompt.

---

## Loading azd Environment Variables in Integration Tests

When running integration tests locally against a provisioned environment, you need the values stored in `.azure/<env>/.env`. Rather than setting them manually, load the file programmatically:

### .NET (C#) — AzdDotEnv pattern

```csharp
// In your test setup class (e.g., in a BaseTest or fixture)
// Requires NuGet: dotenv.net

// Loads .env file from .azure/<env>/ if present; optional so tests still
// work in CI/CD where real environment variables are injected instead
AzdDotEnv.Load(optional: true);

var configuration = new ConfigurationBuilder()
    .AddEnvironmentVariables()
    .Build();

// Usage
string apiUrl = configuration["API_URL"];
string dbHost = configuration["POSTGRES_HOST"];
```

A minimal `AzdDotEnv.cs` implementation:
```csharp
public static class AzdDotEnv
{
    public static void Load(bool optional = false)
    {
        // Walk up from current directory to find .azure/
        var dir = new DirectoryInfo(Directory.GetCurrentDirectory());
        while (dir != null)
        {
            var azureDir = Path.Combine(dir.FullName, ".azure");
            if (Directory.Exists(azureDir))
            {
                // Find the first (or active) environment directory
                var envDir = Directory.GetDirectories(azureDir).FirstOrDefault();
                if (envDir != null)
                {
                    var dotEnvPath = Path.Combine(envDir, ".env");
                    if (File.Exists(dotEnvPath))
                    {
                        DotEnv.Load(new DotEnvOptions(envFilePaths: new[] { dotEnvPath }));
                        return;
                    }
                }
                break;
            }
            dir = dir.Parent;
        }
        if (!optional)
            throw new FileNotFoundException("Could not find .azure/<env>/.env file");
    }
}
```

This pattern works both locally (uses `.env` file) and in CI/CD (uses actual injected environment variables). The `optional: true` flag prevents test failures when running in pipelines where the `.env` file doesn't exist.

### PowerShell — load `.env` in scripts

```powershell
# Load azd .env file for local script execution
function Load-AzdEnv {
    param([string]$EnvName)
    $dotEnvPath = ".azure/$EnvName/.env"
    if (Test-Path $dotEnvPath) {
        Get-Content $dotEnvPath | ForEach-Object {
            if ($_ -match '^\s*(.+?)\s*=\s*"?(.*?)"?\s*$') {
                [System.Environment]::SetEnvironmentVariable($matches[1], $matches[2], 'Process')
            }
        }
        Write-Host "✓ Loaded $dotEnvPath" -ForegroundColor Green
    } elseif (-not $optional) {
        throw "Could not find $dotEnvPath"
    }
}

Load-AzdEnv -EnvName (azd env list --output json | ConvertFrom-Json | Where-Object { $_.IsDefault } | Select-Object -First 1 -ExpandProperty Name)
```

---

## Security Best Practices

1. **Never commit secrets to Git**:
   - Use `azd env set-secret` (not `azd env set`)
   - Add `.azd/` to `.gitignore`

2. **Use Key Vault for production secrets**:
   - API keys, database passwords
   - Accessed via managed identity (no keys in environment)

3. **Rotate secrets regularly**:
   - Change database passwords monthly
   - Rotate API keys quarterly
   - Use `azd env set-secret <name>` to update

4. **Audit who accesses variables**:
   - Log `azd env get-values` usage
   - Restrict `azd env set-secret` to admins
   - Use resource locks for prod configuration

---

## Troubleshooting Variable Issues

| Problem | Diagnosis | Fix |
|---------|-----------|-----|
| `${VAR}` not substituted | Variable not set | Run `azd env set VAR value` |
| Wrong value used | Variable cached from another env | Run `azd env select <name>` to confirm |
| K8s can't find ConfigMap var | ConfigMap not applied | Run `kubectl create configmap ... ` or apply manifest |
| Pod sees old value | ConfigMap updated but pods not restarted | Run `kubectl rollout restart deployment/...` |
| Frontend shows wrong API URL | Vite build-time variable wrong | Rebuild with `--build-arg VITE_API_URL=<correct>` |
