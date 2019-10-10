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
	"strings"

	"github.com/tjeske/containerflight/version"
)

// ContainerFlightDockerClient abstracts the containerflight communication with a moby daemon
type ContainerFlightDockerClient struct {
	*DockerClient
}

type containerFlightConfig interface {
	config
	GetAppConfigFile() string
	GetAppVersion() string
	GetAppName() string
	GetAppDescription() string
}

// NewContainerFlightDockerClient creates a new Docker client using API 1.25 (implemented by Docker 1.13)
func NewContainerFlightDockerClient(config containerFlightConfig) *ContainerFlightDockerClient {

	containerflightVersion := version.ContainerFlightVersion().String()

	return &ContainerFlightDockerClient{DockerClient: NewDockerClient(config, containerflightVersion)}
}

// generate a container label
func (dc *ContainerFlightDockerClient) getDockerContainerLabel() string {
	appNameNormalized := notWordChar.ReplaceAllString(dc.config.(containerFlightConfig).GetAppName(), "")
	if appNameNormalized == "" {
		appNameNormalized = "unknown"
	}
	label := "containerflight_" + strings.ToLower(appNameNormalized) + ":"
	appConfigVersion := dc.config.(containerFlightConfig).GetAppVersion()
	if appConfigVersion != "" {
		label += appConfigVersion
	} else {
		label += "unknown"
	}
	return label
}

// get Docker build command args
func (dc *ContainerFlightDockerClient) getBuildCmdArgs(dockerfile string, dockerBuildCtx string, label string, hashStr string) []string {
	description := dc.config.(containerFlightConfig).GetAppDescription()

	buildCmd := []string{
		dockerBuildCtx,
		"-f", dockerfile,
		"--label", "containerflight=true",
		"--label", "containerflight_appFile=" + dc.config.(containerFlightConfig).GetAppConfigFile(),
		"--label", "containerflight_hash=" + hashStr,
		"--label", "containerflight_cfVersion=" + dc.version,
		"--label", "containerflight_description=" + description,
		"-t", label,
	}

	return buildCmd
}

func (dc *ContainerFlightDockerClient) getRunCmdArgs(imageID string, args []string) []string {

	config := dc.config

	appConfigFile := config.(containerFlightConfig).GetAppConfigFile()
	dockerRunArgs := config.(containerFlightConfig).GetDockerRunArgs()
	containerLabel := dc.getDockerContainerLabel()
	hashStr := dc.getDockerContainerHash()

	runCmdArgs := []string{
		"--rm",
		"--label", "containerflight_appFile=" + appConfigFile,
		"--label", "containerflight_image=" + containerLabel,
		"--label", "containerflight_hash=" + hashStr,
		"--label", "containerflight_version=" + dc.version,
	}

	runCmdArgs = append(runCmdArgs, dockerRunArgs...)
	runCmdArgs = append(runCmdArgs, imageID)
	runCmdArgs = append(runCmdArgs, args...)

	return runCmdArgs
}
