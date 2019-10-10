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
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/docker/cli/cli/command"
	cmd_container "github.com/docker/cli/cli/command/container"
	cmd_build "github.com/docker/cli/cli/command/image"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/term"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/tjeske/containerflight/util"
	"golang.org/x/net/context"
)

// "mock connectors" for unit-tesing
var filesystem = afero.NewOsFs()

type config interface {
	GetDockerfile() string
	GetAppFileDir() string
	GetResolvedAppConfig() string
	GetDockerRunArgs() []string
}

type dockerHttpApiClient interface {
	ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error)
	ImageRemove(ctx context.Context, imageID string, options types.ImageRemoveOptions) ([]types.ImageDeleteResponseItem, error)
}

type dockerCliClient interface {
	command.Cli
}

// DockerClient abstracts the containerflight communication with a moby daemon
type DockerClient struct {
	config    config
	client    dockerHttpApiClient
	dockerCli dockerCliClient
	version   string
}

var notWordChar = regexp.MustCompile("\\W")

// NewDockerClient creates a new Docker client using API 1.25 (implemented by Docker 1.13)
func NewDockerClient(config config, version string) *DockerClient {
	os.Setenv("DOCKER_API_VERSION", "1.25")

	// Docker HTTP API client
	var httpClient *http.Client
	client, err := client.NewClient(client.DefaultDockerHost, "1.30", httpClient, nil)
	util.CheckErr(err)

	// Docker cli client
	stdin, stdout, stderr := term.StdStreams()
	dockerCli := command.NewDockerCli(stdin, stdout, stderr)
	opts := cliflags.NewClientOptions()
	err = dockerCli.Initialize(opts)
	util.CheckErr(err)

	return &DockerClient{config: config, client: client, dockerCli: dockerCli, version: version}
}

// build a Docker container
func (dc *DockerClient) build(dockerBuildCtx string, label string, hashStr string) {

	// remove all previous images
	dc.removeImages(label)

	// create temporary Dockerfile
	tmpDockerFile := dc.createTempDockerFile(dockerBuildCtx, label)
	defer filesystem.Remove(tmpDockerFile.Name())

	cmdDockerRun := cmd_build.NewBuildCommand(dc.dockerCli)
	buildCmdArgs := dc.getBuildCmdArgs(tmpDockerFile.Name(), dockerBuildCtx, label, hashStr)
	cmdDockerRun.SetArgs(buildCmdArgs)
	cmdDockerRun.SilenceErrors = true
	cmdDockerRun.SilenceUsage = true

	log.Debug("execute \"docker build " + strings.Join(buildCmdArgs, " ") + "\"")

	err := cmdDockerRun.Execute()
	util.CheckErr(err)

	tmpDockerFile.Close()
	util.CheckErr(err)
}

