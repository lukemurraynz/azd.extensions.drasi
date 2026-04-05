targetScope = 'resourceGroup'

@description('Azure region for the Event Hubs namespace.')
param location string

@description('Environment name used to stamp tags.')
param environmentName string

@description('Name of the Event Hubs namespace to create.')
param eventHubNamespaceName string

@description('Tags applied to the Event Hubs namespace.')
param tags object

var effectiveTags = union(tags, {
  'azd-env-name': environmentName
})

resource eventHubNamespace 'Microsoft.EventHub/namespaces@2024-01-01' = {
  name: eventHubNamespaceName
  location: location
  tags: effectiveTags
  sku: {
    name: 'Standard'
    tier: 'Standard'
  }
  properties: {
    minimumTlsVersion: '1.2'
  }
}

output eventHubNamespaceId string = eventHubNamespace.id
output eventHubNamespaceFqdn string = eventHubNamespace.properties.serviceBusEndpoint
