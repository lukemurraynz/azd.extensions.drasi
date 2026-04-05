# Event-Driven Messaging Checklist

## Service Bus

- [ ] `disableLocalAuth: true` — Managed Identity only
- [ ] `minimumTlsVersion: '1.2'`
- [ ] SKU appropriate for workload (Standard or Premium)
- [ ] Queues have `maxDeliveryCount` set (3–10)
- [ ] `deadLetteringOnMessageExpiration: true` on all queues
- [ ] `lockDuration` set appropriately (default PT1M)
- [ ] Sessions enabled where ordering is required
- [ ] Duplicate detection enabled where needed

## Event Grid

- [ ] Event subscriptions have retry policy configured
- [ ] Dead-letter destination configured (storage blob)
- [ ] Subject/type filters applied (no catch-all subscriptions)
- [ ] Managed Identity used for delivery authentication

## Event Hubs

- [ ] Partition count set for expected throughput
- [ ] Consumer groups created per consuming application
- [ ] Capture enabled for archival (if needed)
- [ ] `disableLocalAuth: true`

## Message Design

- [ ] Messages use envelope pattern with `id`, `type`, `version`, `correlationId`
- [ ] Events use past tense naming (`OrderCreated`)
- [ ] Commands use imperative naming (`CreateOrder`)
- [ ] Payloads < 64 KB (Claim Check for larger)
- [ ] No secrets or PII in message bodies

## Consumer Reliability

- [ ] All consumers are idempotent
- [ ] Dead-letter queue monitoring and alerting configured
- [ ] Poison message handling implemented
- [ ] Correlation IDs propagated for distributed tracing
- [ ] Error handling distinguishes transient vs permanent failures

## Infrastructure

- [ ] All messaging resources defined in IaC (Bicep/Terraform)
- [ ] RBAC roles assigned (Data Sender/Receiver) instead of connection strings
- [ ] Diagnostic settings configured for monitoring
- [ ] Network access restricted (Private Endpoint or service firewall)
