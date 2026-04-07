# PostgreSQL source template

Scaffolds a Drasi project that watches an Azure Database for PostgreSQL Flexible Server for row-level changes using logical replication.

## What you get

- A Drasi source configured for PostgreSQL (`type: PostgreSQL`)
- A sample continuous query that watches an `orders` table
- A debug/log reaction for observing change events
- Bicep modules that provision:
  - Azure Database for PostgreSQL Flexible Server (Burstable B1ms, logical replication enabled)
  - Shared Drasi infrastructure (AKS, Key Vault, managed identity, Log Analytics)
- A `docker-compose.yml` for local PostgreSQL development
- Environment overlay support (`drasi/environments/dev.yaml`)

## Prerequisites

| Tool | Version |
| ----- | --------- |
| Azure Developer CLI (`azd`) | >= 1.10.0 |
| Drasi CLI | >= 0.10.0 |
| Azure CLI | >= 2.60.0 |
| Docker | >= 24.0 |

## Quick start

```bash
azd drasi init --template postgresql-source
azd drasi validate
azd auth login
azd drasi provision
azd drasi deploy
azd drasi status
```

## Customizing

### Change the watched tables

Edit `drasi/sources/pg-source.yaml` and update the `tables` property with your table names.

### Modify the continuous query

Edit `drasi/queries/watch-orders.yaml`. The query uses Cypher syntax to match nodes from the source.

### Add a production reaction

Replace or extend `drasi/reactions/log-changes.yaml` with a Dapr pub/sub, SignalR, or webhook reaction. See the [Drasi reactions reference](https://drasi.io) for available types.

### Use environment overlays

Place environment-specific overrides in `drasi/environments/<name>.yaml` and deploy with:

```bash
azd drasi deploy --environment <name>
```

## Infrastructure notes

The `infra/modules/postgresql.bicep` module provisions a Burstable B1ms server with public network access and logical replication enabled. This configuration is intended for development. For production, consider private endpoints and a larger SKU.

The admin password uses `uniqueString()` as a placeholder for development only. Replace it with a Key Vault secret reference for production deployments.
