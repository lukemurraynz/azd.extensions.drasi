# main.tf — Root Terraform configuration for a Drasi workload
# Wires the drasi-infra module and exposes its outputs as root-level outputs
# that azd reads via `azd env set` mappings in azure.yaml.

terraform {
  required_version = ">= 1.9.0"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.0"
    }
    azuread = {
      source  = "hashicorp/azuread"
      version = "~> 2.0"
    }
  }

  # Remote backend: uncomment and configure before deploying to shared environments.
  # backend "azurerm" {
  #   resource_group_name  = "terraform-state-rg"
  #   storage_account_name = "tfstatestorage"
  #   container_name       = "tfstate"
  #   key                  = "drasi.terraform.tfstate"
  # }
}

provider "azurerm" {
  features {}
}

provider "azuread" {}

module "drasi_infra" {
  source = "./modules/drasi-infra"

  resource_group_name = var.resource_group_name
  location            = var.location
  environment_name    = var.environment_name
  tags                = var.tags
}
