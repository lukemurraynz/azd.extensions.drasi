# Identity Checklist

## Managed Identity

- [ ] Managed Identity enabled on all Azure compute resources
- [ ] System-Assigned used for single-purpose resources
- [ ] User-Assigned used when sharing identity across resources
- [ ] `DefaultAzureCredential` used in application code
- [ ] No connection strings with embedded credentials in code or config

## RBAC

- [ ] Built-in roles used (no custom roles unless absolutely necessary)
- [ ] Roles scoped to individual resources (not resource group or subscription)
- [ ] Data plane roles used for data access (not management plane roles)
- [ ] Role assignment names use deterministic GUIDs
- [ ] `principalType: 'ServicePrincipal'` set on all service identity assignments

## Local Auth Disabled

- [ ] Application Insights: `DisableLocalAuth: true`
- [ ] Service Bus: `disableLocalAuth: true`
- [ ] Event Hubs: `disableLocalAuth: true`
- [ ] Cosmos DB: `disableLocalAuth: true`
- [ ] Storage: `allowSharedKeyAccess: false`

## Workload Identity (if AKS)

- [ ] OIDC issuer enabled on AKS cluster
- [ ] User-Assigned identity created for each workload
- [ ] Federated credential configured with correct subject
- [ ] Kubernetes service account annotated with identity client ID
- [ ] `azure-workload-identity` webhook installed

## Passwordless Connections

- [ ] Azure SQL uses Entra-only authentication
- [ ] Storage accessed via URI + credential (not connection string)
- [ ] Key Vault accessed via URI + credential
- [ ] Service Bus/Event Hubs use namespace URI + credential

## Infrastructure as Code

- [ ] All identity resources defined in Bicep/Terraform
- [ ] Role assignments created in same deployment as consuming resources
- [ ] No secrets or credentials stored in IaC parameter files
