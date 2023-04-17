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

package tftestenv

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestParseCreatedAtTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "valid",
			input: "x2023-04-22_10h05m15s",
			want:  "2023-04-22 10:05:15",
		},
		{
			name:    "invalid",
			input:   "10h05m15s",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got, want time.Time
			g := NewWithT(t)

			got, err := ParseCreatedAtTime(tt.input)
			g.Expect(err != nil).To(Equal(tt.wantErr))
			if err == nil {
				if tt.want != "" {
					want, err = time.Parse(time.DateTime, tt.want)
					g.Expect(err).ToNot(HaveOccurred())
				}
				g.Expect(got).To(Equal(want))
			}
		})
	}
}
