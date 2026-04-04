---
applyTo: "**/*.cs"
description: "C# and .NET development best practices (Industry Solutions Engineering-aligned) with DDD-lite architectural guardrails"
---

# C# Code Instructions (Industry Solutions Engineering + DDD-lite Guardrails)

## How to Use This File

Follow repository conventions first (.editorconfig, analyzers, Directory.Build.props, Directory.Packages.props).

Apply the rules in this file next.

### Use MCP servers (only when needed)

- **iseplaybook** → ISE (Industry Solutions Engineering) C# and engineering standards
- **context7** → .NET APIs and version-specific behavior
- **microsoft.learn.mcp** → official Microsoft guidance

### Precedence order

- Repo standards → this file → MCP servers

If the repository is not layered (e.g., vertical slices or minimal APIs), apply only the guardrails that fit and do not force structure.

## Engineering Architecture Classification: Control Plane vs Data Plane

Modern cloud-native and platform systems commonly separate responsibilities into **Control Plane** and **Data Plane** components. This classification ensures architectural clarity, prevents over-engineering, and enables safe adoption of automation, orchestration, and agent-based systems.

This classification is **additive** to DDD-lite layering (see below) and does not replace Domain / Application / Infrastructure / Presentation boundaries.

---

### Classification Rule

Every service or component MUST identify whether it primarily operates as:

- Data Plane
- Control Plane
- Hybrid (rare; must document justification)

If unclear, default to **Data Plane**.

---

### Classification Declaration Requirement

New services or major components MUST explicitly document their classification (Data Plane, Control Plane, or Hybrid) in one of the following locations:

- Architecture Decision Record (ADR)
- README or service design documentation
- Service startup logging or diagnostics output (recommended)

If classification changes over time, the change MUST be documented as an architectural decision.

---

## Default Architecture Posture (New Projects)

Default to **DDD-lite** unless the repository clearly uses another style.

### DDD-lite means

- Strong layer and dependency boundaries
- Business invariants in the domain
- Minimal abstractions
- No mandatory CQRS, UoW, sagas, or eventing

### DDD-lite does not mean

- Heavy ceremony
- Over-abstracting “just in case”
- Premature bounded contexts or event sourcing

### Avoid Over-Engineering (Keep it Lean)

- Introduce patterns only when needed: Use CQRS, mediators, sagas, etc., only when there is recurring value. Avoid speculative abstractions.
- Scale architecture to project size: Vertical slices or feature folders are lean by default. Use layers only when needed per use case.
- Prefer “just enough” structure. Evolve it as the domain or team grows.

## Data Plane Definition

Data Plane components execute deterministic workloads and business operations.

### Responsibilities

- Execute business logic
- Process transactional or operational data
- Perform deterministic domain operations
- Provide service or API endpoints
- Execute workflows designed elsewhere
- Interact with persistence and external systems

### Typical Examples

- Business APIs
- Transaction processors
- Microservices
- CRUD services
- Workflow execution engines
- Deployment runners
- Background job processors
- Event handlers or reactions
- Container workloads

---

### Data Plane Guardrails

#### Deterministic Execution

- Data plane services should prioritize predictable, repeatable outcomes.
- Avoid embedding orchestration or decision planning logic beyond domain rules.

#### External Dependency Discipline

- Treat external calls as execution steps, not orchestration coordination.
- Avoid fan-out orchestration across multiple downstream systems unless part of domain invariants.

#### Domain Integrity

- Domain invariants must remain authoritative.
- Data plane components must not accept decisions that violate domain rules.

#### Observability

- Emit telemetry focused on execution performance and dependency health.
- Correlation identifiers must be propagated if provided by upstream control plane.

---

## Control Plane Definition

Control Plane components coordinate, orchestrate, and govern distributed operations across systems.

Note: Control plane orchestration operates at system or cross-service coordination level.
Application layer orchestration remains responsible for use-case or transaction level workflow within a single service boundary.

