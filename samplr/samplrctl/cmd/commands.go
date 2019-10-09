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

package commands

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	verboseFlagName = "verbose"
)

// AddFlags adds flags that are common across all commands.
func AddFlags(s *pflag.FlagSet) {
	// s.BoolP(verboseFlagName, "v", false, "Sets Log level to \"Debug\"")
}

// CobraActionFunc represents a cobra command
type CobraActionFunc func(cmd *cobra.Command, args []string) error

// SamplrActionFunc provides a common type for Cobra functions that require a
// stubby client connection and context.
type SamplrActionFunc func(ctx context.Context, cmd *cobra.Command, args []string) error

// CtxCommand allows the running of a command with a context
func CtxCommand(ctx context.Context, f SamplrActionFunc) CobraActionFunc {
	return func(cmd *cobra.Command, args []string) error {
		return f(ctx, cmd, args)
	}
}
