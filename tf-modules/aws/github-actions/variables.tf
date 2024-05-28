variable "aws_account_id" {
  description = "AWS account ID"
  type        = string
  default     = ""
}

variable "aws_region" {
  description = "AWS region used by the tests"
  type        = string
  default     = "us-east-1"
}

variable "aws_policy_name" {
  description = "Name of the policy with all the required permissions"
  type        = string
}

variable "aws_policy_description" {
  description = "IAM policy description"
  type        = string
}

variable "aws_provision_perms" {
  description = "List of permissions for provisioning the infrastructure"
  type        = list(string)
  default     = []
}

variable "aws_cluster_role_prefix" {
  description = "List of name prefixes of the resources that get cluster permission through IAM pass role"
  type        = list(string)
  default     = []
}

variable "aws_role_name" {
  description = "Name of the role that will be assumed by the GitHub actions"
  type        = string
}

variable "aws_role_description" {
  description = "IAM role description"
  type        = string
}

variable "github_repo_owner" {
  description = "Name of the GitHub owner (org or user) of the target repository"
  type        = string
}

variable "github_project" {
  description = "Name of the GitHub project where the actions run, and secrets/variables are added"
  type        = string
}

variable "github_repo_branch_ref" {
  description = "Reference to the target branch in the GitHub repository. Use * for any branch"
  type        = string
  default     = "ref:refs/heads/main"
}

variable "github_secret_accound_id_name" {
  description = "GitHub secret name for AWS accound ID"
  type        = string
  default     = "AWS_ACCOUNT_ID"
}

variable "github_secret_assume_role_name" {
  description = "GitHub secret name for AWS role name to assume"
  type        = string
  default     = "AWS_ASSUME_ROLE_NAME"
}

variable "github_variable_region_name" {
  description = "GitHub variable name for AWS region"
  type        = string
  default     = "AWS_REGION"
}

variable "github_variable_custom" {
  description = "A map of custom GitHub variables to be created"
  type        = map(string)
  default     = {}
}

variable "github_secret_custom" {
  description = "A map of custom GitHub secrets to be created"
  type        = map(string)
  default     = {}
  sensitive   = true
}
