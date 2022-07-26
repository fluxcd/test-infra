output "kubeconfig" {
  description = "kubeconfig of the created GKE cluster"
  value       = module.gke_auth.kubeconfig_raw
  sensitive   = true
}

output "project" {
  description = "GCP project in which the GKE cluster is created"
  value       = data.google_client_config.this.project
}

output "region" {
  description = "GCP region in which the GKE cluster is created"
  value       = data.google_client_config.this.region
}
