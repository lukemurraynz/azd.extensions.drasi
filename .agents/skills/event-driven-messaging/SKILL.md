---
name: event-driven-messaging
description: >-
  Event-driven architecture patterns using Azure Service Bus, Event Grid, and Event Hubs. Covers message design, dead-letter handling, idempotency, ordering, and Bicep deployment. USE FOR: designing asynchronous communication between services.compatibility: Requires Azure CLI, Azure SDK for target language
---

# Event-Driven Messaging Skill

> **Mandatory:** All messaging services must use Managed Identity for authentication.
> Use RBAC roles (`Azure Service Bus Data Sender/Receiver`, `Azure Event Hubs Data Sender/Receiver`)
> instead of connection strings with shared access keys.

---

## Quick Reference

| Capability       | Description                                                        |
| ---------------- | ------------------------------------------------------------------ |
| Service Bus      | Enterprise messaging with queues, topics, sessions, transactions   |
| Event Grid       | Event routing with push delivery, filtering, and dead-lettering    |
| Event Hubs       | High-throughput event streaming for telemetry and analytics        |
| Message Patterns | Request/reply, pub/sub, competing consumers, saga orchestration    |
| Reliability      | Dead-letter handling, retry policies, idempotency, poison messages |
| Ordering         | Session-based ordering, partition keys, sequence numbers           |

---

## When to Use What

| Scenario                         | Service              | Why                                              |
| -------------------------------- | -------------------- | ------------------------------------------------ |
| Command/task processing          | Service Bus Queue    | Guaranteed delivery, dead-lettering              |
| Pub/sub with filtering           | Service Bus Topic    | Topic subscriptions with SQL/correlation filters |
| Event notification (fire-forget) | Event Grid           | Push delivery, low latency, serverless           |
| High-volume telemetry/streaming  | Event Hubs           | Partitioned, high throughput, replay             |
| Ordered processing               | Service Bus Sessions | Session-based FIFO guarantee                     |
| Cross-service integration        | Event Grid           | System events, custom topics, webhooks           |

> **Sessions caveat:** When `requiresSession: true` is set on a queue or subscription,
> **all** consumers must use session-aware receivers (`ServiceBusSessionProcessor`).
> A non-session-aware receiver will fail with a `ServiceBusException`. You cannot mix
> session and non-session consumers on the same entity.

---

## Standards

| Standard                                         | Purpose                         |
| ------------------------------------------------ | ------------------------------- |
| [Message Design](standards/message-design.md)    | Message structure and contracts |
| [Reliability Patterns](standards/reliability.md) | Dead-letter, retry, idempotency |
| [Checklist](standards/checklist.md)              | Validation checklist            |

---

## Actions

| Action                                                | When to use                   |
| ----------------------------------------------------- | ----------------------------- |
| [Implement Messaging](actions/implement-messaging.md) | Adding messaging to a service |

---

## Service Bus — Bicep

````bicep
resource serviceBus 'Microsoft.ServiceBus/namespaces@2024-01-01' = {
  name: 'sb-${environmentName}'
  location: location
  sku: { name: 'Standard', tier: 'Standard' }
  properties: {
    disableLocalAuth: true  // Managed Identity only
    minimumTlsVersion: '1.2'
  }
}

resource queue 'Microsoft.ServiceBus/namespaces/queues@2024-01-01' = {
  parent: serviceBus
  name: 'orders'
  properties: {
    maxDeliveryCount: 5
    lockDuration: 'PT1M'
    defaultMessageTimeToLive: 'P7D'
    deadLetteringOnMessageExpiration: true
    requiresSession: false
  }
}

> **Warning — lock duration vs processing time:** The default `lockDuration` is 1 minute.
> If your message handler takes longer, the lock expires and the broker **redelivers the
> message to another consumer**, causing duplicate processing. Either increase `lockDuration`
> (max PT5M) or configure `MaxAutoLockRenewalDuration` on the `ServiceBusProcessor`:
>
> ```csharp
> var processor = client.CreateProcessor(queueName, new ServiceBusProcessorOptions
> {
>     MaxAutoLockRenewalDuration = TimeSpan.FromMinutes(10)
> });
> ```

resource topic 'Microsoft.ServiceBus/namespaces/topics@2024-01-01' = {
  parent: serviceBus
  name: 'events'
  properties: {
    defaultMessageTimeToLive: 'P7D'
    enablePartitioning: false
  }
}

resource subscription 'Microsoft.ServiceBus/namespaces/topics/subscriptions@2024-01-01' = {
  parent: topic
  name: 'order-processor'
  properties: {
    maxDeliveryCount: 5
    lockDuration: 'PT1M'
    deadLetteringOnMessageExpiration: true
  }
}
````

---

## Event Grid — System Topic

