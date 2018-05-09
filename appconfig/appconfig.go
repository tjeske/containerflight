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
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/blang/semver"
	yaml "github.com/go-yaml/yaml"
	log "github.com/sirupsen/logrus"
	"github.com/tjeske/containerflight/util"
	"github.com/tjeske/containerflight/version"
)

var logFatalf = log.Fatalf

// specification of an app file
type yamlSpec struct {
	Compatibility string
	Version       string
	Description   string
	Console       bool
	Gui           bool
	Image         struct {
		Base       string
		Dockerfile string
		Storage    struct {
			Driver string
		}
	}
	Runtime struct {
		Driver string
		Docker struct {
			RunArgs []string
		}
	}
}

// AppInfo represents an application config file
type AppInfo struct {
	appConfig      yamlSpec
	env            environment
	resolvedParams map[string]string
}

var parameterRegex = regexp.MustCompile("\\$\\{(.+?)\\}")

// NewAppInfo returns a representation of an application config file
func NewAppInfo(yamlAppConfigFileName string) *AppInfo {

	absYamlAppConfigFileName, err := filepath.Abs(yamlAppConfigFileName)
	util.CheckErr(err)

	yamlAppFileReader, err := os.Open(absYamlAppConfigFileName)
	util.CheckErr(err)

	env := getEnv()

	appConfig := newAppInfo(yamlAppFileReader, env)
	return appConfig
}

// newAppInfo returns a representation of an application config
// use this function for tests
func newAppInfo(yamlAppConfigReader io.Reader, env environment) *AppInfo {

	appConfig := getAppConfig(yamlAppConfigReader)

	resolvedParams := getResolvedParameters(env)

	validate(appConfig)

	return &AppInfo{
		appConfig:      appConfig,
		env:            env,
		resolvedParams: resolvedParams,
	}
}

func validate(appConfig yamlSpec) {
	// check version
	if appConfig.Compatibility != "" {
		cfVersion := version.ContainerFlightVersion()
		parsedRange, err := semver.ParseRange(appConfig.Compatibility)
		util.CheckErrMsg(err, "Version information must match semver 2.0.0 (https://semver.org/)!")
		if !parsedRange(cfVersion) {
			logFatalf("App file is not compatible with current Container Flight version %s!", cfVersion.String())
		}
	}
}

// read and parse app config
func getAppConfig(yamlAppConfigReader io.Reader) yamlSpec {

	// read the app file
	yamlFileBytes, err := ioutil.ReadAll(yamlAppConfigReader)
	util.CheckErr(err)
	str := string(yamlFileBytes)

	// unmarshal yaml file
	appConfig := yamlSpec{}
	err = yaml.UnmarshalStrict([]byte(str), &appConfig)
	util.CheckErr(err)

	return appConfig
}

// map the parameters which can be used in an app file to their corresponding values
func getResolvedParameters(env environment) map[string]string {
	return map[string]string{
		"USERNAME":  env.userName,
		"USERID":    env.userID,
		"GROUPNAME": env.groupName,
		"GROUPID":   env.groupID,
		"HOME":      env.homeDir,
		"PWD":       env.workingDir,
		"SET_PROXY": "ENV http_proxy=${ENV:http_proxy}\n" +
			"ENV https_proxy=${ENV:https_proxy}\n" +
			"ENV no_proxy=${ENV:no_proxy}\n",
		"USER_CTX": "RUN if ! getent group ${GROUPNAME} > /dev/null 2>&1; then \\\n" +
			"        ( \\\n" +
			"            # ubuntu\\\n" +
			"            addgroup -g ${GROUPID} ${GROUPNAME} || \\\n" +
			"            # busybox\\\n" +
			"            addgroup --gid ${GROUPID} ${GROUPNAME} || \\\n" +
			"            # fedora\\\n" +
			"            groupadd --gid ${GROUPID} ${GROUPNAME} \\\n" +
			"        ) > /dev/null 2>&1 ; \\\n" +
			"    fi ; \\\n" +
			"    if ! getent passwd ${USERNAME} > /dev/null 2>&1; then \\\n" +
			"        ( \\\n" +
			"            # fedora\\\n" +
			"            adduser --gid ${GROUPNAME} --uid ${USERID} ${USERNAME} || \\\n" +
			"            # ubuntu\\\n" +
			"            adduser --uid ${USERID} --gecos \"\" --ingroup ${GROUPNAME} --disabled-password ${USERNAME} || \\\n" +
			"            # busybox\\\n" +
			"            adduser -u ${USERID} -D -H -G ${GROUPNAME} ${USERNAME} \\\n" +
			"        ) > /dev/null 2>&1 ; \\\n" +
			"    fi ;\n\n" +
			"USER ${USERNAME}",
	}
}

