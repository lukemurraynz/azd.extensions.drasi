# Configure Private Networking

## Steps

### Step 1 — Deploy VNet and Subnets

Deploy the VNet with dedicated subnets for private endpoints and integration:

```bicep
resource vnet 'Microsoft.Network/virtualNetworks@2025-05-01' = {
  name: vnetName
  location: location
  properties: {
    addressSpace: { addressPrefixes: [ '10.0.0.0/16' ] }
    subnets: [
      {
        name: 'private-endpoints'
        properties: {
          addressPrefix: '10.0.1.0/24'
          networkSecurityGroup: { id: peNsg.id }
          privateEndpointNetworkPolicies: 'Enabled'
        }
      }
      {
        name: 'app-integration'
        properties: {
          addressPrefix: '10.0.2.0/26'
          networkSecurityGroup: { id: appNsg.id }
          delegations: [
            { name: 'webapp', properties: { serviceName: 'Microsoft.Web/serverFarms' } }
          ]
        }
      }
    ]
  }
}
```

### Step 2 — Create Private Endpoints

For each PaaS service, create a private endpoint with DNS:

```bicep
resource privateEndpoint 'Microsoft.Network/privateEndpoints@2025-05-01' = {
  name: '${resourceName}-pe'
  location: location
  properties: {
    subnet: { id: peSubnet.id }
    privateLinkServiceConnections: [
      {
        name: '${resourceName}-plsc'
        properties: {
          privateLinkServiceId: targetResource.id
          groupIds: [ groupId ] // e.g., 'blob', 'vault', 'namespace'
        }
      }
    ]
  }
}
```

Create the corresponding private DNS zone and link it to the VNet.

### Step 3 — Disable Public Access

On each PaaS resource, disable public network access:

```bicep
properties: {
  publicNetworkAccess: 'Disabled'
  networkAcls: {
    defaultAction: 'Deny'
  }
}
```

### Step 4 — Configure Network Security Perimeter (Optional)

For PaaS-to-PaaS communication, create an NSP:

1. Deploy NSP with a profile
2. Associate supported PaaS resources
3. Start in `Learning` mode
4. Review access logs
5. Switch to `Enforced` mode

### Step 5 — Configure VNet Integration

For App Service or Functions that need outbound VNet access:

```bicep
resource vnetIntegration 'Microsoft.Web/sites/networkConfig@2024-11-01' = {
  parent: webApp
  name: 'virtualNetwork'
  properties: {
    subnetResourceId: integrationSubnet.id
    swiftSupported: true
  }
}
```

### Step 6 — Verify

- PaaS resources resolve to private IP addresses from within the VNet
- Public access returns 403/connection refused
- Application connectivity works through private endpoints
- NSP violations logged in Learning mode (if using NSP)
- DNS resolution verified: `nslookup <resource>.privatelink.<service>.windows.net`
