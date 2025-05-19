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

output "full_name" {
  description = "Full name of the GKE cluster"
  value       = "projects/${data.google_client_config.this.project}/locations/${data.google_client_config.this.region}/clusters/${var.name}"
}

output "endpoint" {
  description = "The endpoint of the GKE cluster"
  value       = module.gke_auth.host
}
