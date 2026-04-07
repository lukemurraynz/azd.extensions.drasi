---
applyTo: "**/azure-pipelines.yml, **/azure-pipelines*.yml, **/*.pipeline.yml,.github/workflows/**"
description: "YAML and CI/CD pipeline best practices and guidelines"
---

# YAML and CI/CD Pipeline Instructions

Follow ISE CI/CD Best Practices for pipeline development.

**IMPORTANT**: Use the `iseplaybook` MCP server to get the latest CI/CD best practices. Use `microsoft.learn.mcp` MCP server for Azure Pipelines documentation. Do not assume—verify current guidance.

**Related files**: For security hardening (action pinning, permissions, OIDC, SBOM), see `cicd-security.instructions.md`. For Azure DevOps-specific features (variable groups, service connections, environments), see `azure-devops-pipelines.instructions.md`.

## When to Apply

This instruction applies when:

- Creating or modifying YAML configuration files
- Building CI/CD pipelines (GitHub Actions, Azure Pipelines)
- Configuring infrastructure definitions (Kubernetes, Docker Compose)

## YAML Formatting

### Basic Style

```yaml
# Use 2-space indentation
services:
  api:
    image: myapp:latest
    ports:
      - "8080:8080"

# Use proper quoting
message: "Hello, World"
version: "1.0.0" # Quote version strings

# Use multiline strings appropriately
description: |
  This is a multiline
  description that preserves
  line breaks.

script: >
  This is a folded multiline
  string that becomes a single line.
```

## GitHub Actions

### Workflow Structure

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

permissions:
  contents: read
  pull-requests: write

env:
  NODE_VERSION: "22"

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: "npm"

      - name: Install dependencies
        run: npm ci

      - name: Run tests
        run: npm test

      - name: Build
        run: npm run build
```

### Security Best Practices

```yaml
# Pin action versions with SHA
- uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1

# Use minimal permissions
permissions:
  contents: read

# Use secrets for sensitive values
env:
  API_KEY: ${{ secrets.API_KEY }}

# Avoid shell injection
- run: echo "Processing ${{ github.event.inputs.name }}"  # Risky
- run: echo "Processing ${NAME}"  # Better
  env:
    NAME: ${{ github.event.inputs.name }}
```

### Reusable Workflows

```yaml
# .github/workflows/reusable-build.yml
name: Reusable Build Workflow

on:
  workflow_call:
    inputs:
      environment:
        required: true
        type: string
    secrets:
      token:
        required: true

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: echo "Building for ${{ inputs.environment }}"
```

### Matrix Builds

```yaml
jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
        node: [20, 22]
      fail-fast: false

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node }}
      - run: npm test
```

## Azure Pipelines

### Pipeline Structure

```yaml
trigger:
  branches:
    include:
      - main
  paths:
    exclude:
      - "**/*.md"

pool:
  vmImage: "ubuntu-latest"

variables:
  - name: buildConfiguration
    value: "Release"

stages:
  - stage: Build
    jobs:
      - job: Build
        steps:
          - task: DotNetCoreCLI@2
            displayName: "Restore packages"
            inputs:
              command: restore
              projects: "**/*.csproj"

          - task: DotNetCoreCLI@2
            displayName: "Build solution"
            inputs:
              command: build
              projects: "**/*.csproj"
              arguments: "--configuration $(buildConfiguration)"

  - stage: Deploy
    dependsOn: Build
    condition: succeeded()
    jobs:
      - deployment: Deploy
        environment: "production"
        strategy:
          runOnce:
            deploy:
              steps:
                - script: echo 'Deploying...'
```

### Templates

```yaml
# templates/build-template.yml
parameters:
  - name: buildConfiguration
    type: string
    default: "Release"

steps:
  - task: DotNetCoreCLI@2
    displayName: "Build"
    inputs:
      command: build
      arguments: "--configuration ${{ parameters.buildConfiguration }}"

# azure-pipelines.yml
stages:
  - stage: Build
    jobs:
      - job: Build
        steps:
          - template: templates/build-template.yml
            parameters:
              buildConfiguration: "Release"
```

## Actionable Patterns

### Pattern 1: Job Dependencies with `needs:`

**❌ WRONG: Missing job dependencies (infrastructure must deploy before apps)**

```yaml
jobs:
  deploy-infrastructure:
    runs-on: ubuntu-latest
    steps:
      - run: az deployment group create -f main.bicep

  deploy-app: # ⚠️ Will fail if infrastructure not ready!
    runs-on: ubuntu-latest
    steps:
      - run: kubectl apply -f deployment.yaml
