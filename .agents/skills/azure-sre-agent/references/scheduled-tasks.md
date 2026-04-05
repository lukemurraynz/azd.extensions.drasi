# Scheduled Tasks Operations

Use this guide for recurring automation in production.

## Required Controls

1. Use clear task names and descriptions.
2. Use `Draft the cron for me` for human-readable schedules.
3. Use `Polish instructions` only as draft assistance, then review manually.
4. Always set `Max executions` for bounded workflows.
5. `Max executions` takes precedence over `End date`.

## Mode and Risk Defaults

1. Start new tasks in Review mode.
2. Promote to Autonomous only after successful repeated runs.
3. Keep high-impact tasks in Review unless there is clear blast-radius control.

## Task Categories

### Health Monitoring

Goal: detect degradation early and notify.

Instruction template:

```text
Check health for <resource-scope> over the last 24 hours:
1. Query Azure Resource Health for degraded resources.
2. Query Application Insights/Log Analytics for error rate trends.
3. Correlate with deployment/revision changes if available.
4. Summarize findings with severity and confidence.
5. Notify via <channel> only when anomalies are detected.
```

### Post-Incident Validation

Goal: validate stability after mitigation.

Instruction template:

```text
Run post-mitigation validation for <service> every 1 minute for up to 30 runs:
1. Confirm key health indicators remain stable.
2. On failure, collect logs and escalate immediately.
3. On success, record a completion summary.
4. Generate a final report with pass/fail status and evidence.
```

### Security/Compliance Checks

Goal: recurring guardrails for identity, secrets, and network posture.

Instruction template:

```text
Run weekly security review for <application>:
1. Check authentication and authorization controls.
2. Validate secret handling and managed identity usage.
3. Review access controls and network exposure.
4. Produce prioritized findings and recommended remediation.
```

## Testing Pattern

Before production activation:

1. Validate custom agent behavior in Test playground.
2. Use `Run task now` at least twice.
3. Confirm expected outputs and notification behavior.
4. Confirm no unbounded loop conditions.

## Failure Handling Pattern

Include explicit failure path in task prompt:

1. What counts as failure.
2. Immediate actions.
3. Escalation destination.
4. Stop condition.

## AKS and Container Apps Task Starters

### AKS Cluster Health

```text
Check AKS cluster <cluster-name> in <resource-group>:
1. Verify node Ready status and recent node events.
2. List pods in CrashLoopBackOff/Pending.
3. Query logs for top errors in the last 2 hours.
4. Check API server, ingress, and DNS-related symptoms.
5. Summarize by service impact and recommended next action.
```

### Container Apps Revision Health

```text
Validate Container App <app-name> revision health:
1. Compare current revision latency/error metrics to previous revision.
2. Detect sustained degradation over threshold.
3. Collect revision logs for the suspect revision.
4. Recommend rollback only with evidence and impact statement.
```

## Sources

- Scheduled tasks: https://learn.microsoft.com/en-us/azure/sre-agent/scheduled-tasks
- Workflow automation: https://learn.microsoft.com/en-us/azure/sre-agent/workflow-automation
