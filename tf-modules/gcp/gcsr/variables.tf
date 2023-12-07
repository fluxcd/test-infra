variable "name" {
  description = "Name of the Google Cloud Source repository"
  type        = string
}

variable "tags" {
  description = "Tags for the created resources"
  type        = map(string)
  default     = {}
}