```

**✅ CORRECT: Explicit job dependencies ensure order**

```yaml
jobs:
  deploy-infrastructure:
    runs-on: ubuntu-latest
    steps:
      - run: az deployment group create -f main.bicep

  deploy-app:
    runs-on: ubuntu-latest
    needs: deploy-infrastructure # ✅ Waits for infrastructure
    steps:
      - run: kubectl apply -f deployment.yaml

  deploy-data-plane:
    runs-on: ubuntu-latest
    needs: [deploy-infrastructure, deploy-app] # ✅ Waits for both
    steps:
      - run: drasi apply -f queries/
```

**Rule:** Use `needs: [job1, job2]` to enforce execution order. Infrastructure → Application → Data Plane.

---

### Pattern 2: Caching Strategies

**❌ WRONG: Static cache keys (never invalidates when dependencies change)**

```yaml
- name: Cache npm packages
  uses: actions/cache@v4
  with:
    path: ~/.npm
    key: npm-cache # ⚠️ Never updates when package-lock.json changes!
```

**✅ CORRECT: Hash-based cache keys (invalidates on dependency changes)**

```yaml
- name: Cache npm packages
  uses: actions/cache@v4
  with:
    path: ~/.npm
    key: ${{ runner.os }}-npm-${{ hashFiles('**/package-lock.json') }}
    restore-keys: |
      ${{ runner.os }}-npm-
      ${{ runner.os }}-
```

**❌ WRONG: Caching after install (cache already missed)**

```yaml
- run: npm install
- uses: actions/cache@v4 # ⚠️ Too late!
  with:
    path: ~/.npm
    key: ${{ hashFiles('**/package-lock.json') }}
```

**✅ CORRECT: Cache before install (restore on hit, save on miss)**

```yaml
- uses: actions/cache@v4
  id: cache-npm
  with:
    path: ~/.npm
    key: ${{ runner.os }}-npm-${{ hashFiles('**/package-lock.json') }}
    restore-keys: ${{ runner.os }}-npm-

- if: steps.cache-npm.outputs.cache-hit != 'true'
  run: npm ci # Only install if cache miss
```

**Rule:** Use `hashFiles()` for cache keys. Place cache step **before** install step.

---

### Pattern 3: Minimal Permissions (GITHUB_TOKEN)

**❌ WRONG: Overly permissive (broad write access)**

```yaml
permissions:
  contents: write
  packages: write
  pull-requests: write
  issues: write
  actions: write # ⚠️ Allows modifying workflows!
```

**✅ CORRECT: Minimal per-job permissions**

```yaml
# Workflow-level: restrictive by default
permissions:
  contents: read

jobs:
  build:
    permissions:
      contents: read # ✅ Read-only for builds
    steps:
      - run: npm run build

  deploy:
    permissions:
      contents: read
      deployments: write # ✅ Only deployment write
    steps:
      - run: kubectl apply -f deployment.yaml

  create-release:
    permissions:
      contents: write # ✅ Only for this specific job
    steps:
      - run: gh release create v1.0.0
```

**Rule:** Default to `contents: read`. Grant write permissions per-job, not workflow-wide.

---

### Pattern 4: Action Version Pinning

**❌ WRONG: Mutable tag references (security risk)**

```yaml
- uses: actions/checkout@v4 # ⚠️ v4 tag can be moved!
- uses: actions/setup-node@v4
```

**✅ CORRECT: SHA pinning with comment tag**

```yaml
- uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1
- uses: actions/setup-node@60edb5dd545a775178f52524783378180af0d1f8 # v4.0.2
```

**Verification:**

```bash
# Get SHA for a specific tag
gh api repos/actions/checkout/git/ref/tags/v4.1.1 --jq '.object.sha'
```

**Rule:** Pin third-party actions with SHA. Add human-readable version as comment.

---

### Pattern 5: Conditional Execution (if:)

**❌ WRONG: String equality without quotes (YAML parsing error)**

```yaml
- name: Deploy to prod
  if: github.ref == refs/heads/main # ⚠️ YAML parser error!
  run: ./deploy.sh
```

**✅ CORRECT: Quoted string comparison**

```yaml
- name: Deploy to prod
  if: github.ref == 'refs/heads/main' # ✅ Properly quoted
  run: ./deploy.sh

