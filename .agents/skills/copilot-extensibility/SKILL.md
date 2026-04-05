---
name: copilot-extensibility
description: >-
  Build declarative agents, custom engine agents, API plugins, and MCP-based actions for Microsoft 365 Copilot. USE FOR: creating M365 Copilot agents, adding API plugins to declarative agents, integrating MCP servers with Copilot, choosing between declarative and custom engine agents, configuring agent manifests, building Copilot connectors, designing Action UX experiences.
license: MIT
---

# Microsoft 365 Copilot Extensibility

Use this skill when building agents, plugins, or connectors that extend Microsoft 365 Copilot. This covers the full extensibility surface: declarative agents (Copilot-hosted), custom engine agents (self-hosted), API plugins, MCP server integrations, and Copilot connectors.

For backend AI agent implementation patterns (MAF, orchestration, tool integration), also use:

- `.github/skills/microsoft-agent-framework/SKILL.md`
- `.github/skills/maf-agentic-patterns/SKILL.md`

For designing your existing app to be consumable by agents and Copilot, also use:

- `.github/skills/app-as-skill/SKILL.md`

## When To Use

- Building a declarative agent for M365 Copilot
- Adding API plugins or MCP server actions to a declarative agent
- Choosing between declarative vs custom engine agent
- Creating agent manifests (declarative agent manifest, plugin manifest)
- Building Copilot connectors to ingest external data into Microsoft Graph
- Designing interactive UI widgets for declarative agents
- Publishing agents to the Microsoft 365 app store

## Agent Type Decision

| Factor             | Declarative Agent                                     | Custom Engine Agent                                               |
| ------------------ | ----------------------------------------------------- | ----------------------------------------------------------------- |
| Orchestrator       | Copilot's built-in orchestrator and models            | Your own orchestrator and models                                  |
| Hosting            | Hosted in Microsoft 365 (no infra needed)             | Self-hosted (Azure AI Foundry, Azure Bot Service, Copilot Studio) |
| Compliance         | Inherits M365 compliance, RAI, and security           | You own compliance, RAI, and security                             |
| Proactive messages | Not supported (user-initiated only)                   | Supported (agents can trigger actions autonomously)               |
| Channels           | M365 Copilot Chat, Teams, Word, Excel, Outlook        | M365 Copilot Chat, Teams, and external apps                       |
| Customization      | Limited to instructions, knowledge, and actions       | Fully customizable (models, orchestration, UX)                    |
| Tooling            | Agent Builder (low-code) or Agents Toolkit (pro-code) | Copilot Studio, Agents Toolkit, or custom code                    |

**Decision:** Use declarative agents when you want Copilot to orchestrate your domain knowledge and API calls within the M365 experience. Use custom engine agents when you need full control over the AI model, orchestration logic, or proactive behavior.

## Declarative Agent Anatomy

A declarative agent combines three configurable layers:

1. **Custom instructions** — shape Copilot's responses for your domain
2. **Custom knowledge** — connect M365 data sources (SharePoint, Teams messages, OneDrive, Copilot connectors, email, embedded files)
3. **Custom actions** — integrate APIs via plugins (OpenAPI or MCP) to read and write external systems

### Manifest Structure

```
appPackage/
├── manifest.json          # Teams/M365 app manifest (references the DA manifest)
├── declarativeAgent.json  # Declarative agent manifest (v1.6)
├── apiPlugin.json         # Plugin manifest (v2.4) — if using API actions
├── apiSpecFile.yaml       # OpenAPI spec for API plugin endpoints
└── color.png / outline.png
```

### Declarative Agent Manifest (v1.6)

Key properties:

- `instructions` — system prompt shaping agent behavior
- `capabilities` — 12 knowledge source types: Web search, OneDrive/SharePoint, Copilot connectors, Graphic art, Code interpreter, Dataverse, Teams messages, Email, People, Scenario models, Meetings, Embedded knowledge
- `actions` — references to plugin manifests (API plugins or MCP servers)
- `conversation_starters` — up to 12 starter prompts
- `behavior_overrides` — control knowledge source prioritization and model knowledge usage
- `disclaimers` — legal or usage disclaimers shown to users
- `worker_agents` — reference other declarative agents that this agent can delegate to (v1.6)
- `sensitivity_label` — Purview sensitivity label for agents with embedded files (v1.6)
- `user_overrides` — allow users to modify configured capabilities at runtime (v1.6)

