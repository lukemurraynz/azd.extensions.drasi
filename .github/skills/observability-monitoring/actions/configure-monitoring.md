# Action: Configure Monitoring Infrastructure

Set up Azure Monitor, Application Insights, alerting, and dashboards via IaC.

---

## Step 1 — Deploy Log Analytics + Application Insights

```bicep
param location string = resourceGroup().location
param environmentName string

resource logAnalytics 'Microsoft.OperationalInsights/workspaces@2025-07-01' = {
  name: 'log-${environmentName}'
  location: location
  properties: {
    sku: { name: 'PerGB2018' }
    retentionInDays: 30
  }
}

resource appInsights 'Microsoft.Insights/components@2020-02-02' = {
  name: 'appi-${environmentName}'
  location: location
  kind: 'web'
  properties: {
    Application_Type: 'web'
    WorkspaceResourceId: logAnalytics.id
    DisableLocalAuth: true
  }
}

output appInsightsConnectionString string = appInsights.properties.ConnectionString
output logAnalyticsWorkspaceId string = logAnalytics.id
```

---

## Step 2 — Configure Diagnostic Settings for Azure Resources

```bicep
resource diagSettings 'Microsoft.Insights/diagnosticSettings@2021-05-01-preview' = {
  name: 'send-to-log-analytics'
  scope: targetResource  // e.g., App Service, Container App, etc.
  properties: {
    workspaceId: logAnalytics.id
    logs: [
      { categoryGroup: 'allLogs', enabled: true }
    ]
    metrics: [
      { category: 'AllMetrics', enabled: true }
    ]
  }
}
```

---

## Step 3 — Create Action Group

```bicep
resource actionGroup 'Microsoft.Insights/actionGroups@2023-01-01' = {
  name: 'ag-ops-${environmentName}'
  location: 'global'
  properties: {
    groupShortName: 'ops'
    enabled: true
    emailReceivers: [
      {
        name: 'ops-team'
        emailAddress: 'ops@contoso.com'
        useCommonAlertSchema: true
      }
    ]
  }
}
```

---

## Step 4 — Create Alert Rules

### Error Rate Alert

```bicep
resource errorRateAlert 'Microsoft.Insights/metricAlerts@2018-03-01' = {
  name: 'alert-error-rate-${environmentName}'
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
          name: 'HighErrorRate'
          metricName: 'requests/failed'
          operator: 'GreaterThan'
          threshold: 10
          timeAggregation: 'Count'
        }
      ]
    }
    actions: [{ actionGroupId: actionGroup.id }]
  }
}
```

### Availability Test

```bicep
resource availabilityTest 'Microsoft.Insights/webtests@2022-06-15' = {
  name: 'avail-${environmentName}'
  location: location
  kind: 'standard'
  tags: {
    'hidden-link:${appInsights.id}': 'Resource'
  }
  properties: {
    SyntheticMonitorId: 'avail-${environmentName}'
    Name: 'Health Endpoint'
    Enabled: true
    Frequency: 300
    Timeout: 30
    Kind: 'standard'
    Locations: [
      { Id: 'emea-au-syd-edge' }   // Australia East
      { Id: 'us-va-ash-azr' }      // US East
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

## Step 5 — Create Workbook

Deploy a standard operational workbook or create manually in the portal:

1. Navigate to **Azure Monitor → Workbooks → New**
2. Add sections: Request Overview, Dependency Health, Exceptions, Infrastructure
3. Use parameterised time range and resource selectors
4. Save to a shared resource group for team access

---

## Step 6 — Validate

1. Confirm Log Analytics workspace is receiving data: `Heartbeat | take 10`
2. Confirm Application Insights shows live metrics
3. Trigger a test alert and verify notification delivery
4. Verify availability test is running from expected locations
5. Confirm diagnostic settings exist for all Azure resources in the resource group

---

## Completion Criteria

- [ ] Log Analytics workspace and Application Insights deployed via IaC
- [ ] Diagnostic settings configured for all Azure resources
- [ ] Action group created with appropriate receivers
- [ ] Error rate and availability alerts active
- [ ] Operational workbook created or deployed
- [ ] Notifications verified end-to-end
