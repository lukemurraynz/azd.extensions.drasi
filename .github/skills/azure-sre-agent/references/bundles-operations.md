# Bundle Operations Guide

Use this guide to evolve capabilities without changing the core skill.

## Bundle-First Rule

When adding or removing capability:

1. Update bundles first.
2. Update bundle catalog second.
3. Update core skill routing only if navigation changes.

This keeps core `SKILL.md` stable and prevents capability sprawl.

## Standard Bundle Components

A capability bundle can include:

1. `agents/`
2. `response-plans/`
3. `scheduled-tasks/`
4. `hooks/`
5. `connectors/`
6. `checklists/`

## Capability Assembly Pattern

Assemble production agents by composing bundles:

1. Start with `base-core`.
2. Add one or more domain bundles:
   - `aks-production`
   - `containerapps-production`
   - `drasi-aks-production`
3. Add policy bundles:
   - `governance-kt`
4. Add integration bundles:
   - `connectors-observability`

## Change Workflow

1. Introduce new capability in a new or existing bundle.
2. Add/adjust `bundle.yaml` metadata.
3. Register in `bundles/catalog.yaml`.
4. Add/update acceptance checks.
5. Test with:
   - historical incident simulation
   - scheduled task "Run task now"
   - hook behavior validation

## Versioning Guidance

1. Patch version: non-breaking tweaks.
2. Minor version: additive new resources.
3. Major version: breaking behavior changes.

Document breaking changes in bundle notes.

## Bundle Validation Checklist

Before marking a bundle production-ready:

1. Resource names are environment-neutral.
2. No hardcoded personal IDs/emails/tokens.
3. At least one test scenario is documented.
4. Required dependencies are declared in `bundle.yaml`.
5. Rollback path is documented for write actions.

## Practical Outcome

This model lets you:

1. create new agents quickly by composing bundles,
2. add capability without editing `SKILL.md`, and
3. keep long-term maintenance manageable as scope expands.

For composition examples, see [capability-matrix.md](./capability-matrix.md).
