---
name: dotnet-aspire
description: >-
  Configure .NET Aspire for cloud-native application orchestration, service defaults,
  client integrations, and deployment publishers.
  USE FOR: add Aspire to existing project, create greenfield Aspire solution, configure
  service defaults, wire client integrations, set up AppHost resource graph, deploy via
  azd or aspire publish, integration test distributed apps, or choose compute environment.
compatibility: ".NET 10, Aspire 13.0+ (latest: 13.2), Azure Developer CLI (azd)"
---

# .NET Aspire Skill

Patterns for building cloud-native .NET applications with .NET Aspire — covering
the orchestrator (AppHost), service defaults, client integrations, deployment publishers,
and integration testing.

---

## Architecture: Two Halves

.NET Aspire has two distinct halves that serve different purposes:

| Half                                | Ships to Production? | Purpose                                                   |
| ----------------------------------- | -------------------- | --------------------------------------------------------- |
| **AppHost** (orchestrator)          | No                   | Local dev orchestration, resource graph, dashboard, ports |
| **Service Defaults + Integrations** | Yes                  | OTEL, health checks, resilience, service discovery        |

The AppHost is a `.csproj` that references your services and infrastructure resources. It
runs locally (or in CI) to orchestrate containers, databases, and projects. It **never**
ships to production.

Service Defaults (`AddServiceDefaults()`) and client integration NuGet packages configure
production-grade telemetry, resilience, and connectivity using standard
`Microsoft.Extensions.*` APIs — no vendor lock-in.

---

## Quick Reference

| Capability                | Description                                                            |
| ------------------------- | ---------------------------------------------------------------------- |
| Service Defaults          | Single call to wire OTEL, health checks, resilience, service discovery |
| Client Integrations       | NuGet packages that register typed clients with connection strings     |
| AppHost Resource Graph    | Declarative resource dependencies with `WithReference` and `WaitFor`   |
| Health-Check Dependencies | `WaitFor()` / `WaitForStart()` for startup ordering                    |
| Compute Environments      | Publishers for ACA, Kubernetes, Docker Compose, App Service            |
| OTLP Protocol Config      | gRPC (default) or HTTP Protobuf via `WithOtlpExporter()`               |
| Integration Testing       | `DistributedApplicationTestingBuilder` for full-graph tests            |
| Aspire Dashboard          | Local OTEL dashboard for traces, logs, metrics during development      |
| Emulators                 | `RunAsEmulator()` for Azure services (Azurite, Event Hubs, etc.)       |

---

## Standards

| Standard                                                            | Purpose                                                 |
| ------------------------------------------------------------------- | ------------------------------------------------------- |
| [Service Defaults](standards/service-defaults.md)                   | Anatomy of `AddServiceDefaults()` and what it registers |
| [Client Integrations](standards/client-integrations.md)             | NuGet integration patterns and connection conventions   |
| [Production Considerations](standards/production-considerations.md) | What ships, risks, connection string conventions        |
| [Compute Environments](standards/compute-environments.md)           | Publisher patterns for ACA, K8s, Compose, App Service   |
| [Checklist](standards/checklist.md)                                 | Validation checklist before shipping                    |

---

## Actions

| Action                                                      | When to use                                         |
| ----------------------------------------------------------- | --------------------------------------------------- |
| [Add Aspire to Existing Project](actions/add-to-project.md) | Retrofitting Aspire into an existing .NET solution  |
| [Create Aspire Solution](actions/create-solution.md)        | Greenfield Aspire project from scratch              |
| [Deploy with Aspire](actions/deploy.md)                     | Publishing to ACA, K8s, Docker Compose, App Service |

---

## Service Defaults — What `AddServiceDefaults()` Does

A single extension method in your `ServiceDefaults` project that registers:

1. **OpenTelemetry** — tracing, metrics, and logging exporters (OTLP)
2. **Health checks** — `/health/ready` and `/health/live` endpoints
3. **Resilience** — `ConfigureHttpClientDefaults` with `AddStandardResilienceHandler()`
4. **Service discovery** — `AddServiceDiscovery()` for `http://servicename` URIs

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
    return builder;
}
```

Every service project calls `builder.AddServiceDefaults()` in `Program.cs`. This is the
code that **ships to production** and replaces manual OTEL, health check, resilience, and
service discovery setup.

> **Cross-reference:** For manual OTEL/health check patterns without Aspire, see
> [observability-monitoring](../observability-monitoring/SKILL.md) and
> [dotnet-backend-patterns](../dotnet-backend-patterns/SKILL.md).

---

## AppHost Resource Graph

The AppHost defines your distributed application's topology:

```csharp
var builder = DistributedApplication.CreateBuilder(args);

