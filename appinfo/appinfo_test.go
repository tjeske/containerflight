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

package appinfo

import (
	"fmt"
	"testing"

	yaml "github.com/go-yaml/yaml"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/tjeske/containerflight/util"
	"github.com/tjeske/containerflight/version"
)

func init() {
	// emulate file system
	filesystem = afero.NewMemMapFs()

	util.GetWorkingDir = func() string {
		return "/myworkingdir"
	}
}

func TestEmpty(t *testing.T) {
	testAppConfigAssert(t, "", "")
}

// ---

func TestResolvedAppConfig(t *testing.T) {
	appConfigStr := "description: home = ${HOME}"
	appInfo := NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)

	resolvedAppConfigStr := appInfo.GetResolvedAppConfig()

	// analyze the resolved app config file again
	appInfo2 := NewFakeAppInfo(&filesystem, "/testAppFile2", resolvedAppConfigStr)

	assert.Equal(t, fmt.Sprintf("%#v", ""), fmt.Sprintf("%#v", appInfo2.GetAppVersion()))
	assert.Equal(t, fmt.Sprintf("%#v", "home = /home"), fmt.Sprintf("%#v", appInfo2.GetAppDescription()))
}

// ---

func TestAppFileDir(t *testing.T) {
	filesystem.Mkdir("/appFileDir", 0755)

	appInfo := NewFakeAppInfo(&filesystem, "/appFileDir/testAppFile", "")

	appFileDir := appInfo.GetAppFileDir()

	assert.Equal(t, "/appFileDir", appFileDir)
}

func TestAppConfigFile(t *testing.T) {
	appInfo := NewFakeAppInfo(&filesystem, "/testAppFile", "")

	appConfigFile := appInfo.GetAppConfigFile()

	assert.Equal(t, "/testAppFile", appConfigFile)
}

// ---

func TestCompatibilityMatch(t *testing.T) {
	cfVersion := version.ContainerFlightVersion()
	appConfigStr := "compatibility: " + cfVersion.String()
	testAppConfigAssert(t, appConfigStr, appConfigStr)
}

func TestCompatibilityMustFail(t *testing.T) {
	testForLogFatal(t, func() {
		appConfigStr := "compatibility: 999.0.0"
		testAppConfig(t, appConfigStr, appConfigStr)
	})
}

// ---

func TestAppName(t *testing.T) {
	appInfo := NewFakeAppInfo(&filesystem, "/testAppFile", "name: myName")

	appName := appInfo.GetAppName()

	assert.Equal(t, "myName", appName)
}

func TestAppNameNotSet(t *testing.T) {
	appInfo := NewFakeAppInfo(&filesystem, "/testAppFile", "")

	appName := appInfo.GetAppName()

	assert.Equal(t, "testAppFile", appName)
}

// ---

func TestAppVersion(t *testing.T) {
	appInfo := NewFakeAppInfo(&filesystem, "/testAppFile", "version: 0.1")

	appVersion := appInfo.GetAppVersion()

	assert.Equal(t, "0.1", appVersion)
}

// ---

func TestAppDescription(t *testing.T) {
	appInfo := NewFakeAppInfo(&filesystem, "/testAppFile", "description: some description")

	appDescription := appInfo.GetAppDescription()

	assert.Equal(t, "some description", appDescription)
}

// ---

func TestIsConsoleAppNotSet(t *testing.T) {
	appInfo := NewFakeAppInfo(&filesystem, "/testAppFile", "")

	assert.Equal(t, true, appInfo.IsConsoleApp())
}

func TestIsConsoleAppSetFalse(t *testing.T) {
	appInfo := NewFakeAppInfo(&filesystem, "/testAppFile", "console: false")

	assert.Equal(t, false, appInfo.IsConsoleApp())
}

func TestIsConsoleAppSetTrue(t *testing.T) {
	appInfo := NewFakeAppInfo(&filesystem, "/testAppFile", "console: true")

	assert.Equal(t, true, appInfo.IsConsoleApp())
}

// ---

