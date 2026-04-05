---
name: spa-endpoint-configuration
description: >-
  Configure API and WebSocket endpoints in Vite/React SPAs without recurring build-time injection mistakes. USE FOR: fix ERR_NAME_NOT_RESOLVED, configure VITE_API_URL, set up SPA Docker build args, validate frontend bundle URLs, troubleshoot wrong API endpoint, rebuild frontend with correct URLs.
---

# SPA Endpoint Configuration

> **Canonical source** for Vite/SPA build-time environment variable injection patterns. Other skills and instructions reference this skill for the full Docker build-arg workflow.

## Problem Solved

Vite-based SPAs (React, Vue, Svelte) use build-time environment variable injection. Developers often hardcode URLs in Dockerfile `ARG` defaults, or forget to pass `VITE_*` build args, causing `ERR_NAME_NOT_RESOLVED` errors when deployed to different environments. Restarting pods doesn't fix this—the wrong URL is baked into compiled JS.

This skill ensures correct, reusable SPA endpoint configuration across dev, staging, and production.

---

## When to Use This Skill

- **Creating a new Vite/React/SPA project** → Use to set up endpoint injection from the start
- **Encountering `ERR_NAME_NOT_RESOLVED` in browser** → Use to diagnose and fix hardcoded URLs
- **Deploying to multiple environments** → Use to ensure each environment gets the correct API/WebSocket URLs
- **Configuring Docker image builds** → Use to enforce zero hardcoded service hostnames in Dockerfile

---

## Standard Patterns

### Pattern A: Relative Paths (Ingress Pattern)

**When:** Frontend and API share the same public hostname (behind ingress).

**Setup:**

