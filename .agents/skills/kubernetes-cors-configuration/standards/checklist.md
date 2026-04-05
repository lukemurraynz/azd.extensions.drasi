# AKS CORS Configuration Checklist

Use this checklist when implementing CORS configuration for APIs and SPAs deployed to AKS.

## Pre-Deployment: Infrastructure Setup

### DNS Labels (Foundation for Stable CORS Origins)

- [ ] **Frontend Service Has DNS Label**
  - Annotation: `service.beta.kubernetes.io/azure-dns-label-name: "frontend-emergency-alerts-dev"`
  - DNS name: `frontend-emergency-alerts-dev.australiaeast.cloudapp.azure.com`
  - Resolved IP: matches LoadBalancer EXTERNAL-IP
  - Test: `nslookup frontend-emergency-alerts-dev.australiaeast.cloudapp.azure.com`

- [ ] **API Service Has DNS Label**
  - Annotation: `service.beta.kubernetes.io/azure-dns-label-name: "api-emergency-alerts-dev"`
  - DNS name: `api-emergency-alerts-dev.australiaeast.cloudapp.azure.com`
  - Test: `curl -s http://api-emergency-alerts-dev.australiaeast.cloudapp.azure.com/health/ready`

- [ ] **SignalR Service Has DNS Label** (if separate from API)
  - Annotation: `service.beta.kubernetes.io/azure-dns-label-name: "signalr-emergency-alerts-dev"`
  - Prevents port collisions with API service

- [ ] **DNS Verification Complete**
  - All three services resolve via DNS (not just IP)
  - No broken DNS lookups (all nslookup commands return valid IPs)

### Azure Setup

- [ ] **Managed Identity Created**
  - Identity name: `<api-name>-identity`
  - Assigned "App Configuration Data Reader" role
  - Federated credential created linking K8s service account to managed identity
  - Command: `az identity federated-credential create --issuer <oidc-issuer> --subject system:serviceaccount:<ns>:<sa>`

- [ ] **Workload Identity Enabled on AKS**
  - OIDC issuer is configured: `az aks show --query oidcIssuerProfile.issuerUrl`
  - Workload Identity add-on enabled (check cluster settings)
  - Service account has `azure.workload.identity/client-id` annotation

- [ ] **Azure App Configuration Has CORS Origins**
  - Key: `Cors:AllowedOrigins`
  - Value format: `http://frontend-host:port,http://another-origin:port`
  - Label: `development` or `production` (matching deployment label)
  - Stored without trailing slashes (e.g., `http://localhost:3000` not `http://localhost:3000/`)
  - Command: `az appconfig kv show -n <app-config-name> --key "Cors:AllowedOrigins"`

- [ ] **Azure App Configuration Connection**
  - Endpoint format: `https://<config-name>.azconfig.io`
  - Accessible from AKS cluster (no connectivity issues)
  - Verified: `kubectl exec -it <pod> -- curl https://<endpoint>/...`

### Kubernetes Setup

- [ ] **Service Account Created with Workload Identity**

  ```yaml
  serviceAccountName: <app-name>
  annotations:
    azure.workload.identity/client-id: <client-id>
  ```

- [ ] **Deployment Uses Service Account**
  - Pod template includes: `serviceAccountName: <app-name>`
  - Pod labels include: `azure.workload.identity/use: "true"`
  - No secrets mounted (Workload Identity supplies credentials)

- [ ] **Container Environment Variables Set**
  - `AppConfig__Endpoint=https://<config-name>.azconfig.io`
  - `AppConfig__Label=<development|production>`
  - `ASPNETCORE_ENVIRONMENT=Production` (or equivalent for your framework)

- [ ] **Health Checks Configured**
  - Readiness probe: `/health/ready` returns 200 only after CORS loaded
  - Liveness probe: `/health/live` checks basic connectivity
  - Startup probe: Graceful period for slow initialization
  - All three probes configured in deployment

- [ ] **Resource Limits Set**
  - CPU request: `200m` (95th percentile of actual usage)
  - Memory request: `512Mi` (typical for API)
  - Memory limit: `1Gi` (1.5-2x request)
  - No CPU limits set (prevents throttling)

---

## Application Code

### ASP.NET Core

- [ ] **App Configuration Connected in Program.cs**

  ```csharp
  options.Connect(new Uri(appConfigEndpoint), new DefaultAzureCredential())
  ```

- [ ] **DefaultAzureCredential Used** (not ManagementIdentityCredential or custom keys)
  - Automatically uses Workload Identity token from `/var/run/secrets/azure/tokens/token`
  - Falls back to other credential sources in development

- [ ] **CORS Configured from App Configuration**

  ```csharp
  var corsOrigins = builder.Configuration["Cors:AllowedOrigins"];
  options.AddDefaultPolicy(policy =>
    policy.WithOrigins(corsOrigins.Split(',', ...))
  );
  ```

