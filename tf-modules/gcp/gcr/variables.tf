variable "name" {
  description = "Name of Google Artifact Registry Repository to create"
  type        = string
}

variable "gcr_region" {
  description = "Region of GCR"
  type        = string
  default     = "" // Empty default to use gcr.io.
}