func TestDockerfileBasic(t *testing.T) {

	appConfigStr :=
		"image:\n" +
			"    dockerfile: |\n" +
			"        ${APP_FILE_DIR}\n" +
			"        ${USERNAME}\n" +
			"        ${USERID}\n" +
			"        ${GROUPNAME}\n" +
			"        ${GROUPID}\n" +
			"        ${HOME}\n" +
			"        ${PWD}\n"

	expDockerfile := fmt.Sprintf(dockerFileTmpl,
		"/\n"+
			"testuser\n"+
			"1234\n"+
			"testgroup\n"+
			"5678\n"+
			"/home\n"+
			"/myworkingdir\n")

	appInfo := NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)
	assert.Equal(t, expDockerfile, appInfo.GetDockerfile())
}

func TestDockerfileApt(t *testing.T) {

	appConfigStr :=
		"image:\n" +
			"    dockerfile: |\n" +
			"        ${APT_INSTALL(pkg1, pkg2)}\n"

	expDockerfile := fmt.Sprintf(dockerFileTmpl,
		"RUN apt-get update && \\\n"+
			"    export DEBIAN_FRONTEND=noninteractive && \\\n"+
			"    apt-get install -y pkg1 pkg2 && \\\n"+
			"    rm -rf /var/lib/apt/lists/*\n")

	appInfo := NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)
	assert.Equal(t, expDockerfile, appInfo.GetDockerfile())
}

func TestDockerfileAdd(t *testing.T) {

	filesystem.Mkdir("/foo", 0755)
	afero.WriteFile(filesystem, "/foo/bar", []byte("Hello\nWorld!\n"), 0644)

	appConfigStr :=
		"image:\n" +
			"    dockerfile: |\n" +
			"        ${ADD(/foo/bar, /to)}\n"

	expDockerfile := fmt.Sprintf(dockerFileTmpl,
		"RUN echo 'Hello\\n\\\nWorld!\\n\\\n' > \"/to\"\n")

	appInfo := NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)
	assert.Equal(t, expDockerfile, appInfo.GetDockerfile())
}

func TestDockerfileFromFileAbsolute(t *testing.T) {
	filesystem.Mkdir("/foo", 0755)
	afero.WriteFile(filesystem, "/foo/Dockerfile", []byte("RUN script.sh"), 0644)

	appConfigStr :=
		"image:\n" +
			"    dockerfile: file:///foo/Dockerfile"

	expDockerfile := fmt.Sprintf(dockerFileTmpl, "RUN script.sh")

	appInfo := NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)
	assert.Equal(t, expDockerfile, appInfo.GetDockerfile())
}

func TestDockerfileFromFileRelative(t *testing.T) {
	filesystem.Mkdir("foo", 0755)
	afero.WriteFile(filesystem, "/foo/Dockerfile", []byte("RUN script.sh"), 0644)

	appConfigStr :=
		"image:\n" +
			"    dockerfile: file://../foo/Dockerfile"

	expDockerfile := fmt.Sprintf(dockerFileTmpl, "RUN script.sh")

	appInfo := NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)
	assert.Equal(t, expDockerfile, appInfo.GetDockerfile())
}

func TestDockerfileFromFileNotFound(t *testing.T) {
	testForLogFatal(t, func() {
		appConfigStr :=
			"image:\n" +
				"    dockerfile: file://notthere/Dockerfile"

		appInfo := NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)
		appInfo.GetDockerfile()
	})
}
func TestDockerfileError(t *testing.T) {

	appConfigStr :=
		"image:\n" +
			"    dockerfile: |\n" +
			"        ${UNKNOWN_KEY}\n"

	expDockerfile := fmt.Sprintf(dockerFileTmpl, "<<ERROR!>>\n")

	appInfo := NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)
	assert.Equal(t, expDockerfile, appInfo.GetDockerfile())
}

// ---

func TestDockerRunArgsEmpty(t *testing.T) {

	appConfigStr := ""

	expDockerRunArgs := []string{"-v", "/myworkingdir:/myworkingdir", "-ti", "-h", "flybydocker", "-w", "/myworkingdir"}

	appInfo := NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)
	assert.Equal(t, expDockerRunArgs, appInfo.GetDockerRunArgs())
}

