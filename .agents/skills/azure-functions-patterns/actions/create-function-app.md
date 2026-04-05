# Create Function App

## Steps

### Step 1 — Choose Trigger and Hosting Plan

Determine the trigger type and hosting plan:

| Hosting Plan     | Best For                                     |
| ---------------- | -------------------------------------------- |
| Consumption      | Event-driven, low-cost, auto-scale to zero   |
| Flex Consumption | Per-function scaling, VNet, larger instances |
| Premium          | VNet, long-running, pre-warmed instances     |
| Dedicated        | Existing App Service Plan, predictable costs |

### Step 2 — Deploy Infrastructure

Deploy the Function App with identity-based storage:

```bicep
param location string = resourceGroup().location
param functionAppName string
param appInsightsConnectionString string

resource storageAccount 'Microsoft.Storage/storageAccounts@2025-08-01' = {
  name: replace('${functionAppName}st', '-', '')
  location: location
  sku: { name: 'Standard_LRS' }
  kind: 'StorageV2'
  properties: {
    allowSharedKeyAccess: false
    minimumTlsVersion: 'TLS1_2'
    supportsHttpsTrafficOnly: true
  }
}

resource hostingPlan 'Microsoft.Web/serverfarms@2025-03-01' = {
  name: '${functionAppName}-plan'
  location: location
  sku: {
    name: 'Y1'
    tier: 'Dynamic'
  }
}

resource functionApp 'Microsoft.Web/sites@2025-03-01' = {
  name: functionAppName
  location: location
  kind: 'functionapp'
  identity: {
    type: 'SystemAssigned'
  }
  properties: {
    serverFarmId: hostingPlan.id
    httpsOnly: true
    siteConfig: {
      netFrameworkVersion: 'v10.0'
      ftpsState: 'Disabled'
      minTlsVersion: '1.2'
      appSettings: [
        { name: 'AzureWebJobsStorage__accountName', value: storageAccount.name }
        { name: 'FUNCTIONS_EXTENSION_VERSION', value: '~4' }
        { name: 'FUNCTIONS_WORKER_RUNTIME', value: 'dotnet-isolated' }
        { name: 'APPLICATIONINSIGHTS_CONNECTION_STRING', value: appInsightsConnectionString }
      ]
    }
  }
}

// Assign Storage Blob Data Owner for identity-based storage access
resource storageRoleAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(storageAccount.id, functionApp.id, 'b7e6dc6d-f1e8-4753-8033-0f276bb0955b')
  scope: storageAccount
  properties: {
    principalId: functionApp.identity.principalId
    roleDefinitionId: subscriptionResourceId(
      'Microsoft.Authorization/roleDefinitions',
      'b7e6dc6d-f1e8-4753-8033-0f276bb0955b' // Storage Blob Data Owner
    )
    principalType: 'ServicePrincipal'
  }
}
```

### Step 3 — Create the Function Project

```bash
# Create isolated worker project
func init MyFunctionApp --worker-runtime dotnet-isolated --target-framework net10.0

# Add a function
cd MyFunctionApp
func new --name MyFunction --template "HTTP trigger" --authlevel anonymous
```

### Step 4 — Configure Identity-Based Connections

For each Azure service the function accesses, use identity-based connection settings:

```json
{
  "Values": {
    "AzureWebJobsStorage__accountName": "mystorageaccount",
    "ServiceBusConnection__fullyQualifiedNamespace": "mynamespace.servicebus.windows.net"
  }
}
```

Assign appropriate RBAC roles to the Function App managed identity for each service.

### Step 5 — Deploy and Verify

```bash
# Deploy via Azure CLI
az functionapp deployment source config-zip \
  --resource-group mygroup \
  --name myfunctionapp \
  --src ./publish.zip

# Or deploy via azd
azd up
```

Verify:

- Function executes successfully for target trigger
- Managed identity connects to all dependent services
- Application Insights receives telemetry
- No connection strings or keys in application settings
