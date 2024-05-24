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

import "github.com/rebuy-de/aws-nuke/v2/pkg/types"

type MockResource struct {
	ARN         string
	Tags        types.Properties
	RemoveError error
}

func (mr MockResource) Remove() error                { return mr.RemoveError }
func (mr MockResource) String() string               { return mr.ARN }
func (mr MockResource) Properties() types.Properties { return mr.Tags }

func MockResourceWithTags(arn string, props map[string]string) MockResource {
	return MockResource{
		ARN:  arn,
		Tags: types.Properties(props),
	}
}
