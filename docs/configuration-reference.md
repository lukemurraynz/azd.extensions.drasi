# Configuration reference

This document describes every YAML entity type the extension reads and the full set of supported fields.

## File layout

```text
<project-root>/
└── drasi/
    ├── drasi.yaml               # Manifest (required)
    ├── sources/
    │   └── *.yaml               # Source entity files
    ├── queries/
    │   └── *.yaml               # ContinuousQuery entity files
    ├── reactions/
    │   └── *.yaml               # Reaction entity files
    ├── middleware/
    │   └── *.yaml               # Middleware entity files
    └── environments/
        └── <name>.yaml          # Environment overlay files (optional)
```

The manifest controls which files are included. By default all YAML files in the four entity directories are loaded.

## Choosing a template

Use this guide to pick the right starter template for `azd drasi init --template <name>`.

| Template | Use when | Data source | What you get |
| -------- | -------- | ----------- | ------------ |
| `postgresql-source` | You need to react to changes in PostgreSQL via logical replication | PostgreSQL (logical replication) | PostgreSQL source, sample query, debug reaction, AKS + Key Vault + PostgreSQL infra |
| `event-hub-routing` | You need to route events from Azure Event Hubs through Drasi queries | Azure Event Hubs | Event Hub source, routing query, reaction, AKS + Key Vault + Event Hub infra |
| `blank` | You want an empty Drasi project structure with Bicep infra | N/A | Empty drasi/ directory structure, blank Bicep modules |
| `blank-terraform` | You want an empty Drasi project structure with Terraform infra | N/A | Empty drasi/ directory structure, blank Terraform modules |

If you are evaluating Drasi for the first time, start with `postgresql-source`. It includes working infrastructure and sample components that you can deploy end-to-end in a single session. The `cosmos-change-feed` template is experimental — see its README for known limitations.

## drasi.yaml (manifest)

```yaml
apiVersion: v1          # required; must be "v1"

includes:               # optional; glob patterns relative to drasi/
  - kind: source
    pattern: "sources/*.yaml"
  - kind: continuousquery
    pattern: "queries/*.yaml"
  - kind: reaction
    pattern: "reactions/*.yaml"
  - kind: middleware
    pattern: "middleware/*.yaml"

environments:           # optional; map of environment name to overlay file path
  prod: environments/prod.yaml
  staging: environments/staging.yaml

featureFlags:           # optional; map of flag name to bool
  enableAuditLog: true
  experimentalDeploy: false
```

Fields:

| Field | Type | Required | Description |
| ----- | ---- | -------- | ----------- |
| `apiVersion` | string | yes | Schema version. Must be `"v1"`. |
| `includes` | list | no | Override the default glob patterns. If omitted, all YAML files in each entity subdirectory are loaded. |
| `includes[].kind` | string | yes | One of `source`/`sources`, `continuousquery`/`queries`, `reaction`/`reactions`, `middleware`/`middlewares`. Singular and plural forms are both accepted. |
| `includes[].pattern` | string | yes | Glob pattern relative to `drasi/`. |
| `environments` | map | no | Maps environment name to overlay file path. |
| `featureFlags` | map | no | Boolean feature flags consumed by the extension. |
| `secretMappings` | list | no | Azure Key Vault to Kubernetes Secret mappings applied during `azd drasi deploy`. See [Secret mappings](#secret-mappings). |

## Source

```yaml
apiVersion: v1          # required
kind: Source            # required; case-insensitive
name: pg-source         # required; alphanumeric, hyphens, underscores
spec:
  kind: PostgreSQL      # required; identifies the Drasi source plugin
  properties:           # required; key-value map of source-specific config
    connectionString:
      secretRef:
        vaultName: my-keyvault
        secretName: pg-connection-string
    database: drasidb
    tables: public.orders
```

Fields:

| Field | Type | Required | Description |
| ----- | ---- | -------- | ----------- |
| `apiVersion` | string | yes | Must be `"v1"`. |
| `kind` | string | yes | Must be `"Source"` (case-insensitive). |
| `name` | string | yes | Unique identifier within the project. |
| `spec.kind` | string | yes | Drasi source plugin name (e.g. `PostgreSQL`, `CosmosDb`). |
| `spec.properties` | map | yes | Plugin-specific configuration values. Each value uses the [Value type](#value-types). |

## ContinuousQuery

```yaml
apiVersion: v1
kind: ContinuousQuery
name: watch-orders      # required; alphanumeric, hyphens, underscores
spec:
  mode: query           # required; "query" or "view"
  sources:              # required; at least one source subscription
    subscriptions:
      - id: pg-source
  query: >              # required; Cypher query body
    MATCH (o:public_orders)
    RETURN o.id AS orderId, o.status AS status
```

Fields:

| Field | Type | Required | Description |
| ----- | ---- | -------- | ----------- |
| `apiVersion` | string | yes | Must be `"v1"`. |
| `kind` | string | yes | Must be `"ContinuousQuery"` (case-insensitive). |
| `name` | string | yes | Unique identifier within the project. |
| `spec.mode` | string | yes | Query execution mode. Use `"query"` for most cases. |
| `spec.sources.subscriptions` | list | yes | Source IDs this query reads from. Each entry must have an `id` field. |
| `spec.query` | string | yes | Cypher query body. |

## Reaction

```yaml
apiVersion: v1
kind: Reaction
name: log-changes       # required; alphanumeric, hyphens, underscores
spec:
  kind: Debug           # required; identifies the Drasi reaction plugin
  queries:              # required; map of query name to optional filter expression
    watch-orders: ""
```

Fields:

| Field | Type | Required | Description |
| ----- | ---- | -------- | ----------- |
| `apiVersion` | string | yes | Must be `"v1"`. |
| `kind` | string | yes | Must be `"Reaction"` (case-insensitive). |
| `name` | string | yes | Unique identifier within the project. |
| `spec.kind` | string | yes | Drasi reaction plugin name (e.g. `Debug`, `Dapr`, `SignalR`, `StorageQueue`). |
| `spec.queries` | map | yes | Map of ContinuousQuery names to optional filter expressions. Use an empty string for no filter. |

## Middleware

```yaml
apiVersion: v1
kind: Middleware
name: my-middleware      # required; alphanumeric, hyphens, underscores
spec:
  kind: Mask             # required; identifies the Drasi middleware plugin
  config:                # optional; key-value map of plugin-specific config
    fields:
      value: "creditCardNumber,ssn"
```

Fields:

| Field | Type | Required | Description |
| ----- | ---- | -------- | ----------- |
| `apiVersion` | string | yes | Must be `"v1"`. |
| `kind` | string | yes | Must be `"Middleware"` (case-insensitive). |
| `name` | string | yes | Unique identifier within the project. |
| `spec.kind` | string | yes | Drasi middleware plugin name. |
| `spec.config` | map | no | Plugin-specific configuration values. Each value uses the [Value type](#value-types). |

## Value types

All property and config values use the `Value` type. Four forms are supported:

**Plain string:**

```yaml
connectionString:
  value: "Server=myserver;Database=mydb;"
```

**Key Vault secret reference:**

```yaml
connectionString:
  secretRef:
    vaultName: my-keyvault       # Key Vault name (not URI)
    secretName: cosmos-conn-str  # Secret name within the vault
```

At deploy time the extension fetches the secret from Key Vault using the **caller's Azure CLI identity** (the account returned by `az account show`) and creates a Kubernetes Secret. The caller must have the **Key Vault Secrets User** role (or equivalent) on the Key Vault. The component YAML is rewritten to reference that Kubernetes Secret before being applied. No secret value ever touches the file system.

**Environment variable reference:**

```yaml
connectionString:
  envRef:
    name: COSMOS_CONNECTION_STRING  # Environment variable name on the runner
```

**Kubernetes Secret reference:**

```yaml
connectionString:
  kind: Secret
  name: cosmos-secrets
  key: account-endpoint
```

Kubernetes Secret references point directly to a key inside a Kubernetes Secret that already exists in the cluster. Use `secretMappings` in `drasi.yaml` to bridge Azure Key Vault secrets to Kubernetes Secrets at deploy time (see the `cosmos-change-feed` template for an example).

## Secret mappings

The `secretMappings` field in `drasi.yaml` bridges Azure Key Vault secrets to Kubernetes Secrets during `azd drasi deploy`. Each mapping fetches a secret from Key Vault and upserts it as a key inside a Kubernetes Secret in the `drasi-system` namespace.

```yaml
secretMappings:
  - vaultName: "$(AZURE_KEY_VAULT_NAME)"
    secretName: cosmos-connection-string
    k8sSecret: cosmos-secrets
    k8sKey: account-endpoint
```

Fields:

| Field | Type | Required | Description |
| ----- | ---- | -------- | ----------- |
| `vaultName` | string | yes | Azure Key Vault name. Supports `$(ENV_VAR)` expansion from azd environment. |
| `secretName` | string | yes | Secret name in Key Vault. |
| `k8sSecret` | string | yes | Target Kubernetes Secret name. |
| `k8sKey` | string | yes | Key within the Kubernetes Secret. |
| `namespace` | string | no | Kubernetes namespace. Defaults to `drasi-system`. |

The extension fetches each secret using `az keyvault secret show` under the caller's Azure CLI session. The caller must have the `Key Vault Secrets User` role on the Key Vault.

## Environment overlays

Overlays let you change parameter values per environment without duplicating entity files. Place overlay files in `drasi/environments/` and reference them in `drasi.yaml`.

```yaml
# drasi/environments/prod.yaml
apiVersion: v1

parameters:
  database:
    value: "orders-prod"
  connectionString:
    secretRef:
      vaultName: prod-keyvault
      secretName: cosmos-prod-conn
```

Apply an overlay at deploy time:

```bash
azd drasi deploy --environment prod
```

The resolver merges the overlay parameters into matching entity property values. Overlay keys must match the `id` of existing properties.

### Overlay capabilities and limits

Overlays support two operations. They can override parameter values, and they can exclude components from the deployment. They cannot add new components. If you need a component that exists only in one environment, maintain a separate entity file for that environment.

To exclude specific components from a deployment in a given environment, add a `components` section with an `exclude` list:

```yaml
# drasi/environments/staging.yaml
apiVersion: v1

parameters:
  database:
    value: "orders-staging"

components:
  exclude:
    - kind: reaction
      id: production-alerting
```

Components listed in `exclude` are removed from the deployment plan before any apply or delete actions run.

## Feature flags

Feature flags are defined in `drasi.yaml` and read by the extension at runtime. No flag affects Drasi runtime behavior directly; they gate extension-level behavior only.

| Flag                 | Default | Effect                                                                        |
| -------------------- | ------- | ----------------------------------------------------------------------------- |
| `enableAuditLog`     | `false` | Emit structured audit events to Log Analytics for each deploy action.         |
| `experimentalDeploy` | `false` | Reserved for future use. Currently a no-op.                                   |

## Authentication and identity

The extension uses different identities depending on the operation:

| Operation | Identity used | How it authenticates |
| --------- | ------------- | -------------------- |
| `provision` (Bicep deployment) | Caller's Azure CLI identity | `az account show` — user or service principal |
| `deploy` (Key Vault secret fetch) | Caller's Azure CLI identity | `az keyvault secret show` runs under the logged-in `az` session |
| `deploy` (Kubernetes apply) | Caller's kubeconfig context | `kubectl apply` uses the active kubeconfig context |
| `status`, `logs`, `diagnose` | Caller's kubeconfig context | `kubectl` / `drasi` CLI commands |

During `provision`, the extension resolves the caller's Entra ID object ID and passes it as `principalId` to Bicep. This grants the caller AKS RBAC Cluster Admin so `kubectl` and `drasi` commands work after provisioning. Both user accounts and service principals are supported.

During `deploy`, the `secretMappings` in `drasi.yaml` are processed by running `az keyvault secret show` for each mapping. This means the **caller** (not the cluster workload identity) must have read access to the Key Vault secrets. In CI pipelines, ensure the service principal used by `azd auth login` has the `Key Vault Secrets User` role on the target Key Vault.
