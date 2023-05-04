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
	"errors"
	"fmt"
	"strings"

	"github.com/fluxcd/test-infra/tftestenv"
)

// queryGCP returns a GCP command for querying all the resources in a specific
// format compatible with the resource type.
func queryGCP(binPath, jqPath, project, labelKey, labelVal string) string {
	return fmt.Sprintf(`%[1]s asset search-all-resources --project %[3]s --query='labels.%[4]s=%[5]s' --format=json |
			%[2]s '.[] |
			{"name": "\(.displayName)", "type": "\(.assetType)", "location": "\(.location)", "resourceGroup": "%[3]s", "tags": .labels}' |
			%[2]s -s '.'`,
		binPath, jqPath, project, labelKey, labelVal)
}

// deleteGCPArtifactRepositoryCmd returns a gcloud command for deleting a Google
// Artifact Repository instance.
func deleteGCPArtifactRepositoryCmd(binPath, project, name, location string) string {
	return fmt.Sprintf(`%[1]s artifacts repositories delete %[3]s --project %[2]s --location %[4]s --quiet`,
		binPath, project, name, location)
}

// deleteGCPClusterCmd returns a gcloud command for deleting a GKE cluster.
func deleteGCPClusterCmd(binPath, project, name, location string) string {
	return fmt.Sprintf(`%[1]s container clusters delete %[3]s --project %[2]s --location %[4]s --quiet`,
		binPath, project, name, location)
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

// deleteGCPCluster deletes a GKE cluster.
func deleteGCPCluster(ctx context.Context, cliPath string, res resource) error {
	_, err := tftestenv.RunCommandWithOutput(ctx, "./",
		deleteGCPClusterCmd(cliPath, res.ResourceGroup, res.Name, res.Location),
		tftestenv.RunCommandOptions{AttachConsole: true},
	)
	return err
}

// deleteGCPArtifactRepository deletes a Google Artifact Repository.
func deleteGCPArtifactRepository(ctx context.Context, cliPath string, res resource) error {
	_, err := tftestenv.RunCommandWithOutput(ctx, "./",
		deleteGCPArtifactRepositoryCmd(cliPath, res.ResourceGroup, res.Name, res.Location),
		tftestenv.RunCommandOptions{AttachConsole: true},
	)
	return err
}
