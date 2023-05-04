# reaper

Find orphan/zombie/stale cloud resources and delete them.

## Requirements

- Cloud provider CLI - `aws`, `az` or `gcloud`
- `jq`

## Usage

Query the resources by providing the cloud provider name(`provider`) and the tag key(`tagkey`) and value(`tagval`):

```console
$ go run ./ -provider gcp -tagkey environment -tagval dev
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
$ go run ./ -provider gcp -tagkey environment -tagval dev -ojson
```

**TODO:**
- Filtering of resources based on their age.
- Delete the resources.
