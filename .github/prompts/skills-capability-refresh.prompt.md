---
agent: agent
description: Refresh skills and instructions against the latest documented capabilities and known pitfalls
tools:
  [
    read,
    edit,
    search,
    web,
    browser,
    "azure-mcp/*",
    "bicep/*",
    "context7/*",
    azure/search,
    "iseplaybook/*",
    "microsoft.learn.mcp/*",
    "MRC-MCP-Server/*",
  ]
---

Please perform a **documentation-driven capability refresh** for this repository’s AI customization assets.

Scope:

- `.github/skills/**/SKILL.md`
- `.github/skills/**/standards/**/*.md`
- `.github/instructions/**/*.md`
- `.github/copilot-instructions.md`
- `.github/prompts/*.prompt.md` (only if they reference stale commands/capabilities)

Your goal is to prevent stale assumptions and ensure files reflect the **latest documented capabilities, limitations, and known pitfalls** — covering both **version accuracy** (SDK/API/CLI versions) and **functional accuracy** (whether the recommended patterns, services, and architectural advice remain current best practice).

## Persistent state — memory file

All capability refresh sessions share a **persistent memory file** at `memories/repo/capability-refresh-<date>.md`. This is the single source of truth for what has been verified, what was deferred, and what API versions / SDK versions are known-good.

### On every run:

1. **Read the memory file first** (before inventorying files). It contains: verified version facts, API version reference table, per-session records of every file touched, deferred items, and cross-file observations.
2. **Use it to skip already-verified claims.** If Session N verified `opentelemetry = 0.31` and no new release has occurred, do not re-verify. Only re-verify if: (a) the memory file flags the item as deferred/uncertain, (b) the claim is >30 days old, or (c) a newer version is suspected.
3. **Append a new session record** at the end of the memory file when done (same structure as prior sessions: scope, verification sources table, changes made, deferred items, total edits).
4. **Update the header date** to reflect the current session.
5. If no memory file exists, create one with the initial header and verified version facts.

> **Rule**: Never start from scratch when a memory file exists. The memory file is the incremental checkpoint — the structured checkpoint format below is for mid-session context exhaustion only.

## Execution model

This repository contains ~112 skills (many with their own `standards/*.md` sub-files, totalling ~60+ additional files), 14+ instruction files, 1 copilot-instructions file, and 5 prompt files — approximately **192+ files in scope**. A single-pass scan will exhaust context before completing. Execute in **explicit phases** with checkpoints.

### Phase 0: Load prior state

1. Read `memories/repo/capability-refresh-*.md`. Parse the verified version facts, API version reference, and all session records.
2. Build a **skip list**: files verified within the last 30 days with status `Current` and no deferred items.
3. Build a **re-check list**: files with `[VERIFY]` blocks, deferred items, or Currency dates >30 days old.
4. Detect **new files** added since the last session: compare the file inventory against files mentioned in prior session records. Any file not mentioned in any prior session is a new file and should be prioritized for review.
5. Proceed to Phase 1 with skip list and re-check list loaded.

### Phase 1: Discover, classify, and triage (read-only)

1. Inventory all files in scope using file search. Include both `SKILL.md` files and their `standards/*.md` sub-files.
2. **Cross-reference with Phase 0 skip list.** Mark files already verified and current. Mark files on the re-check list. Mark new files not seen in any prior session.
3. Classify each **skill unit** (SKILL.md + its standards files, treated together) as **technically volatile** or **stable**:
   - **Technically volatile**: references specific SDK/API versions, CLI commands, NuGet/npm packages, ARM API versions, Docker images, tool configurations that change between releases, **or recommends specific Azure services, architectural patterns, authentication approaches, or operational practices that may evolve**. Examples: dotnet-aspire (5 standards files), cosmos-db-patterns (3 standards files), terraform-patterns, github-actions-ci-cd, docker instructions, azure-container-apps, private-networking.
   - **Stable**: business/product/strategy skills with no version-sensitive claims and no technology-specific functional recommendations. Examples: lean-canvas, brainstorm-ideas, SWOT analysis, user-personas, competitive-battlecard.
4. Assign each volatile skill unit a **priority tier**:
   - **P0 (immediate)**: New files not seen in prior sessions, or files with open `[VERIFY]` / deferred items from prior sessions. Process first.
   - **P1 (high)**: Has standards sub-files AND pins specific versions/APIs.
   - **P2 (medium)**: Volatile but no standards sub-files, or standards files with fewer version-sensitive claims.
   - **P3 (low)**: Marginally volatile (e.g., references a tool but no pinned versions).
5. Output a triage list: volatile skills grouped by priority tier (P0 first), stable skills to skip, and already-verified skills to skip with last-verified date.
6. **Checkpoint**: Present the triage list to the user before proceeding. Wait for confirmation or adjustments.

