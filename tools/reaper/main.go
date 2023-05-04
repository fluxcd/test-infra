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
	"errors"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/fluxcd/test-infra/tftestenv"
)

// resource is a common representation of a cloud resource with the minimal
// attributes needed to uniquely identify them.
type resource struct {
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	Location      string            `json:"location"`
	Tags          map[string]string `json:"tags"`
	ResourceGroup string            `json:"resourceGroup"`
}

// awsResource is a representation of AWS resource data obtained by the
// Resource Groups Tagging API. This is used as an intermediate representation
// before converting to resource.
type awsResource struct {
	ResourceARN string
	Tags        []map[string]string
}

// queryGCP returns a GCP command for querying all the resources in a specific
// format compatible with the resource type.
func queryGCP(binPath, jqPath, project, labelKey, labelVal string) string {
	return fmt.Sprintf(`%[1]s asset search-all-resources --project %[3]s --query='labels.%[4]s=%[5]s' --format=json |
			%[2]s '.[] |
			{"name": "\(.displayName)", "type": "\(.assetType)", "location": "\(.location)", "resourceGroup": "%[3]s", "tags": .labels}' |
			%[2]s -s '.'`,
		binPath, jqPath, project, labelKey, labelVal)
}

// queryAzureGroups returns an Azure command for querying all the resource
// groups in a specific format compatible with the resource type.
func queryAzureGroups(binPath, jqPath, tagKey, tagVal string) string {
	return fmt.Sprintf(`%[1]s group list --tag '%[3]s=%[4]s' |
			%[2]s '.[] |
			{name, type, tags, location}' |
			%[2]s -s '.'`,
		binPath, jqPath, tagKey, tagVal)
}

// queryAzureResources returns an Azure command for querying all the resources
// in a specific format compatible with the resource type.
func queryAzureResources(binPath, jqPath, tagKey, tagVal string) string {
	return fmt.Sprintf(`%[1]s resource list --tag '%[3]s=%[4]s' |
			%[2]s '.[] |
			{name, type, tags, location}' |
			%[2]s -s '.'`,
		binPath, jqPath, tagKey, tagVal)
}

// queryAWS returns an AWS command for querying all the resources in a specific
// format compatible with the resource type.
func queryAWS(binPath, jqPath, tagKey, tagVal string) string {
	return fmt.Sprintf(`%[1]s resourcegroupstaggingapi get-resources --tag-filters Key=%[3]s,Values=%[4]s |
			%[2]s '.ResourceTagMappingList'`,
		binPath, jqPath, tagKey, tagVal)
}