- [ ] **Cors middleware placed before authentication**
  - Order: `app.UseCors()` → `app.UseAuthentication()` → `app.UseAuthorization()`

- [ ] **Health checks implemented**
  - `/health/ready` checks CORS origins are loaded
  - `/health/live` basic service health
  - Added to health checks: `builder.Services.AddHealthChecks()`

- [ ] **Dynamic refresh enabled (optional)**
  - `ConfigureRefresh()` registers `Cors:AllowedOrigins` key
  - Refresh interval set to 5 minutes
  - Middleware: `await refreshConfig.TryRefreshAsync()`

### Node.js/Express

- [ ] **App Configuration Client Imported**

  ```javascript
  const { AppConfigurationClient } = require("@azure/app-configuration");
  const { DefaultAzureCredential } = require("@azure/identity");
  ```

- [ ] **CORS Origins Loaded on Startup**

  ```javascript
  const setting = await client.getConfigurationSetting({
    key: "Cors:AllowedOrigins",
    label: process.env.AppConfig__Label,
  });
  ```

- [ ] **CORS Middleware Configured**

  ```javascript
  app.use(
    cors({
      origin: allowedOrigins.split(",").map((o) => o.trim()),
      credentials: true,
    }),
  );
  ```

- [ ] **cors() placed before authentication routes**
  - Order: `app.use(cors(...))` → `app.use(auth)` → routes

- [ ] **Health endpoints implemented**
  - `GET /health/ready` → 200 if CORS loaded, 503 if not
  - `GET /health/live` → 200 always (unless service down)

### Errors Handled

- [ ] **App Configuration connection failures handled gracefully**
  - Try-catch around `client.getConfigurationSetting()`
  - Falls back to default CORS origins if connection fails
  - Logs warning: "Failed to load CORS from App Configuration"
  - Service still starts (not blocking deployment)

- [ ] **Missing CORS origins handled**
  - Fallback value: `http://localhost:3000,http://localhost:5173`
  - Logged to console on startup
  - Readiness probe returns 503 (unhealthy) until real origins loaded

---

## Deployment

- [ ] **Workload Identity Token Verified**
  - Pod has token: `kubectl exec <pod> -- ls /var/run/secrets/azure/tokens/token`
  - Token is valid: `kubectl exec <pod> -- cat /var/run/secrets/azure/tokens/token | jwtdecode`

- [ ] **Pods Running and Healthy**
  - `kubectl get pods -l app=<app-name>` → all `1/1 Running`
  - Restart count is low (0 or 1 on initial deployment)
  - Readiness probe: all pods `Ready`

- [ ] **Deployment Applied with Zero Downtime**
  - Strategy: `RollingUpdate` with `maxUnavailable: 0`
  - Rollout completed: `kubectl rollout status deployment/<app-name>`
  - No 502/503 errors during update

---

## Testing & Validation

### CORS Preflight Test

- [ ] **CURL Test: Preflight Request**

  ```bash
  curl -i -X OPTIONS http://<api-url>/api/v1/endpoint \
    -H "Origin: http://frontend-host:port" \
    -H "Access-Control-Request-Method: GET"
  ```

  - Response: `200 OK` or `204 No Content`
  - Header present: `Access-Control-Allow-Origin: http://frontend-host:port`
  - No `CORS policy` errors in output

- [ ] **CURL Test: Actual Request**

  ```bash
  curl -i -X GET http://<api-url>/api/v1/endpoint \
    -H "Origin: http://frontend-host:port"
  ```

  - Response: `200` (or appropriate status, not 403)
  - Header present: `Access-Control-Allow-Origin`

- [ ] **Browser Test: Frontend Makes API Call**
  - Open frontend in browser: `http://frontend-host:port`
  - Open DevTools → Network tab
  - Make an API call (e.g., GET /api/v1/alerts)
  - Request shows: `Status: 200` (green, not red)
  - Console shows: No `CORS policy` errors
  - Response data displayed correctly in UI

- [ ] **Cross-Origin Request Works**
  - Frontend origin: `http://frontend-host:3000`
  - API origin: `http://api-host:8080`
  - Browser request succeeds (not blocked by CORS)

### App Configuration Connection Test

- [ ] **Pod Can Read CORS Origins from App Configuration**

  ```bash
  kubectl exec <pod> -it -- bash -c "curl -s http://localhost:8080/health/ready | jq ."
  ```

  - Output shows: `corsOrigins` field populated (not null/empty)
  - Status: `healthy`

- [ ] **App Configuration Update Reflected**
  - Update CORS origin in App Configuration: `az appconfig kv set ...`
  - Wait 5 minutes (dynamic refresh interval)
  - Pod's `/health/ready` endpoint shows updated origin
  - No pod restart required (dynamic refresh working)