```bicep
resource eventGridTopic 'Microsoft.EventGrid/systemTopics@2025-02-15' = {
  name: 'evgt-${environmentName}'
  location: location
  properties: {
    source: storageAccount.id
    topicType: 'Microsoft.Storage.StorageAccounts'
  }
}

resource eventSubscription 'Microsoft.EventGrid/systemTopics/eventSubscriptions@2025-02-15' = {
  parent: eventGridTopic
  name: 'blob-created'
  properties: {
    destination: {
      endpointType: 'AzureFunction'
      properties: {
        resourceId: functionApp.id
      }
    }
    filter: {
      subjectBeginsWith: '/blobServices/default/containers/uploads'
      includedEventTypes: ['Microsoft.Storage.BlobCreated']
    }
    retryPolicy: {
      maxDeliveryAttempts: 5
      eventTimeToLiveInMinutes: 1440
    }
    deadLetterDestination: {
      endpointType: 'StorageBlob'
      properties: {
        resourceId: storageAccount.id
        blobContainerName: 'dead-letters'
      }
    }
  }
}
```

---

## Message Processing Pattern

```csharp
// Idempotent message handler with dead-letter support
public async Task ProcessMessageAsync(ProcessMessageEventArgs args)
{
    var messageId = args.Message.MessageId;

    // Idempotency check
    if (await _idempotencyStore.HasBeenProcessedAsync(messageId))
    {
        await args.CompleteMessageAsync(args.Message);
        return;
    }

    try
    {
        var order = args.Message.Body.ToObjectFromJson<OrderEvent>();
        await _orderService.ProcessAsync(order);
        await _idempotencyStore.MarkProcessedAsync(messageId);
        await args.CompleteMessageAsync(args.Message);
    }
    catch (Exception ex) when (args.Message.DeliveryCount >= 4)
    {
        _logger.LogError(ex, "Message {MessageId} exceeded max retries, dead-lettering", messageId);
        await args.DeadLetterMessageAsync(args.Message, "MaxRetriesExceeded", ex.Message);
    }
}
```

---

## Principles

1. **Idempotent consumers** — every handler must safely process the same message twice.
2. **Dead-letter everything** — configure dead-letter queues/destinations on all subscriptions.
3. **Schema versioning** — use envelope pattern with `type` and `version` fields.
4. **Managed Identity** — disable local auth; use RBAC for all service access.
5. **Correlate across boundaries** — propagate `OperationId` in message properties.
6. **Right-size the service** — Service Bus for commands, Event Grid for events, Event Hubs for streams.

---

## References

- [Azure Service Bus overview](https://learn.microsoft.com/azure/service-bus-messaging/service-bus-messaging-overview)
- [Azure Event Grid overview](https://learn.microsoft.com/azure/event-grid/overview)
- [Azure Event Hubs overview](https://learn.microsoft.com/azure/event-hubs/event-hubs-about)
- [Cloud messaging patterns](https://learn.microsoft.com/azure/architecture/patterns/category/messaging)

---

## Related Skills

- **azure-container-apps** — Container Apps with queue-based scaling
- **azure-functions-patterns** — Function triggers for Service Bus and Event Grid
- **observability-monitoring** — Distributed tracing across message boundaries
- **identity-managed-identity** — RBAC roles for messaging services

---

## Currency and Verification

- **Date checked:** 2026-03-31
- **API versions updated:**
  - Service Bus: `2024-01-01` (GA) — upgraded from `2022-10-01-preview`
  - Event Grid systemTopics: `2025-02-15` (GA) — upgraded from `2023-12-15-preview`
- **Sources:** [Service Bus ARM reference](https://learn.microsoft.com/azure/templates/microsoft.servicebus/namespaces), [Event Grid ARM reference](https://learn.microsoft.com/azure/templates/microsoft.eventgrid/systemtopics)
- **Verification steps:**
  1. Verify Service Bus API version: `az provider show --namespace Microsoft.ServiceBus --query "resourceTypes[?resourceType=='namespaces'].apiVersions" -o tsv`
  2. Verify Event Grid API version: `az provider show --namespace Microsoft.EventGrid --query "resourceTypes[?resourceType=='systemTopics'].apiVersions" -o tsv`

### Known Pitfalls

| Area                             | Pitfall                                                                                              | Mitigation                                                                                             |
| -------------------------------- | ---------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------ |
| Dead-letter queue                | Messages exceeding `MaxDeliveryCount` silently move to dead-letter queue with no default alerting    | Monitor dead-letter queue depth with Azure Monitor alerts; implement dead-letter processing            |
| Event Grid retry exhaustion      | Default retry is 24 hours with exponential backoff; failed events are dropped after retry exhaustion | Configure dead-letter destination on Event Grid subscriptions for undeliverable events                 |
| Session ordering                 | Service Bus sessions guarantee FIFO per session ID, not globally across the topic                    | Use session IDs consistently; don't assume global ordering without sessions                            |
| Duplicate delivery               | Both Service Bus and Event Grid deliver at-least-once; duplicates are possible during retries        | Implement idempotent message handlers; use deduplication with `MessageId` or event checkpoints         |
| Large message payloads           | Service Bus Standard is limited to 256 KB per message; Premium supports up to 100 MB                 | Use claim-check pattern (store payload in Blob Storage, send reference) for large payloads             |
| Managed identity for Service Bus | Connection strings with SAS keys are still common in examples but violate least-privilege            | Use managed identity with `Azure Service Bus Data Sender/Receiver` roles instead of connection strings |
