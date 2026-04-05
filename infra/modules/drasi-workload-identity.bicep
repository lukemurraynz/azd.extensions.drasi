targetScope = 'resourceGroup'

@description('Client ID of the user-assigned managed identity bound to the Drasi workload.')
param uamiClientId string

@description('Namespace containing the Drasi Kubernetes resources.')
param drasiNamespace string = 'drasi-system'

// NOTE: The Go provision command consumes these outputs and patches the Drasi
// ServiceAccount and workload pod specs after Azure resources are deployed.
output serviceAccountAnnotation string = 'azure.workload.identity/client-id: ${uamiClientId}'
output podLabel string = 'azure.workload.identity/use: "true"'
output namespace string = drasiNamespace
