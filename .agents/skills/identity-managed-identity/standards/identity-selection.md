# Identity Selection Standard

## Decision Matrix

| Scenario                                       | Identity Type        | Reason                                   |
| ---------------------------------------------- | -------------------- | ---------------------------------------- |
| Single App Service accessing Key Vault         | System-Assigned      | Simple, lifecycle managed automatically  |
| Multiple Container Apps sharing storage access | User-Assigned        | One identity, one role assignment        |
| AKS pods accessing Azure resources             | Workload Identity    | Federated credential, no secrets in pods |
| CI/CD pipeline deploying to Azure              | Federated Credential | OIDC token exchange, no stored secrets   |
| Local development                              | Azure CLI / VS Code  | DefaultAzureCredential handles this      |
| Cross-tenant access                            | User-Assigned + Fed  | Federated with external IdP              |

---

## System-Assigned Identity

**When to Use:**

- Resource has a single, clear purpose
- No need to share the identity
- Acceptable to create role assignments after the resource exists

**Bicep Pattern:**

```bicep
resource appService 'Microsoft.Web/sites@2025-03-01' = {
  name: appName
  location: location
  identity: { type: 'SystemAssigned' }
  // ...
}

// Role assignment depends on the resource being created first
resource roleAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(storageAccount.id, appService.id, storageBlobDataContributor)
  scope: storageAccount
  properties: {
    roleDefinitionId: subscriptionResourceId(
      'Microsoft.Authorization/roleDefinitions', storageBlobDataContributor)
    principalId: appService.identity.principalId
    principalType: 'ServicePrincipal'
  }
}
```

---

## User-Assigned Identity

**When to Use:**

- Multiple resources need the same access
- Role assignments must exist before the resource is created
- Identity needs to survive resource recreation

**Bicep Pattern:**

```bicep
resource managedIdentity 'Microsoft.ManagedIdentity/userAssignedIdentities@2024-11-30' = {
  name: 'id-${serviceName}'
  location: location
}

// Role assignment can be created before the consuming resource
resource roleAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(storageAccount.id, managedIdentity.id, storageBlobDataContributor)
  scope: storageAccount
  properties: {
    roleDefinitionId: subscriptionResourceId(
      'Microsoft.Authorization/roleDefinitions', storageBlobDataContributor)
    principalId: managedIdentity.properties.principalId
    principalType: 'ServicePrincipal'
  }
}

// Assign to a Container App
resource containerApp 'Microsoft.App/containerApps@2025-07-01' = {
  name: 'ca-api'
  location: location
  identity: {
    type: 'UserAssigned'
    userAssignedIdentities: {
      '${managedIdentity.id}': {}
    }
  }
  // ...
}
```

---

## Workload Identity Federation

**When to Use:**

- AKS workloads accessing Azure resources
- GitHub Actions deploying to Azure (OIDC)
- External identity providers

**AKS Setup:**

1. Enable OIDC issuer on the AKS cluster
2. Create a user-assigned managed identity
3. Create a federated identity credential
4. Annotate the Kubernetes service account

```bash
# Get OIDC issuer URL
oidcUrl=$(az aks show --name $aksName --resource-group $rg --query "oidcIssuerProfile.issuerUrl" -o tsv)

# Create federated credential
az identity federated-credential create \
  --name "fed-${serviceName}" \
  --identity-name "id-${serviceName}" \
  --resource-group $rg \
  --issuer "$oidcUrl" \
  --subject "system:serviceaccount:${namespace}:${serviceAccountName}" \
  --audience "api://AzureADTokenExchange"
```

---

## Resources That Support `disableLocalAuth`

| Resource Type        | Property                       |
| -------------------- | ------------------------------ |
| Application Insights | `DisableLocalAuth: true`       |
| Service Bus          | `disableLocalAuth: true`       |
| Event Hubs           | `disableLocalAuth: true`       |
| Cosmos DB            | `disableLocalAuth: true`       |
| Storage Account      | `allowSharedKeyAccess: false`  |
| Azure SQL            | Microsoft Entra-only auth mode |

**Always disable local auth when the resource supports it.**

---

## Rules

1. Default to System-Assigned identity unless sharing or pre-provisioning is needed.
2. Use User-Assigned identity for multi-resource scenarios or blue/green deployments.
3. Use Workload Identity Federation for AKS and CI/CD — never store service principal secrets.
4. Disable local auth on every resource that supports it.
5. Use `DefaultAzureCredential` in application code — it handles all environments.
