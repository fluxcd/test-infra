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
  description = "Principal ID of the AKS cluster for role assignment to allow image pull from registry"
  type        = string
}

variable "resource_group" {
  description = "Name of the resource group in which the ACR instance is created"
  type        = string
}
