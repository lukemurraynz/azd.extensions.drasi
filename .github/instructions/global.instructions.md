---
applyTo: "**/*"
description: Archetype-agnostic global development standards and Copilot guidance for all repositories
---

# Global Development Standards (Archetype‑Agnostic)

These are organization-wide rules that apply to every repository and project, regardless of language, framework, or archetype. Language- or tool-specific files may add details and can override sections explicitly for their scope. Copilot/Agents are configured via `.github/copilot-instructions.md`, which describes **how** they should apply these standards when working in this repository.

## Scope and Precedence

- Applies to all files in the repository unless a more specific instruction file overrides it (for example, Python/Java/Terraform instruction files).
- When guidance conflicts, prefer the most specific rule applicable to the file or technology in question.
- Keep changes minimal and focused; avoid broad refactors unless requested or strictly necessary.
- Operational behaviors for Copilot/Agents live in `.github/copilot-instructions.md`; this file defines engineering standards.

## Core Principles

- Small, focused changes with clear intent and atomic commits
- Clean, readable, maintainable code with meaningful names
- Single Responsibility per function/module where practical
- DRY: prefer reuse over duplication; extract helpers when duplication appears 3+ times
- Deterministic, reproducible builds and tests

## Definition of Done (DoD)

Before you consider a task complete, ensure:

1. Functionality implemented with clear inputs/outputs and errors handled
2. **Automated tests added or updated AND VERIFIED TO COMPILE AND RUN**
   - Unit tests for pure logic; integration tests for I/O
   - Target at least a basic happy path and one edge/error case
   - **Test compilation validated**: Language-specific build commands must exit successfully:
     - .NET: `dotnet build <TestProject>.csproj` exits with code 0
     - TypeScript/JavaScript: `tsc --noEmit` or test runner compiles successfully
     - Python: `pytest --collect-only` or `mypy` on test files
   - **Tests execute successfully**: Run test suite and verify execution (pass/fail, not compilation errors)
   - Documentation-only changes: tests are optional; still ensure examples/commands match current behavior
3. Documentation updated (README, comments/docstrings, ADRs if design decisions changed)
4. Security and privacy concerns reviewed (no secrets; validated inputs; least-privileged access)
5. Lint/format/type checks pass locally and in CI
6. Commit message concise and descriptive; human-authored PRs include context, acceptance criteria, and risk notes

## Testing Standards

- Write tests alongside code changes; prefer fast, deterministic tests
- Organize tests by feature or module; keep fixtures small and explicit
- Mock external calls and system boundaries; avoid network/filesystem unless explicitly needed
- Include at least:
  - Happy path
  - One edge/boundary case
  - One failure/exception path
- Strive for incremental coverage improvement; do not reduce existing coverage without justification
- For TDD expectations, follow the workflow defined in `.github/copilot-instructions.md`

## Static Analysis and Quality Gates

- Treat **critical/high** findings as release blockers unless explicitly triaged as false positives with evidence.
- Keep suppression metadata durable and reviewable: each suppression must include rule id, scope, rationale, owner, and expiration/review date.
- Do not use blanket suppression for entire directories unless files are generated artifacts.
- Generated files (for example ORM migration designer snapshots) must be clearly separated and excluded from complexity/maintainability gates where those gates are not actionable.
- For non-generated production code, enforce maintainability budgets in CI (for example: cyclomatic complexity, function length, duplicate code, and lint severity thresholds).

## Documentation Standards

- Update README or module-level docs when behavior, interfaces, flags, or environment variables change
- Use docstrings/comments to explain "why" when the "what" is not obvious
- Maintain ADRs for significant design or dependency choices
- For new services or major components, explicitly document architecture classification (Data Plane, Control Plane, or Hybrid) in an ADR/README/startup diagnostics output. If classification changes, document it as an architectural decision.
- For HTTP APIs, align contract decisions (versioning, pagination, LROs, idempotency, and error shapes) with `microsoft/api-guidelines` (vNext) appropriate to the plane (Azure data plane vs Graph). For Azure management plane (ARM resource providers), follow the Azure Resource Provider Contract. Document intentional deviations as ADRs.
- Keep examples minimal and runnable

## Security and Compliance

- Never commit secrets, keys, tokens, or credentials; use secret managers and environment injection
- Validate and sanitize all inputs; use parameterized queries for data access
- Prefer least privilege for identities, roles, and API tokens
- Prefer managed identities over hardcoded credentials for service-to-service authentication
- **For Infrastructure as Code**: Avoid post-deployment manual configuration steps when resource properties support the same functionality in Bicep/ARM templates. Post-deployment scripts should only handle data plane operations (e.g., data initialization), not control plane configuration.
- Keep dependencies updated; run vulnerability scans and remediate critical/high findings
- Avoid insecure primitives: raw string SQL, eval/exec, shell with untrusted input, insecure random for security decisions
- Protect PII and sensitive data; log only what’s necessary and sanitize outputs
- Never place secrets or credentials in URL paths/query strings; pass secrets in headers or request bodies only.
- Avoid empty error handlers; catch blocks must either handle meaningfully, log safely with context, or rethrow.
- Before uploading code or content to third-party tools (diagram renderers, pastebins, gists, online formatters), consider whether it could contain sensitive information. Published content may be cached or indexed even if later deleted.
- **Treat client-distributed code as public**: any code shipped to consumers (npm, NuGet, PyPI packages, Docker images, browser bundles) must be assumed fully readable. Do not rely on obfuscation or minification as security controls; LLMs can trivially reverse them.
- **AI system prompts and tool definitions must not ship in client packages or Docker images** unless explicitly classified as Public. Confidential prompts must be loaded at runtime from Azure App Configuration, Key Vault, or equivalent server-side stores. See `distribution-security.instructions.md` for the full framework.
- **Audit published artifacts before release**: run `npm pack --dry-run`, `dotnet pack` inspection, or `docker history` / `dive` to verify no secrets, source maps, PDBs, prompt files, or server-side code leak into distribution artifacts.

