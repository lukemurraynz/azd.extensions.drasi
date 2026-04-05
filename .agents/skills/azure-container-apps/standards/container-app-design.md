# Container App Design Standard

## Architecture Patterns

### Single Container App

Use for standalone services with a single responsibility:

- Web APIs, background workers, scheduled jobs
- One container per app (plus optional Dapr sidecar)
- Scale independently based on workload

### Multi-Service Environment

Use a shared Container App Environment for services that:

- Belong to the same application boundary
- Need internal service-to-service communication
- Share logging infrastructure (Log Analytics)

Each service gets its own Container App within the shared environment.

---

## Container Configuration

### Resource Allocation

| Workload Type     | CPU  | Memory | Min Replicas | Max Replicas |
| ----------------- | ---- | ------ | ------------ | ------------ |
| Web API           | 0.5  | 1Gi    | 1            | 10           |
| Background Worker | 0.25 | 0.5Gi  | 0            | 20           |
| Scheduled Job     | 0.5  | 1Gi    | 0            | 5            |
| CPU-intensive     | 2.0  | 4Gi    | 1            | 10           |

Start small and scale out horizontally. Prefer more replicas at lower resources
over fewer replicas at higher resources.

### Probes

Every container must define probes:

```bicep
probes: [
  {
    type: 'liveness'
    httpGet: {
      path: '/healthz/live'
      port: 8080
    }
    initialDelaySeconds: 5
    periodSeconds: 10
  }
  {
    type: 'readiness'
    httpGet: {
      path: '/healthz/ready'
      port: 8080
    }
    initialDelaySeconds: 10
    periodSeconds: 15
  }
]
```

> [!NOTE]
> Always specify probe timeouts and failure thresholds explicitly. Recommended defaults: `timeoutSeconds: 5`, `failureThreshold: 3`, `periodSeconds: 10`. Overly aggressive probes (e.g., `timeoutSeconds: 1`) cause false restarts under load.

---

## Ingress Configuration

| Setting         | External API                  | Internal Service  |
| --------------- | ----------------------------- | ----------------- |
| `external`      | `true`                        | `false`           |
| `targetPort`    | Application port              | Application port  |
| `transport`     | `http` or `http2`             | `http` or `http2` |
| `allowInsecure` | `false`                       | `false`           |
| Custom domain   | Configure via portal or Bicep | Not needed        |
| IP restrictions | Configure as needed           | VNet scoped       |

---

## Environment Variables

| Pattern                      | Use                           |
| ---------------------------- | ----------------------------- |
| Plain `env` value            | Non-sensitive configuration   |
| Secret reference             | Sensitive values from secrets |
| Managed Identity + Key Vault | Preferred for all secrets     |

```bicep
env: [
  { name: 'ASPNETCORE_ENVIRONMENT', value: 'Production' }
  { name: 'ConnectionString', secretRef: 'db-connection' }
]
```

---

---

## Jobs vs Apps

| Characteristic | Container App                       | Container App Job                        |
| -------------- | ----------------------------------- | ---------------------------------------- |
| Lifecycle      | Runs continuously                   | Runs to completion and exits             |
| Resource type  | `Microsoft.App/containerApps`       | `Microsoft.App/jobs`                     |
| Ingress        | Supported (external/internal)       | Not supported                            |
| Scaling        | Replica-based (HTTP, queue, custom) | Execution-based (manual, schedule, KEDA) |
| Probes         | Liveness + readiness required       | Not applicable                           |
| Revisions      | Multiple revision traffic splitting | Not applicable                           |
| Dapr           | Supported                           | Supported (same environment)             |
| Use cases      | APIs, web apps, background services | Batch processing, ETL, migrations, cron  |

### When to Use Jobs

- **Scheduled tasks** — nightly reports, data aggregation, cleanup routines
- **Event processing** — drain a queue or process blob uploads one-at-a-time
- **Data migrations** — one-off schema changes or data transformations
- **CI/CD tasks** — build, test, or deploy steps triggered by events

### When NOT to Use Jobs

- The workload needs an HTTP endpoint → use a Container App
- The workload needs to run indefinitely → use a Container App with background workers
- You need real-time scaling based on HTTP requests → use a Container App

---

## Rules

1. Use `SystemAssigned` managed identity — avoid storing credentials in env vars or secrets.
2. Pull images via managed identity — set `identity: 'system'` in registry config.
3. Configure both liveness and readiness probes — never skip probes (apps only; jobs do not use probes).
4. Use `activeRevisionsMode: 'Multiple'` for production — enables traffic splitting.
5. Keep containers stateless — use external storage (Blob, Redis, Cosmos) for state.
6. Set resource requests to minimum needed — scale out, not up.
7. Use jobs for run-to-completion workloads — do not use long-running Container Apps as batch processors.
