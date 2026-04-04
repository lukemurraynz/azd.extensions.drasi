---
applyTo: "**/k8s/**/*.yml,**/k8s/**/*.yaml,**/kubernetes/**/*.yml,**/kubernetes/**/*.yaml,**/helm/**/templates/**/*.yml,**/helm/**/templates/**/*.yaml,**/Dockerfile,**/Dockerfile.*,**/*.dockerfile,**/docker-compose*.yml,**/docker-compose*.yaml,**/deploy/**,**/deployment/**,.github/workflows/**"
description: Post-deploy validation checklists, known failure modes, CORS/DNS/SPA verification, and critical configuration change rules for Kubernetes, Docker, and CI/CD deployments
---

# Deployment Safety & Operational Excellence

These checklists apply when deploying, configuring, or reviewing deployment artifacts. They prevent the most common production failures seen in SPA + API + Kubernetes architectures.

## Post-Deploy Validation Checklist (Required)

### DNS Label Verification (Kubernetes)

- [ ] **DNS labels must be verified per service** — Azure doesn't auto-create from annotations; deletion and recreation required
  - Verify: `nslookup <service>-<env>-<suffix>.<region>.cloudapp.azure.com`
  - If "Non-existent domain": Delete services, recreate with annotations, wait 60+ seconds, test again
  - Failure mode: IP address works but DNS name doesn't resolve

### SPA Build-Time Variable Verification (Vite/Next.js)

- [ ] **Frontend image MUST be rebuilt when API URL changes** — env vars baked into compiled JS at build time
  - Rebuild trigger: API DNS label changed → rebuild frontend with `--build-arg VITE_API_URL=<new-dns>`
  - Command: `grep -F "$VITE_API_URL" dist/assets/*.js`
  - Failure mode: Frontend loads but API calls fail with `ERR_NAME_NOT_RESOLVED`
  - Prevention: ALWAYS verify API URL is in compiled bundle before pushing image to registry

### ConfigMap CORS Verification (ASP.NET Core)

- [ ] **ConfigMap CORS origins must match actual frontend DNS label exactly**
  - Get current: `kubectl get configmap <app>-config -n <namespace> -o jsonpath='{.data.Cors__AllowedOrigins}'`
  - **MANDATORY: Restart API pods after any ConfigMap change**: `kubectl rollout restart deployment/<api-deployment>`
  - Failure mode: Preflight OPTIONS returns 403 "No Access-Control-Allow-Origin header"

### Core Deployment Validation

- [ ] DNS resolves for the API public host (e.g., `Resolve-DnsName <api-host>`)
- [ ] API health is reachable: `GET /health/ready` and `GET /health/live`
- [ ] **CORS Preflight Test (MANDATORY)**: Test OPTIONS request returns `Access-Control-Allow-Origin` header
  - Command: `curl -X OPTIONS -H "Origin: <frontend-origin>" <api-url>/api/<endpoint> | grep Access-Control-Allow-Origin`
  - Failure mode: Browser console shows "CORS policy" errors; all API calls blocked
- [ ] If any check fails, STOP and triage before declaring deployment complete

## Copilot Immediate Actions

Apply during implementation and deployment:

- When reviewing .cs files: Check for health endpoints and auth bypass
- When reviewing .bicep files: Run `az bicep build` and validate output
- When reviewing deployment.yaml: Verify imagePullPolicy and resource limits
- When reviewing services/api calls: Check error handling and graceful fallback
- When reviewing any non-test code/config/IaC: Block TODO/FIXME/HACK markers, placeholder tokens, or `NotImplemented*` paths
- When reviewing workflows: Verify linting runs early and build steps have proper dependencies
- **When deploying SPA** (Vite/Next.js): MANDATORY verify compiled bundle contains correct API URL — grep dist/assets for expected URL
- **After K8s deployment**: MANDATORY test CORS preflight — OPTIONS request must return `Access-Control-Allow-Origin` header matching frontend origin
- **After ConfigMap/Secret updates**: MANDATORY restart pods — changes don't auto-apply to running containers

## Common Anti-Patterns

