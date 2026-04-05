---
name: aks-cluster-architecture
description: >-
  AKS cluster architecture decisions including networking topology, node pool strategy, identity, autoscaling, operations, multi-region, and cost optimization. Informed by The AKS Book and Azure roadmap. USE FOR: designing new AKS clusters, planning node pools, choosing CNI and outbound connectivity, implementing workload identity, reviewing cluster architecture for production readiness, or making permanent infrastructure decisions.
---

# AKS Cluster Architecture Skill

> **MUST:** Classify every AKS recommendation by change difficulty before suggesting defaults:
> ❌ **Permanent** (cluster rebuild required) | ⚠️ **Difficult** (disruptive) | ✅ **Reversible** (safe to change later).
>
> **MUST:** Use managed identity for control plane and workload access. Never use service principal secrets.
>
> **DO NOT** assume vanilla Kubernetes behavior on AKS. Prefer Azure-native integrations (Workload Identity, Managed Prometheus, NAT Gateway, Key Vault CSI).

---

## Quick Reference

| Capability | Description |
|---|---|
| Cluster Provisioning | AKS Automatic vs Standard, tier selection, API server exposure |
| Networking | CNI selection, CIDR planning, outbound connectivity, network policies |
| Node Pool Strategy | System vs user pools, VM series, spot instances, NAP/Karpenter |
| Identity & RBAC | Workload Identity, Entra ID integration, Azure RBAC for Kubernetes |
| Autoscaling | HPA, KEDA, VPA, NAP, cluster autoscaler, ScaleProfiles |
| Traffic Management | Gateway API, App Routing, Ingress deprecation, service mesh |
| Observability | Managed Prometheus, ContainerLogV2, correlation IDs, four golden signals |
| Operations | Upgrade strategy, maintenance windows, GitOps, backup/recovery |
| Multi-Region | Fleet Manager, active-active, warm standby, cold standby patterns |
| Cost Optimization | Rightsizing cadence, reserved instances, spot pools, NAP consolidation |

---

## Currency and verification gates

- **Last reviewed:** 2026-03-31
- **Sources:** The AKS Book (March 2026 edition), Azure/AKS GitHub CHANGELOG, AKS roadmap
- Verify AKS API versions: `az provider show --namespace Microsoft.ContainerService --query "resourceTypes[?resourceType=='managedClusters'].apiVersions"`
- Verify Kubernetes versions: `az aks get-versions --location <region> --output table`
- Verify add-on availability: check AKS release notes before deployment

### Deprecations and retirements

| Item | Deadline | Action |
|---|---|---|
| Kubenet CNI | March 31, 2028 | Migrate to Azure CNI Overlay |
| Ingress-NGINX (upstream) | March 2026 (upstream EOL); Microsoft bridge through November 2026 | Migrate to Gateway API via App Routing |
| ContainerLog (v1) | September 30, 2026 | Migrate to ContainerLogV2 |
| AAD Pod Identity | Deprecated | Use Workload Identity |
| Windows Server 2019 nodes | March 1, 2026 | Migrate to Windows Server 2022+ |
| Kubernetes v1.28 LTS | Deprecated | Upgrade to v1.29+ |

### Recent features (2025-2026)

- **AKS Automatic managed system node pools** (preview) with 99.9% pod readiness SLA
- **Gateway API via App Routing** (preview) using meshless Istio control plane (GatewayClass: `approuting-istio`)
- **Node Auto Provisioning** built on Karpenter Azure provider v1.7.2
- **Virtual Machine Node Pools** with multi-SKU support within a single pool
- **Fully managed GPU nodes** (preview) with automatic NVIDIA driver installation
- **Azure Linux 3.0** — default node OS for new clusters; expanded GPU support (NC A100, H100, H200)
- **Deployment Safeguards** (GA) — Azure Policy guardrails applied at cluster level
- **Image Integrity** (preview) — Notary v2 image signature verification via Azure Policy
- **Kubernetes AI Conformance** certification (DRA, gang scheduling, KAITO)
- **ACNS dual-stack** support with Cilium v1.18.6

> **MCP tooling:** Use `@Azure/aks-mcp` for cluster management, validation, and kubectl operations during architecture reviews.

---

## Permanent decisions (cluster rebuild required)

These decisions are ❌ **permanent** or extremely expensive to change. Get them right before `az aks create`.

