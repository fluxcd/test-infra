module "tags" {
  source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/utils/tags"

  tags = var.tags
}

data "google_client_config" "this" {}

data "google_project" "this" {}

resource "google_container_cluster" "primary" {
  name               = var.name
  location           = data.google_client_config.this.region
  initial_node_count = 1
  node_config {
    machine_type = "g1-small"
    disk_size_gb = 10

    # Set the scope to grant the nodes all the API access.
    oauth_scopes = var.oauth_scopes

  }

  workload_identity_config {
    workload_pool = var.enable_wi == false ? null : "${data.google_project.this.project_id}.svc.id.goog"
  }

  resource_labels = module.tags.tags

  lifecycle {
    ignore_changes = [
      # When enabling workload identity, the oauth_scopes in node_config is set
      # to be empty at provision time. But after provision, some default scopes
      # get attached. Reapplying the same config tries to remove the existing
      # oauth scopes for empty scopes. Oauth scopes replacement requires
      # destroying and creating a new cluster. To prevent such cluster
      # recreation, ignore any change in node_config.
      # node_config[0].oauth_scopes can ignore just oauth_scopes of the first
      # node_config, but for simplicity, ignore the whole node_config change.
      node_config
    ]
  }
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
