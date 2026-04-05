# NSP rollout checklist (Learning → Enforced)

Use this checklist for repeatable, low-risk rollout.

## 1) Prepare environment

- Confirm target subscription and resource group.
- Confirm target resources are NSP-onboarded in Azure docs.
- Confirm diagnostics sink exists (Log Analytics workspace).
- Confirm operator RBAC for NSP create/update.
- Confirm **runtime managed identity** RBAC (not only operator RBAC).
- Confirm environment targeting is correct (`dev`/`staging`/`prod`) before toggling flags.
- Confirm required NSP subscription feature flags are `Registered` (not `Pending`) when applicable.

### Verify-first gate (required before any NSP change)

- Build a target resource list and classify each as:
  - `Supported and ready`
  - `Supported but feature-gated/pending`
  - `Unsupported in current subscription/region/path`
- Proceed only when this classification is complete.
- For unsupported or feature-gated items, predefine fallback controls:
  - Private endpoint + private DNS
  - Service firewall/ACL allowlist
  - Public network disabled where feasible

### Preflight commands (recommended)

- Verify active subscription:
  - `az account show`
- Verify azd environment values:
  - `azd env get-values`
- Verify runtime identity role assignments at required scopes:
  - `az role assignment list --assignee <principalId> --all`
- Verify NSP feature registration state:
  - `az feature list --namespace Microsoft.Network --query "[?contains(name, 'AllowNSPInPublicPreview') || contains(name, 'AllowNetworkSecurityPerimeter')].{name:name,state:properties.state}" -o table`
- Verify baseline resource/discovery counts before NSP changes:
  - App-level discovery endpoint response counts
  - Independent graph/management queries where applicable

## 2) Enable Learning mode (repo-specific)

Set azd environment toggles:

- `NIMBUSIQ_ENABLE_NSP=true`
- `NIMBUSIQ_NSP_MODE=Learning`

Then reprovision:

- `azd provision --environment <env> --no-prompt`

## 3) Run traffic and collect evidence

- Execute normal user journeys.
- Execute deployment/automation routines.
- Capture health checks and endpoint traces.
- Review NSP diagnostics for denied/allowed patterns.
- Capture and preserve before/after counts for discovery and critical operations.

## 4) Tune rules

- Update `infrastructure/bicep/modules/nsp.bicep` access rules.
- Re-provision and re-test.
- Repeat until diagnostics show expected allows and no unknown denies for critical flows.
- If a specific association fails due subscription capability, disable that association and continue rollout for supported resources.

## 5) Promote to Enforced mode

Set:

- `NIMBUSIQ_NSP_MODE=Enforced`

Then reprovision:

- `azd provision --environment <env> --no-prompt`

## 6) Post-enforcement validation

- API liveness/readiness checks pass.
- Critical API routes pass.
- Frontend-to-API calls pass.
- Protected resource operations pass.
- CI/CD hooks and scheduled jobs pass.
- Discovery and inventory counts are unchanged or intentionally changed (no silent drop).
- Runtime identity can still enumerate required resources and metadata.

## 7) Break-glass rollback

If production-impacting denies appear:

1. Revert mode to Learning:
   - `NIMBUSIQ_NSP_MODE=Learning`
2. Re-provision immediately.
3. Capture deny diagnostics and patch rules before re-enforcing.
