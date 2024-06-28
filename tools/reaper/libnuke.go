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

package main

import (
	"context"
	"fmt"
	"strings"

	awsnukeconfig "github.com/ekristen/aws-nuke/v3/pkg/config"
	libnukeconfig "github.com/ekristen/libnuke/pkg/config"
	"github.com/ekristen/libnuke/pkg/filter"
	"github.com/ekristen/libnuke/pkg/queue"
	libnukeresource "github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/fluxcd/test-infra/tftestenv"
	"github.com/fluxcd/test-infra/tools/reaper/internal/libnukemod"
)

// getAWSAccountID returns the AWS account ID of the target aws account.
func getAWSAccountID(ctx context.Context, cliPath string) (string, error) {
	output, err := tftestenv.RunCommandWithOutput(ctx, "./",
		fmt.Sprintf(`%s sts get-caller-identity --query "Account" --output text`, cliPath),
		tftestenv.RunCommandOptions{StdoutOnly: true},
	)
	if err != nil {
		return "", fmt.Errorf("%s execution failed to get account ID: %w", cliPath, err)
	}
	id := strings.TrimSpace(string(output))
	if len(id) == 0 {
		return "", fmt.Errorf("could not get aws account ID")
	}
	return id, nil
}

// getLibNukeAWSConfig returns the aws-nuke configuration to be used against flux
// test-infra.
func getLibNukeAWSConfig(accountID string, regions []string) *libnukeconfig.Config {
	nukeRegions := []string{"global"}
	nukeRegions = append(nukeRegions, regions...)

	tagFilter := filter.Filter{
		Property: fmt.Sprintf("tag:%s", tagKey),
		Value:    tagVal,
		Invert:   "true",
	}
	tagRoleFilter := filter.Filter{
		Property: fmt.Sprintf("tag:role:%s", tagKey),
		Value:    tagVal,
		Invert:   "true",
	}
	tagIGWFilter := filter.Filter{
		Property: fmt.Sprintf("tag:igw:%s", tagKey),
		Value:    tagVal,
		Invert:   "true",
	}

	return &libnukeconfig.Config{
		Regions: nukeRegions,
		// Set a fake account in the blocklist to suppress this validation
		// https://github.com/rebuy-de/aws-nuke/blob/v2.25.0/pkg/config/config.go#L121-L125.
		// It is a requirement to set a production account ID in blocklist.
		Blocklist: []string{"999999999999"},
		Accounts: map[string]*libnukeconfig.Account{
			accountID: {
				ResourceTypes: libnukeconfig.ResourceTypes{
					Includes: types.Collection{
						"EC2VPC",
						"EC2SecurityGroup",
						"EC2LaunchTemplate",
						"EC2RouteTable",
						"EC2NetworkInterface",
						"ECRRepository",
						"EC2Volume",
						"EKSNodegroup",
						"EC2Subnet",
						"AutoScalingGroup",
						"EC2Address",
						"EKSCluster",
						"EC2InternetGatewayAttachment",
						"EC2InternetGateway",
						"EC2Instance",
						"EC2NATGateway",
						"IAMRole",
						"IAMRolePolicy",
						"IAMRolePolicyAttachment",
						"IAMPolicy",
						"IAMOpenIDConnectProvider",
					},
				},
				Filters: filter.Filters{
					"EC2VPC":                       []filter.Filter{tagFilter},
					"EC2SecurityGroup":             []filter.Filter{tagFilter},
					"EC2LaunchTemplate":            []filter.Filter{tagFilter},
					"EC2RouteTable":                []filter.Filter{tagFilter},
					"EC2NetworkInterface":          []filter.Filter{tagFilter},
					"ECRRepository":                []filter.Filter{tagFilter},
					"EC2Volume":                    []filter.Filter{tagFilter},
					"EKSNodegroup":                 []filter.Filter{tagFilter},
					"EC2Subnet":                    []filter.Filter{tagFilter},
					"AutoScalingGroup":             []filter.Filter{tagFilter},
					"EC2Address":                   []filter.Filter{tagFilter},
					"EKSCluster":                   []filter.Filter{tagFilter},
					"EC2InternetGatewayAttachment": []filter.Filter{tagIGWFilter},
					"EC2InternetGateway":           []filter.Filter{tagFilter},
					"EC2Instance":                  []filter.Filter{tagFilter},
					"EC2NATGateway":                []filter.Filter{tagFilter},
					"IAMRole":                      []filter.Filter{tagFilter},
					"IAMRolePolicy":                []filter.Filter{tagFilter},
					"IAMRolePolicyAttachment":      []filter.Filter{tagRoleFilter},
					"IAMPolicy":                    []filter.Filter{tagFilter},
					"IAMOpenIDConnectProvider":     []filter.Filter{tagFilter},
				},
			},
		},
	}
}

// libnukeAWSScan configures and scans the target aws account with aws-nuke, and
// returns an instance of libnuke Nuke.
func libnukeAWSScan(ctx context.Context, accountID string) (*libnukemod.Nuke, error) {
	// Parse the regions and set the default region.
	regions := strings.Split(*awsRegions, ",")
	// Use the first region as the default.
	defaultRegion := regions[0]

	parsedConfig := awsnukeconfig.Config{
		Config:                   getLibNukeAWSConfig(accountID, regions),
		BypassAliasCheckAccounts: []string{accountID},
		// TODO: Maybe set custom endpoints?
	}

	n, err := libnukemod.SetUpLibnukeAWS(ctx, accountID, defaultRegion, parsedConfig)
	if err != nil {
		return nil, err
	}

	if err := n.Scan(ctx); err != nil {
		return nil, err
	}

	return n, nil
}

// libnukeItemsToResources converts the items which would be removed to resource
// type.
func libnukeItemsToResources(items []*queue.Item) []resource {
	resources := []resource{}

	for _, item := range items {
		// Only consider items that would be removed.
		if item.State != queue.ItemStateNew {
			continue
		}

		r := resource{}
		rString, ok := item.Resource.(libnukeresource.LegacyStringer)
		if ok {
			r.Name = rString.String()
		}
		r.Location = item.Owner
		r.Type = item.Type
		r.Tags = map[string]string{}
		rProp, ok := item.Resource.(libnukeresource.PropertyGetter)
		if ok {
			r.Tags = rProp.Properties()
		}

		resources = append(resources, r)
	}

	return resources
}
