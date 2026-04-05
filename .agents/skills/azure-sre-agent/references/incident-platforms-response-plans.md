# Incident Platforms and Response Plans

Use this guide for production incident routing behavior.

## Core Constraints

1. Only one incident platform can be active at a time.
2. Azure Monitor is default.
3. Switching to another platform disconnects Azure Monitor.
4. Connecting a platform can create `quickstart_handler`.
5. Overlapping quickstart + custom plans can double-process incidents or misroute.

## Quickstart Plan Handling

When onboarding incident management:

1. Connect platform.
2. Inspect generated quickstart plan.
3. Decide one approach:
   - Keep and customize it.
   - Disable it (`Turn off`).
   - Delete it and use explicit custom plans.
4. Ensure no filter overlap before enabling production plans.

## Response Plan Lifecycle

Use Builder > Incident response plans or Agent Canvas > Table view.

### Create

Set:

- Severity/Priority filters
- Impacted service
- Incident type
- Title contains
- Handler custom agent
- Agent autonomy level

### Operate

- Use `Turn off` for maintenance windows.
- Use `Turn on` to re-enable instantly.
- Keep paused plans for rollback-ready routing.

### Validate

Use historical-incident test mode before production enablement:

1. Select incident sample.
2. Run test.
3. Review generated response.
4. Confirm route, depth, and action quality.

Test mode is read-only.

## Run Mode Normalization

Use this precedence model:

1. Configure mode at response plan or scheduled task.
2. Treat agent-level mode as fallback.
3. Pair mode with RBAC:
   - Review mode does not bypass missing permissions.
   - Autonomous mode still needs permissions, or temporary access flow is requested.

## Production Routing Patterns

### Pattern A: Service split

- `api-high-sev` -> API-focused agent -> Review
- `db-critical` -> DB-focused agent -> Autonomous (only if validated)

### Pattern B: Severity split

- P1/P2 -> Deep diagnostics + remediation path
- P3/P4 -> Evidence collection + recommendation path

### Pattern C: Platform split

- Azure Monitor infrastructure alerts -> infra specialist
- PagerDuty app incidents -> app reliability specialist

## Minimal JSON Payload Pattern (API)

Use placeholders and keep explicit filters:

```json
{
  "id": "service-high-severity",
  "name": "Service High Severity",
  "priorities": ["Sev0", "Sev1"],
  "titleContains": "service-name",
  "handlingAgent": "incident-handler",
  "agentMode": "review",
  "maxAttempts": 3
}
```

## Operational Checklist

1. No overlapping active plans.
2. Every active plan has explicit owner.
3. Every active plan has tested sample incidents.
4. High-risk plans start in Review.
5. Paused/disabled plan strategy documented.

## Sources

- Incident platforms: https://learn.microsoft.com/en-us/azure/sre-agent/incident-platforms
- Incident response plans: https://learn.microsoft.com/en-us/azure/sre-agent/incident-response-plans
- Run modes: https://learn.microsoft.com/en-us/azure/sre-agent/run-modes
