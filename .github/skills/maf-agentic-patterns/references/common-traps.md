# Common Traps and Guardrails

## 1) Assuming Handoff Shares All Context

Trap:

- In handoff flows, agents may not share runtime/session/tool context the way teams expect.

Guardrail:

- Treat handoff payload as an explicit contract.
- Rehydrate required context for each agent.
- Add tests for tool visibility and state continuity across handoff boundaries.

## 2) Streaming Regressions With Middleware

Trap:

- Defining only non-streaming middleware handler can cause streaming paths to degrade.

Guardrail:

- Implement both handlers where needed:
  - non-streaming (`runFunc`)
  - streaming (`runStreamingFunc`)
- Add integration tests that assert chunked streaming behavior.

## 3) AG-UI Approval Message Contamination

Trap:

- Passing approval-specific fields through tool call content can trigger schema/tool-call failures.

Guardrail:

- Strip approval-only keys (`request_approval`, `approval_response`) before final tool-call payload.
- Add regression test for approval required + tool invocation sequence.

## 4) Over-Trusting MCP Servers

Trap:

- MCP tools can expose broad external capabilities and sensitive data paths.

Guardrail:

- Use trusted MCP servers only.
- Restrict scopes and secrets.
- Document data egress expectations and cost implications for third-party tools.

## 5) Sub-Package Preview Drift

Trap:

- MAF core packages are 1.0.0 GA, but sub-packages (A2A, AG-UI, Anthropic, orchestrations) remain preview and can shift APIs between releases.

Guardrail:

- Pin sub-package versions separately from core.
- Maintain compatibility tests for orchestration and protocol contracts.
- Validate upgrade in CI before rollout.

## 6) Picking Complex Pattern Too Early

Trap:

- Jumping directly to dynamic multi-agent patterns for problems solvable by deterministic logic.

Guardrail:

- Start simple, then graduate complexity only when measurable benefit exists.
- Record the reason for each orchestration upgrade in architecture notes.

## 7) Mixing Hosted and Local State Without Ownership

Trap:

- App-level orchestration and Foundry hosted thread history both act as source of truth.

Guardrail:

- Assign one state owner per journey.
- If hybrid is required, define explicit projection/sync boundaries and test replay/recovery.

## 8) Hosted Agent Identity and Configuration Drift

Trap:

- Hosted agent IDs, model deployments, or endpoint configuration drift between environments.

Guardrail:

- Store agent IDs and endpoint settings in environment-backed config.
- Validate at startup and fail fast on missing/incompatible values.
- Add smoke tests for create-thread, send-message, and stream-response paths.

## 9) Exception-Driven Idempotent Provisioning (NimbusIQ learning)

Trap:

- Using broad try/catch around `CreateAIAgentAsync` to detect duplicates, with different SDKs/versions returning different error shapes (400 vs 409 vs different error bodies).

Guardrail:

- Implement get-or-create pattern: call `GetAIAgentAsync` first, create only if not found.
- Cache agent IDs after successful provisioning.
- Treat provisioning failure taxonomy (not-found, already-exists, permission-denied, quota-exceeded) as distinct code paths, not broad catch blocks.

## 10) Unbounded Session/Thread Cache (NimbusIQ learning)

Trap:

- Long-running services cache thread/session references without eviction, causing memory growth proportional to unique conversations.

Guardrail:

- Add max-cardinality and TTL eviction to in-process session caches.
- Emit telemetry for cache size, hit rate, and eviction count.
- Ensure cancelled thread creation does not leave poisoned entries in the cache.

## 11) Missing Observability Until Production Incident (NimbusIQ learning)

Trap:

- Deploying MAF agents without OpenTelemetry tracing, making root-cause analysis of failures impractical in distributed systems.

Guardrail:

- Wire `configure_otel` (Python) or `ActivitySource` (C#) at project setup, not after first incident.
- Verify traces include agent invocation spans, LLM call spans (with token usage), and tool execution spans.
- Connect to Azure Monitor or equivalent before first deployment.

## 12) Orphaned Hosted Agent Runs (NimbusIQ learning)

Trap:

- Not implementing `CancelRunAsync` for hosted agent runs, leaving in-progress runs consuming quota/resources when the client disconnects or times out.

Guardrail:

- Wire `CancellationToken` through all agent run operations.
- Call `CancelRunAsync` on timeout, client disconnect, or user-initiated cancellation.
- Test that cancelled runs reach terminal state and do not block subsequent operations.

## 13) Prompt-Only JSON Instead of Structured Output (NimbusIQ learning)

Trap:

- Relying exclusively on prompt instructions for JSON output format when `ResponseFormat.ForJsonSchema<T>()` is available, leading to parse failures on noncompliant responses.

Guardrail:

- Use `ChatResponseFormat.ForJsonSchema<T>()` for responses that need typed deserialization.
- Reserve prompt-instructed JSON for exploratory/creative outputs where schema rigidity is counterproductive.
- Add deserialization tests that validate the response round-trips to the expected type.

## 14) Hybrid State Ownership Ambiguity (NimbusIQ learning)

Trap:

- Both local orchestrator and Foundry hosted agent claim ownership of conversation state, causing replay/recovery divergence.

Guardrail:

- Designate one state owner per user journey at design time.
- If hybrid is required, define explicit projection/sync boundaries and test replay/recovery.
- Document the boundary in an ADR and enforce it in code review.
