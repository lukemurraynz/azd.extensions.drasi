---
name: azure-container-apps
description: >-
  Azure Container Apps patterns including scaling rules, Dapr integration, revision management, service connectors, managed identity, and Bicep deployment. USE FOR: designing, deploying, or operating containerised workloads on Azure Container Apps.compatibility: Requires Azure CLI with containerapp extension, Docker
---

# Azure Container Apps Skill

> **MUST:** All Container Apps MUST use Managed Identity for Azure service access.
> DO NOT store credentials in environment variables or container images.

---

## Quick Reference

| Capability               | Description                                                               |
| ------------------------ | ------------------------------------------------------------------------- |
| Revision Management      | Blue/green and canary deployments via traffic splitting                   |
| Scaling Rules            | HTTP, queue-based, custom (KEDA) auto-scaling configuration               |
| Dapr Integration         | Sidecar-based service invocation, state, pub/sub, bindings                |
| Managed Identity         | System or user-assigned identity for Azure service authentication         |
| Service Connectors       | Managed connections to Azure databases, storage, and caches               |
| Ingress Configuration    | External/internal ingress, custom domains, session affinity, mTLS         |
| Jobs                     | Event-driven, scheduled, and manual job execution                         |
| Dynamic Sessions         | Hyper-V isolated sandboxed code execution for AI/LLM workloads            |
| Init Containers          | Run setup tasks before application containers start                       |
| Managed OTel Agent       | Environment-level OpenTelemetry routing to App Insights, Datadog, or OTLP |
| Policy-driven Resiliency | Service-to-service retry, circuit breaker, and timeout policies (preview) |
| GPU Workload Profiles    | Dedicated GPU compute for ML inference and training (preview)             |

---

## Currency and verification gates

- Last reviewed: **2026-04-03**
- Verify API versions in Azure template docs before release updates.
- Prefer immutable image tags (commit SHA or semantic version), never `latest` in production manifests.
- Re-validate KEDA scaler auth support (managed identity vs secret) per scaler type before implementation.

### Known Pitfalls

