---
name: typescript-react-patterns
description: >-
  Production-ready patterns for React 19 / TypeScript / Fluent UI v9 / Vite frontend development. USE FOR: implement Fluent UI v9 DataGrid/CommandBar/Dialog, handle SignalR hub reconnect state, parse RFC 9457 Problem Details, configure Vite build-time env vars, write strict TypeScript without 'any', manage async state with discriminated unions.
---

# TypeScript / React Frontend Patterns

Patterns for React 19 / TypeScript strict mode / Fluent UI v9 / Vite frontend development on the Emergency Alerts stack.

---

## When to Use This Skill

- Implementing or reviewing React components using Fluent UI v9
- Setting up or debugging SignalR hub client connections (reconnect state reconciliation)
- Implementing API error handling with RFC 9457 Problem Details
- Configuring Vite build-time environment variables (`VITE_API_URL`, `VITE_SIGNALR_URL`)
- Writing TypeScript with strict mode (no `any`, no `!` non-null assertions)
- Building bulk-action triage UIs (DataGrid, CommandBar, selection state)

---

## 1. TypeScript Strict Mode (REQUIRED)

**Never use `any`.** Use `unknown` with a type guard, or define an explicit type.

```ts
// ❌ WRONG
function processResponse(data: any) {
  return data.alerts;
}

// ✅ CORRECT — unknown + type guard
type AlertsResponse = { alerts: Alert[]; nextLink?: string };

function isAlertsResponse(v: unknown): v is AlertsResponse {
  return (
    typeof v === "object" &&
    v !== null &&
    "alerts" in v &&
    Array.isArray((v as AlertsResponse).alerts)
  );
}

async function fetchAlerts(): Promise<AlertsResponse> {
  const data: unknown = await response.json();
  if (!isAlertsResponse(data)) throw new Error("Unexpected response shape");
  return data;
}
```

**Prohibited patterns:**

| Pattern                       | Why                   | Fix                             |
| ----------------------------- | --------------------- | ------------------------------- |
| `any`                         | Bypasses type safety  | `unknown` + type guard          |
| `@ts-ignore`                  | Hides real errors     | Fix the underlying type         |
| `value!` non-null assertion   | Runtime crash if null | Explicit null check or `?.`     |
| `eval()` / `Function()`       | Code injection        | Never                           |
| Auth tokens in `localStorage` | XSS exfiltration      | In-memory or `HttpOnly` cookies |

---

## 2. API State with Discriminated Unions (REQUIRED)

Never model async state with boolean flags. Use a discriminated union.

```ts
type AlertsState =
  | { status: "loading" }
  | { status: "success"; data: Alert[]; nextLink?: string }
  | { status: "empty" }
  | { status: "error"; message: string; errorCode?: string; data?: Alert[] }
  | { status: "offline"; data: Alert[] };

const [state, setState] = useState<AlertsState>({ status: "loading" });
```

---

## 3. RFC 9457 Problem Details Error Handling (REQUIRED)

All API errors return `application/problem+json` with an `x-error-code` header. Parse both.

```ts
// services/apiInterceptor.ts
type ProblemDetails = {
  type?: string;
  title?: string;
  status?: number;
  detail?: string;
  errorCode?: string;
  traceId?: string;
};

export async function tryReadProblemDetails(res: Response) {
  const headerCode = res.headers.get("x-error-code") ?? undefined;
  try {
    const body: unknown = await res.json();
    const pd = body as ProblemDetails | null;
    const bodyCode =
      typeof pd?.errorCode === "string" ? pd.errorCode : undefined;
    return {
      code: headerCode ?? bodyCode,
      message: pd?.detail ?? pd?.title ?? res.statusText,
      traceId: pd?.traceId,
    };
  } catch {
    return { code: headerCode, message: res.statusText, traceId: undefined };
  }
}

// Usage in a fetch call
const res = await fetch(`${apiBase}/api/v1/alerts?api-version=2024-01-01`);
if (!res.ok) {
  const { code, message, traceId } = await tryReadProblemDetails(res);
  console.error("API error", { code, message, traceId });
  setState({ status: "error", message, errorCode: code });
  return;
}
```

