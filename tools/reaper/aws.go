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
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fluxcd/test-infra/tftestenv"
)

// awsResource is a representation of AWS resource data obtained by the
// Resource Groups Tagging API. This is used as an intermediate representation
// before converting to resource.
type awsResource struct {
	ResourceARN string
	Tags        []map[string]string
}

// queryAWS returns an AWS command for querying all the resources in a specific
// format compatible with the resource type.
func queryAWS(binPath, jqPath, tagKey, tagVal string) string {
	return fmt.Sprintf(`%[1]s resourcegroupstaggingapi get-resources --tag-filters Key=%[3]s,Values=%[4]s |
			%[2]s '.ResourceTagMappingList'`,
		binPath, jqPath, tagKey, tagVal)
}

// getAWSResources queries AWS for resources.
func getAWSResources(ctx context.Context, cliPath, jqPath string) ([]resource, error) {
	output, err := tftestenv.RunCommandWithOutput(ctx, "./",
		queryAWS(cliPath, jqPath, *tagKey, *tagVal),
		tftestenv.RunCommandOptions{},
	)
	if err != nil {
		return nil, err
	}
	return parseAWSJSONResources(output)
}

// parseAWSJSONResources parses the result of resource query into Resource(s).
func parseAWSJSONResources(r []byte) ([]resource, error) {
	// Convert to AWSResources.
	var awsResources []awsResource
	if err := json.Unmarshal(r, &awsResources); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	resources := []resource{}
	for _, ar := range awsResources {
		resources = append(resources, awsResourceToResource(ar))
	}

	return resources, nil
}

// awsResourceToResource converts an AWSResource into a Resource.
func awsResourceToResource(r awsResource) resource {
	// Extract information from the ARN.
	parts := strings.Split(r.ResourceARN, ":")
	rName, rType := parseAWSResourceNameAndType(parts[5])
	rLocation := parts[3]
	accountID := parts[4]

	rTags := map[string]string{}
	for _, t := range r.Tags {
		rTags[t["Key"]] = t["Value"]
	}

	return resource{
		Name:          rName,
		Type:          rType,
		Location:      rLocation,
		Tags:          rTags,
		ResourceGroup: accountID,
	}
}

// parseAWSResourceNameAndType separates resource name and type from the given
// slice of an ARN. For example, for ARN
// "arn:aws:ecr:us-east-2:1111111111:repository/test-repo-flux-test-31457", the
// last slice "repository/test-repo-flux-test-31457" is the input which results
// in "test-repo-flux-test-31457" as the name and "repository" as the type.
func parseAWSResourceNameAndType(name string) (rName, rType string) {
	parts := strings.SplitN(name, "/", 2)
	if len(parts) < 2 {
		// Some resources like "log-group" don't have a name. Use a placeholder
		// name for such resources.
		rType = parts[0]
		rName = "NoName"
	} else {
		rType = parts[0]
		rName = parts[1]
	}

	return rName, rType
}
