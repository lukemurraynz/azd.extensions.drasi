# Add Aspire to an Existing Project

Retrofit .NET Aspire into an existing .NET solution without rewriting your applications.

---

## Prerequisites

- .NET 9+ SDK installed
- Docker Desktop running (for local containers)
- Existing .NET solution with one or more service projects

---

## Steps

### 1. Install Aspire Workload

```bash
dotnet workload update
dotnet workload install aspire
```

### 2. Create the AppHost Project

```bash
dotnet new aspire-apphost -n MyApp.AppHost
dotnet sln add MyApp.AppHost
```

Add project references to your existing services:

```xml
<!-- MyApp.AppHost.csproj -->
<ItemGroup>
  <ProjectReference Include="..\MyApp.Api\MyApp.Api.csproj" />
  <ProjectReference Include="..\MyApp.Worker\MyApp.Worker.csproj" />
</ItemGroup>
```

Configure the resource graph in `Program.cs`:

```csharp
var builder = DistributedApplication.CreateBuilder(args);

var api = builder.AddProject<Projects.MyApp_Api>("api");
var worker = builder.AddProject<Projects.MyApp_Worker>("worker")
    .WithReference(api);

builder.Build().Run();
```

### 3. Create the Service Defaults Project

```bash
dotnet new aspire-servicedefaults -n MyApp.ServiceDefaults
dotnet sln add MyApp.ServiceDefaults
```

Reference it from each service project:

```xml
<!-- MyApp.Api.csproj -->
<ItemGroup>
  <ProjectReference Include="..\MyApp.ServiceDefaults\MyApp.ServiceDefaults.csproj" />
</ItemGroup>
```

### 4. Wire Service Defaults into Each Service

In each service's `Program.cs`, add the call early:

```csharp
var builder = WebApplication.CreateBuilder(args);

builder.AddServiceDefaults();  // Add this line

// ... existing service registration ...

var app = builder.Build();

app.MapDefaultEndpoints();     // Add this line (health checks)

// ... existing middleware and endpoints ...

app.Run();
```

### 5. Add Client Integrations (Optional)

Replace manual client registration with Aspire integrations:

**Before (manual):**

```csharp
builder.Services.AddDbContext<AppDbContext>(options =>
    options.UseNpgsql(builder.Configuration.GetConnectionString("postgres")));
```

**After (Aspire integration):**

```csharp
builder.AddNpgsqlDbContext<AppDbContext>("postgres");
```

And in the AppHost:

```csharp
var postgres = builder.AddPostgres("postgres")
    .WithHealthCheck();
var database = postgres.AddDatabase("mydb");

var api = builder.AddProject<Projects.MyApp_Api>("api")
    .WithReference(database)
    .WaitFor(database);
```

### 6. Add Infrastructure Resources to AppHost

Model your existing infrastructure dependencies:

```csharp
var redis = builder.AddRedis("cache").WithHealthCheck();
var serviceBus = builder.AddAzureServiceBus("messaging");

var api = builder.AddProject<Projects.MyApp_Api>("api")
    .WithReference(redis)
    .WithReference(serviceBus)
    .WaitFor(redis);
```

### 7. Run and Verify

```bash
dotnet run --project MyApp.AppHost
```

The Aspire Dashboard opens automatically at `https://localhost:15xxx` showing all
services, their logs, traces, and metrics.

---

## Incremental Adoption

You do not need to convert everything at once:

1. **Start with Service Defaults only** — get OTEL, health checks, and resilience for free
2. **Add AppHost later** — model your resource graph when you need orchestrated local dev
3. **Add client integrations gradually** — replace manual DI registrations one at a time
4. **Choose a compute environment last** — when you're ready to publish to a deployment target

Each step is independently valuable and does not require the others.
