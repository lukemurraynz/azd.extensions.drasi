---
name: composable-app-architecture
description: >-
  Reference architecture for decomposing applications into composable layers (UX, Logic, Data, Infra) that can be independently deployed and consumed by humans, agents, and M365 Copilot. USE FOR: designing composable applications, decomposing monoliths into composable services, choosing between composable vs monolithic architecture, planning AI-era application structure.
license: MIT
---

# Composable App Architecture

Use this skill when designing a new application (or restructuring an existing one) that should participate in the composable ecosystem: consumable by humans via traditional UI, by M365 Copilot via Action UX, and by autonomous agents operating independently.

This skill is a reference architecture for the **decision** of whether and how to decompose. It does not replace implementation skills.

For implementation, use the appropriate skill:

- `.github/skills/copilot-extensibility/SKILL.md` — building Copilot agents and plugins
- `.github/skills/app-as-skill/SKILL.md` — making your API agent-consumable
- `.github/skills/maf-agentic-patterns/SKILL.md` — agent orchestration patterns
- `.github/skills/dotnet-aspire/SKILL.md` — service composition and orchestration
- `.github/skills/azure-container-apps/SKILL.md` — managed hosting

## When To Use

- Starting a greenfield project and deciding on architecture
- Evaluating whether an existing app should become composable
- Planning which app capabilities to expose as agent skills
- Choosing between monolithic vs composable architecture for a specific project
- Designing the layer boundaries for a composable system

## When NOT To Use

- Building a simple CRUD app with no agent integration planned
- Internal tools with a single consumer (just build it directly)
- When "composable" is aspirational but has no concrete consumer

**The default should be a well-structured monolith with a good API.** Only decompose when you have concrete consumers that need composability.

## The Composable Layers

Traditional applications combine four concerns into one deployment. Composable applications separate them so each can be independently consumed:

```
┌──────────────┐
│     UX       │  How users interact (web UI, Copilot Action UX, Teams)
├──────────────┤
│    Logic     │  Business rules, workflows, orchestration
├──────────────┤
│    Data      │  Storage, retrieval, caching, search
├──────────────┤
│    Infra     │  Hosting, networking, identity, observability
└──────────────┘
```

### Layer Independence

Each layer should be independently:

- **Deployable** — update logic without redeploying UX
- **Consumable** — agents call logic directly without going through the UX
- **Scalable** — scale data tier independently of logic tier
- **Replaceable** — swap UX from web to Copilot plugin without touching logic

### Layer Contracts

Layers communicate through well-defined contracts:

| Boundary         | Contract Type                | Example                        |
| ---------------- | ---------------------------- | ------------------------------ |
| UX ↔ Logic       | REST API (OpenAPI described) | Frontend calls backend API     |
| Logic ↔ Data     | Internal API or ORM          | Service queries database       |
| Logic → External | MCP tools, A2A protocol      | Agent calls external service   |
| Agent → Logic    | MCP tool / API plugin        | Copilot invokes business logic |
| Agent → Data     | Copilot connector            | Agent searches indexed data    |

## Decision Framework

### Should This App Be Composable?

| Signal                                                        | Composable | Monolith |
| ------------------------------------------------------------- | ---------- | -------- |
| Multiple UX surfaces needed (web + Copilot + Teams + agents)  | Yes        | —        |
| Business logic needs to be callable by autonomous agents      | Yes        | —        |
| Data needs to be searchable by Copilot (enterprise knowledge) | Yes        | —        |
| Single team, single UI, internal tool                         | —          | Yes      |
| Prototype or MVP                                              | —          | Yes      |
| No identified agent or Copilot consumer                       | —          | Yes      |

**Decision: Start monolithic with a clean API boundary. Decompose into composable layers only when a second consumer (agent, Copilot, partner app) materializes.**

### Composability Spectrum

Not all-or-nothing. You can adopt composability incrementally:

1. **Level 0 — Monolith with API**: Traditional app. Good API. Not composable, but ready to be.
2. **Level 1 — API as Skill**: Same app, but OpenAPI spec is agent-quality and you publish an API plugin for Copilot. See `app-as-skill` skill.
3. **Level 2 — Data as Knowledge**: Your data is indexed and available to Copilot via Copilot connectors. Agents can search your data without calling your API.
4. **Level 3 — Logic as Service**: Business logic exposed as MCP tools or standalone microservices that multiple consumers (UI, agents, other services) call independently.
5. **Level 4 — Full Decomposition**: All four layers independently deployed and consumed. UX is one of many consumers. Agents compose logic and data independently.

**Most apps should aim for Level 1-2.** Level 3-4 is justified only for platform-scale applications with many consumers.

## Three Deployment Surfaces

Composable apps serve three types of consumers (shown in the architecture reference):

### 1. Action UX (Copilot-Orchestrated)

Copilot composes work tasks across app skills. Your app provides skills; Copilot provides the UX and orchestration.

- **Implementation**: Declarative agent + API plugins
- **Your responsibility**: Well-described API, OpenAPI spec, plugin manifest
- **Copilot's responsibility**: UX rendering, intent detection, orchestration
- **Skill**: `copilot-extensibility`

### 2. Autonomous Agents

Agents operate independently across apps, executing tasks and reporting results to humans and apps.

- **Implementation**: MCP tools wrapping your API, A2A protocol for agent-to-agent
- **Your responsibility**: MCP tool definitions, auth delegation, idempotent endpoints
- **Agent's responsibility**: Task planning, tool selection, error recovery
- **Skill**: `app-as-skill`, `microsoft-agent-framework`

### 3. Traditional App (Human-Driven)

Users interact directly via web UI, mobile app, or desktop app.

