# RBAC Patterns Standard

## Least Privilege Principle

Always assign the narrowest built-in role that satisfies the requirement.
Never use `Owner` or `Contributor` for data plane access — use data-specific roles.

---

## Role Assignment Scope

Assign roles at the most specific scope possible:

| Scope Level    | Use When                                       | Example                              |
| -------------- | ---------------------------------------------- | ------------------------------------ |
| Resource       | Service needs access to one specific resource  | Blob Data Reader on one storage acct |
| Resource Group | Service needs access to all resources in group | Rare — only when truly needed        |
| Subscription   | Platform-level operations only                 | Deployment pipelines only            |

**Never assign data plane roles at subscription scope.**

---

## Common Role Patterns

### Web API accessing storage

```bicep
// Storage Blob Data Contributor — read/write blobs
var storageBlobDataContributor = 'ba92f5b4-2d11-453d-a403-e96b0029c9fe'

resource blobRole 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(storageAccount.id, appIdentity, storageBlobDataContributor)
  scope: storageAccount
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', storageBlobDataContributor)
    principalId: appPrincipalId
    principalType: 'ServicePrincipal'
  }
}
```

### Background worker reading secrets

```bicep
// Key Vault Secrets User — read secrets only
var keyVaultSecretsUser = '4633458b-17de-408a-b874-0445c86b69e6'

resource kvRole 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(keyVault.id, workerIdentity, keyVaultSecretsUser)
  scope: keyVault
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', keyVaultSecretsUser)
    principalId: workerPrincipalId
    principalType: 'ServicePrincipal'
  }
}
```

### Service sending messages

```bicep
// Azure Service Bus Data Sender
var serviceBusDataSender = '69a216fc-b8fb-44d8-bc22-1f3c2cd27a39'

resource sbRole 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(serviceBus.id, senderIdentity, serviceBusDataSender)
  scope: serviceBus
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', serviceBusDataSender)
    principalId: senderPrincipalId
    principalType: 'ServicePrincipal'
  }
}
```

---

## Role Assignment Anti-Patterns

| Anti-Pattern                   | Why It's Wrong                                | Correct Approach                    |
| ------------------------------ | --------------------------------------------- | ----------------------------------- |
| `Owner` for data access        | Grants management plane + data plane          | Use data-specific role              |
| `Contributor` at subscription  | Overly broad; can create/delete resources     | Scope to resource group or resource |
| Connection string with key     | Shared secret; no audit trail per identity    | Managed Identity + RBAC             |
| Shared service principal       | Single blast radius; credential rotation pain | Per-service managed identity        |
| `Storage Account Key Operator` | Allows key access for all data                | `Storage Blob Data Reader/Writer`   |

---

## Role Assignment Naming

Use deterministic GUIDs for role assignment names to ensure idempotent deployments:

```bicep
// Pattern: guid(scope, principalId, roleDefinitionId)
name: guid(storageAccount.id, containerApp.identity.principalId, storageBlobDataContributor)
```

This ensures:

- Same deployment produces the same role assignment name
- Redeployments don't create duplicates
- Different principal/role combinations get unique names

---

## Rules

1. Use built-in roles — create custom roles only when no built-in role fits.
2. Assign at resource scope — never at subscription scope for data access.
3. Use `principalType: 'ServicePrincipal'` — speed up assignment propagation.
4. Use deterministic GUIDs for assignment names — ensure idempotent IaC.
5. Audit role assignments quarterly — remove unused or overprivileged assignments.
6. Prefer read-only roles — use `Data Reader` unless write access is explicitly needed.