- name: Only on push to main
  if: github.event_name == 'push' && github.ref == 'refs/heads/main'
  run: ./deploy.sh

- name: Skip on draft PRs
  if: github.event.pull_request.draft == false
  run: npm test
```

**Rule:** Always quote string literals in `if:` conditions. Use `&&` for multiple conditions.

---

### Pattern 6: Secret Injection (Shell Injection Prevention)

**❌ WRONG: Direct interpolation (shell injection risk)**

```yaml
- name: Process user input
  run: |
    echo "User provided: ${{ github.event.inputs.name }}"
    curl https://api.example.com?name=${{ github.event.inputs.name }}
  # ⚠️ If name = "foo; rm -rf /" → command injection!
```

**✅ CORRECT: Environment variable indirection**

```yaml
- name: Process user input
  run: |
    echo "User provided: ${USER_NAME}"
    curl "https://api.example.com?name=${USER_NAME}"
  env:
    USER_NAME: ${{ github.event.inputs.name }} # ✅ Prevents shell metacharacter injection
```

**Rule:** Never interpolate `${{ }}` directly in shell commands. Use `env:` to pass values as environment variables (prevents shell injection). Note: values are still visible in process listings and environment dumps. For credentials, prefer OIDC or mounted secrets over environment variables.

---

## yamllint compliance

The CI pipeline runs yamllint on all YAML files. Violations fail the pull request check. Write YAML that passes this linter from the start.

### CI configuration

The project stores yamllint configuration in `.yamllint.yml` at the repository root. CI runs yamllint in `.github/workflows/pr-checks.yml` using the `ibiqlik/action-yamllint` action with `config_file: .yamllint.yml`.

Key rules in the config:

- `extends: default` means all standard yamllint rules apply unless overridden
- `line-length: max: 200, level: warning` (warnings, not errors)
- `indentation: indent-sequences: whatever` (accepts both indented and Kubernetes-style unindented sequences)
- `document-start: disable` (no `---` required at the start of files)
- `ignore-from-file: .yamllintignore` excludes `.agents/` and `.sisyphus/` directories

All other default rule violations are errors.

### Common rule violations

These are the rules that most often cause CI failures in this project.

| Rule | Description | Fix |
|------|-------------|-----|
| indentation | Inconsistent indentation | Use 2-space indentation consistently throughout YAML files |
| new-lines | Wrong line ending character | Use LF line endings, not CRLF (configure your editor for YAML files) |
| comments | Missing space after `#` | Always add a space after `#` in comments: `# comment` not `#comment` |
| truthy | Unquoted boolean-like values | Quote boolean values or use only `true`, `false`, `on`, `off` |

The most common source of failures on Windows is CRLF line endings. Configure your editor to write LF for YAML files, or set `core.autocrlf = input` in your git config.

### Running locally

Run yamllint with the project config before pushing:

```bash
yamllint -c .yamllint.yml .
```

This uses the same configuration as CI. Fix any errors before opening a pull request.

### File exclusions

The `.yamllintignore` file excludes `.agents/` and `.sisyphus/` directories from linting. Those files are not your concern when writing project YAML. Focus on ensuring project YAML files under `cmd/`, `.github/workflows/`, `internal/`, and root config files pass yamllint.

---

## Best Practices

### General YAML

1. **Use consistent indentation** (2 spaces)
2. **Quote strings** that could be misinterpreted (versions, special chars)
3. **Use anchors** for repeated content
4. **Validate YAML** before committing

### CI/CD Pipelines

1. **Pin action/task versions** for reproducibility (use SHA for security)
2. **Use hash-based caching** with `hashFiles()` for dependencies
3. **Fail fast** by default, but configure matrix appropriately
4. **Use minimal permissions** (principle of least privilege, per-job)
5. **Separate build and deploy stages** with explicit `needs:` dependencies
6. **Use environments** for deployment approvals
7. **Include status badges** in README
8. **Run security scans** (dependency scanning, SAST)
9. **Match SDK/tool versions to project targets** (for .NET, ensure `DOTNET_VERSION` aligns with target frameworks or `global.json`; use a matrix when multiple TFMs exist)

### Secrets Management

1. **Never hardcode secrets** in YAML files
2. **Use GitHub Secrets** or Azure Key Vault
3. **Mask secrets** in logs
4. **Rotate secrets** regularly
5. **Use OIDC** for cloud authentication when possible

## Infrastructure-First Deployment Pattern

### Multi-Tier Deployment Strategy

