# Pattern Selection Matrix

As of 2026-04-03, Microsoft docs list these orchestration patterns: Sequential, Concurrent, Handoff, Group Chat, Magentic.
Handoff orchestrations are marked `[Experimental]` in .NET 1.0.0.

## Choose By Problem Shape

| Problem shape                                        | Preferred pattern | Why                                      |
| ---------------------------------------------------- | ----------------- | ---------------------------------------- |
| Strict ordered steps with deterministic dependencies | Sequential        | Easiest to reason about and test         |
| Multiple independent analyses with aggregation       | Concurrent        | Reduces end-to-end latency               |
| Specialist routing based on evolving context         | Handoff           | Dynamic delegation with focused experts  |
| Iterative multi-perspective debate/collaboration     | Group Chat        | Supports collaborative reasoning loops   |
| Autonomous self-directing loops                      | Magentic          | Advanced autonomy (not currently for C#) |

## Repo-Oriented Recommendations

- `apps/agent-orchestrator`: default to Sequential or Concurrent for infra analysis and score aggregation.
- Introduce Handoff only when routing criteria are explicit and testable.
- Keep Group Chat for scenarios requiring iterative synthesis, not baseline CRUD/assessment flows.
- Do not design C# features around Magentic until official support lands.

## Hosting Model Selection (Foundry Hosted vs In-Process vs Durable Agents)

| Constraint                                                               | Prefer hosted Foundry agents | Prefer in-process agents | Prefer durable agents (Azure Functions) |
| ------------------------------------------------------------------------ | ---------------------------- | ------------------------ | --------------------------------------- |
| Need managed threads/history                                             | Yes                          | No                       | No (app-managed state)                  |
| Need tight custom runtime branching/state machine control                | No                           | Yes                      | Yes (with Durable Task)                 |
| Need service-side governance envelope (RBAC/network/compliance controls) | Yes                          | Maybe                    | Maybe (Azure Functions auth)            |
| Existing AG-UI event engine already owns state/timeline                  | Maybe (hybrid)               | Yes                      | No                                      |
| Need persistent state across invocations without custom infra            | No                           | No                       | Yes                                     |
| Need auto-generated HTTP endpoints per agent                             | No                           | No                       | Yes (`/api/agents/{name}/run`)          |
| Need checkpoint/resume for long-running workflows                        | Maybe                        | Custom impl              | Yes (built-in)                          |

Durable Agents pattern (Context7-verified):

```csharp
// Agents are registered as durable entities with auto HTTP endpoints
using IHost app = FunctionsApplication
    .CreateBuilder(args)
    .ConfigureFunctionsWebApplication()
    .ConfigureDurableAgents(options =>
    {
        options.AddAIAgent(analysisAgent);
        options.AddAIAgent(assessmentAgent);
    })
    .Build();
app.Run();
// Each agent accessible at: /api/agents/{agentName}/run
```

Rule:

- Pick one primary state owner for each user journey. Do not split authority for the same thread/run without an explicit sync design.

## Decision Flow

1. Can a single deterministic service handle it?
2. If not, are tasks independent?
3. If not, is dynamic delegation truly required?
4. If yes, add Handoff with explicit context contracts.
5. Decide where state lives (in-process or hosted-thread) before coding.

If the answer stays "no" for dynamic delegation, avoid Handoff and keep deterministic orchestration.
