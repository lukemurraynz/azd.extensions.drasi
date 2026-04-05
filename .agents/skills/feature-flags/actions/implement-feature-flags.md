# Implement Feature Flags

## Steps

### Step 1 — Deploy App Configuration Store

Deploy with managed identity and `disableLocalAuth`:

```bicep
resource configStore 'Microsoft.AppConfiguration/configurationStores@2024-06-01' = {
  name: configStoreName
  location: location
  sku: { name: 'standard' }
  identity: { type: 'SystemAssigned' }
  properties: {
    disableLocalAuth: true
  }
}
```

Assign `App Configuration Data Reader` (516239f1-63e1-4d78-a4de-a74fb236a071) to the application's managed identity.

### Step 2 — Add NuGet Packages

```bash
dotnet add package Microsoft.Azure.AppConfiguration.AspNetCore
dotnet add package Microsoft.FeatureManagement.AspNetCore
```

### Step 3 — Configure the Application

```csharp
var builder = WebApplication.CreateBuilder(args);

builder.Configuration.AddAzureAppConfiguration(options =>
{
    options.Connect(
            new Uri(builder.Configuration["AppConfiguration:Endpoint"]),
            new DefaultAzureCredential())
        .UseFeatureFlags(featureOptions =>
        {
            featureOptions.CacheExpirationInterval = TimeSpan.FromSeconds(30);
        });
});

builder.Services.AddAzureAppConfiguration();
builder.Services.AddFeatureManagement()
    .AddFeatureFilter<TargetingFilter>();
builder.Services.AddSingleton<ITargetingContextAccessor, HttpContextTargetingContextAccessor>();

var app = builder.Build();
app.UseAzureAppConfiguration();
```

### Step 4 — Create Feature Flags in App Configuration

Use Azure CLI or the portal:

```bash
# Simple on/off flag
az appconfig feature set \
  --connection-string "$APPCONFIG_CONNECTION_STRING" \
  --feature NewCheckoutFlow \
  --description "New checkout flow for orders"

# Enable with targeting filter
az appconfig feature filter add \
  --connection-string "$APPCONFIG_CONNECTION_STRING" \
  --feature NewCheckoutFlow \
  --filter-name Microsoft.Targeting \
  --filter-parameters Audience.DefaultRolloutPercentage=10
```

### Step 5 — Use Feature Flags in Code

```csharp
public class OrderService
{
    private readonly IFeatureManager _featureManager;
    private readonly ILogger<OrderService> _logger;

    public OrderService(IFeatureManager featureManager, ILogger<OrderService> logger)
    {
        _featureManager = featureManager;
        _logger = logger;
    }

    public async Task<OrderResult> ProcessOrderAsync(Order order)
    {
        var useNewFlow = await _featureManager.IsEnabledAsync("NewCheckoutFlow");
        _logger.LogInformation("Processing order {OrderId}, NewCheckoutFlow={Enabled}",
            order.Id, useNewFlow);

        return useNewFlow
            ? await ProcessNewFlowAsync(order)
            : await ProcessLegacyFlowAsync(order);
    }
}
```

### Step 6 — Verify

- Feature flag appears in App Configuration store
- Application reads flag value dynamically (change in portal reflects within cache interval)
- Targeting filter applies correctly for test users
- Flag evaluation logged in Application Insights
