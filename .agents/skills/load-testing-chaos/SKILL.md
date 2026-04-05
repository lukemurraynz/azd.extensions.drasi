---
name: load-testing-chaos
description: >-
  Azure Load Testing and chaos engineering patterns for resilience validation, performance baselining, and failure injection experiments.
  USE FOR: creating load tests, configuring chaos experiments, validating resilience requirements, establishing performance baselines, or integrating load testing into CI/CD pipelines.
---

# Load Testing & Chaos Engineering

> **Mandatory:** Establish performance baselines before production. Use Azure Load Testing for load/stress tests and Azure Chaos Studio for fault injection and resilience validation.

## Description

Patterns for load testing, performance testing, and chaos engineering — Azure Load Testing with JMeter, performance baselines, Chaos Studio experiments, and resilience validation.

## Capabilities

| Capability            | Details                                              |
| --------------------- | ---------------------------------------------------- |
| Load Testing          | Azure Load Testing with JMeter/Locust scripts        |
| Performance Baselines | Establish and track response time, throughput, RU    |
| Stress Testing        | Find breaking points and degradation limits          |
| Chaos Engineering     | Azure Chaos Studio fault injection experiments       |
| Resilience Validation | Verify retry, circuit breaker, and failover patterns |
| CI/CD Integration     | Automated performance gates in pipelines             |

## Standards

| Standard                                                | Purpose                  |
| ------------------------------------------------------- | ------------------------ |
| [Performance Testing](standards/performance-testing.md) | Test types and baselines |
| [Chaos Engineering](standards/chaos-engineering.md)     | Fault injection patterns |
| [Checklist](standards/checklist.md)                     | Validation checklist     |

## Actions

| Action                                                  | Purpose                 |
| ------------------------------------------------------- | ----------------------- |
| [Run Load Test](actions/run-load-test.md)               | Execute load test       |
| [Run Chaos Experiment](actions/run-chaos-experiment.md) | Execute fault injection |

---

## Azure Load Testing

### Bicep — Azure Load Testing Resource

```bicep
param location string = resourceGroup().location
param loadTestName string
param principalId string

resource loadTest 'Microsoft.LoadTestService/loadTests@2022-12-01' = {
  name: loadTestName
  location: location
  identity: {
    type: 'SystemAssigned'
  }
}

// Load Test Contributor role
resource loadTestRole 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(loadTest.id, principalId, '749a398d-560b-491b-bb21-08571b65f0c2')
  scope: loadTest
  properties: {
    principalId: principalId
    roleDefinitionId: subscriptionResourceId(
      'Microsoft.Authorization/roleDefinitions',
      '749a398d-560b-491b-bb21-08571b65f0c2' // Load Test Contributor
    )
    principalType: 'ServicePrincipal'
  }
}
```

### JMeter Test Script (Basic)

```xml
<?xml version="1.0" encoding="UTF-8"?>
<jmeterTestPlan version="1.2">
  <hashTree>
    <TestPlan guiclass="TestPlanGui" testclass="TestPlan" testname="API Load Test">
      <elementProp name="TestPlan.user_defined_variables" elementType="Arguments"/>
    </TestPlan>
    <hashTree>
      <ThreadGroup guiclass="ThreadGroupGui" testclass="ThreadGroup" testname="Users">
        <intProp name="ThreadGroup.num_threads">50</intProp>
        <intProp name="ThreadGroup.ramp_time">60</intProp>
        <intProp name="ThreadGroup.duration">300</intProp>
        <boolProp name="ThreadGroup.scheduler">true</boolProp>
      </ThreadGroup>
      <hashTree>
        <HTTPSamplerProxy guiclass="HttpTestSampleGui" testclass="HTTPSamplerProxy"
          testname="GET /api/health">
          <stringProp name="HTTPSampler.domain">${target_host}</stringProp>
          <stringProp name="HTTPSampler.port">443</stringProp>
          <stringProp name="HTTPSampler.protocol">https</stringProp>
          <stringProp name="HTTPSampler.path">/api/health</stringProp>
          <stringProp name="HTTPSampler.method">GET</stringProp>
        </HTTPSamplerProxy>
      </hashTree>
    </hashTree>
  </hashTree>
</jmeterTestPlan>
```

### Test Configuration (YAML)

```yaml
# load-test-config.yaml
version: v0.1
testId: api-load-test
testName: API Load Test
testPlan: test-plan.jmx
engineInstances: 1
failureCriteria:
  - avg(response_time_ms) > 500
  - percentage(error) > 5
  - p99(response_time_ms) > 2000
env:
  - name: target_host
    value: myapp.azurewebsites.net
```