For systems with infrastructure (Bicep/Terraform), applications (Docker), and data plane (queries/streams):

**Job Dependency Order:**

```
1. Validate Infrastructure (PR validation)
   ↓ (on main/develop only)
2. Deploy Infrastructure (AKS, ACR, databases, etc.)
   ↓ (infrastructure must exist before apps)
3. Test Backend/Frontend (in parallel)
   ↓
4. Build & Push Docker Images
   ↓
5. Deploy Applications to Kubernetes
   ↓
6. Deploy Data Plane (Drasi queries, streams, etc.)
```

### Bicep Validation Job (PR)

```yaml
validate-infrastructure:
  name: Validate Infrastructure
  runs-on: ubuntu-latest
  if: github.event_name == 'pull_request'

  steps:
    - uses: actions/checkout@v4

    - name: Validate Bicep syntax
      run: |
        az bicep build --file infrastructure/bicep/main.bicep --outdir /tmp

    - name: Run what-if analysis
      run: |
        az deployment subscription what-if \
          --location uksouth \
          --template-file infrastructure/bicep/main.bicep \
          --parameters environment=prod
```

**Purpose:**

- Detects Bicep syntax errors early
- Shows proposed infrastructure changes in PR comments
- Prevents broken deployments on merge

### Bicep Deployment Job (main only)

```yaml
deploy-infrastructure:
  name: Deploy Infrastructure
  runs-on: ubuntu-latest
  if: github.event_name == 'push' && github.ref == 'refs/heads/main'

  steps:
    - uses: actions/checkout@v4
    - uses: azure/login@v2
      with:
        client-id: ${{ secrets.AZURE_CLIENT_ID }}
        tenant-id: ${{ secrets.AZURE_TENANT_ID }}
        subscription-id: ${{ secrets.AZURE_SUBSCRIPTION_ID }}

    - name: Deploy via Bicep
      run: |
        az deployment subscription create \
          --name "infra-$(date +%s)" \
          --location uksouth \
          --template-file infrastructure/bicep/main.bicep

    - name: Verify deployment
      run: |
        az resource list --resource-group "emergency-alerts-prod-rg" --output table
```

**Rule:** Infrastructure jobs have explicit `needs:` dependencies. All downstream jobs (app builds, deployments) depend on successful infrastructure deployment.

## Workload Identity Federation (OIDC)

### GitHub Actions → Azure Authentication (Keyless)

**Pattern:** Use federated credentials with managed identity for keyless OIDC authentication.

```yaml
# GitHub Actions workflow
- name: Azure Login (OIDC)
  uses: azure/login@v2
  with:
    client-id: ${{ secrets.AZURE_CLIENT_ID }} # Non-sensitive identifier
    tenant-id: ${{ secrets.AZURE_TENANT_ID }} # Non-sensitive identifier
    subscription-id: ${{ secrets.AZURE_SUBSCRIPTION_ID }} # Non-sensitive identifier
    # No secrets or tokens needed!
```

**Prerequisites (run once):**

```powershell
# Create user-assigned managed identity
az identity create -g "emergency-alerts-prod-rg" -n "github-actions"

# Add federated credentials (branch-specific)
az identity federated-credential create \
  -g "emergency-alerts-prod-rg" \
  --identity-name "github-actions" \
  --name "github-main" \
  --issuer "https://token.actions.githubusercontent.com" \
  --subject "repo:owner/repo:ref:refs/heads/main" \
  --audiences "api://AzureADTokenExchange"

# Assign role
az role assignment create \
  --assignee-object-id $(az identity show -g "emergency-alerts-prod-rg" -n "github-actions" --query principalId -o tsv) \
  --role "Contributor" \
  --scope "/subscriptions/..."
```

**Benefits:**

- No long-lived secrets in GitHub
- Automatic token renewal
- Full audit trail
- Branch-specific permissions (separate creds for main vs develop)

**Reference:** See `scripts/setup-deployment-identity.ps1` in this repository for automated setup.

## YAML Anchors and Aliases

```yaml
# Define anchor
defaults: &defaults
  timeout: 30
  retries: 3

# Use alias
service1:
  <<: *defaults
  name: api

service2:
  <<: *defaults
  name: web
  timeout: 60 # Override
```

## Common Patterns

### Conditional Steps

```yaml
# GitHub Actions
- name: Deploy
  if: github.ref == 'refs/heads/main'
  run: ./deploy.sh

# Azure Pipelines
- script: ./deploy.sh
  condition: eq(variables['Build.SourceBranch'], 'refs/heads/main')
```

