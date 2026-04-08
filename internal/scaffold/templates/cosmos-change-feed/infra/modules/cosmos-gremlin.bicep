// cosmos-gremlin.bicep — Azure Cosmos DB account with Gremlin API
// Provisions: Cosmos DB account (Gremlin), database, graph container, and Key Vault secrets
// for the account endpoint and primary key.

@description('The Azure region for all resources.')
param location string

@description('Environment name suffix applied to all resource names.')
param environmentName string

@description('Tags to apply to all resources.')
param tags object = {}

@description('Name of the existing Key Vault where Cosmos secrets will be stored.')
param keyVaultName string

var suffix = uniqueString(resourceGroup().id, environmentName)
var accountName = 'drasi-cosmos-${suffix}'
var databaseName = 'drasidb'
var graphName = 'drasi-graph'

// ---------------------------------------------------------------------------
// Cosmos DB Account — Gremlin API
// Session consistency is the default and recommended for most workloads.
// Serverless capacity mode keeps costs low for development/starter templates.
// ---------------------------------------------------------------------------
resource cosmosAccount 'Microsoft.DocumentDB/databaseAccounts@2024-11-15' = {
  name: accountName
  location: location
  tags: union(tags, { component: 'data', 'managed-by': 'azd' })
  kind: 'GlobalDocumentDB'
  properties: {
    databaseAccountOfferType: 'Standard'
    capabilities: [
      { name: 'EnableGremlin' }
      { name: 'EnableServerless' }
    ]
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
  }
}

// ---------------------------------------------------------------------------
// Gremlin Database
// ---------------------------------------------------------------------------
resource gremlinDatabase 'Microsoft.DocumentDB/databaseAccounts/gremlinDatabases@2024-11-15' = {
  parent: cosmosAccount
  name: databaseName
  properties: {
    resource: {
      id: databaseName
    }
  }
}

// ---------------------------------------------------------------------------
// Gremlin Graph — partition key /pk
// Serverless accounts do not require throughput configuration.
// ---------------------------------------------------------------------------
resource gremlinGraph 'Microsoft.DocumentDB/databaseAccounts/gremlinDatabases/graphs@2024-11-15' = {
  parent: gremlinDatabase
  name: graphName
  properties: {
    resource: {
      id: graphName
      partitionKey: {
        paths: ['/pk']
        kind: 'Hash'
      }
    }
  }
}

// ---------------------------------------------------------------------------
// Key Vault Secrets — store Cosmos connection details for Drasi source
// The Drasi resource-provider pod reads these via Workload Identity.
// ---------------------------------------------------------------------------
resource existingKeyVault 'Microsoft.KeyVault/vaults@2023-07-01' existing = {
  name: keyVaultName
}

resource cosmosEndpointSecret 'Microsoft.KeyVault/vaults/secrets@2023-07-01' = {
  parent: existingKeyVault
  name: 'cosmos-account-endpoint'
  properties: {
    value: cosmosAccount.properties.documentEndpoint
  }
}

resource cosmosMasterKeySecret 'Microsoft.KeyVault/vaults/secrets@2023-07-01' = {
  parent: existingKeyVault
  name: 'cosmos-master-key'
  properties: {
    value: cosmosAccount.listKeys().primaryMasterKey
  }
}

// ---------------------------------------------------------------------------
// Outputs
// ---------------------------------------------------------------------------
output accountName string = cosmosAccount.name
output databaseName string = databaseName
output graphName string = graphName
output accountEndpoint string = cosmosAccount.properties.documentEndpoint
