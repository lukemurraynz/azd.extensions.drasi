# Message Design Standard

## Message Envelope

Every message must use an envelope with metadata:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "OrderCreated",
  "version": "1.0",
  "source": "order-service",
  "time": "2024-01-15T10:30:00Z",
  "correlationId": "trace-abc-123",
  "data": {
    "orderId": "ORD-12345",
    "customerId": "CUST-678",
    "total": 149.99
  }
}
```

### Required Envelope Fields

| Field           | Type     | Description                      |
| --------------- | -------- | -------------------------------- |
| `id`            | string   | Unique message identifier (UUID) |
| `type`          | string   | Event/command type name          |
| `version`       | string   | Schema version (semver)          |
| `source`        | string   | Originating service name         |
| `time`          | datetime | UTC timestamp of creation        |
| `correlationId` | string   | Distributed trace correlation ID |
| `data`          | object   | Payload — schema varies by type  |

---

## Schema Versioning

### Strategy: Additive Changes Only

- **Compatible changes** (no version bump): Adding optional fields, adding new event types
- **Breaking changes** (major version): Removing fields, changing field types, renaming fields

### Handling Multiple Versions

```csharp
public async Task ProcessMessageAsync(ProcessMessageEventArgs args)
{
    var envelope = args.Message.Body.ToObjectFromJson<MessageEnvelope>();

    switch (envelope.Type, envelope.Version)
    {
        case ("OrderCreated", "1.0"):
            await HandleOrderCreatedV1(envelope.Data);
            break;
        case ("OrderCreated", "2.0"):
            await HandleOrderCreatedV2(envelope.Data);
            break;
        default:
            _logger.LogWarning("Unknown message type {Type} v{Version}", envelope.Type, envelope.Version);
            await args.DeadLetterMessageAsync(args.Message, "UnknownType");
            break;
    }
}
```

---

## Message Types

### Commands (Imperative — "Do this")

- Named as verb-noun: `CreateOrder`, `ProcessPayment`
- Sent to a specific queue
- Exactly one consumer processes the message
- Sender expects the action to be performed

### Events (Past tense — "This happened")

- Named as noun-past-verb: `OrderCreated`, `PaymentProcessed`
- Published to a topic
- Multiple subscribers may react independently
- Publisher does not know or care about consumers

---

## Message Size

| Service     | Max Message Size                    | Recommendation |
| ----------- | ----------------------------------- | -------------- |
| Service Bus | 256 KB (Standard), 100 MB (Premium) | Keep < 64 KB   |
| Event Grid  | 1 MB                                | Keep < 256 KB  |
| Event Hubs  | 1 MB (Standard), 20 MB (Dedicated)  | Keep < 256 KB  |

For large payloads, use the **Claim Check** pattern:

1. Store payload in Blob Storage
2. Send message with blob URI reference
3. Consumer retrieves payload from blob

---

## Rules

1. Always include `id`, `type`, `version`, and `correlationId` in message envelopes.
2. Use past tense for events (`OrderCreated`), imperative for commands (`CreateOrder`).
3. Keep message payloads small — use Claim Check for > 64 KB.
4. Never include secrets or PII in message bodies.
5. Design schemas for additive-only changes — avoid breaking consumers.
