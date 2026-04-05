targetScope = 'resourceGroup'

@description('Name of the existing user-assigned managed identity.')
param uamiName string

@description('OIDC issuer URL emitted by the AKS cluster.')
param oidcIssuerUrl string

@description('Namespace containing the Drasi resource provider service account.')
param drasiNamespace string = 'drasi-system'

resource uami 'Microsoft.ManagedIdentity/userAssignedIdentities@2023-01-31' existing = {
  name: uamiName
}

resource fedCred 'Microsoft.ManagedIdentity/userAssignedIdentities/federatedIdentityCredentials@2023-01-31' = {
  parent: uami
  name: 'drasi-fedcred'
  properties: {
    issuer: oidcIssuerUrl
    subject: 'system:serviceaccount:${drasiNamespace}:drasi-resource-provider'
    audiences: [
      'api://AzureADTokenExchange'
    ]
  }
}

output federatedCredentialId string = fedCred.id
