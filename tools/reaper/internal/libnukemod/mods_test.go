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

package libnukemod

import (
	"testing"
	"time"

	"github.com/ekristen/libnuke/pkg/queue"
	. "github.com/onsi/gomega"

	"github.com/fluxcd/test-infra/tftestenv"
)

func TestNuke_ApplyRetentionFilter(t *testing.T) {
	// Construct relative time to be used in the test cases.
	now := time.Now().UTC()
	period := "2h"
	beforePeriod := now.Add(-time.Hour * (2 + 1))
	afterPeriod := now.Add(-time.Hour)

	fakeRegion := "zz"

	tests := []struct {
		name           string
		inputPeriod    string
		items          []*queue.Item
		wantErr        bool
		wantItemStates []queue.ItemState
	}{
		{
			name:        "old, new, without createdat, filtered, waiting states",
			inputPeriod: period,
			items: []*queue.Item{
				{
					Resource: MockResourceWithTags("a1", map[string]string{"tag:" + createdat: beforePeriod.Format(tftestenv.CreatedAtTimeLayout)}),
					State:    queue.ItemStateNew,
					Owner:    fakeRegion,
				},
				{
					Resource: MockResourceWithTags("a2", map[string]string{"tag:" + createdat: afterPeriod.Format(tftestenv.CreatedAtTimeLayout)}),
					State:    queue.ItemStateNew,
					Owner:    fakeRegion,
				},
				{
					Resource: MockResource{ARN: "a3"},
					State:    queue.ItemStateNew,
					Owner:    fakeRegion,
				},
				{
					Resource: MockResourceWithTags("a4", map[string]string{"tag:" + createdat: beforePeriod.Format(tftestenv.CreatedAtTimeLayout)}),
					State:    queue.ItemStateFiltered,
					Owner:    fakeRegion,
				},
				{
					Resource: MockResourceWithTags("a5", map[string]string{"tag:" + createdat: beforePeriod.Format(tftestenv.CreatedAtTimeLayout)}),
					State:    queue.ItemStateWaiting,
					Owner:    fakeRegion,
				},
				{
					Resource: MockResourceWithTags("a6", map[string]string{"tag:" + createdat: ""}),
					State:    queue.ItemStateNew,
					Owner:    fakeRegion,
				},
				{
					Resource: MockResourceWithTags("a7", map[string]string{"tag:role:" + createdat: beforePeriod.Format(tftestenv.CreatedAtTimeLayout)}),
					State:    queue.ItemStateNew,
					Owner:    fakeRegion,
				},
				{
					Resource: MockResourceWithTags("a8", map[string]string{"tag:igw:" + createdat: beforePeriod.Format(tftestenv.CreatedAtTimeLayout)}),
					State:    queue.ItemStateNew,
					Owner:    fakeRegion,
				},
			},
			wantItemStates: []queue.ItemState{
				queue.ItemStateNew,
				queue.ItemStateFiltered,
				queue.ItemStateFiltered,
				queue.ItemStateFiltered,
				queue.ItemStateWaiting,
				queue.ItemStateFiltered,
				queue.ItemStateNew,
				queue.ItemStateNew,
			},
		},
		{
			name:        "invalid created at",
			inputPeriod: period,
			items: []*queue.Item{
				{
					Resource: MockResourceWithTags("a2", map[string]string{"tag:" + createdat: "222222"}),
					State:    queue.ItemStateNew,
					Owner:    fakeRegion,
				},
			},
			wantErr: true,
		},
		{
			name:        "invalid period",
			inputPeriod: "",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			n := &Nuke{
				Queue: &queue.Queue{Items: tt.items},
			}
			err := ApplyRetentionFilter(n, tt.inputPeriod)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ApplyRetentionFilter() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				for i, item := range n.Queue.GetItems() {
					g.Expect(item.State).To(Equal(tt.wantItemStates[i]))
				}
			}
		})
	}
}
