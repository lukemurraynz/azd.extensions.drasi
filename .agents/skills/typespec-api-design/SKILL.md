---
name: typespec-api-design
description: >-
  Design, scaffold, and maintain API specifications using TypeSpec with OpenAPI 3.0 output, Azure data-plane patterns, versioning, LROs, and CI integration. WHEN: "create new API spec", "scaffold TypeSpec project", "generate OpenAPI from TypeSpec", "add API versioning", "define long-running operations", "lint API specs", "migrate OpenAPI to TypeSpec". INVOKES: typespec-azure MCP server for verification. FOR SINGLE OPERATIONS: use tsp CLI directly for compile/lint.
license: MIT
---

# TypeSpec API design

TypeSpec is a Microsoft-built language for describing APIs. You write concise `.tsp` files that define models, operations, and interfaces, then emit OpenAPI 3.0/3.1, JSON Schema, Protobuf, or client/server code from a single source of truth.

This skill covers both generic REST APIs (using `@typespec/http` + `@typespec/openapi3`) and Azure-aligned APIs (adding `@azure-tools/typespec-azure-core` or `@azure-tools/typespec-azure-resource-manager`). Azure patterns are the reference model because they encode Microsoft API Guidelines structurally, but the patterns apply to any REST API.

## When to use

- New API surface (cross-team, public, or agent-consumable)
- API redesign where the existing OpenAPI spec is hand-written and drifting
- Azure data-plane or ARM service that needs to pass API review
- Any project where spec-first development is preferred over code-first annotation

## When NOT to use

- Existing stable API with no planned changes (migration cost exceeds value)
- Internal-only API where code-first annotations (Swashbuckle, NSwag, tsoa) are already producing validated specs
- Prototyping where the API shape is not yet decided (use TypeSpec once the shape stabilizes)

## Hard rules

1. **Spec-first**: Write TypeSpec, emit OpenAPI. The generated spec is the contract. Do not hand-edit emitted files.
2. **One source of truth**: Models, operations, and versioning live in `.tsp` files. Do not maintain parallel hand-written OpenAPI alongside TypeSpec.
3. **Document everything**: Every model, property, operation, and parameter must have a doc comment (`/** ... */`) or `@doc()` decorator. TypeSpec Azure linting enforces this.
4. **Semantic operation IDs**: Use verb-noun naming (`listWidgets`, `createOrder`). TypeSpec generates these from interface + operation names automatically. Do NOT use `@operationId()` for Azure APIs (the Azure Style Guide forbids it; the ID is derived from the interface + operation name). For non-Azure APIs, override only when the default is unclear.
5. **Reusable schemas**: Define models under namespaces and reference by name. Do not inline the same shape in multiple operations.
6. **Security schemes**: Always declare `@useAuth()` at the namespace level. Do not leave specs without auth documentation.
7. **Generated files are build artifacts**: Add `tsp-output/` to `.gitignore` (or equivalent). Check in only the `.tsp` source and `tspconfig.yaml`.
8. **CI validation**: Run `tsp compile .` and `tsp lint .` in CI. Treat compiler errors and lint warnings as build failures.

## Integration with App-as-Skill

After generating OpenAPI:

1. Ensure OpenAPI includes:
   - Descriptions on all operations and parameters
   - Examples for request/response
   - Semantic operationIds

2. Invoke the "app-as-skill" skill to:
   - Generate MCP tools
   - Generate Copilot plugin manifests
   - Validate agent-consumability

TypeSpec defines the API. App-as-Skill makes it usable by agents.

## Project scaffolding

### Generic REST API (non-Azure)

```shell
mkdir my-api && cd my-api
tsp init
# Select "Generic REST API" template
# Select @typespec/http and @typespec/openapi3
tsp install
```

Produces this structure:

```
main.tsp
tspconfig.yaml
package.json
node_modules/
tsp-output/
  @typespec/
    openapi3/
      openapi.yaml
```

### Azure data-plane service

```shell
mkdir my-service && cd my-service
tsp init https://aka.ms/typespec/azure-init
# Select "(standalone) Azure Data Plane Service Project"
tsp install
```

### Azure Resource Manager service

```shell
tsp init https://aka.ms/typespec/azure-init
# Select "(standalone) Azure Resource Manager Service Project"
tsp install
```

### Build and watch

```shell
tsp compile .            # one-shot build
tsp compile . --watch    # rebuild on save
```

## tspconfig.yaml

