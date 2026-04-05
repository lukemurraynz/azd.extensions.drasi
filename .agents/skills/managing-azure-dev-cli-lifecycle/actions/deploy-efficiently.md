# Deploy Efficiently

## Purpose

Deploy applications to provisioned infrastructure. Supports multiple deployment targets: Kubernetes (AKS, using container images and kubectl), App Service (ZIP/git deployment), Container Instances, Static Web Apps, and custom targets. Handles configuration injection, health checks, and post-deployment validation.

**Supported Deployment Targets:**
- Kubernetes (AKS) — Container image build, push to ACR, `kubectl apply` manifests
- App Service (Web Apps, API Apps) — ZIP deployment, Docker container deployment, continuous deployment
- Azure Container Instances — Direct container image deployment
- Azure Static Web Apps — SPA deployment with automatic preview environments
- Custom infrastructure — Any target defined in `azure.yaml` postdeploy hooks

---

## Flow

### Step 1: Validate Deployment Readiness ✅

Ensure provisioning is complete and deployment prerequisites are met.

**Checks:**
```powershell
# Pin intended azd environment explicitly
azd env select <env-name>
azd env list

# Verify provisioning completed
azd env get-values | Select-String "acrName|acrLoginServer" | Write-Host

# Check ACR is accessible
$acrName = (azd env get-values | Select-String "acrName").Line -replace "acrName=", ""
az acr show --name $acrName
az acr login --name $acrName

# Check AKS is accessible
$aksName = (azd env get-values | Select-String "aksName|AKS_NAME").Line -replace ".*=", ""
kubectl cluster-info
kubectl get nodes

# Docker daemon must be reachable
docker info
```

**🛑 STOP**: If any check fails, re-run provisioning or troubleshoot before proceeding.

**Success Criteria:**
- [ ] ACR login server is accessible (`az acr show` succeeds)
- [ ] ACR login succeeds (`az acr login --name <acr>` returns success)
- [ ] AKS cluster is reachable (`kubectl cluster-info` succeeds)
- [ ] `kubectl get nodes` shows at least one ready node
- [ ] KUBECONFIG is set or kubectl default context points to target cluster
- [ ] Docker daemon reachable (`docker info` succeeds)

---

### Step 2: Prepare Environment Variables for Build ✅

Set build-time variables needed for container images.

**Commands:**
```powershell
# Get deployment configuration
$azdEnv = @{}
azd env get-values | ForEach-Object {
  if ($_ -match '(.+)=(.*)') {
    $azdEnv[$matches[1].Trim()] = $matches[2].Trim('"')
  }
}

# Get namespace from your deployment manifest or environment
$namespace = $azdEnv['K8S_NAMESPACE'] ?? 'default'  # Replace with your namespace

# Compute API URL dynamically (if using Kubernetes LoadBalancer DNS)
$environment = $azdEnv['AZURE_ENV_NAME'] ?? 'dev'
$location = $azdEnv['AZURE_LOCATION'] ?? '<region>'
$imageTag = (git rev-parse --short HEAD)

# Example: http://api-dev-abc123.<region>.cloudapp.azure.com
$apiUrl = "http://<api-prefix>-$environment-<dns-suffix>.$location.cloudapp.azure.com"

Write-Host "API URL: $apiUrl"
Write-Host "Namespace: $namespace"
Write-Host "ACR: $($azdEnv['acrName'])"
Write-Host "Image tag: $imageTag"
```

**🛑 STOP**: Verify URLs match your DNS labels and ingress configuration.

**Common Issues:**
| Issue | Cause | Fix |
|-------|-------|-----|
| `Missing DNS label suffix` | Deployment manifest not accessible | Ensure infrastructure/k8s/deployment.yaml exists |
| `Wrong URL scheme` | HTTP vs HTTPS mismatch | Check ingress TLS configuration |
| `CORS preflight failures (later)` | Frontend URL not in API CORS allowlist | Verify `Cors__AllowedOrigins` in ConfigMap |

**Success Criteria:**
- [ ] URLs match ingress DNS labels (verify in Azure Portal: LoadBalancer public IP DNS names)
- [ ] No unresolved variables in URLs
- [ ] ACR name is correct

---

### Step 3: Build Container Images 🐳

Validate and build application container images locally.

