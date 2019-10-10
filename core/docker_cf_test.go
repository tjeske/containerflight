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
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/tjeske/containerflight/appinfo"
)

func init() {
	// emulate file system
	filesystem = afero.NewMemMapFs()
}

func newContainerFlightDockerClient(appInfo *appinfo.AppInfo) *ContainerFlightDockerClient {
	return &ContainerFlightDockerClient{DockerClient: newDockerClient(appInfo)}
}

func TestGetBuildCmdArgs(t *testing.T) {
	appConfigStr := "image:\n    dockerfile: |\n        RUN test"
	appInfo := appinfo.NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)

	dockerClient := newContainerFlightDockerClient(appInfo)
	args := dockerClient.getBuildCmdArgs("dockerfile", "dockerBuildCtx", "label", "hashStr")

	expArgs := []string{
		"dockerBuildCtx",
		"-f", "dockerfile",
		"--label", "containerflight=true",
		"--label", "containerflight_appFile=/testAppFile",
		"--label", "containerflight_hash=hashStr",
		"--label", "containerflight_cfVersion=" + dockerClient.version,
		"--label", "containerflight_description=",
		"-t", "label",
	}
	assert.Equal(t, expArgs, args)
}

func TestGetRunCmdArgs(t *testing.T) {
	appConfigStr := ""
	appInfo := appinfo.NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)

	dockerClient := newContainerFlightDockerClient(appInfo)
	args := dockerClient.getRunCmdArgs("123", []string{"arg1", "arg2"})

	expArgs := []string{
		"--rm",
		"--label", "containerflight_appFile=/testAppFile",
		"--label", "containerflight_image=containerflight_testappfile:unknown",
		"--label", "containerflight_hash=562a792d764ddceb355634b2ccee3878edf696021767ff0e8144eab2e2bf035f",
		"--label", "containerflight_version=" + dockerClient.version,
		"-v", "/myworkingdir:/myworkingdir",
		"-ti",
		"-h", "flybydocker",
		"-w", "/myworkingdir",
		"123",
		"arg1",
		"arg2",
	}
	assert.Equal(t, expArgs, args)
}

func TestGetDockerContainerLabel(t *testing.T) {
	appConfigStr := ""
	appInfo := appinfo.NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)

	dockerClient := newContainerFlightDockerClient(appInfo)
	label := dockerClient.getDockerContainerLabel()

	assert.Equal(t, "containerflight_testappfile:unknown", label)
}

func TestGetDockerContainerLabelVersion(t *testing.T) {
	appConfigStr := "version: \"1.2.3\""
	appInfo := appinfo.NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)

	dockerClient := newContainerFlightDockerClient(appInfo)
	label := dockerClient.getDockerContainerLabel()

	assert.Equal(t, "containerflight_testappfile:1.2.3", label)
}
