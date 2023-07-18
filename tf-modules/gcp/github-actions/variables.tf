variable "github_project" {
  description = "Name of the GitHub project to add secrets and variables to"
  type        = string
}

variable "gcp_project_id" {
  description = "GCP project to create the resources in"
  type        = string
  default     = ""
}

variable "gcp_region" {
  description = "GCP region"
  type        = string
  default     = ""
}

variable "gcp_zone" {
  description = "GCP zone"
  type        = string
  default     = ""
}

variable "gcp_service_account_id" {
  description = "GCP Service Account ID"
  type        = string
}

variable "gcp_service_account_name" {
  description = "GCP Service Account display name"
  type        = string
}

variable "gcp_roles" {
  description = "List of permissions to grant to the service account"
  type        = list(string)
  default     = []
}

variable "gcp_encoded_key_path" {
  description = "File path of the encoded GCP json key"
  type        = string
  default     = "encoded.txt"
}

variable "gcp_compressed_key_path" {
  description = "File path of the compressed GCP json key"
  type        = string
  default     = "compressed.txt"
}

variable "github_variable_project_id_name" {
  description = "GitHub variable name for GCP project ID"
  type        = string
  default     = "TF_VAR_gcp_project_id"
}

variable "github_variable_region_name" {
  description = "GitHub variable name for GCP region"
  type        = string
  default     = "TF_VAR_gcp_region"
}

variable "github_variable_zone_name" {
  description = "GitHub variable name for GCP zone"
  type        = string
  default     = "TF_VAR_gcp_zone"
}

variable "github_secret_credentials_name" {
  description = "GitHub secret name for Google application credentials"
  type        = string
  default     = "GOOGLE_APPLICATION_CREDENTIALS"
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