### Responsibilities

- Decide what actions should occur
- Coordinate workflows across services
- Maintain orchestration or lifecycle state
- Apply policy and governance logic
- Provide automation and provisioning orchestration
- Coordinate agents or automated reasoning systems
- Provide developer or platform automation interfaces

---

### Typical Examples

- Deployment orchestration services
- Platform landing zone automation
- Policy engines and governance services
- Workflow orchestration engines
- Internal developer platforms
- Multi-system automation coordinators
- Agent orchestration systems
- Multi-agent platforms
- Infrastructure provisioning coordinators

---

## Control Plane Guardrails

### Decision vs Execution Separation

Control plane components:

- Decide **WHAT** should occur
- Must not directly implement execution logic belonging to data plane services
- Must delegate execution to downstream services or tools
- Require stricter authorization boundaries because they can trigger multi-system side effects

---

### Scalability Constraints

Control plane components SHOULD:

- Prefer declarative reconciliation over imperative step sequencing when feasible
- Prefer asynchronous orchestration over synchronous chaining
- Avoid tight coupling to downstream service latency
- Limit synchronous blocking orchestration across multiple services

---

### Orchestration State Management

Control plane services MUST:

- Maintain explicit workflow or orchestration state
- Support recovery or replay of orchestration steps
- Track correlation identifiers across distributed operations
- Support cancellation and compensation where applicable

Orchestration state persistence MUST:

- Be stored in durable infrastructure storage
- Support versioned schema evolution
- Support replay or step rehydration
- Avoid reliance on transient memory or logs as authoritative state

State MUST NOT be implicitly derived from logs or side effects.

---

### Contract and Tool Governance

Control plane services interact with external services, tools, or workflows.

They MUST:

- Use versioned contracts
- Validate responses before downstream use
- Enforce idempotency or compensation strategies
- Define explicit failure semantics

---

### Observability Requirements

Control plane components MUST:

- Emit orchestration-level telemetry
- Record decision outcomes and transitions
- Maintain auditability of workflow progress
- Surface health status of orchestrated operations

---

### Reliability and Blast Radius

Control plane failures can impact multiple downstream systems. Therefore:

- Downstream calls must define timeouts and cancellation
- Retry policies must be bounded and jittered
- Fan-out orchestration must define concurrency limits
- Long-running orchestration must support checkpointing

---

## Hybrid Services (Rare)

A service that performs both orchestration and execution MUST:

- Clearly separate orchestration modules from execution modules
- Document boundaries explicitly
- Prefer decomposition into separate services when scale or complexity grows

Hybrid services SHOULD be treated as transitional architecture and periodically evaluated for decomposition.

---

## Agent Systems Extension (Control Plane Specialization)

Agent-based systems are a specialized form of Control Plane.

Agent frameworks introduce reasoning, tool selection, and multi-step orchestration capabilities.

Agent guardrails apply ONLY when systems:

- Use LLM or AI reasoning
- Perform dynamic tool or workflow selection
- Maintain conversational or operational memory
- Coordinate multiple autonomous components

---

### Agent Layer Responsibilities

Allowed:

- Planning and decision orchestration
- Tool selection and coordination
- Conversation or workflow management
- Memory coordination
- Reasoning chains

Not Allowed:

- Business domain invariant enforcement
- Direct infrastructure persistence or execution logic
- Tight coupling to tool schemas
- Unbounded or uncontrolled reasoning loops

---

### Tool Integration Rules

- Tools must be accessed via Application or orchestration ports
- Tool contracts must be versioned and schema validated
- Tool responses must be validated before influencing domain or orchestration decisions
- Tools must support idempotency or compensating actions

---

### Prompt and Reasoning Governance

- Prompts must be versioned artifacts
- Prompt construction must be separated from orchestration logic
- Agent outputs must be schema validated
- Reflection or self-correction loops must be bounded
- **Classify every prompt as Public or Confidential** before implementation:
  - **Public**: formatting templates, user-facing help text, response shaping with no competitive advantage. May ship in packages/images.
  - **Confidential**: competitive IP, security-sensitive instructions, orchestration strategy, tool selection criteria. Must NOT ship in packages/images.
