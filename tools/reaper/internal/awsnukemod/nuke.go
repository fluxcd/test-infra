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

The MIT License (MIT)

Copyright (c) 2016 reBuy reCommerce GmbH

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

This file is copied from the source at
https://github.com/rebuy-de/aws-nuke/blob/v2.25.0/cmd/nuke.go, modified to keep
only the functions we need. Function modification is stated in the function
comment explicitly. The extension of the data type defined in this file exist in
mods.go.
*/

package awsnukemod

import (
	"fmt"

	"github.com/rebuy-de/aws-nuke/v2/cmd"
	"github.com/rebuy-de/aws-nuke/v2/pkg/awsutil"
	"github.com/rebuy-de/aws-nuke/v2/pkg/config"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
	"github.com/rebuy-de/aws-nuke/v2/resources"
	"github.com/sirupsen/logrus"
)

type Nuke struct {
	Parameters cmd.NukeParameters
	Account    awsutil.Account
	Config     *config.Nuke

	ResourceTypes types.Collection

	items cmd.Queue
}

func NewNuke(params cmd.NukeParameters, account awsutil.Account) *Nuke {
	n := Nuke{
		Parameters: params,
		Account:    account,
	}

	return &n
}

// Most of this code is covered by the MIT License except for the commented out
// the print statements at the bottom, which are covered by the Apache License.
func (n *Nuke) Scan() error {
	accountConfig := n.Config.Accounts[n.Account.ID()]

	resourceTypes := cmd.ResolveResourceTypes(
		resources.GetListerNames(),
		resources.GetCloudControlMapping(),
		[]types.Collection{
			n.Parameters.Targets,
			n.Config.ResourceTypes.Targets,
			accountConfig.ResourceTypes.Targets,
		},
		[]types.Collection{
			n.Parameters.Excludes,
			n.Config.ResourceTypes.Excludes,
			accountConfig.ResourceTypes.Excludes,
		},
		[]types.Collection{
			n.Parameters.CloudControl,
			n.Config.ResourceTypes.CloudControl,
			accountConfig.ResourceTypes.CloudControl,
		},
	)

	queue := make(cmd.Queue, 0)

	for _, regionName := range n.Config.Regions {
		region := cmd.NewRegion(regionName, n.Account.ResourceTypeToServiceType, n.Account.NewSession)

		items := cmd.Scan(region, resourceTypes)
		for item := range items {
			ffGetter, ok := item.Resource.(resources.FeatureFlagGetter)
			if ok {
				ffGetter.FeatureFlags(n.Config.FeatureFlags)
			}

			queue = append(queue, item)
			err := n.Filter(item)
			if err != nil {
				return err
			}

			// if item.State != cmd.ItemStateFiltered || !n.Parameters.Quiet {
			// 	item.Print()
			// }
		}
	}

	// fmt.Printf("Scan complete: %d total, %d nukeable, %d filtered.\n\n",
	// 	queue.CountTotal(), queue.Count(cmd.ItemStateNew), queue.Count(cmd.ItemStateFiltered))

	n.items = queue

	return nil
}

func (n *Nuke) Filter(item *cmd.Item) error {

	checker, ok := item.Resource.(resources.Filter)
	if ok {
		err := checker.Filter()
		if err != nil {
			item.State = cmd.ItemStateFiltered
			item.Reason = err.Error()

			// Not returning the error, since it could be because of a failed
			// request to the API. We do not want to block the whole nuking,
			// because of an issue on AWS side.
			return nil
		}
	}

	accountFilters, err := n.Config.Filters(n.Account.ID())
	if err != nil {
		return err
	}

	itemFilters, ok := accountFilters[item.Type]
	if !ok {
		return nil
	}

	for _, filter := range itemFilters {
		prop, err := item.GetProperty(filter.Property)
		if err != nil {
			logrus.Warnf(err.Error())
			continue
		}
		match, err := filter.Match(prop)
		if err != nil {
			return err
		}

		if cmd.IsTrue(filter.Invert) {
			match = !match
		}

		if match {
			item.State = cmd.ItemStateFiltered
			item.Reason = "filtered by config"
			return nil
		}
	}

	return nil
}

func (n *Nuke) HandleQueue() {
	listCache := make(map[string]map[string][]resources.Resource)

	for _, item := range n.items {
		switch item.State {
		case cmd.ItemStateNew:
			n.HandleRemove(item)
			item.Print()
		case cmd.ItemStateFailed:
			n.HandleRemove(item)
			n.HandleWait(item, listCache)
			item.Print()
		case cmd.ItemStatePending:
			n.HandleWait(item, listCache)
			item.State = cmd.ItemStateWaiting
			item.Print()
		case cmd.ItemStateWaiting:
			n.HandleWait(item, listCache)
			item.Print()
		}

	}

	fmt.Println()
	fmt.Printf("Removal requested: %d waiting, %d failed, %d skipped, %d finished\n\n",
		n.items.Count(cmd.ItemStateWaiting, cmd.ItemStatePending), n.items.Count(cmd.ItemStateFailed),
		n.items.Count(cmd.ItemStateFiltered), n.items.Count(cmd.ItemStateFinished))
}

func (n *Nuke) HandleRemove(item *cmd.Item) {
	err := item.Resource.Remove()
	if err != nil {
		item.State = cmd.ItemStateFailed
		item.Reason = err.Error()
		return
	}

	item.State = cmd.ItemStatePending
	item.Reason = ""
}

func (n *Nuke) HandleWait(item *cmd.Item, cache map[string]map[string][]resources.Resource) {
	var err error
	region := item.Region.Name
	_, ok := cache[region]
	if !ok {
		cache[region] = map[string][]resources.Resource{}
	}
	left, ok := cache[region][item.Type]
	if !ok {
		left, err = item.List()
		if err != nil {
			item.State = cmd.ItemStateFailed
			item.Reason = err.Error()
			return
		}
		cache[region][item.Type] = left
	}

	for _, r := range left {
		if item.Equals(r) {
			checker, ok := r.(resources.Filter)
			if ok {
				err := checker.Filter()
				if err != nil {
					break
				}
			}

			return
		}
	}

	item.State = cmd.ItemStateFinished
	item.Reason = ""
}
