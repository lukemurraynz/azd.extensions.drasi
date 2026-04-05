targetScope = 'resourceGroup'

@description('Azure region for the Cosmos DB account.')
param location string

@description('Environment name used to stamp tags.')
param environmentName string

@description('Name of the Cosmos DB account to create.')
param cosmosAccountName string

@description('Tags applied to the Cosmos DB account.')
param tags object

var effectiveTags = union(tags, {
  'azd-env-name': environmentName
})

resource cosmosAccount 'Microsoft.DocumentDB/databaseAccounts@2023-11-15' = {
  name: cosmosAccountName
  location: location
  tags: effectiveTags
  kind: 'GlobalDocumentDB'
  properties: {
    capabilities: [
      {
        name: 'EnableGremlin'
      }
    ]
    consistencyPolicy: {
      defaultConsistencyLevel: 'Session'
    }
    databaseAccountOfferType: 'Standard'
    locations: [
      {
        locationName: location
        failoverPriority: 0
      }
    ]
    publicNetworkAccess: 'Enabled'
  }
}

output cosmosId string = cosmosAccount.id
output cosmosEndpoint string = cosmosAccount.properties.documentEndpoint
