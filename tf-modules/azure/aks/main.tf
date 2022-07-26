resource "azurerm_resource_group" "this" {
  name     = var.name
  location = var.location
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
  }
  identity {
    type = "SystemAssigned"
  }
  role_based_access_control_enabled = true
  network_profile {
    network_plugin = "kubenet"
    network_policy = "calico"
  }
}
