# GCR Module

Configuration in this directory creates a Google Artifact Registry Repository
instance and provides a reference to a Google Container Registry that's suitable
for flux tests.

## Usage

```hcl
provider "google" {
  project = "some-project"
  region  = "us-central1"
  zone    = "us-central1-c"
}

resource "random_pet" "suffix" {}

module "gcr" {
    source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/gcp/gcr"

    name = "flux-test-${random_pet.suffix.id}"
}
```
