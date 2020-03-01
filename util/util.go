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

package util

import (
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

// GetWorkingDir returns the current working directory
var GetWorkingDir = func() string {
	workingDir, err := os.Getwd()
	CheckErr(err)
	return workingDir
}

// configurable path separator for unit-tests
var separator = os.PathSeparator

// regex to check for windows drive letters
var winDriveLetterRegex = regexp.MustCompile(`^([a-zA-Z]):/`)

// CheckErr checks for an error
// use this function after calling an error returning function
func CheckErr(err error) {
	if err != nil {
		log.Fatal("ERROR: ", err)
	}
}

// CheckErrMsg checks for an error and prints out a custom error message
// use this function after calling an error returning function
func CheckErrMsg(err error, msg string) {
	if err != nil {
		log.Fatal("ERROR: ", msg+" ("+err.Error()+")")
	}
}

// GetUnixFilePath transforms a unix/windows file path into an unix file path
func GetUnixFilePath(filePath string) string {
	unixFilePath := ToSlash(filePath)
	unixFilePath = winDriveLetterRegex.ReplaceAllString(unixFilePath, `/$1/`)
	return unixFilePath
}

// ToSlash returns the result of replacing each separator character
// in path with a slash ('/') character. Multiple separators are
// replaced by multiple slashes.
func ToSlash(path string) string {
	if separator == '/' {
		return path
	}
	return strings.ReplaceAll(path, string(separator), "/")
}
