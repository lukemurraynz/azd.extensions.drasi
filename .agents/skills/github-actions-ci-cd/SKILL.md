---
name: github-actions-ci-cd
description: >-
  CI/CD workflows for building, testing, and deploying the .NET 10 backend and React/Vite frontend to Azure Container Registry and AKS. USE FOR: set up GitHub Actions for .NET build/test, Docker multi-stage builds, push to ACR, deploy to AKS, validate K8s manifest placeholders, integrate azd provisioning.
---

# GitHub Actions CI/CD

Workflow patterns for the Emergency Alerts project: .NET 10 backend, React 19 / Vite frontend, Azure Container Registry, and AKS deployment.

---

## When to Use This Skill

- Creating or modifying GitHub Actions workflow files
- Setting up CI for .NET build, test, and Docker image build
- Setting up CD for pushing to ACR and deploying to AKS
- Configuring Vite build-time env var injection in Docker builds
- Adding placeholder validation before K8s manifest apply
- Integrating `azd` provisioning into CI/CD
- Troubleshooting pipeline failures (image pull errors, placeholder leaks, failed migrations)

---

## AKS Book Alignment (Future Projects)

- **Prefer pull-based GitOps for deployments**: CI builds/tests/pushes images; CD updates a GitOps repo (image tags / values) and lets the GitOps agent deploy.
- **Use the Flux extension for production** (GA). Use Argo CD extension only for dev/test (preview), or self-managed Argo CD if you must.
- **Use image update automation**: dev/staging can auto-commit; production should raise PRs to a protected branch for approval.
- **Avoid cluster credentials in CI**: GitOps agents authenticate from inside the cluster (Workload Identity).

If you must use push-based deploys (bootstrap, legacy pipelines, or non-prod), the workflow below applies.

---

## Flux Extension GitOps Example (Pull-Based)

**Goal:** CI builds/tests/pushes images → Flux pulls manifests from Git and deploys.

**1) Install Flux extension (Workload Identity):**

```bash
az k8s-extension create \
  --resource-group <rg> \
  --cluster-name <aks-name> \
  --cluster-type managedClusters \
  --name flux \
  --extension-type microsoft.flux \
  --config \
    workloadIdentity.enable=true \
    workloadIdentity.azureClientId=<user-assigned-client-id> \
    workloadIdentity.azureTenantId=<tenant-id>
```

**2) Create federated credential for Flux source controller:**

```bash
az identity federated-credential create \
  --name flux-source-controller \
  --identity-name <uai-name> \
  --resource-group <rg> \
  --issuer <aks-oidc-issuer-url> \
  --subject system:serviceaccount:flux-system:source-controller \
  --audience api://AzureADTokenExchange
```

**3) Create Flux configuration:**

```bash
az k8s-configuration flux create \
  --resource-group <rg> \
  --cluster-name <aks-name> \
  --cluster-type managedClusters \
  --name aks-gitops \
  --namespace flux-system \
  --scope cluster \
  --url https://github.com/<org>/<repo> \
  --branch main \
  --kustomization name=apps path=./clusters/prod prune=true
```

**4) Deploy by Git commit:** update image tags/manifests in Git; Flux reconciles and applies.

---

## 1. Workflow Structure Overview

```
.github/workflows/
  ci.yml          ← PR validation: lint, test, Docker build (no push)
  cd.yml          ← Main branch: push to ACR + deploy to AKS
  infra.yml       ← Bicep what-if on PR, deploy on main
```

---

## 2. CI Workflow (Pull Request)

