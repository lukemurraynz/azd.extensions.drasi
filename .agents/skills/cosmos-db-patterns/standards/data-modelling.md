# Data Modelling

## Document Design Principles

### Embed Related Data (Default)

Embed data that is:

- Read together in a single query
- Updated together atomically
- Bounded in size (won't grow unbounded)

```json
{
  "id": "order-001",
  "partitionKey": "tenant-abc",
  "type": "order",
  "customer": {
    "id": "cust-001",
    "name": "Jane Smith",
    "email": "jane@example.com"
  },
  "items": [
    {
      "productId": "prod-1",
      "name": "Widget",
      "quantity": 2,
      "unitPrice": 29.99
    },
    {
      "productId": "prod-2",
      "name": "Gadget",
      "quantity": 1,
      "unitPrice": 49.99
    }
  ],
  "total": 109.97,
  "status": "completed",
  "createdAt": "2024-01-15T10:30:00Z"
}
```

### Reference When Necessary

Use references when:

- Related data changes independently
- Related data is unbounded (e.g., comments, log entries)
- Related data is queried independently
- Document would exceed 2 MB limit

```json
// Customer document
{ "id": "cust-001", "partitionKey": "cust-001", "type": "customer", "name": "Jane Smith" }

// Order document (references customer)
{ "id": "order-001", "partitionKey": "cust-001", "type": "order", "customerId": "cust-001" }
```

---

## Multi-Type Containers

Store multiple document types in the same container with a discriminator:

```json
{ "id": "cust-001", "partitionKey": "tenant-abc", "type": "customer", "name": "Jane" }
{ "id": "order-001", "partitionKey": "tenant-abc", "type": "order", "total": 109.97 }
```

Query by type:

```sql
SELECT * FROM c WHERE c.partitionKey = 'tenant-abc' AND c.type = 'order'
```

**Benefits:** Reduces container count, enables transactions across types within the same partition.

---

## Denormalisation Patterns

### Materialised View via Change Feed

When data is referenced across partitions, denormalise through change feed processing:

1. Source document changes → change feed fires
2. Function processes change → updates denormalised view in target container/partition

```
[Source Container] → [Change Feed] → [Azure Function] → [View Container]
```

Use this for read-heavy scenarios where query performance matters more than write complexity.

---

## Document Size Guidelines

| Guideline             | Recommendation                            |
| --------------------- | ----------------------------------------- |
| Maximum document size | 2 MB (hard limit)                         |
| Target document size  | Under 100 KB for optimal performance      |
| Embedded arrays       | Keep under 100 items; reference if larger |
| Embedded objects      | Maximum 2-3 levels of nesting             |
| Large text/binary     | Store in Blob Storage, reference by URL   |

---

## Rules

1. Start with embedding — only reference when there's a clear reason.
2. Use a `type` discriminator when storing multiple document types per container.
3. Keep documents under 100 KB for optimal performance.
4. Use change feed for cross-partition denormalisation.
5. Never store large binary data in Cosmos DB — use Blob Storage with a reference.