- Dockerfile: \`ARG VITE_API_URL=\` and \`ARG VITE_SIGNALR_URL=\` (empty)
- Build: \`docker build --build-arg VITE_API_URL="" --build-arg VITE_SIGNALR_URL="" ...\`
- Ingress routes \`/api/\*\` → backend, \`/\*\` → frontend
- Frontend code uses relative paths: \`GET /api/v1/alerts\`

### Pattern B: Absolute URLs (LoadBalancer Pattern)

**When:** Frontend and API have separate public IPs/DNS names (no ingress).

**Setup:**

- Dockerfile: \`ARG VITE_API_URL=\` and \`ARG VITE_SIGNALR_URL=\` (empty)
- Build: \`docker build --build-arg VITE_API_URL="http://api-dns.region.cloudapp.azure.com:8080" --build-arg VITE_SIGNALR_URL="http://api-dns.region.cloudapp.azure.com:8080" ...\`
- K8s Services: LoadBalancer type with public IPs assigned
- Frontend code uses injected base URL: \`GET http://api-dns.../api/v1/alerts\`

### Pattern C: Runtime Config (Dynamic Endpoints)

**When:** Endpoints must change without rebuilding (rare).

**Setup:**

- Serve \`/config.json\` with \`{ "apiUrl": "...", "signalrUrl": "..." }\` from frontend host
- Frontend fetches \`/config.json\` on app boot
- Override \`import.meta.env.VITE_API_URL\` with runtime config
- Requires app code support; verify before claiming this works

---

## Checklist: Prevent Hardcoded URLs

### Dockerfile Review

- [ ] \`ARG VITE_API_URL=\` is **empty** (no default value)
- [ ] \`ARG VITE_SIGNALR_URL=\` is **empty** (no default value)
- [ ] No other \`VITE\_\*\` args have hardcoded service hostnames
- [ ] \`ENV VITE_API_URL=$VITE_API_URL\` passes the arg value
- [ ] \`ENV VITE_SIGNALR_URL=$VITE_SIGNALR_URL\` passes the arg value
- [ ] \`RUN npm run build\` happens AFTER ARG/ENV

### CI/CD / Build Script Review

- [ ] \`--build-arg VITE_API_URL="<actual-value>"\` is in docker build command
- [ ] \`--build-arg VITE_SIGNALR_URL="<actual-value>"\` is in docker build command
- [ ] Build script validates bundle contains the correct URL(s):
      \`\`\`bash
      grep -qF "$VITE_API_URL" dist/assets/*.js || exit 1
  grep -qF "$VITE_SIGNALR_URL" dist/assets/\*.js || exit 1
      \`\`\`
- [ ] Fails loudly if URL is missing or wrong (prevents silent failures)

### Deployment Review

- [ ] K8s frontend Service has correct LoadBalancer IP or Ingress routing
- [ ] DNS labels/hostnames match what was passed at build time
- [ ] Health checks verify frontend loads and makes API calls

### Testing

- [ ] Open frontend in browser
- [ ] Press F12 → Network tab
- [ ] Verify API requests go to correct hostname (`http://actual-api.../api/...`, not `http://api.emergency-alerts.local/...`)
- [ ] No \`ERR_NAME_NOT_RESOLVED\` in console

---

## ADAC Runtime Health (Auto-Detect, Auto-Declare, Auto-Communicate)

**Goal:** Make live data health visible and actionable in the UI.

### Auto-Detect

- Track last-successful API/SSE/SignalR update timestamp.
- Detect connection state (connected, reconnecting, disconnected).
- Flag data as stale after a fixed threshold (for example, 60 seconds).

### Auto-Declare

- Show a visible indicator for connection state and data freshness.
- Expose the same state in a lightweight health/status endpoint (if the SPA has a backend).

### Auto-Communicate

- Display a degraded mode banner with a specific reason (stale data, disconnected, auth failure).
- Avoid silent fallbacks; users should see reduced fidelity.

### ADAC Checklist (Frontend)

- [ ] Last update timestamp stored and rendered
- [ ] Connection state is visible in the UI
- [ ] Stale data threshold is defined and tested
- [ ] Degraded mode message is actionable and specific

---

## Troubleshooting

### **Browser shows `ERR_NAME_NOT_RESOLVED` for all API calls**

**Root cause:** Frontend was built with wrong `VITE_API_URL` or `VITE_SIGNALR_URL`.

**Diagnosis:**

- Open frontend → F12 → Network tab → observe API URL
- Check if URL matches your environment (should be DNS name or IP, not \`localhost\` or \`api.emergency-alerts.local\`)

**Fix:**

1. Identify correct API endpoint: \`http://api-dns.region.cloudapp.azure.com:8080\`
2. Rebuild frontend: \`docker build --build-arg VITE_API_URL="http://api-dns.region.cloudapp.azure.com:8080" --build-arg VITE_SIGNALR_URL="http://api-dns.region.cloudapp.azure.com:8080" -t registry/frontend:new-tag frontend/\`
3. Verify bundle: \`docker run --rm registry/frontend:new-tag grep -o 'http://api-dns[^"]_' dist/assets/_.js\`
4. Push and redeploy: \`docker push registry/frontend:new-tag && kubectl set image ...\`

### **API health endpoints work, but list/create endpoints fail**

**Root cause:** Often CORS or Auth headers. Verify:

- API health check bypasses CORS/Auth (\`/health/\*\` routes)
- Regular API routes have correct CORS headers
- Token/auth headers are included in requests

### **Relative paths not working (Ingress pattern)**

**Root cause:** Ingress routing misconfigured, or frontend using absolute URL.

**Fixes:**

- Verify ingress routes \`/api/\*\` to backend service
- Ensure frontend uses relative paths (\`/api/...\`, not \`http://...\`)
- Check \`apiBaseUrl.ts\` returns empty string when \`VITE_API_URL\` is empty

---

## Examples

### Rebuild Frontend with DNS-Based API URL

\`\`\`bash

# Get the API DNS label assigned to LoadBalancer IP

$apiDns = "api-emergency-alerts-dev.australiaeast.cloudapp.azure.com"
$apiPort = "8080"
$apiUrl = "http://$apiDns:$apiPort"

# Example immutable image tag
export IMAGE_TAG=$(git rev-parse --short HEAD)

# Rebuild with explicit VITE_API_URL

docker build \
 -f frontend/Dockerfile \
 --build-arg VITE_API_URL="$apiUrl" \
  --build-arg VITE_SIGNALR_URL="$apiUrl" \
 -t emergencealerts202602171637.azurecr.io/frontend:${IMAGE_TAG} \
 frontend/

# Validate bundle

docker run --rm emergencealerts202602171637.azurecr.io/frontend:${IMAGE_TAG} \
 grep -F "$apiUrl" dist/assets/\*.js && echo "✅ URL correctly injected"

# Push and redeploy

docker push emergencealerts202602171637.azurecr.io/frontend:${IMAGE_TAG}

kubectl rollout restart deployment/emergency-alerts-frontend -n emergency-alerts
\`\`\`

### CI/CD Validation Step

\`\`\`yaml

- name: Validate Frontend Build
  run: |
  if [[-z "$VITE_API_URL"]]; then
  echo "⚠️ WARNING: VITE*API_URL is empty (relative paths will be used)"
  else
  echo "🔍 Checking if bundle contains $VITE_API_URL..."
      if ! grep -qF "$VITE_API_URL" dist/assets/*.js; then
  echo "❌ ERROR: Built bundle does not contain VITE*API_URL=$VITE_API_URL"
        echo "Compiled JS might use incorrect API endpoint!"
        exit 1
      fi
      echo "✅ Bundle correctly contains API URL"
    fi
    if [[ -n "$VITE_SIGNALR_URL" ]]; then
  echo "🔍 Checking if bundle contains $VITE_SIGNALR_URL..."
      if ! grep -qF "$VITE_SIGNALR_URL" dist/assets/*.js; then
  echo "❌ ERROR: Built bundle does not contain VITE_SIGNALR_URL=$VITE_SIGNALR_URL"
  exit 1
  fi
  echo "✅ Bundle correctly contains SignalR URL"
  fi
  \`\`\`

---

## References

- **Vite Environment Variables:** https://vitejs.dev/guide/env-and-mode.html
- **React + Vite:** https://react.dev/learn/add-react-to-existing-project
- **ISE Instruction:** [.github/instructions/typescript.instructions.md](../../instructions/typescript.instructions.md) → Vite/SPA Build-Time Variables section
- **Docker Instruction:** [.github/instructions/docker.instructions.md](../../instructions/docker.instructions.md) → Vite/SPA Build-Time Environment Variables

---

## Next Steps

1. **Review your Dockerfile** → Ensure all \`VITE\_\*\` args have empty defaults
2. **Update CI/CD build command** → Add explicit \`--build-arg VITE_API_URL=...\` and \`--build-arg VITE_SIGNALR_URL=...\`
3. **Add validation step** → Verify bundle contains correct URL before push
4. **Test deployment** → Open frontend, verify API calls go to correct host
5. **Document your pattern** → README or DEPLOYMENT.md for your team

---

## Currency

- **Date checked:** 2026-03-31
- **Sources:** [Vite Environment Variables](https://vitejs.dev/guide/env-and-mode.html), Microsoft Learn MCP
- **Authoritative references:** [Vite Env and Mode](https://vitejs.dev/guide/env-and-mode.html), Docker multi-stage build docs

### Verification Steps

1. Confirm Vite `import.meta.env` prefix behavior has not changed
2. Verify Docker multi-stage build ARG/ENV propagation behavior
3. Check for new Vite runtime config patterns that may supersede build-time injection

---

## Related Skills

- [TypeScript React Patterns](../typescript-react-patterns/SKILL.md) — Frontend code consuming these endpoints
- [Kubernetes CORS Configuration](../kubernetes-cors-configuration/SKILL.md) — CORS for containerized deployments
- [.NET Backend Patterns](../dotnet-backend-patterns/SKILL.md) — API backend endpoint patterns
