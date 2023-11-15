provider "azurerm" {
  skip_provider_registration = true
  features {}
}

data "azurerm_subscription" "primary" {}
data "azuread_client_config" "current" {}

# Register application, create service principal for it and generate client
# secret.
resource "azuread_application" "app" {
  display_name = var.azure_app_name
  owners       = concat(var.azure_owners, [data.azuread_client_config.current.object_id])
}

resource "azuread_service_principal" "app_sp" {
  client_id = azuread_application.app.client_id
  use_existing   = true
}

resource "azuread_application_password" "app_secret" {
  application_id = resource.azuread_application.app.id
  display_name          = var.azure_app_secret_name
}

# Define custom role.
resource "azurerm_role_definition" "role" {
  name        = var.azure_app_name
  scope       = data.azurerm_subscription.primary.id
  description = var.azure_app_description

  permissions {
    actions = var.azure_permissions
  }

  assignable_scopes = [
    data.azurerm_subscription.primary.id,
  ]
}

# Assign role to the registered app's service principal.
resource "azurerm_role_assignment" "assignment" {
  scope              = data.azurerm_subscription.primary.id
  role_definition_id = azurerm_role_definition.role.role_definition_resource_id
  principal_id       = azuread_service_principal.app_sp.object_id
}

# Add client details in github repo.
resource "github_actions_secret" "client_id" {
  repository      = var.github_project
  secret_name     = var.github_secret_client_id_name
  plaintext_value = azuread_application.app.client_id
}

resource "github_actions_secret" "client_secret" {
  repository      = var.github_project
  secret_name     = var.github_secret_client_secret_name
  plaintext_value = azuread_application_password.app_secret.value
}

resource "github_actions_secret" "subscription_id" {
  repository      = var.github_project
  secret_name     = var.github_secret_subscription_id_name
  plaintext_value = data.azurerm_subscription.primary.subscription_id
}

resource "github_actions_secret" "tenant_id" {
  repository      = var.github_project
  secret_name     = var.github_secret_tenant_id_name
  plaintext_value = azuread_service_principal.app_sp.application_tenant_id
}

resource "github_actions_variable" "location" {
  repository    = var.github_project
  variable_name = var.github_variable_location_name
  value         = var.azure_location
}

resource "github_actions_variable" "custom" {
  for_each = var.github_variable_custom

  repository    = var.github_project
  variable_name = each.key
  value         = each.value
}

resource "github_actions_secret" "custom" {
  # Mark only the key as nonsensitive.
  for_each = nonsensitive(toset(keys(var.github_secret_custom)))

  repository      = var.github_project
  secret_name     = each.key
  plaintext_value = var.github_secret_custom[each.key]
}
