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
	"fmt"

	"github.com/fluxcd/test-infra/tftestenv"
)

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
			{name, type, tags, location, resourceGroup}' |
			%[2]s -s '.'`,
		binPath, jqPath, tagKey, tagVal)
}

// deleteAzureResourceGroupCmd returns an Azure command for deleting a resource
// group.
func deleteAzureResourceGroupCmd(binPath, name string) string {
	return fmt.Sprintf(`%[1]s group delete --name %[2]s --yes`,
		binPath, name)
}

// deleteAzureResourceCmd returns an Azure command for deleting any given
// resource in a resource group.
func deleteAzureResourceCmd(binPath, group, name, rType string) string {
	return fmt.Sprintf(`%[1]s resource delete --resource-group %[2]s --name %[3]s --resource-type "%[4]s"`,
		binPath, group, name, rType)
}

// getAzureResources queries Azure for resources. Azure has two separate APIs
// for listing Resource Groups and all the other resources. Query both and
// combine the result.
func getAzureResources(ctx context.Context, cliPath, jqPath string) ([]resource, error) {
	// Query Resource Groups.
	groupOutput, err := tftestenv.RunCommandWithOutput(ctx, "./",
		queryAzureGroups(cliPath, jqPath, tagKey, tagVal),
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
		queryAzureResources(cliPath, jqPath, tagKey, tagVal),
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

// deleteAzureResourceGroup deletes an Azure resource group.
func deleteAzureResourceGroup(ctx context.Context, cliPath string, res resource) error {
	_, err := tftestenv.RunCommandWithOutput(ctx, "./",
		deleteAzureResourceGroupCmd(cliPath, res.Name),
		tftestenv.RunCommandOptions{AttachConsole: true},
	)
	return err
}

// deleteAzureResource deletes an Azure resource.
// NOTE: This is unused for now as deleting the resource groups deletes
// everything. Use it in the future when there's a need to delete individual
// resources regardless of their resource groups.
func deleteAzureResource(ctx context.Context, cliPath string, res resource) error {
	_, err := tftestenv.RunCommandWithOutput(ctx, "./",
		deleteAzureResourceCmd(cliPath, res.ResourceGroup, res.Name, res.Type),
		tftestenv.RunCommandOptions{AttachConsole: true},
	)
	return err
}
