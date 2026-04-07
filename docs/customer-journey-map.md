# Customer journey map: azd.extensions.drasi

## Persona

**Alex, Azure Platform Engineer**

Alex is a senior platform engineer at a mid-size SaaS company that runs its operational workloads on Azure. The team owns the data platform and is responsible for building event-driven pipelines that propagate changes from production databases to downstream consumers: analytics services, notification systems, and operational dashboards.

**Job to be done:** When a transactional change happens in a data source (Cosmos DB, PostgreSQL, Event Hub), I want to detect it, evaluate it against a standing query, and trigger a downstream reaction, so that I can eliminate polling loops and ad-hoc Functions glue code from the platform.

**Background and context:**

- Comfortable with Kubernetes, AKS, and Azure IaC (Bicep, Terraform).
- Already uses `azd` for other project scaffolding.
- Familiar with CDC concepts from Debezium; has seen Drasi mentioned in CNCF Sandbox listings and a Microsoft blog post.
- Works inside a GitHub Actions CI/CD pipeline and expects infrastructure to be reproducible and code-reviewed.

**Goals:**

- Replace bespoke Event Grid + Azure Functions pipelines with a consistent, observable reactive data layer.
- Keep secrets out of source control.
- Support staging and production environments from the same codebase.
- Monitor pipeline health without custom Grafana dashboards.

**Frustrations going in:**

- Prior attempts with Debezium required running a Kafka cluster just to get CDC working.
- Custom Functions pipelines are hard to test offline and have no concept of a "continuous query."
- Infrastructure for reactive pipelines tends to be glued together manually and is not repeatable.

---

## Journey stages

### Stage 1: Awareness

**Description:** Alex discovers Drasi and the azd extension while researching CDC and reactive data pipeline options for Azure.

**Touchpoints:**

- CNCF Sandbox project listing for Drasi under the Microsoft organization.
- `drasi.io` documentation site explaining Sources, Continuous Queries, and Reactions.
- Microsoft Open Source Blog post describing the azd extension.
- `drasi-project` GitHub organization, leading to `lukemurraynz/azd.extensions.drasi`.
- YouTube talk from a KubeCon session or Azure Developers conference covering Drasi use cases.
- A peer in a Kubernetes Slack workspace mentions the azd extension for deploying Drasi on AKS.

**User actions:**

- Reads the CNCF Sandbox listing and follows the link to `drasi.io`.
- Browses the GitHub repository and scans the README.
- Searches for comparisons between Drasi, Debezium, and custom Event Grid pipelines.
- Watches a demo video showing a continuous query against a Cosmos DB source triggering a SignalR reaction.

**Thoughts and questions:**

- "Is this production-ready or still an early experiment?"
- "How different is this from just wiring up Debezium with a Kafka consumer?"
- "Do I still have to run my own Kafka cluster, or does Drasi replace that entirely?"
- "What does the azd integration actually buy me compared to running Drasi directly?"

**Emotions:** Cautious curiosity. The concept is compelling, but CNCF Sandbox status means it may still be rough. Interest increases when Alex reads that the extension provisions AKS and Key Vault from a single command.

**Pain points:**

- The connection between the Drasi project site, the CNCF listing, and the azd extension GitHub repository is not immediately obvious. Three separate entry points with different levels of detail.
- Documentation on `drasi.io` describes raw Drasi concepts without azd integration context; the azd extension README is where the full workflow becomes clear.

**Opportunities:**

- A single landing page or docs site page that explains the Drasi + azd extension workflow end-to-end.
- A comparison table against Debezium and custom Functions pipelines on the extension README.

---

### Stage 2: Consideration

**Description:** Alex evaluates whether the azd extension approach fits the team's workflow and compares it against alternatives.

**Touchpoints:**

- `lukemurraynz/azd.extensions.drasi` README: prerequisites table, features list, command reference, and example scenarios.
- `specs/001-azd-drasi-extension/quickstart.md`: step-by-step walkthrough from scaffold to status.
- `docs/troubleshooting.md`: error code reference confirming that structured errors exist.
- GitHub Issues and Discussions on the `drasi-project` org.
- Internal team conversation about AKS cost and operational overhead vs. a serverless alternative.

**User actions:**

