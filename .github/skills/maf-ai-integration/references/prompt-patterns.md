# Prompt Patterns (MAF / Azure AI Foundry)

Use these as templates and adapt per feature.

## C# Raw String Literal Pattern

```csharp
var prompt = $$"""
You are an Azure architecture reliability assistant.

Context:
- ServiceGroup: {{serviceGroupName}}
- RecommendationId: {{recommendationId}}
- WafPillar: {{wafPillar}}
- CurrentState: {{currentStateJson}}

Task:
Produce a remediation recommendation grounded in the provided context.

Output contract (JSON only):
{
  "title": "string",
  "rationale": "string",
  "actions": ["string"],
  "confidenceScore": 0.0,
  "confidenceSource": "ai_foundry"
}

Rules:
- Do not invent resources that are not in context.
- Keep actions specific and operational.
- Return valid JSON only.
""";
```

## Grounding Checklist

- Include domain identifiers (service group, resource IDs, recommendation IDs).
- Include policy/business constraints.
- Include explicit response schema.
- Include "JSON only" instruction when parsing downstream.
- Prefer short prompts with concrete context over long generic instructions.

## Structured Output Pattern (Context7-grounded, 2026-03-16)

When downstream consumers need typed deserialization, prefer `ResponseFormat.ForJsonSchema<T>()` over prompt-only JSON enforcement:

```csharp
// Type-safe structured output at agent creation
AIAgent agent = chatClient.AsAIAgent(new ChatClientAgentOptions()
{
    Name = "ExtractionAgent",
    ChatOptions = new()
    {
        Instructions = "Extract architecture assessment data from the provided context.",
        ResponseFormat = ChatResponseFormat.ForJsonSchema<AssessmentResult>()
    }
});

AgentResponse response = await agent.RunAsync(contextInput);
var result = JsonSerializer.Deserialize<AssessmentResult>(
    response.Text, JsonSerializerOptions.Web)!;
```

### Inter-Agent Structured Data Passing

Structured output enables direct message passing between agents without intermediate deserialization:

```csharp
// First agent produces structured output
AgentResponse extractionResponse = await extractionAgent.RunAsync(rawInput);

// Pass the structured output message directly to the next agent
ChatMessage structuredMessage = extractionResponse.Messages.Last();
AgentResponse summaryResponse = await summaryAgent.RunAsync(structuredMessage);
```

### When to Use Which

| Scenario                                  | Approach                                  |
| ----------------------------------------- | ----------------------------------------- |
| Typed deserialization for downstream APIs | `ResponseFormat.ForJsonSchema<T>()`       |
| Inter-agent structured data flow          | `ResponseFormat` + direct message passing |
| Exploratory/creative outputs              | Prompt-instructed JSON                    |
| Human-readable narrative generation       | No JSON constraint                        |
