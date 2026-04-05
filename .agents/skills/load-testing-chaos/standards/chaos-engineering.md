# Chaos Engineering

## Principles

1. **Define steady state** — know what "normal" looks like before injecting faults.
2. **Hypothesise the impact** — predict what will happen before running the experiment.
3. **Minimise blast radius** — start in non-production, scope experiments tightly.
4. **Run in production** — only production validates real resilience (when mature).
5. **Automate experiments** — run as part of regular resilience validation.

---

## Chaos Experiment Design

### Experiment Template

```yaml
experiment:
  name: app-service-stop
  hypothesis: "When App Service is stopped, traffic fails over to secondary instance within 60 seconds"
  steady_state:
    - health endpoint returns 200
    - p95 response time < 500ms
    - error rate < 0.1%
  fault:
    type: App Service Stop
    target: primary-app-service
    duration: 5 minutes
  expected_behaviour:
    - Traffic Manager routes to secondary within 60 seconds
    - Users experience < 30 seconds of degraded service
    - No data loss
  abort_conditions:
    - Error rate > 50% for more than 2 minutes
    - P99 response time > 10 seconds
    - Data corruption detected
  rollback:
    - Restart App Service
    - Verify health endpoint
```

---

## Azure Chaos Studio Setup

### Step 1 — Enable Targets

Enable Chaos Studio on target resources:

```bicep
resource chaosTarget 'Microsoft.Chaos/targets@2025-01-01' = {
  name: 'Microsoft-AppService'
  scope: webApp
  properties: {}
}

resource stopCapability 'Microsoft.Chaos/targets/capabilities@2025-01-01' = {
  parent: chaosTarget
  name: 'AppService-Stop-1.0'
  properties: {}
}
```

### Step 2 — Create Experiment

Define the experiment with steps and branches:

```bicep
resource experiment 'Microsoft.Chaos/experiments@2025-01-01' = {
  name: experimentName
  location: location
  identity: { type: 'SystemAssigned' }
  properties: {
    selectors: [
      { type: 'List', id: 'selector1', targets: [{ type: 'ChaosTarget', id: chaosTarget.id }] }
    ]
    steps: [
      {
        name: 'Inject fault'
        branches: [
          {
            name: 'Main'
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
```

### Step 3 — Assign Permissions

The experiment's managed identity needs appropriate permissions on target resources.

---

## Common Experiments

### Infrastructure Faults

| Experiment         | Validates               | Fault                         |
| ------------------ | ----------------------- | ----------------------------- |
| App Service Stop   | Failover, health probes | `appService:stop/1.0`         |
| VM Shutdown        | Auto-heal, scaling      | `virtualMachine:shutdown/1.0` |
| Network Disconnect | Retry, timeout handling | `networkDisconnect/1.0`       |
| DNS Failure        | DNS fallback, caching   | `dnsFailure/1.0`              |
| CPU Pressure       | Auto-scaling triggers   | `cpuPressure/1.0`             |

### Data Platform Faults

| Experiment            | Validates                            | Fault                    |
| --------------------- | ------------------------------------ | ------------------------ |
| Cosmos DB Failover    | Multi-region reads, connection retry | `cosmosDB:failover/1.0`  |
| Key Vault Access Deny | Secret caching, graceful degrade     | Manual NSG/firewall rule |
| SQL Failover          | Connection retry, read replicas      | `sqlServer:failover/1.0` |

### Container/Kubernetes Faults

| Experiment            | Validates                         | Fault                           |
| --------------------- | --------------------------------- | ------------------------------- |
| Pod Kill (Chaos Mesh) | Self-healing, replica sets        | AKS Chaos Mesh integration      |
| Pod Network Delay     | Timeout handling, circuit breaker | Chaos Mesh network fault        |
| Container Restart     | Startup probes, readiness probes  | Container Apps revision restart |

---

## Maturity Model

| Level   | Environment | Frequency  | Scope                          |
| ------- | ----------- | ---------- | ------------------------------ |
| Level 1 | Dev/Test    | Ad-hoc     | Single resource faults         |
| Level 2 | Staging     | Monthly    | Multi-resource scenarios       |
| Level 3 | Pre-prod    | Weekly     | Full failure domain simulation |
| Level 4 | Production  | Continuous | Game days, automated chaos     |

**Start at Level 1.** Progress only when confidence and observability are sufficient.

---

## Rules

1. Never run chaos experiments without active monitoring and alerting.
2. Document hypothesis, expected behaviour, and abort conditions before every experiment.
3. Start in non-production — production chaos requires mature observability.
4. Use `SystemAssigned` managed identity for experiment permissions.
5. Review experiment results within 24 hours — fix issues before running the next experiment.
6. Gradually increase scope — single service → multi-service → full domain.