func TestDockerRunArgsSetWorkingDir(t *testing.T) {

	appConfigStr :=
		"runtime:\n" +
			"    docker:\n" +
			"        runargs: [\"-w\", \"/newworkingdir\"]"

	expDockerRunArgs := []string{"-v", "/myworkingdir:/myworkingdir", "-w", "/newworkingdir", "-ti", "-h", "flybydocker"}

	appInfo := NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)
	assert.Equal(t, expDockerRunArgs, appInfo.GetDockerRunArgs())
}

func TestDockerRunArgsSetHostname(t *testing.T) {

	appConfigStr :=
		"runtime:\n" +
			"    docker:\n" +
			"        runargs: [\"-h\", \"myhostname\"]"

	expDockerRunArgs := []string{"-v", "/myworkingdir:/myworkingdir", "-h", "myhostname", "-ti", "-w", "/myworkingdir"}

	appInfo := NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)
	assert.Equal(t, expDockerRunArgs, appInfo.GetDockerRunArgs())
}

func TestDockerRunArgsConsole(t *testing.T) {

	appConfigStr := "console: true"

	expDockerRunArgs := []string{"-v", "/myworkingdir:/myworkingdir", "-ti", "-h", "flybydocker", "-w", "/myworkingdir"}

	appInfo := NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)
	assert.Equal(t, expDockerRunArgs, appInfo.GetDockerRunArgs())
}

func TestDockerRunArgsGui(t *testing.T) {

	appConfigStr := "gui: true"

	expDockerRunArgs := []string{"-v", "/myworkingdir:/myworkingdir", "-ti", "-e", "DISPLAY=DISPLAY", "-v", "/tmp/.X11-unix:/tmp/.X11-unix", "-h", "flybydocker", "-w", "/myworkingdir"}

	appInfo := NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)
	assert.Equal(t, expDockerRunArgs, appInfo.GetDockerRunArgs())
}

// ---

var dockerFileTmpl = `ENV http_proxy=http_proxy
ENV https_proxy=https_proxy
ENV no_proxy=no_proxy

%s
RUN if ! getent group testgroup > /dev/null 2>&1; then \
        ( \
            # ubuntu\
            addgroup -g 5678 testgroup || \
            # busybox\
            addgroup --gid 5678 testgroup || \
            # fedora / arch linux\
            groupadd --gid 5678 testgroup \
        ) > /dev/null 2>&1 ; \
    fi ; \
    if ! getent passwd testuser > /dev/null 2>&1; then \
        ( \
            # fedora\
            adduser --gid testgroup --uid 1234 --base-dir "/home" testuser || \
            # ubuntu\
            adduser --home "/home" --uid 1234 --gecos "" --ingroup testgroup --disabled-password testuser || \
            # busybox\
            adduser -h "/home" -u 1234 -D -H -G testgroup testuser || \
            # arch linux\
            useradd --no-user-group --gid 5678 --uid 1234 --home-dir "/home" --create-home testuser \
        ) > /dev/null 2>&1 ; \
    fi ;

USER testuser`

func testAppConfigAssert(t *testing.T, expAppConfigStr string, appConfigStr string) {
	expAppConfig, appConfig := testAppConfig(t, expAppConfigStr, appConfigStr)

	assert.Equal(t, fmt.Sprintf("%#v", expAppConfig), fmt.Sprintf("%#v", appConfig))
}

func testAppConfig(t *testing.T, expAppConfigStr string, appConfigStr string) (yamlSpec, yamlSpec) {
	expAppConfig := yamlSpec{}
	err := yaml.UnmarshalStrict([]byte(expAppConfigStr), &expAppConfig)
	util.CheckErr(err)

	appInfo := NewFakeAppInfo(&filesystem, "/testAppFile", appConfigStr)

	return expAppConfig, appInfo.appConfig
}

func testForLogFatal(t *testing.T, testFunc func()) {

	origLogFatalf := logFatalf
	defer func() { logFatalf = origLogFatalf }()

	numErrors := 0
	logFatalf = func(format string, args ...interface{}) {
		numErrors++
	}

	testFunc()

	if numErrors != 1 {
		t.Errorf("excepted one error, actual %v", numErrors)
	}
}