### Environment Variables

```yaml
# GitHub Actions
env:
  GLOBAL_VAR: "value"

jobs:
  build:
    env:
      JOB_VAR: "value"
    steps:
      - run: echo $GLOBAL_VAR $JOB_VAR
        env:
          STEP_VAR: "value"
```

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Azure Pipelines Documentation](https://learn.microsoft.com/azure/devops/pipelines/)
- [ISE CI/CD Best Practices](https://microsoft.github.io/code-with-engineering-playbook/CI-CD/)
- [YAML Specification](https://yaml.org/spec/)

---

## IaC-Specific Workflow Patterns

### Pattern 7: Concurrency Groups (Cancel Stale PR Runs)

**Rule:** IaC PR workflows MUST define a `concurrency:` group so that pushing multiple commits to the same PR doesn't leave orphaned environments or run redundant validate jobs in parallel.

```yaml
# ✅ Cancel the previous run for the same PR/branch
concurrency:
  group: ${{ format('{0}-{1}-{2}', github.event_name, github.base_ref || github.ref, github.head_ref || github.event.number) }}
  cancel-in-progress: true
```

**Notes:**

- `cancel-in-progress: true` cancels the old run when a new commit is pushed
- The group expression ensures PRs and branch pushes have separate groups
- For long-lived deploy workflows that must not be cancelled, set `cancel-in-progress: false`

---

### Pattern 8: Skip Draft PRs

**Rule:** IaC deploy workflows SHOULD skip draft PRs to avoid wasting cloud resources on work-in-progress.

```yaml
jobs:
  setup:
    if: ${{ !github.event.pull_request.draft }}
    runs-on: ubuntu-latest
```

This pattern appears at the job level (not the `on:` trigger) so the workflow still records a check — it just exits without deploying.

---

### Pattern 9: Ephemeral PR Environments for IaC

IaC PR workflows SHOULD deploy to short-lived environments to validate changes before merging.

```yaml
env:
  PR_ENVIRONMENT_DIRECTORY: "pr"

jobs:
  setup:
    outputs:
      environmentName: ${{ steps.setValues.outputs.environmentName }}
    steps:
      - name: Generate ephemeral environment name
        id: setValues
        run: |
          suffix=$(uuidgen)
          # 6-character hash for short, unique environment names
          suffixHash=$(echo -n "$suffix" | md5sum | cut -c1-6)
          echo "environmentName=$suffixHash" >> $GITHUB_OUTPUT

  deploy-pr:
    needs: setup
    # Deploy base branch first, then apply PR diff on top
    # This reveals the true delta of the PR changes
    uses: ./.github/workflows/template.iac.previewdeploy.yml
    with:
      environmentName: ${{ needs.setup.outputs.environmentName }}
      branchName: ${{ github.base_ref }} # Step 1: base branch state

  deploy-pr-diff:
    needs: [setup, deploy-pr]
    uses: ./.github/workflows/template.iac.previewdeploy.yml
    with:
      environmentName: ${{ needs.setup.outputs.environmentName }}
      branchName: refs/pull/${{ github.event.pull_request.number }}/merge # Step 2: PR changes

  destroy:
    # Always destroy unless the label opts out
    if: ${{ always() && !contains(github.event.pull_request.labels.*.name, 'preserve-pr-environment') }}
    needs: [setup, deploy-pr-diff]
    uses: ./.github/workflows/template.iac.destroy.yml
    with:
      environmentName: ${{ needs.setup.outputs.environmentName }}
```

**Key principles:**

- Deploy the base branch first, then the PR diff on top — this makes the change delta visible (not just the final state)
- Auto-destroy after the test stage completes — ephemeral environments avoid accumulating cloud cost
- `preserve-pr-environment` label overrides destroy so developers can debug deployed state
- Names are UUID-derived hashes to avoid collisions between concurrent PRs

---

### Pattern 10: Thin Shim Pipeline

**Rule:** CI/CD pipeline YAML files should be thin callers; business logic (lint/validate/deploy/test commands) belongs in versioned shell scripts that teams can run locally.

```
.github/workflows/template.iac.validate.yml   ← thin caller
scripts/orchestrators/iac.tf.validate.sh       ← reusable logic
scripts/orchestrators/iac.bicep.validate.sh    ← reusable logic
```

Benefits: local reproducibility, easier updates, and orchestrator portability (GitHub → Azure DevOps without rewriting business logic).
