---
name: identity-managed-identity
description: >-
  Identity and authentication patterns using Managed Identity, DefaultAzureCredential, RBAC, workload identity federation, and passwordless connections. USE FOR: configuring service-to-service authentication, assigning Azure roles, or implementing passwordless access to Azure resources.compatibility: Requires Azure CLI, Azure Identity SDK
---

# Identity & Managed Identity Skill

> **MUST:** Managed Identity is the preferred authentication method for all Azure
> service-to-service communication. DO NOT use connection strings with embedded credentials,
> shared access keys, or service principal secrets in application code.

---

## Quick Reference

| Capability               | Description                                                     |
| ------------------------ | --------------------------------------------------------------- |
| System-Assigned Identity | Lifecycle tied to the Azure resource; one identity per resource |
| User-Assigned Identity   | Independent lifecycle; shared across multiple resources         |
| DefaultAzureCredential   | Automatic credential chain for local dev and production         |
| RBAC Role Assignment     | Least-privilege access via built-in Azure roles                 |
| Workload Identity        | Federated credentials for Kubernetes and external IdPs          |
| Passwordless Connections | Connection strings without credentials using identity tokens    |

---

## Standards

| Standard                                              | Purpose                              |
| ----------------------------------------------------- | ------------------------------------ |
| [Identity Selection](standards/identity-selection.md) | When to use which identity type      |
| [RBAC Patterns](standards/rbac-patterns.md)           | Role assignments and least privilege |
| [Checklist](standards/checklist.md)                   | Validation checklist                 |

---

## Actions

| Action                                              | When to use                 |
| --------------------------------------------------- | --------------------------- |
| [Configure Identity](actions/configure-identity.md) | Setting up managed identity |

---

## DefaultAzureCredential Chain

`DefaultAzureCredential` automatically selects the right credential for the environment:

| Environment        | Credential Used                              |
| ------------------ | -------------------------------------------- |
| Local development  | Azure CLI / Visual Studio / VS Code          |
| Azure VM / App Svc | Managed Identity (system or user-assigned)   |
| Azure Container    | Managed Identity                             |
| AKS                | Workload Identity (via federated token)      |
| CI/CD pipeline     | Environment variables / Federated credential |

### Usage

> **Warning — register as singleton:** `DefaultAzureCredential` caches tokens internally.
> Creating a new instance per request causes token acquisition overhead and may trigger
> local credential timeouts. Register it once in DI:
>
> ```csharp
> builder.Services.AddSingleton<TokenCredential>(new DefaultAzureCredential());
> ```

> **User-assigned identity:** When using a user-assigned managed identity, set the
> `AZURE_CLIENT_ID` environment variable to the identity's client ID — otherwise
> `DefaultAzureCredential` cannot disambiguate between multiple user-assigned identities
> and will fail with a `CredentialUnavailableException`.

```csharp
// .NET
var credential = new DefaultAzureCredential();
var blobClient = new BlobServiceClient(
    new Uri("https://mystorageaccount.blob.core.windows.net"),
    credential);
```

```typescript
// Node.js
import { DefaultAzureCredential } from "@azure/identity";
import { BlobServiceClient } from "@azure/storage-blob";

const credential = new DefaultAzureCredential();
const blobClient = new BlobServiceClient(
  `https://mystorageaccount.blob.core.windows.net`,
  credential,
);
```

```python
# Python
from azure.identity import DefaultAzureCredential
from azure.storage.blob import BlobServiceClient

