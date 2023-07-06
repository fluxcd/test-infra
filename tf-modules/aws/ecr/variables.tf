variable "name" {
  description = "Name of the container registry"
  type        = string
}

variable "region" {
  description = "Region of the container registry"
  default     = ""
  type        = string
}

variable "tags" {
  description = "Tags for the created resources"
  type        = map(string)
  default     = {}
}
