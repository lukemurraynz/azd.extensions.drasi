---
name: azure-functions-patterns
description: >-
  Azure Functions patterns including triggers, bindings, Durable Functions, isolated worker model, deployment strategies, and scaling configuration. USE FOR: building serverless event-driven workloads on Azure Functions.compatibility: Requires Azure Functions Core Tools, Azure CLI
---

# Azure Functions Patterns Skill

> **MUST:** Use the isolated worker model for all new .NET Functions projects.
> The in-process model is deprecated. Use Managed Identity for all Azure service bindings.

---

## Quick Reference

| Capability            | Description                                                  |
| --------------------- | ------------------------------------------------------------ |
| Triggers              | HTTP, Timer, Queue, Service Bus, Event Grid, Blob, Cosmos DB |
| Input/Output Bindings | Declarative data access without boilerplate SDK code         |
| Durable Functions     | Stateful orchestrations, fan-out/fan-in, human interaction   |
| Isolated Worker       | Out-of-process model with full dependency injection support  |
| Scaling               | Consumption, Premium, and Dedicated plan scaling behaviours  |
| Deployment            | Zip deploy, container deploy, azd integration                |

---

## Standards

| Standard                                          | Purpose                           |
| ------------------------------------------------- | --------------------------------- |
| [Function Design](standards/function-design.md)   | Trigger selection and structure   |
| [Durable Patterns](standards/durable-patterns.md) | Orchestration and entity patterns |
| [Checklist](standards/checklist.md)               | Validation checklist              |

---

## Actions

| Action                                                | When to use                   |
| ----------------------------------------------------- | ----------------------------- |
| [Create Function App](actions/create-function-app.md) | Setting up a new Function App |

---

## Trigger Selection

| Event Source          | Trigger Type        | Scaling Behaviour               |
| --------------------- | ------------------- | ------------------------------- |
| HTTP request          | `HttpTrigger`       | Concurrent request count        |
| Timer/schedule        | `TimerTrigger`      | Single instance (singleton)     |
| Storage Queue         | `QueueTrigger`      | Queue length                    |
| Service Bus           | `ServiceBusTrigger` | Message count                   |
| Event Grid            | `EventGridTrigger`  | Event rate                      |
| Blob Storage          | `BlobTrigger`       | New/updated blob count          |
| Cosmos DB change feed | `CosmosDBTrigger`   | Change feed lag                 |
| Event Hubs            | `EventHubTrigger`   | Partition count (max instances) |

---

## Isolated Worker Model (.NET)

> **Warning — logging configuration:** In the isolated worker model, `host.json` log level
> settings **do not apply to the worker process**. They only control the Functions host.
> To configure log levels in your application code, use `ConfigureLogging()` in `Program.cs`:
>
> ```csharp
> .ConfigureLogging(logging =>
> {
>     logging.SetMinimumLevel(LogLevel.Information);
>     logging.AddFilter("Microsoft", LogLevel.Warning);
> })
> ```

```csharp
// Program.cs
var host = new HostBuilder()
    .ConfigureFunctionsWebApplication()
    .ConfigureServices(services =>
    {
        services.AddApplicationInsightsTelemetryWorkerService();
        services.ConfigureFunctionsApplicationInsights();
        services.AddSingleton<IOrderService, OrderService>();
    })
    .Build();

host.Run();
```

```csharp
// Function with dependency injection
public class OrderFunctions
{
    private readonly IOrderService _orderService;
    private readonly ILogger<OrderFunctions> _logger;

    public OrderFunctions(IOrderService orderService, ILogger<OrderFunctions> logger)
    {
        _orderService = orderService;
        _logger = logger;
    }

    [Function("ProcessOrder")]
    public async Task Run(
        [ServiceBusTrigger("orders", Connection = "ServiceBusConnection")] ServiceBusReceivedMessage message,
        FunctionContext context)
    {
        var order = message.Body.ToObjectFromJson<OrderEvent>();
        _logger.LogInformation("Processing order {OrderId}", order.Id);
        await _orderService.ProcessAsync(order);
    }
}
```

> [!NOTE]
> Prefer `ExponentialBackoffRetry` over `FixedDelayRetry` for external API calls to avoid thundering herd:
> ```csharp
> [ExponentialBackoffRetry(5, "00:00:04", "00:15:00")]
> public async Task Run([ServiceBusTrigger("orders")] ServiceBusReceivedMessage message) { }
> ```
> `FixedDelayRetry` is appropriate only for operations where consistent timing matters (e.g., database reconnection with connection pooling).

---

## Bindings with Managed Identity

Use identity-based connections instead of connection strings:

```json
// local.settings.json / App Settings
{
  "ServiceBusConnection__fullyQualifiedNamespace": "mybus.servicebus.windows.net",
  "StorageConnection__blobServiceUri": "https://mystorage.blob.core.windows.net",
  "CosmosConnection__accountEndpoint": "https://mycosmos.documents.azure.com:443/"
}
```

The `__fullyQualifiedNamespace` suffix tells the Functions runtime to use
`DefaultAzureCredential` instead of a connection string.

---

## Hosting Plans

| Plan             | Scale-to-zero | Max Instances | Cold Start | Best For                     |
| ---------------- | ------------- | ------------- | ---------- | ---------------------------- |
| Consumption      | Yes           | 200           | Yes        | Low-traffic, event-driven    |
| Flex Consumption | Yes           | 1000          | Reduced    | Variable traffic, fast scale |
| Premium (EP)     | Optional      | 100           | No         | Predictable traffic, VNet    |
| Dedicated        | No            | Plan-based    | No         | Always-on, existing App Svc  |
| Container Apps   | Yes           | 300           | Yes        | Custom containers, Dapr      |

