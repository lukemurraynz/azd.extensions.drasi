# Production Considerations

Understanding what ships to production and what stays in development is critical
for .NET Aspire adoption.

---

## What Ships to Production

| Component                     | Ships? | Notes                                                   |
| ----------------------------- | ------ | ------------------------------------------------------- |
| Service Defaults library      | Yes    | Standard `Microsoft.Extensions.*` APIs                  |
| Client integration NuGet pkgs | Yes    | Typed clients, health checks, OTEL instrumentation      |
| Your application code         | Yes    | `Program.cs` calls `AddServiceDefaults()` and `Add*()`  |
| AppHost project               | No     | Orchestrator for local dev and CI only                  |
| Aspire Dashboard              | No     | Dev-time OTEL viewer; use App Insights/Grafana in prod  |
| `aspire publish` output       | Maybe  | Generated IaC is a starting point, not production-grade |

---

## No Vendor Lock-In

Service Defaults uses standard .NET APIs:

- `Microsoft.Extensions.Diagnostics.HealthChecks`
- `Microsoft.Extensions.Http.Resilience` (Polly v8)
- `Microsoft.Extensions.ServiceDiscovery`
- `OpenTelemetry.Extensions.Hosting`

If you remove Aspire, you keep working code — you just wire these manually instead
of through `AddServiceDefaults()`.

---

## Risks and Mitigations

### 1. Global Resilience Override

`ConfigureHttpClientDefaults` with `AddStandardResilienceHandler()` applies retry,
circuit-breaker, and timeout to **all** `HttpClient` instances registered via
`IHttpClientFactory`.

**Risk:** Conflicts with services that have custom retry requirements (e.g., payment
gateways that must not retry, or streaming endpoints with long timeouts).

**Mitigation:** Override per named client:

```csharp
builder.Services.AddHttpClient("payment-gateway")
    .ConfigureHttpClient(c => c.Timeout = TimeSpan.FromSeconds(30))
    .RemoveAllResilienceHandlers()   // Remove global defaults
    .AddStandardResilienceHandler(o =>
    {
        o.Retry.MaxRetryAttempts = 0;  // No retries for payments
    });
```

### 2. Connection String Conventions

Aspire injects connection strings as `ConnectionStrings__<name>` environment variables.
In production without AppHost, you must provide these yourself.

**Risk:** Missing or misnamed connection strings cause startup failures.

**Mitigation:**

- Document all expected `ConnectionStrings__*` env vars in your deployment runbook
- Use `builder.Configuration.GetConnectionString("name")` which checks both `:` and `__` separators
- Add fail-fast validation at startup:

```csharp
var connectionString = builder.Configuration.GetConnectionString("postgres")
    ?? throw new InvalidOperationException("Missing ConnectionStrings:postgres");
```

### 3. AppHost-Generated Infrastructure

`aspire publish` generates Bicep, Kubernetes manifests, or Docker Compose files from
the AppHost resource graph. This output is **scaffold quality**.

**Risk:** Deploying generated IaC without review leads to missing security controls,
wrong SKUs, or absent network policies.

**Mitigation:**

- Treat generated IaC as a starting point — apply your Bicep/K8s standards
- Use generated manifests for reference, then maintain production IaC separately
- Review generated Bicep against `azure-deployment-preflight` standards

### 4. Service Discovery in Production

In local dev, the AppHost provides endpoint resolution. In production, service discovery
must be configured explicitly.

**Mitigation:** Use configuration-based service discovery in production:

```json
{
  "Services": {
    "api": {
      "https": ["https://api.myapp.com"]
    }
  }
}
```

Or rely on Kubernetes DNS / ACA internal DNS for service-to-service resolution.

---

## Environment-Specific Configuration

Use `builder.ExecutionContext.IsPublishMode` to distinguish between local dev and
deployment manifest generation:

```csharp
if (builder.ExecutionContext.IsPublishMode)
{
    // Production-specific configuration
    postgres.PublishAsAzurePostgresFlexibleServer();
}
else
{
    // Local dev — use container
    postgres.WithPgAdmin();
}
```

---

## Observability in Production

| Environment | OTEL Backend                   | Configuration                              |
| ----------- | ------------------------------ | ------------------------------------------ |
| Local dev   | Aspire Dashboard (auto)        | AppHost injects OTLP endpoint              |
| Production  | Application Insights / Grafana | Set `OTEL_EXPORTER_OTLP_ENDPOINT`          |
| CI/CD       | Optional — Dashboard or none   | Use `DistributedApplicationTestingBuilder` |

The OTLP exporter configured by Service Defaults sends to whatever endpoint is
configured — it does not depend on the Aspire Dashboard.
