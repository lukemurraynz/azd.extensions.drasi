# Deployment Patterns from Official Samples

This guide captures production-grade patterns seen in official `microsoft/sre-agent`
sample deployments and scripts.

## Pattern 1: Idempotent Post-Provision Setup

Post-provision automation should be safe to re-run.

Use upsert semantics for:

1. Knowledge base uploads
2. Custom agents
3. Incident response plans
4. Scheduled tasks
5. Connectors

Design goal: "fix-forward" without redeploying infrastructure.

## Pattern 2: Eventual Consistency Retries

Some operations are not immediately available after provisioning.

Use bounded retry loops for:

- incident platform availability
- response plan creation/update
- connector readiness

Recommended defaults:

1. 3 attempts
2. 10-30 second delay
3. clear failure message with operator next step

## Pattern 3: Quickstart Cleanup

If quickstart plans are auto-created, remove overlap before custom routes.

Automation step:

1. list active plans
2. detect quickstart handler
3. disable/delete when custom equivalent exists

## Pattern 4: Full Verification Pass

After setup, verify all control-plane objects:

1. knowledge files indexed
2. custom agents present
3. incident platform connected
4. connectors healthy
5. response plans active
6. scheduled tasks active

Return a single readiness summary.

## Pattern 5: Memory-First Investigation

Investigation agents should search memory before fresh diagnostics.

Default flow:

1. memory lookup for similar incidents
2. runbook lookup
3. logs + metrics evidence
4. root-cause hypothesis
5. action proposal

## Pattern 6: Structured Incident Report Templates

Use a fixed report schema across teams:

1. Summary
2. Impact
3. Timeline (UTC)
4. Evidence
5. Root Cause
6. Remediation
7. Action Items
8. References

This supports reliable handoffs and postmortems.

## Pattern 7: Scheduled Issue Triage Workflows

For code-facing operations:

1. scheduled task triggers issue triage agent
2. triage agent classifies and labels issues
3. triage agent posts standardized response comment
4. escalation for P0/P1 style issues

## Pattern 8: KT-Governed Major Incident Handling

For high-severity incidents and production write paths:

1. perform Situation Appraisal before deep investigation
2. produce Problem Analysis with explicit `IS / IS NOT` logic
3. evaluate remediation options with Decision Analysis (`MUST/WANT`)
4. protect execution plan with Potential Problem Analysis

Use hooks to reject incomplete major-incident responses.

## Pattern 9: Starter Lab Deployment (azd-based)

The official starter lab at `microsoft/sre-agent/labs/starter-lab` uses `azd up` for deployment:

1. Deploy: SRE Agent, sample app (Grubify), Log Analytics, App Insights, Alert, Container Registry, Managed Identity.
2. Three scenario tracks: IT Operations, Developer (GitHub), Workflow Automation.
3. Estimated time: ~40 minutes.

Use this as a reference for azd-based SRE Agent provisioning patterns. See https://github.com/microsoft/sre-agent/tree/main/labs/starter-lab.

## AKS and Container Apps Production Pattern

### Split specialist custom agents

1. `aks-diagnostics` for cluster/pod/network diagnostics
2. `containerapps-revision-analyzer` for revision health and rollback evidence
3. `incident-notifier` for outbound notifications and stakeholder updates

### Keep responsibilities separate

1. Diagnose agents collect evidence.
2. Remediation agents execute approved changes.
3. Notifier agents report outcomes.

## Drasi on AKS Pattern

Use this when SRE Agent supports Drasi workloads on AKS:

1. Detect symptom from alert:
   - event processing lag
   - query staleness
   - connector ingestion faults
2. Gather evidence:
   - AKS node/pod health
   - Drasi operator/controller logs
   - backing data plane dependencies
3. Correlate:
   - deployment/revision changes
   - config drift
   - recent cluster/network events
4. Route:
   - infrastructure issue -> AKS remediator path
   - app/workflow issue -> Drasi workflow owner path
5. Report:
   - service impact
   - likely failure domain
   - recommended action and risk

Use [drasi-aks-playbook.md](./drasi-aks-playbook.md) for a ready-to-use
Drasi-specific operating template.

## Sources

- Official samples: https://github.com/microsoft/sre-agent
- Hands-on lab patterns: https://github.com/microsoft/sre-agent/tree/main/samples/hands-on-lab
- Deployment compliance sample: https://github.com/microsoft/sre-agent/tree/main/samples/deployment-compliance