// removeImages destroys all Docker images with the specific label
func (dc *DockerClient) removeImages(label string) {
	client := dc.client
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

// create and populate temporary Dockerfile
func (dc *DockerClient) createTempDockerFile(dockerBuildCtx string, label string) afero.File {
	tmpDockerFile, err := afero.TempFile(filesystem, dockerBuildCtx, label+"_")
	util.CheckErr(err)

	dockerfileContent := dc.config.GetDockerfile()
	_, err = tmpDockerFile.Write([]byte(dockerfileContent))
	util.CheckErr(err)

	return tmpDockerFile
}

// get Docker build command args
func (dc *DockerClient) getBuildCmdArgs(dockerfile string, dockerBuildCtx string, label string, hashStr string) []string {
	buildCmd := []string{
		dockerBuildCtx,
		"-f", dockerfile,
		"--label", "containerflight_hash=" + hashStr,
		"-t", label,
	}

	return buildCmd
}

// Run a Docker container
func (dc *DockerClient) Run(args []string) {
	cmdDockerRun := cmd_container.NewRunCommand(dc.dockerCli)

	config := dc.config
	dockerRunArgs := config.GetDockerRunArgs()

	imageID := dc.getImageID()

	dockerRunCmdArgs := dc.getRunCmdArgs(dockerRunArgs, imageID, args)
	cmdDockerRun.SetArgs(dockerRunCmdArgs)
	cmdDockerRun.SilenceErrors = true
	cmdDockerRun.SilenceUsage = true

	log.Debug("execute \"docker run " + strings.Join(dockerRunCmdArgs, " ") + "\"")

	err := cmdDockerRun.Execute()
	util.CheckErr(err)
}

func (dc *DockerClient) Run2(dockerRunArgs []string, args []string) {
	cmdDockerRun := cmd_container.NewRunCommand(dc.dockerCli)

	imageID := dc.getImageID()

	dockerRunCmdArgs := dc.getRunCmdArgs(dockerRunArgs, imageID, args)
	cmdDockerRun.SetArgs(dockerRunCmdArgs)
	cmdDockerRun.SilenceErrors = true
	cmdDockerRun.SilenceUsage = true

	log.Debug("execute \"docker run " + strings.Join(dockerRunCmdArgs, " ") + "\"")

	err := cmdDockerRun.Execute()
	util.CheckErr(err)
}

// return Docker image Id, if image does not exists build it
func (dc *DockerClient) getImageID() string {

	config := dc.config

	AppFileDir := config.GetAppFileDir()
	containerLabel := dc.getDockerContainerLabel()
	hashStr := dc.getDockerContainerHash()

	imageID, err := dc.getDockerContainerImageID(hashStr)
	if err != nil {
		dc.build(AppFileDir, containerLabel, hashStr)
		imageID, err = dc.getDockerContainerImageID(hashStr)
		util.CheckErr(err)
	}
	return imageID
}

// get Docker run command args
func (dc *DockerClient) getRunCmdArgs(dockerRunArgs []string, imageID string, args []string) []string {
	runCmdArgs := []string{
		"--rm",
	}

	runCmdArgs = append(runCmdArgs, dockerRunArgs...)
	runCmdArgs = append(runCmdArgs, imageID)
	runCmdArgs = append(runCmdArgs, args...)

	return runCmdArgs
}

// getDockerContainerImageID returns the Docker image ID for an app hash value
func (dc *DockerClient) getDockerContainerImageID(hashStr string) (string, error) {
	images, err := dc.client.ImageList(context.Background(), types.ImageListOptions{})
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

// generate a container label
func (dc *DockerClient) getDockerContainerLabel() string {
	return "unknown"
}

// get the corresponding hash value for an app file
func (dc *DockerClient) getDockerContainerHash() string {

	appConfigStr := dc.config.GetResolvedAppConfig()

	hash := sha256.New()

	// hash config file
	appConfigBytes := []byte(appConfigStr)
	hash.Write(appConfigBytes)

	// hash containerflight version
	hash.Write([]byte(dc.version))

	// hash Docker build context if relevant
	dockerBuildCtx := dc.config.GetAppFileDir()
	if dc.isContextUsed() {
		afero.Walk(filesystem, dockerBuildCtx, func(fileName string, fi os.FileInfo, err error) error {

			// return on any error
			if err != nil {
				return err
			}

			// open files for hashing
			fh, err := filesystem.Open(fileName)
			if err != nil {
				return err
			}

			// hash file
			if _, err := io.Copy(hash, fh); err != nil {
				return err
			}

			fh.Close()

			return nil
		})
	}

	hashStr := hex.EncodeToString(hash.Sum(nil))
	return hashStr
}

// isContextUsed returns true if files from the Docker build context should be added to the Docker image
func (dc *DockerClient) isContextUsed() (isUsed bool) {
	dockerfileLines := strings.Split(dc.config.GetDockerfile(), "\n")
	isUsed = false
	for _, dockerfileLine := range dockerfileLines {
		linePreProcessed := strings.ToUpper(strings.TrimSpace(dockerfileLine))
		if strings.HasPrefix(linePreProcessed, "COPY ") || strings.HasPrefix(linePreProcessed, "ADD ") {
			isUsed = true
			break
		}
	}
	return isUsed
}
