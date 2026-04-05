targetScope = 'resourceGroup'

@description('Azure region for the Key Vault.')
param location string

@description('Environment name used to stamp tags.')
param environmentName string

@description('Name of the Key Vault to create.')
@minLength(3)
@maxLength(24)
param keyVaultName string

@description('Tags applied to the Key Vault.')
param tags object

var effectiveTags = union(tags, {
  'azd-env-name': environmentName
})

resource keyVault 'Microsoft.KeyVault/vaults@2023-07-01' = {
  name: keyVaultName
  location: location
  tags: effectiveTags
  properties: {
    tenantId: subscription().tenantId
    enableRbacAuthorization: true
    enablePurgeProtection: true
    enabledForDeployment: false
    enabledForDiskEncryption: false
    enabledForTemplateDeployment: false
    publicNetworkAccess: 'Enabled'
    softDeleteRetentionInDays: 90
    sku: {
      family: 'A'
      name: 'standard'
    }
  }
}

output keyVaultId string = keyVault.id
output keyVaultUri string = keyVault.properties.vaultUri
