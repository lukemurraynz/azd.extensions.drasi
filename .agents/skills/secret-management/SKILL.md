---
name: secret-management
description: >-
  Azure Key Vault secret management with RBAC, managed identity access, secret rotation policies, and integration patterns for AKS and App Service.
  USE FOR: configuring Key Vault RBAC, setting up managed identity secret access, implementing secret rotation, integrating secrets into AKS pods, or connecting App Service to Key Vault.
---

# Secret Management

> **MUST:** Use Azure Key Vault with RBAC access and managed identity. Disable vault access policies in favour of RBAC. DO NOT store secrets in application settings, config files, or source code.

## Description

Patterns for managing secrets, keys, and certificates with Azure Key Vault — RBAC access, managed identity integration, secret rotation, Key Vault references, and CSI driver for AKS.

## Capabilities

| Capability             | Details                                          |
| ---------------------- | ------------------------------------------------ |
| Key Vault Deployment   | Bicep with RBAC, soft delete, purge protection   |
| Secret Access          | Managed identity with RBAC roles                 |
| Key Vault References   | App Service/Functions settings from Key Vault    |
| Secret Rotation        | Automated rotation with Event Grid notifications |
| CSI Secret Store       | AKS pods accessing Key Vault via CSI driver      |
| Certificate Management | TLS certificates stored and auto-renewed         |

## Standards

| Standard                                        | Purpose                     |
| ----------------------------------------------- | --------------------------- |
| [Secret Patterns](standards/secret-patterns.md) | Access and storage patterns |
| [Checklist](standards/checklist.md)             | Validation checklist        |

## Actions

| Action                                                | Purpose              |
| ----------------------------------------------------- | -------------------- |
| [Configure Key Vault](actions/configure-key-vault.md) | Deploy and configure |

---

## Bicep — Key Vault with RBAC

```bicep
param location string = resourceGroup().location
param vaultName string
param principalId string

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
      defaultAction: 'Deny'
      bypass: 'AzureServices'
    }
  }
}

> **Warning — deployment lockout:** Setting `publicNetworkAccess: 'Disabled'` with
> `defaultAction: 'Deny'` can lock out IaC deployments (Bicep, Terraform, CLI) if the
> deploying agent is not on a permitted network or included in `bypass`. Azure DevOps
> hosted agents and GitHub Actions runners are **not** covered by `'AzureServices'` bypass.
> Either deploy from a self-hosted agent within the VNet, add the agent's IP to the
> firewall rules, or use a two-phase deployment (create vault with public access, configure
> secrets, then lock down).

> **Warning — purge protection:** With `enablePurgeProtection: true`, a deleted vault
> retains its name for the `softDeleteRetentionInDays` period (default 90 days).
> Redeploying with the same vault name will fail with a conflict error. Either recover
> the soft-deleted vault (`az keyvault recover`) or use a different name.

// Key Vault Secrets User role
resource secretsUserRole 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(keyVault.id, principalId, '4633458b-17de-408a-b874-0445c86b69e6')
  scope: keyVault
  properties: {
    principalId: principalId
    roleDefinitionId: subscriptionResourceId(
      'Microsoft.Authorization/roleDefinitions',
      '4633458b-17de-408a-b874-0445c86b69e6' // Key Vault Secrets User
    )
    principalType: 'ServicePrincipal'
  }
}
```

> [!WARNING]
> **IaC lockout trap**: Setting `publicNetworkAccess: 'Disabled'` and `networkAcls.defaultAction: 'Deny'` in a single deployment locks out the deploying pipeline from populating secrets. Use a 2-phase pattern: Phase 1 deploys the vault with `Allow` + sets secrets. Phase 2 switches to `Deny` + adds private endpoint and VNet rules.

---

## Key Vault RBAC Roles

