---
name: private-networking
description: >-
  Azure private networking patterns including private endpoints, VNet integration, NSGs, private DNS zones, and network security perimeter configuration.
  USE FOR: configuring private endpoints, setting up VNet integration, designing NSG rules, managing private DNS zones, or implementing network security perimeters for Azure services.
---

# Private Networking

> **MUST:** All production Azure resources must be network-restricted. Use Private Endpoints for PaaS services and Network Security Perimeter (NSP) where supported for simplified perimeter-based access control.

## Description

Patterns for securing Azure resources with private networking — Private Endpoints, VNet integration, Network Security Groups, Network Security Perimeter (NSP), and DNS configuration.

## Capabilities

| Capability                 | Details                                        |
| -------------------------- | ---------------------------------------------- |
| Private Endpoints          | Private IP for PaaS services within VNet       |
| VNet Integration           | Outbound connectivity from compute to VNet     |
| Network Security Groups    | Layer 4 traffic filtering on subnets/NICs      |
| Network Security Perimeter | Perimeter-based access for supported PaaS      |
| Private DNS Zones          | Name resolution for private endpoints          |
| Service Endpoints          | Optimised routing to PaaS (legacy — prefer PE) |

## Standards

| Standard                                                              | Purpose                |
| --------------------------------------------------------------------- | ---------------------- |
| [Network Architecture](standards/network-architecture.md)             | VNet and subnet design |
| [Network Security Perimeter](standards/network-security-perimeter.md) | NSP patterns           |
| [Checklist](standards/checklist.md)                                   | Validation checklist   |

## Actions

| Action                                                          | Purpose          |
| --------------------------------------------------------------- | ---------------- |
| [Configure Private Networking](actions/configure-networking.md) | End-to-end setup |

---

## Private Endpoints

### When to Use

Private Endpoints provide a private IP address for Azure PaaS services within your VNet. Use for:

- Storage Accounts, Key Vault, Cosmos DB, SQL Database, Service Bus, Event Hubs
- App Configuration, Container Registry, Azure Monitor
- Any service that supports `privateEndpointConnections`

### Bicep — Private Endpoint with DNS

```bicep
param location string = resourceGroup().location
param vnetName string
param subnetName string = 'private-endpoints'
param storageAccountName string

resource vnet 'Microsoft.Network/virtualNetworks@2025-05-01' existing = {
  name: vnetName
}

resource peSubnet 'Microsoft.Network/virtualNetworks/subnets@2025-05-01' existing = {
  parent: vnet
  name: subnetName
}

resource storageAccount 'Microsoft.Storage/storageAccounts@2025-08-01' existing = {
  name: storageAccountName
}

resource privateEndpoint 'Microsoft.Network/privateEndpoints@2025-05-01' = {
  name: '${storageAccountName}-pe'
  location: location
  properties: {
    subnet: {
      id: peSubnet.id
    }
    privateLinkServiceConnections: [
      {
        name: '${storageAccountName}-plsc'
        properties: {
          privateLinkServiceId: storageAccount.id
          groupIds: [ 'blob' ]
        }
      }
    ]
  }
}

resource privateDnsZone 'Microsoft.Network/privateDnsZones@2024-06-01' = {
  name: 'privatelink.blob.core.windows.net'
  location: 'global'
}

resource dnsZoneLink 'Microsoft.Network/privateDnsZones/virtualNetworkLinks@2024-06-01' = {
  parent: privateDnsZone
  name: '${vnetName}-link'
  location: 'global'
  properties: {
    registrationEnabled: false
    virtualNetwork: {
      id: vnet.id
    }
  }
}

resource dnsZoneGroup 'Microsoft.Network/privateEndpoints/privateDnsZoneGroups@2025-05-01' = {
  parent: privateEndpoint
  name: 'default'
  properties: {
    privateDnsZoneConfigs: [
      {
        name: 'blob'
        properties: {
          privateDnsZoneId: privateDnsZone.id
        }
      }
    ]
  }
}
```

### Common Private DNS Zone Names

| Service            | DNS Zone                             | Group ID              |
| ------------------ | ------------------------------------ | --------------------- |
| Blob Storage       | `privatelink.blob.core.windows.net`  | `blob`                |
| Key Vault          | `privatelink.vaultcore.azure.net`    | `vault`               |
| Cosmos DB          | `privatelink.documents.azure.com`    | `Sql`                 |
| SQL Database       | `privatelink.database.windows.net`   | `sqlServer`           |
| Service Bus        | `privatelink.servicebus.windows.net` | `namespace`           |
| Event Hubs         | `privatelink.servicebus.windows.net` | `namespace`           |
| App Configuration  | `privatelink.azconfig.io`            | `configurationStores` |
| Container Registry | `privatelink.azurecr.io`             | `registry`            |

---

## VNet Integration

