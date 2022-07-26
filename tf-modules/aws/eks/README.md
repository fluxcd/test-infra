# EKS Module

Configuration in this directory creates an EKS cluster with Managed Node Group
with minimal configurations that's suitable for flux tests.

## Usage

```hcl
provider "aws" {}

resource "random_pet" "suffix" {}

module "eks" {
    source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/aws/eks"

    name = "flux-test-${random_pet.suffix.id}"
}
```
