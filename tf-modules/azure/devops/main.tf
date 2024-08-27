resource "azuredevops_project" "project" {
  name               = var.project_name
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
  description        = var.project_description
}

resource "azuredevops_git_repository" "application" {
  project_id     = azuredevops_project.project.id
  name           = var.repository_name
  default_branch = "refs/heads/main"
  initialization {
    init_type = "Clean"
  }
}
