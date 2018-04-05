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
                adduser --gid ${GROUPNAME} --no-create-home --uid ${USERID} ${USERNAME} || \
                # ubuntu
                adduser --no-create-home --uid ${USERID} --gecos "" --ingroup ${GROUPNAME} --disabled-password ${USERNAME} || \
                # busybox
                adduser -u ${USERID} -D -H -G ${GROUPNAME} ${USERNAME} \
                ) > /dev/null 2>&1 ; \
            fi ;
        
        USER ${USERNAME}`,
	}
}

// read and parse app config file
func getAppConfig(env *environment) yamlSpec {

	// read the app file
	yamlFileBytes, err := ioutil.ReadFile(env.appFile)
	checkErr(err)
	yamlFileStr := string(yamlFileBytes)

	// replace parameters
	resolvedParameters := getResolvedParameters(env)
	re := regexp.MustCompile("\\$\\{(.+?)\\}")
	oldYamlFileStr := ""
	for yamlFileStr != oldYamlFileStr {
		oldYamlFileStr = yamlFileStr
		yamlFileStr = re.ReplaceAllStringFunc(yamlFileStr, func(match string) string {
			trimmedMatch := match[2 : len(match)-1]
			split := strings.Split(trimmedMatch, ":")
			if len(split) == 1 {
				if value, ok := resolvedParameters[split[0]]; ok {
					// ${KEY}
					return value
				}
			} else if len(split) > 1 {
				if split[0] == "ENV" {
					// ${ENV:...}
					return os.Getenv(split[1])
				}
			}
			return "<<ERROR!>>"
		})
	}

	// unmarshal yaml file
	appFile := yamlSpec{}
	err = yaml.UnmarshalStrict([]byte(yamlFileStr), &appFile)
	checkErr(err)
	log.Debug("appFile: %v", appFile)

	return appFile
}
