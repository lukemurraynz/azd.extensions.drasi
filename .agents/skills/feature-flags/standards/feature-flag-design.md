# Feature Flag Design

## Naming Conventions

| Pattern                  | Example                  | Use For                  |
| ------------------------ | ------------------------ | ------------------------ |
| `PascalCase`             | `NewCheckoutFlow`        | Simple on/off flags      |
| `Domain.Feature`         | `Orders.ExpressCheckout` | Domain-scoped flags      |
| `Domain.Feature.Variant` | `UI.Dashboard.LayoutV2`  | Variant/experiment flags |

**Rules:**

- No spaces or special characters
- Descriptive — the name should explain what the flag controls
- Prefix with domain for large applications

---

## Feature Flag Lifecycle

| Phase            | State     | Configuration                     |
| ---------------- | --------- | --------------------------------- |
| Development      | Disabled  | Flag exists, disabled by default  |
| Internal Testing | Targeted  | Enabled for internal users/groups |
| Canary           | Partial   | 5-10% of production users         |
| Gradual Rollout  | Expanding | 25% → 50% → 75% → 100%            |
| Full Rollout     | Enabled   | 100% — flag is always on          |
| Cleanup          | Removed   | Code path removed, flag deleted   |

**Critical:** Every flag must have a cleanup date. Permanent toggles create technical debt.

---

## Flag Types

| Type        | Purpose                            | Example                  |
| ----------- | ---------------------------------- | ------------------------ |
| Release     | Gate a feature for gradual rollout | `NewPaymentProvider`     |
| Experiment  | A/B test with variants             | `CheckoutLayout`         |
| Operational | Circuit breaker, kill switch       | `EnableExternalApiCalls` |
| Permission  | User/group-based access            | `PremiumDashboard`       |

---

## Code Patterns

### Do: Clean conditional logic

```csharp
if (await _featureManager.IsEnabledAsync("NewCheckoutFlow"))
{
    return await _newCheckoutService.ProcessAsync(order);
}
return await _legacyCheckoutService.ProcessAsync(order);
```

### Don't: Nested flag checks

```csharp
// Anti-pattern — deeply nested flags are hard to reason about
if (await _featureManager.IsEnabledAsync("FeatureA"))
{
    if (await _featureManager.IsEnabledAsync("FeatureB"))
    {
        if (await _featureManager.IsEnabledAsync("FeatureC"))
        {
            // What combination is this?
        }
    }
}
```

### Do: Interface-based branching for complex features

```csharp
// Register different implementations based on feature flag
builder.Services.AddScoped<ICheckoutService>(sp =>
{
    var featureManager = sp.GetRequiredService<IFeatureManager>();
    return featureManager.IsEnabledAsync("NewCheckoutFlow").GetAwaiter().GetResult()
        ? sp.GetRequiredService<NewCheckoutService>()
        : sp.GetRequiredService<LegacyCheckoutService>();
});
```

---

## Rules

1. Every flag has a defined lifecycle — create with a cleanup target date.
2. Limit flag nesting — maximum one level of flag dependency.
3. Use variants for A/B testing instead of multiple boolean flags.
4. Log flag evaluations for observability and debugging.
5. Review and remove fully-rolled-out flags within 30 days.
