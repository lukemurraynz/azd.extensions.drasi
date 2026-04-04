---
name: feature-flags
description: >-
  Azure App Configuration feature flags with targeting filters, percentage rollout strategies, and .NET feature management SDK integration.
  USE FOR: implementing feature flags, configuring targeting filters, setting up percentage-based rollouts, integrating .NET feature management, or managing feature lifecycles in Azure App Configuration.
---

# Feature Flags

> **Mandatory:** Use Azure App Configuration for centralised feature flag management. Never hard-code feature flags or use local configuration files in production.

## Description

Patterns for implementing feature flags using Azure App Configuration — gradual rollout, targeting filters, A/B testing, and integration with .NET feature management.

## Capabilities

| Capability            | Details                                            |
| --------------------- | -------------------------------------------------- |
| Feature Management    | Centralised toggle management in App Configuration |
| Targeting Filters     | User, group, and percentage-based rollout          |
| Gradual Rollout       | Percentage-based progressive exposure              |
| Variants              | Multiple values per feature for A/B testing        |
| Configuration Refresh | Dynamic refresh without redeployment               |
| Infrastructure        | Bicep deployment with managed identity             |

## Standards

| Standard                                                | Purpose              |
| ------------------------------------------------------- | -------------------- |
| [Feature Flag Design](standards/feature-flag-design.md) | Naming and lifecycle |
| [Checklist](standards/checklist.md)                     | Validation checklist |

## Actions

| Action                                                        | Purpose          |
| ------------------------------------------------------------- | ---------------- |
| [Implement Feature Flags](actions/implement-feature-flags.md) | End-to-end setup |

---

## App Configuration Setup

### Bicep — App Configuration Store

> **Warning — SKU requirement:** The App Configuration **Free** tier does not support
> feature flags or private endpoints. You must use `sku: { name: 'standard' }` for
> feature flag functionality. The Free tier is limited to configuration key-values only.

```bicep
param location string = resourceGroup().location
param configStoreName string
param principalId string

resource configStore 'Microsoft.AppConfiguration/configurationStores@2024-06-01' = {
  name: configStoreName
  location: location
  sku: { name: 'standard' }
  identity: {
    type: 'SystemAssigned'
  }
  properties: {
    disableLocalAuth: true
    enablePurgeProtection: false
    softDeleteRetentionInDays: 7
  }
}

// App Configuration Data Reader role
resource configReaderRole 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(configStore.id, principalId, '516239f1-63e1-4d78-a4de-a74fb236a071')
  scope: configStore
  properties: {
    principalId: principalId
    roleDefinitionId: subscriptionResourceId(
      'Microsoft.Authorization/roleDefinitions',
      '516239f1-63e1-4d78-a4de-a74fb236a071' // App Configuration Data Reader
    )
    principalType: 'ServicePrincipal'
  }
}

output skuValidation string = configStore.properties.sku.name == 'Free' ? 'ERROR: Feature flags require Standard SKU' : 'OK'
```

---

## .NET Integration

### Program.cs Configuration

```csharp
var builder = WebApplication.CreateBuilder(args);

var endpoint = builder.Configuration["AppConfiguration:Endpoint"];

builder.Configuration.AddAzureAppConfiguration(options =>
{
    options.Connect(new Uri(endpoint), new DefaultAzureCredential())
        .UseFeatureFlags(featureOptions =>
        {
            featureOptions.CacheExpirationInterval = TimeSpan.FromSeconds(30);
        })
        .ConfigureRefresh(refresh =>
        {
            refresh.Register("Sentinel", refreshAll: true)
                .SetCacheExpiration(TimeSpan.FromSeconds(30));
        });
});

> **Sentinel key pattern:** Register a sentinel key (e.g., `"Sentinel"`) with `refreshAll: true`.
> When you want to push configuration changes to all app instances, update the sentinel key's
> value (e.g., set it to the current timestamp). This triggers a refresh of **all** configuration
> values and feature flags, without restarting the application. Without a sentinel key, each
> key is polled individually, which is slower and uses more requests.

> **Cache expiration tradeoff:** The `CacheExpirationInterval` (default 30 seconds) controls
> how often the SDK polls App Configuration. Lower values = faster flag propagation but more
> requests (and cost on Standard tier, which charges per request above the daily free quota).
> For most applications, 30 seconds is a good default. For latency-critical rollouts, use 5–10
> seconds.

builder.Services.AddAzureAppConfiguration();
builder.Services.AddFeatureManagement()
    .AddFeatureFilter<TargetingFilter>();

var app = builder.Build();
app.UseAzureAppConfiguration();
```

