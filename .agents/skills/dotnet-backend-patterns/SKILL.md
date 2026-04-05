---
name: dotnet-backend-patterns
description: >-
  Production-ready patterns for .NET 10 / ASP.NET Core 10 / EF Core 10 / SignalR backend development. USE FOR: implement RFC 9457 error contracts, configure EF Core 10 migrations with PostgreSQL, set up SignalR hubs, apply Azure Identity managed identity, configure health checks, set up CORS, implement resilient HTTP clients.
---

# .NET Backend Patterns

Patterns for .NET 10 / ASP.NET Core 10 / EF Core 10 / SignalR backend development, specific to the Emergency Alerts stack.

---

## When to Use This Skill

- Implementing or reviewing ASP.NET Core controllers, middleware, and services
- Setting up or updating EF Core 10 database migrations (PostgreSQL / Npgsql)
- Configuring SignalR hubs with multi-replica reconnect safety
- Applying Azure Identity / Managed Identity for service-to-service auth
- Implementing RFC 9457 Problem Details error contracts
- Adding or modifying health checks (`/health/ready`, `/health/live`)
- Configuring CORS for the Fluent UI frontend
- Setting up resilient HTTP clients via `Microsoft.Extensions.Http.Resilience`

> **Aspire users:** If using .NET Aspire, `AddServiceDefaults()` replaces manual OTEL,
> health check, resilience, and service discovery setup described in this skill. See
> [dotnet-aspire](../dotnet-aspire/SKILL.md) for Aspire-specific patterns.

---

## 1. RFC 9457 Error Contract (REQUIRED)

All API errors must return `application/problem+json` with an `x-error-code` response header that matches the `errorCode` body extension field.

### Implementation

```csharp
// Helpers/ProblemDetailsHelper.cs
public static class ProblemDetailsHelper
{
    public static IResult NotFound(string detail, string errorCode = "ResourceNotFound") =>
        Results.Problem(
            detail: detail,
            statusCode: StatusCodes.Status404NotFound,
            title: "Resource Not Found",
            extensions: new Dictionary<string, object?>
            {
                ["errorCode"] = errorCode,
                ["traceId"]   = Activity.Current?.Id
            });

    public static IResult BadRequest(string detail, string errorCode = "ValidationFailed") =>
        Results.Problem(
            detail: detail,
            statusCode: StatusCodes.Status400BadRequest,
            title: "Bad Request",
            extensions: new Dictionary<string, object?>
            {
                ["errorCode"] = errorCode,
                ["traceId"]   = Activity.Current?.Id
            });
}
```

```csharp
// Middleware/ExceptionMiddleware.cs — sets the x-error-code header
public async Task InvokeAsync(HttpContext context)
{
    try { await _next(context); }
    catch (Exception ex)
    {
        var errorCode = ex switch
        {
            NotFoundException      => "ResourceNotFound",
            ValidationException    => "ValidationFailed",
            ConflictException      => "Conflict",
            _                      => "InternalError"
        };
        context.Response.Headers["x-error-code"] = errorCode;
        context.Response.ContentType = "application/problem+json";
        context.Response.StatusCode  = ex is NotFoundException ? 404 : 500;
        var problem = new ProblemDetails
        {
            Title   = "An error occurred",
            Detail  = ex.Message,
            Status  = context.Response.StatusCode,
            Extensions = { ["errorCode"] = errorCode, ["traceId"] = Activity.Current?.Id }
        };
        await context.Response.WriteAsJsonAsync(problem);
    }
}
```

### API Versioning Requirement

All endpoints require `api-version=YYYY-MM-DD` query parameter. Return `400 Bad Request` with `x-error-code: MissingApiVersionParameter` or `UnsupportedApiVersionValue` for missing/invalid versions.

---

## 2. Health Checks (REQUIRED for AKS)

```csharp
// Program.cs
builder.Services.AddHealthChecks()
    .AddDbContextCheck<AppDbContext>("database", tags: ["ready"]);

// After app.UseRouting():
app.MapHealthChecks("/health/ready", new HealthCheckOptions
{
    Predicate = c => c.Tags.Contains("ready"),
    ResponseWriter = UIResponseWriter.WriteHealthCheckUIResponse
});
app.MapHealthChecks("/health/live", new HealthCheckOptions
{
    Predicate = _ => false // No health checks — returns 200 if process is alive
});
```

