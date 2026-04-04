---
applyTo: ".github/workflows/**"
description: "DevSecOps pipeline security hardening: supply chain integrity, dependency scanning, SBOM generation, container signing, and secret hygiene for GitHub Actions workflows"
---

# CI/CD Security Hardening Instructions

These rules apply to all GitHub Actions workflows. They complement `yaml.instructions.md` (formatting/structure) and `github-actions-ci-cd` (build/deploy patterns) by adding the security controls that Microsoft 1ES / SDL pipelines enforce by default.

## Precedence

Repo standards → `yaml.instructions.md` → this file → MCP servers (`iseplaybook`, `microsoft.learn.mcp`)

---

## 1. Action Pinning & Supply Chain

### Pin ALL third-party actions to full SHA

```yaml
# ✅ Pinned — immune to tag hijacking
- uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1

# ❌ Tag-only — attacker can retag to malicious commit
- uses: actions/checkout@v4
```

**Rule:** Every `uses:` line MUST include the full 40-character commit SHA followed by a `# vX.Y.Z` comment for readability. The only exception is `actions/` org actions in trusted internal repos.

### Restrict fork PR permissions

```yaml
# Prevent forks from accessing secrets or elevated permissions
on:
  pull_request_target: # ⚠️ Avoid unless you need write access — use pull_request instead
```

**Rule:** Prefer `pull_request` over `pull_request_target`. If `pull_request_target` is required, the workflow MUST NOT check out or execute PR code directly.

### Disable credential persistence on checkout

```yaml
# ✅ Credentials not written to .git/config — reduces token exfiltration surface
- uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1
  with:
    persist-credentials: false

# ❌ Default — persists GITHUB_TOKEN to .git/config where any subsequent step can read it
- uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1
```

**Rule:** ALWAYS set `persist-credentials: false` on `actions/checkout`. Only omit when the job explicitly needs to push commits in-workflow (e.g., auto-fix commit jobs).

---

## 2. Permissions — Least Privilege

```yaml
# ✅ Top-level: restrict everything, then open per-job
permissions: {}

jobs:
  build:
    permissions:
      contents: read
      packages: write # Only if pushing packages
      id-token: write # Only if using OIDC
      attestations: write # Only if generating attestations
```

**Rules:**

- Top-level `permissions: {}` (empty) is REQUIRED — open only what each job needs
- `id-token: write` is REQUIRED for OIDC Azure login (no stored credentials)
- Never grant `contents: write` unless the job creates commits/tags
- `pull-requests: write` only in jobs that post PR comments

---

## 3. Authentication — OIDC, Never Stored Credentials

```yaml
# ✅ OIDC — no secrets stored, auto-rotating, auditable
- uses: azure/login@a457da9ea143d694b1b9c7c869ebb04ebe844ef5 # v2.3.0
  with:
    client-id: ${{ vars.AZURE_CLIENT_ID }}
    tenant-id: ${{ vars.AZURE_TENANT_ID }}
    subscription-id: ${{ vars.AZURE_SUBSCRIPTION_ID }}

# ❌ Service principal with stored secret — rotatable but risky
- uses: azure/login@v2
  with:
    creds: ${{ secrets.AZURE_CREDENTIALS }}
```

**Rules:**

