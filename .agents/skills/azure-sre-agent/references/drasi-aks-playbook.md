# Drasi on AKS Playbook Template

Use this template when creating an SRE Agent for Drasi workloads on AKS.
Adapt resource names, namespaces, and SLO thresholds to your environment.

## Scope

This playbook targets:

1. Drasi control and processing components on AKS
2. event ingestion and processing health
3. query/result freshness and pipeline lag
4. AKS platform dependencies impacting Drasi behavior

## Custom Agent Set

1. `drasi-incident-triage`
   - classify incident and determine likely failure domain
2. `drasi-runtime-diagnostics`
   - inspect Drasi workloads, logs, queue/lag indicators
3. `aks-platform-diagnostics`
   - inspect node/pod/network/storage/control-plane symptoms
4. `drasi-remediation-review`
   - propose and execute approved remediation steps only

## Trigger Design

### Incident plans

1. `drasi-processing-lag`
   - severity: P1/P2
   - keywords: "lag", "stale", "backlog"
   - handler: `drasi-runtime-diagnostics`
2. `drasi-query-staleness`
   - severity: P2/P3
   - keywords: "stale data", "freshness"
   - handler: `drasi-runtime-diagnostics`
3. `drasi-platform-fault`
   - severity: P1/P2
   - keywords: "pod crash", "node not ready", "dns", "ingress"
   - handler: `aks-platform-diagnostics`

### Scheduled tasks

1. 15-minute health probe:
   - backlog/lag trend
   - failed processing rate
   - pod restart anomalies
2. daily resilience report:
   - top failure signatures
   - recurring failure windows
   - capacity pressure indicators

## Diagnostic Workflow

1. Search memory for similar Drasi incidents.
2. Determine failure domain first:
   - application processing logic
   - data/source connector path
   - AKS platform/runtime path
3. Gather evidence:
   - pod status and recent events
   - workload logs and error signatures
   - latency/throughput/freshness trend
4. Correlate with recent changes:
   - deployments/revisions
   - config changes
   - cluster events
5. Produce action recommendation with risk + rollback.

## Evidence Contract

Every incident output must include:

1. impact statement
2. UTC timeline
3. confidence level
4. failing component boundary
5. remediation options (safe-first ordering)

## Prompt Starter: Drasi Runtime Check

```text
Investigate Drasi processing health in AKS namespace <namespace>:
1. Identify unhealthy pods and restart patterns.
2. Summarize top runtime errors in the last 60 minutes.
3. Estimate processing lag/freshness impact from available telemetry.
4. Correlate with recent deployment or config changes.
5. Classify likely failure domain and propose next best action.
```

## Prompt Starter: Drasi Post-Mitigation Validation

```text
Validate Drasi recovery after mitigation for 30 minutes:
1. Run checks every 1 minute for up to 30 executions.
2. Detect regressions in lag, error rate, and pod health.
3. Escalate immediately on regression with evidence.
4. On success, output final pass/fail summary with supporting data.
```

## Governance and Safety

1. Keep write actions in Review mode by default.
2. Add Stop hook approval gate for disruptive actions.
3. Require rollback path before execution.
4. Audit all write-capable tool usage.
5. For P1/P2 incidents, require full KT sections (`SA`, `PA`, `DA`, `PPA`) in outputs.

## Integration Notes

Use this with:

- [aks-containerapps-production.md](./aks-containerapps-production.md)
- [hooks-governance.md](./hooks-governance.md)
- [deployment-patterns.md](./deployment-patterns.md)
- [kt-methodology.md](./kt-methodology.md)
- [kt-templates.md](./kt-templates.md)

This template is intentionally environment-neutral and avoids hardcoded IDs.

## Bundle Mapping

Recommended bundle composition:

1. [`../bundles/base-core`](../bundles/base-core)
2. [`../bundles/aks-production`](../bundles/aks-production)
3. [`../bundles/drasi-aks-production`](../bundles/drasi-aks-production)
4. [`../bundles/governance-kt`](../bundles/governance-kt)
5. [`../bundles/connectors-observability`](../bundles/connectors-observability) (if external integrations are needed)
