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

package snippets

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/devrel-services/samplr"
	commands "github.com/GoogleCloudPlatform/devrel-services/samplr/samplrctl/cmd"
	"github.com/GoogleCloudPlatform/devrel-services/samplr/samplrctl/snippets"
	"github.com/GoogleCloudPlatform/devrel-services/samplr/samplrctl/utils"

	"github.com/spf13/cobra"
)

func AddCommand(ctx context.Context, cmd *cobra.Command) {

	snippets := &cobra.Command{
		Use:   "snippets",
		Short: "interacts with snippets",
		Long:  "interacts with snippets",
	}

	snippetsGet := &cobra.Command{
		Use:   "get",
		Short: "get",
		Long:  "get",
		RunE:  commands.CtxCommand(ctx, getSnippet),
		Args:  cobra.ExactArgs(2),
	}

	snippetsList := &cobra.Command{
		Use:   "list",
		Short: "list",
		Long:  "list",
		RunE:  commands.CtxCommand(ctx, listSnippets),
		Args:  cobra.ExactArgs(1),
	}

	snippets.AddCommand(snippetsGet)
	snippets.AddCommand(snippetsList)
	cmd.AddCommand(snippets)
}

func getSnippet(ctx context.Context, cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("Need exactly 2 arguments, got: %v", len(args))
	}
	fmt.Println("list called")
	d := args[0]
	sName := args[1]

	snps, err := utils.GetSnippets(ctx, d)
	if err != nil {
		return err
	}

	var snp *samplr.Snippet
	for _, s := range snps {
		if s.Name == sName {
			snp = s
			break
		}
	}

	if snp == nil {
		return fmt.Errorf("Could not find a snippet named: %v", sName)
	}

	snippets.OutputSnippet(cmd.OutOrStdout(), snp)

	return nil
}

func listSnippets(ctx context.Context, cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("Need exactly 1 argument, got: %v", len(args))
	}

	d := args[0]

	snps, err := utils.GetSnippets(ctx, d)
	if err != nil {
		return err
	}

	snippets.OutputSnippets(cmd.OutOrStdout(), snps)

	return nil
}
