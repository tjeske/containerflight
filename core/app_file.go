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
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	yaml "github.com/go-yaml/yaml"
	log "github.com/sirupsen/logrus"
)

// specification of an app file
type yamlSpec struct {
	Version string
	Console bool
	Gui     bool
	Docker  struct {
		Dockerfile string
		RunArgs    []string
	}
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

var parameterRegex = regexp.MustCompile("\\$\\{(.+?)\\}")
var setProxyRegex = regexp.MustCompile("(^.*FROM.*?\n)")

// search and replace parameters in string
func replaceParameters(str *string, resolvedParams *map[string]string) {
	oldYamlFileStr := ""
	for *str != oldYamlFileStr {
		oldYamlFileStr = *str
		*str = parameterRegex.ReplaceAllStringFunc(*str, func(match string) string {
			trimmedMatch := match[2 : len(match)-1]
			split := strings.Split(trimmedMatch, ":")
			if len(split) == 1 {
				if value, ok := (*resolvedParams)[split[0]]; ok {
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

// read and parse app config file
func getAppConfig(env *environment) yamlSpec {

	// read the app file
	yamlFileBytes, err := ioutil.ReadFile(env.appFile)
	checkErr(err)
	str := string(yamlFileBytes)

	// unmarshal yaml file
	appFile := yamlSpec{}
	err = yaml.UnmarshalStrict([]byte(str), &appFile)
	checkErr(err)

	resolvedParams := getResolvedParameters(env)

	replacedStr := "${1}" + "\n" + resolvedParams["SET_PROXY"]
	appFile.Docker.Dockerfile = setProxyRegex.ReplaceAllString(appFile.Docker.Dockerfile, replacedStr)
	appFile.Docker.Dockerfile += "\n" + resolvedParams["USER_CTX"]

	defaultDockerArgs := map[string]string{
		"-h": "flybydocker",
		"-w": "${PWD}",
	}
	for _, arg := range appFile.Docker.RunArgs {
		if _, ok := defaultDockerArgs[arg]; ok {
			delete(defaultDockerArgs, arg)
		}
	}

	appFile.Docker.RunArgs = append([]string{"-v", "${PWD}:${PWD}"}, appFile.Docker.RunArgs...)

	if appFile.Console {
		fi, _ := os.Stdin.Stat()
		if (fi.Mode() & os.ModeCharDevice) == 0 {
			// input from pipe
			appFile.Docker.RunArgs = append(appFile.Docker.RunArgs, "-i")
		} else {
			appFile.Docker.RunArgs = append(appFile.Docker.RunArgs, "-ti")
		}
	}

	if appFile.Gui {
		appFile.Docker.RunArgs = append(appFile.Docker.RunArgs,
			"-e", "DISPLAY="+os.Getenv("DISPLAY"),
			"-v", "/tmp/.X11-unix:/tmp/.X11-unix",
		)
	}

	// take default values of unset Docker arguments
	for key, value := range defaultDockerArgs {
		appFile.Docker.RunArgs = append(appFile.Docker.RunArgs, key, value)
	}

	// replace parameters
	replaceParameters(&appFile.Docker.Dockerfile, &resolvedParams)
	for i := range appFile.Docker.RunArgs {
		replaceParameters(&appFile.Docker.RunArgs[i], &resolvedParams)
	}

	log.Debug("appFile: %v", appFile)

	return appFile
}
