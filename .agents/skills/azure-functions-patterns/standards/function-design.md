# Function Design Standard

## Function Structure

### Single Responsibility

Each function handles exactly one trigger. Do not combine multiple triggers in one function.

```csharp
// GOOD — one function, one trigger
[Function("ProcessOrder")]
public async Task ProcessOrder(
    [ServiceBusTrigger("orders", Connection = "ServiceBusConnection")]
    ServiceBusReceivedMessage message) { /* ... */ }

[Function("OrderCreatedNotify")]
public async Task NotifyOrderCreated(
    [ServiceBusTrigger("order-events", "notifications", Connection = "ServiceBusConnection")]
    ServiceBusReceivedMessage message) { /* ... */ }

// BAD — generic handler trying to do multiple things
[Function("HandleAllMessages")]
public async Task HandleAll(/* ... */) { /* switch on message type */ }
```

### Function Sizing

| Guideline         | Target                                        |
| ----------------- | --------------------------------------------- |
| Execution time    | < 5 minutes (Consumption), Premium default 30 min (configurable) |
| Memory usage      | < 256 MB per instance                         |
| Cold start impact | < 3 seconds                                   |
| Dependencies      | Minimal — inject through DI                   |

---

## HTTP Functions

### API Design

```csharp
[Function("GetOrders")]
public async Task<HttpResponseData> GetOrders(
    [HttpTrigger(AuthorizationLevel.Anonymous, "get", Route = "orders")]
    HttpRequestData req)
{
    var orders = await _orderService.GetAllAsync();

    var response = req.CreateResponse(HttpStatusCode.OK);
    await response.WriteAsJsonAsync(orders);
    return response;
}

[Function("CreateOrder")]
public async Task<HttpResponseData> CreateOrder(
    [HttpTrigger(AuthorizationLevel.Anonymous, "post", Route = "orders")]
    HttpRequestData req)
{
    var order = await req.ReadFromJsonAsync<CreateOrderRequest>();
    var created = await _orderService.CreateAsync(order!);

    var response = req.CreateResponse(HttpStatusCode.Created);
    await response.WriteAsJsonAsync(created);
    return response;
}
```

### Authorization

- Use `AuthorizationLevel.Anonymous` when fronted by API Management or App Gateway
- Use `AuthorizationLevel.Function` for service-to-service calls without APIM
- Never use `AuthorizationLevel.Admin` in production code

---

## Timer Functions

```csharp
[Function("CleanupExpiredSessions")]
public async Task Cleanup(
    [TimerTrigger("0 0 2 * * *")] TimerInfo timer)  // daily at 2 AM UTC
{
    _logger.LogInformation("Cleanup started at {Time}", DateTime.UtcNow);

    if (timer.IsPastDue)
    {
        _logger.LogWarning("Timer is past due — may indicate missed execution");
    }

    await _sessionService.CleanupExpiredAsync();
}
```

**Cron expression format:** `{second} {minute} {hour} {day} {month} {day-of-week}`

---

## Error Handling

### Retry Policies (Isolated Worker)

```csharp
[Function("ProcessOrder")]
[FixedDelayRetry(3, "00:00:30")]  // 3 retries, 30 second delay
public async Task ProcessOrder(
    [ServiceBusTrigger("orders", Connection = "ServiceBusConnection")]
    ServiceBusReceivedMessage message) { /* ... */ }
```

### Distinguishing Transient vs Permanent Failures

```csharp
try
{
    await _service.ProcessAsync(data);
}
catch (HttpRequestException ex) when (ex.StatusCode == HttpStatusCode.ServiceUnavailable)
{
    // Transient — let the runtime retry
    throw;
}
catch (ValidationException ex)
{
    // Permanent — don't retry, log and complete/dead-letter
    _logger.LogError(ex, "Validation failed for {MessageId}", message.MessageId);
    // For Service Bus: message will be dead-lettered after maxDeliveryCount
}
```

---

## Rules

1. One function = one trigger — never combine triggers.
2. Keep functions small — delegate business logic to injected services.
3. Use `AuthorizationLevel.Anonymous` when behind APIM — avoid double auth.
4. Configure retry policies explicitly — don't rely on default retry behaviour.
5. Log with correlation — use `ILogger<T>` with semantic templates.
6. Test functions locally — use Azure Functions Core Tools and local settings.
