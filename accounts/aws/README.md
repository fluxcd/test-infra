# AWS Account

This documents how to set up an AWS account, prepare it to use for Flux test
infrastructure and various usage workflows for managing the account.

## New account initial setup

- Once a new AWS account is created, log in as the root user and enable
  multi-factor authentication for the root account, refer
  https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_mfa_enable.html.
- For Billing and Cost Management in the account, enable IAM access to billing,
  refer https://docs.aws.amazon.com/IAM/latest/UserGuide/tutorial_billing.html.
  This will enable the other users to be able to view the billing console if
  they have the necessary permissions. With access to billing console, the users
  would be able to know the cost of their resource usage and help keep the cost
  in control.
- For user management IAM Identity Center is used, which makes it easy to invite
  and manage users in the account. AWS sends invitation email to the users and
  provides an access portal to verify the email address, set up MFA device and
  help log into the account easily. Choose a region, say `us-east-2`, switch to
  the region in the AWS web console and enable IAM Identity Center, refer
  https://docs.aws.amazon.com/SetUp/latest/UserGuide/setup-enableIdC.html. If
  asked, create an AWS Organization and enable the Identity Center as an
  organization, which is the recommended usage by AWS.
  - After enabling IAM Identity Center, go to the IAM Identity Center console
    settings and enable Multi-factor authentication, refer
    https://docs.aws.amazon.com/singlesignon/latest/userguide/mfa-getting-started.html.
    Configure the following options:
      - Under **Prompt users for MFA**, select *Only when the sign-in context
        changes*.
      - Under **Users can authenticate with these MFA types**, select both
        *Security keys* and *Authenticator apps*.
      - Under **If a user does not yet have a registered MFA device**, select
        *Require them to register an MFA device at sign in*.
  - Under the **Authentication** tab in IAM Identity Center settings, configure
    the **Standard authentication** to *Send email OTP for users created from
    API*. This will make sure invitation emails are sent to the users when
    created using terraform or other tooling.
- Some tools like aws-nuke require the AWS account to have an alias set before
  operating on the account. Set an account alias in the IAM Dashboard, refer
  https://docs.aws.amazon.com/IAM/latest/UserGuide/console_account-alias.html.
  It can be set to `fluxcd` or edited if needed in the future.

The above covers the initial setup. Further account setup will be done by code
as described in the following sections.

## Account management

After the initial setup, the account can be managed using terraform
configurations for provisioning and maintaining all the resources.

`main.tf` contains terraform configuration for creating IAM Identity Center
permission sets, groups using the permission sets, their association with the
AWS account, users for web console access, IAM Identity providers which are used
in the tests for authenticating with federated identities and assuming roles
with permissions needed for running the tests.

For first time setup, an IAM user can be created manually with the administrator
access to apply the configurations in `main.tf`. This will create the user
accounts who can log in and use the account.

**NOTE:** Due to a limitation in the AWS Identity Center API, the user accounts
created via API require explicit email verification. Refer
https://github.com/hashicorp/terraform-provider-aws/issues/28102 for details.
Due to this, after creating a new user, an administrator needs to go to the
user's page and click on **Send email verification link** button.

After applying the configuration, the IAM user can be deleted and the created
non-root users can be used to manage the account. Also see
https://docs.aws.amazon.com/signin/latest/userguide/iam-id-center-sign-in-tutorial.html
for details about AWS access portal sign in.

The account can be managed by updating the terraform code and regularly applying
the changes using terraform in a GitHub actions workflow. Updates to users and
resources in the account can be go through the usual GitHub pull request
workflow.
