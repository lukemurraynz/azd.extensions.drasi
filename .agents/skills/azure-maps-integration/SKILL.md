---
name: azure-maps-integration
description: >-
  Production-ready Azure Maps implementation patterns for React/TypeScript frontends and API-backed architectures. USE FOR: auth model selection (Entra ID vs subscription key), Web SDK map-control setup, RBAC role selection, secure deployment guidance.
---

# Azure Maps Integration Patterns

Security-first patterns for Azure Maps integration across frontend and backend systems.

---

## When to Use This Skill

- Adding Azure Maps to a React/TypeScript UI
- Choosing authentication mode for Maps in dev/test/prod
- Selecting least-privilege Azure Maps RBAC roles
- Avoiding client-side secret leakage and auth drift
- Validating Azure Maps config before deployment

---

## 1. Authentication Decision Matrix (REQUIRED)

| Environment / Scenario       | Recommended auth                                           | Why                                                     |
| ---------------------------- | ---------------------------------------------------------- | ------------------------------------------------------- |
| Production web app           | Microsoft Entra ID (`authType: 'aad'`)                     | Identity-based auth, avoids long-lived client-side keys |
| Local dev quickstart / demos | Subscription key (`authType: 'subscriptionKey'`)           | Fast setup when identity plumbing is not available      |
| Server-to-server API calls   | Microsoft Entra token (managed identity/service principal) | Better key hygiene, policy control, auditable access    |

**Rule:** Treat subscription-key auth as a bootstrap option, not the long-term production default for browser clients.

---

## 2. Web SDK Initialization Patterns

### Microsoft Entra ID (Preferred for production)

```ts
import * as atlas from "azure-maps-control";

const map = new atlas.Map("myMap", {
  center: [-122.33, 47.6],
  zoom: 12,
  language: "en-US",
  authOptions: {
    authType: "aad",
    clientId: "<Your Microsoft Entra Client ID>",
    aadAppId: "<Your Microsoft Entra App ID>",
    aadTenant: "<Your Microsoft Entra tenant ID>",
  },
});
```

### Subscription key (Dev/quickstart only)

```ts
import * as atlas from "azure-maps-control";

const map = new atlas.Map("myMap", {
  center: [-122.33, 47.6],
  zoom: 12,
  authOptions: {
    authType: "subscriptionKey",
    subscriptionKey: "<Your Azure Maps Key>",
  },
});
```

> [!WARNING]
> Subscription keys in browser code are visible in DevTools. For development: rotate keys every 7 days. For production: use Entra ID authentication exclusively and disable subscription key access via Azure Portal. If subscription keys must be used temporarily, restrict them via CORS origin policies and monitor usage via Azure Maps metrics.

```javascript
// Required: handle auth and load errors
map.events.add('error', (e) => {
  console.error('Azure Maps error:', e.error?.message);
  // Show user-friendly fallback; do not expose error details
});
```

---

## 3. React Pattern (`react-azure-maps`)

```tsx
import AzureMapsReact from "react-azure-maps";
import { AuthenticationType } from "azure-maps-control";

function AlertMap() {
  const mapKey = import.meta.env.VITE_AZURE_MAPS_KEY as string;

  const options = {
    authOptions: {
      authType: AuthenticationType.subscriptionKey,
      subscriptionKey: mapKey,
    },
    center: [174.7633, -36.8485],
    zoom: 6,
  };

  return <AzureMapsReact options={options} style={{ height: 400 }} />;
}
```

**Important:** if using subscription key in browser code, scope it to non-production contexts and rotate frequently.

---

## 4. RBAC Roles (Least Privilege)

Common built-in roles:

- **Azure Maps Data Reader** (`423170ca-a8f6-4b0f-8487-9e4eb8f49bfa`) — read map-related data
- **Azure Maps Search and Render Data Reader** (`6be48352-4f82-47c9-ad5e-0acacefdb005`) — minimal render + search access for common web scenarios
- **Azure Maps Data Contributor** (`8f5e0ce6-4f7b-4dcf-bddf-e6f48634a204`) — read/write map data

Start with reader-level roles and elevate only with explicit justification.

---

## 5. Security Guardrails

- Do not hardcode production Maps keys in frontend source.
- Do not store Maps secrets in git-tracked files.
- Prefer identity-based auth for production browser workloads.
- If a backend calls Maps APIs, prefer managed identity over shared keys.
- Keep Maps auth and role assignments explicit in deployment documentation.

---

## 6. Validation Checklist

- [ ] Auth mode selected intentionally per environment (prod vs dev)
- [ ] Production path uses Entra ID or approved equivalent identity flow
- [ ] RBAC scoped to least privilege
- [ ] Frontend build/runtime config does not expose privileged secrets
- [ ] Role IDs and permissions validated against current Azure docs

---

## References

- Azure Maps authentication: https://learn.microsoft.com/azure/azure-maps/azure-maps-authentication
- Azure Maps authentication best practices: https://learn.microsoft.com/azure/azure-maps/authentication-best-practices
- Azure Maps map control: https://learn.microsoft.com/azure/azure-maps/how-to-use-map-control
- Azure RBAC built-in roles (web/mobile): https://learn.microsoft.com/azure/role-based-access-control/built-in-roles/web-and-mobile
- react-azure-maps: https://github.com/Azure/react-azure-maps

---

## Known Pitfalls

| Pitfall                                                                | Detail                                                                                              |
| ---------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------- |
| Subscription key in production browser code                            | Keys are visible in browser DevTools; use Entra ID auth for production web apps                     |
| Stale `react-azure-maps` peer deps                                     | Community-maintained; verify compatibility with current `azure-maps-control` before upgrading       |
| RBAC role confusion (`Data Reader` vs `Search and Render Data Reader`) | `Search and Render Data Reader` is narrower (render + search only); prefer it for read-only web UIs |

## Currency and Verification

- **Date checked:** 2026-03-31
- **Evidence:** Microsoft Learn RBAC built-in roles docs + Azure Maps authentication docs (verified via MCP)

| Item                                          | Verified version / value               | Status  |
| --------------------------------------------- | -------------------------------------- | ------- |
| Azure Maps Data Reader GUID                   | `423170ca-a8f6-4b0f-8487-9e4eb8f49bfa` | Current |
| Azure Maps Search and Render Data Reader GUID | `6be48352-4f82-47c9-ad5e-0acacefdb005` | Current |
| Azure Maps Data Contributor GUID              | `8f5e0ce6-4f7b-4dcf-bddf-e6f48634a204` | Current |
| Auth patterns (Entra ID + subscription key)   | Per Azure Maps authentication docs     | Current |
