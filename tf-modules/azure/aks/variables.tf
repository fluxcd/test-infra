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

variable "aad_rbac_tenant_id" {
  description = "Tenant ID for Azure Active Directory RBAC"
  type        = string
  default     = ""
}
