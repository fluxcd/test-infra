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
	"time"

	"github.com/ekristen/libnuke/pkg/queue"

	"github.com/fluxcd/test-infra/tools/reaper/internal/libnukemod"
)

const (
	aws     = "aws"
	azure   = "azure"
	gcp     = "gcp"
	awsnuke = "aws-nuke"
)

// registryTypes maps the registry type resources with their resource.Type value
// in different providers. This is used to identify that a given resource is a
// registry in a particular provider.
var registryTypes map[string]string = map[string]string{
	aws:   "repository",
	azure: "Microsoft.ContainerRegistry/registries",
	gcp:   "artifactregistry.googleapis.com/Repository",
}

// clusterTypes maps the cluster type resource with their resource.Type value in
// different providers. This is used to identify that a given resource is a
// cluster in a particular provider.
var clusterTypes map[string]string = map[string]string{
	aws:   "cluster",
	azure: "Microsoft.ContainerService/managedClusters",
	gcp:   "container.googleapis.com/Cluster",
}

// sourceRepoTypes maps the source repository type resource with their
// resource.Type value in different providers. This is used to identify that a
// given resource is a source repository in a particular provider.
var sourceRepoTypes map[string]string = map[string]string{
	aws:   "",
	azure: "",
	gcp:   "cloud-source-repository",
}

// resource is a common representation of a cloud resource with the minimal
// attributes needed to uniquely identify them.
type resource struct {
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	Location      string            `json:"location"`
	Tags          map[string]string `json:"tags"`
	ResourceGroup string            `json:"resourceGroup"`
}

var (
	supportedProviders = []string{aws, azure, gcp, awsnuke}
	targetProvider     = flag.String("provider", "", fmt.Sprintf("name of the provider %v", supportedProviders))
	gcpProject         = flag.String("gcpproject", "", "GCP project name")
	awsRegions         = flag.String("awsregions", "", "Comma separated list of aws regions for aws-nuke (e.g.: us-east-1,us-east-2). The first entry is used as the default region")
	tags               = flag.String("tags", "", "key-value pair of tag to query with. Only single pair supported at present ('environment=dev')")
	retentionPeriod    = flag.String("retention-period", "", "period for which the resources should be retained (e.g.: 1d, 1h)")
	jsonoutput         = flag.Bool("ojson", false, "JSON output")
	delete             = flag.Bool("delete", false, "delete the resources")
	timeout            = flag.String("timeout", "15m", "timeout")

	// TODO: When adding multiple tags support, get rid of these global tag
	// values and pass the tags between the functions. Implement a cloud
	// provider specific query builder based on a give map of tags.
	tagKey string
	tagVal string
)

