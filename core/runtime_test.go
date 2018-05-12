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
	"testing"

	"github.com/tjeske/containerflight/util"

	"github.com/stretchr/testify/assert"
)

func TestGetRunCmdArgs(t *testing.T) {

	currentDir, err := os.Getwd()
	util.CheckErr(err)

	expDockerRunArgs := []string{
		"--rm",
		"--label", "image=myContainerLabel",
		"--label", "appFile=" + currentDir + string(os.PathSeparator) + "myAppConfigFile",
		"-w", "/newworkingdir",
		"myContainerLabel",
		"arg1",
	}

	dockerRunCmdArgs := getDockerRunCmdArgs(
		[]string{"-w", "/newworkingdir"},
		"myContainerLabel",
		"myAppConfigFile",
		[]string{"arg0", "arg1"},
	)

	assert.Equal(t, expDockerRunArgs, dockerRunCmdArgs)
}
