# Skills & Instructions Catalog

Quick-reference map for all skills, instructions, agents, and prompts. Use this to find the right resource for any task and understand how they connect.

---

## Instruction Files

Instructions are **always-on** rules applied automatically based on file type. They follow a hierarchy: file-specific → language-specific → global → copilot-instructions.

| Instruction                                                                                       | Applies To             | Purpose                                                                                      |
| ------------------------------------------------------------------------------------------------- | ---------------------- | -------------------------------------------------------------------------------------------- |
| [global](instructions/global.instructions.md)                                                     | All code               | Core principles, Definition of Done, testing, security, observability, accessibility         |
| [csharp](instructions/csharp.instructions.md)                                                     | `*.cs`, `*.csproj`     | DDD-lite, Data/Control Plane classification, RFC 9457, health checks, agent systems          |
| [typescript](instructions/typescript.instructions.md)                                             | `*.ts`, `*.tsx`        | Strict TypeScript, Vite env vars, SignalR reconnect, ADAC resilience, WCAG 2.2 AA, Fluent UI |
| [python](instructions/python.instructions.md)                                                     | `*.py`                 | Async/await, type hints, pytest, structured logging, dataclasses                             |
| [powershell](instructions/powershell.instructions.md)                                             | `*.ps1`, `*.psm1`      | Safe automation, dry-run, parameter design, pipeline semantics, ShouldProcess                |
| [bicep](instructions/bicep.instructions.md)                                                       | `*.bicep`              | AVM-first, API version verification, managed identity, symbolic names                        |
| [terraform](instructions/terraform.instructions.md)                                               | `*.tf`                 | AVM policy, moved/import blocks, security defaults, state isolation                          |
| [docker](instructions/docker.instructions.md)                                                     | `Dockerfile*`          | Multi-stage builds, Vite build-args, layer optimization, PostgreSQL encoding                 |
| [kubernetes](instructions/kubernetes-deployment-best-practices.instructions.md)                   | K8s manifests          | AKS guardrails, DNS labels, CORS, health probes, SignalR affinity                            |
| [yaml](instructions/yaml.instructions.md)                                                         | `*.yml`, `*.yaml`      | GitHub Actions & Azure Pipelines formatting, permissions, secret injection                   |
| [azure-devops-pipelines](instructions/azure-devops-pipelines.instructions.md)                     | ADO pipelines          | Stages, templates, security, deployment strategies, performance                              |
| [markdown](instructions/markdown.instructions.md)                                                 | `*.md`                 | Microsoft style, anti-AI fingerprints, README structure, ADR format                          |
| [self-explanatory-code-commenting](instructions/self-explanatory-code-commenting.instructions.md) | All code               | WHY not WHAT, comment decision framework, structured tags                                    |
| [cicd-security](instructions/cicd-security.instructions.md)                                       | `.github/workflows/**` | DevSecOps pipeline hardening: action pinning, OIDC, SBOM, dependency scanning, CodeQL        |

---

## Skills

Skills are **on-demand** domain knowledge loaded when a specific task is detected. Grouped by domain.

### Azure API Management Ecosystem

| Skill                                                              | When to Use                                                             | Related Skills                                       |
| ------------------------------------------------------------------ | ----------------------------------------------------------------------- | ---------------------------------------------------- |
| [azure-apim-architecture](skills/azure-apim-architecture/SKILL.md) | Architecture decisions: VNet mode, Front Door, auth strategy, cost      | → `apim-policy-authoring` → `api-security-review`    |
| [apim-policy-authoring](skills/apim-policy-authoring/SKILL.md)     | Write APIM policy XML: OAuth, JWT, rate limiting, CORS, transforms      | ← `azure-apim-architecture`, → `api-security-review` |
| [api-security-review](skills/api-security-review/SKILL.md)         | Audit APIM for OWASP API Top 10, VNet, Private Link, security controls  | ← `apim-policy-authoring`, → `apiops-deployment`     |
| [apiops-deployment](skills/apiops-deployment/SKILL.md)             | Deploy APIM via Bicep/Terraform, CI/CD pipelines, environment promotion | ← `api-security-review`, uses `azure-defaults`       |

**Workflow chain:** `azure-apim-architecture` → `apim-policy-authoring` → `api-security-review` → `apiops-deployment`

### Azure Infrastructure & Governance

| Skill                                                                                | When to Use                                                               | Related Skills                                                            |
| ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------- | ------------------------------------------------------------------------- |
| [azure-defaults](skills/azure-defaults/SKILL.md)                                     | Naming conventions, regions, tags, AVM policy, security baselines         | **Load first** before any Azure IaC skill                                 |
| [azure-role-selector](skills/azure-role-selector/SKILL.md)                           | Find minimal RBAC roles, generate role assignment Bicep, managed identity | Canonical source for role IDs; used by `bicep` & `terraform` instructions |
| [azure-deployment-preflight](skills/azure-deployment-preflight/SKILL.md)             | Validate Bicep before deploying: syntax, what-if, permissions, RBAC       | Uses `azure-defaults`; first step in `managing-azure-dev-cli-lifecycle`   |
| [azure-troubleshooting](skills/azure-troubleshooting/SKILL.md)                       | Diagnose unhealthy resources: KQL, metrics, severity, remediation         | Complements `waf-assessment`                                              |
| [azure-adr](skills/azure-adr/SKILL.md)                                               | Document architecture decisions with WAF mapping and alternatives         | Captures decisions from `azure-apim-architecture`, `waf-assessment`       |
| [waf-assessment](skills/waf-assessment/SKILL.md)                                     | Assess architecture against WAF five pillars with scoring                 | Feeds into `azure-adr`; expands on `cost-optimization`                    |
| [cost-optimization](skills/cost-optimization/SKILL.md)                               | Analyze costs, right-size, reserved instances, waste detection            | Deeper dive on WAF Cost pillar                                            |
| [private-networking](skills/private-networking/SKILL.md)                             | Private Endpoints, VNet integration, NSGs, Network Security Perimeter     | Uses `azure-defaults`; complements `identity-managed-identity`            |
| [secret-management](skills/secret-management/SKILL.md)                               | Key Vault RBAC, Key Vault references, CSI driver, secret rotation         | Uses `identity-managed-identity`, `azure-role-selector`                   |
| [azure-maps-integration](skills/azure-maps-integration/SKILL.md)                     | Azure Maps auth, Web SDK, RBAC roles, secure deployment                   | Uses `azure-role-selector`, `identity-managed-identity`                   |
| [azure-network-security-perimeter](skills/azure-network-security-perimeter/SKILL.md) | NSP setup, Learning→Enforced rollout, access rules, diagnostics           | Complements `private-networking`; uses `azure-defaults`                   |