func main() {
	flag.Parse()

	// Paths of the cloud provider CLI binaries.
	var awsPath, azPath, gcloudPath string

	var awsNuker *libnukemod.Nuke

	t, err := time.ParseDuration(*timeout)
	if err != nil {
		log.Fatalf("Failed parsing timeout: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()

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

	tagKey, tagVal, err = parseTag(*tags)
	if err != nil {
		log.Fatalf("Failed parsing tags: %v", err)
	}

	// Query resources based on the target provider.
	var resources []resource
	var queryErr error

	switch *targetProvider {
	case aws:
		awsPath, err = exec.LookPath("aws")
		if err != nil {
			log.Fatalln(err)
		}
		resources, queryErr = getAWSResources(ctx, awsPath, jqBinPath)
	case azure:
		azPath, err = exec.LookPath("az")
		if err != nil {
			log.Fatalln(err)
		}
		resources, queryErr = getAzureResources(ctx, azPath, jqBinPath)
	case gcp:
		gcloudPath, err = exec.LookPath("gcloud")
		if err != nil {
			log.Fatalln(err)
		}

		// Unlike other providers, GCP requires a project to be set.
		if *gcpProject == "" {
			log.Println("-gcpproject flag unset. Checking for default gcloud project...")
			p, err := getGCPDefaultProject(ctx, gcloudPath)
			if err != nil {
				log.Fatalf("Failed looking for default gcloud project: %v", err)
			}
			*gcpProject = p
		}
		resources, queryErr = getGCPResources(ctx, gcloudPath, jqBinPath)
	case awsnuke:
		// Get the account ID of the IAM principal using AWS CLI. Since aws-nuke
		// can work on multiple accounts, it explicitly needs the target account
		// ID.
		awsPath, err = exec.LookPath("aws")
		if err != nil {
			log.Fatalln(err)
		}
		awsAccountID, err := getAWSAccountID(ctx, awsPath)
		if err != nil {
			log.Fatalln(err)
		}
		if *awsRegions == "" {
			log.Fatalf("-awsregions flag unset. AWS regions must be set for aws-nuke")
		}

		// Query aws resources using aws-nuke.
		awsNuker, queryErr = libnukeAWSScan(ctx, awsAccountID)
		if queryErr == nil {
			resources = libnukeItemsToResources(awsNuker.Queue.GetItems())
		}
	}
	if queryErr != nil {
		log.Fatalf("Query error: %v", queryErr)
	}

	// Apply the retention period filter.
	if *retentionPeriod != "" && len(resources) > 0 {
		switch *targetProvider {
		case awsnuke:
			if err := libnukemod.ApplyRetentionFilter(awsNuker, *retentionPeriod); err != nil {
				log.Fatalf("Failed to filter resources with retention-period: %v", err)
			}
			// Update the resources if the number of items to be removed has
			// changed.
			if awsNuker.Queue.Count(queue.ItemStateNew) != len(resources) {
				resources = libnukeItemsToResources(awsNuker.Queue.GetItems())
			}
		default:
			resources, err = applyRetentionFilter(resources, *retentionPeriod)
			if err != nil {
				log.Fatalf("Failed to filter resources with retention-period: %v", err)
			}
		}
	}

	// Print only the result to stdout.
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

	// Delete the resources.
	if *delete && len(resources) > 0 {
		log.Printf("Deleting resources...")

		switch *targetProvider {
		case aws:
			log.Println("Unimplemented for provider AWS.")
		case azure:
			groups := getAzureResourceGroups(resources)
			for _, group := range groups {
				if err := deleteAzureResourceGroup(ctx, azPath, group); err != nil {
					log.Fatalf("Failed to delete resource group: %v", err)
				}
			}
		case gcp:
			registries := getRegistries(*targetProvider, resources)
			for _, registry := range registries {
				if err := deleteGCPArtifactRepository(ctx, gcloudPath, registry); err != nil {
					log.Fatalf("Failed to delete registries: %v", err)
				}
			}

			clusters := getClusters(*targetProvider, resources)
			for _, cluster := range clusters {
				if err := deleteGCPCluster(ctx, gcloudPath, cluster); err != nil {
					log.Fatalf("Failed to delete cluster: %v", err)
				}
			}

			srcRepos := getSourceRepos(*targetProvider, resources)
			for _, repo := range srcRepos {
				if err := deleteGCPSourceRepo(ctx, gcloudPath, repo); err != nil {
					log.Fatalf("Failed to delete source repository: %v", err)
				}
			}
		case awsnuke:
			if err := awsNuker.Delete(ctx); err != nil {
				log.Fatalf("Failed to delete resources: %v", err)
			}
		}
	} else if !*delete && len(resources) > 0 {
		// Exit with non-zero exit code when resources are found but not
		// deleted. This is to help detect stale resources when run in CI by
		// failing the job.
		log.Fatal("resources found but not deleted")
	}
}

func getRegistries(provider string, resources []resource) []resource {
	result := []resource{}
	for _, r := range resources {
		if r.Type == registryTypes[provider] {
			result = append(result, r)
		}
	}
	return result
}

func getClusters(provider string, resources []resource) []resource {
	result := []resource{}
	for _, r := range resources {
		if r.Type == clusterTypes[provider] {
			result = append(result, r)
		}
	}
	return result
}

func getSourceRepos(provider string, resources []resource) []resource {
	result := []resource{}
	for _, r := range resources {
		if r.Type == sourceRepoTypes[provider] {
			result = append(result, r)
		}
	}
	return result
}

func getAzureResourceGroups(resources []resource) []resource {
	result := []resource{}
	for _, r := range resources {
		if r.Type == "Microsoft.Resources/resourceGroups" {
			// AKS managed clusters create additional resource groups. Only
			// include independent resource groups.
			if _, ok := r.Tags["aks-managed-cluster-rg"]; !ok {
				result = append(result, r)
			}
		}
	}
	return result
}

// parseJSONResources parses the result of resource query into Resource(s).
func parseJSONResources(r []byte) ([]resource, error) {
	if len(r) == 0 {
		return nil, errors.New("failed to JSON parse empty bytes")
	}
	var resources []resource
	if err := json.Unmarshal(r, &resources); err != nil {
		// When unmarshal fails, provide the full output response to make it
		// easier to see the content of response that it failed to unmarshal.
		return nil, fmt.Errorf("failed to unmarshal: %w, content: %q", err, string(r))
	}
	return resources, nil
}

// parseTag parse tags.
// TODO: Add support for multiple key-value pairs after adding support for
// multiple tag query for all the cloud providers.
func parseTag(tags string) (string, string, error) {
	if tags == "" {
		return "", "", errors.New("-tags flag must be set with tag key")
	}
	parts := strings.Split(tags, "=")
	if len(parts) < 2 {
		return "", "", errors.New("unvalid tags, must have a key and a value separated by '='")
	}
	if len(parts) > 2 {
		return "", "", errors.New("only single key-value tag is supported at present")
	}
	return parts[0], parts[1], nil
}