**Rules:**

- `/health/ready` — Database connection required; `200` means pod can serve traffic.
- `/health/live` — Process-level only; never check external dependencies in liveness.
- Both endpoints **MUST bypass authorization middleware**.

```csharp
// Bypass auth for health endpoints
app.UseWhen(ctx => !ctx.Request.Path.StartsWithSegments("/health"),
    a => a.UseAuthentication().UseAuthorization());
```

---

## 3. EF Core 10 + PostgreSQL (Npgsql)

### Password Encoding (CRITICAL)

Npgsql treats `/`, `+`, `@`, `;` in passwords as special characters. Always URL-encode.

```csharp
// Infrastructure/Extensions/ConnectionStringExtensions.cs
public static string BuildSafeConnectionString(IConfiguration config)
{
    var password = config["Database:Password"] ?? "";
    var encoded  = System.Net.WebUtility.UrlEncode(password);  // Encode BEFORE building
    return new NpgsqlConnectionStringBuilder
    {
        Host     = config["Database:Host"],
        Database = config["Database:Name"],
        Username = config["Database:User"],
        Password = encoded,
        SslMode  = SslMode.Require
    }.ConnectionString;
}
```

### Migrations (Pre-Build Required)

EF Core migrations **must exist in source control before Docker build**. The app calls `dbContext.Database.Migrate()` at startup; missing migration files cause silent HTTP 500 failures.

```bash
# Generate locally BEFORE git push / Docker build
dotnet ef migrations add <MigrationName> \
    --project src/EmergencyAlerts.Infrastructure \
    --startup-project src/EmergencyAlerts.Api

dotnet ef migrations list \
    --project src/EmergencyAlerts.Infrastructure \
    --startup-project src/EmergencyAlerts.Api
```

### Fail-Fast Migration Startup

```csharp
// Program.cs — fail loudly; do NOT silently continue
try
{
    using var scope = app.Services.CreateScope();
    var db = scope.ServiceProvider.GetRequiredService<AppDbContext>();
    db.Database.Migrate();
    Log.Information("✓ Database migrations applied");
}
catch (Exception ex)
{
    Log.Fatal(ex, "❌ FATAL: Database migration failed — stopping");
    Environment.Exit(1);
}
```

### Read-Only Query Pattern

```csharp
// Always use AsNoTracking for queries that don't mutate
var alerts = await _db.Alerts
    .AsNoTracking()
    .Where(a => a.Status == AlertStatus.Active)
    .Select(a => new AlertDto { Id = a.Id, Title = a.Title })
    .ToListAsync(cancellationToken);
```

**Rules:**

- Never return `IQueryable` across layer boundaries — always `.ToListAsync()`.
- Use `AsNoTracking()` at the query origin for read-only paths.
- Project to DTOs via `Select()`; never return EF entities from Application layer.

---

## 4. SignalR Hub Patterns

### Hub Implementation

```csharp
// Hubs/AlertsHub.cs
[Authorize]   // or remove for demo mode
public class AlertsHub : Hub
{
    public override async Task OnConnectedAsync()
    {
        await Groups.AddToGroupAsync(Context.ConnectionId, "all-operators");
        await base.OnConnectedAsync();
    }

    public override async Task OnDisconnectedAsync(Exception? exception)
    {
        await Groups.RemoveFromGroupAsync(Context.ConnectionId, "all-operators");
        await base.OnDisconnectedAsync(exception);
    }

    // Client calls this on connect AND onreconnected
    public async Task SubscribeToDashboard()
    {
        await Groups.AddToGroupAsync(Context.ConnectionId, "dashboard");
    }
}
```

### Hub Registration

```csharp
// Program.cs
builder.Services.AddSignalR();

// CORS must include AllowCredentials() for SignalR browser clients
builder.Services.AddCors(options =>
    options.AddDefaultPolicy(p =>
        p.WithOrigins(allowedOrigins)
         .AllowAnyHeader()
         .AllowAnyMethod()
         .AllowCredentials()));   // Required for SignalR

app.MapHub<AlertsHub>("/api/hubs/alerts");
```

