# outputs.tf — Root module outputs
# azd reads these values after `terraform apply` and stores them as azd
# environment variables so the extension can resolve AKS context, Key Vault
# name, and workload identity details without additional lookups.

output "aks_cluster_name" {
  description = "Name of the AKS cluster."
  value       = module.drasi_infra.aks_cluster_name
}

output "aks_oidc_issuer_url" {
  description = "OIDC issuer URL for the AKS cluster."
  value       = module.drasi_infra.aks_oidc_issuer_url
}

output "aks_resource_id" {
  description = "ARM resource ID of the AKS cluster."
  value       = module.drasi_infra.aks_resource_id
}

output "key_vault_name" {
  description = "Name of the Key Vault instance."
  value       = module.drasi_infra.key_vault_name
}

output "key_vault_uri" {
  description = "URI of the Key Vault."
  value       = module.drasi_infra.key_vault_uri
}

output "log_analytics_workspace_id" {
  description = "ARM resource ID of the Log Analytics workspace."
  value       = module.drasi_infra.log_analytics_workspace_id
}

output "uami_client_id" {
  description = "Client ID of the workload user-assigned managed identity."
  value       = module.drasi_infra.uami_client_id
}

output "uami_principal_id" {
  description = "Principal ID of the workload UAMI."
  value       = module.drasi_infra.uami_principal_id
}

output "uami_resource_id" {
  description = "ARM resource ID of the workload UAMI."
  value       = module.drasi_infra.uami_resource_id
}

output "vnet_resource_id" {
  description = "ARM resource ID of the virtual network."
  value       = module.drasi_infra.vnet_resource_id
}
