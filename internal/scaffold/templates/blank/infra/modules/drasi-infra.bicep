// drasi-infra.bicep — Drasi infrastructure module
// Provisions all Azure resources required to run the Drasi reactive data pipeline runtime:
// Log Analytics, managed identities, Key Vault, NAT Gateway, VNet, and AKS (AVM modules).

@description('The Azure region for all resources.')
param location string

@description('Environment name suffix applied to all resource names.')
param environmentName string

@description('Tags to apply to all resources.')
param tags object

// ---------------------------------------------------------------------------
// Naming
// uniqueString ensures all globally-unique names are deterministic and
// collision-free per resource group + environment combination.
// ---------------------------------------------------------------------------
var suffix = take(uniqueString(resourceGroup().id, environmentName), 10)

var logAnalyticsName     = 'log-drasi-${suffix}'
var workloadUamiName     = 'drasi-id-${suffix}'
var controlPlaneUamiName = 'drasi-cp-${suffix}'
var keyVaultName         = take('kv-drasi-${suffix}', 24)
var publicIpName         = 'pip-drasi-nat-${suffix}'
var natGatewayName       = 'ng-drasi-${suffix}'
var vnetName             = 'vnet-drasi-${suffix}'
var aksName              = 'drasi-aks-${suffix}'
var federatedCredName    = 'drasi-resource-provider-fed'

// ---------------------------------------------------------------------------
// Built-in role definition IDs (subscription-independent format)
// Verified: https://learn.microsoft.com/azure/role-based-access-control/built-in-roles
// ---------------------------------------------------------------------------
var keyVaultSecretsUserRoleId         = '4633458b-17de-408a-b874-0445c86b69e6'
var networkContributorRoleId          = '4d97b98b-1d4f-4787-a291-c67834d212e7'
var monitoringMetricsPublisherRoleId  = '3913510d-42f4-4e42-8a64-420c390055eb'

// ---------------------------------------------------------------------------
// 1. Log Analytics workspace
// AVM: br/public:avm/res/operational-insights/workspace:0.15.0
// ---------------------------------------------------------------------------
module logAnalytics 'br/public:avm/res/operational-insights/workspace:0.15.0' = {
  name: 'log-analytics'
  params: {
    name: logAnalyticsName
    location: location
    tags: tags
    skuName: 'PerGB2018'
    dataRetention: 30
  }
}

// ---------------------------------------------------------------------------
// 2. Workload user-assigned managed identity (used by Drasi resource-provider pod)
// Native resource — no AVM module for standalone UAMI.
// API version verified: az provider show --namespace Microsoft.ManagedIdentity
// ---------------------------------------------------------------------------
resource workloadUami 'Microsoft.ManagedIdentity/userAssignedIdentities@2023-01-31' = {
  name: workloadUamiName
  location: location
  tags: tags
}

// ---------------------------------------------------------------------------
// 3. Control-plane user-assigned managed identity (used by AKS control plane)
// ---------------------------------------------------------------------------
resource controlPlaneUami 'Microsoft.ManagedIdentity/userAssignedIdentities@2023-01-31' = {
  name: controlPlaneUamiName
  location: location
  tags: tags
}

// ---------------------------------------------------------------------------
// 4. Key Vault — RBAC auth, purge protection, 7-day soft delete
// AVM: br/public:avm/res/key-vault/vault:0.13.3
// ---------------------------------------------------------------------------
module keyVault 'br/public:avm/res/key-vault/vault:0.13.3' = {
  name: 'key-vault'
  params: {
    name: keyVaultName
    location: location
    tags: tags
    enableRbacAuthorization: true
    enablePurgeProtection: true
    softDeleteRetentionInDays: 7
    enableVaultForDeployment: false
    enableVaultForDiskEncryption: false
    enableVaultForTemplateDeployment: false
  }
}

// ---------------------------------------------------------------------------
// 5. Key Vault Secrets User — workload UAMI on Key Vault scope
// BCP120 fix: role assignment name must use static inputs calculable at deployment start.
// We use resourceGroup().id + static names rather than module output resourceId.
// ---------------------------------------------------------------------------
resource keyVaultResource 'Microsoft.KeyVault/vaults@2023-07-01' existing = {
  name: keyVaultName
  dependsOn: [keyVault]
}

