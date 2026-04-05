# Service Defaults Standard

The Service Defaults project is a shared class library referenced by every service in your
Aspire solution. It contains a single extension method — `AddServiceDefaults()` — that
wires production-grade cross-cutting concerns.

---

## What `AddServiceDefaults()` Registers

### 1. OpenTelemetry

```csharp
public static IHostApplicationBuilder ConfigureOpenTelemetry(this IHostApplicationBuilder builder)
{
    builder.Logging.AddOpenTelemetry(logging =>
    {
        logging.IncludeFormattedMessage = true;
        logging.IncludeScopes = true;
    });

    builder.Services.AddOpenTelemetry()
        .WithMetrics(metrics =>
        {
            metrics.AddAspNetCoreInstrumentation()
                   .AddHttpClientInstrumentation()
                   .AddRuntimeInstrumentation();
        })
        .WithTracing(tracing =>
        {
            tracing.AddSource(builder.Environment.ApplicationName)
                   .AddAspNetCoreInstrumentation()
                   .AddHttpClientInstrumentation();
        });

    builder.AddOpenTelemetryExporters();
    return builder;
}
```

This replaces manual OTEL registration. The exporter target (OTLP endpoint) is injected
by the AppHost during local dev and by environment variables in production.

### 2. Health Checks

```csharp
public static IHostApplicationBuilder AddDefaultHealthChecks(this IHostApplicationBuilder builder)
{
    builder.Services.AddHealthChecks()
        .AddCheck("self", () => HealthCheckResult.Healthy());

    return builder;
}
```

Map the endpoints in `Program.cs`:

```csharp
app.MapDefaultEndpoints();

// Which typically does:
public static WebApplication MapDefaultEndpoints(this WebApplication app)
{
    app.MapHealthChecks("/health/live", new HealthCheckOptions
    {
        Predicate = _ => false // No dependency checks — just "am I alive?"
    });

    app.MapHealthChecks("/health/ready");

    return app;
}
```

### 3. Resilience

```csharp
builder.Services.ConfigureHttpClientDefaults(http =>
{
    http.AddStandardResilienceHandler();
});
```

This applies retry, circuit-breaker, and timeout policies to **every** `HttpClient`
created via `IHttpClientFactory`. Override per-client when defaults conflict:

> [!IMPORTANT]
> `AddStandardResilienceHandler()` applies default retry and circuit breaker policies to ALL HTTP clients. For payment APIs, long-polling endpoints, or non-idempotent operations, override the defaults per-client:
> ```csharp
> builder.Services.ConfigureHttpClientDefaults(http => {
>     http.AddStandardResilienceHandler();
> });
> // Override for payment service — disable retries on non-idempotent calls
> builder.Services.AddHttpClient("PaymentService")
>     .AddStandardResilienceHandler(options => {
>         options.Retry.MaxRetryAttempts = 0; // Non-idempotent: no retries
>         options.AttemptTimeout.Timeout = TimeSpan.FromSeconds(30);
>     });
> ```

```csharp
builder.Services.AddHttpClient("payment-gateway")
    .ConfigureHttpClient(c => c.Timeout = TimeSpan.FromSeconds(5))
    .AddStandardResilienceHandler(options =>
    {
        options.Retry.MaxRetryAttempts = 2;
    });
```

### 4. Service Discovery

```csharp
builder.Services.AddServiceDiscovery();
builder.Services.ConfigureHttpClientDefaults(http =>
{
    http.AddServiceDiscovery();
});
```

Enables `http://servicename` URIs in `HttpClient` calls. The AppHost injects endpoints
during local dev; in production, configure via environment variables or DNS.

---

## Customizing Service Defaults

The Service Defaults project is **your code** — modify it freely:

- Add custom health checks for domain-specific readiness
- Add additional OTEL instrumentation (e.g., EF Core, MassTransit)
- Adjust resilience defaults for your SLA requirements
- Register shared middleware or DI registrations

```csharp
public static IHostApplicationBuilder AddServiceDefaults(this IHostApplicationBuilder builder)
{
    builder.ConfigureOpenTelemetry();
    builder.AddDefaultHealthChecks();
    builder.Services.AddServiceDiscovery();
    builder.Services.ConfigureHttpClientDefaults(http =>
    {
        http.AddServiceDiscovery();
        http.AddStandardResilienceHandler();
    });

    // Your customizations
    builder.Services.AddHealthChecks()
        .AddNpgSql(builder.Configuration.GetConnectionString("postgres")!);

    return builder;
}
```

---

## Production Behavior

In production (without AppHost running):

- OTLP exporter sends to whatever `OTEL_EXPORTER_OTLP_ENDPOINT` is set to (e.g., Application Insights)
- Service discovery falls back to configuration-based or DNS-based resolution
- Resilience handlers run as configured — no AppHost involvement
- Health checks respond to Kubernetes probes or load balancer health pings
