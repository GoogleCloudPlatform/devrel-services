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

// Package commands contains helpers for working with commands in tests.
package testutil

import (
	"bytes"

	commands "devrel/cloud/devrel-github-service/samplr/samplrctl/cmd"

	"github.com/spf13/cobra"
)

// CreateTestCommand create a Cobra command suitable for use in tests.
func CreateTestCommand() *cobra.Command {
	cmd := &cobra.Command{}
	commands.AddFlags(cmd.PersistentFlags())
	return cmd
}

// Execute will execute the provided command with args and return the output.
func Execute(cmd *cobra.Command, args []string) (string, error) {
	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return output.String(), err
}