- Use OIDC (`id-token: write`) for ALL Azure authentication
- Store `client-id`, `tenant-id`, `subscription-id` as repository **variables** (not secrets — they aren't sensitive)
- If a third-party service doesn't support OIDC, use short-lived tokens via Key Vault with `az keyvault secret show`

---

## 4. Dependency Scanning

### .NET

```yaml
- name: Check for vulnerable NuGet packages
  run: |
    dotnet list package --vulnerable --include-transitive 2>&1 | tee vuln-report.txt
    if grep -qi "has the following vulnerable packages" vuln-report.txt; then
      echo "::error::Vulnerable NuGet packages detected"
      exit 1
    fi
```

### Node.js / npm

```yaml
- name: Audit npm dependencies
  run: npm audit --audit-level=high
  # Use --audit-level=critical for less strict enforcement
```

### Python

```yaml
- name: Check for vulnerable Python packages
  run: |
    pip install pip-audit
    pip-audit --strict --desc
```

### GitHub Dependabot (repo-level)

Ensure `.github/dependabot.yml` exists:

```yaml
version: 2
updates:
  - package-ecosystem: "nuget"
    directory: "/"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 10

  - package-ecosystem: "npm"
    directory: "/frontend"
    schedule:
      interval: "weekly"

  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
    # ⚠️ Critical — catches malicious action updates

  - package-ecosystem: "docker"
    directory: "/"
    schedule:
      interval: "weekly"
```

**Rule:** Every repository MUST have a `dependabot.yml` that covers all ecosystems in use, including `github-actions`.

---

## 5. SBOM Generation

```yaml
- name: Generate SBOM
  uses: anchore/sbom-action@61119d458adab75f756bc0b9e4bde25725f86a7a # v0.17.2
  with:
    artifact-name: sbom-${{ github.sha }}.spdx.json
    output-file: sbom.spdx.json
    format: spdx-json

- name: Upload SBOM as artifact
  uses: actions/upload-artifact@v4
  with:
    name: sbom
    path: sbom.spdx.json
    retention-days: 90
```

**Rules:**

- Production release workflows MUST generate an SBOM
- Use SPDX or CycloneDX format
- Attach the SBOM as a build artifact and optionally publish to a dependency graph

---

## 6. Container Image Signing & Attestation

```yaml
- name: Build and push image
  id: build
  run: |
    docker build -t ${{ env.REGISTRY }}/${{ env.IMAGE }}:${{ github.sha }} .
    docker push ${{ env.REGISTRY }}/${{ env.IMAGE }}:${{ github.sha }}

- name: Generate artifact attestation
  uses: actions/attest-build-provenance@1c608d11d69870c2092266b3f9a6f3abbf17002c # v1.4.3
  with:
    subject-name: ${{ env.REGISTRY }}/${{ env.IMAGE }}
    subject-digest: ${{ steps.build.outputs.digest }}
    push-to-registry: true
```

**Rules:**

- Production container images SHOULD have build provenance attestations (SLSA Level 2+)
- Use `actions/attest-build-provenance` for GitHub-native attestation
- For cosign signing, use `sigstore/cosign-installer` + keyless signing with OIDC

---

## 7. Secret Hygiene

```yaml
# ✅ Environment-scoped secrets with required reviewers
jobs:
  deploy:
    environment: production # Requires approval before secrets are exposed
    steps:
      - run: deploy.sh
        env:
          DB_CONNECTION: ${{ secrets.DB_CONNECTION_STRING }}
```

**Rules:**

- Never echo, log, or write secrets to files — GitHub masks `${{ secrets.* }}` in logs but not in file outputs
- Use **environment protection rules** for production secrets (required reviewers, wait timers)
- Rotate secrets on a defined schedule (90 days recommended)
- Prefer OIDC and managed identity over stored secrets wherever possible
- Never pass secrets via command-line arguments (visible in process listings)

### Secret scanning

Enable GitHub secret scanning and push protection at the repository or organization level:

- **Settings → Code security and analysis → Secret scanning → Enable**
- **Push protection → Enable** — blocks pushes containing detected secrets

### Gitleaks secret scanning with SARIF (IaC / all-language repos)

For IaC repositories and mixed-language repos where GitHub secret scanning may not cover all file types, add Gitleaks to the validate stage:

```yaml
- name: Run Gitleaks
  run: |
    gitleaks detect \
      --source . \
      --report-format sarif \
      --report-path gitleaks-report.sarif \
      --exit-code 1 \
      --log-level info

- uses: actions/upload-artifact@v4
  if: success() || failure()
  with:
    name: gitleaks-report
    path: gitleaks-report.*

- name: Upload Gitleaks SARIF to code scanning
  uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: gitleaks-report.sarif
```

**Rules:**

- Run Gitleaks in the **validate stage** — before any deploy step
- Upload SARIF to GitHub Code Scanning so findings appear in the Security tab
- Set `--exit-code 1` to fail the pipeline on findings
- For PR workflows, scan only PR commits: pass `--log-opts "origin/${{ github.base_ref }}..${{ github.sha }}"` to avoid rescanning the entire history each run
- **For release branches and tags**, run a full-repository scan (no `--log-opts`) to catch secrets anywhere in history:

```yaml
- name: Full-repo gitleaks scan (release gate)
  if: startsWith(github.ref, 'refs/heads/release/') || startsWith(github.ref, 'refs/tags/')
  run: |
    gitleaks detect --source . --log-opts="--all" --exit-code 1
```

### Publish Pipeline Hardening

Before any `npm publish`, `dotnet nuget push`, or `docker push` step, validate artifact contents:

```yaml
- name: Audit package contents before publish
  run: |
    # npm: verify no server code, prompts, source maps, or env files
    npm pack --dry-run 2>&1 | tee /tmp/pack-list.txt
    if grep -E 'src/|test/|\.env|\.map$|prompts/|SKILL\.md|\.agent\.md' /tmp/pack-list.txt; then
      echo "::error::Package contains files that must not be published"
      exit 1
    fi
```

**Rules:**

- Run dry-run/pack audit in CI before every publish step
- Fail the pipeline if server-side code, prompts, debug artifacts (`.map`, `.pdb`), or configuration files appear in the package
- For Docker images, add an image inspection step (see `docker.instructions.md` Section 6)
- See `distribution-security.instructions.md` for the full framework

### Log Sanitization Beyond Built-In Masking

GitHub Actions masks `${{ secrets.* }}` in logs automatically, but derived values, base64-encoded secrets, and secrets in error messages are NOT masked.

- Use `::add-mask::` for any variable derived from a secret
- Redirect sensitive command output to `/dev/null` or a file (not stdout)
- Set `retention-days` on `actions/upload-artifact` to limit exposure window (default 90 days is excessive)
- Never pipe secret values through `jq`, `sed`, or string manipulation without masking the result

```yaml
- name: Mask derived secret
  run: |
    TOKEN=$(az keyvault secret show --name api-key --vault-name $VAULT --query value -o tsv)
    echo "::add-mask::$TOKEN"
    echo "API_TOKEN=$TOKEN" >> "$GITHUB_ENV"
```

---

## 8. CodeQL / Static Analysis

```yaml
- name: Initialize CodeQL
  uses: github/codeql-action/init@v3
  with:
    languages: csharp, javascript

- name: Build (for compiled languages)
  run: dotnet build --no-restore

- name: Perform CodeQL analysis
  uses: github/codeql-action/analyze@v3
  with:
    category: "/language:csharp"
```

**Rules:**

- Enable CodeQL for all supported languages in the repository
- Run on PR and on a weekly schedule (catches new vulnerability patterns)
- Use default query suites; add `security-extended` for higher sensitivity
- Review and triage findings — don't just suppress

---

## 9. Quality Gates for Recurring Defect Classes

Add non-security static checks to block high-frequency regressions:

- React hooks correctness (`react-hooks/rules-of-hooks`, `react-hooks/exhaustive-deps`) as errors.
- Empty catch/no-op error handlers blocked (`no-empty`, equivalent analyzers).
- Promise rejection handling checks enabled (`no-floating-promises` or equivalent).
- Complexity and maintainability budgets enforced for non-generated code.

Generated artifact exclusions are REQUIRED for maintainability gates:

- Exclude known generated files (for example EF migration designer snapshots) from complexity/function-length checks.
- Do not exclude generated files from secret scanning or vulnerability scanning.

---

## 10. Workflow Hardening Checklist

Before merging any workflow change, verify:

| #   | Check                                       | Enforcement                                                                   |
| --- | ------------------------------------------- | ----------------------------------------------------------------------------- |
| 1   | All third-party actions pinned to SHA       | `grep -P 'uses:.*@(?![a-f0-9]{40})' .github/workflows/` should return nothing |
| 2   | Top-level `permissions: {}` set             | Every workflow file                                                           |
| 3   | OIDC for Azure auth                         | No `AZURE_CREDENTIALS` secret in use                                          |
| 4   | Dependency scanning in CI                   | At least one of: `dotnet list --vulnerable`, `npm audit`, `pip-audit`         |
| 5   | Dependabot configured for all ecosystems    | `.github/dependabot.yml` exists and covers the repo                           |
| 6   | Secret scanning + push protection enabled   | Repository settings                                                           |
| 7   | SBOM generated for release builds           | `sbom-action` or equivalent in release workflow                               |
| 8   | No `pull_request_target` with code checkout | Audit all workflows                                                           |
| 9   | Environment protection for production       | `environment: production` with required reviewers                             |
| 10  | CodeQL or equivalent SAST enabled           | Scheduled + PR trigger                                                        |
| 11  | `persist-credentials: false` on checkout    | All `actions/checkout` steps except auto-commit jobs                          |
| 12  | Gitleaks SARIF uploaded to code scanning    | IaC and mixed-language repos (validate stage)                                 |

---

## Related

- [yaml.instructions](yaml.instructions.md) — YAML formatting and pipeline structure
- [github-actions-ci-cd](../skills/github-actions-ci-cd/SKILL.md) — Build/deploy workflow patterns
- [github-actions-terraform](../skills/github-actions-terraform/SKILL.md) — Terraform CI/CD specifics
- [secret-management](../skills/secret-management/SKILL.md) — Key Vault, CSI driver, rotation
