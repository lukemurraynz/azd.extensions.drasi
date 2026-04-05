# Action: Deploy Container App

Deploy a new Container App or update an existing one.

---

## Step 1 — Prepare Container Image

Build and push the container image to Azure Container Registry:

```bash
# Build and push
az acr build --registry $acrName --image "${serviceName}:$(git rev-parse --short HEAD)" .
```

Or use `azd` which handles this automatically:

```bash
azd deploy --service $serviceName
```

---

## Step 2 — Deploy Infrastructure (Bicep)

Deploy the Container App Environment and Container App via Bicep.
See SKILL.md for complete Bicep examples.

```bash
az deployment group create \
  --resource-group $resourceGroup \
  --template-file infra/main.bicep \
  --parameters environmentName=$envName
```

---

## Step 3 — Configure Managed Identity Access

Grant the Container App's managed identity access to required resources:

```bash
# Get the Container App's identity principal ID
principalId=$(az containerapp show --name $appName --resource-group $rg \
  --query identity.principalId -o tsv)

# Grant access to Key Vault
az role assignment create \
  --assignee $principalId \
  --role "Key Vault Secrets User" \
  --scope $keyVaultId

# Grant access to Storage
az role assignment create \
  --assignee $principalId \
  --role "Storage Blob Data Contributor" \
  --scope $storageAccountId
```

---

## Step 4 — Verify Deployment

```bash
# Check container app status
az containerapp show --name $appName --resource-group $rg \
  --query "{status:properties.runningStatus, url:properties.configuration.ingress.fqdn}"

# Check active revisions
az containerapp revision list --name $appName --resource-group $rg \
  --query "[].{name:name, active:properties.active, traffic:properties.trafficWeight}"

# Test health endpoint
curl https://${appUrl}/healthz/ready
```

---

## Step 5 — Configure Traffic Splitting (if canary)

```bash
# Route 10% to new revision
az containerapp ingress traffic set --name $appName --resource-group $rg \
  --revision-weight "${appName}--${oldRevision}=90" "${appName}--${newRevision}=10"

# Monitor for errors, then promote
az containerapp ingress traffic set --name $appName --resource-group $rg \
  --revision-weight "${appName}--${newRevision}=100"
```

---

## Completion Criteria

- [ ] Container image built and pushed to ACR
- [ ] Container App deployed and running
- [ ] Managed identity configured with appropriate role assignments
- [ ] Health endpoints returning 200
- [ ] Ingress accessible (external) or reachable from VNet (internal)
- [ ] Scaling rules active and tested
