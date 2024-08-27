# DevOps Module

Configuration in this directory creates an Azure DevOps Project and repository.

## Usage

Legacy shared modules with their own provider configurations are not compatible
with new features like for_each, count and depends_on as described
[here](https://developer.hashicorp.com/terraform/language/modules/develop/providers#legacy-shared-modules-with-provider-configurations).
To use these features by passing provider configuration to the legacy module,
create version.tf file with the following content -

```hcl
terraform {
  required_providers {
    azuredevops = {
      source = "microsoft/azuredevops"
    }
  }
}
```

In main.tf, create the provider configuration and pass it to the devops module.

```hcl
provider "azuredevops" {
  org_service_url       = "https://dev.azure.com/azuredevops_org"
  personal_access_token = "azuredevops_pat"
}

module "devops" {
  source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/azure/devops"
  providers = {
    azuredevops = azuredevops
  }

  project_name = local.project_name
  repository_name = local.repo_name
}
```
