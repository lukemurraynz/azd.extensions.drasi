---
name: escalation-tracker
description: >-
  Create and manage an escalation dashboard with SLA tracking, decision rights matrix, and resolution workflows. USE FOR: tracking blockers that need leadership attention, managing cross-team dependencies, or monitoring escalation SLAs.
---

# Escalation Tracker: SLA-Monitored Decision Escalation

## Purpose

Create or update an escalation tracker for $ARGUMENTS. Ensure blocked decisions and cross-team issues are surfaced, assigned, SLA-tracked, and resolved, not lost in Slack threads.

## Context

Escalations are decisions or blockers the current team cannot resolve independently. They require someone with broader authority, cross-team coordination, or explicit trade-off approval. Without structured tracking, escalations stall in side channels, lose context, and erode trust. This skill provides a lightweight but rigorous framework for raising, tracking, and closing escalations.

## Instructions

1. **Determine the Action**: The user may want to:
   - **Create** a new escalation
   - **Update** an existing escalation's status
   - **Review** the full escalation dashboard
   - **Establish** the escalation framework for a project/team

2. **If Creating the Framework** (first-time setup):

   ### Decision Rights Matrix

   Define who can resolve what:

   | Decision Type          | Resolver      | Escalation Path           | SLA             |
   | ---------------------- | ------------- | ------------------------- | --------------- |
   | Technical architecture | Tech Lead     | → Engineering Director    | 48h             |
   | Cross-team dependency  | PM            | → Group PM → VP Product   | 72h             |
   | Budget/resourcing      | EM            | → Director of Engineering | 5 business days |
   | Customer commitment    | PM + Sales    | → VP Product + VP Sales   | 24h             |
   | Security/compliance    | Security Lead | → CISO                    | 24h             |
   | Vendor/contract        | EM            | → Procurement → Legal     | 5 business days |

   Customize this matrix based on the team's context.

   ### Severity Levels

   | Severity          | Definition                                    | Response SLA | Resolution SLA   |
   | ----------------- | --------------------------------------------- | ------------ | ---------------- |
   | **S1, Critical** | Blocks launch / revenue / customer commitment | 4h response  | 24h resolution   |
   | **S2, High**     | Blocks sprint goal or cross-team deliverable  | 24h response | 72h resolution   |
   | **S3, Medium**   | Impacts timeline but has workaround           | 48h response | 5 business days  |
   | **S4, Low**      | Process improvement or non-blocking question  | 72h response | 10 business days |

3. **If Creating a New Escalation**:

   ```markdown
   ## ESC-NNNN: [Title]

   | Field               | Value                                            |
   | ------------------- | ------------------------------------------------ |
   | **Severity**        | S1 / S2 / S3 / S4                                |
   | **Status**          | Open / In Progress / Blocked / Resolved / Closed |
   | **Raised by**       | [Name, Date]                                     |
   | **Assigned to**     | [Resolver name/role]                             |
   | **Response SLA**    | [Deadline]                                       |
   | **Resolution SLA**  | [Deadline]                                       |
   | **Escalation path** | [Next level if SLA breached]                     |

   ### Context

   [What is blocked and why it can't be resolved at the current level]

   ### Impact

   - **What is blocked**: [Deliverable, team, timeline]
   - **Downstream effects**: [What else is delayed if this isn't resolved]
   - **Cost of delay**: [Per day/week impact estimate]

   ### Decision Needed

   [Specific question or trade-off that needs to be resolved]

   ### Options on the Table

   | Option      | Pros   | Cons   | Recommended by |
   | ----------- | ------ | ------ | -------------- |
   | A: [Option] | [Pros] | [Cons] | [Who]          |
   | B: [Option] | [Pros] | [Cons] | [Who]          |

   ### Resolution

   - **Decision**: [What was decided (filled when resolved)]
   - **Decided by**: [Who made the call]
   - **Date resolved**: [When]
   - **Follow-up actions**: [Any resulting work items]
   ```

4. **If Reviewing the Dashboard**:

   Generate a summary view:

   ```
   ## Escalation Dashboard, [Date]

   ### SLA Health
   - Open escalations: [N]
   - Within SLA: [N] ([%])
   - SLA breached: [N] ([%])
   - Avg resolution time: [X days]

   ### Active Escalations
   | ID | Title | Severity | Status | Owner | SLA Status | Days Open |
   |---|---|---|---|---|---|---|
   | ESC-0001 | [Title] | S2 | In Progress | [Name] | ✅ On Track | 3 |
   | ESC-0002 | [Title] | S1 | Blocked | [Name] | ❌ Breached | 5 |

   ### Recently Resolved
   | ID | Title | Resolution Time | Decision |
   |---|---|---|---|
   | ESC-0003 | [Title] | 2 days | [Summary] |

   ### Trends
   - Escalations this period: [N] (up/down from [N] last period)
   - Most common category: [Type]
   - Average resolution: [X days] (target: [Y days])
   - SLA compliance: [%]
   ```

5. **Save the Output**:
   - Framework: `docs/escalations/ESCALATION-FRAMEWORK.md`
   - Individual escalations: `docs/escalations/ESC-NNNN-[slug].md`
   - Dashboard: `docs/escalations/DASHBOARD.md`

## Gotchas

- An escalation is not a complaint. It's a structured request for help with a specific decision or blocker. Frame escalations around the decision needed, not the frustration.
- SLA timers start when the escalation is acknowledged by the resolver, not when it's raised. Track both "raised" and "acknowledged" timestamps.
- If the resolver doesn't respond within the response SLA, auto-escalate to the next level in the escalation path. Don't wait politely.
- Closing an escalation without documenting the resolution defeats the purpose. Every closed escalation should have a "Decision" and "Follow-up actions" filled in.

## Related Skills

- **decision-record**, document the resulting decision formally (TD-NNNN) once an escalation is resolved
- **pressure-test**, stress-test the options before the resolver decides
- **post-mortem**, if an escalation reveals a systemic issue, run a post-mortem on the root cause
- **retro**, surface recurring escalation patterns in retrospectives
