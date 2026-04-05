---
name: azd-deployment
description: >-
  Deploy containerized applications to Azure Container Apps using Azure Developer
  CLI (azd), including extensions and AI agent deployment.
  USE FOR: setting up azd projects, writing azure.yaml, creating Bicep for Container
  Apps, configuring ACR remote builds, managing environment variables, idempotent
  deployments, azd extensions, AI agent deployment, or troubleshooting azd up failures.
version: 2.1.0
lastUpdated: 2026-03-28
---

# Azure Developer CLI (azd) Container Apps Deployment

Deploy containerized full-stack applications to Azure Container Apps with infrastructure-as-code, CI/CD automation, versioned deployments, Key Vault secret management, and production-ready operational patterns.

## What's New in v2.1.0 (2026-03-28)

- ✅ **azd extensions framework** (beta) with extension management commands, source configuration, and dev container integration
- ✅ **Azure AI Foundry agent extension** (`azure.ai.agents`) for scaffolding, deploying, and monitoring AI agents
- ✅ **Resolved [VERIFY] blocks** for `alpha.infraParameters` (confirmed still experimental)

## What's New in v2.0 (2026-02-15)

- ✅ **Container image versioning** with `imageTag` parameter (fixes `:latest` anti-pattern)
- ✅ **Key Vault secret references** for secure database connection strings
- ✅ **GitHub Actions validation gates** with Bicep build + what-if analysis
- ✅ **Post-deployment health checks** with retry logic
- ✅ **Lifecycle hooks** (postprovision, predeploy, postdeploy) for automation
- ✅ **Custom workflow configuration** to override default azd up behavior
- ✅ **Circular dependency resolution** strategies (deploy-time vs runtime)
- ✅ **Latest 2024-2025 azd features** (alpha.infraParameters, MCP integration)

