# Run Load Test

## Steps

### Step 1 — Create Test Script

Write a JMeter test plan (`.jmx`) modelling your key user scenarios:

```xml
<ThreadGroup>
  <intProp name="ThreadGroup.num_threads">50</intProp>
  <intProp name="ThreadGroup.ramp_time">60</intProp>
  <intProp name="ThreadGroup.duration">300</intProp>
</ThreadGroup>
```

- `num_threads` — concurrent virtual users
- `ramp_time` — seconds to reach full concurrency
- `duration` — total test duration in seconds

### Step 2 — Define Test Configuration

Create a YAML configuration with failure criteria:

```yaml
version: v0.1
testId: api-load-test
testName: API Load Test
testPlan: test-plan.jmx
engineInstances: 1
failureCriteria:
  - avg(response_time_ms) > 500
  - percentage(error) > 5
  - p99(response_time_ms) > 2000
env:
  - name: target_host
    value: myapp.azurewebsites.net
```

### Step 3 — Run the Test

```bash
# Azure CLI
az load test create \
  --load-test-resource my-load-test \
  --test-id api-load-test \
  --test-plan test-plan.jmx \
  --load-test-config-file config.yaml

az load test-run create \
  --load-test-resource my-load-test \
  --test-id api-load-test \
  --test-run-id run-$(date +%Y%m%d-%H%M%S)
```

### Step 4 — Analyse Results

Review in Azure Portal or CLI:

- Response time distribution (P50, P95, P99)
- Throughput (requests per second)
- Error rate and error types
- Server-side metrics (CPU, memory, RU consumption)
- Failure criteria pass/fail status

### Step 5 — Update Baselines

If performance meets expectations, update the baseline document:

```yaml
baselines:
  response_time:
    p50_ms: 120
    p95_ms: 350
    p99_ms: 800
  throughput:
    rps: 500
  error_rate_percent: 0.02
```

Compare against previous baselines to detect regressions.
