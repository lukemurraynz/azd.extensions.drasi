---
name: api-versioning-governance
description: >-
  API version lifecycle governance: versioning strategies, breaking change rules, deprecation policy, sunset timelines, and Azure API Management version sets. USE FOR: choosing a versioning strategy, defining breaking vs non-breaking changes, creating deprecation policy, configuring APIM version sets, reviewing API contracts for backward compatibility.
license: MIT
metadata:
  author: Azure API Platform Team
  version: "1.0"
  last-updated: "2026-02-27"
  azure-services: "api-management, api-center"
  standards: "OpenAPI 3.x, Azure REST API Guidelines"
---

# API Versioning Governance Skill

Provides governance rules for API version lifecycle management: strategy selection, breaking change classification, deprecation timelines, and Azure API Management version set configuration. Based on [Microsoft REST API Guidelines](https://github.com/microsoft/api-guidelines) and Azure API design patterns.

## When to Use This Skill

| Trigger                           | Use Case                                    |
| --------------------------------- | ------------------------------------------- |
| "How should we version this API?" | Select versioning strategy                  |
| "Is this a breaking change?"      | Classify contract changes                   |
| "Deprecate API v1"                | Create deprecation plan with timeline       |
| "Set up APIM version sets"        | Configure Azure API Management versioning   |
| "API contract review"             | Check backward compatibility before release |
| New API project starting          | Establish versioning policy from day one    |

---

## Versioning Strategy Selection

### Strategy Comparison

| Strategy         | Format Example                          | Pros                                     | Cons                                     | Best For                    |
| ---------------- | --------------------------------------- | ---------------------------------------- | ---------------------------------------- | --------------------------- |
| **URL Path**     | `/api/v1/orders`                        | Obvious, cacheable, easy routing         | Proliferates base URLs, harder to evolve | Public APIs, REST APIs      |
| **Query String** | `/api/orders?api-version=2024-01-15`    | No URL change, Azure-standard            | Easy to miss, not in path-based routing  | Azure-style APIs            |
| **Header**       | `Api-Version: 2`                        | Clean URLs, supports content negotiation | Not visible in browser, harder to test   | Internal/service-to-service |
| **Media Type**   | `Accept: application/vnd.myapp.v2+json` | Most RESTful, per-resource versioning    | Complex, poor tooling support            | Hypermedia APIs (rare)      |

### Decision Guide

```
Is this a public-facing API?
├── Yes → URL Path versioning (most discoverable)
│       └── Use: /api/v{major}/resource
├── Azure-ecosystem API?
│   └── Yes → Query String with date-based versions
│           └── Use: ?api-version=YYYY-MM-DD
└── Internal service-to-service?
    └── Header versioning (clean contract)
        └── Use: Api-Version: {major}.{minor}
```

**Rule:** Pick ONE strategy per API surface and apply it consistently. Never mix strategies within the same API.

---

## Version Format Standards

### Date-Based Versions (Azure-Style)

Format: `YYYY-MM-DD` with optional `-preview` suffix

```
2024-01-15           ← GA stable
2024-06-01-preview   ← Preview (breaking changes allowed)
2025-01-15           ← Next GA (supersedes 2024-01-15)
```

**Rules:**

- GA versions are immutable — no breaking changes ever
- Preview versions MAY have breaking changes between releases
- Preview MUST graduate to GA or be removed within 12 months
- Date reflects the API freeze date, not the release date

### Semantic Versions (URL Path Style)

Format: `v{major}` in URL, full `{major}.{minor}.{patch}` in headers/docs

```
/api/v1/orders       ← Major version in URL
/api/v2/orders       ← Breaking changes = new major
```

**Rules:**

- **Major** (v1 → v2): Breaking changes that require client updates
- **Minor** (1.1 → 1.2): New endpoints/fields, fully backward compatible
- **Patch** (1.2.1 → 1.2.2): Bug fixes, no contract changes
- URL path only includes major version
- Minor/patch communicated via response headers or documentation

---

## Breaking vs Non-Breaking Changes

### Breaking Changes (Require New Version)

| Change Type                  | Example                                      | Why It Breaks                    |
| ---------------------------- | -------------------------------------------- | -------------------------------- |
| Remove endpoint              | `DELETE /api/v1/users/{id}/avatar` removed   | Existing clients get 404         |
| Remove response field        | `user.middleName` removed from response      | Client deserialization fails     |
| Rename field                 | `user.firstName` → `user.givenName`          | Client field mapping breaks      |
| Change field type            | `user.age: string` → `user.age: number`      | Type mismatch / parse failure    |
| Add required request field   | `POST /orders` now requires `currency`       | Existing requests rejected (400) |
| Change HTTP method           | `POST /search` → `GET /search`               | Client request method wrong      |
| Change error format          | Custom errors → RFC 9457 Problem Details     | Client error handling breaks     |
| Tighten validation           | `name: max 100 chars` → `name: max 50 chars` | Previously valid data rejected   |
| Change authentication scheme | API key → OAuth 2.0                          | Auth mechanism incompatible      |
| Change status code semantics | `200` success → `202` accepted               | Client flow logic breaks         |

### Non-Breaking Changes (Safe in Current Version)

| Change Type                  | Example                                      | Why It's Safe                     |
| ---------------------------- | -------------------------------------------- | --------------------------------- | ------------------------------ |
| Add optional request field   | New optional `tags` parameter                | Old clients just don't send it    |
| Add response field           | New `user.avatarUrl` in response             | Clients ignore unknown fields     |
| Add new endpoint             | `GET /api/v1/orders/{id}/history`            | No existing client uses it        |
| Loosen validation            | `name: max 50 chars` → `name: max 100 chars` | Previously valid data still valid |
| Add new enum value           | `status: active                              | inactive`→ add`suspended`         | Safe IF clients handle unknown |
| Add optional response header | `X-Request-Cost: 5`                          | Clients ignore unknown headers    |
| Performance improvement      | Faster response, same contract               | Transparent to clients            |
| Add new error code           | New `RATE_LIMITED` error type                | Safe IF clients handle unknown    |

> ⚠️ **Adding a new enum value is only non-breaking if clients are documented to handle unknown values gracefully.** If clients use strict deserialization (e.g., `System.Text.Json` with `JsonStringEnumConverter` strict mode), this IS breaking. Clarify in your API contract.

---

## Deprecation Policy

### Lifecycle Stages

```
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│  Preview │───▶│  Active  │───▶│Deprecated│───▶│  Sunset  │
│          │    │  (GA)    │    │          │    │ (Removed) │
└──────────┘    └──────────┘    └──────────┘    └──────────┘
     │               │               │               │
  May break    Immutable      Supported,         Endpoint
  at any time  contract       no new features    returns 410
```

| Stage          | Duration                    | Support Level                                     |
| -------------- | --------------------------- | ------------------------------------------------- |
| **Preview**    | Up to 12 months             | Breaking changes allowed. No SLA.                 |
| **Active**     | Until next major version GA | Full support, bug fixes, security patches.        |
| **Deprecated** | Minimum 12 months           | Security patches only. Migration guide published. |
| **Sunset**     | Permanent                   | Endpoint removed or returns `410 Gone`.           |

### Deprecation Communication

When deprecating a version, ALL of the following are required:

#### 1. Response Headers (automated)

```http
HTTP/1.1 200 OK
Deprecation: Sun, 01 Sep 2025 00:00:00 GMT
Sunset: Mon, 01 Sep 2026 00:00:00 GMT
Link: </api/v2/orders>; rel="successor-version"
```

#### 2. API Management Policy (if using APIM)

```xml
<inbound>
    <set-header name="Deprecation" exists-action="override">
        <value>Sun, 01 Sep 2025 00:00:00 GMT</value>
    </set-header>
    <set-header name="Sunset" exists-action="override">
        <value>Mon, 01 Sep 2026 00:00:00 GMT</value>
    </set-header>
    <set-header name="Link" exists-action="override">
        <value>&lt;/api/v2/orders&gt;; rel="successor-version"</value>
    </set-header>
</inbound>
```

#### 3. Documentation & Communication

- Changelog entry with deprecation date and sunset date
- Migration guide: field-by-field mapping from old → new version
- Email/notification to registered API consumers (if applicable)
- API portal banner (Azure API Center or developer portal)

#### 4. Monitoring

Track deprecated version usage to confirm migration progress:

```
Tool: mcp_azure_mcp_applicationinsights
Query: "requests to deprecated API version usage metrics"
```

KQL for tracking deprecated version calls:

```kql
requests
| where url has "/api/v1/"
| summarize CallCount = count(), UniqueCallers = dcount(client_IP) by bin(timestamp, 1d)
| order by timestamp desc
```

---

## Azure API Management Version Sets

### Bicep Template

```bicep
// API version set — groups all versions of an API
resource apiVersionSet 'Microsoft.ApiManagement/service/apiVersionSets@2024-05-01' = {
  parent: apimService
  name: 'orders-api-version-set'
  properties: {
    displayName: 'Orders API'
    versioningScheme: 'Segment'  // 'Segment' | 'Query' | 'Header'
    // versionQueryName: 'api-version'   // Required if scheme is 'Query'
    // versionHeaderName: 'Api-Version'  // Required if scheme is 'Header'
  }
}

// API v1 — active
resource apiV1 'Microsoft.ApiManagement/service/apis@2024-05-01' = {
  parent: apimService
  name: 'orders-api-v1'
  properties: {
    displayName: 'Orders API'
    apiVersion: 'v1'
    apiVersionSetId: apiVersionSet.id
    path: 'orders'
    protocols: ['https']
    // Import from OpenAPI spec
    format: 'openapi+json'
    value: loadTextContent('specs/orders-v1.json')
  }
}

// API v2 — active (successor)
resource apiV2 'Microsoft.ApiManagement/service/apis@2024-05-01' = {
  parent: apimService
  name: 'orders-api-v2'
  properties: {
    displayName: 'Orders API'
    apiVersion: 'v2'
    apiVersionSetId: apiVersionSet.id
    path: 'orders'
    protocols: ['https']
    format: 'openapi+json'
    value: loadTextContent('specs/orders-v2.json')
  }
}
```

### Azure API Center Registration

For API inventory and governance across the organization:

```
Tool: mcp_azure_mcp_documentation
Query: "Azure API Center register API versions lifecycle governance"
```

---

## OpenAPI Specification Versioning

### Required Fields

```yaml
openapi: "3.1.0"
info:
  title: Orders API
  version: "2.0.0" # Semantic version of the API contract
  x-api-version: "v2" # URL version segment (if using path versioning)
  x-lifecycle-stage: active # preview | active | deprecated | sunset
  x-deprecation-date: null # ISO 8601 date when deprecated
  x-sunset-date: null # ISO 8601 date when sunset
  contact:
    name: API Platform Team
    email: api-platform@contoso.com
```

### Contract Diff Validation (CI)

Add to CI pipelines to catch unintentional breaking changes:

```yaml
- name: Check for breaking API changes
  run: |
    npx @openapitools/openapi-diff \
      specs/orders-v2-main.json \
      specs/orders-v2.json \
      --fail-on-incompatible
```

> ⚠️ **Always version your OpenAPI specs in source control.** Comparing HEAD against main branch catches breaking changes before merge. Without CI diffing, breaking changes slip through code review.

---

## Governance Checklist

Before releasing a new API version or deprecating an existing one:

| #   | Check                                              | When              |
| --- | -------------------------------------------------- | ----------------- |
| 1   | Versioning strategy documented and consistent      | New API           |
| 2   | All changes classified as breaking or non-breaking | Every release     |
| 3   | Breaking changes ONLY in new major/date version    | Every release     |
| 4   | OpenAPI spec updated and diffed against previous   | Every release     |
| 5   | Migration guide published for breaking changes     | New major version |
| 6   | `Deprecation` and `Sunset` headers set             | Deprecation       |
| 7   | APIM version set configured correctly              | APIM deployments  |
| 8   | Monitoring query for deprecated version usage      | Deprecation       |
| 9   | Consumer notification sent                         | Deprecation       |
| 10  | 12-month deprecation window respected              | Sunset            |
| 11  | `410 Gone` returned after sunset                   | Post-sunset       |
| 12  | ADR documenting the versioning decision            | New API           |

---

## Guardrails

> ⚠️ **Never make breaking changes to a GA version.** If you need to change the contract, create a new version. "We'll just tell consumers to update" is not a versioning strategy.

> ⚠️ **Don't version too eagerly.** Every active version is a maintenance burden. If a change is non-breaking, add it to the current version. Only create v2 when v1's contract fundamentally can't accommodate the change.

> ⚠️ **Preview versions are not free passes.** Even preview APIs should follow versioning hygiene — date-stamp them and communicate breaking changes. Consumers build on previews and silent breakage erodes trust.

> ⚠️ **`Sunset` without `Deprecation` is a surprise.** Always deprecate first, with documented timeline, before sunsetting. Minimum 12 months between deprecation and sunset for external APIs.

---

## Currency

- **Date checked:** 2026-03-31
- **Sources:** Microsoft Learn MCP (`microsoft_docs_search`)
- **Authoritative references:** [Microsoft API Guidelines (vNext)](https://github.com/microsoft/api-guidelines), [Azure REST API versioning](https://learn.microsoft.com/azure/architecture/best-practices/api-design#versioning-a-restful-web-api)

### Verification Steps

1. Confirm Microsoft API Guidelines versioning recommendations are still current
2. Verify `Sunset` and `Deprecation` header standards align with latest IETF drafts
3. Check APIM version-set configuration syntax against latest GA API version

---

## Related Skills

- [apim-policy-authoring](../apim-policy-authoring/SKILL.md) — Policy implementation for version routing and deprecation headers
- [api-security-review](../api-security-review/SKILL.md) — Security review includes versioning hygiene checks
- [azure-apim-architecture](../azure-apim-architecture/SKILL.md) — APIM design decisions including versioning strategy
- [azure-adr](../azure-adr/SKILL.md) — Document versioning strategy as an ADR
- [dotnet-backend-patterns](../dotnet-backend-patterns/SKILL.md) — ASP.NET Core API versioning implementation
