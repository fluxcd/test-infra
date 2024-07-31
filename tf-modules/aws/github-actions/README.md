# AWS GitHub Actions Secrets and Variables

This terraform module creates AWS policy and role to be used in GitHub actions
by assuming the created role with OIDC federation. The GitHub action assumes the
AWS role by authenticating via GitHub OpenID Connect (OIDC) identity provider,
refer [Use IAM roles to connect GitHub Actions to actions in
AWS](https://aws.amazon.com/blogs/security/use-iam-roles-to-connect-github-actions-to-actions-in-aws/).
This can be made easy by using [Configure AWS
Credentials](https://github.com/marketplace/actions/configure-aws-credentials-action-for-github-actions)
GitHub action.

By default, the following GitHub actions secrets are created:
- `AWS_ACCOUNT_ID`
- `AWS_ASSUME_ROLE_NAME`

and `AWS_REGION` actions variable is created. All these names are
overridable, see `variables.tf`.

It also supports adding custom secrets and variables in addition to the above.

**NOTE:** Overwriting existing GitHub secrets and variables is not supported.

## Usage

```hcl
module "aws_gh_actions" {
  source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/aws/github-actions"

  aws_policy_name        = "test-policy-1"
  aws_policy_description = "For running e2e tests"
  aws_provision_perms    = [
    "ec2:CreateInternetGateway",
    "ec2:CreateLaunchTemplate",
    "ec2:CreateLaunchTemplateVersion",
  ]
  aws_cluster_role_prefix    = [
    "flux-test-",
    "blue-eks-node-group-",
    "green-eks-node-group-"
  ]
  aws_role_name          = "test-role-1"
  aws_role_description   = "Role to be assumed by github actions"
  github_repo_owner      = "fluxcd"
  github_project         = "repo-name"
  github_repo_branch_ref = "ref:refs/heads/main"

  github_variable_custom = {
      "SOME_VAR1" = "some-val1",
      "SOME_var2" = "some-val2"
  }
  github_secret_custom = {
      "SECRET1" = "some-secret1",
      "SECRET2" = "some-secret2"
  }
}
```

## AWS Requirements

Use the following IAM policy document to grant the needed permissions.

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "iam:AttachRolePolicy",
                "iam:CreatePolicy",
                "iam:CreatePolicyVersion",
                "iam:CreateRole",
                "iam:DeletePolicy",
                "iam:DeletePolicyVersion",
                "iam:DeleteRole",
                "iam:DetachRolePolicy",
                "iam:GetPolicy",
                "iam:GetPolicyVersion",
                "iam:GetRole",
                "iam:ListAttachedRolePolicies",
                "iam:ListInstanceProfilesForRole",
                "iam:ListPolicyVersions",
                "iam:ListRolePolicies"
            ],
            "Resource": "*"
        }
    ]
}
```

Since the GitHub actions use GitHub OIDC identity provider, the AWS account must
have GitHub as an existing identity provider, see [Configuring OpenID Connect in
Amazon Web
Services](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/configuring-openid-connect-in-amazon-web-services).
The provider URL is expected to be `https://token.actions.githubusercontent.com`
and the audience `sts.amazonaws.com`, as an account can only have a single
instance of this identity provider. These are hard-coded in the configurations
and should be updated in the source, if needed.

## GitHub Requirements

Create a GitHub fine-grained token for the target repository with the following
repository permissions:
- `Read access to metadata`
- `Read and Write access to actions variables and secrets`

## Provider Configuration

Configure the AWS and GitHub provider with the following environment variables:
```sh
export AWS_ACCESS_KEY_ID=""
export AWS_SECRET_ACCESS_KEY=""

export GITHUB_TOKEN=""
```

Check the respective provider docs for more details.