- **Confidential prompts must be loaded at runtime** from Azure App Configuration, Key Vault, or Foundry agent store via `IPromptProvider` (or equivalent abstraction). Never hardcode as string literals or store as `prompts/*.md` files that get `COPY`'d into Docker images.
- **Never use `<EmbeddedResource>` for Confidential prompt content.** Embedded resources compile into the assembly and are extractable via ILSpy or `dotnet-ildasm`.
- **Audit `dotnet publish` output** for unintended prompt `.md` files. Add `<Content Update="prompts\**" CopyToPublishDirectory="Never" />` for Confidential prompts served from external stores.
- See `distribution-security.instructions.md` for the full framework.

---

### NuGet Package Publish Safeguards

- **Control pack content explicitly** using `Pack="false"` on items that must not ship:

```xml
<ItemGroup>
  <Content Update="prompts\**" Pack="false" />
  <Content Update="appsettings.*.json" Pack="false" />
  <None Update="**\*.md" Pack="false" Condition="'%(Filename)' != 'README'" />
</ItemGroup>
```

- **Use symbol servers for debug symbols**, not PDB inclusion in NuGet packages. Set `<IncludeSymbols>false</IncludeSymbols>` in the `.csproj` and publish symbols to Azure Artifacts or NuGet.org symbol server.
- **Inspect `.nupkg` contents before publish**: the package is a ZIP archive. Verify no `appsettings.*.json`, `prompts/`, PDB, or server-side code leaked in.
- **Set `<PrivateAssets>all</PrivateAssets>`** on analyzer, test, and internal-only package references to prevent transitive leakage.
- Run `dotnet pack --no-build` then extract and review contents in CI before pushing to a feed.

---

### Replay and Auditability (Agent Determinism)

Agent reasoning inputs and outputs MUST be captureable for replay and audit purposes when influencing orchestration or domain outcomes.

Agent decisions affecting persistent or distributed workflows MUST:

- Record input context
- Record prompt or reasoning configuration version
- Record tool invocation results

---

### Memory Governance

Agent memory types must be explicitly defined:

- Episodic (conversation history)
- Semantic (knowledge enrichment)
- Operational (workflow state)

Memory persistence must reside in Infrastructure.

Memory schema changes must include migration strategies.

---

### Multi-Agent Coordination

Agent-to-agent interactions MUST:

- Include correlation identifiers
- Support replay or recovery of coordination workflows
- Define clear ownership boundaries between agents
- Define escalation or fallback behavior

---

## Engineering Decision Guidance

### When To Use Data Plane Only

- Business logic or transactional systems
- Deterministic APIs
- Systems without orchestration or workflow coordination

---

### When To Introduce Control Plane Patterns

- Systems coordinating multiple services
- Automation or provisioning platforms
- Governance or policy enforcement services
- Workflow or orchestration platforms
- Platform developer experience tooling
- Agent or AI orchestration systems

---

### Adoption Guidance

Start with Data Plane architecture by default.

Introduce Control Plane patterns only when orchestration, governance, or automation requirements justify additional complexity.

Agent system guardrails are applied only when AI or reasoning-based orchestration is introduced.

### AI Service Registration Patterns

When integrating optional AI services (e.g., Azure AI Foundry):

- Register expensive AI clients (e.g., `AIChatService`) as **singleton** to avoid per-request overhead.
- Register services that depend on both AI + database as **scoped**, accepting the AI client as a nullable parameter:

```csharp
builder.Services.AddScoped<MyService>(sp =>
    new MyService(
        sp.GetRequiredService<DbContext>(),
        sp.GetRequiredService<ILogger<MyService>>(),
        sp.GetService<AIChatService>())); // nullable when AI not configured
```

