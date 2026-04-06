# main.tf — Drasi infrastructure module
# Provisions all Azure resources required for a Drasi workload:
#   Log Analytics, Managed Identities, Key Vault, NAT Gateway, VNet, AKS,
#   role assignments, and workload identity federation.

terraform {
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
}

data "azurerm_client_config" "current" {}

data "azurerm_resource_group" "current" {
  name = var.resource_group_name
}

locals {
  # Deterministic suffix derived from resource group ID + environment name,
  # mirroring the Bicep uniqueString(resourceGroup().id, environmentName) pattern.
  suffix = substr(sha256("${data.azurerm_resource_group.current.id}${var.environment_name}"), 0, 13)

  aks_name     = "drasi-aks-${local.suffix}"
  kv_name      = "drasi-kv-${local.suffix}"
  law_name     = "drasi-law-${local.suffix}"
  uami_name    = "drasi-id-${local.suffix}"
  cp_uami_name = "drasi-cp-${local.suffix}"
  vnet_name    = "drasi-vnet-${local.suffix}"
  nat_gw_name  = "drasi-natgw-${local.suffix}"
  pip_name     = "drasi-pip-${local.suffix}"
}

# ---------------------------------------------------------------------------
# Log Analytics Workspace (AVM-TF)
# ---------------------------------------------------------------------------
module "log_analytics" {
  source  = "Azure/avm-res-operationalinsights-workspace/azurerm"
  version = "~> 0.5"

  name                = local.law_name
  resource_group_name = data.azurerm_resource_group.current.name
  location            = var.location
  tags                = var.tags
}

# ---------------------------------------------------------------------------
# User-Assigned Managed Identities
# ---------------------------------------------------------------------------

# Workload identity — used by Drasi resource provider pods to authenticate to
# Key Vault and publish metrics via Workload Identity federation.
resource "azurerm_user_assigned_identity" "workload" {
  name                = local.uami_name
  resource_group_name = data.azurerm_resource_group.current.name
  location            = var.location
  tags                = var.tags
}

# Control-plane identity — used as the AKS cluster identity for ARM operations
# (load balancer management, VNet route table updates, etc.).
resource "azurerm_user_assigned_identity" "control_plane" {
  name                = local.cp_uami_name
  resource_group_name = data.azurerm_resource_group.current.name
  location            = var.location
  tags                = var.tags
}

# ---------------------------------------------------------------------------
# Key Vault (AVM-TF)
# ---------------------------------------------------------------------------
module "key_vault" {
  source  = "Azure/avm-res-keyvault-vault/azurerm"
  version = "~> 0.10"

  name                = local.kv_name
  resource_group_name = data.azurerm_resource_group.current.name
  location            = var.location
  tenant_id           = data.azurerm_client_config.current.tenant_id
  sku_name            = "standard"
  tags                = var.tags

  # RBAC auth is enabled by default in the AVM module (legacy_access_policies_enabled = false).
  # Drasi uses workload identity to access secrets.

  # Purge protection and soft-delete prevent accidental permanent deletion.
  purge_protection_enabled   = true
  soft_delete_retention_days = 7
}

# Key Vault Secrets User — lets Drasi pods read secrets without broader permissions.
resource "azurerm_role_assignment" "kv_secrets_user" {
  scope                = module.key_vault.resource_id
  role_definition_name = "Key Vault Secrets User"
  principal_id         = azurerm_user_assigned_identity.workload.principal_id
}

# ---------------------------------------------------------------------------
# Public IP for NAT Gateway (zone-redundant)
# ---------------------------------------------------------------------------
resource "azurerm_public_ip" "nat" {
  name                = local.pip_name
  resource_group_name = data.azurerm_resource_group.current.name
  location            = var.location
  allocation_method   = "Static"
  sku                 = "Standard"
  # Zone-redundant: prefix addresses cover all zones without pinning to one zone.
  zones = ["1", "2", "3"]
  tags  = var.tags
}

# ---------------------------------------------------------------------------
# NAT Gateway
# ---------------------------------------------------------------------------
resource "azurerm_nat_gateway" "main" {
  name                    = local.nat_gw_name
  resource_group_name     = data.azurerm_resource_group.current.name
  location                = var.location
  sku_name                = "Standard"
  idle_timeout_in_minutes = 10
  # Zone 1 for predictable placement; the NAT GW itself is zonal.
  zones = ["1"]
  tags  = var.tags
}

resource "azurerm_nat_gateway_public_ip_association" "main" {
  nat_gateway_id       = azurerm_nat_gateway.main.id
  public_ip_address_id = azurerm_public_ip.nat.id
}

# ---------------------------------------------------------------------------
# Virtual Network (AVM-TF)
# ---------------------------------------------------------------------------
module "vnet" {
  source  = "Azure/avm-res-network-virtualnetwork/azurerm"
  version = "~> 0.17"

  name                = local.vnet_name
  resource_group_name = data.azurerm_resource_group.current.name
  location            = var.location
  tags                = var.tags

  address_space = ["10.0.0.0/16"]

  subnets = {
    snet-aks = {
      name             = "snet-aks"
      address_prefixes = ["10.0.0.0/22"]
      # Associate the NAT Gateway so outbound traffic from AKS nodes uses a
      # predictable public IP instead of ephemeral SNAT.
      nat_gateway = {
        id = azurerm_nat_gateway.main.id
      }
    }
  }
}

