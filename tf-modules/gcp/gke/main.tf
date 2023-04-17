module "tags" {
  source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/utils/tags"

  tags = var.tags
}

data "google_client_config" "this" {}

resource "google_container_cluster" "primary" {
  name               = var.name
  location           = data.google_client_config.this.region
  initial_node_count = 1
  node_config {
    machine_type = "g1-small"
    disk_size_gb = 10

    # Set the scope to grant the nodes all the API access.
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]
  }
  resource_labels = module.tags.tags
}

# auth module to retrieve kubeconfig of the created cluster.
module "gke_auth" {
  source  = "terraform-google-modules/kubernetes-engine/google//modules/auth"
  version = "~> 21"

  project_id   = data.google_client_config.this.project
  cluster_name = var.name
  location     = data.google_client_config.this.region

  depends_on = [google_container_cluster.primary]
}
