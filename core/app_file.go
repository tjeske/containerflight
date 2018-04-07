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
		"SET_PROXY": `ENV http_proxy=${ENV:http_proxy}
        ENV https_proxy=${ENV:https_proxy}`,
		"USER_CTX": `RUN if ! getent group ${GROUPNAME} > /dev/null 2>&1; then \
                ( \
                # ubuntu
                addgroup -g ${GROUPID} ${GROUPNAME} || \
                # busybox
                addgroup --gid ${GROUPID} ${GROUPNAME} || \
                # fedora
                groupadd --gid ${GROUPID} ${GROUPNAME} \
                ) > /dev/null 2>&1 ; \
            fi ; \
            if ! getent passwd ${USERNAME} > /dev/null 2>&1; then \
                ( \
                # fedora
                adduser --gid ${GROUPNAME} --uid ${USERID} ${USERNAME} || \
                # ubuntu
                adduser --uid ${USERID} --gecos "" --ingroup ${GROUPNAME} --disabled-password ${USERNAME} || \
                # busybox
                adduser -u ${USERID} -D -H -G ${GROUPNAME} ${USERNAME} \
                ) > /dev/null 2>&1 ; \
            fi ;
        
        USER ${USERNAME}`,
	}
}

// search and replace parameters in string
func replaceParameters(str *string, resolvedParams *map[string]string) {
	re := regexp.MustCompile("\\$\\{(.+?)\\}")
	oldYamlFileStr := ""
	for *str != oldYamlFileStr {
		oldYamlFileStr = *str
		*str = re.ReplaceAllStringFunc(*str, func(match string) string {
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
							return `RUN apt-get update && \
                            export DEBIAN_FRONTEND=noninteractive && \
                            apt-get install -y` + strings.Join(split, " ") + ` && \
                            rm -rf /var/lib/apt/lists/*`
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

	// replace parameters
	resolvedParams := getResolvedParameters(env)
	replaceParameters(&str, &resolvedParams)

	// unmarshal yaml file
	appFile := yamlSpec{}
	err = yaml.UnmarshalStrict([]byte(str), &appFile)
	checkErr(err)

	appFile.Docker.Dockerfile += "\n" + resolvedParams["USER_CTX"]
	replaceParameters(&appFile.Docker.Dockerfile, &resolvedParams)

	log.Debug("appFile: %v", appFile)

	return appFile
}
