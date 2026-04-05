# Container Apps Checklist

## Environment

- [ ] Container App Environment deployed via IaC (Bicep/Terraform)
- [ ] Log Analytics workspace linked for logging
- [ ] Workload profiles configured (Consumption or Dedicated)
- [ ] VNet integration configured (if private networking required)

## Container App

- [ ] Managed Identity enabled (`SystemAssigned` or `UserAssigned`)
- [ ] ACR image pull via managed identity (not admin credentials)
- [ ] Liveness probe configured
- [ ] Readiness probe configured
- [ ] Resource limits set (CPU and memory)
- [ ] Environment variables use secret references for sensitive values

## Ingress

- [ ] `external` set correctly (true for public, false for internal)
- [ ] `allowInsecure: false` (HTTPS only)
- [ ] Custom domain configured with managed certificate (if public)
- [ ] IP restrictions applied (if applicable)

## Scaling

- [ ] At least one scaling rule defined
- [ ] `minReplicas` and `maxReplicas` set appropriately
- [ ] Scale-to-zero only for non-latency-sensitive workloads
- [ ] Scaling tested under expected load

## Revisions & Deployments

- [ ] `activeRevisionsMode: 'Multiple'` for production (traffic splitting)
- [ ] Traffic splitting strategy defined (canary or blue/green)
- [ ] Rollback procedure documented

## Observability

- [ ] Application Insights connection string configured
- [ ] Structured logging enabled
- [ ] Health endpoints implemented (`/healthz/live`, `/healthz/ready`)
- [ ] Container Apps system logs reviewed for scaling events

## Jobs (if applicable)

- [ ] Correct trigger type selected (Manual, Schedule, or Event)
- [ ] `replicaTimeout` set to at least 2x expected p99 duration
- [ ] `replicaRetryLimit` configured (0 for non-idempotent; 1–3 for transient-safe)
- [ ] `parallelism` and `replicaCompletionCount` set appropriately
- [ ] Event-driven jobs have KEDA scale rules configured
- [ ] `maxExecutions` capped to prevent resource contention
- [ ] Cron expressions account for UTC timezone offset
- [ ] Container exits with code 0 on success, non-zero on failure
- [ ] Managed Identity enabled on the job resource
- [ ] ACR image pull via managed identity (not admin credentials)
- [ ] Execution history monitored for failures

## Security

- [ ] No secrets in environment variables (use Key Vault + managed identity)
- [ ] Container image scanned for vulnerabilities
- [ ] Base image is minimal (alpine/distroless/chiseled)
- [ ] Non-root user in Dockerfile
