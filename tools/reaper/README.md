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
- aws-nuke: See below for an AWS IAM policy document.

For deleting the resources, grant the delete permission for the individual
resources.

In GCP, a new deleter role can be created and assigned to the reaper service
account with the following permissions to delete integration test resources:

- `container.operations.get`
- `container.clusters.delete`
- `artifactregistry.repositories.get`
- `artifactregistry.repositories.delete`

For [aws-nuke][aws-nuke], a new deleter IAM policy
can be created and assigned to the reaper IAM principal with the following
policy document:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "reaper",
            "Effect": "Allow",
            "Action": [
                "iam:ListPolicies",
                "iam:ListRoles",
                "iam:ListOpenIDConnectProviders",
                "iam:ListAttachedRolePolicies",
                "iam:ListAccountAliases",
                "iam:ListRolePolicies",
                "iam:GetRole",
                "iam:GetPolicy",
                "iam:GetOpenIDConnectProvider",
                "ec2:DescribeAddresses",
                "ec2:DescribeInstances",
                "ec2:DescribeLaunchTemplates",
                "ec2:DescribeNatGateways",
                "ec2:DescribeRegions",
                "ec2:DescribeSecurityGroups",
                "ec2:DescribeInternetGateways",
                "ec2:DescribeNetworkInterfaces",
                "ec2:DescribeVpcs",
                "ec2:DescribeVolumes",
                "ec2:DescribeSubnets",
                "ec2:DescribeRouteTables",
                "autoscaling:DescribeAutoScalingGroups",
                "eks:ListClusters",
                "eks:ListNodegroups",
                "eks:DescribeCluster",
                "eks:DescribeNodegroup",
                "ecr:ListTagsForResource",
                "ecr:DescribeRepositories",
                "ec2:DeleteSubnet",
                "ec2:DeleteRouteTable",
                "ec2:DeleteVolume",
                "ec2:DeleteTags",
                "ec2:DeleteInternetGateway",
                "ec2:DetachInternetGateway",
                "ec2:RevokeSecurityGroupEgress",
                "ec2:RevokeSecurityGroupIngress",
                "ec2:DeleteSecurityGroup",
                "ec2:DeleteNatGateway",
                "ec2:DeleteVpc",
                "ec2:ReleaseAddress",
                "ec2:DeleteLaunchTemplate",
                "ec2:TerminateInstances",
                "eks:DeleteCluster",
                "eks:DeleteNodegroup",
                "iam:ListPolicyVersions",
                "iam:DeletePolicyVersion",
                "iam:DeletePolicy",
                "iam:DeleteRolePolicy",
                "iam:DetachRolePolicy",
                "iam:DeleteOpenIDConnectProvider",
                "iam:DeleteRole",
                "ecr:DeleteRepository"
            ],
            "Resource": "*"
        }
    ]
}
```

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
2023/11/04 00:49:45 resources found but not deleted
exit status 1
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

**NOTE:** For AWS, unlike the other providers, a third party tool, [aws-nuke][aws-nuke],
is used. The `aws` provider may be removed in the future. It works in a very
limited manner using the Resource Groups Tagging API. The replacement,
`aws-nuke` provider, is capable of listing and deleting the resources properly.

Use the `-h` flag to list all the available options.

[aws-nuke]: https://github.com/ekristen/aws-nuke
