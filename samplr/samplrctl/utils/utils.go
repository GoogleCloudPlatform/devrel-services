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

package utils

import (
	"context"
	"errors"
	"strings"

	git "github.com/GoogleCloudPlatform/devrel-services/git-go"
	"github.com/GoogleCloudPlatform/devrel-services/samplr"

	"golang.org/x/sync/errgroup"
)

// GetSnippets retrieves the Snippets from the given directory path
func GetSnippets(ctx context.Context, d string) ([]*samplr.Snippet, error) {
	r, err := git.PlainOpen(d)
	if err != nil {
		return nil, err
	}

	remotes, err := r.Remotes()
	if err != nil {
		return nil, err
	}

	orgName := ""
	repoName := ""
	err = remotes.ForEach(func(re *git.Remote) error {
		if re.Config() == nil {
			return errors.New("Remote has a nil Config")
		}
		if re.Config().Name != "origin" {
			return nil
		}

		if len(re.Config().URLs) == 0 {
			return errors.New("No URLs associated with remote")
		}

		parts := strings.Split(re.Config().URLs[0], "/")
		if len(parts) < 3 {
			return errors.New("Remote has an invalid URL")
		}

		repoName = parts[len(parts)-1]
		orgName = parts[len(parts)-2]
		return nil
	})
	if err != nil {
		return nil, err
	}

	refIter, err := r.Branches()
	if err != nil {
		return nil, err
	}

	group, ctx := errgroup.WithContext(ctx)

	snps := make([]*samplr.Snippet, 0)

	refIter.ForEach(func(ref *git.Reference) error {
		if ref.Name() != git.Master {
			return nil
		}
		group.Go(func() error {
			cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
			if err != nil {
				return err
			}
			snps, err = samplr.CalculateSnippets(orgName, repoName, cIter)
			if err != nil {
				return err
			}

			return nil
		})
		return nil
	})

	if err = group.Wait(); err != nil {
		return nil, err
	}
	return snps, nil
}
