variable "name" {
  description = "Name of the AKS cluster and associated resources"
  type        = string
}

variable "location" {
  description = "Location where the AKS cluster is created"
  type        = string
  default     = "eastus"
}

variable "tags" {
  description = "Tags for the created resources"
  type        = map(string)
  default     = {}
}

variable "enable_wi" {
  description = "Enable Workload Identity and OIDC Issuer on the cluster"
  type        = bool
  default     = false
}

variable "enable_aad" {
  description = "Enable integration with Microsoft Entra ID"
  type        = bool
  default     = false
}
