---
name: postgresql-npgsql
description: >-
  Patterns for Azure PostgreSQL Flexible Server with .NET 10 / EF Core 10 / Npgsql. USE FOR: connection string password encoding, EF Core migrations in Docker, fail-fast startup, PostGIS spatial queries, Drasi CDC logical replication setup, Kubernetes Secret management.
license: MIT
---

# PostgreSQL + Npgsql Skill

Patterns for Azure PostgreSQL Flexible Server 17 with EF Core 10 and Npgsql in this project. These patterns address real failure modes documented in the repository's deployment history.

---

## 1. Password Encoding — CRITICAL

**Problem:** PostgreSQL/Npgsql connection strings treat `/`, `+`, `@`, `;` as delimiters. A generated password such as `/yoaZxbB8+sNk66=` will silently corrupt the connection string, producing `Failed to connect` at startup.

**Rule:** Always URL-encode the password **before** embedding it in a connection string.

```csharp
// ✅ CORRECT — encode before building the connection string
var rawPassword = configuration["Database:Password"];
var encodedPassword = System.Net.WebUtility.UrlEncode(rawPassword);

var builder = new NpgsqlConnectionStringBuilder
{
    Host     = configuration["Database:Host"],
    Database = configuration["Database:Name"],
    Username = configuration["Database:User"],
    Password = encodedPassword,   // encoded value
    SslMode  = SslMode.Require,
};

services.AddDbContext<AppDbContext>(o => o.UseNpgsql(builder.ConnectionString));
```

```powershell
# PowerShell — encode password when constructing K8s Secret value
$raw     = az keyvault secret show --name db-password --vault-name $vault --query value -o tsv
$encoded = [System.Net.WebUtility]::UrlEncode($raw)
$connStr = "Server=$host;Username=$user;Password=$encoded;Database=$db;SSL Mode=Require"
$b64     = [Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes($connStr))
```

> [!WARNING]
> The connection string variable contains the database password in plaintext. Never log, echo, or write this variable to files. Base64 encoding (for Kubernetes secrets) is encoding, not encryption. Use `Write-Verbose` sparingly and ensure CI/CD logs mask sensitive variables.

Characters that require encoding:`/`→`%2F`, `+`→`%2B`, `@`→`%40`, `;`→`%3B`, `=`→`%3D`.

---

## 2. EF Core Migrations — Docker Build Requirements

**Problem:** Containers call `Database.Migrate()` at startup but the migrations directory is absent from the image because it was never generated locally or was excluded by `.dockerignore`.

**Rules:**

1. Run `dotnet ef migrations add` **locally** before building the Docker image.
2. Never exclude the `Migrations/` directory in `.dockerignore`.
3. Add a Dockerfile guard that fails the build if migrations are missing.

```bash
# Generate migrations locally (run before docker build)
dotnet ef migrations add InitialCreate \
  --project backend/src/EmergencyAlerts.Infrastructure \
  --startup-project backend/src/EmergencyAlerts.Api

dotnet ef migrations list \
  --project backend/src/EmergencyAlerts.Infrastructure \
  --startup-project backend/src/EmergencyAlerts.Api
```

```dockerfile
# In multi-stage Dockerfile — validate migrations exist
FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build
WORKDIR /src
COPY . .

# Guard: fail the build if no migration files are present
RUN find . -path "*/Migrations/*_*.cs" | grep -v Designer | grep -q . || \
    (echo "ERROR: No EF Core migrations found. Run 'dotnet ef migrations add' locally first." && exit 1)

RUN dotnet publish src/EmergencyAlerts.Api/EmergencyAlerts.Api.csproj \
    -c Release -o /app/publish
```

`.dockerignore` — do NOT add this line:

```
# ❌ Do not exclude migrations
**/Migrations/
```

---

## 3. Fail-Fast Startup Pattern

**Problem:** Wrapping `Database.Migrate()` in a try-catch that logs a warning and continues allows the pod to pass readiness probes while the database schema is absent, producing HTTP 500 on first data request.

```csharp
// ✅ CORRECT — fail fast so Kubernetes restarts the pod and surfaces the error
try
{
    using var scope = app.Services.CreateScope();
    var db = scope.ServiceProvider.GetRequiredService<AppDbContext>();
    logger.LogInformation("Applying EF Core migrations...");
    await db.Database.MigrateAsync();
    logger.LogInformation("✓ Migrations applied");
}
catch (Exception ex)
{
    logger.LogCritical(ex, "❌ Database migration failed — stopping application");
    Environment.Exit(1);   // pod restarts; Kubernetes surfaces CrashLoopBackOff
}
```

Add a database health check so the readiness probe reflects real schema health:

```csharp
builder.Services.AddHealthChecks()
    .AddDbContextCheck<AppDbContext>(
        name: "database",
        failureStatus: HealthStatus.Unhealthy,
        tags: ["ready"]);

app.MapHealthChecks("/health/ready", new HealthCheckOptions
{
    Predicate = c => c.Tags.Contains("ready"),
});
```

---

## 4. Drasi CDC — Logical Replication Configuration

Drasi uses the Debezium PostgreSQL connector, which requires `wal_level = logical` on the server. This is a server-level parameter — it cannot be set per database.

**Bicep (PostgreSQL Flexible Server):**

```bicep
resource pgConfig 'Microsoft.DBforPostgreSQL/flexibleServers/configurations@2025-08-01' = {
  parent: postgresServer
  name: 'wal_level'
  properties: {
    value: 'logical'
    source: 'user-override'
  }
}

resource pgMaxSlots 'Microsoft.DBforPostgreSQL/flexibleServers/configurations@2025-08-01' = {
  parent: postgresServer
  name: 'max_replication_slots'
  properties: { value: '10', source: 'user-override' }
}

resource pgMaxSenders 'Microsoft.DBforPostgreSQL/flexibleServers/configurations@2025-08-01' = {
  parent: postgresServer
  name: 'max_wal_senders'
  properties: { value: '10', source: 'user-override' }
}
```