**Commands:**
```powershell
# Build backend API
Write-Host "Building backend API image..." -ForegroundColor Cyan
docker build -f "<api-dockerfile-path>" `
  -t "<acr>.azurecr.io/api:$imageTag" `
  "<api-build-context>" 2>&1 | Tee-Object -FilePath "build-api-$(Get-Date -Format 'yyyyMMdd-HHmmss').log"

if ($LASTEXITCODE -ne 0) { throw "Backend build failed" }

# Build frontend SPA (with Vite build-time variables)
Write-Host "Building frontend image..." -ForegroundColor Cyan
docker build -f "<frontend-dockerfile-path>" `
  -t "<acr>.azurecr.io/frontend:$imageTag" `
  --build-arg VITE_API_URL="$apiUrl" `
  --build-arg VITE_SIGNALR_URL="$apiUrl/hubs/alerts" `
  "<frontend-build-context>" 2>&1 | Tee-Object -FilePath "build-frontend-$(Get-Date -Format 'yyyyMMdd-HHmmss').log"

if ($LASTEXITCODE -ne 0) { throw "Frontend build failed" }

# Verify build outputs
Write-Host "Verifying built images..." -ForegroundColor Cyan
docker images | Select-String "api|frontend"
```

**Expected Output:**
```
REPOSITORY                              TAG      IMAGE ID      CREATED        SIZE
<acr>.azurecr.io/api                    <imageTag>   abc12345...   5 seconds ago   500MB
<acr>.azurecr.io/frontend               <imageTag>   def67890...   2 seconds ago   150MB
```

**🛑 STOP**: If build fails, check Dockerfile syntax and dependencies. See [troubleshoot-failures](troubleshoot-failures.md).

**⚠️ CRITICAL for SPA Applications (Vite, Next.js, etc.):**
- Build-time variables (Vite: `VITE_*`, Next.js: `NEXT_PUBLIC_*`) are baked into the compiled bundle
- If you omit correct values during build, the SPA will use stale/default values (often `localhost`)
- ❌ **MANDATORY: Verify** the compiled bundle contains correct values (FAIL if missing):
  ```powershell
  # Verify API URL is in compiled bundle
  $imageId = (docker images --format "{{.ID}}" "<acr>.azurecr.io/frontend:$imageTag" | Select-Object -First 1)
  $bundleCheck = docker run --rm $imageId sh -c "grep -r '$apiUrl' /app/dist/assets/*.js" 2>&1
  
  if ($LASTEXITCODE -ne 0 -or $bundleCheck -notmatch $apiUrl) {
    throw "❌ CRITICAL: Built SPA bundle does not contain VITE_API_URL=$apiUrl. Frontend will fail with ERR_NAME_NOT_RESOLVED."
  } else {
    Write-Host "✅ SPA bundle verified: Contains correct API URL: $apiUrl" -ForegroundColor Green
  }
  ```

**Success Criteria:**
- [ ] Both images build without errors (Exit Code = 0)
- [ ] `docker images` shows both api and frontend with the selected immutable tag
- [ ] Frontend build includes correct `VITE_API_URL` in compiled bundle
- [ ] Image sizes are reasonable (no bloated layers)

---

### Step 4: Push Images to ACR 📤

Push built images to Azure Container Registry.

**Commands:**
```powershell
# Log in to ACR
$acrName = "<acr>"
$imageTag = (git rev-parse --short HEAD) # Use the same tag from Step 3
az acr login --name $acrName

# Push backend
Write-Host "Pushing backend API to ACR..." -ForegroundColor Cyan
docker push "$acrName.azurecr.io/api:$imageTag"
if ($LASTEXITCODE -ne 0) { throw "API push failed" }

# Push frontend
Write-Host "Pushing frontend to ACR..." -ForegroundColor Cyan
docker push "$acrName.azurecr.io/frontend:$imageTag"
if ($LASTEXITCODE -ne 0) { throw "Frontend push failed" }

# Verify images in ACR
Write-Host "Verifying ACR images..." -ForegroundColor Cyan
az acr repository list --name $acrName --output table
az acr repository show-tags --name $acrName --repository api --output table
az acr repository show-tags --name $acrName --repository frontend --output table
```

**Expected Output:**
```
Repositories in '<acr>':
  api
  frontend

Tags in '<acr>/api':
  <imageTag>

Tags in '<acr>/frontend':
  <imageTag>
```

**🛑 STOP**: Verify both images appear in ACR before proceeding to deployment.

**Mandatory Rule**: If either `api` or `frontend` image push fails, stop immediately and do not continue to `azd deploy`.

**Common Issues:**
| Issue | Cause | Fix |
|-------|-------|-----|
| `denied: authorization failed` | ACR login expired | Run `az acr login --name <acr>` again |
| `Error response from daemon` | Docker daemon not running | Start Docker Desktop or Docker service |
| `timeout uploading image` | Network issue or large layer | Check network, retry, or push with `--verbose` |

**Success Criteria:**
- [ ] Both images appear in ACR (`az acr repository list`)
- [ ] Both `api` and `frontend` repositories contain expected immutable tag (`$imageTag` or release tag)
- [ ] No `denied` or `authorization` errors

---

### Step 5: Deploy with azd ✅

Trigger the full deployment workflow.

