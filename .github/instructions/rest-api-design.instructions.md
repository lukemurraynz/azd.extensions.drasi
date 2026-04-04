---
applyTo: "**/*.cs,**/*.ts,**/*.tsx,**/*.js,**/*.jsx,**/*.tsp"
description: REST API design standards aligned with Microsoft API Guidelines (vNext) covering versioning, pagination, LROs, idempotency, error shapes by plane (data plane, Graph, ARM), OpenAPI specification quality, and TypeSpec as the preferred spec-first tooling
---

# REST API Design (Microsoft vNext-Aligned)

When adding or changing HTTP APIs, pick the guideline set based on the **plane** and do not mix styles within the same API surface.

## Azure Data Plane (Default for New APIs)

- Required `api-version=YYYY-MM-DD[-preview]` query parameter
- Pagination: `{ "value": [...], "nextLink": "https://..." }` (absolute nextLink, omitted on last page, never null; avoid global counts by default)
- LRO: `202 Accepted` + `operation-location` (absolute URL) + `Retry-After`; `api-version` included in poll URL
- Idempotent POST via `Repeatability-Request-ID` + `Repeatability-First-Sent`
- Errors: RFC 9457 Problem Details (`application/problem+json`) + `x-error-code` header + `errorCode` extension

## Microsoft Graph / OData

- Follow Graph guidelines (OData conventions, `/v1.0` and `/beta`, `@odata.nextLink`, query options)

## Azure Management Plane (ARM Resource Provider APIs)

- Follow the Azure Resource Provider Contract
- Do not apply Azure data plane LRO/header conventions blindly

## OpenAPI Specification Quality

- **Target OpenAPI 3.0.x** unless the full toolchain (APIM, code generators, Copilot plugin manifest) has verified 3.1 support. OpenAPI 3.1 aligns with JSON Schema 2020-12 but tooling support is still inconsistent.
- **Prefer spec-first with TypeSpec** for cross-team or public APIs. Write `.tsp` files, emit OpenAPI via `@typespec/openapi3`, then generate server stubs from the emitted spec. TypeSpec enforces documentation, reusable schemas, and semantic operation IDs structurally. For Azure APIs, use `@azure-tools/typespec-azure-core` (data plane) or `@azure-tools/typespec-azure-resource-manager` (ARM) to get built-in compliance with Azure API Guidelines. Code-first with annotation-generated specs (Swashbuckle, NSwag, tsoa) is acceptable for internal APIs, but the generated spec is still the contract. For full TypeSpec patterns (scaffolding, versioning, LROs, CI integration), see `.github/skills/typespec-api-design/SKILL.md`.
- **Every operation** must have: `operationId` (camelCase, verb-noun, describes the user-level action), `summary` (one sentence), `description` (behavior, side-effects, auth requirements), and at least one response `example`.
- **Every parameter and schema property** must have `description`. Use `format` (`date-time`, `uuid`, `uri`, `email`) to communicate semantics, not just `type: string`.
- **Reusable schemas**: define models under `components/schemas` and reference via `$ref`. Do not inline the same object shape in multiple operations.
- **Enums**: document the meaning of each value in the schema `description`. Use string enums, not integer codes.
- **Nullable handling**: in OpenAPI 3.0.x use `nullable: true`; in 3.1 use `type: ["string", "null"]`. Be explicit — omitting nullability when the API can return null is a contract bug.
- **Security schemes**: define `securitySchemes` in `components` (OAuth2 with correct flows, bearer, or API key). Apply `security` at the operation or global level. Do not leave specs without auth documentation.
- **Spec validation**: lint specs with Spectral, Redocly, or equivalent in CI. At minimum enforce: no missing descriptions, no unused schemas, no invalid examples, security defined.

## Agent-Consumable API Quality

APIs are increasingly consumed by AI agents and M365 Copilot (via API plugins and MCP tools) alongside human-driven UIs. A well-specified OpenAPI document (per the section above) is already agent-consumable. Reinforce: semantic `operationId` names, `description` and `example` on every operation and parameter, and consistent RFC 9457 error responses. Do not create separate "agent endpoints" — one API surface serves all consumers. For detailed guidance, see `.github/skills/app-as-skill/SKILL.md`.

## TypeSpec as Spec-First Tooling

For new API surfaces (or major redesigns), use TypeSpec as the canonical API definition language.

- **Generic REST API**: Use `@typespec/http` + `@typespec/openapi3`. Define models, routes, and operations in `.tsp` files.
- **Azure data-plane API**: Add `@azure-tools/typespec-azure-core` for standard resource lifecycle operations (`ResourceRead`, `ResourceCreateOrUpdate`, `ResourceDelete`, `ResourceList`), LRO templates, versioning decorators (`@added`, `@removed`), and built-in linting that enforces Azure API Guidelines.
- **Azure ARM API**: Add `@azure-tools/typespec-azure-resource-manager` for ARM-specific resource templates and common type versions.
- **Versioning**: Define a `Versions` enum with date-based strings (`"2024-01-15"`) and apply `@versioned()` from the first version. Annotate changes with `@added()` / `@removed()`.
- **CI gate**: Run `tsp compile .` and `tsp lint .` in the build pipeline. Treat failures as build-blocking. Validate emitted OpenAPI with Spectral or equivalent.
- **Do not hand-edit emitted files**: The `tsp-output/` directory is a build artifact. All changes go through `.tsp` source.

Full guidance (project scaffolding, operation templates, LROs, traits, migration from existing OpenAPI): `.github/skills/typespec-api-design/SKILL.md`.

## References

- `microsoft/api-guidelines` (vNext): https://github.com/microsoft/api-guidelines/tree/vNext
- Azure data plane: https://github.com/microsoft/api-guidelines/blob/vNext/azure/Guidelines.md
- Graph: https://github.com/microsoft/api-guidelines/blob/vNext/graph/GuidelinesGraph.md
- ARM RP contract: https://github.com/cloud-and-ai-microsoft/resource-provider-contract
- TypeSpec language: https://typespec.io/docs
- TypeSpec Azure libraries: https://azure.github.io/typespec-azure/docs/intro/
- TypeSpec API design skill: `.github/skills/typespec-api-design/SKILL.md`