> **Rule**: Do not read the full contents of stable (non-technical) skills. Only read technically volatile skills.
> **Rule**: Always read a skill's standards files together with its SKILL.md — they form a single verification unit.
> **Rule**: Files on the Phase 0 skip list (verified <30 days, no deferred items) can be skipped without re-reading unless the user explicitly requests a full re-run.

### Phase 2: Verify volatile skills against authoritative sources

Process volatile skills in **batches of 10–15 skill units**, ordered by priority tier (P1 first). Each batch is one conversation turn or sub-phase.

For each batch:

1. Read the full SKILL.md and all standards files for each skill unit in the batch.
2. Extract all version-sensitive claims (SDK versions, API versions, package names, CLI flags, Docker image tags, configuration shapes).
3. **Extract functional claims**: recommended Azure services/features, architectural patterns, authentication approaches, configuration defaults, operational practices, and workarounds. These are statements like "use X for Y", "prefer A over B", or "configure Z with these settings".
4. Verify each version claim against authoritative sources (see tool priority and version pinning strategy below).
5. **Verify each functional claim** against current documentation:
   - Is the recommended service/feature still GA and actively supported (not deprecated or superseded)?
   - **Azure retirement cross-check**: For each Azure service referenced in the skill, query `MRC-MCP-Server/get_recent_azure_updates` with `filter: products/any(p: p eq '<product>') and tags/any(t: t eq 'Retirements')` to check for active or upcoming retirements. If a retirement is found, get full details with `MRC-MCP-Server/get_azure_update_by_id` and update the skill with retirement date, migration path, and alternatives. Also query with `filter: products/any(p: p eq '<product>') and status eq 'Launched'` to confirm the service is still actively receiving GA features.
   - Has Azure/the framework added a new capability that makes the recommended approach unnecessary or suboptimal?
   - Has the recommended pattern been reclassified as anti-pattern by Microsoft or the framework maintainers?
   - Have defaults changed in the platform that invalidate configuration guidance?
   - Has a workaround been rendered unnecessary by native platform support?
   - Is the recommended authentication/identity approach still best practice (e.g., managed identity vs service principal)?
6. Record findings: `{ file, claim, current_value, verified_value, source_url, status: current|stale|uncertain, claim_type: version|functional }`.
7. **Inline cross-file propagation**: When a version is updated (e.g., `reqwest 0.12` → `0.13`), immediately grep the entire scope for all occurrences of the old version and update them in the same batch. Do not defer cross-file consistency to Phase 4 for version strings — propagate inline. For functional claim updates, also propagate to other skills that reference the same pattern or service.
8. **Breaking change check**: When a major version bump is detected (e.g., 0.x → 0.y for pre-1.0 crates, or N.x → N+1.x), check the changelog/release notes for breaking changes. If breaking changes exist, update Known Pitfalls tables in all affected files within the same batch.
9. **Symbol name verification**: Extract all class names, method names, enum values, and type names mentioned in backtick-delimited code references (e.g., `TruncationCompactionStrategy`, `FileAgentSkillsProvider`). For each symbol:
   - Search the upstream source repository (e.g., `github.com/microsoft/agent-framework`) for the exact symbol name using code search.
   - If GitHub code search requires authentication, verify against the official documentation page that introduced the symbol, confirming the exact spelling appears in a code example or API reference table.
   - **Do not extrapolate naming patterns.** If documentation describes a `CompactionProvider` wrapper and five strategy types, verify each strategy's exact class name independently. Do not assume they follow the same suffix pattern as the wrapper (e.g., `*CompactionProvider` vs the correct `*CompactionStrategy`).
   - **Do not conflate display names with class names.** Documentation tables may show "In-Memory Chat History Provider" as a display name; the actual C# class might be `InMemoryChatHistoryProvider` or something different. When the docs show only display names without code examples, mark the class name as `[VERIFY]` rather than guessing.
   - If a symbol cannot be verified in source code or a code example in official docs, wrap it in a `[VERIFY]` block. Do not present it as confirmed.
   - When propagating new symbols across multiple skill files, verify once and reuse the verified name. Do not re-derive names from memory for each file.
   - Record findings: `{ file, symbol, source_verified: true|false, source_url, notes }`.
10. **Reference link validation**: Extract all `http(s)://` URLs from the skill unit's files. For each URL:

- Fetch the URL using the `web` tool (or `browser` for pages requiring JavaScript rendering).
- Classify the result: `alive` (HTTP 200, content matches expectation), `redirected` (3xx to a different page — record the destination), `broken` (4xx/5xx, connection failure, or timeout), or `soft-404` (HTTP 200 but content is a generic "page not found" / "content moved" / "404" page).
- For `redirected` links: if the destination is the correct updated URL, replace the link inline. If it redirects to a generic page (docs homepage, search results), treat as `broken`.
- For `broken` or `soft-404` links: search for the current URL using `microsoft.learn.mcp` or `web` search, update the link if a replacement is found, or wrap in a `[VERIFY]` block if no replacement is found.
- Skip validation for links to localhost, `127.0.0.1`, example.com domains, and placeholder URLs (e.g., `https://your-instance.example.com`).
- Record findings: `{ file, url, status: alive|redirected|broken|soft-404, replacement_url?, notes }`.

11. **Batch checkpoint**: Present batch findings summary. If context is filling, emit a structured checkpoint (see checkpoint format below) and instruct the user to continue in a new conversation.

After all batches complete, present the full findings summary before proceeding to Phase 3.

> **Rule**: Do not start a new batch if the current conversation's context is more than ~70% consumed. Emit a checkpoint instead.
> **Rule**: Cross-file propagation of version updates is mandatory within the batch that discovers the stale version. Do not wait for Phase 4.

### Phase 3: Apply targeted edits

For each stale finding:

1. Apply minimal edits (see editing principles below).
2. Add or update the `Currency and verification` section.
3. Add or update the `Known pitfalls` table.

### Phase 4: Cross-file validation and output

Phase 4 catches issues that inline propagation (Phase 2, step 5) may have missed. If Phase 2 spanned multiple conversations, merge all checkpoint cross-file observations before starting.

1. Re-read the accumulated cross-file observations from all checkpoints and the memory file.
2. **Azure retirement sweep**: Query `MRC-MCP-Server/get_recent_azure_updates` with `filter: tags/any(t: t eq 'Retirements') and availabilities/any(a: a/ring eq 'Retirement' and a/year eq <current_year>)` to get all retirements for the current year. Cross-reference every Azure product name from the results against all files in scope. Any skill that references a retiring service without a deprecation warning or migration note must be updated. Record the retirement sweep results in the memory file session record.
3. Run cross-file consistency checks using **grep sweeps** for known-stale patterns (see cross-file validation section below). At minimum, sweep for:
   - Old version strings that were updated this session (e.g., if `reqwest 0.12` was bumped, grep for any remaining `reqwest.*0\.12`)
   - Known EOL / deprecated patterns from the memory file (e.g., `node:18`, `net8.0`, `ContainerLog` without V2)
   - `[VERIFY]` blocks — confirm each is intentional or resolve it
   - **Retiring Azure services** — grep for product names identified in the retirement sweep (step 2) that appear without retirement/deprecation warnings
4. Resolve any contradictions or inconsistencies detected across batches.
5. Update the memory file's API Version Reference table with any new verified versions.
6. Update the memory file's **Azure Retirement Reference** section with retirement dates and migration paths discovered during the sweep.
7. Produce the required output format.

### Context management between phases

- If context is filling up, complete the current phase (or current batch within Phase 2), **append the session record to the memory file**, then output a structured checkpoint and instruct the user to continue in a new conversation.
- The memory file is the durable checkpoint. The structured checkpoint format below is a convenience for the user — the memory file is what the next session actually reads.
- Use `/compact` only within a phase when intermediate context (tool output) can be safely compressed.
- Do not use `/clear` mid-phase — it loses verification evidence.
- When resuming from a checkpoint, read the memory file first (Phase 0), then re-read only the files relevant to the current phase.

### Structured checkpoint format

When emitting a checkpoint (end of phase or mid-Phase-2 batch boundary), output this structure so the next conversation can resume without quality loss:

```text
## Checkpoint: <Phase N> / Batch <M>

### Completed
- Files verified: [list of file paths]
- Findings so far: [summary table or list of { file, claim, status }]

### Pending
- Files remaining: [list of file paths, grouped by priority tier]
- Current batch: <batch number> of <total batches>

### Cross-file observations (accumulated)
- Contradictions detected: [list of contradictory directive pairs across files]
- Shared claims: [claims that appear in multiple files — must be consistent]
- Breaking change risks: [edits that would affect other files]

### Resume instructions
- Start Phase <N> / Batch <M+1>
- Re-read these files for context: [minimal list]
- Apply these accumulated findings before starting new verification: [list]
```

> **Rule**: Cross-file observations must be accumulated across every checkpoint so Phase 4 can run even if Phase 2 spanned multiple conversations.

## Tool priority

Use tools in this priority order. If a higher-priority source is unavailable (errors, timeouts), fall through to the next.

