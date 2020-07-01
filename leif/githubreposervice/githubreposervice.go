// Copyright 2020 Google LLC
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

// This package was modeled after the mocking strategy outlined at:
// https://github.com/google/go-github/issues/113#issuecomment-46023864

package githubreposervice

import (
	"context"
	"net/http"

	"github.com/google/go-github/github"
)

// RepoService is an interface defining the needed behaviour of the GitHub client
// This way, the default client may be replaced for testing
type repoService interface {
	GetContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (fileContent *github.RepositoryContent, directoryContent []*github.RepositoryContent, resp *github.Response, err error)
}

// Client is a a wrapper around the GitHub client's RepositoriesService
type Client struct {
	Repositories repoService
}

// NewClient creates a wrapper around the GitHub client's RepositoriesService
// The RepositoriesService can be replaced for unit testing
func NewClient(httpClient *http.Client, repoMock repoService) Client {
	if repoMock != nil {
		return Client{
			Repositories: repoMock,
		}
	}
	client := github.NewClient(httpClient)

	return Client{
		Repositories: client.Repositories,
	}
}
