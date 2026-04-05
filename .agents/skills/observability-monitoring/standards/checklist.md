# Observability Checklist

## Instrumentation

- [ ] OpenTelemetry SDK configured with Azure Monitor exporter
- [ ] Application Insights connection string from environment variable (not hardcoded)
- [ ] `DisableLocalAuth: true` set on Application Insights resource
- [ ] Tracing enabled for HTTP server and client
- [ ] Metrics enabled for HTTP server and client
- [ ] Logging provider connected to OpenTelemetry

## Structured Logging

- [ ] All log messages use semantic templates (no string interpolation)
- [ ] Log levels configured per-category in `appsettings.json` / environment config
- [ ] No PII, secrets, or credentials in log output
- [ ] Correlation IDs (OperationId/TraceId) present in all log entries
- [ ] Appropriate log level assignments (Error for failures, Warning for retries)

## Distributed Tracing

- [ ] W3C Trace Context propagation enabled (default with OpenTelemetry)
- [ ] Cross-boundary correlation for messaging (Service Bus, Event Grid)
- [ ] Custom spans for significant business operations
- [ ] Span attributes include business-relevant identifiers
- [ ] Sampling configured appropriately for environment

## Health Endpoints

- [ ] `/healthz/live` liveness probe — returns 200 if process is running
- [ ] `/healthz/ready` readiness probe — checks downstream dependencies
- [ ] Health check endpoints excluded from Application Insights sampling
- [ ] Container orchestrator probes configured (if applicable)

## Alerting

- [ ] Error rate alert configured (Sev 1)
- [ ] Latency threshold alert configured (Sev 2)
- [ ] Availability test for public endpoints
- [ ] Action groups defined with appropriate notification channels
- [ ] Alert suppression rules for planned maintenance windows

## Dashboards

- [ ] Azure Monitor Workbook with request/error/latency overview
- [ ] Dependency health section in workbook
- [ ] Exception trending section
- [ ] Business metrics section (if applicable)

## Infrastructure as Code

- [ ] Log Analytics workspace defined in Bicep/Terraform
- [ ] Application Insights resource linked to Log Analytics
- [ ] Diagnostic settings configured for all Azure resources
- [ ] Alert rules and action groups defined in IaC
- [ ] Availability tests defined in IaC

## Data Management

- [ ] Log retention period configured (30 days hot, 90 days warm)
- [ ] Daily cap configured to prevent cost overruns
- [ ] Sampling ratio tuned for production volume
- [ ] Archive policy for compliance-required data
