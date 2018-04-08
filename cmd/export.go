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

// exportCmd represents the "export" command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export the container description",
	Long:  `Export the container description`,
}

// dockerCmd represents the "export docker" command
var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Show Dockerfile or \"docker run\" arguments",
	Long:  `Show Dockerfile or "docker run" arguments`,
}

// dockerFileCmd represents the "export docker dockerfile" command
var dockerFileCmd = &cobra.Command{
	Use:   "dockerfile [OPTIONS] APPFILE",
	Short: "Show the processed Dockerfile",
	Long:  `Show the processed Dockerfile of the app container`,
	Args:  cli.RequiresRangeArgs(1, 1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		core.PrintDockerfile(args[0])
	},
}

// dockerRunArgsCmd represents the "export docker runargs" command
var dockerRunArgsCmd = &cobra.Command{
	Use:   "runargs [OPTIONS] APPFILE",
	Short: "Show the Docker run args",
	Long:  `Show the arguments for "docker run" which are used to run the Docker container`,
	Args:  cli.RequiresRangeArgs(1, 1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		core.PrintDockerRunArgs(args[0])
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.AddCommand(dockerCmd)
	dockerCmd.AddCommand(dockerFileCmd)
	dockerCmd.AddCommand(dockerRunArgsCmd)
	flags := dockerFileCmd.Flags()
	flags.SetInterspersed(false)
}