| Decision | Impact | Recommendation |
|---|---|---|
| **Network CIDRs** | Pod CIDR, Service CIDR, DNS IP overlap causes routing conflicts across VNets, peered networks, and on-prem (VPN/ExpressRoute) | Plan CIDRs across all environments before provisioning. Document choices in an ADR. |
| **CNI type** | Changing CNI requires cluster rebuild | Default to **Azure CNI Overlay with Cilium**. Use Dynamic IP Allocation only when external routing to pod IPs is required. |
| **Availability zones** | Cannot add zones to existing cluster | Enable multi-zone (`--zones 1 2 3`) from day one for production. |
| **AKS tier (Automatic vs Standard)** | Not a toggle; different operational models | Choose based on control needs vs operational simplicity. |
| **Cluster name and resource group** | Permanent identifiers | Follow naming conventions from day one. |

### Decisions that are now reversible

| Decision | Why reversible | Notes |
|---|---|---|
| **API server public/private** | API Server VNet Integration allows toggling | Start public with authorized IP ranges; move to private later if needed. |
| **Outbound type** | Can switch between LB SNAT and NAT Gateway | Start with NAT Gateway for production (avoids SNAT exhaustion). |

---

## Networking architecture

### CNI selection (permanent)

Default to **Azure CNI Overlay with Cilium** for new clusters:
- Preserves VNet IP space (pods get overlay IPs, not VNet IPs)
- Supports 5,000 nodes and 200,000 pods per cluster
- Cilium enables eBPF-based network policies, FQDN filtering, and L7 visibility
- Kubenet is retiring March 31, 2028

Use Dynamic IP Allocation only when pods need routable VNet IPs (rare requirement).

### Outbound connectivity

- ❌ Do NOT rely on default load balancer SNAT for production
- Default to **Azure NAT Gateway** (64,512 SNAT ports per public IP vs shared pool with LB)
- Use 2+ public IPs on NAT Gateway for production (129,024 total ports)
- If centralized egress inspection is required, consider Azure Firewall/NVA and document SNAT/throughput trade-offs

**Symptom of SNAT exhaustion:** Intermittent "cannot assign requested address" errors, timeouts under load.

### CIDR planning

Never allow overlap between:
- Pod CIDR and Service CIDR
- VNet address spaces (current and peered)
- On-premises ranges (VPN/ExpressRoute)
- Other cluster CIDRs in the same network topology

### Network policies

- Start with **default-deny ingress** and explicitly allow required traffic
- With Cilium (via ACNS): FQDN filtering and L7 policies available
- Network policies operate at pod level (complement NSG subnet-level controls; do not replace them)

### DNS

- Use **Azure Private DNS Zones** for Azure service endpoints
- Enable **LocalDNS** for production (caches on each node, reduces CoreDNS load and latency)
- Minimize CoreDNS customization; bad changes break the entire cluster instantly

---

## Node pool strategy

### System pool (dedicated, predictable)

- **3 nodes minimum** (fixed, no autoscaling recommended by AKS Book)
- VM size: 4+ vCPU, 4+ GB RAM (e.g., Standard_D4s_v5)
- Taint: `CriticalAddonsOnly=true:NoSchedule` (prevents user workloads)
- Hosts CoreDNS, metrics-server, kube-proxy
- With AKS Automatic managed system pools (preview): Microsoft manages system infrastructure entirely

### Node OS

- Default to **Azure Linux 3.0** for all new node pools (smaller image, faster boot, Microsoft-maintained security patches)
- Ubuntu 24.04 as alternative when specific kernel modules or third-party packages require it
- Set via `--os-sku AzureLinux` on pool creation (permanent per pool)

### Application pools

**Option 1: Manual node pools**
- General purpose: Standard_D8s_v5 or larger
- Specialized pools for memory-intensive (E-series) or compute-intensive (F-series) workloads
- Enable autoscaling: set `minCount` to baseline load, `maxCount` with 20% buffer above observed peak

**Option 2: Node Auto Provisioning (NAP)**
- Dynamically selects optimal VM SKUs based on pending pod requirements
- Consolidation policy continuously bin-packs workloads to reduce cost (30-40% savings typical)
- Default on AKS Automatic; enable via Karpenter add-on on AKS Standard
- **Requires PodDisruptionBudgets** to avoid disruptive consolidation (`minAvailable: n-1` for n-replica services)

### Spot node pools (cost optimization)

- Use for non-critical, interruptible workloads (batch, dev/test, stateless processing)
- Pods can be evicted with 30-second notice
- Always combine with on-demand or reserved pools for critical services
- Set `scaleSetEvictionPolicy: Delete` and define `spotMaxPrice`

### VM series selection

