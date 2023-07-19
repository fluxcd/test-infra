module "tags" {
  source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/utils/tags"

  tags = var.tags
}

resource "azurerm_container_registry" "this" {
  name                = var.name
  resource_group_name = var.resource_group
  location            = var.location
  sku                 = "Basic"
  admin_enabled       = var.admin_enabled
  tags                = module.tags.tags
}

# Add the role to each identity that was passed in.
resource "azurerm_role_assignment" "kubweb_to_acr" {
  count                = length(var.aks_principal_id)
  scope                = azurerm_container_registry.this.id
  role_definition_name = "AcrPull"
  principal_id         = var.aks_principal_id[count.index]
}
