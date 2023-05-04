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
	"flag"
	"fmt"
	"log"
	"os/exec"
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

// parseJSONResources parses the result of resource query into Resource(s).
func parseJSONResources(r []byte) ([]resource, error) {
	var resources []resource
	if err := json.Unmarshal(r, &resources); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}
	return resources, nil
}
