# NSP troubleshooting playbook

## High-signal checks first

1. **Association mode**
   - Verify each resource association mode (`Learning`/Transition vs `Enforced`).
   - Remember: the API uses `Learning` but documentation now calls it "Transition mode."
2. **Rule direction mismatch**
   - Inbound deny often means source isn't permitted.
   - Outbound deny often means destination FQDN/subscription/resource scope isn't permitted.
3. **Resource support**
   - Ensure the failing resource path is NSP-onboarded.
4. **Private endpoint traffic**
   - NSP does **not** block or mediate Private Endpoint traffic. It does log it (`TrafficType=Private`).
   - If you expect PE traffic but see `TrafficType=Public` in logs, the traffic is not reaching the PE (likely a DNS or proxy issue).
5. **Log table**
   - NSP logs go to the `NSPAccessLogs` table in Log Analytics. Each entry includes source, destination, operation, allow/deny, matched rule, and traffic type.

## Common symptom to likely cause

- Works in Transition, fails in Enforced
  - Missing explicit NSP rule that Transition mode tolerated via fallback controls.

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

- Diagnostic logs from associated resources stop arriving in centralized Log Analytics
  - Enforced mode blocks diagnostic log delivery to destinations outside the NSP. Revert to Transition mode or place the diagnostics destination in the same NSP.

- Outbound log entries show DenyAll for intra-NSP traffic that actually works
  - Known logging artifact: outbound calls between resources in the same NSP may be recorded as hitting `DenyAll` even though traffic is allowed. Verify by testing the actual operation, not just the log entry.

## Using NSP logs as a diagnostic tool (even without enforcement)

NSPs in Transition mode provide standardized network visibility that individual PaaS resource logs often lack. Enable NSPs for observability even when enforcement is not yet feasible.

### Proxy/DNS misrouting

**Scenario**: user reports they cannot reach a PaaS resource via Private Endpoint, but nslookup/curl from their machine resolves correctly.

**Root cause**: a forward web proxy sits between the client and the PaaS resource. The proxy resolves DNS independently using its own DNS server, which does not have conditional forwarding for `privatelink.*` domains. The proxy's connection hits the public endpoint, not the PE.

**How NSP logs help**: check `NSPAccessLogs` for entries against the target resource. If you see `TrafficType=Public` from the proxy's egress IP instead of `TrafficType=Private`, the traffic is bypassing the PE. Resolution: configure a proxy bypass for the PaaS privatelink domain.

### PaaS-to-PaaS traffic discovery

**Scenario**: unclear which PaaS services communicate with each other over the public backbone.

**How NSP logs help**: NSP logs capture outbound calls from associated resources, showing source resource, destination, and the operation performed. This surfaces undocumented dependencies (for example, AI Search calling Storage for indexer data, or Storage calling Key Vault for CMK operations).

### Regional VNet integration validation

**Scenario**: compute PaaS with regional VNet integration (App Services, API Management v2) has misconfigured outbound paths.

**How NSP logs help**: if the downstream service-based PaaS is in an NSP, logs show whether inbound traffic arrives as `TrafficType=Public` (VNet integration is misconfigured or DNS is wrong in the integrated VNet) or `TrafficType=Private` (correctly routed through the integrated VNet to a PE).

### AI workload (RAG pattern) validation

**Scenario**: Azure OpenAI/AI Foundry + AI Search + Storage in a RAG pattern. Need to verify all PaaS-to-PaaS flows are within the NSP.

**How NSP logs help**: place all three resources in the same NSP with a single profile. NSP logs surface every operation between them (chunking, indexing, embeddings, search). Inbound IP rules control human/app access. No trusted-services exception or resource-level firewall config needed for inter-service flows.

## Safe recovery pattern

1. Flip to Transition mode (API: `Learning`).
2. Reproduce failing flow and inspect `NSPAccessLogs`.
3. Add minimal explicit allow rules.
4. Re-test under Transition.
5. Re-enable Enforced and verify.

## Recovery commands (examples)

- Re-check active context and environment:
  - `az account show`
  - `azd env get-values`
- Query NSP access logs:
  - `NSPAccessLogs | where TimeGenerated > ago(1h) | where Effect == "Deny" | project TimeGenerated, SourceResourceId, DestinationFqdn, Direction, MatchedRule, TrafficType`
- Re-run role assignment automation for runtime identities when discovery is partial.
- Re-run discovery and compare counts against pre-change baseline.

## Repository-specific reminder

In this repo, NSP is a complement for onboarded PaaS resources.
For non-onboarded or separate data-plane paths, keep using:

- Private networking patterns
- Firewall allowlists
- Service-native network ACLs
