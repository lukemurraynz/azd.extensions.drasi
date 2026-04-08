# Architecture overview

This document describes how the `azd-drasi` extension is structured and how its main flows work.

## Diagram artifact

The canonical architecture diagram is maintained as a draw.io file:

- [docs/diagrams/azd-drasi-solution.drawio](diagrams/azd-drasi-solution.drawio)

It currently contains two sheets:

- **System Overview**
- **Deploy Flow**

## Component diagram (text view)

```text
┌──────────────────────────────────────────────────────────────────────────┐
│ Developer workstation                                                    │
│                                                                          │
│  ┌─────────────────────────┐ gRPC (AZD_SERVER) ┌──────────────────┐    │
│  │ azd-drasi extension     │───────────────────►│ azd host         │    │
│  │ (gRPC)                  │◄───────────────────│                  │    │
│  └──────────┬──────────────┘                    └──────────────────┘    │
│             │                                                           │
│         │ os.Exec (subprocess)                                           │
│         ▼                                                                │
│  ┌─────────────┐                                                        │
│  │ Drasi CLI   │ ────────── kubectl / HTTP ──────────────────────────┐ │
│  └─────────────┘                                                      │ │
└───────────────────────────────────────────────────────────────────────┼──┘
                                                                        │
                                         Azure                           │
                                                                        ▼
                              ┌──────────────────────────────────────────┐
                              │ AKS Runtime                              │
                              │  ┌────────────────────────────────────┐  │
                              │  │ drasi-system namespace              │  │
                              │  │ drasi-api + Dapr sidecars           │  │
                              │  └────────────────────────────────────┘  │
                              └──────────────────────────────────────────┘

                               Key Vault | Log Analytics | Managed Identity
```

Key boundaries:

- The extension binary (`azd-drasi`) communicates with the azd host over a local gRPC channel using the address in `AZD_SERVER`. All environment reads and writes go through this channel, never through direct file I/O.
- The extension never calls Azure APIs directly. Provisioning delegates to `azd` lifecycle hooks and Bicep. Key Vault data-plane access is handled at deploy time by translating secret references into Kubernetes Secrets before handing off to the Drasi CLI.
- The Drasi CLI (`drasi`) is a subprocess. The extension invokes it with `os.Exec`, captures stdout/stderr, and surfaces errors using structured error codes.

## Deploy flow (summary)

The **Deploy Flow** sheet uses three top-level phases:

1. Provision (Bicep/Terraform)
2. Deploy (azd-drasi + CLI)
3. Configure Sources/Queries

## Provision flow (detailed)

```text
azd drasi provision
       │
       ├─ 1. Read environment via gRPC (env name, subscription, resource group)
       │
       ├─ 2. bicep build + az deployment sub create
       │      Creates: AKS, Key Vault, UAMI, Log Analytics, optional ACR/Event Hub
       │
       ├─ 3. drasi init (subprocess)
       │      Installs Drasi runtime on AKS
       │
       └─ 4. Configure FederatedIdentityCredential
              system:serviceaccount:drasi-system:drasi-api → UAMI
```

The provision command stores AKS context name and Key Vault URI in azd environment state via gRPC so that subsequent deploy and status commands can read them without requiring re-authentication.

## Deploy flow with Key Vault translation

```text
azd drasi deploy
       │
       ├─ 1. Pre-deploy validation (same checks as azd drasi validate)
       │
       ├─ 2. Resolve Key Vault secret references
       │      For each component property with secretRef:
       │        GET https://<vaultName>.vault.azure.net/secrets/<secretName>
       │        → write to Kubernetes Secret in drasi-system namespace
       │        → replace secretRef in component spec with k8s secret reference
       │
       ├─ 3. Deploy in dependency order
       │      Sources → ContinuousQueries → Middleware → Reactions
       │      Each component: drasi apply (subprocess) → poll for Online state
       │      Per-component timeout: 5 minutes
       │      Total deploy timeout: 15 minutes
       │
       ├─ 4. Record content hashes in azd environment state
       │      Key format: DRASI_HASH_<KIND>_<ID>
       │      Used for idempotency on subsequent deploys
       │
       └─ 5. Emit structured status output (table or JSON)
```

Teardown follows the reverse deployment order: Reactions → Middleware → ContinuousQueries → Sources.

## Code layout

```text
azd.extensions.drasi/
├── main.go                   # Wire cobra root, execute
├── cmd/                      # One file per command; thin RunE handlers
│   ├── root.go               # Persistent --output and --environment flags
│   ├── validate.go           # commandError type, writeCommandError
│   ├── provision.go          # runProvisionFunc injection var, getEnvValue
│   └── ...
├── internal/
│   ├── config/               # YAML model, loader, resolver, schema validation
│   ├── deployment/           # Deploy engine, hash tracking, teardown ordering
│   ├── drasi/                # drasi CLI subprocess wrapper
│   ├── keyvault/             # Secret reference translation
│   ├── observability/        # OpenTelemetry Azure Monitor exporter wiring
│   ├── output/               # ERR_* constants, ExitCodes map, error formatter
│   ├── scaffold/             # Template rendering (go:embed all:templates)
│   └── validation/           # Per-entity validation rules
├── infra/                    # Bicep modules (AKS, Key Vault, UAMI, etc.)
└── drasi/                    # Default drasi.yaml and entity YAML for templates
```

The `cmd/` layer is intentionally thin. Business logic lives in `internal/`. Command handlers read flags, call internal packages, and use `writeCommandError` to emit structured errors with exit codes.
