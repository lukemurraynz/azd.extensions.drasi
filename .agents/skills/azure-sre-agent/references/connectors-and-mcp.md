# Connectors and MCP Operations

Use this guide for external integrations and MCP reliability.

## Built-In vs Connector Capabilities

With no connectors, agent still has built-in access to:

- Application Insights
- Log Analytics
- Azure Monitor metrics
- Azure Resource Graph
- Azure Resource Manager / Azure CLI
- AKS diagnostics

Add connectors when access is required outside core Azure surfaces.

## Connector Categories

1. Data sources (for non-default telemetry).
2. Source code and knowledge (GitHub, Azure DevOps).
3. Collaboration (Teams, Outlook).
4. Custom MCP servers (third-party and internal systems).

## MCP Health States

Track connector status in Builder > Connectors:

| Status | Meaning | Operator action |
| --- | --- | --- |
| Connected | Healthy, tools available | None |
| Disconnected | Temporary loss, auto-recovery attempts | Observe heartbeat recovery |
| Failed | Unrecoverable or bad config | Fix URL/auth/config |
| Initializing | Connection setup in progress | Wait and re-check |
| Not Available | No running agent instance to check status | Start/restore agent runtime |

### Recovery Behavior

1. Heartbeat check runs every 60 seconds.
2. Transient failures can self-recover.
3. Pre-execution checks verify connection before tool use.

## MCP Tool Assignment Patterns

### Individual tool assignment (precision)

```yaml
mcp_tools:
  - kusto-mcp_kusto_query
  - kusto-mcp_kusto_table_schema
```

### Wildcard assignment (coverage)

```yaml
mcp_tools:
  - kusto-mcp/*
```

Notes:

- Wildcard syntax is `{connection-id}/*`.
- Wildcard support applies to version `26.2.9.0+`.
- Health monitoring status model applies to version `26.1.25.0+`.

## Trust Boundary Guidance

1. Use individual tool lists by default for production.
2. Use wildcard only when:
   - connector trust is high, and
   - automatic adoption of future tools is acceptable.
3. Split custom agents by connector trust domain when possible.

## AKS and Container Apps Connector Set

Typical production set:

1. Built-in Azure tools for metrics/logs/resource state.
2. GitHub connector for source and PR correlation.
3. Optional PagerDuty or ServiceNow incident platform.
4. Optional Grafana/Datadog/Dynatrace MCP for cross-platform observability.

## Troubleshooting Checklist

1. Verify endpoint URL and auth format.
2. Validate token scope and expiration.
3. Confirm external system permissions.
4. Check connector status detail pane for heartbeat and tool count.
5. Retry with minimal single-tool invocation.

## Sources

- Connectors: https://learn.microsoft.com/en-us/azure/sre-agent/connectors
- Workflow automation: https://learn.microsoft.com/en-us/azure/sre-agent/workflow-automation
- Official plugins: https://github.com/Azure/sre-agent-plugins
