---
name: gh-aw-operations
description: >-
  Create, compile, debug, and manage GitHub Agentic Workflows (gh-aw) with
  best practices and common patterns.
  USE FOR: building agentic workflows, debugging gh-aw compilation errors,
  configuring workflow frontmatter, managing workflow tools, or troubleshooting
  GitHub Agentic Workflow failures.
metadata:
  version: 1.0.0
  category: automation
  tags: [github-agentic-workflows, gh-aw, automation, ci-cd, ai-workflows]
---

# GitHub Agentic Workflows Operations Skill

This skill provides expertise in creating, managing, and troubleshooting GitHub Agentic Workflows (gh-aw framework).

## Skill 1: Frontmatter Configuration

### Description

Configure workflow frontmatter with all required and optional fields following gh-aw specifications.

### Implementation Pattern

```yaml
---
# ===== REQUIRED FIELDS =====
on:
  workflow_dispatch:
    inputs:
      parameter_name:
        description: "Clear description of parameter"
        required: false
        type: string|boolean|choice|number
        default: "default_value" # Boolean: "true"/"false" as strings
  schedule:
    - cron: "0 9 * * 1" # Monday 9 AM UTC
  issues:
    types: [opened, labeled]
  pull_request:
    types: [opened, synchronize]

engine: copilot # GitHub Copilot as AI engine

permissions:
  contents: read # NEVER use write — strict mode blocks it
  pull-requests: read # All writes go through safe-outputs
  issues: read

# ===== TOOLS CONFIGURATION =====
tools:
  edit: # bare key — enables file read AND edit
  bash: # bare key — default safe commands
  github:
    toolsets: [pull_requests] # Only include toolsets your workflow needs

# ===== MCP SERVERS (if needed) =====
mcp-servers:
  terraform:
    container: "hashicorp/terraform-mcp-server:0.4.0"
    env:
      TF_LOG: "INFO"
    allowed: ["*"]

# ===== SAFE OUTPUTS =====
safe-outputs:
  create-pull-request:
    title-prefix: "[automated] "
    labels: [automation]
    draft: true
    reviewers: [copilot]
    expires: 14
    fallback-as-issue: true
  create-issue:
    title-prefix: "[bot] "
    labels: [automation]
    expires: 7
  update-issue: null
  add-comment: null

# ===== NETWORK ACCESS =====
network:
  allowed:
    - defaults
    - registry.terraform.io
    - releases.hashicorp.com
    - api.github.com

# ===== IMPORTS (if using reusable agents/skills) =====
imports:
  - owner/repo/.github/agents/agent-name.agent.md@main
  - owner/repo/.github/skills/skill-name/SKILL.md@main
---
```

### Key Rules

- **`edit:` is a bare key** (no value) — enables both reading and writing files. `edit: null` and `edit: true` both fail compilation.
- **`read:` is not a valid tool** — file reading is provided by `edit:`.
- **`bash:` is a bare key** (no value) for default safe commands, or `bash: ["cmd"]` for specific commands.
- **`contents: write` / `pull-requests: write` / `issues: write` are blocked by strict mode** — use `safe-outputs:` for all write operations instead.
- **`pull_request:` trigger requires `types:`** — bare `pull_request:` fails compilation.
- **`workflow_dispatch:` is a bare key** — `workflow_dispatch: null` fails compilation.
- **Toolsets and permissions must align** — `toolsets: [default]` includes the `issues` toolset which requires `issues: read`; only declare the toolsets you actually need.
- Boolean defaults must be strings: `default: "true"` not `default: true`
- Safe-output fields use hyphens: `create-pull-request` not `create_pull_request`
- Imports use `owner/repo/path@ref` format, not raw GitHub URLs
- Safe-outputs can be bare key (default config) or object (custom config)

---

## Skill 2: Safe-Outputs Configuration

### Description

Configure safe-outputs for GitHub operations (PRs, issues, comments) with proper validation and best practices.

