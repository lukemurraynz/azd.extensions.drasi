---
name: maf-agentic-patterns
description: >-
  Design and implement Microsoft Agent Framework orchestration patterns and protocol integrations (AG-UI, MCP, A2A) with production guardrails. USE FOR: selecting between Sequential/Concurrent/Handoff/Group Chat patterns, structuring multi-agent systems, or preventing common MAF traps.
---

# Microsoft Agent Framework Agentic Patterns

Use this skill for architecture/design decisions where agent behavior and coordination patterns are the main concern.

For AI integration implementation details (prompting, confidenceSource propagation, fallback behavior, UI provenance), also use:

- `.github/skills/maf-ai-integration/SKILL.md`

## When To Use

- Choosing orchestration mode (Sequential, Concurrent, Handoff, Group Chat, Magentic)
- Choosing hosting model (in-process agent runtime vs Foundry hosted persistent agents vs Agent Harness)
- Deciding protocol boundary (AG-UI vs MCP vs A2A)
- Designing multi-agent run state, approval flow, and handoff behavior
- Reviewing implementations for known MAF pitfalls

## Hard Rules

1. `MUST`: Start with the simplest possible pattern (single agent or linear flow) before introducing dynamic handoffs.
2. `MUST`: If project language is C#, do not plan Magentic orchestration yet (not supported in current docs).
3. `MUST`: For handoff in workflows, model context transfer explicitly; do not assume shared session/tool state.
4. `MUST`: If middleware is used, provide both non-streaming and streaming handlers when streaming UX is required.
5. `MUST`: Add contract tests for orchestration transitions and handoff/tool invocation behavior.
6. `MUST`: Pin MAF package versions and review release notes before upgrades. Core packages are GA; sub-packages (A2A, AG-UI, orchestrations) remain preview.
7. `MUST`: Explicitly assign state ownership (app-managed vs hosted-thread-managed) before implementation.
8. `MUST`: Validate target SDK surface before implementation; do not introduce builders/types that are not available in the installed package version.
9. `MUST`: Run compile checks after each orchestration edit batch and resolve errors before adding more changes.
10. `MUST`: Reuse existing domain contracts (session/message/result DTOs) and set all required members explicitly.
11. `MUST`: Do not expose orchestration modes in enums, APIs, or config unless each mode has distinct runtime semantics and is exercised on the production path.
12. `MUST`: AG-UI streams must be backed by real workflow/run events or durable orchestration state transitions, not reconstructed snapshots that only approximate execution.
13. `MUST`: If A2A is declared as a supported protocol, implement a real ingress and egress path with validation, persistence, and retrieval; schema-only support is not sufficient.
14. `MUST`: Use a single canonical A2A JSON schema as the source of truth. If runtime validation needs an embedded schema, generate it from the canonical file or add a parity test that fails on drift.
15. `MUST`: Align serializer naming policy and schema property names (for example snake_case). Do not mix naming conventions across DTOs, validators, and API examples.
16. `MUST`: Add API-level A2A contract tests that prove ingress (POST accepted) and egress (GET retrieval/filtering) against persisted runtime messages.
17. `MUST`: In multi-service local development, prefer service discovery over hardcoded endpoints. If Aspire is used, services must resolve dependencies by service name.
18. `MUST`: Apply least-privilege MCP tool scoping per agent role. Do not register global tool sets for all agents when role-specific subsets are possible.
19. `MUST`: Propagate session IDs and correlation IDs end-to-end (UI -> API -> orchestrator -> tool boundary) and include them in logs/events.
20. `MUST`: For solution templates and starter projects, provide one-command local startup and one-command cloud deployment paths.
21. `MUST`: Define one canonical status vocabulary and transition matrix for orchestration outcomes, and reuse it across worker runtime, API contracts, and UI state handling.
22. `MUST`: When a capability is optional or not wired in an environment, return explicit machine-readable mode (`preview`, `unconfigured`, or `unsupported`) rather than implicit behavior.
23. `MUST`: Do not hide missing runtime capabilities behind silent empty responses or compatibility fallbacks that mask contract gaps.
24. `MUST`: For hosted MCP tools, prefer native Agent Framework/Foundry tool resources and run options over bespoke wrapper layers.
25. `MUST NOT`: Add custom wrapper clients that duplicate MCP discovery/invocation semantics when native SDK support is available.
26. `MUST`: Use MCP `tool_filter` to restrict each agent's tool access to the minimum required set rather than registering all available tools globally. Example: `MCPToolset(tool_filter=['read_file', 'list_directory'])`.
27. `MUST`: For tools requiring human approval, set `approval_mode` to `always_require` at tool definition and implement an approval handler that blocks execution until explicit authorization. Do not implement custom approval wrappers when the framework provides native `approval_mode` support.
28. `MUST`: For durable agent hosting (Azure Functions), use `ConfigureDurableAgents` to register agents as durable entities with auto-generated HTTP endpoints. Prefer this over custom HTTP scaffolding for stateful multi-agent services.
29. `MUST`: For long-running orchestrations, implement workflow checkpointing (for example `FileCheckpointStorage` in Python, durable entity state in C#) and test resume-from-checkpoint behavior.
30. `MUST`: When hosted agent runs support cancellation (`CancelRunAsync`), wire cancellation tokens through the orchestration and expose cancel semantics in API contracts.
31. `SHOULD`: For configuration-driven agent behavior, evaluate declarative workflow YAML definitions before building custom code orchestrations. Use declarative workflows for repetitive action sequences that benefit from low-code modification.
32. `MUST`: When bootstrapping hosted agents, implement idempotent get-or-create provisioning; do not rely on exception handling for duplicate detection (learned from NimbusIQ: exception-driven idempotency is fragile across SDK versions).
33. `SHOULD`: Implement in-process session/thread caches with bounded cardinality and TTL eviction for long-running services that reuse hosted agent threads (avoids memory growth observed in NimbusIQ production-style workloads).
34. `MUST`: When using Agent Harness (shell/filesystem tools), run harness-capable agents in isolated environments (containers, sandboxes) and require explicit approval gates before command execution. Do not grant broad host access.
35. `SHOULD`: Use DevUI (preview, browser-based local debugger) during development to visualize agent execution, message flows, tool calls, and orchestration decisions before deploying to production.

## Currency and verification gates

- Last reviewed: **2026-04-03**
- Latest upstream releases at review time: .NET `dotnet-1.0.0` (2026-04-02), Python `python-1.0.0` (2026-04-02)
- Core packages are GA. Sub-packages (A2A, AG-UI, orchestrations) remain preview.
- Handoff orchestrations are marked `[Experimental]` in .NET 1.0.0.
- `ConfigureDurableAgents` (.NET) — verify availability in your installed package version.
- Use **`foundry-mcp`** to verify Azure AI Foundry project config, model deployments, and endpoint availability.
- Before implementing, verify:
  - Provider package compatibility across `.NET` and `Python`
  - SDK type and parameter availability for your installed package version
- If version-dependent behavior is uncertain, include a `[VERIFY]` block in the PR with repro steps.

## OWASP agentic baseline (future-project default)

Map orchestration decisions to OWASP Agentic Top 10 (ASI) controls:

1. `MUST` (ASI01/ASI02): Treat all external prompts, tool outputs, and retrieved documents as untrusted. Validate and constrain before they influence planning or tool selection.
2. `MUST` (ASI03): Use capability-scoped tool catalogs per agent role. High-impact tools (delete, write, deploy, permission change) require explicit approval policies.
3. `MUST` (ASI04): Keep service identity ownership explicit per runtime path and enforce least privilege at the exact execution principal.
4. `MUST` (ASI05): Run agents and MCP connectors in sandboxed boundaries with strict network and filesystem limits; do not grant broad host access.
5. `MUST` (ASI06): Audit trails must redact secrets/tokens/credentials and cap payload size to avoid sensitive data disclosure in logs.
6. `MUST` (ASI07): Grounding and policy references must include provenance metadata (source, timestamp, scope) so poisoned sources can be detected and excluded.
7. `MUST` (ASI08): Bound orchestration loops and tool-call budgets (max turns, max tool calls, max payload size, timeout budget).
8. `MUST` (ASI09): Maintain a dependency inventory for agent/tool/model components (AIBOM/SBOM) and verify provenance before rollout.
9. `MUST` (ASI10): Require human-in-the-loop approval for irreversible or high-risk actions and preserve explainable decision traces.

### Mandatory gap-closure gates (future projects)

These gates are required before sign-off on new MAF projects or major orchestration rewrites:

1. `MUST`: High-impact tool approval gate exists and is tested.
   - Coverage includes at least `delete`, `write/update`, `deploy`, and `permission/role-change` actions.
   - Tests prove requests are blocked without approval and allowed only with valid approval state.
2. `MUST`: Orchestration/tool budgets are explicitly configured and enforced.
   - Minimum controls: max turns, max tool calls, max payload size, and operation timeout budget.
   - Tests prove budget exceedance fails closed with a machine-readable reason.
3. `MUST`: AIBOM/SBOM generation and verification is wired into CI for agent/tool/model dependencies.
   - Build/release gates fail when inventory is missing, integrity/provenance checks fail, or disallowed components are detected.
4. `MUST`: Grounding poisoning-resilience controls are implemented.
   - Source allowlist/trust policy, freshness policy, and provenance metadata checks are required.
   - Runtime behavior must downgrade confidence or reject decisions when evidence is stale or untrusted.

## Identity Ownership Rules (Cross-Service)

These rules prevent protocol/runtime correctness from being undermined by identity drift.

1. `MUST`: Define identity ownership per service at design time:
   - `control-plane-api`: chosen runtime identity
   - `agent-orchestrator`: chosen runtime identity
   - Any worker/sidecar: chosen runtime identity
2. `MUST`: If UAMI is the standard, enforce explicit client-id binding in code and config.
3. `MUST`: Do not assume attached identities imply active runtime selection.
4. `MUST`: Role assignment automation must target the identity that actually acquires tokens at runtime.
5. `MUST`: Keep identity policy symmetric across services unless a documented exception exists.

### azd Up Guardrail

When `azd up` is the deployment path, orchestrator skills must require a post-provision identity gate:

1. Resolve runtime principal IDs for each service revision.
2. Validate required role assignments on AI account/project scopes.
3. Fail deployment validation if runtime principal and RBAC target principal differ.
4. Emit actionable diagnostics (principalId, missing role, scope).

## Prompt and Invocation Baseline (Cross-Language Default)

For Microsoft Agent Framework (MAF) projects, prefer this default pattern:

- Store prompts in version-controlled files (for example `prompts/*.md`).
- Load prompts through a dedicated provider with placeholder binding (for example `IPromptProvider` in C#, equivalent abstraction in other languages).
- Keep runtime invocation native to MAF APIs (for example `AIAgent.RunAsync`, `RunStreamingAsync`, `WorkflowBuilder`, `InProcessExecution`, or language-equivalent primitives).
- Use wrappers only for cross-cutting concerns (telemetry, retries, fallback, validation), not to replace MAF orchestration primitives.

### Prompt Classification (Distribution Security)

Classify every prompt before implementation:

| Classification   | Definition                                                                             | Storage                                   | Ships in artifact? |
| ---------------- | -------------------------------------------------------------------------------------- | ----------------------------------------- | ------------------ |
| **Public**       | Formatting templates, help text, response shaping with no competitive advantage        | Version-controlled files, string literals | Yes (acceptable)   |
| **Confidential** | Competitive IP, security instructions, orchestration strategy, tool selection criteria | External config store                     | **No**             |

**Rules:**

- **Default to Confidential.** Unless explicitly classified as Public, treat all prompts as Confidential.
- **Confidential prompts must not exist in published artifacts.** This means: not in Docker image layers, not as string literals in compiled assemblies, not in npm/NuGet packages, not as `<EmbeddedResource>` items.
- **Load Confidential prompts at runtime** from Azure App Configuration (preferred), Key Vault, or Foundry agent store via `IPromptProvider` or equivalent.
- **Foundry hosted agents**: `CreateAgentAsync(instructions: ...)` stores prompt server-side after creation, but the prompt string still exists in the client binary if hardcoded. Load instruction text from App Configuration at bootstrap.
- **MCP tool catalogs**: Server URLs, allowed tool names, and topology defined in source code are visible in compiled artifacts. For sensitive architectures, load tool configuration from App Configuration or environment variables injected at deploy time.
- See `distribution-security.instructions.md` for the full framework.

## Implementation Safety Guardrails (Future Project Default)

Use this checklist whenever introducing or modernizing orchestration code.

1. Contract-first start:
   - Read existing orchestration contracts before writing code (session/message/result models).
   - Confirm required members and property names in current code, not assumed docs.
2. SDK capability check:
   - Confirm package version and supported APIs in-repo before selecting pattern primitives.
   - If docs and installed package differ, implement to installed package and log upgrade path separately.
3. Incremental compile gates:
   - Add one orchestration component at a time (state model, storage, orchestrator, integration).
   - Run `dotnet build` after each component.
   - Do not continue when build is red.
4. State and resumability safety:
   - Add persistence abstraction before concrete storage implementation.
   - Add migration scripts with indexes and idempotent DDL.
   - Add resume-path tests (skip completed work, continue pending work).
5. Concurrency safety:
   - Use explicit fan-out/fan-in boundaries.
   - Ensure failures are persisted with phase and error context.
   - Ensure in-progress state is cleaned up on success/failure.
6. Contract drift prevention:
   - Keep one canonical schema file for A2A contracts and reference it from validators/tests.
   - Add parity tests for DTO <-> schema required fields and enum/message-type values.
   - Fail CI when a validator-required property is not present in the canonical schema.
7. Runtime outbox coverage:
   - Emit A2A outbox messages from orchestration runtime transitions, not from test-only seed data.
   - Cover mediator fan-out/fan-in runs with assertions for expected message sequence and recipients.

## Failure Patterns To Prevent

- Invented model members in object initializers (for example unknown message/session properties).
- Required members not set on DTOs used by orchestration persistence/messaging.
- Large multi-file scaffolds without intermediate build validation.
- Assuming new orchestration primitives are available without verifying installed package surface.
- Missing migration path for new orchestration state tables.
- Registering an orchestrator in DI without putting it on the actual production execution path.
- Leaving `Consensus`, `Voting`, or similar modes as aliases for sequential ordering instead of implementing real semantics or removing them.
- Emitting AG-UI tool-call events from controller-side database polling when the workflow engine can emit authoritative start/result/error transitions.
- Declaring A2A support with only contracts/validators and no runtime transport or durable inbox/outbox path.
- Maintaining multiple unsynchronized A2A schemas (for example inline validator schema + canonical schema) with different required property names.
- Posting valid A2A payloads that fail validation because serializer naming and schema naming policies are out of sync.
- Verifying A2A only through unit tests while missing API contract tests for status codes, filtering, and retrieval semantics.
- Allowing backend/orchestrator/UI status drift (for example action allowed in UI but rejected by API because state names differ).
- Returning empty/default success payloads when a protocol surface is not implemented instead of advertising explicit preview/unconfigured mode.
- Building protocol adapter wrappers that bypass native hosted MCP tool resource configuration and approval semantics.

## NimbusIQ Project Lessons (Apply to Future MAF Projects)

These patterns were learned during production-style MAF integration and are grounded in real failure modes:

1. **Idempotent agent provisioning**: `CreateAIAgentAsync` does not have first-class upsert semantics. Implement get-then-create with explicit duplicate handling rather than catch-all exception flow. Agent IDs should be cached post-creation.
2. **Session cache hygiene**: In-process thread/session caches in long-running services grow unbounded without TTL and eviction. Add max-cardinality limits and emit telemetry for cache hit/miss.
3. **Hybrid hosted + local state**: When combining Foundry hosted agents (for reasoning) with local orchestrator agents (for control flow), designate one explicit state owner per user journey. Test replay/recovery across the boundary.
4. **Hosted agent config drift**: Agent IDs, model deployments, and endpoint configuration drift between environments. Store in environment-backed config and validate at startup with fail-fast.
5. **Error taxonomy gaps**: Distinguish between unconfigured-capability, not-found, already-exists, permission-denied, and quota-exceeded in agent provisioning paths. Do not use broad exception handling for all.
6. **Cancellation discipline**: Wire `CancellationToken` through all agent/thread operations. Cancelled thread-creation must not poison the session cache.
7. **OpenTelemetry from day one**: Without distributed tracing, production MAF debugging is impractical. Wire `configure_otel` / `ActivitySource` before first deployment, not after first incident.

## Pattern Selection

Reference: [pattern-selection.md](./references/pattern-selection.md)

Quick guidance:

- Sequential: deterministic pipelines with strict dependency order.
- Concurrent: independent analyses where fan-out/fan-in lowers wall-clock time.
- Handoff: one specialist delegates to another dynamically by context.
- Group Chat: collaborative iterative reasoning among multiple agents.
- Magentic: autonomous dynamic conversation loops (currently avoid for C#).
- **Durable Agents (Azure Functions)**: stateful multi-agent hosting with auto HTTP endpoints, Durable Task orchestration, and built-in checkpoint/resume. Use when agents need persistent state across invocations without custom infrastructure.

## Protocol Choice (Boundary Design)

- AG-UI: frontend streaming protocol and UX event model.
- MCP: tool access boundary (bring external systems/tools into agent runtime).
- A2A: inter-agent interoperability across apps/frameworks with task lifecycle.
- **Declarative Workflows**: YAML-based action definitions for low-code orchestration; suitable for repetitive flows that change frequently without code deploys.

Use more than one protocol when needed, but keep each protocol's responsibility clear.

## Reusable Blueprint Patterns (Interview Coach Alignment)

Adopt these as defaults for future MAF projects unless there is a documented reason not to:

- Aspire-first topology for local multi-service runs, including service discovery and health visibility.
- Handoff-capable orchestration with explicit context transfer contracts between specialists.
- MCP tools hosted as independently deployable services and scoped per agent responsibility.
- Session and correlation propagation through every protocol boundary.
- Deployment ergonomics with one-command local run and one-command cloud deploy.

## Common Traps (Do Not Repeat)

Reference: [common-traps.md](./references/common-traps.md)

Top traps:

- Handoff assumptions about shared session/tool visibility
- Middleware streaming degradation due to missing streaming handler
- AG-UI approval payload contamination (`request_approval` / `approval_response`) causing tool-call errors
- Unbounded MCP server trust and data exfiltration risk
- Preview package drift without compatibility tests

## Definition of Done

- Pattern choice is documented and justified.
- Orchestration tests cover happy path plus failure/handoff paths.
- Protocol boundary (AG-UI/MCP/A2A) is explicit in architecture and code comments.
- Known trap checklist is reviewed and signed off.
- The selected orchestration mode is the one actually used by the background worker, API, or other production entrypoint.
- Any advertised protocol surface (AG-UI or A2A) is wired to live runtime behavior, not just DTOs or simulated events.
- A2A contract tests verify accepted ingress payloads and durable outbox retrieval behavior for real orchestration emissions.
- Status transitions are validated by tests against the canonical transition matrix and are consistent across worker, API, and UI.
- Optional capabilities expose explicit mode metadata (`preview`/`unconfigured`) that callers can render and act on.

## Sources

Reference: [sources.md](./references/sources.md)

## Implemented Patterns in This Codebase

### ConcurrentMediatorOrchestrator (Fan-Out/Fan-In)

Located in `apps/agent-orchestrator/`. Uses a mediator pattern to fan out analysis tasks (WAF pillars, FinOps, Sustainability, Reliability) concurrently and aggregate results. Each specialist agent runs independently with explicit context transfer — no shared session state.

### GovernanceMediatorAgent (Negotiation + Dual-Control Approval)

Located in `apps/agent-orchestrator/`. Implements interactive governance negotiation where conflicting constraints (cost vs reliability vs sustainability) are resolved through weighted compromise. Uses dual-control approval flow requiring two independent approvals.

### Learn MCP Grounding

Agent-orchestrator integrates with Learn MCP server to ground recommendations in official Microsoft documentation. Tool results include source URLs for auditability.

### Explicit Context Transfer Between Agents

When one agent hands off to another (e.g., AnalysisOrchestrator → GovernanceMediator), context is passed via typed DTOs containing service group state, scores, and recommendations — not shared session/tool state. This follows Hard Rule #3.

> [!IMPORTANT]
> Multi-agent tool calls must implement circuit breaker patterns. If a tool fails 3 consecutive times within a 60-second window, skip that tool and log the circuit-open event. Resume after a cooldown period (default: 120s). Without this, a single failing tool blocks the entire agent pipeline.

### Production Concurrent-Mediator Default

The agent-orchestrator production path now defaults to `ConcurrentMediator` rather than leaving the mediator flow as an optional branch that is never selected. Future projects should make the intended orchestration mode the default runtime path, not an unused capability.

### AG-UI Backed by Persisted Run Events

The control-plane API should stream AG-UI events from persisted `agent_messages` emitted by the orchestration runtime. Future projects should avoid controller-side synthetic tool-call streams when the worker or orchestration engine can publish authoritative `agent.started`, `result`, and `error` events.

### A2A Requires Runtime Inbox/Outbox

If A2A is a product requirement, implement validated ingress plus durable retrieval over the system-of-record message store. Do not stop at `A2AMessage` DTOs and schema validation alone.

---

## Known Pitfalls

| Area                          | Pitfall                                                                                                 | Mitigation                                                                                                                                                                                                                                                                                                                                                                          |
| ----------------------------- | ------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Multi-agent race conditions   | Concurrent agents modifying shared state (conversation history, tool results) without coordination      | Use optimistic concurrency or explicit locking on shared state stores; design agent interactions as message-passing, not shared-memory                                                                                                                                                                                                                                              |
| Human approval implementation | Human-in-the-loop approval mentioned but no concrete gate implementation                                | Implement approval as a durable workflow step with timeout, escalation, and default-deny; never auto-approve destructive or irreversible actions                                                                                                                                                                                                                                    |
| Agent escalation on blocked   | No defined behavior when an agent cannot proceed (tool unavailable, permission denied, ambiguous input) | Define explicit escalation paths: retry with backoff → fallback agent → human escalation; log blocked state with correlation ID for diagnostics                                                                                                                                                                                                                                     |
| Token budget per agent        | No per-agent token or cost limits; a runaway reasoning loop can exhaust budget                          | Use the framework's compaction strategies (`TruncationCompactionStrategy`, `SlidingWindowCompactionStrategy`, `SummarizationCompactionStrategy`, or a `PipelineCompactionStrategy`) wrapped in `CompactionProvider` and registered as `AIContextProvider` to enforce per-agent context ceilings; monitor cumulative token usage across multi-agent sessions via OpenTelemetry spans |
| SDK sub-package instability   | MAF sub-packages (A2A, AG-UI, Anthropic, orchestrations) remain beta; APIs may change between releases  | Pin sub-package versions in `pyproject.toml` / `.csproj`; verify parameter names against your installed package version; test against minimum supported runtime version in CI                                                                                                                                                                                                       |
| A2A protocol maturity         | A2A reached v1.0.0 but the ecosystem is young; breaking changes may still occur in SDKs and tooling     | Isolate A2A integration behind an adapter interface; version-pin the A2A schema; validate inbound messages; track releases at [a2a-protocol.org](https://a2a-protocol.org)                                                                                                                                                                                                          |

---

## Currency

- **Date checked:** 2026-04-03
- **Sources:** Microsoft Learn MCP (`microsoft_docs_search`), [AG-UI Protocol](https://docs.ag-ui.com)
- **Authoritative references:** [Microsoft Agent Framework](https://learn.microsoft.com/agent-framework/), [A2A Protocol v1.0.0](https://a2a-protocol.org)

### Verification Steps

1. Confirm multi-agent orchestration patterns and ConcurrentMediator API stability
2. Verify AG-UI event specification for agent-to-UI streaming
3. Check A2A protocol maturity and any new interop standards