```yaml
# .github/workflows/ci.yml
name: CI

on:
  pull_request:
    branches: [main]
    paths:
      - "backend/**"
      - "frontend/**"
      - ".github/workflows/ci.yml"

permissions:
  contents: read
  pull-requests: write

env:
  DOTNET_VERSION: "10.0.x"
  NODE_VERSION: "22"

jobs:
  backend-ci:
    name: Backend - Build and Test
    runs-on: ubuntu-latest
    timeout-minutes: 15
    steps:
      - uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1

      - name: Setup .NET ${{ env.DOTNET_VERSION }}
        uses: actions/setup-dotnet@v4
        with:
          dotnet-version: ${{ env.DOTNET_VERSION }}

      - name: Cache NuGet packages
        uses: actions/cache@v4
        with:
          path: ~/.nuget/packages
          key: ${{ runner.os }}-nuget-${{ hashFiles('**/packages.lock.json', '**/*.csproj') }}
          restore-keys: ${{ runner.os }}-nuget-

      - name: Restore
        run: dotnet restore backend/EmergencyAlerts.sln

      - name: Build
        run: dotnet build backend/EmergencyAlerts.sln --no-restore -c Release

      - name: Test
        run: |
          dotnet test backend/EmergencyAlerts.sln \
            --no-build -c Release \
            --logger "trx;LogFileName=results.trx" \
            --collect:"XPlat Code Coverage"

      - name: Publish test results
        uses: dorny/test-reporter@v1
        if: always()
        with:
          name: Backend Tests
          path: "**/results.trx"
          reporter: dotnet-trx

      - name: Validate EF Core migrations exist
        run: |
          find backend -path '*/Migrations/*.cs' -not -name '*Designer.cs' \
            | grep -q . \
            || { echo "❌ No EF Core migrations found — run dotnet ef migrations add first"; exit 1; }

      - name: Docker build — backend (no push on PR)
        run: |
          docker build \
            --no-cache \
            -t emergency-alerts-api:pr-${{ github.event.pull_request.number }} \
            -f backend/src/EmergencyAlerts.Api/Dockerfile \
            backend/

  frontend-ci:
    name: Frontend - Build and Test
    runs-on: ubuntu-latest
    timeout-minutes: 15
    steps:
      - uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1

      - name: Setup Node.js ${{ env.NODE_VERSION }}
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: npm
          cache-dependency-path: frontend/package-lock.json

      - name: Install dependencies
        working-directory: frontend
        run: npm ci

      - name: TypeScript check
        working-directory: frontend
        run: npx tsc --noEmit

      - name: Lint
        working-directory: frontend
        run: npx eslint . --max-warnings 0

      - name: Build (with empty API URL — PR validation only)
        working-directory: frontend
        run: npm run build
        env:
          VITE_API_URL: ""
          VITE_SIGNALR_URL: ""

      - name: Docker build — frontend (no push on PR)
        run: |
          docker build \
            --build-arg VITE_API_URL='' \
            --build-arg VITE_SIGNALR_URL='' \
            -t emergency-alerts-frontend:pr-${{ github.event.pull_request.number }} \
            -f frontend/Dockerfile \
            frontend/
```

---

## 3. CD Workflow (Main Branch → ACR + AKS) — Push-Based (Legacy/Bootstrap)

> **Note:** The AKS Book recommends GitOps (pull-based) for production clusters. Treat this
> push-based workflow as **legacy/bootstrapping** unless you have explicit constraints.

