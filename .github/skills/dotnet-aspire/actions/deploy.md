# Deploy with Aspire

Publish your Aspire application to a deployment target using compute environment publishers.

---

## Choose a Deployment Target

| Target               | Best For                                  | Maturity       |
| -------------------- | ----------------------------------------- | -------------- |
| Azure Container Apps | Default choice; first-class `azd` support | Production     |
| Kubernetes           | Existing K8s clusters with custom tooling | Production     |
| Docker Compose       | Self-hosted or on-prem environments       | Production     |
| Azure App Service    | Simple web apps on PaaS                   | Preview (9.3+) |

---

## Azure Container Apps (Recommended)

### Via Azure Developer CLI

```bash
# Initialize — azd detects Aspire AppHost automatically
azd init

# Provision infrastructure + deploy containers
azd up

# After code changes — redeploy only
azd deploy

# View logs and metrics
azd monitor

# Tear down when done
azd down --purge
```

`azd` reads the AppHost resource graph and:

1. Creates an Azure Container Apps Environment
2. Provisions backing resources (PostgreSQL, Redis, Service Bus, etc.)
3. Builds and pushes container images to Azure Container Registry
4. Deploys Container Apps with correct environment variables and secrets

### Via Publisher API (Advanced)

For programmatic control over the generated infrastructure:

```csharp
var env = builder.AddAzureContainerAppEnvironment("production");

builder.AddProject<Projects.Api>("api")
    .PublishAsAzureContainerApp((infra, app) =>
    {
        app.Template.Scale.MinReplicas = 1;
        app.Template.Scale.MaxReplicas = 10;
    });

// Container App Jobs (9.5+)
builder.AddProject<Projects.Migrator>("migrator")
    .PublishAsAzureContainerAppJob();
```

---

## Kubernetes

### Generate Manifests

```bash
# Install Aspire CLI if not already installed
dotnet tool install -g Aspire.Cli

# Generate Kubernetes manifests
aspire publish --output ./k8s --format kubernetes
```

### Customize via Publisher API

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

### Apply to Cluster

```bash
kubectl apply -f ./k8s/
```

> **Important:** Generated manifests lack production hardening (resource limits, network
> policies, pod disruption budgets). Overlay with Kustomize or Helm before deploying to
> production. See [kubernetes instructions](../../../instructions/kubernetes-deployment-best-practices.instructions.md).

---

## Docker Compose

### Generate Compose File

```bash
aspire publish --format docker-compose --output ./compose
```

### Customize via Publisher API

```csharp
builder.AddDockerComposeEnvironment("env")
    .ConfigureComposeFile(file =>
    {
        file.Name = "my-app";
    })
    .WithDashboard();  // Include Aspire Dashboard

builder.AddContainer("api", "myapp-api:<pinned-version>")
    .PublishAsDockerComposeService((resource, service) =>
    {
        service.Restart = "always";
        service.Labels["traefik.enable"] = "true";
    });
```

### Deploy

```bash
docker compose -f ./compose/docker-compose.yml up -d
```

---

## Azure App Service (9.3+)

```csharp
builder.AddAzureAppServiceEnvironment("env");

builder.AddProject<Projects.Api>("api")
    .WithExternalHttpEndpoints()
    .PublishAsAzureAppServiceWebsite((infra, site) =>
    {
        site.SiteConfig.IsWebSocketsEnabled = true;
    });
```

Requires the project to expose a single public HTTP endpoint. Containers publish
to Azure Container Registry and deploy as Linux Web Apps.

---

## Post-Deployment Validation

After deploying to any target:

1. **Health checks** — verify `/health/ready` and `/health/live` respond with 200
2. **Connection strings** — confirm all `ConnectionStrings__*` env vars are populated
3. **OTEL telemetry** — verify traces/metrics appear in your production backend
4. **Service discovery** — confirm inter-service communication works (DNS or config-based)
5. **Scale rules** — test scaling behavior under load if applicable

---

## Common Deployment Issues

| Symptom                              | Cause                                          | Fix                                              |
| ------------------------------------ | ---------------------------------------------- | ------------------------------------------------ |
| Missing connection strings           | AppHost env vars not replicated to prod config | Set `ConnectionStrings__*` in deployment config  |
| Service discovery failures           | No AppHost in prod; DNS not configured         | Use config-based or DNS service discovery        |
| OTEL data missing                    | No OTLP endpoint configured                    | Set `OTEL_EXPORTER_OTLP_ENDPOINT` env var        |
| Container image pull failures (ACA)  | ACR not linked to ACA environment              | Verify ACR admin or managed identity credentials |
| K8s manifest missing resource limits | Generated manifests are scaffold-quality       | Add limits via Kustomize overlay or Helm values  |
