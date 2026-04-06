---
name: maf-ai-integration
description: >-
  Enforce production-grade Microsoft Agent Framework and Microsoft Foundry integration patterns for .NET + React systems. USE FOR: implementing or reviewing chat/orchestration/recommendation flows that must use real AI APIs, grounded prompts, confidenceSource tracking, deterministic fallbacks, streaming (SSE), secure credential/config binding, strict validation, and UI auditability.
---

# Microsoft Agent Framework AI Integration

Use this skill when building or reviewing AI-backed features in this repo, especially in:

- `apps/agent-orchestrator/`
- `apps/control-plane-api/`
- `apps/frontend/`

This skill exists to prevent "template-first" implementations and enforce observable, enterprise-ready AI behavior.

For architecture and orchestration pattern selection (Sequential vs Concurrent vs Handoff vs Group Chat vs Magentic, plus AG-UI/MCP/A2A choices), also use:

- `.github/skills/maf-agentic-patterns/SKILL.md`

## Quick Rules

1. `MUST`: Use real AI clients (`AIAgent`, `AIProjectClient`, `AzureOpenAIChatClient`, `AzureAIAgentClient`) when AI is configured.
2. `MUST`: Use dependency injection for all AI services and consume through interfaces.
3. `MUST`: Ground prompts with concrete domain context and explicit output schema.
4. `MUST`: Tag all outputs with `confidenceSource` and propagate to API + UI.
5. `MUST`: Wrap AI calls in `try/catch`, log failures, and mark fallback outputs explicitly.
6. `MUST`: Keep chat endpoints streaming-compatible (SSE/chunking) where applicable.
7. `MUST`: Use managed identity or `DefaultAzureCredential`; no hardcoded secrets.
8. `MUST`: Validate with strict build/type checks before concluding work.
9. `SHOULD`: Show AI provenance in UI badges/tooltips and keep auditable traces.
10. `SHOULD`: Pin sub-package versions (A2A, AG-UI, Anthropic, orchestrations remain preview) and guard against preview breaking changes with compatibility tests.
11. `MUST`: Choose and document hosting mode per feature: code-managed agents vs service-hosted Foundry agents.
12. `MUST`: Verify existing API contracts before adding new AI/orchestration DTO usage; do not assume property names.
13. `MUST`: Run compile and test gates incrementally during AI/orchestration changes (not only at the end).
14. `MUST`: Hosted Foundry agent creation, retrieval, and invocation methods must be backed by the installed SDK surface or return an explicit `unconfigured` outcome; do not leave placeholder methods behind stable public APIs.
15. `MUST`: Keep package-family alignment between `Azure.AI.Projects`, Foundry/OpenAI helpers, and MAF AzureAI integration packages; do not mix preview package generations without verifying compile-time compatibility.
16. `MUST`: For A2A-enabled features, keep one canonical schema file and enforce parity between DTO serialization names, validator-required fields, and API contract examples.
17. `MUST`: Add API contract tests for A2A ingress/egress status codes and retrieval filtering. Unit tests alone are not sufficient for protocol claims.
18. `MUST`: Propagate session ID and correlation ID across UI, API, orchestrator, and tool boundaries; include both in structured logs and streamed events.
19. `SHOULD`: Use **`foundry-mcp`** to verify model deployments, endpoint availability, and project configuration before integration work.
20. `MUST`: For provider portability, keep model-provider coupling behind interfaces/options and validate at least one alternate provider path for non-production or fallback mode.
21. `SHOULD`: For local multi-service development, prefer service discovery/runtime composition via Aspire when available instead of hardcoded service URLs.
22. `SHOULD`: For sample and starter implementations, keep operational ergonomics explicit: one command for local startup and one command for cloud deployment.
23. `MUST`: AI/recommendation status values must map to one canonical lifecycle contract across backend services, API DTOs, and frontend state handling.
24. `MUST`: If an AI/protocol capability is unavailable, return explicit mode metadata (`preview`, `unconfigured`, or `unsupported`) and avoid silent empty fallback responses.
25. `MUST`: Recommendation scoring semantics must remain explicit and stable: confidence (certainty), trust (certainty + evidence + freshness), and queue/risk-weighted priority (triage urgency).
26. `MUST`: For Foundry-hosted MCP tools in C#, use the native Agent Framework/Foundry SDK integration (`MCPToolDefinition`, `MCPToolResource`, `ThreadAndRunOptions`, `PersistentAgentsClient`) rather than custom HTTP wrappers.
27. `MUST NOT`: Introduce custom MCP wrapper clients that reimplement tool discovery/invocation over raw HTTP when native MCP tool support exists in the selected SDK surface.
28. `MUST`: Use `ChatResponseFormat.ForJsonSchema<T>()` (C#) or equivalent structured output for type-safe agent responses instead of relying solely on prompt-instructed JSON formatting. Deserialize with `JsonSerializer.Deserialize<T>(response.Text, JsonSerializerOptions.Web)` after invocation.
29. `MUST`: Wire MAF-native OpenTelemetry observability (`configure_otel` in Python; `Activity`/`ActivitySource` in C# with MAF spans) so agent invocations, LLM calls, and tool executions emit distributed traces automatically.
30. `MUST`: For hosted agent runs, implement cancellation support (`CancelRunAsync` / equivalent) and surface cancellation status through API contracts; do not leave orphaned runs.
31. `MUST`: When bootstrapping hosted agents at deploy time, implement idempotent provisioning (get-or-create pattern) instead of exception-driven duplicate detection. Cache agent IDs after successful creation.
32. `SHOULD`: For long-running agent-hosted services, implement in-process session/thread cache with TTL, max-cardinality eviction, and telemetry for cache hit/miss ratios.
33. `SHOULD`: Use DevUI (preview, browser-based local debugger) during development to visualize agent execution traces, message flows, and tool calls before deploying. Reference: [DevUI docs](https://learn.microsoft.com/en-us/agent-framework/devui/).
34. `SHOULD`: For multi-provider setups, verify service connector compatibility per provider (Foundry, Azure OpenAI, OpenAI, Anthropic, Bedrock, Gemini, Ollama). Preview provider packages (`Microsoft.Agents.AI.Anthropic`, `Microsoft.Agents.AI.Bedrock`, `Microsoft.Agents.AI.Gemini`) may have different API surfaces than GA packages.

## Currency and verification gates

- Last reviewed: **2026-04-03**
- Latest upstream releases at review time: .NET `dotnet-1.0.0` (2026-04-02), Python `python-1.0.0` (2026-04-02)
- Core packages are GA. Sub-packages (A2A, AG-UI, Anthropic, orchestrations) remain preview.
- Preview provider packages: `Microsoft.Agents.AI.Anthropic`, `Microsoft.Agents.AI.Bedrock`, `Microsoft.Agents.AI.Gemini`.
- DevUI, Agent Harness, AG-UI/CopilotKit/ChatKit, GitHub Copilot SDK, and Claude Code SDK are preview features.
- `Azure.AI.Projects` upgraded to 2.0.0 GA in .NET 1.0.0 — verify `AIProjectClient` imports if upgrading from beta.
- Before implementing, verify:
  - Provider package compatibility across `.NET` and `Python`
  - SDK type availability for your installed package version
- If version-dependent behavior is uncertain, include a `[VERIFY]` block in the PR with repro steps.

## OWASP agentic implementation checklist (future-project default)

Use this alongside Quick Rules for AI-backed features:

1. `MUST` (ASI01/ASI02): Defend against behavior hijack and prompt injection with strict input handling, schema validation, and instruction/data separation.
2. `MUST` (ASI03): Enforce tool least privilege. Any mutating or high-impact tool path requires explicit approval and audit.
3. `MUST` (ASI04): Keep runtime identity explicit (especially UAMI) and verify RBAC at startup/health gates.
4. `MUST` (ASI05): Execute AI/tool workloads in constrained runtime boundaries; deny-by-default external egress where feasible.
5. `MUST` (ASI06): Redact secrets/credentials/PII from AI traces, MCP arguments/results, and audit logs. Apply payload-size limits.
6. `MUST` (ASI07): Track grounding source provenance and freshness (`url/source`, `lastUpdated`, `retrievedAt`) and downgrade trust when stale/unknown.
7. `MUST` (ASI08): Apply bounded retries, circuit breakers, rate limits, and operation budgets to prevent runaway cost or denial-of-service.
8. `MUST` (ASI09): Maintain model/tool/dependency provenance (AIBOM/SBOM), pin versions, and verify integrity before deployment.
9. `MUST` (ASI10): Preserve explainability and human oversight for critical decisions; never hide high-risk autonomous actions behind silent fallbacks.

### Mandatory gap-closure gates (future projects)

For new MAF-based solutions, these controls are required and must be evidenced in tests/CI:

1. `MUST`: Explicit approval policy for high-impact tool invocations.

- Required action classes: destructive mutations, infrastructure deployment changes, and RBAC/permission updates.
- API/runtime tests must prove deny-by-default behavior when approvals are absent.

2. `MUST`: Per-run execution budgets are configured and enforced.

- Required budgets: max model turns, max tool invocations, max tool payload size, and total runtime timeout.
- Runtime must emit machine-readable budget-exceeded outcomes.

3. `MUST`: AIBOM/SBOM verification is part of CI/CD.

- Pipelines must generate inventory artifacts and fail when provenance/integrity policy checks fail.

4. `MUST`: Grounding-source trust and freshness gates are enforced.

- Source provenance, retrieval time, and freshness thresholds must be evaluated before high-trust decisions.
- Confidence/trust must downgrade (or decision must fail closed) when evidence quality is insufficient.

## Identity and RBAC Consistency (Critical)

Use this section to avoid recurring 401/403 failures caused by identity drift.

1. `MUST`: Choose one runtime identity per service for AI calls and document it explicitly:

- Preferred in Container Apps: **User-Assigned Managed Identity (UAMI)**.
- Acceptable fallback: **System-Assigned Managed Identity (SAMI)** only if intentional and documented.

2. `MUST`: Ensure credential construction matches the chosen identity:

- If UAMI is chosen, set `ManagedIdentityClientId` (or equivalent explicit client-id binding).
- Do not rely on ambient `DefaultAzureCredential` identity selection when both UAMI and SAMI exist.

3. `MUST`: Assign required AI roles to the **exact principal used at runtime** (not just a different attached identity).
4. `MUST`: Keep identity source consistent across app code, app settings, and IaC outputs.
5. `MUST`: During `azd up`, validate role assignments for all execution paths (including `skipContainerApps`/bootstrap paths).

### Required Role Check (AI/Foundry)

Before closing Azure deployment work, verify the runtime principal has the minimum required roles at the correct scope.

- `Cognitive Services OpenAI User` (or service-equivalent inference role)
- `Cognitive Services User` (when required by SDK call path)
- Project/workspace-scoped Foundry role if project APIs are used

### Post-Deploy Identity Verification Checklist

1. Confirm which principal is used at runtime (UAMI client-id or SAMI object-id).
2. Verify role assignments for that principal at AI account/project scopes.
3. Verify AI endpoint audience/scope is correct for the SDK path in use.
4. Confirm logs show successful chat/inference calls and non-zero operation counts.
5. Confirm fallback mode is not silently masking auth errors.

## Build-Safe Delivery Guardrails (Default)

Apply this sequence for all non-trivial AI/orchestration changes.

1. Contract validation first:

- Inspect current request/response/session/message contracts in the target project.
- Record required members and expected enum/property names before implementation.

2. Minimal vertical slice:

- Implement one end-to-end path first (agent call -> contract mapping -> persistence -> API surface).
- Compile and test before scaling to additional agents/features.

3. Red/green batching:

- After each edit batch, run build gates for affected projects.
- Fix all compile errors before introducing new files or patterns.

4. Migration discipline:

- Any new durable orchestration state requires migration + rollback-safe DDL + index strategy.
- Add at least one test covering recovery/resume behavior.

5. Drift prevention:

- If guidance uses preview SDKs, pin exact versions and verify symbols against the installed package.
- Avoid introducing sample-only APIs unless verified in the current solution.

6. Contract parity gates:

- Validate that serialized DTO property names (for example snake_case) match schema required properties.
- Add at least one test that exercises object-path validation and one that exercises raw JSON-path validation.
- Fail fast on naming-policy drift (camelCase vs snake_case) before integration tests run.

## Common Preventable Errors

- Mapping to non-existent DTO members in object initializers.
- Missing required members on orchestrator/session/message contracts.
- Delayed build validation causing cascading syntax/contract errors.
- Adding persistence components without schema migration and recovery tests.
- Leaving `CreateAgentAsync` or prompt-flow methods as pseudo-implementations that never call a real Foundry SDK operation.
- Upgrading `Azure.AI.Projects` without verifying the exact hosted-agent method signatures against the installed package docs or assembly surface.
- Advertising hosted-agent capability when the runtime path still falls back to local prompt execution for every operation.
- Validator/schema divergence where one path requires `messageId` while API/DTOs emit `message_id`.
- Declaring A2A support without proving runtime-emitted outbox messages are retrievable through API contract tests.
- Frontend silently swallowing missing AI/protocol endpoints (for example default empty payload on 404) and masking integration gaps.
- Mixing recommendation status names across services/controllers/UI, causing valid actions to fail at runtime transitions.
- Reintroducing custom MCP wrapper classes around hosted MCP tools instead of using native Agent Framework/Foundry MCP abstractions.
- Relying on prompt-instructed JSON when `ResponseFormat.ForJsonSchema<T>` is available, causing downstream deserialization failures on schema-noncompliant responses.
- Omitting `CancelRunAsync` or equivalent for hosted agent runs, leaving orphaned runs that consume quota.
- Using exception-driven control flow for hosted agent idempotent provisioning instead of get-or-create pattern.
- Missing OpenTelemetry wiring: deploying agents without distributed tracing makes production debugging impossible.
- Caching hosted agent session/thread references without TTL, causing memory growth in long-running services.

## Current Capability Snapshot (Verified 2026-04-03, grounded via Context7 + Microsoft Learn)

- Microsoft Agent Framework core packages reached 1.0.0 GA on 2026-04-02. Sub-packages remain preview.
- Framework capability areas:
  - Agents (tools, MCP integration, model providers, sessions)
  - Workflows (type-safe graph orchestration, checkpointing, HITL)
- Built-in orchestration patterns: Sequential, Concurrent, Handoff, Group Chat, Magentic.
- C# limitation to account for: Magentic orchestration is not yet supported in C#.
- AG-UI integration supports full protocol features including streaming, approvals, shared state, and predictive state updates.
- A2A integration supports agent cards, task-based long-running processes, and cross-framework interoperability.
- MCP support includes local MCP clients and hosted MCP tools for Foundry agents. Tool filtering (`tool_filter`) enables per-agent scoping.
- Foundry Agent Service supports persistent agents with service-managed chat history and server-side tool orchestration.
- **Durable Agents** (Azure Functions hosting): `ConfigureDurableAgents` extension for deterministic multi-agent orchestrations with auto-generated HTTP endpoints (`/api/agents/{agentName}/run`). Supports sequential, parallel, conditional, and human-in-the-loop patterns via Durable Task.
- **Structured Output**: `ChatResponseFormat.ForJsonSchema<T>()` enforces LLM response schema at the API level, enabling type-safe deserialization and inter-agent structured data passing without intermediate parsing.
- **Workflow Checkpointing**: `FileCheckpointStorage` (Python) and durable entity state (C#) enable resume-from-failure for long-running orchestrations.
- **Human-in-the-Loop Tool Approval**: Tools support `approval_mode` (`always_require`, `never_require`) with runtime approval handlers for sensitive operations.
- **OpenTelemetry**: `configure_otel` (Python) / `ActivitySource` (C#) provides automatic distributed tracing of agent invocations, LLM calls (with token usage), and tool execution spans.
- **Declarative Workflows**: YAML-based workflow definitions with actions (`SetValue`, `SendActivity`, etc.) for low-code orchestration; useful for configuration-driven agent behaviors.
- **LRO / Run Cancellation**: `CancelRunAsync` enables graceful cancellation of hosted agent runs.
- **Compaction (Experimental)**: Built-in token budget management via `CompactionProvider` registered as `AIContextProvider`. Five strategies: `TruncationCompactionStrategy`, `SlidingWindowCompactionStrategy`, `ToolResultCompactionStrategy`, `SummarizationCompactionStrategy`, and pipeline composition via `PipelineCompactionStrategy`. Requires `#pragma warning disable MAAI001`.
- **Agent Skills**: `FileAgentSkillsProvider` enables progressive skill disclosure — advertise skill summaries (~100 tokens each), load full instructions on demand (<5000 tokens), and read skill resources as needed. Reduces baseline context cost for agents with many skills. [VERIFY: Only Python code examples found for FileAgentSkillsProvider; C# support not source-verified.]
- **Chat History / Memory Providers**: `InMemoryChatHistoryProvider` (default), `CosmosChatHistoryProvider` (preview) for durable cross-session history, and `ChatHistoryMemoryProvider` (preview) for semantic memory retrieval from conversation history. [VERIFY: CosmosChatHistoryProvider and ChatHistoryMemoryProvider not found in official C# docs or code samples; only Python CosmosHistoryProvider and BoundedChatHistoryProvider/TruncatingChatReducer patterns are source-verified.]

## Intelligence Layer Pattern (Shared AI Services)

When multiple application components need the same AI capabilities (summarization, classification, extraction, scoring), centralize them behind a shared Intelligence Layer rather than embedding AI integration into each component independently.

### When to Centralize

| Signal                                                     | Shared Intelligence Layer | Embedded AI |
| ---------------------------------------------------------- | ------------------------- | ----------- |
| Multiple services call the same model with similar prompts | Yes                       | —           |
| Prompt/model config needs central management and audit     | Yes                       | —           |
| AI capability used by one service only                     | —                         | Yes         |
| Rapid prototyping or spike                                 | —                         | Yes         |

### Implementation

1. `MUST`: Expose AI capabilities through a shared service interface (`IAiCapabilityService` or equivalent), not direct model calls scattered across components.
2. `MUST`: Manage prompts as configuration (Azure App Configuration or Key Vault), not as string literals in each consumer. See distribution-security.instructions.md.
3. `MUST`: Track model deployment, prompt version, and configuration version per request for auditability.
4. `MUST`: Apply the same `confidenceSource` and provenance tracking from Quick Rules regardless of how consumers reach the Intelligence Layer.
5. `SHOULD`: Expose Intelligence Layer capabilities as MCP tools when agents or Copilot need to consume them.
6. `SHOULD`: Support provider portability (swap models or endpoints) behind the shared interface without changing consumers.

### Relationship to Composable Architecture

The Intelligence Layer maps to the composable architecture's shared AI service (see `composable-app-architecture` skill). When building Level 3+ composable systems, the Intelligence Layer should be a standalone service that UX, Logic, and agent consumers all call through the same contract.

For simpler systems (Level 0-1), keeping AI embedded in the application service is fine — extract to a shared layer only when a second consumer needs the same capability.

## Workflow

1. Identify whether the feature is AI-backed (chat, recommendation, enrichment, generation).
2. Decide hosting mode: local/code-managed agent runtime or Foundry hosted/persistent agent.
3. Verify AI service registration exists; if not, add DI + options binding first.
4. Implement AI-first path with grounded prompt and schema-constrained output.
5. Add deterministic fallback path and explicit confidence/provenance tagging.
6. Ensure API contracts include provenance fields and UI renders them clearly.
7. Add/adjust streaming behavior for chat-like endpoints.
8. Run strict validation commands and fix issues in touched scope.

## Required Patterns

### 1) Real AI Integration (No Hidden Heuristic Default)

- If AI configuration is present and valid, the default execution path must call real AI APIs.
- Heuristic/template logic is fallback-only for outage/error scenarios.
- AI logic should be centralized behind interfaces (for testability and replacement).

### 1.1) Foundry Hosted Agents vs Code-Managed Agents

- Use Foundry hosted agents when you need:
  - service-managed threads/history
  - built-in server-side tool orchestration/retry
  - enterprise controls (RBAC/network/content safety/observability) at runtime
- Use code-managed agents when you need:
  - tight in-process control over orchestration and state
  - custom runtime behavior not yet available in service-hosted mode
- Never mix history ownership unintentionally. One feature should have one source of truth for thread state.
- If hosted agents are optional, public methods must communicate that explicitly with a machine-readable status such as `unconfigured`; do not silently emulate hosted behavior with a local string-prompt fallback.

Reference: [foundry-hosted-agents.md](./references/foundry-hosted-agents.md)

### 2) Prompt Engineering and Structured Output

- Prompts must include:
  - domain context (service group, WAF pillar/recommendation context, constraints)
  - requested task and format
  - explicit output contract (fields, enum values, confidence expectations)
- For multi-line prompt templates in C#, use raw string literals (`$$"""`).
- **Prefer `ResponseFormat` over prompt-only JSON enforcement** when the downstream consumer needs typed deserialization:

```csharp
// Type-safe structured output (preferred for inter-agent data flow)
AIAgent agent = chatClient.AsAIAgent(new ChatClientAgentOptions()
{
    Name = "AssessmentAgent",
    ChatOptions = new()
    {
        Instructions = "Assess the architecture...",
        ResponseFormat = ChatResponseFormat.ForJsonSchema<AssessmentResult>()
    }
});
AgentResponse response = await agent.RunAsync(contextInput);
var result = JsonSerializer.Deserialize<AssessmentResult>(response.Text, JsonSerializerOptions.Web)!;
```

- Continue using prompt-instructed JSON for exploratory/creative outputs where schema rigidity is counterproductive.

Reference: [prompt-patterns.md](./references/prompt-patterns.md)

### 3) Confidence Source Tracking

- Every AI-produced object must include:
  - `confidenceScore` (if available)
  - `confidenceSource` (for example: `ai_foundry`, `heuristic_rule`, `template_default`)
- Propagate `confidenceSource` end-to-end:
  - domain model
  - API DTO/response
  - frontend rendering

### 4) Error Handling and Deterministic Fallback

- AI calls must be wrapped in `try/catch`.
- On exception/timeout:
  - emit logs with correlation identifiers
  - return deterministic fallback output
  - set `confidenceSource` to non-AI value
  - preserve a machine-readable reason where appropriate

Reference: [error-fallback-template.md](./references/error-fallback-template.md)

### 5) Streaming and Protocol Compliance

- Chat/conversational responses should support SSE/chunked streaming.
- Flush chunks progressively and preserve event ordering.
- If UI depends on pacing/protocol semantics, keep minimal compliant delays where needed.
- Include correlation IDs in logs/events for troubleshooting.
- When orchestration state is persisted asynchronously by a worker, AG-UI endpoints should stream those persisted runtime events rather than reconstructing tool-call progress from unrelated snapshot queries.
- Include session IDs in logs/events when available so streamed traces are attributable to a single conversation lifecycle.

### 6) Security and Config

- Use `DefaultAzureCredential` or managed identity in hosted environments.
- In production, prefer explicit credentials (for example `ManagedIdentityCredential`) over generic chains when feasible.
- Bind AI endpoint/deployment values via options pattern and environment variables.
- Validate required config at startup and fail fast when mandatory settings are absent.

> [!WARNING]
> **Prompt injection testing is required** before deploying any user-facing AI integration. Establish a baseline test suite that includes: (1) direct injection attempts in user input, (2) indirect injection via tool/function outputs, (3) system prompt extraction attempts. Log all LLM inputs and outputs (redacting PII) to enable post-incident analysis.

### 7) Validation and Build Quality

Before closing work:

- Frontend: run TypeScript build (`tsc -b`) and relevant lint/test commands.
- .NET services: run build/test for affected projects.
- If containerized deployment path changed, run image build validation.

### 8) UI/UX Transparency

- Distinguish AI-generated vs rule/template outputs in UI (badge or label).
- Provide tooltips or details that explain provenance and confidence.
- Preserve audit trail fields for enterprise/compliance review.
- Surface capability mode in UI for optional features (`preview`, `unconfigured`, `unsupported`) so operators understand behavior and risk.

### 9) Canonical Recommendation Lifecycle + Scoring Semantics

- Maintain one canonical recommendation status lifecycle and transition map shared by orchestrator, API, and UI.
- Add contract tests that validate allowed transitions and reject invalid transitions consistently.
- Keep score semantics distinct and documented in contracts/UI labels:
  - `confidence`: recommendation certainty
  - `trust`: certainty + evidence completeness + freshness
  - `queue`/`riskWeightedScore`: triage urgency

Reference: [ui-confidence-audit.md](./references/ui-confidence-audit.md)

## Definition of Done

- AI-first path implemented with real MAF/Azure AI clients.
- Fallback path exists and is explicitly marked as non-AI.
- `confidenceSource` propagated across backend + frontend.
- Streaming behavior is implemented for chat endpoints where expected.
- Config/auth follows managed identity and options pattern.
- Build/type checks pass for modified scope.
- Hosted-agent operations either succeed against the real SDK or return an explicit `unconfigured`/`unsupported` status that downstream code can inspect.
- A2A-enabled features have parity checks ensuring DTO serialization, schema validation, and API examples use the same field names and message types.

## Implemented Patterns in This Codebase

### Nullable AI Service DI (Optional AI)

`AIChatService` is registered as a singleton. Downstream services receive it as a nullable constructor parameter:

```csharp
builder.Services.AddScoped<ExecutiveNarrativeService>(sp =>
    new ExecutiveNarrativeService(
        sp.GetRequiredService<AtlasDbContext>(),
        sp.GetRequiredService<ILogger<ExecutiveNarrativeService>>(),
        sp.GetService<AIChatService>())); // nullable — null when AI not configured
```

Services check `_aiChatService?.IsAIAvailable == true` before calling AI APIs. This ensures the system degrades gracefully without AI Foundry.

### ExecutiveNarrativeService (AI + Rule-Based Fallback)

Located in `apps/control-plane-api/src/Application/Services/ExecutiveNarrativeService.cs`. Generates executive narrative summaries:

- AI path: builds grounded prompt with scores, recommendations, and drift context → calls `AIChatService.GenerateResponseAsync()` → tags `confidenceSource = "ai_foundry"`
- Fallback path: deterministic rule-based summary from score/recommendation/drift data → tags `confidenceSource = "rule_engine"`
- AI call wrapped in try-catch with logged warning on failure

### ScoreSimulationService (Deterministic Projection)

Located in `apps/control-plane-api/src/Application/Services/ScoreSimulationService.cs`. Pure deterministic service (no AI dependency) that projects score impact from hypothetical changes. Generates cost deltas from existing recommendation `EstimatedImpact` JSON and risk deltas from score magnitude changes.

### Singleton AI Client Pattern

`AIChatService` is registered as singleton to avoid expensive client creation per request. Services that use it are scoped (they receive the singleton via DI). Health check (`AiServiceHealthCheck`) reports AI availability as "Degraded" (not unhealthy) since rule-based fallback exists.

### confidenceSource End-to-End

All AI-generated and rule-generated outputs carry `confidenceSource` through: domain result → API DTO → frontend rendering. The frontend uses this to display provenance badges.

### Hosted Foundry Operations Must Be Real Or Explicitly Disabled

In this codebase, prompt execution can use a code-managed `AIAgent`, but hosted agent creation and hosted-flow invocation must not remain placeholder methods. Future projects should implement these methods against the installed Foundry SDK surface or return a durable `unconfigured` result when capability-host support is absent.

### Production UX Should Stream Runtime Events

For agent orchestration UX, first-class AG-UI experiences should be driven by runtime events emitted from the orchestration engine and persisted to the system-of-record, not by controller-side approximations built from point-in-time database reads.

### Provider Portability and Runtime Selection

Future projects should keep provider selection behind interfaces and configuration (for example Foundry default with alternate provider paths for dev/test/fallback). This avoids lock-in at call sites and allows controlled migration between provider backends.

### Aspire-Friendly Local Composition

Where a solution includes multiple services, prefer Aspire composition and service discovery for local development rather than hardcoded endpoint wiring. This improves parity with production-like topology and simplifies diagnostics.

---

## Known Pitfalls

| Area                                       | Pitfall                                                                                                                                                                                                             | Mitigation                                                                                                                                                                                                                                                            |
| ------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| AI endpoint rate-limiting                  | Unbounded calls to Azure AI Foundry or OpenAI endpoints cause cost overruns and 429 throttling                                                                                                                      | Enforce per-request token budgets and per-minute call limits via middleware or `DelegatingHandler`; surface 429 `Retry-After` to callers                                                                                                                              |
| Token credential refresh                   | Bearer token extracted once at startup expires after ~1h; AI SDK calls fail silently                                                                                                                                | Supply a `DelegatingHandler` that calls `credential.GetTokenAsync` per-request instead of caching a static token string                                                                                                                                               |
| Approval workflow patterns                 | Human-in-the-loop approval for AI-driven actions mentioned but no concrete implementation                                                                                                                           | Implement approval gates using durable workflow (e.g., Durable Functions or queue-based) with timeout and fallback; never auto-approve destructive operations                                                                                                         |
| Error propagation across async boundaries  | AI service errors (content filter, token limit, model unavailable) surface as generic 500s                                                                                                                          | Map AI SDK exceptions to domain-specific error types with actionable messages; propagate `x-error-code` headers for client diagnostics                                                                                                                                |
| Prompt injection via user input            | User-supplied text concatenated directly into LLM prompts without sanitization                                                                                                                                      | Separate system prompts from user input; validate and sanitize user content before inclusion; use structured tool-call patterns instead of free-form prompt concatenation                                                                                             |
| **Prompt exposure in published artifacts** | **Prompts in raw string literals are extractable from compiled .NET assemblies via ILSpy or `dotnet-ildasm`. `prompts/*.md` files `COPY`'d into Docker images are extractable via `docker cp` or layer inspection** | **Classify prompts as Public or Confidential. Load Confidential prompts at runtime from Azure App Configuration or Key Vault via `IPromptProvider`, not from files baked into images or string literals in source code. See `distribution-security.instructions.md`** |
| **MCP tool catalog exposure**              | **MCP server URLs, allowed tool names, and topology defined in source code ship with the compiled application and are visible to anyone inspecting the binary or Docker image**                                     | **Accept for non-sensitive tool catalogs; for sensitive architectures, load tool configuration from App Configuration or environment variables injected at deploy time**                                                                                              |

---

## Currency

- **Date checked:** 2026-04-03
- **Sources:** Microsoft Learn MCP (`microsoft_docs_search`), [Azure AI Foundry documentation](https://learn.microsoft.com/azure/ai-studio/)
- **Authoritative references:** [Azure AI Foundry SDK](https://learn.microsoft.com/azure/ai-studio/), [Microsoft Agent Framework](https://learn.microsoft.com/agent-framework/)

### Verification Steps

1. Confirm Azure AI Foundry SDK package versions and any new AI integration patterns
2. Verify hosted agent creation APIs match current Foundry SDK surface
3. Check for new provider-agnostic AI client abstractions in .NET