### Broadcasting from Services

```csharp
// Inject IHubContext<AlertsHub> — not Hub directly
public class AlertService
{
    private readonly IHubContext<AlertsHub> _hub;

    public AlertService(IHubContext<AlertsHub> hub) => _hub = hub;

    public async Task BroadcastAlertCreated(AlertDto alert, CancellationToken ct)
    {
        await _hub.Clients.Group("dashboard")
            .SendAsync("AlertCreated", alert, ct);
    }
}
```

---

## 5. CORS Configuration

> **Kubernetes CORS layer:** [Kubernetes CORS Configuration](../kubernetes-cors-configuration/SKILL.md) covers ConfigMap injection, multi-cluster CORS, and App Configuration integration. This section covers the ASP.NET Core middleware setup.

```csharp
// Program.cs
var allowedOrigins = (configuration["Cors:AllowedOrigins"] ?? "")
    .Split(',', StringSplitOptions.RemoveEmptyEntries | StringSplitOptions.TrimEntries);

builder.Services.AddCors(options =>
    options.AddDefaultPolicy(policy =>
        policy.WithOrigins(allowedOrigins)
              .AllowAnyHeader()
              .AllowAnyMethod()
              .AllowCredentials()));   // Required for SignalR

// Middleware order: Routing → CORS → Auth → Endpoints
app.UseRouting();
app.UseCors();
app.UseAuthentication();
app.UseAuthorization();
```

**ConfigMap injection (Kubernetes):**

```yaml
# k8s deployment ConfigMap
Cors__AllowedOrigins: "${CORS_ALLOWED_ORIGINS}"
```

**CRITICAL:** Restart API pods after any ConfigMap change — CORS origins are read at startup.

---

## 6. Azure Identity (Managed Identity)

```csharp
// Use DefaultAzureCredential for all Azure SDK clients
// Locally: uses az login / Visual Studio credentials
// In AKS: uses Workload Identity (no secrets needed)

builder.Services.AddSingleton(_ =>
    new SecretClient(
        new Uri($"https://{vaultName}.vault.azure.net/"),
        new DefaultAzureCredential()));

// Azure App Configuration
builder.Configuration.AddAzureAppConfiguration(opts =>
    opts.Connect(new Uri(endpoint), new DefaultAzureCredential())
        .UseFeatureFlags());
```

---

## 7. Resilient HTTP Clients

```csharp
// Use Microsoft.Extensions.Http.Resilience (not Polly directly)
builder.Services.AddHttpClient<IExternalService, ExternalService>(c =>
    c.BaseAddress = new Uri(serviceUrl))
    .AddStandardResilienceHandler(opts =>
    {
        opts.Retry.MaxRetryAttempts = 3;
        opts.CircuitBreaker.SamplingDuration = TimeSpan.FromSeconds(30);
    });
```

**Rules:**

- Every outbound HTTP call must have a timeout; never infinite.
- Only retry idempotent operations (GET, PUT with ETag).
- Never retry non-idempotent POSTs unless the API supports `Repeatability-Request-ID`.

---

## 8. Auth / Demo Mode Pattern

```csharp
// Feature flag: Features:DemonstrationMode
// Program.cs: Auth__AllowAnonymous = true disables all auth (dev/demo only)

var allowAnonymous = configuration.GetValue<bool>("Auth:AllowAnonymous");
if (!allowAnonymous)
{
    builder.Services.AddAuthentication(JwtBearerDefaults.AuthenticationScheme)
        .AddJwtBearer(/* ... */);
}

// In controllers: guard demo/admin endpoints to non-prod only
if (!env.IsProduction() || featureFlags.IsEnabled("DemonstrationMode"))
{
    // Demo logic — no secrets/PII
}
```

**Rules:**

- `[AllowAnonymous]` only on health endpoints or explicitly documented public config.
- `Auth__AllowAnonymous=true` is dev/demo only — never in production.
- Demo endpoints that mutate data are disabled in production via `IWebHostEnvironment.IsProduction()`.

