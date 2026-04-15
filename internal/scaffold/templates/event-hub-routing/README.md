# Event Hub routing template

Scaffolds a Drasi project that routes events from Azure Event Hubs to continuous queries for real-time stream processing.

## What you get

- A Drasi source configured for Azure Event Hubs (`kind: EventHub`)
- A sample continuous query that processes event data
- A debug/log reaction for observing query results
- Bicep modules that provision:
  - Azure Event Hubs namespace with an event hub
  - Shared Drasi infrastructure (AKS, Key Vault, managed identity, Log Analytics)
  - Key Vault secret for secure connection string storage
- Environment overlay support (`drasi/environments/dev.yaml`)

## Prerequisites

| Tool | Version |
| ----- | --------- |
| Azure Developer CLI (`azd`) | >= 1.10.0 |
| Drasi CLI | >= 0.10.0 |
| Azure CLI | >= 2.60.0 |

## Quick start

```bash
azd drasi init --template event-hub-routing
azd drasi validate
azd auth login
azd drasi provision
azd drasi deploy
azd drasi status
```

## Customizing

### Send events to Event Hub

Send test events using the Azure CLI:

```bash
az eventhubs eventhub send \
  --namespace-name <your-namespace> \
  --name drasi-events \
  --message '{"id": "1", "data": "test"}' \
  --producer-id test
```

### Modify the continuous query

Edit `drasi/queries/eventhub-query.yaml`. The query uses Cypher syntax to match events from the source.

### Add a production reaction

Replace or extend `drasi/reactions/example-reaction.yaml` with a Dapr pub/sub, SignalR, or webhook reaction. See the [Drasi reactions reference](https://drasi.io) for available types.

## Infrastructure notes

The `infra/modules/eventhub.bicep` module provisions:
- Event Hubs Standard tier namespace
- Event hub with 4 partitions
- Shared access policy with Listen/Send permissions
- Connection string stored in Key Vault

The connection string is auto-stored during `azd drasi provision` and synced to AKS via secretMappings. No secrets are stored in source control.

### Changing the deployment location

Azure Event Hubs availability varies by region. If provisioning fails with a location restriction error:

1. Check available regions: `az eventhubs namespace list-skus --location <region>`
2. Set your preferred location: `azd env set AZURE_LOCATION <region>`
3. Re-run: `azd drasi provision`
