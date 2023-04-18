variable "name" {
  description = "Name of the EKS cluster and associated resources for the cluster"
  type        = string
}

variable "tags" {
  description = "Tags for the created resources"
  type        = map(string)
  default     = {}
}
