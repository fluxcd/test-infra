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

// Code in this file extends the Nuke data type defined in nuke.go and adds
// other helpers for using libnuke.
package libnukemod

import (
	"context"
	"fmt"
	"time"

	"github.com/ekristen/libnuke/pkg/queue"
	"github.com/k1LoW/duration"

	"github.com/fluxcd/test-infra/tftestenv"
)

// TODO: Combine this and the createdat constant in filter.go.
const createdat = "createdat"

// ApplyRetentionFilter applies the retention filter on the Nuke queue items
// that are to be removed. It only alters the items that were already selected
// for removal by checking if the retention period applies to them. If an item
// is to be removed but don't contain the createdat tag, it is filtered to not
// be removed. This function reduces the number of items to be removed or keeps
// them the same as before. It never increases the items to be deleted.
func ApplyRetentionFilter(n *Nuke, period string) error {
	p, err := duration.Parse(period)
	if err != nil {
		return err
	}
	// Subtract period from now in UTC.
	now := time.Now().UTC()
	date := now.Add(-p)

	// Read the createdat tag from resources and check if they were created
	// before date.
	for _, item := range n.Queue.Items {
		if item.State != queue.ItemStateNew {
			continue
		}

		prop, err := getCreatedAt(item)
		// TODO: Maybe return the error if encountered?
		if err != nil || prop == "" {
			// The item will be removed but don't contain the createdat tag.
			// Update the item state to filtered to avoid deleting them.
			item.State = queue.ItemStateFiltered
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
			item.State = queue.ItemStateFiltered
			item.Reason = "filtered by retention-period"
		}
	}

	return nil
}

// getCreatedAt returns the value of createdat tag if present on the given
// resource item. Some resources have variation in the tag key, it attempts to
// handle them and get the createdat tag value.
func getCreatedAt(item *queue.Item) (string, error) {
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

// Delete deletes the resources. This deletes the existing scanned items in
// nuke, skipping a re-scan and summarizes the result of delete.
func (n *Nuke) Delete(ctx context.Context) error {
	if err := n.run(ctx); err != nil {
		return err
	}
	fmt.Printf("Nuke complete: %d failed, %d skipped, %d finished.\n\n",
		n.Queue.Count(queue.ItemStateFailed), n.Queue.Count(queue.ItemStateFiltered), n.Queue.Count(queue.ItemStateFinished))

	return nil
}
