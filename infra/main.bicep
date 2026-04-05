targetScope = 'resourceGroup'

@description('Azure region for all resources in this deployment.')
param location string = resourceGroup().location

@description('Environment name applied to resources and tags.')
param environmentName string

@description('Name of the AKS cluster to create.')
param aksClusterName string

@description('Namespace containing the Drasi Kubernetes workloads.')
param drasiNamespace string = 'drasi-system'

@description('Base name used to derive the Key Vault name.')
param keyVaultName string

@description('Name of the user-assigned managed identity.')
param uamiName string

@description('Base name used to identify the Log Analytics deployment in this template.')
param logAnalyticsWorkspaceName string

@description('Set to true to deploy a private Azure Container Registry.')
param usePrivateAcr bool = false

@description('Base name used to derive the Azure Container Registry name when private ACR is enabled.')
param acrName string = ''

@description('Set to true to deploy the Cosmos DB Gremlin account.')
param enableCosmosDb bool = false

@description('Set to true to deploy the Event Hubs namespace.')
param enableEventHub bool = false

@description('Set to true to deploy the Service Bus namespace.')
param enableServiceBus bool = false

@description('OTLP endpoint URL for Drasi runtime workload telemetry export.')
param otlpEndpoint string = 'http://otel-collector.${drasiNamespace}.svc.cluster.local:4318'

@description('Number of nodes in the AKS system node pool.')
@minValue(1)
param nodeCount int = 3

@description('Base tags supplied by the caller.')
param tags object = {
  'azd-env-name': environmentName
}

var uniqueSuffix = uniqueString(resourceGroup().id)
var effectiveTags = union(tags, {
  displayName: 'Drasi ${environmentName}'
  locationIdentifier: 'az.public.${location}'
  cloud: 'public'
  'azd-env-name': environmentName
})
var resolvedKeyVaultName = take('${take(toLower(keyVaultName), 17)}-${take(uniqueSuffix, 6)}', 24)
var sanitizedAcrBase = toLower(replace(replace(acrName, '-', ''), '_', ''))
var resolvedAcrName = take('${take(sanitizedAcrBase, 44)}${take(uniqueSuffix, 6)}', 50)
var resolvedCosmosName = take('cosmos-${environmentName}-${take(uniqueSuffix, 6)}', 44)
var resolvedEventHubNamespaceName = take('eh-${environmentName}-${take(uniqueSuffix, 6)}', 50)
var resolvedServiceBusNamespaceName = take('sb-${environmentName}-${take(uniqueSuffix, 6)}', 50)

module logAnalytics 'modules/loganalytics.bicep' = {
  name: 'loganalytics-${take(logAnalyticsWorkspaceName, 32)}'
  params: {
    location: location
    environmentName: environmentName
    tags: effectiveTags
  }
}

module keyVault 'modules/keyvault.bicep' = {
  name: 'keyvault-${environmentName}'
  params: {
    location: location
    environmentName: environmentName
    keyVaultName: resolvedKeyVaultName
    tags: effectiveTags
  }
}

module acr 'modules/acr.bicep' = if (usePrivateAcr) {
  name: 'acr-${environmentName}'
  params: {
    location: location
    environmentName: environmentName
    acrName: resolvedAcrName
    tags: effectiveTags
  }
}

module uami 'modules/uami.bicep' = {
  name: 'uami-${environmentName}'
  params: {
    location: location
    environmentName: environmentName
    uamiName: uamiName
    keyVaultId: keyVault.outputs.keyVaultId
    logAnalyticsWorkspaceId: logAnalytics.outputs.workspaceId
    usePrivateAcr: usePrivateAcr
    acrId: usePrivateAcr ? acr!.outputs.acrId : ''
    tags: effectiveTags
  }
}

module aks 'modules/aks.bicep' = {
  name: 'aks-${environmentName}'
  params: {
    location: location
    environmentName: environmentName
    aksClusterName: aksClusterName
    nodeCount: nodeCount
    logAnalyticsWorkspaceId: logAnalytics.outputs.workspaceId
    dcrId: logAnalytics.outputs.dcrId
    tags: effectiveTags
  }
}

module fedCred 'modules/fedcred.bicep' = {
  name: 'fedcred-${environmentName}'
  params: {
    uamiName: uamiName
    oidcIssuerUrl: aks.outputs.aksOidcIssuerUrl
    drasiNamespace: drasiNamespace
  }
  dependsOn: [
    uami
  ]
}

module workloadIdentity 'modules/drasi-workload-identity.bicep' = {
  name: 'workload-identity-${environmentName}'
  params: {
    uamiClientId: uami.outputs.uamiClientId
    drasiNamespace: drasiNamespace
  }
}

module runtimeObservability 'modules/drasi-runtime-observability.bicep' = {
  name: 'runtime-observability-${environmentName}'
  params: {
    otlpEndpoint: otlpEndpoint
    drasiNamespace: drasiNamespace
  }
}

module cosmosDb 'modules/cosmos.bicep' = if (enableCosmosDb) {
  name: 'cosmos-${environmentName}'
  params: {
    location: location
    environmentName: environmentName
    cosmosAccountName: resolvedCosmosName
    tags: effectiveTags
  }
}

module eventHub 'modules/eventhub.bicep' = if (enableEventHub) {
  name: 'eventhub-${environmentName}'
  params: {
    location: location
    environmentName: environmentName
    eventHubNamespaceName: resolvedEventHubNamespaceName
    tags: effectiveTags
  }
}

module serviceBus 'modules/servicebus.bicep' = if (enableServiceBus) {
  name: 'servicebus-${environmentName}'
  params: {
    location: location
    environmentName: environmentName
    serviceBusNamespaceName: resolvedServiceBusNamespaceName
    tags: effectiveTags
  }
}

output aksClusterId string = aks.outputs.aksClusterId
output aksOidcIssuerUrl string = aks.outputs.aksOidcIssuerUrl
output keyVaultUri string = keyVault.outputs.keyVaultUri
output uamiClientId string = uami.outputs.uamiClientId
output logAnalyticsWorkspaceId string = logAnalytics.outputs.workspaceId
