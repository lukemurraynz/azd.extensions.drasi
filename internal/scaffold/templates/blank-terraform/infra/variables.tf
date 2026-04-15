# variables.tf — Input variables for the root Terraform configuration

variable "resource_group_name" {
  type        = string
  description = "Name of the Azure resource group to deploy into. Created externally (e.g. by azd provision)."
}

variable "location" {
  type        = string
  description = "Azure region for all resources. azd sets this from AZURE_LOCATION environment variable."
  default     = "" # Empty default defers to azd's AZURE_LOCATION
}

variable "environment_name" {
  type        = string
  description = "Short environment identifier (e.g. dev, staging, prod). Drives unique suffix generation."
}

variable "tags" {
  type        = map(string)
  description = "Tags applied to all resources."
  default     = {}
}
