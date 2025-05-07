module "tags" {
  source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/utils/tags"

  tags = var.tags
}

data "google_client_config" "this" {}

resource "google_artifact_registry_repository" "this" {
  provider = google-beta

  project       = data.google_client_config.this.project
  location      = data.google_client_config.this.region
  repository_id = var.name
  description   = "example docker repository"
  format        = "DOCKER"
  labels        = module.tags.tags
}
