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

package githubservices

import (
	"context"
	"net/http"

	"github.com/google/go-github/github"
)

// repoService is an interface defining the needed behaviour of the GitHub client
// This way, the default client may be replaced for testing
type repoService interface {
	Get(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error)
	GetContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (fileContent *github.RepositoryContent, directoryContent []*github.RepositoryContent, resp *github.Response, err error)
	ListByOrg(ctx context.Context, org string, opts *github.RepositoryListByOrgOptions) ([]*github.Repository, *github.Response, error)
	ListCollaborators(ctx context.Context, owner, repo string, opts *github.ListCollaboratorsOptions) ([]*github.User, *github.Response, error)
}

type userService interface {
	Get(ctx context.Context, user string) (*github.User, *github.Response, error)
}

// Client is a a wrapper around the GitHub client's RepositoriesService
type Client struct {
	Repositories repoService
	Users        userService
}

// NewClient creates a wrapper around the GitHub client's RepositoriesService
// The RepositoriesService can be replaced for unit testing
func NewClient(httpClient *http.Client, repoMock repoService, userMock userService) Client {
	if repoMock != nil || userMock != nil {
		return Client{
			Repositories: repoMock,
			Users:        userMock,
		}
	}
	client := github.NewClient(httpClient)

	return Client{
		Repositories: client.Repositories,
		Users:        client.Users,
	}
}
