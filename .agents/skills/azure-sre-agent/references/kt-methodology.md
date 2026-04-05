# Kepner-Tregoe (KT) Operations Overlay

Use this overlay so incident handling and decisions follow a consistent
production-grade method across teams and services.

## When KT Is Required

Use KT depth by severity and risk:

1. **P1/P2 incidents**: full KT flow (`SA -> PA -> DA -> PPA`) required.
2. **P3/P4 incidents**: lightweight KT (`SA` + targeted `PA` or `DA`) is sufficient.
3. **Any write action in production**: at minimum include `DA` and `PPA`
   evidence before execution.

## KT Flow Mapping for Azure SRE Agent

### SA — Situation Appraisal

Purpose: classify concerns, set priorities, choose next analysis.

Use in agent workflow:

1. List threats/opportunities and deviations.
2. Separate and clarify concerns.
3. Prioritize by current impact, future impact, and time frame.
4. Choose next step (`PA`, `DA`, or `PPA`) and owners.

### PA — Problem Analysis

Purpose: find true cause of a deviation using `IS / IS NOT` logic.

Use in agent workflow:

1. State deviation clearly.
2. Build `IS / IS NOT` by `WHAT`, `WHERE`, `WHEN`, `EXTENT`.
3. Generate possible causes from distinctions and changes.
4. Test causes against evidence and confirm the most probable cause.

### DA — Decision Analysis

Purpose: choose the best option using explicit objectives.

Use in agent workflow:

1. State decision to be made.
2. Define objectives and classify into `MUST` and `WANT`.
3. Weight `WANT` objectives.
4. Score alternatives, eliminate `No Go` options that fail `MUST`s.
5. Select recommended option with rationale.

### PPA — Potential Problem Analysis

Purpose: protect execution of the chosen action plan.

Use in agent workflow:

1. List potential problems for planned actions.
2. Identify likely causes.
3. Define preventive actions.
4. Define contingent actions and triggers.

## Output Contract (KT-Aligned)

For KT-required incidents, output must include:

1. `Situation Appraisal`
2. `Problem Analysis`
3. `Decision Analysis`
4. `Potential Problem Analysis`
5. `Recommended Action`
6. `Risk and Rollback`
7. `Owners and Timeboxed Next Steps`

## AKS / Container Apps / Drasi Guidance

### AKS

1. SA: split cluster, workload, and dependency concerns.
2. PA: use `IS / IS NOT` across nodes, namespaces, regions, and time windows.
3. DA: compare alternatives (restart, scale, rollback, config fix) with must/want scoring.
4. PPA: define failure triggers and safe rollback before action.

### Container Apps

1. SA: classify impact by service and revision.
2. PA: compare current revision vs known-good revision.
3. DA: decide between rollback, traffic shift, config patch, or wait-and-observe.
4. PPA: protect rollout with threshold triggers and abort plan.

### Drasi on AKS

1. SA: separate ingestion lag, query staleness, and platform instability concerns.
2. PA: isolate whether issue is Drasi runtime, AKS platform, or external dependency.
3. DA: pick safest action (scale/rollback/config correction/restart) using must/want criteria.
4. PPA: define triggers for fallback and contingency paths.

## Governance

For P1/P2 or production write actions:

1. Use a `Stop` hook policy that rejects outputs missing required KT sections.
2. Require explicit rejection reason with remediation guidance.
3. Allow completion only when KT sections are present and coherent.

See [kt-templates.md](./kt-templates.md) for copy-ready worksheets and
[hooks-governance.md](./hooks-governance.md) for enforcement patterns.

## Source

- Derived from user-provided Kepner-Tregoe workbook templates (`SA`, `PA`, `DA`, `PPA` worksheets and question banks).
