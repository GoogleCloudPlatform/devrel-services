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

package samplr

import (
	"context"
	git "github.com/GoogleCloudPlatform/devrel-services/git-go"
	"io/ioutil"
	"os"
	"testing"
)

func TestRepositoryHasSnippet(t *testing.T) {
	cases := []struct {
		Name                string
		URL                 string
		SnippetName         string
		WantMinimumVersions int
	}{
		{
			Name:                "Handles Merges",
			URL:                 "https://github.com/GoogleCloudPlatform/dotnet-docs-samples",
			SnippetName:         "owners/GoogleCloudPlatform/repositories/dotnet-docs-samples/snippets/bigtable_hw_imports/languages/CSHARP",
			WantMinimumVersions: 2,
		},
	}
	for _, c := range cases {
		dirname, err := ioutil.TempDir("", "samplr-")
		if err != nil {
			t.Errorf("Couldnt get a temp dir: %v", err)
			continue
		}
		defer os.RemoveAll(dirname)

		r, err := git.PlainClone(dirname, false, &git.CloneOptions{
			URL: c.URL,
		})
		if err != nil {
			t.Errorf("Error cloning %v: %v", c.URL, err)
			continue
		}
		cor := &Corpus{}
		wgh := watchedGitRepo{
			repository: r,
			c:          cor,
			id:         c.URL,
			snippets:   make(map[string][]*Snippet),
			commits:    make(map[string][]*GitCommit),
		}

		ctx := context.Background()

		err = wgh.Update(ctx)
		if err != nil {
			t.Errorf("Error during update: %v", err)
			return
		}

		var found *Snippet
		// Find the snippet
		wgh.ForEachSnippet(func(snippet *Snippet) error {
			if snippet.Name == c.SnippetName {
				found = snippet
			}
			return nil
		})

		if found == nil {
			t.Errorf("Could not find snippet named: %v", c.SnippetName)
			continue
		}

		if len(found.Versions) < c.WantMinimumVersions {
			t.Errorf("Snippet: %v has %v versions, want: %v", c.SnippetName, len(found.Versions), c.WantMinimumVersions)
			continue
		}
	}
}
