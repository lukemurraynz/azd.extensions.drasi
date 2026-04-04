# Scaling & Performance Standard

## Scaling Rule Types

### HTTP Scaling

Best for web APIs and frontends. Scales based on concurrent HTTP requests per replica.

```bicep
rules: [
  {
    name: 'http-scaling'
    http: {
      metadata: {
        concurrentRequests: '50'
      }
    }
  }
]
```

**Guideline:** Start with `concurrentRequests: 50`, adjust based on load testing.

### Queue-Based Scaling

Best for background processors. Scales based on queue depth.

```bicep
rules: [
  {
    name: 'queue-scaling'
    azureQueue: {
      queueName: 'work-items'
      queueLength: 10
      auth: [
        { secretRef: 'queue-connection', triggerParameter: 'connection' }
      ]
    }
  }
]
```

**Guideline:** Set `queueLength` to the number of messages one replica can process
in 30 seconds.

### Service Bus Scaling

```bicep
rules: [
  {
    name: 'servicebus-scaling'
    custom: {
      type: 'azure-servicebus'
      metadata: {
        queueName: 'orders'
        messageCount: '5'
        namespace: serviceBusNamespace
      }
      auth: [
        { secretRef: 'sb-connection', triggerParameter: 'connection' }
      ]
    }
  }
]
```

---

## Scale-to-Zero

Enable for event-driven workloads to reduce costs:

```bicep
scale: {
  minReplicas: 0
  maxReplicas: 20
  rules: [ /* queue or event trigger */ ]
}
```

**Considerations:**

- First request after scale-to-zero incurs cold start (typically 5–15 seconds)
- Not suitable for latency-sensitive HTTP endpoints
- Ideal for queue processors, event handlers, scheduled jobs

---

## Performance Guidelines

| Metric                | Target          | Action if Exceeded             |
| --------------------- | --------------- | ------------------------------ |
| Request latency (p95) | < 500ms         | Increase replicas or CPU       |
| CPU utilisation       | < 70% sustained | Already well-scaled            |
| Memory utilisation    | < 80% sustained | Increase memory allocation     |
| Scale-out time        | < 30 seconds    | Pre-warm with `minReplicas: 1` |
| Cold start            | < 15 seconds    | Optimise container image size  |

### Container Image Optimisation

1. Use minimal base images (`alpine`, `distroless`, `chiseled`)
2. Multi-stage builds — separate build and runtime stages
3. Target image size < 200MB for fast pull times
4. Layer ordering — dependencies first, application code last

---

## Rules

1. Always define at least one scaling rule — never rely on manual scaling.
2. Set `maxReplicas` to a safe upper bound — prevent runaway scaling costs.
3. Use `minReplicas: 0` only for non-latency-sensitive workloads.
4. Test scaling behaviour under load before production deployment.
5. Monitor scaling events via Container Apps system logs.