### Workload Identity Test

- [ ] **Pod Token Valid and Fresh**

  ```bash
  kubectl exec <pod> -it -- cat /var/run/secrets/azure/tokens/token | head -c 100
  ```

  - Token file exists and is non-empty
  - Contains JWT (three base64-encoded parts separated by dots)

- [ ] **Managed Identity Permissions Verified**
  ```bash
  az role assignment list --assignee <managed-identity-id> --scope <app-config-resource-id>
  ```

  - Role: `App Configuration Data Reader` or equivalent
  - Scope: App Configuration resource

---

## Monitoring & Troubleshooting

- [ ] **Pod Logs Checked for CORS Errors**
  - `kubectl logs -l app=<app-name> -n <namespace> | grep -i cors`
  - No errors about CORS origins missing or malformed

- [ ] **App Configuration Connection Logs Checked**
  - `kubectl logs <pod> | grep -i "app configuration"`
  - No "Connection refused", "403", or "Unauthorized" errors
  - Should see: "✅ Loaded CORS origins from App Configuration"

- [ ] **Container Insights/Azure Monitor Queries Run**
  - CORS failure query: `ContainerLogV2 | where LogMessage contains "Access-Control"`
  - Workload Identity query: `ContainerLogV2 | where LogMessage contains "azure.workload.identity"`
  - No spike in errors indicating widespread CORS failures

- [ ] **Health Probes Checked in Metrics**
  - `kubectl top pods` shows CPU/memory usage is stable
  - Readiness probe failures: `kubectl describe pod <pod>` shows 0 failures
  - Liveness probe not restarting pods (low restart count)

---

## Post-Deployment Updates

### Updating CORS Origins

- [ ] **Update via App Configuration (No Pod Restart)**

  ```powershell
  az appconfig kv set -n <app-config> --key "Cors:AllowedOrigins" \
    --value "http://new-frontend-host:port" \
    --label "production"
  ```

  - Wait 5 minutes for dynamic refresh
  - Test: `curl -i -X OPTIONS ... -H "Origin: http://new-frontend-host:port"`
  - Expected: 200 OK with new origin in header

- [ ] **If Manual Update Needed: Restart Pods**
  ```bash
  kubectl rollout restart deployment/<app-name> -n <namespace>
  kubectl rollout status deployment/<app-name> -n <namespace>
  ```

  - Rolling update with zero downtime
  - All pods running before validating

### Scaling

- [ ] **Pods Scale Correctly**
  - `kubectl scale deployment <app-name> --replicas=3`
  - New pods start healthy (readiness probe passes)
  - CORS works across all replicas
  - No race condition or cache issues

---

## Emergency Procedures

### CORS Broken in Production

1. [ ] **Identify Root Cause** (within 1 minute)
   - Check pod logs: CORS origins missing or outdated?
   - Check App Configuration: CORS origins key exists and has correct value?
   - Check Workload Identity: Pod has valid token?

2. [ ] **Quick Fix: Restart Pods**

   ```bash
   kubectl rollout restart deployment/<app-name> -n <namespace>
   kubectl rollout status deployment/<app-name> -n <namespace>
   ```

   - Pods pick up latest CORS origins
   - Rolling update ensures zero downtime

3. [ ] **Verify Fix**
   ```bash
   curl -i -X OPTIONS http://<api-url>/api/v1/endpoint \
     -H "Origin: http://frontend-host:port" \
     -H "Access-Control-Request-Method: GET"
   ```

   - Expected: 200 OK with correct origin header

### App Configuration Unreachable

- [ ] **Check App Configuration Status**

  ```powershell
  az appconfig show -n <config-name> -g <resource-group>
  ```

  - Status: "Succeeded" (not disabled or updating)

- [ ] **Check Managed Identity Permissions**

  ```bash
  az role assignment list --assignee <client-id>
  ```

  - Role: "App Configuration Data Reader" assigned
  - Scope: covers the App Configuration resource

- [ ] **Check Egress Connectivity (if using NetworkPolicy)**
  - NetworkPolicy allows traffic to `*.azconfig.io:443`
  - No firewall blocking outbound HTTPS

- [ ] **Fallback to Defaults**
  - Application falls back to `http://localhost:3000,http://localhost:5173`
  - Pods become unhealthy (readiness probe fails)
  - Manual fix: Update App Configuration or restart pods with env var override

---

## Sign-Off

- [ ] Entire checklist reviewed and completed
- [ ] Manual testing passed in development environment
- [ ] Manual testing passed in staging environment
- [ ] Monitoring alerts configured and verified
- [ ] Runbook documented for post-deployment troubleshooting
- [ ] Team notified of CORS configuration approach and health check endpoints
