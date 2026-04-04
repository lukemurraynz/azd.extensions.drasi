---
applyTo: "**/*.ts,**/*.tsx,**/*.js,**/*.jsx"
description: "TypeScript/JavaScript + React + Fluent UI v9: correctness-first, performance-aware, production-build truthful"
---

# Vite/SPA Build-Time Environment Variables

**For Vite, React, and other SPA frameworks:**

## Default Vite behavior (non-negotiable)

- Environment variables (e.g., \`VITE_API_URL\`) are injected at **build time**, not at runtime.
- Only variables prefixed with \`VITE\_\` are exposed to client code via \`import.meta.env\`.
- Do **NOT** expect runtime environment variables or ConfigMaps to affect the built bundle.
- Changing pod/container env vars **after build** will **NOT** update the frontend’s API endpoints embedded in the bundle.
- Reference: https://vitejs.dev/guide/env-and-mode.html

## Preferred approach: build-time injection

- Always pass required values at **image build time**, not container runtime.
- Example: \`docker build --build-arg VITE_API_URL=...\`
- Client reads: \`import.meta.env.VITE_API_URL\`
- If SignalR is hosted on a separate public endpoint, inject `VITE_SIGNALR_URL` at build time.

## Allowed escape hatch: runtime config (only when rebuilds are not acceptable)

If endpoints/settings must change without rebuilding:

- Serve a runtime config file (e.g. \`/config.json\` or \`/config.js\`) from the web server/container at startup.
- Fetch it once on app boot and cache in memory (optionally \`sessionStorage\`).
- Document precedence if both runtime config and \`VITE\_\*\` exist (avoid ambiguous “fallback soup”).
- Do **not** claim ConfigMaps work unless the app explicitly reads runtime config at runtime.

**See also:** Docker instructions for multi-stage builds and CI/CD guidance.

---

# TypeScript/JavaScript Copilot Instructions

Follow the ISE JavaScript/TypeScript Code Review Checklist and modern JavaScript best practices.

## Authority and conflicts

- **Source of truth**: repo lint + formatting rules, plus ISE guidance; this file is a repo-specific overlay.
- If these instructions conflict with lint rules or ISE guidance, prefer in this order:
  1. Project lint configuration (ESLint, TypeScript, Prettier)
  2. ISE guidance
  3. This file (open a follow-up PR to reconcile)
- Always use the \`iseplaybook\` MCP server for the latest TypeScript guidance and \`context7\` for React/Node/framework specifics; do not guess, verify.

## General code style

- Use 2-space indentation (match repo).
- Prefer \`const\` over \`let\`; avoid \`var\`.
- Prefer arrow functions for callbacks and inline functions.
- Use template literals for string interpolation and multi-line strings.
- Follow the existing semicolon style; ESLint is the arbiter.
  - Ensure Prettier and ESLint configs are aligned to avoid conflicts.
- For client-side correlation IDs (UI diagnostics), prefer `crypto.randomUUID()` over timestamp + `Math.random()`.

## Modern module + bundle posture (performance-aware)

- Prefer ESM \`import/export\` in frontend code. Avoid \`require()\` in the app bundle.
- Avoid “accidentally importing the world”:
  - Prefer specific imports over whole-library imports where it affects bundle size.
  - Avoid adding heavyweight deps for small utilities (prefer platform APIs / small utilities already in the repo).
- Treat bundle-size regressions as bugs:
  - Code-split large routes/features via dynamic import when appropriate.
  - Avoid barrel re-exports for very large modules if it harms tree-shaking in your setup.

---

# TypeScript best practices

## Types and configuration

- Align to the repo’s \`tsconfig\*.json\` first.
- Prefer enabling \`"strict": true\`.
- Prefer these safety flags **if compatible with the repo**:
  - \`"noUncheckedIndexedAccess": true\`
  - \`"exactOptionalPropertyTypes": true\`
  - \`"noImplicitOverride": true\`
  - \`"verbatimModuleSyntax": true\`

\\_Do not hard-code “2026 baseline” assumptions. Prefer: “read the repo config, then suggest improvements.”\\_

Example snippet (illustrative only; verify against repo standards):
\\\`\\\`\\\`jsonc
{
"compilerOptions": {
"strict": true,
"noUncheckedIndexedAccess": true,
"exactOptionalPropertyTypes": true,
"noImplicitOverride": true,
"verbatimModuleSyntax": true
}
}
\\\`\\\`\\\`

Prefer \`.ts/.tsx\` for new code; for legacy \`.js/.jsx\`, enable \`checkJs\` where feasible and migrate incrementally.

## Strict mode implementation patterns (common pitfalls)

### verbatimModuleSyntax: type-only imports

When \`verbatimModuleSyntax\` is enabled, always use \`import type\` for type-only imports:

\\\`\\\`\\\`ts
// ❌ WRONG
import { User, ApiError } from "./types";

// ✅ CORRECT
import type { User, ApiError } from "./types";

// ✅ Mixed imports
import { apiCall } from "./api";
import type { User } from "./types";

// ✅ Inline type import
import { type Alert, createAlert } from "./services";
\\\`\\\`\\\`

### exactOptionalPropertyTypes: optional props

With \`exactOptionalPropertyTypes\`, optional properties cannot receive \`undefined\` explicitly unless the type includes it. Prefer omitting:

\\\`\\\`\\\`tsx
// ❌ WRONG: error is string | undefined
<Field label="Name" validationMessage={error} />

// ✅ CORRECT: conditional spreading
<Field label="Name" {...(error ? { validationMessage: error } : {})} />

// ✅ CORRECT: explicitly allow undefined (only when truly required)
interface FieldProps {
label: string;
validationMessage?: string | undefined;
}
\\\`\\\`\\\`

### Conditional object construction for strict types

When building objects (headers, config) conditionally, avoid ternaries that combine shapes with different optional keys. Build incrementally instead:

\\\`\\\`\\\`ts
// ❌ WRONG: ternary produces { Authorization: string } | {} — TypeScript infers
// Record<string, string> | { Authorization?: undefined } which fails strict checks
const headers: Record<string, string> = token
? { Authorization: \`Bearer \${token}\` }
: {};

// ✅ CORRECT: build incrementally — always produces Record<string, string>
const headers: Record<string, string> = {};
if (token) {
headers.Authorization = \`Bearer \${token}\`;
}
\\\`\\\`\\\`

This applies to any conditional object construction in `useMemo`, fetch headers, config objects, etc.

### noUncheckedIndexedAccess: safe indexing

With \`noUncheckedIndexedAccess\`, \`arr[i]\` is \`T | undefined\`. Guard before use:

\\\`\\\`\\\`ts
for (let i = 0; i < coords.length - 1; i++) {
const current = coords[i];
const next = coords[i + 1];
if (!current || !next) continue;

if (current.lat === next.lat) {
// ...
}
}
\\\`\\\`\\\`

### Form input assertions

When DOM values map to domain unions, assert carefully:

\\\`\\\`\\\`tsx
import type { AlertSeverity } from "./types";

const handleSeverityChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
setSeverity(e.target.value as AlertSeverity);
};
\\\`\\\`\\\`

## Type design

- Prefer type aliases over interfaces for most shapes (unions are simpler); reserve interface for extension-heavy cases.
- Avoid \`any\`; use \`unknown\`, then narrow via type guards before access.
- Prefer discriminated unions for domain/UI state (e.g. \`loading | success | error | offline\`).

### Type guards

\\\`\\\`\\\`ts
type User = { id: string; name: string; email: string };

function isUser(obj: unknown): obj is User {
return (
typeof obj === "object" &&
obj !== null &&
"id" in obj &&
"name" in obj &&
"email" in obj
);
}
\\\`\\\`\\\`

### Generics, utilities, and boundaries

- Prefer standard utilities (Partial, Pick, Omit, Record, Readonly) over custom helpers.
- Use generics for reusable abstractions, but avoid overly complex conditional types in app code.
- Expose stable types at module boundaries; avoid leaking transport/persistence shapes directly into UI components.

---

# Frontend Reliability (ADAC for UI) (REQUIRED)

Frontend resiliency is about user experience and network reality: Auto-Detect -> Auto-Declare -> Auto-Communicate.

ADAC declarations SHOULD be included in PR descriptions or design artifacts (e.g., ADRs) so they are reviewable and durable.

## Auto-Detect

- Resolve config in one place (build-time `import.meta.env` first, then optional runtime config).
- Detect offline/online via `navigator.onLine` + events (best-effort only).
- Detect slow requests via timeouts (fetch has none by default).
- Detect stale responses (only latest request may update state).
- Detect auth expiry (401/403) and route to re-auth UX.
- Detect throttling (429) and respect `Retry-After` when present.

## Auto-Declare

- Define explicit UI states: `loading | success | empty | error | offline`.
- Declare caching (storage + TTL), retry rules, and cancellation behavior.

## Auto-Communicate

- Always show user-friendly fallback UI (no stack traces).
- Never block critical UI rendering on optional API calls.
- Log safely: endpoint, duration, status, correlation/trace id if provided.
- Never log secrets/tokens or PII.

## Networking non-negotiables

- All fetches must support timeout + cancellation.
- Fetch has no timeout by default; enforce one via `AbortController` + `setTimeout` wrapper (or a shared `fetchJson` helper).
- Example timeout wrapper:
  \\\`\\\`\\\`ts
  function withTimeout(signal: AbortSignal, ms: number) {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), ms);
  if (signal.aborted) controller.abort();

  const abort = () => controller.abort();
  signal.addEventListener("abort", abort, { once: true });

  return {
  signal: controller.signal,
  cleanup: () => {
  clearTimeout(timeoutId);
  signal.removeEventListener("abort", abort);
  },
  };
  }
  \\\`\\\`\\\`
  Example usage:
  \\\`\\\`\\\`ts
  // Usage (always call cleanup in finally):
  const baseController = new AbortController();
  const { signal, cleanup } = withTimeout(baseController.signal, 10_000);

try {
const res = await fetch(url, { signal });
// ...
} finally {
cleanup();
}
\\\`\\\`\\\`

- Prefer a single AbortSignal per operation; if you must combine, ensure abort reasons don't get lost and always cleanup listeners.
- Use `AbortController` to cancel in-flight requests on unmount/navigation.
- Example is illustrative; production calls must use the Networking non-negotiables.
- Never blanket-retry writes. Retry only GET or explicitly idempotent operations.

## Progressive Loading and UX Resilience (REQUIRED)

- **Initial UI shell rendering MUST occur before extension data loads**. Show skeleton/loading states; do not block initial paint.
- **Lazy loading is REQUIRED for optional UI features**. Use `React.lazy()` + `Suspense` for routes and large components.
- **Fallback UI is REQUIRED when extension services are unavailable**. Degrade gracefully; show cached data or placeholder content.
- **Stale response protection is REQUIRED**: Only latest request may update state. Use request tokens or timestamps to detect stale responses.
- **UI interactivity MUST be preserved during partial service failures**. Critical UI surfaces remain operational; optional features degrade.

\\\`\\\`\\\`tsx
// Progressive loading pattern
function DashboardPage() {
const [shellReady, setShellReady] = useState(false);
const [dataState, setDataState] = useState<'loading' | 'success' | 'error'>('loading');

useEffect(() => {
// Shell renders immediately
setShellReady(true);

    // Data loads asynchronously
    fetchDashboardData()
      .then(() => setDataState('success'))
      .catch(() => setDataState('error'));

}, []);

if (!shellReady) {
return <AppShellSkeleton />;
}

return (
<AppShell>
<Navigation />
<Content>
{dataState === 'loading' && <ContentSkeleton />}
{dataState === 'success' && <DashboardContent />}
{dataState === 'error' && <ErrorFallback />}
</Content>
</AppShell>
);
}

// Lazy loading pattern
const AdminPanel = React.lazy(() => import('./panels/AdminPanel'));

function App() {
return (
<Suspense fallback={<PanelSkeleton />}>
<AdminPanel />
</Suspense>
);
}

// Stale response protection pattern
function useLatestRequest<T>(fetchFn: () => Promise<T>) {
const [data, setData] = useState<T | null>(null);
const requestIdRef = useRef(0);

const fetch = useCallback(async () => {
const requestId = ++requestIdRef.current;

    try {
      const result = await fetchFn();

      // Only update if this is still the latest request
      if (requestId === requestIdRef.current) {
        setData(result);
      }
    } catch (error) {
      if (requestId === requestIdRef.current) {
        // Handle error
      }
    }

}, [fetchFn]);

return { data, fetch };
}
\\\`\\\`\\\`

---

# API Error Handling & Graceful Degradation (REQUIRED)

**Frontend applications must handle API failures gracefully.**

Rules:

1. Display a user-friendly fallback UI (no stack traces).
2. Cache or use placeholder data when it helps UX; cache with a TTL (e.g., store `{ data, cachedAt }`, expire after N minutes).
3. Log errors safely for debugging (no secrets/tokens).
4. Never block critical UI rendering on optional API calls.
5. If an API returns RFC 9457 Problem Details (`application/problem+json`), capture `x-error-code` (when present) and treat it as the canonical error contract.

Example: read `x-error-code` + Problem Details (no `any`):

\\\`\\\`\\\`ts
type ProblemDetails = {
type?: unknown;
title?: unknown;
status?: unknown;
detail?: unknown;
instance?: unknown;
errorCode?: unknown;
traceId?: unknown;
};

export async function tryReadProblemDetails(res: Response) {
const headerCode = res.headers.get("x-error-code") ?? undefined;

try {
const body: unknown = await res.json();
const pd = body as ProblemDetails | null | undefined;
const bodyCode = typeof pd?.errorCode === "string" ? pd.errorCode : undefined;
const title = typeof pd?.title === "string" ? pd.title : undefined;
const detail = typeof pd?.detail === "string" ? pd.detail : undefined;
const traceId = typeof pd?.traceId === "string" ? pd.traceId : undefined;

    return { code: headerCode ?? bodyCode, message: detail ?? title ?? res.statusText, traceId };

} catch {
return { code: headerCode, message: res.statusText, traceId: undefined };
}
}
\\\`\\\`\\\`

\\\`\\\`\\\`tsx
type AlertsState =
| { status: "loading" }
| { status: "success"; data: Alert[] }
| { status: "empty" }
| { status: "error"; message: string; data?: Alert[] }
| { status: "offline"; data: Alert[] };

const [state, setState] = useState<AlertsState>({ status: "loading" });

useEffect(() => {
const controller = new AbortController();

fetchAlerts({ signal: controller.signal })
.then((data) => {
setState(data.length ? { status: "success", data } : { status: "empty" });
// TODO: store cachedAt + enforce TTL when reading
localStorage.setItem("alerts_cache", JSON.stringify(data));
})
.catch((err) => {
if ((err as { name?: string } | null)?.name === "AbortError") return;
console.error("Failed to fetch alerts:", err);
const cached = localStorage.getItem("alerts_cache");
const cachedData = cached ? (JSON.parse(cached) as Alert[]) : [];
setState(
cachedData.length
? { status: "offline", data: cachedData }
: { status: "error", message: "Unable to load alerts. Please try again." }
);
});

return () => controller.abort();
}, []);
\\\`\\\`\\\`

---

# Web Security Guardrails (REQUIRED)

- Never store access tokens or refresh tokens in `localStorage`/`sessionStorage`. Prefer in-memory tokens, or `HttpOnly` cookies (and then ensure CSRF protections exist).
- Never use `dangerouslySetInnerHTML` with untrusted content. If you must render HTML, sanitize with a well-maintained sanitizer and keep the allowlist tight.
- Treat all URLs as untrusted input: build query strings via `URL` / `URLSearchParams` (not string concatenation).
- For external links opened in a new tab/window, use `rel="noopener noreferrer"`.

## Content Security Policy (CSP) Compatibility (REQUIRED)

- **Design components to be CSP-compatible**: Avoid inline scripts and inline styles where feasible.
- **Use nonces or hashes** for unavoidable inline scripts/styles when CSP is enforced.
- **Test with strict CSP** during development, but note that Fluent UI/Griffel requires nonce support for CSS-in-JS.
- **Document CSP requirements** for extension deployment (required directives, nonce handling).

\\\`\\\`\\\`tsx
// CSP-friendly approach with nonce support for Fluent UI/Griffel
// Note: makeStyles generates runtime CSS that requires CSP nonces
// CSP header must include: style-src 'self' 'nonce-{random}';

const useStyles = makeStyles({
button: {
backgroundColor: tokens.colorBrandBackground,
},
});

function MyButton() {
const styles = useStyles();
return <Button className={styles.button}>Click</Button>;
}

// Fluent UI/Griffel requires nonce configuration in your root:
// <FluentProvider nonce={nonceValue}>...</FluentProvider>
\\\`\\\`\\\`

## Clickjacking Protection (REQUIRED)

- **Use server-controlled headers** (`X-Frame-Options` and CSP `frame-ancestors`) to control iframe embedding.
- **If embedding third-party content**, use `sandbox` attribute to restrict capabilities.
- **Client-side frame ancestor detection** is best-effort only due to cross-origin restrictions; rely on server headers for enforcement.

\\\`\\\`\\\`tsx
// Sandbox third-party content

<iframe
  src="https://trusted-external.com/widget"
  sandbox="allow-scripts allow-same-origin"
  title="External Widget"
/>
\\\`\\\`\\\`

## Secure URL Construction (REQUIRED)

- **Use platform APIs for URL construction**: `URL` and `URLSearchParams`.
- **Validate and sanitize URL parameters** before navigation or redirection.
- **Never concatenate user input into URLs** without encoding.
- **Never place sensitive data** (tokens, passwords, secrets, one-time codes) in URL path segments or query strings.

\\\`\\\`\\\`tsx
// Safe URL construction
function navigateToResource(resourceId: string, filters: Record<string, string>) {
const url = new URL('/resources', window.location.origin);
url.pathname += `/${encodeURIComponent(resourceId)}`;

Object.entries(filters).forEach(([key, value]) => {
url.searchParams.append(key, value);
});

window.location.href = url.toString();
}
\\\`\\\`\\\`

## Third-Party Content and Threat Modeling (REQUIRED)

- **Treat embedded or third-party content as untrusted**: Isolate via iframes with restrictive sandbox.
- **Perform threat modeling for extension UI surfaces**: Identify attack vectors (XSS, CSRF, injection).
- **Document trust boundaries**: Which data sources are trusted vs untrusted.
- **Use Subresource Integrity (SRI)** for external scripts/stylesheets.

\\\`\\\`\\\`html

<!-- SRI for external resources -->
<script
  src="https://cdn.example.com/library.js"
  integrity="sha384-abc123..."
  crossorigin="anonymous"
></script>

\\\`\\\`\\\`

## Untrusted URL and SSRF controls (REQUIRED)

- Treat user-provided URLs as untrusted input.
- Allow only `https:` unless a documented exception exists.
- Enforce hostname allowlists for server-side fetch/proxy patterns.
- Block private/link-local address ranges (`127.0.0.0/8`, `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`, `169.254.0.0/16`, `::1`, `fc00::/7`).
- Reject ambiguous hosts (embedded credentials, mixed-encoding hostnames, malformed ports).

## Object merge and prototype pollution safety (REQUIRED)

- Do not recursively merge untrusted objects into application state/config.
- Reject dangerous keys before merge: `__proto__`, `prototype`, `constructor`.
- Prefer explicit allowlisted mapping into typed objects over deep merge utilities.

## Package Publish Safeguards (REQUIRED)

- **Use `files` allowlist in `package.json`**, not `.npmignore`. Allowlists fail safe (new files are excluded by default); denylists fail open.

```jsonc
// package.json — only these files ship in the published package
{
  "files": ["dist/", "README.md", "LICENSE"],
}
```

- **Disable source maps for production npm packages.** Set `"sourceMap": false` in `tsconfig.json` for production builds, or strip `.map` files from `dist/` before publish. Source maps expose original source code.
- **Run `npm pack --dry-run` in CI before publish** and verify the file list contains only intended files. No `src/`, `test/`, `.env`, `prompts/`, server-side modules, or agent instruction files should appear.
- **Separate server-side and client-side code** at the package/project level. Do not rely on tree-shaking or bundler configuration to exclude server-only modules from client packages.
- **Never include AI system prompts, tool definitions, or `SKILL.md` files** in the `files` list. See `distribution-security.instructions.md` for the full framework.

---

# API Contract Alignment (Required)

- Frontend endpoints must match the backend contract. Do not call `/health` unless the API explicitly exposes it.
- Prefer `/health/ready` and `/health/live` for Kubernetes-style probes when those endpoints exist.
- If you add or change API paths, update both client calls and server routes in the same change set.
- Avoid multi-shape response parsing in clients. If the API contract is inconsistent, fix the backend or add a single, versioned response type.
- Align enum casing with backend serialization to avoid normalization logic and drift (choose one casing and enforce it end-to-end).
- **Numeric range contracts**: Document and enforce whether numeric fields use 0.0–1.0 or 0–100 scale. Never multiply or divide a value without verifying the source scale. If the backend returns 0–100, do not `* 100` again on the frontend. If the backend returns 0.0–1.0, apply `* 100` exactly once for display. Add a code comment noting the expected range at each transformation point.
- **Display labels vs. API keys**: Keep display text (e.g., "Sustainability Score") separate from API identifiers (e.g., `"Sustainability"`). Use a typed mapping object (`Record<ApiCategory, string>`) to translate between them. Never pass display labels to API calls or use them as filter keys.
- **Field completeness**: When the frontend renders a detail view that depends on a backend field (e.g., `estimatedImpact`, `driftCategory`), verify the backend populates that field for ALL code paths that produce the entity. Empty/null fields that silently hide UI sections are a defect.
- **Category/taxonomy alignment**: When the backend uses multiple taxonomies for the same concept (e.g., `Rule.Category` vs. `DriftCategory`), the frontend filter must use the same taxonomy as the data source being filtered. Mismatches cause "no results" for valid data.
- Avoid hard-coded API version strings in multiple layers; prefer build-time injection or a shared config endpoint.
- If the backend requires `api-version`, include it in requests via a single config source. Do not reconstruct server-driven pagination or LRO URLs: treat `nextLink` and `operation-location` as absolute and call them as-is.
- LRO polling MUST honor `Retry-After` headers when present; when absent, use bounded backoff + jitter (for example, 2s to 10s) to avoid thundering herds.
- LRO polling loops MUST be cancelable on unmount/navigation and must ignore stale poll responses (latest request wins).
- If backend LRO start responses include both `operation-location` and `Location`, clients should prefer `operation-location` and only fall back to `Location` when necessary.

---

# Enterprise Admin UI Posture (Portal-like, Standalone SPA) (Optional)

Apply this section when building resource-management/admin UIs (lists → details → actions) where information density and predictable command surfaces matter.

## Default navigation and layout

- Use a stable app shell (nav + header + content). Keep navigation predictable.
- Prefer list → details → action flows; keep details deep-linkable via routes.
- Prefer a right-side panel/drawer for edits and secondary flows to preserve context; model panel state in the URL when feasible.

## Panel (“blade-like”) behavior (recommended)

- Prefer a single active panel at a time; nested panels are allowed only if the backstack behavior is explicit and testable.
- Keep list context stable when panels open/close: preserve filters/sort/search, selection, and scroll position.
- Make panel state route-driven when feasible:
  - The URL should represent which panel is open (and which resource it targets).
  - Browser Back should close the panel before leaving the underlying page.
  - Deep links should open the correct page + panel state without extra user steps.

## Data grids and lists

- Prefer Fluent UI `DataGrid` for resource lists; support sorting, filtering, and search.
- Debounce search input (300–500ms) and cancel in-flight requests when inputs change.
- For large lists, use virtualization and/or server paging; never sort/filter large arrays inside render.

## Commands, LROs, and non-blocking UX

- Primary actions live in a single command surface per view (toolbar/command bar). Avoid scattered buttons.
- If an API returns `202` + `operation-location`, show progress in-context and poll in the background; respect `Retry-After` on status responses; do not block the whole shell; show a toast on completion/failure.
- **Progress indicators are REQUIRED** for operations returning `operation-location` responses.
- **Polling implementations MUST respect `Retry-After` headers** when present in LRO status responses.
- **Polling cadence MUST be server-driven first** (`Retry-After`), with bounded jittered fallback only when the server does not provide guidance.
- **UI must remain responsive during LRO polling**; use background tasks, not synchronous blocking.
- **Completion notifications or toasts are REQUIRED** for LRO completion or failure.
- **Cancellation support is REQUIRED** when the underlying API supports operation cancellation.

## Errors and supportability

- Show user-friendly messages, but capture `x-error-code` + correlation/trace id for logs and “Copy details” support flows.

## Accessibility and keyboard UX (panels)

- When a panel opens: move focus into the panel and trap focus (dialog/drawer semantics).
- `Esc` closes the panel (unless doing so would lose unsaved work; then confirm).
- When a panel closes: restore focus to the launching control.
- Do not break keyboard navigation or screen reader semantics for the underlying page when a panel is open.

---

# Portal Extension Architecture Requirements

For applications designed as extensible portals or platforms where independent extensions/plugins extend core functionality.

## Extension Isolation (REQUIRED)

- **Extensions MUST be isolated from shell failures**. A crashed extension must not bring down the shell or other extensions.
- **Extension loading MUST use error boundaries** or equivalent isolation mechanisms (try/catch for dynamic imports).
- **Extensions MUST be independently deployable and versionable**. Each extension has its own deployment lifecycle.
- **Shell and extension compatibility MUST use capability discovery** instead of strict version matching. Extensions query available capabilities at runtime.

\\\`\\\`\\\`tsx
// Extension loading with error boundary
const LazyExtension = React.lazy(() =>
import('./extensions/MyExtension').catch((err) => {
console.error('Failed to load extension:', err);
return { default: () => <div>Extension unavailable</div> };
})
);

function ExtensionHost() {
return (
<ErrorBoundary fallback={<div>Extension failed</div>}>
<Suspense fallback={<div>Loading extension...</div>}>
<LazyExtension />
</Suspense>
</ErrorBoundary>
);
}
\\\`\\\`\\\`

## Extension Communication (REQUIRED)

- **Extensions MUST NOT have direct runtime dependencies** on other extensions. No direct imports between extensions.
- **Extensions MUST communicate only through**:
  - Shell-provided services (dependency injection)
  - Event bus mechanisms (pub/sub)
  - Shared state management (shell-managed context)
- **Shell provides the communication contract**; extensions adapt to it.

\\\`\\\`\\\`tsx
// Shell-provided services pattern
interface ShellServices {
telemetry: TelemetryService;
navigation: NavigationService;
notifications: NotificationService;
}

function MyExtension({ services }: { services: ShellServices }) {
const handleAction = () => {
services.telemetry.trackEvent('ExtensionAction', { action: 'submit' });
services.notifications.show('Action completed');
};

return <Button onClick={handleAction}>Submit</Button>;
}
\\\`\\\`\\\`

## Capability Discovery (REQUIRED)

- **Extensions MUST dynamically detect platform capabilities** at runtime, not assume features exist.
- **Feature capability contracts or feature flags are REQUIRED** for optional services.
- **Extensions MUST gracefully degrade** when optional capabilities are unavailable.

\\\`\\\`\\\`tsx
interface ExtensionContext {
capabilities: {
hasNotifications: boolean;
hasAdvancedTelemetry: boolean;
apiVersion: string;
};
services: Partial<ShellServices>;
}

function MyExtension({ context }: { context: ExtensionContext }) {
const notifyUser = (message: string) => {
if (context.capabilities.hasNotifications && context.services.notifications) {
context.services.notifications.show(message);
} else {
// Fallback: use browser alert or inline message
console.log('Notification:', message);
}
};

return <Button onClick={() => notifyUser('Done')}>Action</Button>;
}
\\\`\\\`\\\`

## Extension Lifecycle and Metadata (REQUIRED)

- **Extension onboarding expectations**: Extensions must declare manifest metadata (name, version, capabilities required, entry point).
- **Extension ownership metadata is REQUIRED**: Owner team, support contact, deprecation policy.
- **Backward compatibility expectations**: Breaking changes to shell APIs require deprecation notices and migration guides.
- **Deprecation policy**: Shell must provide at least one major version deprecation window for removed APIs.

Example manifest:
\\\`\\\`\\\`json
{
"name": "my-extension",
"version": "2.1.0",
"owner": "platform-team@example.com",
"requiredCapabilities": ["notifications", "telemetry"],
"optionalCapabilities": ["advancedSearch"],
"entryPoint": "./dist/index.js",
"deprecationPolicy": "follow-shell-versioning"
}
\\\`\\\`\\\`

---

# Cross-Extension Communication Standards

For portals supporting multiple extensions that need coordinated workflows.

## Communication Patterns (REQUIRED)

- **Direct extension-to-extension runtime coupling is FORBIDDEN**. Extensions must not import or directly call each other.
- **Shell-mediated communication patterns are REQUIRED**:
  - Event bus for loosely coupled notifications
  - Shared state management (shell-provided context/store)
  - Shell API contracts for cross-extension coordination
- **Message contracts MUST be version tolerant**. Use additive schema changes; avoid breaking changes.
- **Prefer asynchronous event-driven messaging** over synchronous request/response patterns.

\\\`\\\`\\\`tsx
// Event bus pattern (shell-provided)
interface EventBus {
publish(event: string, payload: unknown): void;
subscribe(event: string, handler: (payload: unknown) => void): () => void;
}

// Extension A (publisher)
function ExtensionA({ eventBus }: { eventBus: EventBus }) {
const notifyChange = () => {
eventBus.publish('resource:updated', {
resourceId: '123',
timestamp: Date.now()
});
};

return <Button onClick={notifyChange}>Update Resource</Button>;
}

// Extension B (subscriber)
function ExtensionB({ eventBus }: { eventBus: EventBus }) {
useEffect(() => {
const unsubscribe = eventBus.subscribe('resource:updated', (payload) => {
console.log('Resource updated:', payload);
// Refresh view
});

    return unsubscribe;

}, [eventBus]);

return <div>Extension B</div>;
}
\\\`\\\`\\\`

## Event Ownership and Documentation (REQUIRED)

- **Event schemas MUST be documented** with type definitions and examples.
- **Event ownership MUST be declared**: Which extension/team owns the event contract.
- **Schema evolution MUST be backward compatible**: Add optional fields; do not remove or rename required fields.
- **Event versioning**: Use event names with versions (`resource:updated:v2`) when breaking changes are unavoidable.

\\\`\\\`\\\`tsx
// Event schema documentation
type ResourceUpdatedEventV1 = {
resourceId: string;
timestamp: number;
// v1 fields
};

type ResourceUpdatedEventV2 = ResourceUpdatedEventV1 & {
userId?: string; // Added in v2, optional for backward compatibility
// Future: DO NOT remove resourceId or timestamp
};
\\\`\\\`\\\`

---

# Telemetry and Observability Requirements

For enterprise portals requiring end-to-end traceability and operational insights.

## Correlation ID Propagation (REQUIRED)

- **Correlation IDs MUST be propagated** across UI → API → telemetry flows.
- **Use W3C trace context** (traceparent, tracestate) when available from backend responses.
- **Generate client-side correlation IDs** using `crypto.randomUUID()` for operations originating in the UI.
- **Include correlation ID in all API requests** (custom header, e.g., `X-Correlation-Id` or `traceparent`).
- **Include correlation ID in telemetry events** for traceability.

\\\`\\\`\\\`tsx
async function fetchWithCorrelation<T>(url: string, options?: RequestInit): Promise<T> {
const correlationId = crypto.randomUUID();

const response = await fetch(url, {
...options,
headers: {
...options?.headers,
'X-Correlation-Id': correlationId,
},
});

if (!response.ok) {
const error = new Error(`Request failed: ${response.status}`);
telemetry.trackException(error, {
correlationId,
url,
status: response.status,
errorCode: response.headers.get('x-error-code') ?? 'unknown',
});
throw error;
}

telemetry.trackEvent('APIRequestSucceeded', {
correlationId,
url,
duration: /_ calculate _/,
});

return response.json() as Promise<T>;
}
\\\`\\\`\\\`

## Critical Operator Actions (REQUIRED)

- **Telemetry events are REQUIRED** for critical operator actions (create, update, delete, approve, execute).
- **Event names MUST follow consistent naming** (e.g., `Resource.Created`, `Workflow.Approved`).
- **Include action metadata**: user identifier (anonymized if needed), resource identifier, outcome (success/failure).

\\\`\\\`\\\`tsx
function handleApproveWorkflow(workflowId: string) {
const startTime = Date.now();

return approveWorkflow(workflowId)
.then(() => {
telemetry.trackEvent('Workflow.Approved', {
workflowId,
duration: Date.now() - startTime,
outcome: 'success',
});
})
.catch((error) => {
telemetry.trackEvent('Workflow.ApprovalFailed', {
workflowId,
duration: Date.now() - startTime,
outcome: 'failure',
errorCode: error.code ?? 'unknown',
});
throw error;
});
}
\\\`\\\`\\\`

## Error Telemetry (REQUIRED)

- **Error telemetry MUST capture `x-error-code`** and trace identifiers when available from API responses.
- **Include context**: operation name, resource identifier, user action leading to error.
- **Do NOT log secrets, tokens, or PII** in telemetry.

\\\`\\\`\\\`tsx
type APIError = Error & {
code?: string;
traceId?: string;
};

async function handleAPIError(error: APIError, context: { operation: string; resourceId?: string }) {
const errorDetails = {
operation: context.operation,
resourceId: context.resourceId,
errorMessage: error.message,
// Extract from error if available
errorCode: error.code ?? 'unknown',
traceId: error.traceId,
timestamp: new Date().toISOString(),
};

telemetry.trackException(error, errorDetails);
}
\\\`\\\`\\\`

## Performance Telemetry (REQUIRED)

- **Performance telemetry is REQUIRED** for major workflows (page load, data fetch, form submission).
- **Track user-perceived latency**: time from action to visible result.
- **Use performance marks** for key milestones.

\\\`\\\`\\\`tsx
function WorkflowPage() {
useEffect(() => {
performance.mark('workflow-page-start');

    return () => {
      performance.mark('workflow-page-end');
      performance.measure('workflow-page-load', 'workflow-page-start', 'workflow-page-end');

      const measure = performance.getEntriesByName('workflow-page-load')[0];
      if (measure) {
        telemetry.trackMetric('PageLoadTime', measure.duration, {
          page: 'WorkflowPage',
        });
      }
    };

}, []);

return <div>Workflow Content</div>;
}
\\\`\\\`\\\`

## Shell Telemetry Services (REQUIRED)

- **Use shell-provided telemetry services** when available; do not initialize separate telemetry clients in extensions.
- **Shell services MUST provide**: trackEvent, trackException, trackMetric, trackPageView.
- **Extensions inherit telemetry configuration** from shell (sampling, routing, filtering).

\\\`\\\`\\\`tsx
interface TelemetryService {
trackEvent(name: string, properties?: Record<string, string | number>): void;
trackException(error: Error, properties?: Record<string, string | number>): void;
trackMetric(name: string, value: number, properties?: Record<string, string | number>): void;
trackPageView(name: string, properties?: Record<string, string | number>): void;
}
\\\`\\\`\\\`

## ADAC Alignment

Telemetry requirements align with ADAC (Auto-Detect → Auto-Declare → Auto-Communicate) reliability principles:

- **Auto-Detect**: Track failures, degraded states, and capability unavailability.
- **Auto-Declare**: Telemetry documents operational mode and degradation.
- **Auto-Communicate**: Telemetry enables operators to diagnose issues without user reports.

---

# React and Fluent UI v9

## UI composition priority (mandatory)

When building UI components:

- Prefer Fluent UI React v9 components and patterns over generic React or HTML solutions.
- Do not introduce custom UI primitives (button, input, dialog, menu, tooltip, etc.) when Fluent UI provides an equivalent.
- Use <Button> instead of <button>, <Input> instead of <input>, etc., unless there is a documented reason.
- Prefer Button with a simple string/short node; use layout wrappers + tokens (e.g., makeStyles, tokens, shorthands) for multi-line button content.

## Design posture (modern, portable, future-proof)

- Use Fluent UI tokens for spacing, color, typography, and motion — avoid raw CSS values when tokens exist.
- Prefer self-contained component “surfaces” (panels, dialogs, regions) over page-centric layouts.
- Avoid reliance on global CSS resets; avoid browser-chrome assumptions.
- Support keyboard and touch; don’t assume mouse-only UX.

## Modern UI baseline (new projects)

- Start with a single theme file that defines brand tokens (colors, typography, radii) and apply it via FluentProvider at the app root.
- Choose intentional typography (display + body pair) and load fonts explicitly; avoid default-only font stacks.
- Build layered surfaces (cards, panels, callouts) with subtle contrast instead of flat monochrome pages.
- Use tokens or CSS variables for gradients/background treatments; keep raw hex values scoped to the theme.
- Add motion deliberately (page load + staggered reveals) and honor prefers-reduced-motion.
- Treat contrast as non-negotiable for status/error messaging; verify against the chosen surface tokens.

## Components and hooks

- Use functional components with hooks; avoid class components.
- Type props explicitly; do not rely on \`React.FC\` for children. Type children explicitly when needed.
- Avoid business-logic-heavy \`useEffect\` in components; prefer extracting logic into custom hooks/services.

\\\`\\\`\\\`tsx
import { Button, Text } from "@fluentui/react-components";

interface UserCardProps {
user: User;
onSelect?: (id: string) => void;
}

const UserCard = ({ user, onSelect }: UserCardProps) => {
const handleClick = () => onSelect?.(user.id);

return (
<Button onClick={handleClick}>
<Text>{user.name}</Text>
<Text>{user.email}</Text>
</Button>
);
};
\\\`\\\`\\\`

### Hooks correctness (required)

**CRITICAL**: Treat exhaustive-deps warnings as correctness bugs, not lint annoyances.

\\\`\\\`\\\`tsx
// ❌ WRONG
useEffect(() => {
fetchUser(userId);
}, []);

// ✅ CORRECT
useEffect(() => {
fetchUser(userId);
}, [userId]);
\\\`\\\`\\\`

If adding deps causes unwanted re-runs, the fix is usually:

- Extract logic into \`useCallback\`/\`useMemo\`
- Move constants outside the component
- Use a ref only when you truly need non-reactive state

#### Rules of Hooks hard stop

- Never call hooks conditionally.
- Never call hooks inside loops, nested functions, or callbacks.
- If per-item hook state is required, extract a child component/custom hook and call hooks at that top level.
- Enforce `react-hooks/rules-of-hooks` as an error in lint/CI.

---

# Performance and accessibility (reduce fat)

## Rendering performance (default rules)

- Avoid creating new objects/functions inline in hot paths and list rows; hoist or memoize.
- Don’t sort/filter/mutate arrays in render without memoization:
  \\\`\\\`\\\`tsx
  const sorted = useMemo(() => [...items].sort(compare), [items, compare]);
  \\\`\\\`\\\`
- Stabilize props for memoized components (callbacks/options objects).
- Avoid duplicated derived state; derive from source state during render.

## Network/runtime performance

- Never block initial render on non-critical API calls.
- Guard against repeated fetch loops caused by unstable dependencies.
- Add stale-response guards/cancellation where it prevents wasted work.

## Accessibility

- Use semantic elements and proper roles/labels.
- Ensure focus management for dialogs/menus.
- Prefer Fluent UI components to get accessibility and theming by default.

## Bundle Governance and Performance Budgets (REQUIRED)

- **Route or extension bundle size budgets MUST be defined**: Track main bundle, route chunks, and extension entry points.
- **Dynamic import is REQUIRED for large optional functionality**: Split code by route, feature flag, or admin-only features.
- **CI MUST detect bundle size regressions**: Fail builds if bundle size increases beyond threshold (e.g., +5% or +50KB).
- **Dependency impact analysis is REQUIRED** when adding packages: Check tree-shaken size, assess alternatives.
- **Unused dependencies MUST be removed** or reviewed: Run `npm ls` / `pnpm list` to audit dependency tree; remove unused packages.

### Bundle Size Targets (Vite + ESM)

- **Main bundle**: <300KB gzipped (app shell + critical routes)
- **Route chunks**: <100KB gzipped per route
- **Extension bundles**: <150KB gzipped per extension
- **Vendor chunk**: <500KB gzipped (shared libraries)

### CI Bundle Size Enforcement

\\\`\\\`\\\`yaml

# .github/workflows/ci.yml

- name: Build and check bundle size
  run: |
  npm run build
  npx vite-bundle-visualizer --json > bundle-stats.json

      # Check main bundle size (gzipped)
      MAIN_SIZE=$(gzip -c dist/assets/index-*.js | wc -c)
      if [ $MAIN_SIZE -gt 307200 ]; then  # 300KB gzipped
        echo "Main bundle exceeds 300KB gzipped: ${MAIN_SIZE} bytes"
        exit 1
      fi

  \\\`\\\`\\\`

### Dependency Audit Pattern

\\\`\\\`\\\`bash

# Before adding a dependency

npm install --dry-run <package>
npx bundlephobia <package> # Check size impact

# After adding dependency

npm run build

# Compare dist/ size before and after

\\\`\\\`\\\`

### Dynamic Import Pattern

\\\`\\\`\\\`tsx
// Large admin feature loaded on-demand
const AdminDashboard = React.lazy(() =>
import('./pages/AdminDashboard')
);

function App() {
const { isAdmin } = useAuth();

return (
<Routes>
<Route path="/" element={<HomePage />} />
{isAdmin && (
<Route
path="/admin"
element={
<Suspense fallback={<LoadingSpinner />}>
<AdminDashboard />
</Suspense>
}
/>
)}
</Routes>
);
}
\\\`\\\`\\\`

## Automated Accessibility Testing (REQUIRED)

- **Automated accessibility testing is REQUIRED in CI**: Use `axe-core`, `jest-axe`, or Playwright accessibility checks.
- **Keyboard navigation validation is REQUIRED** for panel/drawer UI: Tab order, focus trap, Esc to close.
- **Focus restoration validation is REQUIRED** for modal/panel closure: Focus returns to triggering element.
- **Contrast validation using design tokens is REQUIRED**: Verify against WCAG AA contrast ratios (4.5:1 for normal text, 3:1 for large text).
- **WCAG AA baseline is REQUIRED** unless exceptions are formally documented in ADRs.

### Accessibility Testing Pattern

See [typescript-tests.instructions.md](typescript-tests.instructions.md) for accessibility testing code patterns (jest-axe, Playwright keyboard navigation, contrast validation).

---

# Error handling and async patterns

Use \`async/await\` with explicit error paths; always check \`response.ok\` and parse RFC 9457 Problem Details + \`x-error-code\` on failures.

- Empty `catch {}` blocks are not allowed in production code.
- Promise chains must terminate with `.catch(...)` unless intentionally fire-and-forget and explicitly handled.
- Long-lived `useEffect` subscriptions/timers/listeners must return a cleanup function.

\\\`\\\`\\\`ts
async function fetchJson<T>(url: string, signal?: AbortSignal): Promise<T> {
const response = await fetch(url, { signal });

if (!response.ok) {
const { code, message } = await tryReadProblemDetails(response);
throw new Error(code ? `${code}: ${message}` : message);
}

return (await response.json()) as T;
}
\\\`\\\`\\\`

For untrusted responses, prefer runtime validation (e.g. Zod) at boundaries.

---

# Build validation (production is authoritative)

**Production builds are authoritative; do not rely on dev-time validation alone.**

Recommended CI baseline:

- \`npm ci\`
- \`npx tsc --noEmit\`
- \`npm run lint\`
- \`npm run build\`
- tests (unit + optional e2e)

\\\`\\\`\\\`yaml
steps:

- run: npm ci
- run: npx tsc --noEmit
- run: npm run lint
- run: npm run build
- run: npm test -- --coverage
- run: npm run e2e
  \\\`\\\`\\\`

---

# Architecture Decision Record (ADR) Governance

For tracking architectural decisions that impact platform reliability, extensibility, or contract stability.

## ADR Requirements (REQUIRED)

- **ADRs are REQUIRED** for new UI platform patterns (extension architecture, state management, communication patterns).
- **ADRs are REQUIRED** for reliability pattern changes (caching strategies, retry policies, offline UX).
- **ADAC declarations MUST be included in ADRs** for network or resiliency changes: Document Auto-Detect, Auto-Declare, Auto-Communicate decisions.
- **ADRs MUST document** caching, retry, and offline UX strategies with rationale and trade-offs.
- **ADRs MUST be updated** for breaking contract or UI behavior changes (API versioning, event schemas, telemetry contracts).

### ADR Template

\\\`\\\`\\\`markdown

# ADR-XXX: [Title]

## Status

[Proposed | Accepted | Deprecated | Superseded by ADR-YYY]

## Context

What is the problem we are trying to solve? What constraints exist?

## Decision

What did we decide to do?

## Consequences

### Positive

- Benefit 1
- Benefit 2

### Negative

- Trade-off 1
- Trade-off 2

## ADAC Declaration (if applicable)

**Auto-Detect**: [How the system detects failures or degraded states]

**Auto-Declare**: [How operational mode is declared]

**Auto-Communicate**: [How degradation is communicated to operators/users]

## Implementation Notes

[Technical details, migration path, rollback strategy]

## References

[Links to specs, RFCs, related ADRs]
\\\`\\\`\\\`

### Example ADR

\\\`\\\`\\\`markdown

# ADR-042: Extension State Caching Strategy

## Status

Accepted

## Context

Extensions need to maintain state across browser sessions. Current approach uses sessionStorage, which loses data on tab close.

## Decision

Use IndexedDB with 24-hour TTL for extension state persistence.

## Consequences

### Positive

- State survives browser restart
- Larger storage quota (50MB+)
- Asynchronous API (non-blocking)

### Negative

- More complex error handling
- Requires migration from sessionStorage
- IndexedDB not available in private browsing

## ADAC Declaration

**Auto-Detect**: Check `window.indexedDB` availability; fallback to sessionStorage if unavailable.

**Auto-Declare**: Extension logs storage mode on load: "Using IndexedDB" or "Degraded: Using sessionStorage".

**Auto-Communicate**: Show banner if private browsing detected: "Some features unavailable in private mode."

## Implementation Notes

- Add IndexedDB wrapper with TTL enforcement
- Migrate existing sessionStorage data on first load
- Fallback to sessionStorage if IndexedDB fails

## References

- [MDN IndexedDB Guide](https://developer.mozilla.org/en-US/docs/Web/API/IndexedDB_API)
  \\\`\\\`\\\`

---

# Multi-Tenant Cloud Platform Requirements

For applications serving multiple tenants with isolation and compliance requirements.

## Tenant Context Isolation (REQUIRED)

- **Tenant context MUST be isolated** in UI state management. No cross-tenant data leakage in shared state.
- **Tenant switching MUST clear previous tenant state** or namespace state by tenant identifier.
- **Components MUST validate tenant context** before rendering sensitive data.

\\\`\\\`\\\`tsx
// Tenant context pattern
interface TenantContextValue {
tenantId: string;
tenantName: string;
permissions: string[];
}

const TenantContext = createContext<TenantContextValue | null>(null);

function useTenant() {
const context = useContext(TenantContext);
if (!context) {
throw new Error('useTenant must be used within TenantProvider');
}
return context;
}

function ResourceList() {
const { tenantId } = useTenant();

// Data is scoped to tenantId
const { data } = useQuery(['resources', tenantId], () =>
fetchResources(tenantId)
);

return <div>{/_ Render resources _/}</div>;
}
\\\`\\\`\\\`

## Tenant-Aware Caching (REQUIRED)

- **Cache keys MUST include tenant identifier**: Prevent cross-tenant cache pollution.
- **Tenant switching MUST invalidate tenant-specific caches**: Clear localStorage, sessionStorage, IndexedDB entries.
- **Shared caches MUST namespace by tenant**: Use prefixes like `tenant-${tenantId}-${key}`.

\\\`\\\`\\\`tsx
// Tenant-aware caching pattern
function useTenantCache<T>(key: string) {
const { tenantId } = useTenant();
const cacheKey = `tenant-${tenantId}-${key}`;

const getCached = (): T | null => {
const cached = localStorage.getItem(cacheKey);
if (!cached) return null;

    const { data, cachedAt } = JSON.parse(cached);
    const age = Date.now() - cachedAt;

    // 5-minute TTL
    if (age > 5 * 60 * 1000) {
      localStorage.removeItem(cacheKey);
      return null;
    }

    return data as T;

};

const setCached = (data: T) => {
localStorage.setItem(cacheKey, JSON.stringify({
data,
cachedAt: Date.now(),
}));
};

return { getCached, setCached };
}
\\\`\\\`\\\`

## Tenant-Safe Telemetry (REQUIRED)

- **Telemetry MUST NOT log tenant PII** unless explicitly anonymized or hashed.
- **Telemetry properties MUST include tenant identifier** for support and diagnostics.
- **Tenant-specific telemetry sampling** may be configured per tenant tier.

\\\`\\\`\\\`tsx
function trackTenantEvent(eventName: string, properties: Record<string, unknown>) {
const { tenantId, tenantName } = useTenant();

telemetry.trackEvent(eventName, {
...properties,
tenantId, // Safe: identifier only
// DO NOT log: email, phone, address, etc.
});
}
\\\`\\\`\\\`

## Tenant Authorization Boundaries (REQUIRED)

- **UI workflows MUST validate tenant authorization** before mutating actions.
- **API calls MUST include tenant context** in headers or request body.
- **Authorization errors MUST be tenant-safe**: Do not leak existence of resources in other tenants.

\\\`\\\`\\\`tsx
async function deleteResource(resourceId: string) {
const { tenantId } = useTenant();

const response = await fetch(`/api/resources/${resourceId}`, {
method: 'DELETE',
headers: {
'X-Tenant-Id': tenantId,
},
});

if (response.status === 404) {
// Tenant-safe error: do not reveal if resource exists in another tenant
throw new Error('Resource not found');
}

if (!response.ok) {
throw new Error('Delete failed');
}
}
\\\`\\\`\\\`

## Tenant Switching Support (REQUIRED)

- **Tenant switching MUST be supported without full application reload** where feasible.
- **On tenant switch**:
  - Clear tenant-specific state
  - Invalidate tenant-specific caches
  - Re-fetch tenant-specific data
  - Update telemetry context
  - Validate user has access to new tenant

\\\`\\\`\\\`tsx
function TenantSwitcher() {
const { tenantId, setTenant } = useTenant();
const { availableTenants } = useAuth();

const handleTenantSwitch = async (newTenantId: string) => {
// Validate access
if (!availableTenants.some(t => t.id === newTenantId)) {
throw new Error('Access denied');
}

    // Clear tenant state
    clearTenantCaches(tenantId);

    // Switch tenant
    await setTenant(newTenantId);

    // Update telemetry
    telemetry.setContext({ tenantId: newTenantId });

};

return (
<Dropdown
value={tenantId}
onChange={(e, data) => handleTenantSwitch(data.value as string)} >
{availableTenants.map(t => (

<Option key={t.id} value={t.id}>{t.name}</Option>
))}
</Dropdown>
);
}
\\\`\\\`\\\`

---

# Extension Testing Strategy Requirements

See [typescript-tests.instructions.md](typescript-tests.instructions.md) for full extension testing conventions (unit testing, shell mock integration, contract testing with MSW, accessibility coverage, E2E with Playwright).

---

# Testing

See [typescript-tests.instructions.md](typescript-tests.instructions.md) for full testing conventions (Vitest, React Testing Library, Playwright, MSW, accessibility testing, anti-patterns).

---

# Linting, formatting, and monorepo posture

- Keep the codebase lint-clean; only add eslint-disable with justification and a TODO.
- Prefer one formatting flow (Prettier via ESLint or separate) and document it.
- If `eslint-config-prettier` is installed, wire it in `eslint.config.js` to prevent rule conflicts.
- In monorepos, respect package boundaries; keep shared types/config in dedicated packages.

---

# References

- TypeScript docs: https://www.typescriptlang.org/docs/
- React docs: https://react.dev/
- ISE JS/TS code reviews: https://microsoft.github.io/code-with-engineering-playbook/code-reviews/recipes/javascript-and-typescript/
- Fluent UI React v9: https://react.fluentui.dev/
- Vite env vars: https://vitejs.dev/guide/env-and-mode.html
