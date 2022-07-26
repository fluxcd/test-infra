output "registry_url" {
  description = "ACR registry URL"
  value       = azurerm_container_registry.this.login_server
}

output "registry_id" {
  description = "ACR Registry ID"
  value       = azurerm_container_registry.this.id
}