- Inside services, guard AI usage with `_aiChatService?.IsAIAvailable == true`.
- Always wrap AI calls in try-catch; on failure, log a warning and fall back to deterministic logic.
- Tag all outputs with `confidenceSource` (`"ai_foundry"` vs `"rule_engine"`) and propagate through API DTOs to the frontend.
- Register an `IHealthCheck` that reports AI availability as `Degraded` (not `Unhealthy`) when AI is optional, since the system has a rule-based fallback.

## Performance & Efficiency Considerations

Modern compute hides inefficient code. Write software that is lean, measurable, and optimized by default:

- **Measure, don’t guess**: Use profiling tools (e.g., BenchmarkDotNet, VS Profiler) to identify bottlenecks.
- **Optimize hot paths**: Eliminate unnecessary work in loops and critical flows. Cache values, avoid redundant function calls.
- **Minimize allocations**:
  - Consider `struct`/`ref struct` only on profiled hot paths and when it reduces allocations without harming correctness or maintainability
  - Use `ArrayPool<T>`, `Span<T>`, and object pooling in high-throughput or memory-sensitive code
  - Avoid `string +=` in loops; use `StringBuilder` instead
- **Async properly**: Use `async/await` for I/O. Avoid `Task.Run` in ASP.NET Core and never block on async (`.Result`, `.Wait()`).
- **Cancellation**: All request/IO entrypoints and downstream calls should accept and pass `CancellationToken`. Background services/consumers must accept a `CancellationToken` and pass it to all I/O.
- **Don’t return IQueryable**: Never return `IQueryable` across layer/module boundaries. Inside Infrastructure you may compose queries, but you must materialize before returning to Application/Presentation.
- **Benchmark structural changes**: Don’t introduce layers (e.g., mediator pipelines, interceptors) without measuring impact.
- **Balance clarity with optimization**: Encapsulate complex perf code inside simple interfaces. Comment WHY a change improves performance.
- **Cloud cost-aware code**: Every `SaveChanges()` can have a billing impact. Avoid excessive retries, logging, or background fan-out unless needed. Use batching, timeouts, and caching to reduce compute waste.

## Required Preflight (Before Any Code Changes)

### Platform Checks

- Target Framework Moniker (TFM): `net10.0` for new projects
- C# language version: 14 (default with .NET 10 SDK)
- global.json: pin SDK version and configure test runner (see Testing section)
- `<Nullable>enable</Nullable>`
- Directory.Build.props / Directory.Packages.props
- Analyzer packages (ISE): prefer SDK analyzers via `EnableNETAnalyzers`, `AnalysisLevel`, and `AnalysisMode`. Keep `StyleCop.Analyzers` package-based.
- All analyzer, source-generator, and build-tool `PackageReference` items MUST have `PrivateAssets="all"` to prevent transitive leakage to package consumers. Without it, consumers of your library inherit unwanted analyzers.
- **Directory.Build evaluation order**: `Directory.Build.props` → SDK `.props` → `.csproj` → SDK `.targets` → `Directory.Build.targets`. Properties go in `.props`, custom targets and late-bound logic go in `.targets`. Critical pitfall: `$(TargetFramework)` conditions in `.props` files silently fail for single-targeting projects (the property is empty during `.props` evaluation); move TFM-conditional properties to `.targets`.
- **SDK-style project hygiene**: Do not restate SDK defaults (e.g., `<OutputType>Library</OutputType>`, `<EnableDefaultItems>true</EnableDefaultItems>`) as it hides intentional overrides. Do not manually list `.cs` files in SDK-style projects (implicit globbing handles it). Properties repeated across 3+ `.csproj` files belong in `Directory.Build.props`.
- **Central Package Management (CPM)**: When `Directory.Packages.props` uses `ManagePackageVersionsCentrally`, use `GlobalPackageReference` for analyzers (applies to all projects), `VersionOverride` sparingly for projects needing a different version, and always validate with a clean `dotnet restore && dotnet build` after CPM changes.