```yaml
# .github/workflows/cd.yml
name: CD

on:
  push:
    branches: [main]
    paths:
      - "backend/**"
      - "frontend/**"
      - "infrastructure/k8s/**"
      - ".github/workflows/cd.yml"

permissions:
  contents: read
  id-token: write # Required for OIDC / Workload Identity Federation

env:
  DOTNET_VERSION: "10.0.x"
  NODE_VERSION: "22"
  AZURE_SUBSCRIPTION_ID: ${{ vars.AZURE_SUBSCRIPTION_ID }}
  AZURE_TENANT_ID: ${{ vars.AZURE_TENANT_ID }}
  AZURE_CLIENT_ID: ${{ vars.AZURE_CLIENT_ID }}
  ACR_NAME: ${{ vars.ACR_NAME }}
  AKS_RESOURCE_GROUP: ${{ vars.AKS_RESOURCE_GROUP }}
  AKS_CLUSTER_NAME: ${{ vars.AKS_CLUSTER_NAME }}
  K8S_NAMESPACE: emergency-alerts

jobs:
  # 1. Build and push backend
  backend-build:
    name: Build & Push Backend
    runs-on: ubuntu-latest
    timeout-minutes: 15
    outputs:
      image-tag: ${{ steps.meta.outputs.version }}
    steps:
      - uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1

      - name: Azure login (OIDC)
        uses: azure/login@v2
        with:
          client-id: ${{ env.AZURE_CLIENT_ID }}
          tenant-id: ${{ env.AZURE_TENANT_ID }}
          subscription-id: ${{ env.AZURE_SUBSCRIPTION_ID }}

      - name: ACR login
        run: az acr login --name ${{ env.ACR_NAME }}

      - name: Validate EF Core migrations exist
        run: |
          find backend -path '*/Migrations/*.cs' -not -name '*Designer.cs' \
            | grep -q . \
            || { echo "❌ No EF Core migrations — failing build"; exit 1; }

      - name: Set image tag
        id: meta
        run: echo "version=${{ github.sha }}" >> "$GITHUB_OUTPUT"

      - name: Build & push backend
        run: |
          IMAGE="${{ env.ACR_NAME }}.azurecr.io/emergency-alerts-api:${{ steps.meta.outputs.version }}"
          docker build \
            --no-cache \
            -t "$IMAGE" \
            -f backend/src/EmergencyAlerts.Api/Dockerfile \
            backend/
          docker push "$IMAGE"

  # 2. Build and push frontend
  frontend-build:
    name: Build & Push Frontend
    runs-on: ubuntu-latest
    timeout-minutes: 15
    needs: [] # Independent of backend
    outputs:
      image-tag: ${{ steps.meta.outputs.version }}
    steps:
      - uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1

      - name: Azure login (OIDC)
        uses: azure/login@v2
        with:
          client-id: ${{ env.AZURE_CLIENT_ID }}
          tenant-id: ${{ env.AZURE_TENANT_ID }}
          subscription-id: ${{ env.AZURE_SUBSCRIPTION_ID }}

      - name: ACR login
        run: az acr login --name ${{ env.ACR_NAME }}

      - name: Set image tag
        id: meta
        run: echo "version=${{ github.sha }}" >> "$GITHUB_OUTPUT"

      - name: Build & push frontend (with API URL)
        run: |
          API_HOST="${{ vars.VITE_API_URL }}"
          IMAGE="${{ env.ACR_NAME }}.azurecr.io/emergency-alerts-frontend:${{ steps.meta.outputs.version }}"

          docker build \
            --no-cache \
            --build-arg VITE_API_URL="${API_HOST}" \
            --build-arg VITE_SIGNALR_URL="${API_HOST}" \
            -t "$IMAGE" \
            -f frontend/Dockerfile \
            frontend/

          # ⚠️ CRITICAL: Verify URL is baked into bundle before pushing
          if [[ -n "$API_HOST" ]]; then
            if ! docker run --rm "$IMAGE" sh -c "grep -qF '$API_HOST' /app/dist/assets/*.js 2>/dev/null"; then
              echo "❌ Bundle does not contain VITE_API_URL='$API_HOST' — aborting"
              exit 1
            fi
            echo "✅ Bundle correctly contains API URL"
          fi

          docker push "$IMAGE"

  # 3. Deploy to AKS
  deploy:
    name: Deploy to AKS
    runs-on: ubuntu-latest
    timeout-minutes: 20
    needs: [backend-build, frontend-build] # Both images must exist first
    environment: production
    steps:
      - uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1

      - name: Azure login (OIDC)
        uses: azure/login@v2
        with:
          client-id: ${{ env.AZURE_CLIENT_ID }}
          tenant-id: ${{ env.AZURE_TENANT_ID }}
          subscription-id: ${{ env.AZURE_SUBSCRIPTION_ID }}

      - name: Get AKS credentials
        run: |
          az aks get-credentials \
            --resource-group ${{ env.AKS_RESOURCE_GROUP }} \
            --name ${{ env.AKS_CLUSTER_NAME }} \
            --overwrite-existing

      - name: Validate K8s placeholders before apply
        env:
          POSTGRES_HOST: ${{ secrets.POSTGRES_HOST }}
          POSTGRES_USER: ${{ secrets.POSTGRES_USER }}
          POSTGRES_PASSWORD: ${{ secrets.POSTGRES_PASSWORD }}
          POSTGRES_DATABASE: ${{ secrets.POSTGRES_DATABASE }}
          CORS_ALLOWED_ORIGINS: ${{ vars.CORS_ALLOWED_ORIGINS }}
        run: |
          # Substitute and check for unresolved ${VAR} patterns
          for var in POSTGRES_HOST POSTGRES_USER POSTGRES_PASSWORD POSTGRES_DATABASE CORS_ALLOWED_ORIGINS; do
            if [[ -z "${!var}" ]]; then
              echo "❌ Required variable $var is empty — aborting deployment"
              exit 1
            fi
          done
          echo "✅ All required variables are set"

      - name: Apply K8s manifests
        env:
          POSTGRES_HOST: ${{ secrets.POSTGRES_HOST }}
          POSTGRES_USER: ${{ secrets.POSTGRES_USER }}
          POSTGRES_PASSWORD: ${{ secrets.POSTGRES_PASSWORD }}
          POSTGRES_DATABASE: ${{ secrets.POSTGRES_DATABASE }}
          CORS_ALLOWED_ORIGINS: ${{ vars.CORS_ALLOWED_ORIGINS }}
          API_IMAGE: ${{ env.ACR_NAME }}.azurecr.io/emergency-alerts-api:${{ needs.backend-build.outputs.image-tag }}
          FRONT_IMAGE: ${{ env.ACR_NAME }}.azurecr.io/emergency-alerts-frontend:${{ needs.frontend-build.outputs.image-tag }}
        run: |
          pwsh infrastructure/scripts/apply-k8s-manifests.ps1

      - name: Wait for rollout
        run: |
          kubectl rollout status deployment/emergency-alerts-api \
            -n ${{ env.K8S_NAMESPACE }} --timeout=5m
          kubectl rollout status deployment/emergency-alerts-frontend \
            -n ${{ env.K8S_NAMESPACE }} --timeout=5m

      - name: Post-deployment health check
        run: |
          API_URL="${{ vars.VITE_API_URL }}"
          for i in 1 2 3 4 5; do
            STATUS=$(curl -sf -o /dev/null -w "%{http_code}" "${API_URL}/health/ready" || echo "000")
            if [[ "$STATUS" == "200" ]]; then
              echo "✅ API health check passed (attempt $i)"
              break
            fi
            echo "⏳ Attempt $i: /health/ready returned $STATUS — retrying in 15s"
            sleep 15
          done
          [[ "$STATUS" == "200" ]] || { echo "❌ API health check failed after 5 attempts"; exit 1; }
```

