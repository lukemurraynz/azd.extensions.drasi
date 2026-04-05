---
name: observability-monitoring
description: >-
  Application observability and monitoring patterns using Azure Monitor, Application Insights, and OpenTelemetry. Covers structured logging, distributed tracing, metrics collection, alerting rules, dashboards, and KQL queries. USE FOR: instrumenting applications, configuring monitoring infrastructure.compatibility: Requires Azure CLI, Application Insights SDK or OpenTelemetry SDK
---

# Observability & Monitoring Skill

> **MUST:** Use OpenTelemetry as the instrumentation standard. Application Insights
> is the preferred Azure backend. All services MUST emit structured logs, distributed
> traces, and metrics from day one — observability is not an afterthought.

> **Aspire users:** .NET Aspire's `AddServiceDefaults()` auto-configures OTEL exporters,
> structured logging, and the Aspire Dashboard for local dev. See
> [dotnet-aspire](../dotnet-aspire/SKILL.md) for Aspire-specific observability patterns.

---

## Quick Reference

| Capability           | Description                                                       |
| -------------------- | ----------------------------------------------------------------- |
| Structured Logging   | JSON-formatted logs with correlation IDs and semantic fields      |
| Distributed Tracing  | End-to-end request tracing across service boundaries              |
| Metrics Collection   | Custom and platform metrics with dimensional aggregation          |
| Alerting Rules       | Threshold, dynamic, and log-based alert configurations            |
| Dashboards/Workbooks | Azure Monitor Workbooks and dashboards for operational visibility |
| KQL Queries          | Kusto Query Language templates for common diagnostic scenarios    |
| Health Endpoints     | Liveness and readiness probes for orchestrated environments       |

---

## AKS Observability Defaults (AKS Book)

- **Managed Prometheus** for metrics; **Container Insights (Logs & Events preset)** to avoid duplicate metric ingestion.
- **ContainerLogV2** for container logs (target full migration before Q3 2026).
- **Correlation IDs at ingress/gateway**; every service logs and propagates the ID.
- **Structured JSON logs** with consistent field names; reserve `ERROR` for actionable failures.
- **Golden signals** (latency, traffic, errors, saturation) with histogram metrics (not averages).
- **Low-cardinality labels** only; never use user IDs, IPs, timestamps as labels.

---

## Standards

| Standard                                                | Purpose                             |
| ------------------------------------------------------- | ----------------------------------- |
| [Structured Logging](standards/structured-logging.md)   | Log format, correlation, and levels |
| [Distributed Tracing](standards/distributed-tracing.md) | Trace propagation and span design   |
| [Alerting & Dashboards](standards/alerting.md)          | Alert rules, action groups, KQL     |
| [Checklist](standards/checklist.md)                     | Validation checklist                |

---

## Actions

| Action                                                      | When to use                       |
| ----------------------------------------------------------- | --------------------------------- |
| [Instrument Application](actions/instrument-application.md) | Adding observability to a service |
| [Configure Monitoring](actions/configure-monitoring.md)     | Setting up Azure Monitor infra    |

---

## OpenTelemetry + Azure Monitor Setup

### .NET (Recommended)

```csharp
// Program.cs
builder.Services.AddOpenTelemetry()
    .UseAzureMonitor(options =>
    {
        options.ConnectionString = builder.Configuration["APPLICATIONINSIGHTS_CONNECTION_STRING"];
    })
    .WithTracing(tracing => tracing
        .AddAspNetCoreInstrumentation()
        .AddHttpClientInstrumentation())
    .WithMetrics(metrics => metrics
        .AddAspNetCoreInstrumentation()
        .AddHttpClientInstrumentation());
```

### JavaScript / Node.js

```typescript
import { useAzureMonitor } from "@azure/monitor-opentelemetry";

useAzureMonitor({
  azureMonitorExporterOptions: {
    connectionString: process.env.APPLICATIONINSIGHTS_CONNECTION_STRING,
  },
});
```

### Python

```python
from azure.monitor.opentelemetry import configure_azure_monitor

configure_azure_monitor(
    connection_string=os.environ["APPLICATIONINSIGHTS_CONNECTION_STRING"]
)
```

---

## Structured Logging Pattern

Always use structured/semantic logging — never string interpolation in log messages:

```csharp
// GOOD — structured, queryable fields
logger.LogInformation("Order {OrderId} placed by {CustomerId} for {Amount}",
    orderId, customerId, amount);

// BAD — string interpolation loses structure
logger.LogInformation($"Order {orderId} placed by {customerId} for {amount}");
```

### Log Levels

