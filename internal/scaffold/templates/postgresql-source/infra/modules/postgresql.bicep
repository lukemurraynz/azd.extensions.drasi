// postgresql.bicep — Azure Database for PostgreSQL Flexible Server
// Configured with logical replication enabled for Drasi CDC (change data capture).
// NOTE: For production, disable public network access and use a private endpoint.

@description('The Azure region for all resources.')
param location string

@description('Environment name suffix applied to all resource names.')
param environmentName string

@description('Tags to apply to all resources.')
param tags object = {}

// SECURITY: Admin password must be supplied as a secure parameter and stored in Key Vault.
// Never use deterministic functions (uniqueString) for credentials.
@secure()
@description('Administrator login password for the PostgreSQL server. Store in Key Vault for production.')
param administratorLoginPassword string

@description('Public network access setting. Use Disabled with private endpoints for production.')
@allowed([
  'Enabled'
  'Disabled'
])
param publicNetworkAccess string = 'Enabled'

var serverName = 'psql-drasi-${environmentName}-${uniqueString(resourceGroup().id)}'
var databaseName = 'drasidb'
var adminLogin = 'drasiAdmin'

// PostgreSQL Flexible Server with logical replication enabled for Drasi CDC.
// SKU: Burstable B1ms — cost-effective for development and light workloads.
// NOTE: Enable private endpoints and disable public access for production deployments.
resource postgresServer 'Microsoft.DBforPostgreSQL/flexibleServers@2024-08-01' = {
  name: serverName
  location: location
  tags: tags
  sku: {
    name: 'Standard_B1ms'
    tier: 'Burstable'
  }
  properties: {
    administratorLogin: adminLogin
    administratorLoginPassword: administratorLoginPassword
    version: '16'
    storage: {
      storageSizeGB: 32
    }
    backup: {
      backupRetentionDays: 7
      geoRedundantBackup: 'Disabled'
    }
    highAvailability: {
      mode: 'Disabled'
    }
    network: {
      publicNetworkAccess: publicNetworkAccess
    }
  }
}

// Enable logical replication — required for Drasi CDC via PostgreSQL replication slots.
resource walLevelParam 'Microsoft.DBforPostgreSQL/flexibleServers/configurations@2024-08-01' = {
  parent: postgresServer
  name: 'wal_level'
  properties: {
    value: 'logical'
    source: 'user-override'
  }
}

// The Drasi database where your application tables live.
resource drasiDatabase 'Microsoft.DBforPostgreSQL/flexibleServers/databases@2024-08-01' = {
  parent: postgresServer
  name: databaseName
  properties: {
    charset: 'UTF8'
    collation: 'en_US.utf8'
  }
}

output serverFqdn string = postgresServer.properties.fullyQualifiedDomainName
output databaseName string = databaseName
output serverName string = postgresServer.name
