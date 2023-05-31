# ACR Module

Configuration in this directory creates an ACR instance that's suitable for
flux tests. It also takes a principal ID of an AKS cluster and performs role
assignment to allow pulling images from the registry in the cluster nodes. It
expects the resource group to already exist.

## Usage

```hcl
provider "azurerm" {
  features {}
}

resource "random_pet" "suffix" {
  // Since azurerm doesn't allow "-" in registry name, use an alphabet as a
  // separator.
  separator = "o"
}

module "aks" {
    source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/azure/aks"
    ...
}

module "acr" {
    source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/azure/acr"

    name = "fluxTestRepo${random_pet.suffix.id}"
    location = "eastus"
    aks_principal_id = [module.aks.principal_id]
    resource_group = module.aks.resource_group

    depends_on = [module.aks]
}
```