var postgres = builder.AddPostgres("postgres")
    .WithHealthCheck();
var database = postgres.AddDatabase("mydb");

var redis = builder.AddRedis("cache")
    .WithHealthCheck();

var api = builder.AddProject<Projects.Api>("api")
    .WithReference(database)
    .WithReference(redis)
    .WaitFor(database)
    .WaitFor(redis);

var frontend = builder.AddProject<Projects.Frontend>("frontend")
    .WithReference(api)
    .WaitFor(api)
    .WithExternalHttpEndpoints();

builder.Build().Run();
```

### Startup Ordering (9.5+)

Two levels of dependency waiting:

| Method           | Waits For                     | Use Case                            |
| ---------------- | ----------------------------- | ----------------------------------- |
| `WaitForStart()` | Resource enters Running state | Migrators that need container up    |
| `WaitFor()`      | Resource passes health checks | Services that need healthy upstream |

```csharp
var migrator = builder.AddProject<Projects.Migrator>("migrator")
    .WaitForStart(database)   // Start as soon as container runs
    .WithReference(database);

var api = builder.AddProject<Projects.Api>("api")
    .WaitFor(database)        // Wait for healthy database
    .WaitFor(migrator)        // Wait for migration to complete
    .WithReference(database);
```

### External Services (9.5+)

Reference services outside the Aspire graph:

```csharp
var externalApi = builder.AddExternalService("partner-api", "https://api.partner.com")
    .WithHttpHealthCheck("/health/ready");

var api = builder.AddProject<Projects.Api>("api")
    .WaitFor(externalApi)
    .WithReference(externalApi);
```

---

## Emulators for Local Development

Use `RunAsEmulator()` to run Azure service emulators locally:

```csharp
var storage = builder.AddAzureStorage("storage")
    .RunAsEmulator();   // Azurite in Docker

var eventHubs = builder.AddAzureEventHubs("events")
    .RunAsEmulator();   // Event Hubs emulator

var cosmos = builder.AddAzureCosmosDB("cosmos")
    .RunAsEmulator();   // CosmosDB emulator
```

This keeps AppHost self-contained — no Azure subscription needed for local dev.

---

## OTLP Protocol Configuration (9.5+)

Choose between gRPC (default, higher performance) and HTTP Protobuf:

```csharp
var api = builder.AddProject<Projects.Api>("api")
    .WithOtlpExporter(OtlpProtocol.HttpProtobuf);  // For environments blocking gRPC

var worker = builder.AddProject<Projects.Worker>("worker")
    .WithOtlpExporter();  // Default: gRPC
