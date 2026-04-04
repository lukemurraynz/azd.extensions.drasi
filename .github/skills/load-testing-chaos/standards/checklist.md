# Load Testing & Chaos Engineering Checklist

## Load Testing Infrastructure

- [ ] Azure Load Testing resource deployed
- [ ] Managed identity configured with required roles
- [ ] JMeter test scripts created for key scenarios
- [ ] Test configuration YAML defined with failure criteria
- [ ] VNet integration configured (if testing private endpoints)

## Performance Baselines

- [ ] Baseline established for P50, P95, P99 response times
- [ ] Baseline established for throughput (RPS)
- [ ] Baseline established for error rate
- [ ] Baseline established for resource utilisation (CPU, memory, RU)
- [ ] Baselines documented and version-controlled

## Test Scenarios

- [ ] Smoke test defined (minimal load, basic verification)
- [ ] Load test defined (expected production traffic)
- [ ] Stress test defined (beyond expected capacity)
- [ ] Test data representative of production scale
- [ ] User scenarios model actual usage patterns

## CI/CD Integration

- [ ] Smoke test runs on every deployment
- [ ] Load test runs before major releases
- [ ] Failure criteria gate deployments automatically
- [ ] Test results stored for trend analysis

## Chaos Studio Setup

- [ ] Chaos targets enabled on target resources
- [ ] Capabilities registered for required fault types
- [ ] Experiment managed identity has required permissions
- [ ] Experiments deployed via IaC (Bicep/Terraform)

## Chaos Experiment Design

- [ ] Steady state defined before each experiment
- [ ] Hypothesis documented for expected behaviour
- [ ] Abort conditions defined for safety
- [ ] Rollback plan documented
- [ ] Monitoring dashboards active during experiments

## Chaos Maturity

- [ ] Level 1: Single-resource faults in dev/test environment
- [ ] Level 2: Multi-resource scenarios in staging monthly
- [ ] Level 3: Full domain simulation in pre-prod weekly
- [ ] Level 4: Production chaos with automated game days

## Observability During Testing

- [ ] Application Insights monitoring active during tests
- [ ] Alerts configured for test abort conditions
- [ ] Dashboards showing real-time metrics
- [ ] Test results correlated with telemetry data