- **Implementation**: Standard frontend + backend patterns
- **Your responsibility**: Full UX, API, data, hosting
- **Skills**: `typescript-react-patterns`, `dotnet-backend-patterns`

## Intelligence Layer

The intelligence layer is a shared AI service that multiple app components consume, rather than embedding AI logic into each component independently.

### When to Use a Shared Intelligence Layer

| Signal                                                                                    | Shared Layer | Embedded AI |
| ----------------------------------------------------------------------------------------- | ------------ | ----------- |
| Multiple services need the same AI capability (summarization, classification, extraction) | Yes          | —           |
| AI prompts and model config should be managed centrally                                   | Yes          | —           |
| Single service, single AI use case                                                        | —            | Yes         |
| Rapid prototyping                                                                         | —            | Yes         |

### Implementation Pattern

- Centralize AI client, prompt management, and model configuration behind a shared service
- Consumers call the intelligence layer via internal API or MCP tools
- Manage prompts as configuration (Azure App Configuration, Key Vault) not as embedded code
- Track model versions, prompt versions, and configuration changes for auditability

For implementation patterns, see `maf-ai-integration` skill (Intelligence Layer section).

## Managed Host

In the composable model, hosting is abstracted from the application:

| Hosting Model                  | When To Use                               | Skill                      |
| ------------------------------ | ----------------------------------------- | -------------------------- |
| Azure Container Apps           | Self-hosted backend services and APIs     | `azure-container-apps`     |
| M365 Platform (Copilot-hosted) | Declarative agents (no infra needed)      | `copilot-extensibility`    |
| Azure Functions                | Event-driven logic, durable agent hosting | `azure-functions-patterns` |
| AKS                            | Complex workloads requiring K8s control   | `aks-cluster-architecture` |

**Decision**: Default to ACA for self-hosted services and Copilot-hosted for declarative agents. Only use AKS when ACA constraints are not sufficient.

## Hard Rules

1. `MUST NOT`: Decompose into composable layers prematurely. A well-structured monolith with a good API IS the right starting point.
2. `MUST`: Design API contracts (OpenAPI) at layer boundaries before implementation. Contracts are the architecture.
3. `MUST`: Keep one API surface, not separate "human API" and "agent API" endpoints. See `app-as-skill` skill.
4. `MUST`: Choose composability level (0-4) explicitly during design and document the rationale.
5. `MUST`: For Level 2+, define which data becomes Copilot knowledge and which stays API-only. Not all data should be indexed.
6. `MUST`: For Level 3+, define service boundaries using domain-driven design (bounded contexts), not technology layers.
7. `SHOULD`: Use .NET Aspire for local multi-service composition when building Level 3+ systems.
8. `SHOULD`: Evaluate whether a shared Intelligence Layer is warranted before embedding AI into individual services.

## Anti-Patterns

| Pattern                         | Problem                                                                                                 | Fix                                                                                |
| ------------------------------- | ------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| "Composable by default"         | Every simple app gets decomposed into microservices, adding complexity with no consumer                 | Start monolithic. Decompose when a second consumer materializes.                   |
| Dual API surfaces               | "Human API" and "Agent API" diverge over time, creating maintenance burden and inconsistencies          | One API. Describe it well for all consumers.                                       |
| Data firehose to Copilot        | Indexing all data as Copilot knowledge creates noise and degrades retrieval quality                     | Curate which data becomes knowledge. Quality over quantity.                        |
| Premature Intelligence Layer    | Building a shared AI service when only one component uses AI                                            | Embed AI in the component. Extract to shared layer when a second consumer appears. |
| Technology-driven decomposition | Splitting by technology (separate "API service," "worker service," "data service") instead of by domain | Split by business capability (orders, inventory, shipping), not by technical role. |

## Known Pitfalls

| Area               | Pitfall                                                                                        | Mitigation                                                                                |
| ------------------ | ---------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- |
| Over-decomposition | Splitting a small app into 5 microservices adds distributed systems complexity with no benefit | Use the composability spectrum: most apps are Level 0-2                                   |
| Contract drift     | Layer contracts (OpenAPI specs) drift from implementation, breaking agent consumers silently   | Generate specs from code and validate in CI                                               |
| Auth complexity    | Delegated auth across composable layers creates token exchange chains that are hard to debug   | Use managed identity for service-to-service; delegated auth only at the edge (user → API) |
| Observability gaps | Requests spanning multiple composable layers lose trace context                                | Use OpenTelemetry with W3C Trace Context propagation across all layer boundaries          |

---

## Currency

- **Date checked:** 2026-04-02
- **Sources:** Microsoft Build 2025/2026 Composable Apps vision, [M365 Copilot Extensibility](https://learn.microsoft.com/microsoft-365/copilot/extensibility/), [Azure Container Apps](https://learn.microsoft.com/azure/container-apps/), [.NET Aspire](https://learn.microsoft.com/dotnet/aspire/)
- **Authoritative references:** [M365 Copilot Agents Overview](https://learn.microsoft.com/microsoft-365/copilot/extensibility/agents-overview)

### Verification Steps

1. Confirm M365 Copilot extensibility surface and composable app patterns
2. Verify managed hosting options and any new platform-hosted patterns
3. Check for published composable architecture reference implementations

---

## Related Skills

- [Copilot Extensibility](../copilot-extensibility/SKILL.md) — build Copilot agents and plugins
- [App as Skill](../app-as-skill/SKILL.md) — make your API agent-consumable
- [Microsoft Agent Framework](../microsoft-agent-framework/SKILL.md) — build agent backends
- [MAF AI Integration](../maf-ai-integration/SKILL.md) — AI integration and Intelligence Layer
- [.NET Aspire](../dotnet-aspire/SKILL.md) — service composition and orchestration
- [Azure Container Apps](../azure-container-apps/SKILL.md) — managed container hosting
