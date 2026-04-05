---
name: red-team
description: >-
  Think as the competitor: analyze your strategy from the adversary's perspective, generate counter-moves, identify strategic vulnerabilities, and recommend defenses. USE FOR: stress-testing competitive positioning, preparing for market response, or anticipating rival reactions to a launch.
---

> [!NOTE]
> This skill covers **competitive strategy red-teaming** (challenging product assumptions, market positioning, and business model weaknesses from an adversary's perspective). For **security red-teaming** (penetration testing, vulnerability assessment, attack simulation), see the `threat-modelling` skill.

# Red Team: Competitive Counter-Strategy

## Purpose

You are a strategic analyst role-playing as a direct competitor evaluating $ARGUMENTS. Your job is to find every vulnerability in the strategy, generate realistic counter-moves, and then switch back to advise on defenses.

## Context

Red-teaming goes beyond competitive analysis (which observes competitors) and battlecards (which arm sales). A red team actively thinks **as** the competitor — with their resources, constraints, and incentives — to anticipate how they would respond to your moves. This surfaces blind spots that internal analysis misses because it forces perspective-shifting.

## Instructions

1. **Gather Intelligence**: If the user provides documents (strategy, PRD, launch plan, competitive data), read them first. Use web search to research the specified competitor's current strategy, recent moves, funding, product roadmap signals, and leadership commentary.

2. **Establish the Competitor Profile**:
   - **Identity**: Who are we role-playing as? (company, product, market position)
   - **Resources**: Engineering team size, funding, revenue, partnerships
   - **Strategic posture**: Aggressive/defensive, market leader/challenger, growth stage
   - **Known priorities**: Where are they investing? What have they announced?
   - **Constraints**: What can't they easily do? (technical debt, regulatory, brand positioning)

3. **Red Team Analysis** (think as the competitor):

   ### Vulnerability Scan

   For each element of the plan/strategy, ask: "If I were [Competitor], how would I exploit this?"
   - Product gaps the competitor already fills
   - Pricing vulnerabilities (undercut, bundle, freemium)
   - Go-to-market timing windows the competitor could exploit
   - Customer relationships the competitor could leverage
   - Technical advantages the competitor holds

   ### Counter-Move Generation

   Generate 3–5 realistic competitive responses:
   | # | Counter-Move | Likelihood | Impact | Speed to Execute |
   |---|---|---|---|---|
   | 1 | [What they'd do] | [High/Med/Low] | [High/Med/Low] | [Weeks/Months/Quarters] |

   For each counter-move, explain the competitor's reasoning and what internal signals would indicate they're pursuing it.

   ### Attack Scenarios

   Describe 2–3 specific attack scenarios:
   - **Scenario A — [Name]**: [Detailed narrative of how the competitor disrupts the plan]
   - **Scenario B — [Name]**: [Alternative attack vector]
   - **Scenario C — [Name]**: [Unconventional/surprising move]

4. **Switch to Defense** (back to your team's perspective):

   ### Recommended Defenses

   For each high-likelihood counter-move:
   | Threat | Defense | Investment Required | Priority |
   |---|---|---|---|
   | [Counter-move] | [Defensive action] | [Low/Med/High] | [P0/P1/P2] |

   ### Early Warning Signals

   List observable signals that indicate a competitor is executing a counter-move:
   - Job postings, patent filings, partnership announcements
   - Pricing changes, feature launches, marketing shifts
   - Conference talks, blog posts, analyst briefings

   ### Moat Assessment

   Rate the durability of your key differentiators:
   | Differentiator | Defensibility | Time to Copy | Recommended Moat Investment |
   |---|---|---|---|
   | [Advantage] | [Strong/Moderate/Weak] | [Months/Quarters/Years] | [Action to strengthen] |

5. **Structure the Output**:

   ```
   ## Red Team Analysis: [Your Strategy] vs. [Competitor]

   ### Competitor Profile
   [Summary of who we're role-playing as]

   ### Vulnerabilities Identified
   [Numbered list with severity]

   ### Counter-Moves
   [Table of likely competitive responses]

   ### Attack Scenarios
   [Narrative scenarios]

   ### Recommended Defenses
   [Table of defensive actions]

   ### Early Warning Signals
   [Observable indicators]

   ### Moat Assessment
   [Durability of advantages]

   ### Strategic Verdict
   [Overall assessment: How exposed is the strategy?]
   ```

6. **Save the Output**: Save as `RedTeam-[strategy]-vs-[competitor]-[date].md`.

## Gotchas

- Don't strawman the competitor — give them credit for being smart and well-resourced. Weak red teams produce false confidence.
- The competitor's best move might not be a direct feature copy — consider pricing, partnerships, acquisitions, and ecosystem plays.
- Red-teaming a strategy with multiple competitors? Run a separate analysis for each; their responses will differ.
- Time horizon matters: a startup competitor responds differently than an incumbent.

## Related Skills

- **competitor-analysis** — Objective analysis of competitor landscape (observation, not role-play)
- **competitive-battlecard** — Sales-ready materials for head-to-head deals
- **pressure-test** — Seven-lens stress test of the overall plan
- **pre-mortem** — Imagine failure scenarios (broader than competitive, includes internal risks)