### Infrastructure as Code

| Skill                                                                | When to Use                                                            | Related Skills                                              |
| -------------------------------------------------------------------- | ---------------------------------------------------------------------- | ----------------------------------------------------------- |
| [azure-verified-modules](skills/azure-verified-modules/SKILL.md)     | Learn from AVM patterns for custom Terraform modules (reference only)  | Patterns used by `terraform-patterns`                       |
| [terraform-patterns](skills/terraform-patterns/SKILL.md)             | Reusable Terraform patterns: hub-spoke, private endpoints, diagnostics | Uses `azure-defaults`, learns from `azure-verified-modules` |
| [terraform-security-scan](skills/terraform-security-scan/SKILL.md)   | Security scan Terraform: CIS, Azure Security Benchmark, compliance     | Validates code from `terraform-patterns`                    |
| [github-actions-terraform](skills/github-actions-terraform/SKILL.md) | Debug failing Terraform CI/CD: auth, state, plan/apply errors          | Uses `terraform-patterns` for code fixes                    |

### CI/CD & Deployment

| Skill                                                                                | When to Use                                                          | Related Skills                                                   |
| ------------------------------------------------------------------------------------ | -------------------------------------------------------------------- | ---------------------------------------------------------------- |
| [github-actions-ci-cd](skills/github-actions-ci-cd/SKILL.md)                         | GitHub Actions for .NET/React: build, ACR push, AKS deploy           | Uses `dotnet-backend-patterns`, `typescript-react-patterns`      |
| [gh-aw-operations](skills/gh-aw-operations/SKILL.md)                                 | GitHub Agentic Workflows: frontmatter, safe-outputs, MCP integration | Different from traditional `github-actions-ci-cd`                |
| [managing-azure-dev-cli-lifecycle](skills/managing-azure-dev-cli-lifecycle/SKILL.md) | azd provision/deploy/down/purge, multi-env, troubleshooting          | Orchestrates deployment; uses `azure-deployment-preflight` first |
| [azd-deployment](skills/azd-deployment/SKILL.md)                                     | azd + Container Apps: azure.yaml, Bicep infra, ACR remote builds     | Uses `managing-azure-dev-cli-lifecycle` for lifecycle management |
| [load-testing-chaos](skills/load-testing-chaos/SKILL.md)                             | Azure Load Testing, JMeter baselines, Chaos Studio experiments       | Validates `azure-container-apps`, `azure-functions-patterns`     |

### Backend & Frontend Patterns

| Skill                                                                          | When to Use                                                                   | Related Skills                                                                         |
| ------------------------------------------------------------------------------ | ----------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- |
| [dotnet-backend-patterns](skills/dotnet-backend-patterns/SKILL.md)             | ASP.NET Core: RFC 9457, EF Core migrations, SignalR hubs, health checks       | Uses `postgresql-npgsql` for data layer                                                |
| [typescript-react-patterns](skills/typescript-react-patterns/SKILL.md)         | React 19 / Fluent UI v9: strict TS, discriminated unions, SignalR, Vite, a11y | Uses `spa-endpoint-configuration` for URL injection                                    |
| [postgresql-npgsql](skills/postgresql-npgsql/SKILL.md)                         | Npgsql connection strings, EF Core in Docker, PostGIS, Drasi CDC              | Data layer for `dotnet-backend-patterns`                                               |
| [cosmos-db-patterns](skills/cosmos-db-patterns/SKILL.md)                       | Cosmos DB NoSQL: partition keys, data modelling, RBAC, change feed            | Data layer alternative to `postgresql-npgsql`; uses `identity-managed-identity`        |
| [spa-endpoint-configuration](skills/spa-endpoint-configuration/SKILL.md)       | Fix ERR_NAME_NOT_RESOLVED, Vite API URL injection at Docker build time        | Configures endpoints used by `typescript-react-patterns`                               |
| [kubernetes-cors-configuration](skills/kubernetes-cors-configuration/SKILL.md) | CORS in AKS: App Configuration, Workload Identity, ingress layer              | App-layer CORS is in `csharp.instructions`; APIM CORS in `apim-policy-authoring`       |
| [azure-portal-branding](skills/azure-portal-branding/SKILL.md)                 | Azure-portal-style admin UI: shell, blades, filters, data grids, tokens       | Design system for `typescript-react-patterns` implementations                          |
| [feature-flags](skills/feature-flags/SKILL.md)                                 | Azure App Configuration feature flags, targeting, variants, A/B testing       | Uses `identity-managed-identity`; complements `dotnet-backend-patterns`                |
| [async-request-reply](skills/async-request-reply/SKILL.md)                     | 202 Accepted + polling, Retry-After, cancellation, idempotency                | Complements `dotnet-backend-patterns`, `api-versioning-governance`                     |
| [dotnet-aspire](skills/dotnet-aspire/SKILL.md)                                 | .NET Aspire orchestration, service defaults, components, dashboard            | Uses `dotnet-backend-patterns`; orchestrates `postgresql-npgsql`, `cosmos-db-patterns` |
| [ui-design-brain](skills/ui-design-brain/SKILL.md)                             | Production-grade UI patterns from 60+ component conventions, SaaS quality     | Design reference for `typescript-react-patterns`, `azure-portal-branding`              |

### Azure Compute & Runtime

| Skill                                                                | When to Use                                                            | Related Skills                                                          |
| -------------------------------------------------------------------- | ---------------------------------------------------------------------- | ----------------------------------------------------------------------- |
| [azure-container-apps](skills/azure-container-apps/SKILL.md)         | Container Apps: scaling, Dapr, jobs, managed identity, ingress         | Uses `identity-managed-identity`, `observability-monitoring`            |
| [azure-functions-patterns](skills/azure-functions-patterns/SKILL.md) | Azure Functions: isolated worker, Durable Functions, triggers, hosting | Uses `identity-managed-identity`; `event-driven-messaging` for triggers |
| [azure-sre-agent](skills/azure-sre-agent/SKILL.md)                   | Azure SRE Agent: setup, custom agents, skills, connectors, run modes   | Complements `azure-troubleshooting`, `observability-monitoring`         |

