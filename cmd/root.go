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
	"os"

	"github.com/docker/cli/cli"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var debug bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "containerflight",
	Short: "Run applications in a defined and isolated environment",
	Long:  `Run applications in a defined and isolated environment`,
	Args:  cli.RequiresMinArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debug == true {
			log.SetLevel(log.DebugLevel)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately
func Execute() {
	flags := rootCmd.Flags()
	flags.SetInterspersed(false)

	persistentFlags := rootCmd.PersistentFlags()
	persistentFlags.BoolVarP(&debug, "debug", "d", false, "print out debug information")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
