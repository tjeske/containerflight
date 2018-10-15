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

package appinfo

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/tjeske/containerflight/util"
)

type environment struct {
	appConfigFile string
	appFileDir    string
	userName      string
	userID        string
	groupName     string
	groupID       string
	homeDir       string
	workingDir    string
}

// Determine the current environment
var getEnv = func(appConfigFile string) environment {

	absAppConfigFile, err := filepath.Abs(appConfigFile)
	util.CheckErr(err)

	appFileDir := filepath.Dir(absAppConfigFile)

	// current user
	currentUser, err := user.Current()
	util.CheckErr(err)

	groupName, err := user.LookupGroupId(currentUser.Gid)
	util.CheckErr(err)

	workingDir, err := os.Getwd()
	util.CheckErr(err)

	// create environment object
	var env = environment{
		appConfigFile: absAppConfigFile,
		appFileDir:    appFileDir,
		userName:      currentUser.Username,
		userID:        currentUser.Uid,
		groupName:     groupName.Name,
		groupID:       currentUser.Gid,
		homeDir:       currentUser.HomeDir,
		workingDir:    workingDir,
	}
	return env
}
