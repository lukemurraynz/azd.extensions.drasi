# variables.tf — Input variables for the drasi-infra module

variable "resource_group_name" {
  type        = string
  description = "Name of the Azure resource group that will contain all Drasi resources."

  validation {
    condition     = length(var.resource_group_name) >= 1 && length(var.resource_group_name) <= 90
    error_message = "resource_group_name must be between 1 and 90 characters."
  }
}

variable "location" {
  type        = string
  description = "Azure region to deploy resources into. Must match the resource group's region."
}

variable "environment_name" {
  type        = string
  description = "Short environment identifier (e.g. dev, staging, prod). Used to derive unique resource name suffixes."

  validation {
    condition     = length(var.environment_name) >= 1 && length(var.environment_name) <= 24
    error_message = "environment_name must be between 1 and 24 characters."
  }
}

variable "tags" {
  type        = map(string)
  description = "Resource tags applied to all created resources."
  default     = {}
}

variable "drasi_namespace" {
  type        = string
  description = "Kubernetes namespace that Drasi is installed into. Used to scope the Workload Identity federated credential."
  default     = "drasi-system"
}

variable "drasi_service_account_name" {
  type        = string
  description = "Kubernetes service account name used by the Drasi resource-provider pods. Used to scope the Workload Identity federated credential."
  default     = "drasi-resource-provider"
}
