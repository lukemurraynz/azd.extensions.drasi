---
name: azure-network-security-perimeter
description: >-
  Setup, validate, and troubleshoot Azure Network Security Perimeter (NSP) with a safe Transition→Enforced rollout. USE FOR: repeatable NSP onboarding across projects, including access-rule design, diagnostics, production hardening without breaking connectivity, and using NSP logs as an observability tool for troubleshooting PaaS networking issues.
version: 1.3.0
lastUpdated: 2026-04-07
---

# Azure Network Security Perimeter (NSP) Setup, Testing, and Troubleshooting

Use this skill when a user asks to:

- Configure NSP for Azure PaaS resources
- Roll out NSP safely (Transition mode first, then Enforced)
- Validate NSP connectivity before enforcement
- Troubleshoot NSP-related access denials or regressions
- Use NSP diagnostic logs to troubleshoot PaaS networking issues (proxy/DNS, PaaS-to-PaaS, VNet integration)

## Scope and intent

This skill provides a reusable process for NSP operations across projects.

It covers:

1. NSP component model and constraints
2. Readiness and support-matrix validation
3. Initial rollout in Transition mode
4. Diagnostics and rule tuning
5. Promotion to Enforced mode
6. Break-glass and rollback procedures
7. Using NSP logs as an observability and troubleshooting tool

## Terminology note

Microsoft documentation renamed "Learning mode" to "Transition mode." The REST API still uses `Learning` as the enum value. This skill uses "Transition mode" for documentation alignment and notes the API value where relevant.

## NSP component hierarchy

Understanding the primitives is required before configuring anything.

### Network Security Perimeter (top-level resource)

- Lives under `Microsoft.Network` resource provider.
- Regional (not global). Deployed into a resource group.
- Serves as the outer container. Properties are name, location, and tags.

### Profiles (child of NSP)

- A collection of inbound and outbound access rules.
- Each NSP can have 1 or more profiles (recommended fewer than 200).
- Use separate profiles to apply different rule sets to different resources within the same NSP (for example, a "public" profile with IP rules for an externally accessed Key Vault and a "private" profile with no rules for an internal-only Key Vault).

### Access Rules (child of Profile)

- Direction: inbound or outbound.
- Inbound filters: IP address/CIDR, subscription ID.
- Outbound filters: FQDN.
- Hard limit: 200 rules per profile.
- IP prefix capacity: approximately 500 inbound and 500 outbound per profile (observed in testing, compared to 400 for the legacy service firewall).
- Future: the API schema includes placeholders for email, phone, and service tag filtering, but these are not yet usable.

### Resource Associations (child of NSP)

- Link a supported Azure resource to a specific profile with an access mode.
- Resources can be associated to an NSP in a **different subscription**.
- A resource can only be associated to **one profile in Enforced mode** per NSP.
- Three access modes exist:
  1. **Transition** (API value: `Learning`): evaluates rules and logs allow/deny decisions, then falls back to the resource's own service firewall. Default mode.
  2. **Enforced**: NSP rules override the resource's service firewall. All public traffic not explicitly allowed is blocked.
  3. **Audit** (available via API, not in portal documentation): read-only observation of traffic for resources in an NSP you do not control. Useful for security teams auditing without enforcement authority.

### Links (future, not yet functional)

- The API exposes a `Links` resource hinting at future NSP-to-NSP communication.
- This will likely resolve the centralized logging blocker described below.

## Key behavioral rules

1. **NSP rules apply only to public traffic.** Traffic through a Private Endpoint is never blocked or mediated by NSP. NSP logs do capture Private Endpoint traffic (labeled `TrafficType=Private`, matched by `DefaultAllowAll` rule), which is useful for troubleshooting.
2. **Resources within the same NSP can communicate** with each other over the Microsoft public backbone when associated with a managed identity. No additional access rules are needed for intra-NSP PaaS-to-PaaS flows.
3. **`publicNetworkAccess = SecuredByPerimeter`**: a new property value that locks down public access exclusively to NSP control (even if no NSP is associated yet). Use this for greenfield deployments where NSP-only control is intended from day one. The traditional values (`Enabled`, `Disabled`) still apply for non-NSP scenarios.

## Supported vs non-supported controls

NSP works for onboarded Private Link-capable PaaS resources (for example Key Vault, Storage, Log Analytics, AI Services/Foundry, AI Search). NSP support for Microsoft Foundry resources became GA in February 2026.

NSPs focus on service-based PaaS (where you upload data but do not control the executing code). Compute-based PaaS (AKS, App Services, Functions) is not an NSP target today.

