# Connector Token Security and Operations

Use this guide for secure production connector credential practices.

## Security Defaults

1. Use service accounts for production connectors.
2. Avoid personal tokens for long-lived integrations.
3. Grant least privilege required for each connector.
4. Store tokens in secure secret stores and rotate on schedule.
5. Monitor token usage and revoke on suspicion or personnel change.

## PagerDuty Token Pattern

Preferred model:

1. dedicated PagerDuty service account
2. role scoped to required operations (Responder/Observer style)
3. user API token format in connector auth (`Token <token>`)

Do not use account-level API key where user token is required by MCP server.

## Azure Managed Grafana Token Pattern

Preferred model:

1. Grafana service account token for persistent agent connectivity
2. Viewer role by default
3. elevate to Editor/Admin only when required

Alternative:

- Entra ID token for managed identity/service principal, with role assignment.
- Treat Entra token flow as short-lived and refresh-aware.

## Dynatrace Token Pattern

Use platform token with minimum scopes:

1. MCP gateway invoke/read scopes
2. add only required query/problem/security scopes for enabled workflows

Keep scoped token per environment (dev/stage/prod), not global.

## Rotation and Ownership

Set a default connector token policy:

1. owner: team mailbox + on-call group, not individual user
2. rotation cadence: 30-90 days based on risk
3. immediate rotation triggers:
   - role change/offboarding
   - credential leak suspicion
   - failed audit

## Validation After Rotation

After each credential rotation:

1. verify connector state is Connected
2. run one read-only tool test
3. run one workflow test (if applicable)
4. confirm no degraded automations

## Incident Response for Token Failures

When status flips to Failed:

1. check auth header format
2. verify endpoint URL and region
3. verify token scopes/roles
4. rotate token if uncertain
5. retest and document recovery

## Sources

- Official plugin repository: https://github.com/Azure/sre-agent-plugins
- PagerDuty plugin docs: https://github.com/Azure/sre-agent-plugins/tree/main/plugins/pager-duty
- Dynatrace plugin docs: https://github.com/Azure/sre-agent-plugins/tree/main/plugins/dynatrace
- Azure Managed Grafana plugin docs: https://github.com/Azure/sre-agent-plugins/tree/main/plugins/azure-managed-grafana

## Bundle Mapping

Connector templates live in:

- [`../bundles/connectors-observability/connectors/`](../bundles/connectors-observability/connectors/)
