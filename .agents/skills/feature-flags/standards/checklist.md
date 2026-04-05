# Feature Flags Checklist

## App Configuration Store

- [ ] Standard SKU used (Free doesn't support feature flags fully)
- [ ] `disableLocalAuth: true` — managed identity only
- [ ] App Configuration Data Reader role assigned to applications
- [ ] Diagnostic settings configured to Log Analytics

## Feature Flag Definition

- [ ] Flags use consistent naming convention (PascalCase or Domain.Feature)
- [ ] Each flag has a description explaining its purpose
- [ ] Each flag has a planned cleanup date
- [ ] Flag type identified (release, experiment, operational, permission)

## Application Integration

- [ ] `AddAzureAppConfiguration()` configured with feature flags
- [ ] `UseAzureAppConfiguration()` middleware registered
- [ ] `AddFeatureManagement()` registered in DI
- [ ] Cache expiration interval set (30 seconds default)
- [ ] Identity-based connection (endpoint, not connection string)

## Targeting and Rollout

- [ ] Targeting filter registered for percentage/group rollout
- [ ] `ITargetingContextAccessor` implemented for user context
- [ ] Rollout plan defined (internal → canary → gradual → full)
- [ ] Rollback plan documented (disable flag)

## Observability

- [ ] Flag evaluations logged with structured logging
- [ ] Flag state changes tracked in audit log
- [ ] Metrics on flag-gated code paths monitored
- [ ] Alerts configured for operational flags (circuit breakers)

## Lifecycle Management

- [ ] Fully-rolled-out flags cleaned up within 30 days
- [ ] Flag code paths and alternative paths removed after cleanup
- [ ] No permanent feature flags without explicit justification
