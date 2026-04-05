# Action: Instrument Application

Add observability to a new or existing service.

---

## Step 1 — Add OpenTelemetry + Azure Monitor Packages

### .NET

```bash
dotnet add package Azure.Monitor.OpenTelemetry.AspNetCore
```

### Node.js

```bash
npm install @azure/monitor-opentelemetry
```

### Python

```bash
pip install azure-monitor-opentelemetry
```

---

## Step 2 — Configure the SDK

### .NET

```csharp
var builder = WebApplication.CreateBuilder(args);

builder.Services.AddOpenTelemetry()
    .UseAzureMonitor(options =>
    {
        options.ConnectionString = builder.Configuration["APPLICATIONINSIGHTS_CONNECTION_STRING"];
    })
    .WithTracing(tracing => tracing
        .AddAspNetCoreInstrumentation()
        .AddHttpClientInstrumentation()
        .AddSource("MyApp.*"))  // custom ActivitySources
    .WithMetrics(metrics => metrics
        .AddAspNetCoreInstrumentation()
        .AddHttpClientInstrumentation());
```

### Node.js

```typescript
// instrumentation.ts — must be imported before any other module
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

## Step 3 — Add Health Endpoints

```csharp
builder.Services.AddHealthChecks()
    .AddCheck("self", () => HealthCheckResult.Healthy());
    // Add dependency checks as needed:
    // .AddNpgSql(connectionString)
    // .AddAzureBlobStorage(connectionString)

app.MapHealthChecks("/healthz/live", new HealthCheckOptions
{
    Predicate = _ => false  // liveness — no dependency checks
});
app.MapHealthChecks("/healthz/ready");  // readiness — all checks
```

---

## Step 4 — Add Custom Spans for Business Operations

```csharp
private static readonly ActivitySource Source = new("MyApp.Orders");

public async Task<Order> CreateOrderAsync(CreateOrderRequest request)
{
    using var activity = Source.StartActivity("CreateOrder");
    activity?.SetTag("order.customer_id", request.CustomerId);
    activity?.SetTag("order.item_count", request.Items.Count);

    // ... business logic ...

    activity?.SetTag("order.id", order.Id);
    return order;
}
```

---

## Step 5 — Configure Structured Logging

```json
{
  "Logging": {
    "LogLevel": {
      "Default": "Information",
      "Microsoft.AspNetCore": "Warning",
      "Microsoft.EntityFrameworkCore.Database.Command": "Warning"
    }
  }
}
```

Ensure all log calls use semantic templates:

```csharp
logger.LogInformation("Order {OrderId} created with {ItemCount} items", order.Id, items.Count);
```

---

## Step 6 — Verify in Application Insights

1. Run the application and generate traffic.
2. Open Application Insights in the Azure Portal.
3. Verify: **Live Metrics** shows incoming requests.
4. Verify: **Application Map** shows service and dependencies.
5. Verify: **Transaction Search** shows correlated traces.
6. Verify: **Logs** → `AppRequests | take 10` returns data.

---

## Completion Criteria

- [ ] OpenTelemetry SDK initialised with Azure Monitor exporter
- [ ] Health endpoints responding at `/healthz/live` and `/healthz/ready`
- [ ] Custom spans emitting for key business operations
- [ ] Structured logs flowing to Application Insights
- [ ] Application Map shows expected service topology
