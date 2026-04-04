---
description: "Guidelines for GitHub Copilot to produce self-explanatory code with minimal comments. Comments should explain intent and rationale (WHY). Public APIs should use doc comments or docstrings for behavior and usage."
applyTo: "**"
---

# Self-explanatory code commenting instructions

## Core principle

Write code that explains itself through names, types, and small functions. Prefer **no comment** unless it adds durable understanding.

**Comments are for:**

- **WHY** this exists (rationale, constraints, intent)
- **WHAT users need to know** only when producing _public API documentation_ (doc comments / docstrings)
- **Non-obvious tradeoffs** (correctness, performance, security, external quirks)

Comments should help future maintainers understand decisions that are not obvious from the code alone.

---

## Prefer these over comments (in order)

1. **Better names** (variables, functions, types, modules)
2. **Smaller units** (extract function, early returns, reduce nesting)
3. **Stronger types / invariants** (validation, domain types, enums)
4. **Tests** that demonstrate behavior and edge-cases
5. **Doc comments / docstrings** for public API usage & contracts
6. **A comment** explaining WHY / constraints / tradeoffs

---

## Decision framework (fast check)

Before writing a comment, ask:

1. **Is the code already clear?**
   - → No comment.
2. **Could a rename or extraction remove the need?**
   - → Refactor instead.
3. **Does this explain WHY or constraints, not line-by-line behavior?**
   - → Comment may be justified.
4. **Will this still be true in 6 months?**
   - → If not, don’t write it or reframe it.
5. **Is this public API surface?**
   - → Use doc comments / docstrings, not inline comments.

---

## What to comment

### Rationale / intent (WHY)

```js
// We treat "unknown" as non-fatal to keep ingestion resilient during partner outages.
if (status === "unknown") return;

// NZ payroll: holiday pay uses the higher of average weekly earnings (AWE) vs ordinary weekly pay (OWP).
const holiday_pay = max(awe, owp);

// GitHub API: avoid secondary rate limits by spacing burst calls.
await rate_limiter.wait();

// Invariant: items are sorted by created_at ascending; binary search relies on this.

// PERF: This runs on the hot path; avoid allocations by reusing the buffer.

// SECURITY: Validate and normalize user input before building the query to avoid injection.

// Using Floyd–Warshall because we need all-pairs distances (graph is small: n <= 300).

// Matches email-like strings: local@domain.tld (basic validation, not RFC-complete).
const email_pattern = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
```

### Trade-offs and gotchas

When a design choice has a non-obvious cost, surface it in the comment. Use concrete consequences, not vague warnings.

```js
// The catch: this cache never invalidates. Acceptable here because catalog data
// changes at most once per deploy, but do not reuse this pattern for user data.
const catalog = new Map();

// The gotcha: retries are unbounded. Safe for idempotent GETs against this
// internal service, but would cause cascading failures against a shared gateway.
await fetchWithRetry(url);

// This is a permanent decision. Changing the partition key later requires a full
// data migration, so pick based on the highest-cardinality access pattern.
```

### Public APIs: doc comments / docstrings

Inline comments are not a substitute for API documentation.

For public functions, classes, or modules:

- Describe purpose and behavior
- Document parameters, return values, errors, and side-effects
- Avoid internal implementation details
- Use the language-appropriate format: JSDoc for JavaScript/TypeScript, XML doc comments for C#, docstrings for Python, Javadoc for Java

```js
/**
 * Calculates compound interest.
 *
 * @param {number} principal Initial amount invested.
 * @param {number} rate Annual interest rate (0.05 == 5%).
 * @param {number} years Time in years.
 * @param {number} frequency Compounds per year.
 * @returns {number} Final amount.
 * @throws {RangeError} If inputs are negative.
 */
function compound_interest(principal, rate, years, frequency = 1) {
  // implementation
}
```

### Structured annotations

Use consistent tags so intent is searchable and reviewable:

- `NOTE`: surprising behavior or nuance
- `WARNING`: foot-gun or side-effect
- `PERF`: performance-sensitive area
- `SECURITY`: security-sensitive logic
- `DEPRECATED`: replacement and removal timeline

```js
// TODO(ABC-1234): Replace polling with webhooks once partner supports callbacks.
// HACK(DEF-987): Work around upstream bug in v2.1.0; remove after upgrade.
```

> **Rule:** Every TODO, FIXME, HACK, or DEPRECATED comment must include either a tracking reference or a clear removal condition.

---

## Comment style rules

- Use sentence case and proper punctuation
- Be specific and neutral; avoid words like “obviously” or “just”
- Do not include secrets, credentials, or sensitive customer data
- Place comments immediately above the code they describe
- If a change makes a comment inaccurate, update or delete the comment in the same change

---

## Anti-patterns to avoid

- Obvious or redundant narration
  ```js
  let counter = 0; // Initialize counter to zero
  counter++; // Increment counter by one
  ```
- Explaining WHAT instead of improving code
  - If a comment explains line-by-line behavior, refactor the code instead.
- Outdated comments
  - If a comment can easily drift from reality, remove it or encode the rule in code or tests.
- Changelog or authorship comments
  - Version control already tracks this.
  ```js
  // Modified by John on 2023-01-15
  ```
- Decorative dividers
  ```js
  //======================
  // UTILITIES
  //======================
  ```
- Commented-out code
  - Delete it and rely on version control if recovery is needed.
- Removing comments you don't understand
  - A comment that looks pointless may encode a constraint or lesson from a past bug not visible in the current diff.
  - Only remove comments when you are removing the code they describe, or you can confirm they are factually wrong.

---

## Quality checklist (before committing)

- Comment explains WHY, constraints, or tradeoffs
- Public APIs use doc comments or docstrings
- Comment is accurate and likely to remain so
- No sensitive or confidential information included
- TODO / FIXME / HACK comments include tracking or removal criteria
- If the comment existed only because the code was unclear, the code was refactored
