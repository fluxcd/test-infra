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

func TestParseAWSJSONResources(t *testing.T) {
	data := []byte(`
[
  {
    "ResourceARN": "arn:aws:ecr:us-east-2:1111111111:repository/test11111",
    "Tags": [
      {
        "Key": "aaa",
        "Value": "bbb"
      }
    ]
  },
  {
    "ResourceARN": "arn:aws:ec2:us-east-2:1111111111:security-group/sg-0f055036c4fc16dcd",
    "Tags": [
      {
        "Key": "aaa",
        "Value": "bbb"
      }
    ]
  },
  {
    "ResourceARN": "arn:aws:ec2:us-east-2:1111111111:vpc/vpc-0c596f64034f2014f",
    "Tags": [
      {
        "Key": "aaa",
        "Value": "bbb"
      }
    ]
  }
]
`)

	g := NewWithT(t)

	rs, err := parseAWSJSONResources(data)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(len(rs)).To(Equal(3))
}

func TestAWSResourceToResource(t *testing.T) {
	tc := []struct {
		name     string
		resource awsResource
		wantType string
		wantName string
	}{
		{
			name: "resource with name",
			resource: awsResource{
				ResourceARN: "arn:aws:ecr:us-east-2:1111111111:repository/test-repo-flux-test-31457",
				Tags: []map[string]string{
					{
						"Key":   "foo",
						"Value": "bar",
					},
				},
			},
			wantType: "repository",
			wantName: "test-repo-flux-test-31457",
		},
		{
			name: "resource without name",
			resource: awsResource{
				ResourceARN: "arn:aws:ecr:us-east-2:1111111111:log-group",
				Tags: []map[string]string{
					{
						"Key":   "foo",
						"Value": "bar",
					},
				},
			},
			wantType: "log-group",
			wantName: "NoName",
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			r := awsResourceToResource(tt.resource)
			g.Expect(r.Type).To(Equal(tt.wantType))
			g.Expect(r.Name).To(Equal(tt.wantName))
		})
	}
}