### Cross-Cutting Concerns

| Skill                                                                  | When to Use                                                                | Related Skills                                                         |
| ---------------------------------------------------------------------- | -------------------------------------------------------------------------- | ---------------------------------------------------------------------- |
| [identity-managed-identity](skills/identity-managed-identity/SKILL.md) | Managed identity, DefaultAzureCredential, RBAC, identity-based connections | Uses `azure-role-selector`; required by most Azure skills              |
| [observability-monitoring](skills/observability-monitoring/SKILL.md)   | Application Insights, OpenTelemetry, KQL, alerts, dashboards               | Complements `azure-troubleshooting`; used by all runtime skills        |
| [event-driven-messaging](skills/event-driven-messaging/SKILL.md)       | Service Bus, Event Grid, Event Hubs: queues, topics, event routing         | Uses `identity-managed-identity`; triggers `azure-functions-patterns`  |
| [threat-modelling](skills/threat-modelling/SKILL.md)                   | STRIDE/DREAD threat analysis, trust boundaries, DFDs, risk scoring         | Feeds `azure-adr`; complements `api-security-review`, `waf-assessment` |
| [api-versioning-governance](skills/api-versioning-governance/SKILL.md) | API version lifecycle, breaking change rules, deprecation, APIM versions   | Uses `apim-policy-authoring`; complements `api-security-review`        |

### AI & Agent Patterns

| Skill                                                                  | When to Use                                                                            | Related Skills                                                                       |
| ---------------------------------------------------------------------- | -------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| [microsoft-agent-framework](skills/microsoft-agent-framework/SKILL.md) | AI agents, multi-agent workflows, tool integration, MCP/A2A protocols, agent hosting   | Uses `identity-managed-identity`, `observability-monitoring`, `azure-container-apps` |
| [maf-agentic-patterns](skills/maf-agentic-patterns/SKILL.md)           | MAF orchestration patterns: Sequential, Concurrent, Handoff, Group Chat, AG-UI/MCP/A2A | Implements patterns from `microsoft-agent-framework`                                 |
| [maf-ai-integration](skills/maf-ai-integration/SKILL.md)               | Production MAF + Foundry: streaming, grounded prompts, confidence tracking, SSE        | Uses `microsoft-agent-framework`; enforces real AI integration                       |

**Architectural governance:** Agent systems guardrails (tool contracts, prompt versioning, memory governance, replay/auditability, multi-agent coordination) live in [csharp.instructions.md — Agent Systems Extension](instructions/csharp.instructions.md). The skill covers implementation patterns; the instructions cover architectural rules.

### Specialized

| Skill                                          | When to Use                                                         | Related Skills                           |
| ---------------------------------------------- | ------------------------------------------------------------------- | ---------------------------------------- |
| [drasi-queries](skills/drasi-queries/SKILL.md) | Drasi Sources, ContinuousQueries, Reactions, Cypher syntax pitfalls | Uses `postgresql-npgsql` for CDC sources |

### Meta / Tooling

| Skill                                                            | When to Use                                                               | Related Skills                             |
| ---------------------------------------------------------------- | ------------------------------------------------------------------------- | ------------------------------------------ |
| [creating-agent-skill](skills/creating-agent-skill/SKILL.md)     | Create or review SKILL.md files, naming, structure, token budgeting       | Defines structure used by all skills       |
| [creating-devcontainers](skills/creating-devcontainers/SKILL.md) | Create/review .devcontainer configurations, lifecycle hooks, security     | Complements language-specific instructions |
| [test-scenarios](skills/test-scenarios/SKILL.md)                 | Create test scenarios from user stories: objectives, preconditions, steps | Feeds `test-validation-specialist` agent   |

### Product Strategy & Discovery

| Skill                                                                  | When to Use                                                               | Related Skills                                                 |
| ---------------------------------------------------------------------- | ------------------------------------------------------------------------- | -------------------------------------------------------------- |
| [product-vision](skills/product-vision/SKILL.md)                       | Define or refine an inspiring product vision statement                    | First step → `product-strategy` → `create-prd`                 |
| [product-strategy](skills/product-strategy/SKILL.md)                   | 9-section Product Strategy Canvas: segments, costs, value, growth, moat   | Uses `product-vision`; feeds `create-prd`, `lean-canvas`       |
| [create-prd](skills/create-prd/SKILL.md)                               | 8-section PRD: problem, objectives, segments, value props, solution, plan | Uses `product-strategy`; feeds `sprint-plan`, `pre-mortem`     |
| [product-name](skills/product-name/SKILL.md)                           | Brainstorm 5 unique product names aligned to brand and audience           | Uses `product-vision`, `value-proposition`                     |
| [lean-canvas](skills/lean-canvas/SKILL.md)                             | Lean Canvas: problem, solution, metrics, cost, UVP, unfair advantage      | Alternative to `business-model`; complements `startup-canvas`  |
| [startup-canvas](skills/startup-canvas/SKILL.md)                       | Startup Canvas: strategy (9 sections) + business model (costs + revenue)  | Combines `product-strategy` + `business-model` in one artifact |
| [business-model](skills/business-model/SKILL.md)                       | Business Model Canvas: all 9 building blocks                              | Alternative to `lean-canvas`; feeds `monetization-strategy`    |
| [opportunity-solution-tree](skills/opportunity-solution-tree/SKILL.md) | OST: map outcome → opportunities → solutions → experiments                | Uses `product-strategy`; feeds `brainstorm-experiments-*`      |
| [outcome-roadmap](skills/outcome-roadmap/SKILL.md)                     | Transform output-focused roadmap into outcome-focused strategic roadmap   | Uses `product-strategy`, `brainstorm-okrs`                     |

**Workflow chain:** `product-vision` → `product-strategy` → `create-prd` → `pre-mortem` → `sprint-plan`

### User & Customer Research

