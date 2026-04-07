---
name: docs-style
description: >-
  Enforce documentation style rules for this repository: no em dashes, no AI-isms, no promotional tone, Microsoft Writing Style Guide alignment, and README accuracy checks. USE FOR: writing or editing any Markdown file (README, ADR, runbook, template README), reviewing documentation written by agents, or catching style violations before committing.
version: 1.0.0
lastUpdated: 2026-04-07
---

# Documentation Style Rules

These rules apply whenever you write or edit any `.md` file in this repository. They are non-negotiable and override any general writing instinct.

They distill `.github/instructions/coding-standards/markdown/markdown.instructions.md` and `.github/copilot-instructions.md` into the checks that agents most commonly fail.

---

## Punctuation — Hard Rules

### Em dashes: never

Em dashes (—) are the single most common AI writing tell in this codebase. Never use them.

| Instead of | Use |
|---|---|
| `Yes — this works` | `This works.` (just say it) |
| `The tool — which is fast — handles X` | `The tool handles X. It is fast.` |
| `Install it — then run X` | `Install it, then run X.` (comma) |
| `The key concept — overlays` | `The key concept is overlays.` |

For asides, use parentheses. For natural pauses between clauses, use a comma. For emphasis, restructure the sentence.

### Colons in running prose: split the sentence instead

Do not introduce inline lists with a colon in the middle of a sentence.

```
# Wrong
Azure provides two options: Flux and Argo CD.

# Right
Azure provides two managed extensions. Flux handles GitOps workflows. Argo CD provides...
```

### Bold labels as pseudo-headings: don't

```
# Wrong
**ContainerLog migration:** All existing clusters must...

# Right
ContainerLogV2 migration affects all existing clusters...
```

---

## Tone — Hard Rules

### No promotional language

Reject any sentence that sounds like marketing copy.

Banned phrases and patterns:
- "powerful and flexible"
- "seamless experience"
- "game-changer"
- "best-in-class"
- "robust solution"
- "This aims to..."
- "leverages" (use "uses")
- "facilitates" (use "lets" or "helps")
- Abstract nouns: landscape, journey, tapestry, ecosystem (when used figuratively)

### No AI-isms

These are fingerprints of AI-generated text. Reject them:
- Excessive em dashes (covered above)
- Overuse of bold text for words that don't need emphasis
- Emojis in headings or bullet lists (unless the document is explicitly informal)
- Lists that always have exactly three items when the domain has two or four
- Closing phrases like "Let me know if you have questions" or "I hope this helps"
- Hedging every statement with "typically" or "generally" when the behavior is deterministic

### Neutral, not impressive

Documentation explains behavior. It does not sell.

```
# Wrong
This extension wraps that complexity in a powerful set of familiar azd commands.

# Right
This extension wraps that complexity in familiar azd commands.
```

---

## Structure — Hard Rules

### Sentence case for headings

```
# Wrong
## Command Reference

# Right
## Command reference
```

### No punctuation at the end of headings

```
# Wrong
## Quick start guide:

# Right
## Quick start
```

### One H1 per document

---

## README Accuracy — Checklist Before Finishing

Every time you write or edit a README, run through this checklist. Do not mark the work complete until every item passes.

- [ ] **No em dashes** — grep for `—` in the file. Zero results required.
- [ ] **No stale flags** — every `--flag` in tables and examples matches what the code actually accepts. If a flag was removed from the code, remove it from the README too.
- [ ] **No stale template names** — the `--template` flag description lists every template that exists in `internal/scaffold/templates/`. If a new template was added (e.g. `postgresql-source`), add it to the README.
- [ ] **No stale command list** — the command reference section lists every command registered in `cmd/root.go`. If `describe` was added, it appears in the README.
- [ ] **No promotional adjectives** — read each feature bullet. Remove words like "powerful", "seamless", "robust".
- [ ] **Code examples run** — every bash block in the README uses flags and subcommands that exist in the current codebase.

### How to verify flags and templates without running the binary

```bash
# Check what templates exist
ls internal/scaffold/templates/

# Check what commands are registered
grep 'AddCommand' cmd/root.go

# Check what flags a command has
grep 'Flags()' cmd/logs.go
grep 'PersistentFlags()' cmd/root.go
```

---

## Common Mistakes in This Codebase

These have happened before. Check for them explicitly.

| Mistake | Where it appeared | How to catch it |
|---|---|---|
| Em dash in "Yes — this..." | README.md line 46 | `grep '—' *.md` |
| `--follow` flag documented after it was removed | README.md logs section | Cross-check README flags against `cmd/logs.go` |
| New template not added to README | README.md init section | `ls internal/scaffold/templates/` vs `--template` description |
| New command not added to README | README.md command reference | `grep AddCommand cmd/root.go` vs README sections |

---

## Verification Command

Run this before finishing any documentation task:

```bash
# Check for em dashes
grep -r '—' *.md

# Check for common AI phrases
grep -ri 'seamless\|powerful\|robust\|game-changer\|best-in-class\|leverages' *.md
```

Both must return zero results in files you wrote or edited.
