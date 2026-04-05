# Secret Patterns

## When to Use Key Vault vs Identity-Based Connections

| Scenario                                | Approach                          |
| --------------------------------------- | --------------------------------- |
| Azure service supports managed identity | Identity-based connection (no KV) |
| Third-party API key or password         | Store in Key Vault                |
| Database with no MI support             | Store connection string in KV     |
| TLS certificate for custom domain       | Certificate in Key Vault          |
| Encryption key for application data     | Key in Key Vault (HSM-backed)     |

**Rule:** If a managed identity connection is available, don't use Key Vault for that service.

---

## Secret Naming

| Convention             | Example                                 |
| ---------------------- | --------------------------------------- |
| Lowercase with hyphens | `database-password`                     |
| Service-prefixed       | `stripe-api-key`                        |
| Environment-agnostic   | `external-api-key` (not `prod-api-key`) |

Use separate Key Vaults per environment, not separate secret names.

---

## Secret Lifecycle

| Phase      | Action                                         |
| ---------- | ---------------------------------------------- |
| Creation   | Store via IaC or CLI, never portal copy-paste  |
| Access     | Read via managed identity + RBAC               |
| Rotation   | Automated via Event Grid + Function            |
| Expiry     | Set expiration date, alert before expiry       |
| Revocation | Disable secret version, create new version     |
| Deletion   | Soft delete (recoverable for retention period) |

---

## Caching Secrets

Cache secrets in memory to reduce Key Vault calls:

```csharp
// Good — cache with periodic refresh
public class CachedSecretClient
{
    private readonly SecretClient _client;
    private readonly IMemoryCache _cache;

    public async Task<string> GetSecretAsync(string name)
    {
        return await _cache.GetOrCreateAsync(
            $"kv-{name}",
            async entry =>
            {
                entry.AbsoluteExpirationRelativeToNow = TimeSpan.FromMinutes(15);
                var secret = await _client.GetSecretAsync(name);
                return secret.Value.Value;
            });
    }
}
```

**Guidelines:**

- Cache for 5-15 minutes in production
- Shorter cache for secrets that rotate frequently
- Always handle `SecretNotFound` and `AccessDenied` gracefully

---

## Anti-Patterns

| Anti-Pattern                           | Correct Approach                        |
| -------------------------------------- | --------------------------------------- |
| Secrets in `appsettings.json`          | Key Vault reference or MI connection    |
| Secrets in environment variables       | Key Vault reference in app settings     |
| Secrets in source code                 | Key Vault with managed identity         |
| Shared Key Vault across all envs       | Separate vault per environment          |
| Access policy-based vault              | RBAC-based vault                        |
| Wide roles (Administrator for reading) | Narrow roles (Secrets User for reading) |

---

## Rules

1. Prefer identity-based connections over Key Vault secrets.
2. One Key Vault per environment — not one vault with env-prefixed secrets.
3. Cache secrets in memory — don't read from Key Vault on every request.
4. Set expiration dates on all secrets.
5. Use RBAC roles — never vault access policies.