resource kvSecretsUserAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(resourceGroup().id, workloadUamiName, keyVaultSecretsUserRoleId)
  scope: keyVaultResource
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', keyVaultSecretsUserRoleId)
    principalId: workloadUami.properties.principalId
    principalType: 'ServicePrincipal'
  }
}

// ---------------------------------------------------------------------------
// 6. Public IP for NAT Gateway (Standard, Static, zones 1/2/3)
// API version verified: az provider show --namespace Microsoft.Network
// BCP034 fix: zones must be string[] for publicIPAddresses (ARM requirement).
// ---------------------------------------------------------------------------
resource natPublicIp 'Microsoft.Network/publicIPAddresses@2024-10-01' = {
  name: publicIpName
  location: location
  tags: tags
  sku: {
    name: 'Standard'
    tier: 'Regional'
  }
  zones: ['1', '2', '3']
  properties: {
    publicIPAllocationMethod: 'Static'
    publicIPAddressVersion: 'IPv4'
  }
}

// ---------------------------------------------------------------------------
// 7. NAT Gateway (Standard, zone 1, 10-minute idle timeout)
// ---------------------------------------------------------------------------
resource natGateway 'Microsoft.Network/natGateways@2024-10-01' = {
  name: natGatewayName
  location: location
  tags: tags
  sku: {
    name: 'Standard'
  }
  zones: ['1']
  properties: {
    idleTimeoutInMinutes: 10
    publicIpAddresses: [
      { id: natPublicIp.id }
    ]
  }
}

// ---------------------------------------------------------------------------
// 8. Virtual Network — 10.0.0.0/16, AKS subnet 10.0.0.0/22 with NAT GW
// AVM: br/public:avm/res/network/virtual-network:0.8.0
// ---------------------------------------------------------------------------
module vnet 'br/public:avm/res/network/virtual-network:0.8.0' = {
  name: 'virtual-network'
  params: {
    name: vnetName
    location: location
    tags: tags
    addressPrefixes: ['10.0.0.0/16']
    subnets: [
      {
        name: 'snet-aks'
        addressPrefix: '10.0.0.0/22'
        natGatewayResourceId: natGateway.id
      }
    ]
  }
}

// ---------------------------------------------------------------------------
// 9. Network Contributor — control-plane UAMI on VNet scope
// BCP120 fix: use static name inputs (vnetName + uami name) for guid().
// ---------------------------------------------------------------------------
resource vnetResource 'Microsoft.Network/virtualNetworks@2024-05-01' existing = {
  name: vnetName
  dependsOn: [vnet]
}

resource vnetNetworkContributorAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(resourceGroup().id, controlPlaneUamiName, networkContributorRoleId)
  scope: vnetResource
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', networkContributorRoleId)
    principalId: controlPlaneUami.properties.principalId
    principalType: 'ServicePrincipal'
  }
}