| Skill                                                                | When to Use                                                                | Related Skills                                                  |
| -------------------------------------------------------------------- | -------------------------------------------------------------------------- | --------------------------------------------------------------- |
| [user-personas](skills/user-personas/SKILL.md)                       | 3 personas from research: JTBD, pains, gains, unexpected insights          | Feeds `user-stories`, `job-stories`, `customer-journey-map`     |
| [user-stories](skills/user-stories/SKILL.md)                         | 3 C's + INVEST user stories with acceptance criteria                       | Uses `user-personas`; alternative to `job-stories`, `wwas`      |
| [job-stories](skills/job-stories/SKILL.md)                           | JTBD-format: When [situation], I want to [motivation], so I can [outcome]  | Alternative to `user-stories`; uses `user-personas`             |
| [user-segmentation](skills/user-segmentation/SKILL.md)               | Segment users from feedback by behavior, JTBD, and needs                   | Feeds `user-personas`, `ideal-customer-profile`                 |
| [customer-journey-map](skills/customer-journey-map/SKILL.md)         | End-to-end journey: stages, touchpoints, emotions, pain points             | Uses `user-personas`; identifies opportunities for `create-prd` |
| [ideal-customer-profile](skills/ideal-customer-profile/SKILL.md)     | ICP from research: demographics, behaviors, JTBD, needs                    | Feeds `beachhead-segment`; uses `user-segmentation`             |
| [beachhead-segment](skills/beachhead-segment/SKILL.md)               | First market segment: burning pain, willingness to pay, referral potential | Uses `ideal-customer-profile`; feeds `gtm-strategy`             |
| [interview-script](skills/interview-script/SKILL.md)                 | Mom Test interview guide: warm-up, JTBD probing, wrap-up (no pitching)     | Produces data for `summarize-interview`, `user-personas`        |
| [summarize-interview](skills/summarize-interview/SKILL.md)           | Structured interview summary: JTBD, satisfaction signals, action items     | Uses transcript from `interview-script`; feeds `user-personas`  |
| [analyze-feature-requests](skills/analyze-feature-requests/SKILL.md) | Analyze and prioritize feature requests by theme, impact, effort, risk     | Feeds `prioritize-features`; complements `sentiment-analysis`   |

**Workflow chain:** `interview-script` → `summarize-interview` → `user-personas` → `customer-journey-map` → `opportunity-solution-tree`

### Competitive & Market Intelligence

| Skill                                                            | When to Use                                                                  | Related Skills                                                   |
| ---------------------------------------------------------------- | ---------------------------------------------------------------------------- | ---------------------------------------------------------------- |
| [competitor-analysis](skills/competitor-analysis/SKILL.md)       | Competitive landscape: strengths, weaknesses, differentiation opportunities  | Feeds `competitive-battlecard`, `red-team`                       |
| [competitive-battlecard](skills/competitive-battlecard/SKILL.md) | Sales-ready battlecard: positioning, feature comparison, objection handling  | Uses `competitor-analysis`; alternative view to `red-team`       |
| [red-team](skills/red-team/SKILL.md)                             | Adversary role-play: vulnerability scan, counter-moves, defenses             | Uses `competitor-analysis`; complements `pressure-test`          |
| [market-sizing](skills/market-sizing/SKILL.md)                   | TAM/SAM/SOM estimation: top-down and bottom-up approaches                    | Feeds `beachhead-segment`, `gtm-strategy`                        |
| [market-segments](skills/market-segments/SKILL.md)               | 3-5 customer segments: demographics, JTBD, product fit                       | Feeds `user-personas`, `beachhead-segment`                       |
| [porters-five-forces](skills/porters-five-forces/SKILL.md)       | Industry dynamics: rivalry, suppliers, buyers, substitutes, new entrants     | Strategic input for `product-strategy`, `competitive-battlecard` |
| [pestle-analysis](skills/pestle-analysis/SKILL.md)               | Macro environment: Political, Economic, Social, Tech, Legal, Environmental   | Complements `porters-five-forces`, `swot-analysis`               |
| [swot-analysis](skills/swot-analysis/SKILL.md)                   | Strategic SWOT: strengths, weaknesses, opportunities, threats + actions      | Complements `porters-five-forces`, `pestle-analysis`             |
| [ansoff-matrix](skills/ansoff-matrix/SKILL.md)                   | Growth paths: penetration, market dev, product dev, diversification          | Feeds `product-strategy`, `gtm-strategy`                         |
| [value-proposition](skills/value-proposition/SKILL.md)           | 6-part JTBD value prop: Who, Why, What before, How, What after, Alternatives | Feeds `value-prop-statements`, `positioning-ideas`               |
| [value-prop-statements](skills/value-prop-statements/SKILL.md)   | Value prop → marketing, sales, and onboarding copy                           | Uses `value-proposition`; feeds `gtm-strategy`                   |
| [positioning-ideas](skills/positioning-ideas/SKILL.md)           | Product positioning differentiated from competitors with rationale           | Uses `competitor-analysis`, `value-proposition`                  |

**Workflow chain:** `competitor-analysis` → `competitive-battlecard` → `red-team` → `pressure-test`

### Ideation & Experimentation

| Skill                                                                              | When to Use                                                                              | Related Skills                                                 |
| ---------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- | -------------------------------------------------------------- |
| [brainstorm-ideas-new](skills/brainstorm-ideas-new/SKILL.md)                       | Feature ideas for a **new** product from PM, Designer, Engineer perspectives             | For new products; see `brainstorm-ideas-existing` for existing |
| [brainstorm-ideas-existing](skills/brainstorm-ideas-existing/SKILL.md)             | Feature ideas for an **existing** product from multi-perspective ideation                | For existing products; see `brainstorm-ideas-new` for new      |
| [brainstorm-experiments-new](skills/brainstorm-experiments-new/SKILL.md)           | Lean pretotypes for a **new** product: XYZ hypotheses, low-effort validation             | For new products; feeds `identify-assumptions-new`             |
| [brainstorm-experiments-existing](skills/brainstorm-experiments-existing/SKILL.md) | Experiment design for **existing** product: prototypes, A/B tests, spikes                | For existing products; feeds `identify-assumptions-existing`   |
| [identify-assumptions-new](skills/identify-assumptions-new/SKILL.md)               | Risky assumptions for a **new** product across 8 risk categories incl. GTM, Team         | For new products; feeds `prioritize-assumptions`               |
| [identify-assumptions-existing](skills/identify-assumptions-existing/SKILL.md)     | Risky assumptions for a **feature idea** across Value, Usability, Viability, Feasibility | For existing products; feeds `prioritize-assumptions`          |
| [prioritize-assumptions](skills/prioritize-assumptions/SKILL.md)                   | Impact × Risk matrix, suggest experiments per assumption                                 | Uses `identify-assumptions-*`; feeds experiment design         |

