# Compute Environments

Aspire provides compute environment publishers that transform the AppHost resource
graph into deployment-specific artifacts. Each publisher targets a different hosting platform.
Publishers were introduced in Aspire 9.x and are fully supported in Aspire 13.0.

---

## Available Environments

```csharp
builder.AddAzureContainerAppEnvironment("aca");
builder.AddKubernetesEnvironment("k8s");
builder.AddDockerComposeEnvironment("compose");
builder.AddAzureAppServiceEnvironment("appservice");
```

---

## Azure Container Apps (First-Class)

ACA is Aspire's primary deployment target with first-class `azd` support.

### Via azd (Recommended)

```bash
azd init           # Detects Aspire AppHost
azd up             # Provisions resources + deploys containers
azd deploy         # Redeploy after code changes
azd monitor        # View logs and metrics
```

### Via Publisher API

```csharp
var env = builder.AddAzureContainerAppEnvironment("production");

builder.AddProject<Projects.Api>("api")
    .PublishAsAzureContainerApp((infra, app) =>
    {
        app.Template.Scale.MinReplicas = 1;
        app.Template.Scale.MaxReplicas = 10;
    });
```

### Referencing Existing Azure Resources

```csharp
var workspaceName = builder.AddParameter("workspace-name");
var workspaceRg = builder.AddParameter("workspace-rg");

var logWorkspace = builder.AddAzureLogAnalyticsWorkspace("workspace")
    .AsExisting(workspaceName, workspaceRg);

var containerEnv = builder.AddAzureContainerAppEnvironment("production")
    .WithAzureLogAnalyticsWorkspace(logWorkspace);
```

---

## Kubernetes

Generate manifests from the AppHost:

```bash
aspire publish --output ./manifests --format kubernetes
```

Generated artifacts: `deployments.yaml`, `services.yaml`, `configmaps.yaml`, `secrets.yaml`.

### Customizing Generated Manifests

```csharp
builder.AddKubernetesEnvironment("env")
    .WithProperties(env =>
    {
        env.DefaultImagePullPolicy = "Always";
    });

builder.AddProject<Projects.Api>("api")
    .PublishAsKubernetesService(resource =>
    {
        resource.Deployment!.Spec.RevisionHistoryLimit = 5;
    });
```

> **Important:** Generated K8s manifests are a starting point. Apply your Kubernetes
> deployment standards (probes, resource limits, network policies) via Kustomize or Helm
> overlays. See [kubernetes instructions](../../../instructions/kubernetes-deployment-best-practices.instructions.md).

---

## Docker Compose

```csharp
builder.AddDockerComposeEnvironment("env")
    .ConfigureComposeFile(file =>
    {
        file.Name = "my-app";
    });

builder.AddContainer("nginx", "nginx:<pinned-version>")
    .PublishAsDockerComposeService((resource, service) =>
    {
        service.Labels["traefik.enable"] = "true";
        service.Restart = "always";
    });
```

### Dashboard with Docker Compose (9.3+)

```csharp
builder.AddDockerComposeEnvironment("env")
    .WithDashboard();  // Adds Aspire Dashboard to compose output
```

---

## Azure App Service (9.3+)

Deploy .NET projects as containerized Linux Web Apps:

```csharp
builder.AddAzureAppServiceEnvironment("env");

builder.AddProject<Projects.Api>("api")
    .WithExternalHttpEndpoints()
    .PublishAsAzureAppServiceWebsite((infra, site) =>
    {
        site.SiteConfig.IsWebSocketsEnabled = true;
    });
```

Requires projects to expose a single public HTTP endpoint.

---

## Multiple Compute Environments

When you define more than one environment, explicitly assign services:

```csharp
var k8s = builder.AddKubernetesEnvironment("k8s");
var compose = builder.AddDockerComposeEnvironment("compose");

builder.AddProject<Projects.Api>("api")
    .WithComputeEnvironment(k8s);

builder.AddProject<Projects.Worker>("worker")
    .WithComputeEnvironment(compose);
```

Without explicit assignment and multiple environments defined, Aspire throws an
ambiguous environment exception at publish time.

---

## Publishing Commands

| Target         | Command                                             |
| -------------- | --------------------------------------------------- |
| ACA via azd    | `azd up` (auto-detects AppHost)                     |
| ACA via CLI    | `aspire publish --format aca`                       |
| Kubernetes     | `aspire publish --format kubernetes --output ./k8s` |
| Docker Compose | `aspire publish --format docker-compose`            |
| App Service    | Via publisher API + `azd`                           |
