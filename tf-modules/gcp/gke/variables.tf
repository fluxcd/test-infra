variable "name" {
  description = "Name of the GKE cluster and associated resources for the cluster"
  type        = string
}

variable "tags" {
  description = "Tags for the created resources"
  type        = map(string)
  default     = {}
}

variable "enable_wi" {
  default     = false
  type        = bool
  description = "enable workload identity for the cluster"
}

variable "oauth_scopes" {
  type = list(string)
  default = [
    "https://www.googleapis.com/auth/cloud-platform"
  ]
}
