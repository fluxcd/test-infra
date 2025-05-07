output "artifact_repository_id" {
  description = "ID of the Artifact Repository"
  value       = google_artifact_registry_repository.this.repository_id
}