### Using Feature Flags in Code

```csharp
public class OrderController : ControllerBase
{
    private readonly IFeatureManager _featureManager;

    public OrderController(IFeatureManager featureManager)
    {
        _featureManager = featureManager;
    }

    [HttpPost]
    public async Task<IActionResult> CreateOrder(OrderRequest request)
    {
        if (await _featureManager.IsEnabledAsync("NewCheckoutFlow"))
        {
            return await ProcessWithNewFlow(request);
        }

        return await ProcessWithLegacyFlow(request);
    }
}
```

> [!IMPORTANT]
> Feature flag evaluation depends on App Configuration availability. Wrap flag checks with a timeout (500ms recommended) and a fail-closed fallback (flag disabled). If App Configuration is unreachable, the application should continue with default behavior rather than blocking requests.
> ```csharp
> // Fail-closed pattern: if flag evaluation times out, treat as disabled
> var cts = new CancellationTokenSource(TimeSpan.FromMilliseconds(500));
> var enabled = false;
> try { enabled = await featureManager.IsEnabledAsync("MyFeature", cts.Token); }
> catch (OperationCanceledException) { logger.LogWarning("Feature flag timeout — defaulting to disabled"); }
> ```

### Using Feature Flags in Razor

```html
<feature name="NewDashboard">
  <div>New dashboard experience</div>
</feature>
<feature name="NewDashboard" negate="true">
  <div>Classic dashboard</div>
</feature>
```

---

## Targeting Filters

### Percentage Rollout

Roll out a feature to a percentage of users:

```json
{
  "id": "NewCheckoutFlow",
  "enabled": true,
  "conditions": {
    "client_filters": [
      {
        "name": "Microsoft.Targeting",
        "parameters": {
          "Audience": {
            "DefaultRolloutPercentage": 25
          }
        }
      }
    ]
  }
}
```

### User and Group Targeting

Target specific users or groups:

```json
{
  "id": "BetaFeature",
  "enabled": true,
  "conditions": {
    "client_filters": [
      {
        "name": "Microsoft.Targeting",
        "parameters": {
          "Audience": {
            "Users": ["user@example.com"],
            "Groups": [
              { "Name": "BetaTesters", "RolloutPercentage": 100 },
              { "Name": "InternalUsers", "RolloutPercentage": 50 }
            ],
            "DefaultRolloutPercentage": 0
          }
        }
      }
    ]
  }
}
```

### Implementing the Targeting Context

```csharp
public class HttpContextTargetingContextAccessor : ITargetingContextAccessor
{
    private readonly IHttpContextAccessor _httpContextAccessor;

    public HttpContextTargetingContextAccessor(IHttpContextAccessor httpContextAccessor)
    {
        _httpContextAccessor = httpContextAccessor;
    }

    public ValueTask<TargetingContext> GetContextAsync()
    {
        var httpContext = _httpContextAccessor.HttpContext;
        var user = httpContext?.User;

        return new ValueTask<TargetingContext>(new TargetingContext
        {
            UserId = user?.Identity?.Name ?? "anonymous",
            Groups = user?.Claims
                .Where(c => c.Type == "groups")
                .Select(c => c.Value)
                .ToList() ?? new List<string>()
        });
    }
}
```

