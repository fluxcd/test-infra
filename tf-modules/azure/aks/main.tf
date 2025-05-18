module "tags" {
  source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/utils/tags"

  tags = var.tags
}

resource "azurerm_resource_group" "this" {
  name     = var.name
  location = var.location
  tags     = module.tags.tags
}

resource "azurerm_kubernetes_cluster" "this" {
  name                = var.name
  location            = azurerm_resource_group.this.location
  resource_group_name = azurerm_resource_group.this.name
  dns_prefix          = var.name
  default_node_pool {
    name            = "default"
    node_count      = 2
    vm_size         = "Standard_B2s"
    os_disk_size_gb = 30
    upgrade_settings {
      drain_timeout_in_minutes      = 0
      max_surge                     = "10%"
      node_soak_duration_in_minutes = 0
    }
  }
  identity {
    type = "SystemAssigned"
  }
  role_based_access_control_enabled = true
  oidc_issuer_enabled               = var.enable_wi
  workload_identity_enabled         = var.enable_wi
  network_profile {
    network_plugin = "kubenet"
    network_policy = "calico"
  }
  tags = module.tags.tags
  dynamic "azure_active_directory_role_based_access_control" {
    for_each = var.aad_rbac_tenant_id != "" ? [1] : []
    content {
      azure_rbac_enabled = true
      tenant_id          = var.aad_rbac_tenant_id
    }
  }
}
