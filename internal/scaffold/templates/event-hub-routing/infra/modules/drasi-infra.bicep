// drasi-infra.bicep — Core Azure infrastructure for Drasi workloads.
// Provisions: VNet, NAT Gateway, Public IP, AKS (Azure CNI Overlay + Cilium, OIDC + Workload
// Identity), Key Vault, Log Analytics workspace, User-Assigned Managed Identities (control-plane
// and workload), role assignments, and federated identity credential.
//
// All resources use Azure Verified Modules (AVM) public registry modules.
// AKS configuration follows the AKS skill recommendations for production clusters.

@description('The Azure region for all resources.')
param location string

@description('Environment name suffix applied to all resource names.')
param environmentName string

@description('Tags to apply to all resources.')
param tags object

@description('Object ID of the deploying user. Used to assign AKS RBAC Cluster Admin so the user can run kubectl/drasi commands after provisioning.')
param principalId string

@description('Kubernetes namespace where Drasi is installed.')
param drasiNamespace string = 'drasi-system'

@description('Service account name used by the Drasi resource provider.')
param drasiServiceAccountName string = 'drasi-resource-provider'

// ---------------------------------------------------------------------------
// Names — deterministic suffix so redeploys are idempotent
// ---------------------------------------------------------------------------
var suffix = uniqueString(resourceGroup().id, environmentName)
var aksName        = 'drasi-aks-${suffix}'
var kvName         = take('drasi-kv-${suffix}', 24) // Key Vault name max 24 chars
var lawName        = 'drasi-law-${suffix}'
// Drasi workload UAMI — used for Workload Identity federated credential
var uamiName       = 'drasi-id-${suffix}'
// Control-plane UAMI — used as the AKS cluster identity (Network Contributor on VNet)
var cpUamiName     = 'drasi-cp-${suffix}'
var vnetName       = 'drasi-vnet-${suffix}'
var natGwName      = 'drasi-natgw-${suffix}'
var pipName        = 'drasi-pip-${suffix}'
// Subnet address spaces
var vnetPrefix     = '10.0.0.0/16'
var aksSubnetPrefix = '10.0.0.0/22' // /22 = 1022 usable IPs; sufficient for Azure CNI Overlay pod CIDRs

// Role definition GUIDs (built-in, verified against Azure RBAC docs)
// https://learn.microsoft.com/azure/role-based-access-control/built-in-roles
var kvSecretsUserRoleId        = '4633458b-17de-408a-b874-0445c86b69e6' // Key Vault Secrets User
var kvSecretsOfficerRoleId     = 'b86a8fe4-44ce-4948-aee5-eccb2c155cd7' // Key Vault Secrets Officer
var monitoringPublisherRoleId  = '3913510d-42f4-4e42-8a64-420c390055eb' // Monitoring Metrics Publisher
var networkContributorRoleId   = '4d97b98b-1d4f-4787-a291-c67834d212e7' // Network Contributor
var aksRbacClusterAdminRoleId  = 'b1ff04bb-8a4e-4dc4-8eb5-8693973ce19b' // Azure Kubernetes Service RBAC Cluster Admin

// ---------------------------------------------------------------------------
// Log Analytics Workspace — AVM module
// https://github.com/Azure/bicep-registry-modules/tree/main/avm/res/operational-insights/workspace
// ---------------------------------------------------------------------------
module logAnalytics 'br/public:avm/res/operational-insights/workspace:0.15.0' = {
  name: 'law-deployment'
  params: {
    name: lawName
    location: location
    tags: union(tags, { component: 'observability', 'managed-by': 'azd' })
    skuName: 'PerGB2018'
    dataRetention: 30
  }
}

// ---------------------------------------------------------------------------
// Drasi workload User-Assigned Managed Identity — AVM module
// Used for Workload Identity federated credential; grants pod access to Key Vault.
// https://github.com/Azure/bicep-registry-modules/tree/main/avm/res/managed-identity/user-assigned-identity
// ---------------------------------------------------------------------------
module uami 'br/public:avm/res/managed-identity/user-assigned-identity:0.5.0' = {
  name: 'uami-deployment'
  params: {
    name: uamiName
    location: location
    tags: union(tags, { component: 'identity', 'managed-by': 'azd' })
  }
}

