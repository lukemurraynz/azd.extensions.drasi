# Hook Authoring Best Practices

Best practices for writing azd [hooks](https://learn.microsoft.com/en-us/azure/developer/azure-developer-cli/azd-extensibility) — scripts that run at lifecycle events (`preprovision`, `postprovision`, `predown`, `postdeploy`, etc.).

---

## Hook Registration (`azure.yaml`)

```yaml
hooks:
  postprovision:
    windows:
      shell: pwsh
      run: ./hooks/postprovision.ps1
    posix:
      shell: sh
      run: ./hooks/postprovision.sh
  predown:
    windows:
      shell: pwsh
      run: ./hooks/predown.ps1
    posix:
      shell: sh
      run: ./hooks/predown.sh
```

Use `windows` / `posix` to support both platforms. Prefer PowerShell for Azure-heavy scripts (better `$LASTEXITCODE` handling); bash for simple shell commands.

---

## PowerShell Hook Template

### Prefer Parameters with `$env:` Defaults

Write hooks as functions with parameters defaulting to environment variables. This allows the script to be tested independently without running azd:

```powershell
# ✅ RECOMMENDED: Parameters with env var defaults
param(
    [Parameter(Mandatory = $false)]
    [string]$SubscriptionId = $env:AZURE_SUBSCRIPTION_ID,

    [Parameter(Mandatory = $false)]
    [string]$ResourceGroup = $env:AZURE_RESOURCE_GROUP,

    [Parameter(Mandatory = $false)]
    [string]$EnvironmentName = $env:AZURE_ENV_NAME,

    [Parameter(Mandatory = $false)]
    [string]$Location = $env:AZURE_LOCATION
)

# When run by azd: env vars set from .azure/<env>/.env automatically
# When run manually: pass as args or ensure env vars are set via azd env get-values
```

**Why this matters**: Without parameters, the script can only be tested by running azd. With parameter defaults, you can invoke it directly with explicit arguments for unit testing and debugging:

```powershell
# Run manually during debugging (no azd required)
./hooks/postprovision.ps1 -SubscriptionId "my-sub-id" -ResourceGroup "my-rg"
```

### Always Sync Azure CLI Subscription in Hooks

azd uses its own authentication context. If your hook calls `az` commands, explicitly sync the subscription to avoid operating against the wrong Azure account:

```powershell
# Sync Azure CLI to the same subscription as azd
az account set --subscription $SubscriptionId
if ($LASTEXITCODE -ne 0) {
    throw "Unable to set Azure subscription '$SubscriptionId'. " +
          "Ensure you're logged in with: az login"
}

Write-Host "✓ Azure CLI set to subscription: $SubscriptionId" -ForegroundColor Green
```

**Why**: `azd auth login` and `az login` maintain separate credential caches. A hook targeting `$AZURE_SUBSCRIPTION_ID` may silently operate against whatever subscription `az` last authenticated to.

### Validate Required Variables Early

Fail fast on missing required inputs:

```powershell
param(
    [string]$SubscriptionId = $env:AZURE_SUBSCRIPTION_ID,
    [string]$ResourceGroup  = $env:AZURE_RESOURCE_GROUP
)

# Validate before doing any work
if ([string]::IsNullOrEmpty($SubscriptionId)) {
    throw "AZURE_SUBSCRIPTION_ID is not set. Run: azd env set AZURE_SUBSCRIPTION_ID <value>"
}
if ([string]::IsNullOrEmpty($ResourceGroup)) {
    throw "AZURE_RESOURCE_GROUP is not set. Run: azd provision first to populate it."
}
```

### Full Hook Template

```powershell
<#
.SYNOPSIS
    Post-provision hook: configures <what this hook does>.
.DESCRIPTION
    Runs after `azd provision` completes. Requires Azure CLI to be authenticated.
.PARAMETER SubscriptionId
    Azure subscription ID. Defaults to AZURE_SUBSCRIPTION_ID env var.
.PARAMETER ResourceGroup
    Resource group name. Defaults to AZURE_RESOURCE_GROUP env var.
#>
param(
    [Parameter(Mandatory = $false)]
    [string]$SubscriptionId = $env:AZURE_SUBSCRIPTION_ID,

    [Parameter(Mandatory = $false)]
    [string]$ResourceGroup = $env:AZURE_RESOURCE_GROUP,

    [Parameter(Mandatory = $false)]
    [string]$EnvironmentName = $env:AZURE_ENV_NAME
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# Validate required inputs
if ([string]::IsNullOrEmpty($SubscriptionId)) { throw "AZURE_SUBSCRIPTION_ID not set" }
if ([string]::IsNullOrEmpty($ResourceGroup))  { throw "AZURE_RESOURCE_GROUP not set" }

# Sync Azure CLI subscription
az account set --subscription $SubscriptionId
if ($LASTEXITCODE -ne 0) {
    throw "Failed to set Azure subscription to '$SubscriptionId'. Is Azure CLI authenticated?"
}

Write-Host "Running post-provision for environment '$EnvironmentName' in '$ResourceGroup'..." -ForegroundColor Cyan

# --- Your hook logic here ---

Write-Host "✓ Post-provision complete" -ForegroundColor Green
```

---

## Reading Output Variables from Bicep

After provisioning, Bicep `output` values are written to the environment's `.env` file and available as `$env:VARIABLE_NAME` in hooks:

```bicep
// In main.bicep
output POSTGRES_HOST string = postgresServer.properties.fullyQualifiedDomainName
output ACR_NAME string = containerRegistry.name
output AKS_CLUSTER_NAME string = aksCluster.name
```

```powershell
# In postprovision.ps1 — Bicep outputs are available as env vars
param(
    [string]$PostgresHost  = $env:POSTGRES_HOST,
    [string]$AcrName       = $env:ACR_NAME,
    [string]$AksClusterName = $env:AKS_CLUSTER_NAME
)
```

---

## Hooks for Mandatory Cleanup (`predown`)

For resources azd cannot delete automatically (Entra ID, DNS records, etc.), add a `predown` hook:

```yaml
# azure.yaml
hooks:
  predown:
    windows:
      shell: pwsh
      run: ./hooks/predown-cleanup.ps1
```

```powershell
# hooks/predown-cleanup.ps1
param(
    [string]$EnvironmentId = $env:AZD_ENV_ID,   # Set during postprovision from Bicep output
    [int]$MaxRetries = 5
)

Write-Host "Running predown cleanup for environment ID: $EnvironmentId" -ForegroundColor Cyan

# Example: remove Entra ID app registrations tagged with this environment ID
$apps = az ad app list --filter "tags/any(t:t eq 'azd-env-id: $EnvironmentId')" --show-deleted-items | ConvertFrom-Json

foreach ($app in $apps) {
    Write-Host "Removing app registration: $($app.displayName) ($($app.appId))"
    
    for ($attempt = 1; $attempt -le $MaxRetries; $attempt++) {
        az ad app delete --id $app.appId
        if ($LASTEXITCODE -eq 0) { break }
        
        Write-Warning "Delete attempt $attempt/$MaxRetries failed; retrying in $($attempt * 5)s..."
        Start-Sleep -Seconds ($attempt * 5)
    }
    
    # Permanently remove from soft-deleted items to allow re-deployment with same name
    # (Retry needed: Entra ID eventual consistency across regions can be flaky)
    for ($attempt = 1; $attempt -le $MaxRetries; $attempt++) {
        az ad app delete --id $app.appId --permanent 2>$null
        if ($LASTEXITCODE -eq 0) { break }
        Start-Sleep -Seconds ($attempt * 5)
    }
}

Write-Host "✓ Predown cleanup complete" -ForegroundColor Green
```

> **Note on flakiness**: Entra ID operations are eventually consistent. If your azd pipeline runs in a different Azure region from your tenant's primary region (e.g., tenant in West Europe, CI runner in East US), delete operations can succeed locally but fail in CI or vice versa. The retry pattern above accounts for this. See [this known issue](https://github.com/Azure/azure-cli/issues/32467).

---

## Checklist for New Hooks

- [ ] Parameters declared with `$env:VAR` defaults (testable independently)
- [ ] `az account set --subscription $SubscriptionId` called early (CLI auth synced)
- [ ] `$LASTEXITCODE -ne 0` checked after every `az` call
- [ ] `Set-StrictMode -Version Latest` and `$ErrorActionPreference = 'Stop'` set
- [ ] Hook registered in `azure.yaml` for both `windows` and `posix`
- [ ] Script path exists relative to `azure.yaml` (test with `Test-Path ./hooks/yourscript.ps1`)
- [ ] Hook documented in README (what it does, what env vars it uses)