---

## Azure Chaos Studio

### Bicep — Enable Chaos Target

```bicep
// Enable Chaos Studio on a resource (e.g., App Service)
resource chaosTarget 'Microsoft.Chaos/targets@2025-01-01' = {
  name: 'Microsoft-AppService'
  scope: webApp
  properties: {}
}

// Enable a specific capability (fault type)
resource chaosCapability 'Microsoft.Chaos/targets/capabilities@2025-01-01' = {
  parent: chaosTarget
  name: 'AppService-Stop-1.0'
  properties: {}
}
```

### Bicep — Chaos Experiment

```bicep
resource chaosExperiment 'Microsoft.Chaos/experiments@2025-01-01' = {
  name: '${webAppName}-stop-experiment'
  location: location
  identity: {
    type: 'SystemAssigned'
  }
  properties: {
    selectors: [
      {
        type: 'List'
        id: 'selector1'
        targets: [
          {
            type: 'ChaosTarget'
            id: chaosTarget.id
          }
        ]
      }
    ]
    steps: [
      {
        name: 'Step 1 — Stop App Service'
        branches: [
          {
            name: 'Branch 1'
            actions: [
              {
                type: 'continuous'
                name: 'urn:csci:microsoft:appService:stop/1.0'
                duration: 'PT5M'
                selectorId: 'selector1'
                parameters: []
              }
            ]
          }
        ]
      }
    ]
  }
}

// Grant the experiment identity the required role on the target
resource chaosRoleAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(chaosExperiment.id, webApp.id, 'contributor')
  scope: webApp
  properties: {
    principalId: chaosExperiment.identity.principalId
    roleDefinitionId: subscriptionResourceId(
      'Microsoft.Authorization/roleDefinitions',
      'b24988ac-6180-42a0-ab88-20f7382dd24c' // Contributor
    )
    principalType: 'ServicePrincipal'
  }
}
```

> [!WARNING]
> The `Contributor` role grants broader permissions than needed for fault injection. Prefer a custom role with only the required fault provider permissions (e.g., `Microsoft.Compute/virtualMachines/restart/action` for VM faults). If using `Contributor`, document the trade-off in the Bicep template and restrict the scope to the target resource group only.

### Common Chaos Faults

| Fault                     | Target      | Duration | What It Tests                    |
| ------------------------- | ----------- | -------- | -------------------------------- |
| App Service Stop          | App Service | 5 min    | Failover, health probes          |
| Network Disconnect        | VM / VMSS   | 5 min    | Network resilience               |
| DNS Failure               | VM / VMSS   | 5 min    | DNS fallback                     |
| Cosmos DB Failover        | Cosmos DB   | Varies   | Multi-region failover            |
| Key Vault Deny Access     | Key Vault   | 5 min    | Secret caching, graceful degrade |
| AKS Chaos Mesh (Pod Kill) | AKS         | 5 min    | Pod restart, self-healing        |
| CPU Pressure              | VM / VMSS   | 10 min   | Auto-scaling, throttling         |

---

## Performance Baselines

| Metric            | Target            | Source               |
| ----------------- | ----------------- | -------------------- |
| P50 Response Time | < 200 ms          | Application Insights |
| P95 Response Time | < 500 ms          | Application Insights |
| P99 Response Time | < 2000 ms         | Application Insights |
| Error Rate        | < 0.1%            | Application Insights |
| Throughput        | > baseline RPS    | Azure Load Testing   |
| RU Consumption    | < 80% provisioned | Cosmos DB metrics    |
| CPU Utilisation   | < 70% average     | Azure Monitor        |

---

## AKS Book Load Testing Cadence

- **Revenue-critical apps**: quarterly full-scale tests to **2x expected peak**, plus event-driven tests before major releases or campaigns.
- **Non-critical apps**: annual comprehensive tests + pre-release smoke tests.
- **Always document** the max sustainable load and the first saturated resource.

---

## Principles

1. **Baseline before optimising** — establish performance metrics before making changes.
2. **Test in production-like environments** — use representative data and infrastructure.
3. **Automate performance gates** — fail CI/CD pipelines when baselines are breached.
4. **Start chaos small** — begin with non-production, single-fault experiments.
5. **Always have a hypothesis** — define what you expect to happen before running chaos.
6. **Monitor during experiments** — watch dashboards and alerts during chaos runs.

## References

