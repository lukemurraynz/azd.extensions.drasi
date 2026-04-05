---
description: "Distribution-time security: prevent secrets, system prompts, debug artifacts, and proprietary logic from leaking through published packages, Docker images, CI/CD artifacts, or git history"
applyTo: "**/package.json,**/*.csproj,**/*.fsproj,**/*.nuspec,**/pyproject.toml,**/setup.py,**/setup.cfg,**/Dockerfile,**/Dockerfile.*,**/*.dockerfile,.github/workflows/**,**/docker-compose*.yml,**/compose*.yml"
---

# Distribution-Time Security

> **MCP verification**: Use `iseplaybook` for supply-chain and CI/CD security best practices. Use `microsoft.learn.mcp` for Azure-specific secret management and container registry guidance.

This file addresses threats that occur **at build/publish/deploy time**, not runtime. Runtime security (authentication, authorization, input validation, injection prevention) is covered in `global.instructions.md` and language-specific instruction files.

**Threat model context**: The Anthropic Claude Code incident (2025) demonstrated that client-distributed npm packages can be trivially deobfuscated by LLMs, exposing source code, system prompts, and tool definitions. Obfuscation, minification, and bundling provide zero meaningful security against determined extraction. Any code, prompt, or configuration shipped to clients must be treated as fully readable.

## Core Principle

**Client-distributed code is public code.** Any artifact shipped to consumers (npm packages, NuGet packages, PyPI wheels, Docker images, browser bundles, mobile app bundles, Electron apps) must be treated as fully readable by adversaries. Design your security boundary around what stays server-side, not around obscuring what ships client-side.

---

## T1: Package Publish Safeguards

Published packages are the highest-risk distribution vector because they are designed to be downloaded and inspected.

### Rules (All Package Ecosystems)

- **Use allowlists, not denylists.** Specify exactly which files to include rather than trying to exclude sensitive files. Denylist approaches fail because new sensitive files are not excluded by default.
- **Audit package contents before every publish.** Run the ecosystem's dry-run/pack command and review the file list.
- **Never publish from local machines.** Publish only from CI/CD pipelines with controlled build environments.
- **Separate server-side and client-side code** at the module/project level. Do not rely on tree-shaking or bundler configuration to keep server code out of client packages.

### npm / Node.js

```jsonc
// package.json — allowlist pattern (preferred over .npmignore)
{
  "files": ["dist/", "README.md", "LICENSE"],
}
```

- Use the `files` field in `package.json` (allowlist) instead of `.npmignore` (denylist).
- Run `npm pack --dry-run` in CI before publish and verify the file list contains only intended files.
- Disable source maps for production npm packages: set `sourceMap: false` in `tsconfig.json` for the production build, or strip `.map` files from the `dist/` output.
- Never include `src/`, `test/`, `.env`, `prompts/`, `*.md` (except README/LICENSE/CHANGELOG), or server-side modules in the `files` list.

### NuGet / .NET

```xml
<!-- In .csproj — control what gets packed -->
<PropertyGroup>
  <!-- Exclude PDBs from the package; use a symbol server instead -->
  <IncludeSymbols>false</IncludeSymbols>
  <DebugType>embedded</DebugType> <!-- Or 'none' for published packages -->
</PropertyGroup>

<ItemGroup>
  <!-- Prevent content files from being packed unless explicitly intended -->
  <Content Update="prompts\**" Pack="false" />
  <Content Update="appsettings.*.json" Pack="false" />
  <None Update="**\*.md" Pack="false" />
</ItemGroup>
```

- Run `dotnet pack --no-build` then inspect the `.nupkg` (it is a ZIP) for unintended files before publish.
- Never pack `appsettings.*.json`, `prompts/`, embedded resources containing secrets or system prompts, or PDB files.
- Use symbol servers (Azure Artifacts, NuGet.org symbol server) for debug symbols instead of shipping PDBs in packages.
- Set `<PrivateAssets>all</PrivateAssets>` on analyzer, test, and internal-only package references to prevent transitive leakage.