- Reviews the prerequisites table: `azd >= 1.10.0`, `drasi >= 0.10.0`, `Azure CLI >= 2.60.0`, `Go >= 1.22`, `Docker >= 24.0`, `kubectl >= 1.28`.
- Checks the team's CI environment against those versions.
- Reads the "Why this exists" section and the features list.
- Explores available templates: `cosmos-change-feed`, `event-hub-routing`, `query-subscription`, `postgresql-source`, `blank`, `blank-terraform`.
- Considers the Terraform template as a potential path for teams with existing Terraform state.
- Evaluates the Key Vault secret reference pattern (no secrets in source control) against the team's security requirements.
- Reads the error code reference to assess how debuggable failures will be.

**Thoughts and questions:**

- "Six prerequisites is more than I expected. We'll need Go and kubectl on the CI runner."
- "The `blank-terraform` template is interesting. Can I use this alongside our existing Terraform modules?"
- "The `cosmos-change-feed` template is exactly our use case. What does the generated query YAML look like?"
- "Key Vault secret references are reassuring. That aligns with how we handle secrets everywhere else."
- "Does `azd drasi provision` idempotent? Can I safely re-run it if something fails mid-way?"

**Emotions:** Analytical and measured. The feature set is genuinely appealing. The prerequisites list introduces the first moment of pause: the Go requirement is unexpected for what appears to be a CLI tool, not an application.

**Pain points:**

- Six prerequisites with specific minimum versions. Some (Go, kubectl) are not standard in all CI environments and require a setup step before any Drasi work can happen.
- OIDC/Workload Identity configuration is mentioned as a requirement but not fully explained in the README.
- Template contents are not previewed in the documentation; Alex has to install the extension to see what the generated YAML looks like.

**Opportunities:**

- Inline the generated directory structure for each template in the README (the quickstart does this for `cosmos-change-feed`; the README does not).
- Add a CI environment setup section documenting how to install all prerequisites in a GitHub Actions runner.
- Clarify whether Go is needed as an application dependency or only as a build tool for the extension itself.

---

### Stage 3: Acquisition

**Description:** Alex decides to proceed, installs the extension, and verifies the local setup is working.

**Touchpoints:**

- Terminal / shell environment.
- `azd extension install azd-drasi` from the official azd extensions registry (or via `azd extension source add` for GitHub Releases).
- `azd drasi --help` to confirm installation.
- `azd drasi version` to confirm the version matches expectations.
- The devcontainer configuration in `.devcontainer/devcontainer.json` as an alternative that pre-installs all tools.

**User actions:**

- Runs `azd extension install azd-drasi`.
- Runs `azd drasi --help` to confirm the command surface.
- Checks `drasi --version` and `az version` to confirm prerequisites meet the minimum.
- Opens the repository in VS Code and considers the devcontainer as a way to avoid local tool setup.
- Shares the devcontainer approach with a colleague as a faster onramp.

**Thoughts and questions:**

- "The install itself was fast. Good."
- "The devcontainer pre-installs everything. That would have saved me 20 minutes on the prerequisites."
- "I should verify my `drasi` CLI version before going further."

**Emotions:** Relieved that the install step itself is simple. Mild frustration if any prerequisite version check fails immediately after installation, especially `ERR_DRASI_CLI_NOT_FOUND` or `ERR_DRASI_CLI_VERSION`, which surface only at runtime rather than at install time.

**Pain points:**

- Prerequisite versions are only checked at runtime, not at install time. A developer can successfully install the extension and not discover a missing or outdated `drasi` CLI until they run their first command.
- The devcontainer is the best onramp for zero-friction setup, but it is documented in the README under a "Build and run" section that reads as contributor-focused rather than as a user recommendation.

**Opportunities:**

- Run a prerequisite check as part of `azd extension install` or surface it as a first-time `azd drasi check` command.
- Promote the devcontainer in the README's "Quick start" section as the recommended path for new users.

---

### Stage 4: Onboarding

**Description:** Alex scaffolds a project, configures the Drasi manifest files, validates offline, and provisions infrastructure for the first time.

**Touchpoints:**

- `azd drasi init --template cosmos-change-feed`: generates `drasi/`, `infra/`, `azure.yaml`.
- Generated files: `drasi/drasi.yaml`, `drasi/sources/cosmos-source.yaml`, `drasi/queries/order-changes.yaml`, `drasi/reactions/pubsub-reaction.yaml`, `infra/main.bicep`, `infra/modules/aks.bicep`, `infra/modules/keyvault.bicep`.
- `azd drasi validate` and `azd drasi validate --strict`: offline validation before any cloud call.
- `azd auth login`: Azure credential setup.
- `azd drasi provision`: 8-12 minute provisioning of AKS, Key Vault, UAMI, Log Analytics, and the Drasi runtime.
- `postprovision` lifecycle hook: `waitForDrasiReady` health check runs automatically.

