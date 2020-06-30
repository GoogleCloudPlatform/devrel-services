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

package leif

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/go-github/github"
)

const sloConfigFileName = "issue_slo_rules.json"

var ErrNoContent = errors.New("no content found")
var ErrNotAFile = errors.New("not a file")
var GoGitHubErr = errors.New("an error came from go github")

type noContentError struct {
	path string
	org  string
	repo string
	err  error
}

func (e *noContentError) Error() string {
	return fmt.Sprintf("The path %v in %v/%v did not return any content", e.path, e.org, e.repo)
}

func (e *noContentError) Unwrap() error {
	return e.err
}

type notAFileError struct {
	path string
	org  string
	repo string
	Err  error
}

func (e *notAFileError) Error() string {
	return fmt.Sprintf("The path %v in %v/%v does not correspond to a file", e.path, e.org, e.repo)
}

func (e *notAFileError) Unwrap() error {
	return e.Err
}

// githubRepoService is an interface defining the needed behaviour of the GitHub client
// This way, the default client may be replaced for testing
type gitHubRepoService interface {
	GetContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (fileContent *github.RepositoryContent, directoryContent []*github.RepositoryContent, resp *github.Response, err error)
}

// githubClient is a a wrapper around the GitHub client's RepositoriesService
type gitHubClient struct {
	Repositories gitHubRepoService
}

// NewGithubClient creates a wrapper around the GitHub client's RepositoriesService
// The RepositoriesService can be replaced for unit testing
func NewGitHubClient(httpClient *http.Client, repoMock gitHubRepoService) gitHubClient {
	if repoMock != nil {
		return gitHubClient{
			Repositories: repoMock,
		}
	}
	client := github.NewClient(httpClient)

	return gitHubClient{
		Repositories: client.Repositories,
	}
}

// Repository represents a GitHub repository and stores its SLO rules
type Repository struct {
	name           string
	SLOFileContent string
	SLORules       []*SLORule
}

// ParseSLOs transforms the string format of the SLO config file into structured SLO rules
func (repo *Repository) ParseSLOs() error {
	slos, err := unmarshalSLOs([]byte(repo.SLOFileContent))
	if err != nil {
		return err
	}

	repo.SLORules = slos

	return nil
}

func (repo *Repository) findSLODoc(ctx context.Context, orgName string, repoName string, ghClient *gitHubClient) (lastPathLookedAt string, e error) {
	var ghErrorResponse *github.ErrorResponse

	path := ".github/" + sloConfigFileName

	file, err := fetchFile(ctx, orgName, repoName, path, ghClient)

	if errors.As(err, &ghErrorResponse) && ghErrorResponse.Response.StatusCode == 404 {
		// SLO config not found, look for file in org:
		repoName = ".github"
		path = sloConfigFileName
		file, err = fetchFile(ctx, orgName, repoName, path, ghClient)
	}
	if err != nil {
		return fmt.Sprintf("%v/%v/%v", orgName, repoName, path), err
	}

	repo.SLOFileContent = file

	return fmt.Sprintf("%v/%v/%v", orgName, repoName, path), nil
}

func fetchFile(ctx context.Context, orgName string, repoName string, filePath string, ghClient *gitHubClient) (string, error) {

	content, _, _, err := ghClient.Repositories.GetContents(ctx, orgName, repoName, filePath, nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", GoGitHubErr, err)
	}
	if content == nil {
		error := noContentError{path: filePath, org: orgName, repo: repoName, err: ErrNoContent}
		return "", &error
	}
	if content.GetType() != "file" {
		error := notAFileError{path: filePath, org: orgName, repo: repoName, Err: ErrNotAFile}

		return "", &error
	}

	return content.GetContent()
}
