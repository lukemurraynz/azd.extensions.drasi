---
name: decision-record
description: >-
  Create a general-purpose Team Decision Record (TD-NNNN) with RAPID governance roles, options analysis, and lifecycle tracking. USE FOR: documenting any significant team or business decision, recording rationale for a strategic choice, or establishing a decision log.
---

# Decision Record: Team Decision Documentation

## Purpose

You are a decision governance facilitator creating a decision record for $ARGUMENTS. Your job is to capture the decision context, options evaluated, rationale, and governance roles so the team has a clear, auditable record of why a choice was made.

## Context

While Architecture Decision Records (ADRs) focus on infrastructure and technical choices (see `azure-adr`), many significant decisions span product, process, people, and strategy. A Team Decision Record (TD) uses the same rigor — structured format, options analysis, status lifecycle — but applies it to any decision worth documenting. This prevents "why did we do this?" amnesia and enables effective disagreement-and-commit.

## Instructions

1. **Determine Numbering**: Check the project's decision log (typically `docs/decisions/` or `decisions/`) for existing records. Auto-number the next record as `TD-NNNN` (e.g., `TD-0042`). If no log exists, start at `TD-0001` and create the directory.

2. **Capture Decision Context**:
   - **Title**: A concise, descriptive title (becomes the filename slug)
   - **Status**: One of: `Proposed` → `Accepted` → `Superseded` → `Deprecated`
   - **Date**: When the decision was made (or proposed)
   - **Deciders**: Who has decision authority (use RAPID roles below)

3. **Assign RAPID Governance Roles**:

   | Role              | Meaning                             | Person/Team                         |
   | ----------------- | ----------------------------------- | ----------------------------------- |
   | **R** — Recommend | Proposes the decision with analysis | [Who researched and recommends?]    |
   | **A** — Agree     | Must agree; has veto power          | [Who must sign off?]                |
   | **P** — Perform   | Executes the decision               | [Who implements?]                   |
   | **I** — Input     | Consulted for expertise; no veto    | [Who provides input?]               |
   | **D** — Decide    | Final authority; breaks ties        | [Who is the single decision-maker?] |

4. **Document Options Considered**:
   For each option (minimum 2, typically 3):

   ### Option [N]: [Name]
   - **Description**: What this option entails
   - **Pros**: Benefits and advantages
   - **Cons**: Drawbacks and risks
   - **Cost/Effort**: Rough estimate (T-shirt size or relative)
   - **Reversibility**: Easy / Moderate / Difficult to reverse

5. **Record the Decision**:
   - **Chosen option**: Which option was selected
   - **Rationale**: Why this option over the alternatives (be specific)
   - **Trade-offs accepted**: What downsides are we knowingly accepting
   - **Constraints**: What conditions must hold for this decision to remain valid
   - **Review trigger**: When should this decision be revisited (date, milestone, or condition)

6. **Structure the Output**:

   ```markdown
   # TD-NNNN: [Decision Title]

   | Field               | Value                                         |
   | ------------------- | --------------------------------------------- |
   | **Status**          | Proposed / Accepted / Superseded / Deprecated |
   | **Date**            | YYYY-MM-DD                                    |
   | **Decider (D)**     | [Name/Role]                                   |
   | **Recommender (R)** | [Name/Role]                                   |
   | **Agrees (A)**      | [Names/Roles]                                 |
   | **Input (I)**       | [Names/Roles]                                 |
   | **Performers (P)**  | [Names/Roles]                                 |
   | **Supersedes**      | [TD-XXXX or N/A]                              |
   | **Superseded by**   | [TD-XXXX or N/A]                              |

   ## Context

   [Why is this decision needed? What is the problem or opportunity?]

   ## Options Considered

   ### Option 1: [Name]

   - **Pros**: ...
   - **Cons**: ...
   - **Cost**: ...
   - **Reversibility**: ...

   ### Option 2: [Name]

   ...

   ### Option 3: [Name]

   ...

   ## Decision

   We chose **Option [N]: [Name]**.

   ### Rationale

   [Why this option was selected]

   ### Trade-offs Accepted

   [What downsides we're knowingly accepting]

   ### Constraints / Assumptions

   [Conditions that must hold for this decision to remain valid]

   ### Review Trigger

   [When to revisit: date, milestone, or condition]

   ## Consequences

   ### Positive

   - [Expected benefit]

   ### Negative

   - [Accepted downside]

   ### Risks

   - [Risk and mitigation]
   ```

7. **Save the Output**: Save as `docs/decisions/TD-NNNN-[slug].md` (e.g., `docs/decisions/TD-0042-switch-to-event-driven-architecture.md`).

## Gotchas

- A decision record is not a proposal — it documents a decision that has been (or is being) made. If the decision is still exploratory, set status to `Proposed` and identify who needs to move it to `Accepted`.
- Don't skip the "Options Considered" section even if the choice seems obvious — the record should show alternatives were evaluated.
- RAPID roles need exactly one **D** (Decide). If you can't identify a single decision-maker, that's a governance problem to surface.
- When a decision is superseded, update the original record's status and add a `Superseded by` link — don't just create a new record in isolation.

## Related Skills

- **azure-adr** — Architecture Decision Records scoped to Azure infrastructure (uses WAF mapping)
- **pressure-test** — Stress-test the options before deciding
- **escalation-tracker** — Track decisions that are escalated or blocked
- **okr-alignment-check** — Verify the decision aligns with strategic objectives
