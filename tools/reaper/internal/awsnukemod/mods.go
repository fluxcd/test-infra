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

// Code in this file extends the Nuke data type defined in nuke.go.
package awsnukemod

import (
	"fmt"
	"log"
	"time"

	"github.com/k1LoW/duration"
	"github.com/rebuy-de/aws-nuke/v2/cmd"

	"github.com/fluxcd/test-infra/tftestenv"
)

// TODO: Combine this and the createdat constant in filter.go.
const createdat = "createdat"

// Items returns the resource items that Nuke has observed.
func (n *Nuke) Items() cmd.Queue {
	return n.items
}

// ApplyRetentionFilter applies the retention filter on the aws-nuke items that
// are to be removed. It only alters the items that were already selected for
// removal by checking if the retention period applies to them. If an item is to
// be removed but don't contain the createdat tag, it is filtered to not be
// removed. This function reduces the number of items to be removed or keeps
// them the same as before. It never increases the items to be deleted.
func (n *Nuke) ApplyRetentionFilter(period string) error {
	p, err := duration.Parse(period)
	if err != nil {
		return err
	}
	// Subtract period from now in UTC.
	now := time.Now().UTC()
	date := now.Add(-p)

	// Read the createdat tag from resources and check if they were created
	// before date.
	for _, item := range n.items {
		if item.State != cmd.ItemStateNew {
			continue
		}

		prop, err := getCreatedAt(item)
		// TODO: Maybe return the error if encountered?
		if err != nil || prop == "" {
			// The item will be removed but don't contain the createdat tag.
			// Update the item state to filtered to avoid deleting them.
			item.State = cmd.ItemStateFiltered
			item.Reason = "filtered by retention-period"
			continue
		}

		createdat, err := tftestenv.ParseCreatedAtTime(prop)
		if err != nil {
			return err
		}
		// If the item was not created before the retention period, filter it
		// out to not be removed.
		if !createdat.Before(date) {
			item.State = cmd.ItemStateFiltered
			item.Reason = "filtered by retention-period"
		}
	}

	return nil
}

// getCreatedAt returns the value of createdat tag if present on the given
// resource item. Some resources have variation in the tag key, it attempts to
// handle them and get the createdat tag value.
func getCreatedAt(item *cmd.Item) (string, error) {
	// TODO: Maybe base this on the type of resource with switch case.
	propPrefix := []string{"tag", "tag:role", "tag:igw"}
	for _, pp := range propPrefix {
		prop, err := item.GetProperty(fmt.Sprintf("%s:%s", pp, createdat))
		if err != nil {
			return "", err
		}
		if prop != "" {
			return prop, nil
		}
	}
	return "", nil
}

// GatherResources runs aws-nuke scan to gather all the resource details.
func (n *Nuke) GatherResources() error {
	err := n.Config.ValidateAccount(n.Account.ID(), n.Account.Aliases())
	if err != nil {
		return err
	}

	return n.Scan()
}

// Delete deletes the resources. This is based on the upstream Nuke.Run() which
// deletes the resources at the end.
func (n *Nuke) Delete() error {
	failCount := 0
	waitingCount := 0

	for {
		n.HandleQueue()

		if n.items.Count(cmd.ItemStatePending, cmd.ItemStateWaiting, cmd.ItemStateNew) == 0 && n.items.Count(cmd.ItemStateFailed) > 0 {
			if failCount >= 2 {
				log.Println("There are resources in failed state, but none are ready for deletion, anymore.")
				fmt.Println()

				for _, item := range n.items {
					if item.State != cmd.ItemStateFailed {
						continue
					}

					item.Print()
					log.Println(item.Reason)
				}

				return fmt.Errorf("failed")
			}

			failCount = failCount + 1
		} else {
			failCount = 0
		}
		if n.Parameters.MaxWaitRetries != 0 && n.items.Count(cmd.ItemStateWaiting, cmd.ItemStatePending) > 0 && n.items.Count(cmd.ItemStateNew) == 0 {
			if waitingCount >= n.Parameters.MaxWaitRetries {
				return fmt.Errorf("max wait retries of %d exceeded", n.Parameters.MaxWaitRetries)
			}
			waitingCount = waitingCount + 1
		} else {
			waitingCount = 0
		}
		if n.items.Count(cmd.ItemStateNew, cmd.ItemStatePending, cmd.ItemStateFailed, cmd.ItemStateWaiting) == 0 {
			break
		}

		time.Sleep(5 * time.Second)
	}

	fmt.Printf("Nuke complete: %d failed, %d skipped, %d finished.\n\n",
		n.items.Count(cmd.ItemStateFailed), n.items.Count(cmd.ItemStateFiltered), n.items.Count(cmd.ItemStateFinished))

	return nil
}
