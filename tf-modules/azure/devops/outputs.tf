output "repo_url" {
  description = "Azure Devops Git repository HTTPS url"
  value       = azuredevops_git_repository.application.remote_url
}

output "project_id" {
  description = "Azure Devops Project ID"
  value       = azuredevops_project.project.id
}

output "ssh_url" {
  description = "Azure Devops Git repository SSH url"
  value       = azuredevops_git_repository.application.ssh_url
}
