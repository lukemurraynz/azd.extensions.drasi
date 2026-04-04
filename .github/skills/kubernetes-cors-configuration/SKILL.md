---
name: kubernetes-cors-configuration
description: >-
  Configure CORS correctly in Azure Kubernetes Service (AKS) environments using Azure-native services and best practices. Covers CORS implementation patterns for production AKS clusters using App Configuration, Workload Identity, and Kubernetes resource management. USE FOR: fix CORS errors in AKS, configure Access-Control-Allow-Origin, set up CORS with Azure App Configuration, debug preflight OPTIONS failures, implement CORS at ingress layer.
---

# AKS CORS Configuration Skill

**Purpose**: Configure CORS correctly in Azure Kubernetes Service (AKS) environments using Azure-native services and best practices from "The AKS Book." This skill covers CORS implementation patterns that work reliably in production AKS clusters using App Configuration, Workload Identity, and proper Kubernetes resource management.

**Apply this skill when:**

- Setting up cross-origin requests between frontend SPAs and backend APIs in AKS
- Integrating CORS configuration with Azure App Configuration and Key Vault
- Debugging CORS errors and preventing `No 'Access-Control-Allow-Origin' header` issues
- Implementing CORS at the ingress/gateway layer (Gateway API, Application Routing add-on, Application Gateway for Containers, Istio)
- Designing CORS for multi-deployment scenarios with rolling updates
- Implementing least-privilege identity patterns for CORS origins access

**AKS Prerequisites:**

- Application CORS middleware configured (ASP.NET Core with `AddCors`, Express with `cors`, etc.)
- Kubernetes manifests for deployments, services, and ingress/gateway resources
- Azure App Configuration or Key Vault for centralized configuration management
- Workload Identity enabled on AKS cluster (OIDC issuer configured)
- Understanding of CORS preflight requests (OPTIONS method) and allowed origins

---

## Problem: Why CORS Fails in Kubernetes

> **ASP.NET Core CORS middleware:** See `.NET Backend Patterns` skill (§5 CORS Configuration) for the C# middleware setup. This skill covers the Kubernetes deployment layer: ConfigMaps, envsubst, and multi-cluster CORS origin management.

1. **ConfigMap/Environment Variable Missing**: CORS origins not passed to the API container
2. **Placeholder Not Substituted**: Manifests contain `${CORS_ALLOWED_ORIGINS}` but the actual origin was never injected
3. **Wrong Origin Format**: Missing protocol (`http://` vs `https://`), port, or protocol-port mismatch
4. **Hardcoded Localhost Fallback**: API defaults to `localhost:3000` but frontend is at external DNS
5. **Pod Restart Doesn't Refresh Config**: Environment variables set post-deployment won't apply to running pods

---

## Prerequisites: Configure DNS Labels for AKS Services

> **AKS Book default (future projects):** Prefer a **single public entrypoint** (Gateway API / Ingress)
> and attach the DNS label to that gateway service. Use per-service `LoadBalancer` services only
> for non-HTTP protocols or transitional setups.

Before configuring CORS origins, ensure your frontend and API services have **stable DNS names** (not just IP addresses). AKS LoadBalancer services with DNS labels provide:

- ✅ **Stable hostname** (`frontend-service.australiaeast.cloudapp.azure.com`)
- ✅ **Survives IP redeployments** (IPs can change, DNS labels don't)
- ✅ **Clean CORS origin** (`http://frontend-service.australiaeast.cloudapp.azure.com` instead of IP)
- ✅ **Certificate-friendly** (Let's Encrypt can validate the domain)

### Preferred: DNS Label on Gateway/Ingress Service

If you run a Gateway API or ingress controller (recommended), place the DNS label on the gateway
service and use that host for CORS origins:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: gateway-public
  namespace: ingress-system
  annotations:
    # Single public entrypoint for HTTP(S)
    service.beta.kubernetes.io/azure-dns-label-name: "gateway-emergency-alerts-dev"
spec:
  type: LoadBalancer
  ports:
    - name: http
      port: 80
      targetPort: 8080
    - name: https
      port: 443
      targetPort: 8443
  selector:
    app: gateway
```

**Use this host for CORS origins**:

```powershell
$env:CORS_ALLOWED_ORIGINS = "https://gateway-emergency-alerts-dev.australiaeast.cloudapp.azure.com"
```

### If You Must Use Direct LoadBalancer Services (Exceptions)

**In your K8s service YAML:**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: emergency-alerts-frontend
  namespace: emergency-alerts
  labels:
    app: emergency-alerts-frontend
  annotations:
    # Request an Azure-provided DNS name
    # Format: <label>.<region>.cloudapp.azure.com
    service.beta.kubernetes.io/azure-dns-label-name: "frontend-emergency-alerts-dev"
spec:
  type: LoadBalancer
  ports:
    - name: http
      port: 80
      targetPort: 3000
      protocol: TCP
  selector:
    app: emergency-alerts-frontend
```

**For API service:**

```yaml
annotations:
  service.beta.kubernetes.io/azure-dns-label-name: "api-emergency-alerts-dev"
```

**For SignalR service (keep separate from API to avoid DNS collisions):**

```yaml
annotations:
  service.beta.kubernetes.io/azure-dns-label-name: "signalr-emergency-alerts-dev"
```

### Add DNS Label to Existing Service (kubectl)

If service already exists without DNS label:

```powershell
# Frontend
kubectl patch service emergency-alerts-frontend -n emergency-alerts \
  -p '{"metadata":{"annotations":{"service.beta.kubernetes.io/azure-dns-label-name":"frontend-emergency-alerts-dev"}}}'

# API
kubectl patch service emergency-alerts-api -n emergency-alerts \
  -p '{"metadata":{"annotations":{"service.beta.kubernetes.io/azure-dns-label-name":"api-emergency-alerts-dev"}}}'
```

### Verify DNS Label is Provisioned

**Step 1: Wait for DNS provisioning** (typically 30-60 seconds):

```powershell
# Check service has external IP
kubectl get svc emergency-alerts-frontend -n emergency-alerts -o wide
# Expected: EXTERNAL-IP shows public IP (e.g., 20.227.31.148)

# Wait for DNS to resolve
Start-Sleep -Seconds 40

# Verify DNS resolves to the public IP
nslookup frontend-emergency-alerts-dev.australiaeast.cloudapp.azure.com
# Expected: Name: frontend-emergency-alerts-dev.australiaeast.cloudapp.azure.com
#           Address: 20.227.31.148
```

**Step 2: Test connectivity via DNS**:

```powershell
# HTTP test
curl -s -o /dev/null -w "HTTP Status: %{http_code}\n" `
  http://frontend-emergency-alerts-dev.australiaeast.cloudapp.azure.com/

# Expected: HTTP Status: 200 (or 301/302 redirect, anything not 5xx)
```

**Step 3: Use DNS in CORS origins** (from now on):

```powershell
# Don't use IP addresses for CORS origins anymore
# ❌ WRONG: $env:CORS_ALLOWED_ORIGINS = "http://20.227.31.148"
# ✅ RIGHT:
$env:CORS_ALLOWED_ORIGINS = "http://frontend-emergency-alerts-dev.australiaeast.cloudapp.azure.com"
```

### Why Separate DNS Labels per Service?

| Service      | DNS Label                       | Why                                      |
| ------------ | ------------------------------- | ---------------------------------------- |
| **Frontend** | `frontend-emergency-alerts-dev` | SPA origin for CORS                      |
| **API**      | `api-emergency-alerts-dev`      | Consumed by frontend app                 |
| **SignalR**  | `signalr-emergency-alerts-dev`  | Separate from API; avoids port collision |

Using separate DNS labels prevents port conflicts and makes scaling/updates cleaner (one service can be updated without affecting others).

---

## Pattern 1: AKS-Native Pattern - Azure App Configuration + Workload Identity

**Recommended for AKS production clusters.** CORS origins are stored in Azure App Configuration (centralized, audited, multi-cluster), retrieved via Workload Identity (zero secrets in Kubernetes), with optional dynamic refresh without pod restarts. Aligns with "The AKS Book" best practices: Chapter 3 (Identity), Chapter 6 (Production Deployment), Chapter 10 (Traffic Management).

### 1.1 Prerequisites: Enable Workload Identity on Cluster

Verify OIDC issuer is enabled (if not, cluster setup must enable it once):

```powershell
$aksClusterName = "emergency-alerts-dev-aks"
$resourceGroup = "emergency-alerts-dev-rg"

# Check if OIDC issuer is already enabled
$oidcIssuer = az aks show -g $resourceGroup -n $aksClusterName --query "oidcIssuerProfile.issuerUrl" -o tsv

if ($oidcIssuer) {
    Write-Host "✅ OIDC issuer already enabled: $oidcIssuer"
} else {
    Write-Host "⚠️ OIDC issuer not enabled. Enabling now (non-disruptive, ~60 seconds)..."
    az aks update -g $resourceGroup -n $aksClusterName --enable-oidc-issuer
    Write-Host "✅ OIDC issuer enabled"
}
```

### 1.2 Create Azure Managed Identity and Federated Credential

Create a managed identity for the API service account and link it via OIDC federation:

```powershell
$resourceGroup = "emergency-alerts-dev-rg"
$aksClusterName = "emergency-alerts-dev-aks"
$identityName = "emergency-alerts-api-identity"
$namespace = "emergency-alerts"
$serviceAccountName = "emergency-alerts-api"

# Create managed identity
az identity create -g $resourceGroup -n $identityName

$clientId = az identity show -g $resourceGroup -n $identityName --query "clientId" -o tsv
$principalId = az identity show -g $resourceGroup -n $identityName --query "principalId" -o tsv

# Get OIDC issuer URL
$oidcIssuer = az aks show -g $resourceGroup -n $aksClusterName --query "oidcIssuerProfile.issuerUrl" -o tsv

# Create federated credential linking K8s service account to managed identity
az identity federated-credential create \
  -g $resourceGroup \
  --identity-name $identityName \
  --name "${namespace}-${serviceAccountName}-fed-cred" \
  --issuer $oidcIssuer \
  --subject "system:serviceaccount:${namespace}:${serviceAccountName}" \
  --audiences "api://AzureADTokenExchange"

Write-Host "✅ Managed Identity: $identityName"
Write-Host "   Client ID: $clientId"
Write-Host "   Service Account: ${namespace}/${serviceAccountName}"
Write-Host "   OIDC Issuer: $oidcIssuer"
```

### 1.3 Store CORS Origins in Azure App Configuration

Create or update the App Configuration instance with CORS origins (centralized, audited, shared across AKS clusters):

```powershell
$appConfigName = "emergencyalerts2-prod-config"
$resourceGroup = "emergency-alerts-dev-rg"

# CORS origins for different environments
$corsOriginsDev = "http://frontend-emergency-alerts-dev.australiaeast.cloudapp.azure.com:3000,http://localhost:3000"
$corsOriginsProd = "https://frontend.emergencyalerts.com"  # Production HTTPS

# Store in App Configuration
az appconfig kv set \
  -n $appConfigName \
  --key "Cors:AllowedOrigins" \
  --value $corsOriginsDev \
  --label "development" \
  --yes

az appconfig kv set \
  -n $appConfigName \
  --key "Cors:AllowedOrigins" \
  --value $corsOriginsProd \
  --label "production" \
  --yes

# Verify
Write-Host "✅ CORS origins stored in App Configuration:"
az appconfig kv show -n $appConfigName --key "Cors:AllowedOrigins" --label "development" --query "value" -o tsv
```

**Why App Configuration over Kubernetes ConfigMaps (per AKS Book Ch. 6):**

- ✅ **Centralized**: Single source of truth across multiple AKS clusters and environments
- ✅ **Audited**: Azure Activity Log tracks every CORS origin change
- ✅ **No Kubernetes Secrets**: Workload Identity eliminates credential management
- ✅ **Dynamic Refresh**: Application can reload CORS origins every 5 minutes without pod restart
- ✅ **Multi-deployment**: Same configuration used across canary, staging, production
- ✅ **Least Privilege**: Managed identity has only `App Configuration Data Reader` role (RBAC via Azure, not Kubernetes)

### 1.4 Grant App Configuration Data Reader Role to Managed Identity

```powershell
$appConfigResourceId = az appconfig show -n $appConfigName -g $resourceGroup --query "id" -o tsv
$principalId = az identity show -g $resourceGroup -n $identityName --query "principalId" -o tsv

# Assign role (built-in role ID: 516239f1-63e1-4108-9a7a-da0e6b6b6b49)
az role assignment create \
  --role "App Configuration Data Reader" \
  --assignee-object-id $principalId \
  --scope $appConfigResourceId

Write-Host "✅ Granted 'App Configuration Data Reader' role to managed identity"
```

### 1.5 Create Kubernetes Service Account with Workload Identity Annotation

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: emergency-alerts-api
  namespace: emergency-alerts
  annotations:
    # Link this K8s service account to the Azure managed identity
    # Workload Identity uses OIDC federation to get tokens (no credentials stored)
    azure.workload.identity/client-id: <MANAGED_IDENTITY_CLIENT_ID>
```

Apply it:

```bash
kubectl apply -f serviceaccount.yaml
```

### 1.6 Update Deployment with Workload Identity

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: emergency-alerts-api
  namespace: emergency-alerts
  labels:
    app: emergency-alerts-api
spec:
  replicas: 2
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 0 # Zero downtime during CORS config updates
      maxSurge: 1 # One extra pod during rolling update
  selector:
    matchLabels:
      app: emergency-alerts-api
  template:
    metadata:
      labels:
        app: emergency-alerts-api
        # Enable Workload Identity for this pod
        azure.workload.identity/use: "true"
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "5000"
    spec:
      serviceAccountName: emergency-alerts-api # ✓ SA with Workload Identity annotation
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
      containers:
        - name: api
          image: ${ACR_NAME}.azurecr.io/api:${IMAGE_TAG} # Pipeline-injected; never hardcode ACR hostnames
          imagePullPolicy: Always
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop:
                - ALL
          ports:
            - name: http
              containerPort: 8080
          env:
            - name: ASPNETCORE_ENVIRONMENT
              value: "Production"
            # App Configuration endpoint - no credentials in this env var
            # Workload Identity + DefaultAzureCredential handles authentication
            - name: AppConfig__Endpoint
              value: "https://emergencyalerts2-prod-config.azconfig.io"
            # App Configuration label for this cluster
            - name: AppConfig__Label
              value: "development"
          livenessProbe:
            httpGet:
              path: /health/live
              port: http
            initialDelaySeconds: 10
            periodSeconds: 30
            timeoutSeconds: 5
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /health/ready
              port: http
            initialDelaySeconds: 5
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 3
          resources:
            requests:
              cpu: 200m # 95th percentile of actual usage
              memory: 512Mi
            limits:
              # CPU limits can trigger throttling; omit unless you have strong reason
              # (per AKS Book Ch. 6: "Don't set CPU limits; set memory limits to 1.5-2x requests")
              memory: 1Gi # Upper bound for OOMKill
          volumeMounts:
            - name: tmp
              mountPath: /tmp
      volumes:
        - name: tmp
          emptyDir: {}
```

Apply it:

```bash
kubectl apply -f deployment.yaml
kubectl rollout status deployment/emergency-alerts-api -n emergency-alerts
```

### 1.7 Application Code Reads CORS from App Configuration

**ASP.NET Core Example:**

```csharp
// In Program.cs
using Azure.Identity;
using Microsoft.Extensions.Configuration.AzureAppConfiguration;

var builder = WebApplication.CreateBuilder(args);

// Connect to App Configuration using Workload Identity + DefaultAzureCredential
var appConfigEndpoint = builder.Configuration["AppConfig__Endpoint"];
if (!string.IsNullOrEmpty(appConfigEndpoint))
{
    // DefaultAzureCredential automatically uses:
    // 1. Environment variables (if set)
    // 2. Managed Identity (IMDS, from pod spec)
    // 3. Workload Identity (OIDC federation, from pod annotation + federated credential)
    // 4. Azure CLI credentials (local development)

    builder.Configuration.AddAzureAppConfiguration(options =>
    {
        options.Connect(new Uri(appConfigEndpoint), new DefaultAzureCredential())
            .Select("*")  // Load all keys
            .Select("*", builder.Configuration["AppConfig__Label"])  // Label-specific overrides
            .ConfigureRefresh(refresh =>
            {
                // Optional: Dynamic refresh every 5 minutes
                // Reloads CORS origins without pod restart
                refresh.Register("Cors:AllowedOrigins", refreshAll: false)
                    .SetRefreshInterval(TimeSpan.FromMinutes(5));
            });
    });
}

// Add App Configuration services for dynamic refresh (if enabled)
builder.Services.AddAzureAppConfiguration();

// Configure CORS from App Configuration
var corsOrigins = builder.Configuration["Cors:AllowedOrigins"] ??
    "http://localhost:3000,http://localhost:5173";

builder.Services.AddCors(options =>
{
    options.AddDefaultPolicy(policy =>
    {
        policy
            .WithOrigins(corsOrigins.Split(',', StringSplitOptions.TrimEntries))
            .AllowAnyHeader()
            .AllowAnyMethod()
            .AllowCredentials();  // Required for SignalR
    });
});

var app = builder.Build();

// Enable dynamic refresh middleware (optional)
if (builder.Configuration is IConfigurationRefresh refreshConfig)
{
    app.Use(async (context, next) =>
    {
        await refreshConfig.TryRefreshAsync();
        await next();
    });
}

app.UseCors();  // ✓ MUST come before auth middleware
```

**Node.js/Express Example:**

```javascript
const { DefaultAzureCredential } = require("@azure/identity");
const { AppConfigurationClient } = require("@azure/app-configuration");
const cors = require("cors");
const express = require("express");

const app = express();
const appConfigEndpoint = process.env.AppConfig__Endpoint;
const appConfigLabel = process.env.AppConfig__Label || "development";

let allowedOrigins = "http://localhost:3000";

// Load CORS origins from App Configuration using Workload Identity
async function initializeCorsOrigins() {
  if (appConfigEndpoint) {
    try {
      const client = new AppConfigurationClient(
        appConfigEndpoint,
        new DefaultAzureCredential(), // Workload Identity
      );
      const setting = await client.getConfigurationSetting({
        key: "Cors:AllowedOrigins",
        label: appConfigLabel,
      });
      allowedOrigins = setting.value;
      console.log(
        `✅ Loaded CORS origins from App Configuration: ${allowedOrigins}`,
      );

      // Optional: Refresh CORS every 5 minutes
      setInterval(
        async () => {
          try {
            const refreshedSetting = await client.getConfigurationSetting({
              key: "Cors:AllowedOrigins",
              label: appConfigLabel,
            });
            if (refreshedSetting.value !== allowedOrigins) {
              allowedOrigins = refreshedSetting.value;
              console.log(`✅ Refreshed CORS origins: ${allowedOrigins}`);
            }
          } catch (error) {
            // WARNING: Do not silently continue after configuration load failure.
            // A failed CORS config reload should: (1) keep the last-known-good config,
            // (2) log with correlation ID, (3) alert via health check degradation.
            console.warn(`⚠️ Failed to refresh CORS origins: ${error.message}`);
          }
        },
        5 * 60 * 1000,
      ); // Refresh every 5 minutes
    } catch (error) {
      console.warn(
        `⚠️ Failed to load CORS origins from App Configuration: ${error.message}`,
      );
    }
  }
}

// Initialize on startup
initializeCorsOrigins();

// Configure CORS middleware
app.use(
  cors({
    origin: allowedOrigins.split(",").map((o) => o.trim()),
    credentials: true,
    methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
    allowedHeaders: ["Content-Type", "Authorization"],
  }),
);
```

### 1.8 Deploy and Verify Workload Identity

```powershell
# Apply deployment
kubectl apply -f deployment.yaml -n emergency-alerts

# Wait for pods to be ready
kubectl rollout status deployment/emergency-alerts-api -n emergency-alerts --timeout=300s

# Verify pods are running with Workload Identity
$podName = kubectl get pods -l app=emergency-alerts-api -n emergency-alerts -o jsonpath='{.items[0].metadata.name}'

# Check that Workload Identity token can be obtained (pod should have ./var/run/secrets/azure/tokens/token)
kubectl exec -it $podName -n emergency-alerts -- ls -la /var/run/secrets/azure/tokens/ 2>/dev/null || echo "Token file not found - Workload Identity may not be active"

# Test pod can access App Configuration (should return health check with CORS origins loaded)
kubectl exec -it $podName -n emergency-alerts -- bash -c "curl -s http://localhost:8080/health/ready | jq ."

# Test CORS preflight request
$apiServiceIp = kubectl get svc emergency-alerts-api -n emergency-alerts -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
curl -i -X OPTIONS "http://${apiServiceIp}:8080/api/v1/alerts" `
  -H "Origin: http://frontend-emergency-alerts-dev.australiaeast.cloudapp.azure.com:3000" `
  -H "Access-Control-Request-Method: GET"

# Expected: 200 OK with Access-Control-Allow-Origin header
```

### 1.9 Advantages of AKS-Native Pattern

- **Zero Kubernetes Secrets**: Workload Identity eliminates credential storage (no secrets.yaml needed)
- **Centralized Configuration**: Single App Configuration instance shared across multiple AKS clusters
- **Audit Trail**: Azure Activity Log records every CORS origin change with timestamp, user, and change details
- **Dynamic Refresh**: Application can reload CORS origins every 5 minutes without pod restart
- **Least Privilege**: Managed identity has only `App Configuration Data Reader` role (no `*` permissions)
- **Multi-Cluster Support**: Same configuration used across development, staging, production clusters
- **Production-Grade**: Aligns with "The AKS Book" best practices for identity, configuration, and deployment
- **No ConfigMap Synchronization**: No risk of placeholder variables (`${VAR}`) not being substituted (App Configuration API does substitution)

---

## Azure Monitor & Container Insights: Detecting CORS Failures

Monitor and troubleshoot CORS errors using Azure Monitor and Container Insights (per AKS Book Ch. 11 - Observability).

### Monitoring Setup

**Prerequisite**: Container Insights enabled on AKS cluster:

```powershell
az aks enable-addons -g emergency-alerts-dev-rg -n emergency-alerts-dev-aks -a monitoring
```

### CORS Failure Detection: KQL Queries

> **ContainerLogV2 required**: `ContainerLog` retires **30 September 2026**.
> All new Container Insights deployments default to `ContainerLogV2`.
> See [Container insights log schema](https://learn.microsoft.com/azure/azure-monitor/containers/container-insights-logs-schema) for migration.

**Query 1: Count CORS preflight failures (OPTIONS returning non-2xx)**

```kusto
ContainerLogV2
| where TimeGenerated > ago(1h)
| where LogMessage contains "OPTIONS" or LogMessage contains "Access-Control"
| where LogMessage contains "403" or LogMessage contains "401" or LogMessage contains "400"
| summarize FailureCount = count() by bin(TimeGenerated, 5m), PodName
| render timechart
```

**Query 2: CORS errors in application logs**

```kusto
ContainerLogV2
| where TimeGenerated > ago(1h)
| where LogMessage contains "CORS" or LogMessage contains "Access-Control"
| where LogMessage contains "error" or LogMessage contains "denied"
| project TimeGenerated, PodName, LogMessage
| order by TimeGenerated desc
```

**Query 3: Pods unable to connect to App Configuration**

```kusto
ContainerLogV2
| where TimeGenerated > ago(1h)
| where LogMessage contains "App Configuration" or LogMessage contains "AppConfig"
| where LogMessage contains "error" or LogMessage contains "failed" or LogMessage contains "403" or LogMessage contains "401"
| summarize count() by PodName, LogMessage
| render barchart
```

### Alert Rules

Create alerts to notify when CORS failures spike:

```powershell
# Alert: High CORS failure rate
az monitor metrics alert create \
  -g emergency-alerts-dev-rg \
  -n "CORS Failures Alert" \
  --scopes "/subscriptions/<sub-id>/resourcegroups/emergency-alerts-dev-rg/providers/microsoft.containerservice/managedclusters/emergency-alerts-dev-aks" \
  --condition "avg ContainerLogV2 > 10" \
  --description "Alert when CORS failures exceed 10 in 5 minutes" \
  --enabled true
```

### Troubleshooting with Container Insights

**In Azure Portal:**

1. Navigate to AKS cluster → Insights → Containers
2. Check pod logs for App Configuration connection errors
3. Verify managed identity assignments (Identity section)
4. Check NetworkPolicy (if enabled) isn't blocking egress to App Configuration

---

## ADAC Health Contract (Auto-Detect, Auto-Declare, Auto-Communicate)

**Purpose:** Ensure CORS + config readiness is visible to operators and users.

### Auto-Detect

- Track App Configuration connectivity and last refresh time.
- Detect missing or empty CORS origins during startup and refresh.

### Auto-Declare

- `/health/ready` returns structured status with reasons (e.g., `corsOriginsLoaded=false`).
- Include a data freshness field (timestamp or age) for config refresh.

### Auto-Communicate

- When CORS config is missing or stale, expose a degraded mode reason.
- Frontend should surface a visible banner or status indicator when API reports degraded mode.

### ADAC Checklist (API)

- [ ] `/health/ready` includes CORS/config state and freshness
- [ ] `/health/live` remains fast and minimal
- [ ] Degraded mode reasons are user-visible (not just logs)

---

## Health Probes for CORS Services

Implement all three probe types (per AKS Book Ch. 6 - Production Deployment) to ensure CORS service health and prevent traffic to pods with stale CORS configuration.

### Readiness Probe: Delay Traffic Until CORS Loaded

The readiness probe prevents traffic to pods that haven't yet loaded CORS origins from App Configuration:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: emergency-alerts-api
spec:
  template:
    spec:
      containers:
        - name: api
          image: ${ACR_NAME}.azurecr.io/api:${IMAGE_TAG} # Use immutable tags; avoid :latest

          # Readiness probe: Pod receives traffic only after CORS + config is loaded
          readinessProbe:
            httpGet:
              path: /health/ready
              port: 8080
            initialDelaySeconds: 5 # Wait 5 seconds before first probe (App Config load time)
            periodSeconds: 10 # Check every 10 seconds
            timeoutSeconds: 5 # Timeout if no response in 5 seconds
            failureThreshold: 3 # Mark unhealthy after 3 consecutive failures

          # Liveness probe: Restart pod if service becomes unresponsive
          livenessProbe:
            httpGet:
              path: /health/live
              port: 8080
            initialDelaySeconds: 30 # Startup grace period (60-120 seconds typical)
            periodSeconds: 30 # Check every 30 seconds
            failureThreshold: 3 # Restart after 3 consecutive failures

          # Startup probe: Allow old pods time to fully initialize (with App Config refresh)
          startupProbe:
            httpGet:
              path: /health/ready
              port: 8080
            failureThreshold: 30 # Allow up to 30 * 10 = 300 seconds (~5 minutes) to start
            periodSeconds: 10
```

### Health Endpoint: /health/ready

The `/health/ready` endpoint must check that CORS origins are loaded before returning 200 OK:

**ASP.NET Core Example:**

```csharp
// In Program.cs
// Implement a health check that validates CORS configuration is loaded

builder.Services.AddHealthChecks()
    .AddCheck("cors-loaded", async () =>
    {
        var corsOrigins = configuration["Cors:AllowedOrigins"];
        if (string.IsNullOrEmpty(corsOrigins))
        {
            return HealthCheckResult.Unhealthy("CORS origins not loaded from App Configuration");
        }
        return HealthCheckResult.Healthy("CORS configuration loaded");
    })
    .AddCheck("appconfig-connection", async () =>
    {
        try {
            // Verify App Configuration is connected (attempt to read a key)
            var appConfigUri = configuration["AppConfig__Endpoint"];
            if (string.IsNullOrEmpty(appConfigUri)) {
                return HealthCheckResult.Healthy("App Configuration not configured (using defaults)");
            }
            // Ping App Configuration to ensure connectivity
            using var httpClient = new HttpClient();
            var response = await httpClient.GetAsync($"{appConfigUri}/operations/check-health");
            return response.IsSuccessStatusCode
                ? HealthCheckResult.Healthy("Connected to App Configuration")
                : HealthCheckResult.Unhealthy($"App Configuration returned {response.StatusCode}");
        } catch (Exception ex) {
            return HealthCheckResult.Unhealthy($"Failed to connect to App Configuration: {ex.Message}");
        }
    });

// Map health endpoints
app.MapHealthChecks("/health/ready", new HealthCheckOptions { ResponseWriter = WriteReadinessResponse });
app.MapHealthChecks("/health/live", new HealthCheckOptions { ResponseWriter = WriteLivenessResponse });

static Task WriteReadinessResponse(HttpContext context, HealthReport report)
{
    context.Response.ContentType = "application/json";
    var json = JsonConvert.SerializeObject(new
    {
        status = report.Status.ToString(),
        checks = report.Entries
    });
    return context.Response.WriteAsync(json);
}

static Task WriteLivenessResponse(HttpContext context, HealthReport report)
{
    context.Response.ContentType = "application/json";
    if (report.Status != HealthStatus.Healthy)
        context.Response.StatusCode = StatusCodes.Status503ServiceUnavailable;
    return context.Response.WriteAsync("OK");
}
```

**Node.js/Express Example:**

```javascript
const express = require("express");
const app = express();

let corsOriginsCached = null;

app.get("/health/ready", (req, res) => {
  // Check if CORS origins are loaded
  if (!corsOriginsCached || corsOriginsCached === "http://localhost:3000") {
    return res.status(503).json({
      status: "unhealthy",
      reason: "CORS origins not loaded from App Configuration",
    });
  }

  // Check if App Configuration endpoint is reachable
  const appConfigEndpoint = process.env.AppConfig__Endpoint;
  if (!appConfigEndpoint) {
    return res.status(200).json({
      status: "healthy",
      reason: "Using default CORS configuration",
    });
  }

  res.status(200).json({
    status: "healthy",
    corsOrigins: corsOriginsCached,
    appConfigEndpoint: appConfigEndpoint,
  });
});

app.get("/health/live", (req, res) => {
  // Simple liveness check: can we respond?
  res.status(200).send("OK");
});
```

### Rolling Updates with Zero Downtime

When updating CORS configuration, use `maxUnavailable: 0` to ensure no downtime:

```yaml
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 0 # ✓ Never remove pods before new ones are ready
      maxSurge: 1 # One extra pod during update
  template:
    spec:
      terminationGracePeriodSeconds: 30 # Graceful shutdown time
```

---

## Pattern 2: Environment Variable Injection (Post-Deployment)

Use this when the deployment is already running and you need to update CORS without redeploying.

### 2.1 Set Environment on Running Deployment

```powershell
# Update the deployment with new CORS origins
kubectl set env deployment/api \
  -n your-namespace \
  "Cors__AllowedOrigins=http://frontend-emergency-alerts-dev.australiaeast.cloudapp.azure.com:3000"

# This triggers a rolling update - old pods are replaced with new ones
# that read the new environment variable on startup
```

### 2.2 Verify Rollout

```bash
# Watch the rolling update
kubectl rollout status deployment/api -n your-namespace --timeout=300s

# Check new pods are running
kubectl get pods -l app=api -n your-namespace
```

### 2.3 Validate CORS After Restart

```bash
# Preflight test (as above)
curl -i -X OPTIONS http://<api-ip>:8080/api/v1/endpoint \
  -H "Origin: http://frontend-emergency-alerts-dev.australiaeast.cloudapp.azure.com:3000" \
  -H "Access-Control-Request-Method: GET"
```

---

## Pattern 3: Ingress-Based CORS (Reverse Proxy Pattern)

When using an Ingress controller, CORS can be handled by the ingress instead of the API.

### 3.1 Ingress with CORS Annotations (NGINX Example)

> [!WARNING]
> **Retiring November 2026.** The upstream Ingress-NGINX project is deprecated (March 2026). Microsoft provides critical security patches through 30 November 2026, after which NGINX ingress controller is no longer supported in the Application Routing add-on. Migrate to Gateway API via App Routing, Application Gateway for Containers, or Istio Service Mesh. See [Azure Update 555839](https://azure.microsoft.com/updates/555839).

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: api-ingress
  namespace: your-namespace
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/enable-cors: "true"
    nginx.ingress.kubernetes.io/cors-allow-origin: "http://frontend.example.com:3000,https://frontend.example.com"
    nginx.ingress.kubernetes.io/cors-allow-credentials: "true"
    nginx.ingress.kubernetes.io/cors-allow-methods: "GET, POST, PUT, DELETE, OPTIONS"
spec:
  tls:
    - hosts:
        - api.example.com
      secretName: api-tls
  rules:
    - host: api.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: api
                port:
                  number: 8080
```

### 3.2 Advantage of Ingress CORS

- CORS headers added by the ingress, not the application
- Reduces app complexity
- Centralized CORS policy management
- Can be updated without redeploying the app

---

## Checklist: Kubernetes CORS Configuration

### Pre-Deployment

- [ ] **Get Frontend Origin**: Determine the exact frontend URL including protocol and port
  - Example: `http://frontend-emergency-alerts-dev.australiaeast.cloudapp.azure.com:3000`
  - **Common Mistake**: Using `http://localhost:3000` in prod, or omitting the port

- [ ] **Set Env Variable**: Before applying K8s manifests, set `CORS_ALLOWED_ORIGINS`

  ```powershell
  $env:CORS_ALLOWED_ORIGINS = "http://frontend-emergency-alerts-dev.australiaeast.cloudapp.azure.com:3000"
  ```

- [ ] **ConfigMap Prepared**: ConfigMap or deployment env spec includes CORS config

  ```yaml
  Cors__AllowedOrigins: "${CORS_ALLOWED_ORIGINS}" # Will be substituted
  ```

- [ ] **Apply Script Validates**: CI/CD or deployment script substitutes placeholders
  ```powershell
  (Get-Content deployment.yaml) -replace '\$\{CORS_ALLOWED_ORIGINS\}', $env:CORS_ALLOWED_ORIGINS | kubectl apply -f -
  ```

### Post-Deployment Validation

- [ ] **ConfigMap Exists**: `kubectl get configmap app-config -n your-namespace`
  - Verify content: `kubectl get configmap app-config -n your-namespace -o jsonpath='{.data.Cors__AllowedOrigins}'`
  - Should output: `http://frontend-emergency-alerts-dev.australiaeast.cloudapp.azure.com:3000`

- [ ] **Pods Running**: `kubectl get pods -l app=api -n your-namespace`
  - All should be `1/1 Running` with 0 restarts (or low restart count)

- [ ] **Preflight Request Succeeds**:

  ```bash
  curl -i -X OPTIONS http://<api-ip>:8080/api/v1/alerts \
    -H "Origin: http://frontend-emergency-alerts-dev.australiaeast.cloudapp.azure.com:3000" \
    -H "Access-Control-Request-Method: GET"
  ```

  - Should return `200 OK` (or `204 No Content`)
  - Should include `Access-Control-Allow-Origin` header matching the origin

- [ ] **Browser Test**: Open frontend in browser, F12 Network tab
  - Make an API call
  - Verify `GET /api/v1/alerts` returns `200` (not red/failed)
  - Verify NO CORS error in Console

---

## Troubleshooting: Common CORS Failures

| Error                                                                          | Root Cause                                             | Solution                                                                                 |
| ------------------------------------------------------------------------------ | ------------------------------------------------------ | ---------------------------------------------------------------------------------------- |
| `No 'Access-Control-Allow-Origin' header`                                      | ConfigMap/env var not set or not matching              | Check: `kubectl get configmap app-config -o jsonpath='{.data.Cors__AllowedOrigins}'`     |
| `Origin 'http://localhost:3000' is not allowed by Access-Control-Allow-Origin` | API origin list doesn't include frontend URL           | Verify frontend actual URL (not localhost if deployed) and update `CORS_ALLOWED_ORIGINS` |
| CORS works locally, fails in K8s                                               | Hardcoded localhost in app code                        | Check for `WithOrigins("http://localhost:*")` in code; use config instead                |
| Restarting pods doesn't fix CORS                                               | Environment variable not passed to deployment          | Use `kubectl set env` or apply new deployment with env var                               |
| Port mismatch (e.g., `:3000` vs `:80`)                                         | Origin includes port but API allows without it         | Ensure origins match exactly: `http://host.com:3000` ≠ `http://host.com`                 |
| `URIAssemblyQualifiedName` in parsing                                          | Non-URI format in CORS_ALLOWED_ORIGINS                 | Verify format: `http://host.com:port` (no trailing slash)                                |
| Wildcard `*` doesn't work with credentials                                     | CORS policy with credentials requires explicit origins | Set explicit origins: `http://frontend.example.com` (not `*`)                            |

---

## Real Example: Emergency Alerts API

### Scenario

- API: `http://20.167.110.5:8080` (LoadBalancer IP) or `http://api-emergency-alerts-dev.australiaeast.cloudapp.azure.com:8080` (DNS)
- Frontend: `http://20.227.31.148:3000` (LoadBalancer IP) or `http://frontend-emergency-alerts-dev.australiaeast.cloudapp.azure.com:3000` (DNS)

### Setup

**1. Create ConfigMap**:

```powershell
$env:CORS_ALLOWED_ORIGINS = "http://frontend-emergency-alerts-dev.australiaeast.cloudapp.azure.com:3000"

kubectl create configmap emergency-alerts-config `
  --from-literal=Cors__AllowedOrigins=$env:CORS_ALLOWED_ORIGINS `
  --from-literal=ASPNETCORE_ENVIRONMENT=Production `
  -n emergency-alerts
```

**2. Deployment references ConfigMap**:

```yaml
envFrom:
  - configMapRef:
      name: emergency-alerts-config
```

**3. Test**:

```bash
curl -i -X OPTIONS http://20.167.110.5:8080/api/v1/alerts \
  -H "Origin: http://frontend-emergency-alerts-dev.australiaeast.cloudapp.azure.com:3000" \
  -H "Access-Control-Request-Method: GET"

# Expected:
# HTTP/1.1 200 OK
# Access-Control-Allow-Origin: http://frontend-emergency-alerts-dev.australiaeast.cloudapp.azure.com:3000
# ...
```

**4. Browser test**:

- Open: `http://frontend-emergency-alerts-dev.australiaeast.cloudapp.azure.com:3000`
- F12 → Network tab
- Verify API requests succeed (no red CORS failures)

---

## Key Takeaways

1. **CORS origins must be set BEFORE or WITH deployment**, not after
2. **ConfigMap approach is preferred** for maintainability and consistency
3. **Pod restarts are required** for environment variable changes to take effect (ConfigMap changes don't auto-reload)
4. **Test with curl first** before debugging in the browser
5. **Port matching is critical** - `http://host:3000` is different from `http://host:80`
6. **Always use explicit origins** when `AllowCredentials` is true (SignalR, sessions, etc.)
7. **Ingress-based CORS** is an alternative that centralizes CORS policy at the network boundary

---

## CI/CD Integration: Safeguard Pattern

Add this to your CI/CD pipeline to prevent CORS misconfigurations:

```yaml
# Azure Pipeline example
- name: Validate CORS Configuration
  script: |
    # Ensure environment variable is set
    if [ -z "$CORS_ALLOWED_ORIGINS" ]; then
      echo "ERROR: CORS_ALLOWED_ORIGINS not set"
      exit 1
    fi

    # Validate URI format (http://host:port)
    if ! echo "$CORS_ALLOWED_ORIGINS" | grep -E '^https?://[^:]+:[0-9]+$'; then
      echo "ERROR: CORS_ALLOWED_ORIGINS format invalid. Expected: http://host:port"
      exit 1
    fi

    # Substitute into manifest and validate
    envsubst < deployment.yaml | kubectl dry-run=client -f -

    echo "✅ CORS configuration validated"
```

---

## Currency and Verification

- **Date checked:** 2026-03-31

| Component                                         | Status                       | Notes                                                                                                                                                       |
| ------------------------------------------------- | ---------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ContainerLogV2 KQL schema                         | ✅ Current                   | `ContainerLog` retires 30 Sep 2026; all queries use `ContainerLogV2`                                                                                        |
| Workload Identity pattern                         | ✅ Current                   | OIDC federation + `azure.workload.identity/use` label                                                                                                       |
| `service.beta.kubernetes.io/azure-dns-label-name` | ✅ Current (beta annotation) | Stable despite `beta` in annotation name; no GA rename announced                                                                                            |
| Azure App Configuration SDK                       | ✅ Current                   | `DefaultAzureCredential` + `AddAzureAppConfiguration` pattern                                                                                               |
| NGINX Ingress CORS annotations                    | ⚠️ Retiring Nov 2026         | Upstream Ingress-NGINX deprecated March 2026; Microsoft bridge through November 2026. Migrate to Gateway API via App Routing. See Azure Update ID `555839`. |
| Gateway API on AKS                                | ⚠️ Preview                   | Managed Gateway API CRDs require `aks-preview` extension + `ManagedGatewayAPIPreview` feature flag                                                          |
| ASP.NET Core CORS middleware                      | ✅ Current                   | `AddCors` + `UseCors` pattern (ASP.NET Core 10)                                                                                                             |

**Last reviewed:** 2026-03-17
**Verification:** `mcp_microsoft_doc_microsoft_docs_search` for ContainerLogV2 schema, Gateway API AKS preview status

---

## Known Pitfalls

| #   | Pitfall                                                        | Impact                                                                           | Prevention                                                                                             |
| --- | -------------------------------------------------------------- | -------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------ |
| 1   | ConfigMap changes don't auto-reload in running pods            | CORS origins stay stale until pods are restarted                                 | Always `kubectl rollout restart deployment/<name>` after ConfigMap updates                             |
| 2   | CORS origin format must match exactly (protocol + host + port) | `http://host:3000` ≠ `http://host` ≠ `https://host:3000`                         | Validate with `curl -X OPTIONS -H "Origin: <exact-origin>"` before deploying                           |
| 3   | Wildcard `*` origin incompatible with `AllowCredentials`       | Browser rejects response; SignalR/cookies fail silently                          | Always use explicit origins when `AllowCredentials()` is enabled                                       |
| 4   | `ContainerLog` table retires 30 September 2026                 | KQL queries stop returning data after retirement                                 | Migrate all queries to `ContainerLogV2` (uses `LogMessage` not `LogEntry`, `PodName` not `Computer`)   |
| 5   | Placeholder `${CORS_ALLOWED_ORIGINS}` not substituted          | API starts with literal `${CORS_ALLOWED_ORIGINS}` string; all CORS requests fail | Add pipeline validation step: `grep -qF '${' manifests/*.yaml && exit 1`                               |
| 6   | DNS label provisioning delay (30–60 seconds)                   | CORS origin set to DNS name but name doesn't resolve yet                         | Wait and verify with `nslookup` before setting CORS origins                                            |
| 7   | Hardcoded ACR hostnames in deployment manifests                | Images fail to pull if registry changes; violates org policy                     | Use pipeline-injected `${ACR_NAME}` variables; lint manifests for `.azurecr.io` literals               |
| 8   | Gateway API on AKS is still preview                            | Requires feature flag registration and preview CLI extension                     | Use Ingress or Application Routing add-on for production until GA; document preview opt-in if accepted |

---

**References:**

- [Fetch Living Standard: CORS protocol](https://fetch.spec.whatwg.org/#http-cors-protocol)
- [MDN: CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)
- [Microsoft API Guidelines: CORS](https://github.com/microsoft/api-guidelines)
- [ASP.NET Core CORS Documentation](https://docs.microsoft.com/aspnet/core/security/cors)
- [Kubernetes ConfigMap](https://kubernetes.io/docs/concepts/configuration/configmap/)
- [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)
- [Container Insights Log Schema (ContainerLogV2)](https://learn.microsoft.com/azure/azure-monitor/containers/container-insights-logs-schema)
- [Managed Gateway API CRDs on AKS (preview)](https://learn.microsoft.com/azure/aks/managed-gateway-api)

---

## Related Skills

- [.NET Backend Patterns](../dotnet-backend-patterns/SKILL.md) — ASP.NET Core CORS middleware setup
- [SPA Endpoint Configuration](../spa-endpoint-configuration/SKILL.md) — Frontend endpoint and build-arg patterns
- [Azure Defaults](../azure-defaults/SKILL.md) — Baseline AKS configuration
