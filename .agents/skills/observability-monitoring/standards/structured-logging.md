# Structured Logging Standard

## Format

All application logs must be structured JSON. Never use string interpolation in log message templates.

### Required Fields

Every log entry must contain:

| Field         | Source        | Description                                    |
| ------------- | ------------- | ---------------------------------------------- |
| `Timestamp`   | Framework     | UTC ISO 8601 timestamp                         |
| `Level`       | Framework     | Trace/Debug/Information/Warning/Error/Critical |
| `Message`     | Developer     | Semantic template with placeholders            |
| `OperationId` | OpenTelemetry | Distributed trace correlation ID               |
| `SpanId`      | OpenTelemetry | Current span within the trace                  |
| `ServiceName` | Configuration | Identifies the emitting service                |

### Recommended Fields

| Field         | Description                            |
| ------------- | -------------------------------------- |
| `UserId`      | Authenticated user identifier (hashed) |
| `TenantId`    | Multi-tenant context                   |
| `Environment` | dev / staging / production             |
| `Version`     | Application version or commit SHA      |

---

## Log Level Guidelines

| Level       | When to Use                               | Example                                        |
| ----------- | ----------------------------------------- | ---------------------------------------------- |
| Trace       | Step-by-step execution detail (dev only)  | "Entering method {MethodName}"                 |
| Debug       | Diagnostic data useful during development | "Cache miss for key {CacheKey}"                |
| Information | Normal operational milestones             | "Order {OrderId} created successfully"         |
| Warning     | Recoverable anomaly; system continues     | "Retry {Attempt} for {ServiceName}"            |
| Error       | Operation failed; requires attention      | "Payment failed for Order {OrderId}: {Reason}" |
| Critical    | System-wide failure; data loss risk       | "Database connection pool exhausted"           |

### Level Configuration by Environment

| Environment | Minimum Level |
| ----------- | ------------- |
| Development | Debug         |
| Staging     | Information   |
| Production  | Information   |

> **Production cost control:** `Information` is a safe default, but high-traffic services
> (> 500 RPS) should raise the default to `Warning` and selectively enable `Information`
> only for application-specific categories. Configure per-category overrides in
> `appsettings.Production.json` and use adaptive sampling (see Observability SKILL.md).

---

## Correlation

### Automatic Correlation

OpenTelemetry propagates `TraceId` and `SpanId` automatically through HTTP headers (`traceparent`). Application Insights maps these to `OperationId` and `ParentId`.

### Manual Correlation

When crossing non-HTTP boundaries (queues, file drops), propagate context explicitly:

```csharp
// Publishing to a queue — embed trace context
var activity = Activity.Current;
message.Properties["traceparent"] = activity?.Id;

// Consuming from a queue — restore trace context
using var activity = new Activity("ProcessMessage");
if (message.Properties.TryGetValue("traceparent", out var traceParent))
{
    activity.SetParentId((string)traceParent);
}
activity.Start();
```

---

## Sensitive Data

Never log:

- Passwords, tokens, API keys, or secrets
- Full credit card numbers or SSNs
- Personally identifiable information (PII) without consent
- Connection strings with credentials

Use redaction or hashing for any user-identifying fields logged for diagnostics.

```csharp
// PII hashing pattern — hash identifiers before logging
using System.Security.Cryptography;

private static string HashPii(string value)
{
    var bytes = SHA256.HashData(System.Text.Encoding.UTF8.GetBytes(value));
    return Convert.ToHexString(bytes)[..12]; // First 12 chars sufficient for correlation
}

logger.LogInformation("User {HashedUserId} performed action", HashPii(userId));
```

---

## Rules

1. Use semantic message templates — `"Order {OrderId} processed"` not `$"Order {orderId} processed"`.
2. Include correlation IDs in every log entry — framework handles this when OpenTelemetry is configured.
3. Log at the boundary — entry/exit of service calls, not inside tight loops.
4. Set log levels per-category in configuration — never hardcode minimum levels.
5. Do not catch and log exceptions that will be logged by middleware — avoid duplicate logging.