### Common Safe-Output Patterns

#### Pattern: Pull Request Creation

```yaml
safe-outputs:
  create-pull-request:
    title-prefix: "[type] "
    labels: [automation, bot-generated]
    draft: true
    reviewers: [copilot]
    expires: 14 # Auto-close after 14 days
    fallback-as-issue: true # Create issue if PR fails
    base-branch: main # Target branch
```

**When to use:**

- Automated code changes
- Dependency upgrades
- Code generation
- Refactoring

#### Pattern: Issue Creation

```yaml
safe-outputs:
  create-issue:
    title-prefix: "[report] "
    labels: [automation, needs-review]
    assignees: [copilot]
    expires: 7
    group: true # Group multiple issues as sub-issues
    close-older-issues: true # Close previous issues from same workflow
```

**When to use:**

- Reporting findings
- Tracking tasks
- Alerting on issues
- Documentation requests

#### Pattern: Issue/PR Comments

```yaml
safe-outputs:
  add-comment:
    target: "triggering" # Comment on triggering issue/PR
    max: 3
    hide-older-comments: true # Hide previous comments from same workflow
```

**When to use:**

- Status updates
- Analysis results
- Bot responses
- Progress reporting

#### Pattern: PR Reviews

```yaml
safe-outputs:
  create-pull-request-review-comment:
    max: 10
    side: "RIGHT"
    footer: "if-body" # Only show footer when review has body
  submit-pull-request-review:
    max: 1
    footer: false
```

**When to use:**

- Code review automation
- Inline suggestions
- Security analysis
- Style checking

#### Pattern: Label Management

```yaml
safe-outputs:
  add-labels:
    allowed: [bug, enhancement, documentation]
    max: 3
  remove-labels:
    allowed: [needs-triage, stale]
    max: 3
```

**When to use:**

- Workflow status tracking
- Issue classification
- Automation state management

### Available Safe-Output Types

- `create-pull-request` - Create PRs with code changes
- `create-issue` - Create issues
- `update-issue` - Update issue title/body/status
- `close-issue` - Close issues
- `create-discussion` - Create discussions
- `update-discussion` - Update discussions
- `close-discussion` - Close discussions
- `add-comment` - Add comments to issues/PRs/discussions
- `hide-comment` - Hide comments
- `add-labels` - Add labels
- `remove-labels` - Remove labels
- `add-reviewer` - Add PR reviewers
- `create-pull-request-review-comment` - Add PR review comments
- `submit-pull-request-review` - Submit PR review
- `update-pull-request` - Update PR title/body
- `dispatch-workflow` - Trigger other workflows
- `create-project`, `update-project` - Manage GitHub Projects
- `upload-asset` - Upload files to orphaned branch

---

## Skill 3: MCP Server Integration

### Description

Configure and use Model Context Protocol (MCP) servers for specialized tool access.

### Common MCP Server Configurations

#### Terraform MCP Server

```yaml
mcp-servers:
  terraform:
    container: "hashicorp/terraform-mcp-server:0.4.0"
    env:
      TF_LOG: "INFO"
    allowed: ["*"] # or specific tools
```

**Available operations:**

- Read Terraform configurations
- Analyze provider versions
- Detect breaking changes
- Generate upgrade recommendations

#### Azure MCP Server

```yaml
mcp-servers:
  azure:
    container: "mcr.microsoft.com/azure-sdk/azure-mcp:<pinned-version>"
    env:
      AZURE_SUBSCRIPTION_ID: ${{ secrets.AZURE_SUBSCRIPTION_ID }}
```

**Available operations:**

- Azure resource queries
- Best practices validation
- Documentation lookup
- Terraform Azure integration

#### Custom MCP Server

```yaml
mcp-servers:
  custom:
    container: "your-org/your-mcp-server:v1.0.0"
    env:
      API_KEY: ${{ secrets.API_KEY }}
    allowed: ["specific_tool1", "specific_tool2"]
```

