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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUnixFilePathWindows(t *testing.T) {
	separator = '\\'

	assert.Equal(t, "/a/b", GetUnixFilePath("/a/b"))
	assert.Equal(t, "a/b", GetUnixFilePath("a/b"))
	assert.Equal(t, "a", GetUnixFilePath("a"))
	assert.Equal(t, "/c/a/b", GetUnixFilePath("c:\\a\\b"))
	assert.Equal(t, "/a/b", GetUnixFilePath("\\a\\b"))
	assert.Equal(t, "a/b", GetUnixFilePath("a\\b"))
}
