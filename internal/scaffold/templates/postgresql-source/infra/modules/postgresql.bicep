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

// Allow Azure services (including AKS) to connect to PostgreSQL.
// Start/end IP of 0.0.0.0 is the Azure-internal signal for "Allow Azure services".
// For production, use private endpoints instead.
resource allowAzureServices 'Microsoft.DBforPostgreSQL/flexibleServers/firewallRules@2024-08-01' = {
  parent: postgresServer
  name: 'AllowAllAzureServicesAndResourcesWithinAzureIps'
  properties: {
    startIpAddress: '0.0.0.0'
    endIpAddress: '0.0.0.0'
  }
}

// ---------------------------------------------------------------------------
// Deployment Script — bootstrap the database schema and grant REPLICATION role.
// Uses `az postgres flexible-server execute` from the AzureCLI deployment script
// container. This command routes through the ARM management plane so the script
// does not require direct network connectivity to the PostgreSQL server.
// Requires a user-assigned managed identity (no special Azure roles needed — auth
// is via PostgreSQL admin credentials, not Azure RBAC).
// ---------------------------------------------------------------------------

@description('Resource ID of a user-assigned managed identity for the deployment script.')
param scriptIdentityId string

// The deployment script runs psql to:
// 1. Create the orders table (IF NOT EXISTS) — required by the sample continuous query.
// 2. Grant REPLICATION to the admin role — required for Drasi PostgreSQL CDC.
// The script is idempotent and safe to re-run on subsequent provisions.
resource dbBootstrap 'Microsoft.Resources/deploymentScripts@2023-08-01' = {
  name: 'pg-bootstrap-${uniqueString(resourceGroup().id)}'
  location: location
  kind: 'AzureCLI'
  identity: {
    type: 'UserAssigned'
    userAssignedIdentities: {
      '${scriptIdentityId}': {}
    }
  }
  properties: {
    azCliVersion: '2.67.0'
    retentionInterval: 'PT1H'
    timeout: 'PT10M'
    cleanupPreference: 'OnSuccess'
    environmentVariables: [
      { name: 'PG_SERVER_NAME', value: postgresServer.name }
      { name: 'PG_ADMIN_USER', value: adminLogin }
      { name: 'PG_DATABASE', value: databaseName }
      { name: 'PG_ADMIN_PASSWORD', secureValue: administratorLoginPassword }
    ]
    scriptContent: '''
      set -euo pipefail

      echo "Creating orders table..."
      az postgres flexible-server execute \
        --name "$PG_SERVER_NAME" \
        --admin-user "$PG_ADMIN_USER" \
        --admin-password "$PG_ADMIN_PASSWORD" \
        --database-name "$PG_DATABASE" \
        --querytext "CREATE TABLE IF NOT EXISTS public.orders (id SERIAL PRIMARY KEY, status VARCHAR(50) NOT NULL DEFAULT 'pending', created_at TIMESTAMPTZ NOT NULL DEFAULT NOW());"

      echo "Granting REPLICATION role..."
      az postgres flexible-server execute \
        --name "$PG_SERVER_NAME" \
        --admin-user "$PG_ADMIN_USER" \
        --admin-password "$PG_ADMIN_PASSWORD" \
        --database-name "$PG_DATABASE" \
        --querytext "ALTER ROLE \"$PG_ADMIN_USER\" REPLICATION;"

      echo "Database bootstrap completed successfully"
    '''
  }
  dependsOn: [
    drasiDatabase
    walLevelParam
    allowAzureServices
  ]
}

output serverFqdn string = postgresServer.properties.fullyQualifiedDomainName
output databaseName string = databaseName
output serverName string = postgresServer.name
