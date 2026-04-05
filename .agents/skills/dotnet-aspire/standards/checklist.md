# Aspire Readiness Checklist

Validate before marking Aspire adoption or deployment complete.

---

## Service Defaults

- [ ] `AddServiceDefaults()` called in every service project's `Program.cs`
- [ ] `MapDefaultEndpoints()` called after `app.Build()` for health check endpoints
- [ ] Service Defaults project referenced by all service projects
- [ ] OpenTelemetry exporter configured for production backend (App Insights / OTLP endpoint)
- [ ] Custom resilience overrides applied for clients that must not use global retry defaults
- [ ] Health checks include domain-specific readiness probes (database, cache, queues)
- [ ] Health check endpoints do not leak internal IP addresses, service versions, or secrets

## Client Integrations

- [ ] Connection string names match between AppHost resources and service `Add*()` calls
- [ ] All `ConnectionStrings__<name>` env vars documented for production deployment
- [ ] Fail-fast validation added for required connection strings at startup
- [ ] Client integrations use typed settings callbacks for environment-specific config

## AppHost

- [ ] Resource graph models all services, databases, caches, and message brokers
- [ ] `WaitFor()` / `WaitForStart()` ordering prevents startup race conditions
- [ ] `WithHealthCheck()` added to infrastructure resources
- [ ] Emulators configured for Azure services used in local dev (`RunAsEmulator()`)
- [ ] External services declared with health checks where applicable (9.5+)
- [ ] AppHost project excluded from Docker production images and deployment artifacts

## Deployment

- [ ] Compute environment chosen and publisher configured (ACA / K8s / Compose / App Service)
- [ ] Generated IaC reviewed against production Bicep/K8s standards — not deployed as-is
- [ ] For ACA: `azd init` detects AppHost and `azd up` completes successfully
- [ ] For K8s: generated manifests overlaid with probes, resource limits, network policies
- [ ] Service discovery configured for production (DNS, config-based, or ACA internal DNS)
- [ ] OTLP endpoint configured for production telemetry backend
- [ ] Aspire Dashboard NOT deployed to production

## Integration Testing

- [ ] At least one `DistributedApplicationTestingBuilder` test validates the full resource graph
- [ ] Test health endpoints return 200 for all services
- [ ] Integration tests run in CI (may require Docker-in-Docker or similar)

## Security

- [ ] No secrets in AppHost code (use `AddParameter("name", secret: true)`)
- [ ] Connection strings in production use managed identity where possible
- [ ] Health check endpoints bypass authentication middleware