### Network Configuration for MCP Servers

```yaml
network:
  allowed:
    - defaults # GitHub APIs
    - registry.terraform.io
    - releases.hashicorp.com
    - management.azure.com
    - custom-api.example.com
```

---

## Skill 4: Workflow Compilation & Debugging

### Description

Compile workflows and resolve common compilation errors.

### Compilation Commands

```bash
# Compile single workflow
gh aw compile workflow-name

# Compile all workflows
gh aw compile

# Force recompile
gh aw compile --force workflow-name

# Validate without compiling
gh aw validate workflow-name
```

### Common Compilation Errors & Fixes

#### Error: Invalid tool value for `edit`

**Problem:**

```yaml
tools:
  read: null   # ❌ 'read' is not a valid tool
  edit: null   # ❌ null not valid for edit
  edit: true   # ❌ true not valid for edit
```

**Solution:**

```yaml
tools: # ✅ bare keys
  edit:
  bash:
```

#### Error: Write permissions blocked by strict mode

**Problem:**

```yaml
permissions:
  contents: write # ❌ blocked by strict mode
  pull-requests: write # ❌ blocked by strict mode
```

**Solution:**

```yaml
permissions: # ✅ read-only
  contents: read
  pull-requests: read
safe-outputs: # ✅ write operations go here
  create-pull-request: null
```

#### Error: `pull_request` trigger requires `types`

**Problem:**

```yaml
on:
  pull_request: # ❌ bare null not valid
  workflow_dispatch: null # ❌ null not valid
```

**Solution:**

```yaml
on:
  workflow_dispatch: # ✅ bare key
  pull_request: # ✅ explicit types
    types: [opened, synchronize, reopened]
```

#### Error: Missing permission for toolset

**Problem:**

```yaml
tools:
  github:
    toolsets: [default] # includes 'issues' toolset
permissions:
  contents: read # ❌ missing 'issues: read', compiler warns
```

**Solution:**

```yaml
tools:
  github:
    toolsets: [pull_requests] # ✅ only what you need
permissions:
  contents: read
  pull-requests: read
```

#### Error: "got array, want object" for tools

**Problem:**

```yaml
tools: ["bash", "read", "edit"] # ❌ Wrong
```

**Solution:**

```yaml
tools: # ✅ Correct
  bash:
  edit:
```

#### Error: Boolean default must be string

**Problem:**

```yaml
workflow_dispatch:
  inputs:
    enabled:
      type: boolean
      default: true # ❌ Wrong
```

**Solution:**

```yaml
workflow_dispatch:
  inputs:
    enabled:
      type: boolean
      default: "true" # ✅ Correct (quoted)
```

#### Error: Invalid safe-output field name

**Problem:**

```yaml
safe-outputs:
  create_pull_request: null # ❌ Wrong (underscore)
```

**Solution:**

```yaml
safe-outputs:
  create-pull-request: null # ✅ Correct (hyphen)
```

#### Error: Import download failed

**Problem:**

```yaml
imports:
  - https://raw.githubusercontent.com/owner/repo/main/agent.md # ❌ Wrong
```

**Solution:**

```yaml
imports:
  - owner/repo/.github/agents/agent.md@main # ✅ Correct
```

#### Error: Safe-outputs format

**Problem:**

```yaml
safe-outputs: [create-issue, create-pull-request] # ❌ Wrong (array)
```

**Solution:**

```yaml
safe-outputs: # ✅ Correct (object)
  create-issue: null
  create-pull-request: null
```

---

## Skill 5: Workflow Testing & Deployment

### Description

Deploy and test agentic workflows in GitHub Actions.

### Deployment Checklist

1. **Compile Workflow**

   ```bash
   gh aw compile workflow-name
   ```

   - Verify 0 errors
   - Check warnings (informational only)

