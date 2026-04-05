# GitHub Copilot Coding Agent Instructions

This file governs how GitHub Copilot behaves when it autonomously picks up issues and works on them without direct human supervision.

## Capability Self-Check

Before starting work on an issue, rate your fit:

| Fit                        | When                                                                                                        | Action                                                            |
| -------------------------- | ----------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| 🟢 **Proceed**             | Issue is clear, scoped, and within your skillset                                                            | Start work following the workflow below                           |
| 🟡 **Proceed with caveat** | Issue is mostly clear but has ambiguous edges                                                               | Start on the clear parts; surface ambiguity in the PR description |
| 🔴 **Decline**             | Issue requires domain expertise you do not have, is underspecified, or conflicts with existing architecture | Comment on the issue explaining what is unclear or out of scope   |

Treat 🔴 issues as blocked — comment and wait for clarification rather than guessing.

## Branch Naming

```
feature/{issue-number}-{short-slug}
fix/{issue-number}-{short-slug}
chore/{issue-number}-{short-slug}
```

**Examples:**

- `feature/42-add-cosmos-skill`
- `fix/17-broken-devcontainer-path`
- `chore/88-update-bicep-api-versions`

Use `feature/` for new capabilities, `fix/` for bugs, `chore/` for maintenance or updates with no behaviour change.

## Specialist Agent Routing

Engage the appropriate agent before or during implementation:

| Issue type                                          | Agent to engage                 |
| --------------------------------------------------- | ------------------------------- |
| Azure infrastructure design or service selection    | `azure-architect`               |
| New skill creation or skill compliance review       | `creating-agent-skill-agent`    |
| DevContainer setup or tooling configuration         | `creating-devcontainers-agent`  |
| Architecture diagrams needed                        | `diagram-smith`                 |
| Documentation updates (README, ADRs)                | `documentation-specialist`      |
| Security review of IaC or code changes              | `security-specialist`           |
| Tests not compiling or test/implementation mismatch | `test-validation-specialist`    |
| Root cause unclear or system behaving unexpectedly  | `troubleshooting-specialist`    |
| Backlog item needs acceptance criteria              | `backlog-refinement-specialist` |

For issues spanning multiple domains, engage `team-orchestrator` to coordinate parallel agent work.

## Workflow

### 1. Understand before starting

- Read the full issue description, comments, and any linked issues
- Check `memories/` for relevant prior decisions before doing any design work
- Identify which files and agent skills are relevant — read `.github/agents/` to understand available specialists

### 2. Plan before coding

For issues that are non-trivial (more than a single-file edit):

1. Write a brief plan as a comment on the issue (or in the PR description if creating the branch immediately)
2. List the files you expect to change
3. Identify any decisions that need to be made and state your intended choice

### 2.1 Delegation Prompt Contract (Mandatory)

When delegating work to a specialist agent, provide a complete prompt contract.

Required sections (all required):

1. **TASK**: One atomic objective
2. **EXPECTED OUTCOME**: Concrete deliverable
3. **REQUIRED SKILLS**: Explicit skill list (or `[]`)
4. **REQUIRED TOOLS**: Must-use tools and constraints
5. **MUST DO**: Hard requirements
6. **MUST NOT DO**: Prohibited actions
7. **CONTEXT**: Relevant files, patterns, references, and boundaries

If any section is missing, stop and complete the prompt before delegation.

Additional hard requirements:

- Delegation prompts must be self-contained (no hidden dependency on prior conversation context).
- Do not use lazy phrasing such as "based on your findings". Summarize findings into a concrete spec before delegation.
- After delegation, report launch status only. Do not infer or predict results before worker completion.

### 3. Implementation

- Follow all standards in `.github/instructions/global.instructions.md`
- Follow language-specific standards in `.github/instructions/coding-standards/`
- Engage specialist agents (see routing table above) rather than improvising domain work
- Never simulate, stub, or fake functionality — surface the real error and fix the root cause

### 4. Verification

Before opening a PR:

- Run any existing linters, builds, or tests — do not add new tooling without approval
- Engage `test-validation-specialist` if tests were added or modified
- Engage `security-specialist` if the change touches authentication, secrets, IaC, or public-facing endpoints

### 4.1 User-Facing Verification Gate (Mandatory)

If a change affects user-facing behavior, do not mark work complete until behavior is executed and verified:

- **Frontend/UI**: Run the app path and verify interactions, console health, and responsive behavior.
- **Backend/API**: Execute happy-path and malformed-input calls; verify status and response shape.
- **CLI/Script**: Run representative commands including `--help` and at least one invalid input case.
- **Config/Runtime**: Start the service (or equivalent dry run) and confirm updated configuration is applied.

### 4.2 Configuration/Hook Verification Gate

For settings, hook, or automation-policy changes:

- Validate syntax/schema first (JSON/YAML/schema checks).
- Verify runtime activation by triggering at least one relevant event path when feasible.
- If runtime activation cannot be verified in-session, explicitly document what was validated, what is still unverified, and the exact manual verification step.

### 5. Pull Request

**Title format:** `{type}({scope}): {short description}` — e.g., `feat(skills): add ab-test-analysis skill`

**PR body must include:**

```markdown
## Summary

<What this PR does and why>

## Changes

- <file or area>: <what changed>

## Testing

<How the change was verified>

## Decisions made

<Any design or implementation choices made autonomously — see below>

## Related

Closes #<issue number>
```

**Labels to apply:**

| Condition                          | Label           |
| ---------------------------------- | --------------- |
| New agent or skill                 | `enhancement`   |
| Bug fix                            | `bug`           |
| Documentation only                 | `documentation` |
| Needs human review before merge    | `needs-review`  |
| Contains an architectural decision | `architecture`  |

## Decision Capture

When you make a design or implementation decision autonomously during issue work, record it so it persists across sessions and is visible to reviewers:

1. **In the PR description** — list decisions under "Decisions made" in the PR body
2. **In `memories/`** — write a brief decision record for anything architectural:

```markdown
# Decision: <short title>

Date: <YYYY-MM-DD>
Issue: #<number>
Context: <what problem prompted this decision>
Decision: <what was decided and why>
Alternatives considered: <what else was evaluated>
Trade-offs: <downsides accepted>
```

## Guardrails

- ✅ **Always:** Read `memories/` before designing anything — prior decisions may already answer your question
- ✅ **Always:** Engage a specialist agent rather than doing domain work yourself
- ✅ **Always:** Capture decisions in the PR body and in `memories/` when they are architectural
- ✅ **Always:** Surface ambiguity in PR descriptions rather than silently resolving it
- ⚠️ **Ask first:** Before refactoring files outside the direct scope of the issue
- ⚠️ **Ask first:** Before adding a new dependency, tool, or file pattern not already present in the repo
- 🚫 **Never:** Merge without human approval when the `needs-review` label is applied
- 🚫 **Never:** Simulate, stub, or fake functionality — fix the real problem
- 🚫 **Never:** Commit secrets, credentials, or hardcoded endpoints
- 🚫 **Never:** Suppress linting warnings or test failures without documenting why