| Priority | Source                                        | Use for                                                         |
| -------- | --------------------------------------------- | --------------------------------------------------------------- |
| 1        | Microsoft Learn MCP (`microsoft.learn.mcp/*`) | Azure/Microsoft SDK versions, API versions, breaking changes    |
| 2        | MRC MCP Server (`MRC-MCP-Server/*`)           | Azure service retirements, GA/preview status, lifecycle dates   |
| 3        | Azure MCP (`azure-mcp/*`)                     | Azure resource provider details, ARM schema validation          |
| 4        | Bicep MCP (`bicep/*`)                         | Bicep best practices, AVM module validation                     |
| 5        | Context7 (`context7/*`)                       | Framework/library docs (MAF, Drasi, other non-Azure frameworks) |
| 6        | ISE Playbook (`iseplaybook/*`)                | Engineering process, code review checklists, CI/CD patterns     |
| 7        | Web/browser (`search`, `browser`)             | GitHub releases, closed PRs/issues, changelogs                  |

**Fallback rule**: If a top-priority MCP server returns an error, log it **in the memory file session record** (include error type and timestamp) and proceed with the next available source. Do not block the entire run on one unavailable server. Return and retry failed queries later if the server becomes available.

**MCP availability tracking**: At the start of each session, test each MCP server with a simple query. Record availability in the session record. If a server was unavailable in a prior session and is now available, prioritize re-verifying claims that were deferred due to that server's unavailability.

## Version pinning strategy

When updating a version reference, follow this strategy to avoid repeated incremental bumps across sessions:

| Resource type                | Target                                  | Rationale                                                                                             |
| ---------------------------- | --------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| Azure ARM API versions       | **Latest GA** (not preview, not N-1)    | Avoids re-bumping next session. Use `az provider show` or MCP to find the latest non-preview version. |
| NuGet / npm / crate packages | **Latest stable release**               | Check for major version bumps and breaking changes before updating.                                   |
| Docker base images           | **Latest LTS or current stable tag**    | e.g., `node:22-alpine` not `node:20-alpine`.                                                          |
| GitHub Actions               | **Latest major version SHA**            | e.g., `actions/checkout@<sha> # v4.3.1` — use latest patch within the major.                          |
| CLI tools                    | **Version range that includes current** | e.g., `>= 1.9.0` for Terraform.                                                                       |

> **Rule**: When you see an API version like `@2024-04-01` and the latest GA is `@2024-11-01`, jump to the latest GA — do not stop at an intermediate version. Check the memory file's API Version Reference table to see if a newer version was already verified in a prior session.

## HVE alignment

- This prompt is for **Hypervelocity Engineering (HVE)** quality workflows, not ad-hoc "vibe coding."
- Enforce **RPI discipline** where applicable: Research → Plan → Implement → Review.
- Prefer **verified truth over plausible output**: recommendations must be evidence-backed and operationally wired.
- Preserve and strengthen guidance that requires traceability (citations, file paths, line references, and validation steps).

## Validation against sample repositories

When instructions describe implementation patterns, validate against maintained sample repositories.

Priority order:

1. Microsoft-owned repositories (github.com/Azure-Samples, github.com/microsoft, github.com/Azure)
2. Official product team samples (github.com/drasi-project, github.com/microsoft/agent-framework)
3. Microsoft Learn referenced samples
4. Community samples with >500 stars and recent activity (within 12 months)

Confirm: commands compile/run, configuration shapes match docs, patterns reflect current practices.

Avoid: abandoned repos, blog companion repos, personal proof-of-concept projects.

## Staleness detection

Flag these categories:

