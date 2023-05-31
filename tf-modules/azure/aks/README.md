# AKS Module

Configuration in this directory creates an AKS cluster with the minimal
configurations that's suitable for flux tests.

__NOTE:__ When enabling Workload Identity, please ensure that the current
subscription has the `EnableWorkloadIdentityPreview` feature flag registered.
For more information, please see [the docs](https://learn.microsoft.com/en-us/azure/aks/workload-identity-deploy-cluster#register-the-enableworkloadidentitypreview-feature-flag).

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
