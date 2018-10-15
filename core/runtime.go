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
	"fmt"
	"strings"

	"github.com/tjeske/containerflight/appinfo"
)

// PrintDockerfile loads an app file and dump the processed dockerfile
func PrintDockerfile(yamlAppConfigFileName string) {

	appInfo := appinfo.NewAppInfo(yamlAppConfigFileName)
	dockerfile := appInfo.GetDockerfile()

	fmt.Println(dockerfile)
}

// PrintDockerRunArgs show the resulting "docker run" arguments
func PrintDockerRunArgs(yamlAppConfigFileName string) {

	appInfo := appinfo.NewAppInfo(yamlAppConfigFileName)
	dockerClient := NewDockerClient(appInfo)

	imageID := dockerClient.getImageID()
	dockerRunCmdArgs := dockerClient.getRunCmdArgs(imageID, []string{})

	fmt.Println("\"docker run\" will be called with the following arguments:\n" + strings.Join(dockerRunCmdArgs, " "))
}

// Build creates an app container image.
func Build(yamlAppConfigFileName string) {

	appInfo := appinfo.NewAppInfo(yamlAppConfigFileName)
	dockerClient := NewDockerClient(appInfo)

	AppFileDir := appInfo.GetAppFileDir()

	containerLabel := dockerClient.getDockerContainerLabel()

	hashStr := dockerClient.getDockerContainerHash()

	dockerClient.build(AppFileDir, containerLabel, hashStr)
}

// Run starts an app in a container.
// If the container does not exists it is built upfront.
func Run(yamlAppConfigFileName string, args []string) {

	appInfo := appinfo.NewAppInfo(yamlAppConfigFileName)
	dockerClient := NewDockerClient(appInfo)

	dockerClient.run(args)
}
