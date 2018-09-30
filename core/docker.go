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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	cmd_container "github.com/docker/cli/cli/command/container"
	cmd_build "github.com/docker/cli/cli/command/image"
	cliflags "github.com/docker/cli/cli/flags"
	log "github.com/sirupsen/logrus"
	"github.com/tjeske/containerflight/appinfo"
	"github.com/tjeske/containerflight/util"
	"github.com/tjeske/containerflight/version"

	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/term"
)

// abstracts the containerflight communication with a moby daemon
type dockerClient struct {
	appInfo *appinfo.AppInfo
	client  *client.Client
}

// create a new Docker client using API 1.25 (implemented by Docker 1.13)
func newDockerClient(appInfo *appinfo.AppInfo) *dockerClient {
	os.Setenv("DOCKER_API_VERSION", "1.25")

	var httpClient *http.Client
	client, err := client.NewClient(client.DefaultDockerHost, "1.30", httpClient, nil)
	util.CheckErr(err)

	return &dockerClient{appInfo: appInfo, client: client}
}

// build a Docker container
func (dockerClient *dockerClient) build(dockerBuildCtx string, label string, hash string) {

	dockerClient.removeImages(label)

	stdin, stdout, stderr := term.StdStreams()
	// dockerFileStream := ioutil.NopCloser(strings.NewReader(dockerfileContent))
	dockerCli := command.NewDockerCli(stdin, stdout, stderr)

	opts := cliflags.NewClientOptions()
	err := dockerCli.Initialize(opts)
	util.CheckErr(err)

	tmpDockerFile, err := ioutil.TempFile(dockerBuildCtx, label+"_")
	util.CheckErr(err)

	defer os.Remove(tmpDockerFile.Name())

	dockerfileContent := dockerClient.appInfo.GetDockerfile()
	_, err = tmpDockerFile.Write([]byte(dockerfileContent))
	util.CheckErr(err)

	cfVersion := version.ContainerFlightVersion().String()

	buildCmd := []string{
		dockerBuildCtx,
		"-f", tmpDockerFile.Name(),
		"--label", "containerflight=true",
		"--label", "containerflight_appFile=" + dockerClient.appInfo.GetAppConfigFile(),
		"--label", "containerflight_hash=" + hash,
		"--label", "containerflight_version=" + cfVersion,
		"-t", label,
	}

	cmdDockerRun := cmd_build.NewBuildCommand(dockerCli)
	cmdDockerRun.SetArgs(buildCmd)
	cmdDockerRun.SilenceErrors = true
	cmdDockerRun.SilenceUsage = true

	log.Debug("execute \"docker build " + strings.Join(buildCmd, " ") + "\"")

	err = cmdDockerRun.Execute()
	util.CheckErr(err)

	tmpDockerFile.Close()
	util.CheckErr(err)
}

// run a Docker container
func (dockerClient *dockerClient) run(args []string) {
	dockerRunArgs := dockerClient.getDockerRunCmdArgs(args)

	stdin, stdout, stderr := term.StdStreams()
	dockerCli := command.NewDockerCli(stdin, stdout, stderr)

	opts := cliflags.NewClientOptions()
	err := dockerCli.Initialize(opts)
	util.CheckErr(err)

	cmdDockerRun := cmd_container.NewRunCommand(dockerCli)
	cmdDockerRun.SetArgs(dockerRunArgs)
	cmdDockerRun.SilenceErrors = true
	cmdDockerRun.SilenceUsage = true

	log.Debug("execute \"docker run " + strings.Join(dockerRunArgs, " ") + "\"")

	err = cmdDockerRun.Execute()
	util.CheckErr(err)
}

func (dockerClient *dockerClient) getDockerRunCmdArgs(args []string) []string {

	appInfo := dockerClient.appInfo

	appConfigFile := appInfo.GetAppConfigFile()
	dockerRunArgs := appInfo.GetDockerRunArgs()
	AppFileDir := appInfo.GetAppFileDir()
	containerLabel := dockerClient.getDockerContainerLabel()
	hashStr := dockerClient.getDockerContainerHash()

	imageID, err := dockerClient.getDockerContainerImageID(hashStr)
	if err != nil {
		dockerClient.build(AppFileDir, containerLabel, hashStr)
		imageID, err = dockerClient.getDockerContainerImageID(hashStr)
		util.CheckErr(err)
	}

	cfVersion := version.ContainerFlightVersion().String()

	runCmdArgs := []string{
		"--rm",
		"--label", "containerflight_appFile=" + appConfigFile,
		"--label", "containerflight_image=" + containerLabel,
		"--label", "containerflight_hash=" + hashStr,
		"--label", "containerflight_version=" + cfVersion,
	}

	runCmdArgs = append(runCmdArgs, dockerRunArgs...)
	runCmdArgs = append(runCmdArgs, imageID)
	if len(args) > 1 {
		runCmdArgs = append(runCmdArgs, args[1:]...)
	}

	return runCmdArgs
}

// getDockerContainerImageID returns the Docker image ID for an app hash value
func (dockerClient *dockerClient) getDockerContainerImageID(hashStr string) (string, error) {
	images, err := dockerClient.client.ImageList(context.Background(), types.ImageListOptions{})
	util.CheckErr(err)
	imageID := ""
	for _, image := range images {
		imgHash := image.Labels["containerflight_hash"]
		if hashStr == imgHash {
			imageID = image.ID
			break
		}
	}
	if imageID != "" {
		return imageID, nil
	}
	return "", fmt.Errorf("cannot find image with ID `%s`", hashStr)
}

// removeImages destroys all Docker images with the specific label
func (dockerClient *dockerClient) removeImages(label string) {
	client := dockerClient.client
	images, err := client.ImageList(context.Background(), types.ImageListOptions{})
	util.CheckErr(err)

	for _, image := range images {
		tagFound := false
		for _, tag := range image.RepoTags {
			if tag == label {
				tagFound = true
				break
			}
		}
		if tagFound {
			// remove image
			options := types.ImageRemoveOptions{Force: true, PruneChildren: true}
			client.ImageRemove(context.Background(), image.ID, options)
			util.CheckErr(err)
		}
	}
}

// generate a container label
func (dockerClient *dockerClient) getDockerContainerLabel() string {

	label := "containerflight_" + filepath.Base(dockerClient.appInfo.GetAppConfigFile()) + ":"
	appConfigVersion := dockerClient.appInfo.GetVersion()
	if appConfigVersion != "" {
		label += appConfigVersion
	} else {
		label += "unknown"
	}
	return label
}

// get the corresponding hash value for an app file
func (dockerClient *dockerClient) getDockerContainerHash() string {

	appConfigStr := dockerClient.appInfo.GetResolvedAppConfig()
	cfVersion := version.ContainerFlightVersion().String()

	appConfigBytes := []byte(appConfigStr)
	// hash config file
	hash := sha256.New()
	hash.Write(appConfigBytes)
	hash.Write([]byte(cfVersion))
	hashStr := hex.EncodeToString(hash.Sum(nil))
	return hashStr
}
