output "repository_url" {
  description = "URL of the Git repository"
  value       = google_sourcerepo_repository.this.url
}
