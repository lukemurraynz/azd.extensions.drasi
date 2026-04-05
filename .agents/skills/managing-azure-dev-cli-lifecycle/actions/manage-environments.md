# Manage Environments

## Purpose

Set up and maintain multi-environment configurations (dev/staging/prod), switch between environments safely, and manage secrets and configuration across deployments.

---

## Flow

### Step 1: Understand azd Environment Concepts 📚

Learn how azd organizes environments.

**Key Concepts:**

| Concept | Definition | Example |
|---------|-----------|---------|
| **Environment** | A named deployment configuration | `dev`, `staging`, `prod` |
| **Environment Directory** | `.azd/environment/<env>` | `.azd/environment/dev` |
| **.env File** | DOTENV-style variables for the environment | `AZURE_LOCATION=<region>` |
| **Secrets** | Sensitive values (passwords, tokens) | `POSTGRES_PASSWORD`, `API_KEY` |
| **Current Environment** | The environment that `azd` commands apply to | Set with `azd env select` |

**Example Structure:**
```
.azd/
├── environment/
│   ├── dev/
│   │   └── .env          # dev configuration
│   ├── staging/
│   │   └── .env          # staging configuration
│   └── prod/
│       └── .env          # prod configuration
└── <other azd state>
```

**Key Principle**: Each environment is isolated; changing one doesn't affect others.

---

### Step 2: Create a New Environment 🆕

Set up a named environment with its own configuration.

**Commands:**
```powershell
# Create new environment
Write-Host "Creating new environment..." -ForegroundColor Cyan
azd env new <env-name>              # e.g., azd env new staging

# Expected output:
# New environment '<env-name>' created.
# Use 'azd env select' to make it current.

# List all environments
azd env list

# Expected output:
# dev         (current)
# staging
# prod
```

**🛑 STOP**: Verify the environment appears in `azd env list`.

**Success Criteria:**
- [ ] `azd env list` shows new environment
- [ ] Environment directory exists: `.azd/environment/<env-name>/`
- [ ] `.env` file is created in the environment directory

---

### Step 3: Configure Environment Variables 🔧

Set up location, subscription, and other variables for the environment.

**Commands:**
```powershell
# Select the environment to configure
azd env select <env-name>

# Set critical variables
azd env set AZURE_SUBSCRIPTION_ID "<subscription-uuid>"
azd env set AZURE_LOCATION "<region>"                  # e.g., eastus, westeurope, southeastasia
azd env set AZURE_ENV_NAME "<env-name>"                # Should match environment name

# Set optional variables (specific to your app)
azd env set KEYVAULT_NAME_SUFFIX "<suffix>"
azd env set AZURE_MAPS_AAD_APP_ID "<app-id>"
azd env set EMAIL_TEST_RECIPIENTS "<email@example.com>"

# Verify configuration
azd env get-values | Sort-Object
```

**Expected Output:**
```
AZURE_ENV_NAME=staging
AZURE_LOCATION=eastus
AZURE_SUBSCRIPTION_ID=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
KEYVAULT_NAME_SUFFIX=123abc
...
```

**🛑 STOP**: Verify variables are set correctly before moving to provisioning.

**Common Variables by Environment:**
| Variable | Dev | Staging | Prod |
|----------|-----|---------|------|
| `AZURE_LOCATION` | <dev-region> | <staging-region> | <prod-region> |
| `AZURE_ENV_NAME` | dev | staging | prod |
| Replica count | 1 | 2 | 3+ |
| Database tier | B (basic) | S (standard) | P (premium) |

**Success Criteria:**
- [ ] `AZURE_SUBSCRIPTION_ID` set to correct subscription
- [ ] `AZURE_LOCATION` is a valid Azure region
- [ ] `AZURE_ENV_NAME` matches environment name (avoids confusion)

---

### Step 4: Set Secrets Securely 🔐

Store sensitive values without exposing them in `.env` files.

**Commands:**
```powershell
# Set a secret (prompts for input, not echoed to console)
azd env set-secret <secret-name>

# Examples:
azd env set-secret POSTGRES_PASSWORD
azd env set-secret API_KEY
azd env set-secret JWT_SECRET

# Expected: Prompt will ask for value (hidden input)
# Enter value for [secret-name]:

# Verify secret is stored (value not displayed)
azd env get-values | Select-String "POSTGRES"
# Output: POSTGRES_PASSWORD=<hidden>  (not shown in logs)

# Retrieve secret value (only in scripts, not logged by default)
$secret = azd env get-values | Select-String "POSTGRES_PASSWORD" | ForEach-Object { $_ -replace ".*=", "" }
# Use $secret in your deployment scripts
```