**User actions:**

- Runs `azd drasi init --template cosmos-change-feed` from a new project directory.
- Opens `drasi/sources/cosmos-source.yaml` and sets the Cosmos DB connection string as a Key Vault secret reference.
- Edits `drasi/queries/order-changes.yaml` to write an openCypher continuous query matching order document changes.
- Runs `azd drasi validate` to check for `ERR_VALIDATION_FAILED` or `ERR_MISSING_REFERENCE` before touching the cluster.
- Runs `azd drasi validate --strict` to catch warnings as errors before provisioning.
- Runs `azd auth login` and confirms the correct subscription is selected.
- Runs `azd drasi provision` and watches the log output for the 8-12 minute provisioning window.
- Sees the `waitForDrasiReady` hook confirm that the Drasi API is ready before the command exits.

**Thoughts and questions:**

- "The generated YAML is clean and annotated. The openCypher query stub gives me a real starting point."
- "Offline validation is exactly what I wanted. I can iterate on the query syntax without wasting a cluster call."
- "Twelve minutes is a long time to wait to find out whether the provision worked."
- "I need to configure a FederatedIdentityCredential for the Drasi service account. Where does that happen?"
- "The Bicep modules look reasonable. Can I extend `main.bicep` without conflicting with what the extension manages?"

**Emotions:** Engaged and productive during YAML editing and offline validation. A clear wait-and-hope feeling during provisioning. Anxiety peaks around minute 8 if no progress indicators are visible in the terminal.

**Pain points:**

- Eight to twelve minutes of provisioning is unavoidable for AKS, but the user experience during that window relies heavily on whatever log output `azd drasi provision` streams. A silent provision or one with sparse output makes it feel longer.
- The FederatedIdentityCredential configuration for `system:serviceaccount:drasi-system:drasi-api` is handled by the provision step but is not clearly documented as automatic in the quickstart. Engineers who read the Drasi docs directly may try to configure it manually and collide with what the extension does.
- Environment overlays (`drasi/environments/<name>.yaml`) cannot be used to remove components in a lower environment, only to override parameters. This limitation is not surfaced until a developer tries to scope down a production manifest for a dev environment.

**Moments of truth:**

- `azd drasi validate` returning clean output is the first strong signal that the tool works as described. This is often the moment Alex decides the extension is worth the investment.
- `azd drasi provision` completing successfully is the point of no return: Alex now has real Azure resources and a working Drasi runtime.

**Opportunities:**

- Stream provisioning progress in named phases ("Creating AKS cluster...", "Installing Drasi runtime...", "Running health check...") so engineers can track where they are in the 8-12 minute window.
- Add a note to the quickstart clarifying that `FederatedIdentityCredential` configuration is handled automatically by `provision`.
- Document the environment overlay limitation clearly: overlays support parameter overrides, not component removal.

---

### Stage 5: Engagement

**Description:** Alex has a working Drasi pipeline and is using it regularly, iterating on queries, deploying to multiple environments, and building the pattern into team workflows.

**Touchpoints:**

- `azd drasi deploy` and `azd drasi deploy --dry-run`: applying component changes.
- `azd drasi status` and `azd drasi status --kind continuousquery --output json`: monitoring component health.
- `azd drasi logs --kind continuousquery --component order-changes`: streaming live query output.
- `drasi/environments/staging.yaml` and `drasi/environments/prod.yaml`: environment overlays with per-environment Key Vault references.
- `azd drasi deploy --environment staging` and `azd drasi deploy --environment prod`: multi-environment deploys.
- GitHub Actions workflow using `azd drasi validate` as a pre-provision gate in CI.
- `azd drasi deploy --output json` piped into monitoring scripts.

**User actions:**

- Iterates on openCypher query syntax using `azd drasi validate` before each deploy.
- Uses `azd drasi deploy --dry-run` to preview component changes in a staging environment.
- Adds a second continuous query (`azd drasi init --template query-subscription`) as a new use case emerges.
- Sets up `drasi/environments/staging.yaml` with staging-specific Key Vault references and confirms the overlay applies correctly.
- Integrates `azd drasi validate` into the GitHub Actions CI pipeline as a pre-deploy gate.
- Uses `--output json` on status and diagnose commands to feed a team dashboard.

