# AKS Module

Configuration in this directory creates an AKS cluster with minimal
configurations that's suitable for flux tests.

## Usage

```hcl
provider "azurerm" {
  features {}
}

resource "random_pet" "suffix" {}

module "aks" {
    source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/azure/aks"

    name = "flux-test-${random_pet.suffix.id}"
    location = "eastus"
} 
```