**Environment-Specific Secrets:**
```powershell
# Each environment can have different secrets
azd env select dev
azd env set-secret POSTGRES_PASSWORD "dev-password-123"

azd env select staging
azd env set-secret POSTGRES_PASSWORD "staging-password-456"  # Different for each environment

azd env select prod
azd env set-secret POSTGRES_PASSWORD "prod-password-xyz"     # Different for each environment
```

**🛑 STOP**: Never commit `.azd/environment/<env>/.env` files to Git if they contain secrets.

**Best Practice - Add to `.gitignore`:**
```bash
# .gitignore
.azd/
.env
.env.local
*.sw
*secrets*
```

**Success Criteria:**
- [ ] Secrets are set via `azd env set-secret` (not `azd env set`)
- [ ] `.gitignore` blocks `.azd/` and `.env` files
- [ ] Sensitive values are never visible in logs or `git status`

---

### Step 5: Switch Between Environments 🔄

Change the active environment.

**Commands:**
```powershell
# List all environments
azd env list
# Output shows:
# dev         (current)
# staging
# prod

# Switch to a different environment
azd env select staging

# Verify switch (staging should show as current)
azd env list
# Output shows:
# dev
# staging     (current)
# prod

# Confirm active environment
azd env get-values | Select-String "AZURE_ENV_NAME"
# Output: AZURE_ENV_NAME=staging
```

**🛑 STOP**: Always verify you're on the correct environment before running `azd provision` or `azd down`.

**Common Mistake Prevention:**
```powershell
# SAFE: Always confirm environment before critical operations
Write-Host "Current environment:" -ForegroundColor Yellow
azd env get-values | Select-String "AZURE_ENV_NAME|AZURE_LOCATION"
$confirm = Read-Host "Proceed? (yes/no)"
if ($confirm -eq 'yes') {
  azd provision
} else {
  Write-Host "Cancelled. Switch environment: azd env select <name>"
}
```

**Success Criteria:**
- [ ] `azd env list` shows correct environment marked as `(current)`
- [ ] `azd env get-values` shows variables for correct environment
- [ ] No accidental deployments to wrong environment

---

### Step 6: Validate Environment Configuration ✅

Ensure environment is ready for provisioning.

**Commands:**
```powershell
# Comprehensive validation
Write-Host "=== Environment Validation ===" -ForegroundColor Cyan

# Check all critical variables are set
$critical = @('AZURE_SUBSCRIPTION_ID', 'AZURE_LOCATION', 'AZURE_ENV_NAME')
$missing = @()
$critical | ForEach-Object {
  $value = azd env get-values | Select-String $_ | ForEach-Object { $_ -replace ".*=", "" }
  if ([string]::IsNullOrEmpty($value)) {
    $missing += $_
  } else {
    Write-Host "✓ $_=$value" -ForegroundColor Green
  }
}

if ($missing.Count -gt 0) {
  Write-Error "Missing variables: $($missing -join ', ')"
  exit 1
}

# Verify subscription
Write-Host "Verifying subscription access..." -ForegroundColor Cyan
$subId = azd env get-values | Select-String "AZURE_SUBSCRIPTION_ID" | ForEach-Object { $_ -replace ".*=", "" }
az account set --subscription "$subId"
az account show --query "{ subscriptionId: subscriptionId, displayName: name }" --output table

# Verify region is valid
Write-Host "Verifying location..." -ForegroundColor Cyan
$location = azd env get-values | Select-String "AZURE_LOCATION" | ForEach-Object { $_ -replace ".*=", "" }
az account list-locations --query "[?name=='$location'].displayName" --output table
if ($LASTEXITCODE -ne 0) {
  Write-Error "Invalid location: $location"
}

# Check for duplicate resources across environments
Write-Host "Checking for resource name collisions..." -ForegroundColor Cyan
$currentEnv = azd env get-values | Select-String "AZURE_ENV_NAME" | ForEach-Object { $_ -replace ".*=", "" }
Write-Host "Current environment: $currentEnv (resource names should be unique)"
```

**Expected Output:**
```
=== Environment Validation ===
✓ AZURE_SUBSCRIPTION_ID=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
✓ AZURE_LOCATION=<region>
✓ AZURE_ENV_NAME=staging

Subscription: My Dev Subscription (ID: xxxx...)
Location: <region> (valid)
```

**Common Issues:**
| Issue | Fix |
|-------|-----|
| `Missing variables` | Run `azd env set VARIABLE value` |
| `Invalid location` | Check: `az account list-locations --query "[].name"` |
| `Subscription access denied` | Verify IAM role: `az role assignment list --scope /subscriptions/<sub-id>` |

**Success Criteria:**
- [ ] All critical variables set
- [ ] Subscription accessible
- [ ] Location is valid
- [ ] No resource name collisions

