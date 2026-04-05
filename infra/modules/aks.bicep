targetScope = 'resourceGroup'

@description('Azure region for the AKS cluster.')
param location string

@description('Environment name used in resource tags and derived names.')
param environmentName string

@description('Name of the AKS cluster to create.')
param aksClusterName string

@description('Number of nodes in the system node pool.')
@minValue(1)
param nodeCount int = 3

@description('Resource ID of the Log Analytics workspace used by the AKS monitoring add-on.')
param logAnalyticsWorkspaceId string

@description('Resource ID of the data collection rule associated with the cluster.')
param dcrId string

@description('Tags applied to the AKS cluster and extension resources.')
param tags object

var effectiveTags = union(tags, {
  'azd-env-name': environmentName
})
var dnsPrefixBase = toLower(replace(aksClusterName, '_', '-'))
var dnsPrefix = take(dnsPrefixBase, 54)

resource aksCluster 'Microsoft.ContainerService/managedClusters@2024-02-01' = {
  name: aksClusterName
  location: location
  tags: effectiveTags
  identity: {
    type: 'SystemAssigned'
  }
  properties: {
    kubernetesVersion: '1.28'
    dnsPrefix: dnsPrefix
    enableRBAC: true
    oidcIssuerProfile: {
      enabled: true
    }
    securityProfile: {
      workloadIdentity: {
        enabled: true
      }
    }
    agentPoolProfiles: [
      {
        name: 'systempool'
        mode: 'System'
        count: nodeCount
        vmSize: 'Standard_D4s_v5'
        osType: 'Linux'
        type: 'VirtualMachineScaleSets'
      }
    ]
    addonProfiles: {
      omsagent: {
        enabled: true
        config: {
          logAnalyticsWorkspaceResourceID: logAnalyticsWorkspaceId
        }
      }
    }
    networkProfile: {
      networkPlugin: 'azure'
    }
  }
}

resource containerLogAssociation 'Microsoft.Insights/dataCollectionRuleAssociations@2022-06-01' = {
  scope: aksCluster
  name: 'containerLogV2Association'
  properties: {
    dataCollectionRuleId: dcrId
    description: 'Associates the AKS cluster with the ContainerLogV2 data collection rule.'
  }
}

output aksClusterId string = aksCluster.id
output aksOidcIssuerUrl string = aksCluster.properties.oidcIssuerProfile.issuerURL
output aksFqdn string = aksCluster.properties.fqdn
