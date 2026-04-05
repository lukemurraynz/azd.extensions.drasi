# Distributed Tracing Standard

## Overview

Distributed tracing provides end-to-end visibility across service boundaries.
OpenTelemetry is the instrumentation standard; Application Insights is the Azure backend.

---

## Trace Propagation

### W3C Trace Context (Default)

OpenTelemetry uses W3C `traceparent` and `tracestate` headers by default.
No custom header propagation needed for HTTP calls.

```
traceparent: 00-{traceId}-{spanId}-{flags}
```

### Cross-Boundary Propagation

For non-HTTP boundaries (Service Bus, Event Grid, storage queues), embed
trace context in message properties:

```csharp
// Publish — embed context
var message = new ServiceBusMessage(payload);
if (Activity.Current is { } act)
{
    message.ApplicationProperties["Diagnostic-Id"] = act.Id;
}

// Consume — restore context
using var activity = new Activity("Process " + message.Subject);
if (message.ApplicationProperties.TryGetValue("Diagnostic-Id", out var diagId))
{
    activity.SetParentId((string)diagId);
}
activity.Start();
```

---

## Span Design Guidelines

### Naming

| Pattern                   | Example                     |
| ------------------------- | --------------------------- |
| `{verb} {noun}`           | `GET /api/orders`           |
| `{service}.{operation}`   | `OrderService.CreateOrder`  |
| `{messaging}.{operation}` | `ServiceBus.ProcessMessage` |

### Attributes

Add semantic attributes to spans for queryability:

```csharp
using var activity = activitySource.StartActivity("ProcessOrder");
activity?.SetTag("order.id", orderId);
activity?.SetTag("order.amount", amount);
activity?.SetTag("customer.region", region);
```

### Span Boundaries

Create spans at:

- Service entry points (HTTP handlers, message handlers)
- Outbound calls (HTTP clients, database, queue sends)
- Significant internal operations (batch processing, complex calculations)

Do **not** create spans for:

- Simple property access or mapping
- Iterations within a loop (use events instead)
- Trivial helper methods

---

## Application Map

Application Insights automatically builds an application map from dependency
tracking. Ensure all outbound calls use instrumented clients:

| Technology        | Instrumentation                                     |
| ----------------- | --------------------------------------------------- |
| HttpClient (.NET) | `AddHttpClientInstrumentation()` — automatic        |
| SQL Server        | `AddSqlClientInstrumentation()` — automatic         |
| Azure SDK         | Built-in OpenTelemetry support — automatic          |
| gRPC              | `AddGrpcClientInstrumentation()` — add explicitly   |
| Redis             | `AddRedisInstrumentation()` — add via StackExchange |

---

## Sampling

### Default Strategy

Use adaptive sampling in Application Insights to control volume without
losing critical data:

```csharp
builder.Services.AddOpenTelemetry()
    .UseAzureMonitor(options =>
    {
        options.SamplingRatio = 0.1f;  // 10% in high-volume production
    });
```

### Preserve Critical Traces

Never sample out:

- Error traces (5xx responses, exceptions)
- Traces exceeding latency thresholds
- Health check-triggered alerts

---

## Rules

1. Use W3C Trace Context — do not invent custom correlation headers.
2. Propagate trace context across all boundaries — HTTP, messaging, and storage.
3. Name spans with `{verb} {noun}` convention — avoid generic names like "Process".
4. Add business-relevant attributes — `order.id`, `tenant.id`, `feature.flag`.
5. Keep span depth reasonable — typically 3–5 levels per service.
6. Sample in production — 10–20% for high-throughput services; 100% in dev/staging.