// ---------------------------------------------------------------------------
// Control-plane User-Assigned Managed Identity — AVM module
// Used as the AKS cluster identity per AKS skill: pre-create so Network Contributor
// can be assigned before the cluster is provisioned.
// ---------------------------------------------------------------------------
module cpUami 'br/public:avm/res/managed-identity/user-assigned-identity:0.5.0' = {
  name: 'cp-uami-deployment'
  params: {
    name: cpUamiName
    location: location
    tags: union(tags, { component: 'aks-control-plane-identity', 'managed-by': 'azd' })
  }
}

// ---------------------------------------------------------------------------
// Key Vault (RBAC-authorised, purge-protected) — AVM module
// https://github.com/Azure/bicep-registry-modules/tree/main/avm/res/key-vault/vault
// ---------------------------------------------------------------------------
module keyVault 'br/public:avm/res/key-vault/vault:0.13.3' = {
  name: 'kv-deployment'
  params: {
    name: kvName
    location: location
    tags: union(tags, { component: 'secrets', 'managed-by': 'azd' })
    enableRbacAuthorization: true
    enableSoftDelete: true
    softDeleteRetentionInDays: 7
    enablePurgeProtection: true
    // Grant the Drasi workload UAMI Key Vault Secrets User so pods can read secrets.
    roleAssignments: [
      {
        principalId: uami.outputs.principalId
        roleDefinitionIdOrName: kvSecretsUserRoleId
        principalType: 'ServicePrincipal'
      }
      // Grant the deploying user Key Vault Secrets Officer so the provision step can
      // store secrets and the deploy step can read them via `az keyvault secret show`.
      {
        principalId: principalId
        roleDefinitionIdOrName: kvSecretsOfficerRoleId
        principalType: 'User'
      }
    ]
  }
}

// ---------------------------------------------------------------------------
// Public IP Address — AVM module (used by NAT Gateway)
// Standard SKU + Static required for NAT Gateway compatibility.
// Zone-redundant across all three zones.
// https://github.com/Azure/bicep-registry-modules/tree/main/avm/res/network/public-ip-address
// ---------------------------------------------------------------------------
module publicIp 'br/public:avm/res/network/public-ip-address:0.9.1' = {
  name: 'pip-deployment'
  params: {
    name: pipName
    location: location
    tags: union(tags, { component: 'network', 'managed-by': 'azd' })
    skuName: 'Standard'
    skuTier: 'Regional'
    publicIPAllocationMethod: 'Static'
    publicIPAddressVersion: 'IPv4'
    // Development: no zone pinning. For production, add zones: [1, 2, 3].
  }
}

// ---------------------------------------------------------------------------
// NAT Gateway — AVM module
// AKS skill: outbound via NAT Gateway to prevent SNAT port exhaustion.
// Zone-redundant deployment (no zone pinning) — matches multi-zone AKS pools.
// https://github.com/Azure/bicep-registry-modules/tree/main/avm/res/network/nat-gateway
// ---------------------------------------------------------------------------
module natGateway 'br/public:avm/res/network/nat-gateway:2.0.1' = {
  name: 'nat-gw-deployment'
  params: {
    name: natGwName
    location: location
    tags: union(tags, { component: 'network', 'managed-by': 'azd' })
    // Zone-redundant: -1 means no zone pinning, matching multi-zone AKS pools
    availabilityZone: -1
    publicIpResourceIds: [
      publicIp.outputs.resourceId
    ]
    idleTimeoutInMinutes: 10
  }
}

// ---------------------------------------------------------------------------
// Virtual Network — AVM module
// AKS skill: dedicated VNet for Azure CNI Overlay networking.
// NAT Gateway is associated at the subnet level.
// https://github.com/Azure/bicep-registry-modules/tree/main/avm/res/network/virtual-network
// ---------------------------------------------------------------------------
module vnet 'br/public:avm/res/network/virtual-network:0.8.0' = {
  name: 'vnet-deployment'
  params: {
    name: vnetName
    location: location
    tags: union(tags, { component: 'network', 'managed-by': 'azd' })
    addressPrefixes: [
      vnetPrefix
    ]
    subnets: [
      {
        name: 'snet-aks'
        addressPrefix: aksSubnetPrefix
        // Associate NAT Gateway for outbound traffic (AKS skill requirement)
        natGatewayResourceId: natGateway.outputs.resourceId
      }
    ]
  }
}

