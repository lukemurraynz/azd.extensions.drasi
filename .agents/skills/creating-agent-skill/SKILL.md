---
name: creating-agent-skill
description: >-
  Create and review Agent Skills that follow the open Agent Skills specification and best practices. USE FOR: create new skill, generate SKILL.md, build skill instructions, review skill compliance, improve skill quality, structure skill directory, write skill description.
license: MIT
---

# Creating Agent Skills

## Overview

This skill provides capabilities for creating and reviewing Agent Skills that comply with the [Agent Skills specification](https://agentskills.io/specification) and follow [authoring best practices](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices).

Agent Skills are a lightweight, open format for extending AI agent capabilities with specialized knowledge and workflows. A skill is a folder containing a `SKILL.md` file with metadata and instructions, optionally bundled with scripts, references, and assets.

## Capabilities

| Capability | Action              | Description                                               |
| ---------- | ------------------- | --------------------------------------------------------- |
| Create     | `actions/create.md` | Generate a new, specification-compliant Agent Skill       |
| Review     | `actions/review.md` | Analyse an existing skill for compliance and improvements |

## Standards

This skill bundles the following standards in `standards/`:

| Standard             | File                      | Description                                            |
| -------------------- | ------------------------- | ------------------------------------------------------ |
| Specification        | `specification.md`        | Core Agent Skills format specification                 |
| Naming               | `naming.md`               | Naming conventions for skills and files                |
| Descriptions         | `descriptions.md`         | Writing effective skill descriptions                   |
| Structure            | `structure.md`            | Directory structure and progressive disclosure         |
| Instructions         | `instructions.md`         | Writing effective skill instructions                   |
| Clarifying Questions | `clarifying-questions.md` | Questioning framework for gathering skill requirements |
| Checklist            | `checklist.md`            | Consolidated compliance and quality checklist          |

## Principles

### 1. Concise is Key

The context window is a shared resource. Only add context the agent doesn't already have. Challenge each piece of information: "Does the agent really need this explanation?"

### 2. Progressive Disclosure

Structure skills for efficient context use:

- **Metadata** (~100 tokens): `name` and `description` loaded at startup
- **Instructions** (<5000 tokens recommended): Full `SKILL.md` body loaded when activated
- **Resources** (as needed): Additional files loaded only when required

### 3. Set Appropriate Degrees of Freedom

Match specificity to task fragility:

- **High freedom**: Text-based instructions for flexible tasks
- **Medium freedom**: Pseudocode or parameterised scripts
- **Low freedom**: Exact scripts for fragile, error-prone operations

### 4. Test with Real Usage

Build evaluations before extensive documentation. Test with representative tasks. Iterate based on observed agent behavior, not assumptions.

## Usage

1. Load this skill manifest
2. Identify the required capability (create or review)
3. Load the bundled standards from `standards/`
4. Execute the action following `actions/<capability>.md`

## References

- [Agent Skills Specification](https://agentskills.io/specification)
- [What are Agent Skills?](https://agentskills.io/what-are-skills)
- [Skill Authoring Best Practices](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices)
- [Example Skills](https://github.com/anthropics/skills)
- [Reference Library](https://github.com/agentskills/agentskills/tree/main/skills-ref)

---

## Related Skills

- [Creating Dev Containers](../creating-devcontainers/SKILL.md) — Companion meta-skill for dev environment setup

---

## Known Pitfalls

| Area | Pitfall | Mitigation |
| --- | --- | --- |
| Description length | Descriptions exceeding 1024 characters (spec maximum) are rejected; overly long descriptions waste the ~100 token metadata budget | Keep descriptions 150-300 characters with clear "USE FOR:" triggers; validate against the 1024-character spec limit |
| Context budget waste | Bundling >5000 tokens in SKILL.md loads unnecessary context at activation | Structure as ~500 tokens overview, ~2000 tokens key patterns; defer detail to referenced files |
| Standards as dead docs | `standards/` directories created but never referenced in SKILL.md; agents do not auto-discover them | Explicitly reference each standard inline so agents know they exist |
| Frontmatter YAML folding | Multi-line `description` using `>-` can concatenate unrelated YAML keys into the description text | Validate frontmatter parses correctly; check that `description` does not contain stray key-value pairs |

---

## Currency

- **Date checked:** 2026-03-31
- **Sources:** [Agent Skills Specification](https://agentskills.io/specification), [Skill Authoring Best Practices](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices)
- **Authoritative references:** [Agent Skills spec](https://agentskills.io/specification)

### Verification Steps

1. Confirm Agent Skills specification version and any new required fields
2. Verify SKILL.md schema requirements (description length, USE FOR triggers)
3. Check for new skill capability categories or metadata fields