**Workflow chain:** `brainstorm-ideas-new` → `identify-assumptions-new` → `prioritize-assumptions` → `brainstorm-experiments-new`

### Metrics, OKRs & Prioritization

| Skill                                                                  | When to Use                                                                         | Related Skills                                                   |
| ---------------------------------------------------------------------- | ----------------------------------------------------------------------------------- | ---------------------------------------------------------------- |
| [brainstorm-okrs](skills/brainstorm-okrs/SKILL.md)                     | Team OKRs aligned with company objectives: qualitative + measurable                 | Feeds `okr-alignment-check`; uses `product-strategy`             |
| [okr-alignment-check](skills/okr-alignment-check/SKILL.md)             | Score work against OKRs, detect strategic drift, identify coverage gaps             | Uses `brainstorm-okrs`; complements `pressure-test`              |
| [north-star-metric](skills/north-star-metric/SKILL.md)                 | North Star + 3-5 input metrics, business game classification, 7-criteria validation | Feeds `metrics-dashboard`; aligns with `brainstorm-okrs`         |
| [metrics-dashboard](skills/metrics-dashboard/SKILL.md)                 | Dashboard design: key metrics, data sources, visualization, alert thresholds        | Uses `north-star-metric`; complements `observability-monitoring` |
| [cohort-analysis](skills/cohort-analysis/SKILL.md)                     | Retention curves, feature adoption, churn patterns by cohort                        | Feeds `metrics-dashboard`; complements `sentiment-analysis`      |
| [sentiment-analysis](skills/sentiment-analysis/SKILL.md)               | User feedback segments with sentiment scores, JTBD, satisfaction insights           | Complements `cohort-analysis`, `analyze-feature-requests`        |
| [ab-test-analysis](skills/ab-test-analysis/SKILL.md)                   | Statistical significance, sample size, confidence intervals, ship/stop decisions    | Uses `brainstorm-experiments-*`; feeds `metrics-dashboard`       |
| [prioritization-frameworks](skills/prioritization-frameworks/SKILL.md) | Reference: RICE, ICE, Kano, MoSCoW, Opportunity Score + 4 more with formulas        | Framework guide for `prioritize-features`                        |
| [prioritize-features](skills/prioritize-features/SKILL.md)             | Backlog prioritization: impact, effort, risk, strategic alignment, top 5            | Uses `prioritization-frameworks`; feeds `sprint-plan`            |

### Go-to-Market & Growth

| Skill                                                          | When to Use                                                                | Related Skills                                            |
| -------------------------------------------------------------- | -------------------------------------------------------------------------- | --------------------------------------------------------- |
| [gtm-strategy](skills/gtm-strategy/SKILL.md)                   | GTM plan: channels, messaging, success metrics, launch timeline            | Uses `beachhead-segment`, `value-proposition`             |
| [gtm-motions](skills/gtm-motions/SKILL.md)                     | 7 GTM motions: Inbound, Outbound, Paid, Community, Partners, ABM, PLG      | Complements `gtm-strategy`; feeds `marketing-ideas`       |
| [marketing-ideas](skills/marketing-ideas/SKILL.md)             | 5 creative, cost-effective marketing ideas with channels and messaging     | Uses `gtm-motions`; feeds campaign execution              |
| [pricing-strategy](skills/pricing-strategy/SKILL.md)           | Pricing models, competitive analysis, willingness-to-pay, price elasticity | Uses `competitor-analysis`; feeds `monetization-strategy` |
| [monetization-strategy](skills/monetization-strategy/SKILL.md) | 3-5 revenue models: audience fit, risks, validation experiments            | Uses `pricing-strategy`, `business-model`                 |
| [growth-loops](skills/growth-loops/SKILL.md)                   | 5 flywheel types: Viral, Usage, Collaboration, UGC, Referral               | Feeds `north-star-metric`; complements `gtm-strategy`     |

**Workflow chain:** `beachhead-segment` → `value-proposition` → `gtm-strategy` → `gtm-motions` → `marketing-ideas`

### Decision & Risk Governance

| Skill                                                    | When to Use                                                                      | Related Skills                                             |
| -------------------------------------------------------- | -------------------------------------------------------------------------------- | ---------------------------------------------------------- |
| [decision-record](skills/decision-record/SKILL.md)       | Team Decision Records (TD-NNNN): RAPID roles, options analysis, status lifecycle | Alternative to `azure-adr` for non-Azure decisions         |
| [pressure-test](skills/pressure-test/SKILL.md)           | Seven-lens stress test: assumptions, resources, timeline, competition, failure   | Complements `pre-mortem`; feeds `decision-record`          |
| [pre-mortem](skills/pre-mortem/SKILL.md)                 | Pre-launch risk: Tigers / Paper Tigers / Elephants, blocking vs fast-follow      | Before launch; see `post-mortem` for after incidents       |
| [post-mortem](skills/post-mortem/SKILL.md)               | Blameless incident analysis: timeline, 5 Whys, TTD/TTM/TTR, action items         | After incidents; see `pre-mortem` for before launch        |
| [escalation-tracker](skills/escalation-tracker/SKILL.md) | SLA-monitored escalation dashboard: severity S1-S4, decision rights matrix       | Tracks blockers; feeds `decision-record`                   |
| [stakeholder-map](skills/stakeholder-map/SKILL.md)       | Power/interest grid, communication strategies per quadrant                       | Feeds `escalation-tracker`, `decision-record`              |
| [retro](skills/retro/SKILL.md)                           | Sprint retrospective: Start/Stop/Continue or 4Ls or Sailboat, action items       | Captures learnings; feeds `decision-record`, `sprint-plan` |

**Workflow chain:** `pre-mortem` → `pressure-test` → `decision-record` → `escalation-tracker` (if blocked)

### Sprint & Delivery

| Skill                                          | When to Use                                                        | Related Skills                                             |
| ---------------------------------------------- | ------------------------------------------------------------------ | ---------------------------------------------------------- |
| [sprint-plan](skills/sprint-plan/SKILL.md)     | Sprint planning: capacity, story selection, dependencies, risks    | Uses `prioritize-features`; feeds `user-stories` or `wwas` |
| [wwas](skills/wwas/SKILL.md)                   | Why-What-Acceptance backlog items with strategic context           | Alternative to `user-stories`; uses `sprint-plan`          |
| [release-notes](skills/release-notes/SKILL.md) | User-facing release notes from tickets/PRDs/changelogs by category | End of sprint; uses `sprint-plan`, `create-prd`            |