The configuration file controls emitters, output directories, and linting. Example for a generic API with OpenAPI 3.0 output:

```yaml
emit:
  - "@typespec/openapi3"
options:
  "@typespec/openapi3":
    emitter-output-dir: "{output-dir}/openapi"
    openapi-versions:
      - "3.0.0"
```

For Azure APIs, add the linter ruleset to catch Azure API Guidelines violations at compile time:

```yaml
emit:
  - "@typespec/openapi3"
linter:
  extends:
    - "@azure-tools/typespec-azure-core/all"
options:
  "@typespec/openapi3":
    emitter-output-dir: "{output-dir}/openapi3"
```

### Autorest vs OpenAPI3 emitter (Azure SDK tooling)

If generating Azure SDKs (via AutoRest or `tsp-client`), use `@azure-tools/typespec-autorest` instead of (or alongside) `@typespec/openapi3`. The OpenAPI3 emitter does NOT include `x-ms-*` extensions (pageable, LRO, client name) that Azure SDK generators require.

```yaml
emit:
  - "@azure-tools/typespec-autorest"
  - "@typespec/openapi3"
linter:
  extends:
    - "@azure-tools/typespec-azure-core/all"
options:
  "@azure-tools/typespec-autorest":
    emitter-output-dir: "{output-dir}/swagger"
  "@typespec/openapi3":
    emitter-output-dir: "{output-dir}/openapi3"
```

For non-Azure APIs that don't use Azure SDK tooling, `@typespec/openapi3` alone is sufficient.

## Core patterns

### Service namespace and metadata

```typespec
import "@typespec/http";
import "@typespec/openapi3";

using Http;

@service(#{ title: "Contoso Widget Manager" })
@server("https://api.contoso.com", "Production endpoint")
namespace Contoso.WidgetManager;
```

For Azure data-plane, add Azure Core:

```typespec
import "@typespec/http";
import "@typespec/rest";
import "@typespec/versioning";
import "@azure-tools/typespec-azure-core";

using Http;
using Rest;
using Versioning;
using Azure.Core;
```

### Defining models

```typespec
/** A widget managed by the service. */
model Widget {
  /** The widget's display name. */
  name: string;

  /** Manufacturer identifier. */
  manufacturerId: string;

  /** Current operational status. */
  status: "active" | "inactive" | "maintenance";

  /** Weight in kilograms. */
  @minValue(0)
  weight: float64;

  /** When the widget was created. */
  createdAt: utcDateTime;
}
```

Key decorators for model properties:

| Decorator                         | Purpose                                                                                                | Example                               |
| --------------------------------- | ------------------------------------------------------------------------------------------------------ | ------------------------------------- |
| `@minValue(n)` / `@maxValue(n)`   | Numeric range constraints                                                                              | `@minValue(0) age: int32`             |
| `@minLength(n)` / `@maxLength(n)` | String length constraints                                                                              | `@maxLength(50) name: string`         |
| `@pattern(regex)`                 | String pattern validation                                                                              | `@pattern("^[A-Z]{2}$") code: string` |
| `@format(name)`                   | Semantic format hint (NOT recommended for Azure; use specific scalar types like `url`, `uuid` instead) | `@format("email") email: string`      |
| `@secret`                         | Marks sensitive fields                                                                                 | `@secret apiKey: string`              |
| `@doc(text)`                      | Documentation (alternative to `/** */`)                                                                | `@doc("User email") email: string`    |

### Defining resources (Azure-style, applicable to any REST API)

```typespec
/** A widget resource. */
@resource("widgets")
model Widget {
  /** The widget name, used as the resource key. */
  @key("widgetName")
  @visibility(Lifecycle.Read)
  name: string;

  /** Manufacturer identifier. */
  manufacturerId: string;
}
```

- `@resource("widgets")` sets the collection name and route segment.
- `@key("widgetName")` marks the path parameter.
- `@visibility(Lifecycle.Read)` means the property appears only in responses.
- Path parameters should use `@maxLength` and `@pattern` to constrain allowed values (Azure Style Guide requirement).

```typespec
@key("widgetName")
@maxLength(64)
@pattern("^[a-zA-Z0-9-]+$")
@visibility(Lifecycle.Read)
name: string;
```

### Defining operations (generic REST)

