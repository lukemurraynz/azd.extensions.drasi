---
applyTo: "**/k8s/**/*.yml,**/k8s/**/*.yaml,**/kubernetes/**/*.yml,**/kubernetes/**/*.yaml,**/helm/**/templates/**/*.yml,**/helm/**/templates/**/*.yaml"
description: "Comprehensive AKS-first best practices for Kubernetes manifests and deployment operations. Covers controllers, networking, probes, resources, scaling, observability, and security."
---

# Kubernetes (AKS-First) Deployment Best Practices

## Your Mission

As GitHub Copilot, you are an expert in **production Kubernetes on AKS**. Your mission is to guide developers in crafting optimal manifests and operational patterns that are **reliable, secure, cost-aware, and scalable**. Prioritize decisions that are safe at scale and **call out decisions that are hard to reverse**.

### Authority and Conflicts

- Follow repository standards first (linting, policies, ADRs, platform guardrails).
- If guidance conflicts, prefer (in order): **repo policy** → **platform policy** → **this file**.
- If the user asks for a quick POC, still keep security defaults, but you may relax cost/HA complexity _explicitly and intentionally_.

---

## 0. Non-Negotiable Guardrails

### 0.1 Production vs POC Defaults

- **Production**: optimize for resilience, secure-by-default, predictable operations, and safe rollouts.
- **POC**: optimize for learning speed while keeping secure baselines (no public admin endpoints, no plaintext secrets, no privileged containers).

### 0.2 AKS Architecture Decisions

For cluster-level architecture decisions (networking topology, node pool strategy, identity, CNI selection, outbound connectivity, availability zones, and other permanent infrastructure choices), invoke the **aks-cluster-architecture** skill. This instruction covers **manifest-level guardrails** for Kubernetes YAML files.

---

## 1. Pods

### Principle

Pods are the smallest deployable unit. In production, Pods should be managed by controllers (Deployments/StatefulSets/DaemonSets/Jobs).

### Guidance for Copilot

- Design Pods for **one primary container** plus tightly-coupled sidecars only.
- Always define `resources.requests` (and justify `limits`).
- Implement health checks (`readinessProbe`, `livenessProbe`, and `startupProbe` where appropriate).
- Prefer `securityContext` hardening (non-root, no privilege escalation, drop capabilities).

### Pro Tip

Avoid deploying Pods directly; always use higher-level controllers.

---

## 2. Deployments

### Principle

Deployments manage stateless workloads with rolling updates/rollbacks.

### Guidance for Copilot

- Use Deployments for stateless apps.
- Always define:
  - `replicas` (or use HPA with a sensible baseline)
  - `selector.matchLabels` that exactly matches `template.metadata.labels`
  - rolling update strategy: `maxSurge` / `maxUnavailable`
- Always add:
  - `revisionHistoryLimit` (e.g., 3–10)
  - `terminationGracePeriodSeconds`
  - `preStop` hook if the app needs graceful shutdown

### Update Strategy Defaults (Production)

- Default to rolling updates with conservative values:
  - `maxUnavailable: 0` for critical services
  - `maxSurge: 1` (or a small percentage for large replicas)

---

## 3. Services

### Principle

Services expose a stable endpoint for a set of Pods.

### Guidance for Copilot

- Prefer `ClusterIP` for internal services.
- Use `LoadBalancer` only when required; prefer routing through a gateway/ingress layer.
- Avoid exposing the same app via both Ingress/Gateway and `LoadBalancer` Services. If multiple public entry points are required, document the justification and controls.
- Ensure selectors match Pod labels and that ports are correctly named for clarity.

### 3.1 Realtime/SignalR Services (Non-Negotiable for Multi-Replica)

- For SignalR or other realtime hubs, assume connection establishment spans multiple requests (`/negotiate` + transport).
- If using negotiate/SSE/LongPolling, require sticky routing (for example, `sessionAffinity: ClientIP` at Service, or equivalent ingress affinity/session policy).
- If sticky routing is not available or not acceptable, prefer a managed realtime backplane/service (for example, Azure SignalR Service) instead of direct pod fan-out.
- Do not configure separate public endpoints for API and realtime traffic unless both endpoints are validated from outside the cluster (DNS, TCP reachability, negotiate `200`, and transport success).

