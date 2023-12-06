variable "devops_org_name" {
  description = "Name of the Azure Devops organization"
  type        = string
}

variable "devops_project_id" {
  description = "ID of the Azure Devops project"
  type = string
}

variable "devops_git_repository" {
  description = "Name of the Azure Devops Git repository"
  type = string
}

variable "pat" {
  description = "Name of the Personal Access Token to access the Azure Devops organization"
  type = string
}

variable "tags" {
  description = "Tags for the created resources"
  type        = map(string)
  default     = {}
}
