# Source Map (Authoritative Basis)

This map ties key guidance in this skill to authoritative sources.

## Platform and Routing

1. Incident platform behavior, single active platform, quickstart behavior:
   - https://learn.microsoft.com/en-us/azure/sre-agent/incident-platforms
2. Response plan lifecycle, turn off/on, testing mode, quickstart overlap warning:
   - https://learn.microsoft.com/en-us/azure/sre-agent/incident-response-plans
3. Run mode semantics and permission interaction:
   - https://learn.microsoft.com/en-us/azure/sre-agent/run-modes

## Scheduled Automation

1. Scheduled task controls (`Draft the cron for me`, `Polish instructions`, `Max executions` precedence):
   - https://learn.microsoft.com/en-us/azure/sre-agent/scheduled-tasks
2. Workflow design practices, `Run task now`, custom agent trigger model:
   - https://learn.microsoft.com/en-us/azure/sre-agent/workflow-automation

## Custom Agents and Skills

1. Custom agent behavior, testing, tool assignment, knowledge base limits:
   - https://learn.microsoft.com/en-us/azure/sre-agent/sub-agents
2. Skill activation model and constraints:
   - https://learn.microsoft.com/en-us/azure/sre-agent/skills

## Connectors and MCP

1. Connector categories, health states, heartbeat behavior, wildcard syntax and version notes:
   - https://learn.microsoft.com/en-us/azure/sre-agent/connectors
2. Official plugin catalog and connector-specific setup guidance:
   - https://github.com/Azure/sre-agent-plugins

## Governance with Hooks

1. Hook events (`Stop`, `PostToolUse`), formats, limits, best practices:
   - https://learn.microsoft.com/en-us/azure/sre-agent/agent-hooks
2. Hook API workflow and v2 extended-agent examples:
   - https://learn.microsoft.com/en-us/azure/sre-agent/tutorial-agent-hooks

## Real Deployment Patterns

1. Official end-to-end samples and labs:
   - https://github.com/microsoft/sre-agent
2. Hands-on lab post-provision and verification patterns:
   - https://github.com/microsoft/sre-agent/tree/main/samples/hands-on-lab
3. Deployment-compliance sample (hooks + idempotent setup + policy workflows):
   - https://github.com/microsoft/sre-agent/tree/main/samples/deployment-compliance
4. This skill's full-capability blueprint patterns:
   - `references/production-blueprints.md`
   - `hooks/kt-completeness-gate.yaml`
5. Bundle-first extension artifacts:
   - `bundles/catalog.yaml`
   - `bundles/README.md`
   - `references/bundles-operations.md`
   - `references/capability-matrix.md`

## Pricing and Regions

1. AAU billing model, always-on cost, token-based active flow, monthly allocation limit:
   - https://learn.microsoft.com/en-us/azure/sre-agent/pricing-billing
2. Supported regions (East US 2, Sweden Central, Australia East):
   - https://learn.microsoft.com/en-us/azure/sre-agent/supported-regions
3. Billing model update announcement (token-based active flow, effective April 15, 2026):
   - https://techcommunity.microsoft.com/blog/appsonazureblog/an-update-to-the-active-flow-billing-model-for-azure-sre-agent/4507866

## GA Announcement and New Capabilities

1. GA announcement (March 10, 2026):
   - https://techcommunity.microsoft.com/blog/appsonazureblog/announcing-general-availability-for-the-azure-sre-agent/4500682
2. What's new in GA release:
   - https://techcommunity.microsoft.com/blog/appsonazureblog/whats-new-in-azure-sre-agent-in-the-ga-release/4500779
3. Plugin Marketplace and Skills:
   - https://learn.microsoft.com/en-us/azure/sre-agent/skills
4. MCP connectors (including pre-configured partners):
   - https://sre.azure.com/docs/capabilities/mcp-connectors
5. Starter lab (azd-based deployment):
   - https://github.com/microsoft/sre-agent/tree/main/labs/starter-lab

## Official Resources

1. Product documentation: https://aka.ms/sreagent/docs
2. Self-paced labs: https://aka.ms/sreagent/lab
3. Technical videos: https://aka.ms/sreagent/youtube
4. Home page: https://www.azure.com/sreagent
5. X (Twitter): https://x.com/azuresreagent

## Notes

1. Community references are optional context only.
2. If a claim cannot be tied to the sources above, treat it as a local assumption and label it.
3. KT templates in this skill are derived from a user-provided workbook:
   `KT_Templates.xlsx` (`SA QU/WORKSHEET`, `PA QU/WORKSHEET`, `DA QU/WORKSHEET`, `PPA QU/WORKSHEET`).
