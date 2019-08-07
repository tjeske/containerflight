// Copyright © 2018 Tobias Jeske
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package version

import (
	"github.com/blang/semver"
	"github.com/tjeske/containerflight/util"
)

var versionStr = "0.2.2"

// ContainerFlightVersion returns the current containerflight version
func ContainerFlightVersion() semver.Version {
	containerFlightVersion, err := semver.Make(versionStr)
	util.CheckErr(err)
	return containerFlightVersion
}