Do not change TFM, SDK, or language version unless explicitly requested.

### Architecture Preflight (DDD-lite)

Briefly state:

- Core domain concept(s)
- Layers impacted: Domain / Application / Infrastructure / Presentation
- Where invariants live
- Integration/eventing needs
- External dependencies: Only introduce resilience libraries if justified by usage/load

Adapt if the repo uses a non-DDD architecture—do not force structure.

## DDD-lite Layering Rules

### Domain Layer

Allowed:

- Entities, aggregates, value objects
- Stateless domain services
- Domain exceptions/invariants

Not Allowed:

- EF Core
- I/O, DI, config, logging
- `[Json*]` attributes or serialization logic

Performance Rules:

- Keep domain methods side-effect free (except internal state)
- Avoid direct time/ID/randomness in logic you need deterministic tests for; inject via ports (`IClock`, `IIdGenerator`) at the seam (Application/Infrastructure) and pass values in
- Avoid allocating collections/objects unnecessarily inside domain logic
- Favor value-based types (`record`, `readonly struct`) where immutable logic suffices

C# 14 Awareness:

- The `field` contextual keyword (field-backed properties) conflicts with identifiers named `field`. If a domain entity has a property or variable named `field`, use `@field` to reference the identifier or `this.field` for instance members. Prefer renaming the identifier to avoid ambiguity.
- Extension members (`extension` blocks) can define static extension methods and instance/static extension properties. Do not use extension members to add domain logic that should live inside the aggregate.

### Application Layer

Allowed:

- Orchestration and workflow
- Input validation
- Authorization
- Mapping to/from Domain
- Ports to persistence/externals

Performance Rules:

- Avoid duplicating domain logic in handlers
- Queries must be materialized before leaving Infrastructure/Application boundaries (never leak `IQueryable`)
- Return DTOs, not EF entities or `IQueryable`
- Apply `AsNoTracking()` for read-only queries at the query origin (where EF is used), not after materialization
- Filter and project as early as possible to reduce payload

### Infrastructure Layer

Allowed:

- EF Core repositories/mappings
- HTTP/messaging adapters
- Caching, telemetry, retry policies

Performance Rules:

- All outbound network calls must have explicit timeouts and propagate cancellation. Default guidance: 2–10s for typical HTTP calls unless contract requires longer; never infinite. For streaming/long-running operations, document the contract and use a different timeout explicitly.
- Use `Microsoft.Extensions.Http.Resilience` (v10+) for retry/backoff only after timeouts/cancellation are defined and calls are idempotent. The deprecated `Microsoft.Extensions.Http.Polly` package must not be used.
- When using `AddStandardResilienceHandler()`, call `options.Retry.DisableForUnsafeHttpMethods()` to prevent retries on POST/PUT/PATCH/DELETE/CONNECT. Alternatively use `options.Retry.DisableFor(HttpMethod.Post, ...)` for fine-grained control.
- Prefer per-downstream policies; avoid global "retry everything"
- Retries must be bounded and jittered; never infinite retry
- Never retry non-idempotent writes unless the operation is explicitly repeatable/idempotent (for example via repeatability headers or an explicit idempotency mechanism).
- Prefer typed `HttpClient` with timeouts + circuit breakers per service
- Pool large buffers/objects (`ArrayPool`, custom pools) for repeated workloads
- Prefer projection to DTOs over `Include` when possible; use eager loading (`Include`) when read models require navigation props
- For writes, prefer explicit transactions only when needed and avoid multiple `SaveChanges` per request
- Use `AsSplitQuery` only when required and measure impact
- Profile queries regularly for indexes and N+1 patterns

### Query Construction Safety (Required)

- Never build SQL (or SQL-like) statements by concatenating/interpolating user-controlled input.
- For EF Core, prefer LINQ expression composition over raw SQL string assembly.
- For raw SQL, use parameterized queries only; never inline values.
- For KQL/ARG-style query builders where parameterization is unavailable, enforce strict allowlists, normalize case, and escape values with a dedicated helper before interpolation.
- Dynamic filter builders must validate permitted fields/operators and reject unknown keys.