### PyPI / Python

```toml
# pyproject.toml
[tool.setuptools.packages.find]
include = ["mypackage*"]
exclude = ["tests*", "prompts*", "scripts*"]

[tool.setuptools.package-data]
mypackage = ["py.typed"]
# Do NOT include: *.md, *.env, prompts/*, config/*
```

- Use explicit `include` in package discovery. Never rely on default auto-discovery for packages with server-side components.
- Run `python -m build --sdist` then inspect the archive for unintended files.
- Use `MANIFEST.in` with `prune` directives for additional denylist safety, but prefer structural separation (server code in a separate non-published package).

---

## T2: Docker Image Hygiene

Docker images are distribution artifacts. Even "private" images in a container registry are accessible to anyone with registry pull access.

### Rules

- **Use multi-stage builds.** The final stage must contain only the runtime, compiled output, and required configuration. Build tools, source code, test files, and prompt files must not be copied to the final stage.
- **Use BuildKit `--mount=type=secret`** for build-time secrets (registry credentials, API keys, private feeds). Secrets mounted this way never persist in image layers.
- **Inspect images before push.** Run `docker history <image>` and/or use [`dive`](https://github.com/wagoodman/dive) in CI to verify no secrets, source maps, PDBs, prompt files, or source code appear in the final image layers.
- **Exclude sensitive files from build context** via `.dockerignore`:

```dockerignore
# .dockerignore — prevent sensitive files from entering the build context
.git
.env
.env.*
**/*.pdb
**/*.map
**/prompts/
**/*.agent.md
**/SKILL.md
**/.github/
**/appsettings.Development.json
**/appsettings.Local.json
**/*.pfx
**/*.key
**/node_modules/
```

### BuildKit Secret Mounts

```dockerfile
# syntax=docker/dockerfile:1
# Build-time secret that never persists in layers
RUN --mount=type=secret,id=nuget_token \
    dotnet restore --configfile /run/secrets/nuget_token
```

```bash
# Build command
DOCKER_BUILDKIT=1 docker build --secret id=nuget_token,src=./nuget.config .
```

### CI Image Inspection Step

```yaml
# GitHub Actions example
- name: Inspect image for leaked artifacts
  run: |
    docker history ${{ env.IMAGE_TAG }} --no-trunc
    # Fail if source maps, PDBs, or prompt files are found
    docker create --name inspect-target ${{ env.IMAGE_TAG }}
    if docker export inspect-target | tar -t | grep -E '\.(map|pdb)$|/prompts/|SKILL\.md|\.agent\.md'; then
      echo "::error::Production image contains debug artifacts or prompt files"
      exit 1
    fi
    docker rm inspect-target
```

---

## T3: Git History Secret Remediation

Secrets "deleted" from the current tree still exist in git history and are accessible to anyone with repository clone access.

### Prevention

- Enable **GitHub push protection** (or equivalent) to block secret commits before they reach the remote.
- Run **gitleaks** (or equivalent) as a pre-commit hook and in CI on every PR.
- Run a **full-repository gitleaks scan** on release branches before tagging releases (not just PR-scoped incremental scans).

### Remediation (After Accidental Commit)

If a secret was committed, even if subsequently "deleted":

1. **Rotate the credential immediately.** Do not wait for history cleanup. Assume it is compromised.
2. **Scrub git history** using `git filter-repo` (preferred) or BFG Repo-Cleaner. Do not use `git filter-branch` (deprecated, slow, error-prone).
3. **Force-push** the cleaned history and notify all collaborators to re-clone or rebase.
4. **Verify cleanup** by cloning a fresh copy and running gitleaks against all history.
5. **Audit access logs** for any usage of the compromised credential between commit and rotation.

```bash
# Scrub a file from all history (after rotating the secret)
git filter-repo --path secrets.json --invert-paths
git push --force-with-lease --all
git push --force-with-lease --tags
```

---

## T4: CI/CD Artifact and Log Protection

CI/CD pipelines can leak secrets through build artifacts, log output, and workflow artifacts.

### Rules

- **Validate artifact content before upload.** Before `actions/upload-artifact` (or equivalent), verify the artifact does not contain secrets, source code, PDBs, `.map` files, or prompt files.
- **Never echo secrets to logs.** GitHub Actions masks `${{ secrets.* }}` automatically, but custom variables derived from secrets, base64-encoded secrets, and secrets in error messages are NOT masked. Use `::add-mask::` for derived values.
- **Set artifact retention limits.** Use `retention-days` on `actions/upload-artifact` to limit exposure window. Default retention of 90 days is excessive for most build artifacts.
- **Scope gitleaks scans** appropriately: PR-scoped for incremental checks, full-repo for release branches.

```yaml
# Mask derived secrets
- name: Configure deployment
  run: |
    CONNECTION_STRING=$(az keyvault secret show --name db-conn --vault-name $VAULT --query value -o tsv)
    echo "::add-mask::$CONNECTION_STRING"
    echo "DB_CONNECTION=$CONNECTION_STRING" >> "$GITHUB_ENV"

# Upload with retention limit and content validation
- name: Upload build artifacts
  run: |
    # Verify no secrets or debug artifacts in upload directory
    if find ./artifacts -name '*.pdb' -o -name '*.map' -o -name '.env' -o -name 'appsettings.Development.json' | head -1 | grep -q .; then
      echo "::error::Artifacts contain debug/secret files"
      exit 1
    fi
- uses: actions/upload-artifact@v4
  with:
    name: build-output
    path: ./artifacts
    retention-days: 7
```

### Full-Repo Secret Scan on Release

```yaml
# Run on release branches, not just PRs
- name: Full-repo gitleaks scan
  if: startsWith(github.ref, 'refs/heads/release/') || startsWith(github.ref, 'refs/tags/')
  uses: gitleaks/gitleaks-action@v2
  with:
    args: detect --source . --log-opts="--all"
  env:
    GITLEAKS_LICENSE: ${{ secrets.GITLEAKS_LICENSE }}
```

---

## T5: AI System Prompt and Tool Definition Protection

AI agent system prompts, tool definitions, and orchestration instructions are intellectual property that can reveal competitive strategy, security posture, and system architecture.

### Prompt Classification

Classify every prompt and agent instruction as one of:

| Classification   | Definition                                                                                                       | Storage                                                                         | Ships in artifact? |
| ---------------- | ---------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------- | ------------------ |
| **Public**       | Formatting templates, user-facing help text, response shaping that provides no competitive advantage             | Version-controlled files, string literals                                       | Yes (acceptable)   |
| **Confidential** | Competitive IP, security-sensitive instructions, business logic, orchestration strategy, tool selection criteria | External config store (Azure App Configuration, Key Vault, Foundry agent store) | **No**             |

### Rules

- **Default to Confidential.** Unless a prompt is explicitly classified as Public, treat it as Confidential.
- **Confidential prompts must not exist in published artifacts.** This means they must not be in Docker image layers, NuGet packages, npm packages, compiled assemblies (as string literals), or embedded resources.
- **Load Confidential prompts at runtime** from:
  - Azure App Configuration (with managed identity)
  - Azure Key Vault (for short, high-sensitivity prompts)
  - Foundry agent store (`CreateAgentAsync` with externally-loaded instructions)
  - Server-side API endpoint (for complex prompt assemblies)
- **Foundry hosted agents**: `CreateAgentAsync(instructions: ...)` stores the prompt server-side after creation, but the prompt string still exists in the client binary if hardcoded. Load the instruction text from App Configuration or Key Vault at bootstrap.
- **MCP tool catalogs**: MCP server URLs, allowed tool names, and topology defined in source code are visible in compiled artifacts. For sensitive architectures, load tool configuration from App Configuration or environment variables injected at deploy time.
- **Prompt files in Docker images**: If `prompts/*.md` files are `COPY`'d into the Docker image, they are extractable via `docker cp`, layer inspection, or `docker export`. Exclude them from `.dockerignore` if Confidential, or load from a mounted volume or config store at runtime.
- **String literals in compiled code**: Prompts embedded as C# raw string literals or TypeScript template literals are visible via ILSpy, `dotnet-ildasm`, or JavaScript source inspection. Use `IPromptProvider` (or equivalent) backed by an external store for Confidential prompts.

### Anti-Patterns

- Storing Confidential prompts in `prompts/*.md` and `COPY`'ing them into Docker images
- Hardcoding system prompt text in `CreateAgentAsync(instructions: "You are a...")` calls
- Using `<EmbeddedResource>` for prompt `.md` files that compile into the assembly
- Relying on minification or obfuscation to hide prompts in JavaScript bundles
- Treating all prompts the same (either over-protecting Public prompts or under-protecting Confidential ones)

---

## T6: Obfuscation Is Not Security

**Position statement**: Obfuscation, minification, and code bundling are build optimizations, not security controls. They provide zero meaningful protection against extraction of source code, prompts, secrets, or business logic from client-distributed artifacts.

### Context

LLMs can deobfuscate minified JavaScript, reverse-compile .NET assemblies, and reconstruct readable source code from bundled artifacts with minimal effort. The cost of reverse engineering dropped from "skilled specialist spending days" to "anyone with LLM access spending minutes."

### Rules

- **Never rely on obfuscation to protect intellectual property or secrets.** If something must stay confidential, keep it server-side.
- **Never store secrets, API keys, or credentials in client-distributed code** with the assumption that obfuscation hides them.
- **Do not use code obfuscation tools as a security measure.** Use them for code size optimization only, and document that they provide no security benefit.
- **Design for transparency**: assume all client code is open source from a security perspective. Build competitive advantage in model quality, data, infrastructure, and user experience, not in prompt secrecy.

---

## Pre-Commit and CI Integration

### Pre-Commit Hooks

The `.pre-commit-config.yaml` should include:

```yaml
- repo: https://github.com/pre-commit/pre-commit-hooks
  hooks:
    - id: detect-private-key # Catch private keys before commit
    - id: check-added-large-files
      args: ["--maxkb=500"] # Flag suspiciously large files

- repo: https://github.com/gitleaks/gitleaks
  hooks:
    - id: gitleaks # Detect secrets in staged changes
```

### CI Validation Checklist

Before any publish or deploy step in CI:

1. Source maps (`.map` files) are not in production bundles or packages
2. PDB files are not in NuGet packages or Docker images (use symbol servers)
3. `prompts/` directory is not in npm `files` list, NuGet package, or Docker final stage
4. `.env` files are not in any published artifact
5. `appsettings.Development.json` / `appsettings.Local.json` are not in Docker images
6. `SKILL.md`, `.agent.md`, and agent instruction files are not in Docker images
7. Full-repo gitleaks scan passes on release branches
8. Docker image inspection (`docker history` or `dive`) shows no leaked layers

---

## Cross-References

- [global.instructions.md](global.instructions.md) — Core security principles
- [docker.instructions.md](coding-standards/docker/docker.instructions.md) — Multi-stage builds, layer security
- [cicd-security.instructions.md](coding-standards/yaml/cicd-security.instructions.md) — Pipeline secret hygiene, SBOM, signing
- [typescript.instructions.md](coding-standards/typescript/typescript.instructions.md) — npm publish safeguards, source maps
- [csharp.instructions.md](coding-standards/csharp/csharp.instructions.md) — NuGet publish, PDB strategy, agent prompt classification
- [secret-management SKILL](../skills/secret-management/SKILL.md) — Runtime secret access patterns
- [threat-modelling SKILL](../skills/threat-modelling/SKILL.md) — Distribution trust boundary and threat catalogue
