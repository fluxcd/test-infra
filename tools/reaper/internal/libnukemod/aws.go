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

// Code in this file uses the Nuke data type defined in nuke.go to provider
// helpers for aws-nuke.
package libnukemod

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/ekristen/aws-nuke/v3/pkg/awsutil"
	"github.com/ekristen/aws-nuke/v3/pkg/config"
	awsnuke "github.com/ekristen/aws-nuke/v3/pkg/nuke"
	_ "github.com/ekristen/aws-nuke/v3/resources" // Register all the AWS resources.
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/scanner"
	"github.com/ekristen/libnuke/pkg/types"
	"github.com/sirupsen/logrus"
)

// SetUpLibnukeAWS configures and returns Nuke for AWS. This is based on the
// aws-nuke nuke command.
func SetUpLibnukeAWS(ctx context.Context, accountID string, defaultRegion string, cfg config.Config) (*Nuke, error) {
	// Read aws credentials from the environment and validate.
	creds := &awsutil.Credentials{}
	creds.AccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	creds.SecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	creds.Profile = os.Getenv("AWS_PROFILE")
	creds.SessionToken = os.Getenv("AWS_SESSION_TOKEN")
	creds.AssumeRoleArn = os.Getenv("AWS_ASSUME_ROLE_ARN")
	creds.RoleSessionName = os.Getenv("AWS_ASSUME_ROLE_SESSION_NAME")
	creds.ExternalID = os.Getenv("AWS_ASSUME_ROLE_EXTERNAL_ID")

	if err := creds.Validate(); err != nil {
		return nil, err
	}

	// Initialize the libnuke configuration based on libnuke config.New()
	// manually since we aren't reading the configuration from a file path.
	if cfg.Log == nil {
		cfg.Log = logrus.WithContext(ctx)
	}
	cfg.Deprecations = registry.GetDeprecatedResourceTypeMapping()
	if err := cfg.ResolveDeprecations(); err != nil {
		return nil, err
	}
	cfg.Blocklist = cfg.ResolveBlocklist()

	if defaultRegion != "" {
		awsutil.DefaultRegionID = defaultRegion

		partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), defaultRegion)
		if !ok {
			if cfg.CustomEndpoints.GetRegion(defaultRegion) == nil {
				err := fmt.Errorf(
					"the custom region '%s' must be specified in the configuration 'endpoints'"+
						" to determine its partition", defaultRegion)
				logrus.WithError(err).Errorf("unable to resolve partition for region: %s", defaultRegion)
				return nil, err
			}
		}

		awsutil.DefaultAWSPartitionID = partition.ID()
	}

	// Create the AWS account object.
	account, err := awsutil.NewAccount(creds, cfg.CustomEndpoints)
	if err != nil {
		return nil, err
	}

	// Get filters for the account.
	filters, err := cfg.Filters(account.ID())
	if err != nil {
		return nil, err
	}

	params := &Parameters{
		// Hide filtered resources.
		Quiet: true,
	}

	// Instantiate libnuke.
	n := New(params, filters, cfg.Settings)

	// TODO: Register version?
	n.RegisterValidateHandler(func() error {
		return cfg.ValidateAccount(account.ID(), account.Aliases(), true)
	})

	// Get any specific account level configuration
	accountConfig := cfg.Accounts[account.ID()]

	// Resolve the resource types to be used for the nuke process based on the parameters, global configuration, and
	// account level configuration.
	resourceTypes := types.ResolveResourceTypes(
		registry.GetNames(),
		[]types.Collection{
			n.Parameters.Includes,
			cfg.ResourceTypes.GetIncludes(),
			accountConfig.ResourceTypes.GetIncludes(),
		},
		[]types.Collection{
			n.Parameters.Excludes,
			cfg.ResourceTypes.Excludes,
			accountConfig.ResourceTypes.Excludes,
		},
		[]types.Collection{
			n.Parameters.Alternatives,
			cfg.ResourceTypes.GetAlternatives(),
			accountConfig.ResourceTypes.GetAlternatives(),
		},
		registry.GetAlternativeResourceTypeMapping(),
	)

	// If the user has specified the "all" region, then we need to get the enabled regions for the account
	// and use those. Otherwise, we will use the regions that are specified in the configuration.
	if slices.Contains(cfg.Regions, "all") {
		cfg.Regions = account.Regions()

		logrus.Info(
			`"all" detected in region list, only enabled regions and "global" will be used, all others ignored`)

		if len(cfg.Regions) > 1 {
			logrus.Warnf(`additional regions defined along with "all", these will be ignored!`)
		}

		logrus.Infof("The following regions are enabled for the account (%d total):", len(cfg.Regions))

		printableRegions := make([]string, 0)
		for i, region := range cfg.Regions {
			printableRegions = append(printableRegions, region)
			if i%6 == 0 { // print 5 regions per line
				logrus.Infof("> %s", strings.Join(printableRegions, ", "))
				printableRegions = make([]string, 0)
			} else if i == len(cfg.Regions)-1 {
				logrus.Infof("> %s", strings.Join(printableRegions, ", "))
			}
		}
	}

	// Register the scanners for each region that is defined in the configuration.
	for _, regionName := range cfg.Regions {
		// Step 1 - Create the region object
		region := awsnuke.NewRegion(regionName, account.ResourceTypeToServiceType, account.NewSession)

		// Step 2 - Create the scannerActual object
		scannerActual := scanner.New(regionName, resourceTypes, &awsnuke.ListerOpts{
			Region: region,
		})

		// Step 3 - Register a mutate function that will be called to modify the lister options for each resource type
		// see pkg/nuke/resource.go for the MutateOpts function. Its purpose is to create the proper session for the
		// proper region.
		regMutateErr := scannerActual.RegisterMutateOptsFunc(awsnuke.MutateOpts)
		if regMutateErr != nil {
			return nil, regMutateErr
		}

		// Step 4 - Register the scannerActual with the nuke object
		regScanErr := n.RegisterScanner(awsnuke.Account, scannerActual)
		if regScanErr != nil {
			return nil, regScanErr
		}
	}

	return n, nil
}
