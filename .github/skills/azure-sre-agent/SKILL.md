---
name: azure-sre-agent
description: >-
  Guide creation, configuration, and operation of Azure SRE Agent for automated
  incident response, scheduled tasks, custom agents, connectors, and governance
  hooks. USE FOR: setting up Azure SRE Agent, routing incidents with response
  plans, building scheduled automation, integrating GitHub/PagerDuty/ServiceNow/MCP,
  and operationalizing production-safe SRE workflows at sre.azure.com, including
  AKS, Container Apps, and Drasi-on-AKS operating patterns.
  Model provider preference: Azure OpenAI first; Anthropic only as explicit fallback.
compatibility: Requires Azure subscription with Owner or User Access Administrator role
---

# Azure SRE Agent Operator Playbook

Use this skill for real deployments of Azure SRE Agent. Keep this top-level file
lean and decision-oriented. Open the linked `references/` documents for deep
implementation details and copy-ready templates.

## Bundle-First Extension Model

Extend capability through `bundles/` before changing core skill guidance.

Use:

- [Bundle Operations Guide](./references/bundles-operations.md)
- [Bundle Catalog](./bundles/catalog.yaml)
- [Capability Matrix](./references/capability-matrix.md)
- [Bundle README](./bundles/README.md)

## Trigger Conditions

Use this skill when the user asks to:

- Create or operate Azure SRE Agent in Azure.
- Configure incident platforms, response plans, or scheduled tasks.
- Build custom agents, skills, or MCP connector workflows.
- Build production-grade coverage for AKS, Container Apps, and Drasi-on-AKS services.
- Standardize incident handling with Kepner-Tregoe (KT) methodology.
- Introduce governance controls (approval gates, audits, hook policies).
- Productionize existing SRE Agent proof-of-concepts.

## Model Provider Preference

When this skill recommends or configures LLM providers:

1. Use **Azure OpenAI first** for default and net-new setups.
2. Use **Anthropic only as fallback** when:
   - the user explicitly requests Anthropic/Claude, or
   - a required capability is unavailable in Azure OpenAI for the target region/subscription.
3. If fallback is used, document:
   - why Azure OpenAI was not used, and
   - a concrete path to return to Azure OpenAI.

## Non-Negotiable Platform Rules

Apply these rules in every deployment:

1. **Incident platform scope**
   - Only one incident platform can be active at a time.
   - Azure Monitor is default; switching platforms disconnects Azure Monitor.
2. **Quickstart overlap**
   - Connecting incident management can create `quickstart_handler`.
   - Delete or disable overlaps before enabling custom response plans.
3. **Run mode precedence**
   - Configure run mode on response plans and scheduled tasks.
   - Treat agent-level mode as fallback only.
4. **Response plan testing**
   - Use historical-incident test mode before production enablement.
   - Test mode is read-only.
5. **Scheduled task limits**
   - Set `Max executions` for bounded automation.
   - `Max executions` takes precedence over `End date`.
6. **Connector operations**
   - Monitor MCP connector health states and heartbeat behavior.
   - Use wildcard MCP tool assignment only when the trust boundary is acceptable.
7. **AKS and Container Apps discipline**
   - Separate diagnosis, remediation, and notification roles.
   - Require explicit evidence and rollback path before high-impact changes.
8. **KT process discipline**
   - For P1/P2 and production write actions, require KT sections (`SA`, `PA`, `DA`, `PPA`).
   - Use lightweight KT for lower-severity incidents.

See:

- [Incident + Response Plan Ops](./references/incident-platforms-response-plans.md)
- [Scheduled Task Ops](./references/scheduled-tasks.md)
- [Connectors + MCP Ops](./references/connectors-and-mcp.md)

## Billing and Cost Control Model

Azure SRE Agent uses token-based pricing for active usage, aligned with modern AI platform billing models used across Azure AI and GitHub Copilot.

