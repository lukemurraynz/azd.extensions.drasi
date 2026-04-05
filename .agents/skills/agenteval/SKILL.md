---
name: agenteval
description: >-
  Evaluate AI agents using AgentEval (.NET). Fluent assertion-based evaluation framework covering tool usage, performance SLAs, stochastic reliability, RAG quality, red team security (OWASP LLM Top 10), Responsible AI, and model comparison. Built for Microsoft.Extensions.AI, Semantic Kernel, and Microsoft Agent Framework. USE FOR: validating agent behavior, enforcing quality gates, benchmarking models, and integrating evaluation into CI/CD.
license: MIT
---

# AgentEval: Fluent Evaluation Framework for AI Agents

> [AgentEval](https://agenteval.dev) is a .NET-native evaluation framework using fluent assertions to validate agent behavior, performance, and safety.

---

# 🧭 CORE MENTAL MODEL (CRITICAL)

AgentEval is NOT just testing.

It is **fluent evaluation over execution results**:

```csharp
result.Should()
    .HaveCalledTool("AuthenticateUser")
    .BeforeTool("FetchUserData")
    .WithArgument("method", "OAuth2");
```

Everything is evaluated through:

- `result.ToolUsage`
- `result.Performance`
- `result.Response`
- `result.Trace`
- evaluation runners (stochastic, red team, etc.)

---

# ⚙️ PREREQUISITES

- .NET 8+
- Azure OpenAI or OpenAI-compatible endpoint
- Test framework (xUnit/NUnit/MSTest)
- NuGet:
```bash
dotnet add <test-project> package AgentEval --prerelease
```

---

# 🚀 QUICK START

```csharp
[Fact]
public async Task Agent_should_call_expected_tool()
{
    var result = await agent.RunAsync("Get weather in Auckland");

    result.ToolUsage.Should()
        .HaveCalledTool("get_weather")
        .WithArgument("location", "Auckland");
}
```

---

# 🧠 EVALUATION DIMENSIONS

## 1. Tool Usage (PRIMARY CAPABILITY)

```csharp
result.ToolUsage.Should()
    .HaveCalledTool("tool_name")
    .WithArgument("key", "value")
    .BeforeTool("next_tool")
    .HaveNoErrors();
```

Validates:
- Tool selection
- Argument correctness
- Call ordering
- Execution errors

---

## 2. Performance SLAs

```csharp
result.Performance.Should()
    .HaveFirstTokenUnder(TimeSpan.FromSeconds(2))
    .HaveTotalDurationUnder(TimeSpan.FromSeconds(10))
    .HaveEstimatedCostUnder(0.05m);
```

Metrics:
- TTFT (time to first token)
- Total latency
- Token usage
- Estimated cost

---

## 3. Stochastic Evaluation (NON-DETERMINISM CONTROL)

```csharp
var result = await stochasticRunner.RunStochasticTestAsync(
    agent,
    testCase,
    new StochasticOptions(Runs: 30, SuccessRateThreshold: 0.9)
);

result.Statistics.SuccessRate.Should().BeGreaterThan(0.9);
```

Purpose:
- Measure consistency
- Detect regressions
- Validate reliability

---

## 4. Model Comparison

```csharp
var result = await comparer.CompareModelsAsync(
    factories: [gpt4o, gpt4oMini],
    testCases: testSuite
);

Console.WriteLine(result.ToMarkdown());
```

Outputs:
- Accuracy
- Cost efficiency
- Ranking

---

## 5. RAG Quality Metrics

- Faithfulness (grounding)
- Relevance
- Context precision/recall

Used to detect:
- Hallucination
- Retrieval issues

---

## 6. Red Team Security

```csharp
var result = await agent.QuickRedTeamScanAsync();

result.Should()
    .HavePassed()
    .HaveMinimumScore(80);
```

Coverage:
- Prompt injection
- Jailbreaks
- PII leakage
- System prompt extraction
- OWASP LLM Top 10

---

## 7. Responsible AI

Evaluate:
- Toxicity
- Bias
- Misinformation

Used for:
- Compliance
- Enterprise governance

---

## 8. Behavioral Policies

```csharp
new BehavioralPolicy()
    .NeverCallTool("delete_record")
    .MustConfirmBefore("update_record");
```

---

## 9. Multi-Turn Evaluation

```csharp
var conversation = new ConversationBuilder()
    .User("Book a flight")
    .AssertAgentCalls("search_flights")
    .User("Upgrade to business")
    .AssertAgentCalls("update_booking");
```

---

## 10. Trace Recording + Replay (IMPORTANT)

```csharp
// Record once (real execution)
var recorder = new TraceRecordingAgent(agent);
await recorder.ExecuteAsync("Book a flight");

// Replay (deterministic, CI-safe)
var replayer = new TraceReplayingAgent(trace);
await replayer.ReplayNextAsync();
```

Use cases:
- CI pipelines
- Debugging failures
- Reproducibility

---

# 🧪 CLI USAGE

```bash
agenteval eval --config eval-config.json --output results/
```

Use for:
- Running eval suites outside code
- CI/CD pipelines
- Standardised evaluation runs

---

# 🔌 FRAMEWORK INTEGRATION

## Microsoft.Extensions.AI

```csharp
var evaluable = chatClient.AsEvaluableAgent();
```

## Semantic Kernel

```csharp
var evaluable = skAgent.AsMAFAgent().AsEvaluable();
```

## Dependency Injection (OPTIONAL)

```csharp
services.AddAgentEval();
```

Use carefully:
- Allowed in controlled environments
- Avoid leaking into production runtime unnecessarily

---

# 🔁 CI/CD INTEGRATION

## Minimal Pipeline

```yaml
- name: Run AgentEval
  run: dotnet test
```

## Advanced

- Run stochastic tests nightly
- Run red team scans pre-release
- Export reports (SARIF, Markdown, PDF)

---

# 🚦 QUALITY GATES (RECOMMENDED)

```yaml
agenteval:
  success_rate: 0.90
  tool_accuracy: 0.95
  latency_seconds: 10
  cost_per_request: 0.05
  red_team_score: 80
```

---

# ⚠️ COMMON FAILURE MODES

| Issue | Cause |
|------|------|
| Inconsistent results | No stochastic evaluation |
| Hidden tool bugs | No tool validation |
| Slow responses | No performance assertions |
| Security gaps | No red team scans |
| Hallucinations | Weak RAG grounding |

---

# 📚 KEY CAPABILITIES SUMMARY

- Fluent assertions (`Should()`)
- Tool trajectory validation
- Performance + cost SLAs
- Stochastic evaluation
- Model comparison
- Red team security
- Responsible AI metrics
- Trace replay
- CLI execution
- Cross-framework support

---

# 🧾 VERSIONING

- Pin NuGet version (prerelease)
- Monitor breaking changes
- Validate API changes before upgrades

---

# ✅ SUCCESS CRITERIA

An agent is production-ready when:

- Tool usage is correct and deterministic
- Performance SLAs are met
- Stochastic success rate ≥ 90%
- No critical security issues
- Outputs are safe and grounded