| Workload type | VM series | Notes |
|---|---|---|
| General purpose | D-series (D4s_v5, D8s_v5) | Default choice for most workloads |
| Memory-optimised | E-series | Caches, in-memory databases |
| Compute-optimised | F-series | CPU-bound processing |
| GPU/AI/ML | NC/ND-series | NVIDIA GPU workloads; Azure Linux 3.0 supports A100, H100, H200 |
| Spot/interruptible | Any series with spot pricing | Cost savings for fault-tolerant workloads |

### Key fact: scheduling overhead

Kubernetes schedules by **resource requests**, not actual usage. System components consume approximately 2.5% CPU and 10% memory per node. A 50-node cluster loses equivalent capacity of 1-2 full nodes to overhead.

---

## Identity and access

### Cluster identity

- **Control plane:** User-assigned managed identity (pre-create, grant Network Contributor on VNet)
- **Kubelet:** System-assigned managed identity (for ACR pulls, Key Vault access)

### Workload Identity (pod-to-Azure authentication)

Microsoft Entra Workload ID replaces deprecated AAD Pod Identity:
1. Enable OIDC issuer on cluster
2. Create user-assigned managed identity for the workload
3. Create federated credential linking Kubernetes service account to the managed identity
4. Annotate service account: `azure.workload.identity/client-id: <client-id>`
5. Application uses `DefaultAzureCredential` (no secrets stored)

### User and admin access

- Enable **Microsoft Entra ID integration** (`--enable-aad`)
- **Disable local accounts** on production clusters (`--disable-local-accounts`)
- Use **Azure RBAC for Kubernetes** (4 built-in roles managed centrally in Azure: Cluster Admin, Admin, Writer, Reader)
- Create Entra groups: cluster-admins and developers

---

## Autoscaling strategy

### Scaling decision matrix

| Signal type | Scaler | When to use |
|---|---|---|
| CPU/memory utilisation | HPA | Stateless workloads with request-proportional load |
| Queue depth / event count | KEDA | Message-driven or event-driven workloads |
| Custom business metrics | KEDA or HPA (custom metrics) | Domain-specific scaling (active sessions, pending jobs) |
| Pod resource requests drift | VPA (Off/Initial mode) | Rightsizing requests based on observed usage |
| Node-level capacity | NAP/Karpenter | Heterogeneous workloads needing diverse VM SKUs |
| Node-level capacity (homogeneous) | Cluster Autoscaler | Fixed-SKU node pools with predictable workloads |

### HPA timing awareness

Scale delay from "HPA decides to scale" to "pod serving traffic" is approximately 90-120 seconds:
- Metric check interval → replica count increase → pod scheduling → image pull → container startup → probe pass → endpoint registration

### Rightsizing (continuous practice)

- Compare actual usage (P95/P99 from Prometheus/Container Insights) against `resources.requests` regularly
- Use VPA in **Off** mode to generate sizing baselines without automatic mutation
- Target request-to-usage ratio between 1.1x and 1.5x; ratios above 2x indicate significant over-provisioning
- Set resource requests to P95 of actual usage over 1 week
- CPU limits: avoid setting them (artificial throttling). Memory limits: 1.5-2x requests

---

## Traffic management

### Gateway API (recommended for new designs)

Gateway API replaces deprecated Ingress-NGINX (upstream EOL March 2026; Microsoft bridge through November 2026).

**AKS App Routing with Gateway API (preview as of March 2026):**
- Uses meshless Istio control plane (no sidecar injection)
- GatewayClass: `approuting-istio`
- AKS auto-provisions: Envoy Deployment, LoadBalancer Service, HPA (2-5 replicas), PDB
- Limitations in preview: DNS/TLS management not yet supported; no SNI passthrough

**Implementation options for AKS:**

| Option | Best for | Trade-offs |
|---|---|---|
| App Routing + Gateway API | Most workloads, minimal overhead | Preview; DNS/TLS manual |
| Application Gateway for Containers (AGC) | WAF integration, Azure-native | ~$0.25/hr + traffic cost |
| Istio service mesh | mTLS, traffic splitting, observability | Full mesh complexity |

---

## Observability

### Logging

- Migrate to **ContainerLogV2** before September 2026 retirement
- Use **Basic logs tier** for ContainerLogV2 (70-80% cheaper than analytics tier)
- Log levels: ERROR (actionable failures), WARN (handled anomalies), INFO (audit trail), DEBUG (off in production)

### Metrics

- Use **Azure Monitor managed Prometheus** (18-month retention, externally hosted)
- Avoid duplicate ingestion: use **Logs and Events preset** with managed Prometheus (not Standard preset)
- Track four golden signals: Latency (P50/P95/P99), Traffic (req/sec), Errors (rate %), Saturation (80% threshold)

