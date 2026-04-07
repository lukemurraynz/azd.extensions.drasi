---
applyTo: "**/*.md,**/docs/**"
description: "Documentation and Markdown best practices aligned with the Industry Solutions Engineering Playbook, Microsoft Writing Style Guide, and built-in humanization guardrails"
---

# Documentation and Markdown Instructions

Write documentation that is clear, specific, and written by a human for other humans.
Technical accuracy is mandatory. Readability, clarity, and tone are equally mandatory.

**IMPORTANT**: Use `microsoft.learn.mcp` MCP server to verify current Microsoft Writing Style Guide conventions and terminology. Use `iseplaybook` MCP server for ISE documentation standards. Do not assume style rules — verify against current guidance.

This instruction set aligns with:

- Microsoft ISE (Industry Solutions Engineering) Engineering Playbook
- Microsoft Writing Style Guide
- Explicit anti-AI and humanization rules

---

## When to Apply

Apply these instructions when:

- Creating or editing Markdown files
- Writing README files
- Producing technical documentation
- Writing Architecture Decision Records (ADRs)
- Editing documentation generated or assisted by an LLM

---

## Core Principles (Non-Negotiable)

1. Clarity over cleverness
   Say what something is and does. Avoid inflated or abstract language.

2. Specific beats impressive
   Prefer concrete details over general claims.

3. Neutral, not promotional
   Documentation explains behavior. It does not market solutions.

4. Human voice is allowed
   First person is acceptable when it improves clarity or honesty.

5. No AI fingerprints
   If a sentence sounds like a press release, rewrite it.

---

## Microsoft Writing Style Alignment (Mandatory)

All documentation must follow these Microsoft Writing Style Guide principles:

- Use short, simple sentences.
- Use sentence case for headings and labels.
- Prefer active voice.
- Use imperative verbs in procedures.
- Write in a conversational, natural tone.
- Be concise and remove filler words.
- Avoid idioms, slang, and cultural references.
- Use the Oxford (serial) comma in lists.
- Use bias-free and inclusive language.
- Avoid punctuation at the end of headings.

If Microsoft guidance conflicts with stylistic preference, Microsoft guidance wins.

---

## README Structure

Use this as a baseline, not a rigid template.

    # Project name

    One or two plain sentences explaining what the project does.

    ## Why this exists

    Brief context. What problem it solves and for whom.

    ## Features

    - Concrete capabilities only
    - Avoid aspirational or promotional language

    ## Prerequisites

    - Explicit versions and requirements

    ## Installation

    Step-by-step instructions that actually work.

    ## Usage

    Minimal, copy-paste-ready examples.

    ## Configuration

    Key options, defaults, and trade-offs.

    ## Contributing

    How to contribute and where to start.

    ## License

    License information.

Avoid:

- “This project aims to…”
- “A powerful, flexible, and scalable solution…”
- “It's a game-changer…”
- Generic “Future improvements” sections

---

## Markdown Formatting Standards

### Headings

- Use sentence case
- One H1 per document
- No punctuation at the end of headings

Example:

    # Project overview

    ## How authentication works

    ### Token validation flow

---

### Code Examples

- Examples must compile or run
- Prefer minimal examples over completeness

Example:

    def hello():
        return "Hello, world"

    npm install
    npm run build

---

### Lists

Use lists for facts, not filler.

Example:

    - Supports OAuth2 token validation
    - Caches tokens for five minutes
    - Logs failed authentication attempts

Avoid forcing ideas into groups of three unless necessary.

---

### Links

- Use descriptive link text
- Prefer relative links inside repositories

Example:

    [Deployment guide](../guides/deployment.md)

Avoid:

- “Click here”
- Bare URLs in prose

---

### Tables

Use tables only when comparison matters.

Example:

    | Option  | Default | Notes                      |
    |--------|---------|----------------------------|
    | timeout| 30s     | Increase for slow networks |

---

### Images

- Images must add information
- Always include meaningful alt text

Example:

    ![Request flow between API and worker](images/request-flow.png)

---

## Technical Documentation

### API Documentation

Focus on behavior, not ceremony.

Example:

    ## GET /api/users/{id}

    Returns a user record by ID.

    ### Parameters

    | Name | Type   | Required | Description     |
    |------|--------|----------|-----------------|
    | id   | string | Yes      | User identifier |

    ### Response

    {
      "id": "123",
      "name": "Jane Doe"
    }

    ### Errors

    | Code | When it happens       |
    |------|-----------------------|
    | 404  | User does not exist   |
    | 500  | Database unavailable |

Avoid:

- “Plays a crucial role”
- “Ensures a seamless experience”

---

## Architecture Decision Records (ADR)

ADRs must be factual and honest.

Example:

    # ADR-001: Use PostgreSQL for user data

    ## Status
    Accepted

    ## Context
    We need transactional guarantees and relational queries.

    ## Decision
    We will use PostgreSQL.

    ## Consequences

    ### Positive
    - Strong relational support
    - Mature tooling

    ### Negative
    - Requires operational maintenance

Do not include:

- Strategic alignment
- Future-proofing
- Industry best practice without evidence

---

## Humanization Rules (Mandatory)

### Prefer plain language

Use:

- is / are / has / does
- Direct, concrete phrasing

Avoid:

- serves as
- stands as
- represents
- abstract nouns like landscape, journey, tapestry

---

### Avoid promotional tone

Bad:

    A powerful and flexible solution designed to enhance productivity.

Better:

    The tool batches requests and retries failures automatically.

---

### Avoid vague attribution

Do not write:

- “Experts believe…”
- “Industry reports suggest…”

Either cite a source or remove the claim.

---

### Avoid formulaic sections

Avoid headings like:

- Challenges and future outlook
- Key takeaways
- Final thoughts

