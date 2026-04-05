# Network Architecture

## VNet Design

### Subnet Layout

| Subnet               | CIDR  | Delegation                   | Purpose                        |
| -------------------- | ----- | ---------------------------- | ------------------------------ |
| `app-integration`    | `/27` (or `/26` recommended) | `Microsoft.Web/serverFarms`  | App Service/Functions outbound |
| `container-apps`     | `/27`+ (workload profiles) / `/23` (consumption-only) | `Microsoft.App/environments` | Container Apps infrastructure  |
| `private-endpoints`  | `/24` | None                         | Private Endpoints for PaaS     |
| `default`            | `/24` | None                         | VMs, other compute             |
| `AzureBastionSubnet` | `/26` | None (name must be exact)    | Azure Bastion access           |

### Bicep — VNet with Subnets

```bicep
param location string = resourceGroup().location
param vnetName string
param vnetAddressPrefix string = '10.0.0.0/16'

resource nsgPrivateEndpoints 'Microsoft.Network/networkSecurityGroups@2025-05-01' = {
  name: '${vnetName}-pe-nsg'
  location: location
  properties: {
    securityRules: []
  }
}

resource nsgAppIntegration 'Microsoft.Network/networkSecurityGroups@2025-05-01' = {
  name: '${vnetName}-app-nsg'
  location: location
  properties: {
    securityRules: []
  }
}

resource vnet 'Microsoft.Network/virtualNetworks@2025-05-01' = {
  name: vnetName
  location: location
  properties: {
    addressSpace: {
      addressPrefixes: [ vnetAddressPrefix ]
    }
    subnets: [
      {
        name: 'private-endpoints'
        properties: {
          addressPrefix: '10.0.1.0/24'
          networkSecurityGroup: { id: nsgPrivateEndpoints.id }
          privateEndpointNetworkPolicies: 'Enabled'
        }
      }
      {
        name: 'app-integration'
        properties: {
          addressPrefix: '10.0.2.0/27'
          networkSecurityGroup: { id: nsgAppIntegration.id }
          delegations: [
            {
              name: 'webapp'
              properties: {
                serviceName: 'Microsoft.Web/serverFarms'
              }
            }
          ]
        }
      }
      {
        name: 'container-apps'
        properties: {
          addressPrefix: '10.0.4.0/27'
          delegations: [
            {
              name: 'containerapp'
              properties: {
                serviceName: 'Microsoft.App/environments'
              }
            }
          ]
        }
      }
    ]
  }
}
```

---

## Disabling Public Access

After configuring private endpoints, disable public network access on PaaS resources:

| Resource           | Property                          |
| ------------------ | --------------------------------- |
| Storage Account    | `publicNetworkAccess: 'Disabled'` |
| Key Vault          | `publicNetworkAccess: 'disabled'` |
| Cosmos DB          | `publicNetworkAccess: 'Disabled'` |
| SQL Database       | `publicNetworkAccess: 'Disabled'` |
| Service Bus        | `publicNetworkAccess: 'Disabled'` |
| App Configuration  | `publicNetworkAccess: 'Disabled'` |
| Container Registry | `publicNetworkAccess: 'Disabled'` |

```bicep
resource storageAccount 'Microsoft.Storage/storageAccounts@2025-08-01' = {
  name: storageName
  location: location
  properties: {
    publicNetworkAccess: 'Disabled'
    networkAcls: {
      defaultAction: 'Deny'
    }
  }
}
```

---

## DNS Resolution Flow

```
Application → Azure DNS (168.63.129.16)
  → Private DNS Zone (privatelink.blob.core.windows.net)
    → Private IP (10.0.1.5) → Private Endpoint → Storage Account
```

When DNS zones are linked to the VNet, Azure DNS automatically resolves private endpoint FQDNs to their private IP addresses.

---

## Rules

1. Use a dedicated subnet for private endpoints — don't mix with compute resources.
2. Disable public access on PaaS resources after private endpoints are configured.
3. Associate NSGs with every subnet, even if rules are initially empty.
4. Plan IP address space for growth — it's difficult to resize subnets later.
5. Use hub-spoke topology for multi-VNet architectures.