### App Service / Functions VNet Integration

```bicep
resource vnetIntegration 'Microsoft.Web/sites/networkConfig@2024-11-01' = {
  parent: webApp
  name: 'virtualNetwork'
  properties: {
    subnetResourceId: integrationSubnet.id
    swiftSupported: true
  }
}

> **Route-all behaviour:** By default, App Service VNet integration only routes
> RFC 1918 traffic (10.x, 172.16.x, 192.168.x) through the VNet. To route **all**
> outbound traffic (including public internet) through the VNet (e.g., for firewall
> inspection), set the `vnetRouteAllEnabled` site property to `true` or
> `WEBSITE_VNET_ROUTE_ALL=1` app setting.
```

**Subnet requirements:**

- Delegated to `Microsoft.Web/serverFarms`
- Minimum `/27` for new App Service integration subnets (`/28` for some existing subnets); `/26` recommended for scale
- Separate from private endpoint subnets

### Container Apps VNet Integration

> **Warning — internal environments and custom domains:** Setting `internal: true` on a
> Container Apps environment restricts all ingress to the VNet. Custom domains with
> public DNS will **not** route to internal environments. You need Azure Application
> Gateway, Azure Front Door, or a reverse proxy in front of the internal environment
> to expose apps with custom domains to the internet.

```bicep
resource containerAppEnv 'Microsoft.App/managedEnvironments@2025-07-01' = {
  name: envName
  location: location
  properties: {
    vnetConfiguration: {
      infrastructureSubnetId: infraSubnet.id
      internal: true
    }
  }
}
```

**Subnet requirements:**

- Workload profiles environments support smaller minimums (commonly `/27`); consumption-only environments may require larger ranges such as `/23`
- Delegated to `Microsoft.App/environments`

---

## Network Security Groups

```bicep
resource nsg 'Microsoft.Network/networkSecurityGroups@2025-05-01' = {
  name: '${subnetName}-nsg'
  location: location
  properties: {
    securityRules: [
      {
        name: 'AllowHttpsInbound'
        properties: {
          priority: 100
          direction: 'Inbound'
          access: 'Allow'
          protocol: 'Tcp'
          sourceAddressPrefix: 'Internet'
          sourcePortRange: '*'
          destinationAddressPrefix: '*'
          destinationPortRange: '443'
        }
      }
      {
        name: 'DenyAllInbound'
        properties: {
          priority: 4096
          direction: 'Inbound'
          access: 'Deny'
          protocol: '*'
          sourceAddressPrefix: '*'
          sourcePortRange: '*'
          destinationAddressPrefix: '*'
          destinationPortRange: '*'
        }
      }
    ]
  }
}
```

---

## NAT Gateway (Outbound-Only Scenarios)

Use a NAT Gateway when VNet-integrated compute (Container Apps, App Service, Functions) needs a **static outbound IP** for allowlisting by third-party APIs or firewalls.

```bicep
resource publicIp 'Microsoft.Network/publicIPAddresses@2025-05-01' = {
  name: 'pip-nat-${workloadName}'
  location: location
  sku: { name: 'Standard' }
  properties: {
    publicIPAllocationMethod: 'Static'
  }
}

resource natGateway 'Microsoft.Network/natGateways@2025-05-01' = {
  name: 'ng-${workloadName}'
  location: location
  sku: { name: 'Standard' }
  properties: {
    idleTimeoutInMinutes: 4
    publicIpAddresses: [
      { id: publicIp.id }
    ]
  }
}

// Associate with the subnet used by your compute
resource subnet 'Microsoft.Network/virtualNetworks/subnets@2025-05-01' = {
  parent: vnet
  name: 'snet-apps'
  properties: {
    addressPrefix: '10.0.1.0/24'
    natGateway: { id: natGateway.id }
    // ... delegation, NSG, etc.
  }
}
```

**When to use NAT Gateway vs default outbound:**

| Scenario                                           | Use NAT Gateway?  |
| -------------------------------------------------- | ----------------- |
| Third-party API requires IP allowlisting           | Yes               |
| General outbound internet (no IP requirement)      | No — default SNAT |
| Multiple subnets sharing one outbound IP           | Yes               |
| High-volume outbound connections (SNAT exhaustion) | Yes               |

> **Warning — default outbound access retirement:** Azure is retiring default outbound
> access for new VMs and VMSS created after 30 September 2025. Existing resources are
> unaffected, but new deployments should explicitly configure NAT Gateway, Load Balancer
> outbound rules, or instance-level public IPs.

---

## Principles

1. **Private Endpoints over service endpoints** — they provide stronger isolation and work across VNet peering.
2. **Disable public access** on PaaS services once private endpoints are configured.
3. **Use Network Security Perimeter** where supported — simplifies perimeter-based access control for PaaS-to-PaaS communication.
4. **Centralise Private DNS Zones** — share zones across VNets via links, don't duplicate.
5. **Least-privilege NSG rules** — deny all by default, allow only required traffic.
6. **Plan subnet sizing** — account for delegation requirements and growth.
7. **Use NAT Gateway for static outbound IPs** — avoid relying on default outbound access for IP-dependent integrations.

