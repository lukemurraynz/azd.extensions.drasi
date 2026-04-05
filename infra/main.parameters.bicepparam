using './main.bicep'

param environmentName = readEnvironmentVariable('AZURE_ENV_NAME', 'dev')
param location = readEnvironmentVariable('AZURE_LOCATION', 'eastus')
param aksClusterName = 'aks-${readEnvironmentVariable('AZURE_ENV_NAME', 'dev')}'
param keyVaultName = 'kv-drasi-${readEnvironmentVariable('AZURE_ENV_NAME', 'dev')}'
param uamiName = 'uami-drasi-${readEnvironmentVariable('AZURE_ENV_NAME', 'dev')}'
param logAnalyticsWorkspaceName = 'law-drasi-${readEnvironmentVariable('AZURE_ENV_NAME', 'dev')}'
param usePrivateAcr = false
param enableCosmosDb = false
param enableEventHub = false
param enableServiceBus = false