2. **Commit Files**

   ```bash
   git add .github/workflows/workflow-name.md
   git add .github/workflows/workflow-name.lock.yml
   git commit -m "feat: Add workflow-name agentic workflow"
   git push
   ```

3. **Enable PR Creation** (if using `create-pull-request`)
   - Go to: Settings → Actions → General
   - Check: ✓ Allow GitHub Actions to create and approve pull requests
   - Save

4. **Configure Secrets** (if needed)
   - Go to: Settings → Secrets and variables → Actions
   - Add required secrets (API keys, tokens, etc.)

5. **Test Workflow**
   - Navigate to: Actions tab
   - Select workflow
   - Click "Run workflow"
   - Provide inputs
   - Monitor execution

### Testing Strategies

#### Strategy 1: Manual Trigger Testing

```yaml
on:
  workflow_dispatch:
    inputs:
      dry_run:
        description: "Dry run mode (no actual changes)"
        type: boolean
        default: "true"
```

Test without making real changes first.

#### Strategy 2: Branch Testing

```yaml
on:
  push:
    branches:
      - test/**
```

Test on dedicated test branches before enabling on main.

#### Strategy 3: Label-Based Testing

```yaml
on:
  issues:
    types: [labeled]
  # Only run when "test-workflow" label is added
```

Control execution via labels during testing.

---

## Skill 6: Import Reusable Agents & Skills

### Description

Use imports to leverage reusable agents and skills from other repositories.

### Import Patterns

#### Import External Agent

```yaml
imports:
  - thomast1906/github-copilot-skills-terraform/.github/agents/terraform-provider-upgrade.agent.md@main
```

**In markdown instructions:**

```markdown
Use your imported Terraform provider upgrade expertise to analyze
the current provider versions and recommend upgrades.
```

#### Import External Skill

```yaml
imports:
  - owner/repo/.github/skills/code-review/SKILL.md@main
  - owner/repo/.github/skills/security-scan/SKILL.md@main
```

**In markdown instructions:**

```markdown
Apply your code review and security scanning skills to analyze
the changes in this pull request.
```

#### Import Multiple Resources

```yaml
imports:
  - owner/repo/.github/agents/analyzer.agent.md@main
  - owner/repo/.github/skills/reporting/SKILL.md@main
  - owner/repo/.github/skills/validation/SKILL.md@main
```

### Import Best Practices

1. **Pin to specific refs when stable:**

   ```yaml
   imports:
     - owner/repo/.github/agents/agent.md@v1.2.0 # Tag
     - owner/repo/.github/agents/agent.md@abc1234 # Commit SHA
   ```

2. **Use main/master for latest:**

   ```yaml
   imports:
     - owner/repo/.github/agents/agent.md@main
   ```

3. **Imports are cached in `.github/aw/imports/`**
   - Committed to repository
   - Updated on recompilation
   - Version-locked for consistency

---

## Skill 7: Common Workflow Patterns

### Description

Pre-built patterns for common automation scenarios.

### Pattern: Dependency Upgrade Workflow

```yaml
---
on:
  workflow_dispatch:
    inputs:
      target_version:
        description: 'Target version (leave empty for latest)'
        required: false
        type: string
  schedule:
    - cron: '0 9 * * 1'  # Weekly Monday 9 AM

engine: copilot

permissions:
  contents: read
  pull-requests: read

tools:
  edit:
  github:
    toolsets: [pull_requests]

safe-outputs:
  create-pull-request:
    title-prefix: "[dependency] "
    labels: [dependencies, automation]
    draft: true
    reviewers: [copilot]
    expires: 14
  create-issue:
    title-prefix: "[dependency] "
    labels: [dependencies, blocked]

network:
  allowed:
    - defaults
    - registry.npmjs.org
    - registry.terraform.io
---

# Dependency Upgrade Workflow

Automatically upgrade project dependencies to latest compatible versions.

## Your Task

1. Scan current dependencies
2. Check for available upgrades
3. Analyze breaking changes
4. Create PR with upgrades or issue if blocked
```

