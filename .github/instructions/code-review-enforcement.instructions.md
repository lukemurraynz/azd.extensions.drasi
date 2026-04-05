---
applyTo: "**/*.cs,**/*.bicep,**/k8s/**/*.yml,**/k8s/**/*.yaml,**/kubernetes/**/*.yml,**/kubernetes/**/*.yaml,**/*.ts,**/*.tsx,**/*.js,**/*.jsx,**/Dockerfile,**/Dockerfile.*,**/*.dockerfile,.github/workflows/**"
description: Per-technology code review enforcement rules for PRs covering ASP.NET Core, Bicep, Kubernetes, TypeScript/React, Docker, and CI/CD workflows
---

# Code Review Enforcement Rules

These rules apply to all PRs. Each section maps to a technology and its corresponding coding-standards instruction file. Use these checklists during implementation and code review.

## 1. ASP.NET Core Apps (C#)

- [ ] Health check endpoints `/health/ready` and `/health/live` mapped with JSON responses (if Kubernetes target)
- [ ] Authorization middleware bypasses `/health/*` paths (case-insensitive)
- [ ] Authentication configured when using `HttpContext.User` / `[Authorize]` (AddAuthentication + UseAuthentication)
- [ ] No `[AllowAnonymous]` on mutating endpoints; demo endpoints gated to non-prod
- [ ] `Auth__AllowAnonymous` is false/omitted in production config
- [ ] ETag/`If-Match` enforced when supported
- [ ] Rate-limit/auth bypass headers restricted to dev/test only
- [ ] Token-protected callbacks fail closed when token is missing in non-dev
- [ ] InMemory DB fallback disabled in production
- [ ] Rate limiting is cluster-aware (no static in-memory store without eviction)
- [ ] Error contract is consistent: RFC 9457 Problem Details (`application/problem+json`) + `x-error-code` header + `errorCode` extension (and `x-error-code` matches `errorCode`)
- [ ] API versioning uses required `api-version=YYYY-MM-DD[-preview]` query param (no version in path); missing/unsupported returns `400` with `x-error-code` set to `MissingApiVersionParameter` / `UnsupportedApiVersionValue`
- [ ] List pagination uses `{ "value": [...], "nextLink": "https://..." }` (absolute `nextLink`, omitted on last page, never null; avoid global counts by default)
- [ ] LRO uses `202 Accepted` + `operation-location` (absolute URL) for polling; `api-version` included in poll URL
- [ ] LRO status monitor responses (GET `operation-location`) include `Retry-After` when not terminal
- [ ] POST create/actions are idempotent; if retries are possible, support `Repeatability-Request-ID` + `Repeatability-First-Sent` (or an explicit idempotency mechanism)
- [ ] Expensive clients (e.g., Azure SDK clients) are reused via DI
- Reference: `.github/instructions/coding-standards/csharp/csharp.instructions.md`

## 2. Bicep Infrastructure Code

- [ ] No unused parameters (checked by `az bicep build`)
- [ ] No preview API versions unless explicitly justified in PR comment
- [ ] API versions verified to exist in `az provider` output or learn.microsoft.com
- [ ] Deterministic resource naming (avoid hardcoded names like "config" or "aks-federation")
- [ ] CI/CD runs `az bicep build` with warning-as-error enforcement
- Reference: `.github/instructions/coding-standards/bicep/bicep.instructions.md`

## 3. Kubernetes Manifests

- [ ] `imagePullPolicy: Always` for all containers (prevents stale image caching)
- [ ] Liveness and readiness probes configured (endpoints must bypass auth)
- [ ] Resource requests/limits defined for CPU and memory
- [ ] No hardcoded image tags (use semantic versioning or commit SHA)
- [ ] Ingress/Gateway is the only public entry point unless a second one is explicitly justified
- [ ] NetworkPolicies do not allow broad `0.0.0.0/0` egress without documented rationale
- Reference: `.github/instructions/coding-standards/kubernetes/kubernetes-deployment-best-practices.instructions.md`

## 4. TypeScript/React Frontend

- [ ] API calls wrapped in error handlers; failures don't crash app
- [ ] Discriminated union state for loading/success/error (not boolean flags)
- [ ] Fallback UI shown when API unavailable (not blank screen)
- [ ] localStorage or cache used for offline scenarios where applicable
- [ ] Client health endpoints match backend contract (`/health/ready` and `/health/live` if present)
- [ ] Correlation IDs use `crypto.randomUUID()` (not `Math.random`)
- [ ] ESLint/Prettier integration is wired if Prettier is installed
- [ ] Client response parsing matches a single API contract shape
- [ ] When present, API failures surface/log `x-error-code` + correlation/trace id and parse RFC 9457 Problem Details consistently (including `errorCode`/`traceId` extensions if present)
- [ ] Auth tokens are not stored in `localStorage`/`sessionStorage` (prefer in-memory or `HttpOnly` cookies)
- [ ] LRO polling respects `Retry-After` when present
- [ ] **CRITICAL - Multi-Host Deployments**: If frontend and API are on separate hosts:
  - Frontend **must be built** with explicit `--build-arg VITE_API_URL=<api-host>`
  - Verify compiled bundle contains the expected URLs: `grep -qF "$API_HOST" dist/assets/*.js`
  - If missing: **DO NOT DEPLOY** — URLs are baked at build time; restarting containers won't fix wrong URLs
  - If JSON parse error (`Unexpected token '<'`): frontend is calling itself (SPA) instead of backend; rebuild with correct base URL
- Reference: `.github/instructions/coding-standards/typescript/typescript.instructions.md`

## 5. Docker Builds

- [ ] Multi-stage builds with minimal runtime layer
- [ ] No build secrets exposed in final image (use BuildKit secrets)
- [ ] Base image tag is specific version (not `latest`)
- [ ] Dockerfile tested locally: `docker build --no-cache`
- Reference: `.github/instructions/coding-standards/docker/docker.instructions.md`

## 6. CI/CD Workflows

- [ ] Linting steps run BEFORE build (fail-on-error: true)
- [ ] Tests run with coverage collection
- [ ] Docker build includes security scanning
- [ ] Deployment approval gate requires manual review before prod
- [ ] SDK/tooling versions match project targets (for .NET, DOTNET_VERSION aligns with TFM/global.json)
- Reference: `.github/instructions/coding-standards/yaml/yaml.instructions.md`