### Correlation IDs

Generate at ingress, propagate through all services. Enables rapid root-cause analysis across the service chain.

---

## Operations

### Upgrade strategy

- Set `autoUpgradeProfile.upgradeChannel: stable` for automatic Kubernetes version upgrades
- Set `nodeOSUpgradeChannel: NodeImage` for node OS patches
- Configure **maintenance windows** to restrict upgrades to off-peak hours (e.g., Sunday 02:00-06:00 UTC)

### GitOps deployment

- Use **pull-based GitOps** instead of push (CI/CD pushing kubectl to cluster)
- **Flux Extension** (GA, recommended) vs Argo CD (preview, not yet production-ready on AKS)
- Benefits: single source of truth, drift prevention (reconciliation reverts manual changes within ~10 minutes), audit trail, easy rollback via `git revert`

### Backup and recovery

- **Configuration:** Git repos (GitOps eliminates need for config backup)
- **Data:** Volume snapshots (fast, RPO 4-8h) + application-level backups (consistency)
- **Infrastructure:** Terraform/Bicep (IaC enables cluster recreation in 20-40 minutes)
- **Tool:** Azure Backup for AKS (unless multi-cloud requires Velero)
- **Test recovery regularly.** Untested backups fail when you need them most.

### ACR integration

- Use managed identity: `az aks update --attach-acr <acr-name>`
- Enable ABAC for repository-level isolation between environments

### Deployment safeguards

- Enable at cluster level: `az aks update --safeguards-level Warning` (or `Enforcement` for strict mode)
- Validates pod security, resource limits, and anti-patterns against Azure Policy built-in initiative
- Use `Warning` during development to surface issues; `Enforcement` in production to block non-compliant deployments

### Image integrity

- Verify container image signatures at admission using Azure Policy + Notary v2
- Sign images in ACR with `az acr sign` (requires notation CLI and signing certificate)
- Policy blocks unsigned or tampered images from running in the cluster

---

## Multi-region strategies

### Azure Kubernetes Fleet Manager

- Centralized hub managing up to 100 member clusters across regions
- Hub cost: 2.7-4% of total multi-region spend (break-even at 2-3 member clusters)
- Enables: centralized resource propagation, orchestrated updates
- DNS load balancing and Layer 4 LB remain preview (no production SLA)

### Recovery patterns

| Pattern | Cost overhead | Recovery time | Use when |
|---|---|---|---|
| **Cold standby** | Minimal | 50-60 min | Non-critical workloads; IaC recreates cluster during incident |
| **Warm standby** | ~40% increase | 12-15 min | Pre-configured secondary with zero-replica apps |
| **Active-active** | ~2x cost | 60-90 sec | Revenue-critical workloads requiring near-zero downtime |

For revenue-critical active-active, use Azure Traffic Manager or Front Door (not Fleet Manager DNS LB, which is preview).

---

## Cost optimization

- **NAP consolidation:** 30-40% cost reduction through intelligent bin-packing (10-15% for already high-utilization clusters)
- **Reserved instances:** Use for system pool and predictable baseline capacity
- **Spot instances:** 60-80% savings for interruptible workloads
- **Log ingestion:** Basic tier for ContainerLogV2 = 70-80% cost reduction vs analytics tier
- **Managed Prometheus:** Use instead of self-hosted Prometheus at scale
- **Artifact Streaming (preview):** 50-80% faster startup for large ML/AI container images
- **Rightsizing cadence:** Monthly review of request-to-usage ratios; automate alerts when drift exceeds 2x
- **Enable AKS Cost Analysis add-on** + budgets; run monthly orphaned resource audits

---

## Pre-creation checklist

### Permanent decisions (get these right first)

- [ ] Networking: Azure CNI Overlay + Cilium (pod CIDR, service CIDR, DNS IP planned and documented)
- [ ] Zones: Multi-zone across 3 availability zones
- [ ] Identities: User-assigned managed identity for control plane (grant Network Contributor on VNet)
- [ ] API Server: VNet Integration enabled; start public with authorized IPs
- [ ] Outbound: NAT Gateway with 2+ public IPs (not load balancer SNAT)
- [ ] Cluster name and resource group follow naming conventions
- [ ] AKS tier decision documented (Automatic vs Standard)
- [ ] Node OS: Azure Linux 3.0 (`--os-sku AzureLinux`) unless Ubuntu is specifically required

### Post-creation setup

