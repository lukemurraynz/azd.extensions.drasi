---
name: microsoft-agent-framework
description: >-
  Microsoft Agent Framework patterns for building AI agents and multi-agent workflows. Covers agent creation, tool integration, workflow orchestration, MCP/A2A protocols, hosting, and observability. USE FOR: building AI agents, multi-agent systems, tool-using agents, agent workflows, MCP server integration, agent hosting on Azure.compatibility: Requires .NET 10+ or Python 3.10+, Azure OpenAI or OpenAI API access
---

# Microsoft Agent Framework Skill

> **Core packages reached 1.0.0 GA** (2026-04-02). Sub-packages (A2A, AG-UI, Anthropic, orchestrations) remain in preview.
> Pin package versions and monitor the [changelog](https://github.com/microsoft/agent-framework/releases).

> **Cross-reference:** Architectural governance for agent systems (tool contracts, prompt
> versioning, memory governance, replay/auditability, multi-agent coordination) is defined
> in [csharp.instructions.md — Agent Systems Extension](../../instructions/csharp.instructions.md).
> This skill covers **implementation patterns**; the instructions cover **architectural guardrails**.

> **Cross-reference:** Operational lifecycle for Foundry agents (deploy, invoke, observe,
> trace, troubleshoot) is covered by the [microsoft-foundry skill](../../../microsoft-foundry/SKILL.md)
> (user-level skill). This skill covers **SDK patterns and code**; the Foundry skill covers
> **platform operations**.

---

## Quick Reference

| Capability        | Description                                                             |
| ----------------- | ----------------------------------------------------------------------- |
| Agents            | LLM-powered units with instructions, tools, and conversation management |
| Workflows         | Graph-based orchestration with executors, edges, and checkpointing      |
| Function Tools    | Strongly-typed .NET/Python functions exposed to agents                  |
| Agent Composition | Agent-as-function-tool pattern for hierarchical orchestration           |
| Agent Skills      | Portable instruction packages with progressive disclosure loading       |
| Service Connectors | First-party connectors for Foundry, Azure OpenAI, OpenAI, Anthropic, Bedrock, Gemini, Ollama |
| Middleware Hooks  | Intercept, transform, extend agent behavior at every execution stage    |
| Compaction        | Token-budget management via truncation, summarization, and pipelines    |
| Memory Providers  | Pluggable chat history (Cosmos DB) and cross-conversation memory        |
| MCP Integration   | Model Context Protocol for external tool servers                        |
| A2A Protocol      | Agent-to-Agent communication across services and languages              |
| Declarative Agents | YAML-defined agents with instructions, tools, memory, and orchestration topology |
| Migration Assistants | Code analysis and step-by-step migration from Semantic Kernel and AutoGen |
| Hosting           | ASP.NET, Azure Functions, Container Apps, Azure AI Foundry              |
| Observability     | OpenTelemetry traces, conversation metrics, model usage tracking        |

---

## Currency and verification gates

- Last reviewed: **2026-04-07**
- Latest upstream releases at review time: .NET `dotnet-1.0.0` (2026-04-02), Python `python-1.0.0` (2026-04-02)
- `Microsoft.Agents.AI.Workflows` remains `1.0.0-rc1`, not full GA.
- Preview provider packages include `Microsoft.Agents.AI.Anthropic`, `Microsoft.Agents.AI.Bedrock`, and `Microsoft.Agents.AI.Gemini`.
- Agent Harness, DevUI, AG-UI, GitHub Copilot SDK, and Claude Code SDK are preview features.
- **1.0.0 GA breaking changes (.NET)**:
  - `Azure.AI.Projects` upgraded to 2.0.0 GA — verify import paths if upgrading from beta
  - `OpenAIAssistantClientExtensions` class removed
  - Handoff orchestrations marked `[Experimental]`
  - `Microsoft.Extensions.AI` 10.4.0+ required
- **1.0.0 GA breaking changes (Python)**:
  - `BaseContextProvider` and `BaseHistoryProvider` aliases removed
  - `Message(text=...)` constructor parameter removed — use `Message(contents=[...])`
  - `agent-framework-openai` and `agent-framework-foundry` extracted as separate packages (stable)
  - Sub-packages (A2A, AG-UI, Anthropic, orchestrations, etc.) remain beta (`1.0.0b*`)
- Core packages no longer require `--prerelease` (.NET) or `--pre` (Python).
- Before implementing, verify:
  - Provider package compatibility across `.NET` and `Python`
  - Protocol support needed for your scenario (`MCP`, `A2A`, `AG-UI`)
  - Sub-package versions if using preview integrations
- If version-dependent behavior is uncertain, include a `[VERIFY]` block in the PR with repro steps.

---

## Standards

| Standard                                      | Purpose                                           |
| --------------------------------------------- | ------------------------------------------------- |
| [Agent Design](standards/agent-design.md)     | Agent vs workflow decisions, composition, tools   |
| [Responsible AI](standards/responsible-ai.md) | Content safety, grounding, PII, prompt governance |
| [Checklist](standards/checklist.md)           | Production readiness validation                   |

---

## Actions

| Action                                  | When to use                      |
| --------------------------------------- | -------------------------------- |
| [Create Agent](actions/create-agent.md) | Building a new agent or workflow |

---

## Related Skills

| Skill                                                    | Purpose                                                                       |
| -------------------------------------------------------- | ----------------------------------------------------------------------------- |
| [AgentEval](../agenteval/SKILL.md)                       | **Required.** Evaluation, red team, and responsible AI testing for MAF agents |
| [MAF Agentic Patterns](../maf-agentic-patterns/SKILL.md) | Orchestration patterns (handoff, group chat, Magentic-One)                    |
| [MAF AI Integration](../maf-ai-integration/SKILL.md)     | AI client setup, model selection, tool calling                                |
| microsoft-foundry (user-level)                           | Foundry operational lifecycle: deploy, invoke, observe, trace, troubleshoot   |

> **Cross-reference:** When creating or modifying MAF agents, always run AgentEval functional evaluation
> and red team scans before deployment. The [create-agent](actions/create-agent.md) action includes
> AgentEval as Step 9, and the [production checklist](standards/checklist.md) requires passing scores.

## Resources

- [Official docs](https://learn.microsoft.com/en-us/agent-framework/)
- [Awesome Microsoft Agent Framework](https://github.com/webmaxru/awesome-microsoft-agent-framework)
- [GitHub repository](https://github.com/microsoft/agent-framework)
- [Migration from Semantic Kernel](https://learn.microsoft.com/en-us/agent-framework/migration-guide/from-semantic-kernel)
- [Migration from AutoGen](https://learn.microsoft.com/en-us/agent-framework/migration-guide/from-autogen)
- [DevUI](https://learn.microsoft.com/en-us/agent-framework/devui/)
- [NuGet (.NET)](https://www.nuget.org/packages/Microsoft.Agents.AI/)
- [PyPI (Python)](https://pypi.org/project/agent-framework/)
- [Discord Community](https://aka.ms/foundry/discord)
- [AgentEval](https://agenteval.dev/) — Evaluation and red team framework for AI agents

### MCP Tooling

- **`foundry-mcp`** — Validate Azure AI Foundry model deployments, endpoints, and project configuration
- **`@Azure/aks-mcp`** — Validate AKS cluster readiness when deploying agents to Kubernetes

---

## Core Concepts

### Agent vs Workflow

Microsoft Agent Framework has two primary building blocks:

| Concept      | When to Use                                                                             |
| ------------ | --------------------------------------------------------------------------------------- |
| **Agent**    | Open-ended tasks, conversational interactions, autonomous tool selection                |
| **Workflow** | Well-defined multi-step processes, deterministic control flow, multi-agent coordination |

> **Decision rule:** If you can write a function to handle the task, do that instead of
> using an AI agent. Use agents only when the task requires reasoning, judgment, or dynamic
> tool selection.

### Agent Types

| Type                      | Provider         | Use Case                                            |
| ------------------------- | ---------------- | --------------------------------------------------- |
| `ChatClientAgent`         | Any IChatClient  | Wraps any `IChatClient` implementation              |
| Foundry Responses Agent   | Azure AI Foundry | You own the agent definition; simple, flexible setup via `AIProjectClient.AsAIAgent()` |
| Foundry Agent (versioned) | Azure AI Foundry | Server-managed; agent definitions are created and versioned on Foundry, retrieved by name |
| Foundry Hosted Agent      | Azure AI Foundry | Containerized agent code deployed to Foundry Agent Service (managed runtime, scaling, observability) |
| Azure OpenAI              | Azure OpenAI     | Azure-hosted GPT models with enterprise controls    |
| OpenAI                    | OpenAI           | Direct OpenAI API access                            |
| Anthropic                 | Anthropic        | Claude models via Anthropic API                     |
| A2A Proxy                 | Any A2A server   | Proxy to a remote agent via Agent-to-Agent protocol |

> **Foundry agent type selection:** Use Responses Agent for quick prototyping where you control the agent definition in code.
> Use Foundry Agent (versioned) when agent definitions should be server-managed, versioned, and retrievable by name.
> Use Foundry Hosted Agent when you need managed container hosting with built-in scaling, observability, and enterprise controls.
> See [microsoft-foundry skill](../microsoft-foundry/SKILL.md) for operational lifecycle (deploy, invoke, observe, trace, troubleshoot).

### Service Connectors

Agent Framework ships with first-party connectors for multiple model providers.

| Provider          | Package (.NET)                         | Status  |
| ----------------- | -------------------------------------- | ------- |
| Microsoft Foundry | `Microsoft.Agents.AI.Foundry`          | Stable  |
| Azure OpenAI      | `Microsoft.Agents.AI.OpenAI`           | Stable  |
| OpenAI            | `Microsoft.Agents.AI.OpenAI`           | Stable  |
| Anthropic Claude  | `Microsoft.Agents.AI.Anthropic`        | Preview |
| Amazon Bedrock    | `Microsoft.Agents.AI.Bedrock`          | Preview |
| Google Gemini     | `Microsoft.Agents.AI.Gemini`           | Preview |
| Ollama            | Via `Microsoft.Extensions.AI` + Ollama | Preview |

Python equivalents install via `pip install agent-framework` (core) plus provider-specific extras.

Reference: [Service connectors docs](https://learn.microsoft.com/en-us/agent-framework/agents/)

### Foundry Hosting Adapter Packages

When deploying Agent Framework agents as Foundry Hosted Agents (containerized, managed runtime),
use the hosting adapter packages to wrap your agent code for the Foundry Agent Service.

| Package (.NET)                             | Purpose                                     |
| ------------------------------------------ | ------------------------------------------- |
| `Azure.AI.AgentServer.Core`               | Core hosting adapter runtime                |
| `Azure.AI.AgentServer.AgentFramework`     | Agent Framework integration for hosted agents |

| Package (Python)                           | Purpose                                     |
| ------------------------------------------ | ------------------------------------------- |
| `azure-ai-agentserver-core`               | Core hosting adapter runtime                |
| `azure-ai-agentserver-agentframework`     | Agent Framework integration for hosted agents |
| `azure-ai-agentserver-langgraph`          | LangGraph integration for hosted agents     |

These packages are separate from the Agent Framework SDK packages. They bridge your agent
code to Foundry's managed hosting environment. See the
[microsoft-foundry skill](../microsoft-foundry/SKILL.md) for the full deploy workflow
(`azd ai agent init` → `azd provision` → `azd deploy`).

Reference: [Hosted agents docs](https://learn.microsoft.com/en-us/azure/foundry/agents/concepts/hosted-agents)

### Middleware Hooks

The middleware pipeline intercepts agent execution at every stage. This includes content safety filters,
logging, compliance policies, and custom logic, without modifying agent prompts.

```csharp
// Example: Add a logging middleware to an agent
var agent = chatClient
    .AsBuilder()
    .Use(async (messages, options, next, ct) =>
    {
        logger.LogInformation("Agent invoked with {Count} messages", messages.Count);
        var response = await next(messages, options, ct);
        logger.LogInformation("Agent responded with {Tokens} tokens", response.Usage?.TotalTokenCount);
        return response;
    })
    .BuildAIAgent(new ChatClientAgentOptions
    {
        Name = "LoggedAgent",
        ChatOptions = new() { Instructions = "You are a helpful assistant." }
    });
```

Middleware supports both non-streaming and streaming handlers. When streaming UX is required,
provide both handlers to avoid degraded output.

Reference: [Middleware docs](https://learn.microsoft.com/en-us/agent-framework/agents/middleware/)

---

## Agent Creation Patterns

### .NET — Azure OpenAI Agent

```csharp
using Azure.AI.OpenAI;
using Azure.Identity;
using Microsoft.Agents.AI;
using Microsoft.Extensions.AI;

// Production: use ManagedIdentityCredential, not DefaultAzureCredential
var credential = new ManagedIdentityCredential();
var aoaiClient = new AzureOpenAIClient(
    new Uri("https://<resource>.openai.azure.com/"),
    credential);

var agent = aoaiClient
    .GetChatClient("gpt-4.1-mini")
    .AsAIAgent(instructions: "You are a helpful assistant that answers questions concisely.");
```

> **Warning:** Microsoft docs explicitly recommend `ManagedIdentityCredential` in production.
> `DefaultAzureCredential` is acceptable for local development only. See
> [identity-managed-identity](../identity-managed-identity/SKILL.md) for full guidance.

> [!NOTE]
> `ManagedIdentityCredential` works only in Azure-hosted environments. For local development, use `DefaultAzureCredential` which chains multiple credential types (Azure CLI, Visual Studio, environment variables) before falling back to managed identity. Example:
>
> ```csharp
> var credential = builder.Environment.IsDevelopment()
>     ? new DefaultAzureCredential()
>     : new ManagedIdentityCredential(clientId);
> ```

### .NET — OpenAI Agent

```csharp
using OpenAI;
using Microsoft.Agents.AI;
using Microsoft.Extensions.AI;

var client = new OpenAIClient(Environment.GetEnvironmentVariable("OPENAI_API_KEY"));
var agent = client
    .GetChatClient("gpt-4.1-mini")
    .AsAIAgent(instructions: "You are a helpful assistant.");
```

### .NET — Foundry Agent (Simplified)

```csharp
using Microsoft.Agents.AI;
using Microsoft.Agents.AI.Foundry;
using Azure.Identity;

var agent = new AIProjectClient(endpoint: "https://your-project.services.ai.azure.com")
    .GetResponsesClient("gpt-4.1-mini")
    .AsAIAgent(
        name: "Assistant",
        instructions: "You are a helpful assistant.");

Console.WriteLine(await agent.RunAsync("What is the capital of New Zealand?"));
```

> **Note:** `AIProjectClient` from `Microsoft.Agents.AI.Foundry` provides the simplest path
> to creating agents backed by Foundry-hosted models. For production, pass a
> `ManagedIdentityCredential` to the client constructor.

### .NET — Foundry Agent (Versioned, Server-Managed)

Use this pattern when agent definitions are managed on Foundry and retrieved by name.
Agents are versioned server-side, enabling updates without redeploying your application.

```csharp
using Microsoft.Agents.AI;
using Microsoft.Agents.AI.Foundry;
using Azure.Identity;

var aiProjectClient = new AIProjectClient(
    new Uri("https://your-project.services.ai.azure.com"),
    new ManagedIdentityCredential());

// Retrieve a server-managed agent by name
ProjectsAgentRecord agentRecord = await aiProjectClient
    .AgentAdministrationClient.GetAgentAsync("SupportAgent");

// Wrap as an AIAgent for use with Agent Framework
FoundryAgent agent = aiProjectClient.AsAIAgent(agentRecord);

Console.WriteLine(await agent.RunAsync("How do I reset my password?"));
```

> **When to use versioned agents:** Use this pattern when agent instructions, tools, and
> configuration should be managed centrally on Foundry (updated via portal or API) rather
> than baked into your application code. This enables non-developer updates to agent behavior.

### Python — Azure OpenAI Agent

```python
import asyncio
from agent_framework import Agent
from agent_framework.azure import AzureOpenAIChatClient
from azure.identity import ManagedIdentityCredential

credential = ManagedIdentityCredential()
client = AzureOpenAIChatClient(credential=credential)

agent = client.as_agent(
    name="assistant",
    instructions="You are a helpful assistant.",
)
```

> **Note:** Core Python packages (`agent-framework`, `agent-framework-core`, `agent-framework-openai`,
> `agent-framework-foundry`) are 1.0.0 stable. Install with `pip install agent-framework` (no `--pre`).
> Sub-packages for A2A, AG-UI, Anthropic, and other integrations remain beta and require `--pre`.
> Import paths are unified under `agent_framework`.

### Conversation Loop

```csharp
var session = await agent.CreateSessionAsync();

// Single-turn
Console.WriteLine(await agent.RunAsync("What is the weather in Auckland?", session));

// Multi-turn — pass the same session to preserve conversation history
Console.WriteLine(await agent.RunAsync("What about tomorrow?", session));
```

---

## Tool Integration

### Function Tools (.NET)

Expose .NET methods as tools the agent can call:

```csharp
using System.ComponentModel;
using Microsoft.Agents.AI;
using Microsoft.Extensions.AI;

[Description("Gets the current weather for a city")]
static string GetWeather([Description("The city name")] string city)
{
    return $"Weather in {city}: 18°C, partly cloudy";
}

// Register tools with agent via the tools parameter
var agent = chatClient
    .AsAIAgent(
        instructions: "You are a weather assistant.",
        tools: [AIFunctionFactory.Create(GetWeather)]);
```

### Tool Output Schema Validation (Required)

Treat tool output as untrusted input. Validate shape and required fields before using
tool results for routing, writes, or external side effects.

```csharp
using System.Text.Json;

public sealed record SearchResult(string Id, string Title, string SourceUrl);

static bool TryParseSearchResult(string raw, out SearchResult? result, out string? error)
{
    try
    {
        result = JsonSerializer.Deserialize<SearchResult>(raw);
        if (result is null || string.IsNullOrWhiteSpace(result.Id) || string.IsNullOrWhiteSpace(result.SourceUrl))
        {
            error = "Invalid tool payload: required fields missing.";
            return false;
        }
        error = null;
        return true;
    }
    catch (JsonException ex)
    {
        result = null;
        error = $"Invalid JSON from tool: {ex.Message}";
        return false;
    }
}
```

Python example (Pydantic):

```python
from pydantic import BaseModel, HttpUrl, ValidationError

class SearchResult(BaseModel):
    id: str
    title: str
    source_url: HttpUrl

def parse_result(raw: dict) -> SearchResult | None:
    try:
        return SearchResult.model_validate(raw)
    except ValidationError:
        return None
```

If validation fails, return a safe error to the agent and stop the side-effect path.

### Tool Approval (Human-in-the-Loop)

```csharp
// Wrap sensitive functions in ApprovalRequiredAIFunction
AIFunction deleteFunction = AIFunctionFactory.Create(DatabaseTools.DeleteRecord);
AIFunction approvalRequired = new ApprovalRequiredAIFunction(deleteFunction);

var agent = chatClient
    .AsAIAgent(
        instructions: "You are a helpful assistant.",
        tools: [approvalRequired]);

// After each run, check for pending approval requests
AgentSession session = await agent.CreateSessionAsync();
AgentResponse response = await agent.RunAsync("Delete record 42", session);

var approvalRequests = response.Messages
    .SelectMany(m => m.Contents)
    .OfType<FunctionApprovalRequestContent>()
    .ToList();

if (approvalRequests.Count > 0)
{
    var request = approvalRequests.First();
    Console.WriteLine($"Approve '{request.FunctionCall.Name}'? (y/n)");
    bool approved = Console.ReadLine()?.Trim().ToLower() == "y";
    var approval = new ChatMessage(ChatRole.User, [request.CreateResponse(approved)]);
    Console.WriteLine(await agent.RunAsync(approval, session));
}
```

### Tool Approval (Python)

For sensitive tools in Python, require explicit approval at the tool declaration level:

```python
from agent_framework import tool

@tool(approval_mode="always_require")
def sensitive_action(param: str) -> str:
    """Perform a sensitive action that requires human approval."""
    return f"Executed with {param}"
```

When tool approval is enabled, check `result.user_input_requests` after each run.
Process approval requests until all pending requests are resolved:

```python
result = await agent.run("Do something sensitive", tools=sensitive_action)
if result.user_input_requests:
    for req in result.user_input_requests:
        approved = True  # get real user input in production
        approval_msg = Message(role="user", contents=[req.create_response(approved)])
        result = await agent.run([original_query, Message(role="assistant", contents=[req]), approval_msg])
```

### Agent-as-Function-Tool Composition

Compose agents hierarchically by exposing one agent as a tool for another:

```csharp
// Specialist agent
var weatherAgent = chatClient
    .AsAIAgent(
        instructions: "You are a weather specialist.",
        tools: [AIFunctionFactory.Create(WeatherTools.GetWeather)]);

// Orchestrator agent uses specialist as a tool
var orchestrator = chatClient
    .AsAIAgent(
        instructions: "You coordinate tasks. Use the weather specialist for weather queries.",
        tools: [weatherAgent.AsAIFunction()]);
```

### MCP Server Integration

MAF supports two MCP hosting modes. Choose based on who manages the MCP server lifecycle:

| Mode                           | When to use                                                                        | Transport                                      |
| ------------------------------ | ---------------------------------------------------------------------------------- | ---------------------------------------------- |
| **Foundry-hosted MCP**         | Foundry manages the server-side connection; you want enterprise controls, no infra | `MCPToolDefinition` + `PersistentAgentsClient` |
| **Local / bring-your-own MCP** | Self-hosted or third-party MCP server you connect to yourself                      | `StdioClientTransport` (MCP C# SDK)            |

#### Hosted MCP Tools (Azure AI Foundry) — .NET

Use `MCPToolDefinition` with `PersistentAgentsClient` — no custom HTTP wrappers. Foundry
manages the connection to the remote MCP server and executes tools server-side.

Required packages:

```bash
dotnet add package Azure.AI.Agents.Persistent --prerelease
dotnet add package Microsoft.Agents.AI.Foundry
```

```csharp
using Azure.AI.Agents.Persistent;
using Azure.Identity;
using Microsoft.Agents.AI;

var endpoint = Environment.GetEnvironmentVariable("AZURE_FOUNDRY_PROJECT_ENDPOINT")
    ?? throw new InvalidOperationException("AZURE_FOUNDRY_PROJECT_ENDPOINT is not set.");
var model = Environment.GetEnvironmentVariable("AZURE_FOUNDRY_PROJECT_MODEL_ID") ?? "gpt-4.1-mini";

// Define the hosted MCP tool — Foundry connects to the server, not your app.
var mcpTool = new MCPToolDefinition(
    serverLabel: "microsoft_learn",
    serverUrl: "https://learn.microsoft.com/api/mcp");
mcpTool.AllowedTools.Add("microsoft_docs_search");

// Production: use ManagedIdentityCredential, not DefaultAzureCredential.
var persistentAgentsClient = new PersistentAgentsClient(endpoint, new ManagedIdentityCredential());

var agentMetadata = await persistentAgentsClient.Administration.CreateAgentAsync(
    model: model,
    name: "MicrosoftLearnAgent",
    instructions: "You answer questions by searching the Microsoft Learn content only.",
    tools: [mcpTool]);

// Retrieve as an AIAgent for use with Agent Framework.
AIAgent agent = await persistentAgentsClient.GetAIAgentAsync(agentMetadata.Value.Id);

// Link the MCP resource at run time and configure approval.
var runOptions = new ChatClientAgentRunOptions()
{
    ChatOptions = new()
    {
        RawRepresentationFactory = (_) => new ThreadAndRunOptions()
        {
            ToolResources = new MCPToolResource(serverLabel: "microsoft_learn")
            {
                RequireApproval = new MCPApproval("never"),
            }.ToToolResources()
        }
    }
};

var session = await agent.CreateSessionAsync();
var response = await agent.RunAsync(
    "How do I create an Azure Function app?",
    session,
    runOptions);
Console.WriteLine(response);

// Always clean up persistent agents when they are no longer needed.
await persistentAgentsClient.Administration.DeleteAgentAsync(agent.Id);
```

**Key points:**

- `serverLabel` is a stable identifier used to link `MCPToolDefinition` → `MCPToolResource` at run time
- `AllowedTools` restricts which tools on that server the agent may invoke (allowlist)
- `RequireApproval("never")` lets the agent call tools automatically; use `"always"` for HITL
- Delete persistent agents when finished to avoid resource and billing accumulation
- Reference: [Hosted MCP tools docs](https://learn.microsoft.com/en-us/agent-framework/agents/tools/hosted-mcp-tools?pivots=programming-language-csharp)

#### MCP Trust-Boundary Policy (Required)

For both hosted and local MCP setups:

- Allowlist approved MCP servers by URL/host and reject unknown servers.
- Restrict tool surface with explicit `AllowedTools` lists (never "all tools").
- Pass auth at runtime from managed identity or secret stores, never from prompts or hardcoded strings.
- Validate every MCP tool response schema before using it in business logic.
- Re-authorize sensitive operations after tool output (tool success does not imply user authorization).
- Log MCP calls as metadata only (tool name, duration, status); avoid full payload logging by default.

#### Local MCP Server (stdio) — .NET

Use `StdioClientTransport` from the MCP C# SDK when you run the MCP server as a child process.

Required package:

```bash
dotnet add package ModelContextProtocol --prerelease
```

```csharp
using ModelContextProtocol.Client;
using Microsoft.Agents.AI;
using Azure.AI.OpenAI;
using Azure.Identity;

// Launches the MCP server as a child process over stdin/stdout.
await using var mcpClient = await McpClientFactory.CreateAsync(
    new StdioClientTransport(new()
    {
        Name = "FileSystemServer",
        Command = "npx",
        Arguments = ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/dir"],
    }));

var mcpTools = await mcpClient.ListToolsAsync().ConfigureAwait(false);

var agent = new AzureOpenAIClient(new Uri(endpoint), new ManagedIdentityCredential())
    .GetChatClient(deploymentName)
    .AsAIAgent(
        instructions: "Use available tools to help the user.",
        tools: [.. mcpTools.Cast<AITool>()]);

Console.WriteLine(await agent.RunAsync("List files in the current directory."));
```

- Reference: [Local MCP tools docs](https://learn.microsoft.com/en-us/agent-framework/agents/tools/local-mcp-tools?pivots=programming-language-csharp)

#### Learn MCP (HTTP/Streamable) — Python (Agent Framework)

Use `MCPStreamableHTTPTool` and pass it directly as a tool to the agent.
For local MCP servers, use `MCPStdioTool` (subprocess transport) or `MCPWebsocketTool` (WebSocket transport):

```python
import asyncio
from agent_framework import Agent, MCPStreamableHTTPTool
from agent_framework.azure import AzureAIAgentClient
from azure.identity.aio import AzureCliCredential

async def learn_mcp_example():
    async with (
        AzureCliCredential() as credential,
        MCPStreamableHTTPTool(
            name="Microsoft Learn MCP",
            url="https://learn.microsoft.com/api/mcp",
        ) as learn_mcp,
        Agent(
            chat_client=AzureAIAgentClient(credential=credential),
            name="DocsAgent",
            instructions="Answer Microsoft documentation questions.",
        ) as agent,
    ):
        result = await agent.run(
            "How do I create an Azure Function app?",
            tools=learn_mcp,
        )
        print(result)

if __name__ == "__main__":
    asyncio.run(learn_mcp_example())
```

### Reliability Defaults (Model + Tool Calls)

Apply these defaults unless you have service-specific requirements:

| Control         | Default                                      | Notes                                                        |
| --------------- | -------------------------------------------- | ------------------------------------------------------------ |
| Model timeout   | 30s interactive / 90s background             | Fail fast and return actionable fallback messaging           |
| Tool timeout    | 10s read tools / 30s write tools             | Keep write tools explicit and approval-gated                 |
| Retry policy    | Max 3 attempts, exponential backoff + jitter | Retry transient errors only (`429`, `5xx`, network timeouts) |
| Circuit breaker | Open after 5 consecutive failures in 60s     | Enter degraded mode and stop immediate retries               |
| Turn limit      | Max 8-12 turns per request                   | Prevent runaway loops and token exhaustion                   |

If these thresholds are exceeded, stop autonomous execution and escalate to human review.

### Compaction (Token Budget Management)

Use the built-in compaction framework to manage conversation history size. Compaction
applies only to agents with in-memory history (not Foundry hosted agents or Responses API
with `store` enabled, which manage context server-side).

> **Experimental.** Requires `#pragma warning disable MAAI001` in .NET.

**Built-in strategies** (from gentlest to most aggressive):

| Strategy                          | Effect                                                            | LLM required? | Best for                                  |
| --------------------------------- | ----------------------------------------------------------------- | ------------- | ----------------------------------------- |
| `ToolResultCompactionStrategy`    | Collapses old tool-call groups into `[Tool calls: ...]` summaries | No            | Reclaiming space from verbose tool output |
| `SummarizationCompactionStrategy` | LLM-summarizes older conversation spans                           | Yes           | Long conversations where context matters  |
| `SlidingWindowCompactionStrategy` | Drops oldest user turns (respects turn boundaries)                | No            | Hard turn-count limits                    |
| `TruncationCompactionStrategy`    | Drops oldest non-system message groups                            | No            | Emergency token-budget backstops          |
| `PipelineCompactionStrategy`      | Chains strategies sequentially (gentle → aggressive)              | Depends       | Layered compaction with fallbacks         |

**Recommended production pipeline:**

```csharp
#pragma warning disable MAAI001

IChatClient summarizerClient = openAIClient.GetChatClient("gpt-4.1-mini").AsIChatClient();

PipelineCompactionStrategy compaction = new(
    new ToolResultCompactionStrategy(CompactionTriggers.TokensExceed(0x200)),
    new SummarizationCompactionStrategy(summarizerClient, CompactionTriggers.TokensExceed(0x500)),
    new SlidingWindowCompactionStrategy(CompactionTriggers.TurnsExceed(4)),
    new TruncationCompactionStrategy(CompactionTriggers.TokensExceed(0x8000)));

AIAgent agent = agentChatClient
    .AsBuilder()
    .UseAIContextProviders(new CompactionProvider(compaction))
    .BuildAIAgent(new ChatClientAgentOptions
    {
        Name = "Assistant",
        ChatOptions = new() { Instructions = "You are a helpful assistant." }
    });
```

**Key decisions:**

- Register `CompactionProvider` via `UseAIContextProviders` on the `ChatClientBuilder`
  (not via `ChatClientAgentOptions.AIContextProviders`) so compaction runs inside the
  tool-calling loop without polluting persisted chat history.
- Use a smaller/cheaper model for `SummarizationCompactionStrategy` to reduce cost.
- Set `MinimumPreserved` to keep recent turns visible (defaults: Truncation=32, SlidingWindow=1, Tool=2, Summarization=4).

Reference: [Compaction docs](https://learn.microsoft.com/agent-framework/agents/conversations/compaction)

### Model Fallback Strategy (429/5xx)

1. Retry the same model with bounded backoff for transient failures.
2. If failures persist, switch to a pre-approved fallback deployment/model.
3. Return degraded-mode status to the caller (include correlation ID).
4. Escalate to human review when fallback also fails or policy checks block completion.

---

## Agent Skills

Agent Skills are portable packages of instructions, scripts, and resources that give
agents specialized capabilities via a progressive disclosure pattern. The framework
loads skills through `FileAgentSkillsProvider`, which discovers `SKILL.md` files and
exposes three tools to the agent: `load_skill`, `read_skill_resource`, and `run_skill_script`.

### Progressive Disclosure (3-Stage)

| Stage              | Token cost        | Trigger                                                                 |
| ------------------ | ----------------- | ----------------------------------------------------------------------- |
| **Advertise**      | ~100 tokens/skill | Skill names and descriptions injected into system prompt at run start   |
| **Load**           | < 5000 tokens     | Agent calls `load_skill` when a task matches a skill domain             |
| **Read resources** | As needed         | Agent calls `read_skill_resource` for reference docs, templates, assets |

### Setup (.NET)

```csharp
using Microsoft.Agents.AI;

var skillsProvider = new FileAgentSkillsProvider(
    skillPath: Path.Combine(AppContext.BaseDirectory, "skills"));

AIAgent agent = new AIProjectClient(
    new Uri(endpoint), new ManagedIdentityCredential())
    .AsAIAgent(new ChatClientAgentOptions
    {
        Name = "SkillsAgent",
        ChatOptions = new()
        {
            ModelId = deploymentName,
            Instructions = "You are a helpful assistant.",
        },
        AIContextProviders = [skillsProvider],
    });
```

### Multiple Skill Directories

```csharp
var skillsProvider = new FileAgentSkillsProvider(
    skillPaths: [
        Path.Combine(AppContext.BaseDirectory, "company-skills"),
        Path.Combine(AppContext.BaseDirectory, "team-skills"),
    ]);
```

### Skill vs Workflow Decision

| Signal                                    | Use Skill | Use Workflow |
| ----------------------------------------- | --------- | ------------ |
| AI decides execution approach             | ✅        |              |
| Steps must execute in guaranteed order    |           | ✅           |
| Operations are idempotent/low-risk        | ✅        |              |
| Steps produce non-idempotent side effects |           | ✅           |
| Single-domain, focused task               | ✅        |              |
| Multi-step with human approvals           |           | ✅           |

Reference: [Agent Skills docs](https://learn.microsoft.com/agent-framework/agents/skills/)

---

## Declarative Agents and Workflows

Define agents and orchestration topology in version-controlled YAML files, then load and
run them with a single API call. Suitable for repetitive flows that change frequently
without code deploys.

```yaml
# agent.yaml
name: SupportAgent
instructions: You are a customer support agent.
model: gpt-4.1-mini
tools:
  - name: search_docs
    type: mcp
    server_url: https://learn.microsoft.com/api/mcp
memory:
  provider: foundry
```

```csharp
// Load and run a declarative agent
var agent = await AgentFactory.CreateFromYamlAsync("agent.yaml", chatClient);
Console.WriteLine(await agent.RunAsync("How do I reset my password?"));
```

Reference: [Declarative agents docs](https://learn.microsoft.com/en-us/agent-framework/agents/declarative)

---

## Chat History and Memory Providers

The framework provides pluggable `AIContextProvider` implementations for persistent
conversation history and cross-conversation memory. Agent Memory can be backed by
Foundry Agent Service, Mem0, Redis, Neo4j, or custom stores.

### Chat History Providers

Customize how conversation history is stored when the agent manages its own history
(not applicable to Foundry hosted agents or Responses API with `store` enabled).

| Provider                      | Status  | Use case                                   |
| ----------------------------- | ------- | ------------------------------------------ |
| `InMemoryChatHistoryProvider` | Preview | Default; single-process, no persistence    |
| `CosmosChatHistoryProvider`   | Preview | Durable cross-session history in Cosmos DB |

### Memory AI Context Providers

Extract and recall memories across conversations (semantic memory):

| Provider                    | Status  | Use case                                                           |
| --------------------------- | ------- | ------------------------------------------------------------------ |
| `ChatHistoryMemoryProvider` | Preview | Extracts memories from messages; retrieves relevant ones per query |
| Mem0 Provider               | Preview | Cross-conversation memory via Mem0                                 |
| Redis Provider              | Preview | Redis-backed memory and retrieval                                  |
| Foundry Memory              | Preview | Managed memory via Microsoft Foundry Agent Service                 |

### RAG AI Context Providers

| Provider                | Status  | Use case                                           |
| ----------------------- | ------- | -------------------------------------------------- |
| Neo4j GraphRAG Provider | Preview | Graph-based retrieval augmented generation          |
| Text Search Provider    | Preview | Vector/keyword search over document stores          |

Mem0, Redis, and Foundry Memory providers listed above also support RAG retrieval patterns.

Context providers are registered via `AIContextProviders` on `ChatClientAgentOptions`
or via `UseAIContextProviders` on `ChatClientBuilder`.

Reference: [Integrations docs](https://learn.microsoft.com/agent-framework/integrations/)

---

## Workflow Patterns

### Core Workflow Concepts

| Concept      | Description                                                    |
| ------------ | -------------------------------------------------------------- |
| **Executor** | A processing unit (agent, function, or external call)          |
| **Edge**     | Connection between executors with optional conditional routing |
| **Event**    | Lifecycle hook for observability (started, completed, error)   |
| **Builder**  | Fluent API for constructing the directed graph                 |

### Sequential Workflow

```csharp
using Microsoft.Agents.AI.Workflows;

var workflow = new WorkflowBuilder()
    .AddExecutor("research", researchAgent)
    .AddExecutor("summarize", summaryAgent)
    .AddEdge("research", "summarize")
    .Build();

var result = await workflow.InvokeAsync(input);
```

### Conditional Routing

```csharp
var workflow = new WorkflowBuilder()
    .AddExecutor("classify", classifierAgent)
    .AddExecutor("handle-billing", billingAgent)
    .AddExecutor("handle-technical", technicalAgent)
    .AddConditionalEdge("classify", result =>
        result.Contains("billing") ? "handle-billing" : "handle-technical")
    .Build();
```

### Concurrent Execution

```csharp
var workflow = new WorkflowBuilder()
    .AddExecutor("fetch-weather", weatherAgent)
    .AddExecutor("fetch-news", newsAgent)
    .AddExecutor("summarize", summaryAgent)
    // Both run in parallel, then converge
    .AddEdge(WorkflowBuilder.START, "fetch-weather")
    .AddEdge(WorkflowBuilder.START, "fetch-news")
    .AddEdge("fetch-weather", "summarize")
    .AddEdge("fetch-news", "summarize")
    .Build();
```

### Checkpointing (Durable Workflows)

For long-running or mission-critical workflows, use the Durable Task extension:

```csharp
// Integrates with Azure Durable Task for checkpointing and recovery
builder.Services.AddDurableTaskWorkflow(options =>
{
    options.AddWorkflow<MyAgentWorkflow>();
});
```

---

## Multi-Agent Orchestration Patterns

| Pattern        | Description                                                  | Use Case                                             |
| -------------- | ------------------------------------------------------------ | ---------------------------------------------------- |
| **Sequential** | Agents process in fixed order, each building on the previous | Research → Draft → Review pipeline                   |
| **Concurrent** | Multiple agents run in parallel, results merged              | Parallel data gathering from different sources       |
| **Hand-off**   | Agent transfers control to a specialist based on context     | Triage agent routing to billing/technical/sales      |
| **Magentic**   | Dynamic multi-agent discussion with a moderator              | Brainstorming, debate, collaborative problem-solving |

---

## Hosting

### ASP.NET Web API

```csharp
var builder = WebApplication.CreateBuilder(args);

// Register agent as a service
builder.Services.AddSingleton(sp =>
{
    var credential = new ManagedIdentityCredential();
    var client = new AzureOpenAIClient(
        new Uri(builder.Configuration["AzureOpenAI:Endpoint"]!),
        credential);

    return client
        .GetChatClient(builder.Configuration["AzureOpenAI:DeploymentName"]!)
        .AsAIAgent(instructions: "You are a helpful assistant.");
});

var app = builder.Build();

app.MapPost("/chat", async (AIAgent agent, ChatRequest request) =>
{
    var session = await agent.CreateSessionAsync();
    var response = await agent.RunAsync(request.Message, session);
    return Results.Ok(new { reply = response.ToString() });
});
```

### Azure Functions (Isolated Worker)

```csharp
[Function("ChatAgent")]
public async Task<HttpResponseData> Run(
    [HttpTrigger(AuthorizationLevel.Function, "post")] HttpRequestData req)
{
    var request = await req.ReadFromJsonAsync<ChatRequest>();
    var session = await _agent.CreateSessionAsync();
    var response = await _agent.RunAsync(request!.Message, session);

    var httpResponse = req.CreateResponse(HttpStatusCode.OK);
    await httpResponse.WriteAsJsonAsync(new { reply = response.ToString() });
    return httpResponse;
}
```

---

## Infrastructure (Bicep)

### Azure OpenAI with Managed Identity

```bicep
param location string = resourceGroup().location
param openAiName string
param deploymentName string = 'gpt-4.1-mini'

resource openAi 'Microsoft.CognitiveServices/accounts@2025-12-01' = {
  name: openAiName
  location: location
  kind: 'OpenAI'
  sku: { name: 'S0' }
  properties: {
    publicNetworkAccess: 'Disabled'
    disableLocalAuth: true // Force Managed Identity — no API keys
    networkAcls: {
      defaultAction: 'Deny'
    }
  }
}

resource deployment 'Microsoft.CognitiveServices/accounts/deployments@2025-12-01' = {
  parent: openAi
  name: deploymentName
  properties: {
    model: {
      format: 'OpenAI'
      name: 'gpt-4.1-mini'
      version: '2025-04-14'
    }
  }
  sku: {
    name: 'GlobalStandard'
    capacity: 10
  }
}

// Grant the Container App identity access
resource roleAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(openAi.id, containerAppIdentity.id, cognitiveServicesOpenAIUserRole)
  scope: openAi
  properties: {
    roleDefinitionId: cognitiveServicesOpenAIUserRole
    principalId: containerAppIdentity.properties.principalId
    principalType: 'ServicePrincipal'
  }
}

var cognitiveServicesOpenAIUserRole = subscriptionResourceId(
  'Microsoft.Authorization/roleDefinitions',
  '5e0bd9bd-7b93-4f28-af87-19fc36ad61bd' // Cognitive Services OpenAI User
)
```

> **Warning — `disableLocalAuth: true`:** Always disable API key access and use RBAC.
> See [identity-managed-identity](../identity-managed-identity/SKILL.md) and
> [azure-role-selector](../azure-role-selector/SKILL.md) for role assignment patterns.

---

## Observability

### OpenTelemetry Integration

```csharp
builder.Services.AddOpenTelemetry()
    .WithTracing(tracing => tracing
        .AddSource("Microsoft.Agents.*")
        .AddAspNetCoreInstrumentation()
        .AddHttpClientInstrumentation()
        .AddOtlpExporter())
    .WithMetrics(metrics => metrics
        .AddMeter("Microsoft.Agents.*")
        .AddAspNetCoreInstrumentation()
        .AddHttpClientInstrumentation()
        .AddOtlpExporter());
```

### Key Metrics to Monitor

| Metric                  | Description                                  | Alert Threshold            |
| ----------------------- | -------------------------------------------- | -------------------------- |
| Agent invocation count  | Total agent calls                            | Baseline ± 3σ              |
| Tool call success rate  | Percentage of tool calls that succeed        | < 95% → warning            |
| Token usage per request | Input + output tokens consumed               | Budget-dependent           |
| Conversation turn count | Average turns per conversation               | > 10 → review agent design |
| Latency P95             | 95th percentile end-to-end response time     | > 10s → investigate        |
| Model error rate        | 429 (throttled) and 500 errors from provider | > 1% → alert               |

> **Cross-reference:** See [observability-monitoring](../observability-monitoring/SKILL.md) for
> Application Insights, KQL dashboards, and alert configuration patterns.

---

## Protocols

### MCP (Model Context Protocol)

Connect agents to external tool servers. The framework supports both hosted and local MCP:

- **Hosted MCP:** Cloud-hosted tool servers managed by the provider
- **Local MCP:** Self-hosted servers using stdio or SSE transport

### A2A (Agent-to-Agent Protocol)

Enable cross-service agent communication:

```csharp
// Create a proxy to a remote agent
var remoteAgent = new A2AAgent(
    new Uri("https://remote-agent-service.example.com/.well-known/agent.json"));

// Use the remote agent as a tool
var orchestrator = chatClient
    .AsAIAgent(
        instructions: "Coordinate with the remote specialist.",
        tools: [remoteAgent.AsAIFunction()]);
```

---

## Migration from Semantic Kernel and AutoGen

For teams coming from Semantic Kernel or AutoGen, migration assistants analyze existing
code and generate step-by-step migration plans.

| Source Framework | Guide |
| ---------------- | ----- |
| Semantic Kernel  | [Migration guide](https://learn.microsoft.com/en-us/agent-framework/migration-guide/from-semantic-kernel) |
| AutoGen          | [Migration guide](https://learn.microsoft.com/en-us/agent-framework/migration-guide/from-autogen) |

Key migration notes.

- Agent Framework unifies Semantic Kernel's enterprise foundations with AutoGen's orchestrations
- Core `IChatClient` abstraction from `Microsoft.Extensions.AI` is the shared integration point
- Existing Semantic Kernel plugins can be adapted as Agent Framework function tools
- AutoGen multi-agent patterns map to Agent Framework orchestrations (Sequential, Concurrent, Handoff, Group Chat)

## Preview Features (v1.0 Release)

These features shipped with v1.0 but remain in preview. APIs may evolve based on feedback.

### DevUI

Browser-based local debugger for visualizing agent execution, message flows, tool calls,
and orchestration decisions in real time. Install and launch alongside your agent application
during development.

Reference: [DevUI docs](https://learn.microsoft.com/en-us/agent-framework/devui/)

### AG-UI / CopilotKit / ChatKit

Stream agent output to frontend surfaces with adapters for CopilotKit and ChatKit,
including tool execution status and human-in-the-loop flows.

Reference: [AG-UI docs](https://learn.microsoft.com/en-us/agent-framework/integrations/ag-ui/)

### GitHub Copilot SDK and Claude Code SDK

Use GitHub Copilot or Claude Code as an agent harness directly from Agent Framework
orchestration code. These SDKs handle autonomous agent loops (planning, tool execution,
file edits, session management) and Agent Framework wraps them, composing coding-capable
agents alongside other agents in multi-agent workflows.

Reference: [GitHub Copilot provider docs](https://learn.microsoft.com/en-us/agent-framework/agents/providers/github-copilot)

### Agent Harness

Customizable local runtime giving agents access to shell, filesystem, and messaging loop.
Supports both local execution (with approval gates) and hosted execution in managed
container environments.

**Local shell with approvals (.NET)**

```csharp
using System.ComponentModel;
using System.Diagnostics;

[Description("Execute a shell command locally.")]
static string RunBash([Description("Bash command")] string command)
{
    using var process = Process.Start(new ProcessStartInfo
    {
        FileName = "/bin/bash",
        ArgumentList = { "-lc", command },
        RedirectStandardOutput = true,
        RedirectStandardError = true,
        UseShellExecute = false,
    });
    process!.WaitForExit(30_000);
    return $"stdout:\n{process.StandardOutput.ReadToEnd()}\nstderr:\n{process.StandardError.ReadToEnd()}\nexit_code:{process.ExitCode}";
}

var agent = chatClient.AsAIAgent(
    name: "ShellAgent",
    instructions: "Use tools when needed. Avoid destructive commands.",
    tools: [new ApprovalRequiredAIFunction(AIFunctionFactory.Create(RunBash))]);
```

> **Security:** Run shell-capable agents in isolated environments. Keep explicit approval
> gates before command execution.

Reference: [Agent Harness docs](https://devblogs.microsoft.com/agent-framework/agent-harness-in-agent-framework/)

### Foundry Hosted Agent Integration

Run Agent Framework agents as managed services on Microsoft Foundry or as Azure Durable
Functions with automatic checkpointing and HTTP endpoint generation.

**Built-in server-side tools** (run on Foundry, not in your code):

| Tool              | Description                                       |
| ----------------- | ------------------------------------------------- |
| File Search       | Retrieval over uploaded files and vector stores    |
| Code Interpreter  | Sandboxed code execution with file I/O            |
| Web Search (Bing) | Grounded web search via Bing                      |
| Memory            | Managed cross-conversation memory                 |
| MCP Servers       | Remote MCP tools hosted and connected server-side |

These tools are declared via `MCPToolDefinition`, `FileSearchToolDefinition`, or
`CodeInterpreterToolDefinition` on the agent definition. They execute server-side
on the Foundry runtime, not in your application process.

**Hosted agent lifecycle:** create, start, update, stop, delete. Deployable via
`azd ai agent init` followed by `azd provision` and `azd deploy`.

**Replica sizing:** Hosted agents support CPU/memory replica size configuration.
Autoscaling is managed by the Foundry runtime.

**Observability:** End-to-end tracing and metrics via Application Insights integration.
Log streaming available through the REST API.

**Identity/Security:** Entra identity with RBAC. Content filters applied at the model level.

> **Caveat:** Hosted agents (preview) do NOT support private networking. Prompt agents
> and workflow agents do. Plan network topology accordingly.

> **Operational lifecycle:** For deploy, invoke, observe, trace, and troubleshoot workflows,
> see the [microsoft-foundry skill](../../../microsoft-foundry/SKILL.md) (user-level skill).
> This skill covers SDK patterns and code; the Foundry skill covers platform operations.

Reference: [Foundry hosted agents](https://learn.microsoft.com/en-us/azure/foundry/agents/concepts/hosted-agents) |
[Foundry integration docs](https://learn.microsoft.com/en-us/agent-framework/integrations/azure-functions) |
[Sample repo](https://github.com/Azure-Samples/foundry-hosted-agents-dotnet-demo)

---

## Common Pitfalls

> **Warning — Token budget exhaustion:** Agents with many tools or long system prompts
> consume significant input tokens per invocation. Monitor token usage and keep tool
> descriptions concise. Consider splitting into specialist agents if exceeding context limits.

> **Warning — Unbounded conversation loops:** Multi-turn agents without turn limits can
> enter infinite retry loops. Always set a maximum turn count or implement circuit-breaker
> logic per the [csharp.instructions.md](../../instructions/csharp.instructions.md)
> Agent Systems Extension bounded reasoning rule.

> **Warning — Tool exception leaking:** Unhandled exceptions in tool functions surface
> raw error messages to the LLM, which may hallucinate recovery actions. Catch tool
> exceptions and return structured error descriptions.

> **Warning — Unvalidated tool output:** Treating MCP/function output as trusted can
> propagate malformed or adversarial data into business actions. Validate tool output
> schemas before writes, approvals, or routing.

> **Warning — Ignored approval requests:** In Python flows, tool calls can remain pending
> if `user_input_requests` are not processed. Always implement an approval handling loop
> until no approval requests remain.

> **Warning — DefaultAzureCredential in production:** The framework samples use
> `DefaultAzureCredential` for convenience. In production, switch to
> `ManagedIdentityCredential` to avoid credential chain ambiguity and latency.
> See [identity-managed-identity](../identity-managed-identity/SKILL.md).

> **Warning — Missing fallback policy:** Repeated `429`/`5xx` responses without a
> fallback plan can cascade into user-visible outages. Implement bounded retries plus
> approved fallback deployments and degraded-mode responses.

> **Warning — Hosted agent private networking:** Foundry hosted agents (preview) do not
> support private networking. If your architecture requires VNet isolation, use prompt
> agents or workflow agents instead, or self-host via Container Apps / Azure Functions.

## Currency and Verification

- **Date checked:** 2026-03-31 (verified via Microsoft Learn MCP — ARM template references)
- **Compatibility:** Azure Bicep, ARM templates, Azure AI Services
- **Sources:** [Microsoft.CognitiveServices ARM reference](https://learn.microsoft.com/azure/templates/microsoft.cognitiveservices/accounts)
- **Verification steps:**
  1. Run `az provider show --namespace Microsoft.CognitiveServices --query "resourceTypes[?resourceType=='accounts'].apiVersions" -o tsv` and confirm `2025-12-01` is listed
  2. Run `az bicep build --file <your-bicep-file>` to validate syntax

> **Note:** Some packages referenced in this skill (`Azure.AI.Agents.Persistent`, `ModelContextProtocol`) remain in preview. Core `Microsoft.Agents.AI.*` packages are 1.0.0 GA. Verify package versions with `dotnet list package --outdated` periodically.
