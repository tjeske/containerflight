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
	"github.com/blang/semver"
	yaml "github.com/go-yaml/yaml"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/tjeske/containerflight/util"
	"github.com/tjeske/containerflight/version"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// "mock connectors" for unit-tesing
var logFatalf = log.Fatalf
var filesystem = afero.NewOsFs()

// specification of an app file
type yamlSpec struct {
	Compatibility string
	Name          string
	Version       string
	Description   string
	Console       bool
	Gui           bool

	Image struct {
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

var parameterRegex = regexp.MustCompile("\\$\\{[[:word:]]+(\\(.*?\\))?\\}")
var parameterSplitRegex = regexp.MustCompile(`(?P<name>[[:word:]]+)(\((?P<args>.+)\))?`)

// NewAppInfo returns a representation of an application config file
func NewAppInfo(appConfigFile string) *AppInfo {

	absAppConfigFile, err := filepath.Abs(appConfigFile)
	util.CheckErr(err)

	yamlAppFileReader, err := filesystem.Open(absAppConfigFile)
	util.CheckErr(err)

	env := getEnv(absAppConfigFile)

	appConfig := getAppConfig(yamlAppFileReader)

	resolvedParams := getResolvedParameters(env)

	validate(appConfig)

	return &AppInfo{
		appConfig:      appConfig,
		env:            env,
		resolvedParams: resolvedParams,
	}
}

// NewFakeAppInfo returns a fake representation of an application config file for unit-testing
func NewFakeAppInfo(fs *afero.Fs, appConfigFile string, appConfigStr string) *AppInfo {
	origFS := filesystem
	defer func() { filesystem = origFS }()
	filesystem = *fs

	// mock environment variables
	getEnvVar = func(name string) string {
		return name
	}

	// mock environment
	getEnv = func(appConfigFile string) environment {
		absAppConfigFile, err := filepath.Abs(appConfigFile)
		util.CheckErr(err)

		appFileDir := filepath.Dir(absAppConfigFile)

		// create environment object
		var env = environment{
			appConfigFile: absAppConfigFile,
			appFileDir:    appFileDir,
			userName:      "testuser",
			userID:        "1234",
			groupName:     "testgroup",
			groupID:       "5678",
			homeDir:       "/home",
			workingDir:    "/myworkingdir",
		}
		return env
	}

	afero.WriteFile(filesystem, appConfigFile, []byte(appConfigStr), 0644)

	return NewAppInfo(appConfigFile)
}

// validate app config file
func validate(appInfoConfig yamlSpec) {
	// check version
	if appInfoConfig.Compatibility != "" {
		cfVersion := version.ContainerFlightVersion()
		parsedRange, err := semver.ParseRange(appInfoConfig.Compatibility)
		util.CheckErrMsg(err, "Version information must match semver 2.0.0 (https://semver.org/)!")
		if !parsedRange(cfVersion) {
			logFatalf("App file is not compatible with current containerflight version %s!", cfVersion.String())
		}
	}
}

// read and parse app config file
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
		"APP_FILE_DIR": env.appFileDir,
		"USERNAME":     env.userName,
		"USERID":       env.userID,
		"GROUPNAME":    env.groupName,
		"GROUPID":      env.groupID,
		"HOME":         env.homeDir,
		"PWD":          env.workingDir,
		"SET_PROXY": "ENV http_proxy=${ENV(http_proxy)}\n" +
			"ENV https_proxy=${ENV(https_proxy)}\n" +
			"ENV no_proxy=${ENV(no_proxy)}\n",
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
			"            adduser --gid ${GROUPNAME} --uid ${USERID} --base-dir \"${HOME}\" ${USERNAME} || \\\n" +
			"            # ubuntu\\\n" +
			"            adduser --home \"${HOME}\" --uid ${USERID} --gecos \"\" --ingroup ${GROUPNAME} --disabled-password ${USERNAME} || \\\n" +
			"            # busybox\\\n" +
			"            adduser -h \"${HOME}\" -u ${USERID} -D -H -G ${GROUPNAME} ${USERNAME} \\\n" +
			"        ) > /dev/null 2>&1 ; \\\n" +
			"    fi ;\n\n" +
			"USER ${USERNAME}",
	}
}

