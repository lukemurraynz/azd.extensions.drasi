# Built-in Azure Role Catalog (Verified IDs — Azure Public Cloud)

> All IDs verified against `az role definition list` on Azure Public Cloud.  
> **Re-verify** with `az role definition list --query "[?name=='<id>'].roleName"` before use in any new environment.  
> Sovereign clouds (Government, China) may have different IDs.

---

## Azure Container Registry

| Role | Definition ID | Grants |
|------|--------------|--------|
| AcrPull | `7f951dda-4ed3-4680-a7ca-43fe172d538d` | Pull container images and Helm charts |
| AcrPush | `8311e382-0749-4cb8-b61a-304f252e45ec` | Push and pull container images |
| AcrDelete | `c2f4ef07-c644-48eb-af81-4b1b4947fb11` | Delete images from ACR |
| AcrImageSigner | `6cef56e8-d556-48e5-a04f-b8e64114680f` | Sign images in ACR |

**Typical assignment:** AKS kubelet managed identity → `AcrPull` on the ACR resource.

---

## Azure Kubernetes Service

| Role | Definition ID | Grants |
|------|--------------|--------|
| AKS Cluster User | `4abbcc35-e782-43d8-92c5-2d3f1bd2253f` | Access cluster via `az aks get-credentials` |
| AKS Cluster Admin | `b1ff04bb-8a4e-4dc4-8eb5-8693973ce19e` | Full cluster admin including node access |
| AKS Service Mesh Admin | `194cd0eb-af40-4d1e-bbac-57082f0f6e1b` | Manage Azure Service Mesh configuration |

---

## Azure Key Vault (RBAC model)

| Role | Definition ID | Grants |
|------|--------------|--------|
| Key Vault Administrator | `00482a5a-887f-4fb3-b363-3b7fe8e74483` | All key, secret, and certificate operations |
| Key Vault Secrets User | `4633458b-17de-408a-b874-0445c86b69e6` | Read secret values (data-plane read only) |
| Key Vault Secrets Officer | `b86a8fe4-44ce-4948-aee5-eccb2c155cd7` | Create, delete, and read secrets |
| Key Vault Crypto User | `12338af0-0e69-4776-bea7-57ae8d297424` | Encrypt and decrypt with keys |
| Key Vault Certificates Officer | `a4417e6f-fecd-4de8-b567-7b0420556985` | Manage certificates |
| Key Vault Reader | `21090545-7ca7-4776-b22c-e363652d74d2` | List key vault and its properties (no secret values) |

**Important:** Key Vault RBAC model must be enabled on the vault (`enableRbacAuthorization: true`). Access policies model uses different mechanisms.

---

## Azure Storage

| Role | Definition ID | Grants |
|------|--------------|--------|
| Storage Blob Data Contributor | `ba92f5b4-2d11-453d-a403-e96b0029c9fe` | Read, write, and delete blobs |
| Storage Blob Data Reader | `2a2b9908-6ea1-4ae2-8e65-a410df84e7d1` | Read blobs only |
| Storage Blob Data Owner | `b7e6dc6d-f1e8-4753-8033-0f276bb3955b` | Full blob access including POSIX ACL |
| Storage Queue Data Contributor | `974c5e8b-45b9-4653-ba55-5f855dd0fb88` | Read, write, and delete queue messages |
| Storage Table Data Contributor | `0a9a7e1f-b9d0-4cc4-a60d-0319b160aaa3` | Read, write, and delete table entities |

---

## Azure App Configuration

| Role | Definition ID | Grants |
|------|--------------|--------|
| App Configuration Data Reader | `516239f1-63e1-4d78-a4de-a74fb236a071` | Read configuration values (no keys or labels management) |
| App Configuration Data Owner | `5ae67dd6-50cb-40e7-96ff-dc2bfa4b606b` | Full read/write/delete on configuration data |

**Typical assignment:** API workload identity → `App Configuration Data Reader` on the store resource.

---

## Azure Database for PostgreSQL Flexible Server

| Role | Definition ID | Grants |
|------|--------------|--------|
| Contributor | `b24988ac-6180-42a0-ab88-20f7382dd24c` | Create, update, delete the server resource |

> PostgreSQL Flexible Server does not have data-plane RBAC roles. Database-level access is controlled by PostgreSQL role-based authentication (pg_hba.conf, GRANT statements). For Entra ID authentication, enable `activeDirectoryAuth` on the server and assign the `Azure AD Administrator` login.

---

## Azure Maps

| Role | Definition ID | Grants |
|------|--------------|--------|
| Azure Maps Data Reader | `423170ca-a8f6-4b0f-8487-9e4eb8f49bfa` | Read geospatial data (tiles, geocoding, routing) |
| Azure Maps Data Contributor | `8f5e0ce6-4f7b-4dcf-bddf-e6f48634a204` | Read and write geospatial data |

**Typical assignment:** Frontend SPA (via managed identity or API backend) → `Azure Maps Data Reader` on the Maps account.

---

## Azure SignalR Service

| Role | Definition ID | Grants |
|------|--------------|--------|
| SignalR App Server | `420fcaa2-552c-430f-98ca-3264be4806c7` | Connect to SignalR Service as an app server (negotiate + broadcast) |
| SignalR REST API Owner | `fd53cd77-2268-407a-8f46-7e7863d0f521` | Full access to SignalR Service REST API |
| SignalR REST API Reader | `ddde6b66-c0df-4114-a159-3618637b3035` | Read-only access to SignalR Service REST API |
| SignalR Service Owner | `7e4f1700-ea5a-4f59-8f37-079cfe29dce3` | Full management of SignalR Service |

---

## Azure Monitor / Diagnostics

| Role | Definition ID | Grants |
|------|--------------|--------|
| Monitoring Metrics Publisher | `3913510d-42f4-4e42-8a64-420c390055eb` | Publish custom metrics to Azure Monitor |
| Monitoring Reader | `43d0d8ad-25c7-4714-9337-8ba259a9fe05` | Read monitoring data and settings |
| Log Analytics Contributor | `92aaf0da-9dab-42b6-933f-5efcd4f11b15` | Create, delete, and read Log Analytics workspaces and data |
| Log Analytics Reader | `73c42c96-874c-492b-b04d-ab87d138a893` | View and query Log Analytics data |

---

## General Purpose Roles

| Role | Definition ID | Grants |
|------|--------------|--------|
| Owner | `8e3af657-a8ff-443c-a75c-2fe8c4bcb635` | Full resource management + role assignments. **Avoid.** |
| Contributor | `b24988ac-6180-42a0-ab88-20f7382dd24c` | Create and manage resources. No role assignment. |
| Reader | `acdd72a7-3385-48ef-bd42-f606fba81ae7` | View resources. No writes. |
| User Access Administrator | `18d7d88d-d35e-4fb5-a5c3-7773c20a72d8` | Manage role assignments. No resource management. |

---

## Verification Command

```bash
# Verify a specific role ID is valid in the current subscription
az role definition list \
  --query "[?name=='<definition-id>'].[roleName, name, description]" \
  --output table

# List all built-in roles for a specific service
az role definition list \
  --query "[?contains(roleName, 'AKS')].[roleName, name]" \
  --output table
```
