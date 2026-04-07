# Capability Matrix

Use this matrix to compose or trim capabilities without editing the core skill.

| Bundle | Primary Capabilities | Resource Types |
| --- | --- | --- |
| [`../bundles/base-core`](../bundles/base-core) | baseline incident routing, diagnostics, remediation review, notification | agents, response plans, scheduled tasks, hooks, checklist |
| [`../bundles/aks-production`](../bundles/aks-production) | AKS incident diagnostics and daily health monitoring | agents, response plans, scheduled tasks |
| [`../bundles/containerapps-production`](../bundles/containerapps-production) | Container Apps diagnostics and revision regression analysis | agents, response plans, scheduled tasks |
| [`../bundles/drasi-aks-production`](../bundles/drasi-aks-production) | Drasi lag/staleness triage and AKS-correlated diagnostics | agents, response plans, scheduled tasks |
| [`../bundles/governance-kt`](../bundles/governance-kt) | approval gates, tool audit, KT completeness policy | hooks |
| [`../bundles/connectors-observability`](../bundles/connectors-observability) | connector templates for GitHub/PagerDuty/Dynatrace/Grafana/Datadog/New Relic/Splunk/Elasticsearch/Hawkeye/AWS | connectors |

## Recommended Compositions

### Minimal production baseline

1. `base-core`
2. `governance-kt`

### AKS production

1. `base-core`
2. `aks-production`
3. `governance-kt`
4. `connectors-observability` (as needed)

### Container Apps production

1. `base-core`
2. `containerapps-production`
3. `governance-kt`
4. `connectors-observability` (as needed)

### Drasi on AKS

1. `base-core`
2. `aks-production`
3. `drasi-aks-production`
4. `governance-kt`
5. `connectors-observability` (as needed)
