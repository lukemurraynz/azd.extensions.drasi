# Azure SRE Agent Bundles

Bundles are modular capability packs used to build production SRE Agents
without editing the core skill file.

Use bundles to add, remove, or evolve capabilities over time.

## Why Bundles

1. Keep `SKILL.md` stable and concise.
2. Isolate capabilities by domain (AKS, Container Apps, Drasi, governance, connectors).
3. Make capability changes additive and versionable.
4. Improve reuse across multiple agents/environments.

## Bundle Layout

Each bundle folder should include:

1. `bundle.yaml` (metadata + capability manifest)
2. Optional `agents/` definitions
3. Optional `response-plans/` definitions
4. Optional `scheduled-tasks/` definitions
5. Optional `hooks/` definitions
6. Optional `connectors/` templates
7. Optional `checklists/` operational checks

## Current Bundles

- `base-core`
- `aks-production`
- `containerapps-production`
- `drasi-aks-production`
- `governance-kt`
- `connectors-observability`

See [catalog.yaml](./catalog.yaml) for bundle index and ownership.

## How to Extend

1. Create a new bundle directory with `bundle.yaml`.
2. Add only capability-specific resources (avoid duplicating unrelated resources).
3. Register the bundle in [catalog.yaml](./catalog.yaml).
4. Link it from `references/bundles-operations.md`.
5. Add acceptance tests/checklist for operational readiness.

## Naming Guidance

1. Use lowercase and hyphens only.
2. Name by outcome or domain (for example: `drasi-aks-production`).
3. Keep resources environment-neutral using placeholders.

## Upgrade Guidance

1. Prefer adding new bundles over editing many existing bundles.
2. If breaking changes are necessary, bump `version` and document migration.
3. Keep deprecated bundles readable until replacement bundles are adopted.
