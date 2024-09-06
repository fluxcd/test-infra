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
	"time"

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

// queryGCPSourceRepos returns a GCP command for querying all the source
// repositories in a specific format compatible with the resource type.
func queryGCPSourceRepos(binPath, jqPath, project string) string {
	// TODO: Figure out a better way to detect the age of the repository.
	// Currently, all the repositories are deleted with fixed now-1hr createdat
	// time.
	now := time.Now().UTC()
	createdat := now.Add(-time.Hour)
	tagVal := createdat.Format(tftestenv.CreatedAtTimeLayout)
	return fmt.Sprintf(`%[1]s source repos list --project %[3]s --format=json |
			%[2]s '.[] |
			{name, "type": "cloud-source-repository", "tags": { "createdat": "%[4]s" }}' |
			%[2]s -s '.'`,
		binPath, jqPath, project, tagVal)
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

// deleteGCPSourceRepoCmd returns a gcloud command for deleting cloud source
// repository.
func deleteGCPSourceRepoCmd(binPath, project, name string) string {
	return fmt.Sprintf(`%[1]s source repos delete %[3]s --project %[2]s --quiet`,
		binPath, project, name)
}

func getGCPSourceRepos(ctx context.Context, cliPath, jqPath string) ([]resource, error) {
	output, err := tftestenv.RunCommandWithOutput(ctx, "./",
		queryGCPSourceRepos(cliPath, jqPath, *gcpProject),
		tftestenv.RunCommandOptions{},
	)
	if err != nil {
		return nil, err
	}
	return parseJSONResources(output)
}

// getGCPResources queries GCP for resources.
func getGCPResources(ctx context.Context, cliPath, jqPath string) ([]resource, error) {
	result := []resource{}
	output, err := tftestenv.RunCommandWithOutput(ctx, "./",
		queryGCP(cliPath, jqPath, *gcpProject, tagKey, tagVal),
		tftestenv.RunCommandOptions{},
	)
	if err != nil {
		return nil, err
	}
	r, err := parseJSONResources(output)
	if err != nil {
		return nil, err
	}
	result = append(result, r...)

	sr, err := getGCPSourceRepos(ctx, cliPath, jqPath)
	if err != nil {
		return nil, err
	}
	result = append(result, sr...)

	return result, nil
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

// deleteGCPSourceRepo deletes a Google cloud source repository.
func deleteGCPSourceRepo(ctx context.Context, cliPath string, res resource) error {
	_, err := tftestenv.RunCommandWithOutput(ctx, "./",
		deleteGCPSourceRepoCmd(cliPath, *gcpProject, res.Name),
		tftestenv.RunCommandOptions{AttachConsole: true},
	)
	return err
}
