data "aws_caller_identity" "current" {}
data "aws_ssoadmin_instances" "current" {}

# Create a permission set for administrator access.
resource "aws_ssoadmin_permission_set" "admin" {
  name             = "AdministratorAccess"
  instance_arn     = tolist(data.aws_ssoadmin_instances.current.arns)[0]
  description      = "To be used to grant administrator access to users and groups."
  session_duration = "PT8H"
  # TODO: Decide and add tags.
}

# Create a group for administrators.
resource "aws_identitystore_group" "admin" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.current.identity_store_ids)[0]
  display_name      = "Admin"
  description       = "Admin Group"
}

# Assign the admin group and permission set.
resource "aws_ssoadmin_account_assignment" "admin_account_assignment" {
  instance_arn       = tolist(data.aws_ssoadmin_instances.current.arns)[0]
  permission_set_arn = aws_ssoadmin_permission_set.admin.arn

  principal_id   = aws_identitystore_group.admin.group_id
  principal_type = "GROUP"

  target_id   = data.aws_caller_identity.current.account_id
  target_type = "AWS_ACCOUNT"
}

# Attach a AdministratorAccess managed policy to the administrator permission
# set.
# NOTE: Since this attachment affects accounts the permission set is associated
# with, it has to depends on the account assignment.
resource "aws_ssoadmin_managed_policy_attachment" "admin" {
  depends_on = [aws_ssoadmin_account_assignment.admin_account_assignment]

  instance_arn       = tolist(data.aws_ssoadmin_instances.current.arns)[0]
  managed_policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess"
  permission_set_arn = aws_ssoadmin_permission_set.admin.arn
}

# Create user and assign a group.
resource "aws_identitystore_user" "darkowlzz" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.current.identity_store_ids)[0]

  display_name = "Sunny"
  user_name    = "darkowlzz"

  name {
    given_name  = "Sunny"
    family_name = "Sunny"
  }

  emails {
    value   = "strainer_exception937@simplelogin.com"
    primary = true
  }
}
resource "aws_identitystore_group_membership" "darkowlzz_admin" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.current.identity_store_ids)[0]
  group_id          = aws_identitystore_group.admin.group_id
  member_id         = aws_identitystore_user.darkowlzz.user_id
}

# Register GitHub OIDC identity provider.
resource "aws_iam_openid_connect_provider" "github" {
  url = "https://token.actions.githubusercontent.com"

  client_id_list = [
    "sts.amazonaws.com",
  ]

  # For obtaining the thumbprint, refer
  # https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc_verify-thumbprint.html.
  # Another easier way to obtain this is from the AWS IAM Identity Provider web
  # console. When a provider is added through the web console, the thumbprint is
  # optional. AWS automatically obtains it and shows it in the console.
  thumbprint_list = ["1b511abead59c6ce207077c0bf0e0043b1382612"]

  # TODO: Decide and add tags.
}
