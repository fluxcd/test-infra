variable "name" {
  description = "Name of the ACR instance"
  type        = string
}

variable "location" {
  description = "Location where the ACR instance is created"
  type        = string
  default     = "eastus"
}

variable "aks_principal_id" {
  description = "A list of Principal IDs for role assignment to allow image pull from registry"
  type        = list(string)
  default     = []
}

variable "resource_group" {
  description = "Name of the resource group in which the ACR instance is created"
  type        = string
}

variable "admin_enabled" {
  description = "Specifies whether the admin user is enabled in the registry."
  type        = bool
  default     = false
}

variable "tags" {
  description = "Tags for the created resources"
  type        = map(string)
  default     = {}
}
