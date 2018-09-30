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

package cmd

import (
	"github.com/tjeske/containerflight/core"

	"github.com/docker/cli/cli"
	"github.com/spf13/cobra"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build [OPTIONS] APPFILE",
	Short: "Build a containerflight app image",
	Long:  `Build a containerflight app image`,
	Args:  cli.RequiresMinArgs(1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		core.Build(args[0])
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	flags := buildCmd.Flags()
	flags.SetInterspersed(false)
}
