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
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	cmd_container "github.com/docker/cli/cli/command/container"
	cmd_build "github.com/docker/cli/cli/command/image"
	cliflags "github.com/docker/cli/cli/flags"
	log "github.com/sirupsen/logrus"

	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/pkg/term"
)

// generate a container label
func getDockerContainerLabel(appConfigFileName string, dockerFile string) string {

	dockerFileAsBytes := []byte(dockerFile)

	// hash config file
	hash := sha256.New()
	hash.Write(dockerFileAsBytes)

	label := "containerflight_" + filepath.Base(appConfigFileName) + "_" + hex.EncodeToString(hash.Sum(nil))
	return label
}

// abstracts the Container Flight communication with a moby daemon
type dockerClient struct {
	containerLabel string
}

// create a new Docker client using API 1.25 (implemented by Docker 1.13)
func newDockerClient(containerLabel string) *dockerClient {
	os.Setenv("DOCKER_API_VERSION", "1.25")
	return &dockerClient{containerLabel: containerLabel}
}

// build a Docker container
func (dockerClient *dockerClient) build(dockerfileContent *string) {

	_, stdout, stderr := term.StdStreams()
	dockerFileStream := ioutil.NopCloser(strings.NewReader(*dockerfileContent))
	dockerCli := command.NewDockerCli(dockerFileStream, stdout, stderr)

	opts := cliflags.NewClientOptions()
	err := dockerCli.Initialize(opts)
	checkErr(err)

	buildCmd := []string{"-", "-t", dockerClient.containerLabel}

	cmdDockerRun := cmd_build.NewBuildCommand(dockerCli)
	cmdDockerRun.SetArgs(buildCmd)
	cmdDockerRun.SilenceErrors = true
	cmdDockerRun.SilenceUsage = true

	log.Debug("execute \"docker build " + strings.Join(buildCmd, " ") + "\"")

	err = cmdDockerRun.Execute()
	checkErr(err)
}

// run a Docker container
func (dockerClient *dockerClient) run(runCmd *[]string) {
	stdin, stdout, stderr := term.StdStreams()
	dockerCli := command.NewDockerCli(stdin, stdout, stderr)

	opts := cliflags.NewClientOptions()
	err := dockerCli.Initialize(opts)
	checkErr(err)

	cmdDockerRun := cmd_container.NewRunCommand(dockerCli)
	cmdDockerRun.SetArgs(*runCmd)
	cmdDockerRun.SilenceErrors = true
	cmdDockerRun.SilenceUsage = true

	log.Debug("execute \"docker run " + strings.Join(*runCmd, " ") + "\"")

	err = cmdDockerRun.Execute()
	checkErr(err)
}
