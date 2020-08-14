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
	"errors"

	"github.com/google/go-github/github"
)

// MockGithubRepositoryService is a struct that can replace github.RepositoriesService for testing
type MockGithubRepositoryService struct {
	Content    *github.RepositoryContent
	DirContent []*github.RepositoryContent
	Users      []*github.User
	Response   *github.Response
	Error      error
	Owner      string
	Repo       string
}

// Get mocks the original github.RepositoriesService.Get() by returning the given mocked response and error
func (mgc *MockGithubRepositoryService) Get(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error) {
	return nil, mgc.Response, mgc.Error
}

// GetContents mocks the original github.RepositoriesService.GetContents()
// Checks whether the owner is correct and returns the mocked content, response and error
func (mgc *MockGithubRepositoryService) GetContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error) {
	if owner != mgc.Owner {
		return nil, nil, nil, errors.New("owner did not equal expected owner: was: " + owner)
	}
	return mgc.Content, mgc.DirContent, mgc.Response, mgc.Error
}

// ListByOrg mocks the original github.RepositoriesService.ListByOrg()
// Checks whether the owner is correct and returns the mocked error
func (mgc *MockGithubRepositoryService) ListByOrg(ctx context.Context, org string, opts *github.RepositoryListByOrgOptions) ([]*github.Repository, *github.Response, error) {
	if org != mgc.Owner {
		return nil, nil, errors.New("org did not equal expected owner: was: " + org)
	}
	return nil, nil, mgc.Error
}

// ListCollaborators mocks the original github.RepositoriesService.ListCollaborators()
// Checks whether the owner is correct and returns the mocked error
func (mgc *MockGithubRepositoryService) ListCollaborators(ctx context.Context, owner, repo string, opts *github.ListCollaboratorsOptions) ([]*github.User, *github.Response, error) {
	if owner != mgc.Owner {
		return nil, nil, errors.New("org did not equal expected owner: was: " + owner)
	}
	return mgc.Users, nil, mgc.Error
}