---

## 4. Traffic Management (Ingress, Gateway API, Mesh)

### Guidance for Copilot

- Prefer **Gateway API** for new designs when supported by the platform and controller.
  - ⚠️ **AKS note (as of 2026-03-17):** Managed Gateway API CRDs on AKS are **preview** — require `aks-preview` extension 19.0.0b4+, `ManagedGatewayAPIPreview` feature flag, and a supported implementation (e.g., Istio add-on asm-1-26+). Use for dev/test; evaluate readiness before production adoption. See [Managed Gateway API on AKS](https://learn.microsoft.com/azure/aks/managed-gateway-api).
- Use Ingress only when constrained by existing controllers or platform standards.
- When adopting Gateway API, treat **GatewayClass + Gateway + HTTPRoute** as the baseline model.
- Plan TLS ownership explicitly: certificates terminate on the **Gateway**, which may change per-team self-service and centralize responsibility.
- Select a controller intentionally; Gateway API behavior and supported features vary by controller implementation.
- Prefer TLS everywhere; avoid plaintext HTTP in production.
- Avoid controller-specific annotations unless required and documented.
- If using a service mesh, only recommend it when justified (mTLS, retries/timeouts/traffic splitting, observability), and acknowledge complexity trade-offs.
- For realtime/WebSocket workloads, explicitly validate controller support and timeout behavior for upgrade requests before recommending topology changes.

---

## 5. Configuration and Secrets

### 5.1 ConfigMaps

- Use ConfigMaps for non-sensitive config.
- Prefer mounting as files for complex config, env vars for simple flags.
- Never store secrets in ConfigMaps.
- Do not ship security-bypass flags (for example, `Auth__AllowAnonymous=true`) in production ConfigMaps.

### 5.2 Secrets

- Use Kubernetes Secrets for sensitive data but treat them as baseline—not the final answer for production.
- Avoid injecting secrets as env vars when possible (they can leak via process dumps and tooling).
- Prefer mounting secrets as files or using external secret integration.

### 5.3 External Secrets and Azure Key Vault (AKS)

- For production on AKS, prefer **Azure Key Vault CSI Driver** or **External Secrets Operator** (as per platform standards).
- Prefer **Workload Identity** over legacy identity mechanisms.

---

## 6. Health Checks and Probes (Operationally Correct)

### Probe Semantics (Non-Negotiable)

- **Readiness probes matter most for safe rollouts.**
- **Liveness probes MUST NOT check external dependencies** (DB, downstream APIs).
- Readiness probes MAY validate critical dependencies if it reflects “can serve traffic.”
- Use `startupProbe` for slow-start apps to avoid restart loops.

### Recommended Defaults

- Avoid overly aggressive probing that causes restart storms.
- Include sensible timeouts and thresholds:
  - `timeoutSeconds: 1–5`
  - `periodSeconds: 5–15`
  - `failureThreshold` and `initialDelaySeconds` tuned to app behavior

---

## 7. Resource Management (Requests vs Limits)

### 7.1 Requests (Mandatory)

- Always define CPU and memory **requests** for every container.
- Requests drive scheduling, HPA behavior, and capacity planning.

### 7.2 Limits (Use Carefully)

- Memory limits can cause OOMKills; apply only with understanding and monitoring.
- Avoid CPU limits unless required; they can throttle performance unexpectedly.
- If you recommend limits, explain the trade-off and the monitoring required.

### 7.3 QoS Awareness

- Prefer Burstable/Guaranteed via correctly set requests/limits.
- Avoid BestEffort in production.

### 7.4 Rightsizing (Continuous Practice)

- Treat rightsizing as an ongoing operational practice, not a one-time task.
- Compare actual usage (P95/P99 CPU and memory from Prometheus / Container Insights) against `resources.requests` regularly.
- Use VPA in **Off** (recommendation-only) mode to generate sizing baselines without automatic mutation.
- Target a request-to-usage ratio between 1.1x and 1.5x; ratios above 2x indicate significant over-provisioning.
- Resize node pools to match workload profiles: memory-optimised (E-series) for caches, compute-optimised (F-series) for CPU-bound, general purpose (D-series) as default.
- Incorporate rightsizing reviews into sprint or operational cadences; automate alerts when request-to-usage drift exceeds thresholds.

---

## 8. Scaling and Autoscaling

### 8.1 Horizontal Pod Autoscaler (HPA)

- Prefer HPA for stateless workloads with variable load.
- Ensure `resources.requests` are set; HPA percentage targets depend on them.
- Define sensible `minReplicas` / `maxReplicas` and scaling behaviour to avoid flapping.
- Use `behavior.scaleDown.stabilizationWindowSeconds` (default 300s) to prevent premature scale-down.
- Prefer custom or external metrics (request latency, queue depth) over raw CPU when they better represent demand.
- When combining HPA with VPA, run VPA in **Off** or **Initial** mode only; Auto mode and HPA conflict on the same resource dimension.

### 8.2 KEDA (Event-Driven Autoscaling)

- Use KEDA for workloads where CPU/memory utilisation is not the right scaling signal.
- KEDA provides 70+ built-in scalers: Azure Service Bus, Event Hub, Azure Monitor, Prometheus, Redis, Kafka, and more.
- On AKS, prefer the **KEDA add-on** (`az aks update --enable-keda`) for managed lifecycle and support.
- KEDA creates and manages HPA objects automatically; do not define a separate HPA for the same Deployment.
- Architecture: Event source → KEDA ScaledObject → KEDA controller creates HPA → KEDA metrics adapter exposes external metrics.
- Configure `pollingInterval`, `cooldownPeriod`, and `minReplicaCount` / `maxReplicaCount` per ScaledObject.
- Scale-to-zero is supported for queue/event workloads; set `minReplicaCount: 0` and configure `idleReplicaCount` if a warm standby is needed.

### 8.3 Vertical Pod Autoscaler (VPA)

- Use VPA to right-size pod resource requests over time.
- VPA modes:
  - **Off**: Generates recommendations only (safe for observation; no mutations).
  - **Initial**: Sets requests at pod creation; does not update running pods.
  - **Auto**: Evicts and recreates pods to apply new requests (disruptive; use with caution).
- Start with **Off** mode in production to collect recommendations before enabling mutations.
- VPA and HPA must not target the same resource dimension (CPU or memory) simultaneously.
- Review VPA recommendations as part of regular rightsizing cadence.

### 8.4 Node Auto Provisioning (NAP) / Karpenter

- NAP is the AKS-managed implementation of Karpenter for automatic node provisioning and bin-packing.
- NAP selects right-sized VM SKUs from configured families (e.g., D-series general purpose, E-series memory-optimised) based on pending pod requirements.
- Consolidation policy continuously bin-packs workloads onto fewer nodes to reduce cost and fragmentation.
- Use `NodePool` CRDs to define constraints: allowed instance families, availability zones, capacity type (on-demand vs spot), and taints/labels.
- NAP is the default node scaling mechanism on **AKS Automatic**. On **AKS Standard**, enable via the Karpenter add-on.
- Prefer NAP over cluster autoscaler for heterogeneous workloads that benefit from VM SKU diversity.

### 8.5 VM Node Pool ScaleProfiles (GA)

- ScaleProfiles allow multiple same-family VM SKUs within a single node pool for greater capacity flexibility and resilience.
- If the primary SKU is capacity-constrained in a zone, AKS falls back to alternative SKUs in the profile.
- Define ScaleProfiles via `az aks nodepool add --scale-profile` or Bicep/ARM properties.

### 8.6 VM Node Pool Autoscaling and SKU Flexibility (Preview)

- Single-SKU node pools support standard cluster autoscaler autoscaling.
- Mixed-SKU node pools (via ScaleProfiles) currently support manual scaling; autoscaling support is in public preview.
- Mixed-SKU pools improve capacity allocation reliability and cost optimisation across availability zones.
- Reference: [https://aka.ms/aks/vm-node-pool](https://aka.ms/aks/vm-node-pool)

### 8.7 Scaling Decision Matrix

| Signal Type | Scaler | When to Use |
|---|---|---|
| CPU / Memory utilisation | HPA | Stateless workloads with request-proportional load |
| Queue depth / Event count | KEDA | Message-driven or event-driven workloads |
| Custom business metrics | KEDA or HPA (custom metrics) | Domain-specific scaling (e.g., active sessions, pending jobs) |
| Pod resource requests drift | VPA (Off / Initial) | Rightsizing requests based on observed usage |
| Node-level capacity | NAP / Karpenter | Heterogeneous workloads needing diverse VM SKUs |
| Node-level capacity (homogeneous) | Cluster Autoscaler | Fixed-SKU node pools with predictable workloads |

---

## 9. Availability: Pod Disruption Budgets (Mandatory for Production)

### Guidance for Copilot

- For any production Deployment with `replicas > 1`, define a **PodDisruptionBudget**.
- Ensure PDB won’t block upgrades/node drains:
  - Avoid setting `minAvailable` too high
  - Ensure enough replicas to satisfy PDB during rollouts

---

## 10. Scheduling and Failure Domains

### Guidance for Copilot

- Use **topology spread constraints** to distribute pods across zones/nodes where applicable.
- Use anti-affinity for critical replicas when necessary, but prefer topology spread constraints for simplicity.
- Use `nodeSelector`, `tolerations`, and `affinity` intentionally and document why.
- Separate system and workload concerns via node pools (AKS standard practice).

### 10.2 Configurable Scheduler Bin Packing Profiles

- Kubernetes supports scheduler scoring plugins to influence how pods are placed across nodes.
- The `NodeResourcesFit` plugin with `RequestedToCapacityRatio` scoring strategy enables bin-packing to maximise node utilisation and reduce cost.

#### Example Scheduler Profile Configuration

```yaml
apiVersion: kubescheduler.config.k8s.io/v1
kind: KubeSchedulerConfiguration
profiles:
  - schedulerName: default-scheduler
    plugins:
      score:
        enabled:
          - name: NodeResourcesFit
            weight: 10
    pluginConfig:
      - name: NodeResourcesFit
        args:
          apiVersion: kubescheduler.config.k8s.io/v1
          kind: NodeResourcesFitArgs
          scoringStrategy:
            type: RequestedToCapacityRatio
            resources:
              - name: cpu
                weight: 8
              - name: memory
                weight: 1
            requestedToCapacityRatio:
              shape:
                - utilization: 0
                  score: 0
                - utilization: 30
                  score: 8
                - utilization: 50
                  score: 10
                - utilization: 85
                  score: 10
                - utilization: 90
                  score: 5
                - utilization: 100
                  score: 0
```

#### Bin Packing Guidance

- **Resource weights** control which dimension the scheduler prioritises for packing. A CPU weight of 8 vs memory weight of 1 strongly favours filling CPU capacity first.
- **Utilisation curve shape** defines how desirable a node is at a given utilisation level. The example above peaks at 50–85% utilisation and drops off sharply above 90% to avoid hot-spotting.
- Trade-off: aggressive bin-packing improves cost efficiency but increases blast radius if a packed node fails. Balance packing with PodDisruptionBudgets and topology spread constraints.
- On **AKS**, custom scheduler profiles require a self-managed scheduler sidecar or an AKS feature flag (check current AKS documentation for managed scheduler profile support). NAP/Karpenter consolidation provides an alternative bin-packing approach at the node lifecycle level.
- Validate scoring configuration in non-production clusters before applying to production.

---

## 11. Networking

For AKS networking architecture decisions (CNI selection, CIDR planning, outbound connectivity, NAT Gateway), invoke the **aks-cluster-architecture** skill.

### 11.1 Network Policies

- Recommend default-deny policies with explicit allow rules.
- Keep rules minimal and documented (traffic patterns, namespaces, ports).
- Avoid broad egress like `0.0.0.0/0` in “restricted” policies; if required, treat it as open egress and document the risk.

---

## 12. Security Best Practices

### 12.1 Pod and Container Security Context (Default Hardened)

- `runAsNonRoot: true`
- `allowPrivilegeEscalation: false`
- `readOnlyRootFilesystem: true` where possible
- Drop capabilities: `capabilities.drop: ["ALL"]`
- Set `seccompProfile: RuntimeDefault` when available
- Avoid `hostNetwork`, `hostPID`, `hostIPC` unless explicitly required and reviewed

### 12.2 RBAC (Least Privilege)

- Use dedicated ServiceAccounts per workload where needed.
- Prefer minimal Roles over broad ClusterRoles.
- Regularly review and prune RBAC bindings.

### 12.3 Image Security

- Avoid `:latest` tags; use immutable tags or digests.
- Prefer minimal base images (distroless where possible).
- Integrate vulnerability scanning in CI and define action thresholds.

### 12.4 Admission / Policy

- Prefer policy-as-code enforcement (e.g., Gatekeeper/Kyverno) where adopted.
- Align to Pod Security Standards (baseline/restricted) and platform policies.

---

## 13. Observability (AKS-Realistic)

### Guidance for Copilot

- Logs: write to `STDOUT`/`STDERR`; include correlation IDs.
- Metrics: expose Prometheus metrics where possible; keep label cardinality under control.
- Alerts: favor symptom-based alerting (latency, errors, saturation, traffic), not noisy signals.
- Always include enough structured context in logs for incident response.

---

## 14. Deployment Patterns: CI/CD and GitOps

### Guidance for Copilot

- Prefer declarative deployments with version control as the source of truth.
- For GitOps:
  - Keep environment overlays clear (kustomize/helm/overlays) and avoid branch-per-env unless required.
  - Ensure image update strategy is explicit and auditable.

---

## 15. AKS Operational Guardrails

For AKS cluster-level operational decisions (tier selection, identity, availability zones, upgrade strategy, GitOps, multi-region), invoke the **aks-cluster-architecture** skill.

---

## 16. Kubernetes Manifest Review Checklist

- [ ] Are `apiVersion` and `kind` correct?
- [ ] Are `metadata.name`, `labels`, and `selectors` consistent?
- [ ] Are `resources.requests` set for all containers?
- [ ] Are probes correctly configured (readiness semantics correct; no dependency checks in liveness)?
- [ ] Is rolling update strategy safe (`maxUnavailable` / `maxSurge`)?
- [ ] Is a PDB present for production workloads with replicas > 1?
- [ ] Are secrets in Secrets/external managers (never ConfigMaps)?
- [ ] Are security contexts hardened (non-root, no privilege escalation, drop caps)?
- [ ] Are NetworkPolicies considered and documented?
- [ ] Are image tags immutable (no `:latest`) and pull policy appropriate?
- [ ] Is observability in place (structured logs, metrics, correlation IDs)?
- [ ] Are scheduling constraints intentional (topology spread, affinity, tolerations)?
- [ ] Is autoscaling configured (HPA, KEDA, or NAP/Karpenter) with sensible min/max?
- [ ] Are VPA recommendations reviewed and rightsizing tracked as an operational practice?
- [ ] If using custom scheduler bin packing profiles, are they validated in non-production first?
- [ ] Are VM Node Pool ScaleProfiles configured for capacity resilience where applicable?

---

## 17. Troubleshooting Playbook (Fast Triage)

### 17.1 Pods Pending / CrashLoopBackOff

- `kubectl describe pod` → events
- `kubectl logs` (and previous logs)
- Check requests too high/low, image pull errors, missing config/secrets, probe misconfiguration

### 17.2 Pods Not Ready / Service Unavailable

- Validate readiness probe logic and timings
- Confirm the app listens on expected port inside the container
- Check endpoints for the Service (`kubectl get endpoints`)

### 17.3 Service Not Accessible

- Verify Service selectors and labels match
- Validate NetworkPolicies aren’t blocking traffic
- If using gateway/ingress, check controller/gateway logs and resource status

### 17.4 OOMKilled / Resource Exhaustion

- Validate memory requests and limits
- Check memory leak patterns / load-driven spikes
- Consider VPA recommendations and right-sizing

### 17.5 Intermittent Outbound Failures (AKS)

- Suspect SNAT exhaustion first under load:
  - timeouts, “cannot assign requested address”, inconsistent failures
- Validate outbound model (NAT Gateway vs LB vs firewall/NVA)
