targetScope = 'resourceGroup'

@description('Azure region for the Azure Container Registry.')
param location string

@description('Environment name used to stamp tags.')
param environmentName string

@description('Name of the Azure Container Registry to create.')
param acrName string

@description('Tags applied to the Azure Container Registry.')
param tags object

var effectiveTags = union(tags, {
  'azd-env-name': environmentName
})

resource containerRegistry 'Microsoft.ContainerRegistry/registries@2023-07-01' = {
  name: acrName
  location: location
  tags: effectiveTags
  sku: {
    name: 'Premium'
  }
  properties: {
    adminUserEnabled: false
    publicNetworkAccess: 'Enabled'
  }
}

output acrId string = containerRegistry.id
output acrLoginServer string = containerRegistry.properties.loginServer
