# Action: Implement Messaging

Add event-driven messaging to a service using Azure Service Bus, Event Grid, or Event Hubs.

---

## Step 1 — Choose the Right Service

| Need                             | Use               |
| -------------------------------- | ----------------- |
| Reliable command/task processing | Service Bus Queue |
| Pub/sub with filtering           | Service Bus Topic |
| System event notifications       | Event Grid        |
| High-throughput streaming        | Event Hubs        |

---

## Step 2 — Deploy Infrastructure

Deploy the messaging infrastructure via Bicep. See SKILL.md for complete examples.

Key settings:

- `disableLocalAuth: true` on all namespace resources
- Dead-lettering enabled on all queues and subscriptions
- `maxDeliveryCount` set between 3–10

---

## Step 3 — Configure Authentication (Managed Identity)

```bash
# Grant sender role
az role assignment create \
  --assignee $senderPrincipalId \
  --role "Azure Service Bus Data Sender" \
  --scope $serviceBusId

# Grant receiver role
az role assignment create \
  --assignee $receiverPrincipalId \
  --role "Azure Service Bus Data Receiver" \
  --scope $serviceBusId
```

---

## Step 4 — Implement Producer

```csharp
// Using DefaultAzureCredential (Managed Identity)
var client = new ServiceBusClient(
    "{namespace}.servicebus.windows.net",
    new DefaultAzureCredential());

var sender = client.CreateSender("orders");

var message = new ServiceBusMessage(JsonSerializer.SerializeToUtf8Bytes(new
{
    id = Guid.NewGuid().ToString(),
    type = "CreateOrder",
    version = "1.0",
    source = "api-service",
    time = DateTimeOffset.UtcNow,
    correlationId = Activity.Current?.Id,
    data = orderRequest
}));

await sender.SendMessageAsync(message);
```

---

## Step 5 — Implement Consumer

```csharp
var processor = client.CreateProcessor("orders", new ServiceBusProcessorOptions
{
    MaxConcurrentCalls = 10,
    AutoCompleteMessages = false
});

processor.ProcessMessageAsync += async args =>
{
    var messageId = args.Message.MessageId;

    if (await _idempotencyStore.HasBeenProcessedAsync(messageId))
    {
        await args.CompleteMessageAsync(args.Message);
        return;
    }

    try
    {
        var envelope = args.Message.Body.ToObjectFromJson<MessageEnvelope>();
        await _handler.HandleAsync(envelope);
        await _idempotencyStore.MarkProcessedAsync(messageId);
        await args.CompleteMessageAsync(args.Message);
    }
    catch (Exception ex)
    {
        _logger.LogError(ex, "Failed processing message {MessageId}", messageId);
        // Let Service Bus handle retry via delivery count
    }
};

processor.ProcessErrorAsync += args =>
{
    _logger.LogError(args.Exception, "Service Bus processor error");
    return Task.CompletedTask;
};

await processor.StartProcessingAsync();
```

---

## Step 6 — Configure Monitoring

- Set up dead-letter queue depth alerts (see reliability standard)
- Enable diagnostic settings on the namespace
- Verify messages flow end-to-end with distributed tracing

---

## Completion Criteria

- [ ] Messaging infrastructure deployed via IaC
- [ ] Managed Identity authentication configured (no connection strings)
- [ ] Producer sending messages with envelope pattern
- [ ] Consumer processing messages idempotently
- [ ] Dead-letter handling and alerting configured
- [ ] Distributed tracing correlation verified end-to-end
