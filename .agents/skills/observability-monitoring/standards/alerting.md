# Alerting & Dashboards Standard

## Alert Design

### Alert on Symptoms, Not Causes

| Alert Type   | Good Example                        | Bad Example             |
| ------------ | ----------------------------------- | ----------------------- |
| Availability | "Error rate > 5% for 5 min"         | "Pod restarted"         |
| Latency      | "p95 response time > 2s for 10 min" | "CPU > 80%"             |
| Saturation   | "Queue depth > 1000 for 15 min"     | "Memory > 90%"          |
| Correctness  | "Payment success rate < 99%"        | "Exception count > 100" |

### Severity Levels

| Severity | Response       | Example                         | Action Group                |
| -------- | -------------- | ------------------------------- | --------------------------- |
| Sev 0    | Page on-call   | Complete service outage         | SMS + Phone + Email + Teams |
| Sev 1    | Respond 15 min | Degraded for >5% of users       | Email + Teams               |
| Sev 2    | Respond 1 hour | Non-critical component degraded | Email + Teams               |
| Sev 3    | Next business  | Warning threshold approached    | Email                       |

---

## Alert Types in Azure Monitor

### Metric Alerts

```bicep
resource alert 'Microsoft.Insights/metricAlerts@2018-03-01' = {
  name: 'high-error-rate'
  location: 'global'
  properties: {
    severity: 1
    enabled: true
    scopes: [appInsights.id]
    evaluationFrequency: 'PT5M'
    windowSize: 'PT15M'
    criteria: {
      'odata.type': 'Microsoft.Azure.Monitor.SingleResourceMultipleMetricCriteria'
      allOf: [
        {
          name: 'ErrorRate'
          metricName: 'requests/failed'
          operator: 'GreaterThan'
          threshold: 5
          timeAggregation: 'Count'
        }
      ]
    }
    actions: [{ actionGroupId: actionGroup.id }]
  }
}
```

### Log-Based Alerts (KQL)

```kql
// Alert: High exception rate by component
AppExceptions
| where TimeGenerated > ago(15m)
| summarize ExceptionCount = count() by AppRoleInstance
| where ExceptionCount > 50
```

### Availability Tests

```bicep
resource availabilityTest 'Microsoft.Insights/webtests@2022-06-15' = {
  name: 'health-check'
  location: location
  kind: 'standard'
  properties: {
    SyntheticMonitorId: 'health-check'
    Name: 'Health Endpoint Check'
    Enabled: true
    Frequency: 300  // 5 minutes
    Timeout: 30
    Kind: 'standard'
    Locations: [
      { Id: 'us-va-ash-azr' }
      { Id: 'emea-au-syd-edge' }
    ]
    Request: {
      RequestUrl: 'https://${appUrl}/healthz/live'
      HttpVerb: 'GET'
    }
    ValidationRules: {
      ExpectedHttpStatusCode: 200
    }
  }
}
```

---

## Action Groups

```bicep
resource actionGroup 'Microsoft.Insights/actionGroups@2023-01-01' = {
  name: 'ops-team'
  location: 'global'
  properties: {
    groupShortName: 'ops'
    enabled: true
    emailReceivers: [
      { name: 'ops-email', emailAddress: 'ops@contoso.com', useCommonAlertSchema: true }
    ]
  }
}
```

---

## Dashboard & Workbook Patterns

### Recommended Workbook Sections

| Section          | Content                                          |
| ---------------- | ------------------------------------------------ |
| Overview         | Request rate, error rate, p50/p95/p99 latency    |
| Dependencies     | External call latency and failure rate           |
| Exceptions       | Top exceptions by count, trending new exceptions |
| Infrastructure   | CPU, memory, pod/instance count, restarts        |
| Business Metrics | Domain-specific KPIs (orders/min, active users)  |

### Key KQL for Workbooks

```kql
// Request latency percentiles
AppRequests
| where TimeGenerated > ago(1h)
| summarize
    p50 = percentile(DurationMs, 50),
    p95 = percentile(DurationMs, 95),
    p99 = percentile(DurationMs, 99)
  by bin(TimeGenerated, 5m)
| render timechart
```

```kql
// Dependency health heatmap
AppDependencies
| where TimeGenerated > ago(1h)
| summarize
    SuccessRate = round(countif(Success == true) * 100.0 / count(), 1),
    AvgDuration = round(avg(DurationMs), 0)
  by Target, bin(TimeGenerated, 15m)
```

---

## Rules

1. Every alert must have an action group — silent alerts are worse than no alerts.
2. Use suppression windows for maintenance — do not disable alerts permanently.
3. Review alert noise monthly — high-volume low-action alerts erode on-call trust.
4. Workbooks over dashboards — workbooks support parameters, drill-down, and sharing.
5. Alert on SLO breach, not individual metric spikes — align with service objectives.
