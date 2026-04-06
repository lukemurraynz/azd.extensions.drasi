# Create Agent Action

Step-by-step guide for creating a new Microsoft Agent Framework agent or workflow.

---

## Prerequisites

- .NET 10+ SDK or Python 3.10+
- Azure OpenAI resource (or OpenAI API key for development)
- Azure CLI authenticated (`az login`)

---

## Step 1: Install Packages

### .NET

```bash
dotnet add package Microsoft.Agents.AI.OpenAI
dotnet add package Azure.AI.OpenAI
dotnet add package Azure.Identity
```

For workflows:

```bash
dotnet add package Microsoft.Agents.AI.Workflows
```

For **Foundry-hosted persistent agents** (includes hosted MCP tools support):

```bash
dotnet add package Azure.AI.Agents.Persistent --prerelease
dotnet add package Microsoft.Agents.AI.Foundry
```

For **local MCP servers** (stdio/SSE transport via MCP C# SDK):

```bash
dotnet add package ModelContextProtocol --prerelease
```

For OpenTelemetry:

```bash
dotnet add package OpenTelemetry.Extensions.Hosting
dotnet add package OpenTelemetry.Exporter.OpenTelemetryProtocol
```

### Python

```bash
pip install agent-framework
```

---

## Step 2: Determine Agent vs Workflow

Ask these questions:

1. **Is the task open-ended or conversational?** → Agent
2. **Are there well-defined sequential or branching steps?** → Workflow
3. **Can the task be handled by a single function?** → Plain code (no agent needed)

See [agent-design.md](../standards/agent-design.md) for the full decision framework.

---

## Step 3: Create the Agent

### Option A: Single Agent with Tools

```csharp
// Program.cs or DI registration
var credential = new ManagedIdentityCredential(); // production
var aoaiClient = new AzureOpenAIClient(
    new Uri(config["AzureOpenAI:Endpoint"]!),
    credential);

var agent = aoaiClient
    .GetChatClient(config["AzureOpenAI:DeploymentName"]!)
    .AsAIAgent(
        instructions: "You are a [role]. You help with [domain].",
        tools: [AIFunctionFactory.Create(MyTools.ToolMethodA), AIFunctionFactory.Create(MyTools.ToolMethodB)]);
```

### Option B: Multi-Agent with Composition

```csharp
// Specialist agents
var specialistA = chatClient
    .AsAIAgent(
        instructions: "You are a specialist in [area A].",
        tools: [AIFunctionFactory.Create(AreaATools.Method1)]);

var specialistB = chatClient
    .AsAIAgent(
        instructions: "You are a specialist in [area B].",
        tools: [AIFunctionFactory.Create(AreaBTools.Method1)]);

// Orchestrator uses specialists as tools
var orchestrator = chatClient
    .AsAIAgent(
        instructions: """
        You coordinate tasks. Route to the appropriate specialist:
        - Use specialist_a for [area A] questions
        - Use specialist_b for [area B] questions
        """,
        tools: [
            specialistA.AsAIFunction(),
            specialistB.AsAIFunction()]);
```

### Option C: Workflow

```csharp
var workflow = new WorkflowBuilder()
    .AddExecutor("step1", agent1)
    .AddExecutor("step2", agent2)
    .AddExecutor("step3", agent3)
    .AddEdge("step1", "step2")
    .AddEdge("step2", "step3")
    .Build();
```

### Option E: Foundry Quick Start (simplest path)

Use `AIProjectClient` from `Microsoft.Agents.AI.Foundry` for the fastest getting-started experience.

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

For production, pass a `ManagedIdentityCredential` to the `AIProjectClient` constructor.
When you need hosted MCP tools, persistent threads, or server-side tool execution, graduate to Option D.

### Option D: Foundry-Hosted Agent with MCP Tools (preferred for remote tool servers)

Use `MCPToolDefinition` when you want Foundry to manage the MCP server connection server-side.
Do **not** use custom HTTP wrappers or the MCP C# SDK client for this pattern.

```csharp
using Azure.AI.Agents.Persistent;
using Azure.Identity;
using Microsoft.Agents.AI;

var endpoint = Environment.GetEnvironmentVariable("AZURE_FOUNDRY_PROJECT_ENDPOINT")
    ?? throw new InvalidOperationException("AZURE_FOUNDRY_PROJECT_ENDPOINT is not set.");
var model = Environment.GetEnvironmentVariable("AZURE_FOUNDRY_PROJECT_MODEL_ID") ?? "gpt-4.1-mini";

// Declare the hosted MCP tool — Foundry owns the server connection.
var mcpTool = new MCPToolDefinition(
    serverLabel: "microsoft_learn",
    serverUrl: "https://learn.microsoft.com/api/mcp");
mcpTool.AllowedTools.Add("microsoft_docs_search");

// Production: ManagedIdentityCredential, not DefaultAzureCredential.
var client = new PersistentAgentsClient(endpoint, new ManagedIdentityCredential());

var agentMetadata = await client.Administration.CreateAgentAsync(
    model: model,
    name: "MyAgent",
    instructions: "You are a helpful assistant.",
    tools: [mcpTool]);

AIAgent agent = await client.GetAIAgentAsync(agentMetadata.Value.Id);

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
var response = await agent.RunAsync("Your question here", session, runOptions);
Console.WriteLine(response);

// Delete persistent agents when no longer needed — they persist and incur cost.
await client.Administration.DeleteAgentAsync(agent.Id);
```

See the full pattern and key-point explanations in [SKILL.md — Hosted MCP Tools](../SKILL.md).

### Option F: Foundry Agent (versioned, server-managed definition)

Use this when agent definitions are managed and versioned on the Foundry server (created via
portal, API, or CI/CD). Your code retrieves the definition by name rather than defining it locally.

```csharp
using Microsoft.Agents.AI;
using Microsoft.Agents.AI.Foundry;
using Azure.Identity;

var projectClient = new AIProjectClient(
    endpoint: "https://your-project.services.ai.azure.com",
    credential: new ManagedIdentityCredential());

// Retrieve the server-managed agent definition by name.
var adminClient = projectClient.AgentAdministrationClient;
var agentRecord = await adminClient.GetAgentAsync("my-versioned-agent");

// Wrap it as an AIAgent — instructions, tools, and model are all server-defined.
AIAgent agent = projectClient.AsAIAgent(agentRecord);

Console.WriteLine(await agent.RunAsync("What is the capital of New Zealand?"));
```

Use this pattern when:
- Agent definitions are updated independently of code deployments
- Multiple environments share the same code but point to different agent versions
- Agent configuration is managed by a platform team or through the Foundry portal

---

## Step 4: Add Responsible AI Guardrails

1. Include scope and refusal rules in agent instructions
2. Add tool approval for any destructive operations
3. Configure content safety at the Azure OpenAI deployment level
4. Validate tool output schemas before using results for writes/routing
5. Add MCP server/tool allowlists (never open-ended tool access)

See [responsible-ai.md](../standards/responsible-ai.md).

---

## Step 5: Configure Reliability and Fallback

Set baseline reliability controls before production:

- Model timeout budget (interactive vs background requests)
- Tool timeout budget (read vs write tools)
- Retries for transient failures only (`429`, `5xx`, network timeout)
- Circuit breaker for repeated failures
- Approved fallback model/deployment and degraded response message

---

## Step 6: Add Observability

```csharp
builder.Services.AddOpenTelemetry()
    .WithTracing(t => t
        .AddSource("Microsoft.Agents.*")
        .AddAspNetCoreInstrumentation()
        .AddOtlpExporter())
    .WithMetrics(m => m
        .AddMeter("Microsoft.Agents.*")
        .AddOtlpExporter());
```

---

## Step 7: Deploy Infrastructure

1. Deploy Azure OpenAI with `disableLocalAuth: true` — use the Bicep template in [SKILL.md](../SKILL.md)
2. Create Managed Identity for the hosting resource
3. Assign `Cognitive Services OpenAI User` role
4. Configure Private Endpoint if in VNet

Use [azure-defaults](../../azure-defaults/SKILL.md) for naming and tagging.
Use [azure-deployment-preflight](../../azure-deployment-preflight/SKILL.md) to validate before deploying.

---

## Step 8: Host the Agent

Choose a hosting model:

| Host             | Best For                                  | Reference                                                           |
| ---------------- | ----------------------------------------- | ------------------------------------------------------------------- |
| ASP.NET Web API  | REST/gRPC agent endpoints                 | See SKILL.md Hosting section                                        |
| Azure Functions  | Event-driven, per-invocation billing      | [azure-functions-patterns](../../azure-functions-patterns/SKILL.md) |
| Container Apps   | Scalable containers, Dapr, jobs           | [azure-container-apps](../../azure-container-apps/SKILL.md)         |
| Azure AI Foundry | Managed agent hosting with built-in tools | Official docs                                                       |

---

## Step 9: Test with AgentEval

Run AgentEval evaluation and red team scans against the agent before deployment.
See the [agenteval skill](../../agenteval/SKILL.md) for full setup and configuration.

> **Test project only.** All AgentEval code lives in your test project (e.g., `MyAgent.Tests`),
> not the production agent project. AgentEval must never ship in production binaries.
> **No mocks.** All evaluations must run against real LLM endpoints — do not mock, stub,
> or replay cached responses. Real model behavior is what you are testing.

**Minimum test gates:**

1. **Functional evaluation** — validate tool usage correctness, response quality, and behavioral policy compliance
2. **Red team scan** — run at `Quick` intensity minimum for PR validation, `Comprehensive` for release

```csharp
// In MyAgent.Tests project — requires real Azure OpenAI endpoint
[Fact]
[Trait("Category", "AgentEval")]
public async Task Agent_passes_functional_evaluation()
{
    var results = await agent.EvaluateAsync(new EvaluationOptions
    {
        Metrics = { new ToolUsageMetric(), new ResponseQualityMetric() }
    });
    Assert.True(results.OverallScore >= 0.8);
}

[Fact]
[Trait("Category", "RedTeam")]
public async Task Agent_passes_red_team_scan()
{
    var redTeam = await agent.RedTeamAsync(new RedTeamOptions
    {
        Intensity = RedTeamIntensity.Quick,
        AttackTypes = RedTeamAttackType.All,
        OutputFormat = ReportFormat.Sarif
    });
    Assert.Equal(0, redTeam.CriticalFindings);
}
```

---

## Step 10: Validate

Run through the [production readiness checklist](../standards/checklist.md).