### Writing & Legal

| Skill                                            | When to Use                                                             | Related Skills                |
| ------------------------------------------------ | ----------------------------------------------------------------------- | ----------------------------- |
| [grammar-check](skills/grammar-check/SKILL.md)   | Proofread text for grammar, logic, and flow errors without full rewrite | Applies to any written output |
| [draft-nda](skills/draft-nda/SKILL.md)           | Draft NDA: information types, jurisdiction, clauses for legal review    | Complements `privacy-policy`  |
| [privacy-policy](skills/privacy-policy/SKILL.md) | Draft privacy policy: data types, GDPR, jurisdiction, compliance        | Complements `draft-nda`       |

### Utilities

| Skill                                                  | When to Use                                                              | Related Skills                                               |
| ------------------------------------------------------ | ------------------------------------------------------------------------ | ------------------------------------------------------------ |
| [dummy-dataset](skills/dummy-dataset/SKILL.md)         | Generate realistic test data: CSV, JSON, SQL, Python with constraints    | Feeds `sql-queries` for testing                              |
| [sql-queries](skills/sql-queries/SKILL.md)             | Natural language → SQL: BigQuery, PostgreSQL, MySQL from schema/diagrams | Uses schema from `postgresql-npgsql` or `cosmos-db-patterns` |
| [review-resume](skills/review-resume/SKILL.md)         | PM resume review: XYZ+S formula, keyword optimization, tailoring         | Standalone utility                                           |
| [summarize-meeting](skills/summarize-meeting/SKILL.md) | Meeting transcript → structured notes: decisions, summary, action items  | Complements `summarize-interview`                            |

---

## Agents

Agents **orchestrate workflows** by combining skills and instructions for complex multi-step tasks.

| Agent                                                                    | Purpose                                                              | Related Skills/Instructions                                          |
| ------------------------------------------------------------------------ | -------------------------------------------------------------------- | -------------------------------------------------------------------- |
| [azure-architect](agents/azure-architect.agent.md)                       | WAF assessment, IaC review, cost optimization                        | `waf-assessment`, `cost-optimization`, `azure-adr`, `azure-defaults` |
| [backlog-refinement](agents/backlog-refinement.agent.md)                 | INVEST validation, t-shirt sizing, Definition of Ready               | Standalone — works with all agent outputs                            |
| [cleanup-specialist](agents/cleanup-specialist.agent.md)                 | Technical debt, code smells, duplication, maintainability            | Language instructions, `self-explanatory-code-commenting`            |
| [code-review](agents/code-review.agent.md)                               | Security, performance, maintainability, correctness review           | All language instructions, `api-security-review`                     |
| [creating-agent-skill](agents/creating-agent-skill.agent.md)             | Create/review SKILL.md files                                         | `creating-agent-skill` skill                                         |
| [creating-devcontainers](agents/creating-devcontainers.agent.md)         | DevContainer configurations                                          | `creating-devcontainers` skill                                       |
| [diagram-smith](agents/diagram-smith.md)                                 | .drawio diagrams: C4, BPMN 2.0, flowcharts as mxGraph XML            | Used by `documentation-specialist`                                   |
| [documentation-specialist](agents/documentation-specialist.agent.md)     | Diátaxis documentation, freshness validation, diagram integration    | `markdown.instructions`, `diagram-smith` agent                       |
| [security-specialist](agents/security-specialist.agent.md)               | OWASP vulnerabilities, auth review, data protection, remediation     | `api-security-review`, `terraform-security-scan`, `threat-modelling` |
| [test-validation-specialist](agents/test-validation-specialist.agent.md) | Test compilation, signature matching, mock contracts, TDD compliance | Language instructions                                                |
| [troubleshooting-specialist](agents/troubleshooting-specialist.agent.md) | Kepner-Tregoe root cause analysis, P0 fast-path, diagnostics         | `azure-troubleshooting`, `managing-azure-dev-cli-lifecycle`          |

---

## Prompts

Prompts are **one-shot triggers** that initiate specific workflows.

| Prompt                                                                                       | Purpose                                                               |
| -------------------------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| [code-review](prompts/code-review.prompt.md)                                                 | Initiate structured code review with MCP-backed checklists            |
| [refactor](prompts/refactor.prompt.md)                                                       | Analyze code for SOLID/DRY/KISS violations with before/after examples |
| [test-generation](prompts/test-generation.prompt.md)                                         | Generate tests using appropriate framework (xUnit/Jest/pytest)        |
| [ai-prompt-engineering-safety-review](prompts/ai-prompt-engineering-safety-review.prompt.md) | Safety/bias/security review of AI prompts with PyRIT-aligned analysis |

---

## Overlap Resolution: Which Layer Owns What?

### CORS Configuration

| Layer                  | Owner                                                                          | When                                            |
| ---------------------- | ------------------------------------------------------------------------------ | ----------------------------------------------- |
| APIM Gateway           | [apim-policy-authoring](skills/apim-policy-authoring/SKILL.md)                 | API Management handles all external API traffic |
| Kubernetes Ingress     | [kubernetes-cors-configuration](skills/kubernetes-cors-configuration/SKILL.md) | Direct AKS ingress without APIM                 |
| Application Middleware | [csharp.instructions](instructions/csharp.instructions.md)                     | ASP.NET Core `AddCors()` for app-level control  |

**Rule:** Configure CORS at exactly one layer. Multiple layers cause double-header issues.

### Health Checks

| Layer             | Owner                                                                                        | Scope                                     |
| ----------------- | -------------------------------------------------------------------------------------------- | ----------------------------------------- |
| Application       | [csharp.instructions](instructions/csharp.instructions.md)                                   | `/health/ready`, `/health/live` endpoints |
| Kubernetes Probes | [kubernetes instructions](instructions/kubernetes-deployment-best-practices.instructions.md) | `readinessProbe`, `livenessProbe` on pods |
| Container         | [docker.instructions](instructions/docker.instructions.md)                                   | `HEALTHCHECK` in Dockerfile               |

### Vite/SPA Build-Time Variables

