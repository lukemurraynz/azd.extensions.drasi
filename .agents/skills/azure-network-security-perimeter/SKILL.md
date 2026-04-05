---
name: azure-network-security-perimeter
description: >-
  Setup, validate, and troubleshoot Azure Network Security Perimeter (NSP) with a safe Learning→Enforced rollout. USE FOR: repeatable NSP onboarding across projects, including access-rule design, diagnostics, and production hardening without breaking connectivity.
version: 1.2.0
lastUpdated: 2026-03-14
---

# Azure Network Security Perimeter (NSP) Setup, Testing, and Troubleshooting

Use this skill when a user asks to:

- Configure NSP for Azure PaaS resources
- Roll out NSP safely (Learning/Transition mode first)
- Validate NSP connectivity before enforcement
- Troubleshoot NSP-related access denials or regressions

## Scope and intent

This skill provides a reusable process for NSP operations across projects.

It covers:

1. Readiness and support-matrix validation
2. Initial rollout in Learning mode
3. Diagnostics and rule tuning
4. Promotion to Enforced mode
5. Break-glass and rollback procedures

## Supported vs non-supported controls

NSP works for onboarded Private Link-capable PaaS resources (for example Key Vault, Storage, Log Analytics, and AI services as supported in Azure docs).

Do not rely on NSP alone for resources not onboarded in your architecture path. For those, use resource-native controls:

- Private endpoints + private DNS
- Public network disabled where feasible
- Firewall/IP allowlists where private-only is not yet possible

### Practical support matrix policy (for every project)

Use this matrix as a starting point, but never assume support without runtime verification:

- **Typically onboarded in this solution path**: Key Vault, Storage, Log Analytics, AI Services/Foundry.
- **May vary by subscription/feature state**: ML workspace association, preview-gated paths, or region-specific rollouts.
- **Fallback-required paths**: any resource/association that fails onboarding or association deployment.

Mandatory behavior:

1. Verify support and feature registration **before making NSP changes**.
2. If a resource/association is unsupported, do **not** block rollout of supported resources.
3. Route unsupported resources to fallback controls (private endpoints first, then service ACL/firewall).

See [references/rollout-checklist.md](references/rollout-checklist.md) for preflight verification steps.

## Standard safe rollout pattern (default)

### Phase 0 — Readiness

1. Confirm target resources are NSP-onboarded.
2. Confirm current traffic paths (app → data plane, app → control plane, CI/CD paths).
3. Ensure diagnostics destination exists (Log Analytics preferred).
4. Confirm RBAC for perimeter/profile/rule/association operations.

### Phase 1 — Learning mode

> [!IMPORTANT]
> Run Learning Mode for a minimum of **14 days** before switching to Enforced mode. This ensures all legitimate traffic patterns are captured in diagnostic logs. Shorter observation periods miss infrequent but valid access patterns (e.g., monthly batch jobs, disaster recovery tests).

1. Create or reuse an NSP perimeter and profile.
2. Associate target resources with `accessMode = Learning`.
3. Keep existing network controls active (firewalls/private endpoints) while observing flows.
4. Run normal workload traffic and operational jobs.
5. Collect denied/allowed flow diagnostics.

### Phase 2 — Rule tuning

1. Add explicit inbound/outbound access rules required by observed traffic.
2. Validate critical user journeys and operational dependencies.
3. Re-run traffic and ensure no unexpected denies remain.

### Phase 3 — Enforced mode

1. Promote associations from Learning to Enforced.
2. Repeat smoke tests and health checks.
3. Monitor diagnostics for regressions.

### Phase 4 — Operate

1. Keep diagnostics enabled and reviewed.
2. Treat new integrations as change-managed NSP rule updates.
3. Revalidate NSP posture during each release window.

## Project-specific usage for this repository

This repo already includes:

- `infrastructure/bicep/modules/nsp.bicep`
- Toggles in `infrastructure/bicep/main.bicep` and `main.bicepparam`

### Recommended toggle sequence

1. Enable NSP + Learning mode:
   - `NIMBUSIQ_ENABLE_NSP=true`
   - `NIMBUSIQ_NSP_MODE=Learning`
2. Provision and observe traffic.
3. Add/adjust rules in `modules/nsp.bicep`.
4. Promote:
   - `NIMBUSIQ_NSP_MODE=Enforced`

See [references/rollout-checklist.md](references/rollout-checklist.md) for command-level steps.

## Validation checklist (must pass before Enforced)

- Resource health endpoints are healthy.
- App read/write paths to protected resources succeed.
- CI/CD and platform automation still work.
- No unexplained deny events in diagnostics during representative traffic.
- Break-glass rollback procedure is documented and tested.

## Troubleshooting quick map

- Symptoms: intermittent 403/401, timeouts, or dependency-only failures after rollout
  - Check association mode and active profile rules.
  - Check diagnostics for blocked direction (inbound vs outbound).
  - Confirm private endpoint traffic path assumptions.

- Symptoms: “works in Learning, fails in Enforced”
  - Missing explicit rule; Learning tolerated it via fallback controls.

- Symptoms: no effect observed
  - Resource may not be onboarded/supported for NSP in that path.

See [references/troubleshooting.md](references/troubleshooting.md) for a detailed playbook.

## Field learnings to carry into future projects

These are recurring pitfalls observed during real rollout and troubleshooting:

1. **RBAC can hide resources and produce false confidence**

- Limited identity scope can make discovery look complete when it is only partial.
- Validate discovery counts against independent sources before sign-off.

2. **Different discovery surfaces show different coverage**

- Some resources appear in relationship-based data while not appearing in expected resource container queries.
- Use multiple evidence paths (graph relationships + management API + app-level discovery output).

3. **Runtime identity is what matters, not operator identity**

- Validate permissions for the identity used by deployed workloads (managed identity), not only your interactive CLI login.

4. **Learning mode is mandatory for low-risk rollout**

- Skipping Learning mode often causes avoidable enforcement outages due to missing explicit rules.

5. **Environment drift causes misleading outcomes**

- Ensure toggles and scripts target the deployed environment name (for example `prod` vs `dev`) before evaluating rollout results.

6. **NSP complements but does not replace service-native controls**

- Keep private endpoints, firewall allowlists, and service ACLs where NSP does not apply.

7. **Feature registration propagation can block rollout**

- Subscription feature flags may show `Pending` for extended periods; plan for propagation delay and retries.

8. **Association support can vary by resource/service path**

- Some associations may fail in a subscription even when perimeter creation succeeds. Keep associations configurable per resource and enable incrementally.

## Definition of done for NSP change

- Learning-mode soak completed with diagnostics reviewed.
- Required access rules implemented and peer-reviewed.
- Enforced-mode validation passed for critical user and ops paths.
- Rollback path tested and documented.
