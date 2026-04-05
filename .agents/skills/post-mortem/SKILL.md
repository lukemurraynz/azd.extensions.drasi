---
name: post-mortem
description: >-
  Conduct a blameless post-mortem analysis after an incident, failure, or missed goal — root cause analysis, timeline reconstruction, contributing factors, and actionable follow-ups. USE FOR: analyzing why something went wrong, learning from production incidents, or documenting lessons after a missed target.
---

# Post-Mortem: Blameless Incident & Failure Analysis

## Purpose

You are a post-mortem facilitator conducting a blameless analysis of $ARGUMENTS. Your job is to reconstruct what happened, identify root causes and contributing factors, and produce actionable follow-ups that prevent recurrence — without assigning blame to individuals.

## Context

A post-mortem analyzes something that already happened (an incident, a missed launch, a failed initiative) to extract learning. It differs from a **pre-mortem** (which imagines future failure before launch) and a **retro** (which reflects on a sprint's process). Post-mortems focus on understanding causation, not just correlation, and produce systemic improvements rather than individual accountability.

## Instructions

1. **Gather Incident Data**: If the user provides files (incident reports, logs, timelines, metrics, Slack exports, status page history), read them thoroughly. Ask for any missing context:
   - What happened?
   - When was it detected? By whom?
   - What was the impact (users affected, revenue lost, SLA breached)?
   - When was it resolved?

2. **Reconstruct the Timeline**:

   Build a detailed, factual timeline in UTC:

   | Time (UTC)       | Event           | Actor/System | Source                    |
   | ---------------- | --------------- | ------------ | ------------------------- |
   | YYYY-MM-DD HH:MM | [What happened] | [Who/what]   | [Log, alert, Slack, etc.] |

   Include:
   - First signal / trigger event
   - Detection (automated alert vs. human notice)
   - Escalation points
   - Investigation steps taken
   - Mitigation actions
   - Full resolution
   - Communication milestones (status page, customer notification)

3. **Measure Impact**:

   | Metric                       | Value                                     |
   | ---------------------------- | ----------------------------------------- |
   | **Duration**                 | Time from first impact to full resolution |
   | **Time to Detect (TTD)**     | First impact → detection                  |
   | **Time to Mitigate (TTM)**   | Detection → impact reduced                |
   | **Time to Resolve (TTR)**    | Detection → full resolution               |
   | **Users/customers affected** | [Count or percentage]                     |
   | **Revenue impact**           | [Estimate if applicable]                  |
   | **SLA impact**               | [Which SLAs were breached]                |
   | **Data loss**                | [Yes/No — describe if yes]                |

4. **Root Cause Analysis**:

   Use the **5 Whys** technique to drill past symptoms to root causes:
   - **Why did [symptom] happen?** → Because [cause 1]
   - **Why did [cause 1] happen?** → Because [cause 2]
   - **Why did [cause 2] happen?** → Because [cause 3]
   - Continue until you reach a systemic root cause (not a human error)

   Identify:
   - **Root cause**: The fundamental systemic issue
   - **Contributing factors**: Conditions that made the incident worse or more likely
   - **Triggers**: The specific event that initiated the incident

   **Blameless principle**: Root causes are always systemic (missing guardrails, inadequate testing, unclear runbooks) — never "person X made a mistake." If a human error was involved, ask why the system allowed that error to cause impact.

5. **Assess What Went Well**:
   Not everything fails during an incident. Identify:
   - Detection mechanisms that worked
   - Runbooks or procedures that helped
   - Communication that was effective
   - Quick thinking or improvisation that reduced impact

6. **Generate Action Items**:

   | Priority | Action Item | Type                      | Owner   | Deadline | Prevents Recurrence Of       |
   | -------- | ----------- | ------------------------- | ------- | -------- | ---------------------------- |
   | P0       | [Action]    | [Prevent/Detect/Mitigate] | [Owner] | [Date]   | [Which root cause or factor] |
   | P1       | [Action]    | [Prevent/Detect/Mitigate] | [Owner] | [Date]   | [Which root cause or factor] |

   Classify each action by type:
   - **Prevent**: Stop the root cause from occurring (best)
   - **Detect**: Catch it faster next time (reduce TTD)
   - **Mitigate**: Reduce impact when it does happen (reduce TTM)

   Limit to 5–7 action items. More than that won't get completed.

7. **Structure the Output**:

   ```
   ## Post-Mortem: [Incident Title]
   **Date of incident**: YYYY-MM-DD
   **Post-mortem date**: YYYY-MM-DD
   **Severity**: S1 / S2 / S3 / S4
   **Status**: Draft / Reviewed / Final

   ### Executive Summary
   [2-3 sentence summary: what happened, impact, root cause, current status]

   ### Impact
   [Impact metrics table]

   ### Timeline
   [Chronological event table]

   ### Root Cause Analysis
   [5 Whys chain]
   - **Root cause**: [Systemic issue]
   - **Contributing factors**: [List]
   - **Trigger**: [Specific event]

   ### What Went Well
   - [Positive observations]

   ### What Went Poorly
   - [Areas for improvement]

   ### Action Items
   [Prioritized action table]

   ### Lessons Learned
   - [Key takeaways for the broader team/org]

   ### Follow-Up Schedule
   - [ ] [Date]: Review action item completion
   - [ ] [Date]: Validate fixes with load test / chaos experiment
   ```

8. **Save the Output**: Save as `PostMortem-[incident-slug]-[date].md`.

## Gotchas

- **Blameless means blameless.** If your post-mortem identifies "human error" as a root cause, dig deeper. The root cause is the system that allowed human error to have impact.
- **Hindsight bias is real.** The team made decisions with the information they had at the time. Evaluate actions in context, not with perfect hindsight.
- **Follow-up is the whole point.** A post-mortem without tracked action items is just storytelling. Ensure every action item has an owner and deadline.
- **Don't wait too long.** Post-mortems lose value as memory fades. Conduct within 5 business days of resolution.
- **Share broadly.** The value of a post-mortem scales with how many people learn from it. Default to sharing across the engineering org.

## Related Skills

- **pre-mortem** — Imagine failure before it happens (proactive counterpart to post-mortem)
- **retro** — Sprint-level reflection on process (broader scope than incident-specific)
- **escalation-tracker** — If the incident revealed escalation failures, improve the framework
- **decision-record** — Document key decisions made during or after the incident as formal TDs
- **threat-modelling** — If the incident was security-related, update the threat model
