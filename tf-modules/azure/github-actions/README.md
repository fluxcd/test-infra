# Azure GitHub Actions Secrets and Variables

Configuration in this directory registers an Azure app to be used in CI,
generates a client secret for the app, creates a custom role with the requested
permissions and assigns the role to the Azure app service account. The app
client details are then written to GitHub actions secrets and variables of a
given repository.

By default, the following GitHub actions secrets are created:
- `ARM_CLIENT_ID`
- `ARM_CLIENT_SECRET`
- `ARM_SUBSCRIPTION_ID`
- `ARM_TENANT_ID`

and `TF_VAR_azure_location` actions variable is created. All of these names
are overridable, see `variables.tf`.

It also supports adding custom secrets and variables in addition to the above.

**NOTE:** Overwriting existing GitHub secrets and variables are not supported.

## Usage

```hcl
module "azure_gh_actions" {
    source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/azure/github-actions"

    azure_owners = ["owner-id-1", "owner-id-2"]
    azure_app_name = "test-app-1"
    azure_app_description = "A test app."
    azure_permissions = [
        "Microsoft.Kubernetes/*",
        "Microsoft.ContainerRegistry/*"
    ]
    azure_location = "eastus"

    github_project = "repo-name"

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

## Azure Requirements

AzureAD permissions:
- To use a Service Account, grant the following Microsoft Graph API permissions:
  - `Application.ReadWrite.OwnedBy`
  - `Directory.Read.All`
- To use a User account, ensure that the user has one of the following
  directory roles:
  - `Application Administrator`
  - `Global Administrator`

Refer [azuread_application docs](https://registry.terraform.io/providers/hashicorp/azuread/latest/docs/resources/application)
and [azuread_service_principal docs](https://registry.terraform.io/providers/hashicorp/azuread/latest/docs/resources/service_principal)
for more details.

AzureRM permissions:
- The following IAM permissions are required when using a Service Account:
  - `Microsoft.Resources/subscriptions/read`
  - `Microsoft.Authorization/roleDefinitions/write`
  - `Microsoft.Authorization/roleDefinitions/delete`
  - `Microsoft.Authorization/roleAssignments/read`
  - `Microsoft.Authorization/roleAssignments/write`
  - `Microsoft.Authorization/roleAssignments/delete`

## GitHub Requirements

Create a GitHub fine-grained token for the target repository with the following
repository permissions:
- `Read access to metadata`
- `Read and Write access to actions variables and secrets`

## Provider Configuration

Configure the AzureRM, AzureAD and Github providers with the following
environment variables:
```sh
export ARM_CLIENT_ID=""
export ARM_CLIENT_SECRET=""
export ARM_SUBSCRIPTION_ID=""
export ARM_TENANT_ID=""

export GITHUB_TOKEN=""
```

**NOTE:** When using Azure user account, logging in using the Azure CLI should
be enough to configure the Azure providers.

Check the respective provider docs for more details.