### Presentation Layer

- For HTTP APIs, apply `microsoft/api-guidelines` (vNext) appropriate to the plane (Azure data plane vs Graph). Document deviations as ADRs.
- DTO mapping and validation
- REST Level 2 (resources + HTTP verbs + correct status codes); Level 3 (optional) adds HATEOAS
- API versioning MUST use a required `api-version` query parameter (`YYYY-MM-DD` or `YYYY-MM-DD-preview`); validate strictly/invariant and never version via URL path segments
- Error responses MUST use RFC 9457 Problem Details (`application/problem+json`) and include a stable `x-error-code` response header. Include an `errorCode` extension field in the body and ensure `x-error-code` matches `errorCode`.
- DTO versioning outside the domain
- Large collections MUST use server-driven pagination with `{ "value": [...], "nextLink": "https://..." }` (absolute URL). Omit `nextLink` on the last page (never null) and include required query params (including `api-version`) in `nextLink`.
- Long-running operations (LRO) MUST return `202 Accepted` with an `operation-location` response header (absolute URL) for polling; validate inputs before returning `202` and include `api-version` in the poll URL.
- LRO start responses SHOULD also include `Location` with the same absolute status monitor URL for broad client/proxy compatibility.
- LRO status monitor responses (GET to `operation-location`) MUST include a `Retry-After` header (integer seconds) when the operation is not terminal, so clients can back off correctly.
- LRO status schemas MUST be consistent across endpoints and use explicit terminal/non-terminal values (recommended: `NotStarted | Running | Succeeded | Failed | Canceled`).
- If operation cancellation is supported, expose a documented cancel endpoint and require idempotent cancel semantics.
- POST operations that create resources or perform actions MUST be idempotent; when retries are possible, support `Repeatability-Request-ID` and `Repeatability-First-Sent` (repeatable requests) with a tracked window of at least 5 minutes. If an operation does not support repeatability headers, return `501 Not Implemented` when valid repeatability headers are present.
- For Minimal APIs, use `AddValidation()` to enable built-in request validation with `System.ComponentModel.DataAnnotations`. Validation errors integrate with `IProblemDetailsService` for consistent RFC 9457 error responses. Record types are supported as validated parameters.
- **OpenAPI 3.1**: ASP.NET Core 10 defaults to OpenAPI 3.1 via OpenAPI.NET v2.0. Breaking changes from prior versions: schema entities are now interfaces (`IOpenApiSchema` instead of `OpenApiSchema`), the `Nullable` property is removed (check `JsonSchemaType.Null` on `OpenApiSchema.Type` instead), and `OpenApiAny` is replaced by `JsonNode`. Update all schema/operation/document transformers accordingly. Use `options.OpenApiVersion = OpenApiSpecVersion.OpenApi3_1` explicitly. YAML format is available via `.yaml`/`.yml` suffix on `MapOpenApi`.

Performance Rules:

- Avoid fat controllers—offload orchestration to Application
- Do not return full aggregates unless required—trim payloads
- Avoid large object graph serialization unless justified
- Avoid returning a global `totalCount` by default for large sets (expensive). Prefer `nextLink` pagination and expose count only when required.

## Security & Auth Guardrails (ASP.NET Core)

