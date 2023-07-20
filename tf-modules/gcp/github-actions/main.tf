data "google_client_config" "this" {}

locals {
  # Set the provider values from input variables and provider configuration.
  project_id = var.gcp_project_id == "" ? data.google_client_config.this.project : var.gcp_project_id
  region     = var.gcp_region == "" ? data.google_client_config.this.region : var.gcp_region
  zone       = var.gcp_zone == "" ? data.google_client_config.this.zone : var.gcp_zone
}

# Create service account, grant permissions and generate the JSON key.
resource "google_service_account" "service_account" {
  account_id   = var.gcp_service_account_id
  display_name = var.gcp_service_account_name
}

resource "google_project_iam_member" "role_binding" {
  for_each = toset(var.gcp_roles)

  project = local.project_id
  role    = each.key
  member  = "serviceAccount:${google_service_account.service_account.email}"
}

resource "google_service_account_key" "key" {
  service_account_id = google_service_account.service_account.name

  # The JSON key is base64 encoded and prettified. Before using the key in
  # GitHub action, it has to be decoded and compressed. Write the key in a local
  # file and operate on it.
  provisioner "local-exec" {
    command = "echo ${self.private_key} > ${var.gcp_encoded_key_path}"
  }
  provisioner "local-exec" {
    command = "cat encoded.txt | base64 -d | jq -r tostring > ${var.gcp_compressed_key_path}"
  }
}

# Load the compressed JSON key.
data "local_sensitive_file" "compressed" {
  filename = var.gcp_compressed_key_path

  depends_on = [google_service_account_key.key]
}

# Add variables and secrets in github repo.
resource "github_actions_variable" "project_id" {
  repository    = var.github_project
  variable_name = var.github_variable_project_id_name
  value         = local.project_id
}

resource "github_actions_variable" "region" {
  repository    = var.github_project
  variable_name = var.github_variable_region_name
  value         = local.region
}

resource "github_actions_variable" "zone" {
  repository    = var.github_project
  variable_name = var.github_variable_zone_name
  value         = local.zone
}

resource "github_actions_secret" "credentials" {
  repository      = var.github_project
  secret_name     = var.github_secret_credentials_name
  plaintext_value = data.local_sensitive_file.compressed.content
}

resource "github_actions_variable" "custom" {
  for_each = var.github_variable_custom

  repository    = var.github_project
  variable_name = each.key
  value         = each.value
}

resource "github_actions_secret" "custom" {
  # Mark only the key as nonsensitive.
  for_each = nonsensitive(toset(keys(var.github_secret_custom)))

  repository      = var.github_project
  secret_name     = each.key
  plaintext_value = var.github_secret_custom[each.key]
}
