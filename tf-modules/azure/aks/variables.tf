variable "name" {
  description = "Name of the AKS cluster and associated resources"
  type        = string
}

variable "location" {
  description = "Location where the AKS cluster is created"
  type        = string
  default     = "eastus"
}
