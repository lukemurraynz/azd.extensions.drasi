# Agent Design Standards

## Agent vs Workflow Decision Framework

Before building, determine the right abstraction:

| Signal                                   | Use Agent | Use Workflow | Use Plain Code |
| ---------------------------------------- | --------- | ------------ | -------------- |
| Task requires reasoning or judgment      | ✅        |              |                |
| Task has well-defined sequential steps   |           | ✅           |                |
| Task is deterministic with no ambiguity  |           |              | ✅             |
| Multiple agents must coordinate          |           | ✅           |                |
| User intent must be interpreted          | ✅        |              |                |
| Steps require conditional branching      |           | ✅           |                |
| Task can be handled by a single function |           |              | ✅             |

> **Rule:** If you can write a function to handle the task, do that instead of using an
> AI agent. Agents add latency, cost, and non-determinism.

---

## Agent Design Principles

### Single Responsibility

Each agent should have a focused purpose. Prefer multiple specialist agents composed via
agent-as-function-tool over a single monolithic agent.

**Anti-pattern:**

```
"You are a customer service agent that handles billing, technical support,
account management, compliance questions, and product recommendations."
```

**Preferred:**

```
"You are a triage agent. Determine the user's intent and route to the
appropriate specialist."
```

With specialist agents for billing, technical, accounts, etc., each exposed via `.AsAIFunction()`.

### Instruction Design

- Keep instructions concise — every token counts against the context budget
- State what the agent should do, not what it should avoid
- Include output format expectations when structured output is needed
- Version instructions as code artifacts (not inline strings in production)

### Tool Design

- **Descriptive names:** `GetCustomerOrderHistory` not `GetData`
- **Typed parameters:** Use strongly-typed parameters with `[Description]` attributes
- **Bounded scope:** Each tool should do one thing. Avoid god-tools that accept mode switches
- **Idempotent when possible:** Tool calls may be retried by the agent
- **Structured errors:** Return descriptive error strings, not exceptions. The agent will see
  the error text and try to recover
- **No side-channel state:** Tools should not modify shared mutable state. Return results explicitly

```csharp
// Good: specific, typed, described
[Description("Gets order history for a customer by their email address. Returns last 10 orders.")]
public static async Task<string> GetOrderHistory(
    [Description("Customer email address")] string email)
{
    // validate, query, return structured result
}

// Bad: vague, untyped, no description
public static string Query(string input) { ... }
```

### Tool Count Budget

Each tool consumes context tokens (name + description + parameters). Monitor total tool
token cost:

| Tool Count | Guidance                                              |
| ---------- | ----------------------------------------------------- |
| 1–5        | Fine for most agents                                  |
| 6–15       | Test that the agent reliably selects the right tool   |
| 16+        | Split into specialist agents; tool selection degrades |

### Context and Token Budget Policy

- Keep system prompts short and procedural; move long policy text to Agent Skills
  (loaded on demand via `FileAgentSkillsProvider` progressive disclosure).
- Keep individual tool descriptions concise and action-oriented.
- Set per-request turn limits (for example 8-12) for autonomous loops.
- Use built-in compaction strategies (`PipelineCompactionStrategy`) to manage growing
  conversation history instead of manual summarization. Register `CompactionProvider`
  via `UseAIContextProviders` so compaction runs inside the tool-calling loop.
- Monitor token spikes and split responsibilities before prompt/tool bloat causes instability.
- See [SKILL.md — Compaction](../SKILL.md) for strategy selection and pipeline setup.

---

## Agent Composition Patterns

### Hierarchical (Orchestrator + Specialists)

An orchestrator agent routes to specialist agents exposed as function tools:

```
Orchestrator
├── WeatherAgent.AsAIFunction()
├── CalendarAgent.AsAIFunction()
└── EmailAgent.AsAIFunction()
```

- The orchestrator's system prompt describes when to use each specialist
- Each specialist has focused tools and instructions
- Token budget stays manageable per agent

### Workflow-Based (Graph Orchestration)

Use workflows when the execution order is known and deterministic:

```
[START] → [Classify] → (billing?) → [BillingAgent]
                      → (technical?) → [TechAgent]
                      → (other?) → [GeneralAgent]
```

- Use conditional edges for routing logic
- Capture intermediate results for audit trails
- Enable checkpointing for long-running processes

### Protocol-Based (A2A)

Use Agent-to-Agent protocol when agents run in different services:

- Each service exposes an agent card at `/.well-known/agent.json`
- Agents discover each other's capabilities via the agent card
- Communication is language-agnostic (a .NET orchestrator can call a Python specialist)

---

## Conversation Management

### Session-Based Conversation Management

`AgentSession` tracks conversation state. Pass the same session for multi-turn exchanges:

```csharp
var session = await agent.CreateSessionAsync();
Console.WriteLine(await agent.RunAsync("Hello", session));
Console.WriteLine(await agent.RunAsync("Follow up", session)); // same session preserves history
```

### Conversation Boundaries

- Set maximum turn limits to prevent runaway conversations
- Implement session timeouts for user-facing agents
- Clear conversation history when context shifts to avoid confusion

### Memory Scoping

Follow the memory governance rules from [csharp.instructions.md](../../../instructions/csharp.instructions.md):

| Memory Type | Scope               | Persistence                             | Framework Provider                   |
| ----------- | ------------------- | --------------------------------------- | ------------------------------------ |
| Episodic    | Single conversation | `InMemoryChatHistoryProvider` (default) | Built-in                             |
| Episodic    | Cross-session       | `CosmosChatHistoryProvider`             | `Microsoft.Agents.AI.CosmosNoSql`    |
| Semantic    | Cross-conversation  | `ChatHistoryMemoryProvider`             | Built-in (extracts/recalls memories) |
| Operational | Workflow state      | Durable Task checkpoint                 | Azure Functions hosting              |

### Compaction

For long-running conversations with in-memory history, register a `CompactionProvider`
to manage token budget automatically. See [SKILL.md — Compaction](../SKILL.md).

---

## Provider Selection

| Provider         | Strengths                                         | Considerations                    |
| ---------------- | ------------------------------------------------- | --------------------------------- |
| Azure OpenAI     | Enterprise security, VNet, content filtering, SLA | Requires Azure subscription       |
| Azure AI Foundry | Managed agents, hosted tools, model catalog       | Higher-level abstraction          |
| OpenAI Direct    | Latest models, fastest feature availability       | No VNet, data leaves your network |
| Anthropic        | Strong reasoning, long context                    | No Azure-native integration       |
| Ollama           | Local/offline, no data egress, free               | Limited model capabilities        |

> **Default for production:** Azure OpenAI with Managed Identity, `disableLocalAuth: true`,
> and Private Endpoint. See [private-networking](../../private-networking/SKILL.md) and
> [identity-managed-identity](../../identity-managed-identity/SKILL.md).