### Pattern: Code Analysis & Reporting

```yaml
---
on:
  workflow_dispatch:
  pull_request:
    types: [opened, synchronize]

engine: copilot

permissions:
  contents: read
  pull-requests: read

tools:
  edit:
  github:
    toolsets: [pull_requests]

safe-outputs:
  add-comment:
    target: "triggering"
  create-pull-request-review-comment:
    max: 10
  add-labels:
    allowed: [needs-review, lgtm, requires-changes]
---

# Code Analysis Workflow

Analyze code quality and provide feedback on pull requests.

## Your Task

Review the PR changes and provide:
1. Code quality feedback
2. Security concerns
3. Best practice suggestions
4. Inline comments on specific issues
```

### Pattern: IssueOps/ChatOps

```yaml
---
on:
  issues:
    types: [opened, labeled]
  issue_comment:
    types: [created]

engine: copilot

permissions:
  contents: read
  issues: read

tools:
  github:
    toolsets: [issues]

safe-outputs:
  add-comment: null
  update-issue: null
  add-labels:
    allowed: [in-progress, completed, needs-info]
  close-issue: null
---

# IssueOps Workflow

Respond to issue commands and manage issue lifecycle.

## Your Task

When an issue is created or commented on:
1. Parse commands (e.g., /analyze, /report, /close)
2. Execute the requested operation
3. Report results as comment
4. Update issue labels/status
```

---

## Skill 8: Troubleshooting Guide

### Description

Diagnose and fix common issues with agentic workflows.

### Issue: Workflow Not Appearing in Actions UI

**Symptoms:**

- Workflow file exists but doesn't show in Actions tab
- Can't manually trigger workflow

**Diagnosis:**

```bash
# Check if lock file was generated
ls -la .github/workflows/*.lock.yml

# Check if files are committed
git status

# Verify compilation succeeded
gh aw compile workflow-name
```

**Solutions:**

1. Ensure both `.md` and `.lock.yml` are committed and pushed
2. Verify `workflow_dispatch` trigger is configured
3. Check repository Settings → Actions to ensure workflows are enabled

### Issue: PR Creation Fails

**Symptoms:**

```
Warning: Failed to create pull request: GitHub Actions is not permitted to create or approve pull requests.
```

**Solution:**

1. Go to: Settings → Actions → General
2. Scroll to: "Workflow permissions"
3. Check: ✓ Allow GitHub Actions to create and approve pull requests
4. Click Save

### Issue: MCP Server Not Working

**Symptoms:**

- Tools from MCP server not available
- Connection timeouts

**Diagnosis:**

```yaml
# Check container image is correct
mcp-servers:
  terraform:
    container: "hashicorp/terraform-mcp-server:0.4.0" # Verify version
```

**Solutions:**

1. Verify container image exists and is accessible
2. Check network allowlist includes required domains
3. Verify environment variables are set correctly
4. Check container logs in workflow run

### Issue: Import Not Resolving

**Symptoms:**

```
Error: Failed to download import: owner/repo/path@ref
```

**Solutions:**

1. Verify repository is public or accessible with token
2. Check path is correct: `.github/agents/name.agent.md`
3. Verify branch/tag/SHA exists
4. Use correct format: `owner/repo/path@ref` (not raw URL)

---

## Skill 9: Performance Optimization

### Description

Optimize workflow execution time and resource usage.

### Optimization Techniques

#### 1. Minimize MCP Server Usage

```yaml
# Only allow specific tools needed
mcp-servers:
  terraform:
    allowed: ["analyze_version", "detect_breaking_changes"] # Not "*"
```

#### 2. Set Appropriate Timeouts

```yaml
# Don't run indefinitely
on:
  workflow_dispatch:
    inputs:
      timeout_minutes:
        description: "Maximum execution time"
        default: "30"
```

