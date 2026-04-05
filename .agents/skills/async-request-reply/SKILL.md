---
name: async-request-reply
description: >-
  Implement and review asynchronous Request-Reply APIs (202 Accepted + status polling) with consistent contracts, Retry-After handling, cancellation, and idempotency. USE FOR: long-running operations in APIs and client polling workflows.
---

# Async Request-Reply Skill (Azure-style LRO)

Use this skill when an operation is long-running and should not block a request thread.

## When to Use This Skill

- A create/action request can take longer than normal API latency targets
- Work is queued or processed in background workers
- You need resilient client polling and operation tracking
- You are designing or reviewing `202 Accepted` API flows

## Contract Baseline (Required)

### Start operation endpoint

- Validate request first; only return `202 Accepted` when work has been accepted
- Return headers:
  - `operation-location`: absolute URL of status monitor endpoint (**required**)
  - `Location`: same absolute status URL for broad client compatibility (**recommended**)
  - `Retry-After`: polling hint in seconds (**required**)
- Return body fields:
  - `operationId` (stable server-generated identifier)
  - `status` (initially `NotStarted` or `Running`)
  - optional links (`statusUrl`, `resourceLocation`) for diagnostics

### Status monitor endpoint

- Returns operation resource by ID
- Supported statuses: `NotStarted | Running | Succeeded | Failed | Canceled`
- While non-terminal (`NotStarted` / `Running`), include `Retry-After`
- Terminal behavior:
  - `Succeeded`: include result or `resourceLocation`
  - `Failed`: return RFC 9457 Problem Details with `x-error-code`
  - `Canceled`: return a clear terminal cancellation payload

### Cancellation endpoint (when supported)

- `POST .../operations/{id}:cancel` or equivalent
- Idempotent behavior: repeated cancel requests do not fail unexpectedly
- Status transitions must be explicit and auditable

## Idempotency and Repeatability (Required)

- POST create/action operations should be idempotent when retries are likely
- Support repeatability headers:
  - `Repeatability-Request-ID`
  - `Repeatability-First-Sent`
- Keep a replay window (minimum 5 minutes)
- If repeatability headers are sent but unsupported, return `501 Not Implemented`

## Server Implementation Checklist

- Persist operation state in durable storage (not only memory)
- Track timestamps: created, lastUpdated, completed
- Include correlation IDs in logs and operation records
- Ensure worker retries are bounded and jittered
- Never retry non-idempotent writes without idempotency controls
- Expose consistent error contract: RFC 9457 + `x-error-code`

## Client Polling Checklist

- Poll `operation-location` as absolute URL (do not rebuild URL)
- Respect `Retry-After` when provided
- Use bounded fallback interval + jitter when missing (e.g., 2s–10s)
- Keep UI responsive while polling (background polling only)
- Surface completion/failure/cancellation to user via clear notification
- Support cancellation if API exposes it

> [!IMPORTANT]
> Client polling must use bounded backoff with jitter:
> 1. Honor `Retry-After` header when present (server knows best)
> 2. When absent, use exponential backoff: start 2s, max 30s, jitter ±20%
> 3. Set a maximum poll count (e.g., 60 attempts = ~10 min with backoff) before reporting timeout to the user
> 4. Cancel polling on unmount/navigation (AbortController pattern)

## API Review Quick Checks

- Does `202` response include `operation-location` + `Retry-After`?
- Is `Location` also present for compatibility?
- Is status schema consistent across all operation types?
- Are terminal states explicit and final?
- Is failure payload RFC 9457 with stable `x-error-code`?
- Are repeatability headers honored on POST?
- Is cancellation behavior defined and idempotent?

## Minimal Example Shape

`POST /tasks/recompute` → `202 Accepted`

- Headers:
  - `operation-location: https://api.contoso.com/operations/abc123?api-version=2026-01-01`
  - `Location: https://api.contoso.com/operations/abc123?api-version=2026-01-01`
  - `Retry-After: 5`
- Body:
  - `{ "operationId": "abc123", "status": "Running" }`

`GET /operations/abc123`

- `200` + `{ "operationId": "abc123", "status": "Running" }` (+ `Retry-After`)
- terminal:
  - success: `{ "operationId": "abc123", "status": "Succeeded", "resourceLocation": "..." }`
  - failure: RFC 9457 + `x-error-code`

## Common Pitfalls to Prevent

- Returning `202` before request validation is complete
- Missing `Retry-After`, causing aggressive client polling
- Returning relative `operation-location` URLs
- Using inconsistent status names between endpoints
- Omitting cancellation semantics for long-running actions
- Returning ad-hoc error shapes from status monitor

## References

- Azure async request-reply pattern
- Azure API LRO guidelines (`operation-location`, repeatability, pagination)
- RFC 9457 Problem Details

---

## Currency

- **Date checked:** 2026-03-31
- **Sources:** Microsoft Learn MCP (`microsoft_docs_search`), IETF RFCs
- **Authoritative references:** [Azure async request-reply pattern](https://learn.microsoft.com/azure/architecture/patterns/async-request-reply), [RFC 9457](https://www.rfc-editor.org/rfc/rfc9457)

### Verification Steps

1. Confirm Azure LRO header conventions (`operation-location`, `Retry-After`) are unchanged
2. Verify RFC 9457 Problem Details is still the current error format standard
3. Check for new Azure SDK LRO polling helpers or deprecations