Do not rely on NSP alone for resources not onboarded in your architecture path. For those, use resource-native controls:

- Private endpoints + private DNS
- Public network disabled where feasible
- Firewall/IP allowlists where private-only is not yet possible

### Practical support matrix policy (for every project)

Use this matrix as a starting point, but never assume support without runtime verification:

- **Typically onboarded in this solution path**: Key Vault, Storage, Log Analytics, AI Services/Foundry, AI Search.
- **May vary by subscription/feature state**: ML workspace association, preview-gated paths, or region-specific rollouts.
- **Fallback-required paths**: any resource/association that fails onboarding or association deployment.

Mandatory behavior:

1. Verify support and feature registration **before making NSP changes**.
2. If a resource/association is unsupported, do **not** block rollout of supported resources.
3. Route unsupported resources to fallback controls (private endpoints first, then service ACL/firewall).

See [references/rollout-checklist.md](references/rollout-checklist.md) for preflight verification steps.

## Diagnostic logging

NSP diagnostic settings are configured per NSP resource. Logs go to the `NSPAccessLogs` table in Log Analytics.

Each log entry includes: source resource, destination resource, operation performed, allow/deny effect, the specific rule that matched, and traffic type (Public or Private).

> [!WARNING]
> **Centralized logging blocker**: in Enforced mode, NSP blocks diagnostic log delivery from associated resources to Log Analytics Workspaces, Storage Accounts, or Event Hubs that are outside the NSP. If you use a centralized Log Analytics Workspace (most enterprises do), keep NSPs in Transition mode until cross-NSP linking becomes available. Alternatively, place the diagnostics destination in the same NSP (but this is impractical for centralized models).

NSP logs are valuable for both security and operations. Do not lock them behind a SIEM. Make them available to operations teams for day-to-day troubleshooting.

## Standard safe rollout pattern (default)

### Phase 0 — Readiness

1. Confirm target resources are NSP-onboarded.
2. Confirm current traffic paths (app → data plane, app → control plane, CI/CD paths).
3. Ensure diagnostics destination exists (Log Analytics preferred).
4. Confirm RBAC for perimeter/profile/rule/association operations.
5. For greenfield resources, consider deploying with `publicNetworkAccess = SecuredByPerimeter` from the start.

### Phase 1 — Transition mode

> [!IMPORTANT]
> Run Transition mode (API value: `Learning`) for a minimum of **14 days** before switching to Enforced mode. This ensures all legitimate traffic patterns are captured in diagnostic logs. Shorter observation periods miss infrequent but valid access patterns (e.g., monthly batch jobs, disaster recovery tests).

1. Create or reuse an NSP perimeter and profile.
2. Associate target resources with `accessMode = Learning` (Transition mode).
3. Keep existing network controls active (firewalls/private endpoints) while observing flows.
4. Run normal workload traffic and operational jobs.
5. Collect denied/allowed flow diagnostics from `NSPAccessLogs`.
6. Verify diagnostic log delivery from associated resources still reaches your centralized Log Analytics Workspace (it should, because Transition mode does not block it).

### Phase 2 — Rule tuning

1. Add explicit inbound/outbound access rules required by observed traffic.
2. Validate critical user journeys and operational dependencies.
3. Re-run traffic and ensure no unexpected denies remain.
4. For intra-NSP PaaS-to-PaaS flows, confirm resources are associated with managed identities (no additional rules needed).

### Phase 3 — Enforced mode

> [!WARNING]
> Before enabling Enforced mode, confirm your diagnostic log destinations are either in the same NSP or that you accept the centralized logging interruption until cross-NSP links are available.

1. Promote associations from Transition to Enforced.
2. Repeat smoke tests and health checks.
3. Monitor diagnostics for regressions.
4. Verify diagnostic log delivery still works for your monitoring pipeline.

### Phase 4 — Operate

1. Keep diagnostics enabled and reviewed.
2. Treat new integrations as change-managed NSP rule updates.
3. Revalidate NSP posture during each release window.
4. Use NSP logs proactively for troubleshooting (see "NSP as an observability tool" below).

## Project-specific usage for this repository

This repo already includes:

- `infrastructure/bicep/modules/nsp.bicep`
- Toggles in `infrastructure/bicep/main.bicep` and `main.bicepparam`

### Recommended toggle sequence

1. Enable NSP + Transition mode:
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
- Diagnostic log delivery from associated resources to centralized Log Analytics still works (or you have accepted the interruption).

## NSP as an observability tool

Even without enforcement, NSPs in Transition mode provide standardized network visibility that individual PaaS resource logs often lack. Consider deploying NSPs for observability alone when enforcement is not yet feasible.

