# Performance Testing

## Test Types

| Test Type   | Purpose                                       | Duration  | Load Pattern         |
| ----------- | --------------------------------------------- | --------- | -------------------- |
| Smoke Test  | Verify basic functionality under minimal load | 1-2 min   | 1-5 users            |
| Load Test   | Validate expected production traffic          | 5-15 min  | Ramp to expected RPS |
| Stress Test | Find breaking points                          | 10-30 min | Ramp beyond capacity |
| Soak Test   | Detect memory leaks, resource exhaustion      | 1-4 hours | Steady expected load |
| Spike Test  | Validate auto-scaling and recovery            | 5-10 min  | Sudden burst         |

---

## Performance Baselines

### Establishing Baselines

1. Run a load test at expected production traffic
2. Record P50, P95, P99 response times
3. Record throughput (requests per second)
4. Record error rate
5. Record resource utilisation (CPU, memory, RU)

### Baseline Template

```yaml
# performance-baseline.yaml
service: order-api
date: 2024-01-15
environment: staging
baselines:
  response_time:
    p50_ms: 120
    p95_ms: 350
    p99_ms: 800
  throughput:
    rps: 500
  error_rate_percent: 0.02
  resources:
    cpu_avg_percent: 45
    memory_avg_percent: 60
    cosmos_ru_avg: 2500
```

### Failure Criteria

Define thresholds that fail the test:

```yaml
failureCriteria:
  - avg(response_time_ms) > 500 # Average response too slow
  - percentage(error) > 5 # Too many errors
  - p99(response_time_ms) > 2000 # Tail latency too high
  - avg(requests_per_sec) < 100 # Throughput too low
```

---

## Test Design

### Realistic Scenarios

Model tests on actual user behaviour:

```
Scenario: E-commerce checkout flow
  - 60% Browse products (GET /api/products)
  - 25% View product detail (GET /api/products/{id})
  - 10% Add to cart (POST /api/cart)
  - 4% Checkout (POST /api/orders)
  - 1% Payment callback (POST /api/payments/callback)
```

### Data Considerations

| Aspect                  | Recommendation                                     |
| ----------------------- | -------------------------------------------------- |
| Test data               | Use production-scale data sets                     |
| User identity           | Rotate unique test users to avoid caching bias     |
| Payload variation       | Vary request payloads to test different code paths |
| Geographic distribution | Use multiple Azure Load Testing engine regions     |

---

## CI/CD Integration

### GitHub Actions

```yaml
- name: Run Load Test
  uses: azure/load-testing@v1
  with:
    loadTestConfigFile: "tests/load/config.yaml"
    loadTestResource: "my-load-test-resource"
    resourceGroup: "my-resource-group"
```

### Azure DevOps

```yaml
- task: AzureLoadTest@1
  inputs:
    azureSubscription: "my-subscription"
    loadTestConfigFile: "tests/load/config.yaml"
    loadTestResource: "my-load-test-resource"
    resourceGroup: "my-resource-group"
```

---

## Rules

1. Run smoke tests on every deployment — catch regressions early.
2. Run full load tests before major releases.
3. Use failure criteria to gate deployments automatically.
4. Test with production-representative data and scale.
5. Track baselines over time — performance should not regress.
