output "kubeconfig" {
  description = "kubeconfig of the created AKS cluster"
  value       = azurerm_kubernetes_cluster.this.kube_config_raw
  sensitive   = true
}

output "principal_id" {
  description = "Azure Principal ID of the AKS cluster"
  value       = azurerm_kubernetes_cluster.this.kubelet_identity[0].object_id
}

output "kubelet_client_id" {
  description = "Azure Client ID of the AKS cluster"
  value       = azurerm_kubernetes_cluster.this.kubelet_identity[0].client_id
}

output "resource_group" {
  description = "Azure Resource Group in which the AKS cluster is created"
  value       = azurerm_resource_group.this.name
}

output "cluster_oidc_url" {
  description = "The OIDC Issuer URL of the AKS cluster"
  value       = azurerm_kubernetes_cluster.this.oidc_issuer_url
}

output "cluster_id" {
  description = "The ID of the AKS cluster"
  value       = azurerm_kubernetes_cluster.this.id
}
