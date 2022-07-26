output "gcr_repository_url" {
  description = "URL of the GCR repository"
  value       = data.google_container_registry_repository.this.repository_url
}

output "artifact_repository_id" {
  description = "ID of the Artifact Repository"
  value       = google_artifact_registry_repository.this.repository_id
}
