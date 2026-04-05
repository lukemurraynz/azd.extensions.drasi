# Performance

## Point Reads vs Queries

| Operation        | RU Cost       | When to Use                           |
| ---------------- | ------------- | ------------------------------------- |
| Point read       | 1 RU (< 1 KB) | Know `id` and partition key           |
| Single-partition | Low           | Query within one logical partition    |
| Cross-partition  | High          | Avoid — fan-out across all partitions |

**Always prefer point reads:**

```csharp
// Point read — 1 RU, always efficient
var response = await container.ReadItemAsync<Order>(
    id: "order-001",
    partitionKey: new PartitionKey("tenant-abc"));

// Single-partition query — efficient
var query = container.GetItemQueryIterator<Order>(
    new QueryDefinition("SELECT * FROM c WHERE c.status = @status")
        .WithParameter("@status", "completed"),
    requestOptions: new QueryRequestOptions
    {
        PartitionKey = new PartitionKey("tenant-abc")
    });
```

---

## Indexing Policy

### Custom Indexing (Recommended)

Exclude paths you don't query on and include only what you need:

```json
{
  "indexingMode": "consistent",
  "includedPaths": [
    { "path": "/status/?" },
    { "path": "/createdAt/?" },
    { "path": "/type/?" }
  ],
  "excludedPaths": [{ "path": "/*" }],
  "compositeIndexes": [
    [
      { "path": "/status", "order": "ascending" },
      { "path": "/createdAt", "order": "descending" }
    ]
  ]
}
```

### Composite Indexes

Required for:

- `ORDER BY` on multiple properties
- Filters with `ORDER BY` on a different property
- Queries with multiple range filters

---

## RU Optimisation

| Technique                  | Impact                                 |
| -------------------------- | -------------------------------------- |
| Point reads over queries   | 1 RU vs variable                       |
| Exclude unused index paths | Reduces write RUs by 10-30%            |
| Projection (`SELECT c.id`) | Reduces read RUs                       |
| Pagination                 | `MaxItemCount` limits per-request cost |
| Co-locate related data     | Single read instead of multiple joins  |

### Pagination

```csharp
var queryOptions = new QueryRequestOptions
{
    MaxItemCount = 25,
    PartitionKey = new PartitionKey("tenant-abc")
};

var iterator = container.GetItemQueryIterator<Order>(query, requestOptions: queryOptions);

while (iterator.HasMoreResults)
{
    var page = await iterator.ReadNextAsync();
    // Process page.Resource
    // page.RequestCharge shows RU cost
    // page.ContinuationToken for next page
}
```

---

## Throughput Patterns

| Mode        | Best For                                | Billing        |
| ----------- | --------------------------------------- | -------------- |
| Serverless  | Development, low/sporadic traffic       | Per-request RU |
| Autoscale   | Variable workloads, predictable scaling | Per-hour max   |
| Provisioned | Steady, predictable workloads           | Fixed RU/s     |

**Default:** Serverless for development and low-traffic. Autoscale for production.

---

## Rules

1. Use point reads for known `id` + partition key lookups.
2. Always include the partition key in queries.
3. Customise indexing policy — exclude paths you don't query.
4. Use composite indexes for multi-property ORDER BY.
5. Monitor RU consumption per query and optimise the most expensive ones.
