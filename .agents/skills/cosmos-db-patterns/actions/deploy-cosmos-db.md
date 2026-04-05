# Deploy Cosmos DB

## Steps

### Step 1 — Choose Throughput Mode

| Mode        | Best For                                   |
| ----------- | ------------------------------------------ |
| Serverless  | Development, sporadic traffic, < 1000 RU/s |
| Autoscale   | Production, variable workloads             |
| Provisioned | Steady workloads, predictable cost         |

### Step 2 — Design Partition Key

Select a partition key based on:

- Most common query filter (`WHERE c.partitionKey = @value`)
- High cardinality (many distinct values)
- Even distribution of storage and throughput

### Step 3 — Deploy Infrastructure

```bicep
param location string = resourceGroup().location
param accountName string
param databaseName string = 'appdb'
param principalId string

resource cosmosAccount 'Microsoft.DocumentDB/databaseAccounts@2025-10-15' = {
  name: accountName
  location: location
  kind: 'GlobalDocumentDB'
  properties: {
    databaseAccountOfferType: 'Standard'
    disableLocalAuth: true
    consistencyPolicy: {
      defaultConsistencyLevel: 'Session'
    }
    locations: [
      {
        locationName: location
        failoverPriority: 0
        isZoneRedundant: false
      }
    ]
    capabilities: [
      { name: 'EnableServerless' }
    ]
  }
}

resource database 'Microsoft.DocumentDB/databaseAccounts/sqlDatabases@2025-10-15' = {
  parent: cosmosAccount
  name: databaseName
  properties: {
    resource: { id: databaseName }
  }
}

resource container 'Microsoft.DocumentDB/databaseAccounts/sqlDatabases/sqlContainers@2025-10-15' = {
  parent: database
  name: 'items'
  properties: {
    resource: {
      id: 'items'
      partitionKey: {
        paths: [ '/partitionKey' ]
        kind: 'Hash'
        version: 2
      }
      indexingPolicy: {
        indexingMode: 'consistent'
        includedPaths: [ { path: '/*' } ]
        excludedPaths: [ { path: '/"_etag"/?' } ]
      }
    }
  }
}

// Assign Cosmos DB Built-in Data Contributor
resource dataContributorRole 'Microsoft.DocumentDB/databaseAccounts/sqlRoleAssignments@2025-10-15' = {
  parent: cosmosAccount
  name: guid(cosmosAccount.id, principalId, '00000000-0000-0000-0000-000000000002')
  properties: {
    roleDefinitionId: '${cosmosAccount.id}/sqlRoleDefinitions/00000000-0000-0000-0000-000000000002'
    principalId: principalId
    scope: cosmosAccount.id
  }
}
```

### Step 4 — Configure Application

Use identity-based connection (no connection string):

```csharp
// Program.cs
builder.Services.AddSingleton(sp =>
{
    var endpoint = builder.Configuration["CosmosDb:Endpoint"];
    return new CosmosClient(endpoint, new DefaultAzureCredential());
});
```

Application setting:

```json
{
  "CosmosDb__Endpoint": "https://myaccount.documents.azure.com:443/"
}
```

### Step 5 — Verify

- Account deployed with `disableLocalAuth: true`
- RBAC data plane role assigned to application managed identity
- Application connects using `DefaultAzureCredential`
- Diagnostic settings forwarding to Log Analytics
- Queries include partition key filter
