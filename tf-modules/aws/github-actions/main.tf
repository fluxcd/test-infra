data "aws_caller_identity" "current" {}

locals {
  # Set the provider values from input variables and provider configuration.
  account_id = var.aws_account_id == "" ? data.aws_caller_identity.current.account_id : var.aws_account_id

  # Construct a list of role ARN from the given role prefixes to use in cluster
  # permissions below.
  clusterperms_resources = [for prefix in var.aws_cluster_role_prefix : "arn:aws:iam::${local.account_id}:role/${prefix}*"]
}

data "aws_iam_policy_document" "policy_doc" {
  # Permissions for provisioning the infrastructure.
  statement {
    sid       = "testinfra"
    actions   = var.aws_provision_perms
    resources = ["*"]
  }

  # Pass cluster permissions to the following roles: cluster, node-groups, etc.
  statement {
    sid       = "clusterperms"
    actions   = ["iam:PassRole"]
    resources = local.clusterperms_resources
  }
}

resource "aws_iam_policy" "policy" {
  name        = var.aws_policy_name
  policy      = data.aws_iam_policy_document.policy_doc.json
  description = var.aws_policy_description
}

# Create assume role policy document for defining the trust relationship with
# GitHub OIDC and the target repository.
data "aws_iam_policy_document" "assume_role_doc" {
  statement {
    # Create trusted identity of type Web identity with github as the provider.
    actions = ["sts:AssumeRoleWithWebIdentity"]
    principals {
      type        = "Federated"
      identifiers = ["arn:aws:iam::${local.account_id}:oidc-provider/token.actions.githubusercontent.com"]
    }
    # Set the audience to STS.
    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = ["sts.amazonaws.com"]
    }
    # Set the GitHub repository.
    condition {
      test     = "StringLike"
      variable = "token.actions.githubusercontent.com:sub"
      values   = ["repo:${var.github_repo_owner}/${var.github_project}:${var.github_repo_branch_ref}"]
    }
  }
}

# Create a role to assume by github actions.
resource "aws_iam_role" "github_actions_role" {
  name               = var.aws_role_name
  assume_role_policy = data.aws_iam_policy_document.assume_role_doc.json
  description        = var.aws_role_description
}

# Attach the policy to the role.
resource "aws_iam_role_policy_attachment" "github_actions_role_attachment" {
  role       = aws_iam_role.github_actions_role.name
  policy_arn = aws_iam_policy.policy.arn
}

# Add a GitHub secret variable for the AWS account ID.
resource "github_actions_secret" "account_id" {
  repository      = var.github_project
  secret_name     = var.github_secret_accound_id_name
  plaintext_value = local.account_id
}

# Add a GitHub secret for the AWS role name to assume.
resource "github_actions_secret" "role_name" {
  repository      = var.github_project
  secret_name     = var.github_secret_assume_role_name
  plaintext_value = var.aws_role_name
}

# Add a GitHub variable for the AWS region.
resource "github_actions_variable" "region" {
  repository    = var.github_project
  variable_name = var.github_variable_region_name
  value         = var.aws_region
}

resource "github_actions_variable" "custom" {
  for_each = var.github_variable_custom

  repository    = var.github_project
  variable_name = each.key
  value         = each.value
}

resource "github_actions_secret" "custom" {
  # Mark only the key as nonsensitive.
  for_each = nonsensitive(toset(keys(var.github_secret_custom)))

  repository      = var.github_project
  secret_name     = each.key
  plaintext_value = var.github_secret_custom[each.key]
}
