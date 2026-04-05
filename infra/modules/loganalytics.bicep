targetScope = 'resourceGroup'

@description('Azure region for the monitoring resources.')
param location string

@description('Environment name used in resource naming.')
param environmentName string

@description('Tags applied to monitoring resources.')
param tags object

var uniqueSuffix = uniqueString(resourceGroup().id)
var effectiveTags = union(tags, {
  'azd-env-name': environmentName
})
var workspaceName = take('log-${environmentName}-${take(uniqueSuffix, 6)}', 63)
var monitorWorkspaceName = take('amw-${environmentName}-${take(uniqueSuffix, 6)}', 63)
var dataCollectionRuleName = take('dcr-${environmentName}-${take(uniqueSuffix, 6)}', 64)

resource logAnalyticsWorkspace 'Microsoft.OperationalInsights/workspaces@2023-09-01' = {
  name: workspaceName
  location: location
  tags: effectiveTags
  properties: {
    retentionInDays: 30
    sku: {
      name: 'PerGB2018'
    }
  }
}

resource prometheusWorkspace 'Microsoft.Monitor/accounts@2023-04-03' = {
  name: monitorWorkspaceName
  location: location
  tags: effectiveTags
  properties: {}
}

resource containerLogRule 'Microsoft.Insights/dataCollectionRules@2023-03-11' = {
  name: dataCollectionRuleName
  location: location
  tags: effectiveTags
  kind: 'Linux'
  properties: {
    dataSources: {
      extensions: [
        {
          name: 'ContainerInsightsExtension'
          extensionName: 'ContainerInsights'
          streams: [
            'Microsoft-ContainerLogV2'
          ]
          extensionSettings: {
            dataCollectionSettings: {
              enableContainerLogV2: true
              interval: '1m'
              namespaceFilteringMode: 'Off'
            }
          }
        }
      ]
    }
    destinations: {
      logAnalytics: [
        {
          name: 'logAnalyticsDestination'
          workspaceResourceId: logAnalyticsWorkspace.id
        }
      ]
    }
    dataFlows: [
      {
        streams: [
          'Microsoft-ContainerLogV2'
        ]
        destinations: [
          'logAnalyticsDestination'
        ]
      }
    ]
  }
}

output workspaceId string = logAnalyticsWorkspace.id
output workspaceCustomerId string = logAnalyticsWorkspace.properties.customerId
output dcrId string = containerLogRule.id
output prometheusWorkspaceId string = prometheusWorkspace.id
