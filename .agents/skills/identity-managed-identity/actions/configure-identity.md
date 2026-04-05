# Action: Configure Identity

Set up Managed Identity and RBAC for an Azure service.

---

## Step 1 — Enable Managed Identity

### System-Assigned (Default Choice)

```bash
# App Service
az webapp identity assign --name $appName --resource-group $rg

# Container App
az containerapp identity assign --name $appName --resource-group $rg --system-assigned

# Function App
az functionapp identity assign --name $appName --resource-group $rg
```

### User-Assigned (Shared Identity)

```bash
# Create the identity
az identity create --name "id-${serviceName}" --resource-group $rg --location $location

# Assign to a Container App
az containerapp identity assign --name $appName --resource-group $rg \
  --user-assigned "id-${serviceName}"
```

---

## Step 2 — Assign RBAC Roles

```bash
# Get the principal ID
principalId=$(az containerapp show --name $appName --resource-group $rg \
  --query identity.principalId -o tsv)

# Assign roles at resource scope
az role assignment create \
  --assignee $principalId \
  --role "Key Vault Secrets User" \
  --scope "/subscriptions/$sub/resourceGroups/$rg/providers/Microsoft.KeyVault/vaults/$kvName"

az role assignment create \
  --assignee $principalId \
  --role "Storage Blob Data Contributor" \
  --scope "/subscriptions/$sub/resourceGroups/$rg/providers/Microsoft.Storage/storageAccounts/$storageName"
```

---

## Step 3 — Disable Local Auth

```bash
# Service Bus
az servicebus namespace update --name $sbName --resource-group $rg \
  --disable-local-auth true

# Storage (disable shared key)
az storage account update --name $storageName --resource-group $rg \
  --allow-shared-key-access false

# Cosmos DB
az cosmosdb update --name $cosmosName --resource-group $rg \
  --disable-key-based-metadata-write-access true
```

---

## Step 4 — Update Application Code

Replace credential-based connections with `DefaultAzureCredential`:

```csharp
// Before (BAD)
var client = new BlobServiceClient("DefaultEndpointsProtocol=https;AccountKey=...");

// After (GOOD)
var client = new BlobServiceClient(
    new Uri("https://mystorageaccount.blob.core.windows.net"),
    new DefaultAzureCredential());
```

---

## Step 5 — Verify

```bash
# Test role assignment
az role assignment list --assignee $principalId --all \
  --query "[].{role:roleDefinitionName, scope:scope}" -o table

# Verify the app can access resources
# Deploy and test health endpoints or run integration tests
```

---

## Completion Criteria

- [ ] Managed Identity enabled on the compute resource
- [ ] RBAC roles assigned at resource scope (not subscription)
- [ ] Local auth disabled on target resources
- [ ] Application code uses `DefaultAzureCredential`
- [ ] No connection strings with embedded credentials remain
- [ ] Access verified end-to-end