### Plugin Manifest (v2.4)

Supports two action types:

- **OpenAPI-based** — standard REST API calls described by an OpenAPI spec
- **MCP server-based** — connect to any MCP-compatible server via URL

Key properties per function:

- `name`, `description` — used by Copilot to decide when to invoke
- `parameters` — mapped from OpenAPI or MCP tool schema
- `response_semantics` — control how Copilot renders results (Adaptive Cards, file references)
- `confirmation` — require user confirmation before execution (critical for mutations)

## Hard Rules

1. `MUST`: Choose agent type (declarative vs custom engine) based on the decision table above before implementation.
2. `MUST`: Use [Microsoft 365 Agents Toolkit](https://aka.ms/M365AgentsToolkit) (v6.3+) for pro-code declarative agent development.
3. `MUST`: Use declarative agent manifest v1.6 for new agents. Older versions (v1.4, v1.5) are supported but miss capabilities like `worker_agents`, `sensitivity_label`, and `embedded knowledge`.
4. `MUST`: Use plugin manifest v2.4 for API plugins (adds MCP server support and enhanced response semantics).
5. `MUST`: Write specific, behavior-shaping instructions in `instructions` field. Generic instructions like "be helpful" are wasted tokens.
6. `MUST`: For API plugins, provide an accurate OpenAPI spec with rich operation descriptions. Copilot uses these descriptions to decide when to call your API. Poor descriptions = poor invocation accuracy.
7. `MUST`: Set `confirmation.type` to `"AdaptiveCard"` or `"message"` for any mutating API operation (POST, PUT, PATCH, DELETE). Never auto-execute mutations without user confirmation.
8. `MUST`: Set `response_semantics` for each function to control how results render. Use Adaptive Cards for structured data and plain text for simple responses.
9. `MUST`: For MCP server plugins, use a publicly accessible HTTPS URL. Local MCP servers are for development only.
10. `MUST`: Implement OAuth 2.0 or API key authentication for API plugins that access protected resources. Define auth in the `auth` section of the plugin manifest.
11. `MUST NOT`: Treat API plugins and MCP plugins as separate API surfaces. They should wrap the same underlying API. One API, multiple access paths.
12. `MUST`: Test agents locally using Agents Toolkit's local debug before deploying to a tenant.
13. `MUST`: For custom engine agents, implement the Bot Framework protocol for Teams/Copilot channel compatibility.
14. `SHOULD`: Use `behavior_overrides` to prioritize provided knowledge sources over general Copilot knowledge when domain accuracy matters.
15. `SHOULD`: Provide conversation starters that demonstrate the agent's capabilities to users.
16. `SHOULD`: Keep instructions under 8000 characters. Longer instructions degrade response quality.

## MCP Server Integration Pattern

Declarative agents can consume MCP servers directly via the plugin manifest:

```json
{
  "schema_version": "v2.4",
  "name_for_human": "Project Tracker",
  "functions": [
    {
      "name": "list_projects",
      "description": "List all active projects with status and timeline",
      "mcp_server": {
        "url": "https://api.example.com/mcp/"
      }
    }
  ]
}
```

This means any MCP-compatible backend (including MAF agents with MCP endpoints) can be surfaced as Copilot actions without building a separate API plugin.

## Custom Engine Agent Pattern

For agents requiring custom orchestration (MAF-based, Azure AI Foundry-based):

1. Build the agent backend using MAF (see `microsoft-agent-framework` skill)
2. Host on Azure Container Apps, Azure Functions, or Azure Bot Service
3. Register as a custom engine agent in the Teams/M365 app manifest
4. Optionally use Copilot APIs (Chat API, Retrieval API, Search API) to access M365 enterprise data

Custom engine agents can send proactive messages and support asynchronous workflows, which declarative agents cannot.

## Converting Between Agent Types

- **Declarative → Custom Engine**: Use when you outgrow Copilot's orchestrator (need proactive messages, custom models, or complex workflows). Follow the conversion guide at `learn.microsoft.com/microsoft-365/copilot/extensibility/convert-declarative-agent`.
- **Declarative → Copilot Studio**: Use "Copy to full experience" feature for advanced lifecycle management, analytics, and governance controls.

## Knowledge Source Selection

| Source                    | Best For                                                  |
| ------------------------- | --------------------------------------------------------- |
| SharePoint (sites, files) | Organizational documents, policies, procedures            |
| OneDrive                  | Personal files and team-shared documents                  |
| Teams messages            | Channel conversations, meeting chats                      |
| Email                     | Communication history and context                         |
| Meetings                  | Meeting transcripts and action items                      |
| Copilot connectors        | External system data (ServiceNow, Salesforce, Jira, etc.) |
| Embedded files            | Static reference documents shipped with the agent         |
| Dataverse                 | Structured business data                                  |
| People                    | Organizational directory data                             |
| Web content               | Public website information                                |

## Interactive UI Widgets (March 2026+)

Declarative agents can render interactive UI widgets using the OpenAI Apps SDK. Widgets render inline or full-screen within M365 Copilot, providing richer experiences than Adaptive Cards alone.

## Copilot APIs (Preview)

For custom engine agents that need M365 enterprise data:

- **Chat API** — multi-turn conversations with Copilot, grounded in enterprise search
- **Retrieval API** — retrieve content from SharePoint and Copilot connectors
- **Search API** — semantic search across OneDrive content

## Known Pitfalls

| Area                    | Pitfall                                                                                                                                              | Mitigation                                                                                                                                               |
| ----------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Instruction quality     | Generic or overly long instructions cause Copilot to ignore domain-specific behavior and fall back to general responses                              | Write specific, concise instructions (under 8000 chars) that define persona, constraints, and output format                                              |
| OpenAPI descriptions    | Poor operation descriptions in OpenAPI specs cause Copilot to invoke the wrong API or skip invocation entirely                                       | Write descriptions from the user's perspective ("Find open tickets assigned to the current user") not the developer's ("GET /api/tickets?assignee={id}") |
| Auth configuration      | Missing or misconfigured OAuth causes silent 401 failures that Copilot reports as "I couldn't find that information"                                 | Test auth flow end-to-end in Agents Toolkit local debug before deploying                                                                                 |
| Mutation safety         | API plugins that auto-execute POST/PUT/DELETE without confirmation risk unintended data changes                                                      | Always set `confirmation` for mutating operations                                                                                                        |
| Knowledge source limits | Too many knowledge sources dilute retrieval quality. Copilot retrieves from all sources equally by default                                           | Use `behavior_overrides` to prioritize key sources; keep sources focused and non-overlapping                                                             |
| MCP server availability | MCP servers must be publicly accessible over HTTPS for deployed agents; local servers only work in debug                                             | Deploy MCP servers to a hosted environment with proper TLS before publishing the agent                                                                   |
| Manifest version drift  | Using outdated manifest versions misses capabilities (MCP support requires plugin v2.4, `worker_agents` requires DA v1.6, meetings require DA v1.5+) | Pin to latest stable manifest version (DA v1.6, plugin v2.4) and check what's-new page before starting                                                   |

---

## Currency

- **Date checked:** 2026-04-02
- **Sources:** [M365 Copilot Extensibility Overview](https://learn.microsoft.com/microsoft-365/copilot/extensibility/agents-overview), [What's New](https://learn.microsoft.com/microsoft-365/copilot/extensibility/whats-new), [Plugin Manifest v2.4](https://learn.microsoft.com/microsoft-365/copilot/extensibility/plugin-manifest-2.4), [Declarative Agent Manifest v1.6](https://learn.microsoft.com/microsoft-365/copilot/extensibility/declarative-agent-manifest-1.6)
- **Authoritative references:** [Microsoft Learn — M365 Copilot Extensibility](https://learn.microsoft.com/microsoft-365/copilot/extensibility/)

### Verification Steps

1. Confirm latest declarative agent manifest schema version (currently v1.6)
2. Confirm latest plugin manifest schema version (currently v2.4)
3. Check for new knowledge source types or capabilities (v1.6 added embedded knowledge)
4. Verify Agents Toolkit version requirements (currently v6.3+)
5. Check custom engine agent GA status (GA since July 2025) and any new API surfaces

---

## Related Skills

- [Microsoft Agent Framework](../microsoft-agent-framework/SKILL.md) — build agent backends with MAF
- [MAF Agentic Patterns](../maf-agentic-patterns/SKILL.md) — orchestration pattern selection
- [MAF AI Integration](../maf-ai-integration/SKILL.md) — AI client integration patterns
- [App as Skill](../app-as-skill/SKILL.md) — design apps to be consumable by agents and Copilot