| Role                           | GUID                                   | Access                  |
| ------------------------------ | -------------------------------------- | ----------------------- |
| Key Vault Administrator        | `00482a5a-887f-4fb3-b363-3b7fe8e74483` | Full management         |
| Key Vault Secrets Officer      | `b86a8fe4-44ce-4948-aee5-eccb2c155cd7` | Manage secrets          |
| Key Vault Secrets User         | `4633458b-17de-408a-b874-0445c86b69e6` | Read secrets            |
| Key Vault Certificates Officer | `a4417e6f-fecd-4de8-b567-7b0420556985` | Manage certificates     |
| Key Vault Crypto Officer       | `14b46e9e-c2b7-41b4-b07b-48a6ebf60603` | Manage keys             |
| Key Vault Crypto User          | `12338af0-0e69-4776-bea7-57ae8d297424` | Use keys for crypto ops |

---

## Accessing Secrets

### .NET with DefaultAzureCredential

```csharp
var client = new SecretClient(
    new Uri("https://myvault.vault.azure.net/"),
    new DefaultAzureCredential());

KeyVaultSecret secret = await client.GetSecretAsync("database-password");
string value = secret.Value;
```

### Key Vault References (App Service / Functions)

Reference Key Vault secrets directly in application settings:

```bicep
resource webApp 'Microsoft.Web/sites@2025-03-01' = {
  name: appName
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

**Requirements:**

- App Service/Function has system-assigned managed identity
- Identity has `Key Vault Secrets User` role on the vault

> **Troubleshooting:** When a Key Vault reference fails (wrong identity, network
> restriction, or missing RBAC), the app setting value shows the raw
> `@Microsoft.KeyVault(SecretUri=...)` string instead of the secret value. Check
> the identity, RBAC role, and network access. App Service shows reference status
> under **Configuration > Application settings > Source** column.

### AKS CSI Secret Store Driver

> **AKS Book default:** Use Key Vault CSI driver with **Workload Identity**. Secrets mount as files
> (not stored in etcd) and auto-refresh roughly every 2 minutes. Use `secretObjects` only when
> environment variables are mandatory (secrets then live in etcd and require pod restart).

#### Preferred: Workload Identity (pod-level)

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: my-app
  namespace: my-namespace
  annotations:
    azure.workload.identity/client-id: "<managed-identity-client-id>"
```

```yaml
# Pod (or Deployment) template metadata
metadata:
  labels:
    azure.workload.identity/use: "true"
```

```yaml
# SecretProviderClass (Workload Identity)
apiVersion: secrets-store.csi.x-k8s.io/v1
kind: SecretProviderClass
metadata:
  name: azure-kv-secrets-wi
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
          objectName: database-password
          objectType: secret
```

#### Fallback: Kubelet Identity (node-level user-assigned MI)

```yaml
# SecretProviderClass
apiVersion: secrets-store.csi.x-k8s.io/v1
kind: SecretProviderClass
metadata:
  name: azure-kv-secrets
spec:
  provider: azure
  parameters:
    usePodIdentity: "false"
    useVMManagedIdentity: "true"
    userAssignedIdentityID: "<managed-identity-client-id>"
    keyvaultName: "myvault"
    objects: |
      array:
        - |
          objectName: database-password
          objectType: secret
        - |
          objectName: tls-certificate
          objectType: cert
    tenantId: "<tenant-id>"
---
# Pod mounting secrets
apiVersion: v1
kind: Pod
metadata:
  name: my-app
spec:
  containers:
    - name: app
      image: myapp:<tag>
      volumeMounts:
        - name: secrets
          mountPath: "/mnt/secrets"
          readOnly: true
  volumes:
    - name: secrets
      csi:
        driver: secrets-store.csi.k8s.io
        readOnly: true
        volumeAttributes:
          secretProviderClass: "azure-kv-secrets"
```

---

## Secret Rotation

### Event Grid Notification on Secret Near Expiry