// search and replace parameters in string
func (cfg *AppInfo) replaceParameters(str *string) {
	oldYamlFileStr := ""
	for *str != oldYamlFileStr {
		oldYamlFileStr = *str
		*str = parameterRegex.ReplaceAllStringFunc(*str, func(match string) string {
			trimmedMatch := match[2 : len(match)-1]
			split := parameterSplitRegex.FindStringSubmatch(trimmedMatch)

			if split[2] == "" {
				if value, ok := cfg.resolvedParams[split[0]]; ok {
					// ${KEY}
					return value
				}
			} else {
				switch split[1] {
				case "ENV":
					{
						// ${ENV(...)}
						return getEnvVar(split[3])
					}
				case "APT_INSTALL":
					{
						// ${APT_INSTALL(...)}
						args := strings.Split(split[3], ",")
						for i := range args {
							args[i] = strings.TrimSpace(args[i])
						}
						if len(split) >= 1 {
							return "RUN apt-get update && \\\n" +
								"    export DEBIAN_FRONTEND=noninteractive && \\\n" +
								"    apt-get install -y " + strings.Join(args, " ") + " && \\\n" +
								"    rm -rf /var/lib/apt/lists/*"
						}
					}
				}
			}
			return "<<ERROR!>>"
		})
	}
}

// GetResolvedAppConfig returns the resolved app file
func (cfg *AppInfo) GetResolvedAppConfig() string {

	appConfigByte, err := yaml.Marshal(&cfg.appConfig)
	util.CheckErr(err)

	appConfigStr := string(appConfigByte)

	// replace parameters
	cfg.replaceParameters(&appConfigStr)

	return appConfigStr

}

// GetAppFileDir returns the app file directory
func (cfg *AppInfo) GetAppFileDir() string {
	return cfg.env.appFileDir
}

// GetAppConfigFile returns the app file
func (cfg *AppInfo) GetAppConfigFile() string {
	return cfg.env.appConfigFile
}

// GetAppName returns the name of the application
func (cfg *AppInfo) GetAppName() string {
	name := cfg.appConfig.Name

	// replace parameters
	cfg.replaceParameters(&name)

	if name == "" {
		name = filepath.Base(cfg.env.appConfigFile)
	}

	return name
}

// GetAppVersion returns the version of an application file
func (cfg *AppInfo) GetAppVersion() string {
	version := cfg.appConfig.Version

	// replace parameters
	cfg.replaceParameters(&version)

	return version
}

// GetAppDescription returns the description of an application file
func (cfg *AppInfo) GetAppDescription() string {
	description := cfg.appConfig.Description

	// replace parameters
	cfg.replaceParameters(&description)

	return description
}

// GetDockerfile returns for an app file the resolved dockerfile
func (cfg *AppInfo) GetDockerfile() string {
	dockerfileFinal := ""
	re := regexp.MustCompile("^docker://")
	baseImage := re.ReplaceAllString(cfg.appConfig.Image.Base, "FROM ")
	if baseImage != "" {
		dockerfileFinal += baseImage + "\n\n"
	}

	dockerfile := cfg.handleDockerfileLoad(cfg.appConfig.Image.Dockerfile)

	dockerfileFinal += cfg.resolvedParams["SET_PROXY"] + "\n" +
		dockerfile + "\n" +
		cfg.resolvedParams["USER_CTX"]

	// replace parameters
	cfg.replaceParameters(&dockerfileFinal)

	log.Debug("dockerfile: ", dockerfileFinal)

	return dockerfileFinal
}

// deal with "file://" notation in image -> dockerfile
func (cfg *AppInfo) handleDockerfileLoad(dockerfile string) string {
	if len(strings.Split(dockerfile, "\n")) == 1 {
		split := regexp.MustCompile("^file://").Split(strings.TrimSpace(dockerfile), 2)
		if len(split) == 2 {
			userFileName := split[1]

			// try to interpret as an absolute path
			rawData, err := afero.ReadFile(filesystem, userFileName)
			if err != nil {
				// cannot open file -> try to interpret as a relative path
				fileName := filepath.Join(cfg.env.appFileDir, userFileName)
				rawData, err = afero.ReadFile(filesystem, fileName)
				if err != nil {
					logFatalf("Cannot read file \"" + fileName + "\"!")
				}
			}
			dockerfile = string(rawData)
		}
	}
	return dockerfile
}

// GetDockerRunArgs returns for an app file the resolved docker run arguments
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
	dockerRunArgs = append(
		[]string{"-v", "${PWD}:${PWD}"},
		cfg.appConfig.Runtime.Docker.RunArgs...,
	)

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
			"-e", "DISPLAY="+getEnvVar("DISPLAY"),
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

	// use absolute dirs for volumes so that duplicated mount points can be detected by Docker
	for i := range dockerRunArgs {
		if strings.TrimSpace(dockerRunArgs[i]) == "-v" && i+1 < len(dockerRunArgs) {
			dirs := strings.Split(dockerRunArgs[i+1], ":")
			if len(dirs) == 2 {
				hostPath, _ := filepath.Abs(dirs[0])
				containerPath, _ := filepath.Abs(dirs[1])
				i++
				dockerRunArgs[i] = hostPath + ":" + containerPath
			}

		}
		cfg.replaceParameters(&dockerRunArgs[i])
	}

	log.Debug("dockerRunArgs: ", dockerRunArgs)

	return dockerRunArgs
}

var getEnvVar = func(name string) string {
	return os.Getenv(name)
}
