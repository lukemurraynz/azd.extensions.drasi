# Project Role Assignments (Emergency Alerts)

Pre-computed RBAC assignments for service-to-service authentication in this repository.

| Principal | Target Resource | Role | Definition ID | Reason |
|-----------|----------------|------|--------------|--------|
| AKS kubelet managed identity | Azure Container Registry | AcrPull | `7f951dda-4ed3-4680-a7ca-43fe172d538d` | Pull backend/frontend images at pod startup |
| AKS workload identity | Azure Key Vault | Key Vault Secrets User | `4633458b-17de-408a-b874-0445c86b69e6` | Read DB credentials and API keys at runtime |
| AKS workload identity | Azure App Configuration | App Configuration Data Reader | `516239f1-63e1-4d78-a4de-a74fb236a071` | Read CORS origins, feature flags, and connection settings |
| AKS workload identity | Azure Maps Account | Azure Maps Data Reader | `423170ca-a8f6-4b0f-8487-9e4eb8f49bfa` | Frontend geocoding and tile requests via API backend |
| AKS workload identity | Azure SignalR Service | SignalR App Server | `420fcaa2-552c-430f-98ca-3264be4806c7` | Negotiate and broadcast real-time alert events to browser clients |
| GitHub Actions managed identity | Azure Container Registry | AcrPush | `8311e382-0749-4cb8-b61a-304f252e45ec` | Push built images to ACR in CD pipeline |
| GitHub Actions managed identity | AKS Cluster | AKS Cluster User | `4abbcc35-e782-43d8-92c5-2d3f1bd2253f` | Run `kubectl apply` for K8s manifest deployment |
| GitHub Actions managed identity | Resource Group | Contributor | `b24988ac-6180-42a0-ab88-20f7382dd24c` | `azd provision` creates Bicep-managed resources |
