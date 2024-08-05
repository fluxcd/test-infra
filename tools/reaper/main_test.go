/*
Copyright 2023 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestParseJSONResources(t *testing.T) {
	tc := []struct {
		name      string
		data      []byte
		wantItems int
	}{
		{
			name: "GCP resources",
			data: []byte(`
[
    {
        "assetType": "artifactregistry.googleapis.com/Repository",
        "createTime": "2023-04-22T20:27:00Z",
        "displayName": "projects/test-gcp/locations/asia-south1/repositories/www",
        "folders": [
            "folders/538536782197",
            "folders/881303735266"
        ],
        "labels": {
            "aaa": "bbb"
        },
        "location": "asia-south1",
        "name": "//artifactregistry.googleapis.com/projects/test-gcp/locations/asia-south1/repositories/www",
        "organization": "organizations/1111111",
        "parentAssetType": "cloudresourcemanager.googleapis.com/Project",
        "parentFullResourceName": "//cloudresourcemanager.googleapis.com/projects/test-gcp",
        "project": "projects/2222222",
        "updateTime": "2023-04-22T20:27:00Z"
    },
    {
        "assetType": "artifactregistry.googleapis.com/Repository",
        "createTime": "2023-04-22T19:10:27Z",
        "displayName": "projects/test-gcp/locations/asia-south1/repositories/qqq",
        "folders": [
            "folders/538536782197",
            "folders/881303735266"
        ],
        "labels": {
            "aaa": "bbb"
        },
        "location": "asia-south1",
        "name": "//artifactregistry.googleapis.com/projects/test-gcp/locations/asia-south1/repositories/qqq",
        "organization": "organizations/1111111",
        "parentAssetType": "cloudresourcemanager.googleapis.com/Project",
        "parentFullResourceName": "//cloudresourcemanager.googleapis.com/projects/test-gcp",
        "project": "projects/2222222",
        "updateTime": "2023-04-22T19:10:27Z"
    }
]
`),
			wantItems: 2,
		},
		{
			name: "Azure resources",
			data: []byte(`
[
    {
        "changedTime": null,
        "createdTime": null,
        "extendedLocation": null,
        "id": "/subscriptions/11111111-1111-1111-1111-1111111111/resourceGroups/test-1/providers/Microsoft.ContainerRegistry/registries/test123zzz",
        "identity": null,
        "kind": "",
        "location": "centralindia",
        "managedBy": "",
        "name": "test123zzz",
        "plan": null,
        "properties": null,
        "provisioningState": null,
        "resourceGroup": "test-1",
        "sku": {
            "capacity": null,
            "family": null,
            "model": null,
            "name": "Basic",
            "size": null,
            "tier": "Basic"
        },
        "tags": {
            "aaa": "bbb"
        },
        "type": "Microsoft.ContainerRegistry/registries"
    },
    {
        "changedTime": null,
        "createdTime": null,
        "extendedLocation": null,
        "id": "/subscriptions/11111111-1111-1111-1111-1111111111/resourceGroups/test-1/providers/Microsoft.KeyVault/vaults/aaa11",
        "identity": null,
        "kind": "",
        "location": "centralindia",
        "managedBy": "",
        "name": "aaa11",
        "plan": null,
        "properties": null,
        "provisioningState": null,
        "resourceGroup": "test-1",
        "sku": null,
        "tags": {
            "aaa": "bbb"
        },
        "type": "Microsoft.KeyVault/vaults"
    }
]
`),
			wantItems: 2,
		},
		{
			name: "Azure resource group",
			data: []byte(`
[
  {
    "id": "/subscriptions/11111111-1111-1111-1111-1111111111/resourceGroups/test-1",
    "location": "centralindia",
    "managedBy": null,
    "name": "test-1",
    "properties": {
      "provisioningState": "Succeeded"
    },
    "tags": {
      "aaa": "bbb",
      "ccc": "ddd"
    },
    "type": "Microsoft.Resources/resourceGroups"
  }
]
`),
			wantItems: 1,
		},
		{
			name: "GCP source repository",
			data: []byte(`
[
  {
    "name": "projects/cncf-flux/repos/fleet-infra-sure-marmot",
    "type": "cloud-source-repository",
    "tags": {
      "createdat": "aaaa"
    }
  }
]
`),
			wantItems: 1,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			rs, err := parseJSONResources(tt.data)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(len(rs)).To(Equal(tt.wantItems))
		})
	}
}

func TestParseTag(t *testing.T) {
	tc := []struct {
		name     string
		tags     string
		wantTags map[string]string
		wantErr  bool
	}{
		{
			name: "single tag",
			tags: "foo1=bar1",
			wantTags: map[string]string{
				"foo1": "bar1",
			},
		},
		{
			name:    "multiple tag",
			tags:    "foo1=bar1,foo2=bar2",
			wantErr: true,
		},
		{
			name:    "no tag",
			tags:    "",
			wantErr: true,
		},
		{
			name:    "no val",
			tags:    "foo1",
			wantErr: true,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			k, v, err := parseTag(tt.tags)
			g.Expect(err != nil).To(Equal(tt.wantErr))
			if err == nil {
				g.Expect(tt.wantTags[k]).To(Equal(v))
			}
		})
	}
}
