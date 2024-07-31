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

// Package libnukemod contains copies of code from the libnuke project
// https://github.com/ekristen/libnuke and modifications to it. In order to
// integrate with reaper, the resources observed by libnuke needed to be
// converted into the resource data type of the reaper, so that the list of
// resources can be printed in a coherent manner across all the different
// providers. For this, the Nuke.Run() command, which combines scan and delete,
// had to be split into separate steps. Hence, the mods.go adds Delete() to
// Nuke.
// The Nuke.Scan() function prints all the scanned resources. This breaks the
// reaper interface. Scan() is modified to not print the resources.
// To support the retain-period feature of reaper, aws-nuke needs to understand
// the custom timestamp that test-env uses. Since the default aws-nuke filters
// can't be appended without copying and modifying more code,
// ApplyRetentionFilter() is introduced. This allows applying the filter
// on the items after gathering all the resources and before deleting them.
package libnukemod