#### 3. Limit Safe-Output Operations

```yaml
safe-outputs:
  create-pull-request-review-comment:
    max: 10 # Limit number of comments
```

#### 5. Use Concise Instructions

```markdown
# ❌ Avoid overly verbose instructions

Please very carefully analyze every single line...

# ✅ Be concise and clear

Analyze changes and report findings.
```

---

## Skill 10: Security Best Practices

### Description

Implement secure configurations for agentic workflows.

### Security Checklist

#### Minimal Permissions

```yaml
permissions:
  contents: read
  pull-requests: read # Match toolsets; all writes go through safe-outputs
  issues: read

safe-outputs:
  create-pull-request: null # Write operations live here, not in permissions
  create-issue: null
```

**Rule:** Strict mode blocks `write` permissions. Use `safe-outputs:` for all write operations.

#### Secret Management

```yaml
mcp-servers:
  custom:
    env:
      API_KEY: ${{ secrets.API_KEY }} # Use secrets, not hardcoded values
```

#### Network Restrictions

```yaml
network:
  allowed:
    - defaults
    - specific-domain.com # Only allow required domains
    # Don't use wildcards or overly broad allowlists
```

#### Input Validation

```yaml
on:
  workflow_dispatch:
    inputs:
      target:
        type: choice # Use choice instead of string when possible
        options: [dev, staging, prod]
```

#### Safe-Output Restrictions

```yaml
safe-outputs:
  add-labels:
    allowed: [bug, enhancement] # Restrict which labels can be added
  add-reviewer:
    reviewers: [team-lead, copilot] # Restrict who can be assigned
```

---

## Quick Reference Card

### File Structure

```
.github/
├── workflows/
│   ├── workflow-name.md          # Your workflow definition
│   └── workflow-name.lock.yml    # Generated (committed)
├── agents/
│   └── agent-name.agent.md       # Reusable agents
└── skills/
    └── skill-name/
        └── SKILL.md               # Reusable skills
```

### Essential Commands

```bash
# Compile workflow
gh aw compile workflow-name

# Compile all workflows
gh aw compile

# Validate workflow
gh aw validate workflow-name

# List workflows
gh aw list
```

### Frontmatter Quick Template

```yaml
---
on:
  workflow_dispatch: # bare key — no null
  pull_request:
    types: [opened, synchronize, reopened] # always specify types
engine: copilot
permissions:
  contents: read # never write — strict mode blocks it
  pull-requests: read # match only what your toolsets need
tools:
  edit: # bare key — enables file read/edit
  github:
    toolsets: [pull_requests] # only what you need
safe-outputs:
  create-pull-request: null # all write ops go here
network:
  allowed: [defaults]
---
```

### Common Safe-Outputs

- `create-pull-request` - Code changes
- `create-issue` - Reports/alerts
- `add-comment` - Status updates
- `update-issue` - Track progress
- `add-labels` - State management

### Debug Workflow Issues

1. Check compilation: `gh aw compile workflow-name`
2. Verify files committed: `git status`
3. Check Actions enabled: Settings → Actions
4. Enable PR creation: Settings → Actions → General
5. Review workflow logs in Actions tab

---

## Related Skills

- [Drasi Queries](../drasi-queries/SKILL.md) — Continuous query definitions used by workflows
- [GitHub Actions CI/CD](../github-actions-ci-cd/SKILL.md) — General GitHub Actions patterns

---

## Currency

- **Date checked:** 2026-03-31
- **Sources:** [GitHub Advanced Workflows CLI](https://github.com/gh-aw/gh-aw), Microsoft Learn MCP
- **Authoritative references:** [gh-aw documentation](https://github.com/gh-aw/gh-aw)

### Verification Steps

1. Confirm gh-aw CLI version and any new workflow commands
2. Verify compilation and deployment patterns match current gh-aw release
3. Check for new workflow template types or deprecations
