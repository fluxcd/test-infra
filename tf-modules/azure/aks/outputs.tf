output "kubeconfig" {
  description = "kubeconfig of the created AKS cluster"
  value       = azurerm_kubernetes_cluster.this.kube_config_raw
  sensitive   = true
}

output "principal_id" {
  description = "Azure Principal ID of the AKS cluster"
  value       = azurerm_kubernetes_cluster.this.kubelet_identity[0].object_id
}

output "resource_group" {
  description = "Azure Resource Group in which the AKS cluster is created"
  value       = azurerm_resource_group.this.name
}
