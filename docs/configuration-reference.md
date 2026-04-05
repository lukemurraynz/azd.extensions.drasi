# Configuration reference

This document describes every YAML entity type the extension reads and the full set of supported fields.

## File layout

```
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
|-------|------|----------|-------------|
| `apiVersion` | string | yes | Schema version. Must be `"v1"`. |
| `includes` | list | no | Override the default glob patterns. If omitted, all YAML files in each entity subdirectory are loaded. |
| `includes[].kind` | string | yes | One of `source`, `continuousquery`, `reaction`, `middleware`. |
| `includes[].pattern` | string | yes | Glob pattern relative to `drasi/`. |
| `environments` | map | no | Maps environment name to overlay file path. |
| `featureFlags` | map | no | Boolean feature flags consumed by the extension. |

## Source

```yaml
apiVersion: v1          # required
kind: Source            # required; case-insensitive
id: cosmos-source       # required; alphanumeric, hyphens, underscores
sourceKind: CosmosDb    # required; identifies the Drasi source plugin

properties:             # required; key-value map of source-specific config
  accountEndpoint:
    value: "https://my-cosmos.documents.azure.com:443/"
  database:
    value: "orders"
  container:
    value: "events"
  connectionString:
    secretRef:
      vaultName: my-keyvault
      secretName: cosmos-connection-string
```

Fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | yes | Must be `"v1"`. |
| `kind` | string | yes | Must be `"Source"` (case-insensitive). |
| `id` | string | yes | Unique identifier within the project. |
| `sourceKind` | string | yes | Drasi source plugin name (e.g. `CosmosDb`, `PostgreSql`). |
| `properties` | map | yes | Plugin-specific configuration values. |

## ContinuousQuery

```yaml
apiVersion: v1
kind: ContinuousQuery
id: order-changes
queryLanguage: Cypher    # required

sources:                 # required; at least one source reference
  - id: cosmos-source

query: |                 # required; the query body in the declared queryLanguage
  MATCH (o:Order)
  WHERE o.status = 'pending'
  RETURN o.id, o.total

joins:                   # optional
  - type: inner
    keys:
      - label: Order
        field: customerId
      - label: Customer
        field: id

reactions:               # optional; reaction IDs that subscribe to this query
  - pubsub-reaction

autoStart: true          # optional; default false
```

Fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | yes | Must be `"v1"`. |
| `kind` | string | yes | Must be `"ContinuousQuery"` (case-insensitive). |
| `id` | string | yes | Unique identifier. |
| `queryLanguage` | string | yes | Query language. Must be `"Cypher"`. |
| `sources` | list | yes | Source IDs this query reads from. |
| `query` | string | yes | Query body. |
| `joins` | list | no | Join specifications for multi-source queries. |
| `reactions` | list | no | Reaction IDs to notify when results change. |
| `autoStart` | bool | no | Start the query automatically on deploy. Default `false`. |

## Reaction

```yaml
apiVersion: v1
kind: Reaction
id: pubsub-reaction
reactionKind: Dapr       # required; identifies the Drasi reaction plugin

config:                  # optional; key-value map of reaction-specific config
  pubsubName:
    value: "drasi-pubsub"
  topic:
    value: "order-events"
  signingKey:
    secretRef:
      vaultName: my-keyvault
      secretName: reaction-signing-key
```

Fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | yes | Must be `"v1"`. |
| `kind` | string | yes | Must be `"Reaction"` (case-insensitive). |
| `id` | string | yes | Unique identifier. |
| `reactionKind` | string | yes | Drasi reaction plugin name (e.g. `Dapr`, `SignalR`, `StorageQueue`). |
| `config` | map | no | Plugin-specific configuration values. |

## Middleware

```yaml
apiVersion: v1
kind: Middleware
id: my-middleware
middlewareKind: Mask     # required; identifies the Drasi middleware plugin

config:                  # optional; key-value map
  fields:
    value: "creditCardNumber,ssn"
```

Fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | yes | Must be `"v1"`. |
| `kind` | string | yes | Must be `"Middleware"` (case-insensitive). |
| `id` | string | yes | Unique identifier. |
| `middlewareKind` | string | yes | Drasi middleware plugin name. |
| `config` | map | no | Plugin-specific configuration values. |

## Value types

All property and config values use the `Value` type. Three forms are supported:

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

At deploy time the extension fetches the secret from Key Vault using the provisioned managed identity and creates a Kubernetes Secret. The component YAML is rewritten to reference that Kubernetes Secret before being applied. No secret value ever touches the file system.

**Environment variable reference:**

```yaml
connectionString:
  envRef:
    name: COSMOS_CONNECTION_STRING  # Environment variable name on the runner
```

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

## Feature flags

Feature flags are defined in `drasi.yaml` and read by the extension at runtime. No flag affects Drasi runtime behavior directly; they gate extension-level behavior only.

| Flag | Default | Effect |
|------|---------|--------|
| `enableAuditLog` | `false` | Emit structured audit events to Log Analytics for each deploy action. |
| `experimentalDeploy` | `false` | Reserved for future use. Currently a no-op. |