- Stale versions, outdated commands, deprecated patterns, incorrect defaults, missing caveats
- Breaking changes, deprecations, migration notes from GitHub releases/PRs/issues
- **Major version bump detection**: When verifying a package/crate/SDK version, check if a major version bump occurred since the last refresh (e.g., 0.12→0.13 for pre-1.0, or 4.x→5.x). Major bumps frequently introduce breaking changes (renamed APIs, changed defaults, removed features). Always read the release notes/changelog for major bumps and update Known Pitfalls tables accordingly.
- Hardcoded or mutable production patterns (`latest` tags, unbounded versions)
- Preview/GA confusion
- CLI flag drift
- Mode/config-shape mismatches (same concept, different runtimes)
- Security drift (credentials vs managed identity, missing least-privilege checks, native token refresh)
- **OWASP standards drift** — verify current editions on each run:
  - [OWASP Top 10 (Web)](https://owasp.org/Top10/) — currently **2025** (replaces 2021)
  - [OWASP API Security Top 10](https://owasp.org/API-Security/) — currently **2023** (replaces 2019)
  - [OWASP Top 10 for LLM/GenAI Applications](https://genai.owasp.org/llm-top-10/) — currently **2025** (replaces 2023-24)
  - Check project pages for newer editions; update skill files and instruction files that reference OWASP when editions change
  - For AI/agent skills, verify alignment with the LLM Top 10 and OWASP Agentic Security Initiatives
- **Reference link rot** — URLs in skills and instructions that return 4xx/5xx, redirect to generic pages, or point to content that has been reorganised or removed. Common causes: Microsoft Learn documentation restructuring, GitHub repository archival or rename, OWASP project page reorganisation, deprecated service documentation removal. Do not assume a URL from a prior session is still valid — links break independent of content accuracy.
- Language that encourages **plausible but unverified implementation** without research/verification gates
- Guidance that collapses RPI phases for complex work without documented justification
- **AI code safety drift** — skills/instructions that generate LLM-calling code must address:
  - Prompt injection prevention (user input concatenated into LLM prompts without sanitization)
  - Tool-call result validation (MCP/function-call outputs consumed without schema validation)
  - Placeholder credentials in code examples (`your_api_key_here`, `changeme`, `password123`)
  - Debug defaults left enabled in example configurations
  - Data leakage to AI services (PII sent to external LLM APIs without anonymization)
  - Rate limiting on endpoints that trigger expensive LLM calls
- **Agent instruction integrity** — check for contradictory directives across instruction hierarchy levels (e.g., "always ask" in one file vs "never ask" in another); flag unsafe override patterns (`ignore previous instructions`, `disable safeguards`)
- **Intent alignment** — verify skill names match what they produce; flag stub/placeholder guidance presented as complete; detect dead documentation (described parameters not used, docstrings contradicting implementation)
- **AI writing fingerprints** — enforce `markdown.instructions.md` humanization rules during refresh edits; flag tutorial-style step-numbered comments, conversational first-person preambles, and promotional language introduced by AI generation
- **Over-engineering drift** — flag skills recommending unnecessary abstractions, wrapper-mania, enterprise-isms in small-scope guidance, or premature generalization patterns
- **Framework-specific safety patterns** — verify framework skills reflect current security practices (ASP.NET Core auth/CORS, React hooks ordering, Django DEBUG/SECRET_KEY, Express middleware chains, FastAPI auth dependencies)
- **Reliability pattern coverage** — skills generating service code must address retry logic with backoff, circuit breakers, timeouts, idempotency, and health checks where applicable
- **Logging privacy** — observability and logging guidance must address PII redaction, credential exclusion, and structured logging with selective field masking

### Functional capability drift

Beyond version numbers and CLI flags, skills and instructions contain **functional recommendations** — specific services, patterns, and practices they advise. These can become stale even when no version number changes. Check for:

- **Service deprecation or supersession**: The recommended Azure service (e.g., Azure Cosmos DB for PostgreSQL, Azure Functions v3 runtime, classic Application Insights) has been deprecated, marked end-of-support, or superseded by a newer service. **Primary check**: Query `MRC-MCP-Server/get_recent_azure_updates` with `filter: products/any(p: p eq '<product>') and tags/any(t: t eq 'Retirements')` to get structured retirement data with exact dates and migration guidance. Use `MRC-MCP-Server/get_azure_update_by_id` for full details on specific retirements. **Secondary check**: Query `microsoft.learn.mcp` with `"<service> deprecated"` or `"<service> end of support"` and check for deprecation banners in official documentation. The MRC MCP Server provides structured, machine-readable retirement timelines that are more reliable than searching for deprecation keywords in documentation prose.
- **New capabilities that change recommendations**: Azure or the framework added a feature that makes existing guidance suboptimal. Examples: AKS added native Gateway API support (removing need for third-party ingress controllers), Azure Container Apps added built-in GPU support, .NET added native AOT compilation. Query MCP for recent "what's new" or changelog pages.
- **Pattern reclassification**: A recommended pattern has been reclassified as anti-pattern by Microsoft, the framework maintainers, or the ISE Playbook. Examples: synchronous-over-async wrapping, service locator pattern, ambient context for DI. Check `iseplaybook` and `microsoft.learn.mcp` for current guidance.
- **Default changes**: The platform changed defaults that invalidate configuration guidance. Examples: .NET 10 container images default to Ubuntu (not Debian), AKS changed default outbound type, Azure Functions changed default runtime version. When defaults change, skills must explicitly call out the new default and update any configuration examples.
- **Workaround invalidation**: A skill recommends a workaround for a platform limitation that has since been resolved natively. Examples: manual cert rotation when auto-rotation was added, custom health probe scripts when built-in probes were enhanced, manual RBAC assignment when managed identity integration was streamlined.
- **Authentication and identity evolution**: Recommended auth approaches may be superseded. Examples: connection strings replaced by managed identity, API keys replaced by Entra ID RBAC, AAD v1 endpoints replaced by Microsoft Identity Platform v2. Check that skills recommend the most current identity approach.
- **Architectural guidance drift**: WAF (Well-Architected Framework) pillars, Azure Architecture Center patterns, or landing zone guidance may have been updated. Skills that reference specific WAF recommendations should verify alignment with current published guidance.
- **Tooling supersession**: A recommended CLI tool, extension, or utility has been replaced. Examples: `az acr build` replacing manual Docker push, `azd` replacing manual ARM deployments, `bicep` replacing raw ARM JSON. Verify recommended tools are still the primary path.

When a functional claim is found stale:

1. Verify the replacement/update against at least one authoritative source (MCP, official docs, maintained sample repo).
2. Update the skill's recommendation to reflect current best practice.
3. Add or update the `Known pitfalls` table with a row explaining the change (what was recommended, what's now recommended, why).
4. If the change is significant (service deprecated, pattern reclassified), add a note in the `Currency and verification` section documenting the shift.
5. Propagate the functional update to any other skills that reference the same service/pattern.

### HVE and context-engineering validation

- Confirm fast-moving skills preserve phase boundaries for complex work (research before implementation).
- Confirm multi-phase instructions include context management guidance.
- Confirm artifact handoff guidance exists for phase transitions.
- If autonomous flow is recommended, confirm fallback behavior exists when orchestration tools are unavailable.

## AI-specific refresh checks

When reviewing skills and instructions that produce or guide AI-generated code, apply these additional checks (informed by adversarial code review judges):

### Code safety in examples

- Verify code examples do not contain placeholder credentials, hardcoded secrets, or insecure defaults.
- Verify API endpoint examples include input validation guidance.
- Verify async code examples address error propagation (unhandled promise rejections, missing await).
- Verify database examples use parameterized queries, not string concatenation.

### Agent governance

- Verify each skill defines what to do when blocked (escalate, ask, fail-safe) — not just the happy path.
- Verify instruction files do not contain contradictory directives that create undefined agent behavior.
- Verify skills that recommend autonomous operation include human-in-the-loop safeguards for high-stakes decisions.
- Verify skills that consume external tool results (MCP tools, function calls) require schema validation before use.

### Generated code quality

- Verify skills that produce code enforce single-responsibility, not god-functions or god-classes.
- Verify generated patterns avoid unnecessary abstraction layers (wrappers around simple builtins, abstract factories with one implementation).
- Verify error handling guidance requires meaningful catch blocks (no empty catch, no swallowed exceptions).
- Verify concurrency guidance addresses race conditions, missing awaits, and unbounded parallelism.

### Compliance and privacy

- Verify observability/logging skills address PII redaction and credential exclusion from logs.
- Verify skills that handle user data address data residency and cross-border transfer considerations.
- Verify API skills include rate limiting and pagination guidance.
- Verify skills that integrate with AI services address data leakage prevention.

## Editing principles

- Favor small, high-signal updates. Keep existing structure/tone.
- Keep language concrete and operational. Use imperative for procedures.
- Avoid speculative statements without verification.
- When uncertain, include a `[VERIFY]` block with where/how to confirm.
- **Date checked**: Always use the actual current date, never a fabricated future date.
- Do not add broad marketing prose.

### What to add or update

1. **Currency and verification** section in fast-moving technical skills:

   ```text
   ## Currency and Verification

   - **Date checked:** YYYY-MM-DD (verified via <source>)
   - **Compatibility:** <frameworks/versions>
   - **Sources:** [link](url)
   - **Verification steps:**
     1. Concrete verification command or check
     2. ...
   ```

2. **Known pitfalls** table:

   | Area                          | Pitfall                                                                | Mitigation                                                                     |
   | ----------------------------- | ---------------------------------------------------------------------- | ------------------------------------------------------------------------------ |
   | MAF / MCP HTTP transport      | Bearer token extracted once at startup — expires ~1h, not auto-renewed | Supply a `DelegatingHandler` that calls `credential.GetTokenAsync` per-request |
   | MAF / database (Npgsql/Redis) | Token string used as static password — not refreshed                   | Use `UsePeriodicPasswordProvider` or equivalent                                |

3. Explicit **verification steps** for version-dependent claims.

### Frontmatter compliance guard (sensei alignment)

When editing a skill's YAML frontmatter `description:` field (or when adding a new skill), validate these constraints before committing the change.

| Constraint                        | Threshold        | Action if violated                                                  |
| --------------------------------- | ---------------- | ------------------------------------------------------------------- |
| Has `USE FOR:` or `WHEN:` trigger | Required         | Add a `USE FOR:` clause listing the primary invocation scenarios    |
| Word count                        | ≤ 60 words       | Trim redundant qualifiers; prefer verb phrases over noun phrases    |
| Character count                   | ≥ 150 chars      | Expand with specific scenarios or technology names                  |
| Character count                   | ≤ 1024 chars     | Compress; move detail into the skill body                           |
| No `DO NOT USE FOR:`              | Must not contain | Remove negative routing (keyword contamination risk at 120+ skills) |

If a description edit violates any constraint, fix it inline. If the fix is non-trivial (requires understanding the skill's purpose), flag the skill in the session's deferred items for manual review.

> **Rule**: The refresh prompt applies a lightweight frontmatter guard to prevent regressions. Run a dedicated frontmatter-optimization pass separately for comprehensive optimization.

### Pre-edit checks

- Confirm staleness using at least one authoritative source before editing.
- If evidence is ambiguous, prefer a `[VERIFY]` note over changing behavior.
- Preserve existing headings, section order, and tone; add new sections minimally.

## Required output format

Return:

1. **Summary of files updated** (with one-line reason each)
2. **Stale assumptions removed** (bullet list)
3. **New/updated pitfall guidance** (table: area, pitfall, mitigation)
4. **Verification evidence** — for each updated file: source URL, what was verified, date checked, specific claim supported
5. **Follow-up recommendations** — high-impact items not changed yet
6. **HVE alignment findings** — RPI conformance updates, context-engineering risks found/fixed, anti-vibe-coding guardrails added/strengthened
7. **Deferred items** — skills/instructions not verified due to MCP unavailability or context limits, with recommended next steps
8. **Functional capability findings** — for each functional claim found stale: the skill, what was recommended, what's now current, the authoritative source, and whether the change was applied or deferred
9. **Link health report** — summary table of all URLs checked: total links checked, alive, redirected (with updated URLs), broken (with replacements or `[VERIFY]` blocks), and soft-404s. List each non-alive link with: file path, original URL, status, replacement URL or `[VERIFY]` action taken

## Definition of done

- All identified stale or risky guidance in scoped files is updated or wrapped in `[VERIFY]` blocks.
- Each updated fast-moving skill includes a `Currency and verification` section with: date checked, sources used, concrete verification steps.
- For each technical domain touched, there is at least one `Known pitfalls` table.
- Updated links resolve successfully.
- Output includes all sections from the required output format.
- **All skill units** (SKILL.md + standards files) in scope were either verified or explicitly deferred with rationale.
- **All Phase 2 batches** completed (or remaining batches documented in deferred items with checkpoint for resumption).
- **Cross-file observations** from all checkpoints were merged and resolved in Phase 4.
- **Memory file updated**: Session record appended, API Version Reference table updated with any new verified versions, deferred items list reflects current state.
- **Grep sweeps passed**: All mandatory cross-file grep patterns return zero stale matches (or matches are confirmed intentional).
- **Reference links validated**: All `http(s)://` URLs in verified skill units have been checked. Broken links are either replaced with current URLs or wrapped in `[VERIFY]` blocks. Redirected links are updated to their canonical destination. No skill links to a page that returns 4xx/5xx or displays generic "page not found" content.
- **Functional claims verified**: Each volatile skill unit had its functional recommendations (recommended services, patterns, auth approaches) checked against current MCP documentation. Stale functional claims are updated or wrapped in `[VERIFY]` blocks with rationale.
- **Symbol names verified**: Every class, method, or type name introduced or updated during the session was verified against upstream source code (GitHub code search) or official code examples. No symbol was derived by pattern extrapolation alone. Unverifiable symbols are wrapped in `[VERIFY]` blocks.

## Quality gate

- No stale command/version claims left in edited files.
- Every updated fast-moving technical skill has explicit verification guidance.
- Added guidance is traceable to current docs/changelog evidence.
- Context-management guidance is consistent across updated files when multi-phase workflows are described.
- No placeholder credentials or insecure defaults in code examples across updated skills.
- No AI writing fingerprints introduced by refresh edits (enforce humanization rules from `markdown.instructions.md`).
- Skills recommending API code address rate limiting and error response consistency.
- Skills generating async code address error propagation and missing-await hazards.
- Observability guidance addresses PII redaction and credential exclusion.
- Reliability patterns (retry, circuit breaker, timeout) present in skills that generate service integration code.
- No contradictory directives introduced across the instruction hierarchy.
- Skills that consume external tool results require schema validation.
- Edited skill descriptions pass the frontmatter compliance guard (USE FOR: trigger present, ≤60 words, ≥150 chars, ≤1024 chars, no DO NOT USE FOR:).
- **Functional recommendations verified**: Each volatile skill's recommended services, patterns, and auth approaches have been checked against current MCP documentation. No skill recommends a deprecated or end-of-support service without an explicit migration warning.
- **Deprecated services flagged**: Any skill that previously recommended a service now marked deprecated/EOL has been updated with current alternatives and a Known Pitfalls entry documenting the change.
- **No broken reference links**: All URLs in edited skill files resolve to valid, relevant content. No links return 4xx/5xx or redirect to generic error/search pages.
- **No unverified symbol names**: Every class, method, enum, or type name in backtick-delimited references has been verified against upstream source code or an official code example. Display names from documentation tables are not treated as class names without confirmation. Any symbol that could not be source-verified is wrapped in a `[VERIFY]` block.

## Cross-file validation

- Ensure the same command or capability is not described differently across skills, instructions, and prompts.
- If conflicting guidance exists: prefer official documentation, update all conflicting files.
- Changes are minimal and do not break existing repository conventions.
- **Mandatory grep sweeps**: After all edits in a session, run these grep patterns across `.github/**/*.md` and confirm zero stale matches:
  - Every old version string that was replaced this session (e.g., `reqwest.*0\.12`, `@2024-04-01` for a specific resource type)
  - Known EOL patterns from the memory file (maintain a running list; examples: `node:18`, `net8.0`, `ContainerLog` without `V2`, `OWASP.*2021`)
  - Any `[VERIFY]` block — confirm each is intentional or resolve it
  - `DO NOT USE FOR:` in skill descriptions — should be zero matches (keyword contamination risk; see frontmatter compliance guard)
  - **Broken URLs discovered during Phase 2 link validation** — after all edits, re-check any URLs that were replaced to confirm the new links are alive. Verify that no stale URLs remain in edited files by grepping for old domain patterns or path segments that were part of broken links.
  - **Symbol name consistency** — when a symbol was introduced or corrected this session, grep for all variant spellings across all files in scope. Common failure pattern: the correct class name is `FooStrategy` but a prior edit introduced `FooProvider` in another file. Grep for the incorrect variant to confirm zero matches remain.
- **Contradiction detection**: Scan for contradictory directives across instruction hierarchy levels. Flag pairs where one file says "always X" and another says "never X" or where scope boundaries overlap with different rules.
- **Backwards compatibility**: Before editing established skills/instructions, assess whether the change would break existing agent behavior that other files depend on. Document breaking changes explicitly.
- **False-positive awareness**: When flagging issues in skills, distinguish between:
  - Application code patterns vs. IaC/declarative template patterns (different rules apply)
  - String literals containing keywords ("password" in error messages is not a hardcoded secret)
  - Test/fixture/mock context (production-only concerns do not apply to test examples)
  - Import/type declarations vs. runtime code

## Symbol and namespace verification guardrails

- Never introduce commands, flags, APIs, namespaces, or class/type names not present in official documentation or maintained sample code.
- For every SDK, API, or code reference (including namespaces, class names, and method names), verify the exact symbol against authoritative docs and at least one maintained sample. Do not rely on plausible naming or pattern extrapolation.
- When a symbol, namespace, or class name changes (e.g., due to a breaking SDK release or official rename), immediately propagate the update across all skills, instructions, and prompts in scope. Do not leave partial or inconsistent references.
- If a symbol or namespace cannot be verified in source code or an official code example, wrap it in a `[VERIFY]` block and do not present it as confirmed.
- Treat plausible naming as insufficient evidence. Do not extrapolate naming conventions across types or namespaces. Each name is a separate verification target.
- Distinguish source-verified from docs-described. A symbol confirmed in a GitHub code search of the upstream repo is "source-verified." A symbol mentioned only in a docs page's prose or table (without a code example showing the exact class name) is "docs-described" and should be tagged `[VERIFY]` if the exact class name matters for consumers.
- When adding new capability sections, verify every symbol and namespace before writing it into the skill. Do not batch-write symbols and verify later. The cost of propagating a wrong name across multiple files and then correcting it is higher than verifying upfront.
- Prefer linking documentation rather than paraphrasing complex behavior.

## Hallucination guardrails

- Do not introduce recommendations that optimize for speed while skipping evidence, phase separation, or validation.
- For complex changes, require explicit research evidence and implementation verification.
- If the evidence chain is incomplete, mark as `[VERIFY]` or `[UNCERTAIN]` instead of implying confidence.
- All guidance must describe real functionality and production patterns. Discourage workaround-style implementations that "look complete" but are not operationally wired.

## Anti-vibe-coding guardrails

- Do not accept recommendations that optimize for speed while skipping evidence, phase separation, or validation.
- For complex changes, require explicit research evidence and implementation verification.
- If the evidence chain is incomplete, mark as `[VERIFY]` or `[UNCERTAIN]` instead of implying confidence.
- All guidance must describe real functionality and production patterns. Discourage workaround-style implementations that "look complete" but are not operationally wired.