# Network Contributor on the VNet lets the AKS control-plane manage load
# balancers, route tables, and public IPs within the VNet.
resource "azurerm_role_assignment" "vnet_network_contributor" {
  scope                = module.vnet.resource_id
  role_definition_name = "Network Contributor"
  principal_id         = azurerm_user_assigned_identity.control_plane.principal_id
}

# ---------------------------------------------------------------------------
# AKS Cluster
# ---------------------------------------------------------------------------
resource "azurerm_kubernetes_cluster" "main" {
  name                = local.aks_name
  resource_group_name = data.azurerm_resource_group.current.name
  location            = var.location
  dns_prefix          = local.aks_name
  tags                = var.tags

  # Use the control-plane UAMI for ARM operations (load balancer, routes, etc.).
  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.control_plane.id]
  }

  # System node pool — reserved for kube-system and Drasi control-plane pods.
  default_node_pool {
    name       = "systempool"
    node_count = 3
    vm_size    = "Standard_D4s_v5"
    os_sku     = "AzureLinux"

    # Zone-spread for system pool: spread across all three zones.
    zones = ["1", "2", "3"]

    # Restricts this pool to kube-system and critical addon pods.
    # Equivalent to the CriticalAddonsOnly=true:NoSchedule taint.
    # node_taints is not valid on the default_node_pool in azurerm v4;
    # use only_critical_addons_enabled instead.
    only_critical_addons_enabled = true

    vnet_subnet_id = module.vnet.subnets["snet-aks"].resource_id

    upgrade_settings {
      max_surge = "10%"
    }
  }

  # Azure CNI Overlay with Cilium for high pod-density without consuming VNet
  # address space per pod; Cilium provides eBPF-accelerated data plane.
  network_profile {
    network_plugin      = "azure"
    network_plugin_mode = "overlay"
    network_data_plane  = "cilium"
    # network_policy must be set to "cilium" alongside network_data_plane = "cilium"
    # to activate Cilium's eBPF-based NetworkPolicy enforcement. Without this,
    # Cilium is installed as the data plane only; policy enforcement is not activated.
    network_policy = "cilium"
    pod_cidr       = "192.168.0.0/16"
    service_cidr   = "172.16.0.0/16"
    dns_service_ip = "172.16.0.10"
    outbound_type  = "userAssignedNATGateway"
  }

  # OIDC issuer is required for Workload Identity federation.
  oidc_issuer_enabled = true

  # Workload Identity lets Drasi pods exchange Kubernetes service account
  # tokens for Entra ID tokens without storing credentials in pods.
  workload_identity_enabled = true

  # Disk CSI driver is required by Drasi for persistent volume claims.
  storage_profile {
    disk_driver_enabled = true
  }

  # Disable local (certificate-based) accounts — all access via Entra ID.
  local_account_disabled = true

  azure_active_directory_role_based_access_control {
    azure_rbac_enabled = true
  }

  # Auto-upgrade: stable channel keeps nodes on the latest stable patch.
  automatic_upgrade_channel = "stable"

  # NodeImage channel: keeps node OS images up-to-date independently of k8s.
  node_os_upgrade_channel = "NodeImage"

  # Container Insights for cluster-level metrics and log collection.
  oms_agent {
    log_analytics_workspace_id = module.log_analytics.resource_id
  }
}

# Workload node pool — Drasi sources, queries, and reactions run here.
resource "azurerm_kubernetes_cluster_node_pool" "workload" {
  name                  = "workload"
  kubernetes_cluster_id = azurerm_kubernetes_cluster.main.id
  vm_size               = "Standard_D4s_v5"
  node_count            = 2
  os_sku                = "AzureLinux"
  mode                  = "User"
  zones                 = ["1", "2", "3"]
  vnet_subnet_id        = module.vnet.subnets["snet-aks"].resource_id
  tags                  = var.tags

  upgrade_settings {
    max_surge = "10%"
  }
}

# ---------------------------------------------------------------------------
# Role assignments for the workload identity
# ---------------------------------------------------------------------------

# Monitoring Metrics Publisher — allows the workload UAMI (used by the
# Container Insights agent) to publish custom metrics to Azure Monitor.
resource "azurerm_role_assignment" "monitoring_metrics_publisher" {
  scope                = data.azurerm_resource_group.current.id
  role_definition_name = "Monitoring Metrics Publisher"
  principal_id         = azurerm_user_assigned_identity.workload.principal_id
}

# ---------------------------------------------------------------------------
# Workload Identity federation
# ---------------------------------------------------------------------------

# Federated credential: the AKS OIDC issuer can exchange Kubernetes SA tokens
# (namespace/drasi-system, SA/drasi-resource-provider) for Entra ID tokens
# scoped to the workload UAMI.
resource "azurerm_federated_identity_credential" "drasi" {
  name                = "drasi-federated-credential"
  resource_group_name = data.azurerm_resource_group.current.name
  parent_id           = azurerm_user_assigned_identity.workload.id
  audience            = ["api://AzureADTokenExchange"]
  issuer              = azurerm_kubernetes_cluster.main.oidc_issuer_url
  subject             = "system:serviceaccount:${var.drasi_namespace}:${var.drasi_service_account_name}"
}