| Level       | Use for                               | Alert? |
| ----------- | ------------------------------------- | ------ |
| Trace       | Verbose diagnostic detail             | No     |
| Debug       | Development-time diagnostics          | No     |
| Information | Normal operational events             | No     |
| Warning     | Unexpected but recoverable situations | Watch  |
| Error       | Failures requiring attention          | Yes    |
| Critical    | System-wide failures, data loss risk  | Page   |

> **Warning — log volume at scale:** `Information`-level logging generates **significant
> data volume** in high-traffic applications (thousands of requests/second). Application
> Insights charges per GB ingested. A single `LogInformation` per request at 1,000 RPS
> can generate 5–10 GB/day. Use adaptive sampling and set minimum log levels per category
> in production to control costs:
>
> ```csharp
> // Program.cs — configure sampling
> builder.Services.AddOpenTelemetry()
>     .UseAzureMonitor(options =>
>     {
>         options.ConnectionString = connectionString;
>         options.SamplingRatio = 0.1f; // Sample 10% of requests
>     });
>
> // appsettings.Production.json — raise log levels for noisy categories
> // "Logging": {
> //   "LogLevel": {
> //     "Default": "Warning",
> //     "Microsoft.AspNetCore": "Warning",
> //     "YourApp": "Information"
> //   }
> // }
> ```

---

## Key KQL Templates

### Error Rate Over Time

```kql
AppExceptions
| where TimeGenerated > ago(24h)
| summarize ErrorCount = count() by bin(TimeGenerated, 15m), ProblemId
| order by TimeGenerated desc
| render timechart
```

### Slow Dependency Calls

```kql
AppDependencies
| where TimeGenerated > ago(1h)
| where DurationMs > 1000
| summarize avg(DurationMs), count() by Target, Name
| order by avg_DurationMs desc
```

### Request Success Rate

```kql
AppRequests
| where TimeGenerated > ago(1h)
| summarize Total = count(), Failed = countif(Success == false) by bin(TimeGenerated, 5m)
| extend SuccessRate = round((Total - Failed) * 100.0 / Total, 2)
| render timechart
```

### End-to-End Transaction Search

```kql
union AppRequests, AppDependencies, AppTraces, AppExceptions
| where OperationId == '{operationId}'
| order by TimeGenerated asc
| project TimeGenerated, Type = $table, Name, DurationMs, Message, Success
```

---

## Bicep — Application Insights + Log Analytics

```bicep
resource logAnalytics 'Microsoft.OperationalInsights/workspaces@2025-07-01' = {
  name: logAnalyticsName
  location: location
  properties: {
    sku: { name: 'PerGB2018' }
    retentionInDays: 30
  }
}

resource appInsights 'Microsoft.Insights/components@2020-02-02' = {
  name: appInsightsName
  location: location
  kind: 'web'
  properties: {
    Application_Type: 'web'
    WorkspaceResourceId: logAnalytics.id
    DisableLocalAuth: true  // Managed Identity only
  }
}
```

---

## Bicep — Alert Rules

Define alert rules in IaC to ensure monitoring survives deployments:

```bicep
resource errorRateAlert 'Microsoft.Insights/metricAlerts@2018-03-01' = {
  name: 'alert-${appInsightsName}-error-rate'
  location: 'global'
  properties: {
    severity: 2 // Warning
    enabled: true
    scopes: [ appInsights.id ]
    evaluationFrequency: 'PT5M'
    windowSize: 'PT15M'
    criteria: {
      'odata.type': 'Microsoft.Azure.Monitor.SingleResourceMultipleMetricCriteria'
      allOf: [
        {
          name: 'HighErrorRate'
          metricName: 'requests/failed'
          operator: 'GreaterThan'
          threshold: 10
          timeAggregation: 'Count'
          criterionType: 'StaticThresholdCriterion'
        }
      ]
    }
    actions: [
      { actionGroupId: actionGroup.id }
    ]
  }
}

resource actionGroup 'Microsoft.Insights/actionGroups@2023-01-01' = {
  name: 'ag-${appInsightsName}-oncall'
  location: 'global'
  properties: {
    groupShortName: 'OnCall'
    enabled: true
    emailReceivers: [
      {
        name: 'OnCallTeam'
        emailAddress: oncallEmail
        useCommonAlertSchema: true
      }
    ]
  }
}
```

> **Tip:** Alert on symptoms (error rate, latency percentile) not causes (CPU, memory).
> Symptom-based alerts fire when users are impacted, reducing alert fatigue.

---

## Health Endpoints

Implement health checks for orchestrated environments (Kubernetes, Container Apps):