**Commands:**
```powershell
# Deploy (uses postdeploy hooks defined in `azure.yaml`)
Write-Host "Deploying to AKS..." -ForegroundColor Cyan
azd deploy

# Expected output (from postdeploy hook):
# 🔄 Post-deployment: Rendering configuration...
# 🔄 Post-deployment: Applying manifests/scripts...
# ✅ Post-deployment steps completed.
```

**🛑 STOP**: If deployment fails with exit code non-zero, do NOT manually apply manifests. See [troubleshoot-failures](troubleshoot-failures.md).

**Mandatory Rule**: Postdeploy scripts must fail-fast on failed image push, failed secret retrieval, failed `kubectl apply`, or unresolved placeholder substitution.

**Success Criteria:**
- [ ] Exit code = 0
- [ ] `azd deploy` completes without errors
- [ ] K8s manifests applied successfully
- [ ] Ingress created (check: `kubectl get ingress`)
- [ ] Services have external IPs (check: `kubectl get svc`)

---

### Step 6: Monitor Post-Deployment Health 🏥

Verify application is healthy after deployment.

**Commands:**
```powershell
# Check pod status
Write-Host "Checking pod status..." -ForegroundColor Cyan
kubectl get pods -n <your-namespace> --watch
# Exit watch with Ctrl+C when pods are Running

# Check services
Write-Host "Checking services..." -ForegroundColor Cyan
kubectl get svc -n <your-namespace>

# Check ingress (if using Kubernetes Ingress)
Write-Host "Checking ingress..." -ForegroundColor Cyan
kubectl get ingress -n <your-namespace>

# Get API external IP and test health endpoint
$apiServiceName = "<api-service-name>"
$apiPort = "<api-port>"
$apiSvc = kubectl get svc $apiServiceName -n <your-namespace> -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null
if ($apiSvc) {
  Write-Host "API Service IP: $apiSvc" -ForegroundColor Cyan

  # Test health endpoints
  Write-Host "Testing API health endpoints..." -ForegroundColor Cyan
  curl -i "http://$apiSvc:$apiPort/health/ready"
  curl -i "http://$apiSvc:$apiPort/health/live"
}
```

**Expected Output:**
```
# Pods
NAME                       READY   STATUS    RESTARTS   AGE
api-deployment-xxxxx       1/1     Running   0          2m
frontend-deployment-xxxxx  1/1     Running   0          2m

# Services (should show EXTERNAL-IP, not <pending>)
SERVICE              TYPE           CLUSTER-IP     EXTERNAL-IP     PORT(S)
<api-service-name>   LoadBalancer   10.0.x.x       40.x.y.z        8080:30xxx/TCP
<web-service-name>   LoadBalancer   10.0.x.y       40.a.b.c        3000:30yyy/TCP

# Health check
HTTP/1.1 200 OK
Content-Type: application/json
```

**🛑 STOP**: If pods are stuck in `Pending` or `CrashLoopBackOff`, troubleshoot before considering deployment successful. See [troubleshoot-failures](troubleshoot-failures.md).

**Common Issues:**
| Issue | Cause | Fix |
|-------|-------|-----|
| Pods in `Pending` | No available nodes / node capacity exhausted | Check node status: `kubectl get nodes` |
| Pods in `ImagePullBackOff` | Image not in ACR or AKS can't access ACR | Verify ACR auth and image tags |
| Services stuck with `<pending>` | LoadBalancer IP allocation slow | Wait 2-5 mins or check Azure quota |
| Health endpoint returns 502/503 | Service dependencies unavailable (DB, cache, etc.) | Check dependent service logs: `kubectl logs <pod>` |
| Unknown namespace | Namespace doesn't exist | Create it: `kubectl create namespace <your-namespace>` |
| `ImagePullBackOff` persists after image becomes available | Existing pods still on failed pull cycle | `kubectl rollout restart deployment/<deployment-name> -n <your-namespace>` |

---

### Step 6b: ❌ MANDATORY CORS Preflight Validation 🔐

**REQUIRED BEFORE DEPLOYMENT COMPLETE**: Verify CORS headers are configured correctly. This catches the most common frontend-to-API communication issue.

**Why This Step**: CORS misconfiguration is a silent failure—the API works, but browser blocks requests with "No 'Access-Control-Allow-Origin' header". Without this test, deployment appears successful but frontend can't fetch data.

**Commands:**
```powershell
# Get API endpoint
$apiHost = (kubectl get svc <api-service-name> -n <your-namespace> -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null) ?? "api-<env>-<suffix>.<region>.cloudapp.azure.com"
$frontendOrigin = "http://frontend-<env>-<suffix>.<region>.cloudapp.azure.com"

Write-Host "Testing CORS preflight..." -ForegroundColor Cyan
Write-Host "API Host: $apiHost" -ForegroundColor Gray
Write-Host "Frontend Origin: $frontendOrigin" -ForegroundColor Gray

# Test CORS preflight for any REST endpoint
$response = Invoke-WebRequest -Uri "http://$apiHost/api/v1/alerts?page=1" `
  -Method OPTIONS `
  -Headers @{"Origin" = $frontendOrigin} `
  -SkipHttpErrorCheck

