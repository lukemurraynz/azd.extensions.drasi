---
name: cosmos-db-patterns
description: >-
  Azure Cosmos DB design patterns including partition key selection, RBAC data-plane authentication, change feed processing, and performance/cost optimization.
  USE FOR: choosing partition keys, configuring Cosmos DB RBAC, implementing change feed processors, optimizing RU consumption, or designing multi-region Cosmos DB architectures.
---

# Cosmos DB Patterns

> **MUST:** Use RBAC data plane authentication with `disableLocalAuth: true`. DO NOT use primary keys or connection strings for data access.

> **Warning — Data Explorer:** Setting `disableLocalAuth: true` blocks the Azure Portal
> Data Explorer from querying data (it uses key-based auth internally). To browse data
> in the portal, temporarily re-enable local auth or use Azure Cosmos DB Explorer
> (standalone tool) with Entra ID authentication.

## Description

Patterns and best practices for Azure Cosmos DB for NoSQL — partition key design, data modelling, change feed, provisioning, and integration with managed identity.

## Capabilities

| Capability           | Details                                            |
| -------------------- | -------------------------------------------------- |
| Data Modelling       | Document design, embedding vs referencing          |
| Partition Key Design | Hot partition avoidance, hierarchical keys         |
| Query Optimisation   | Point reads, cross-partition queries, indexing     |
| Change Feed          | Event-driven processing, materialised views        |
| Consistency Levels   | Session, Eventual, Strong, Bounded Staleness       |
| Infrastructure       | Bicep deployment with RBAC and diagnostics         |
| Cost Management      | RU budgeting, serverless vs provisioned throughput |

## Standards

| Standard                                      | Purpose                     |
| --------------------------------------------- | --------------------------- |
| [Data Modelling](standards/data-modelling.md) | Document design patterns    |
| [Performance](standards/performance.md)       | Query and indexing patterns |
| [Checklist](standards/checklist.md)           | Validation checklist        |

## Actions

| Action                                          | Purpose                 |
| ----------------------------------------------- | ----------------------- |
| [Deploy Cosmos DB](actions/deploy-cosmos-db.md) | Provision and configure |

---

## Partition Key Design

Choose a partition key that distributes reads and writes evenly:

| Strategy                   | Example                         | When to Use                             |
| -------------------------- | ------------------------------- | --------------------------------------- |
| Natural key                | `/tenantId`                     | Multi-tenant — queries scoped to tenant |
| Synthetic key              | `/partitionKey` (composed)      | No single high-cardinality property     |
| Hierarchical partition key | `/tenantId`, `/region`, `/year` | Large tenants with further subdivision  |

**Good partition key properties:**

- High cardinality (many distinct values)
- Even distribution of storage and throughput
- Frequently used in WHERE clauses

**Anti-pattern:** Using a low-cardinality field like `/status` or `/country` creates hot partitions.

---

## Data Modelling

### Embed Related Data (Default)

Embed when data is read together and changed together:

```json
{
  "id": "order-001",
  "partitionKey": "tenant-abc",
  "customer": {
    "name": "Jane Smith",
    "email": "jane@example.com"
  },
  "items": [
    { "productId": "prod-1", "quantity": 2, "unitPrice": 29.99 },
    { "productId": "prod-2", "quantity": 1, "unitPrice": 49.99 }
  ],
  "total": 109.97,
  "status": "completed"
}
```

### Reference Separate Documents

Reference when data is updated independently or grows unbounded:

```json
// Order document
{
  "id": "order-001",
  "partitionKey": "tenant-abc",
  "customerId": "customer-001",
  "status": "completed"
}

// Order line items (separate documents, same partition)
{
  "id": "line-001",
  "partitionKey": "tenant-abc",
  "orderId": "order-001",
  "productId": "prod-1",
  "quantity": 2
}
```

---

## RBAC Data Plane Access

Use built-in Cosmos DB data plane roles:

| Role                                | GUID                                   | Access                  |
| ----------------------------------- | -------------------------------------- | ----------------------- |
| Cosmos DB Built-in Data Reader      | `00000000-0000-0000-0000-000000000001` | Read all data           |
| Cosmos DB Built-in Data Contributor | `00000000-0000-0000-0000-000000000002` | Read and write all data |

```bicep
resource sqlRoleAssignment 'Microsoft.DocumentDB/databaseAccounts/sqlRoleAssignments@2025-10-15' = {
  parent: cosmosAccount
  name: guid(cosmosAccount.id, principalId, '00000000-0000-0000-0000-000000000002')
  properties: {
    roleDefinitionId: '${cosmosAccount.id}/sqlRoleDefinitions/00000000-0000-0000-0000-000000000002'
    principalId: principalId
    scope: cosmosAccount.id
  }
}
```

---

## Bicep — Cosmos DB Account with Database and Container