---

## 9. Pagination (List Endpoints)

All list endpoints use `{ "value": [...], "nextLink": "https://..." }`:

```csharp
// Presentation layer — Minimal API example
app.MapGet("/api/v1/alerts", async (
    [FromQuery] string? cursor,
    [FromQuery] int pageSize = 50,
    [FromQuery] string apiVersion = "",
    IAlertService svc,
    HttpContext ctx,
    CancellationToken ct) =>
{
    if (string.IsNullOrEmpty(apiVersion))
        return Results.BadRequest(ProblemDetailsHelper.MissingApiVersion());

    var page = await svc.GetPageAsync(cursor, pageSize, ct);
    var baseUrl = $"{ctx.Request.Scheme}://{ctx.Request.Host}";
    return Results.Ok(new
    {
        value    = page.Items,
        nextLink = page.NextCursor is null
            ? null
            : $"{baseUrl}/api/v1/alerts?cursor={page.NextCursor}&pageSize={pageSize}&api-version={apiVersion}"
    });
});
```

**Rules:**

- `nextLink` is absolute URL, omitted on last page (never `null`).
- Include `api-version` in `nextLink`.
- Never return `totalCount` by default (expensive for large tables).

---

## Checklist: Before Merging Backend Changes

- [ ] All error responses use RFC 9457 Problem Details (`application/problem+json`)
- [ ] `x-error-code` header matches `errorCode` body field
- [ ] Liveness/readiness probes bypass auth middleware
- [ ] DB passwords URL-encoded before Npgsql connection string
- [ ] EF Core migrations committed to source control before Docker build
- [ ] Startup migration uses fail-fast (`Environment.Exit(1)`) — no silent catch
- [ ] `AsNoTracking()` on all read-only queries; no `IQueryable` leaked
- [ ] CORS includes `AllowCredentials()` for SignalR hubs
- [ ] `DefaultAzureCredential` used for all Azure SDK clients (no hardcoded keys)
- [ ] `Auth__AllowAnonymous` is false/absent in production ConfigMap
- [ ] List endpoints use `{ value, nextLink }` pagination shape
- [ ] All outbound HTTP clients have timeouts and resilience pipelines
- [ ] ADAC: health endpoints declare degradation reasons, no silent fallbacks

---

## 10. ADAC Resilience: Auto-Detect → Auto-Declare → Auto-Communicate (REQUIRED)

Backend reliability follows the ADAC triad. See `csharp.instructions.md` for the full pattern and `typescript.instructions.md` for the frontend counterpart.

### Auto-Detect

- Identify the execution context: request pipeline, background service, queue consumer, scheduled job.
- Identify failure boundaries: database, external HTTP calls, cache, message bus.
- Track dependency health per external service (last success, consecutive failures, latency).

### Auto-Declare

- Health endpoints MUST declare **why** they are unhealthy, not just return 503.
- Use `HealthCheckResult.Degraded("reason")` when a non-critical dependency fails but partial service is possible.
- When reliability posture changes, include the ADAC declaration in PR descriptions or ADRs.

### Auto-Communicate

- Callers get consistent error contracts: RFC 9457 Problem Details + `x-error-code` + structured `degradationReason` when serving partial data.
- Operators get actionable telemetry: dependency name, duration, outcome, correlation ID.
- No silent fallbacks — if a dependency is down, the health endpoint and response headers must say so.

### Graceful Degradation Patterns

1. **Skip optional enrichment** — If a non-critical service (e.g., notifications, recommendations) is unavailable, serve the core response and set `x-degraded: {service}` header.
2. **Serve cached data** — If the database is read-only or slow, serve from cache and declare staleness via `x-data-age` header.
3. **Circuit breaker transparency** — When `Microsoft.Extensions.Http.Resilience` opens a circuit, log the dependency, duration, and reason. Never swallow the failure.

### ADAC Checklist (Backend)

- [ ] Health endpoints return structured reasons (not just HTTP status)
- [ ] `Degraded` result used for non-critical dependency failures
- [ ] No silent catch blocks in startup or middleware
- [ ] Response headers declare degraded state when serving partial data
- [ ] Telemetry includes dependency name, duration, outcome, and correlation ID