- If code relies on `HttpContext.User`, `[Authorize]`, or role checks, the app **must** configure authentication with `AddAuthentication(...)` and `UseAuthentication()`; treat missing auth scheme configuration as a defect.
- `[AllowAnonymous]` is **only** permitted for health endpoints or explicitly documented public config endpoints; never use it for mutating endpoints.
- Demo/test endpoints that mutate data must be **disabled outside dev/test**. Enforce with environment checks or compile-time guards; in production, do not map the endpoints (or return 404/403).
- Any bypass switches (headers/query flags like `X-Bypass-RateLimit`) must be **gated to dev/test** and protected by an allowlist or internal auth. Public bypasses are not allowed.
- Token-protected callbacks (e.g., Drasi reactions) must **fail closed** when the token is missing/invalid in non-dev environments.
- If ETags/`If-Match` headers are accepted, **enforce** them and return `412 Precondition Failed` on mismatch. Add tests for concurrency behavior.
- In-memory DB fallback is **dev/test only**. In production, missing connection strings should fail fast with a clear startup error.
- Avoid buffering response bodies in logging middleware unless you actually log them; if you must capture, enforce size limits.
- If request logging excludes health endpoints, ensure the skip path matches the actual probe routes.
- Use structured logging in hot paths (e.g., `logger.LogInformation("... {Thing}", thing)`), avoid interpolated strings, and never log PII/secrets.
- For very hot paths, consider `LoggerMessage` source-generated logging.
- Rate limiting must be **cluster-aware** for multi-replica deployments. Avoid static in-memory dictionaries without eviction; prefer ASP.NET Core rate limiting with a distributed store, or an external gateway/ingress limiter. Ensure reset headers reflect the configured window duration and enforce required roles/claims if configured.
- Reuse expensive clients (e.g., `EmailClient`, `DefaultAzureCredential`, `HttpClient`) via DI and singleton/typed client patterns; avoid per-request instantiation.
- Use a **single error response shape** across middleware and controllers: RFC 9457 Problem Details (`application/problem+json`) plus an `x-error-code` header. Include an `errorCode` extension field in the body and ensure `x-error-code` equals `errorCode` (error codes are part of the contract; do not add new ones without bumping `api-version`).
- If `api-version` is used, it MUST be required and validated strictly (`YYYY-MM-DD` or `YYYY-MM-DD-preview`). Missing/unsupported versions MUST return `400 Bad Request` with `x-error-code` set to `MissingApiVersionParameter` / `UnsupportedApiVersionValue`.
- Never place secrets, passwords, or tokens in URL path/query values. Use headers or body payloads.
- Do not use empty catch blocks in production code. Handle, log safely, or rethrow.

### Telemetry & Correlation (OpenTelemetry)

- Services MUST accept and propagate W3C trace context headers (`traceparent`, `tracestate`, `baggage`) and must not reject unknown headers just because they are unrecognized.
- Error responses SHOULD include a `traceId` extension field in RFC 9457 Problem Details that matches the active trace id, so operators can correlate frontend reports to backend traces.
- Logs/metrics/traces MUST include the trace id, request path, status code, duration, and `x-error-code` when present.

## Reliability First: Auto-Detect → Auto-Declare → Auto-Communicate (ADAC)

Apply to reliability-affecting changes. Keep it brief and concrete.

ADAC declarations SHOULD be included in PR descriptions or design artifacts (e.g., ADRs) so they are reviewable and durable.

### 1) Auto-Detect (understand what you're touching)

- Identify the execution context: request pipeline, background service, queue consumer, scheduled job.
- Identify failure boundaries: network calls, storage calls, serialization, concurrency, external dependencies.
- Identify hot spots: missing timeouts/cancellation, sync-over-async, retries without idempotency, unbounded fan-out, ambiguous error handling, implicit time/ID/randomness.
- Identify contract risks: public APIs, DTOs, persistence schema, message formats.
- Identify **data format contracts**: numeric scales (0.0–1.0 vs. 0–100), category/enum values consumed by frontend, and JSON field names that frontend components depend on.

### Cross-Layer Data Contract Rules

