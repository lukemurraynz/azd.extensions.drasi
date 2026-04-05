# Secret Management Checklist

## Key Vault Configuration

- [ ] `enableRbacAuthorization: true` — no access policies
- [ ] `enableSoftDelete: true` — always enabled
- [ ] `enablePurgeProtection: true` — enabled in production
- [ ] `publicNetworkAccess: 'Disabled'` — private endpoint only in production
- [ ] `networkAcls.bypass: 'AzureServices'` — allow trusted services
- [ ] Diagnostic settings configured to Log Analytics

## RBAC Access

- [ ] Managed identity used for all application access
- [ ] Least-privilege roles assigned (`Secrets User` for read, not `Administrator`)
- [ ] Role assignments scoped to vault (not subscription)
- [ ] No service principal secrets for Key Vault access

## Secret Management

- [ ] Secrets stored in Key Vault, not app settings or config files
- [ ] Secret names follow naming convention (lowercase with hyphens)
- [ ] Expiration dates set on all secrets
- [ ] Secret versions used (not overwriting same version)
- [ ] Identity-based connections preferred over Key Vault secrets

## Key Vault References

- [ ] App Service/Functions use Key Vault references for secrets
- [ ] System-assigned managed identity has `Key Vault Secrets User` role
- [ ] Key Vault reference syntax correct: `@Microsoft.KeyVault(SecretUri=...)`

## Secret Rotation

- [ ] Event Grid subscription for `SecretNearExpiry` events
- [ ] Rotation function deployed and tested
- [ ] Rotation verified with zero-downtime
- [ ] Alert on rotation failure

## AKS CSI Driver (if applicable)

- [ ] CSI Secret Store driver enabled on cluster
- [ ] Workload Identity configured for pods that access Key Vault
- [ ] SecretProviderClass configured with correct identity
- [ ] Pods mount secrets as volumes (not environment variables)
- [ ] Secret sync to Kubernetes Secret configured (if needed)

## Infrastructure as Code

- [ ] Key Vault deployed via Bicep/Terraform
- [ ] RBAC role assignments in IaC
- [ ] Private endpoint configured in IaC
- [ ] No secrets stored in IaC files or parameter files
