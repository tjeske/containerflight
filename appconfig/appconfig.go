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
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/blang/semver"
	yaml "github.com/go-yaml/yaml"
	log "github.com/sirupsen/logrus"
	"github.com/tjeske/containerflight/util"
)

// specification of an app file
type yamlSpec struct {
	Compatibility string
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

type appConfig struct {
	appConfig      yamlSpec
	env            *environment
	resolvedParams map[string]string
}

var parameterRegex = regexp.MustCompile("\\$\\{(.+?)\\}")

func NewAppConfig(yamlAppFileName *string, cfVersion semver.Version) *appConfig {

	absYamlAppFileName, err := filepath.Abs(*yamlAppFileName)
	util.CheckErr(err)

	env := getEnv(absYamlAppFileName)

	appFile := getAppFile(env)

	// check version
	if appFile.Compatibility != "" {
		parsedRange, err := semver.ParseRange(appFile.Compatibility)
		util.CheckErrMsg(err, "Version information must match semver 2.0.0 (https://semver.org/)!")
		if !parsedRange(cfVersion) {
			log.Fatal("App file is not compatible with current Container Flight version " + cfVersion.String() + "!")
		}
	}

	resolvedParams := getResolvedParameters(env)

	return &appConfig{
		appConfig:      appFile,
		env:            env,
		resolvedParams: resolvedParams,
	}
}

// read and parse app config file
func getAppFile(env *environment) yamlSpec {

	// read the app file
	yamlFileBytes, err := ioutil.ReadFile(env.appFile)
	util.CheckErr(err)
	str := string(yamlFileBytes)

	// unmarshal yaml file
	appFile := yamlSpec{}
	err = yaml.UnmarshalStrict([]byte(str), &appFile)
	util.CheckErr(err)

	return appFile
}

// map the parameters which can be used in an app file to their corresponding values
func getResolvedParameters(env *environment) map[string]string {
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

// search and replace parameters in string
func (cfg *appConfig) replaceParameters(str *string) {
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
						return os.Getenv(split[1])
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

func (cfg *appConfig) GetDockerfile() (dockerfile string) {
	re := regexp.MustCompile("^docker://")
	dockerfile = re.ReplaceAllString(cfg.appConfig.Image.Base, "FROM ") + "\n\n"

	dockerfile += cfg.resolvedParams["SET_PROXY"] + "\n" + cfg.appConfig.Image.Dockerfile + "\n" + cfg.resolvedParams["USER_CTX"]

	// replace parameters
	cfg.replaceParameters(&dockerfile)

	log.Debug("dockerfile: %v", dockerfile)

	return dockerfile
}

func (cfg *appConfig) GetDockerRunArgs() (dockerRunArgs []string) {
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
	for key, value := range defaultDockerArgs {
		dockerRunArgs = append(dockerRunArgs, key, value)
	}

	// replace parameters
	for i := range dockerRunArgs {
		cfg.replaceParameters(&dockerRunArgs[i])
	}

	log.Debug("dockerRunArgs: %v", dockerRunArgs)

	return dockerRunArgs
}