---

## 4. SignalR Hub Client Lifecycle (CRITICAL)

### Connection Setup

```ts
// services/dashboardHubService.ts
import * as signalR from "@microsoft/signalr";

export function createHubConnection(baseUrl: string) {
  return new signalR.HubConnectionBuilder()
    .withUrl(`${baseUrl}/api/hubs/alerts`)
    .withAutomaticReconnect([0, 2000, 5000, 10000, 30000])
    .configureLogging(signalR.LogLevel.Information)
    .build();
}
```

### Reconnect State Reconciliation (CRITICAL)

`withAutomaticReconnect` silently reconnects but the **new pod has no memory of prior server-side group memberships**. Always re-invoke subscriptions in `onreconnected`.

```ts
// ❌ WRONG — client shows Connected but receives no events after pod restart
connection.onreconnected((id) => {
  console.debug("Reconnected:", id);
  // Missing: re-subscribe to server-side groups
});

// ✅ CORRECT — re-invoke on every reconnect
connection.onreconnected(async (_id) => {
  try {
    await connection.invoke("SubscribeToDashboard");
  } catch (err) {
    console.error("Re-subscribe failed:", err);
  }
});
```

### Full Lifecycle Pattern

```tsx
// hooks/useAlertHub.ts
import { useEffect, useRef } from "react";
import * as signalR from "@microsoft/signalr";

export function useAlertHub(
  baseUrl: string,
  onAlertCreated: (alert: Alert) => void,
) {
  const connRef = useRef<signalR.HubConnection | null>(null);

  useEffect(() => {
    const conn = new signalR.HubConnectionBuilder()
      .withUrl(`${baseUrl}/api/hubs/alerts`)
      .withAutomaticReconnect()
      .build();

    connRef.current = conn;
    conn.on("AlertCreated", onAlertCreated);

    const subscribe = async () => {
      try {
        await conn.invoke("SubscribeToDashboard");
      } catch (err) {
        console.error("Subscribe failed:", err);
      }
    };

    conn.onreconnected(() => {
      void subscribe();
    });

    conn
      .start()
      .then(() => subscribe())
      .catch((err) => console.error("Hub start failed:", err));

    return () => {
      conn.off("AlertCreated", onAlertCreated);
      void conn.stop();
    };
  }, [baseUrl]); // ← only reconnect when baseUrl changes
}
```

---

## 5. Fluent UI v9 Patterns (REQUIRED)

Always use Fluent UI v9 components. Never create custom buttons, inputs, or dialogs.

### DataGrid for Alert Triage

```tsx
import {
  DataGrid,
  DataGridHeader,
  DataGridHeaderCell,
  DataGridBody,
  DataGridRow,
  DataGridCell,
  createTableColumn,
  TableColumnDefinition,
  TableRowId,
  SelectionItemId,
} from "@fluentui/react-components";
import type { Alert } from "../types/Alert";

const columns: TableColumnDefinition<Alert>[] = [
  createTableColumn<Alert>({
    columnId: "title",
    renderHeaderCell: () => "Title",
    renderCell: (item) => item.title,
  }),
  createTableColumn<Alert>({
    columnId: "severity",
    renderHeaderCell: () => "Severity",
    renderCell: (item) => <SeverityBadge severity={item.severity} />,
  }),
];

function AlertGrid({ alerts }: { alerts: Alert[] }) {
  const [selectedRows, setSelectedRows] = useState<Set<SelectionItemId>>(
    new Set(),
  );

  // Clear selection when alerts change (page/filter/sort)
  useEffect(() => setSelectedRows(new Set()), [alerts]);

  return (
    <DataGrid
      items={alerts.filter((a) => a.alertId != null)} // Guard against undefined key
      columns={columns}
      selectionMode="multiselect"
      selectedItems={selectedRows}
      onSelectionChange={(_, data) => setSelectedRows(data.selectedItems)}
      getRowId={(item) => item.alertId}
      sortable
    >
      <DataGridHeader>
        <DataGridRow>
          {({ renderHeaderCell }) => (
            <DataGridHeaderCell>{renderHeaderCell()}</DataGridHeaderCell>
          )}
        </DataGridRow>
      </DataGridHeader>
      <DataGridBody<Alert>>
        {({ item, rowId }) => (
          <DataGridRow<Alert> key={rowId}>
            {({ renderCell }) => (
              <DataGridCell>{renderCell(item)}</DataGridCell>
            )}
          </DataGridRow>
        )}
      </DataGridBody>
    </DataGrid>
  );
}
```

