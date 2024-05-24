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

package awsnukemod

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/rebuy-de/aws-nuke/v2/cmd"

	"github.com/fluxcd/test-infra/tftestenv"
)

func TestNuke_ApplyRetentionFilter(t *testing.T) {
	// Construct relative time to be used in the test cases.
	now := time.Now().UTC()
	period := "2h"
	beforePeriod := now.Add(-time.Hour * (2 + 1))
	afterPeriod := now.Add(-time.Hour)

	fakeRegion := cmd.Region{Name: "zz"}

	tests := []struct {
		name           string
		inputPeriod    string
		items          cmd.Queue
		wantErr        bool
		wantItemStates []cmd.ItemState
	}{
		{
			name:        "old, new, without createdat, filtered, waiting states",
			inputPeriod: period,
			items: []*cmd.Item{
				{
					Resource: MockResourceWithTags("a1", map[string]string{"tag:" + createdat: beforePeriod.Format(tftestenv.CreatedAtTimeLayout)}),
					State:    cmd.ItemStateNew,
					Region:   &fakeRegion,
				},
				{
					Resource: MockResourceWithTags("a2", map[string]string{"tag:" + createdat: afterPeriod.Format(tftestenv.CreatedAtTimeLayout)}),
					State:    cmd.ItemStateNew,
					Region:   &fakeRegion,
				},
				{
					Resource: MockResource{ARN: "a3"},
					State:    cmd.ItemStateNew,
					Region:   &fakeRegion,
				},
				{
					Resource: MockResourceWithTags("a4", map[string]string{"tag:" + createdat: beforePeriod.Format(tftestenv.CreatedAtTimeLayout)}),
					State:    cmd.ItemStateFiltered,
					Region:   &fakeRegion,
				},
				{
					Resource: MockResourceWithTags("a5", map[string]string{"tag:" + createdat: beforePeriod.Format(tftestenv.CreatedAtTimeLayout)}),
					State:    cmd.ItemStateWaiting,
					Region:   &fakeRegion,
				},
				{
					Resource: MockResourceWithTags("a6", map[string]string{"tag:" + createdat: ""}),
					State:    cmd.ItemStateNew,
					Region:   &fakeRegion,
				},
				{
					Resource: MockResourceWithTags("a7", map[string]string{"tag:role:" + createdat: beforePeriod.Format(tftestenv.CreatedAtTimeLayout)}),
					State:    cmd.ItemStateNew,
					Region:   &fakeRegion,
				},
				{
					Resource: MockResourceWithTags("a8", map[string]string{"tag:igw:" + createdat: beforePeriod.Format(tftestenv.CreatedAtTimeLayout)}),
					State:    cmd.ItemStateNew,
					Region:   &fakeRegion,
				},
			},
			wantItemStates: []cmd.ItemState{
				cmd.ItemStateNew,
				cmd.ItemStateFiltered,
				cmd.ItemStateFiltered,
				cmd.ItemStateFiltered,
				cmd.ItemStateWaiting,
				cmd.ItemStateFiltered,
				cmd.ItemStateNew,
				cmd.ItemStateNew,
			},
		},
		{
			name:        "invalid created at",
			inputPeriod: period,
			items: []*cmd.Item{
				{
					Resource: MockResourceWithTags("a2", map[string]string{"tag:" + createdat: "222222"}),
					State:    cmd.ItemStateNew,
					Region:   &fakeRegion,
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
				items: tt.items,
			}
			err := n.ApplyRetentionFilter(tt.inputPeriod)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Nuke.ApplyRetentionFilter() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				for i, item := range n.Items() {
					g.Expect(item.State).To(Equal(tt.wantItemStates[i]))
				}
			}
		})
	}
}
