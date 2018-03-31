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

package core

import (
	"os"
	"os/user"
	"path/filepath"
)

type environment struct {
	appName    string
	appFile    string
	userName   string
	userID     string
	groupName  string
	groupID    string
	homeDir    string
	workingDir string
}

// Determine the current environment
func getEnv(appFile string) *environment {

	// current user
	currentUser, err := user.Current()
	checkErr(err)

	groupName, err := user.LookupGroupId(currentUser.Gid)
	checkErr(err)

	workingDir, err := os.Getwd()
	checkErr(err)

	// create environment object
	var env = environment{
		appName:    filepath.Base(appFile),
		appFile:    appFile,
		userName:   currentUser.Username,
		userID:     currentUser.Uid,
		groupName:  groupName.Name,
		groupID:    currentUser.Gid,
		homeDir:    currentUser.HomeDir,
		workingDir: workingDir}
	return &env
}
