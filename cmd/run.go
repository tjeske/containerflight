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

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [OPTIONS] APPFILE",
	Short: "Run a containerflight app",
	Long:  `Run a containerflight app`,
	Args:  cli.RequiresMinArgs(1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 1 {
			core.Run(args[0], args[1:])
		} else {
			core.Run(args[0], []string{})
		}

	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	flags := runCmd.Flags()
	flags.SetInterspersed(false)
}