// ---------------------------------------------------------------------------
// 10. AKS cluster — Azure CNI Overlay + Cilium, Workload Identity, OIDC
// AVM: br/public:avm/res/container-service/managed-cluster:0.13.0
//
// AVM 0.13.0 parameter mapping (verified from module main.bicep):
//   - oidcIssuerEnabled        → enableOidcIssuerProfile
//   - enableWorkloadIdentity   → securityProfile.workloadIdentity.enabled
//   - enableAzureRbac          → enableRBAC
//   - autoUpgradeChannel       → autoUpgradeProfile.upgradeChannel
//   - nodeOsUpgradeChannel     → autoUpgradeProfile.nodeOSUpgradeChannel
//   - diskCSIDriverEnabled     → enableStorageProfileDiskCSIDriver
//   - agentPools (required)    → primaryAgentPoolProfiles
//   - agentPools (additional)  → agentPools
//   - availabilityZones        → int[] (not string[])
// ---------------------------------------------------------------------------
module aks 'br/public:avm/res/container-service/managed-cluster:0.13.0' = {
  name: 'aks-cluster'
  params: {
    name: aksName
    location: location
    tags: tags

    // Identity — control-plane UAMI.
    managedIdentities: {
      userAssignedResourceIds: [controlPlaneUami.id]
    }

    // OIDC issuer (required for Workload Identity federation).
    enableOidcIssuerProfile: true

    // Workload Identity via securityProfile.
    securityProfile: {
      workloadIdentity: {
        enabled: true
      }
    }

    // Disable local accounts; use Entra RBAC only.
    disableLocalAccounts: true
    enableRBAC: true

    // Auto-upgrade: keep cluster on stable channel, nodes on latest OS image.
    autoUpgradeProfile: {
      upgradeChannel: 'stable'
      nodeOSUpgradeChannel: 'NodeImage'
    }

    // Disk CSI driver (required for Drasi persistent volumes).
    enableStorageProfileDiskCSIDriver: true

    // Networking — Azure CNI Overlay + Cilium dataplane + Cilium network policy.
    networkPlugin: 'azure'
    networkPluginMode: 'overlay'
    networkDataplane: 'cilium'
    networkPolicy: 'cilium'
    podCidr: '192.168.0.0/16'
    serviceCidr: '172.16.0.0/16'
    dnsServiceIP: '172.16.0.10'
    outboundType: 'userAssignedNATGateway'

    // OMS / Log Analytics integration.
    omsAgentEnabled: true
    monitoringWorkspaceResourceId: logAnalytics.outputs.resourceId

    // System node pool — only critical addons, 3 nodes, zones 1/2/3.
    // availabilityZones is int[] in AVM 0.13.0.
    primaryAgentPoolProfiles: [
      {
        name: 'system'
        mode: 'System'
        count: 3
        vmSize: 'Standard_D4s_v5'
        osType: 'Linux'
        osSKU: 'AzureLinux'
        availabilityZones: [1, 2, 3]
        nodeTaints: ['CriticalAddonsOnly=true:NoSchedule']
        vnetSubnetResourceId: '${vnet.outputs.resourceId}/subnets/snet-aks'
        enableAutoScaling: false
      }
    ]

    // User/workload node pool — 2 nodes, zones 1/2/3.
    agentPools: [
      {
        name: 'workload'
        mode: 'User'
        count: 2
        vmSize: 'Standard_D4s_v5'
        osType: 'Linux'
        osSKU: 'AzureLinux'
        availabilityZones: [1, 2, 3]
        vnetSubnetResourceId: '${vnet.outputs.resourceId}/subnets/snet-aks'
        enableAutoScaling: false
      }
    ]
  }
}

// ---------------------------------------------------------------------------
// 11. Monitoring Metrics Publisher — workload UAMI on resource group scope
// ---------------------------------------------------------------------------
resource monitoringRoleAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(resourceGroup().id, workloadUami.id, monitoringMetricsPublisherRoleId)
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', monitoringMetricsPublisherRoleId)
    principalId: workloadUami.properties.principalId
    principalType: 'ServicePrincipal'
  }
}

// ---------------------------------------------------------------------------
// 12. Federated credential — binds workload UAMI to the Drasi resource-provider
// service account so the pod can acquire Azure tokens without a client secret.
// Subject: system:serviceaccount:drasi-system:drasi-resource-provider
// BCP321 fix: aks.outputs.oidcIssuerUrl is nullable — use ! (non-null assertion)
// since OIDC is explicitly enabled above and the value will always be present.
// ---------------------------------------------------------------------------
resource federatedCredential 'Microsoft.ManagedIdentity/userAssignedIdentities/federatedIdentityCredentials@2023-01-31' = {
  name: federatedCredName
  parent: workloadUami
  properties: {
    audiences: ['api://AzureADTokenExchange']
    issuer: aks.outputs.oidcIssuerUrl!
    subject: 'system:serviceaccount:drasi-system:drasi-resource-provider'
  }
}

// ---------------------------------------------------------------------------
// Outputs — consumed by main.bicep and written to azd env state.
// ---------------------------------------------------------------------------
@description('Name of the AKS cluster.')
output aksClusterName string = aks.outputs.name

@description('Key Vault URI for secret references.')
output keyVaultUri string = keyVault.outputs.uri

@description('Resource ID of the Log Analytics workspace.')
output logAnalyticsWorkspaceId string = logAnalytics.outputs.resourceId

@description('OIDC issuer URL for federated identity configuration.')
output oidcIssuerUrl string = aks.outputs.oidcIssuerUrl!

@description('Client ID of the workload UAMI (used by drasi-resource-provider pod annotation).')
output kubeletClientId string = workloadUami.properties.clientId