> ⚠️ **`alpha.infraParameters`** is an experimental `azd` feature. Alpha features may change or be removed between azd releases without notice. Verify availability in your installed version with `azd config list alpha` or check the [azd changelog](https://github.com/Azure/azure-dev/releases).

- ✅ **Comprehensive anti-patterns** section with real-world fixes

## Quick Start

```bash
# Initialize and deploy
azd auth login
azd init                    # Creates azure.yaml and .azure/ folder
azd env new <env-name>      # Create environment (dev, staging, prod)
azd up                      # Provision infra + build + deploy
```

## Core File Structure

```
project/
├── azure.yaml              # azd service definitions + hooks
├── infra/
│   ├── main.bicep          # Root infrastructure module
│   ├── main.parameters.json # Parameter injection from env vars
│   └── modules/
│       ├── container-apps-environment.bicep
│       └── container-app.bicep
├── .azure/
│   ├── config.json         # Default environment pointer
│   └── <env-name>/
│       ├── .env            # Environment-specific values (azd-managed)
│       └── config.json     # Environment metadata
└── src/
    ├── frontend/Dockerfile
    └── backend/Dockerfile
```

## azure.yaml Configuration

### Minimal Configuration

```yaml
name: azd-deployment
services:
  backend:
    project: ./src/backend
    language: python
    host: containerapp
    docker:
      path: ./Dockerfile
      remoteBuild: true
```

### Full-Stack Multi-Service Pattern (Production-Ready)

```yaml
name: atlas
metadata:
  template: atlas@0.0.1-beta

infra:
  provider: bicep
  path: infrastructure/bicep
  module: main

services:
  control-plane-api:
    project: ./apps/control-plane-api
    language: csharp
    host: containerapp
    docker:
      path: ./apps/control-plane-api/Dockerfile
      context: ./apps/control-plane-api

  agent-orchestrator:
    project: ./apps/agent-orchestrator
    language: csharp
    host: containerapp
    docker:
      path: ./apps/agent-orchestrator/Dockerfile
      context: ./apps/agent-orchestrator

  frontend:
    project: ./apps/frontend
    language: js
    host: containerapp
    docker:
      path: ./apps/frontend/Dockerfile
      context: ./apps/frontend
      buildArgs:
        - VITE_CONTROL_PLANE_API_BASE_URL=${CONTROL_PLANE_API_BASE_URL}

pipeline:
  provider: github

hooks:
  postprovision:
    shell: sh
    run: |
      # Capture infrastructure outputs and set as environment variables
      API_URL=$(azd env get-values | grep CONTROL_PLANE_API_URL | cut -d'=' -f2 | tr -d '"')
      azd env set CONTROL_PLANE_API_BASE_URL "$API_URL"
      echo "✅ API URL configured: $API_URL"

  postdeploy:
    shell: sh
    run: |
      # Health check after deployment
      API_URL=$(azd env get-values | grep CONTROL_PLANE_API_URL | cut -d'=' -f2 | tr -d '"')
      if curl -sf "${API_URL}/health" > /dev/null 2>&1; then
        echo "✅ Deployment successful - API is healthy"
      else
        echo "❌ Health check failed"
        exit 1
      fi
```

### Full Configuration with Hooks (Legacy Example)

```yaml
name: azd-deployment
metadata:
  template: my-project@1.0.0

infra:
  provider: bicep
  path: ./infra

azure:
  location: eastus2

services:
  frontend:
    project: ./src/frontend
    language: ts
    host: containerapp
    docker:
      path: ./Dockerfile
      context: .
      remoteBuild: true

  backend:
    project: ./src/backend
    language: python
    host: containerapp
    docker:
      path: ./Dockerfile
      context: .
      remoteBuild: true

hooks:
  preprovision:
    shell: sh
    run: |
      echo "Before provisioning..."

  postprovision:
    shell: sh
    run: |
      echo "After provisioning - set up RBAC, etc."

  postdeploy:
    shell: sh
    run: |
      echo "Frontend: ${SERVICE_FRONTEND_URI}"
      echo "Backend: ${SERVICE_BACKEND_URI}"
```

### Key azure.yaml Options

| Option                      | Description                                            |
| --------------------------- | ------------------------------------------------------ |
| `remoteBuild: true`         | Build images in Azure Container Registry (recommended) |
| `context: .`                | Docker build context relative to project path          |
| `host: containerapp`        | Deploy to Azure Container Apps                         |
| `infra.provider: bicep`     | Use Bicep for infrastructure                           |
| `buildArgs`                 | Pass environment variables to Docker build             |
| `pipeline.provider: github` | Enable GitHub Actions integration                      |

## Lifecycle Hooks (NEW in v2.0)

Lifecycle hooks allow custom automation at key deployment stages. Use hooks to capture infrastructure outputs, run database migrations, configure RBAC, or perform health checks.

### Available Hooks

| Hook            | When It Runs                        | Common Use Cases                                   |
| --------------- | ----------------------------------- | -------------------------------------------------- |
| `preprovision`  | Before infrastructure deployment    | Validate prerequisites, check quotas               |
| `postprovision` | After infrastructure, before deploy | Capture outputs, configure RBAC, run DB migrations |
| `predeploy`     | Before code deployment              | Run database migrations, warm caches               |
| `postdeploy`    | After code deployment               | Health checks, smoke tests, notifications          |

### Pattern: Capturing Infrastructure Outputs

**Problem:** Frontend needs API URL but infrastructure outputs aren't available until after provisioning.

**Solution:** Use `postprovision` hook to capture Bicep outputs and set as environment variables:

```yaml
hooks:
  postprovision:
    shell: sh
    run: |
      # Get API URL from Bicep outputs
      API_URL=$(azd env get-values | grep CONTROL_PLANE_API_URL | cut -d'=' -f2 | tr -d '"')

      # Set as environment variable for frontend build
      azd env set CONTROL_PLANE_API_BASE_URL "$API_URL"

      echo "✅ API URL configured: $API_URL"
```

Then reference in `azure.yaml`:

```yaml
services:
  frontend:
    docker:
      buildArgs:
        - VITE_CONTROL_PLANE_API_BASE_URL=${CONTROL_PLANE_API_BASE_URL}
```

### Pattern: Database Migrations

```yaml
hooks:
  predeploy:
    shell: sh
    run: |
      # Run EF Core migrations before deploying new code
      CONNECTION_STRING=$(azd env get-values | grep DATABASE_CONNECTION_STRING | cut -d'=' -f2)

      cd apps/api
      dotnet ef database update --connection "$CONNECTION_STRING"
      echo "✅ Database migrations applied"
```

### Pattern: Post-Deployment Health Checks

```yaml
hooks:
  postdeploy:
    shell: sh
    run: |
      API_URL=$(azd env get-values | grep API_URL | cut -d'=' -f2 | tr -d '"')

      MAX_RETRIES=10
      for i in $(seq 1 $MAX_RETRIES); do
        if curl -sf "${API_URL}/health" > /dev/null 2>&1; then
          echo "✅ Deployment successful - API is healthy"
          exit 0
        fi
        echo "⏳ Waiting for API (attempt $i/$MAX_RETRIES)..."
        sleep 10
      done

      echo "❌ Health check failed after $MAX_RETRIES attempts"
      exit 1
```

### Platform-Specific Hooks

Use `windows` or `posix` shells for platform-specific logic:

```yaml
hooks:
  postprovision:
    windows:
      shell: pwsh
      run: |
        Write-Host "Running on Windows"
        $apiUrl = azd env get-values | Select-String "API_URL" | ForEach-Object { $_.Line.Split('=')[1].Trim('"') }
        azd env set CONTROL_PLANE_API_BASE_URL $apiUrl

    posix:
      shell: sh
      run: |
        echo "Running on Linux/macOS"
        API_URL=$(azd env get-values | grep API_URL | cut -d'=' -f2 | tr -d '"')
        azd env set CONTROL_PLANE_API_BASE_URL "$API_URL"
```

## Circular Dependency Resolution

**Problem:** Frontend build needs API URL from infrastructure outputs, but infrastructure must be provisioned before building images.

### Strategy 1: Deploy-Time Configuration (Recommended)

Use `postprovision` hook + Container Apps upsert pattern:

1. Provision infrastructure (outputs API URL)
2. Capture API URL in `postprovision` hook
3. Build frontend with API URL from environment variable
4. Deploy services

**Advantages:**

- Configuration baked into image at build time
- No runtime overhead
- Works with Vite/React/Next.js static builds

**Example:**

```yaml
hooks:
  postprovision:
    shell: sh
    run: |
      API_URL=$(azd env get-values | grep CONTROL_PLANE_API_URL | cut -d'=' -f2 | tr -d '"')
      azd env set CONTROL_PLANE_API_BASE_URL "$API_URL"

services:
  frontend:
    docker:
      buildArgs:
        - VITE_CONTROL_PLANE_API_BASE_URL=${CONTROL_PLANE_API_BASE_URL}
```

### Strategy 2: Runtime Configuration (Alternative)

Frontend fetches configuration from API at startup:

1. Provision infrastructure
2. Deploy services with placeholder config
3. Frontend calls `/api/config` at startup to get actual URLs

**Advantages:**

- No circular dependency
- Can update config without rebuild

**Disadvantages:**

- Extra API call on every page load
- Requires additional backend endpoint

**Example:**

```typescript
// frontend/src/config.ts
let apiBaseUrl: string | null = null;

export async function getApiBaseUrl(): Promise<string> {
  if (apiBaseUrl) return apiBaseUrl;

  const response = await fetch("/api/v1/config");
  const config = await response.json();
  apiBaseUrl = config.apiBaseUrl;
  return apiBaseUrl;
}
```

### Comparison Table

| Aspect                | Deploy-Time          | Runtime                                    |
| --------------------- | -------------------- | ------------------------------------------ |
| Performance           | ✅ Fastest           | ⚠️ Extra API call                          |
| Rebuild required?     | ✅ Yes               | ❌ No                                      |
| Configuration updates | ⚠️ Requires redeploy | ✅ Immediate                               |
| SPA compatibility     | ✅ Perfect           | ⚠️ Needs backend endpoint                  |
| Recommended for       | Production apps      | Admin portals with frequent config changes |

## Custom Workflow Configuration (NEW in 2024-2025)

Override default `azd up` behavior with custom workflows:

```yaml
workflows:
  up:
    steps:
      - azd: provision
      - run: echo "Custom step between provision and package"
      - azd: package
      - run: echo "Running tests before deploy"
      - run: npm test
      - azd: deploy
      - run: echo "Deployment complete"

  custom-workflow:
    steps:
      - azd: provision
      - run: ./custom-script.sh
      - azd: deploy --service api
```

Then run: `azd workflow run custom-workflow`

## Latest azd Features (2024-2025)

### Alpha Features (Enable Explicitly)

> ⚠️ **Alpha features are experimental.** They may change or be removed between azd releases without notice. Run `azd config list alpha` to see currently available alpha features, and confirm against the [azd changelog](https://github.com/Azure/azure-dev/releases) before relying on them in production pipelines.

**Custom Bicep Parameters:**

```yaml
# azure.yaml
alpha:
  infraParameters:
    imageTag: ${IMAGE_TAG}
    postgresAdminPassword: ${POSTGRES_ADMIN_PASSWORD}
```

Or via CLI:

```bash
azd config set alpha.infraParameters '{"imageTag": "abc123", "postgresAdminPassword": "..."}'
```

**Interactive Mode Control:**

```bash
# Disable interactive prompts in hooks
azd hooks run postprovision --no-prompt

# CI/CD-friendly deployment
azd up --no-prompt
```

### Testing Hooks Locally

```bash
# Test a specific hook without full deployment
azd hooks run postprovision

# Debug hook execution
azd hooks run postdeploy --debug
```

## azd Extensions (Beta)

azd extensions are modular components that extend azd functionality beyond its core capabilities. Extensions are currently in beta and provide additional commands, integrations, and tooling for specialized workflows.

### Extension Management Commands

```bash
# List available and installed extensions
azd extension list

# Install an extension
azd extension install <name>

# Upgrade an installed extension
azd extension upgrade <name>

# Remove an extension
azd extension uninstall <name>
```

### Extension Source Management

azd ships with the official registry preconfigured. You can add additional sources for development or custom extensions.

```bash
# List configured extension sources
azd extension source list

# Add the dev registry (opt-in, unsigned binaries — testing only)
azd extension source add dev -t url -l https://aka.ms/azd/extensions/registry/dev

# Add a custom source (URL-based)
azd extension source add my-source -t url -l <registry-url>

# Add a custom source (file-based, for local development)
azd extension source add local-dev -t file -l <path-to-local-registry>

# Remove an extension source
azd extension source remove dev
```

| Registry | URL | Signed | Use Case |
| --- | --- | --- | --- |
| Official | `https://aka.ms/azd/extensions/registry` | Yes | Production use (preconfigured) |
| Dev | `https://aka.ms/azd/extensions/registry/dev` | No | Testing pre-release extensions only |

### Dev Container Integration

Extensions can be auto-installed during dev container build by specifying them in `devcontainer.json`:

```json
{
  "features": {
    "ghcr.io/azure/azure-dev/azd:latest": {
      "extensions": "azure.ai.agents"
    }
  }
}
```

Multiple extensions can be specified as a comma-separated list.

### Required Versions in azure.yaml

Declare required extension versions in `azure.yaml` so that `azd up` fails fast if the extension is not installed:

```yaml
requiredVersions:
  extensions:
    azure.ai.agents: latest
```

## Azure AI Foundry Agent Extension

The `azure.ai.agents` extension enables scaffolding, deploying, invoking, and monitoring AI agents built with the [Agent Framework](https://github.com/microsoft/agent-framework) on Azure AI Foundry.

- **Extension ID:** `azure.ai.agents`
- **Requires:** azd version 1.21.3+
- **Install:** `azd extension install azure.ai.agents` (auto-installs when using `azd ai agent` commands)

### Key Commands

```bash
# Scaffold a new agent project from starter templates
azd ai agent init

# Invoke an agent with a prompt
azd ai agent invoke --prompt "Summarize this quarter's results"

# Deploy and run agents
azd ai agent run

# Monitor deployed agents
azd ai agent monitor

# Show agent details
azd ai agent show
```

### azure.yaml Configuration for Agent Services

```yaml
services:
  my-agent:
    host: azure.ai.agent
    config:
      model:
        deployments:
          - name: gpt-4o
            model:
              name: gpt-4o
              version: "2024-08-06"
            sku:
              name: GlobalStandard
              capacity: 10
```

### Key Environment Variables

| Variable | Description |
| --- | --- |
| `AZURE_AI_ACCOUNT_NAME` | Azure AI account name |
| `AZURE_AI_PROJECT_NAME` | Azure AI project name |
| `AZURE_AI_FOUNDRY_PROJECT_ENDPOINT` | Full endpoint URL for the Foundry project |

### Security

The extension uses managed identities by default with Entra ID authentication. Role assignments (Cognitive Services User, etc.) are created automatically during provisioning.

### Limitations

Hosted agents are currently limited to specific Azure regions. Check the [Foundry AI agent extension docs](https://learn.microsoft.com/azure/developer/azure-developer-cli/extensions/azure-ai-foundry-extension) for current region availability before planning deployments.

### Agent Definitions

Agent definitions and starter templates come from the [Agent Framework repo](https://github.com/microsoft/agent-framework).

## Environment Variables Flow

### Three-Level Configuration

1. **Local `.env`** - For local development only
2. **`.azure/<env>/.env`** - azd-managed, auto-populated from Bicep outputs
3. **`main.parameters.json`** - Maps env vars to Bicep parameters

### Parameter Injection Pattern

```json
// infra/main.parameters.json
{
  "parameters": {
    "environmentName": { "value": "${AZURE_ENV_NAME}" },
    "location": { "value": "${AZURE_LOCATION=eastus2}" },
    "azureOpenAiEndpoint": { "value": "${AZURE_OPENAI_ENDPOINT}" }
  }
}
```

Syntax: `${VAR_NAME}` or `${VAR_NAME=default_value}`

### Setting Environment Variables

```bash
# Set for current environment
azd env set AZURE_OPENAI_ENDPOINT "https://my-openai.openai.azure.com"
azd env set AZURE_SEARCH_ENDPOINT "https://my-search.search.windows.net"

# Set during init
azd env new prod
azd env set AZURE_OPENAI_ENDPOINT "..."
```

### Bicep Output → Environment Variable

```bicep
// In main.bicep - outputs auto-populate .azure/<env>/.env
output SERVICE_FRONTEND_URI string = frontend.outputs.uri
output SERVICE_BACKEND_URI string = backend.outputs.uri
output BACKEND_PRINCIPAL_ID string = backend.outputs.principalId
```

## Idempotent Deployments

### Why azd up is Idempotent

1. **Bicep is declarative** - Resources reconcile to desired state
2. **Remote builds tag uniquely** - Image tags include deployment timestamp
3. **ACR reuses layers** - Only changed layers upload

### Preserving Manual Changes

Custom domains added via Portal can be lost on redeploy. Preserve with hooks:

```yaml
hooks:
  preprovision:
    shell: sh
    run: |
      # Save custom domains before provision
      if az containerapp show --name "$FRONTEND_NAME" -g "$RG" &>/dev/null; then
        az containerapp show --name "$FRONTEND_NAME" -g "$RG" \
          --query "properties.configuration.ingress.customDomains" \
          -o json > /tmp/domains.json
      fi

  postprovision:
    shell: sh
    run: |
      # Verify/restore custom domains
      if [ -f /tmp/domains.json ]; then
        echo "Saved domains: $(cat /tmp/domains.json)"
      fi
```

### Handling Existing Resources

```bicep
// Reference existing ACR (don't recreate)
resource containerRegistry 'Microsoft.ContainerRegistry/registries@2025-11-01' existing = {
  name: containerRegistryName
}

// Set customDomains to null to preserve Portal-added domains
customDomains: empty(customDomainsParam) ? null : customDomainsParam
```

## Container App Service Discovery

Internal HTTP routing between Container Apps in same environment:

```bicep
// Backend reference in frontend env vars
env: [
  {
    name: 'BACKEND_URL'
    value: 'http://ca-backend-${resourceToken}'  // Internal DNS
  }
]
```

Frontend nginx proxies to internal URL:

```nginx
location /api {
    proxy_pass $BACKEND_URL;
}
```

## Managed Identity & RBAC

### Enable System-Assigned Identity

```bicep
resource containerApp 'Microsoft.App/containerApps@2025-07-01' = {
  identity: {
    type: 'SystemAssigned'
  }
}

output principalId string = containerApp.identity.principalId
```

### Post-Provision RBAC Assignment

```yaml
hooks:
  postprovision:
    shell: sh
    run: |
      PRINCIPAL_ID="${BACKEND_PRINCIPAL_ID}"

      # Azure OpenAI access
      az role assignment create \
        --assignee-object-id "$PRINCIPAL_ID" \
        --assignee-principal-type ServicePrincipal \
        --role "Cognitive Services OpenAI User" \
        --scope "$OPENAI_RESOURCE_ID" 2>/dev/null || true

      # Azure AI Search access
      az role assignment create \
        --assignee-object-id "$PRINCIPAL_ID" \
        --role "Search Index Data Reader" \
        --scope "$SEARCH_RESOURCE_ID" 2>/dev/null || true
```

## Container Image Versioning (CRITICAL)

**Anti-Pattern:** Using `:latest` tag makes deployments non-deterministic and impossible to roll back.

**Solution:** Parameterize image tags in Bicep and pass commit SHA from CI/CD.

### Bicep Pattern

```bicep
@description('Container image tag (e.g., commit SHA or semantic version)')
param imageTag string = 'latest'  // Default for local dev, override in CI/CD

module controlPlaneApi 'br/public:avm/res/app/container-app:0.21' = {
  params: {
    containers: [
      {
        name: 'control-plane-api'
        image: '${acr.outputs.loginServer}/atlas-control-plane-api:${imageTag}'  // ✅ Parameterized tag
        // ... other container config
      }
    ]
  }
}
```

### CI/CD Integration (GitHub Actions)

```yaml
- name: Configure azd with image tag
  run: |
    azd config set alpha.infraParameters '{"imageTag": "${{ github.sha }}"}'

- name: Deploy with versioned images
  run: azd up --no-prompt
```

**Result:** Every deployment uses immutable image tag (commit SHA) for reproducible deployments and easy rollback.

## Key Vault Secret Management (CRITICAL)

**Anti-Pattern:** Plain text passwords in Container App environment variables are visible in deployment history and configuration.

**Solution:** Use Key Vault secret references with managed identity authentication.

### Bicep Pattern

```bicep
// 1. Create Key Vault with secrets
module keyVault 'br/public:avm/res/key-vault/vault:0.13' = {
  params: {
    name: 'kv-${uniqueSuffix}'
    location: location
    sku: 'standard'
    enableRbacAuthorization: true

    secrets: [
      {
        name: 'postgres-admin-password'
        value: postgresAdminPassword
      }
      {
        name: 'postgres-connection-string'
        value: 'Host=${postgresql.outputs.fqdn};Database=atlas;Username=atlasadmin;Password=${postgresAdminPassword};SSL Mode=Require'
      }
    ]
  }
}

// 2. Grant managed identity Key Vault Secrets User role
module keyVaultSecretsRoleAssignment 'br/public:avm/ptn/authorization/resource-role-assignment:0.1.1' = {
  params: {
    principalId: managedIdentity.outputs.principalId
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', '4633458b-17de-408a-b874-0445c86b69e6') // Key Vault Secrets User
    resourceId: keyVault.outputs.resourceId
  }
}

// 3. Container App references Key Vault secret
module controlPlaneApi 'br/public:avm/res/app/container-app:0.21' = {
  params: {
    managedIdentities: {
      userAssignedResourceIds: [managedIdentity.outputs.resourceId]
    }

    secrets: {
      secureList: [
        {
          name: 'postgres-connection-string'
          keyVaultUrl: '${keyVault.outputs.uri}secrets/postgres-connection-string'
          identity: managedIdentity.outputs.resourceId
        }
      ]
    }

    containers: [
      {
        name: 'control-plane-api'
        image: '${acr.outputs.loginServer}/atlas-control-plane-api:${imageTag}'
        env: [
          {
            name: 'ConnectionStrings__DefaultConnection'
            secretRef: 'postgres-connection-string'  // ✅ References Key Vault secret
          }
        ]
      }
    ]
  }
}
```

**Security Benefits:**

- ✅ Passwords never visible in Container App configuration
- ✅ Deployment history doesn't expose secrets
- ✅ Secrets rotation doesn't require application redeployment
- ✅ Managed identity authentication (no service principal secrets)

## Deployment Validation & Health Checks (CRITICAL)

### GitHub Actions Validation Job

Run **before** actual deployment as a PR gate:

```yaml
jobs:
  validate:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - name: Validate Bicep syntax
        run: |
          az bicep build --file infrastructure/bicep/main.bicep --outdir /tmp
          echo "✅ Bicep syntax valid"

      - uses: azure/login@v2
        with:
          client-id: ${{ secrets.AZURE_CLIENT_ID }}
          tenant-id: ${{ secrets.AZURE_TENANT_ID }}
          subscription-id: ${{ secrets.AZURE_SUBSCRIPTION_ID }}

      - name: What-If Analysis
        run: |
          az deployment sub what-if \
            --location uksouth \
            --template-file infrastructure/bicep/main.bicep \
            --parameters \
              environment=prod \
              projectName=atlas \
              imageTag=${{ github.sha }}
```

### Post-Deployment Health Checks

```yaml
deploy:
  needs: validate
  environment: production # Requires manual approval

  steps:
    # ... azd up deployment ...

    - name: Health Check with Retries
      run: |
        API_URL=$(azd env get-values | grep apiUrl | cut -d'=' -f2 | tr -d '"')

        MAX_RETRIES=10
        RETRY_COUNT=0
        until [ $RETRY_COUNT -ge $MAX_RETRIES ]; do
          if curl -sf "${API_URL}/health" > /dev/null 2>&1; then
            echo "✅ API is healthy"
            exit 0
          fi
          RETRY_COUNT=$((RETRY_COUNT+1))
          echo "⏳ Waiting for API (attempt $RETRY_COUNT/$MAX_RETRIES)..."
          sleep 10
        done

        echo "❌ Health check failed after $MAX_RETRIES attempts"
        exit 1
```

## Anti-Patterns to Avoid (Lessons from Code Reviews)

### ❌ Anti-Pattern 1: Hardcoded `:latest` Tags

**Problem:**

```bicep
containers: [
  {
    name: 'api'
    image: '${acr.loginServer}/my-api:latest'  // ❌ Mutable tag
  }
]
```

**Impact:** Non-deterministic deployments, impossible to roll back, can't identify running version.

**Fix:** Parameterize image tags and pass commit SHA from CI/CD (see Container Image Versioning section above).

---

### ❌ Anti-Pattern 2: Plain Text Secrets in Environment Variables

**Problem:**

```bicep
env: [
  {
    name: 'DATABASE_CONNECTION_STRING'
    value: 'Host=...;Password=MyPassword123;...'  // ❌ Visible in config
  }
]
```

**Impact:** Secrets exposed in deployment history, configuration UI, and logs.

**Fix:** Use Key Vault secret references (see Key Vault Secret Management section above).

---

### ❌ Anti-Pattern 3: Invalid URL Outputs for Services Without Ingress

**Problem:**

```bicep
// Service has no ingress configured
module agentOrchestrator 'container-app.bicep' = {
  params: {
    ingressExternal: false  // Internal only
    // No ingressTargetPort specified
  }
}

// But outputs try to expose HTTPS URL
output agentOrchestratorUrl string = 'https://${agentOrchestrator.outputs.fqdn}'  // ❌ fqdn doesn't exist!
```

**Impact:** Deployment fails because Container Apps without ingress don't have an FQDN.

**Fix:** Remove invalid URL outputs for internal-only services, or add ingress configuration.

---

### ❌ Anti-Pattern 4: Suppressing Errors with `|| true`

**Problem:**

```bash
azd env new production --subscription $SUBSCRIPTION_ID || true  # ❌ Masks real errors
```

**Impact:** Legitimate failures (permissions, configuration issues) are silently ignored.

**Fix:**

```bash
if azd env show production >/dev/null 2>&1; then
  echo "Environment exists, skipping creation"
else
  azd env new production --subscription $SUBSCRIPTION_ID  # ✅ Fails on real errors
fi
```

---

### ❌ Anti-Pattern 5: Relative API Paths with Separate FQDNs

**Problem:**

```yaml
# azure.yaml
services:
  frontend:
    docker:
      buildArgs:
        - VITE_API_BASE_URL=/api/v1 # ❌ Relative path
```

**Impact:** Frontend deployed to `frontend-xyz.azurecontainerapps.io` tries to call `/api/v1` on its own domain instead of API domain `api-xyz.azurecontainerapps.io`.

**Fix:** Use full HTTPS URL from infrastructure outputs (see Circular Dependency Resolution section above).

---

### ❌ Anti-Pattern 6: No Deployment Validation

**Problem:**

```yaml
# GitHub Actions workflow
jobs:
  deploy:
    steps:
      - run: azd up --no-prompt # ❌ Deploys directly to production
```

**Impact:** Syntax errors, invalid parameters, or breaking changes deploy to production without review.

**Fix:** Add validation job with Bicep build + what-if analysis (see Deployment Validation section above).

---

### ❌ Anti-Pattern 7: No Health Checks After Deployment

**Problem:**

```yaml
steps:
  - run: azd up --no-prompt
  # ❌ Workflow succeeds even if services are unhealthy
```

**Impact:** Deployment marked successful but applications are crashing or unreachable.

**Fix:** Add post-deployment health checks with retry logic (see Health Checks section above).

---

### ❌ Anti-Pattern 8: Ignoring Hook Failures

**Problem:**

```yaml
hooks:
  postprovision:
    shell: sh
    run: |
      az role assignment create ... || true  # ❌ Hides failures
      ./migration.sh || true  # ❌ Hides migration failures
```

**Impact:** Critical setup steps fail silently, application deployed in broken state.

**Fix:** Only use `|| true` for idempotent operations that are safe to retry (e.g., RBAC assignments that may already exist). Never suppress migration or data setup failures.

---

### ❌ Anti-Pattern 9: Not Setting Image Pull Policy

**Problem:**

```bicep
// Bicep doesn't explicitly set imagePullPolicy
containers: [
  {
    image: '${acr.loginServer}/my-api:${imageTag}'
    // ❌ Defaults may cache old images
  }
]
```

**Impact:** Container Apps may serve stale images even after redeployment.

**Fix:** While Container Apps don't expose explicit imagePullPolicy, ensure unique image tags (commit SHA) to force fresh pulls.

---

### ❌ Anti-Pattern 10: Not Handling Environment Variable Circular Dependencies

**Problem:**

```yaml
# Frontend needs API URL at build time
services:
  frontend:
    docker:
      buildArgs:
        - VITE_API_URL=${API_URL} # ❌ API_URL not available until after provision
```

**Impact:** Build fails or uses wrong/empty API URL.

**Fix:** Use `postprovision` hook to capture infrastructure outputs before deploy (see Circular Dependency Resolution section above).

---

### ❌ Anti-Pattern 11: Inconsistent Environment Variable Names Across Files

**Problem:**

```bicep
// main.bicepparam uses:
param projectName = readEnvironmentVariable('NIMBUSIQ_PROJECT_NAME', 'nimbusiq')

// But post-deploy script reads:
$ProjectName = $envValues['AZURE_PROJECT_NAME'] ?? 'atlas'  # ❌ Different name!
```

**Impact:** Script uses default value instead of the one set by the user. Deployment succeeds but post-deploy automation finds wrong resources or creates mismatched names.

**Fix:** Establish a canonical env var name per concept and grep the repo before adding references:

```bash
# Before adding a new env var reference, verify consistency:
grep -r 'PROJECT_NAME' --include='*.bicepparam' --include='*.ps1' --include='*.yaml' --include='*.sh'
```

**Rule:** When `.bicepparam` reads `NIMBUSIQ_PROJECT_NAME`, all hooks and scripts must read `NIMBUSIQ_PROJECT_NAME` first (with fallback chain if backward compatibility needed).

---

### ❌ Anti-Pattern 12: Soft-Deleted Azure Resources Blocking Fresh Deployments

**Problem:** Previous `azd down` or manual deletion leaves Key Vault, Cognitive Services, or App Configuration in soft-deleted state. A fresh `azd up` with the same project name fails with `ConflictError` or `NameUnavailable`.

**Impact:** Fresh environment deployment blocked by ghost resources from prior environments.

**Fix:**

1. Use `uniqueString(subscription().subscriptionId, projectName, environment)` for globally unique names (already standard)
2. When reusing the same project name, purge soft-deleted resources first:
   ```bash
   az keyvault purge --name <vault-name>
   az cognitiveservices account purge --name <name> --resource-group <rg> --location <loc>
   ```
3. Or use a different `NIMBUSIQ_PROJECT_NAME` value for the new environment

## Common Commands

```bash
# Environment management
azd env list                        # List environments
azd env select <name>               # Switch environment
azd env get-values                  # Show all env vars
azd env set KEY value               # Set variable

# Deployment
azd up                              # Full provision + deploy
azd provision                       # Infrastructure only
azd deploy                          # Code deployment only
azd deploy --service backend        # Deploy single service

# Debugging
azd show                            # Show project status
az containerapp logs show -n <app> -g <rg> --follow  # Stream logs
```

## Reference Files

- **Bicep patterns**: See [references/bicep-patterns.md](references/bicep-patterns.md) for Container Apps modules with AVM
- **Troubleshooting**: See [references/troubleshooting.md](references/troubleshooting.md) for common deployment issues
- **azure.yaml schema**: See [references/azure-yaml-schema.md](references/azure-yaml-schema.md) for comprehensive options
- **Acceptance criteria**: See [references/acceptance-criteria.md](references/acceptance-criteria.md) for deployment validation

## Critical Reminders (Updated for v2.0)

### Must Do ✅

1. **Parameterize image tags** - Use `imageTag` parameter in Bicep, pass commit SHA from CI/CD (not `:latest`)
2. **Use Key Vault secret references** - Never plain text passwords in Container App environment variables
3. **Add deployment validation** - Bicep build + what-if analysis before production deployments
4. **Implement health checks** - Verify services are healthy after deployment with retry logic
5. **Use `remoteBuild: true`** - Local builds fail on M1/ARM Macs deploying to AMD64 Container Apps
6. **Capture infrastructure outputs** - Use `postprovision` hook to set environment variables from Bicep outputs
7. **Configure GitHub environment protection** - Require manual approval for production deployments
8. **Use managed identity for ACR + Key Vault** - Grant AcrPull and Key Vault Secrets User roles

### Must Not Do ❌

1. **Never hardcode `:latest` tags** - Makes deployments non-deterministic and impossible to roll back
2. **Never expose secrets in outputs** - Use Key Vault references instead of plain text
3. **Never output URLs for services without ingress** - Container Apps without ingress don't have FQDNs
4. **Never suppress legitimate errors** - Only use `|| true` for idempotent operations (RBAC assignments)
5. **Never use relative API paths with separate FQDNs** - Frontend needs full HTTPS URLs for cross-domain calls
6. **Never skip what-if analysis** - Prevents invalid deployments and shows impact preview
7. **Never deploy without health checks** - Deployment may succeed but services could be unhealthy

### Best Practices 🌟

1. **Bicep outputs auto-populate `.azure/<env>/.env`** - Don't manually edit these files
2. **Use `azd env set` for secrets** - Not main.parameters.json defaults
3. **Service tags (`azd-service-name`)** are required - azd uses these to find Container Apps
4. **Use lifecycle hooks** - `postprovision` for outputs, `predeploy` for migrations, `postdeploy` for health checks
5. **Follow circular dependency patterns** - Deploy-time config (hooks) or runtime config (API endpoint)
6. **Test hooks locally** - Use `azd hooks run <hook-name>` before committing
7. **Use AVM modules** - Azure Verified Modules provide production-ready, tested patterns
8. **Monitor deployment telemetry** - Use Azure Monitor and Application Insights for observability

## Currency and verification

- **Date checked:** 2026-03-31
- **AZD latest stable:** 1.23.10 (2026-03-17) — verify at [Azure Dev CLI releases](https://github.com/Azure/azure-dev/releases)
- **azd extensions:** Beta — verify at [azd extensions overview](https://learn.microsoft.com/azure/developer/azure-developer-cli/extensions/overview)
- **azure.ai.agents extension:** Requires azd 1.21.3+ — verify at [Foundry AI agent extension](https://learn.microsoft.com/azure/developer/azure-developer-cli/extensions/azure-ai-foundry-extension)
- **Sources:** [AZD docs](https://learn.microsoft.com/azure/developer/azure-developer-cli/), [AZD GitHub releases](https://github.com/Azure/azure-dev/releases), [azd extensions overview](https://learn.microsoft.com/azure/developer/azure-developer-cli/extensions/overview)
- **Verification steps:** Run `azd version`; run `azd extension list` to check available extensions; check the release notes for breaking changes before upgrading.

### Known pitfalls

| Area                           | Pitfall                                                                                                              | Mitigation                                                                                       |
| ------------------------------ | -------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------ |
| `alpha.infraParameters`        | This is an experimental feature; `azd config set alpha.infraParameters` may not exist in older or newer azd versions | Run `azd config list alpha` to check availability; pin your azd version in CI with `AZD_VERSION` |
| `:latest` container tags       | Non-deterministic deployments; impossible to roll back                                                               | Always pass a specific tag (commit SHA) via the `imageTag` Bicep parameter                       |
| `remoteBuild: true`            | Required for ARM64/M1 Macs building AMD64 Container Apps; skipping this causes arch-mismatch crashes                 | Set `remoteBuild: true` in `azure.yaml` for all Container App services                           |
| Key Vault soft-delete          | Re-deploying with the same Key Vault name within the purge-protection window fails                                   | Use `uniqueString()` in Bicep names or purge the soft-deleted vault before redeploying           |
| Service tag `azd-service-name` | Missing tag means azd cannot find the Container App to update                                                        | Every Container App resource used by azd must have the `azd-service-name` tag set                |
| Hook errors silently ignored   | If a lifecycle hook script exits non-zero, azd may warn but continue by default                                      | Use `failOnError: true` in hook config and test hooks with `azd hooks run <hook>`                |
| Extensions beta status | Extensions are currently in beta; commands and behavior may change between azd releases | Pin azd version in CI; test extension workflows in dev before production |
| Dev registry unsigned binaries | Extensions from the dev registry (`https://aka.ms/azd/extensions/registry/dev`) are NOT signed | Only use dev registry for testing; use official registry for production |
| Foundry agent region limits | Hosted agents may be limited to specific Azure regions | Check [Foundry docs](https://learn.microsoft.com/azure/developer/azure-developer-cli/extensions/azure-ai-foundry-extension) for current region availability before planning deployments |
| `requiredVersions` in azure.yaml | If `requiredVersions.extensions` specifies an extension, `azd up` will fail if the extension is not installed | Use dev container auto-install or add `azd extension install` to CI setup |

## Version History

- **v2.1.0** (2026-03-28): Added azd extensions framework (beta), Azure AI Foundry agent extension (`azure.ai.agents`), extension source management, dev container integration, resolved [VERIFY] blocks
- **v2.0.1** (2026-03-07): Added Currency and verification section, [VERIFY] blocks for alpha.infraParameters, updated pitfall table
- **v2.0.0** (2026-02-15): Added image versioning, Key Vault patterns, lifecycle hooks, circular dependencies, validation gates, health checks, anti-patterns, latest 2024-2025 features
- **v1.0.0** (2024): Initial skill with basic azd + Container Apps deployment patterns
