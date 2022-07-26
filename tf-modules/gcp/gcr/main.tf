data "google_client_config" "this" {}

data "google_container_registry_repository" "this" {
  region = var.gcr_region
}

resource "google_artifact_registry_repository" "this" {
  provider = google-beta

  project       = data.google_client_config.this.project
  location      = data.google_client_config.this.region
  repository_id = var.name
  description   = "example docker repository"
  format        = "DOCKER"
}