// ---------------------------------------------------------------------------
// Network Contributor on VNet for the AKS control-plane UAMI
// AKS skill: control-plane identity needs Network Contributor to manage VNet
// resources (NICs, load balancers) during cluster operations.
// Using a direct roleAssignment avoids re-deploying the VNet module (which causes
// a DeploymentActive conflict on the subnet sub-deployment).
// guid() inputs use known var names so the name is calculable at deployment start.
// ---------------------------------------------------------------------------
resource vnetRef 'Microsoft.Network/virtualNetworks@2024-10-01' existing = {
  name: vnetName
  dependsOn: [vnet]
}

resource vnetNetworkContributorRole 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(resourceGroup().id, cpUamiName, networkContributorRoleId)
  scope: vnetRef
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', networkContributorRoleId)
    principalId: cpUami.outputs.principalId
    principalType: 'ServicePrincipal'
  }
}

// ---------------------------------------------------------------------------
// AKS Managed Cluster — AVM module
// AKS skill compliance:
//   - Azure CNI Overlay + Cilium network plugin (replaces retiring Kubenet)
//   - User-assigned control-plane identity (pre-created, Network Contributor on VNet)
//   - System node pool: 3 nodes, Standard_D4s_v5, AzureLinux 3.0,
//     CriticalAddonsOnly taint, availability zones [1,2,3]
//   - disableLocalAccounts: true (production hardening)
//   - autoUpgradeProfile: stable + nodeOSUpgradeChannel: NodeImage
//   - OIDC issuer + Workload Identity for Drasi federated credential
// https://github.com/Azure/bicep-registry-modules/tree/main/avm/res/container-service/managed-cluster
// ---------------------------------------------------------------------------
module aks 'br/public:avm/res/container-service/managed-cluster:0.13.0' = {
  name: 'aks-deployment'
  params: {
    name: aksName
    location: location
    tags: union(tags, { component: 'compute', 'managed-by': 'azd' })

    // AKS skill: user-assigned identity for control plane (not system-assigned)
    managedIdentities: {
      userAssignedResourceIds: [
        cpUami.outputs.resourceId
      ]
    }

    // OIDC issuer — required for Drasi federated credential on workload UAMI
    enableOidcIssuerProfile: true

    // Kubernetes RBAC — explicit for clarity
    enableRBAC: true

    // AKS skill: Azure CNI Overlay + Cilium (replaces Kubenet, non-reversible decision)
    networkPlugin: 'azure'
    networkPluginMode: 'overlay'
    networkDataplane: 'cilium'
    networkPolicy: 'cilium'

    // Explicit outbound type — required when using BYO NAT Gateway
    outboundType: 'userAssignedNATGateway'

    // Non-overlapping CIDRs — must not overlap with VNet (10.0.0.0/16) or subnet (10.0.0.0/22)
    serviceCidr: '10.2.0.0/16'
    dnsServiceIP: '10.2.0.10'

    // AKS skill: Managed AAD integration — required for disableLocalAccounts and Azure RBAC on cluster
    // aadProfile enables Microsoft Entra ID authentication; Azure RBAC replaces Kubernetes RBAC for authz.
    aadProfile: {
      managed: true
      enableAzureRBAC: true
    }

    // AKS skill: disable local accounts in production (requires managed AAD above)
    disableLocalAccounts: true

    // AKS skill: automatic upgrade channels
    autoUpgradeProfile: {
      upgradeChannel: 'stable'
      nodeOSUpgradeChannel: 'NodeImage'
    }

    // Wire Log Analytics for Container Insights via OMS agent
    omsAgentEnabled: true
    omsAgentUseAADAuth: true
    monitoringWorkspaceResourceId: logAnalytics.outputs.resourceId

    // Workload Identity — Drasi pods exchange Kubernetes SATs for Entra ID tokens
    securityProfile: {
      workloadIdentity: {
        enabled: true
      }
    }

    // Storage profile — enable Azure Disk CSI driver so StatefulSets (redis, mongo)
    // can provision PersistentVolumeClaims using the 'default' StorageClass.
    enableStorageProfileDiskCSIDriver: true

    // System node pool — development configuration.
    //   - Single node with Standard_D2s_v5 to minimise vCPU quota usage.
    //   - Azure Linux (osSku: AzureLinux)
    //   - CriticalAddonsOnly taint keeps workloads on the user pool.
    //   - For production, increase count to 3, use D4s_v5+, and add availability zones.
    primaryAgentPoolProfiles: [
      {
        name: 'systempool'
        count: 1
        vmSize: 'Standard_D2s_v5'
        osType: 'Linux'
        osSKU: 'AzureLinux'
        mode: 'System'
        enableAutoScaling: false
        vnetSubnetResourceId: '${vnet.outputs.resourceId}/subnets/snet-aks'
        nodeTaints: [
          'CriticalAddonsOnly=true:NoSchedule'
        ]
      }
    ]

    // User node pool — Dapr, Drasi, and application workloads schedule here.
    // Development configuration: single node to minimise quota usage.
    // For production, increase count to 3, use D4s_v5+, and add availability zones.
    agentPools: [
      {
        name: 'workload'
        count: 1
        vmSize: 'Standard_D2s_v5'
        osType: 'Linux'
        osSKU: 'AzureLinux'
        mode: 'User'
        enableAutoScaling: false
        vnetSubnetResourceId: '${vnet.outputs.resourceId}/subnets/snet-aks'
        // No nodeTaints — all workloads (Dapr, Drasi) must be able to schedule here.
        nodeTaints: []
      }
    ]
  }
  dependsOn: [vnetNetworkContributorRole]
}