var getEnvVar = func(name string) string {
	return os.Getenv(name)
}

// search and replace parameters in string
func (cfg *AppInfo) replaceParameters(str *string) {
	oldYamlFileStr := ""
	for *str != oldYamlFileStr {
		oldYamlFileStr = *str
		*str = parameterRegex.ReplaceAllStringFunc(*str, func(match string) string {
			trimmedMatch := match[2 : len(match)-1]
			split := strings.Split(trimmedMatch, ":")
			if len(split) == 1 {
				if value, ok := cfg.resolvedParams[split[0]]; ok {
					// ${KEY}
					return value
				}
			} else if len(split) > 1 {
				switch split[0] {
				case "ENV":
					{
						// ${ENV:...}
						return getEnvVar(split[1])
					}
				case "APT_INSTALL":
					{
						// ${APT_INSTALL:...}
						split := strings.Split(split[1], ",")
						if len(split) >= 1 {
							return "RUN apt-get update && \\\n" +
								"    export DEBIAN_FRONTEND=noninteractive && \\\n" +
								"    apt-get install -y" + strings.Join(split, " ") + " && \\\n" +
								"    rm -rf /var/lib/apt/lists/*"
						}
					}
				}
			}
			return "<<ERROR!>>"
		})
	}
}

// GetDockerfile returns for an application file the resolved dockerfile
func (cfg *AppInfo) GetDockerfile() (dockerfile string) {
	re := regexp.MustCompile("^docker://")
	dockerfile = re.ReplaceAllString(cfg.appConfig.Image.Base, "FROM ")
	if dockerfile != "" {
		dockerfile += "\n\n"
	}

	dockerfile += cfg.resolvedParams["SET_PROXY"] + "\n" +
		cfg.appConfig.Image.Dockerfile + "\n" +
		cfg.resolvedParams["USER_CTX"]

	// replace parameters
	cfg.replaceParameters(&dockerfile)

	log.Debug("dockerfile: %v", dockerfile)

	return dockerfile
}

// GetDockerRunArgs returns for an application file the resolved docker run arguments
func (cfg *AppInfo) GetDockerRunArgs() (dockerRunArgs []string) {
	defaultDockerArgs := map[string]string{
		"-h": "flybydocker",
		"-w": "${PWD}",
	}
	for _, arg := range cfg.appConfig.Runtime.Docker.RunArgs {
		if _, ok := defaultDockerArgs[arg]; ok {
			delete(defaultDockerArgs, arg)
		}
	}
	dockerRunArgs = append([]string{
		"-v", "${PWD}:${PWD}"},
		cfg.appConfig.Runtime.Docker.RunArgs...)

	if cfg.appConfig.Console {
		fi, _ := os.Stdin.Stat()
		if (fi.Mode() & os.ModeCharDevice) == 0 {
			// input from pipe
			dockerRunArgs = append(dockerRunArgs, "-i")
		} else {
			dockerRunArgs = append(dockerRunArgs, "-ti")
		}
	}

	if cfg.appConfig.Gui {
		dockerRunArgs = append(dockerRunArgs,
			"-e", "DISPLAY="+os.Getenv("DISPLAY"),
			"-v", "/tmp/.X11-unix:/tmp/.X11-unix",
		)
	}

	// take default values of unset Docker arguments
	keys := make([]string, 0)
	for k := range defaultDockerArgs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := defaultDockerArgs[key]
		dockerRunArgs = append(dockerRunArgs, key, value)
	}

	// replace parameters
	for i := range dockerRunArgs {
		cfg.replaceParameters(&dockerRunArgs[i])
	}

	log.Debug("dockerRunArgs: %v", dockerRunArgs)

	return dockerRunArgs
}