- [Azure Load Testing](https://learn.microsoft.com/en-us/azure/load-testing/overview-what-is-azure-load-testing)
- [Azure Chaos Studio](https://learn.microsoft.com/en-us/azure/chaos-studio/chaos-studio-overview)
- [Chaos Studio fault library](https://learn.microsoft.com/en-us/azure/chaos-studio/chaos-studio-fault-library)
- [Load testing in CI/CD](https://learn.microsoft.com/en-us/azure/load-testing/quickstart-add-load-test-cicd)
- [Well-Architected — Reliability testing](https://learn.microsoft.com/en-us/azure/well-architected/reliability/testing-strategy)

## Related Skills

- [Observability & Monitoring](../observability-monitoring/SKILL.md) — Monitor during tests
- [Azure Container Apps](../azure-container-apps/SKILL.md) — Scaling validation
- [Azure Functions Patterns](../azure-functions-patterns/SKILL.md) — Function scaling tests

---

## Currency and Verification

- **Date checked:** 2026-03-31
- **API versions used:**
  - `Microsoft.Chaos` (targets, capabilities, experiments): `2025-01-01` GA
  - `Microsoft.LoadTestService/loadTests`: `2022-12-01` GA (latest stable)
  - `Microsoft.Authorization/roleAssignments`: `2022-04-01`
  - Chaos Studio REST API: `2025-01-01`
- **Sources:** [Chaos Studio ARM reference](https://learn.microsoft.com/azure/templates/microsoft.chaos/targets), [Load Test ARM reference](https://learn.microsoft.com/azure/templates/microsoft.loadtestservice/loadtests), [Chaos Studio limitations](https://learn.microsoft.com/azure/chaos-studio/chaos-studio-limitations)
- **Verification steps:**
  1. Verify Chaos API version: `az provider show --namespace Microsoft.Chaos --query "resourceTypes[?resourceType=='targets'].apiVersions" -o tsv`
  2. Verify Load Test API version: `az provider show --namespace Microsoft.LoadTestService --query "resourceTypes[?resourceType=='loadTests'].apiVersions" -o tsv`
  3. Check Chaos Studio region availability: [Supported regions](https://azure.microsoft.com/global-infrastructure/services/?products=chaos-studio)

### Known Pitfalls

| Area                              | Pitfall                                                                                                                                                                            | Mitigation                                                                                                                                                    |
| --------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Region support                    | Chaos experiments only run in [supported regions](https://azure.microsoft.com/global-infrastructure/services/?products=chaos-studio); targeting unsupported regions silently fails | Verify target resources are in supported regions before creating experiments                                                                                  |
| Resource Move                     | Experiment resources don't support Resource Move; must copy JSON and recreate                                                                                                      | Copy experiment JSON to use in other subscriptions/resource groups/regions                                                                                    |
| Agent network access              | Agent-based faults require outbound access to Chaos Studio regional endpoints                                                                                                      | Allowlist endpoints per [network security docs](https://learn.microsoft.com/azure/chaos-studio/chaos-studio-permissions-security#network-security)            |
| Network Disconnect fault          | Only affects **new** connections; existing active connections persist                                                                                                              | Restart services/processes to force connection breaks during experiments                                                                                      |
| No CLI/PowerShell modules         | No dedicated Azure CLI or PowerShell modules for Chaos Studio                                                                                                                      | Use `az rest` with the [REST API](https://learn.microsoft.com/rest/api/chaosstudio/) for automation                                                           |
| VMSS orchestration mode           | Service-direct faults only work with Uniform VMSS orchestration mode                                                                                                               | For Flexible mode, target individual VM instances with ARM VM shutdown fault                                                                                  |
| Linux network faults              | `sch_netem` kernel module may be missing on RHEL-based VMs, preventing latency/packet-loss faults                                                                                  | Verify with `modinfo sch_netem`; install `kernel-modules-extra` if missing                                                                                    |
| Least-privilege roles             | No built-in least-privileged roles for fault injection on resources                                                                                                                | Create custom roles or assign existing built-in roles per [fault providers docs](https://learn.microsoft.com/azure/chaos-studio/chaos-studio-fault-providers) |
| NSG v1.1 `flushConnection`        | Enabling `flushConnection` parameter causes `FlushingNetworkSecurityGroupConnectionIsNotEnabled` error                                                                             | Disable `flushConnection` or use NSG Security Rule v1.0 fault                                                                                                 |
| Cross-subscription VNet injection | VNet injection requires experiment and target VNet in the same subscription                                                                                                        | Create experiment in the same subscription as the target VNet                                                                                                 |