```typespec
@route("/widgets")
interface Widgets {
  /** List all widgets. */
  @get list(@query filter?: string): Widget[];

  /** Get a widget by ID. */
  @get read(@path id: string): Widget;

  /** Create a new widget. */
  @post create(@body widget: Widget): Widget;

  /** Update an existing widget. */
  @put update(@path id: string, @body widget: Widget): Widget;

  /** Delete a widget. */
  @delete remove(@path id: string): void;
}
```

### Defining operations (Azure Core standard lifecycle)

Azure Core provides type-safe operation templates that encode Azure API Guidelines:

```typespec
alias ServiceTraits = SupportsRepeatableRequests &
  SupportsConditionalRequests &
  SupportsClientRequestId;

alias Operations = Azure.Core.ResourceOperations<ServiceTraits>;

interface Widgets {
  /** Fetch a Widget by name. */
  getWidget is Operations.ResourceRead<Widget>;

  /** Create or update a Widget. */
  createOrUpdateWidget is Operations.ResourceCreateOrUpdate<Widget>;

  /** Delete a Widget. */
  deleteWidget is Operations.ResourceDelete<Widget>;

  /** List Widget resources. */
  listWidgets is Operations.ResourceList<Widget>;
}
```

Available standard operations:

| Operation template                         | HTTP                     | Description                      |
| ------------------------------------------ | ------------------------ | -------------------------------- |
| `ResourceRead<T>`                          | GET                      | Read a single resource           |
| `ResourceCreateOrUpdate<T>`                | PATCH (merge)            | Upsert a resource                |
| `ResourceCreateOrReplace<T>`               | PUT                      | Create or replace                |
| `ResourceCreateWithServiceProvidedName<T>` | POST                     | Create with server-generated key |
| `ResourceDelete<T>`                        | DELETE                   | Delete a resource                |
| `ResourceList<T>`                          | GET (collection)         | List resources with pagination   |
| `ResourceAction<T, TReq, TResp>`           | POST (action)            | Custom instance action           |
| `ResourceCollectionAction<T, TReq, TResp>` | POST (collection action) | Custom collection action         |

### Child resources

```typespec
/** A part belonging to a widget. */
@resource("parts")
@parentResource(Widget)
model WidgetPart {
  @key("partName")
  name: string;

  /** The part number. */
  number: string;
}
```

This produces routes like `/widgets/{widgetName}/parts/{partName}`.

### Long-running operations (LROs)

```typespec
interface Widgets {
  /** Get status of a Widget operation. */
  getWidgetOperationStatus is Operations.GetResourceOperationStatus<Widget, never>;

  /** Fetch a Widget by name. */
  getWidget is Operations.ResourceRead<Widget>;

  /** Create or replace a Widget asynchronously. */
  @pollingOperation(Widgets.getWidgetOperationStatus)
  createOrUpdateWidget is Operations.LongRunningResourceCreateOrReplace<Widget>;

  /** Delete a Widget asynchronously. */
  @pollingOperation(Widgets.getWidgetOperationStatus)
  deleteWidget is Operations.LongRunningResourceDelete<Widget>;

  /** List Widget resources. */
  listWidgets is Operations.ResourceList<Widget>;
}
```

The status monitor operation must be defined before the LRO operations that reference it.

Available LRO templates:

| Template                                              | Description              |
| ----------------------------------------------------- | ------------------------ |
| `LongRunningResourceCreateOrReplace<T>`               | Async PUT create/replace |
| `LongRunningResourceCreateOrUpdate<T>`                | Async PATCH upsert       |
| `LongRunningResourceDelete<T>`                        | Async DELETE             |
| `LongRunningResourceAction<T, TReq, TResp>`           | Async custom action      |
| `LongRunningResourceCollectionAction<T, TReq, TResp>` | Async collection action  |

### Custom actions

```typespec
/** Schedule a widget for repairs. */
op scheduleRepairs is Operations.ResourceAction<
  Widget,
  WidgetRepairRequest,
  WidgetRepairResponse
>;
```

Produces route: `/widgets/{widgetName}:scheduleRepairs`

### Error responses

For generic APIs, define an error model:

```typespec
@error
model ErrorResponse {
  @statusCode code: 404 | 500;
  /** The error message. */
  message: string;
}

op getWidget(@path id: string): Widget | ErrorResponse;
```

For Azure APIs, use the built-in error types or customize them:

```typespec
alias Operations = Azure.Core.ResourceOperations<ServiceTraits, ErrorResponse>;
```

### Singleton resources

A singleton resource has exactly one instance (e.g., analytics for a widget):

