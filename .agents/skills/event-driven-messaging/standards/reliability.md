# Reliability Patterns Standard

## Idempotency

Every message consumer must be idempotent — safely handle the same message multiple times.

### Implementation Patterns

**Deduplication Store (Recommended)**

```csharp
public async Task<bool> ProcessIfNewAsync(string messageId, Func<Task> handler)
{
    if (await _store.ExistsAsync(messageId))
        return false;  // already processed

    await handler();
    await _store.AddAsync(messageId, TimeSpan.FromDays(7));
    return true;
}
```

**Natural Idempotency**

Design operations to be naturally idempotent:

- `SET status = 'Completed'` instead of `INCREMENT counter`
- `UPSERT` instead of `INSERT`
- Use `If-Match` ETags for conditional updates

**Service Bus Deduplication**

Enable on the queue for automatic message-ID-based deduplication:

```bicep
properties: {
  requiresDuplicateDetection: true
  duplicateDetectionHistoryTimeWindow: 'PT10M'
}
```

---

## Dead-Letter Handling

### Configuration

Every queue and subscription must have dead-lettering enabled:

```bicep
properties: {
  maxDeliveryCount: 5
  deadLetteringOnMessageExpiration: true
  deadLetteringOnFilterEvaluationExceptions: true  // topics only
}
```

### Dead-Letter Processing

Monitor and reprocess dead-lettered messages:

```csharp
// Read from dead-letter queue
var dlqClient = serviceBusClient.CreateReceiver(queueName,
    new ServiceBusReceiverOptions { SubQueue = SubQueue.DeadLetterQueue });

var messages = await dlqClient.ReceiveMessagesAsync(maxMessages: 10);
foreach (var msg in messages)
{
    _logger.LogWarning("Dead-lettered: {Reason} — {Description}",
        msg.DeadLetterReason, msg.DeadLetterErrorDescription);

    // Investigate, fix, and resubmit if appropriate
}
```

### Dead-Letter Alerting

Create an alert for dead-letter queue depth:

```kql
AzureMetrics
| where ResourceProvider == "MICROSOFT.SERVICEBUS"
| where MetricName == "DeadletteredMessages"
| where Total > 0
| summarize MaxDLQ = max(Total) by Resource, bin(TimeGenerated, 5m)
```

---

## Retry Policies

### Client-Side Retries

```csharp
var clientOptions = new ServiceBusClientOptions
{
    RetryOptions = new ServiceBusRetryOptions
    {
        Mode = ServiceBusRetryMode.Exponential,
        MaxRetries = 3,
        Delay = TimeSpan.FromSeconds(1),
        MaxDelay = TimeSpan.FromSeconds(30)
    }
};
```

### Server-Side Retries (Event Grid)

```bicep
retryPolicy: {
  maxDeliveryAttempts: 5
  eventTimeToLiveInMinutes: 1440  // 24 hours
}
```

---

## Poison Message Handling

Messages that consistently fail processing:

1. **Detect:** Check `DeliveryCount` against `maxDeliveryCount`
2. **Isolate:** Let Service Bus auto-move to dead-letter queue
3. **Alert:** Monitor dead-letter queue depth
4. **Investigate:** Read dead-letter reason and error description
5. **Resolve:** Fix the root cause, then resubmit messages

```csharp
if (args.Message.DeliveryCount >= maxDeliveryCount - 1)
{
    _logger.LogError("Message {Id} is a poison message after {Count} attempts",
        args.Message.MessageId, args.Message.DeliveryCount);
    await args.DeadLetterMessageAsync(args.Message,
        deadLetterReason: "PoisonMessage",
        deadLetterErrorDescription: "Exceeded retry threshold");
    return;
}
```

---

## Ordering Guarantees

### Service Bus Sessions (FIFO)

```bicep
properties: {
  requiresSession: true
}
```

```csharp
// Send with session ID for ordering
var message = new ServiceBusMessage(payload)
{
    SessionId = orderId  // all messages for this order processed in order
};
```

### Event Hubs Partitions

Messages with the same partition key go to the same partition and are processed in order:

```csharp
var eventData = new EventData(payload);
await producerClient.SendAsync(new[] { eventData },
    new SendEventOptions { PartitionKey = customerId });
```

> **Event Hubs checkpointing frequency:** Checkpointing after every event causes Azure
> Storage throttling (each checkpoint is a blob write). Checkpoint too infrequently and
> you get excessive duplicate processing on restart. A good default: checkpoint every
> **N events** (e.g., 100) or every **T seconds** (e.g., 30), whichever comes first.
> The `EventProcessorClient` does not auto-checkpoint — you must call
> `UpdateCheckpointAsync` explicitly.

---

## Rules

1. Every consumer must be idempotent — duplicate delivery is normal, not exceptional.
2. Always configure dead-letter queues — unprocessable messages must not block the queue.
3. Alert on dead-letter queue depth — dead letters require investigation, not silence.
4. Use exponential backoff for retries — never retry at a fixed interval.
5. Use sessions for ordering — do not assume queue ordering without sessions enabled.
6. Set `maxDeliveryCount` between 3 and 10 — too low causes false dead-letters, too high delays detection.