### Severity Badge Colors

| Severity | Fluent Token                    | Color  |
| -------- | ------------------------------- | ------ |
| Extreme  | `colorPaletteRedBorder2`        | Red    |
| Severe   | `colorPaletteOrangeBackground2` | Orange |
| Moderate | `colorPaletteYellowBackground2` | Yellow |
| Minor    | `colorPaletteGreenBackground2`  | Green  |
| Unknown  | `colorNeutralBackground3`       | Gray   |

### CommandBar for Bulk Actions

```tsx
import { Toolbar, ToolbarButton, Divider } from "@fluentui/react-components";
import {
  CheckmarkCircleRegular,
  DismissCircleRegular,
} from "@fluentui/react-icons";

function AlertCommandBar({
  selectedCount,
  onApprove,
  onReject,
}: {
  selectedCount: number;
  onApprove: () => void;
  onReject: () => void;
}) {
  const disabled = selectedCount === 0;
  return (
    <Toolbar>
      <ToolbarButton
        icon={<CheckmarkCircleRegular />}
        disabled={disabled}
        onClick={onApprove}
      >
        Approve ({selectedCount})
      </ToolbarButton>
      <Divider vertical />
      <ToolbarButton
        icon={<DismissCircleRegular />}
        disabled={disabled}
        onClick={onReject}
      >
        Reject ({selectedCount})
      </ToolbarButton>
    </Toolbar>
  );
}
```

### Confirmation Dialog

```tsx
import {
  Dialog,
  DialogSurface,
  DialogTitle,
  DialogContent,
  DialogActions,
  DialogTrigger,
  Button,
} from "@fluentui/react-components";

function ConfirmBulkAction({
  open,
  action,
  count,
  onConfirm,
  onCancel,
}: {
  open: boolean;
  action: "Approve" | "Reject";
  count: number;
  onConfirm: () => void;
  onCancel: () => void;
}) {
  return (
    <Dialog open={open}>
      <DialogSurface>
        <DialogTitle>Confirm {action}</DialogTitle>
        <DialogContent>
          {action} {count} alert{count !== 1 ? "s" : ""}? This cannot be undone.
        </DialogContent>
        <DialogActions>
          <Button appearance="primary" onClick={onConfirm}>
            {action}
          </Button>
          <DialogTrigger disableButtonEnhancement>
            <Button onClick={onCancel}>Cancel</Button>
          </DialogTrigger>
        </DialogActions>
      </DialogSurface>
    </Dialog>
  );
}
```

---

## 6. Bulk Operations with Promise.allSettled (REQUIRED)

Always use `Promise.allSettled` for bulk actions. Never `Promise.all` — one failure cancels all.

```ts
async function handleBulkApprove(
  alertIds: string[],
  apiBase: string,
): Promise<{ succeeded: number; failed: number }> {
  const results = await Promise.allSettled(
    alertIds.map((id) =>
      fetch(`${apiBase}/api/v1/alerts/${id}/approve?api-version=2024-01-01`, {
        method: "POST",
      }),
    ),
  );

  const succeeded = results.filter((r) => r.status === "fulfilled").length;
  const failed = results.length - succeeded;
  return { succeeded, failed };
}
```

---

## 7. Vite Build-Time Environment Variables (CRITICAL)

