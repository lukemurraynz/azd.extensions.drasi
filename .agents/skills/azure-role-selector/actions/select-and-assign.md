# Action: Select and Assign an Azure RBAC Role

## Inputs

- **Principal**: Managed identity, service principal, or user (provide object/principal ID or resource name)
- **Target resource**: Azure resource the principal needs to access (resource ID or symbolic Bicep name)
- **Required permissions**: Plain-language description of what the principal needs to do (e.g., "read blobs", "pull container images", "read Key Vault secrets")
- **Scope level**: `resource` (preferred), `resourceGroup`, or `subscription` (use sparingly)

---

## Step 1 — Discover the Minimal Built-in Role

### 1a. Check `standards/built-in-roles.md` first

Scan the role catalog for an exact or close match to the required permissions. Prefer roles whose name ends in `Reader` or `User` for read-only access and `Contributor` for read-write. Avoid `Owner` entirely unless managing other identities is explicitly required.

### 1b. Use `Azure MCP/documentation` to search the full catalog

```
Search: "Azure built-in roles <service> least privilege"
```

Evaluate the top candidates:
- Does the role grant *only* what is needed (no extra scopes)?
- Is there a data-plane role narrower than the control-plane `Contributor`?
- Is this a preview or deprecated role? Prefer stable, GA roles.

### 1c. Verify the role definition ID

```bash
# Confirm the role exists and get its ID
az role definition list \
  --query "[?contains(roleName, '<role name candidate>')].[roleName, name]" \
  --output table
```

> **Never hardcode a role ID from memory.** Always verify with the command above or in `standards/built-in-roles.md`.

---

## Step 2 — Determine the Correct Scope

| Scenario | Recommended scope |
|----------|-------------------|
| AKS pod pulling from ACR | Resource group scope on ACR |
| App reading Key Vault secrets | Resource scope on Key Vault |
| Managed identity reading App Configuration | Resource scope on App Configuration store |
| GitHub Actions deploying to AKS | Resource scope on AKS cluster |
| Blanket access for dev environment | Resource group scope (never subscription in prod) |

---

## Step 3 — Generate the CLI Assignment Command

Use `Azure MCP/extension_cli_generate` or construct manually:

```bash
az role assignment create \
  --assignee-object-id <principal-object-id> \
  --assignee-principal-type ServicePrincipal \
  --role "<role-definition-id>" \
  --scope "<resource-id>"
```

**Important:** Pass `--assignee-object-id` (not `--assignee`) to avoid ambiguity with service principals and managed identities.

---

## Step 4 — Generate the Bicep `roleAssignment` Resource

Every role assignment must be declared as infrastructure code, not applied ad-hoc:

```bicep
// <Principal resource name> needs <role name> on <target resource name>
// Reason: <plain language justification>
resource roleAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  // Scope: apply to the target resource (not the resource group)
  scope: targetResource
  // Name: deterministic GUID prevents duplicate assignments on re-deploy
  name: guid(targetResource.id, principalResource.id, '<role-definition-id>')
  properties: {
    roleDefinitionId: resourceId(
      'Microsoft.Authorization/roleDefinitions',
      '<role-definition-id>'
    )
    principalId: principalResource.identity.principalId
    principalType: 'ServicePrincipal'  // or 'User' / 'Group'
  }
}
```

### Common Bicep patterns for this project

#### AKS → ACR (image pull)

```bicep
// AKS kubelet managed identity needs AcrPull on the container registry
resource aksAcrPull 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  scope: containerRegistry
  name: guid(containerRegistry.id, aksCluster.properties.identityProfile.kubeletidentity.objectId, '7f951dda-4ed3-4680-a7ca-43fe172d538d')
  properties: {
    roleDefinitionId: resourceId('Microsoft.Authorization/roleDefinitions', '7f951dda-4ed3-4680-a7ca-43fe172d538d')
    principalId: aksCluster.properties.identityProfile.kubeletidentity.objectId
    principalType: 'ServicePrincipal'
  }
}
```

#### Workload Identity → Key Vault (secret read)

```bicep
// AKS workload identity reads secrets from Key Vault for app credentials
resource wlKeyVaultReader 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  scope: keyVault
  name: guid(keyVault.id, workloadIdentity.properties.principalId, '4633458b-17de-408a-b874-0445c86b69e6')
  properties: {
    roleDefinitionId: resourceId('Microsoft.Authorization/roleDefinitions', '4633458b-17de-408a-b874-0445c86b69e6')
    principalId: workloadIdentity.properties.principalId
    principalType: 'ServicePrincipal'
  }
}
```

#### App → App Configuration (data read)

```bicep
// API pod reads configuration values from Azure App Configuration
resource appConfigReader 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  scope: appConfigStore
  name: guid(appConfigStore.id, workloadIdentity.properties.principalId, '516239f1-63e1-4d78-a4de-a74fb236a071')
  properties: {
    roleDefinitionId: resourceId('Microsoft.Authorization/roleDefinitions', '516239f1-63e1-4d78-a4de-a74fb236a071')
    principalId: workloadIdentity.properties.principalId
    principalType: 'ServicePrincipal'
  }
}
```

---

## Step 5 — Create a Custom Role (Last Resort Only)

Only create a custom role when **no built-in role** satisfies the requirement without granting excess permissions.

```bash
# List existing custom roles to avoid duplicates
az role definition list --custom-role-only true --query "[].roleName" --output table
```

```json
{
  "Name": "Emergency Alerts Operator",
  "IsCustom": true,
  "Description": "Can read and update alert status but cannot delete or administer the system.",
  "Actions": [
    "Microsoft.AlertsManagement/alerts/read",
    "Microsoft.AlertsManagement/alerts/changestate/action"
  ],
  "NotActions": [],
  "DataActions": [],
  "NotDataActions": [],
  "AssignableScopes": [
    "/subscriptions/<subscription-id>/resourceGroups/<resource-group>"
  ]
}
```

```bash
az role definition create --role-definition custom-role.json
```

> **Document** the custom role in an ADR. Custom roles must be reviewed during security audits.

---

## Validation Checklist

- [ ] Role was discovered using `az role definition list` — not hardcoded from memory
- [ ] Role is the narrowest built-in role that satisfies the permission requirement
- [ ] Scope is at resource level (not subscription) unless explicitly justified
- [ ] `principalType: 'ServicePrincipal'` set for managed identities (prevents RBAC propagation delays)
- [ ] `name: guid(...)` uses deterministic inputs — idempotent on re-deploy
- [ ] Bicep comment explains the assignment (resource, role name, reason)
- [ ] Role assignment verified post-deploy: `az role assignment list --scope <resource-id>`
