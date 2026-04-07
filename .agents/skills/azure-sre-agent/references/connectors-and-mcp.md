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
3. Collaboration (Teams, Outlook, Slack, Jira).
4. Incident management (PagerDuty, ServiceNow).
5. Observability (Datadog, New Relic, Splunk, Elasticsearch, Dynatrace, Grafana).
6. Cloud infrastructure (AWS).
7. Custom MCP servers (third-party and internal systems).

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

## MCP Transport Types

Two transport modes are supported:

| Transport | Use case |
| --- | --- |
| Streamable-HTTP | Cloud-hosted MCP servers with HTTPS endpoints |
| Stdio | Locally hosted MCP servers running as sidecar processes |

Choose Streamable-HTTP for SaaS connectors (Datadog, Splunk, etc.) and Stdio for self-hosted or internal MCP servers.

## AKS and Container Apps Connector Set

Typical production set:

1. Built-in Azure tools for metrics/logs/resource state.
2. GitHub or Azure DevOps connector for source and PR correlation.
3. Optional PagerDuty or ServiceNow incident platform.
4. Optional Datadog/New Relic/Splunk/Dynatrace/Grafana MCP for cross-platform observability.
5. Optional Slack or Jira for collaboration and work tracking.

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
- MCP connectors (pre-configured partners): https://sre.azure.com/docs/capabilities/mcp-connectors
