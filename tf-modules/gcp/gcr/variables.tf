variable "name" {
  description = "Name of Google Artifact Registry Repository to create"
  type        = string
}

variable "tags" {
  description = "Tags for the created resources"
  type        = map(string)
  default     = {}
}