// ---------------------------------------------------------------------------
// Monitoring Metrics Publisher role on Log Analytics — direct roleAssignment resource
// Grants the Drasi workload UAMI publish rights for Container Insights custom metrics.
// Using a direct roleAssignment avoids re-deploying the LAW module (which causes a
// Conflict error when the workspace is still provisioning).
// ---------------------------------------------------------------------------
resource lawMetricsPublisherRole 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  // guid() inputs must be calculable at deployment start — use known names, not module outputs
  name: guid(resourceGroup().id, uamiName, monitoringPublisherRoleId)
  scope: resourceGroup()
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', monitoringPublisherRoleId)
    principalId: uami.outputs.principalId
    principalType: 'ServicePrincipal'
  }
}

// ---------------------------------------------------------------------------
// AKS RBAC Cluster Admin — deploying user on AKS cluster scope
// Required because the cluster uses disableLocalAccounts + enableAzureRBAC.
// Without this, the deploying user cannot run kubectl or drasi commands.
// ---------------------------------------------------------------------------
resource aksRef 'Microsoft.ContainerService/managedClusters@2024-09-01' existing = {
  name: aksName
  dependsOn: [aks]
}

resource aksClusterAdminAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(resourceGroup().id, principalId, aksRbacClusterAdminRoleId)
  scope: aksRef
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', aksRbacClusterAdminRoleId)
    principalId: principalId
    principalType: 'User'
  }
}

// ---------------------------------------------------------------------------
// Federated Identity Credential — allows the Drasi resource-provider pod to
// exchange a Kubernetes ServiceAccountToken for an Entra ID access token.
// AVM UAMI module does not expose federated credentials; use child resource.
// API version: 2024-11-30 (verified stable)
// ---------------------------------------------------------------------------
resource federatedCredential 'Microsoft.ManagedIdentity/userAssignedIdentities/federatedIdentityCredentials@2024-11-30' = {
  name: '${uamiName}/drasi-federation'
  properties: {
    issuer: aks.outputs.oidcIssuerUrl!
    subject: 'system:serviceaccount:${drasiNamespace}:${drasiServiceAccountName}'
    audiences: [
      'api://AzureADTokenExchange'
    ]
  }
}

// ---------------------------------------------------------------------------
// Outputs — consumed by main.bicep and mapped to azd env state
// ---------------------------------------------------------------------------
output aksClusterName string = aks.outputs.name
output aksOidcIssuerUrl string = aks.outputs.oidcIssuerUrl!
output keyVaultName string = keyVault.outputs.name
output keyVaultUri string = keyVault.outputs.uri
output logAnalyticsWorkspaceId string = logAnalytics.outputs.resourceId
output uamiClientId string = uami.outputs.clientId
output uamiPrincipalId string = uami.outputs.principalId
output uamiResourceId string = uami.outputs.resourceId
output vnetResourceId string = vnet.outputs.resourceId
