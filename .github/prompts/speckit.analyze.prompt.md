---
agent: speckit.analyze
---

# SpecKit analyze prompt

Perform a non-destructive cross-artifact analysis of `spec.md`, `plan.md`, and `tasks.md`.

## Required input loading

Load these files before analysis:

- `specs/<feature>/spec.md`
- `specs/<feature>/plan.md`
- `specs/<feature>/tasks.md`
- `.specify/memory/constitution.md` (if present)

## Mandatory analysis checks

1. Constitution coverage
   - Every constitution-level `MUST` or non-negotiable rule in scope must be represented by at least one requirement in `spec.md` and at least one task in `tasks.md`.
   - Flag missing mappings as `CRITICAL`.

2. Requirement to task coverage
   - Every FR and SC in `spec.md` must map to one or more task IDs.
   - Mark partial coverage when a task exists but does not fully enforce the behavior.

3. Edge case closure
   - Each edge case in `spec.md` must map to at least one test task and one implementation task.
   - If not, mark `HIGH`.

4. Command-surface consistency
   - Validate that command contracts are consistent across artifacts.
   - Check global flag parity (`--environment`, `--output json`) across all commands, including stubs.
   - Check default behavior versus alias flags (for example streaming defaults versus `--follow`).

5. Error-model consistency
   - Ensure error codes referenced in `spec.md` exist in `tasks.md` and are covered by tests.
   - If an error code is required by spec but absent in implementation planning, mark `HIGH`.

6. Runtime-versus-tooling observability
   - Distinguish runtime workload telemetry requirements from extension or host process telemetry.
   - Flag plans that only instrument the CLI process while runtime telemetry is required.

7. Version governance
   - Detect duplicated hardcoded version floors across files.
   - Recommend a single source of truth with references.

8. Plan and task structure coherence
   - Ensure phase numbering and dependency narratives are internally consistent.
   - Validate task summary totals against the actual task list.

## Required output format

Provide a compact report with these sections:

1. `Issue table` with columns: `ID`, `Category`, `Severity`, `Location(s)`, `Summary`, `Recommendation`
2. `Coverage summary` mapping each FR and SC to task IDs
3. `Constitution alignment` list
4. `Unmapped tasks` list
5. `Metrics` (`requirements considered`, `coverage %`, `critical count`)
6. `Recommended next actions` in priority order

## Severity rubric

- `CRITICAL`: violates constitution non-negotiable or release-blocking requirement
- `HIGH`: missing behavior that can cause incorrect implementation
- `MEDIUM`: ambiguity, inconsistency, or likely drift risk
- `LOW`: duplication, readability, or maintainability improvement

## Constraints

- Do not edit files.
- Do not invent missing artifacts.
- Use exact file paths and line references when available.
