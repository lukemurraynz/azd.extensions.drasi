# Cosmos DB Checklist

## Account Configuration

- [ ] `disableLocalAuth: true` — no primary keys for data access
- [ ] Consistency level set appropriately (Session default)
- [ ] Serverless or autoscale based on workload pattern
- [ ] Diagnostic settings configured to Log Analytics
- [ ] Backup policy configured (continuous preferred)

## Database and Container Design

- [ ] Partition key ensures even distribution of reads and writes
- [ ] Partition key is high cardinality (many distinct values)
- [ ] Partition key used in most queries as a filter
- [ ] Hierarchical partition key considered for large tenants
- [ ] Container-per-workload, not container-per-entity type

## Data Modelling

- [ ] Related data embedded by default
- [ ] References used only for unbounded or independently updated data
- [ ] `type` discriminator for multi-type containers
- [ ] Documents under 100 KB target size
- [ ] Large binary data stored in Blob Storage with reference

## Indexing

- [ ] Custom indexing policy — excluded unused paths
- [ ] Composite indexes for multi-property ORDER BY
- [ ] Spatial indexes only if geo-queries needed
- [ ] Indexing mode set to `consistent`

## Query Performance

- [ ] Point reads used for known id + partition key
- [ ] All queries include partition key filter
- [ ] Cross-partition queries avoided or paginated
- [ ] Projections used to reduce response size
- [ ] `MaxItemCount` set for pagination

## Identity and Security

- [ ] RBAC data plane roles assigned (Built-in Data Reader/Contributor)
- [ ] Managed identity used for all application access
- [ ] No connection strings in application configuration
- [ ] Network restrictions configured (private endpoint or IP rules)

## Change Feed

- [ ] Lease container created for change feed processing
- [ ] Change feed processor handles errors without dropping events
- [ ] Materialised views updated via change feed (not sync queries)
- [ ] Change feed functions use identity-based connections

## Infrastructure as Code

- [ ] Cosmos DB account deployed via Bicep/Terraform
- [ ] Database and containers managed through IaC
- [ ] RBAC role assignments in IaC
- [ ] Diagnostic settings in IaC
