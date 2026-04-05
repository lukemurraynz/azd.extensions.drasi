---
name: app-as-skill
description: >-
  Design applications so their capabilities are consumable as skills by agents and M365 Copilot. One API surface, multiple consumers. USE FOR: exposing app capabilities as MCP tools or API plugins, making an existing API agent-consumable, designing skill manifests for existing services, auth delegation for agent-to-app calls, testing agent consumption of app APIs.
license: MIT
---

# App as Skill

Use this skill when you want an existing or new application to expose its capabilities so that AI agents and M365 Copilot can consume them, without building a separate "agent API." The core principle: **one API surface, well-described, serves both humans and agents.**

For building the Copilot-side integration (declarative agents, plugin manifests), also use:

- `.github/skills/copilot-extensibility/SKILL.md`

For building standalone AI agents that consume skills, also use:

- `.github/skills/microsoft-agent-framework/SKILL.md`

## When To Use

- Making an existing REST API consumable by M365 Copilot or autonomous agents
- Adding MCP tool definitions that wrap existing API endpoints
- Writing OpenAPI specs optimized for agent consumption
- Designing auth delegation so agents call your API with delegated tokens
- Testing that agents can discover. understand, and use your API effectively

## Integration with TypeSpec

If the API is defined using TypeSpec:

1. Treat TypeSpec as the single source of truth
2. Run `tsp compile` to generate OpenAPI
3. Use the generated OpenAPI as the input for:
   - MCP tool generation
   - Copilot plugin manifests
4. NEVER modify generated OpenAPI directly
5. If changes are required, update TypeSpec and recompile

## Core Principle: One API, Multiple Consumers

```
┌─────────────────────────────────────────────┐
│              Your Application               │
│                                             │
│   ┌───────────────────────────────────┐     │
│   │         Single REST API           │     │
│   │   (well-described, well-typed)    │     │
│   └──────┬──────────┬──────────┬──────┘     │
│          │          │          │             │
│     ┌────┴───┐ ┌────┴───┐ ┌───┴─────┐      │
│     │ Web UI │ │ Agent  │ │ Copilot │      │
│     │(human) │ │ (MCP)  │ │(plugin) │      │
│     └────────┘ └────────┘ └─────────┘      │
└─────────────────────────────────────────────┘
```

You do NOT build a second API. You make your existing API discoverable and self-describing. Agents consume the same routes, same contracts, same auth.

## Hard Rules

1. `MUST NOT`: Create separate "agent endpoints" or "agent controllers." Agents call the same API surface as your UI and other consumers.
2. `MUST`: Write rich OpenAPI descriptions on every operation. Agent orchestrators (Copilot, MAF) use these descriptions to decide when to call your API. Poor descriptions = poor invocation.
3. `MUST`: Use semantic operation IDs in OpenAPI specs that describe the action from the user's perspective (e.g., `listOpenTickets`, `createInvoice`, `getProjectStatus`) not implementation names (e.g., `get_v2_tickets_filtered`).
4. `MUST`: Include `description` on every parameter and response schema property. Agents need these to map natural language to API parameters.
5. `MUST`: Include `examples` in OpenAPI request/response schemas. Examples help agent orchestrators generate correct payloads.
6. `MUST`: Return structured, typed responses (not HTML, not untyped JSON blobs). Agents parse structured data; they cannot parse rendered HTML.
7. `MUST`: Use RFC 9457 Problem Details for errors. Agents need actionable error information to decide whether to retry, ask the user, or fail.
8. `MUST`: Support idempotent operations where applicable. Agents retry on transient failures. POST endpoints should support `Repeatability-Request-ID` or equivalent.
9. `MUST`: Support delegated auth (OAuth 2.0 on-behalf-of or token exchange) so agents act with the user's identity and permissions. Do not create service-level "agent accounts" with elevated permissions.
10. `MUST NOT`: Rely on session state, cookies, or CSRF tokens for API auth. Agents are stateless HTTP clients that pass bearer tokens.
11. `MUST`: Expose capabilities at granular, composable endpoints. Prefer `GET /projects/{id}/tasks` over `GET /projects/{id}?include=tasks,members,timeline,budget`. Agents compose multiple small calls; they don't parse mega-responses well.
12. `SHOULD`: Implement pagination following Microsoft API Guidelines (`value`/`nextLink` pattern). Agents handle paginated responses natively if the pattern is standard.
13. `SHOULD`: Include `x-{app}-capability` or equivalent metadata in API responses that helps agents understand what actions are available on returned objects.