### Hosting Plan Decision Tree

1. **.NET 10 on Linux?** → MUST use Flex Consumption (Consumption plan not supported)
2. **VNet integration required?** → Premium (EP) or Flex Consumption
3. **Low-traffic, event-driven?** → Consumption (cost-optimised scale-to-zero)
4. **Variable traffic, fast scale?** → Flex Consumption
5. **Predictable traffic, always-on?** → Premium or Dedicated

> **Warning — .NET 10 on Linux Consumption:** .NET 10 is NOT supported on the Linux
> Consumption plan. Use the Flex Consumption plan instead for .NET 10 on Linux.
> See [Migrate Consumption plan apps to Flex Consumption](https://learn.microsoft.com/azure/azure-functions/migration/migrate-plan-consumption-to-flex).

---

## Bicep — Function App

```bicep
resource functionApp 'Microsoft.Web/sites@2025-03-01' = {
  name: 'func-${environmentName}'
  location: location
  kind: 'functionapp,linux'
  identity: { type: 'SystemAssigned' }
  properties: {
    serverFarmId: appServicePlan.id
    siteConfig: {
      linuxFxVersion: 'DOTNET-ISOLATED|10.0'
      appSettings: [
        { name: 'AzureWebJobsStorage__blobServiceUri', value: 'https://${storageAccount.name}.blob.core.windows.net' }
        { name: 'AzureWebJobsStorage__queueServiceUri', value: 'https://${storageAccount.name}.queue.core.windows.net' }
        { name: 'AzureWebJobsStorage__tableServiceUri', value: 'https://${storageAccount.name}.table.core.windows.net' }
        { name: 'FUNCTIONS_EXTENSION_VERSION', value: '~4' }
        { name: 'FUNCTIONS_WORKER_RUNTIME', value: 'dotnet-isolated' }
        { name: 'APPLICATIONINSIGHTS_CONNECTION_STRING', value: appInsights.properties.ConnectionString }
      ]
    }
  }
}
```

---

## azure.yaml Integration (azd)

```yaml
services:
  functions:
    project: ./src/functions
    host: function
    language: csharp
```

---

## Principles

1. **One function, one responsibility** — each function handles exactly one trigger.
2. **Isolated worker model** — use out-of-process for .NET; leverage full DI capabilities.
3. **Identity-based connections** — use `__fullyQualifiedNamespace` suffix, not connection strings.
4. **Idempotent handlers** — queue and event triggers may deliver messages more than once.
5. **Right-size the plan** — Consumption for spiky, Premium for VNet, Flex for fast scale.
6. **Structured logging** — use `ILogger<T>` with semantic templates throughout.
7. **Durable Functions TaskHub isolation** — when using deployment slots, each slot must use a
   unique `TaskHub` name in `host.json`. Shared TaskHub names cause orchestration state
   cross-contamination between slots (a common source of mysterious failures).

---

## Currency and Verification

- **Date checked:** 2026-03-31 (verified via Microsoft Learn MCP — ARM template references)
- **Compatibility:** Azure Bicep, ARM templates, Azure Functions
- **Sources:**
  - [Microsoft.Web/serverfarms](https://learn.microsoft.com/azure/templates/microsoft.web/serverfarms)
  - [Microsoft.Storage/storageAccounts](https://learn.microsoft.com/azure/templates/microsoft.storage/storageaccounts)
- **Verification steps:**
  1. Run `az provider show --namespace Microsoft.Web --query "resourceTypes[?resourceType=='serverfarms'].apiVersions" -o tsv` and confirm `2025-03-01` is listed
  2. Run `az provider show --namespace Microsoft.Storage --query "resourceTypes[?resourceType=='storageAccounts'].apiVersions" -o tsv` and confirm `2025-08-01` is listed
  3. Run `az bicep build --file <your-bicep-file>` to validate syntax

### Known Pitfalls

| Area                        | Pitfall                                                                                   | Mitigation                                                    |
| --------------------------- | ----------------------------------------------------------------------------------------- | ------------------------------------------------------------- | -------------------------- |
| .NET 10 + Linux Consumption | .NET 10 is not supported on Linux Consumption plan                                        | Use Flex Consumption plan for .NET 10 on Linux                |
| In-process model            | In-process model is deprecated; end-of-support November 10, 2026                          | Migrate to isolated worker model                              |
| linuxFxVersion              | Using wrong version string (e.g., `8.0` when targeting `net10.0`) causes runtime mismatch | Set `DOTNET-ISOLATED\|10.0` for .NET 10 projects |
| Durable TaskHub isolation   | Shared TaskHub names across deployment slots cause state cross-contamination              | Use unique `TaskHub` name per slot in `host.json`             |
| Logging in isolated worker  | `host.json` log level settings do not apply to the worker process                         | Configure log levels in `Program.cs` via `ConfigureLogging()` |

---

## References

- [Azure Functions overview](https://learn.microsoft.com/azure/azure-functions/functions-overview)
- [Isolated worker model](https://learn.microsoft.com/azure/azure-functions/dotnet-isolated-process-guide)
- [Identity-based connections](https://learn.microsoft.com/azure/azure-functions/functions-reference#configure-an-identity-based-connection)
- [Durable Functions](https://learn.microsoft.com/azure/azure-functions/durable/durable-functions-overview)
- [Hosting plans comparison](https://learn.microsoft.com/azure/azure-functions/functions-scale)

---

## Related Skills

- **event-driven-messaging** — Service Bus and Event Grid integration
- **observability-monitoring** — Application Insights for Functions
- **identity-managed-identity** — Managed Identity bindings
- **managing-azure-dev-cli-lifecycle** — azd deployment with `function` host
