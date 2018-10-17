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
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/tjeske/containerflight/appinfo"
	"github.com/tjeske/containerflight/util"
	"golang.org/x/net/context"
	"io"
	"regexp"
	"testing"
)

type mockHttpApiClient struct {
	imageRepo []types.ImageSummary
}

func init() {
	// emulate file system
	filesystem = afero.NewMemMapFs()

	// fake version number to have fixed hash values
	containerflightVersion = "x.y.z"
}

func newMockHttpApiClient() *mockHttpApiClient {
	imageRepo := []types.ImageSummary{
		{
			Containers: -1,
			Created:    -1,
			ID:         "sha256:123",
			Labels: map[string]string{
				"containerflight_hash": "123",
			},
			ParentID:    "sha256:1234",
			RepoDigests: []string{},
			RepoTags: []string{
				"containerflight_donotremove:testingversion",
			},
		},
		{
			Containers: -1,
			Created:    -1,
			ID:         "sha256:456",
			Labels: map[string]string{
				"containerflight_hash": "456",
			},
			ParentID:    "sha256:4567",
			RepoDigests: []string{},
			RepoTags: []string{
				"containerflight_testing:testingversion",
			},
		},
		{
			Containers: -1,
			Created:    -1,
			ID:         "sha256:789",
			Labels: map[string]string{
				"containerflight_hash": "789",
			},
			ParentID:    "sha256:7890",
			RepoDigests: []string{},
			RepoTags: []string{
				"containerflight_testing:testingversion",
			},
		},
	}

	return &mockHttpApiClient{imageRepo}
}

func (c *mockHttpApiClient) ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error) {
	return c.imageRepo, nil
}

func (c *mockHttpApiClient) ImageRemove(ctx context.Context, imageID string, options types.ImageRemoveOptions) ([]types.ImageDeleteResponseItem, error) {
	respItems := []types.ImageDeleteResponseItem{}
	for i, el := range c.imageRepo {
		if el.ID == imageID {
			c.imageRepo = append(c.imageRepo[:i], c.imageRepo[i+1:]...)
			respItems = append(respItems, types.ImageDeleteResponseItem{Deleted: el.ID})
		}
	}

	return respItems, nil
}

func newDockerClient(appInfo *appinfo.AppInfo) *DockerClient {

	// Docker HTTP API client
	client := newMockHttpApiClient()

	// Docker cli client
	var dockerCli dockerCliClient

	return &DockerClient{appInfo: appInfo, client: client, dockerCli: dockerCli}
}
func TestRemoveImages(t *testing.T) {
	dockerClient := newDockerClient(&appinfo.AppInfo{})
	dockerClient.removeImages("containerflight_testing:testingversion")

	httpApiClient := dockerClient.client.(*mockHttpApiClient)

	assert.Equal(t, 1, len(httpApiClient.imageRepo))
	assert.Equal(t, "containerflight_donotremove:testingversion", httpApiClient.imageRepo[0].RepoTags[0])
}

func TestRemoveImagesUnknown(t *testing.T) {
	dockerClient := newDockerClient(&appinfo.AppInfo{})
	dockerClient.removeImages("containerflight_testingUnknown:testingversion")

	httpApiClient := dockerClient.client.(*mockHttpApiClient)

	// nothing gets deleted
	assert.Equal(t, 3, len(httpApiClient.imageRepo))
}

func TestCreateTempDockerFile(t *testing.T) {
	appConfigStr := "image:\n    dockerfile: |\n        RUN test"
	appInfo := appinfo.NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)

	dockerClient := newDockerClient(appInfo)
	tmpDockerFile := dockerClient.createTempDockerFile(".", "testlabel")

	tmpDockerFile.Seek(0, io.SeekStart)
	rawData, err := afero.ReadAll(tmpDockerFile)
	util.CheckErr(err)

	assert.Regexp(t, regexp.MustCompile("RUN test"), string(rawData))
}

func TestGetBuildCmdArgs(t *testing.T) {
	appConfigStr := "image:\n    dockerfile: |\n        RUN test"
	appInfo := appinfo.NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)

	dockerClient := newDockerClient(appInfo)
	args := dockerClient.getBuildCmdArgs("dockerfile", "dockerBuildCtx", "label", "hashStr")

	expArgs := []string{
		"dockerBuildCtx",
		"-f", "dockerfile",
		"--label", "containerflight=true",
		"--label", "containerflight_appFile=/testAppFile",
		"--label", "containerflight_hash=hashStr",
		"--label", "containerflight_cfVersion=" + containerflightVersion,
		"--label", "containerflight_description=",
		"-t", "label",
	}
	assert.Equal(t, expArgs, args)
}

func TestGetRunCmdArgs(t *testing.T) {
	appConfigStr := ""
	appInfo := appinfo.NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)

	dockerClient := newDockerClient(appInfo)
	args := dockerClient.getRunCmdArgs("123", []string{"arg1", "arg2"})

	expArgs := []string{
		"--rm",
		"--label", "containerflight_appFile=/testAppFile",
		"--label", "containerflight_image=containerflight_testAppFile:unknown",
		"--label", "containerflight_hash=562a792d764ddceb355634b2ccee3878edf696021767ff0e8144eab2e2bf035f",
		"--label", "containerflight_version=" + containerflightVersion,
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

func TestGetDockerContainerImageID(t *testing.T) {
	appConfigStr := ""
	appInfo := appinfo.NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)

	dockerClient := newDockerClient(appInfo)
	imageID, err := dockerClient.getDockerContainerImageID("456")

	assert.Equal(t, "sha256:456", imageID)
	assert.Equal(t, nil, err)
}

func TestGetDockerContainerImageIDNotFound(t *testing.T) {
	appConfigStr := ""
	appInfo := appinfo.NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)

	dockerClient := newDockerClient(appInfo)
	imageID, err := dockerClient.getDockerContainerImageID("notfound")

	assert.Equal(t, "", imageID)
	assert.Equal(t, errors.New("cannot find image with ID `notfound`"), err)
}

func TestGetDockerContainerLabel(t *testing.T) {
	appConfigStr := ""
	appInfo := appinfo.NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)

	dockerClient := newDockerClient(appInfo)
	label := dockerClient.getDockerContainerLabel()

	assert.Equal(t, "containerflight_testAppFile:unknown", label)
}

func TestGetDockerContainerLabelVersion(t *testing.T) {
	appConfigStr := "version: \"1.2.3\""
	appInfo := appinfo.NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)

	dockerClient := newDockerClient(appInfo)
	label := dockerClient.getDockerContainerLabel()

	assert.Equal(t, "containerflight_testAppFile:1.2.3", label)
}

func TestGetDockerContainerHash(t *testing.T) {
	appConfigStr := ""
	appInfo := appinfo.NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)

	dockerClient := newDockerClient(appInfo)
	hashStr := dockerClient.getDockerContainerHash()

	assert.Equal(t, "562a792d764ddceb355634b2ccee3878edf696021767ff0e8144eab2e2bf035f", hashStr)
}

func TestGetDockerContainerHashWithContext(t *testing.T) {
	afero.WriteFile(filesystem, "/foo.bar", []byte("some data"), 0644)

	appConfigStr := "image:\n    dockerfile: |\n        COPY foo.bar"
	appInfo := appinfo.NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)

	dockerClient := newDockerClient(appInfo)
	hashStr := dockerClient.getDockerContainerHash()

	assert.Equal(t, "c90e2a76c380fae4b63ec88566a327637cfd6fc3f26f88cdc0137961b02d10d9", hashStr)
}