credential = DefaultAzureCredential()
blob_client = BlobServiceClient(
    account_url="https://mystorageaccount.blob.core.windows.net",
    credential=credential
)
```

---

## System vs User-Assigned Identity

| Criterion              | System-Assigned                        | User-Assigned                    |
| ---------------------- | -------------------------------------- | -------------------------------- |
| Lifecycle              | Tied to the resource                   | Independent                      |
| Sharing                | One resource only                      | Multiple resources               |
| Bicep setup            | `identity: { type: 'SystemAssigned' }` | Reference existing identity      |
| Best for               | Single-resource, simple scenarios      | Shared identity, pre-provisioned |
| Role assignment timing | After resource creation                | Before or after                  |

**Default choice:** Use System-Assigned unless you need to share the identity or
pre-provision role assignments.

---

## Bicep — Managed Identity + RBAC

```bicep
// System-assigned identity on a Container App
resource containerApp 'Microsoft.App/containerApps@2025-07-01' = {
  name: 'ca-api'
  location: location
  identity: { type: 'SystemAssigned' }
  // ... rest of configuration
}

// Role assignment — Key Vault Secrets User
resource kvRoleAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(keyVault.id, containerApp.id, keyVaultSecretsUser)
  scope: keyVault
  properties: {
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', keyVaultSecretsUser)
    principalId: containerApp.identity.principalId
    principalType: 'ServicePrincipal'
  }
}

> **RBAC propagation delay:** Azure RBAC role assignments use eventual consistency.
> After a `roleAssignment` is created, it can take **up to 10 minutes** for the permission
> to propagate. Applications that call Azure APIs immediately after deployment may receive
> `403 Forbidden` until propagation completes. Build retry logic into startup code and
> do not treat initial 403s as permanent failures.
```

> [!WARNING]
> Azure RBAC assignments can take up to 10 minutes to propagate. Applications that call Azure APIs immediately after role assignment may receive 403 errors. Implement retry with exponential backoff (initial delay 5s, max 60s, 5 attempts) for the first operation after identity configuration.

### Common RBAC Role IDs

| Role                                | ID                                     | Use For                 |
| ----------------------------------- | -------------------------------------- | ----------------------- |
| Key Vault Secrets User              | `4633458b-17de-408a-b874-0445c86b69e6` | Reading secrets         |
| Storage Blob Data Contributor       | `ba92f5b4-2d11-453d-a403-e96b0029c9fe` | Read/write blobs        |
| Storage Blob Data Reader            | `2a2b9908-6ea1-4ae2-8e65-a410df84e7d1` | Read-only blob access   |
| Azure Service Bus Data Sender       | `69a216fc-b8fb-44d8-bc22-1f3c2cd27a39` | Send messages           |
| Azure Service Bus Data Receiver     | `4f6d3b9b-027b-4f4c-9142-0e5a2a2247e0` | Receive messages        |
| Cosmos DB Built-in Data Contributor | `00000000-0000-0000-0000-000000000002` | Read/write Cosmos data  |
| Azure SQL DB Contributor            | `9b7fa17d-e63e-47b0-bb0a-15c516ac86ec` | SQL database management |

---

## Passwordless Connection Strings

> [!NOTE]
> Register `DefaultAzureCredential` as a **singleton** in DI to avoid per-request credential negotiation overhead. Example: `builder.Services.AddSingleton<TokenCredential>(new DefaultAzureCredential());`

### Azure SQL

```csharp
// No password — uses managed identity token
"Server=tcp:myserver.database.windows.net;Database=mydb;Authentication=Active Directory Managed Identity;"
```

### Azure PostgreSQL

```csharp
// Token-based authentication
var credential = new DefaultAzureCredential();
var token = await credential.GetTokenAsync(new TokenRequestContext(
    new[] { "https://ossrdbms-aad.database.windows.net/.default" }));

var connectionString = $"Host=myserver.postgres.database.azure.com;Database=mydb;Username=managed-identity-name;Password={token.Token};SSL Mode=Require;";
```

### Azure Storage

```csharp
// No connection string needed — URI + credential
new BlobServiceClient(new Uri("https://account.blob.core.windows.net"), new DefaultAzureCredential());
```

---

## Workload Identity (AKS / External)

For AKS or external Kubernetes clusters, use workload identity federation:

> **AKS Book note:** AAD Pod Identity is deprecated; use Workload Identity for all AKS pods.

```bicep
resource userIdentity 'Microsoft.ManagedIdentity/userAssignedIdentities@2024-11-30' = {
  name: 'id-${serviceName}'
  location: location
}