- [ ] Microsoft Entra integration enabled; local accounts disabled
- [ ] Workload Identity OIDC issuer enabled
- [ ] System node pool: 3 nodes, Standard_D4s_v5+, CriticalAddonsOnly taint, Azure Linux OS
- [ ] Application pools: NAP or manual D8s_v5+ with autoscaling
- [ ] Pod Security Standards: Baseline enforcement on namespaces
- [ ] Deployment safeguards: `Warning` level enabled (consider `Enforcement` for production)
- [ ] ContainerLogV2 + managed Prometheus configured
- [ ] GitOps deployment (Flux extension or equivalent)
- [ ] Backup extension installed + restore testing plan
- [ ] Maintenance windows configured for off-peak hours
- [ ] ACR attached via managed identity

---

## Known pitfalls

| Area | Pitfall | Mitigation |
|---|---|---|
| SNAT exhaustion | Load balancer SNAT ports cause "cannot assign requested address" under load | Use NAT Gateway with 2+ public IPs |
| HPA without requests | Pods with no CPU requests show 0% utilization; HPA will not scale them | Always set `resources.requests` on all containers |
| DNS misconfiguration | Bad CoreDNS changes break entire cluster instantly | Test in dev cluster first; have immediate rollback plan |
| Network policy vs NSG | NSGs are too coarse-grained for workload traffic rules | Use Kubernetes network policies (Cilium for L7) |
| GitOps manual overrides | Flux/ArgoCD reconciliation reverts `kubectl` changes within ~10 minutes | Use GitOps workflow for all changes; document emergency break-glass procedure |
| Untested backups | Restore failures only discovered during actual incidents | Test recovery quarterly at minimum |
| Fleet Manager preview features | DNS LB and L4 LB have no production SLA | Use Azure Traffic Manager for revenue-critical multi-region |
| Spot evictions | 30-second notice; no guarantee of graceful shutdown | Use PDBs; never run stateful or critical services on spot pools |
| System pool mixed workloads | User pods on system pool cause evictions under resource pressure | Taint system pool with CriticalAddonsOnly; run user workloads on separate pools |

---

## Principles

1. **Permanent decisions first** — Get networking, zones, and identity right before `az aks create`
2. **Managed identity everywhere** — Workload Identity for pods; managed identity for control plane and kubelet
3. **Zone redundancy by default** — Multi-zone node pools and control plane for production
4. **System pool isolation** — Dedicated, tainted system pool; separate user workloads
5. **NAT Gateway for outbound** — Prevent SNAT exhaustion before it becomes a production incident
6. **Azure CNI Overlay + Cilium** — Default networking stack for new clusters
7. **Autoscaling enabled** — Never use static node counts in production; prefer NAP for heterogeneous workloads
8. **Rightsizing is continuous** — Monthly review of request-to-usage ratios with VPA recommendations
9. **GitOps for deployments** — Pull-based with Flux; version control as single source of truth
10. **Observe from day one** — Managed Prometheus + ContainerLogV2 configured at cluster creation

---

## References

- [AKS best practices](https://learn.microsoft.com/azure/aks/best-practices)
- [AKS networking concepts](https://learn.microsoft.com/azure/aks/concepts-network)
- [Workload Identity overview](https://learn.microsoft.com/azure/aks/workload-identity-overview)
- [Node Auto Provisioning](https://learn.microsoft.com/azure/aks/node-autoprovision)
- [Gateway API on AKS](https://learn.microsoft.com/azure/aks/istio-gateway-api)
- [AKS release notes](https://github.com/Azure/AKS/releases)
- The AKS Book: The Real-World Guide to Azure Kubernetes Service (Richard Hooper, March 2026)

### MCP Tooling

- **`@Azure/aks-mcp`** — Cluster management, kubectl operations, and validation during architecture reviews
- **`drawio`** — Generate architecture diagrams (C4, network topology) via the diagram-smith agent

---

## Related skills

- **cost-optimization** — Azure-wide cost analysis, reserved instances, rightsizing framework
- **private-networking** — VNet integration, NSGs, Private Link, network isolation patterns
- **observability-monitoring** — Managed Prometheus, KQL queries, alerting, health checks
- **identity-managed-identity** — Workload Identity, RBAC, Entra ID integration
- **azure-container-apps** — Alternative to AKS for simpler containerized workloads
- **azure-defaults** — Region selection, tagging standards, naming conventions

---

## Companion instruction

This skill covers **architectural planning decisions** (cluster design, networking topology, node pools, identity). For **Kubernetes manifest editing guardrails** (YAML fields, probes, security contexts, resource requests), see the companion instruction at `.github/instructions/coding-standards/kubernetes/kubernetes-deployment-best-practices.instructions.md` which auto-attaches when editing K8s YAML files.