---

## Troubleshooting

### **API returns 500 on all endpoints after deploy**

Most likely cause: EF Core migration failed silently.

```bash
kubectl logs deployment/emergency-alerts-api -n emergency-alerts | grep -E "FATAL|migration|Database"
```

If migration failed: check DB password encoding and firewall rules.

### **SignalR connection succeeds but no events arrive**

Client reconnected to a new pod that has no memory of group membership.
**Fix:** Frontend must call `SubscribeToDashboard()` in `onreconnected` handler.

### **CORS error: No Access-Control-Allow-Origin**

Check in this order:

1. `Cors__AllowedOrigins` ConfigMap value matches exact frontend origin
2. API pods restarted after ConfigMap update: `kubectl rollout restart deployment/emergency-alerts-api`
3. CORS middleware is before `UseAuthentication()` in pipeline
4. `AllowCredentials()` is present in the CORS policy

---

## References

- **csharp.instructions.md**: `.github/instructions/csharp.instructions.md`
- **EF Core 10 docs**: https://learn.microsoft.com/ef/core/what-is-new/ef-core-10.0/whatsnew
- **SignalR hub patterns**: https://learn.microsoft.com/aspnet/core/signalr/hubs
- **RFC 9457 Problem Details**: https://www.rfc-editor.org/rfc/rfc9457
- **Microsoft.Extensions.Http.Resilience**: https://learn.microsoft.com/dotnet/core/resilience/http-resilience

---

## Currency and verification

- **Date checked:** 2026-03-31
- **Sources:** [.NET What's New](https://learn.microsoft.com/dotnet/core/whats-new/dotnet-10/overview), [EF Core 10 docs](https://learn.microsoft.com/ef/core/what-is-new/ef-core-10.0/whatsnew), [ASP.NET Core release notes](https://github.com/dotnet/aspnetcore/releases)
- **Versions verified:** .NET 10 GA (LTS, supported until November 2028), EF Core 10.0.x (GA, LTS), ASP.NET Core 10 GA, Npgsql 10.x
- **Verification steps:** Run `dotnet --version` and check `global.json` against [.NET release index](https://github.com/dotnet/core/releases).

### Known pitfalls

| Area                                   | Pitfall                                                                                                                                             | Mitigation                                                                                                                                                                    |
| -------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| .NET 10 vs .NET 9                      | .NET 10 is GA (LTS, November 2025). EF Core 10 is the matching ORM — do not mix EF Core 9 with .NET 10 targets unless pinning for a specific reason | Upgrade EF Core packages to 10.x when targeting `net10.0`; check [EF Core 10 breaking changes](https://learn.microsoft.com/ef/core/what-is-new/ef-core-10.0/breaking-changes) |
| EF Core migrations                     | Missing migration files at Docker build time cause silent HTTP 500 on startup                                                                       | Commit all migrations to source control; verify `dotnet ef migrations list` passes locally before push                                                                        |
| SignalR multi-replica                  | Clients reconnect to a new pod with no group membership memory                                                                                      | Frontend must re-call `SubscribeToDashboard()` in the `onreconnected` handler                                                                                                 |
| Npgsql SSL                             | Default SSL mode changed between Npgsql 7 and 8; wrong `sslmode` causes connection refused                                                          | Set `sslmode=require` (or `prefer`) explicitly in the connection string; test against the actual DB                                                                           |
| `Microsoft.Extensions.Http.Resilience` | Retry policies silently retry non-idempotent POST requests                                                                                          | Use `AddStandardHedgingHandler()` only for idempotent methods or configure per-method policies                                                                                |

---

## Related Skills

- [PostgreSQL Npgsql](../postgresql-npgsql/SKILL.md) — Database patterns for .NET backends
- [API Security Review](../api-security-review/SKILL.md) — Authentication and authorization patterns
- [Kubernetes CORS Configuration](../kubernetes-cors-configuration/SKILL.md) — CORS setup for deployed APIs
- [TypeScript React Patterns](../typescript-react-patterns/SKILL.md) — Frontend consuming this API