---

## 4. Infrastructure Workflow (Bicep)

```yaml
# .github/workflows/infra.yml
name: Infrastructure

on:
  push:
    branches: [main]
    paths: ["infrastructure/bicep/**"]
  pull_request:
    paths: ["infrastructure/bicep/**"]

permissions:
  contents: read
  id-token: write
  pull-requests: write

jobs:
  validate:
    name: Validate Bicep
    runs-on: ubuntu-latest
    if: github.event_name == 'pull_request'
    steps:
      - uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1

      - uses: azure/login@v2
        with:
          client-id: ${{ vars.AZURE_CLIENT_ID }}
          tenant-id: ${{ vars.AZURE_TENANT_ID }}
          subscription-id: ${{ vars.AZURE_SUBSCRIPTION_ID }}

      - name: Check for preview APIs
        run: |
          if grep -rE '@[0-9]{4}-[0-9]{2}-[0-9]{2}-preview' infrastructure/bicep/; then
            echo "❌ Preview API versions detected — use stable versions only"
            exit 1
          fi

      - name: Build Bicep (lint)
        run: az bicep build --file infrastructure/bicep/main.bicep --outdir /tmp

      - name: What-if analysis
        run: |
          az deployment sub what-if \
            --location australiaeast \
            --template-file infrastructure/bicep/main.bicep \
            --parameters environment=dev

  deploy-infra:
    name: Deploy Infrastructure
    runs-on: ubuntu-latest
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1

      - uses: azure/login@v2
        with:
          client-id: ${{ vars.AZURE_CLIENT_ID }}
          tenant-id: ${{ vars.AZURE_TENANT_ID }}
          subscription-id: ${{ vars.AZURE_SUBSCRIPTION_ID }}

      - name: Deploy Bicep
        run: |
          az deployment sub create \
            --name "infra-$(date +%s)" \
            --location australiaeast \
            --template-file infrastructure/bicep/main.bicep \
            --parameters environment=prod
```

