targetScope = 'resourceGroup'

@description('Azure region for the managed identity.')
param location string

@description('Environment name used to stamp tags.')
param environmentName string

@description('Name of the user-assigned managed identity.')
param uamiName string

@description('Resource ID of the Key Vault that the identity should read secrets from.')
param keyVaultId string

@description('Resource ID of the Log Analytics workspace used for monitoring role assignment scope.')
param logAnalyticsWorkspaceId string

@description('Set to true to grant AcrPull on the private ACR resource.')
param usePrivateAcr bool = false

@description('Resource ID of the Azure Container Registry when private ACR is enabled.')
param acrId string = ''

@description('Tags applied to the managed identity.')
param tags object

var effectiveTags = union(tags, {
  'azd-env-name': environmentName
})
var keyVaultSecretsUserRoleId = '4633458b-17de-408a-b874-0445c86b69e6'
var monitoringMetricsPublisherRoleId = '3913510d-42f4-4e42-8a64-420c390055eb'
var acrPullRoleId = '7f951dda-4ed3-4680-a7ca-43fe172d538d'

resource uami 'Microsoft.ManagedIdentity/userAssignedIdentities@2023-01-31' = {
  name: uamiName
  location: location
  tags: effectiveTags
}

resource keyVault 'Microsoft.KeyVault/vaults@2023-07-01' existing = {
  name: last(split(keyVaultId, '/'))
}

resource logAnalyticsWorkspace 'Microsoft.OperationalInsights/workspaces@2023-09-01' existing = {
  name: last(split(logAnalyticsWorkspaceId, '/'))
}

resource acr 'Microsoft.ContainerRegistry/registries@2023-07-01' existing = if (usePrivateAcr) {
  name: last(split(acrId, '/'))
}

resource keyVaultSecretsUserAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  scope: keyVault
  name: guid(keyVault.id, uami.id, keyVaultSecretsUserRoleId)
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', keyVaultSecretsUserRoleId)
    principalId: uami.properties.principalId
    principalType: 'ServicePrincipal'
  }
}

resource monitoringMetricsPublisherAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  scope: logAnalyticsWorkspace
  name: guid(logAnalyticsWorkspace.id, uami.id, monitoringMetricsPublisherRoleId)
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', monitoringMetricsPublisherRoleId)
    principalId: uami.properties.principalId
    principalType: 'ServicePrincipal'
  }
}

resource acrPullAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = if (usePrivateAcr) {
  scope: acr
  name: guid(acr.id, uami.id, acrPullRoleId)
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', acrPullRoleId)
    principalId: uami.properties.principalId
    principalType: 'ServicePrincipal'
  }
}

output uamiId string = uami.id
output uamiPrincipalId string = uami.properties.principalId
output uamiClientId string = uami.properties.clientId
