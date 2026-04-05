# Create a New Aspire Solution

Start a greenfield .NET Aspire solution from scratch.

---

## Prerequisites

- .NET 9+ SDK installed
- Docker Desktop running
- Aspire workload installed: `dotnet workload install aspire`

---

## Steps

### 1. Create the Solution

```bash
dotnet new aspire -n MyApp
cd MyApp
```

This creates:

```
MyApp/
├── MyApp.sln
├── MyApp.AppHost/           # Orchestrator (dev only)
│   └── Program.cs
├── MyApp.ServiceDefaults/   # Shared OTEL/health/resilience
│   └── Extensions.cs
├── MyApp.ApiService/        # Sample API project
│   └── Program.cs
└── MyApp.Web/               # Sample frontend project
    └── Program.cs
```

### 2. Model Your Resources

Edit `MyApp.AppHost/Program.cs` to define your distributed application:

```csharp
var builder = DistributedApplication.CreateBuilder(args);

// Infrastructure
var postgres = builder.AddPostgres("postgres")
    .WithHealthCheck()
    .WithPgAdmin();  // Dev tooling
var database = postgres.AddDatabase("appdb");

var redis = builder.AddRedis("cache")
    .WithHealthCheck()
    .WithRedisCommander();  // Dev tooling

// Services
var api = builder.AddProject<Projects.MyApp_ApiService>("api")
    .WithReference(database)
    .WithReference(redis)
    .WaitFor(database)
    .WaitFor(redis);

var web = builder.AddProject<Projects.MyApp_Web>("web")
    .WithReference(api)
    .WaitFor(api)
    .WithExternalHttpEndpoints();

builder.Build().Run();
```

### 3. Add Client Integrations to Services

In `MyApp.ApiService/Program.cs`:

```csharp
var builder = WebApplication.CreateBuilder(args);

builder.AddServiceDefaults();
builder.AddNpgsqlDbContext<AppDbContext>("appdb");
builder.AddRedisDistributedCache("cache");

var app = builder.Build();
app.MapDefaultEndpoints();
// ... your API endpoints ...
app.Run();
```

### 4. Run Locally

```bash
dotnet run --project MyApp.AppHost
```

The Aspire Dashboard opens showing all services, containers, logs, and traces.

### 5. Add More Projects

```bash
dotnet new webapi -n MyApp.WorkerService
dotnet sln add MyApp.WorkerService
```

Reference Service Defaults and add to AppHost:

```csharp
var worker = builder.AddProject<Projects.MyApp_WorkerService>("worker")
    .WithReference(database)
    .WaitFor(database);
```

---

## Project Structure Conventions

| Project                  | Purpose                       | Ships to Prod? |
| ------------------------ | ----------------------------- | -------------- |
| `*.AppHost`              | Resource graph orchestration  | No             |
| `*.ServiceDefaults`      | Shared cross-cutting concerns | Yes            |
| `*.ApiService` / `*.Web` | Your application logic        | Yes            |
| `*.Tests`                | Integration + unit tests      | No             |

---

## Adding Azure Services

Use emulators for local development:

```csharp
var storage = builder.AddAzureStorage("storage")
    .RunAsEmulator();

var blobs = storage.AddBlobs("blobs");
var queues = storage.AddQueues("queues");

var serviceBus = builder.AddAzureServiceBus("messaging");
```

In production, these reference real Azure resources provisioned via `azd` or your Bicep.

---

## Next Steps

- Configure a compute environment for deployment — see [Deploy with Aspire](deploy.md)
- Add integration tests — see [Integration Testing](../SKILL.md#integration-testing)
- Review production considerations — see [Production Considerations](../standards/production-considerations.md)
