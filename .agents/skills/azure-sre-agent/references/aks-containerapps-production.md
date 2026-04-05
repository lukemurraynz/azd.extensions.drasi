# AKS and Container Apps Production Playbook

Use this guide when building SRE Agent workflows for AKS and Container Apps.

## Objectives

1. Reduce mean time to detect and resolve.
2. Separate diagnosis from remediation.
3. Keep production actions auditable and approval-aware.
4. Preserve rollback paths for every write action.

## Recommended Custom Agent Topology

1. `aks-diagnostics`
   - node health, pod states, cluster events, control-plane symptoms
2. `containerapps-diagnostics`
   - revision comparison, app logs, latency/error trends
3. `remediation-executor`
   - write actions only after explicit policy checks/approval
4. `incident-notifier`
   - stakeholder updates (email/Teams/ticket sync)

## Tool Assignment Guidance

1. Keep read tools with diagnostics agents.
2. Keep write tools isolated in remediation agent.
3. Add source connector tools only to code-analysis agents.
4. Add notification tools only to notifier agents.

## Response Plan Design

### AKS examples

1. `aks-critical`
   - filter: P1/P2, service contains AKS app domain
   - handler: `aks-diagnostics`
   - mode: Review (default), later Autonomous for proven low-risk actions
2. `aks-capacity-degradation`
   - filter: node pressure and scheduling failures
   - handler: `aks-diagnostics`
   - mode: Review

### Container Apps examples

1. `ca-http-5xx`
   - filter: high error rate alerts
   - handler: `containerapps-diagnostics`
   - mode: Review
2. `ca-revision-regression`
   - filter: post-deployment latency/error spikes
   - handler: `containerapps-diagnostics`
   - mode: Review

## Scheduled Task Baselines

1. AKS daily health:
   - node readiness
   - CrashLoopBackOff/Pending pods
   - top error signatures by namespace
2. Container Apps revision checks:
   - compare current vs previous revision metrics
   - detect sustained degradation windows
3. Weekly drift/security check:
   - RBAC and workload identity sanity
   - exposed endpoints and ingress posture

## Evidence Minimum for Production Actions

Require these before remediation:

1. impact statement (who is affected)
2. timestamped evidence (metrics/logs, UTC)
3. confidence level and known uncertainty
4. rollback strategy and expected side effects

## Example Incident Summary Template

```text
Summary:
- Service: <service>
- Environment: <env>
- Severity: <sev>
- User Impact: <impact>

Evidence (UTC):
- Metrics: <key trend>
- Logs: <top failures>
- Platform state: <cluster/revision symptoms>

Root Cause Hypothesis:
- <hypothesis>

Recommended Action:
- <action>
- Risk: <risk>
- Rollback: <rollback>
```

## Governance Defaults

1. Add approval-gate hooks for disruptive AKS/Container Apps actions.
2. Keep high-risk actions in Review mode.
3. Emit audit trace for all write-capable tools.
4. For P1/P2 incidents, require KT sections (`SA`, `PA`, `DA`, `PPA`) in final output.

## KT Overlay (Recommended for Production)

1. SA: prioritize cluster/workload/dependency concerns.
2. PA: isolate true cause with `IS / IS NOT`.
3. DA: compare remediation alternatives with `MUST` and weighted `WANT`.
4. PPA: define preventive and contingent actions plus triggers.

Use [kt-methodology.md](./kt-methodology.md) and [kt-templates.md](./kt-templates.md).

## Bundle Mapping

Use these bundles for implementation:

1. [`../bundles/base-core`](../bundles/base-core)
2. [`../bundles/aks-production`](../bundles/aks-production)
3. [`../bundles/containerapps-production`](../bundles/containerapps-production)
4. [`../bundles/governance-kt`](../bundles/governance-kt)
5. [`../bundles/connectors-observability`](../bundles/connectors-observability) (optional, integration-driven)

## Sources

- Custom agents: https://learn.microsoft.com/en-us/azure/sre-agent/sub-agents
- Workflow automation: https://learn.microsoft.com/en-us/azure/sre-agent/workflow-automation
- Scheduled tasks: https://learn.microsoft.com/en-us/azure/sre-agent/scheduled-tasks
- Official sample repo: https://github.com/microsoft/sre-agent
