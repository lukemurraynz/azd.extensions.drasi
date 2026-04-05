# Production Readiness Checklist

## Agent Design

- [ ] Confirmed agent is the right abstraction (not a plain function or workflow)
- [ ] Agent has focused, single-responsibility instructions
- [ ] Tool count per agent is ≤ 15 (split into specialists if exceeding)
- [ ] Tool descriptions are concise and accurate
- [ ] All tool parameters have `[Description]` attributes
- [ ] Maximum conversation turn limit is configured
- [ ] Agent instructions are versioned artifacts (not inline strings)
- [ ] Framework package versions are pinned and compatible with the chosen provider SDK
- [ ] Current release notes reviewed for breaking API changes before implementation
- [ ] Core packages are 1.0.0 GA; sub-packages verified for stability if used

## Authentication & Identity

- [ ] Using `ManagedIdentityCredential` (not `DefaultAzureCredential`) in production
- [ ] Azure OpenAI has `disableLocalAuth: true` (API keys disabled)
- [ ] RBAC role is `Cognitive Services OpenAI User` (minimum privilege)
- [ ] No API keys or secrets in environment variables or code
- [ ] Cross-reference: [identity-managed-identity](../../identity-managed-identity/SKILL.md)

## Responsible AI

- [ ] Azure AI Content Safety filters enabled on the deployment
- [ ] Agent instructions include scope boundaries and refusal rules
- [ ] Tested against prompt injection attacks
- [ ] Tested against jailbreak attempts
- [ ] PII is redacted from logs and telemetry
- [ ] AI-generated content is disclosed to end users
- [ ] Cross-reference: [responsible-ai.md](responsible-ai.md)

## AgentEval Testing

> AgentEval is a test-time dependency only (lives in your test project, never production code).
> All evaluations must run against real LLM endpoints — no mocks or replayed responses.

- [ ] AgentEval installed in test project only (not in any production `.csproj`)
- [ ] AgentEval functional evaluation passes with overall score ≥ 0.8 (real LLM endpoint)
- [ ] AgentEval red team scan passes with zero critical findings (real LLM endpoint)
- [ ] Behavioral policy compliance validated (NeverCallTool, MustConfirmBefore)
- [ ] Responsible AI metrics meet thresholds (toxicity, bias, misinformation)
- [ ] SARIF report exported and reviewed (upload to GitHub Security tab for CI/CD)
- [ ] Cross-reference: [agenteval skill](../../skills/agenteval/SKILL.md)

## Tool Safety

- [ ] Destructive tools have human-in-the-loop approval (`WithToolApproval`)
- [ ] Python sensitive tools use `@tool(approval_mode="always_require")` where applicable
- [ ] Approval flow handles pending `user_input_requests` until fully resolved
- [ ] Tool exceptions are caught and return structured error descriptions
- [ ] Tool outputs are schema-validated before writes, routing, or external side effects
- [ ] Sensitive actions re-check authorization after tool output (do not trust tool success alone)
- [ ] Tools are idempotent or have compensating actions
- [ ] Tool responses are validated before influencing agent decisions
- [ ] No raw exceptions are exposed to the LLM

## MCP Tools (if applicable)

- [ ] **Hosted MCP (Foundry)**: uses `MCPToolDefinition` + `PersistentAgentsClient` — no custom HTTP wrappers or `McpClientFactory` for hosted endpoints
- [ ] `AllowedTools` on `MCPToolDefinition` is an explicit allowlist (not open-ended)
- [ ] `serverLabel` is consistent between `MCPToolDefinition` and `MCPToolResource`
- [ ] `RequireApproval` is set intentionally (`"never"` ↔ fully automated; `"always"` ↔ HITL)
- [ ] MCP server URL/host is allowlisted and verified at startup
- [ ] Persistent agents are deleted (`DeleteAgentAsync`) when no longer needed
- [ ] **Local MCP (stdio)**: uses `StdioClientTransport` + `McpClientFactory.CreateAsync` with `await using` disposal
- [ ] MCP server auth headers are passed at run time (not hardcoded in config)
- [ ] MCP payload logging is metadata-only by default (no raw sensitive payloads)

## Networking & Security

- [ ] Azure OpenAI is behind Private Endpoint (no public access)
- [ ] Container App / App Service has VNet integration
- [ ] Cross-reference: [private-networking](../../private-networking/SKILL.md)

## Observability

- [ ] OpenTelemetry configured with `Microsoft.Agents.*` sources
- [ ] Agent invocation count metric is tracked
- [ ] Tool call success rate metric is tracked
- [ ] Token usage per request metric is tracked
- [ ] Model error rate (429/500) is tracked with alerts
- [ ] Conversation turn count is monitored
- [ ] P95 latency is monitored with threshold alerts
- [ ] Cross-reference: [observability-monitoring](../../observability-monitoring/SKILL.md)

## Reliability & Fallback

- [ ] Model timeout budgets are defined (interactive vs background)
- [ ] Tool timeout budgets are defined (read vs write tools)
- [ ] Retries use bounded exponential backoff and target transient failures only
- [ ] Circuit breaker policy is implemented for repeated failures
- [ ] Fallback deployment/model strategy exists for repeated `429`/`5xx`
- [ ] Degraded-mode response contract is defined for end users and callers

## Compaction (if in-memory history)

- [ ] Compaction strategy selected (truncation, sliding window, tool result collapse, summarization, or pipeline)
- [ ] `CompactionProvider` registered via `UseAIContextProviders` on `ChatClientBuilder` (not on `ChatClientAgentOptions`) to avoid polluting persisted history
- [ ] `MinimumPreserved` floor set appropriately per strategy
- [ ] `PipelineCompactionStrategy` orders strategies from gentlest to most aggressive
- [ ] Summarization uses a separate, cheaper model
- [ ] `#pragma warning disable MAAI001` acknowledged (compaction is experimental)

## Agent Skills (if applicable)

- [ ] Skills loaded via `FileAgentSkillsProvider` with explicit directory paths
- [ ] `SKILL.md` files are < 500 lines; detailed reference material in `references/` subdirectory
- [ ] Skill scripts reviewed for security before deployment (treat as third-party code)
- [ ] Skill names match directory names (lowercase, hyphens, no consecutive hyphens)

## Infrastructure

- [ ] Azure OpenAI Bicep uses `disableLocalAuth: true`
- [ ] Verify disableLocalAuth: true is set on all AI service resources in Bicep/Terraform
- [ ] Model deployment specifies explicit `version` (pinned)
- [ ] Capacity (TPM) is sized for expected load
- [ ] Regional deployment matches data residency requirements
- [ ] Cross-reference: [azure-defaults](../../azure-defaults/SKILL.md)

## Workflows (if applicable)

- [ ] Workflow has defined start and end nodes
- [ ] Conditional edges cover all possible outcomes (no dead ends)
- [ ] Checkpointing is enabled for long-running workflows
- [ ] Error edges handle executor failures gracefully
- [ ] Intermediate results are logged for audit

## Multi-Agent (if applicable)

- [ ] All agent-to-agent interactions include correlation identifiers
- [ ] Agent ownership boundaries are clearly defined
- [ ] Escalation / fallback behaviour is defined for each agent
- [ ] A2A agent cards are served at `/.well-known/agent.json`
- [ ] Cross-reference: [csharp.instructions.md Agent Systems Extension](../../../instructions/csharp.instructions.md)

## Hosting

- [ ] Health check endpoint is configured (`/health/ready`, `/health/live`)
- [ ] Graceful shutdown handles in-flight agent conversations
- [ ] Auto-scaling is configured based on request rate
- [ ] Deployment uses blue/green or canary strategy