$corsHeader = $response.Headers["Access-Control-Allow-Origin"]
$credentialsHeader = $response.Headers["Access-Control-Allow-Credentials"]

if ($corsHeader -eq $frontendOrigin) {
  Write-Host "✅ CORS configured correctly: Access-Control-Allow-Origin = $corsHeader" -ForegroundColor Green
} else {
  Write-Host "❌ CORS FAILED: Expected '$frontendOrigin' but got '$corsHeader'" -ForegroundColor Red
  Write-Host "Action: Update K8s ConfigMap with correct Cors__AllowedOrigins and restart API pods:" -ForegroundColor Yellow
  Write-Host "  kubectl rollout restart deployment/<api-deployment-name> -n <your-namespace>" -ForegroundColor Yellow
  throw "CORS validation failed; cannot proceed."
}

if ($credentialsHeader -eq "true") {
  Write-Host "✅ CORS credentials allowed" -ForegroundColor Green
} else {
  Write-Host "⚠️  CORS credentials: $credentialsHeader (may be expected)" -ForegroundColor Yellow
}
```

**Expected Output:**
```
Testing CORS preflight...
API Host: api-dev-abc123.<region>.cloudapp.azure.com
Frontend Origin: http://frontend-dev-xyz789.<region>.cloudapp.azure.com
✅ CORS configured correctly: Access-Control-Allow-Origin = http://frontend-dev-xyz789.<region>.cloudapp.azure.com
✅ CORS credentials allowed
```

**🛑 STOP**: If CORS test fails, deployment is NOT complete. Fix the ConfigMap and restart pods before proceeding to Step 7.

**Success Criteria:**
- [ ] CORS preflight returns `200 OK` (or `405 Method Not Allowed` is acceptable if endpoint doesn't support OPTIONS)
- [ ] `Access-Control-Allow-Origin` header matches frontend origin exactly (no typos, correct scheme/domain/port)
- [ ] No `ERR_NAME_NOT_RESOLVED` in browser console when frontend loads
- [ ] No "CORS policy" errors in browser console for API calls

**Success Criteria:**
- [ ] Both pods in `Running` state with `READY 1/1` (or expected replica count)
- [ ] All services show `EXTERNAL-IP` (not `<pending>`)
- [ ] Health endpoints return `200 OK`
- [ ] Post-deployment tests pass (if available)

---

### Step 7: Run Post-Deployment Tests 🧪

Validate end-to-end functionality.

**Commands:**
```powershell
# Run integration tests (if available)
Write-Host "Running integration tests..." -ForegroundColor Cyan
# Replace with your actual test command:
# .NET: dotnet test <project>/bin/Release/<project>.dll
# Node: npm test
# Python: pytest

# Browser runtime prerequisite for Playwright-based smoke tests
npx playwright install chromium-headless-shell

cd tests
if (Test-Path "./run-tests.sh") {
  ./run-tests.sh
} else {
  Write-Host "No integration test script found. Skipping."
}

if ($LASTEXITCODE -ne 0) {
  Write-Warning "Some tests failed; investigate before marking deployment complete."
} else {
  Write-Host "All tests passed!" -ForegroundColor Green
}
```

**Success Criteria:**
- [ ] All integration tests pass (or known failures are tracked in GitHub)
- [ ] No `failed` test result
- [ ] Exit code = 0

---

## Common Patterns

### One-Command Deployment (for CI/CD)

```powershell
# After azd provision succeeds, deploy with:
azd deploy
```

The postdeploy hook in `azure.yaml` handles your project-specific steps (for example, manifest application, configuration substitution, and integration setup).

### Manual K8s Deployment (if postdeploy is disabled)

```powershell
cd infrastructure
./scripts/<apply-manifests-script>.ps1
kubectl apply -f k8s/<ingress-or-routing-manifest>.yaml
```

### Rollback to Previous Deployment

```powershell
# AKS rollback (Kubernetes native)
kubectl rollout undo deployment/api-deployment -n <your-namespace>
kubectl rollout undo deployment/frontend-deployment -n <your-namespace>

# Wait for rollback
kubectl rollout status deployment/api-deployment -n <your-namespace>
kubectl rollout status deployment/frontend-deployment -n <your-namespace>
```

---

## Next Steps

✅ After successful deployment and health checks, deployment is complete.

❌ If deployment fails, see [troubleshoot-failures](troubleshoot-failures.md).

🧹 To clean up, see [cleanup-completely](cleanup-completely.md).