| Area                       | Pitfall                                                                                       | Mitigation                                                                                        |
| -------------------------- | --------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------- |
| `minReplicas: 0` with HTTP | Scale-to-zero causes cold start latency; first request may timeout                            | Set `minReplicas: 1` for user-facing HTTP services; use `0` only for event-driven processing      |
| `imagePullPolicy` default  | Default `IfNotPresent` caches stale images when using mutable tags                            | Always use `imagePullPolicy: Always` or immutable image tags (commit SHA)                         |
| KEDA scaler auth           | Not all KEDA scalers support managed identity authentication; some require connection strings | Verify scaler-specific auth support in [KEDA docs](https://keda.sh/docs/scalers/) per scaler type |
| Cron expressions in UTC    | Schedule cron expressions are always evaluated in UTC, not local time                         | Convert desired local schedule to UTC before configuring cron triggers                            |
| Jobs don't support ingress | Container App Jobs have no ingress or DNS name; cannot receive HTTP requests                  | Use Container Apps (not Jobs) for HTTP workloads; jobs are for batch/event-driven processing      |
| Jobs don't support Dapr    | Container App Jobs cannot use Dapr sidecar for service invocation, pub/sub, or state          | Use Container Apps (not Jobs) for Dapr workloads; jobs communicate via HTTP/SDK directly          |
| Revision traffic split     | Deploying a new revision with 100% traffic before health check passes causes downtime         | Use traffic splitting with `latestRevision: false`; validate new revision before shifting traffic |
| ACA Express limitations    | Express mode lacks managed identity, secrets, Dapr, KEDA, jobs, health probes, sidecar, VNet  | Use standard Container Apps environments for anything beyond simple HTTP-only web apps            |

---

## Standards

| Standard                                                  | Purpose                         |
| --------------------------------------------------------- | ------------------------------- |
| [Container App Design](standards/container-app-design.md) | Architecture and configuration  |
| [Scaling & Performance](standards/scaling.md)             | Auto-scaling rules and patterns |
| [Checklist](standards/checklist.md)                       | Validation checklist            |

---

## Actions

| Action                                                  | When to use                          |
| ------------------------------------------------------- | ------------------------------------ |
| [Deploy Container App](actions/deploy-container-app.md) | Creating or updating a Container App |

---

## Container App Environment

```bicep
resource containerAppEnv 'Microsoft.App/managedEnvironments@2025-07-01' = {
  name: 'cae-${environmentName}'
  location: location
  properties: {
    appLogsConfiguration: {
      destination: 'log-analytics'
      logAnalyticsConfiguration: {
        customerId: logAnalytics.properties.customerId
        sharedKey: logAnalytics.listKeys().primarySharedKey
      }
    }
    workloadProfiles: [
      { name: 'Consumption', workloadProfileType: 'Consumption' }
    ]
  }
}
```

---

## Container App with Managed Identity

Assume an explicit image tag parameter in your Bicep file:

```bicep
param imageTag string
```

```bicep
resource containerApp 'Microsoft.App/containerApps@2025-07-01' = {
  name: 'ca-${serviceName}'
  location: location
  identity: {
    type: 'SystemAssigned'
  }
  properties: {
    managedEnvironmentId: containerAppEnv.id
    configuration: {
      activeRevisionsMode: 'Multiple'  // enables traffic splitting
      ingress: {
        external: true
        targetPort: 8080
        transport: 'http'
        traffic: [
          { latestRevision: true, weight: 100 }
        ]
      }
      registries: [
        {
          server: acr.properties.loginServer
          identity: 'system'  // pull images via managed identity
        }
      ]
    }
    template: {
      containers: [
        {
          name: serviceName
          image: '${acr.properties.loginServer}/${serviceName}:${imageTag}'
          resources: {
            cpu: json('0.5')
            memory: '1Gi'
          }
          env: [
            { name: 'APPLICATIONINSIGHTS_CONNECTION_STRING', value: appInsights.properties.ConnectionString }
          ]
        }
      ]
      scale: {
        minReplicas: 1
        maxReplicas: 10
        rules: [
          {
            name: 'http-scaling'
            http: { metadata: { concurrentRequests: '50' } }
          }
        ]
      }
    }
  }
}
```

---

## Health Probes

Configure liveness, readiness, and startup probes for every container to enable reliable scaling and self-healing.

```bicep
template: {
  containers: [
    {
      name: serviceName
      image: '${acr.properties.loginServer}/${serviceName}:${imageTag}'
      resources: {
        cpu: json('0.5')
        memory: '1Gi'
      }
      probes: [
        {
          type: 'liveness'
          httpGet: {
            path: '/healthz'
            port: 8080
          }
          periodSeconds: 10
          failureThreshold: 3
          initialDelaySeconds: 5
        }
        {
          type: 'readiness'
          httpGet: {
            path: '/ready'
            port: 8080
          }
          periodSeconds: 5
          failureThreshold: 3
          initialDelaySeconds: 3
        }
        {
          type: 'startup'
          httpGet: {
            path: '/healthz'
            port: 8080
          }
          periodSeconds: 5
          failureThreshold: 30
          initialDelaySeconds: 0
        }
      ]
    }
  ]
}
```

> **Startup probes prevent premature liveness kills:** Use a startup probe with a high
> `failureThreshold` for containers that take time to initialize (e.g., loading ML models,
> running EF Core migrations). The liveness probe does not run until the startup probe succeeds.

---

## Key Vault Secret References

Source secrets from Azure Key Vault via managed identity instead of storing values inline.
The platform automatically retrieves and injects secret values.

```bicep
resource containerApp 'Microsoft.App/containerApps@2025-07-01' = {
  name: 'ca-${serviceName}'
  location: location
  identity: {
    type: 'UserAssigned'
    userAssignedIdentities: {
      '${managedIdentity.id}': {}
    }
  }
  properties: {
    managedEnvironmentId: containerAppEnv.id
    configuration: {
      secrets: [
        {
          name: 'db-connection-string'
          keyVaultUrl: 'https://${keyVault.name}${environment().suffixes.keyvaultDns}/secrets/db-connection'
          identity: managedIdentity.id
        }
        {
          name: 'api-key'
          keyVaultUrl: 'https://${keyVault.name}${environment().suffixes.keyvaultDns}/secrets/api-key'
          identity: managedIdentity.id
        }
      ]
    }
    template: {
      containers: [
        {
          name: serviceName
          image: '${acr.properties.loginServer}/${serviceName}:${imageTag}'
          env: [
            { name: 'ConnectionStrings__Default', secretRef: 'db-connection-string' }
            { name: 'API_KEY', secretRef: 'api-key' }
          ]
        }
      ]
    }
  }
}
```

> **RBAC requirement:** The managed identity must have the `Key Vault Secrets User` role
> on the Key Vault. See [secret-management](../secret-management/SKILL.md).

### Secrets Volume Mounts

Mount secrets as files instead of environment variables when the consuming application
reads configuration from the filesystem (e.g., TLS certificates, config files).

```bicep
configuration: {
  secrets: [
    {
      name: 'tls-cert'
      keyVaultUrl: 'https://${keyVault.name}${environment().suffixes.keyvaultDns}/secrets/tls-cert'
      identity: managedIdentity.id
    }
  ]
}
template: {
  containers: [
    {
      name: serviceName
      image: '${acr.properties.loginServer}/${serviceName}:${imageTag}'
      volumeMounts: [
        { volumeName: 'secrets-vol', mountPath: '/mnt/secrets' }
      ]
    }
  ]
  volumes: [
    {
      name: 'secrets-vol'
      storageType: 'Secret'
    }
  ]
}
```

> All secrets defined in the configuration are mounted as individual files in the volume.
> File names match the secret names.

---

## Scaling Rules

| Trigger       | Use Case                     | Configuration Key                 |
| ------------- | ---------------------------- | --------------------------------- |
| HTTP          | Web APIs, frontends          | `concurrentRequests`              |
| Azure Queue   | Background processing        | `queueName`, `queueLength`        |
| Service Bus   | Message-driven microservices | `queueName`, `messageCount`       |
| Custom (KEDA) | Any KEDA-supported scaler    | KEDA scaler metadata              |
| Cron/Schedule | Scheduled batch jobs         | `start`, `end`, `desiredReplicas` |

### Queue-Based Scaling

> **Warning — cold start with HTTP:** Setting `minReplicas: 0` for HTTP-serving
> Container Apps causes cold start delays of **5–15 seconds** when scaling from zero.
> Users experience timeouts or slow first requests. Use `minReplicas: 1` (or higher)
> for any app serving HTTP traffic. Reserve `minReplicas: 0` for background processors
> and queue consumers only.

```bicep
scale: {
  minReplicas: 0  // scale to zero when queue is empty
  maxReplicas: 20
  rules: [
    {
      name: 'queue-scaling'
      azureQueue: {
        queueName: 'work-items'
        queueLength: 10
        auth: [
          { secretRef: 'queue-connection', triggerParameter: 'connection' }
        ]
      }
    }
  ]
}
```

> **Prefer identity-based auth for KEDA scalers:** The `secretRef` approach above requires
> storing a connection string as a Container Apps secret. Where supported, use managed
> identity authentication with KEDA scalers instead. For Azure Storage Queue, the KEDA
> `azure-queue` scaler supports `accountName` + managed identity via pod identity or
> workload identity, eliminating secret management for the scaler.

````

---

## Revision Management

### Traffic Splitting (Canary)

```bash
# Deploy new revision
az containerapp update --name $appName --resource-group $rg \
  --image "${registry}/${image}:v2"

# Split traffic: 90% stable, 10% canary
az containerapp ingress traffic set --name $appName --resource-group $rg \
  --revision-weight "ca-api--stable=90" "ca-api--canary=10"

# Promote canary to 100%
az containerapp ingress traffic set --name $appName --resource-group $rg \
  --revision-weight "ca-api--canary=100"
````

---

## Dapr Integration

```bicep
configuration: {
  dapr: {
    enabled: true
    appId: serviceName
    appPort: 8080
    appProtocol: 'http'
  }
}
```

### Service Invocation

```csharp
// Call another Container App via Dapr
var client = DaprClient.CreateInvokerHttpClient(appId: "order-service");
var response = await client.GetAsync("/api/orders/123");
```

---

## Init Containers

Init containers run to completion before application containers start. Use them for
setup tasks that must succeed before the app serves traffic.

```bicep
template: {
  initContainers: [
    {
      name: 'db-migrate'
      image: '${acr.properties.loginServer}/migrations:${imageTag}'
      resources: {
        cpu: json('0.25')
        memory: '0.5Gi'
      }
      command: [ 'dotnet', 'ef', 'database', 'update' ]
      env: [
        { name: 'ConnectionStrings__Default', secretRef: 'db-connection-string' }
      ]
    }
  ]
  containers: [
    {
      name: serviceName
      image: '${acr.properties.loginServer}/${serviceName}:${imageTag}'
      resources: {
        cpu: json('0.5')
        memory: '1Gi'
      }
    }
  ]
}
```

Common use cases:

- Database schema migrations
- Configuration or certificate pre-fetching
- Waiting for dependent services to become available
- Populating shared volume mounts with seed data

> **Init containers block app startup:** If an init container fails, the replica is restarted.
> Keep init containers fast and idempotent to avoid blocking deployments.

---

## Session Affinity

Route all requests from the same client to the same Container App replica. Required for
stateful workloads that store in-memory session state.

```bicep
ingress: {
  external: true
  targetPort: 8080
  transport: 'http'
  stickySessions: {
    affinity: 'sticky'
  }
}
```

> **Prefer stateless design:** Session affinity reduces load balancing effectiveness and
> complicates scaling. Use external session stores (Redis, Cosmos DB) instead when possible.
> Session affinity is appropriate for legacy apps that cannot be refactored.

---

## Policy-driven Resiliency (Preview)

Configure service-to-service retry, circuit breaker, and timeout policies at the platform level
without code changes. Policies apply to outbound requests from a Container App.

```bicep
resource resiliencyPolicy 'Microsoft.App/containerApps/resiliencyPolicies@2025-07-01' = {
  parent: containerApp
  name: 'default-resiliency'
  properties: {
    httpRetryPolicy: {
      maxRetries: 3
      retryBackOff: {
        initialDelayInMilliseconds: 1000
        maxIntervalInMilliseconds: 10000
      }
      matches: {
        headers: []
        httpStatusCodes: [ 502, 503, 504 ]
      }
    }
    timeoutPolicy: {
      responseTimeoutInSeconds: 30
      connectionTimeoutInSeconds: 5
    }
    circuitBreakerPolicy: {
      consecutiveErrors: 5
      intervalInSeconds: 10
      maxEjectionPercent: 50
    }
  }
}
```

> **Preview status:** Resiliency policies are in public preview. Verify API version support
> before production use.

---

## Container Apps Jobs

Jobs are Container Apps that run to completion and exit, unlike services that run continuously.
Use the `Microsoft.App/jobs` resource type (not `Microsoft.App/containerApps`).

> **Warning — Jobs do NOT support Dapr or ingress:** Container App Jobs cannot use the Dapr
> sidecar for service invocation, pub/sub, or state management. Jobs also have no ingress or
> DNS name. Use Container Apps (not Jobs) for workloads requiring Dapr or HTTP endpoints.

### Required Permissions

| Role                           | Purpose                                   |
| ------------------------------ | ----------------------------------------- |
| `Container Apps Contributor`   | Create and manage container apps and jobs |
| `Monitoring Reader` (optional) | View monitoring data for jobs             |

For granular control, create a custom role with these actions:

- `microsoft.app/jobs/start/action`
- `microsoft.app/jobs/read`
- `microsoft.app/jobs/execution/read`

### Trigger Types

| Trigger      | When to Use                                                | Configuration Key                       |
| ------------ | ---------------------------------------------------------- | --------------------------------------- |
| **Manual**   | On-demand tasks: data migrations, ad-hoc processing        | `manualTriggerConfig`                   |
| **Schedule** | Recurring tasks: nightly reports, hourly cleanup           | `scheduleTriggerConfig.cronExpression`  |
| **Event**    | Reactive tasks: queue drain, blob processing, webhook work | `eventTriggerConfig` + KEDA scale rules |

### Key Configuration Properties

| Property                 | Description                                             | Default |
| ------------------------ | ------------------------------------------------------- | ------- |
| `replicaTimeout`         | Max seconds a replica can run before forced termination | 1800    |
| `replicaRetryLimit`      | Number of retries for a failed replica                  | 0       |
| `parallelism`            | Number of replicas to run in parallel per execution     | 1       |
| `replicaCompletionCount` | Number of replicas that must complete for job success   | 1       |

### Scheduled Job (Bicep)

```bicep
resource scheduledJob 'Microsoft.App/jobs@2025-07-01' = {
  name: 'job-${jobName}-scheduled'
  location: location
  identity: {
    type: 'SystemAssigned'
  }
  properties: {
    environmentId: containerAppEnv.id
    configuration: {
      triggerType: 'Schedule'
      replicaTimeout: 1800
      replicaRetryLimit: 1
      scheduleTriggerConfig: {
        cronExpression: '0 2 * * *'  // Daily at 2:00 AM UTC
        parallelism: 1
        replicaCompletionCount: 1
      }
      registries: [
        {
          server: acr.properties.loginServer
          identity: 'system'
        }
      ]
    }
    template: {
      containers: [
        {
          name: jobName
          image: '${acr.properties.loginServer}/${jobName}:${imageTag}'
          resources: {
            cpu: json('0.5')
            memory: '1Gi'
          }
          env: [
            { name: 'APPLICATIONINSIGHTS_CONNECTION_STRING', value: appInsights.properties.ConnectionString }
          ]
        }
      ]
    }
  }
}
```

> **Warning — cron expressions are UTC:** Schedule cron expressions are always evaluated
> in UTC. A job configured as `0 2 * * *` runs at 2:00 AM UTC, not local time. Account for
> timezone offsets (e.g., NZST is UTC+12, so 2:00 AM NZST = `0 14 * * *` UTC of the
> previous day).

### Event-Driven Job (Bicep)

```bicep
resource eventJob 'Microsoft.App/jobs@2025-07-01' = {
  name: 'job-${jobName}-event'
  location: location
  identity: {
    type: 'UserAssigned'
    userAssignedIdentities: {
      '${managedIdentity.id}': {}
    }
  }
  properties: {
    environmentId: containerAppEnv.id
    configuration: {
      triggerType: 'Event'
      replicaTimeout: 600
      replicaRetryLimit: 2
      eventTriggerConfig: {
        scale: {
          minExecutions: 0
          maxExecutions: 10
          pollingInterval: 30
          rules: [
            {
              name: 'queue-trigger'
              type: 'azure-servicebus'
              metadata: {
                queueName: 'work-items'
                messageCount: '5'
                namespace: serviceBusNamespace.name
              }
              auth: [
                {
                  secretRef: 'sb-connection'
                  triggerParameter: 'connection'
                }
              ]
            }
          ]
        }
      }
      secrets: [
        {
          name: 'sb-connection'
          value: serviceBusConnectionString
        }
      ]
      registries: [
        {
          server: acr.properties.loginServer
          identity: managedIdentity.id
        }
      ]
    }
    template: {
      containers: [
        {
          name: jobName
          image: '${acr.properties.loginServer}/${jobName}:${imageTag}'
          resources: {
            cpu: json('0.25')
            memory: '0.5Gi'
          }
        }
      ]
    }
  }
}
```

> **Prefer managed identity for KEDA triggers:** The example above uses `secretRef` for the
> KEDA scaler connection. Where the KEDA scaler supports identity-based auth (e.g., Service
> Bus with `azure-servicebus` scaler), use workload identity instead to eliminate secret
> management. See [identity-managed-identity](../identity-managed-identity/SKILL.md).

### Manual Job (Bicep)

```bicep
resource manualJob 'Microsoft.App/jobs@2025-07-01' = {
  name: 'job-${jobName}-manual'
  location: location
  identity: {
    type: 'SystemAssigned'
  }
  properties: {
    environmentId: containerAppEnv.id
    configuration: {
      triggerType: 'Manual'
      replicaTimeout: 3600
      replicaRetryLimit: 0
      manualTriggerConfig: {
        parallelism: 1
        replicaCompletionCount: 1
      }
      registries: [
        {
          server: acr.properties.loginServer
          identity: 'system'
        }
      ]
    }
    template: {
      containers: [
        {
          name: jobName
          image: '${acr.properties.loginServer}/${jobName}:${imageTag}'
          resources: {
            cpu: json('1')
            memory: '2Gi'
          }
        }
      ]
    }
  }
}
```

### Starting a Manual Job

```bash
# Start a manual job execution
az containerapp job start --name "job-migrate-manual" --resource-group $rg

# List execution history
az containerapp job execution list --name "job-migrate-manual" --resource-group $rg
```

### Overriding Job Configuration Per Execution

Override the container config (image, env vars, command) for a single execution without
changing the job's default configuration.

```bash
# Export current job template
az containerapp job show --name "job-migrate-manual" --resource-group $rg \
  --query "properties.template" --output yaml > job-template.yaml

# Edit job-template.yaml to override env vars, image, or command, then start:
az containerapp job start --name "job-migrate-manual" --resource-group $rg \
  --yaml job-template.yaml
```

> **Warning — full template replacement:** When you override configuration, the job's entire
> template is replaced with the provided YAML. Include ALL required settings (image, resources,
> env vars) in the override file, not just the fields you want to change.

### Jobs Best Practices

1. **Set `replicaTimeout` conservatively** — if the job exceeds the timeout, the replica is
   terminated. Set it to at least 2x your expected p99 duration. Maximum is 86400 (24 hours).
2. **Use `replicaRetryLimit` for transient failures** — set to 1–3 for jobs that may hit
   transient errors (network, throttling). Keep at 0 for jobs that are not idempotent.
3. **Jobs share the Container App Environment** — jobs and apps in the same environment share
   VNet and Log Analytics. Use a shared environment to reduce overhead. Note that jobs
   cannot use Dapr even though app-level Dapr components exist in the environment.
4. **Exit code matters** — the container must exit with code 0 for success. Non-zero exit
   codes trigger a retry (if `replicaRetryLimit > 0`) or mark the execution as failed.
5. **Right-size `maxExecutions` for event-driven jobs** — setting `maxExecutions` too high
   can cause resource contention. Start with 5–10 and increase based on observed throughput.
6. **Use `pollingInterval` to control scale responsiveness** — default is 30 seconds. Lower
   values (e.g., 10s) improve reaction time but increase KEDA API calls to the event source.
7. **Log to stdout/stderr** — job container logs are captured by Log Analytics when output
   is written to stdout/stderr. Do not rely on file-based logging inside the container.

> **Warning — jobs do not support ingress:** Container App Jobs do not have ingress or
> custom domains. They are fire-and-forget compute. If you need to expose an HTTP endpoint
> that also does background processing, use a Container App (not a job) with queue-based
> scaling.

> **Warning — `replicaCompletionCount` vs `parallelism`:** `replicaCompletionCount` is the
> total number of replicas that must complete successfully, while `parallelism` is how many
> run concurrently. For example, `replicaCompletionCount: 10` with `parallelism: 5` runs
> 10 replicas in two batches of 5. Each replica gets the same container config — use
> environment variables or queue-based work distribution to differentiate work per replica.

---

## Dynamic Sessions

Dynamic sessions provide fast, Hyper-V isolated sandboxed environments for running
untrusted code. Two session pool types are available:

| Pool Type            | When to Use                                                       | Image Required |
| -------------------- | ----------------------------------------------------------------- | -------------- |
| **Code Interpreter** | LLM-generated code, user-submitted scripts, educational sandboxes | No (built-in)  |
| **Custom Container** | Custom runtimes, proprietary tools, specialized OS/library needs  | Yes            |

### Key Concepts

- **Session Pool** — prewarmed pool of ready-to-use sessions with near-instant allocation
- **Session** — ephemeral, isolated execution environment (auto-destroyed after cooldown)
- **Session Identifier** — string you define (user ID, conversation ID) that routes requests to the same session

### Session Pool (Bicep)

```bicep
resource sessionPool 'Microsoft.App/sessionPools@2025-07-01' = {
  name: 'sp-${poolName}'
  location: location
  properties: {
    poolManagementType: 'Dynamic'
    containerType: 'PythonLTS'  // 'PythonLTS' for code interpreter, 'CustomContainer' for BYOC
    dynamicPoolConfiguration: {
      cooldownPeriodInSeconds: 300
    }
    scaleConfiguration: {
      maxConcurrentSessions: 100
    }
  }
}
```

### Authentication

Sessions require Microsoft Entra tokens. The calling identity must have the
`Azure ContainerApps Session Executor` role on the session pool.

### LLM Framework Integrations

| Framework       | Package                                    |
| --------------- | ------------------------------------------ |
| LangChain       | `langchain-azure-dynamic-sessions`         |
| LlamaIndex      | `llama-index-tools-azure-code-interpreter` |
| Semantic Kernel | `semantic-kernel` (0.9.8-b1+)              |

### Code Execution (REST API)

```http
POST https://<REGION>.dynamicsessions.io/subscriptions/<SUB>/resourceGroups/<RG>/sessionPools/<POOL>/executions?api-version=2025-10-02-preview&identifier=<SESSION_ID>
Content-Type: application/json
Authorization: Bearer <TOKEN>

{
  "properties": {
    "codeInputType": "inline",
    "executionType": "synchronous",
    "code": "print('Hello from sandboxed session')"
  }
}
```

> **Session identifiers are sensitive:** The session identifier controls which session
> a request routes to. Ensure your application maps each user/tenant to their own
> session identifier to prevent unauthorized data access across sessions.

> **File upload limit:** Session file uploads are capped at 128 MB. Exceeding this
> returns HTTP 413.

---

## Managed OpenTelemetry Agent

The environment-level managed OTel agent routes traces, logs, and metrics to one or more
destinations without running your own collector. No additional compute cost.

### Destinations

| Destination            | Traces | Metrics | Logs |
| ---------------------- | ------ | ------- | ---- |
| Azure App Insights     | Yes    | No      | Yes  |
| Datadog                | Yes    | Yes     | Yes  |
| OTLP endpoint (custom) | Yes    | Yes     | Yes  |

### Environment Configuration (Bicep)

```bicep
resource containerAppEnv 'Microsoft.App/managedEnvironments@2025-07-01' = {
  name: 'cae-${environmentName}'
  location: location
  properties: {
    appInsightsConfiguration: {
      connectionString: appInsights.properties.ConnectionString
    }
    openTelemetryConfiguration: {
      tracesConfiguration: {
        destinations: [ 'appInsights' ]
        includeDapr: true
      }
      logsConfiguration: {
        destinations: [ 'appInsights' ]
      }
      metricsConfiguration: {
        destinations: [ 'appInsights' ]
        includeKeda: true
      }
    }
  }
}
```

### Auto-injected Environment Variables

The agent injects these variables into every container at runtime:

| Variable                      | Purpose                           |
| ----------------------------- | --------------------------------- |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | Base OTLP endpoint URL            |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | Transport protocol (`grpc`)       |
| `OTEL_RESOURCE_ATTRIBUTES`    | Container app resource attributes |

> **SDK instrumentation still required:** The managed agent routes data but does not
> generate it. You must instrument your application with the OpenTelemetry SDK to
> emit traces, metrics, and logs.

### Known Limitations

- App Insights destination does not accept metrics (traces and logs only)
- Agent runs as a single replica (no HA configuration)
- Only gRPC transport protocol is supported
- Agent config is environment-level (cannot split by app in the same environment)

---

## Environment Modes: Standard vs Express

### ACA Express (Private Preview)

Express mode provides the fastest deployment path with opinionated defaults and scale-from-zero.
However, it has significant limitations that make it unsuitable for most production architectures.

| Feature                   | Standard        | Express              |
| ------------------------- | --------------- | -------------------- |
| Scale to zero             | Yes             | Yes                  |
| Managed Identity          | Yes             | **No**               |
| Secrets / Key Vault       | Yes             | **No**               |
| Dapr integration          | Yes             | **No**               |
| KEDA autoscaling          | Yes             | **No**               |
| Jobs                      | Yes             | **No**               |
| Health probes             | Yes             | **No**               |
| Init / sidecar containers | Yes             | **No**               |
| VNet integration          | Yes             | **No**               |
| OpenTelemetry agent       | Yes             | **No**               |
| Custom domains            | Yes             | **No**               |
| App-to-app communication  | Yes             | **No**               |
| GPU profiles              | Yes             | **No**               |
| Region availability       | All ACA regions | West Central US only |

**Decision: Use Standard environments** for any workload requiring identity, secrets, Dapr,
event-driven scaling, jobs, or service-to-service communication. Express is appropriate only
for rapid prototyping of simple HTTP-only web apps and AI gateway frontends.

---

## GPU Workload Profiles (Preview)

Dedicated GPU workload profiles enable ML inference, model training, and GPU-accelerated
compute within the Container Apps environment.

```bicep
resource containerAppEnv 'Microsoft.App/managedEnvironments@2025-07-01' = {
  name: 'cae-${environmentName}'
  location: location
  properties: {
    workloadProfiles: [
      { name: 'Consumption', workloadProfileType: 'Consumption' }
      { name: 'gpu-t4', workloadProfileType: 'NC4as-T4-V3' }
    ]
  }
}
```

```bicep
resource gpuApp 'Microsoft.App/containerApps@2025-07-01' = {
  name: 'ca-${serviceName}-gpu'
  location: location
  properties: {
    managedEnvironmentId: containerAppEnv.id
    workloadProfileName: 'gpu-t4'
    template: {
      containers: [
        {
          name: serviceName
          image: '${acr.properties.loginServer}/${serviceName}:${imageTag}'
          resources: {
            cpu: json('4')
            memory: '28Gi'
          }
        }
      ]
    }
  }
}
```

> **Preview status:** GPU workload profiles are in public preview. Not all regions support
> GPU SKUs. Verify availability in your target region before planning workloads.

> **Cost implications:** GPU profiles use dedicated compute (billed per instance), not
> consumption-based. GPU apps do not scale to zero.

---

## azure.yaml Integration (azd)

```yaml
services:
  api:
    project: ./src/api
    host: containerapp
    language: csharp
    docker:
      path: ./Dockerfile
      context: .
```

---

## Principles

1. **Managed Identity everywhere** — use `identity: 'system'` for ACR pulls and Azure SDK auth.
2. **Scale to zero** — configure `minReplicas: 0` for queue processors and batch jobs (not HTTP endpoints).
3. **Revision-based deployments** — use traffic splitting for safe rollouts, never in-place updates.
   Use `activeRevisionsMode: 'Multiple'` to enable traffic splitting. `'Single'` mode
   deactivates old revisions immediately — no gradual rollout or instant rollback.
4. **Right-size containers** — start with 0.25 CPU / 0.5Gi and scale out, not up.
5. **Structured health probes** — configure liveness and readiness probes for every container.
6. **Environment per stage** — separate Container App Environments for dev, staging, production.

---

## References

- [Azure Container Apps overview](https://learn.microsoft.com/azure/container-apps/overview)
- [Scaling in Container Apps](https://learn.microsoft.com/azure/container-apps/scale-app)
- [Dapr integration](https://learn.microsoft.com/azure/container-apps/dapr-overview)
- [Managed identity in Container Apps](https://learn.microsoft.com/azure/container-apps/managed-identity)
- [Container Apps Bicep reference](https://learn.microsoft.com/azure/templates/microsoft.app/containerapps)
- [Container Apps Jobs](https://learn.microsoft.com/azure/container-apps/jobs)
- [Dynamic sessions](https://learn.microsoft.com/azure/container-apps/sessions)
- [Code interpreter sessions](https://learn.microsoft.com/azure/container-apps/sessions-code-interpreter)
- [OpenTelemetry agents](https://learn.microsoft.com/azure/container-apps/opentelemetry-agents)
- [Resiliency policies](https://learn.microsoft.com/azure/container-apps/resiliency)
- [Health probes](https://learn.microsoft.com/azure/container-apps/health-probes)
- [Init containers](https://learn.microsoft.com/azure/container-apps/containers#init-containers)
- [Key Vault secret references](https://learn.microsoft.com/azure/container-apps/manage-secrets#reference-secret-from-key-vault)
- [GPU workload profiles](https://learn.microsoft.com/azure/container-apps/workload-profiles-overview)

---

## Related Skills

- **managing-azure-dev-cli-lifecycle** — azd hosting with `containerapp` target
- **observability-monitoring** — Application Insights integration
- **identity-managed-identity** — Managed Identity patterns
- **azure-defaults** — Region and tagging standards