```bicep
param location string = resourceGroup().location
param accountName string
param databaseName string = 'appdb'
param containerName string = 'items'

resource cosmosAccount 'Microsoft.DocumentDB/databaseAccounts@2025-10-15' = {
  name: accountName
  location: location
  kind: 'GlobalDocumentDB'
  properties: {
    databaseAccountOfferType: 'Standard'
    disableLocalAuth: true
    consistencyPolicy: {
      defaultConsistencyLevel: 'Session'
    }
    locations: [
      {
        locationName: location
        failoverPriority: 0
        isZoneRedundant: false
      }
    ]
    capabilities: [
      { name: 'EnableServerless' }
    ]
  }
}

> **Serverless limitation:** Serverless accounts have a **maximum response size of 4 MB**
> per query. Large result sets must be paginated using continuation tokens. Cross-partition
> queries on serverless are limited to a single round-trip per partition.

resource database 'Microsoft.DocumentDB/databaseAccounts/sqlDatabases@2025-10-15' = {
  parent: cosmosAccount
  name: databaseName
  properties: {
    resource: { id: databaseName }
  }
}

resource container 'Microsoft.DocumentDB/databaseAccounts/sqlDatabases/sqlContainers@2025-10-15' = {
  parent: database
  name: containerName
  properties: {
    resource: {
      id: containerName
      partitionKey: {
        paths: [ '/partitionKey' ]
        kind: 'Hash'
        version: 2
      }
      indexingPolicy: {
        indexingMode: 'consistent'
        includedPaths: [
          { path: '/*' }
        ]
        excludedPaths: [
          { path: '/"_etag"/?' }
        ]
      }
    }
  }
}
```

---

## Change Feed

Process changes as a stream of events:

```csharp
// Azure Functions change feed trigger (isolated worker)
[Function("ProcessChanges")]
public async Task Run(
    [CosmosDBTrigger(
        databaseName: "appdb",
        containerName: "items",
        Connection = "CosmosDb",
        LeaseContainerName = "leases",
        CreateLeaseContainerIfNotExists = true)]
    IReadOnlyList<MyDocument> changes,
    FunctionContext context)
{
    var logger = context.GetLogger("ProcessChanges");

    foreach (var doc in changes)
    {
        logger.LogInformation("Changed document: {Id}", doc.Id);
        // Process change — update materialised view, trigger event, etc.
    }
}
```

> [!WARNING]
> Change feed processor errors are silent by default. Always register an error handler:
> ```csharp
> .WithErrorNotification((leaseToken, exception) => {
>     logger.LogError(exception, "Change feed error for lease {LeaseToken}", leaseToken);
>     return Task.CompletedTask;
> })
> ```
> Without this, transient errors (throttling, partition splits) cause the processor to silently stop processing.

**Connection setting** (identity-based):

```json
{
  "CosmosDb__accountEndpoint": "https://myaccount.documents.azure.com:443/"
}
```

---

## Consistency Levels

| Level             | Guarantee                           | Latency | Cost (RU) |
| ----------------- | ----------------------------------- | ------- | --------- |
| Strong            | Linearisable reads                  | Higher  | 2x        |
| Bounded Staleness | Reads lag by k versions or t time   | Medium  | 2x        |
| Session           | Read-your-writes within session     | Low     | 1x        |
| Consistent Prefix | Reads never see out-of-order writes | Low     | 1x        |
| Eventual          | No ordering guarantee               | Lowest  | 1x        |

**Default:** Session consistency — balances read-your-writes guarantee with performance.

---

## Principles

1. **Design around the partition key** — it determines performance, cost, and scalability.
2. **Embed by default, reference when necessary** — minimise cross-document reads.
3. **Use point reads over queries** — `ReadItemAsync` is always 1 RU for items under 1 KB.
4. **Disable local authentication** — use RBAC data plane roles with managed identity.
5. **Monitor RU consumption** — set alerts for RU usage approaching provisioned limits.
6. **Use change feed for event-driven patterns** — avoid polling.
7. **Connection mode** — use Direct mode (default in SDK v3+) for production.
   Gateway mode is required when behind corporate firewalls that block direct TCP.
   Direct mode provides lower latency and higher throughput.

## References

- [Cosmos DB for NoSQL documentation](https://learn.microsoft.com/en-us/azure/cosmos-db/nosql/)
- [Partition key design](https://learn.microsoft.com/en-us/azure/cosmos-db/partitioning-overview)
- [Data modelling](https://learn.microsoft.com/en-us/azure/cosmos-db/nosql/modeling-data)
- [Change feed](https://learn.microsoft.com/en-us/azure/cosmos-db/change-feed)
- [RBAC for data plane](https://learn.microsoft.com/en-us/azure/cosmos-db/how-to-setup-rbac)

## Related Skills

- [Identity & Managed Identity](../identity-managed-identity/SKILL.md) — RBAC role assignments
- [Azure Functions Patterns](../azure-functions-patterns/SKILL.md) — Change feed triggers
- [Event-Driven Messaging](../event-driven-messaging/SKILL.md) — Change feed to messaging

---

## Currency and Verification

- **Date checked:** 2026-03-31 (verified via Microsoft Learn MCP — ARM template reference for Microsoft.DocumentDB)
- **Compatibility:** Azure Bicep, ARM templates
- **Sources:** [Microsoft.DocumentDB ARM reference](https://learn.microsoft.com/azure/templates/microsoft.documentdb/databaseaccounts)
- **Verification steps:**
  1. Run `az provider show --namespace Microsoft.DocumentDB --query "resourceTypes[?resourceType=='databaseAccounts'].apiVersions" -o tsv` and confirm `2025-10-15` is listed
  2. Run `az bicep build --file <your-bicep-file>` to validate syntax

### Known Pitfalls

| Area                     | Pitfall                                                                                          | Mitigation                                                                                         |
| ------------------------ | ------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------- |
| `disableLocalAuth: true` | Azure Portal Data Explorer stops working (uses key-based auth)                                   | Use standalone Cosmos DB Explorer with Entra ID, or temporarily re-enable local auth for debugging |
| Serverless response size | Maximum 4 MB per query response; silent truncation if exceeded                                   | Use continuation tokens for pagination; monitor response sizes                                     |
| API version drift        | `2025-10-15` is current but newer versions may add features (e.g., vector search enhancements) | Verify with `az provider show` before deploying; consider upgrading for new feature access         |