- **Numeric scale**: Choose one scale (0.0–1.0 or 0–100) per field and document it in the DTO. Do not mix scales within the same response (e.g., `score` in 0–100 but `dimensions` in 0.0–1.0).
- **DTO field completeness**: When adding a field to a DTO or response model, populate it in ALL code paths that produce that entity. Generators, factories, and query projections must all set the field. A nullable DTO field that is never populated is a defect—either populate it or remove it from the contract.
- **Category-relevant data**: When a response contains per-category detail (e.g., dimension breakdown, top factors), tailor the data to the specific category. Returning identical generic data for all categories defeats the purpose of the detail view.
- **Taxonomy consistency**: If the same concept (e.g., violation category) has multiple taxonomies in the domain model, the DTO exposed to clients must pick one and use it consistently. Do not expose one taxonomy in summary endpoints and a different one in detail endpoints.

### 2) Auto-Declare (state reliability intent in the change)

- What can fail?
- How do we degrade?
- How do we limit blast radius?
- What is idempotent?
- What signals prove it?

If nothing changes, explicitly state "Reliability unchanged" and why.

### 3) Auto-Communicate (make failures legible)

- Callers get consistent error contracts (status codes + `x-error-code` + RFC 9457 Problem Details) or result types.
- Operators get actionable telemetry (dependency name, duration, outcome, correlation id).
- Include "how to validate" (tests, minimal checks, metrics).
- Backend error contracts (status codes, `x-error-code`, error body shape, and headers like correlation-id) are considered public API contracts and must remain stable for frontend consumers.

## Testing and Determinism

See [csharp-tests.instructions.md](csharp-tests.instructions.md) for full testing conventions (MSTest v4, integration testing, determinism, anti-patterns, MTP configuration).

## Maintainability Budgets and Generated-Code Triage

- For non-generated production code, target cyclomatic complexity <= 15 and function length <= 120 lines (justify and document exceptions).
- Prefer decomposition into focused helpers over adding branches to already-complex methods.
- Do not apply complexity gates to generated artifacts (for example `**/Migrations/*.Designer.cs`, `*ModelSnapshot.cs`); keep generated code excluded via analyzer/CI configuration.
- If a complexity warning is suppressed, require a reason, owner, and review date.

## Data Access Performance Guardrails

- Avoid `await`-inside-loop data access when calls are independent; batch or prefetch where possible.
- For EF Core read paths, project once and hydrate related data in a bounded number of queries.
- Treat N+1 query patterns as correctness/performance defects, not optional optimizations.

## Engineering Principles

- Default to internal visibility; expose public APIs deliberately
- Avoid speculative abstractions
- Comment **why**, not **what**
- Prefer smaller, safer diffs
- Do not wrap existing framework types without value
- Align with .NET platform features before introducing libraries

## Final Self-Check (Before Proposing Changes)

✅ Boundaries respected (Domain is pure)
✅ No IQueryable or EF entities leaked through Application
✅ Invariants enforced in Domain
✅ Mapping/serialization kept out of Domain
✅ Time/ID/randomness deterministic where tests require it (ports at seams)
✅ Health checks present (if ASP.NET Core)
✅ Health endpoints are reachable; logging/rate-limit skip paths match `/health/*`
✅ External services use try/catch + fallback
✅ Runtime mode auto-detected, declared in logs, and communicated via health/error contracts
✅ Minimal safe diffs
✅ Repo conventions followed
✅ API contracts (REST Level 2+)
✅ Collections paginated
✅ LRO start responses include `operation-location` (+ `Location` where applicable) and status monitor uses `Retry-After`
✅ Performance considered (hot paths, allocations)
✅ Tests updated or added
✅ Auth pipeline configured when using `HttpContext.User` / `[Authorize]`
✅ No anonymous access on mutating endpoints (demo endpoints gated to non-prod)
✅ ETag/`If-Match` enforced when supported
✅ No prod fallbacks to InMemory DB
✅ No public bypass flags for auth/rate limits
✅ OpenAPI transformers updated for OpenAPI.NET v2.0 (interfaces, `JsonNode`, no `Nullable` property)
✅ Minimal API validation configured via `AddValidation()` when using DataAnnotations
✅ Resilience handler uses `DisableForUnsafeHttpMethods()` for non-idempotent methods