| Concern                      | Owner                                                                          |
| ---------------------------- | ------------------------------------------------------------------------------ |
| Build-arg Docker mechanics   | [docker.instructions](instructions/docker.instructions.md)                     |
| TypeScript client-side usage | [typescript.instructions](instructions/typescript.instructions.md)             |
| Endpoint URL injection       | [spa-endpoint-configuration](skills/spa-endpoint-configuration/SKILL.md) skill |

### Managed Identity & RBAC

| Concern                        | Owner                                                                                       |
| ------------------------------ | ------------------------------------------------------------------------------------------- |
| Identity patterns & auth flows | [identity-managed-identity](skills/identity-managed-identity/SKILL.md) skill (patterns)     |
| Role selection & IDs           | [azure-role-selector](skills/azure-role-selector/SKILL.md) skill (canonical role catalog)   |
| Bicep role assignments         | [bicep.instructions](instructions/bicep.instructions.md) (references role-selector)         |
| Terraform role assignments     | [terraform.instructions](instructions/terraform.instructions.md) (references role-selector) |
| Key Vault & secret access      | [secret-management](skills/secret-management/SKILL.md) skill (Key Vault RBAC roles)         |

### Observability & Telemetry

| Concern                      | Owner                                                                                      |
| ---------------------------- | ------------------------------------------------------------------------------------------ |
| OpenTelemetry & App Insights | [observability-monitoring](skills/observability-monitoring/SKILL.md) skill (authoritative) |
| ADAC resilience telemetry    | Language instruction files (`csharp`, `typescript`)                                        |
| KQL & diagnostics            | [azure-troubleshooting](skills/azure-troubleshooting/SKILL.md) skill (troubleshooting KQL) |

### Networking & Security Perimeter

| Concern                    | Owner                                                                                     |
| -------------------------- | ----------------------------------------------------------------------------------------- |
| Private Endpoints & VNet   | [private-networking](skills/private-networking/SKILL.md) skill                            |
| Network Security Perimeter | [azure-network-security-perimeter](skills/azure-network-security-perimeter/SKILL.md)      |
| NSGs & subnet design       | [private-networking](skills/private-networking/SKILL.md) (network architecture standards) |
| APIM VNet integration      | [azure-apim-architecture](skills/azure-apim-architecture/SKILL.md)                        |

### Risk Analysis: Pre-mortem vs Pressure-test vs Red-team

| Skill           | Focus                       | When                                                                  |
| --------------- | --------------------------- | --------------------------------------------------------------------- |
| `pre-mortem`    | Launch risk categorization  | Before shipping: Tigers / Paper Tigers / Elephants, blocking vs track |
| `pressure-test` | Seven-lens stress analysis  | Before committing: assumptions, resources, timeline, competition      |
| `red-team`      | Adversary role-play         | Before competitive launch: counter-moves, vulnerabilities, defenses   |
| `post-mortem`   | Blameless incident analysis | After failure: timeline, 5 Whys, TTD/TTM/TTR, prevent/detect/mitigate |

**Rule:** `pre-mortem` is launch-specific risk. `pressure-test` is general strategy stress testing. `red-team` is competitor-focused. `post-mortem` is backward-looking.

### Canvas & Business Model Formats

| Skill            | Focus                                  | When                                                     |
| ---------------- | -------------------------------------- | -------------------------------------------------------- |
| `lean-canvas`    | 1-page lean startup hypothesis         | Early-stage validation, quick business model sketch      |
| `business-model` | Full Business Model Canvas (9 blocks)  | Established business, comprehensive model documentation  |
| `startup-canvas` | Strategy (9 sections) + business model | New venture combining strategy + economics in one canvas |

**Rule:** `lean-canvas` for speed, `business-model` for depth, `startup-canvas` for new ventures needing both.

### Backlog Item Formats

| Skill          | Format                                  | When                                           |
| -------------- | --------------------------------------- | ---------------------------------------------- |
| `user-stories` | As a [role], I want [goal], so [value]  | Traditional agile, role-based requirements     |
| `job-stories`  | When [situation], I want [motivation]   | JTBD context, situation-driven requirements    |
| `wwas`         | Why-What-Acceptance + strategic context | Outcome-focused items with strategic alignment |

**Rule:** Team preference. `user-stories` for role-clarity, `job-stories` for situation-clarity, `wwas` for strategic traceability.

### New vs Existing Product Skills

| Domain             | New Product Variant          | Existing Product Variant          |
| ------------------ | ---------------------------- | --------------------------------- |
| Ideation           | `brainstorm-ideas-new`       | `brainstorm-ideas-existing`       |
| Experiments        | `brainstorm-experiments-new` | `brainstorm-experiments-existing` |
| Assumption Mapping | `identify-assumptions-new`   | `identify-assumptions-existing`   |

**Rule:** `-new` variants include startup-specific categories (GTM, Team, Strategy). `-existing` variants focus on Value/Usability/Viability/Feasibility.

### Decision Records: Azure vs General

| Skill             | Scope                        | When                                                   |
| ----------------- | ---------------------------- | ------------------------------------------------------ |
| `azure-adr`       | Azure Architecture Decisions | WAF-mapped decisions about Azure services and patterns |
| `decision-record` | General Team Decisions (TD)  | Business, process, strategy, or technology decisions   |

**Rule:** Use `azure-adr` for Azure architecture. Use `decision-record` for everything else (team processes, business strategy, tool selection).

---

## ADAC Resilience Coverage (Auto-Detect → Auto-Declare → Auto-Communicate)

The ADAC triad is the resilience and observability pattern used across the stack. These are the authoritative sources and where each layer's guidance lives:

| Layer                    | Source                                                                               | What It Covers                                                                          |
| ------------------------ | ------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------- |
| **C# / .NET**            | [csharp.instructions](instructions/csharp.instructions.md)                           | Full ADAC definition, execution context, failure boundaries, error contracts, telemetry |
| **TypeScript / React**   | [typescript.instructions](instructions/typescript.instructions.md)                   | Frontend ADAC: connection state, stale data, offline mode, fetch resilience             |
| **Backend skill**        | [dotnet-backend-patterns](skills/dotnet-backend-patterns/SKILL.md)                   | Health endpoint degradation reasons, graceful degradation patterns                      |
| **Frontend skill**       | [typescript-react-patterns](skills/typescript-react-patterns/SKILL.md)               | UI connection state, stale-data thresholds, degraded-mode banners                       |
| **SPA config**           | [spa-endpoint-configuration](skills/spa-endpoint-configuration/SKILL.md)             | Runtime health indicators, data freshness, ADAC checklist                               |
| **CI/CD**                | [github-actions-ci-cd](skills/github-actions-ci-cd/SKILL.md)                         | Post-deployment ADAC validation, health body inspection                                 |
| **API gateway**          | [apim-policy-authoring](skills/apim-policy-authoring/SKILL.md)                       | Gateway-level failure detection, structured 503 responses, Retry-After                  |
| **Troubleshooting**      | [azure-troubleshooting](skills/azure-troubleshooting/SKILL.md)                       | ADAC violation diagnosis, common anti-patterns table                                    |
| **Deployment lifecycle** | [managing-azure-dev-cli-lifecycle](skills/managing-azure-dev-cli-lifecycle/SKILL.md) | ADAC as deployment principle, Definition of Done                                        |
| **Preflight**            | [azure-deployment-preflight](skills/azure-deployment-preflight/SKILL.md)             | Pre-deployment validation of ADAC contracts                                             |
| **Code review**          | [code-review agent](agents/code-review.agent.md)                                     | ADAC checklist items per language                                                       |

---

## Common Workflows

### Deploy .NET + React to AKS

1. `azure-defaults` → naming, tags, region
2. `azure-deployment-preflight` → validate Bicep
3. `managing-azure-dev-cli-lifecycle` → `azd provision`
4. `github-actions-ci-cd` → build & push images
5. `kubernetes-cors-configuration` → verify CORS
6. `spa-endpoint-configuration` → verify frontend URLs

### Build APIM API Marketplace

1. `azure-apim-architecture` → design decisions
2. `apim-policy-authoring` → implement policies
3. `api-security-review` → audit security
4. `apiops-deployment` → deploy to environments
5. `azure-adr` → document key decisions

### Terraform Azure Infrastructure

1. `azure-defaults` → naming, security baselines
2. `azure-verified-modules` → learn AVM patterns
3. `terraform-patterns` → implement infrastructure
4. `terraform-security-scan` → validate security
5. `github-actions-terraform` → CI/CD pipeline

### Troubleshoot Production Issue

1. `troubleshooting-specialist` agent → systematic diagnosis
2. `azure-troubleshooting` skill → KQL, metrics, remediation
3. `azure-sre-agent` → automate incident response with Azure SRE Agent
4. `observability-monitoring` → dashboards, alerts, OpenTelemetry traces
5. `azure-adr` → document the decision/fix

### Deploy Container Apps with Private Networking

1. `azure-defaults` → naming, tags, region
2. `identity-managed-identity` → managed identity & RBAC
3. `private-networking` → VNet, Private Endpoints, NSP
4. `azure-container-apps` → Container Apps environment, scaling, Dapr
5. `secret-management` → Key Vault for external secrets
6. `observability-monitoring` → Application Insights, alerts
7. `load-testing-chaos` → load test baselines, chaos experiments

### Build AI Agent Service

1. `azure-defaults` → naming, tags, region
2. `microsoft-agent-framework` → agent design, tool integration, workflows
3. `identity-managed-identity` → managed identity for Azure OpenAI
4. `azure-container-apps` or `azure-functions-patterns` → hosting
5. `private-networking` → Private Endpoint for Azure OpenAI
6. `observability-monitoring` → OpenTelemetry traces, model usage metrics
7. `load-testing-chaos` → load test agent endpoints

### Implement Feature Flags for Gradual Rollout

1. `feature-flags` → App Configuration store, targeting filters
2. `identity-managed-identity` → managed identity for App Configuration access
3. `dotnet-backend-patterns` → service integration patterns
4. `observability-monitoring` → flag evaluation telemetry

### Event-Driven Azure Functions

1. `azure-functions-patterns` → function design, triggers, Durable Functions
2. `event-driven-messaging` → Service Bus, Event Grid setup
3. `identity-managed-identity` → identity-based trigger connections
4. `cosmos-db-patterns` → change feed triggers, data modelling
5. `secret-management` → external API keys in Key Vault

### Secure API Lifecycle

1. `api-versioning-governance` → versioning strategy, breaking change rules
2. `azure-apim-architecture` → APIM design decisions
3. `apim-policy-authoring` → deprecation headers, version routing
4. `threat-modelling` → STRIDE analysis of API attack surface
5. `api-security-review` → OWASP API Top 10 audit
6. `azure-adr` → document versioning and security decisions

### Validate a New Product Idea

1. `lean-canvas` → quick hypothesis on problem, solution, UVP, channels
2. `identify-assumptions-new` → map risky assumptions across 8 categories
3. `prioritize-assumptions` → Impact × Risk matrix, pick top assumptions
4. `brainstorm-experiments-new` → design pretotypes to validate cheaply
5. `pressure-test` → seven-lens stress analysis before committing
6. `pre-mortem` → categorize launch risks as Tigers / Paper Tigers / Elephants
7. `okr-alignment-check` → verify alignment with team/company OKRs

### Competitive Response Analysis

1. `competitor-analysis` → strengths, weaknesses, differentiation landscape
2. `competitive-battlecard` → sales-ready comparison and objection handling
3. `red-team` → adversary role-play: counter-moves, vulnerabilities, defenses
4. `pressure-test` → stress-test your response strategy
5. `decision-record` → document the competitive response decision

### Sprint Discovery to Delivery

1. `user-personas` → build personas from research data
2. `opportunity-solution-tree` → map outcome → opportunities → solutions
3. `prioritize-features` → rank by impact, effort, risk, alignment
4. `sprint-plan` → capacity, story selection, dependencies
5. `user-stories` or `wwas` → write backlog items
6. `release-notes` → summarize what shipped

### Decision Governance

1. `stakeholder-map` → identify decision makers and communication plan
2. `pressure-test` → stress-test the proposal before deciding
3. `decision-record` → document with RAPID roles and options analysis
4. `escalation-tracker` → track blockers with SLA monitoring
5. `okr-alignment-check` → verify decision aligns with strategic goals

### Incident Learning Cycle

1. `post-mortem` → blameless analysis: timeline, 5 Whys, TTD/TTM/TTR
2. `retro` → team reflection: what worked, what didn't, action items
3. `decision-record` → document process changes as team decisions
4. `sprint-plan` → incorporate follow-up actions into next sprint