```bicep
resource kvEventSubscription 'Microsoft.EventGrid/eventSubscriptions@2025-02-15' = {
  name: 'secret-near-expiry'
  scope: keyVault
  properties: {
    destination: {
      endpointType: 'AzureFunction'
      properties: {
        resourceId: rotationFunction.id
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

### Rotation Function Pattern

```csharp
[Function("RotateSecret")]
public async Task Run(
    [EventGridTrigger] EventGridEvent eventGridEvent)
{
    var secretName = eventGridEvent.Subject;

    // 1. Generate new credential in the target service
    var newPassword = await _targetService.RotateCredentialAsync();

    // 2. Update the secret in Key Vault
    await _secretClient.SetSecretAsync(secretName, newPassword);

    // 3. Verify the new secret works
    await _targetService.VerifyCredentialAsync(newPassword);
}
```

---

## Principles

1. **RBAC over access policies** — `enableRbacAuthorization: true` always.
2. **Least-privilege roles** — use `Key Vault Secrets User` for read-only, not `Administrator`.
3. **Managed identity for access** — no client secrets or certificates for Key Vault authentication.
4. **Soft delete and purge protection** — always enabled in production.
5. **Automate rotation** — secrets should rotate before they expire.
6. **Prefer identity-based connections** — many Azure services support managed identity directly, eliminating the need for Key Vault secrets entirely.

## Currency and Verification

- **Date checked:** 2026-03-31
- **Compatibility:** Azure Key Vault, AKS CSI driver, App Service/Functions Key Vault references
- **Sources:** [Key Vault docs](https://learn.microsoft.com/azure/key-vault/general/overview), [Key Vault RBAC](https://learn.microsoft.com/azure/key-vault/general/rbac-guide), [CSI Secret Store driver](https://learn.microsoft.com/azure/aks/csi-secrets-store-driver)
- **Verification steps:**
  1. Verify Key Vault API version: `az provider show --namespace Microsoft.KeyVault --query "resourceTypes[?resourceType=='vaults'].apiVersions" -o tsv`
  2. Verify RBAC access model: `az keyvault show --name <vault> --query properties.enableRbacAuthorization`
  3. Test secret access: `az keyvault secret show --vault-name <vault> --name <secret>`

### Known Pitfalls

| Area                   | Pitfall                                                                                              | Mitigation                                                                                              |
| ---------------------- | ---------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------- |
| Vault name reuse       | Recreating a Key Vault with the same name during the purge-protection window (default 90 days) fails | Use `uniqueString()` in Bicep names or purge the soft-deleted vault: `az keyvault purge --name <vault>` |
| Access policy vs RBAC  | Mixing vault access policies and RBAC causes confusing permission precedence                         | Use RBAC exclusively (`enableRbacAuthorization: true`); disable vault access policies                   |
| CSI driver sync delay  | CSI Secret Store syncs on pod start only by default; secret rotation isn't picked up                 | Enable `rotationPollInterval` on SecretProviderClass for periodic re-sync                               |
| Key Vault throttling   | Key Vault limits to 4000 transactions per 10 seconds per vault; burst requests get 429 errors        | Cache secrets at application startup; avoid per-request Key Vault calls                                 |
| Secret version pinning | Referencing the latest secret version means rotations take effect immediately (no rollback)          | Pin secret version in Key Vault references for controlled rollouts; update version after validation     |

## References

- [Azure Key Vault documentation](https://learn.microsoft.com/en-us/azure/key-vault/general/overview)
- [Key Vault RBAC](https://learn.microsoft.com/en-us/azure/key-vault/general/rbac-guide)
- [Key Vault references in App Service](https://learn.microsoft.com/en-us/azure/app-service/app-service-key-vault-references)
- [CSI Secret Store driver](https://learn.microsoft.com/en-us/azure/aks/csi-secrets-store-driver)
- [Secret rotation](https://learn.microsoft.com/en-us/azure/key-vault/secrets/tutorial-rotation)

## Related Skills

- [Identity & Managed Identity](../identity-managed-identity/SKILL.md) — RBAC role assignments
- [Private Networking](../private-networking/SKILL.md) — Private endpoints for Key Vault
- [Azure Functions Patterns](../azure-functions-patterns/SKILL.md) — Rotation functions