**Thoughts and questions:**

- "The dry-run is valuable before touching a prod cluster."
- "The dependency ordering (sources -> queries -> middleware -> reactions) is enforced automatically. That is one less thing to think about."
- "I want to update a source configuration. The delete-then-apply pattern means the source goes offline briefly. I need to plan that."
- "The environment overlay for prod uses a different Key Vault. The overlay pattern works, but I can't remove the dev-only reaction from the prod manifest without creating a separate manifest."

**Emotions:** Satisfied with the core workflow. Friction emerges around the delete-then-apply behavior for component updates (a source or query goes offline during a redeploy) and around environment overlay limitations.

**Pain points:**

- Component updates use delete-then-apply semantics. There is no in-place update path. A source going offline during a pipeline update causes brief query interruptions, which matters in a production pipeline.
- The `ERR_COMPONENT_TIMEOUT` (5 minutes per component) and `ERR_TOTAL_TIMEOUT` (15 minutes total) limits become real constraints when deploying many components. A team that grows to 10 or more Drasi components will hit the total timeout on a full redeploy.
- Environment overlays support parameter overrides but cannot remove components. A staging manifest that inherits all production components (including production-only reactions) requires a workaround: a separate staging `drasi.yaml` or conditional logic outside the overlay system.
- The `ERR_DEPLOY_IN_PROGRESS` error occurs if a CI job retries quickly after a failure. Manual lock removal from azd environment state is not a great recovery experience.

**Opportunities:**

- Document the delete-then-apply behavior and its pipeline impact explicitly so engineers can plan maintenance windows.
- Add a `--timeout` override to `deploy` to unblock teams with large component counts.
- Extend the environment overlay model to support component inclusion/exclusion, not just parameter overrides.
- Auto-clear stale `ERR_DEPLOY_IN_PROGRESS` locks after a configurable timeout.

---

### Stage 6: Retention

**Description:** Alex has integrated the extension into the team's standard platform toolkit and is managing it long-term. Decisions about staying vs. churning happen here.

**Touchpoints:**

- `azd drasi upgrade --force`: upgrading the Drasi runtime on existing clusters.
- `azd drasi diagnose` and `azd drasi diagnose --output json`: routine and incident-triggered health checks.
- `azd drasi status --kind source` and `azd drasi logs --kind source --component my-source`: operational monitoring.
- Kubernetes network policies applied by the extension: `drasi-default-deny`, `drasi-allow-internal`, `drasi-allow-dns`, `drasi-allow-azure-api-egress`, `drasi-allow-dapr-sidecar`, `drasi-allow-k8s-api`, `drasi-allow-datastores`, `drasi-allow-common-data-egress`.
- `kubectl logs -n drasi-system -l drasi.io/component-id=<id>`: lower-level debugging when `diagnose` is not enough.
- Release notes and changelog for new Drasi CLI or extension versions.
- Community channels: GitHub Discussions, CNCF Slack.

**User actions:**

- Runs `azd drasi upgrade --force` when a new Drasi CLI version is released; confirms the `--force` gate prevents accidental upgrades.
- Uses `azd drasi diagnose` as the first step whenever a component appears stuck.
- Adds `azd drasi diagnose --output json` to a monitoring script that runs nightly.
- Investigates a `drasi-allow-datastores` policy gap when a new Cosmos DB account is added in a different region.
- Opens a GitHub Discussion when `ERR_DAPR_NOT_READY` surfaces after an AKS node pool upgrade.

**Thoughts and questions:**

- "The five-point diagnose command caught a Key Vault auth regression after a managed identity role assignment expired. That alone justified the tooling."
- "The runtime upgrade was smooth, but I want to know what changes before I run it. I need a changelog or a `--dry-run` mode for upgrade."
- "The network policies are correct in principle, but adding a new datastore requires understanding which policy to update. The docs don't cover that case."
- "How do I know when a new extension version is available? There's no notification mechanism."

**What keeps Alex using the tool:**

- Offline validation (`azd drasi validate`) before any cloud change reduces wasted provisioning cycles.
- The five-point `azd drasi diagnose` check resolves most incidents without kubectl spelunking.
- Single-command provision (`azd drasi provision`) is idempotent and re-runnable, which matters after partially failed runs.
- Key Vault secret reference translation means no secrets ever touch the manifest files.
- Environment overlays reduce duplication between staging and production configurations.
- The structured error codes (`ERR_KV_AUTH_FAILED`, `ERR_AKS_CONTEXT_NOT_FOUND`, etc.) make CI failures readable without digging into logs.

