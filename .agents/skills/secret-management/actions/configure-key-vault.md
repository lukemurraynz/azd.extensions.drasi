# Configure Azure Key Vault

## Steps

### 1. Deploy Key Vault with RBAC

```bicep
param location string = resourceGroup().location
param vaultName string
param managedIdentityPrincipalId string

resource keyVault 'Microsoft.KeyVault/vaults@2025-05-01' = {
  name: vaultName
  location: location
  properties: {
    tenantId: subscription().tenantId
    sku: {
      family: 'A'
      name: 'standard'
    }
    enableRbacAuthorization: true
    enableSoftDelete: true
    softDeleteRetentionInDays: 90
    enablePurgeProtection: true
    publicNetworkAccess: 'Disabled'
    networkAcls: {
      bypass: 'AzureServices'
      defaultAction: 'Deny'
    }
  }
}

// Secrets User role for application identity
resource secretsUserRole 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(keyVault.id, managedIdentityPrincipalId, '4633458b-17de-408a-b874-0445c86b69e6')
  scope: keyVault
  properties: {
    roleDefinitionId: subscriptionResourceId(
      'Microsoft.Authorization/roleDefinitions',
      '4633458b-17de-408a-b874-0445c86b69e6' // Key Vault Secrets User
    )
    principalId: managedIdentityPrincipalId
    principalType: 'ServicePrincipal'
  }
}
```

### 2. Create Secrets

Use Azure CLI to set secrets — never store secret values in IaC:

```bash
az keyvault secret set \
  --vault-name $VAULT_NAME \
  --name "external-api-key" \
  --value "$API_KEY_VALUE" \
  --expires "2025-12-31T00:00:00Z"
```

### 3. Configure Application Access

**Option A — Key Vault References (App Service / Functions):**

```bicep
resource appService 'Microsoft.Web/sites@2025-03-01' = {
  // ...existing app config...
  properties: {
    siteConfig: {
      appSettings: [
        {
          name: 'ExternalApi__ApiKey'
          value: '@Microsoft.KeyVault(SecretUri=${keyVault.properties.vaultUri}secrets/external-api-key)'
        }
      ]
    }
  }
}
```

**Option B — Direct SDK Access (.NET):**

```csharp
builder.Configuration.AddAzureKeyVault(
    new Uri(builder.Configuration["KeyVault:VaultUri"]!),
    new DefaultAzureCredential());
```

**Option C — CSI Secret Store Driver (AKS):**

**Preferred (AKS Book): Workload Identity**

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: my-app
  namespace: my-namespace
  annotations:
    azure.workload.identity/client-id: "<managed-identity-client-id>"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    metadata:
      labels:
        azure.workload.identity/use: "true"
```

```yaml
# SecretProviderClass (Workload Identity)
apiVersion: secrets-store.csi.x-k8s.io/v1
kind: SecretProviderClass
metadata:
  name: azure-keyvault-secrets-wi
spec:
  provider: azure
  parameters:
    usePodIdentity: "false"
    clientID: "<managed-identity-client-id>" # Workload Identity
    keyvaultName: "<vault-name>"
    tenantId: "<tenant-id>"
    objects: |
      array:
        - |
          objectName: external-api-key
          objectType: secret
```

**Fallback: Kubelet Identity (node-level user-assigned MI)**

```yaml
apiVersion: secrets-store.csi.x-k8s.io/v1
kind: SecretProviderClass
metadata:
  name: azure-keyvault-secrets
spec:
  provider: azure
  parameters:
    usePodIdentity: "false"
    useVMManagedIdentity: "true"
    userAssignedIdentityID: "<managed-identity-client-id>"
    keyvaultName: "<vault-name>"
    tenantId: "<tenant-id>"
    objects: |
      array:
        - |
          objectName: external-api-key
          objectType: secret
```

### 4. Configure Secret Rotation (Optional)

```bicep
// Event Grid subscription for near-expiry notifications
resource eventSubscription 'Microsoft.EventGrid/eventSubscriptions@2025-02-15' = {
  name: 'secret-rotation'
  scope: keyVault
  properties: {
    destination: {
      endpointType: 'AzureFunction'
      properties: {
        resourceId: rotationFunctionId
      }
    }
    filter: {
      includedEventTypes: [
        'Microsoft.KeyVault.SecretNearExpiry'
      ]
    }
  }
}
```

### 5. Verify

```bash
# Confirm RBAC is enabled
az keyvault show --name $VAULT_NAME --query "properties.enableRbacAuthorization"

# List role assignments
az role assignment list --scope $(az keyvault show --name $VAULT_NAME --query id -o tsv) -o table

# Test secret access (with permitted identity)
az keyvault secret show --vault-name $VAULT_NAME --name "external-api-key" --query "value"
```
