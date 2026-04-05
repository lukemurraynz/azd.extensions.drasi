# Durable Functions Patterns

## When to Use Durable Functions

| Pattern            | Use Case                                      |
| ------------------ | --------------------------------------------- |
| Function Chaining  | Sequential multi-step workflows               |
| Fan-out/Fan-in     | Parallel processing with aggregation          |
| Async HTTP API     | Long-running operations with polling          |
| Monitor            | Periodic polling until a condition is met     |
| Human Interaction  | Approval workflows with timeout               |
| Sub-orchestrations | Composing complex workflows from simpler ones |

---

## Function Chaining

```csharp
[Function("OrderOrchestrator")]
public static async Task<OrderResult> RunOrchestrator(
    [OrchestrationTrigger] TaskOrchestrationContext context)
{
    var order = context.GetInput<Order>();

    var validated = await context.CallActivityAsync<Order>("ValidateOrder", order);
    var processed = await context.CallActivityAsync<Order>("ProcessPayment", validated);
    var shipped = await context.CallActivityAsync<Order>("ArrangeShipping", processed);
    var notified = await context.CallActivityAsync<OrderResult>("NotifyCustomer", shipped);

    return notified;
}

[Function("ValidateOrder")]
public async Task<Order> ValidateOrder(
    [ActivityTrigger] Order order) { /* ... */ }
```

---

## Fan-out/Fan-in

```csharp
[Function("BatchProcessor")]
public static async Task<int[]> RunBatch(
    [OrchestrationTrigger] TaskOrchestrationContext context)
{
    var items = context.GetInput<List<WorkItem>>();

    // Fan out — process all items in parallel
    var tasks = items.Select(item =>
        context.CallActivityAsync<int>("ProcessItem", item));

    // Fan in — wait for all results
    var results = await Task.WhenAll(tasks);

    return results;
}
```

---

## Async HTTP API (Long-Running)

```csharp
[Function("StartLongRunning")]
public static async Task<HttpResponseData> HttpStart(
    [HttpTrigger(AuthorizationLevel.Anonymous, "post")] HttpRequestData req,
    [DurableClient] DurableTaskClient client)
{
    var input = await req.ReadFromJsonAsync<ProcessRequest>();
    var instanceId = await client.ScheduleNewOrchestrationInstanceAsync(
        "LongRunningOrchestrator", input);

    return await client.CreateCheckStatusResponseAsync(req, instanceId);
}
```

The `CreateCheckStatusResponse` returns a 202 with polling URLs:

- `statusQueryGetUri` — check orchestration status
- `sendEventPostUri` — send events to the orchestration
- `terminatePostUri` — terminate the orchestration

---

## Human Interaction (Approval)

```csharp
[Function("ApprovalOrchestrator")]
public static async Task<string> RunApproval(
    [OrchestrationTrigger] TaskOrchestrationContext context)
{
    await context.CallActivityAsync("SendApprovalRequest", context.InstanceId);

    using var cts = new CancellationTokenSource();
    var approvalEvent = context.WaitForExternalEvent<bool>("ApprovalResponse");
    var timeout = context.CreateTimer(TimeSpan.FromHours(24), cts.Token);

    var winner = await Task.WhenAny(approvalEvent, timeout);

    if (winner == approvalEvent)
    {
        cts.Cancel();
        return approvalEvent.Result ? "Approved" : "Rejected";
    }

    return "TimedOut";
}
```

---

## Orchestrator Rules

Orchestrator functions must be **deterministic**:

| Rule                    | Why                                    |
| ----------------------- | -------------------------------------- |
| No I/O operations       | Use activities for any external calls  |
| No `DateTime.UtcNow`    | Use `context.CurrentUtcDateTime`       |
| No `Guid.NewGuid()`     | Use `context.NewGuid()`                |
| No `Task.Delay`         | Use `context.CreateTimer()`            |
| No thread-blocking      | Use `await` for all async operations   |
| No static mutable state | Orchestrator may replay multiple times |

---

## Rules

1. Use activities for all side effects — orchestrators must be deterministic.
2. Use fan-out/fan-in instead of sequential loops — parallelise where possible.
3. Set timeouts on human interaction patterns — never wait indefinitely.
4. Use sub-orchestrations for complex workflows — keep orchestrators readable.
5. Monitor orchestration history size — purge completed instances regularly.
