# Microsoft Foundry and Hosted Agents Guidance

This reference covers when to use Foundry-hosted persistent agents versus code-managed agents.

## Decision Matrix

| Need                                                                            | Prefer                                                |
| ------------------------------------------------------------------------------- | ----------------------------------------------------- |
| Service-managed threads/history, server-side execution, enterprise governance   | Foundry hosted persistent agents                      |
| Fine-grained in-process orchestration and custom control flow                   | Code-managed agents                                   |
| Existing AG-UI orchestration already in app with domain-specific state machines | Code-managed agents (or hybrid with clear boundaries) |

## SDK/Client Orientation (.NET)

- Code-managed path commonly uses:
  - `AzureOpenAIChatClient` / `AzureAIAgentClient`
  - `AIAgent`
- Hosted persistent path typically adds:
  - `Microsoft.Agents.AI.Foundry`
  - `Azure.AI.Agents.Persistent` (still preview)
  - persistent agent/thread model APIs
- `Azure.AI.Projects` 2.0.0 GA is required for hosted agent integration as of 1.0.0.

## Configuration Musts

- Endpoint and deployment/model values must come from environment-backed options.
- Treat hosted-agent IDs and thread IDs as durable identifiers; persist and validate them.
- Include correlation IDs for every run/thread operation to support incident debugging.

## Common Hosted-Agent Traps

1. Double-owning conversation state:
   - App stores full transcript and also relies on hosted thread history without reconciliation.
2. Leaking provider-specific DTOs across layers:
   - Keep domain contracts stable; map Foundry payloads at boundaries.
3. Credential ambiguity:
   - Local/dev works with `DefaultAzureCredential` chain but production identity differs.
4. Hidden coupling between tools and thread context:
   - Validate tool calls using integration tests against representative production configs.
5. Exception-driven idempotent provisioning (NimbusIQ learning):
   - `CreateAIAgentAsync` duplicate behavior varies across SDK versions (400 vs 409 vs different error bodies).
   - Implement get-or-create pattern instead: `GetAIAgentAsync` first, create only if not found.
   - Cache agent IDs after provisioning to avoid repeated API calls.
6. Orphaned hosted runs (NimbusIQ learning):
   - Forgetting to `CancelRunAsync` when client disconnects leaves runs consuming quota.
   - Wire `CancellationToken` through all run operations and cancel on timeout/disconnect.
7. Thread cache memory growth (NimbusIQ learning):
   - Long-running services caching thread references without TTL grow unbounded.
   - Add max-cardinality eviction and emit cache telemetry.
8. Missing OpenTelemetry from the start (NimbusIQ learning):
   - Debugging distributed MAF failures without tracing is impractical.
   - Wire `configure_otel` / `ActivitySource` before first deployment.

## Hosted Agent Idempotent Provisioning Pattern (Context7-grounded)

```csharp
// Preferred: get-or-create (not exception-driven duplicate detection)
AIAgent? existing = null;
try
{
    existing = await persistentAgentsClient.GetAIAgentAsync(cachedAgentId);
}
catch (RequestFailedException ex) when (ex.Status == 404)
{
    // Agent not found; create below
}

if (existing is null)
{
    existing = await persistentAgentsClient.CreateAIAgentAsync(
        model,
        name: "assessment-agent",
        instructions: "...");
    // Cache the ID for subsequent calls
    _agentIdCache[agentName] = existing.Id;
}
```

## Recommended Hybrid Boundary

- Orchestrator app owns run lifecycle, domain state, and UX events (AG-UI).
- Foundry hosted agents own model interaction and managed conversation memory for selected tasks.
- Keep interface boundary explicit (`IAIChatService`/`IAIEnrichmentService`) so runtime mode can switch per use case.
