module "tags" {
  source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/utils/tags"

  tags = var.tags
}

provider "azuredevops" {
  org_service_url = "https://dev.azure.com/${var.devops_org_name}"
  personal_access_token = var.pat
}

resource "azuredevops_git_repository" "application" {
  project_id = var.devops_project_id
  name       = var.devops_git_repository
  default_branch = "refs/heads/main"
  initialization {
    init_type = "Clean"
  }
}
