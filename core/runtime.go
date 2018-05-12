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
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/tjeske/containerflight/appconfig"
	"github.com/tjeske/containerflight/util"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// PrintDockerRunArgs show the resulting "docker run" arguments
func PrintDockerRunArgs(yamlAppConfigFileName string) {

	appInfo := appconfig.NewAppInfo(yamlAppConfigFileName)
	dockerfile := appInfo.GetDockerfile()
	dockerRunArgs := appInfo.GetDockerRunArgs()

	containerLabel := getDockerContainerLabel(yamlAppConfigFileName, dockerfile)
	dockerRunCmdArgs := getDockerRunCmdArgs(dockerRunArgs, containerLabel, yamlAppConfigFileName, []string{})

	fmt.Println("\"docker run\" will be called with the following arguments:\n" + strings.Join(dockerRunCmdArgs, " "))
}

// Run starts an app in a container.
// If the container does not exists it is built upfront.
func Run(args []string) {

	yamlAppConfigFileName := args[0]

	appInfo := appconfig.NewAppInfo(yamlAppConfigFileName)
	dockerfile := appInfo.GetDockerfile()
	dockerRunArgs := appInfo.GetDockerRunArgs()

	var httpClient *http.Client
	cli, err := client.NewClient(client.DefaultDockerHost, "1.30", httpClient, nil)
	util.CheckErr(err)

	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	util.CheckErr(err)

	containerLabel := getDockerContainerLabel(yamlAppConfigFileName, dockerfile)
	fullContainerLabel := containerLabel + ":latest"

	found := false
	for _, image := range images {
		for _, repoTag := range image.RepoTags {
			if repoTag == fullContainerLabel {
				found = true
			}
		}
	}

	dockerClient := newDockerClient(containerLabel, &yamlAppConfigFileName)
	if !found {
		dockerClient.build(&dockerfile)
	}

	dockerRunCmdArgs := getDockerRunCmdArgs(dockerRunArgs, containerLabel, yamlAppConfigFileName, args)

	dockerClient.run(&dockerRunCmdArgs)
}

func getDockerRunCmdArgs(dockerRunArgs []string, containerLabel string, appFile string, args []string) []string {

	absAppFile, err := filepath.Abs(appFile)
	util.CheckErr(err)

	runCmdArgs := []string{
		"--rm",
		"--label", "image=" + containerLabel,
		"--label", "appFile=" + absAppFile,
	}

	runCmdArgs = append(runCmdArgs, dockerRunArgs...)
	runCmdArgs = append(runCmdArgs, containerLabel)
	if len(args) > 1 {
		runCmdArgs = append(runCmdArgs, args[1:]...)
	}

	return runCmdArgs
}