---

### Step 7: Provision and Deploy to Environment 🚀

Deploy to the selected environment.

**Commands:**
```powershell
# Final confirmation
Write-Host "About to deploy to:" -ForegroundColor Yellow
azd env get-values | Select-String "AZURE_ENV_NAME|AZURE_LOCATION"
$proceed = Read-Host "Type 'deploy' to confirm"

if ($proceed -eq 'deploy') {
  # Provision infrastructure
  azd provision

  # Deploy application
  azd deploy

  Write-Host "✅ Environment deployed: $(azd env get-values | Select-String 'AZURE_ENV_NAME')" -ForegroundColor Green
} else {
  Write-Host "Cancelled"
}
```

**🛑 STOP**: This will create or update Azure resources. Ensure you're on the correct environment.

**Success Criteria:**
- [ ] Provisioning succeeds (exit code 0)
- [ ] Deployment succeeds (exit code 0)
- [ ] Resources appear in target subscription/resource group
- [ ] Application is accessible and healthy

---

## Common Multi-Environment Patterns

### Development → Staging → Production Workflow

```powershell
# Local development (mostly manual)
azd env select dev
azd provision      # Once, when infrastructure changes
azd deploy         # Every code change

# Prepare for staging
azd env select staging
azd provision      # Create staging infrastructure
azd deploy         # Deploy staging

# Production release
azd env select prod
azd provision      # Create prod infrastructure (usually one-time)
azd deploy         # Deploy prod (rare, usually CI/CD driven)
```

### Parameter Differences Across Environments

```powershell
# Use azure.yaml or Bicep parameter files for environment-specific values

# Example in azure.yaml:
# services:
#   api:
#     project: ./<path-to-api-project>
#     language: dotnet
# infra:
#   module: main
#   parameters:
#     replicaCount: ${REPLICA_COUNT:1}  # Dev: 1, Staging: 2, Prod: 3

# Override with environment variables:
azd env select dev
azd env set REPLICA_COUNT 1

azd env select staging
azd env set REPLICA_COUNT 2

azd env select prod
azd env set REPLICA_COUNT 3
```

### Automated Multi-Environment Deployment (CI/CD)

```powershell
# GitHub Actions / Azure Pipelines example
$envs = @('dev', 'staging', 'prod')

$envs | ForEach-Object {
  $env = $_
  Write-Host "Deploying to $env" -ForegroundColor Cyan
  
  # Authenticate
  azd auth login --client-id "$env:CLIENT_ID" --client-secret "$env:CLIENT_SECRET" --tenant-id "$env:TENANT_ID"
  
  # Select environment
  azd env select $env
  
  # Deploy
  azd provision
  azd deploy
}
```

---

## Environment Cleanup

### Delete an Environment

```powershell
# Delete local environment only (keep Azure resources)
azd env delete <env-name>

# Delete Azure resources + local state
azd env select <env-name>
azd down --purge
azd env delete <env-name>
```

### Archive an Old Environment

```powershell
# Back up environment configuration
Write-Host "Backing up environment: <env-name>"
Copy-Item ".azd/environment/<env-name>" -Destination "backup-<env-name>-$(Get-Date -Format 'yyyyMMdd')" -Recurse

# Then delete
azd env delete <env-name>
```

---

## Troubleshooting Multi-Environment Issues

| Problem | Cause | Fix |
|---------|-------|-----|
| Wrong environment deployed | Selected wrong environment | Run `azd env list` before `azd provision` |
| Resource name conflicts | Same env name used in multiple subscriptions | Use region/subscription suffix: `dev-east` vs `dev-west` |
| Secrets not available | Secret set in wrong environment | Verify with `azd env select` before `azd env set-secret` |
| Configuration drift | Manual changes made outside azd | Track changes in `azure.yaml`, always deploy from code |

---

## Best Practices

1. **Always verify environment before critical ops**: `azd env list` before `azd down`
2. **Use meaningful names**: `dev`, `staging`, `prod` (not `myenv1`, `myenv2`)
3. **Keep `.azd/` out of Git**: Add to `.gitignore`
4. **Document environment differences**: Create a table (SKU, replicas, regions)
5. **Use CI/CD for consistency**: Automate deployments to avoid manual errors
6. **Back up secrets**: Securely store secret values (not in code)
7. **Test cleanup**: Regularly test `azd down --purge` in dev to ensure it works

---

## Next Steps

✅ After setting up environments:
- Deploy to each environment in order: [deploy-efficiently](deploy-efficiently.md)
- To add more environments, repeat steps 2-7

❌ If environment setup fails: [troubleshoot-failures](troubleshoot-failures.md)
