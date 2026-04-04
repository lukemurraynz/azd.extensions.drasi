---
applyTo: "**/azure-pipelines.yml, **/azure-pipelines*.yml, **/*.pipeline.yml"
description: "Best practices for Azure DevOps Pipeline YAML files"
---

# Azure DevOps Pipeline YAML Best Practices

**IMPORTANT**: Use `microsoft.learn.mcp` MCP server for Azure Pipelines task reference and current task versions. Use `iseplaybook` MCP server for ISE CI/CD best practices. Do not assume task versions are current — verify with MCP or [Azure Pipelines task reference](https://learn.microsoft.com/azure/devops/pipelines/tasks/reference/).

For general YAML structure, job dependencies, caching, and secret injection patterns, see `yaml.instructions.md`. For security controls (action pinning, permissions, OIDC), see `cicd-security.instructions.md`. This file covers Azure DevOps-specific features only.

## Task Version Currency (Non-Negotiable)

Always verify task versions before use. Outdated tasks may lack security patches or features.

| Task            | Current Version | Notes                                   |
| --------------- | --------------- | --------------------------------------- |
| `UseDotNet`     | `@3`            | `@2` is deprecated                      |
| `DotNetCoreCLI` | `@2`            | Still current; verify with MCP          |
| `AzureWebApp`   | `@1`            | Check for newer versions                |
| `AzureCLI`      | `@2`            | Preferred for scripted Azure operations |

**Rule:** Before referencing any task, verify the latest version via `microsoft.learn.mcp` or the [task reference docs](https://learn.microsoft.com/azure/devops/pipelines/tasks/reference/).

## Azure DevOps-Specific Features

### Variable Groups and Secure Variables

```yaml
variables:
  - group: shared-config # Linked to Azure Key Vault or static group
  - group: $(Environment)-secrets # Dynamic group selection (requires pipeline parameter)
  - name: buildConfiguration
    value: "Release"
```

- Link variable groups to Azure Key Vault for secret management.
- Mark sensitive variables as `isSecret: true` in the group.
- Use `$(variableName)` syntax (not `${{ }}`) for runtime resolution of secrets.

### Service Connections

- Use Workload Identity federation (OIDC) for Azure service connections where available.
- Scope service connections to specific resource groups (not entire subscriptions).
- Require approval and checks on service connections used for production deployments.
- Prefer managed identities over service principal secrets.

### Environment Approvals and Checks

```yaml
stages:
  - stage: DeployProduction
    dependsOn: DeployStaging
    condition: succeeded()
    jobs:
      - deployment: Deploy
        environment: production # Requires approval gate configured in Azure DevOps
        strategy:
          runOnce:
            deploy:
              steps:
                - download: current
                  artifact: drop
```

- Configure approval gates on environments (not in YAML — managed in Azure DevOps UI).
- Use branch control checks to restrict which branches can deploy to production.
- Add business hours checks for production deployments when appropriate.

### Template Reuse (extends and includes)

```yaml
# pipeline.yml — thin shim that extends a shared template
trigger:
  branches:
    include: [main]

extends:
  template: templates/dotnet-build-deploy.yml
  parameters:
    solution: "**/*.sln"
    buildConfiguration: "Release"
    environment: "staging"
```

- Use `extends` for complete pipeline inheritance with security enforcement.
- Use `template` references for reusable step/job/stage sequences.
- Version templates in a shared repository; pin with `@refs/tags/v1.0`.
- Keep pipeline YAML files thin — logic belongs in templates.

### Agent Pool Selection

```yaml
pool:
  vmImage: "ubuntu-latest" # Microsoft-hosted (recommended for most builds)
  # OR
  name: "Self-Hosted-Pool" # Self-hosted (for network access or custom tooling)
```

- Prefer Microsoft-hosted agents (`ubuntu-latest`) unless you need private network access or custom tooling.
- For self-hosted agents, pin to specific OS versions and maintain agent updates.
- Use demands to ensure agents have required capabilities.

## Example: .NET Build and Deploy

```yaml
trigger:
  branches:
    include: [main, develop]
  paths:
    exclude: [docs/*, README.md]

variables:
  - group: shared-variables
  - name: buildConfiguration
    value: "Release"

stages:
  - stage: Build
    displayName: "Build and Test"
    jobs:
      - job: Build
        pool:
          vmImage: "ubuntu-latest"
        steps:
          - task: UseDotNet@3
            displayName: "Install .NET SDK"
            inputs:
              version: "10.x"

          - task: DotNetCoreCLI@2
            displayName: "Restore"
            inputs:
              command: "restore"
              projects: "**/*.csproj"

          - task: DotNetCoreCLI@2
            displayName: "Build"
            inputs:
              command: "build"
              projects: "**/*.csproj"
              arguments: "--configuration $(buildConfiguration) --no-restore"

          - task: DotNetCoreCLI@2
            displayName: "Test"
            inputs:
              command: "test"
              projects: "**/*Tests.csproj"
              arguments: '--configuration $(buildConfiguration) --no-restore --collect:"XPlat Code Coverage"'

  - stage: Deploy
    displayName: "Deploy to Staging"
    dependsOn: Build
    condition: and(succeeded(), eq(variables['Build.SourceBranch'], 'refs/heads/main'))
    jobs:
      - deployment: DeployToStaging
        environment: "staging"
        strategy:
          runOnce:
            deploy:
              steps:
                - download: current
                  artifact: drop
                - task: AzureWebApp@1
                  displayName: "Deploy to Azure Web App"
                  inputs:
                    azureSubscription: "staging-service-connection"
                    appType: "webApp"
                    appName: "myapp-staging"
                    package: "$(Pipeline.Workspace)/drop/**/*.zip"
```

## Anti-Patterns

| Anti-Pattern                                     | Fix                                               |
| ------------------------------------------------ | ------------------------------------------------- |
| Using deprecated task versions (`UseDotNet@2`)   | Verify current versions via `microsoft.learn.mcp` |
| Hardcoding secrets in YAML                       | Use variable groups linked to Key Vault           |
| Overly broad triggers (trigger on all branches)  | Use path and branch filters                       |
| Monolithic pipelines (500+ lines)                | Split into templates with `extends`               |
| Missing environment approval gates               | Configure approvals in Azure DevOps environments  |
| Service connections scoped to full subscriptions | Scope to resource groups                          |
