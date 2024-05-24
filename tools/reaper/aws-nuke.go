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
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/rebuy-de/aws-nuke/v2/cmd"
	"github.com/rebuy-de/aws-nuke/v2/pkg/awsutil"
	"github.com/rebuy-de/aws-nuke/v2/pkg/config"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
	awsresources "github.com/rebuy-de/aws-nuke/v2/resources"

	"github.com/fluxcd/test-infra/tftestenv"
	"github.com/fluxcd/test-infra/tools/reaper/internal/awsnukemod"
)

// getAWSAccountID returns the AWS account ID of the target aws account.
func getAWSAccountID(ctx context.Context, cliPath string) (string, error) {
	output, err := tftestenv.RunCommandWithOutput(ctx, "./",
		fmt.Sprintf(`%s sts get-caller-identity --query "Account" --output text`, cliPath),
		tftestenv.RunCommandOptions{StdoutOnly: true},
	)
	if err != nil {
		return "", err
	}
	id := strings.TrimSpace(string(output))
	if len(id) == 0 {
		return "", fmt.Errorf("could not get aws account ID")
	}
	return id, nil
}

// getAWSNukeConfig returns the aws-nuke configuration to be used against flux
// test-infra.
func getAWSNukeConfig(accountID string, regions []string) *config.Nuke {
	nukeRegions := []string{"global"}
	nukeRegions = append(nukeRegions, regions...)

	tagFilter := config.Filter{
		Property: fmt.Sprintf("tag:%s", tagKey),
		Value:    tagVal,
		Invert:   "true",
	}
	tagRoleFilter := config.Filter{
		Property: fmt.Sprintf("tag:role:%s", tagKey),
		Value:    tagVal,
		Invert:   "true",
	}
	tagIGWFilter := config.Filter{
		Property: fmt.Sprintf("tag:igw:%s", tagKey),
		Value:    tagVal,
		Invert:   "true",
	}

	return &config.Nuke{
		Regions: nukeRegions,
		// Set a fake account in the blocklist to suppress this validation
		// https://github.com/rebuy-de/aws-nuke/blob/v2.25.0/pkg/config/config.go#L121-L125.
		// It is a requirement to set a production account ID in blocklist.
		AccountBlocklist: []string{"999999999999"},
		Accounts: map[string]config.Account{
			accountID: {
				ResourceTypes: config.ResourceTypes{
					Targets: types.Collection{
						"EC2VPC",
						"EC2SecurityGroup",
						"EC2LaunchTemplate",
						"EC2RouteTable",
						"EC2NetworkInterface",
						"ECRRepository",
						"EC2Volume",
						"EKSNodegroups",
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
				Filters: config.Filters{
					"EC2VPC":                       []config.Filter{tagFilter},
					"EC2SecurityGroup":             []config.Filter{tagFilter},
					"EC2LaunchTemplate":            []config.Filter{tagFilter},
					"EC2RouteTable":                []config.Filter{tagFilter},
					"EC2NetworkInterface":          []config.Filter{tagFilter},
					"ECRRepository":                []config.Filter{tagFilter},
					"EC2Volume":                    []config.Filter{tagFilter},
					"EKSNodegroups":                []config.Filter{tagFilter},
					"EC2Subnet":                    []config.Filter{tagFilter},
					"AutoScalingGroup":             []config.Filter{tagFilter},
					"EC2Address":                   []config.Filter{tagFilter},
					"EKSCluster":                   []config.Filter{tagFilter},
					"EC2InternetGatewayAttachment": []config.Filter{tagIGWFilter},
					"EC2InternetGateway":           []config.Filter{tagFilter},
					"EC2Instance":                  []config.Filter{tagFilter},
					"EC2NATGateway":                []config.Filter{tagFilter},
					"IAMRole":                      []config.Filter{tagFilter},
					"IAMRolePolicy":                []config.Filter{tagFilter},
					"IAMRolePolicyAttachment":      []config.Filter{tagRoleFilter},
					"IAMPolicy":                    []config.Filter{tagFilter},
					"IAMOpenIDConnectProvider":     []config.Filter{tagFilter},
				},
			},
		},
	}
}

// awsnukeScan configures and scans the target aws account with aws-nuke, and
// returns an instance of aws-nuke.
func awsnukeScan(accountID string) (*awsnukemod.Nuke, error) {
	// Parse the regions and set the default region.
	regions := strings.Split(*awsRegions, ",")
	// Use the first region as the default.
	defaultRegion := regions[0]

	var creds awsutil.Credentials

	// Read aws credentials from the environment and validate.
	creds.AccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	creds.SecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	creds.Profile = os.Getenv("AWS_PROFILE")
	creds.SessionToken = os.Getenv("AWS_SESSION_TOKEN")
	creds.AssumeRoleArn = os.Getenv("AWS_ROLE_ARN")
	if creds.HasProfile() && creds.HasKeys() {
		return nil, fmt.Errorf("please provide either AWS_PROFILE or " +
			"AWS_ACCESS_KEY_ID with AWS_SECRET_ACCESS_KEY and optionally " +
			"AWS_SESSION_TOKEN environment variables")
	}

	nukeCfg := getAWSNukeConfig(accountID, regions)

	if defaultRegion != "" {
		awsutil.DefaultRegionID = defaultRegion
		switch defaultRegion {
		case endpoints.UsEast1RegionID, endpoints.UsEast2RegionID, endpoints.UsWest1RegionID, endpoints.UsWest2RegionID:
			awsutil.DefaultAWSPartitionID = endpoints.AwsPartitionID
		case endpoints.UsGovEast1RegionID, endpoints.UsGovWest1RegionID:
			awsutil.DefaultAWSPartitionID = endpoints.AwsUsGovPartitionID
		case endpoints.CnNorth1RegionID, endpoints.CnNorthwest1RegionID:
			awsutil.DefaultAWSPartitionID = endpoints.AwsCnPartitionID
		default:
			if nukeCfg.CustomEndpoints.GetRegion(defaultRegion) == nil {
				err := fmt.Errorf("the custom region '%s' must be specified in the configuration 'endpoints'", defaultRegion)
				return nil, err
			}
		}
	}

	account, err := awsutil.NewAccount(creds, nukeCfg.CustomEndpoints)
	if err != nil {
		return nil, err
	}
	params := cmd.NukeParameters{
		Quiet: true,
	}
	n := awsnukemod.NewNuke(params, *account)
	n.Config = nukeCfg

	err = n.GatherResources()
	return n, err
}

// awsnukeItemsToResources converts the items which would be removed to resource
// type.
func awsnukeItemsToResources(items cmd.Queue) []resource {
	resources := []resource{}

	for _, item := range items {
		// Only consider items that would be removed.
		if item.State != cmd.ItemStateNew {
			continue
		}

		r := resource{}
		rString, ok := item.Resource.(awsresources.LegacyStringer)
		if ok {
			r.Name = rString.String()
		}
		r.Location = item.Region.Name
		r.Type = item.Type
		r.Tags = map[string]string{}
		rProp, ok := item.Resource.(awsresources.ResourcePropertyGetter)
		if ok {
			r.Tags = rProp.Properties()
		}

		resources = append(resources, r)
	}

	return resources
}
