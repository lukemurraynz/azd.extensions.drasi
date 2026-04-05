# Cleanup Ordering

Safe deletion sequence for complex stacks to avoid orphans and errors.

---

## Dependency Graph

```
User Traffic / Clients
    ↓
Ingress (K8s)
    ↓
Services (K8s LoadBalancer)
    ↓
Pods (Deployments, StatefulSets)
    ↓
PersistentVolumeClaims (Data)
    ↓
Databases (PostgreSQL, Cosmos DB, etc.)
    ↓
AKS Cluster
    ↓
ACR (Container Registry)
    ↓
Key Vault, Storage, Networking
    ↓
Resource Group
```

---

## Safe Cleanup Sequence

**DELETE IN THIS ORDER** to avoid orphans and "resource in use" errors:

### Phase 1: Stop Traffic (1-2 mins)

1. **Stop Ingress Traffic**
   ```powershell
   kubectl delete ingress -n <your-namespace> --all
   ```
   - Prevents new connections
   - Allows graceful pod shutdown

2. **Drain Connections**
   ```powershell
   kubectl scale deployment <api-deployment-name> -n <your-namespace> --replicas 0
   kubectl scale deployment <web-deployment-name> -n <your-namespace> --replicas 0
   
   # Wait for graceful termination
   kubectl wait --for=delete pod -l app=<api-app-label> -n <your-namespace> --timeout=30s || $true
   ```
   - Allows in-flight requests to complete
   - Stops accepting new connections

### Phase 2: Delete Application Resources (2-3 mins)

3. **Delete Deployments & Services**
   ```powershell
   kubectl delete deployment -n <your-namespace> --all
   kubectl delete svc -n <your-namespace> --all
   kubectl delete configmap -n <your-namespace> --all
   kubectl delete secret -n <your-namespace> --all
   ```
   - Removes app layer
   - Releases LoadBalancer IPs
   - Cleans up configuration

4. **Delete Persistent Volumes**
   ```powershell
   kubectl delete pvc -n <your-namespace> --all
   kubectl delete pv -n <your-namespace> --all  # If not auto-deleted
   ```
   - Detaches storage
   - Allows database cleanup (if storage-backed)

5. **Delete Kubernetes Namespace**
   ```powershell
   kubectl delete namespace <your-namespace> --grace-period=30
   kubectl wait --for=delete namespace/<your-namespace> --timeout=60s || $true
   ```
   - Cascades deletion of all K8s objects
   - Cleans up LB public IPs from resource group

### Phase 3: Delete Data Layer (5-10 mins)

6. **Delete Databases** (or backup first)
   ```powershell
   # If PostgreSQL
   az postgres server delete --name "postgres-$env" --resource-group "$rg" --yes
   
   # If Cosmos DB
   az cosmosdb delete --name "cosmos-$env" --resource-group "$rg" --yes
   ```
   - Avoids orphaned data
   - Releases storage
   - **Optional**: Back up first per Step 3 of [cleanup-completely](../actions/cleanup-completely.md)

7. **Delete Managed Disks** (if not tied to VMs)
   ```powershell
   az disk list --resource-group "$rg" --query "[].name" --output tsv | ForEach-Object {
     az disk delete --name "$_" --resource-group "$rg" --yes
   }
   ```
   - Prevents "disk in use" errors
   - Reduces storage costs

### Phase 4: Delete Infrastructure (5-15 mins)

8. **Delete AKS Cluster**
   ```powershell
   # azd down handles this
   az aks delete --name "<your-aks-name>" --resource-group "$rg" --yes
   ```
   - Takes longest; do last
   - Releases compute resources
   - Deletes associated subnets

9. **Delete ACR**
   ```powershell
   az acr delete --name "<acr-name>" --resource-group "$rg" --yes
   ```
   - Deletes all images
   - Releases storage
   - Allows resource group deletion

