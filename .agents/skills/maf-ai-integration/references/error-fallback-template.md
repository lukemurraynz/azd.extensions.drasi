# Error and Fallback Template

Use this pattern for AI-backed methods.

```csharp
public async Task<RecommendationResult> GenerateAsync(InputModel input, CancellationToken ct)
{
    var correlationId = Activity.Current?.Id ?? Guid.NewGuid().ToString("n");

    try
    {
        var aiResponse = await _aiChatService.GenerateRecommendationAsync(input, ct);

        return new RecommendationResult
        {
            Title = aiResponse.Title,
            Rationale = aiResponse.Rationale,
            Actions = aiResponse.Actions,
            ConfidenceScore = aiResponse.ConfidenceScore,
            ConfidenceSource = "ai_foundry",
            CorrelationId = correlationId
        };
    }
    catch (Exception ex)
    {
        // Differentiate transient (retry) vs permanent (log and fail) errors
        _logger.LogError(ex,
            "AI recommendation generation failed. CorrelationId={CorrelationId}, RecommendationId={RecommendationId}",
            correlationId,
            input.RecommendationId);

        // Deterministic fallback: safe, explainable, non-AI.
        return new RecommendationResult
        {
            Title = "Review configuration baseline",
            Rationale = "Fallback recommendation generated due to AI unavailability.",
            Actions = new[] { "Validate configuration against approved baseline." },
            ConfidenceScore = 0.35m,
            ConfidenceSource = "heuristic_rule",
            CorrelationId = correlationId,
            FallbackReason = "ai_unavailable"
        };
    }
}
```

## Rules

- Never return fallback output with `confidenceSource = ai_foundry`.
- Log failures with correlation IDs.
- Keep fallback deterministic and auditable.

