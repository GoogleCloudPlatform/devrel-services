// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"os"

	"devrel/cloud/devrel-github-service/samplr/samplrctl/cmd"
	"devrel/cloud/devrel-github-service/samplr/samplrctl/cmd/completion"
	"devrel/cloud/devrel-github-service/samplr/samplrctl/cmd/snippets"
	"devrel/cloud/devrel-github-service/samplr/samplrctl/cmd/snippetversions"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.Level = logrus.DebugLevel
	log.Out = os.Stdout
}

func main() {
	ctx := context.Background()

	cmd := Command(ctx, "samplrctl")
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// Command returns a *cobra.Command setup with the common set of commands
// and configuration already done.
func Command(ctx context.Context, name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:           name,
		Short:         "Command line interface for samplr",
		SilenceUsage:  false,
		SilenceErrors: false,
		Args:          cobra.MinimumNArgs(1),
	}
	cmd.SetOutput(os.Stdout)

	commands.AddFlags(cmd.PersistentFlags())

	completion.AddCommand(ctx, cmd)
	snippets.AddCommand(ctx, cmd)
	snippetversions.AddCommand(ctx, cmd)

	updateHelpFlag(cmd)

	return cmd
}

func updateHelpFlag(cmd *cobra.Command) {
	cmd.Flags().BoolP("help", "h", false, "Help for "+cmd.Name())
	for _, c := range cmd.Commands() {
		updateHelpFlag(c)
	}
}