---

## 5. Go CI/CD Workflow

### 5a. Go CI (Build, Test, Lint, Vuln Scan)

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:

# Deny all permissions at workflow level; grant per job only what is needed.
permissions: {}

jobs:
  ci:
    name: Build · Test · Lint
    runs-on: ubuntu-latest
    timeout-minutes: 15
    permissions:
      contents: read
    steps:
      - uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683 # v4.2.2
        with:
          persist-credentials: false

      - uses: actions/setup-go@f111f3307d8850f501ac008e886eec1fd1932a34 # v5.3.0
        with:
          go-version-file: go.mod
          cache: true

      - name: Build
        run: go build ./...

      - name: Test (with race detector)
        run: go test -race -coverprofile=coverage.out ./...

      - name: golangci-lint
        uses: golangci/golangci-lint-action@4afd733a84b1f43292c63897423277bb7f4313a9 # v6.5.2
        with:
          version: latest

      - name: Vulnerability scan
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...
```

> **[VERIFY]** SHA pins shown here may be outdated. Verify current SHAs at
> `https://github.com/<owner>/<repo>/tags` for each action before committing.

### 5b. Go Release Workflow (Cross-Platform Build + SBOM)

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags: ["v*.*.*"]

permissions: {}

jobs:
  release:
    name: Build & Release
    runs-on: ubuntu-latest
    timeout-minutes: 20
    permissions:
      contents: write # required to create GitHub release assets
    steps:
      - uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683 # v4.2.2
        with:
          persist-credentials: false

      - uses: actions/setup-go@f111f3307d8850f501ac008e886eec1fd1932a34 # v5.3.0
        with:
          go-version-file: go.mod
          cache: true

      - name: Build cross-platform binaries
        run: |
          VERSION="${GITHUB_REF_NAME}"
          LDFLAGS="-X main.version=${VERSION} -s -w"
          mkdir -p dist
          for GOOS in linux darwin windows; do
            for GOARCH in amd64 arm64; do
              EXT=""
              [[ "$GOOS" == "windows" ]] && EXT=".exe"
              OUT="dist/my-extension-${GOOS}-${GOARCH}${EXT}"
              GOOS=$GOOS GOARCH=$GOARCH go build -ldflags "$LDFLAGS" -o "$OUT" .
              echo "Built $OUT"
            done
          done

      - name: Generate SBOM
        uses: anchore/sbom-action@f325610c9f50a54015d37c8d16cb3b0e2c8f4de # v0.18.0
        with:
          path: dist/
          artifact-name: sbom.spdx.json
          output-file: dist/sbom.spdx.json

      - name: Create GitHub Release
        uses: softprops/action-gh-release@c062e08bd532815e2082a85e87e3ef29c3e6d191 # v2.2.1
        with:
          files: dist/*
```

### 5c. `dependabot.yml` for Go + GitHub Actions

```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: gomod
    directory: /
    schedule:
      interval: weekly
    # Group all minor/patch updates into one PR
    groups:
      go-minor-patch:
        patterns: ["*"]
        update-types: ["minor", "patch"]

  - package-ecosystem: github-actions
    directory: /
    schedule:
      interval: weekly
    groups:
      actions-minor-patch:
        patterns: ["*"]
        update-types: ["minor", "patch"]
```

---

## 6. Key Rules & Anti-Patterns

### Job Dependency Order

```
Infrastructure deploy (bicep)
    ↓
Backend build + push  (parallel with)  Frontend build + push
    ↓ (both must succeed)
Deploy to AKS
    ↓
Health check validation
```

Always use `needs: [backend-build, frontend-build]` on the deploy job.

### DOTNET_VERSION Alignment

The workflow `DOTNET_VERSION: '10.0.x'` must match the `global.json` and `Dockerfile` base image (`mcr.microsoft.com/dotnet/sdk:10.0`). Mismatch causes silent test failures.

### Image Tag Strategy

- Use **commit SHA** as the authoritative tag (`${{ github.sha }}`).
- Keep deployment manifests pinned to immutable tags only.
- Never deploy `:latest` in production K8s manifests — always use the SHA tag via `needs.<job>.outputs.image-tag`.

### Secrets vs Vars

| Type        | GitHub Setting    | Use For                               |
| ----------- | ----------------- | ------------------------------------- |
| `secrets.*` | Encrypted secrets | Passwords, connection strings, tokens |
| `vars.*`    | Plain variables   | ACR names, resource groups, API URLs  |

---

## 6. Required GitHub Repository Settings

### Secrets (Settings → Secrets → Actions)

| Secret              | Description                            |
| ------------------- | -------------------------------------- |
| `POSTGRES_HOST`     | PostgreSQL Flexible Server hostname    |
| `POSTGRES_USER`     | PostgreSQL admin username              |
| `POSTGRES_PASSWORD` | PostgreSQL password (URL-encoded safe) |
| `POSTGRES_DATABASE` | Database name                          |

### Variables (Settings → Variables → Actions)

| Variable                | Description                          |
| ----------------------- | ------------------------------------ |
| `AZURE_SUBSCRIPTION_ID` | Azure subscription ID                |
| `AZURE_TENANT_ID`       | Azure tenant ID                      |
| `AZURE_CLIENT_ID`       | Managed identity client ID for OIDC  |
| `ACR_NAME`              | ACR name (without `.azurecr.io`)     |
| `AKS_RESOURCE_GROUP`    | AKS resource group name              |
| `AKS_CLUSTER_NAME`      | AKS cluster name                     |
| `VITE_API_URL`          | Public API DNS label URL             |
| `CORS_ALLOWED_ORIGINS`  | Comma-separated allowed CORS origins |

---

## 7. Workload Identity Federation (Keyless Auth)

No long-lived secrets in GitHub — use OIDC with Managed Identity.

```bash
# One-time setup (run locally or in setup script)
az identity create \
    -g "emergency-alerts-rg" \
    -n "github-actions-identity"

PRINCIPAL_ID=$(az identity show \
    -g "emergency-alerts-rg" \
    -n "github-actions-identity" \
    --query principalId -o tsv)

# Assign Contributor role (scoped to resource group)
az role assignment create \
    --assignee-object-id "$PRINCIPAL_ID" \
    --role Contributor \
    --scope "/subscriptions/${SUB_ID}/resourceGroups/emergency-alerts-rg"

# Federated credential for main branch
az identity federated-credential create \
    -g "emergency-alerts-rg" \
    --identity-name "github-actions-identity" \
    --name "github-main" \
    --issuer "https://token.actions.githubusercontent.com" \
    --subject "repo:OWNER/REPO:ref:refs/heads/main" \
    --audiences "api://AzureADTokenExchange"
```

---

## 8. Troubleshooting Common Failures

### ErrImagePull / ImagePullBackOff

Cause: Hardcoded ACR hostname or wrong tag in K8s manifest.

```bash
# Check what image the pod is trying to pull
kubectl describe pod -n emergency-alerts -l app=emergency-alerts-api \
    | grep -E "Image:|Failed|Error"
```

Fix: Ensure manifests use pipeline-injected `$ACR_NAME` variable, not a hardcoded hostname.

### Frontend shows blank page / ERR_NAME_NOT_RESOLVED

Cause: `VITE_API_URL` not injected at Docker build time.

```bash
# Check if URL is in the compiled bundle
docker run --rm myacr.azurecr.io/emergency-alerts-frontend:<sha-tag> \
    sh -c "grep -oE 'https?://[a-z0-9.-]+' /app/dist/assets/*.js | head -5"
```

Fix: Rebuild with `--build-arg VITE_API_URL=<url>` and `--no-cache`.

### Database migration failed (500 on all endpoints)

```bash
kubectl logs deployment/emergency-alerts-api -n emergency-alerts \
    | grep -E "FATAL|migration|Migration"
```

Common causes:

- Missing migration files (not committed before Docker build)
- Password contains `/` or `+` and is not URL-encoded
- PostgreSQL firewall blocks AKS LoadBalancer IP

### CORS preflight returns 403

```bash
# Test from outside the cluster
curl -i -X OPTIONS \
    -H "Origin: http://frontend-host.australiaeast.cloudapp.azure.com" \
    http://api-host.australiaeast.cloudapp.azure.com/api/v1/alerts \
    | grep -i "Access-Control"
```

Fix: Restart API pods after ConfigMap update:

```bash
kubectl rollout restart deployment/emergency-alerts-api -n emergency-alerts
```

---

## 9. ADAC Validation in CI/CD (REQUIRED)

Deployment pipelines are the last gate before runtime. They MUST validate that ADAC contracts are in place.

### Auto-Detect (CI stage)

- Verify health endpoints exist in source: `/health/ready` and `/health/live` routes must be registered.
- Check that EF Core migrations exist before Docker build (detect missing migration files).
- Validate environment variables are injected (no `${UNRESOLVED}` placeholders).

### Auto-Declare (CD stage — post-deployment)

- The existing `/health/ready` retry loop validates that the backend **declares** its readiness.
- Extend the health check to inspect the response body — a structured response with `status` and `entries` is preferred over a bare 200.
- If the health endpoint returns `Degraded`, treat it as a warning (log but don't fail) so partial deployments are visible.

### Auto-Communicate (CD stage — smoke test)

- After health passes, verify the frontend can reach the API (not just that the API pod is up).
- Validate that the compiled frontend bundle contains the correct `VITE_API_URL` (existing bundle verification step).
- If any post-deployment check fails, the pipeline output MUST state **what** failed and **why** (not just exit code 1).

### ADAC Checklist (CI/CD)

- [ ] Health endpoints validated in post-deployment step
- [ ] Health response body inspected (structured response, not just HTTP 200)
- [ ] Frontend bundle contains correct API URL
- [ ] Pipeline failure messages include specific reason, not just exit code
- [ ] Degraded health is logged as warning, not silently passed

---

## Checklist: Before Adding a New Workflow

### Universal (all workflows)

- [ ] `permissions: {}` empty block at **workflow** level; each job grants only what it needs
- [ ] ALL third-party `uses:` entries pinned to a full 40-char SHA with inline `# vX.Y.Z` comment
- [ ] `persist-credentials: false` on every `actions/checkout` step
- [ ] `timeout-minutes` set on every job (15 for build/test, 20 for deploy)
- [ ] `needs:` dependencies enforce Infrastructure → Build → Deploy order
- [ ] Secrets in `secrets.*`, plain config in `vars.*`
- [ ] `dependabot.yml` exists covering all package ecosystems + `github-actions`
- [ ] Release workflow generates and attaches SBOM (`anchore/sbom-action`, SHA-pinned)

### .NET / React specific

- [ ] `DOTNET_VERSION` matches `global.json` and Dockerfile base image
- [ ] `VITE_API_URL` passed as `--build-arg` with `--no-cache`
- [ ] Bundle verification step present for frontend builds
- [ ] EF Core migration existence check before backend Docker build
- [ ] Placeholder validation before `kubectl apply`
- [ ] Post-deployment `/health/ready` check with retry loop
- [ ] ADAC: health response body inspected, failure messages include specific reason

### Go specific

- [ ] `go test -race ./...` (race detector enabled)
- [ ] `golangci-lint run` step present
- [ ] `govulncheck ./...` step present
- [ ] `go-version-file: go.mod` used (not a hardcoded version string)
- [ ] Release workflow builds all required GOOS/GOARCH combinations

---

## References

- **yaml.instructions.md**: `.github/instructions/coding-standards/yaml/yaml.instructions.md`
- **docker.instructions.md**: `.github/instructions/coding-standards/docker/docker.instructions.md`
- **SPA-Endpoint-Configuration skill**: `.github/skills/spa-endpoint-configuration/SKILL.md`
- **managing-azure-dev-cli-lifecycle skill**: `.github/skills/managing-azure-dev-cli-lifecycle/SKILL.md`
- **GitHub Actions OIDC for Azure**: https://learn.microsoft.com/azure/developer/github/connect-from-azure
- **dotnet/sdk image tags**: https://hub.docker.com/_/microsoft-dotnet-sdk

---

## Node.js / TypeScript Variant

For Node.js/TypeScript backends, replace the .NET-specific steps with:

```yaml
# CI — Node.js build & test
jobs:
  build-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@<pin-to-sha>
      - uses: actions/setup-node@<pin-to-sha>
        with:
          node-version-file: ".nvmrc"
          cache: "npm" # or 'pnpm' if using pnpm
      - run: npm ci
      - run: npm run lint
      - run: npm test -- --coverage
      - run: npm run build

  docker-build:
    needs: build-test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@<pin-to-sha>
      - run: |
          docker build -t ${{ vars.ACR_NAME }}.azurecr.io/${{ vars.IMAGE_NAME }}:${{ github.sha }} \
            --build-arg NODE_ENV=production .
```

**Key differences from .NET:**

- Use `.nvmrc` for version pinning (equivalent to `global.json`)
- `npm ci` (not `npm install`) for deterministic builds
- Pin actions to SHA, same as .NET workflows
- For pnpm: replace `cache: 'npm'` with `cache: 'pnpm'` and `npm ci` with `pnpm install --frozen-lockfile`

---

## Currency and verification

- **Date checked:** 2026-03-31
- **Sources:** [Node.js releases](https://nodejs.org/en/about/releases), [.NET release schedule](https://dotnet.microsoft.com/platform/support/policy/dotnet-core), [GitHub Actions runner images](https://github.com/actions/runner-images/releases)
- **Versions verified:** .NET 10 SDK (GA/LTS), Node.js 22 (Maintenance LTS), Node.js 24 (Active LTS as of May 2025)
- **Verification steps:** Check [Node.js releases page](https://nodejs.org/en/about/releases) for current Active LTS; run `node --version` and `dotnet --version` in the workflow.

> **Node.js 22 note:** Node.js 22 is currently **Maintenance LTS** (active through April 2027, but no new features). Node.js 24 is now **Active LTS** (recommended for new projects). If using `NODE_VERSION: "22"` today, it will continue to work but plan a migration to 24 before April 2027.

### Known pitfalls

| Area                               | Pitfall                                                                                                                                                                            | Mitigation                                                                                                                                                   |
| ---------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `DOTNET_VERSION` mismatch          | `global.json` SDK version, Dockerfile base image, and `DOTNET_VERSION` env var diverge → silent test failures or build errors                                                      | Keep all three in sync; run `dotnet --version` as a CI step to surface drift early                                                                           |
| Node.js LTS drift                  | Using Maintenance LTS instead of Active LTS means missing performance/security improvements                                                                                        | Upgrade `NODE_VERSION` when the current version moves from Active LTS to Maintenance LTS                                                                     |
| `npm install` vs `npm ci`          | `npm install` can silently update `package-lock.json`; `npm ci` enforces the lockfile                                                                                              | Always use `npm ci` in CI; never use `npm install`                                                                                                           |
| Action version pinning             | Using mutable tags (e.g., `@v4`) lets an upstream attacker hijack the action                                                                                                       | Pin all third-party actions to a full SHA (`@<sha> # v4.x.y`)                                                                                                |
| Non-checkout actions in this skill | Examples for `actions/setup-dotnet`, `actions/setup-node`, `actions/cache`, `azure/login`, and `dorny/test-reporter` use tag-only references (`@v4`, `@v2`, `@v1`) for readability | Before using in real workflows, replace with SHA-pinned versions per `cicd-security.instructions.md` (look up current SHAs from each action's releases page) |
| `pull_request_target`              | Workflows using `pull_request_target` with code checkout can expose secrets to fork PRs                                                                                            | Prefer `pull_request`; only use `pull_request_target` when write access is explicitly required                                                               |

---

## Related Skills

- [GitHub Actions Terraform](../github-actions-terraform/SKILL.md) — Terraform-specific CI/CD workflows
- [Managing Azure Dev CLI Lifecycle](../managing-azure-dev-cli-lifecycle/SKILL.md) — azd-based deployment pipelines
- [SPA Endpoint Configuration](../spa-endpoint-configuration/SKILL.md) — Build-time variable injection in CI