## Dependencies and Build

- Pin or bracket versions prudently to ensure reproducibility
- Prefer widely-used, actively maintained libraries
- Remove unused dependencies; avoid transitive bloat
- Keep builds cacheable; fail fast on lint/type errors
- Maintain supply-chain hygiene: prefer lockfiles and/or provenance/signing where supported

## Version Control and PR Hygiene

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```text
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`, `ci`

**Examples:**

- `feat(api): add user authentication endpoint`
- `fix(auth): resolve token expiration issue`
- `docs(readme): update installation instructions`

### Commit Best Practices

- Write atomic commits with imperative subject lines under ~72 chars
- Keep commits focused; reference issue IDs where applicable (e.g., `fixes #123`)
- Use branches for work; keep PRs small and focused with clear checklists

### Code Review Guidelines

- Be respectful and constructive in code reviews
- Provide specific, actionable feedback; focus on improving the code, not the person
- Ask questions when requirements are unclear; explain the "why" behind suggestions
- In reviews, be kind, specific, and actionable; request changes with concrete suggestions

### Pull Request (PR) Standards

- Create clear, focused PRs with descriptive titles
- Include acceptance criteria and testing notes
- Keep PRs small (< 400 lines when possible)
- Provide context for reviewers in the description
- Link related issues

## Observability and Operations

- Use structured logging; include request correlation IDs at boundaries
- Avoid logging secrets and sensitive payloads
- Fail with actionable errors; include context for troubleshooting
- Consider basic performance characteristics and document any non-obvious tradeoffs
- Reliability posture (ADAC): auto-detect runtime config/dependency state, auto-declare the resolved mode (logs/state), and auto-communicate degradation (health/UX/metrics). No silent fallbacks for required services. Include ADAC declarations in PR descriptions or ADRs when reliability posture changes.
- If LLM/agent decisions influence orchestration or domain outcomes, record inputs, prompt/config version, and tool invocation results so decisions are replayable and auditable.

## Accessibility and Inclusion

- Write inclusive language in code, docs, and UI strings
- Ensure docs and examples are accessible (headings, links, contrast in images)

## Working with Copilot/Agents

- Leverage repository prompts and agents when available for planning, security review, and test generation.
- For detailed Copilot/Agents behavior (instruction hierarchy, MCP usage, task planning), see `.github/copilot-instructions.md`.
- Reference prompt and agent files using the full URLs above when embedding links in Markdown.
- Prefer minimal, targeted suggestions; validate generated code (human- or tool-authored) with tests and linters.
- Use minimal context: read only files needed to complete the task, and reuse existing patterns before creating new ones.
- Verify exact API/class/method/command names against authoritative docs or maintained samples before presenting them as valid.
- If exact symbol verification is not possible, fail closed: do not assert the name, add a `[VERIFY]` block with confirmation steps.
- If verification fails, mark the response as `[UNCERTAIN]` and defer to human review.
- **Do NOT create markdown files to document changes**: Make changes directly and provide brief text summaries. Only create documentation files (README, ADR, runbooks) when explicitly requested or updating existing docs.

## Configuration Management

- **Externalise all environment-specific configuration**: no environment names, connection strings, feature flags, or service URLs hard-coded in application code; use environment variables, app configuration services, or mounted config files
- **Provide safe defaults for local development**: default values must not connect to production systems; wrong defaults should fail loudly, not silently degrade
- **Validate configuration at startup**: mandatory keys must be checked before the application accepts traffic; emit clear error messages identifying exactly which keys are missing
- **Backwards-compatible config changes**: treat config key renames the same as public API renames — support both old and new keys through a transition period; log a deprecation warning when the old key is used
- **No secrets in committed configuration files**: `.env` files, `appsettings.json`, and Helm values files must never contain real credentials; use placeholder values and document how to obtain real values via the project README
- **Document every configuration key**: maintain a reference table (README or wiki) listing each key, its type, default, required/optional status, and which environments need it overridden
- **One artifact, many environments**: build once, deploy everywhere; configuration differences belong in environment-specific injection (environment variables, Key Vault references, Config Map), not in separate build artefacts
- **Prefer typed/structured configuration**: use strongly typed config objects (e.g., `IOptions<T>` in .NET, `pydantic.BaseSettings` in Python, `zod` in TypeScript) rather than raw string lookups; parse and validate at the boundary, not at the point of use

---

By following these global standards, teams maintain consistency across repositories while still allowing language- and framework-specific instruction files to tailor the details where appropriate.