> **Canonical reference:** [SPA Endpoint Configuration](../spa-endpoint-configuration/SKILL.md) covers the full Docker build-arg pattern, CI/CD integration, and troubleshooting. This section covers the TypeScript/React application code side.

Vite injects `VITE_*` variables at **build time**, not runtime. Changing pod env vars after build does nothing.

```ts
// src/services/apiBaseUrl.ts
export function getApiBaseUrl(): string {
  // Empty default allows detection of missing injection at dev time
  return (import.meta.env.VITE_API_URL as string) ?? "";
}

export function getSignalRBaseUrl(): string {
  return (import.meta.env.VITE_SIGNALR_URL as string) ?? getApiBaseUrl();
}
```

```dockerfile
# frontend/Dockerfile
ARG VITE_API_URL=       # MUST be empty default — injected at build time
ARG VITE_SIGNALR_URL=

ENV VITE_API_URL=$VITE_API_URL
ENV VITE_SIGNALR_URL=$VITE_SIGNALR_URL

RUN npm run build
```

```bash
# Build with explicit args — required when hosts change
IMAGE_TAG=$(git rev-parse --short HEAD)
docker build --no-cache \
    --build-arg VITE_API_URL="http://api-dev-abc123.australiaeast.cloudapp.azure.com" \
    --build-arg VITE_SIGNALR_URL="http://api-dev-abc123.australiaeast.cloudapp.azure.com" \
    -t myacr.azurecr.io/frontend:${IMAGE_TAG} \
    -f frontend/Dockerfile frontend/
```

**Verification (CI required):**

```bash
if [[ -n "$VITE_API_URL" ]]; then
    grep -qF "$VITE_API_URL" dist/assets/*.js \
        || { echo "❌ Bundle missing VITE_API_URL"; exit 1; }
fi
```

---

## 8. Fetch with Timeout and Cancellation (REQUIRED)

`fetch` has no built-in timeout. Always use `AbortController`.

```ts
export async function fetchWithTimeout<T>(
  url: string,
  options: RequestInit & { timeoutMs?: number } = {},
): Promise<T> {
  const { timeoutMs = 10_000, signal: outerSignal, ...rest } = options;
  const controller = new AbortController();
  const timerId = setTimeout(() => controller.abort(), timeoutMs);

  if (outerSignal) {
    outerSignal.addEventListener("abort", () => controller.abort(), {
      once: true,
    });
  }

  try {
    const res = await fetch(url, { ...rest, signal: controller.signal });
    if (!res.ok) {
      const { code, message } = await tryReadProblemDetails(res);
      throw Object.assign(new Error(message), { errorCode: code });
    }
    return (await res.json()) as T;
  } finally {
    clearTimeout(timerId);
  }
}
```

---

## 9. Alert Severity Sort Order

Severity priority queue order for the triage inbox:

```ts
const SEVERITY_ORDER: Record<string, number> = {
  extreme: 0,
  severe: 1,
  moderate: 2,
  minor: 3,
  unknown: 4,
};

function sortByPriority(alerts: Alert[]): Alert[] {
  return [...alerts].sort((a, b) => {
    const oa = SEVERITY_ORDER[a.severity.toLowerCase()] ?? 4;
    const ob = SEVERITY_ORDER[b.severity.toLowerCase()] ?? 4;
    return oa - ob;
  });
}
```

Severity values from API: `Extreme`, `Severe`, `Moderate`, `Minor`, `Unknown`.

---

## 10. Azure Maps Integration

Azure Maps guidance has moved to a dedicated skill:

- [Azure Maps Integration](../azure-maps-integration/SKILL.md) — Auth model selection, secure Web SDK setup, RBAC role guidance, and production guardrails

---

## 11. ADAC Resilience: Auto-Detect → Auto-Declare → Auto-Communicate (REQUIRED)

Frontend reliability follows the ADAC triad. See `typescript.instructions.md` for the full pattern and `csharp.instructions.md` for the backend counterpart.

### Auto-Detect