```

---

## Integration Testing

Use `DistributedApplicationTestingBuilder` to spin up the full resource graph in tests:

```csharp
[Fact]
public async Task ApiReturnsHealthy()
{
    var appHost = await DistributedApplicationTestingBuilder
        .CreateAsync<Projects.AppHost>();

    await using var app = await appHost.BuildAsync();
    await app.StartAsync();

    var httpClient = app.CreateHttpClient("api");
    var response = await httpClient.GetAsync("/health/ready");

    response.EnsureSuccessStatusCode();
}
```

This boots real containers and services — use sparingly for integration tests, not unit tests.

---

## When to Use Aspire vs Not

### Use Aspire When

- Building multi-service .NET applications with databases, caches, or message brokers
- You want turnkey OTEL, resilience, and service discovery without manual wiring
- Local dev experience matters — Aspire Dashboard is excellent for debugging
- Deploying to Azure Container Apps (first-class `azd` support)
- Integration testing distributed systems

### Do NOT Use Aspire When

- Single-project APIs with no external dependencies
- Non-.NET services dominate the architecture (Aspire is .NET-centric)
- You need fine-grained control over every OTEL/resilience configuration
- Production Kubernetes with complex Helm charts (Aspire K8s manifests are a starting point)

---

## Production Risks and Mitigations

| Risk                                              | Mitigation                                                           |
| ------------------------------------------------- | -------------------------------------------------------------------- |
| `AddStandardResilienceHandler()` applies globally | Override per-client with named `HttpClient` configurations           |
| Connection strings via env vars only              | Use `ConnectionStrings__<name>` convention; document in runbooks     |
| AppHost-generated Bicep is scaffold-quality       | Treat published IaC as a starting point; apply your Bicep standards  |
| K8s manifests lack production hardening           | Overlay with Kustomize or Helm; add probes, limits, network policies |
| Aspire Dashboard is dev-only                      | Use Application Insights / Grafana in production                     |

---

## Currency and Verification

- **Date checked:** 2026-03-31
- **Compatibility:** .NET 10, Aspire 13.2+ (rebranded from ".NET Aspire" to "Aspire" as of 13.0; latest: [What's new in Aspire 13.2](https://aspire.dev/whats-new/aspire-13-2/))
- **Sources:** [Aspire documentation](https://learn.microsoft.com/dotnet/aspire/), [Aspire 13.0 upgrade guide](https://learn.microsoft.com/dotnet/aspire/get-started/upgrade-to-aspire-13), [Aspire 13.0 breaking changes](https://learn.microsoft.com/dotnet/aspire/compatibility/13.0/), [Aspire 13.2 release notes](https://aspire.dev/whats-new/aspire-13-2/), [Aspire.Hosting NuGet](https://www.nuget.org/packages/Aspire.Hosting/)
- **Verification steps:**
  1. Check latest Aspire NuGet package version: `dotnet package search Aspire.Hosting --prerelease`
  2. Verify `AddServiceDefaults()` API surface in [ServiceDefaults docs](https://learn.microsoft.com/dotnet/aspire/fundamentals/service-defaults)
  3. Verify compute environment publisher list: `dotnet run -- --publisher` in AppHost project
  4. Check breaking changes in [Aspire breaking changes](https://learn.microsoft.com/dotnet/aspire/compatibility/breaking-changes)

### Aspire 13.0 Breaking Changes (from 9.x)

- **SDK declaration simplified:** Use `Sdk="Aspire.AppHost.Sdk/13.0.0"` — remove explicit `Aspire.Hosting.AppHost` package reference.
- **Target framework:** `.NET 10` (`net10.0`) required for Aspire 13.0.
- **Azure Storage APIs renamed:** `AddAzureBlobClient` → `AddAzureBlobServiceClient`, `AddAzureQueueClient` → `AddAzureQueueServiceClient`, `AddAzureTableClient` → `AddAzureTableServiceClient`.
- **DefaultAzureCredential on ACA/App Service:** Now defaults to `ManagedIdentityCredential` only (sets `AZURE_TOKEN_CREDENTIALS` environment variable).
- **Single-file AppHosts:** Aspire 13.0 supports `#:sdk Aspire.AppHost.Sdk@13.0.0` with `#:package` directives — no `.csproj` required for prototyping.
- **Upgrade path:** If upgrading from 8.x, must go through 9.x first. Review breaking changes for all intermediate versions.
- **Aspire CLI:** New `aspire update` command for upgrading packages.

> **Aspire evolves rapidly.** Before using patterns from this skill in a new project, verify the API surface against the [Aspire 13.0 upgrade guide](https://learn.microsoft.com/dotnet/aspire/get-started/upgrade-to-aspire-13). Key areas of change: publisher APIs, integration package names, Azure Storage API renames, and credential defaults.

### Known Pitfalls

| Area                                      | Pitfall                                                                                                                                                | Mitigation                                                                              |
| ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------- |
| NuGet package renames                     | Aspire integration packages have been renamed between major versions (e.g., Azure Storage: `AddAzureBlobClient` → `AddAzureBlobServiceClient` in 9.4+) | Check NuGet for the current package name before adding references                       |
| `RunAsEmulator()` port conflicts          | Multiple emulators can bind to the same default port on the host                                                                                       | Use `.WithHostPort()` to assign explicit ports if running multiple emulators            |
| `WaitFor()` vs `WaitForStart()`           | Confusing the two causes either premature starts or unnecessary delays                                                                                 | `WaitForStart()` = container running; `WaitFor()` = health check passed                 |
| AppHost-generated manifests in production | Generated K8s/Bicep lacks hardening (probes, limits, network policies)                                                                                 | Always overlay with production-grade manifests; treat generated output as scaffold only |
| Service discovery in production           | `http://servicename` URIs require DNS or service mesh configuration outside Aspire                                                                     | Configure Kubernetes service DNS or explicit connection strings for production          |
