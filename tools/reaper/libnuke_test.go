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

	"github.com/ekristen/libnuke/pkg/queue"
	. "github.com/onsi/gomega"

	"github.com/fluxcd/test-infra/tools/reaper/internal/libnukemod"
)

func Test_libnukeItemsToResources(t *testing.T) {
	fakeRegion1 := "aa"
	fakeRegion2 := "bb"
	fakeResourceType := "Foo"

	tests := []struct {
		name  string
		items []*queue.Item
		want  []resource
	}{
		{
			name: "only converts the items to be deleted",
			items: []*queue.Item{
				{
					Resource: libnukemod.MockResource{ARN: "a1"},
					State:    queue.ItemStateFailed,
					Owner:    fakeRegion1,
					Type:     fakeResourceType,
				},
				{
					Resource: libnukemod.MockResource{ARN: "a2"},
					State:    queue.ItemStateNew,
					Owner:    fakeRegion1,
					Type:     fakeResourceType,
				},
				{
					Resource: libnukemod.MockResourceWithTags("a3", map[string]string{"o": "p"}),
					State:    queue.ItemStateFiltered,
					Owner:    fakeRegion1,
					Type:     fakeResourceType,
				},
				{
					Resource: libnukemod.MockResource{ARN: "a4"},
					State:    queue.ItemStateFinished,
					Owner:    fakeRegion1,
					Type:     fakeResourceType,
				},
				{
					Resource: libnukemod.MockResource{ARN: "a5"},
					State:    queue.ItemStatePending,
					Owner:    fakeRegion1,
					Type:     fakeResourceType,
				},
				{
					Resource: libnukemod.MockResource{ARN: "a6"},
					State:    queue.ItemStatePending,
					Owner:    fakeRegion1,
					Type:     fakeResourceType,
				},
				{
					Resource: libnukemod.MockResourceWithTags("a7", map[string]string{"m": "n"}),
					State:    queue.ItemStateNew,
					Owner:    fakeRegion2,
					Type:     fakeResourceType,
				},
			},
			want: []resource{
				{
					Name:     "a2",
					Type:     fakeResourceType,
					Location: fakeRegion1,
					Tags:     nil,
				},
				{
					Name:     "a7",
					Type:     fakeResourceType,
					Location: fakeRegion2,
					Tags:     map[string]string{"m": "n"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			got := libnukeItemsToResources(tt.items)
			g.Expect(got).To(Equal(tt.want))
		})
	}
}