### Use case: proxy/DNS misrouting

When users report they cannot reach a PaaS resource via Private Endpoint, the actual problem is often a forward web proxy resolving DNS independently and hitting the public endpoint. NSP logs (`NSPAccessLogs`) surface this immediately: an inbound entry with `TrafficType=Public` from an unexpected IP (the proxy egress) instead of the expected `TrafficType=Private` entry. This avoids hours of nslookup/curl debugging that only tests the client's own DNS, not the proxy's.

### Use case: PaaS-to-PaaS traffic discovery

PaaS-to-PaaS communication paths are often undocumented. NSP logs capture outbound calls from associated resources, surfacing source, destination, and operation. This helps identify unexpected dependencies (for example, AI Search calling Storage or Key Vault over the public backbone).

### Use case: regional VNet integration validation

Compute PaaS with regional VNet integration (App Services, API Management v2) sometimes has misconfigured outbound paths where traffic exits via the public backbone instead of the integrated VNet. NSP logs on the downstream service-based PaaS show whether traffic arrives as Public (misconfigured) or Private (correctly routed through VNet integration + Private Endpoint).

## Troubleshooting quick map

- Symptoms: intermittent 403/401, timeouts, or dependency-only failures after rollout
  - Check association mode and active profile rules.
  - Check `NSPAccessLogs` for blocked direction (inbound vs outbound) and the rule that matched.
  - Confirm private endpoint traffic path assumptions (NSP does not block PE traffic, but logs it).

- Symptoms: "works in Transition, fails in Enforced"
  - Missing explicit rule; Transition mode tolerated it via fallback controls.

- Symptoms: no effect observed
  - Resource may not be onboarded/supported for NSP in that path.

- Symptoms: diagnostic logs from associated resources stop arriving
  - Enforced mode blocks log delivery to destinations outside the NSP. Revert to Transition mode or place the destination in the same NSP.

See [references/troubleshooting.md](references/troubleshooting.md) for a detailed playbook.

## Field learnings to carry into future projects

These are recurring pitfalls observed during real rollout and troubleshooting:

1. **RBAC can hide resources and produce false confidence**

- Limited identity scope can make discovery look complete when it is only partial.
- Validate discovery counts against independent sources before sign-off.

1. **Different discovery surfaces show different coverage**

- Some resources appear in relationship-based data while not appearing in expected resource container queries.
- Use multiple evidence paths (graph relationships + management API + app-level discovery output).

1. **Runtime identity is what matters, not operator identity**

- Validate permissions for the identity used by deployed workloads (managed identity), not only your interactive CLI login.

1. **Transition mode is mandatory for low-risk rollout**

- Skipping Transition mode often causes avoidable enforcement outages due to missing explicit rules.

1. **Environment drift causes misleading outcomes**

- Ensure toggles and scripts target the deployed environment name (for example `prod` vs `dev`) before evaluating rollout results.

1. **NSP complements but does not replace service-native controls**

- Keep private endpoints, firewall allowlists, and service ACLs where NSP does not apply.

1. **Feature registration propagation can block rollout**

- Subscription feature flags may show `Pending` for extended periods; plan for propagation delay and retries.

1. **Association support can vary by resource/service path**

- Some associations may fail in a subscription even when perimeter creation succeeds. Keep associations configurable per resource and enable incrementally.

1. **Enforced mode blocks centralized diagnostic log delivery**

- If your Log Analytics Workspace, Storage Account, or Event Hub is not in the same NSP, Enforced mode will stop diagnostic log delivery from associated resources. This is the most common reason to stay in Transition mode for now.

1. **NSP rules are resilient during geo-failover**

- Per the Product Group: NSP rules persist on resources during GRS failover even if the NSP's own region is down. You cannot modify rules during that time, but enforcement continues.

1. **Outbound log entries for intra-NSP traffic may show false DenyAll**

- Observed in testing: outbound calls between resources in the same NSP that are actually allowed may be logged as hitting the `DenyAll` rule. This appears to be a logging artifact, not an actual block.

1. **No replacement for Trusted Microsoft Services exception yet**

- NSPs do not currently offer an equivalent to the "Allow Trusted Microsoft Services" toggle. For resources that depend on this, keep the service firewall with the trusted services exception enabled and leave the NSP in Transition mode. Service tag support in access rules may address this in the future.

## Definition of done for NSP change

- Transition-mode soak completed with diagnostics reviewed.
- Required access rules implemented and peer-reviewed.
- Centralized diagnostic log delivery impact assessed and accepted.
- Enforced-mode validation passed for critical user and ops paths.
- Rollback path tested and documented.
