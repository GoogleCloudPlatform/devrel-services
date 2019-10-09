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

// Package completion provides shell completion capabilities for the CLI.
package completion

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

var (
	errUnspecifiedShell = errors.New("Shell not specified")
	errTooManyArgs      = errors.New("Too many arguments. Expected only the shell type")
	completionShells    = map[string]func(w io.Writer, cmd *cobra.Command) error{
		"bash": runCompletionBash,
	}
)

type unsupportedShellErr struct {
	shellName string
}

func (e *unsupportedShellErr) Error() string {
	return fmt.Sprintf("Unsupported shell type %q", e.shellName)
}

// AddCommand adds the completion sub-command to the passed in root command.
func AddCommand(ctx context.Context, root *cobra.Command) {
	shells := []string{}
	for s := range completionShells {
		shells = append(shells, s)
	}

	completion := &cobra.Command{
		Use:       "completion",
		Short:     "Generate shell completion",
		Long:      `Generate shell completion.`,
		ValidArgs: shells,
		RunE:      runCompletion,
	}

	root.AddCommand(completion)
}

func runCompletion(c *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errUnspecifiedShell
	}
	if len(args) > 1 {
		return errTooManyArgs
	}
	gen, found := completionShells[args[0]]
	if !found {
		return &unsupportedShellErr{shellName: args[0]}
	}
	return gen(c.OutOrStdout(), c)
}

func runCompletionBash(w io.Writer, cmd *cobra.Command) error {
	return cmd.Root().GenBashCompletion(w)
}