| Anti-Pattern                                     | Consequence                        | Prevention                                                     |
| ------------------------------------------------ | ---------------------------------- | -------------------------------------------------------------- |
| Unused Bicep parameters                          | Lint failure, dead code            | Enforce `az bicep build` in CI                                 |
| Hardcoded federated credential names             | Duplicate on redeploy              | Use deterministic naming (`uniqueString()`)                    |
| Health probes behind auth middleware             | Probes fail, pods restart          | Bypass `/health/*` paths in auth                               |
| No try-catch on external service init            | App crashes if service unavailable | Wrap startup connections with graceful fallback                |
| `imagePullPolicy: IfNotPresent`                  | Stale cached images deployed       | Always use `imagePullPolicy: Always`                           |
| ASPNETCORE_URLS with HTTPS in container          | Certificate error at startup       | Use `http://+:8080` in containers                              |
| SPA built without explicit API URL               | `ERR_NAME_NOT_RESOLVED` at runtime | Build with `--build-arg VITE_API_URL=<host>`, verify in bundle |
| TODO/FIXME/HACK or placeholders in non-test code | Hidden incomplete implementation   | Remove placeholders; wire full production integration          |

## Critical Configuration Changes (Mandatory Validation)

When modifying any of these **categories** of files, run the project's validation checklist:

| File Category                | Why It's Critical                                                            |
| ---------------------------- | ---------------------------------------------------------------------------- |
| API startup / `Program.cs`   | Endpoint routes, middleware order, CORS, auth configuration                  |
| Frontend API base URL config | Frontend → API communication; wrong URL = `ERR_NAME_NOT_RESOLVED`            |
| Frontend Dockerfile          | Build-time env vars (e.g., `VITE_API_URL`) baked into compiled JS            |
| K8s ConfigMaps / Secrets     | CORS origins, connection strings, env vars — pods must restart after changes |
| K8s Ingress / Gateway        | Request routing rules; mismatches cause 404s or CORS failures                |
| IaC database modules         | Firewall rules for compute ↔ database connectivity                           |
| Deployment hooks / scripts   | Variable substitution, DNS label computation, environment setup              |

## Known Failure Modes

Applicable to any SPA + API + K8s project:

| Symptom                                 | Root Cause                                   | Fix                                                     |
| --------------------------------------- | -------------------------------------------- | ------------------------------------------------------- |
| `ERR_NAME_NOT_RESOLVED`                 | Frontend built without correct API URL       | Rebuild with `--build-arg VITE_API_URL=<host>`          |
| `Unexpected token '<'`                  | Frontend receives HTML 404 instead of JSON   | Fix API URL format or ingress routing                   |
| `No Access-Control-Allow-Origin header` | API missing CORS config or not restarted     | Update ConfigMap, restart API pods                      |
| CORS preflight returns 400              | Env var not set before manifest substitution | Set `CORS_ALLOWED_ORIGINS` before applying manifests    |
| K8s Secret has placeholder text         | Variable substitution script didn't run      | Verify pipeline sets all database env vars before apply |

## CORS Configuration Checklist

- [ ] API has CORS policy configured (e.g., `options.AddDefaultPolicy` in ASP.NET Core)
- [ ] K8s ConfigMap has `Cors__AllowedOrigins` with correct placeholder format
- [ ] Deployment pipeline sets CORS env vars before applying manifests
- [ ] CORS origin matches frontend DNS exactly (no trailing slash, correct protocol/port)
- [ ] CORS preflight test passes: `curl -i -X OPTIONS <API_URL> -H "Origin: <FRONTEND_URL>" | grep Access-Control`
- [ ] API pods restarted after ConfigMap update

## Database Configuration Checklist

- [ ] K8s secrets have placeholder variables for host, user, password
- [ ] Deployment pipeline sets ALL database env vars before applying secrets
- [ ] IaC configures database firewall rules for compute connectivity
- [ ] Connection string verified after deployment (decode base64 secret)
- [ ] Database connectivity tested from compute (e.g., `psql`, `sqlcmd`)
- [ ] API pods restarted after Secret update