10. **Delete Remaining Services** (Key Vault, Storage, Networking)
    ```powershell
    # Usually cascade-deleted by resource group deletion
    # If needed manually:
    az keyvault delete --name "kv-$env" --resource-group "$rg" --yes
    az storage account delete --name "sa$env" --resource-group "$rg" --yes
    ```

### Phase 5: Delete Resource Group (1-2 mins)

11. **Delete Resource Group** (final cleanup)
    ```powershell
    az group delete --name "$rg" --yes
    ```
    - Deletes all remaining resources
    - Cascades to all child resources
    - Clears Azure billing

---

## Why This Order?

| Phase | Why Delete in This Order |
|-------|--------------------------|
| **Stop Traffic** | Prevents "resource in use" errors; allows graceful shutdown; shorter pod termination grace period |
| **App Resources** | Releases LoadBalancer IPs; frees network; simplifies storage detach |
| **Persistent Data** | Databases can be large/slow to delete; depends on being detached from pods first |
| **Infrastructure** | AKS deletion takes longest; do last; depends on subnets being cleared first |
| **ACR, Key Vault** | No pods depend on these after deletion; safe to delete anytime after pods gone |
| **Resource Group** | Final catch-all; cascades remaining stragglers |

---

## Automated Cleanup (Using azd)

`azd down` handles much of this, but may fail if:
- K8s resources not deleted first
- Database has active connections
- Resource locks prevent deletion

**Safe sequence with azd**:

```powershell
# Optional: Manual prep (if needed)
kubectl delete namespace <your-namespace> --force --grace-period=0  # If K8s stuck

# Primary cleanup
azd down --no-prompt    # Attempts full deletion

# If azd down fails:
# 1. Check remaining resources: az resource list --resource-group "$rg" --output table
# 2. Delete stuck resources manually (see Phase 1-4 above)
# 3. Retry: az group delete --name "$rg" --yes
```

---

## Partial Failure Recovery

**Scenario: `azd down` fails partway through**

Remaining resources by typical failure point:

| Failure Point | Remaining | Recovery |
|---|---|---|
| K8s cluster deletion | AKS, subnets, NICs | `az aks delete --name ... --yes`; retry resource group delete |
| Database deletion | PostgreSQL, snapshots, backups | Check locks: `az lock list --resource-group "$rg"`; manual delete |
| Storage account deletion | Blobs, file shares, queues | `az storage account delete --name ... --yes --force-delete-contained-namespaces` |
| Generic partial failure | Unspecified | See [troubleshoot-failures](../actions/troubleshoot-failures.md) Step 9 |

---

## Verification After Cleanup

```powershell
# Should return empty or not-found
az resource list --resource-group "$rg" --output table

# Should return error: not found
az group show --name "$rg"

# Should show 0 resources
az resource count --resource-group "$rg"
```

---

## Cost Impact

Cleanup cost in order:
1. **AKS cluster** (largest) — compute stops immediately
2. **Databases** (large) — storage freed after backup finalized
3. **ACR, Storage, Key Vault** (small) — minimal cost
4. **Networking, IPs** (tiny) — negligible

**Expect billing adjustment**: 24-48 hours for final reconciliation in Azure Cost Analysis.

---

## Checklist for Safe Cleanup

- [ ] Backup critical data (Step 3 of [cleanup-completely](../actions/cleanup-completely.md))
- [ ] Delete K8s namespace first: `kubectl delete namespace ...`
- [ ] Verify K8s deletion: `kubectl get namespace ...` (should be not-found)
- [ ] Run `azd down --no-prompt`
- [ ] Monitor resource group: `az group show` (wait 2-5 mins for final deletion)
- [ ] Verify no orphans: `az resource list --resource-group "$rg"` (should be empty)
- [ ] Check for resource locks: `az lock list --resource-group "$rg"` (should be empty)
- [ ] Confirm billing stops: Azure Cost Analysis (may take 24 hrs to update)
