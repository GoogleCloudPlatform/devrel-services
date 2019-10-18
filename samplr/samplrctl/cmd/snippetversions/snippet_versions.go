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

package snippetversions

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/devrel-services/samplr"
	commands "github.com/GoogleCloudPlatform/devrel-services/samplr/samplrctl/cmd"
	"github.com/GoogleCloudPlatform/devrel-services/samplr/samplrctl/snippetversions"
	"github.com/GoogleCloudPlatform/devrel-services/samplr/samplrctl/utils"

	"github.com/spf13/cobra"
)

func AddCommand(ctx context.Context, cmd *cobra.Command) {

	snippetVersions := &cobra.Command{
		Use:   "versions",
		Short: "versions",
		Long:  "versions",
	}

	snippetVersionsGet := &cobra.Command{
		Use:   "get",
		Short: "get",
		Long:  "get",
		RunE:  commands.CtxCommand(ctx, getSnippetVersion),
		Args:  cobra.ExactArgs(3),
	}

	snippetVersionsList := &cobra.Command{
		Use:   "list",
		Short: "list",
		Long:  "list",
		RunE:  commands.CtxCommand(ctx, listSnippetVersions),
		Args:  cobra.ExactArgs(2),
	}

	snippetVersions.AddCommand(snippetVersionsGet, snippetVersionsList)
	cmd.AddCommand(snippetVersions)
}

func getSnippetVersion(ctx context.Context, cmd *cobra.Command, args []string) error {
	return nil
}

func listSnippetVersions(ctx context.Context, cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("Need exactly 2 argumetns, got: %v", len(args))
	}

	dir := args[0]
	sName := args[1]

	snps, err := utils.GetSnippets(ctx, dir)
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

	snippetversions.OutputSnippetVersions(cmd.OutOrStderr(), snp.Versions)

	return nil
}
