# ECR Module

Configuration in this directory creates an ECR instance that's suitable for
flux tests.

## Usage

```hcl
provider "aws" {}

resource "random_pet" "suffix" {}

module "ecr" {
    source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/aws/ecr"

    name = "flux-test-${random_pet.suffix.id}"
}
```
