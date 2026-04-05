# Network Security Perimeter

## Overview

Azure Network Security Perimeter (NSP) provides perimeter-based access control for Azure PaaS services. Instead of configuring private endpoints and firewall rules on each individual resource, NSP defines a security boundary that controls inbound and outbound access for all associated resources.

**When to use NSP:**

- Multiple PaaS resources that need to communicate with each other
- Simplified network security management without individual private endpoint configuration
- Centralised access control for PaaS-to-PaaS communication

**When to use Private Endpoints instead:**

- Connecting compute resources (VMs, App Service, Container Apps) to PaaS
- Cross-region or cross-VNet connectivity
- Resources not yet supported by NSP

---

## Supported Services

| Service                 | Resource Type                                    |
| ----------------------- | ------------------------------------------------ |
| Azure Storage           | `Microsoft.Storage/storageAccounts`              |
| Azure Key Vault         | `Microsoft.KeyVault/vaults`                      |
| Azure Cosmos DB         | `Microsoft.DocumentDB/databaseAccounts`          |
| Azure SQL Database      | `Microsoft.Sql/servers`                          |
| Azure Event Hubs        | `Microsoft.EventHub/namespaces`                  |
| Azure Service Bus       | `Microsoft.ServiceBus/namespaces`                |
| Azure App Configuration | `Microsoft.AppConfiguration/configurationStores` |
| Azure Monitor           | `Microsoft.Insights/components`                  |
| Azure AI Services       | `Microsoft.CognitiveServices/accounts`           |

> Check [latest NSP supported services](https://learn.microsoft.com/en-us/azure/private-link/network-security-perimeter-concepts#onboarded-private-link-resources) as more services are continuously added.

---

## Bicep — Network Security Perimeter

```bicep
param location string = resourceGroup().location
param perimeterName string

resource nsp 'Microsoft.Network/networkSecurityPerimeters@2025-05-01' = {
  name: perimeterName
  location: location
}

// Profile defines the access rules
resource nspProfile 'Microsoft.Network/networkSecurityPerimeters/profiles@2025-05-01' = {
  parent: nsp
  name: 'default'
  location: location
}

// Inbound access rule — allow specific public IP ranges
resource inboundRule 'Microsoft.Network/networkSecurityPerimeters/profiles/accessRules@2025-05-01' = {
  parent: nspProfile
  name: 'allow-corporate-ips'
  location: location
  properties: {
    direction: 'Inbound'
    addressPrefixes: [
      '203.0.113.0/24' // Corporate IP range
    ]
  }
}

// Outbound access rule — allow FQDN-based access
resource outboundRule 'Microsoft.Network/networkSecurityPerimeters/profiles/accessRules@2025-05-01' = {
  parent: nspProfile
  name: 'allow-external-api'
  location: location
  properties: {
    direction: 'Outbound'
    fullyQualifiedDomainNames: [
      'api.example.com'
    ]
  }
}
```

---

## Resource Association

Associate PaaS resources with the NSP:

```bicep
// Associate a Storage Account with the NSP
resource nspAssociation 'Microsoft.Network/networkSecurityPerimeters/resourceAssociations@2025-05-01' = {
  parent: nsp
  name: '${storageAccount.name}-association'
  location: location
  properties: {
    privateLinkResource: {
      id: storageAccount.id
    }
    profile: {
      id: nspProfile.id
    }
    accessMode: 'Enforced'
  }
}
```

### Access Modes

| Mode       | Behaviour                                       |
| ---------- | ----------------------------------------------- |
| `Enforced` | Only traffic matching NSP rules is allowed      |
| `Learning` | All traffic allowed, but violations are logged  |
| `Audit`    | Existing rules apply, NSP violations are logged |

**Recommended approach:** Start with `Learning` mode to identify traffic patterns, then switch to `Enforced`.

---

## NSP with Private Endpoints

NSP and Private Endpoints are complementary:

- **NSP** controls PaaS-to-PaaS communication within the perimeter boundary
- **Private Endpoints** provide private IP connectivity from VNet-based compute

```
[VNet Compute] → Private Endpoint → [Storage Account]
                                          ↕ NSP perimeter
                                     [Key Vault] (same NSP)
                                          ↕ NSP perimeter
                                     [Cosmos DB] (same NSP)
```

Resources within the same NSP can communicate freely without additional configuration.

---

## Diagnostic Logging

```bicep
resource nspDiagnostics 'Microsoft.Insights/diagnosticSettings@2021-05-01-preview' = {
  name: '${perimeterName}-diag'
  scope: nsp
  properties: {
    workspaceId: logAnalyticsWorkspace.id
    logs: [
      {
        category: 'NspAccessLogs'
        enabled: true
        retentionPolicy: { enabled: true, days: 90 }
      }
    ]
  }
}
```

---

## Rules

1. Start with Learning mode — understand traffic patterns before enforcing.
2. Associate all supported PaaS resources in a workload to the same NSP.
3. Use NSP for PaaS-to-PaaS, Private Endpoints for compute-to-PaaS.
4. Enable diagnostic logging on the NSP to monitor access patterns.
5. Review NSP access logs regularly for unexpected traffic.
6. Keep NSP access rules minimal — allow only required inbound/outbound traffic.