var (
	supportedProviders = []string{"aws", "azure", "gcp"}
	targetProvider     = flag.String("provider", "", fmt.Sprintf("name of the provider %v", supportedProviders))
	gcpProject         = flag.String("gcpproject", "", "GCP project name")
	tagKey             = flag.String("tagkey", "", "tag key to query with")
	tagVal             = flag.String("tagval", "", "tag value to query with")
	retentionPeriod    = flag.String("retention-period", "", "period for which the resources should be retained (e.g.: 1d, 1h)")
	jsonoutput         = flag.Bool("ojson", false, "JSON output")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	jqBinPath, err := exec.LookPath("jq")
	if err != nil {
		log.Fatalln(err)
	}

	// Flag validation.
	if *targetProvider == "" {
		log.Fatalf("-provider flag must be set to one of %v", supportedProviders)
	}
	var supported bool
	for _, p := range supportedProviders {
		if p == *targetProvider {
			supported = true
			break
		}
	}
	if !supported {
		log.Fatalf("Unsupported provider %q, must be one of %v", *targetProvider, supportedProviders)
	}

	if *tagKey == "" {
		log.Fatalf("-tagkey flag must be set with tag key")
	}

	if *tagVal == "" {
		log.Fatalf("-tagval flag must be set with tag value")
	}

	// Query resources based on the target provider.
	var resources []resource
	var queryErr error

	switch *targetProvider {
	case "aws":
		path, err := exec.LookPath("aws")
		if err != nil {
			log.Fatalln(err)
		}
		resources, queryErr = getAWSResources(ctx, path, jqBinPath)
	case "azure":
		path, err := exec.LookPath("az")
		if err != nil {
			log.Fatalln(err)
		}
		resources, queryErr = getAzureResources(ctx, path, jqBinPath)
	case "gcp":
		path, err := exec.LookPath("gcloud")
		if err != nil {
			log.Fatalln(err)
		}

		// Unlike other providers, GCP requires a project to be set.
		if *gcpProject == "" {
			log.Println("-gcpproject flag unset. Checking for default gcloud project...")
			p, err := getGCPDefaultProject(ctx, path)
			if err != nil {
				log.Fatalf("Failed looking for default gcloud project: %v", err)
			}
			*gcpProject = p
		}
		resources, queryErr = getGCPResources(ctx, path, jqBinPath)
	}
	if queryErr != nil {
		log.Fatalf("Query error: %v", queryErr)
	}

	// Print only the result to stdout.
	if *retentionPeriod != "" {
		resources, err = applyRetentionFilter(resources, *retentionPeriod)
		if err != nil {
			log.Fatalf("Failed to filter resources with retention-period: %v", err)
		}
	}

	if *jsonoutput {
		out, err := json.MarshalIndent(resources, "", "  ")
		if err != nil {
			log.Fatalf("Failed to JSON marshal result: %v", err)
		}
		fmt.Println(string(out))
	} else {
		fmt.Println("Total resources found:", len(resources))
		for _, r := range resources {
			fmt.Printf("%s: %s\n", r.Type, r.Name)
		}
	}
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

// getAzureResources queries Azure for resources. Azure has two separate APIs
// for listing Resource Groups and all the other resources. Query both and
// combine the result.
func getAzureResources(ctx context.Context, cliPath, jqPath string) ([]resource, error) {
	// Query Resource Groups.
	groupOutput, err := tftestenv.RunCommandWithOutput(ctx, "./",
		queryAzureGroups(cliPath, jqPath, *tagKey, *tagVal),
		tftestenv.RunCommandOptions{},
	)
	if err != nil {
		return nil, err
	}
	groupResources, err := parseJSONResources(groupOutput)
	if err != nil {
		return nil, err
	}

	// Query all the resources.
	resourceOutput, err := tftestenv.RunCommandWithOutput(ctx, "./",
		queryAzureResources(cliPath, jqPath, *tagKey, *tagVal),
		tftestenv.RunCommandOptions{},
	)
	if err != nil {
		return nil, err
	}
	allResources, err := parseJSONResources(resourceOutput)
	if err != nil {
		return nil, err
	}

	return append(groupResources, allResources...), nil
}

// getGCPResources queries GCP for resources.
func getGCPResources(ctx context.Context, cliPath, jqPath string) ([]resource, error) {
	output, err := tftestenv.RunCommandWithOutput(ctx, "./",
		queryGCP(cliPath, jqPath, *gcpProject, *tagKey, *tagVal),
		tftestenv.RunCommandOptions{},
	)
	if err != nil {
		return nil, err
	}
	return parseJSONResources(output)
}

// getGCPDefaultProject queries for the gcloud default/current project.
func getGCPDefaultProject(ctx context.Context, cliPath string) (string, error) {
	// Read only the stdout for valid project value or empty result.
	project, err := tftestenv.RunCommandWithOutput(ctx, "./",
		fmt.Sprintf("%s config get-value project", cliPath),
		tftestenv.RunCommandOptions{StdoutOnly: true},
	)
	if err != nil {
		return "", err
	}
	p := strings.TrimSpace(string(project))
	if p == "" {
		return "", errors.New("no default GCP project found")
	}
	return p, nil
}

// parseJSONResources parses the result of resource query into Resource(s).
func parseJSONResources(r []byte) ([]resource, error) {
	var resources []resource
	if err := json.Unmarshal(r, &resources); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}
	return resources, nil
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
