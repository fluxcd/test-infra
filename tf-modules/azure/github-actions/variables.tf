variable "azure_owners" {
  description = "Object IDs of the Azure resource owners"
  type        = list(string)
  default     = []
}

variable "github_project" {
  description = "Name of the GitHub project to add secrets to"
  type        = string
}

variable "azure_app_name" {
  description = "Name of the Azure app to register"
  type        = string
}

variable "azure_app_description" {
  description = "Description of the Azure app, will be used to provide context to the role assigned to the app"
  type        = string
}

variable "azure_app_secret_name" {
  description = "Name of the Azure app secret"
  type        = string
  default     = "tf-generated-secret"
}

variable "azure_permissions" {
  description = "List of permissions to grant to the created app"
  type        = list(string)
  default     = []
}

variable "github_secret_client_id_name" {
  description = "GitHub secret name for Azure app client ID"
  type        = string
  default     = "ARM_CLIENT_ID"
}

variable "github_secret_client_secret_name" {
  description = "GitHub secret name for Azure app client secret"
  type        = string
  default     = "ARM_CLIENT_SECRET"
}

variable "github_secret_subscription_id_name" {
  description = "GitHub secret name for Azure subscription ID"
  type        = string
  default     = "ARM_SUBSCRIPTION_ID"
}

variable "github_secret_tenant_id_name" {
  description = "GitHub secret name for Azure tenant ID"
  type        = string
  default     = "ARM_TENANT_ID"
}

variable "github_variable_location_name" {
  description = "GitHub variable name for Azure location"
  type        = string
  default     = "TF_VAR_azure_location"
}

variable "azure_location" {
  description = "Azure location used by the tests"
  type        = string
  default     = "eastus"
}

variable "github_variable_custom" {
  description = "A map of custom GitHub variables to be created"
  type        = map(string)
  default     = {}
}

variable "github_secret_custom" {
  description = "A map of custom GitHub secrets to be created"
  type        = map(string)
  default     = {}
  sensitive   = true
}
