/*
Copyright 2024 The Flux authors

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
	"github.com/rebuy-de/aws-nuke/v2/cmd"

	"github.com/fluxcd/test-infra/tools/reaper/internal/awsnukemod"
)

func Test_awsnukeItemsToResources(t *testing.T) {
	fakeRegion1 := &cmd.Region{Name: "aa"}
	fakeRegion2 := &cmd.Region{Name: "bb"}
	fakeResourceType := "Foo"

	tests := []struct {
		name  string
		items cmd.Queue
		want  []resource
	}{
		{
			name: "only converts the items to be deleted",
			items: []*cmd.Item{
				{
					Resource: awsnukemod.MockResource{ARN: "a1"},
					State:    cmd.ItemStateFailed,
					Region:   fakeRegion1,
					Type:     fakeResourceType,
				},
				{
					Resource: awsnukemod.MockResource{ARN: "a2"},
					State:    cmd.ItemStateNew,
					Region:   fakeRegion1,
					Type:     fakeResourceType,
				},
				{
					Resource: awsnukemod.MockResourceWithTags("a3", map[string]string{"o": "p"}),
					State:    cmd.ItemStateFiltered,
					Region:   fakeRegion1,
					Type:     fakeResourceType,
				},
				{
					Resource: awsnukemod.MockResource{ARN: "a4"},
					State:    cmd.ItemStateFinished,
					Region:   fakeRegion1,
					Type:     fakeResourceType,
				},
				{
					Resource: awsnukemod.MockResource{ARN: "a5"},
					State:    cmd.ItemStatePending,
					Region:   fakeRegion1,
					Type:     fakeResourceType,
				},
				{
					Resource: awsnukemod.MockResource{ARN: "a6"},
					State:    cmd.ItemStatePending,
					Region:   fakeRegion1,
					Type:     fakeResourceType,
				},
				{
					Resource: awsnukemod.MockResourceWithTags("a7", map[string]string{"m": "n"}),
					State:    cmd.ItemStateNew,
					Region:   fakeRegion2,
					Type:     fakeResourceType,
				},
			},
			want: []resource{
				{
					Name:     "a2",
					Type:     fakeResourceType,
					Location: fakeRegion1.Name,
					Tags:     nil,
				},
				{
					Name:     "a7",
					Type:     fakeResourceType,
					Location: fakeRegion2.Name,
					Tags:     map[string]string{"m": "n"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			got := awsnukeItemsToResources(tt.items)
			g.Expect(got).To(Equal(tt.want))
		})
	}
}