```typespec
@resource("analytics")
@parentResource(Widget)
model WidgetAnalytics {
  @key("analyticsId")
  id: "current";

  /** The number of uses of the widget. */
  useCount: int64;

  /** The number of times the widget was repaired. */
  repairCount: int64;
}

op getAnalytics is Operations.ResourceRead<WidgetAnalytics>;
```

Route: `/widgets/{widgetName}/analytics/current` (key parameter excluded since it's a literal).

### Non-resource RPC operations

For operations that don't follow REST resource patterns, use `RpcOperation`:

```typespec
/** Analyze text for sentiment. */
op analyzeText is Operations.RpcOperation<
  { @body text: string },
  { sentiment: float64; confidence: float64 }
>;
```

### Security and authentication

```typespec
@useAuth(BearerAuth)
namespace MyService;
```

For OAuth2 / Entra ID:

```typespec
/** Microsoft Entra ID OAuth2 flow. */
model EntraIDToken
  is OAuth2Auth<[
    {
      type: OAuth2FlowType.authorizationCode;
      authorizationUrl: "https://login.microsoftonline.com/common/oauth2/v2.0/authorize";
      tokenUrl: "https://login.microsoftonline.com/common/oauth2/v2.0/token";
      scopes: ["https://api.contoso.com/.default"];
    }
  ]>;

@useAuth(EntraIDToken)
namespace Contoso.WidgetManager;
```

For API key authentication (e.g., Cognitive Services):

```typespec
/** API key passed via header. */
model AzureKey is ApiKeyAuth<ApiKeyLocation.header, "Ocp-Apim-Subscription-Key">;

@useAuth(AzureKey)
namespace Contoso.WidgetManager;
```

## Versioning

Add versioning support from the first version. Define a `Versions` enum and apply `@versioned`:

```typespec
@service(#{ title: "Contoso Widget Manager" })
@versioned(Contoso.WidgetManager.Versions)
namespace Contoso.WidgetManager;

/** Contoso Widget Manager API versions. */
enum Versions {
  /** Initial GA release. */
  v2024_01_15: "2024-01-15",
}
```

When you release a new version, add to the enum and annotate changes with `@added` / `@removed`:

```typespec
enum Versions {
  v2024_01_15: "2024-01-15",
  v2024_06_01: "2024-06-01",
}

model Widget {
  @key("widgetName")
  @visibility(Lifecycle.Read)
  name: string;

  /** Added in v2024_06_01. */
  @added(Versions.v2024_06_01)
  color?: string;

  /** Removed in v2024_06_01. */
  @removed(Versions.v2024_06_01)
  manufacturerId: string;
}
```

For Azure APIs with `@useDependency`, pin each service version to a specific Azure.Core version.

## Model composition

TypeSpec supports several composition patterns that map to OpenAPI constructs:

| Pattern                     | TypeSpec                           | OpenAPI result                               |
| --------------------------- | ---------------------------------- | -------------------------------------------- |
| Copy properties             | `...Address` (spread)              | Flat schema with inlined properties          |
| Inheritance                 | `model Dog extends Pet`            | `allOf` with `$ref` to parent                |
| Same shape, no relationship | `model Shipping is Address`        | Independent schema with same properties      |
| Discriminated unions        | `@discriminator("kind") model Pet` | `discriminator` object with mappings         |
| Union types                 | `Dog \| Cat \| Hamster`            | `anyOf` (default) or `oneOf` (with `@oneOf`) |

## Customizing operations with traits

Apply cross-cutting concerns to all operations via service traits:

```typespec
alias ServiceTraits = SupportsRepeatableRequests &
  SupportsConditionalRequests &
  SupportsClientRequestId &
  RequestHeadersTrait<{
    /** Custom correlation header. */
    @header("X-Correlation-Id") correlationId?: string;
  }>;

alias Operations = ResourceOperations<ServiceTraits>;
```

### Pagination query parameters

Azure Core provides standard query parameter types for list operations. Use them instead of defining custom parameters:

```typespec
op listWidgets is Operations.ResourceList<
  Widget,
  ListQueryParametersTrait<
    StandardListQueryParameters &
    SelectQueryParameter &
    ExpandQueryParameter
  >
>;
```

Available pagination/filtering parameters:

| Parameter type              | Query param    | Description                         |
| --------------------------- | -------------- | ----------------------------------- |
| `TopQueryParameter`         | `$top`         | Maximum number of items to return   |
| `SkipQueryParameter`        | `$skip`        | Number of items to skip             |
| `MaxPageSizeQueryParameter` | `$maxpagesize` | Maximum page size                   |
| `FilterQueryParameter`      | `$filter`      | OData-style filter expression       |
| `OrderByQueryParameter`     | `$orderby`     | Sort order                          |
| `SelectQueryParameter`      | `$select`      | Fields to include in response       |
| `ExpandQueryParameter`      | `$expand`      | Related resources to include inline |

`StandardListQueryParameters` combines `TopQueryParameter`, `SkipQueryParameter`, `MaxPageSizeQueryParameter`, `FilterQueryParameter`, and `OrderByQueryParameter`.

## CI/CD integration

### GitHub Actions

```yaml
- name: Install TypeSpec
  run: npm ci

- name: Compile TypeSpec
  run: npx tsp compile .

- name: Lint TypeSpec
  run: npx tsp lint .

- name: Validate generated OpenAPI
  run: npx spectral lint tsp-output/@typespec/openapi3/openapi.yaml
```

### Azure DevOps

```yaml
- task: NodeTool@0
  inputs:
    versionSpec: "22.x"

- script: npm ci
  displayName: Install TypeSpec

- script: npx tsp compile .
  displayName: Compile TypeSpec

- script: npx tsp lint .
  displayName: Lint TypeSpec

- task: PublishPipelineArtifact@1
  displayName: Publish TypeSpec Output
  inputs:
    targetPath: "tsp-output"
    artifact: "tsp-output"
    publishLocation: "pipeline"
```

### Validation checklist for CI

- `tsp compile .` exits 0 (no compiler errors)
- `tsp lint .` exits 0 (no lint warnings when using Azure rules)
- Generated OpenAPI passes Spectral or equivalent linter
- Generated spec matches expected version count (one per `Versions` enum member)
- Generated spec has no missing descriptions (enforced by TypeSpec Azure linting)

## Migration from existing OpenAPI

TypeSpec provides a CLI tool to convert existing OpenAPI 3.x to TypeSpec:

```shell
npx @typespec/openapi3 convert --output-dir ./typespec ./existing-openapi.yaml
```

After conversion:

1. Review generated `.tsp` files for correctness
2. Replace inline schemas with named models
3. Add doc comments to all models and operations
4. Set up `tspconfig.yaml` with desired emitters
5. Compare re-emitted OpenAPI against original to verify fidelity

## Known pitfalls

| Area                         | Pitfall                                                                                        | Mitigation                                                                                           |
| ---------------------------- | ---------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| Hand-editing output          | Editing `tsp-output/` files directly; changes are overwritten on next compile                  | Never edit generated files. Change `.tsp` source instead                                             |
| Missing docs                 | Forgetting doc comments on properties; Azure linting catches this but generic projects may not | Enable `@typespec/best-practices` linter or add `@doc()` checks                                      |
| LRO ordering                 | Status monitor op must appear before LRO ops in the interface                                  | Define `getOperationStatus` first in the interface block                                             |
| Version enum naming          | Using non-date version strings for Azure APIs                                                  | Follow `YYYY-MM-DD` format for Azure data plane; use semantic versions for non-Azure                 |
| Emitter version mismatch     | Mixing incompatible versions of `@typespec/compiler` and emitter packages                      | Use `tsp install` for dependency resolution; pin versions in `package.json`                          |
| Tooling support gap          | OpenAPI 3.1 vs 3.0 tooling differences                                                         | Target OpenAPI 3.0.x unless full toolchain has verified 3.1 support                                  |
| Wrong emitter for Azure SDKs | Using `@typespec/openapi3` when Azure SDK generation is needed; loses `x-ms-*` extensions      | Use `@azure-tools/typespec-autorest` for Azure SDK tooling; `openapi3` for non-Azure consumers       |
| `@operationId` in Azure      | Explicitly setting `@operationId` on Azure operations; conflicts with auto-derived IDs         | Remove `@operationId`; let TypeSpec derive from interface + operation name (Azure Style Guide)       |
| `@format` in Azure           | Using `@format("uuid")` instead of the `uuid` scalar type; `@format` is open-ended             | Use specific scalar types (`uuid`, `url`, `eTag`, `ipV4Address`, `ipV6Address`) instead of `@format` |
| Missing path constraints     | Path parameters without `@maxLength` and `@pattern`; clients can send invalid IDs              | Add `@maxLength` and `@pattern` to all `@key` properties                                             |

## Related skills

- [REST API Design](../../instructions/rest-api-design.instructions.md) for Microsoft API Guidelines alignment
- [App as Skill](../app-as-skill/SKILL.md) for making APIs agent-consumable
- [API Versioning Governance](../api-versioning-governance/SKILL.md) for version lifecycle policy
- [APIM Policy Authoring](../apim-policy-authoring/SKILL.md) for importing TypeSpec-generated specs into APIM

## Azure Style Guide summary

Key rules from the [Azure TypeSpec Style Guide](https://azure.github.io/typespec-azure/docs/reference/azure-style-guide/) that apply broadly:

| Rule                                                    | Rationale                                                        |
| ------------------------------------------------------- | ---------------------------------------------------------------- |
| Import and use `@azure-tools/typespec-azure-core`       | Enforces Azure API Guidelines structurally                       |
| Import and use `@typespec/versioning`                   | All Azure specs must be versioned from v1                        |
| Do NOT use `@operationId`                               | Operation ID is auto-derived from interface + operation name     |
| Do NOT use `@format`                                    | Use specific scalar types (`uuid`, `url`, `eTag`, etc.) instead  |
| Add `@maxLength` + `@pattern` to path params            | Constrains resource identifiers to valid characters              |
| Group operations in interfaces named as plural nouns    | e.g., `interface Widgets {}`, `interface VirtualMachines {}`     |
| Security scheme required                                | Every spec must define `@useAuth` with `oauth2` or `apiKey` type |
| OAuth2 scopes must use `<resource-URI>/.default` format | Standard Azure scope pattern                                     |

### Azure Core linter rules

The `@azure-tools/typespec-azure-core/all` ruleset enforces 30+ rules including:

- `documentation-required` — doc comments on all models, properties, operations
- `auth-required` — security scheme must be defined
- `operation-missing-api-version` — operations need an api-version parameter
- `non-breaking-versioning` — only backward-compatible changes between versions
- `use-standard-operations` — operations should use Azure.Core templates
- `use-standard-names` — operation names follow naming conventions
- `no-generic-numeric` — use `int32`, `int64`, `float64` not `integer` or `numeric`
- `casing-style` — enforce proper casing conventions
- `key-visibility-required` — key properties need lifecycle visibility

Enable in `tspconfig.yaml`:

```yaml
linter:
  extends:
    - "@azure-tools/typespec-azure-core/all"
```

## References

- [TypeSpec language](https://typespec.io/docs)
- [TypeSpec Azure libraries](https://azure.github.io/typespec-azure/docs/intro/)
- [Azure TypeSpec Style Guide](https://azure.github.io/typespec-azure/docs/reference/azure-style-guide/)
- [Azure Core linter rules](https://azure.github.io/typespec-azure/docs/libraries/azure-core/reference/linter/)
- [TypeSpec playground](https://typespec.io/playground/)
- [Azure TypeSpec playground](https://azure.github.io/typespec-azure/playground)
- [OpenAPI 3 emitter reference](https://typespec.io/docs/emitters/openapi3/openapi/)
- [Autorest emitter (Azure SDK)](https://github.com/Azure/typespec-azure/tree/main/packages/typespec-autorest)
- [Microsoft API Guidelines (vNext)](https://github.com/microsoft/api-guidelines/blob/vNext/azure/Guidelines.md)
- [Build pipeline guide](https://azure.github.io/typespec-azure/docs/howtos/rest-api-publish/buildpipelines)
- [TypeSpec GitHub repository](https://github.com/microsoft/typespec)
- [TypeSpec Azure GitHub repository](https://github.com/Azure/typespec-azure)

## Currency

- **Date checked:** 2026-04-02
- **Sources:** [TypeSpec docs](https://typespec.io/docs), [TypeSpec Azure docs](https://azure.github.io/typespec-azure/docs/intro/)
- **Authoritative references:** [TypeSpec 1.10+](https://typespec.io/release-notes/typespec-1-10-0/), [Azure TypeSpec libraries](https://azure.github.io/typespec-azure/)
- **MCP verification:** Use the `typespec-azure` MCP server to verify operation templates, decorator names, and linter rules against the current `Azure/typespec-azure` repo before generating TypeSpec code. Do not rely on cached knowledge for version-specific behavior.

### Verification steps

1. Confirm `@typespec/compiler` latest stable version and any new built-in decorators
2. Verify Azure Core operation templates for any additions or deprecations
3. Check OpenAPI 3 emitter for 3.1 support status and configuration changes
4. Verify `tsp init` template names have not changed
