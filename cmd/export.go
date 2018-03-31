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

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export the container description",
	Long:  `Export the container description`,
}

// exportCmd represents the export dockerfile command
var dockerFileCmd = &cobra.Command{
	Use:   "dockerfile [OPTIONS] FLYFILE",
	Short: "Show the processed Dockerfile",
	Long:  `Show the processed Dockerfile of the app container`,
	Args:  cli.RequiresRangeArgs(1, 1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		core.PrintDockerfile(args[0])
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.AddCommand(dockerFileCmd)
	flags := dockerFileCmd.Flags()
	flags.SetInterspersed(false)
}
