# reaper

Find orphan/zombie/stale cloud resources and delete them.

## Requirements

- Cloud provider CLI - `aws`, `az` or `gcloud`
- `jq`

## Permissions

For listing the resources, readonly access to all the resources is needed.
- AWS: Use the builtin `AWSResourceGroupsReadOnlyAccess` IAM policy .
- Azure: Use the builtin `Reader` IAM role.
- GCP: Use the builtin `Cloud Asset Viewer` IAM role.

For deleting the resources, grant the delete permission for the individual
resources.

## Usage

Query the resources by providing the cloud provider name(`provider`) and the
tags key-value pair(`tags`):

```console
$ go run ./ -provider gcp -tags 'environment=dev'
2023/04/28 01:59:36 -gcpproject flag unset. Checking for default gcloud project...
Total resources found: 11
compute.googleapis.com/Instance: gke-flux-test-full-dove-default-pool-dcded0bb-9df3
compute.googleapis.com/Disk: gke-flux-test-full-dove-default-pool-dcded0bb-9df3
compute.googleapis.com/Instance: gke-flux-test-full-dove-default-pool-42290fc9-651q
compute.googleapis.com/Disk: gke-flux-test-full-dove-default-pool-42290fc9-651q
compute.googleapis.com/Instance: gke-flux-test-full-dove-default-pool-f5fadbab-dlxq
compute.googleapis.com/Disk: gke-flux-test-full-dove-default-pool-f5fadbab-dlxq
compute.googleapis.com/InstanceTemplate: gke-flux-test-full-dove-default-pool-f5fadbab
compute.googleapis.com/InstanceTemplate: gke-flux-test-full-dove-default-pool-42290fc9
compute.googleapis.com/InstanceTemplate: gke-flux-test-full-dove-default-pool-dcded0bb
container.googleapis.com/Cluster: flux-test-full-dove
artifactregistry.googleapis.com/Repository: projects/darkowlzz-gcp/locations/us-central1/repositories/flux-test-full-dove
```

JSON output is also supported:

```console
$ go run ./ -provider gcp -tags 'environment=dev' -ojson
```

JSON output contains more detailed information about the resources.

In order to filter the resources by their age, pass the `-retention-period`
flag:

```console
$ go run ./ -provider gcp -tags 'environment=dev' -retention-period 3d
```

The above command would list the resources that are older than 3 days.

In order to delete these resources, pass the `-delete` flag.

**NOTE:** Deleting resources is fully supported in Azure and GCP. Due to the
complexity of deleting the resources created in AWS, it's not implemented yet.
The test infrastructure for AWS involves a lot of individual components that
have to be managed independently, compared of Azure and GCP where resources
related to a cluster are related to one another and can be deleted all together.
If and when the complexity of the AWS test infrastructure is simplified,
deleting the resources can be easily implemented similar to the other providers.
Another issue that contributes to it is the stale resources that are reported
when listing resources via the Resource Groups Tagging API which makes it hard
to find out if a resource still exists or has been deleted without describing
the individual resource and checking their status.

Use the `-h` flag to list all the available options.