```csharp
// Program.cs
builder.Services.AddHealthChecks()
    .AddCheck("self", () => HealthCheckResult.Healthy())
    .AddAzureBlobStorage(connectionString)  // dependency checks
    .AddNpgSql(connectionString);

app.MapHealthChecks("/healthz/live", new() { Predicate = _ => false });  // liveness
app.MapHealthChecks("/healthz/ready");  // readiness — includes dependencies
```

---

## Principles

1. **Instrument from day one** — observability is a first-class requirement, not a post-incident addition.
2. **Correlate everything** — every log, trace, and metric must carry an operation/correlation ID.
3. **Structured over unstructured** — semantic log fields enable querying; string messages do not.
4. **Alert on symptoms, not causes** — alert on user-facing impact (error rate, latency) not internal metrics.
5. **Dashboards for context** — workbooks provide context during incidents; build them before you need them.
6. **Retain proportionally** — hot data 30 days, warm 90 days, archive beyond; cost-optimise retention.

---

## Usage

Activate this skill when:

- Instrumenting a new service or application
- Adding Application Insights or OpenTelemetry to an existing project
- Configuring alert rules, action groups, or dashboards
- Writing KQL queries for diagnostics or reporting
- Reviewing observability coverage in a pull request

---

## References

- [Azure Monitor OpenTelemetry overview](https://learn.microsoft.com/azure/azure-monitor/app/opentelemetry-enable)
- [Application Insights for ASP.NET Core](https://learn.microsoft.com/azure/azure-monitor/app/asp-net-core)
- [KQL quick reference](https://learn.microsoft.com/kusto/query/kql-quick-reference)
- [Azure Monitor Workbooks](https://learn.microsoft.com/azure/azure-monitor/visualize/workbooks-overview)
- [Health checks in ASP.NET Core](https://learn.microsoft.com/aspnet/core/host-and-deploy/health-checks)

---

## Related Skills

- **azure-troubleshooting** — KQL templates, health checks, remediation playbooks
- **azure-defaults** — Region and tagging standards for monitoring resources
- **cost-optimization** — Log retention and data volume cost management

---

## Currency and Verification

- **Date checked:** 2026-03-31 (verified via Microsoft Learn MCP — ARM template references)
- **Compatibility:** Azure Bicep, ARM templates, Azure Monitor
- **Sources:**
  - [Microsoft.OperationalInsights/workspaces](https://learn.microsoft.com/azure/templates/microsoft.operationalinsights/workspaces)
  - [Microsoft.Insights/components](https://learn.microsoft.com/azure/templates/microsoft.insights/components) — `@2020-02-02` confirmed as latest GA
  - [Microsoft.Insights/metricAlerts](https://learn.microsoft.com/azure/templates/microsoft.insights/metricalerts) — `@2018-03-01` confirmed as latest GA
- **Verification steps:**
  1. Run `az provider show --namespace Microsoft.OperationalInsights --query "resourceTypes[?resourceType=='workspaces'].apiVersions" -o tsv` and confirm `2025-07-01` is listed
  2. Run `az bicep build --file <your-bicep-file>` to validate syntax

## Known Pitfalls

| Area | Pitfall | Mitigation |
|---|---|---|
| Insights/components API | `@2020-02-02` appears old but is genuinely the latest GA version | Do not update to a non-existent newer version; verify with `az provider show` |
| Insights/metricAlerts API | `@2018-03-01` appears old but is the only GA version (preview `@2024-03-01-preview` exists) | Use `@2018-03-01` for production; only use preview if specific preview features are required |
| OpenTelemetry version drift   | OpenTelemetry .NET SDK and Azure Monitor Exporter versions must match; mixing causes missing telemetry | Pin both packages to the same release train; check [OpenTelemetry .NET releases](https://github.com/open-telemetry/opentelemetry-dotnet/releases) |
| Application Insights sampling | Default adaptive sampling can drop important telemetry under load                                      | Configure fixed-rate sampling or exclude critical telemetry types from sampling                                                                   |
| Log Analytics ingestion delay | Data may take 5–15 minutes to appear in Log Analytics queries                                          | Don't rely on real-time KQL queries for incident detection; use metric alerts for time-critical scenarios                                         |
| Action group rate limits      | Action groups throttle notifications (e.g., 1 email per 5 minutes per address)                         | Use multiple action groups or notification channels for critical alerts                                                                           |
| Custom metrics cardinality    | High-cardinality dimensions (user IDs, request IDs) on custom metrics cause cost explosion             | Use low-cardinality dimensions only (status code, region, service name) for custom metrics                                                        |