End documents when the useful information ends.

---

### Avoid AI-isms

Do not use:

- Excessive em dashes
- Overuse of bold text
- Emojis in headings or lists
- Curly quotes
- Chatbot phrases like “Let me know if…”

---

### Punctuation Precision

These rules prevent the most common AI-generated formatting patterns.

**Em dashes:** Never use em dashes (—) or substitute them with hyphens (-) used as em dashes.

- For asides or clarifications, use parentheses
- For natural pauses between clauses, use commas
- ❌ "Public clusters—with authorized IP ranges—work well"
- ✅ "Public clusters (with authorized IP ranges) work well"

**Colons in running text:** Never use a colon to introduce an inline list within a sentence.

- Break the sentence after stating that something exists, then describe each item separately.
- ❌ "Azure provides two options: the Flux extension and the Argo CD extension."
- ✅ "Azure provides two managed extensions. The Flux extension offers... The Argo CD extension provides..."
- ❌ "Three approaches solve this: SOPS encrypts secrets, Sealed Secrets encrypts with cluster keys, and ESO synchronizes from Key Vault."
- ✅ "Three approaches solve this. SOPS encrypts secrets... Sealed Secrets encrypts... ESO synchronizes..."

**Bold labels with colons:** Do not use bold labels followed by a colon and inline text as pseudo-headings.

- ❌ "**ContainerLog to ContainerLogV2 migration:** All existing clusters..."
- ✅ "ContainerLogV2 migration affects all existing clusters..."
- Bold is for emphasis within sentences, not for creating inline headings.

---

### Decision-Focused Writing

When documenting choices (architecture, tooling, configuration), use a decision-focused structure rather than tutorial-style walkthroughs.

**Pattern:** Problem → Options → Trade-offs → Clear Recommendation

**Decision statements:** Conclude decision sections with a bold declaration.

- Format: `**Decision: Use X because Y.**` followed by brief reasoning.
- Place the decision statement after explaining trade-offs, not before.

**Show what matters, not every option:**

- ❌ Show full CLI commands with 20 parameters
- ✅ Explain the three characteristics that matter and why
- ❌ Tutorial-style walkthroughs of every configuration option
- ✅ Decision framework explaining when to choose each option

**Code blocks in decision-focused docs:**

- Purpose is to illustrate concepts, not provide step-by-step configuration
- One example command is enough to show syntax; do not repeat the pattern
- If showing a command, explain why the parameters matter, not just what they do

---

## Writing for Developers

1. Start with why, then how
2. Assume the reader is competent but busy
3. Explain failure modes and what typically goes wrong
4. Link instead of repeating content
5. Remove anything that does not help someone do the work
6. Distinguish between decisions that are hard to change and decisions that are easy to change
7. Acknowledge trade-offs honestly ("The catch:", "The gotcha:") rather than presenting one option as universally correct
8. Value what works in production over theoretical perfection
9. Use concrete scenarios and specific numbers instead of abstract principles

---

## Accessibility

- Logical heading order
- Descriptive link text
- Meaningful alt text
- No reliance on color alone

---

## Common Markdown Patterns

### Collapsible sections

Example:

    <details>
    <summary>Why retries are limited</summary>

    Unlimited retries caused cascading failures during load testing.
    </details>

---

### Callouts (GitHub)

Example:

    > [!NOTE]
    > This migration is irreversible.

    > [!WARNING]
    > Running this script will delete existing data.

Use sparingly and only when the callout adds clarity.

---

## markdownlint compliance

The CI pipeline runs `markdownlint-cli2` on all Markdown files. Violations fail the pull request check. Write Markdown that passes this linter from the start.

### Configuration

The project uses three markdownlint-related files:

- `.markdownlint.json` disables three rules project-wide: MD013 (line length), MD036 (emphasis used as a heading), and MD041 (first line must be a top-level heading).
- `.markdownlintignore` excludes agent and instruction files from linting: `.github/agents/**`, `.github/instructions/**`, and `.agents/**`. Files outside those paths are checked.
- CI runs via `DavidAnson/markdownlint-cli2-action` in `.github/workflows/pr-checks.yml` against all `*.md` files not excluded by the ignore file.

### Common rule violations

These are the rules that most often cause CI failures in this project.

| Rule | Description | Fix |
|------|-------------|-----|
| MD034 | Bare URLs in text | Wrap in angle brackets `<https://example.com>` or use `[text](url)` |
| MD040 | Fenced code blocks without language | Add a language identifier after the opening triple backticks |
| MD060 | Table column style inconsistency | Ensure the separator row matches column alignment consistently |

**MD034 example:**

```markdown
# Wrong
See https://example.com for details.

# Right
See <https://example.com> for details.
```

**MD040 example:**

````markdown
# Wrong
```
npm install
```

# Right
```bash
npm install
```
````

### Running locally

To run the linter before pushing:

```bash
npx markdownlint-cli2 "**/*.md" "!.github/**/*.md" "!memories/**/*.md" "!.agents/**/*.md"
```

This mirrors the CI glob pattern. Fix any violations before opening a pull request.

---

## Quality Bar Before Merging

Before finalizing documentation, confirm:

- Does this sound like a person explaining something they understand?
- Would I trust this if I did not write it?
- Can I remove 20% without losing meaning?
- Are there sentences included only to sound good?

If yes, rewrite.

---

## References

- Microsoft Writing Style Guide
  https://learn.microsoft.com/style-guide/welcome/
- ISE Engineering Playbook – Documentation
  https://microsoft.github.io/code-with-engineering-playbook/documentation/
- GitHub Markdown Guide
  https://docs.github.com/en/get-started/writing-on-github
- Wikipedia: Signs of AI Writing
  https://en.wikipedia.org/wiki/Wikipedia:Signs_of_AI_writing