The Drasi PostgreSQL source also requires the `azure.extensions` server parameter to include `postgis` if spatial data types are used:

```bicep
resource pgExtensions 'Microsoft.DBforPostgreSQL/flexibleServers/configurations@2025-08-01' = {
  parent: postgresServer
  name: 'azure.extensions'
  properties: { value: 'postgis', source: 'user-override' }
}
```

---

## 5. PostGIS Spatial Queries (EF Core)

```csharp
// Program.cs — enable NetTopologySuite for spatial types
builder.Services.AddDbContext<AppDbContext>(o =>
    o.UseNpgsql(connectionString, npgsql =>
        npgsql.UseNetTopologySuite()));
```

```csharp
// Entity
public class Alert
{
    public int Id { get; set; }
    public NetTopologySuite.Geometries.Point? Location { get; set; }
}

// EF Core query — find alerts within 50 km of a point
var nearby = await db.Alerts
    .Where(a => a.Location != null &&
                a.Location.Distance(referencePoint) < 50_000)
    .ToListAsync();
```

```sql
-- Verify PostGIS extension installed
SELECT PostGIS_Version();

-- Confirm spatial index exists
SELECT indexname FROM pg_indexes
WHERE tablename = 'Alerts' AND indexdef LIKE '%gist%';
```

---

## 6. Kubernetes Secret — Connection String

```yaml
# infrastructure/k8s/secrets.yaml
# Placeholders are substituted by apply-k8s-manifests.ps1 before kubectl apply
apiVersion: v1
kind: Secret
metadata:
  name: emergency-alerts-secrets
  namespace: emergency-alerts
type: Opaque
stringData:
  ConnectionStrings__Default: "Server=${POSTGRES_HOST};Port=${POSTGRES_PORT};Database=${POSTGRES_DATABASE};Username=${POSTGRES_USER};Password=${POSTGRES_PASSWORD};SSL Mode=Require"
```

```powershell
# Verify no placeholders remain after apply
$decoded = kubectl get secret emergency-alerts-secrets -n emergency-alerts \
  -o jsonpath='{.data.ConnectionStrings__Default}' | base64 -d
if ($decoded -match '\$\{') {
    Write-Error "Unresolved placeholder in Secret: $decoded"
    exit 1
}
```

---

## 7. Common Failure Modes

| Symptom                                   | Root Cause                                                             | Fix                                                                               |
| ----------------------------------------- | ---------------------------------------------------------------------- | --------------------------------------------------------------------------------- |
| `Failed to connect to <host>:5432`        | Password with `/` or `+` not URL-encoded                               | URL-encode password before building connection string                             |
| `error NETSDK1004: Assets file not found` | `--no-restore` used without prior `dotnet restore` in that build stage | Add `dotnet restore` before `dotnet build --no-restore`                           |
| Pod healthy but API returns HTTP 500      | Migration silently failed; schema absent                               | Use fail-fast pattern — `Environment.Exit(1)` on migration failure                |
| `No migrations found` in container        | `Migrations/` excluded from Docker build context                       | Remove `**/Migrations/` from `.dockerignore`; add Dockerfile guard                |
| `replication slot already exists`         | Drasi source restarted without dropping old slot                       | Run `SELECT pg_drop_replication_slot('<slot>');` before re-deploying Drasi source |
| Slow spatial queries                      | Missing GiST index on geometry column                                  | `CREATE INDEX ON "Alerts" USING GIST ("Location");`                               |

---

## Currency and Verification

- **Date checked:** 2026-03-31
- **API version used:** `2025-08-01` for `Microsoft.DBforPostgreSQL/flexibleServers/configurations`
- **Compatibility:** .NET 10 / EF Core 10 / Npgsql 10.x / PostgreSQL Flexible Server 17
- **Sources:** [PostgreSQL Flexible Server ARM reference](https://learn.microsoft.com/azure/templates/microsoft.dbforpostgresql/flexibleservers), [Npgsql docs](https://www.npgsql.org/doc/)
- **Verification steps:**
  1. Verify API version: `az provider show --namespace Microsoft.DBforPostgreSQL --query "resourceTypes[?resourceType=='flexibleServers'].apiVersions" -o tsv`
  2. Verify PostgreSQL version: `az postgres flexible-server show --name <server> --resource-group <rg> --query version`
  3. Verify Npgsql compatibility: check [Npgsql releases](https://github.com/npgsql/npgsql/releases)

### Known Pitfalls

| Area                 | Pitfall                                                                             | Mitigation                                                     |
| -------------------- | ----------------------------------------------------------------------------------- | -------------------------------------------------------------- |
| Password encoding    | `/`, `+`, `@`, `;` in passwords corrupt Npgsql connection strings                   | URL-encode password before building connection string          |
| `wal_level` change   | Changing `wal_level` to `logical` requires a server restart                         | Plan for downtime or use blue-green deployment                 |
| Migrations in Docker | `Migrations/` directory excluded from Docker build context                          | Remove `**/Migrations/` from `.dockerignore`                   |
| EF Core version      | EF Core 10 ships with .NET 10 — do not mix EF Core 9 packages with `net10.0` target | Upgrade all `Microsoft.EntityFrameworkCore.*` packages to 10.x |

---

## Related Skills

- [.NET Backend Patterns](../dotnet-backend-patterns/SKILL.md) — EF Core and API patterns using Npgsql
- [Drasi Queries](../drasi-queries/SKILL.md) — Continuous queries over PostgreSQL data
