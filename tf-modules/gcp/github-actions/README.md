# Google Cloud Platform GitHub Actions Secrets and Variables

This terraform module registers a GCP service account to be used in CI, grants
the provided IAM permissions to the service account and generates a JSON key for
the service account. The JSON key along with the provider details are then
written to GitHub actions secrets and variables of the given repository.

By default, the following GitHub actions variables are created:
- `TF_VAR_gcp_project_id`
- `TF_VAR_gcp_region`
- `TF_VAR_gcp_zone`

and `GOOGLE_APPLICATION_CREDENTIALS` actions secret is created. All of these
names are overridable, see `variables.tf`.

It also supports adding custom secrets and variables in addition to the above.

**NOTE:** Overwriting existing GitHub secrets and variables are not supported.

## Usage

```hcl
provider "google" {}

module "gcp_gh_actions" {
    source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/gcp/github-actions"

    gcp_service_account_id = "test-sa-1"
    gcp_service_account_name = "test-sa-1"
    gcp_roles = [
        "roles/compute.instanceAdmin.v1",
        "roles/container.admin"
    ]

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

**NOTE:** The JSON key is generated and downloaded locally and is compressed
using jq before writing to GitHub repository. This leaves behind a local copy of
the JSON key. Ensure that they are deleted after the usage. See variable
`gcp_encoded_key_path` and `gcp_compressed_key_path` variables for the file
paths. Since the service account key is only generated the first time, deleting
these files leave no reference of the key and subsequent terraform apply or
destroy would fail without their existence.

## Requirements

This module depends on `base64` and `jq` for decoding and compressing the JSON
key before writing them to GitHub repository. These are required to be
preinstalled on the host machine.

### GCP Requirements

Grant the following IAM permissions:
- Project IAM Admin - `roles/resourcemanager.projectIamAdmin`
- Service Account Admin - `roles/iam.serviceAccountAdmin`
- Service Account Key Admin - `roles/iam.serviceAccountKeyAdmin`

### GitHub Requirements

Create a GitHub fine-grained token for the target repository with the following
repository permissions:
- `Read access to metadata`
- `Read and Write access to actions variables and secrets`

## Provider Configuration

Configure the Google and GitHub provider with the following environment
variables:
```sh
export GOOGLE_PROJECT=""
export GOOGLE_REGION=""
export GOOGLE_ZONE=""

export GITHUB_TOKEN=""
```

Also use `GOOGLE_APPLICATION_CREDENTIALS` when authenticating with a service
account.

Check the respective provider docs for more details.