- Track the timestamp of the last successful API/SignalR update.
- Detect connection state: `connected | reconnecting | disconnected`.
- Flag data as stale when no update arrives within a defined threshold (e.g., 60 seconds).
- Detect auth expiry (401/403) and route to re-auth UX.
- Detect throttling (429) and respect `Retry-After`.

### Auto-Declare

- Define explicit UI states: `loading | success | empty | error | offline`.
- Store connection state and last-update timestamp in a shared context/store.
- Expose the same state in telemetry (structured logs with endpoint, duration, status, correlation ID).

### Auto-Communicate

- Display a degraded-mode banner with a specific reason (stale data, disconnected, auth failure).
- Never silently fall back — users must see reduced fidelity.
- Never block critical UI rendering on optional API calls.
- Log safely: endpoint, duration, status, correlation/trace ID. Never log secrets, tokens, or PII.

### ADAC Checklist (Frontend)

- [ ] Last-update timestamp stored and rendered
- [ ] Connection state is visible in the UI
- [ ] Stale-data threshold is defined and tested
- [ ] Degraded-mode message is actionable and specific
- [ ] No silent fallbacks — every failure path has a user-visible indicator

---

## 12. Accessibility Patterns — WCAG 2.2 AA (REQUIRED)

All frontend code must meet WCAG 2.2 Level AA. Use Fluent UI v9 components — they provide baseline accessibility, but you must still handle focus management, ARIA labels, and dynamic content correctly.

### Fluent UI v9 accessible component usage

```tsx
import {
  Button,
  Tooltip,
  Dialog,
  DialogSurface,
  DialogTitle,
  DialogBody,
  DialogActions,
} from "@fluentui/react-components";
import { DeleteRegular } from "@fluentui/react-icons";

// Icon-only button: aria-label is REQUIRED
function DeleteButton({
  onDelete,
  resourceName,
}: {
  onDelete: () => void;
  resourceName: string;
}) {
  return (
    <Tooltip content={`Delete ${resourceName}`} relationship="label">
      <Button
        icon={<DeleteRegular />}
        aria-label={`Delete ${resourceName}`}
        appearance="subtle"
        onClick={onDelete}
      />
    </Tooltip>
  );
}

// Dialog with focus trap (automatic in Fluent Dialog)
function ConfirmDeleteDialog({
  open,
  onConfirm,
  onCancel,
  resourceName,
}: {
  open: boolean;
  onConfirm: () => void;
  onCancel: () => void;
  resourceName: string;
}) {
  return (
    <Dialog
      open={open}
      onOpenChange={(_, data) => {
        if (!data.open) onCancel();
      }}
    >
      <DialogSurface aria-label={`Confirm deletion of ${resourceName}`}>
        <DialogTitle>Delete {resourceName}?</DialogTitle>
        <DialogBody>This action cannot be undone.</DialogBody>
        <DialogActions>
          <Button appearance="secondary" onClick={onCancel}>
            Cancel
          </Button>
          <Button appearance="primary" onClick={onConfirm}>
            Delete
          </Button>
        </DialogActions>
      </DialogSurface>
    </Dialog>
  );
}
```

### Live region for real-time updates (SignalR)

```tsx
import { useRef } from "react";

function SignalRStatusAnnouncer({
  connectionState,
}: {
  connectionState: "connected" | "reconnecting" | "disconnected";
}) {
  const previousState = useRef(connectionState);

  const message =
    connectionState !== previousState.current
      ? `Connection ${connectionState}`
      : undefined;

  previousState.current = connectionState;

  return (
    <div
      role="status"
      aria-live="polite"
      aria-atomic="true"
      style={{
        position: "absolute",
        width: 1,
        height: 1,
        overflow: "hidden",
        clip: "rect(0,0,0,0)",
      }}
    >
      {message}
    </div>
  );
}
```

### DataGrid keyboard and screen reader support

