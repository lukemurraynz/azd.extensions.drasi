targetScope = 'resourceGroup'

@description('Azure region for the Service Bus namespace.')
param location string

@description('Environment name used to stamp tags.')
param environmentName string

@description('Name of the Service Bus namespace to create.')
param serviceBusNamespaceName string

@description('Tags applied to the Service Bus namespace.')
param tags object

var effectiveTags = union(tags, {
  'azd-env-name': environmentName
})

resource serviceBusNamespace 'Microsoft.ServiceBus/namespaces@2024-01-01' = {
  name: serviceBusNamespaceName
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

output serviceBusNamespaceId string = serviceBusNamespace.id
output serviceBusEndpoint string = serviceBusNamespace.properties.serviceBusEndpoint