## Making Your API Agent-Consumable: Checklist

### 1. OpenAPI Spec Quality

Your OpenAPI spec is the primary interface agent orchestrators use to understand your API. Treat it as a first-class product artifact.

- [ ] Every operation has a clear `summary` and `description` written from the user's perspective
- [ ] Operation IDs are semantic and action-oriented (`listActiveProjects`, not `getProjectsV2`)
- [ ] Every parameter has a `description` and, where applicable, `example` values
- [ ] Response schemas have `description` on every property
- [ ] Error responses use RFC 9457 Problem Details with `type` URIs that agents can match
- [ ] `servers` section includes the production base URL (agents need absolute URLs)
- [ ] Spec is validated and published alongside the API (not generated ad-hoc)

### 2. MCP Tool Wrapping

To expose your API as MCP tools (for MAF agents, Copilot MCP plugins, or any MCP-compatible consumer):

```json
{
  "name": "list_active_projects",
  "description": "List all active projects the user has access to, with status and timeline",
  "inputSchema": {
    "type": "object",
    "properties": {
      "status": {
        "type": "string",
        "enum": ["active", "completed", "archived"],
        "description": "Filter by project status. Defaults to active."
      }
    }
  }
}
```

Key principles:

- MCP tool definitions wrap existing API endpoints. The tool calls your API internally.
- Tool names should match OpenAPI operation IDs where possible.
- Tool descriptions should match OpenAPI operation descriptions.
- Keep input schemas minimal. Only expose parameters agents actually need.

### 3. Auth Delegation

Agents must act on behalf of users, not as super-users:

- **M365 Copilot (API plugins)**: Configure OAuth 2.0 in the plugin manifest. Copilot handles the auth flow and passes tokens to your API.
- **MAF agents (MCP tools)**: The agent runtime passes the user's delegated token when calling your MCP tools. Your API validates the token as normal.
- **Service-to-service**: If an agent needs to call your API without a user context (batch processing, scheduled tasks), use managed identity with minimal RBAC scope.

### 4. Response Design for Agents

Agents parse responses differently than humans:

- Return typed, flat-ish JSON. Deeply nested responses are harder for agents to navigate.
- Include IDs on all returned objects so agents can make follow-up calls.
- Include `_links` or action hints that tell agents what operations are available on returned objects.
- Keep response payloads focused. Agents process everything returned; large payloads waste tokens and degrade accuracy.

### 5. Error Design for Agents

Agents need errors they can act on:

```json
{
  "type": "https://api.example.com/errors/insufficient-permissions",
  "title": "Insufficient Permissions",
  "status": 403,
  "detail": "User does not have 'project.write' permission on project 'proj-123'.",
  "instance": "/projects/proj-123/tasks"
}
```

- Use specific `type` URIs so agents can match error categories programmatically.
- Include `detail` with enough context for the agent to explain the problem to the user.
- For 422 validation errors, include field-level error details so agents can fix and retry.

## Testing Agent Consumption

Before declaring your API agent-ready, test with real agent consumers:

1. **OpenAPI lint**: Run spectral or equivalent linter with "agent-consumable" rules (descriptions on all operations/params, semantic operationIds, examples present).
2. **Copilot plugin test**: Add your API as a plugin to a declarative agent and test with natural language prompts that should trigger each operation.
3. **MCP tool test**: Register your MCP tools with an MAF agent and verify tool discovery, invocation, and response parsing.
4. **Auth flow test**: Verify delegated auth works end-to-end (token acquisition, token validation, permission enforcement, token refresh).
5. **Error handling test**: Send invalid requests and verify agents receive actionable error responses, not generic 500s.
6. **Idempotency test**: Send the same mutating request twice with the same idempotency key and verify correct behavior (no duplicates, no errors).

## Anti-Patterns

| Pattern                             | Problem                                                            | Fix                                                                                              |
| ----------------------------------- | ------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------ |
| Separate `/api/agent/` routes       | Creates two API surfaces to maintain, test, and secure             | Remove agent routes. Improve OpenAPI descriptions on existing routes.                            |
| Agent-specific DTOs                 | Drift between "human DTOs" and "agent DTOs" causes inconsistencies | One DTO per operation. If agents need less data, use sparse fieldsets (`?fields=id,name,status`) |
| Service account for agents          | Over-privileged access, no per-user audit trail                    | Use delegated auth (OBO/token exchange). Agents act as the user.                                 |
| HTML error pages                    | Agents cannot parse HTML error responses                           | Return RFC 9457 JSON error responses on all API routes (including 404, 500)                      |
| Missing OpenAPI descriptions        | Agents can't determine when to call your API                       | Treat OpenAPI descriptions as product copy. Review them like UI text.                            |
| Mega-endpoints returning everything | Agents waste tokens parsing irrelevant data and lose accuracy      | Decompose into focused endpoints. Let agents compose multiple calls.                             |

## Known Pitfalls

| Area                     | Pitfall                                                                                       | Mitigation                                                                                                        |
| ------------------------ | --------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| OpenAPI spec drift       | Generated specs diverge from actual API behavior over time, causing agent invocation failures | Generate specs from code (Swashbuckle, NSwag, FastAPI) and validate in CI                                         |
| Auth token lifetime      | Delegated tokens expire during long agent workflows, causing mid-operation auth failures      | Implement token refresh in the MCP tool layer or use short-lived operations                                       |
| Rate limiting            | Agents retry aggressively and can exhaust API rate limits faster than human users             | Implement per-client rate limits with `Retry-After` headers; agents respect standard retry semantics              |
| Breaking API changes     | Agents break silently on API changes because there's no UI to show deprecation warnings       | Use API versioning. Keep deprecated endpoints active during migration. Communicate via OpenAPI `deprecated` flag. |
| Over-exposing operations | Exposing all API operations as agent tools creates confusion; agents invoke wrong tools       | Curate which operations become tools. Start with read operations, add mutations incrementally.                    |

---

## Currency

- **Date checked:** 2026-04-02
- **Sources:** [M365 Copilot Extensibility — Plugins Overview](https://learn.microsoft.com/microsoft-365/copilot/extensibility/overview-plugins), [MCP Protocol](https://modelcontextprotocol.io/), [Microsoft API Guidelines (vNext)](https://github.com/microsoft/api-guidelines/tree/vNext)
- **Authoritative references:** [Plugin Manifest v2.4](https://learn.microsoft.com/microsoft-365/copilot/extensibility/plugin-manifest-2.4), [MCP Specification](https://modelcontextprotocol.io/specification)

### Verification Steps

1. Confirm plugin manifest v2.4 schema for MCP server integration
2. Verify OAuth 2.0 on-behalf-of flow for API plugins
3. Check for any new agent-consumability requirements in Microsoft API Guidelines

---

## Related Skills

- [Copilot Extensibility](../copilot-extensibility/SKILL.md) — build the Copilot-side integration
- [REST API Design](../../instructions/rest-api-design.instructions.md) — API design standards
- [Microsoft Agent Framework](../microsoft-agent-framework/SKILL.md) — build agent backends
- [MAF AI Integration](../maf-ai-integration/SKILL.md) — AI client integration patterns