**Churn risks:**

- `ERR_COMPONENT_TIMEOUT` and `ERR_TOTAL_TIMEOUT` become regular occurrences as the number of components grows, creating a perception that the tool does not scale.
- Complex OIDC/Workload Identity/FederatedIdentityCredential requirements create recurring friction when rotating credentials, adding new service principals, or when the AKS cluster is recreated.
- The environment overlay limitation (can't remove components per environment) forces bespoke solutions that undermine the "single codebase, multiple environments" promise.
- No upgrade preview: `azd drasi upgrade --force` gives no changelog or diff before committing the upgrade.
- Drasi being a CNCF Sandbox project means API stability is not guaranteed. A breaking change in the Drasi CLI or API could force a full re-evaluation.

**Opportunities:**

- Add `azd drasi upgrade --dry-run` to preview what version change would be applied.
- Build a `azd drasi status --watch` mode that refreshes at an interval without polling manually.
- Document which Kubernetes network policies to update when adding new data source connections.
- Add an `azd extension update` notification or a version check in `azd drasi version` that flags when a newer release is available.

---

### Stage 7: Advocacy

**Description:** Alex recommends the extension to peers and presents it internally as the team's standard approach for reactive data pipelines.

**Touchpoints:**

- Internal tech talk or architecture review: Alex presents the scaffold-to-production workflow using the quickstart as a live demo.
- Pull request templates updated to require `azd drasi validate` output in PR descriptions.
- Team wiki: Alex writes a platform playbook entry documenting `azd drasi init`, the template selection guide, and the multi-environment overlay pattern.
- Peer referrals: colleagues in adjacent teams adopt the extension for `postgresql-source` and `event-hub-routing` use cases.
- GitHub star and a comment on the `lukemurraynz/azd.extensions.drasi` Discussions.
- Conference talk proposal mentioning the Drasi + azd workflow.

**User actions:**

- Runs a live demo of `azd drasi init --template cosmos-change-feed` followed by `azd drasi validate --strict` and `azd drasi deploy --dry-run` to show the workflow without spending 12 minutes on provisioning.
- Writes a platform playbook entry with recommended templates for each use case.
- Creates a PR in the team's shared `devcontainer` configuration to add the Drasi CLI and azd extension as standard tools.
- Files a GitHub Discussion asking about the extension's roadmap for supporting multiple `drasi.yaml` manifests per project.

**Thoughts and questions:**

- "The demo lands well because the `validate` and `dry-run` commands work without a live cluster. That matters for onboarding."
- "I want to direct colleagues to a 'getting started' page that doesn't require them to read the full spec."
- "I'd like to see a contributed template for Azure SQL / SQL Server CDC, which our team also uses."

**What drives advocacy:**

- The offline-first workflow (validate then dry-run) makes the extension demo-friendly and low-risk to evaluate.
- The structured error codes and `diagnose` command reduce the "it failed, now what?" problem that discourages adoption.
- Single-command provisioning is a strong demonstration of value for engineers currently managing AKS + Key Vault + Dapr setup manually.

**Pain points:**

- There is no contributed templates gallery or "which template should I use" decision guide. Peer recommendations are informal.
- The quickstart is accurate but assumes a local setup with all prerequisites. A hosted demo environment (like a GitHub Codespace badge) would reduce friction for evaluators who don't want to install six tools locally.

**Opportunities:**

- Add a GitHub Codespace "Open in Codespace" badge to the README that pre-installs all prerequisites.
- Create a templates decision guide in the docs: "Use `cosmos-change-feed` for..., use `event-hub-routing` for..., use `postgresql-source` for...".
- Publish a community templates contribution guide to encourage `azure-sql-source` and similar additions.

---

## Critical moments

### Aha moment

`azd drasi validate` returning clean output after the developer writes their first openCypher continuous query. The moment they confirm that the query structure is correct, references are resolved, and no circular dependencies exist, all without touching a cluster or spending a penny, the value proposition of the extension becomes concrete. The offline-first design is what separates this tool from raw Drasi CLI use and from custom Functions pipelines.

### Moments of truth

1. **First validate pass (Stage 4):** `azd drasi validate` runs clean. The developer commits to provisioning. This is where intent converts to investment.

2. **First successful provision (Stage 4):** `azd drasi provision` completes with the `waitForDrasiReady` hook confirming a healthy Drasi API. The developer now has a working AKS cluster and Drasi runtime provisioned from a single command. This is the point at which the "single command to production-ready infra" claim is proven or disproven.

3. **First multi-environment deploy (Stage 5):** `azd drasi deploy --environment staging` followed by `azd drasi deploy --environment prod` work from the same manifest with overlay differences. This is the moment the team validates that the tool supports their actual workflow, not just a demo.

4. **First incident recovery (Stage 6):** `azd drasi diagnose` identifies the root cause of a stuck component without requiring kubectl log spelunking. This converts a frustrating production incident into a structured debugging experience and strongly reinforces retention.

### Churn triggers

1. **Prerequisite version failure at runtime:** `ERR_DRASI_CLI_NOT_FOUND` or `ERR_DRASI_CLI_VERSION` on first command after install, without any install-time warning. The developer expected the extension to work after `azd extension install azd-drasi` and instead hits an opaque runtime error.

2. **Silent or sparse provision output:** `azd drasi provision` takes 8-12 minutes. If progress indicators are minimal, engineers conclude the command has hung and interrupt it, leaving a partially created resource group.

3. **`ERR_COMPONENT_TIMEOUT` on a growing pipeline:** Teams that grow beyond a handful of Drasi components hit component and total timeouts regularly. Without a `--timeout` override, the tool appears to have a hard cap on pipeline complexity.

4. **Environment overlay cannot remove components:** A team tries to scope a production manifest down to a staging-safe configuration using overlays and discovers overlays only support parameter overrides. The workaround (separate manifests) undermines the multi-environment workflow that drove adoption.

5. **FederatedIdentityCredential rotation complexity:** Rotating OIDC credentials or recreating the AKS cluster breaks the Workload Identity federation. Recovery requires understanding the exact FederatedIdentityCredential configuration that `provision` applies, which is not surfaced in the user-facing documentation.

6. **`ERR_DEPLOY_IN_PROGRESS` with no automatic recovery:** A CI job that retries after failure hits a stale deploy lock. Manual recovery from azd environment state is a poor experience in an automated pipeline.

---

## Journey map table

| Stage         | Touchpoint                                                                                  | User action                                                      | Emotion                                                                             | Pain point                                                                                         | Opportunity                                                               |
| ------------- | ------------------------------------------------------------------------------------------- | ---------------------------------------------------------------- | ----------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------- |
| Awareness     | CNCF Sandbox listing, `drasi.io`, GitHub README                                             | Reads docs, compares to Debezium                                 | Cautious curiosity                                                                  | Three separate entry points; no unified overview                                                   | Single landing page linking Drasi + azd extension workflow                |
| Consideration | README prerequisites table, quickstart, error code reference                                | Reviews prerequisites, evaluates templates                       | Analytical; pauses at six prerequisites                                             | Go + kubectl not standard in all CI environments                                                   | CI environment setup section in README                                    |
| Acquisition   | `azd extension install azd-drasi`, `azd drasi --help`                                       | Installs extension, verifies command surface                     | Relieved; risk of immediate runtime error if prerequisites missing                  | Prerequisite checks deferred to runtime                                                            | Prerequisite check at install time or via `azd drasi check`               |
| Onboarding    | `azd drasi init --template cosmos-change-feed`, `azd drasi validate`, `azd drasi provision` | Scaffolds project, writes queries, provisions infrastructure     | Engaged during YAML editing; anxious during 8-12 min provision                      | Silent provisioning window; FederatedIdentityCredential setup not clearly documented               | Named provisioning phases in log output                                   |
| Engagement    | `azd drasi deploy`, `azd drasi status`, `azd drasi logs`, environment overlays              | Iterates on queries, deploys to staging and prod                 | Satisfied with core workflow; frustrated by delete-then-apply and overlay limits    | No in-place component update; overlay can't remove components; `ERR_DEPLOY_IN_PROGRESS` stale lock | `--timeout` override on deploy; component inclusion/exclusion in overlays |
| Retention     | `azd drasi upgrade`, `azd drasi diagnose`, network policy management                        | Upgrades runtime, runs health checks, investigates incidents     | Confident when diagnose resolves incidents; anxious around upgrades without preview | No upgrade preview; network policy docs don't cover adding new datastores                          | `azd drasi upgrade --dry-run`; network policy extension guide             |
| Advocacy      | Internal tech talks, PR template updates, peer referrals                                    | Demos offline workflow, writes platform playbook, onboards peers | Proud and productive                                                                | No templates decision guide; no Codespace badge for zero-friction evaluation                       | Codespace badge in README; templates decision guide in docs               |

---

## Prioritized improvements

### Quick wins

These can be addressed with documentation, configuration, or small code changes.

**1. Stream named provisioning phases**

`azd drasi provision` should emit named phase markers ("Creating AKS cluster...", "Configuring Workload Identity...", "Installing Drasi runtime...", "Running health check...") so engineers know where they are in the 8-12 minute window. This requires no architecture change and directly addresses the most common source of anxiety during first use.

**2. Promote the devcontainer as the first onramp**

Move the devcontainer mention to the top of the "Quick start" section in the README with a clear sentence: "The fastest way to get all prerequisites is to open this repository in the VS Code Dev Container." Currently it sits under a contributor-focused "Build and run" section that most users skip.

**3. Add a templates decision guide**

A short table or paragraph in the docs explaining when to use each template (`cosmos-change-feed` for Cosmos DB CDC, `postgresql-source` for PostgreSQL logical replication, `event-hub-routing` for Event Hub fan-out, `query-subscription` for notification workflows, `blank` for custom configurations) removes a question that currently requires reading multiple quickstart files.

**4. Clarify FederatedIdentityCredential automation**

Add a note to the quickstart that `azd drasi provision` automatically creates the FederatedIdentityCredential for `system:serviceaccount:drasi-system:drasi-api`. Engineers who also read the raw Drasi docs will try to do this manually and create conflicts.

**5. Document the environment overlay limitation**

The limitation ("overlays support parameter overrides, not component removal") should appear explicitly in the configuration reference and in the multi-environment workflow section of the README. Currently a developer discovers it by failing.

**6. Add a GitHub Codespace badge**

A "Open in GitHub Codespaces" badge in the README creates a zero-install evaluation path for engineers who don't want to set up six prerequisites locally before deciding whether the tool fits their needs.

**7. Document which network policies to update for new datastore connections**

A short section in `docs/troubleshooting.md` or a dedicated network policy reference explaining that `drasi-allow-datastores` must be updated when adding new data source endpoints removes a common operational confusion point.

---

### Deeper investments

These require product or engineering decisions beyond documentation.

**8. Prerequisite check at install time or first-run**

An `azd drasi check` command (or an automatic check on first use) that validates all prerequisites (azd version, drasi CLI version and path, Azure CLI version, Docker, kubectl) with clear remediation hints for each failure would eliminate the class of frustrating early failures where the developer installs the extension, runs their first command, and receives `ERR_DRASI_CLI_NOT_FOUND`.

**9. `--timeout` override on `azd drasi deploy`**

Allow `azd drasi deploy --timeout 30m` (or per-component timeout flags) so teams with large pipelines can override the 5-minute per-component and 15-minute total defaults. Without this, `ERR_COMPONENT_TIMEOUT` and `ERR_TOTAL_TIMEOUT` become hard ceilings on pipeline complexity.

**10. Component inclusion/exclusion in environment overlays**

Extend `drasi/environments/<name>.yaml` to support a `components.exclude` list so a staging overlay can suppress production-only reactions without requiring a separate `drasi.yaml`. This is the most commonly requested workaround in multi-environment workflows.

**11. `azd drasi upgrade --dry-run`**

Show the current and target Drasi runtime version and a summary of what the upgrade would change before committing. The `--force` gate already exists; a `--dry-run` mode would give the engineer the information they need to decide whether `--force` is appropriate without reading external changelogs.

**12. Auto-clear stale `ERR_DEPLOY_IN_PROGRESS` locks**

After a configurable timeout (for example, 30 minutes), automatically release an in-progress deploy lock from azd environment state. A CI job that fails and retries should not require manual state surgery to recover.

**13. In-place component update semantics (longer term)**

Explore whether Drasi's API supports patch or update operations for sources and reactions that would allow the extension to apply changes without a delete-then-apply cycle. Even partial support (for example, in-place updates for reaction configuration, delete-then-apply only for sources that require connection reconfiguration) would reduce pipeline downtime during iterative development.

---

_Recommended next step: Use this map as the basis for a Miro or FigJam session with the engineering and developer experience teams to prioritize improvements against the backlog._
