variable "project_name" {
  description = "The name of the Azure DevOps project"
  type        = string
}

variable "project_description" {
  description = "The description of the Azure DevOps project"
  type        = string
  default     = "Test Project for Flux E2E test - Managed by Terraform"
}

variable "repository_name" {
  description = "The name of the Azure DevOps repository"
  type        = string
}