resource federatedCredential 'Microsoft.ManagedIdentity/userAssignedIdentities/federatedIdentityCredentials@2024-11-30' = {
  parent: userIdentity
  name: 'fed-${serviceName}'
  properties: {
    issuer: aksCluster.properties.oidcIssuerProfileUrl
    subject: 'system:serviceaccount:${namespace}:${serviceAccountName}'
    audiences: ['api://AzureADTokenExchange']
  }
}
```

---

## Principles

1. **Managed Identity first** — always prefer managed identity over secrets or keys.
2. **Least privilege** — assign the narrowest built-in role that satisfies the requirement.
3. **Disable local auth** — set `disableLocalAuth: true` on resources that support it.
4. **DefaultAzureCredential** — use it everywhere; it works locally and in production.
5. **No secrets in code** — never hardcode credentials, keys, or connection strings.
6. **Scope role assignments** — assign at the resource level, not subscription or resource group.

---

## References

- [Managed identities overview](https://learn.microsoft.com/entra/identity/managed-identities-azure-resources/overview)
- [DefaultAzureCredential](https://learn.microsoft.com/dotnet/azure/sdk/authentication/credential-chains)
- [Azure built-in roles](https://learn.microsoft.com/azure/role-based-access-control/built-in-roles)
- [Passwordless connections](https://learn.microsoft.com/azure/developer/intro/passwordless-overview)
- [Workload identity federation](https://learn.microsoft.com/entra/workload-id/workload-identity-federation)

---

## Currency and Verification

- **Date checked:** 2026-03-31
- **API version used:** `Microsoft.ManagedIdentity/userAssignedIdentities@2024-11-30`
- **Compatibility:** .NET 10, Azure.Identity SDK 1.x, Azure CLI, Bicep
- **Sources:** [Managed identities overview](https://learn.microsoft.com/entra/identity/managed-identities-azure-resources/overview), [Azure Identity SDK](https://learn.microsoft.com/dotnet/azure/sdk/authentication/credential-chains)
- **Verification steps:**
  1. Verify API version: `az provider show --namespace Microsoft.ManagedIdentity --query "resourceTypes[?resourceType=='userAssignedIdentities'].apiVersions" -o tsv`
  2. Check Azure.Identity SDK version: `dotnet list package | grep Azure.Identity`
  3. Verify role assignment propagation: allow 5+ minutes after `az role assignment create` before testing access

### Known Pitfalls

| Area                                   | Pitfall                                                                                                     | Mitigation                                                                                             |
| -------------------------------------- | ----------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------ |
| `DefaultAzureCredential` in production | Credential chain tries multiple providers sequentially; causes latency and ambiguity in failure diagnostics | Use `ManagedIdentityCredential` directly in production; reserve `DefaultAzureCredential` for local dev |
| Token expiration                       | Entra ID tokens expire after ~1 hour; using a cached token string as a static password fails silently       | Use `UsePeriodicPasswordProvider` (Npgsql) or `TokenCredential`-based clients that auto-refresh        |
| Role assignment propagation            | RBAC assignments take up to 5 minutes to propagate; immediate access attempts fail with 403                 | Add retry logic or wait step after role assignment; do not fail deployment on first 403                |
| Federated credential naming            | Duplicate federated credential subjects across environments cause `AADSTS700016` errors                     | Use deterministic naming with `uniqueString()` and distinct subjects per environment                   |
| System vs User-Assigned lifecycle      | System-assigned identity is deleted when the resource is deleted; dependent role assignments break          | Use user-assigned identity for resources with shared role assignments or cross-service dependencies    |

## Related Skills

- **azure-role-selector** — Detailed role selection guidance
- **azure-container-apps** — Container Apps with managed identity
- **azure-functions-patterns** — Functions with managed identity bindings
- **event-driven-messaging** — RBAC for Service Bus and Event Hubs
