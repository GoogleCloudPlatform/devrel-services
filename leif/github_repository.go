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
	"path"

	"github.com/GoogleCloudPlatform/devrel-services/leif/githubreposervice"

	"github.com/google/go-github/github"
)

const gitHubDir = ".github"
const sloConfigFileName = "issue_slo_rules.json"

var ErrGoGitHub = errors.New("an error came from go github")
var ErrNoContent = errors.New("no content found")
var ErrNotAFile = errors.New("not a file")

type goGitHubErr github.ErrorResponse

func (e *goGitHubErr) Error() string {
	return e.Message
}

func (e *goGitHubErr) Unwrap() error {
	return ErrGoGitHub
}

type noContentError struct {
	path  string
	owner string
	repo  string
	err   error
}

func (e *noContentError) Error() string {
	return fmt.Sprintf("The path %v in %v/%v did not return any content", e.path, e.owner, e.repo)
}

func (e *noContentError) Unwrap() error {
	return e.err
}

type notAFileError struct {
	path  string
	owner string
	repo  string
	err   error
}

func (e *notAFileError) Error() string {
	return fmt.Sprintf("The path %v in %v/%v does not correspond to a file", e.path, e.owner, e.repo)
}

func (e *notAFileError) Unwrap() error {
	return e.err
}

// Repository represents a GitHub repository and stores its SLO rules
type Repository struct {
	name     string
	SLORules []*SLORule
}

func findSLODoc(ctx context.Context, owner Owner, repoName string, ghClient *githubreposervice.Client) ([]*SLORule, error) {
	var ghErrorResponse *goGitHubErr

	p := path.Join(gitHubDir, sloConfigFileName)

	file, err := fetchFile(ctx, owner.name, repoName, p, ghClient)

	if errors.As(err, &ghErrorResponse) && ghErrorResponse.Response.StatusCode == 404 {
		// SLO config not found, get SLO rules from owner:
		return owner.SLORules, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error finding SLO config: %w", err)
	}

	return unmarshalSLOs([]byte(file))
}

func fetchFile(ctx context.Context, ownerName string, repoName string, filePath string, ghClient *githubreposervice.Client) (string, error) {
	content, _, _, err := ghClient.Repositories.GetContents(ctx, ownerName, repoName, filePath, nil)
	if err != nil {
		var ghErrorResponse *github.ErrorResponse

		if errors.As(err, &ghErrorResponse) {
			e := goGitHubErr(*ghErrorResponse)
			return "", &e
		}
		return "", err
	}
	if content == nil {
		return "", &noContentError{path: filePath, owner: ownerName, repo: repoName, err: ErrNoContent}
	}
	if content.GetType() != "file" {
		return "", &notAFileError{path: filePath, owner: ownerName, repo: repoName, err: ErrNotAFile}
	}

	return content.GetContent()
}
