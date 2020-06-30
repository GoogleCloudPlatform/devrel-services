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
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-github/github"
)

var file = "file"
var base64 = "base64"
var fileName = sloConfigFileName
var jsonString = `{"some sort of json"}`

type MockGithubRepositoryService struct {
	Content    *github.RepositoryContent
	DirContent []*github.RepositoryContent
	Response   *github.Response
	Error      error
	Owner      string
	Repo       string
}

func (mgc *MockGithubRepositoryService) GetContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error) {
	if owner != mgc.Owner {
		return nil, nil, nil, errors.New("owner did not equal expected owner: was: " + owner)
	}
	if repo != mgc.Repo && repo != ".github" {
		return nil, nil, nil, errors.New("repo did not equal expected repo: was: " + repo)
	}
	return mgc.Content, mgc.DirContent, mgc.Response, mgc.Error
}

func TestFetchFile(t *testing.T) {
	tests := []struct {
		name        string
		orgName     string
		repoName    string
		filePath    string
		mockContent *github.RepositoryContent
		mockError   error
		expected    string
		wantErr     error
	}{
		{
			name:     "Fetch empty file",
			orgName:  "Google",
			repoName: "MyRepo",
			filePath: "file.json",
			mockContent: &github.RepositoryContent{
				Type:     &file,
				Encoding: &base64,
				Name:     &fileName,
				Content:  new(string),
			},
			mockError: nil,
			expected:  "",
			wantErr:   nil,
		},
		{
			name:     "Fetches file with content",
			orgName:  "Google",
			repoName: "MyRepo",
			filePath: "file.json",
			mockContent: &github.RepositoryContent{
				Type:    &file,
				Name:    &fileName,
				Content: &jsonString,
			},
			mockError: nil,
			expected:  jsonString,
			wantErr:   nil,
		},
		{
			name:     "Fails if content is not a file",
			orgName:  "Google",
			repoName: "MyRepo",
			filePath: "directory",
			mockContent: &github.RepositoryContent{
				Type:     new(string),
				Encoding: &base64,
				Name:     &fileName,
				Content:  new(string),
			},
			mockError: nil,
			expected:  "",
			wantErr:   ErrNotAFile,
		},
		{
			name:        "Errors if no file content provided",
			orgName:     "Google",
			repoName:    "MyRepo",
			filePath:    "no-content",
			mockContent: nil,
			mockError:   nil,
			expected:    "",
			wantErr:     ErrNoContent,
		},
		{
			name:     "Errors if file is not found",
			orgName:  "Google",
			repoName: "MyRepo",
			filePath: "dna",
			mockContent: &github.RepositoryContent{
				Type:     &file,
				Encoding: &base64,
				Name:     &fileName,
				Content:  new(string),
			},
			mockError: &github.ErrorResponse{
				Response: &http.Response{
					StatusCode: 404,
					Status:     "404 Not Found",
					Request:    &http.Request{},
				},
				Message: "GH error"},
			expected: "",
			wantErr:  GoGitHubErr,
		},
	}
	for _, test := range tests {
		mock := new(MockGithubRepositoryService)
		mock.Owner = test.orgName
		mock.Repo = test.repoName
		mock.Content = test.mockContent
		mock.Error = test.mockError

		client := NewGitHubClient(nil, mock)

		ctx := context.Background()
		got, gotErr := fetchFile(ctx, test.orgName, test.repoName, test.filePath, &client)

		if !reflect.DeepEqual(got, test.expected) {
			t.Errorf("%v did not pass.\n\tWant:\t%v\n\tGot:\t%v", test.name, test.expected, got)
		}

		if !errors.Is(gotErr, test.wantErr) {
			t.Errorf("%v did not pass.\n\tWant Err: %v \n\tGot Err: %v", test.name, test.wantErr, gotErr)
		}
	}
}

func TestFindSLODoc(t *testing.T) {
	tests := []struct {
		name        string
		owner       Owner
		repoName    string
		mockContent *github.RepositoryContent
		mockError   error
		expected    []*SLORule
		wantErr     bool
	}{
		// {
		// 	name:     "Find empty file returns empty rules",
		// 	owner:    Owner{name: "Google"},
		// 	repoName: "MyRepo",
		// 	mockContent: &github.RepositoryContent{
		// 		Type:     &file,
		// 		Encoding: &base64,
		// 		Name:     &fileName,
		// 		Content:  new(string),
		// 	},
		// 	mockError: nil,
		// 	expected:  nil,
		// 	wantErr:   false,
		// },
		{
			name:     "Find file with malformed content fails",
			owner:    Owner{name: "Google"},
			repoName: "MyRepo",
			mockContent: &github.RepositoryContent{
				Type:    &file,
				Name:    &fileName,
				Content: &jsonString,
			},
			mockError: nil,
			expected:  nil,
			wantErr:   true,
		},
		{
			name: "File not found takes owner rules",
			owner: Owner{
				name: "Google",
				SLORules: []*SLORule{&SLORule{
					AppliesTo: AppliesTo{
						Issues: true,
						PRs:    false,
					},
					ComplianceSettings: ComplianceSettings{
						ResponseTime:   time.Hour,
						ResolutionTime: time.Second,
						Responders:     Responders{Contributors: "WRITE"},
					},
				}},
			},
			repoName:    "MyRepo",
			mockContent: nil,
			mockError: &github.ErrorResponse{
				Response: &http.Response{
					StatusCode: 404,
					Status:     "404 Not Found",
					Request:    &http.Request{},
				},
				Message: "GH error"},
			expected: []*SLORule{&SLORule{
				AppliesTo: AppliesTo{
					Issues: true,
					PRs:    false,
				},
				ComplianceSettings: ComplianceSettings{
					ResponseTime:   time.Hour,
					ResolutionTime: time.Second,
					Responders:     Responders{Contributors: "WRITE"},
				},
			}},
			wantErr: false,
		},
	}

	for _, test := range tests {
		mock := new(MockGithubRepositoryService)

		mock.Owner = test.owner.name
		mock.Repo = test.repoName
		mock.Content = test.mockContent
		mock.Error = test.mockError
		client := NewGitHubClient(nil, mock)

		got, gotErr := findSLODoc(context.Background(), test.owner, test.repoName, &client)

		if !reflect.DeepEqual(got, test.expected) {
			t.Errorf("%v did not pass.\n\tWant:\t%v\n\tGot:\t%v", test.name, test.expected, got)
		}

		if (gotErr != nil && !test.wantErr) || (gotErr == nil && test.wantErr) {
			t.Errorf("%v did not pass.\n\tWant Err: %v \n\tGot Err: %v", test.name, test.wantErr, gotErr)
		}
	}
}