```tsx
import {
  DataGrid,
  DataGridHeader,
  DataGridRow,
  DataGridHeaderCell,
  DataGridBody,
  DataGridCell,
  TableColumnDefinition,
  createTableColumn,
} from "@fluentui/react-components";

// Fluent DataGrid provides keyboard navigation (arrow keys, Home, End) and
// ARIA grid semantics automatically. Ensure:
// 1. Every column has a clear header label
// 2. Row selection uses selectionMode prop (not custom checkboxes)
// 3. Action buttons inside cells have aria-label with row context

const columns: TableColumnDefinition<Resource>[] = [
  createTableColumn({
    columnId: "name",
    renderHeaderCell: () => "Resource Name",
    renderCell: (item) => item.name,
  }),
  createTableColumn({
    columnId: "actions",
    renderHeaderCell: () => "Actions",
    renderCell: (item) => (
      <Button
        aria-label={`Edit ${item.name}`}
        icon={<EditRegular />}
        appearance="subtle"
        size="small"
      />
    ),
  }),
];
```

### Skip navigation link

```tsx
// Place as the first focusable element in App.tsx
function SkipToContent() {
  return (
    <a
      href="#main-content"
      style={{
        position: "absolute",
        left: "-10000px",
        top: "auto",
        width: "1px",
        height: "1px",
        overflow: "hidden",
      }}
      onFocus={(e) => {
        e.currentTarget.style.position = "static";
        e.currentTarget.style.width = "auto";
        e.currentTarget.style.height = "auto";
      }}
      onBlur={(e) => {
        e.currentTarget.style.position = "absolute";
        e.currentTarget.style.width = "1px";
        e.currentTarget.style.height = "1px";
      }}
    >
      Skip to main content
    </a>
  );
}

// Target element
// <main id="main-content" tabIndex={-1}> ... </main>
```

### Reduced motion support

```tsx
const prefersReducedMotion = window.matchMedia(
  "(prefers-reduced-motion: reduce)",
).matches;

// Use in animation logic
const transitionDuration = prefersReducedMotion ? "0ms" : "200ms";
```

### Accessibility testing checklist

- [ ] All pages navigable by keyboard alone (Tab, Shift+Tab, Enter, Space, Escape, Arrow keys)
- [ ] Focus indicator visible on every interactive element
- [ ] Screen reader announces page structure (headings, landmarks, lists)
- [ ] Icon-only buttons have `aria-label` describing the action
- [ ] Form errors programmatically associated via `aria-describedby` or Fluent `Field`
- [ ] Colour contrast verified: ≥ 4.5:1 normal text, ≥ 3:1 large text / UI components
- [ ] No information conveyed by colour alone
- [ ] Live regions announce dynamic content changes (toasts, status updates)
- [ ] `prefers-reduced-motion` respected for animations
- [ ] Touch targets ≥ 44×44px (or ≥ 24×24px with spacing for dense admin UIs)

---

## Checklist: Before Merging Frontend Changes

- [ ] No `any` types — `unknown` + type guards used instead
- [ ] No `@ts-ignore` comments
- [ ] No auth tokens in `localStorage`/`sessionStorage`
- [ ] All async state uses discriminated union (`loading | success | error | offline`)
- [ ] RFC 9457 error parsing extracts both `x-error-code` header and `errorCode` body field
- [ ] SignalR `onreconnected` re-invokes all server-side subscription methods
- [ ] ADAC: connection state visible, stale-data threshold defined, no silent fallbacks
- [ ] All Fluent UI components used; no raw `<button>` / `<input>` / `<dialog>`
- [ ] Bulk operations use `Promise.allSettled`
- [ ] Alert filters remove items with `alertId == null` before rendering DataGrid
- [ ] Selection state cleared when alerts array changes (page/filter/sort)
- [ ] `VITE_API_URL` and `VITE_SIGNALR_URL` passed as Docker `--build-arg` (empty Dockerfile defaults)
- [ ] `crypto.randomUUID()` used for correlation IDs (not `Math.random()`)
- [ ] Accessibility: all interactive elements keyboard-reachable, icon-only buttons have `aria-label`
- [ ] Accessibility: colour contrast ≥ 4.5:1 (normal text) / ≥ 3:1 (large text + UI components)
- [ ] Accessibility: `prefers-reduced-motion` respected; live regions used for dynamic updates

