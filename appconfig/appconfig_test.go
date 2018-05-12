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

package appconfig

import (
	"fmt"
	"strings"
	"testing"

	"github.com/tjeske/containerflight/version"

	"github.com/stretchr/testify/assert"
	"github.com/tjeske/containerflight/util"

	yaml "github.com/go-yaml/yaml"
)

// fake environment
var env = environment{
	userName:   "testuser",
	userID:     "1234",
	groupName:  "testgroup",
	groupID:    "5678",
	homeDir:    "/home",
	workingDir: "/myworkingdir",
}

func TestEmpty(t *testing.T) {
	testAppConfigAssert(t, "", "")
}

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

func TestDockerfileBasic(t *testing.T) {

	appConfigStr :=
		"image:\n" +
			"    dockerfile: |\n" +
			"        ${USERNAME}\n" +
			"        ${USERID}\n" +
			"        ${GROUPNAME}\n" +
			"        ${GROUPID}\n" +
			"        ${HOME}\n" +
			"        ${PWD}\n"

	expDockerfile := fmt.Sprintf(dockerFileTmpl,
		"testuser\n"+
			"1234\n"+
			"testgroup\n"+
			"5678\n"+
			"/home\n"+
			"/myworkingdir\n")

	testDockerfile(t, appConfigStr, expDockerfile)
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

	testDockerfile(t, appConfigStr, expDockerfile)
}

func TestDockerfileError(t *testing.T) {

	appConfigStr :=
		"image:\n" +
			"    dockerfile: |\n" +
			"        ${UNKNOWN_KEY}\n"

	expDockerfile := fmt.Sprintf(dockerFileTmpl, "<<ERROR!>>\n")

	testDockerfile(t, appConfigStr, expDockerfile)
}

// ---

func TestDockerRunArgsEmpty(t *testing.T) {

	appConfigStr := ""

	expDockerRunArgs := []string{"-v", "/myworkingdir:/myworkingdir", "-h", "flybydocker", "-w", "/myworkingdir"}

	testDockerRunArgs(t, appConfigStr, expDockerRunArgs)
}

func TestDockerRunArgsSetWorkingDir(t *testing.T) {

	appConfigStr :=
		"runtime:\n" +
			"    docker:\n" +
			"        runargs: [\"-w\", \"/newworkingdir\"]"

	expDockerRunArgs := []string{"-v", "/myworkingdir:/myworkingdir", "-w", "/newworkingdir", "-h", "flybydocker"}

	testDockerRunArgs(t, appConfigStr, expDockerRunArgs)
}

func TestDockerRunArgsSetHostname(t *testing.T) {

	appConfigStr :=
		"runtime:\n" +
			"    docker:\n" +
			"        runargs: [\"-h\", \"myhostname\"]"

	expDockerRunArgs := []string{"-v", "/myworkingdir:/myworkingdir", "-h", "myhostname", "-w", "/myworkingdir"}

	testDockerRunArgs(t, appConfigStr, expDockerRunArgs)
}

func TestDockerRunArgsConsole(t *testing.T) {

	appConfigStr := "console: true"

	expDockerRunArgs := []string{"-v", "/myworkingdir:/myworkingdir", "-ti", "-h", "flybydocker", "-w", "/myworkingdir"}

	testDockerRunArgs(t, appConfigStr, expDockerRunArgs)
}

func TestDockerRunArgsGui(t *testing.T) {

	appConfigStr := "gui: true"

	expDockerRunArgs := []string{"-v", "/myworkingdir:/myworkingdir", "-e", "DISPLAY=DISPLAY", "-v", "/tmp/.X11-unix:/tmp/.X11-unix", "-h", "flybydocker", "-w", "/myworkingdir"}

	testDockerRunArgs(t, appConfigStr, expDockerRunArgs)
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
            # fedora\
            groupadd --gid 5678 testgroup \
        ) > /dev/null 2>&1 ; \
    fi ; \
    if ! getent passwd testuser > /dev/null 2>&1; then \
        ( \
            # fedora\
            adduser --gid testgroup --uid 1234 testuser || \
            # ubuntu\
            adduser --uid 1234 --gecos "" --ingroup testgroup --disabled-password testuser || \
            # busybox\
            adduser -u 1234 -D -H -G testgroup testuser \
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

	yamlAppConfigReader := strings.NewReader(appConfigStr)
	appInfo := newAppInfo(yamlAppConfigReader, env)

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

func testDockerfile(t *testing.T, appConfigStr string, expDockerfile string) {

	origGetEnvVar := getEnvVar
	defer func() { getEnvVar = origGetEnvVar }()

	getEnvVar = func(name string) string {
		return name
	}

	yamlAppConfigReader := strings.NewReader(appConfigStr)
	appInfo := newAppInfo(yamlAppConfigReader, env)

	assert.Equal(t, expDockerfile, appInfo.GetDockerfile())
}

func testDockerRunArgs(t *testing.T, appConfigStr string, expDockerRunArgs []string) {

	origGetEnvVar := getEnvVar
	defer func() { getEnvVar = origGetEnvVar }()

	getEnvVar = func(name string) string {
		return name
	}

	yamlAppConfigReader := strings.NewReader(appConfigStr)
	appInfo := newAppInfo(yamlAppConfigReader, env)

	assert.Equal(t, expDockerRunArgs, appInfo.GetDockerRunArgs())
}
