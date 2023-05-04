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
	"time"

	. "github.com/onsi/gomega"

	"github.com/fluxcd/test-infra/tftestenv"
)

func TestApplyRetentionFilter(t *testing.T) {
	// Construct relative time to be used in the test cases.
	now := time.Now().UTC()
	period := "2h"
	beforePeriod := now.Add(-time.Hour * (2 + 1))
	afterPeriod := now.Add(-time.Hour)

	tc := []struct {
		name              string
		inputPeriod       string
		resources         []resource
		wantResourceCount int
		wantErr           bool
	}{
		{
			name:        "old, new, without createdat",
			inputPeriod: period,
			resources: []resource{
				{
					Name: "foo1",
					Tags: map[string]string{createdat: beforePeriod.Format(tftestenv.CreatedAtTimeLayout)},
				},
				{
					Name: "foo2",
					Tags: map[string]string{createdat: beforePeriod.Format(tftestenv.CreatedAtTimeLayout)},
				},
				{
					Name: "foo3",
					Tags: map[string]string{createdat: afterPeriod.Format(tftestenv.CreatedAtTimeLayout)},
				},
				{
					Name: "foo3",
					Tags: map[string]string{"foo": "bar"},
				},
			},
			wantResourceCount: 2,
		},
		{
			name:        "invalid createdat",
			inputPeriod: period,
			resources: []resource{
				{
					Name: "foo1",
					Tags: map[string]string{createdat: "22222"},
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

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			result, err := applyRetentionFilter(tt.resources, tt.inputPeriod)
			g.Expect(err != nil).To(Equal(tt.wantErr))
			if err == nil {
				g.Expect(len(result)).To(Equal(tt.wantResourceCount))
			}
		})
	}
}
