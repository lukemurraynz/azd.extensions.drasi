# NSP troubleshooting playbook

## High-signal checks first

1. **Association mode**
   - Verify each resource association mode (`Learning` vs `Enforced`).
2. **Rule direction mismatch**
   - Inbound deny often means source isn’t permitted.
   - Outbound deny often means destination FQDN/subscription/resource scope isn’t permitted.
3. **Resource support**
   - Ensure the failing resource path is NSP-onboarded.
4. **Private endpoint assumptions**
   - Private endpoint traffic is handled differently from public traffic controls.

## Common symptom → likely cause

- Works in Learning, fails in Enforced
  - Missing explicit NSP rule that Learning mode tolerated via fallback controls.

- Resource still accessible despite expected block
  - Resource not onboarded for NSP on that path, or traffic is not traversing the controlled path.

- Sudden 403/timeout after rule changes
  - Overly broad deny due to rule typo/scope mismatch/FQDN mismatch.

- Only a subset of expected resources/groups appear after rollout
  - Runtime identity RBAC is too narrow, or discovery surface/query path is incomplete for that resource type.
  - Validate using independent methods and compare counts before/after.

- Script reports success but rollout targets unexpected environment
  - Environment resolution drift (`dev` vs `prod`) caused toggles/role assignments to apply to a different target than expected.

- Deployment fails with "feature not yet available for given subscription"
  - NSP feature registration is incomplete (`Pending`) or not supported for a targeted association path in the subscription.

- Perimeter creates successfully but one association fails
  - Resource-specific association support can differ; proceed with supported associations and keep the failing one behind a toggle.

## Safe recovery pattern

1. Flip to Learning mode.
2. Reproduce failing flow and inspect diagnostics.
3. Add minimal explicit allow rules.
4. Re-test under Learning.
5. Re-enable Enforced and verify.

## Recovery commands (examples)

- Re-check active context and environment:
  - `az account show`
  - `azd env get-values`
- Re-run role assignment automation for runtime identities when discovery is partial.
- Re-run discovery and compare counts against pre-change baseline.

## Repository-specific reminder

In this repo, NSP is a complement for onboarded PaaS resources.
For non-onboarded or separate data-plane paths, keep using:

- Private networking patterns
- Firewall allowlists
- Service-native network ACLs
