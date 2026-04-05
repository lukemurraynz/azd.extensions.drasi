# Client Integrations Standard

Client integrations are NuGet packages that register typed clients, connection factories,
and health checks for specific technologies. They ship to production alongside your
application code.

---

## Connection String Convention

Aspire uses a single convention: `ConnectionStrings__<name>` environment variable or
`ConnectionStrings:<name>` in configuration. The `<name>` matches the resource name
defined in the AppHost.

```csharp
// AppHost
var postgres = builder.AddPostgres("orders-db");

// Service — uses "orders-db" as connection name
builder.AddNpgsqlDataSource("orders-db");
```

At runtime, the connection string resolves from:

1. `ConnectionStrings__orders-db` environment variable (injected by AppHost in dev)
2. `ConnectionStrings:orders-db` in `appsettings.json` or other config providers

---

## Common Integration Packages

### Database

```csharp
// PostgreSQL via Npgsql
builder.AddNpgsqlDataSource("postgres");

// PostgreSQL via EF Core
builder.AddNpgsqlDbContext<AppDbContext>("postgres");

// SQL Server via EF Core
builder.AddSqlServerDbContext<AppDbContext>("sqldb");

// Azure Cosmos DB
builder.AddAzureCosmosClient("cosmos");
```

### Caching

```csharp
// Redis (distributed cache)
builder.AddRedisDistributedCache("cache");

// Redis (output cache)
builder.AddRedisOutputCache("cache");

// Redis (general client)
builder.AddRedisClient("cache");
```

### Messaging

```csharp
// Azure Service Bus
builder.AddAzureServiceBusClient("messaging");

// Azure Event Hubs
builder.AddAzureEventHubProducerClient("events", settings =>
{
    settings.EventHubName = "orders";
});

// RabbitMQ
builder.AddRabbitMQClient("rabbitmq");
```

### Storage and Identity

```csharp
// Azure Blob Storage
builder.AddAzureBlobClient("blobs");

// Azure Key Vault
builder.AddAzureKeyVaultClient("secrets");

// Azure App Configuration
builder.AddAzureAppConfiguration("config");
```

---

## What Each Integration Registers

Every client integration package typically registers:

1. **Typed client or connection factory** — via DI (e.g., `NpgsqlDataSource`, `ServiceBusClient`)
2. **Health check** — automatically added to the health check pipeline
3. **OpenTelemetry instrumentation** — traces and metrics for the client library
4. **Resilience** — connection retry and circuit-breaker policies where appropriate

You get all four by calling one `Add*` method.

---

## Configuration Override Pattern

Client integrations accept a settings callback for customization:

```csharp
builder.AddNpgsqlDataSource("postgres", settings =>
{
    settings.DisableHealthChecks = false;
    settings.DisableTracing = false;
    settings.DisableMetrics = false;
});

builder.AddRedisClient("cache", settings =>
{
    settings.ConnectionString = "override-if-needed";
});
```

---

## AppHost Resource ↔ Integration Mapping

| AppHost Resource             | Service Integration                    |
| ---------------------------- | -------------------------------------- |
| `AddPostgres("pg")`          | `AddNpgsqlDataSource("pg")`            |
| `AddRedis("cache")`          | `AddRedisDistributedCache("cache")`    |
| `AddAzureServiceBus("sb")`   | `AddAzureServiceBusClient("sb")`       |
| `AddAzureEventHubs("eh")`    | `AddAzureEventHubProducerClient("eh")` |
| `AddAzureCosmosDB("cosmos")` | `AddAzureCosmosClient("cosmos")`       |
| `AddAzureStorage("storage")` | `AddAzureBlobClient("storage")`        |
| `AddRabbitMQ("rmq")`         | `AddRabbitMQClient("rmq")`             |

The name string must match between AppHost and service integration.

---

## Without Aspire AppHost

Client integration packages work standalone — you just provide the connection string
via configuration instead of relying on AppHost injection:

```json
{
  "ConnectionStrings": {
    "postgres": "Host=myserver;Database=mydb;Username=user;Password=secret"
  }
}
```

This makes Aspire integrations safe to adopt incrementally.
