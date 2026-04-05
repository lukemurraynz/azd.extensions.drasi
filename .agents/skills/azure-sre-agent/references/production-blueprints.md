# Production Blueprints (Full Capability)

Use these blueprint patterns to create practical Azure SRE Agents that use the
platform end-to-end rather than only partial features.

## Blueprint A: Core Production Baseline

### Goal

Deploy a production-ready agent with safe automation, diagnostics, and governance.

### Required Building Blocks

1. Incident platform connected (single active platform).
2. Response plans with explicit filters and ownership.
3. Scheduled tasks for proactive monitoring.
4. Custom agent topology with role separation:
   - diagnostics
   - remediation
   - notification
5. Connectors for source, ticketing, and collaboration as needed.
6. Hook controls for approval and audit.
7. KT enforcement for major incidents and production write paths.

### Minimum Acceptance Criteria

1. Historical-incident tests pass.
2. "Run task now" tests pass.
3. Connector health is green for all required integrations.
4. At least one rollback path is documented and tested.
5. P1/P2 response includes full KT sections.

### Bundle Mapping

- [`../bundles/base-core`](../bundles/base-core)
- [`../bundles/governance-kt`](../bundles/governance-kt)

## Blueprint B: AKS Production Agent

### Custom Agents

1. `aks-diagnostics`
2. `aks-remediation-review`
3. `incident-notifier`

### Trigger Set

1. Incident plans:
   - node not ready
   - scheduling pressure
   - workload crash loops
2. Scheduled tasks:
   - periodic cluster health scan
   - post-mitigation stability validation

### Required Data/Evidence

1. node/pod states and event timeline
2. workload error signatures
3. change correlation (deployments/config)
4. impact and blast-radius assessment

### Bundle Mapping

- [`../bundles/base-core`](../bundles/base-core)
- [`../bundles/aks-production`](../bundles/aks-production)
- [`../bundles/governance-kt`](../bundles/governance-kt)

## Blueprint C: Container Apps Production Agent

### Custom Agents

1. `containerapps-diagnostics`
2. `revision-health-analyzer`
3. `remediation-review`
4. `incident-notifier`

### Trigger Set

1. incident plans:
   - elevated 5xx
   - revision regression
2. scheduled tasks:
   - revision health comparison
   - security/compliance drift checks

### Required Data/Evidence

1. current vs prior revision metrics
2. revision logs and failure signatures
3. confidence-scored root cause hypothesis
4. rollback plan before execution

### Bundle Mapping

- [`../bundles/base-core`](../bundles/base-core)
- [`../bundles/containerapps-production`](../bundles/containerapps-production)
- [`../bundles/governance-kt`](../bundles/governance-kt)

## Blueprint D: Drasi on AKS Production Agent

### Custom Agents

1. `drasi-incident-triage`
2. `drasi-runtime-diagnostics`
3. `aks-platform-diagnostics`
4. `remediation-review`

### Trigger Set

1. incident plans:
   - processing lag
   - query staleness
   - platform instability
2. scheduled tasks:
   - 15-minute health probe
   - post-mitigation validation (30-minute bounded loop)

### Required Data/Evidence

1. Drasi runtime logs + lag/freshness indicators
2. AKS platform state and resource pressure
3. dependency failure signals
4. KT-based decision rationale for production changes

### Bundle Mapping

- [`../bundles/base-core`](../bundles/base-core)
- [`../bundles/aks-production`](../bundles/aks-production)
- [`../bundles/drasi-aks-production`](../bundles/drasi-aks-production)
- [`../bundles/governance-kt`](../bundles/governance-kt)
- [`../bundles/connectors-observability`](../bundles/connectors-observability) (as needed)

## Capability Coverage Checklist

Use this to ensure full platform use:

1. Incident platform configured and tested.
2. Response plans lifecycle implemented (create/disable/test/re-enable).
3. Scheduled tasks implemented with bounded execution.
4. Custom agent handoff topology deployed.
5. Connector portfolio deployed and healthy.
6. Hooks deployed:
   - approval gate
   - audit gate
   - KT completeness gate
7. KT templates operationalized in major incidents.
8. Security posture in place (least privilege + token rotation).
9. Verification report generated after setup.

## Practical Build Sequence (Operator Runbook)

Use this sequence to stand up a useful production agent quickly:

1. Provision agent and baseline RBAC.
2. Connect incident platform and remove quickstart overlap.
3. Create diagnostics custom agents (AKS/Container Apps/Drasi as needed).
4. Create remediation-review custom agent with write tools isolated.
5. Create notifier custom agent and test outbound channels.
6. Add response plans and test against historical incidents.
7. Add scheduled tasks and run "Run task now" tests.
8. Apply hooks:
   - remediation approval
   - tool audit
   - KT completeness gate
9. Run acceptance tests and publish readiness summary.

## Copy-Ready Custom Agent Stubs

### Diagnostics (read-first)

```yaml
api_version: azuresre.ai/v1
kind: AgentConfiguration
spec:
  name: diagnostics-agent
  system_prompt: |
    Diagnose incidents with evidence-first workflow.
    Collect logs, metrics, and timeline in UTC before proposing actions.
  handoff_description: Evidence-first diagnostics for production incidents
  tools:
    - QueryLogAnalyticsByResourceId
    - QueryAppInsightsByResourceId
    - ListAvailableMetrics
    - GetMetricsTimeSeriesAnalysis
    - RunAzCliReadCommands
```

### Remediation Review (write-isolated)

```yaml
api_version: azuresre.ai/v1
kind: AgentConfiguration
spec:
  name: remediation-review
  system_prompt: |
    Propose remediation with blast-radius and rollback.
    Execute only when mode, permission, and policy gates are satisfied.
  handoff_description: Controlled remediation execution with rollback discipline
  tools:
    - RunAzCliWriteCommands
    - RunAzCliReadCommands
```

### Notifier

```yaml
api_version: azuresre.ai/v1
kind: AgentConfiguration
spec:
  name: incident-notifier
  system_prompt: |
    Send concise, structured status updates with evidence and next actions.
  handoff_description: Notification and stakeholder communication specialist
  tools:
    - SendOutlookEmail
```

## Operational Acceptance Tests

Before production handoff, confirm:

1. P1 simulated incident routes to correct custom agent.
2. Diagnostics output includes evidence + UTC timeline.
3. Write action proposal includes blast-radius and rollback.
4. KT gate rejects missing `SA/PA/DA/PPA` for P1/P2.
5. Scheduled tasks run and stop as configured (`Max executions` respected).
6. Connectors recover correctly from transient disconnect.
7. Final verification report includes all active plans, tasks, connectors, and hooks.

## Related References

- [bundles-operations.md](./bundles-operations.md)
- [incident-platforms-response-plans.md](./incident-platforms-response-plans.md)
- [scheduled-tasks.md](./scheduled-tasks.md)
- [connectors-and-mcp.md](./connectors-and-mcp.md)
- [hooks-governance.md](./hooks-governance.md)
- [kt-methodology.md](./kt-methodology.md)
- [kt-templates.md](./kt-templates.md)
- [aks-containerapps-production.md](./aks-containerapps-production.md)
- [drasi-aks-playbook.md](./drasi-aks-playbook.md)
