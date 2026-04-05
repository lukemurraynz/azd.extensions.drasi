# Pre-flight & Deployment Checklists

## Pre-Provisioning Checklist

Go/no-go criteria before running `azd provision`.

- [ ] **Azure Credentials Valid**
  - `az account show` returns your account
  - No "authorization failed" errors
  - `azd auth login` completed (or rerun if expired)

- [ ] **Tools Installed & Updated**
  - `azd version` returns a supported/stable release (prefer latest stable from [Install azd](https://learn.microsoft.com/azure/developer/azure-developer-cli/install-azd))
  - `az version` returns a current Azure CLI release (prefer latest stable from [Install Azure CLI](https://learn.microsoft.com/cli/azure/install-azure-cli))
  - `az bicep version` (bundled with az)
  - `kubectl version --client` matches AKS target
  - `docker info` succeeds (daemon reachable, not just CLI present)

- [ ] **azure.yaml Valid**
  - File exists in repository root
  - YAML syntax is correct (no indentation errors)
  - All declared services have valid `project` paths
  - Infrastructure `module` points to valid Bicep file

- [ ] **Bicep Valid**
  - Run: `az bicep build <path-to-main.bicep>`
  - No syntax errors or warnings
  - All required parameters documented

- [ ] **Environment Configured**
  - `azd env select <env-name>` executed explicitly (no implicit environment assumptions)
  - `azd env list` shows current environment
  - `AZURE_SUBSCRIPTION_ID` set and correct
  - `AZURE_LOCATION` set to valid Azure region
  - `AZURE_ENV_NAME` set (e.g., dev, staging, prod)
  - Optional uniqueness suffixes reviewed for global-name resources (`APPCONFIG_NAME_SUFFIX`, `KEYVAULT_NAME_SUFFIX`)

- [ ] **Subscription Permissions**
  - Your role = Contributor or Owner
  - No temporary role removals in progress (wait 1-2 mins)
  - Can create resource groups: `az group create --name test-$RANDOM --location "<location>" && az group delete --name test-$RANDOM --yes`

- [ ] **What-If Reviewed**
  - Run: `azd provision --preview`
  - No deletion of unintended resources
  - No permission (`Forbidden`) errors
  - No API version or resource type errors

---

## Deployment Checklist

Go/no-go criteria before running `azd deploy`.

- [ ] **Provisioning Complete**
  - All infrastructure resources exist in Azure Portal
  - AKS cluster is in "Succeeded" state (if using AKS)
  - ACR exists and is "Enabled" (if using container images)
  - Database is ready (if applicable)

- [ ] **AKS Access Configured**
  - `kubectl cluster-info` succeeds
  - `kubectl get nodes` shows ≥1 ready node
  - KUBECONFIG points to correct cluster

- [ ] **ACR Accessible**
  - `az acr show --name <acr-name>` succeeds
  - `az acr login --name <acr-name>` succeeds
  - Can access registry: `docker login <acr>.azurecr.io`
  - Push permissions validated: `docker push <acr>.azurecr.io/<repo>:<tag>` succeeds for each repo used in deployment

- [ ] **Container Images Ready** (if pre-built)
  - Latest image tags exist: `az acr repository show-tags --name <acr> --repository api`
  - No pending builds or failed builds

- [ ] **Kubernetes Namespace Created** (if required)
  - Run: `kubectl create namespace <your-namespace>` (or apply from manifest)
  - Verify: `kubectl get namespace <your-namespace>`

- [ ] **Environment Variables Set for Build**
  - `apiUrl` (frontend API URL) computed and verified
  - `frontendUrl` computed for CORS configuration
  - DNS label suffix matches ingress DNS names

- [ ] **Secrets Resolution Path Validated**
  - Primary secret path works (for example, Key Vault read for database password)
  - Fallback path verified if primary path can fail (for example, `azd env` secret fallback)
  - No placeholder values remain in rendered K8s secrets/config

- [ ] **Secrets Stored** (if manually required)
  - Database password set: `kubectl create secret generic db-credentials ...`
  - API keys set in Key Vault
  - All secrets are non-empty

---

## Post-Deployment Checklist

Go/no-go criteria after `azd deploy` completes.

**Immediate (within 2 mins)**

- [ ] **Pods Running**
  - `kubectl get pods -n <your-namespace>` shows ≥ 1 Running pod per service
  - No `CrashLoopBackOff` or `ImagePullBackOff`
  - READY count shows 1/1 for each pod

- [ ] **Services Created**
  - `kubectl get svc -n <your-namespace>` shows LoadBalancer services
  - External IPs assigned (not `<pending>`)
  - If `<pending>`, wait 2-5 mins then `kubectl get svc --watch`

- [ ] **No Error Events**
  - `kubectl get events -n <your-namespace>` shows no `Warning` or `Error`
  - Pod restart count = 0
  - No failed image pulls for required repos (for example, `api`, `frontend`)

**Short-term (within 5 mins)**

- [ ] **Health Endpoints Responding**
  - Get API external IP: `kubectl get svc <api-service-name> -n <your-namespace> -o jsonpath='{.status.loadBalancer.ingress[0].ip}'`
  - Test: `curl http://<ip>:8080/health/ready`
  - Expected: HTTP 200 OK with JSON response

- [ ] **CORS Configured** (for frontend communication)
  - Get API URL from ingress or service
  - Test preflight: `curl -i -X OPTIONS http://<api-url>/api/v1/alerts -H "Origin: http://frontend-url"`
  - Expected: `Access-Control-Allow-Origin` header present

- [ ] **Ingress Rules Applied**
  - `kubectl get ingress -n <your-namespace>` shows ingress created
  - DNS labels applied to public IPs
  - Azure Portal shows LoadBalancer public IP with DNS name

**Verification (before marking deployment complete)**

- [ ] **End-to-End Test**
  - Frontend is accessible: `curl http://frontend-url/`
  - API is accessible: `curl http://api-url/<your-health-or-status-route>`
  - No 404, 403, or 500 errors

- [ ] **ADAC Runtime Declaration**
  - Health endpoints include data freshness and connection status (if applicable)
  - UI shows degraded mode reason when live data is unavailable

- [ ] **Database Connected** (if applicable)
  - Check API logs for successful DB connection: `kubectl logs -l app=<api-app-label> -n <your-namespace> | grep -i database`
  - No "connection refused" or "timeout" errors

- [ ] **No Pod Restarts**
  - `kubectl get pods -n <your-namespace> -o wide` shows RESTARTS = 0
  - If RESTARTS > 0, investigate with: `kubectl logs -p <pod-name>` (previous logs)

---

## Definition of Done (Deployment)

Mark deployment complete only when all checks below pass:

- [ ] `azd up` or `azd deploy` exits with code 0
- [ ] All target pods are `Running` and ready (`kubectl get pods`)
- [ ] API health endpoint returns HTTP 200
- [ ] Frontend root endpoint returns HTTP 200
- [ ] CORS preflight from frontend origin to API endpoint succeeds
- [ ] Backend automated tests pass (at minimum solution/unit scope)
- [ ] Frontend unit tests pass
- [ ] Smoke E2E tests pass (and required browser runtime is installed, e.g., Playwright browsers)
- [ ] Static analysis gate executed (Judges or equivalent) with clearly recorded profile settings (confidence threshold, AST enabled/disabled)
- [ ] Any remaining critical/high findings are either fixed or explicitly triaged with owner + follow-up
- [ ] ADAC declared in runtime surfaces (health endpoints and UI) with actionable degraded mode reasons

---

## Cleanup Checklist

Go/no-go criteria before and after `azd down`.

**Before Cleanup**

- [ ] **Backup Complete** (if needed)
  - Database backed up: `pg_dump ... > backup.sql`
  - Configuration documented or exported
  - Important logs captured (`kubectl logs ...`)

- [ ] **Resource Inventory Captured**
  - Cleanup manifest created with `az resource list` output
  - Resource count documented (for verification after cleanup)

- [ ] **Environment Correct**
  - `azd env list` shows correct current environment
  - `azd env get-values | grep AZURE_ENV_NAME` matches intended cleanup environment (not prod!)

- [ ] **Approval Obtained** (if production)
  - Stakeholders notified of cleanup
  - Maintenance window approved
  - No active deployments or user traffic

**During Cleanup**

- [ ] **Kubernetes Namespace Deleted**
  - Run: `kubectl delete namespace <your-namespace>`
  - Verify: `kubectl get namespace <your-namespace>` returns not-found

- [ ] **azd down Succeeds**
  - Exit code = 0
  - "Resource group deleted successfully" message
  - No `Error: ...` in output

**After Cleanup**

- [ ] **Resource Group Deleted**
  - `az group show --name <rg>` returns not-found (or error 404)
  - Azure Portal no longer shows resource group

- [ ] **Local State Cleaned** (if using `--purge`)
  - `.azd/environment/<env>/` no longer exists
  - `azd env list` no longer shows deleted environment

- [ ] **No Orphaned Resources**
  - `az resource list --query "[?resourceGroup=='<rg>']"` returns empty
  - No ACR images remaining (or acceptable retention policy)
  - No managed identities lingering in subscription

- [ ] **Cost Impact Verified**
  - Check Azure Cost Analysis (may take 24 hrs to update)
  - Confirm no unexpected charges for stranded resources

---

## Troubleshooting Quick Links

| Checklist Item Failed | See Action |
|----------------------|-----------|
| "What-If shows errors" | [troubleshoot-failures.md](../actions/troubleshoot-failures.md) → Step 3 |
| "Pods stuck in `Pending`" | [troubleshoot-failures.md](../actions/troubleshoot-failures.md) → Common errors: Services stuck with `<pending>`  |
| "CORS header missing" | [troubleshoot-failures.md](../actions/troubleshoot-failures.md) → Common errors: CORS policy |
| "`azd down` fails" | [troubleshoot-failures.md](../actions/troubleshoot-failures.md) → Cleanup failures |
| "Exit code 1 everywhere" | [troubleshoot-failures.md](../actions/troubleshoot-failures.md) → Step 5: Diagnose Specific Failures |

---

## Template Authoring Checklist

Use this checklist when creating or publishing a new azd template.

### azure.yaml Metadata

- [ ] **Version set in `azure.yaml`** — Track template versions in the metadata field:
  ```yaml
  name: my-template
  metadata:
    template: my-template@1.0.0
  ```
- [ ] **azure.yaml validated against schema** — Use the [official JSON schema](https://github.com/Azure/azure-dev/blob/main/schemas/v1.0/azure.yaml.json) for IDE validation
- [ ] **All services have correct `project` paths** pointing to buildable app directories

### Bicep Parameters

- [ ] **Sensible defaults in `main.parameters.json`** — Use `${VAR=default}` syntax so users don't have to `azd env set` every parameter on first deploy (see [environment-variables.md](environment-variables.md)):
  ```json
  { "value": "${SOME_PARAM=my-default-value}" }
  ```
- [ ] **Optional resources use boolean param + output roundtrip** — So values persist across re-deployments without re-prompting the user
- [ ] **Resource names use deterministic unique suffix** — Use `uniqueString(subscription().subscriptionId, environmentName, location)` to prevent global name collisions without manual suffix management

### Hooks

- [ ] **All hooks use parameter defaults from env vars** — So scripts can be tested without running azd (see [authoring-hooks.md](authoring-hooks.md))
- [ ] **`az account set --subscription` called first** in every hook that uses Azure CLI
- [ ] **Hooks documented in README** — Explain what each hook does, what it requires, and what it produces
- [ ] **Entra ID cleanup has `predown` hook** — If the template creates app registrations or service principals (see [common-traps.md](common-traps.md) — Trap #14)

### Dependency Management

- [ ] **Automated dependency updates configured** — Use [Renovate](https://docs.renovatebot.com/) (preferred over Dependabot for azd templates because Renovate supports Bicep resource version updates). Dependabot currently cannot update Bicep API versions.
  - Renovate config example: `.github/renovate.json`
    ```json
    {
      "$schema": "https://docs.renovatebot.com/renovate-schema.json",
      "extends": ["config:base"],
      "bicep": { "enabled": true }
    }
    ```
- [ ] **CI pipeline validates dependency updates** — Renovate PRs should trigger the full build/deploy/verify/cleanup pipeline automatically

### README

- [ ] **Three-command quick start** documented:
  ```bash
  azd init --template <org>/<template>
  azd auth login
  azd up
  ```
- [ ] **Required tools listed** (beyond azd itself — e.g., Docker, kubectl)
- [ ] **Required Azure permissions documented** (Contributor, specific RBAC roles, etc.)
- [ ] **Configuration parameters documented** with how to override via `azd env set`
- [ ] **Troubleshooting section** for known deployment issues (especially purge/soft-delete issues)
- [ ] **Hook descriptions** included (what each hook does)
- [ ] **Known limitations** documented (e.g., multi-region deployment caveats — Trap #15)

### Publishing

- [ ] **Tagged release created** using `gh release create` with auto-generated notes:
  ```bash
  gh release create "v1.0.0" --generate-notes --target main --latest
  ```
- [ ] **Template published to [Awesome azd](https://azure.github.io/awesome-azd/)** (if reusable by the community)
