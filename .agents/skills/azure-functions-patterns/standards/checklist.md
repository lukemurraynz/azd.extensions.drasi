# Azure Functions Checklist

## Function Design

- [ ] Each function has a single responsibility
- [ ] Functions use the isolated worker model
- [ ] Functions are stateless — external state in storage/database
- [ ] HTTP functions define explicit routes and HTTP methods
- [ ] Timer functions use appropriate cron schedules
- [ ] Functions complete within timeout limits (5 min Consumption, 10 min Premium)

## Triggers and Bindings

- [ ] Trigger type matches the workload pattern
- [ ] Identity-based connections used (`__fullyQualifiedNamespace` suffix)
- [ ] No connection strings or keys in application settings
- [ ] Service Bus/Queue triggers use appropriate `maxConcurrentCalls`
- [ ] Blob triggers use Event Grid-based triggering for reliability

## Hosting

- [ ] Hosting plan matches workload requirements (Consumption/Flex/Premium/Dedicated)
- [ ] Flex Consumption configured for per-function scaling
- [ ] Premium plan used for VNet integration or long-running functions
- [ ] Deployment slots configured for production workloads

## Identity and Security

- [ ] System-assigned managed identity enabled
- [ ] RBAC roles assigned for all accessed resources (Storage, Service Bus, Key Vault)
- [ ] Function-level authorisation configured for HTTP triggers
- [ ] `disableLocalAuth` not available for Functions — use API Management for key management

## Durable Functions

- [ ] Orchestrator functions are deterministic (no I/O, no random, no DateTime.Now)
- [ ] Activities used for all side effects
- [ ] Timeouts set on all external event waits
- [ ] Instance history purged for completed orchestrations
- [ ] Sub-orchestrations used for complex workflows

## Error Handling

- [ ] Retry policies configured for transient failures
- [ ] Dead-letter queues monitored for failed messages
- [ ] Structured logging with correlation IDs
- [ ] Application Insights connected for monitoring

## Infrastructure as Code

- [ ] Function App deployed via Bicep/Terraform
- [ ] Application settings managed through IaC (not portal)
- [ ] Storage account deployed with managed identity access
- [ ] Diagnostic settings configured for Application Insights
