# GKE Module

Configuration in this directory creates an GKE cluster with minimal
configurations that's suitable for flux tests. An OAuth scope is set in the
cluster to provide GCP API access to all the nodes, allowing them to pull images
from container registries.

## Usage

```hcl
provider "google" {
  project = "some-project"
  region  = "us-central1"
  zone    = "us-central1-c"
}

resource "random_pet" "suffix" {}

module "gke" {
    source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/gcp/gke"

    name = "flux-test-${random_pet.suffix.id}"
}
```

**:warning: WARNING:** This module ignores changes to the
`google_container_cluster.node_config` using [lifecycle
meta-argument](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#ignore_changes).
Please consider updating the lifecycle arguments when there's a need to update
the `node_config` post provision.
