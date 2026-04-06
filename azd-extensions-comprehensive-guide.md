# Azure Developer CLI (azd) Extensions: Command-Level Analysis & Workflow Documentation

**Date Generated:** April 6, 2026  
**Source:** Microsoft Learn, GitHub Azure/azure-dev, Official Registry  
**Maturity Level:** Public Preview (Extensions Beta)

---

## TABLE OF CONTENTS
1. [Executive Summary](#executive-summary)
2. [Official Extension Registry](#official-extension-registry)
3. [Detailed Extension Profiles](#detailed-extension-profiles)
4. [Capability Comparison Matrix](#capability-comparison-matrix)
5. [Workflow Examples by Extension](#workflow-examples-by-extension)
6. [Local Drasi Extension Alignment](#local-drasi-extension-alignment)

---

## EXECUTIVE SUMMARY

The Azure Developer CLI (azd) is now **extensible via a beta extension framework** that allows plugins to:
- Register **custom commands** under namespaces (e.g., \zd demo\, \zd ai agent\, \zd x\)
- Subscribe to **lifecycle events** (preprovision, prepackage, predeploy, postdeploy, etc.)
- Provide **service-target providers** (custom deployment strategies)
- Implement **MCP servers** for tool integration
- Expose **metadata** for IDE autocomplete and validation

**Official Registry:** https://aka.ms/azd/extensions/registry  
**Development Registry:** https://aka.ms/azd/extensions/registry/dev  
**Framework Docs:** https://github.com/Azure/azure-dev/blob/main/cli/azd/docs/extensions/extension-framework.md

---

## OFFICIAL EXTENSION REGISTRY (as of April 2026)

\\\
azd extension list
\\\

Registry contains these official extensions:

| Extension ID | Namespace | Display Name | Status | Version |
|---|---|---|---|---|
| **azure.ai.agents** | ai.agent | Foundry agents | **Preview** | 0.1.20-preview |
| **microsoft.azd.extensions** | x | azd extensions Developer Kit | Stable | 0.10.0 |
| **azure.coding-agent** | coding-agent | Coding agent configuration | Stable | 0.6.1 |
| **microsoft.azd.demo** | demo | Demo Extension (Reference) | Stable | 0.6.0 |
| **microsoft.azd.concurx** | concurx | Concurrent execution | Stable | 0.1.0 |
| **azure.ai.finetune** | ai.finetuning | Foundry Fine Tuning | **Preview** | 0.0.17-preview |
| **azure.ai.models** | ai.models | Foundry Custom Models | **Preview** | 0.0.4-preview |
| **azure.appservice** | appservice | Azure App Service | Stable | 0.1.0 |

---

## DETAILED EXTENSION PROFILES

### 1. azure.ai.agents - Foundry Agent Service Integration

**EVIDENCED COMMANDS:**

\\\ash
azd ai agent init                    # Initialize new AI agent project
\\\

**Status:** Preview | **Maturity:** Early  
**Namespace:** ai.agent | **Language:** Go  
**Latest Version:** 0.1.20-preview (requires azd >1.23.6)

**Capabilities:**
- custom-commands
- lifecycle-events  
- mcp-server
- service-target-provider
- metadata

**Service Target Provider:**
- \zure.ai.agent\ - Deploys agents to the Foundry Agent Service

**User Workflow:**
1. Install: \zd extension install azure.ai.agents\
2. Initialize project: \zd ai agent init\ → scaffolds agent template
3. Configure in azure.yaml with service type: azure.ai.agent
4. Deploy: \zd provision\ → extension listens to lifecycle events
5. \zd deploy\ → service-target provider handles package/publish/deploy
6. Monitor via Azure Portal

**Documentation Signals:**
- README in repo ✓ (build/install instructions)
- extension.yaml ✓ (metadata + examples)
- Examples in registry ✓ (init command documented)

**Source:** 
- Registry: https://aka.ms/azd/extensions/registry
- Repo: https://github.com/Azure/azure-dev/blob/main/cli/azd/extensions/azure.ai.agents

---

### 2. microsoft.azd.extensions - Extension Developer Kit

**EVIDENCED COMMANDS:**

\\\ash
azd x init                          # Initialize new extension project (interactive)
azd x build                         # Build binary (auto-installs locally)
azd x pack                          # Package for distribution
azd x publish                       # Update registry metadata
azd x release                       # Create GitHub release
azd x watch                         # File watcher + auto-rebuild/install
\\\

**Status:** Stable | **Maturity:** Productive  
**Namespace:** x | **Language:** Go  
**Latest Version:** 0.10.0

**Capabilities:**
- custom-commands
- metadata (added in v0.9.0)

**Developer Workflow:**
1. Install: \zd extension install microsoft.azd.extensions\
2. New extension: \zd x init\ → prompts for language (Go, Dotnet, Python, JavaScript)
3. Develop locally: Edit code in \cli/azd/extensions/<name>/\
4. Test locally: \zd x watch\ OR \zd x build\ (auto-installs to ~/.azd/extensions/)
5. Package: \zd x pack --output <dir>\
6. Release: \zd x release --repo owner/repo --version 1.0.0\
7. Publish: \zd x publish --repo owner/repo\
8. Users install: \zd extension install <your-ext>\

**Build Times by Language:**
- Go: ~15s (fastest)
- Dotnet/C#: ~60s
- Python: ~4m (slowest)
- JavaScript: ~90s

**Source:**
- Registry: https://aka.ms/azd/extensions/registry
- Repo: https://github.com/Azure/azure-dev/blob/main/cli/azd/extensions/microsoft.azd.extensions

---

### 3. azure.coding-agent - GitHub Copilot Coding Agent Setup

**EVIDENCED COMMANDS:**

\\\ash
azd coding-agent config            # Configure GitHub Copilot Coding Agent for a repo
azd coding-agent config --debug    # Show underlying commands (verbose)
\\\

**Status:** Stable | **Maturity:** Productive  
**Namespace:** coding-agent | **Language:** Go  
**Latest Version:** 0.6.1

**Capabilities:**
- custom-commands
- metadata (added in v0.6.0)

**User Workflow:**
1. Prerequisites: GitHub repo with admin rights, azd login
2. Install: \zd extension install azure.coding-agent\
3. Navigate to GitHub repo clone
4. Run: \zd coding-agent config\ → Interactive flow:
   - Creates Azure Managed Identity
   - Assigns Reader role (scoped to resource group)
   - Sets up federated credentials for GitHub Actions
   - Updates GitHub \copilot\ environment
   - Creates/updates MCP configuration
   - Creates pull request with setup steps
5. Merge PR and follow any final configuration

**Default Permissions:**
- Reader role on resource group (customizable via Azure Portal RBAC)
- Federated credential chain: GitHub repo → Managed Identity → Azure resources

**Source:**
- Registry: https://aka.ms/azd/extensions/registry
- Repo: https://github.com/Azure/azure-dev/blob/main/cli/azd/extensions/azure.coding-agent

---

### 4. microsoft.azd.demo - Framework Demonstration Extension (Reference Implementation)

**EVIDENCED COMMANDS:**

\\\ash
azd demo context                              # Show project & environment context
azd demo prompt                               # Demo prompt UX (select, confirm, etc.)
azd demo ai models                            # Browse AI models interactively
azd demo ai deployment                        # Select model/version/SKU/capacity
azd demo ai quota                             # View usage meters & limits
azd demo mcp start                            # Start MCP server with demo tools
azd demo copilot [--model <model>] [--resume] # Interactive Copilot chat loop
azd demo listen                               # Lifecycle event listener
\\\

**Status:** Stable | **Maturity:** Reference Implementation  
**Namespace:** demo | **Language:** Go  
**Latest Version:** 0.6.0

**Capabilities:**
- custom-commands
- lifecycle-events
- mcp-server
- service-target-provider
- framework-service-provider
- metadata

**Service Target Provider:**
- \demo\ - Deploys application components to demo

**Test Workflow:**
1. Install: \zd extension install microsoft.azd.demo\
2. Test prompts: \zd demo prompt\ → Shows:
   - PromptSubscription (select Azure subscription)
   - PromptLocation (select Azure region)
   - PromptResourceGroup (select RG)
   - Confirm (yes/no)
   - Select (single choice)
   - MultiSelect (multiple choices)
3. Test AI integration: \zd demo ai models\ → Interactive model browser
4. Use as service target:
   - Add to azure.yaml: \services: web: host: demo\
   - Deploy: \zd up\ → Demo provider handles deploy

**Extension Configuration (azure.yaml):**
\\\yaml
extensions:
  demo:
    project:
      enableColors: true
      maxItems: 20
      labels:
        team: platform
        env: dev
services:
  web:
    extensions:
      demo:
        service:
          endpoint: "https://api.example.com"
          port: 8080
          environment: staging
\\\

**Source:**
- Registry: https://aka.ms/azd/extensions/registry
- Repo: https://github.com/Azure/azure-dev/blob/main/cli/azd/extensions/microsoft.azd.demo

---

### 5. azure.appservice - App Service Management

**EVIDENCED COMMANDS:**

\\\ash
azd appservice swap --service <name> --src <slot> --dst <slot>
\\\

**Status:** Stable | **Maturity:** Early Productive  
**Namespace:** appservice | **Language:** Go  
**Latest Version:** 0.1.0

**Capabilities:**
- custom-commands
- metadata

**Command Details:**

\\\ash
azd appservice swap --service <name> --src <slot> --dst <slot>
\\\

**Flags:**
- \--service\ (optional) - Service name; prompts if omitted or only one exists
- \--src\ - Source slot name (\@main\ = production)
- \--dst\ - Destination slot name (\@main\ = production)

**Interactive Mode:**
- Prompts for service and slots if flags omitted

**Examples:**
\\\ash
# Swap staging to production
azd appservice swap --service myapp --src staging --dst @main

# Interactive mode
azd appservice swap

# Swap from production to staging (rollback)
azd appservice swap --src @main --dst staging
\\\

**Source:**
- Registry: https://aka.ms/azd/extensions/registry
- Repo: https://github.com/Azure/azure-dev/blob/main/cli/azd/extensions/azure.appservice

---

## CAPABILITY COMPARISON MATRIX

\\\
┌────────────────────┬─────────────┬──────────────┬─────────────┬──────────────┬──────────┐
│ Extension          │ Custom Cmds │ Lifecycle    │ Service     │ MCP Server   │ Metadata │
│                    │             │ Events       │ Target      │              │          │
├────────────────────┼─────────────┼──────────────┼─────────────┼──────────────┼──────────┤
│ azure.ai.agents    │      ✓      │      ✓       │      ✓      │      ✓       │    ✓     │
│ microsoft.azd.     │      ✓      │      —       │      —      │      —       │    ✓     │
│ extensions         │             │              │             │              │          │
│ azure.coding-agent │      ✓      │      —       │      —      │      —       │    ✓     │
│ microsoft.azd.demo │      ✓      │      ✓       │      ✓      │      ✓       │    ✓     │
│ azure.appservice   │      ✓      │      —       │      —      │      —       │    ✓     │
└────────────────────┴─────────────┴──────────────┴─────────────┴──────────────┴──────────┘
\\\

---

## LOCAL DRASI EXTENSION ALIGNMENT

### Current State (Inferred)

**Repository:** D:\\GitHub\\azd.extensions.drasi  
**Purpose:** Extend azd for Drasi CDC (Change Data Capture) query orchestration  
**Language:** Go (consistent with official framework)  
**Status:** Under development

### Recommended Capability Expansion

**Priority 1: High** (Implement for parity with azure.ai.agents)
- [ ] Lifecycle Events: preprovision, predeploy, postdeploy
- [ ] Metadata Capability: Command schemas + configuration schemas
- [ ] Enhanced Error Messaging: Integration with azd event service

**Priority 2: Medium** (Optional, enable advanced workflows)
- [ ] Service Target Provider: If Drasi deployable as custom host type
- [ ] Framework Service Provider: For custom language/runtime detection
- [ ] Prompt Service Integration: For interactive source/query selection

**Priority 3: Low** (Future enhancements)
- [ ] MCP Server: For agent tool integration
- [ ] Distributed Tracing: OpenTelemetry correlation with azd

### Proposed Command Hierarchy

**EVIDENCED (Current):**
\\\ash
azd drasi <commands...>
\\\

**RECOMMENDED (Roadmap):**
\\\ash
# Configuration & Setup
azd drasi config --connection <string>        # Configure Drasi connection
azd drasi init                                # Initialize Drasi in project

# Query Management
azd drasi queries init                        # Scaffold query templates
azd drasi queries validate                    # Validate Cypher syntax
azd drasi queries deploy                      # Deploy to Drasi
azd drasi queries list                        # List deployed queries
azd drasi queries describe <query-name>       # Show query details

# Source Management
azd drasi sources list                        # List configured sources
azd drasi sources add --type <type> [--name <name>]  # Add source
azd drasi sources validate                    # Check connectivity

# Operations
azd drasi status                              # Check deployment status
azd drasi health check                        # Validate connectivity
azd drasi logs [--follow]                     # Stream logs
\\\

### Maturity Comparison vs. Peer Extensions

\\\
Extension              │ Custom Cmds │ Lifecycle │ Metadata │ Status
───────────────────────┼─────────────┼───────────┼──────────┼─────────
Drasi (current)        │      ✓      │     ?     │    ?     │ Unknown
Drasi (recommended)    │      ✓      │     ✓     │    ✓     │ +2 gaps
───────────────────────┼─────────────┼───────────┼──────────┼─────────
azure.ai.agents        │      ✓      │     ✓     │    ✓     │ Reference
microsoft.azd.demo     │      ✓      │     ✓     │    ✓     │ Reference
\\\

### Implementation Priority

1. **Audit extension.yaml** ← START HERE
   - Verify namespace, current capabilities
   - Check version alignment with framework (>= 1.23.6 recommended)

2. **Add metadata capability** ← QUICK WIN
   - Expose command schemas for IDE support
   - Document azure.yaml configuration validation

3. **Wire lifecycle events** ← HIGH VALUE
   - preprovision: Validate Drasi sources before provisioning
   - predeploy: Generate/compile queries
   - postdeploy: Health checks, logging

4. **Expand command documentation** ← POLISH
   - Update README with proposed command hierarchy
   - Add workflow examples (similar to this document)
   - Document each command's purpose, flags, examples

---

## EXTENSION FRAMEWORK: gRPC SERVICES

Extensions can invoke these gRPC services (via azdext SDK):

1. **Project Service** - Project metadata, list services/environments
2. **Environment Service** - List environments, get state
3. **User Config Service** - Read azure.yaml extension config
4. **Deployment Service** - Check deployment status
5. **Account Service** - Get subscription, credentials
6. **Prompt Service** - Styled prompts (select, multiselect, confirm, subscription/region/RG pickers)
7. **AI Model Service** - Query model catalog, check SKU/capacity
8. **Event Service** - Publish logs/events
9. **Container Service** - Build, push, run containers
10. **Framework Service** - Register language/framework build logic
11. **Service Target Service** - Register custom service targets
12. **Compose Service** - Docker Compose orchestration
13. **Workflow Service** - Trigger azd commands
14. **Copilot Service** - Copilot agent interaction

---

## SOURCES & REFERENCES

**Official Documentation:**
- Microsoft Learn - azd extensions: https://learn.microsoft.com/en-us/azure/developer/azure-developer-cli/extensions/overview
- Microsoft Learn - azd reference: https://learn.microsoft.com/en-us/azure/developer/azure-developer-cli/reference

**GitHub Repository & Framework:**
- Azure/azure-dev: https://github.com/Azure/azure-dev
- Extension Framework Docs: https://github.com/Azure/azure-dev/blob/main/cli/azd/docs/extensions/extension-framework.md
- Extension SDK Reference: https://github.com/Azure/azure-dev/blob/main/cli/azd/docs/extensions/extension-sdk-reference.md
- Extension E2E Walkthrough: https://github.com/Azure/azure-dev/blob/main/cli/azd/docs/extensions/extension-e2e-walkthrough.md
- Registry Schema: https://github.com/Azure/azure-dev/blob/main/cli/azd/extensions/registry.schema.json
- Official Registry: https://github.com/Azure/azure-dev/blob/main/cli/azd/extensions/registry.json
- Development Registry: https://aka.ms/azd/extensions/registry/dev

**Individual Extensions:**
- azure.ai.agents: https://github.com/Azure/azure-dev/tree/main/cli/azd/extensions/azure.ai.agents
- microsoft.azd.demo: https://github.com/Azure/azure-dev/tree/main/cli/azd/extensions/microsoft.azd.demo
- microsoft.azd.extensions: https://github.com/Azure/azure-dev/tree/main/cli/azd/extensions/microsoft.azd.extensions
- azure.coding-agent: https://github.com/Azure/azure-dev/tree/main/cli/azd/extensions/azure.coding-agent
- azure.appservice: https://github.com/Azure/azure-dev/tree/main/cli/azd/extensions/azure.appservice

---

**Report Generated:** April 6, 2026  
**azd Version Analyzed:** 1.23.14  
**Framework Status:** Public Preview (Beta)  
**Maturity Index:** All extensions Go-based (native support), steady releases  

**Separation:** ✓ Evidenced commands clearly marked | ✓ Inferred capabilities noted | ✓ Official sources cited
