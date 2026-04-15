// eventhub.bicep — Azure Event Hubs for Drasi routing source
// Provisions an Event Hubs namespace, event hub, and shared access policy.
// The connection string is stored in Key Vault for secure secret sync to AKS.

@description('The Azure region for all resources.')
param location string

@description('Environment name suffix applied to all resource names.')
param environmentName string

@description('Tags to apply to all resources.')
param tags object = {}

@description('Resource ID of a user-assigned managed identity for the deployment script.')
param scriptIdentityId string = ''

// Event Hubs namespace name (max 50 chars, alphanumeric and hyphens)
var namespaceName = 'evh-${environmentName}-${uniqueString(resourceGroup().id)}'
var eventHubName = 'drasi-events'
var consumerGroup = '$Default'

// Throughput settings for development (auto-inflate disabled)
var messageRetentionDays = 1
var partitionCount = 4

// Event Hubs namespace
resource eventHubNamespace 'Microsoft.EventHub/namespaces@2024-01-01' = {
  name: namespaceName
  location: location
  tags: tags
  sku: {
    name: 'Standard'
    tier: 'Standard'
  }
  properties: {
    minimumTlsVersion: '1.2'
    publicNetworkAccess: 'Enabled'
  }
}

// Event Hub with routing enabled for Drasi
resource eventHub 'Microsoft.EventHub/namespaces/eventhubs@2024-01-01' = {
  parent: eventHubNamespace
  name: eventHubName
  properties: {
    messageRetentionInDays: messageRetentionDays
    partitionCount: partitionCount
    status: 'Active'
  }
}

// Default consumer group (Drasi will use this)
resource consumerGroupResource 'Microsoft.EventHub/namespaces/eventhubs/consumergroups@2024-01-01' = {
  parent: eventHub
  name: consumerGroup
  properties: {}
}

// Shared access policy with Listen and Send permissions
resource sharedAccessPolicy 'Microsoft.EventHub/namespaces/eventhubs/authorizationRules@2024-01-01' = {
  parent: eventHub
  name: 'drasi-policy'
  properties: {
    rights: [
      'Listen'
      'Send'
    ]
  }
}

// Retrieve the connection string for the shared access policy
resource namespaceAuthorizationRule 'Microsoft.EventHub/namespaces/authorizationRules@2024-01-01' existing = {
  parent: eventHubNamespace
  name: 'RootManageSharedAccessKey'
}

var connectionString = listKeys(namespaceAuthorizationRule.id, namespaceAuthorizationRule.apiVersion).primaryConnectionString

// Store the connection string in Key Vault
resource kvRef 'Microsoft.KeyVault/vaults@2023-07-01' existing = {
  name: 'drasi-kv-${uniqueString(resourceGroup().id)}'
}

resource eventHubSecret 'Microsoft.KeyVault/vaults/secrets@2023-07-01' = {
  parent: kvRef
  name: 'eventhub-connection-string'
  properties: {
    value: connectionString
  }
}

// Allow Azure services (including AKS) to access Event Hubs
resource networkRuleSet 'Microsoft.EventHub/namespaces/networkRuleSets@2024-01-01' = {
  parent: eventHubNamespace
  name: 'default'
  properties: {
    trustedServiceAccessEnabled: true
    defaultAction: 'Allow'
    virtualNetworkRules: []
    ipRules: []
  }
}

output namespaceName string = eventHubNamespace.name
output namespaceId string = eventHubNamespace.id
output eventHubName string = eventHub.name
output eventHubId string = eventHub.id
output connectionString string = connectionString
output consumerGroup string = consumerGroup
output fqdn string = eventHubNamespace.properties.serviceBusEndpoint