## References

- [Private Endpoint documentation](https://learn.microsoft.com/en-us/azure/private-link/private-endpoint-overview)
- [VNet integration for App Service](https://learn.microsoft.com/en-us/azure/app-service/overview-vnet-integration)
- [Network Security Perimeter](https://learn.microsoft.com/en-us/azure/private-link/network-security-perimeter-concepts)
- [Private DNS Zones](https://learn.microsoft.com/en-us/azure/private-link/private-endpoint-dns)
- [NSG documentation](https://learn.microsoft.com/en-us/azure/virtual-network/network-security-groups-overview)

---

## Currency and Verification

- **Date checked:** 2026-03-31 (verified via Microsoft Learn MCP — ARM template references)
- **Compatibility:** Azure Bicep, ARM templates
- **Sources:**
  - [Microsoft.Network ARM reference](https://learn.microsoft.com/azure/templates/microsoft.network/virtualnetworks)
  - [Microsoft.Storage ARM reference](https://learn.microsoft.com/azure/templates/microsoft.storage/storageaccounts)
- **Verification steps:**
  1. Run `az provider show --namespace Microsoft.Network --query "resourceTypes[?resourceType=='virtualNetworks'].apiVersions" -o tsv` and confirm `2025-05-01` is listed
  2. Run `az provider show --namespace Microsoft.Storage --query "resourceTypes[?resourceType=='storageAccounts'].apiVersions" -o tsv` and confirm `2025-08-01` is listed
  3. Run `az bicep build --file <your-bicep-file>` to validate syntax

### API versions used in this file

| Resource type                                             | API version  | Status  |
| --------------------------------------------------------- | ------------ | ------- |
| `Microsoft.Network/virtualNetworks`                       | `2025-05-01` | Current |
| `Microsoft.Network/privateEndpoints`                      | `2025-05-01` | Current |
| `Microsoft.Network/privateDnsZones`                       | `2024-06-01` | Current |
| `Microsoft.Network/privateDnsZones/virtualNetworkLinks`   | `2024-06-01` | Current |
| `Microsoft.Network/privateEndpoints/privateDnsZoneGroups` | `2025-05-01` | Current |
| `Microsoft.Network/networkSecurityGroups`                 | `2025-05-01` | Current |
| `Microsoft.Network/publicIPAddresses`                     | `2025-05-01` | Current |
| `Microsoft.Network/natGateways`                           | `2025-05-01` | Current |
| `Microsoft.Storage/storageAccounts`                       | `2025-08-01` | Current |
| `Microsoft.App/managedEnvironments`                       | `2025-07-01` | Current |
| `Microsoft.Web/sites/networkConfig`                       | `2024-11-01` | Current |

## Known Pitfalls

| Pitfall                                                   | Symptom                                                           | Fix                                                                                                                                            |
| --------------------------------------------------------- | ----------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| Private endpoint subnet overlaps with delegation          | Deployment fails with subnet conflict                             | Use separate subnets for private endpoints and VNet integration (delegation)                                                                   |
| Missing private DNS zone link                             | Private endpoint IP resolves but hostname returns public IP       | Link the private DNS zone to the VNet; verify with `nslookup` from within VNet                                                                 |
| Container Apps `internal: true` with public custom domain | Custom domain returns connection refused or timeout               | Place Azure Application Gateway or Front Door in front of internal environment                                                                 |
| Default outbound access retirement (after Sep 2025)       | New VMs/VMSS have no outbound connectivity                        | Explicitly configure NAT Gateway, Load Balancer outbound rules, or instance-level public IP                                                    |
| VNet integration only routes RFC 1918 by default          | Outbound calls to public IPs bypass VNet (no firewall inspection) | Set `vnetRouteAllEnabled: true` or `WEBSITE_VNET_ROUTE_ALL=1` on the App Service                                                               |
| Container Apps subnet too small                           | Environment provisioning fails                                    | Size Container Apps subnets by environment type (`/27` or larger for workload profiles; `/23` commonly used for consumption-only environments) |
| App Service subnet too small                              | VNet integration fails to allocate addresses                      | Use `/27` minimum (`/28` only in specific existing-subnet scenarios); prefer `/26` for headroom                                                |

## Related Skills

- [Identity & Managed Identity](../identity-managed-identity/SKILL.md) — RBAC complements network restrictions
- [Azure Container Apps](../azure-container-apps/SKILL.md) — VNet integration for Container Apps
- [Azure Functions Patterns](../azure-functions-patterns/SKILL.md) — VNet integration for Functions
