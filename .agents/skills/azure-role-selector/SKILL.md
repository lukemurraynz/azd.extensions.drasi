---
name: azure-role-selector
description: >-
  Identify the least-privilege Azure RBAC role for a given set of permissions
  and generate CLI and Bicep role assignment code.
  USE FOR: choosing Azure roles, assigning RBAC permissions, finding built-in
  roles, creating custom role definitions, or troubleshooting authorization errors.
allowed-tools:
  [
    "Azure MCP/documentation",
    "Azure MCP/bicepschema",
    "Azure MCP/extension_cli_generate",
    "Azure MCP/get_bestpractices",
  ]
---

Use 'Azure MCP/documentation' tool to find the minimal role definition that matches the desired permissions the user wants to assign to an identity (If no built-in role matches the desired permissions, use 'Azure MCP/extension_cli_generate' tool to create a custom role definition with the desired permissions). Use 'Azure MCP/extension_cli_generate' tool to generate the CLI commands needed to assign that role to the identity and use the 'Azure MCP/bicepschema' and the 'Azure MCP/get_bestpractices' tool to provide a Bicep code snippet for adding the role assignment.

## Managed Identity Guidance (required)

When recommending RBAC assignments, always clarify **which managed identity** is actually used at runtime and target that identity:

- **User-assigned managed identity (UAMI)** is used when an environment variable like `AZURE_CLIENT_ID` is set or a specific client ID is configured in the workload.
- **System-assigned managed identity (SAMI)** is used only when no user-assigned identity is specified.

**Rule:** If the workload uses a UAMI, assign the role to the UAMI principal — assigning only to the SAMI will **not** fix runtime authorization errors.

### Discover Runtime Identity

Before assigning roles, confirm which identity the workload actually uses at runtime:

```bash
# Check if UAMI is configured (AZURE_CLIENT_ID set = UAMI)
az containerapp show -n <app> -g <rg> --query "identity"

# List current role assignments for the identity
az role assignment list --assignee <principal-id> --output table

# Verify token acquisition works
az account get-access-token --resource https://vault.azure.net
```

**Scope guidance:**

- Assign roles at the **narrowest scope** that satisfies the request (resource > resource group > subscription).
- For services that require both control-plane and data-plane access, check if **two roles** are required on different scopes.

When providing examples, show both:

1. how to discover the **effective identity** used by the workload, and
2. how to apply the role assignment to that identity at the correct scope.

---

## Currency and Verification

- **Date checked:** 2026-03-31
- **Compatibility:** Azure CLI, Bicep, Azure RBAC (Entra ID)
- **Sources:** [Azure built-in roles](https://learn.microsoft.com/azure/role-based-access-control/built-in-roles), [Custom role definitions](https://learn.microsoft.com/azure/role-based-access-control/custom-roles), [Managed identity overview](https://learn.microsoft.com/entra/identity/managed-identities-azure-resources/overview)
- **Verification steps:**
  1. List built-in roles: `az role definition list --query "[?roleType=='BuiltInRole'].{name:roleName, id:name}" -o table`
  2. Verify role exists: `az role definition list --name "Key Vault Secrets User" -o table`
  3. Check role assignment propagation: allow 5+ minutes after `az role assignment create`