Core Billing Model
Two cost components
Always-on cost
Fixed at 4 AAUs per agent-hour
Charged regardless of activity
Active usage cost
Based on tokens consumed
Driven by:
incident investigations
response plan execution
scheduled tasks
MCP tool interactions
AAU (Azure Agent Unit)
Standard billing unit for agent activity
Token consumption maps to AAU usage
Critical Cost Risks

Treat the following as production risks, not just cost concerns:

Autonomous execution loops
Recursive or repeated reasoning chains can exponentially increase token usage
High-context workflows
KT methodology (SA/PA/DA/PPA) significantly increases prompt size
MCP wildcard tool access
Broad tool access can trigger unnecessary or repeated calls
Scheduled task misconfiguration
Missing Max executions can result in unbounded cost
Mandatory Cost Controls

Apply these in all deployments:

Budget enforcement
Set AAU budget per agent
Use auto-suspension as a safety boundary
Execution discipline
Default all:
response plans → Review
scheduled tasks → bounded runs
Promote to Autonomous only after cost validation
Prompt efficiency
Minimize unnecessary context in:
KT templates
incident summaries
Prefer structured data over verbose text
Connector scoping
Avoid MCP wildcard (*) unless justified
Limit tool surface area to reduce unnecessary calls

## Safe Default Rollout

Use this order unless the user asks for a different rollout:

1. **Provision**
   - Create agent in supported region with managed identity and observability resources.
   - Register `Microsoft.App` provider if needed.
2. **Baseline access**
   - Start with Reader-oriented access and expand to write roles intentionally.
3. **Connect incident source**
   - Configure incident platform.
   - Remove quickstart overlap before custom routing.
4. **Attach focused workers**
   - Build custom agents with minimal tools per agent role.
   - Keep execution boundaries clear (diagnose vs remediate vs notify).
5. **Configure triggers**
   - Add response plans and scheduled tasks in Review mode first.
   - Move proven paths to Autonomous selectively.
6. **Validation**
   - Test in playground.
   - Use historical incident testing and "Run task now."
7. **Operational hardening**
   - Add hooks for approvals/audit.
   - Apply token least-privilege and rotation standards.

Deep guidance:

- [Bundle Operations Guide](./references/bundles-operations.md)
- [Bundle Catalog](./bundles/catalog.yaml)
- [Production Full-Capability Blueprints](./references/production-blueprints.md)
- [Deployment Patterns From Official Samples](./references/deployment-patterns.md)
- [Hooks and Governance Controls](./references/hooks-governance.md)
- [Connector Token Security](./references/connector-token-security.md)
- [AKS + Container Apps Production Playbook](./references/aks-containerapps-production.md)
- [Drasi on AKS Playbook Template](./references/drasi-aks-playbook.md)
- [KT Methodology Overlay](./references/kt-methodology.md)
- [KT Worksheets and Templates](./references/kt-templates.md)

## Design Defaults for Real Deployments

Use these defaults unless the user requests otherwise:

1. **Memory-first triage**
   - Search memory for similar incidents before new diagnostics.
2. **Structured incident reporting**
   - Use consistent sections: Summary, Impact, Timeline, Evidence, Root Cause, Remediation, Action Items, References.
3. **Idempotent setup scripts**
   - Prefer upsert behavior and safe re-runs for post-provision steps.
4. **Eventual consistency aware**
   - Add bounded retries for APIs that may not be immediately ready.
5. **Verification pass required**
   - Verify all configured objects (KB, subagents, connectors, response plans, tasks) before handoff.
6. **Environment-neutral examples**
   - Never hardcode personal email, tenant-specific IDs, or one-off resource names in reusable templates.
7. **KT severity policy**
   - Full KT (`SA -> PA -> DA -> PPA`) for P1/P2 incidents.
   - Lightweight KT for P3/P4 unless elevated by risk.
8. **Bundle-first capability changes**
   - Add or modify capability in bundles, not in core skill body.
   - Register new capability packs in `bundles/catalog.yaml`.

## Reference Routing

Open only what is needed:

| If the user asks about... | Open this file |
| --- | --- |
| Incident platform setup, response plans, run mode behavior, quickstart conflicts | [references/incident-platforms-response-plans.md](./references/incident-platforms-response-plans.md) |
| Task schedules, cron drafting, instruction quality, bounded runs | [references/scheduled-tasks.md](./references/scheduled-tasks.md) |
| MCP health states, wildcard assignment, connector troubleshooting | [references/connectors-and-mcp.md](./references/connectors-and-mcp.md) |
| How to add/remove capabilities without touching core skill | [references/bundles-operations.md](./references/bundles-operations.md) |
| Bundle registry and capability map | [bundles/catalog.yaml](./bundles/catalog.yaml) |
| Bundle composition by use case | [references/capability-matrix.md](./references/capability-matrix.md) |
| End-to-end production implementation patterns using full platform capability | [references/production-blueprints.md](./references/production-blueprints.md) |
| Idempotent setup, retries, cleanup of defaults, post-setup verification | [references/deployment-patterns.md](./references/deployment-patterns.md) |
| Stop/PostToolUse hooks, approval gates, audit patterns, hook safety defaults | [references/hooks-governance.md](./references/hooks-governance.md) |
| Ready-to-apply KT Stop hook artifact | [hooks/kt-completeness-gate.yaml](./hooks/kt-completeness-gate.yaml) |
| PagerDuty/Grafana/Dynatrace token least-privilege, rotation, ownership | [references/connector-token-security.md](./references/connector-token-security.md) |
| AKS and Container Apps production setup and operating defaults | [references/aks-containerapps-production.md](./references/aks-containerapps-production.md) |
| Drasi workloads running on AKS (incident and task templates) | [references/drasi-aks-playbook.md](./references/drasi-aks-playbook.md) |
| KT process rules and severity-based application | [references/kt-methodology.md](./references/kt-methodology.md) |
| Copy-ready KT worksheets (`SA`, `PA`, `DA`, `PPA`) | [references/kt-templates.md](./references/kt-templates.md) |
| Traceability from guidance to authoritative sources | [references/source-map.md](./references/source-map.md) |

## YAML Guidance (Normalized)

Use these compatibility rules:

1. **Modes**
   - Set `Review`/`Autonomous` on response plans and scheduled tasks.
   - Do not rely on custom agent YAML `agent_type` to control runtime mode.
2. **Hooks**
   - Hook configuration uses v2 extended-agent APIs/capabilities.
   - Agent Canvas YAML view can omit hook details; manage hooks in Builder > Hooks or v2 APIs.
3. **MCP tools**
   - Individual tools: precise control.
   - Wildcard `{connection-id}/*`: broad trust and automatic inclusion of future tools.

## Output Standards

When producing deployment recommendations or runbooks with this skill:

1. Include explicit mode and permission assumptions.
2. Call out blast radius for any write action.
3. Include rollback or disable path (`Turn off` plan, pause task, revert config).
4. Include verification steps and expected evidence.
5. Separate facts from assumptions.
6. When KT is required, include `Situation Appraisal`, `Problem Analysis`, `Decision Analysis`, and `Potential Problem Analysis` headings explicitly.

## Source and Currency Policy

Authority priority:

1. Microsoft Learn (`learn.microsoft.com/azure/sre-agent/*`)
2. Official repositories:
   - `microsoft/sre-agent`
   - `Azure/sre-agent-plugins`
3. Community sources only as optional context, never as primary authority.

Currency checks:

- Re-verify core behavior quarterly or before production cutover:
  - incident platform behavior
  - run mode semantics
  - connector status model
  - hook schema/API behavior
  - available regions and pricing

For source traceability, see [references/source-map.md](./references/source-map.md).

## Related Skills

- [Azure Troubleshooting](../azure-troubleshooting/SKILL.md) for ad-hoc diagnostics and KQL workflows.
- [Cost Optimization](../cost-optimization/SKILL.md) for spend-focused remediation and ROI planning.
- [Post Mortem](../post-mortem/SKILL.md) for incident retrospective and follow-up actions.
