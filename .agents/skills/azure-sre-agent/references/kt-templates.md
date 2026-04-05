# KT Templates for Azure SRE Agent

Copy and use these templates when KT methodology is required.

## 1) Situation Appraisal (SA)

```markdown
## Situation Appraisal

### Identify Concerns
- Threats/opportunities:
- Deviations occurring:
- Decisions required:
- Plans to implement:
- Changes anticipated:
- What bothers us about this situation:

### Separate and Clarify
- Concern #1:
  - What exactly is it:
  - Evidence:
  - Related deviations/decisions/plans:
- Concern #2:
  - What exactly is it:
  - Evidence:
  - Related deviations/decisions/plans:

### Set Priority (Current/Future/Time)
| Concern | Current Impact | Future Impact | Time Frame | Priority |
| --- | --- | --- | --- | --- |
| | | | | |

### Plan Next Steps
- Analysis needed now (`PA`, `DA`, `PPA`):
- What must be done and when:
- Who is needed for information:
- Who is needed for approval/implementation:
```

## 2) Problem Analysis (PA)

```markdown
## Problem Analysis

### State the Problem
- Object/group with deviation:
- Deviation observed:
- Sensory/telemetry evidence:

### IS / IS NOT Specification
| Dimension | IS | IS NOT | Distinctions | Changes |
| --- | --- | --- | --- | --- |
| WHAT (object) | | | | |
| WHAT (deviation) | | | | |
| WHERE (geo/system) | | | | |
| WHERE (on object/component) | | | | |
| WHEN (first observed) | | | | |
| WHEN (pattern since) | | | | |
| WHEN (lifecycle stage) | | | | |
| EXTENT (how many/size/trend) | | | | |

### Possible Causes
1.
2.
3.

### Evaluate Possible Causes
- Cause tested:
- How it explains IS and IS NOT:
- Assumptions required:
- Confidence:

### Confirm True Cause
- Verification method:
- Observed confirmation:
- Most probable cause:
```

## 3) Decision Analysis (DA)

```markdown
## Decision Analysis

### Clarify Purpose
- Decision statement (choice + result + key modifiers):
- Decision owner:
- Decision deadline:

### Objectives
| Objective | Measure | Type (`MUST`/`WANT`) | Weight (for WANT) |
| --- | --- | --- | --- |
| | | | |

### Alternatives
1.
2.
3.

### MUST Screen
| Alternative | Meets All MUSTs? | Notes |
| --- | --- | --- |
| | | |

### WANT Scoring
| Alternative | Weighted Score | Key Strengths | Key Weaknesses |
| --- | --- | --- | --- |
| | | | |

### Recommendation
- Selected option:
- Why selected:
- Why non-selected options were rejected:
```

## 4) Potential Problem Analysis (PPA)

```markdown
## Potential Problem Analysis

### Action Plan to Protect
- End result to achieve:
- Critical actions:

### Potential Problems
| Potential Problem | Probability | Seriousness | Likely Causes |
| --- | --- | --- | --- |
| | | | |

### Preventive Actions
| Likely Cause | Preventive Action | Owner | Due Date |
| --- | --- | --- | --- |
| | | | |

### Contingent Actions and Triggers
| Potential Problem | Trigger (How we know) | Contingent Action | Owner |
| --- | --- | --- | --- |
| | | | |
```

## 5) Combined Incident Output Template (KT Complete)

```markdown
# Incident KT Analysis

## Incident Summary
- Service:
- Environment:
- Severity:
- User impact:
- Time detected (UTC):

## Situation Appraisal
[SA content]

## Problem Analysis
[PA content]

## Decision Analysis
[DA content]

## Potential Problem Analysis
[PPA content]

## Recommended Action
- Action:
- Why:
- Expected outcome:

## Risk and Rollback
- Primary risks:
- Rollback trigger:
- Rollback action:

## Owners and Next Steps
- Owner(s):
- Immediate steps:
- Verification plan:
```

## KT Question Prompts (from workbook structure)

Use these prompts when the model needs nudges:

### SA prompts

1. What deviations are occurring?
2. Which concern should we work on first?
3. What is the deadline and when do we need to start?
4. Which analysis is needed next?

### PA prompts

1. What object has the deviation?
2. What is observed in `IS` that is absent in `IS NOT`?
3. What distinctions/changes suggest a cause?
4. Which cause best explains both `IS` and `IS NOT`?

### DA prompts

1. What do we need to decide?
2. Which objectives are mandatory (`MUST`)?
3. How should wants be weighted?
4. Which alternative best meets weighted objectives?

### PPA prompts

1. When we execute this plan, what could go wrong?
2. What likely causes could trigger each problem?
3. What preventive action reduces probability?
4. What trigger activates each contingent action?