Register: `builder.Services.AddSingleton<ITargetingContextAccessor, HttpContextTargetingContextAccessor>();`

---

## Variants (A/B Testing)

Define multiple values for a feature:

```json
{
  "id": "CheckoutLayout",
  "enabled": true,
  "variants": [
    { "name": "LayoutA", "configuration_value": "single-page" },
    { "name": "LayoutB", "configuration_value": "multi-step" }
  ],
  "allocation": {
    "percentile": [
      { "variant": "LayoutA", "from": 0, "to": 50 },
      { "variant": "LayoutB", "from": 50, "to": 100 }
    ],
    "default_when_enabled": "LayoutA"
  }
}
```

```csharp
var variant = await _featureManager.GetVariantAsync("CheckoutLayout");
var layout = variant?.Configuration?.Get<string>(); // "single-page" or "multi-step"
```

---

## Principles

1. **Centralise flag management** — use App Configuration, not local config files.
2. **Use managed identity** — `disableLocalAuth: true` on the config store.
3. **Set cache expiration** — balance freshness with performance (30 seconds default).
4. **Clean up flags** — remove flags after full rollout to avoid permanent toggles.
5. **Use targeting for safe rollout** — start with internal users, expand gradually.
6. **Track flag usage** — log which flags are evaluated for observability.

## Currency and Verification

- **Date checked:** 2026-03-31
- **Compatibility:** .NET 10, Azure App Configuration, `Microsoft.FeatureManagement` SDK
- **Sources:** [Azure App Configuration feature flags](https://learn.microsoft.com/azure/azure-app-configuration/concept-feature-management), [.NET feature management](https://learn.microsoft.com/azure/azure-app-configuration/use-feature-flags-dotnet-core)
- **Verification steps:**
  1. Check feature management SDK: `dotnet list package | grep Microsoft.FeatureManagement`
  2. Verify App Config API version: `az provider show --namespace Microsoft.AppConfiguration --query "resourceTypes[?resourceType=='configurationStores'].apiVersions" -o tsv`

### Known Pitfalls

| Area                        | Pitfall                                                                            | Mitigation                                                                                             |
| --------------------------- | ---------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------ |
| Cache expiration            | Default 30-second cache means flag changes take up to 30 seconds to propagate      | Acceptable for most scenarios; reduce for real-time kill switches, increase for stability              |
| Permanent toggles           | Feature flags left in production code after full rollout become tech debt          | Schedule flag cleanup; audit flags quarterly; remove after stable full rollout                         |
| Targeting filter complexity | Complex targeting rules (percentage + user lists) are hard to debug in production  | Log flag evaluation results with targeting context; use App Config portal to inspect active rules      |
| `disableLocalAuth` impact   | Disabling local auth on App Config store blocks Azure Portal data plane operations | Use managed identity or Entra ID for all access; enable local auth temporarily for emergency debugging |
| Variant configuration drift | Feature flag variants defined in code may differ from App Config store definitions | Use App Config as single source of truth; don't define variant values in `appsettings.json`            |

## References

- [Azure App Configuration feature flags](https://learn.microsoft.com/en-us/azure/azure-app-configuration/concept-feature-management)
- [.NET feature management](https://learn.microsoft.com/en-us/azure/azure-app-configuration/use-feature-flags-dotnet-core)
- [Targeting filter](https://learn.microsoft.com/en-us/azure/azure-app-configuration/howto-targetingfilter-aspnet-core)
- [Feature flag variants](https://learn.microsoft.com/en-us/azure/azure-app-configuration/howto-feature-filters-aspnet-core)

## Related Skills

- [Identity & Managed Identity](../identity-managed-identity/SKILL.md) — RBAC for App Configuration
- [Observability & Monitoring](../observability-monitoring/SKILL.md) — Flag evaluation tracking
