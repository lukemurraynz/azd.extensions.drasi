# outputs.tf — Outputs from the drasi-infra module
# Exposed so the root module can forward them as azd environment values and
# so the azd extension can resolve AKS context, Key Vault name, and UAMI
# client-id after provisioning completes.

output "aks_cluster_name" {
  description = "Name of the AKS cluster."
  value       = azurerm_kubernetes_cluster.main.name
}

output "aks_oidc_issuer_url" {
  description = "OIDC issuer URL for the AKS cluster. Required to configure Workload Identity federation."
  value       = azurerm_kubernetes_cluster.main.oidc_issuer_url
}

output "aks_resource_id" {
  description = "ARM resource ID of the AKS cluster."
  value       = azurerm_kubernetes_cluster.main.id
}

output "key_vault_name" {
  description = "Name of the Key Vault instance."
  value       = module.key_vault.name
}

output "key_vault_uri" {
  description = "URI of the Key Vault (https://<name>.vault.azure.net/)."
  # The AVM Key Vault module exposes the vault's properties via `resource` output.
  value = module.key_vault.resource.properties.vaultUri
}

output "log_analytics_workspace_id" {
  description = "ARM resource ID of the Log Analytics workspace."
  value       = module.log_analytics.resource_id
}

output "uami_client_id" {
  description = "Client ID of the workload user-assigned managed identity. Set as AZURE_CLIENT_ID in Drasi pods."
  value       = azurerm_user_assigned_identity.workload.client_id
}

output "uami_principal_id" {
  description = "Principal (object) ID of the workload UAMI. Used for RBAC role assignments."
  value       = azurerm_user_assigned_identity.workload.principal_id
}

output "uami_resource_id" {
  description = "ARM resource ID of the workload UAMI."
  value       = azurerm_user_assigned_identity.workload.id
}

output "vnet_resource_id" {
  description = "ARM resource ID of the virtual network."
  value       = module.vnet.resource_id
}