---

## Troubleshooting

### **SignalR JSON parse error: `SyntaxError: Unexpected token '<'`**

Frontend is calling its own origin, not the backend.
**Fix:** Rebuild frontend with correct `--build-arg VITE_API_URL=<actual-api-host>`.

### **DataGrid shows blank rows or wrong selection counts**

Cause: `alertId` is undefined for some items, causing React key instability.
**Fix:** Filter before rendering: `items={alerts.filter(a => a.alertId != null)}`.

### **Bulk approve shows 0 selected after action completes**

Expected behavior — selection is cleared in `useEffect` when the `alerts` array updates after refresh. Reapply sort after refresh to maintain queue order.

---

## References

- **typescript.instructions.md**: `.github/instructions/typescript.instructions.md`
- **Fluent UI v9**: https://react.fluentui.dev/
- **Fluent UI icon browser (flicon.io)**: https://flicon.io/ — searchable reference for `@fluentui/react-icons` icon names
- **SignalR JS client**: https://learn.microsoft.com/aspnet/core/signalr/javascript-client
- **Vite env vars**: https://vitejs.dev/guide/env-and-mode.html
- **Azure Maps React**: https://github.com/Azure/react-azure-maps

---

## Related Skills

- [Azure Maps Integration](../azure-maps-integration/SKILL.md) — Security-first Maps auth, RBAC, and integration patterns
- [SPA Endpoint Configuration](../spa-endpoint-configuration/SKILL.md) — Build-time environment variable injection
- [.NET Backend Patterns](../dotnet-backend-patterns/SKILL.md) — API backend this frontend consumes
- [Azure Portal Branding](../azure-portal-branding/SKILL.md) — Fluent UI theming and branding

---

## Currency and Verification

- **Date checked:** 2026-03-31
- **Compatibility:** React 19.x, TypeScript 5.x (strict mode), Fluent UI v9, Vite 8.x
- **Sources:** [React docs](https://react.dev/), [Fluent UI v9](https://react.fluentui.dev/), [Vite docs](https://vitejs.dev/), [react-azure-maps](https://github.com/Azure/react-azure-maps)
- **Verification steps:**
  1. Check React version: `npm list react`
  2. Check Fluent UI version: `npm list @fluentui/react-components`
  3. Verify Vite env vars baked at build time: `grep -F "VITE_API_URL" dist/assets/*.js`
  4. Check Vite version: `npm list vite` (Vite 8 released; Vite 6 projects should plan migration)

### Known Pitfalls

| Area                           | Pitfall                                                                                       | Mitigation                                                                                       |
| ------------------------------ | --------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------ |
| Vite env vars                  | `VITE_*` variables are baked at build time, not runtime; changing them requires a rebuild     | Rebuild Docker image when API URLs change; verify with `grep -F "VITE_API_URL" dist/assets/*.js` |
| React 19 `defaultProps`        | React 19 deprecates `defaultProps` for function components                                    | Use default parameters in function signatures instead of `defaultProps`                          |
| Fluent UI v9 bundle size       | Importing from barrel exports (`@fluentui/react-components`) can bloat bundles                | Import specific components; verify bundle size with `npx vite-bundle-visualizer`                 |
| Error boundary gaps            | Missing error boundaries cause full app crashes on component errors                           | Wrap route-level components in `<ErrorBoundary>` with fallback UI                                |
| `useEffect` cleanup            | Missing cleanup for subscriptions/timers causes memory leaks and state-after-unmount warnings | Return cleanup function from `useEffect`; use `AbortController` for fetch calls                  |
| `localStorage` for auth tokens | Storing tokens in `localStorage`/`sessionStorage` exposes them to XSS                         | Keep tokens in memory or `HttpOnly` cookies; never persist in browser storage                    |
| Vite 6 → 8 migration           | Vite 8 released with breaking changes; Vite 6 projects on older config may need updates       | Review [Vite 8 migration guide](https://vitejs.dev/guide/migration); test build before upgrading |
