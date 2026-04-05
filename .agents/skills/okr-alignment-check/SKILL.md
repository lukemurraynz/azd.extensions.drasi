---
name: okr-alignment-check
description: >-
  Score work items, features, or initiatives against team OKRs to detect strategic drift, measure alignment, and identify coverage gaps. USE FOR: validating a roadmap against OKRs, checking if a feature request aligns with objectives, or auditing quarterly focus.
---

# OKR Alignment Check: Strategic Alignment Validator

## Purpose

You are a strategic alignment analyst evaluating $ARGUMENTS against the team's OKRs. Your job is to quantify how well current work connects to stated objectives, surface misalignment, and identify OKRs that lack sufficient investment.

## Context

Teams set ambitious OKRs but frequently drift — shipping features that feel productive but don't move key results. An alignment check is a periodic audit that scores work against objectives, catches strategic drift early, and ensures resource allocation matches stated priorities. This complements OKR creation (see `brainstorm-okrs`); alignment checking validates execution against intent.

## Instructions

1. **Gather OKRs and Work Items**:
   - Ask for or locate the team's current OKRs (objectives + key results with targets)
   - Gather the work items to evaluate: backlog items, roadmap features, sprint plans, or a specific initiative
   - If provided as files, read them thoroughly

2. **Score Each Work Item for Alignment**:

   For each work item or initiative, assess:

   | Field                 | Assessment                                                              |
   | --------------------- | ----------------------------------------------------------------------- |
   | **Work Item**         | [Name/description]                                                      |
   | **Primary OKR**       | Which objective + key result does this most directly serve?             |
   | **Alignment Score**   | 0–10 (see rubric below)                                                 |
   | **Key Result Impact** | How much does this move the target metric? (None / Low / Medium / High) |
   | **Confidence**        | How certain is the causal link? (Proven / Hypothesized / Speculative)   |

   **Alignment Score Rubric**:
   | Score | Meaning |
   |---|---|
   | 9–10 | Directly moves a key result; clear causal link with evidence |
   | 7–8 | Strong connection to an objective; measurable but indirect impact |
   | 5–6 | Loosely related; supports the objective's spirit but not a specific KR |
   | 3–4 | Tangential; might contribute but connection is speculative |
   | 1–2 | No meaningful connection to any current OKR |
   | 0 | Actively works against an objective or is pure overhead |

3. **OKR Coverage Analysis**:

   For each OKR, assess how much active work supports it:

   | Objective | Key Result | Target   | Work Items Mapped | Coverage                    | Risk                         |
   | --------- | ---------- | -------- | ----------------- | --------------------------- | ---------------------------- |
   | [O1]      | [KR1.1]    | [Target] | [Count + names]   | [Strong/Adequate/Thin/None] | [On Track/At Risk/Off Track] |

   Flag any OKR with **Thin** or **None** coverage — these are objectives the team claims to care about but isn't investing in.

4. **Strategic Drift Detection**:

   Calculate the alignment distribution:
   - **% of work scoring 7+**: Core aligned work
   - **% of work scoring 4–6**: Loosely aligned (watch zone)
   - **% of work scoring 0–3**: Unaligned work (drift zone)

   **Health assessment**:
   | Distribution | Health |
   |---|---|
   | 70%+ core aligned | Healthy — team is focused |
   | 50–70% core aligned | Drifting — need to re-prioritize |
   | <50% core aligned | Misaligned — urgent strategy review needed |

5. **Generate Recommendations**:
   - **Cut or defer**: Work items scoring 0–3 with no strategic justification
   - **Investigate**: Work items scoring 4–6 — can they be reframed to better serve an OKR?
   - **Double down**: Under-invested OKRs that need more work items
   - **Missing bets**: Are there key results with no work items at all?

6. **Structure the Output**:

   ```
   ## OKR Alignment Check: [Team/Quarter]

   ### Summary
   - Total work items evaluated: [N]
   - Core aligned (7+): [N] ([%])
   - Watch zone (4–6): [N] ([%])
   - Drift zone (0–3): [N] ([%])
   - **Health: [Healthy / Drifting / Misaligned]**

   ### Work Item Alignment Scores
   | Work Item | Primary OKR | Score | KR Impact | Confidence |
   |---|---|---|---|---|
   | [Item] | [OKR] | [0–10] | [H/M/L/None] | [Proven/Hyp/Spec] |

   ### OKR Coverage
   | Objective | Key Result | Coverage | Risk |
   |---|---|---|---|
   | [O1] | [KR1.1] | [Strong/Adequate/Thin/None] | [Status] |

   ### Recommendations
   #### Cut or Defer
   - [Work items with justification]

   #### Reframe
   - [Work items that could better serve OKRs]

   #### Increase Investment
   - [Under-served OKRs]

   #### Missing Work
   - [OKRs with no mapped work items]
   ```

7. **Save the Output**: Save as `OKRAlignmentCheck-[team]-[quarter]-[date].md`.

## Gotchas

- Not all valuable work maps to OKRs — maintenance, tech debt, and compliance are real. Flag them separately as "strategic overhead" rather than scoring them 0.
- Alignment ≠ priority. A perfectly aligned item can still be low priority if the KR it serves is already on track.
- Don't force-fit alignment — a score of 2 is honest. The goal is clarity, not making everything look aligned.
- OKRs themselves might be wrong. If most valuable work scores low, consider whether the OKRs need updating rather than the work.

## Related Skills

- **brainstorm-okrs** — Create OKRs (this skill validates alignment after they exist)
- **pressure-test** — Stress-test the strategy behind the OKRs
- **decision-record** — Document the decision to re-prioritize based on alignment findings
- **backlog-refinement** agent — Refine the backlog incorporating alignment insights
