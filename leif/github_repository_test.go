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

	goGH "github.com/google/go-github/v32/github"
)

var file = "file"
var base64 = "base64"
var fileName = sloConfigFileName
var jsonString = `{"some sort of json"}`

var notAFileErr *notAFileError
var ghErrorResponse *goGH.ErrorResponse
var noContentErr *noContentError

type MockGithubRepositoryService struct {
	Content    *goGH.RepositoryContent
	DirContent []*goGH.RepositoryContent
	Response   *goGH.Response
	Error      error
	Org        string
	Repo       string
}

func (mgc *MockGithubRepositoryService) GetContents(ctx context.Context, org, repo, path string, opts *goGH.RepositoryContentGetOptions) (*goGH.RepositoryContent, []*goGH.RepositoryContent, *goGH.Response, error) {
	if org != mgc.Org {
		return nil, nil, nil, errors.New("org did not equal expected org: was: " + org)
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
		mockContent *goGH.RepositoryContent
		mockError   error
		expected    string
		wantErr     error
	}{
		{
			name:     "Fetch empty file",
			orgName:  "Google",
			repoName: "MyRepo",
			filePath: "file.json",
			mockContent: &goGH.RepositoryContent{
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
			mockContent: &goGH.RepositoryContent{
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
			mockContent: &goGH.RepositoryContent{
				Type:     new(string),
				Encoding: &base64,
				Name:     &fileName,
				Content:  new(string),
			},
			mockError: nil,
			expected:  "",
			wantErr:   notAFileErr,
		},
		{
			name:        "Errors if no file content provided",
			orgName:     "Google",
			repoName:    "MyRepo",
			filePath:    "no-content",
			mockContent: nil,
			mockError:   nil,
			expected:    "",
			wantErr:     noContentErr,
		},
		{
			name:     "Errors if file is not found",
			orgName:  "Google",
			repoName: "MyRepo",
			filePath: "dna",
			mockContent: &goGH.RepositoryContent{
				Type:     &file,
				Encoding: &base64,
				Name:     &fileName,
				Content:  new(string),
			},
			mockError: &goGH.ErrorResponse{Response: &http.Response{StatusCode: 404}},
			expected:  "",
			wantErr:   ghErrorResponse,
		},
	}
	for _, test := range tests {
		mock := new(MockGithubRepositoryService)

		mock.Org = test.orgName
		mock.Repo = test.repoName
		mock.Content = test.mockContent
		mock.Error = test.mockError
		client := NewGitHubClient(nil, mock)

		ctx := context.Background()
		got, gotErr := fetchFile(ctx, test.orgName, test.repoName, test.filePath, &client)

		if !reflect.DeepEqual(got, test.expected) {
			t.Errorf("%v did not pass.\n\tWant:\t%v\n\tGot:\t%v", test.name, test.expected, got)
		}
		if (test.wantErr == nil && gotErr != nil) || (test.wantErr != nil && reflect.TypeOf(gotErr) != reflect.TypeOf(test.wantErr)) {
			t.Errorf("%v did not pass.\n\tWant Err: %v \n\tGot Err: %v", test.name, reflect.TypeOf(test.wantErr), reflect.TypeOf(gotErr))
		}
	}
}

func TestFindSLODoc(t *testing.T) {
	tests := []struct {
		name         string
		orgName      string
		repoName     string
		mockContent  *goGH.RepositoryContent
		mockError    error
		currentRepo  *Repository
		expectedRepo *Repository
		expectedPath string
		wantErr      error
	}{
		{
			name:     "Find file",
			orgName:  "Google",
			repoName: "MyRepo",
			mockContent: &goGH.RepositoryContent{
				Type:     &file,
				Encoding: &base64,
				Name:     &fileName,
				Content:  new(string),
			},
			mockError:   nil,
			currentRepo: &Repository{},
			expectedRepo: &Repository{
				SLOFileContent: "",
			},
			expectedPath: "Google/MyRepo/.github/" + sloConfigFileName,
			wantErr:      nil,
		},
		{
			name:     "Find file with content",
			orgName:  "Google",
			repoName: "MyRepo",
			mockContent: &goGH.RepositoryContent{
				Type:    &file,
				Name:    &fileName,
				Content: &jsonString,
			},
			mockError:   nil,
			currentRepo: &Repository{},
			expectedRepo: &Repository{
				SLOFileContent: jsonString,
			},
			expectedPath: "Google/MyRepo/.github/" + sloConfigFileName,
			wantErr:      nil,
		},
		{
			name:     "File not found fails after looking at org level",
			orgName:  "Google",
			repoName: "MyRepo",
			mockContent: &goGH.RepositoryContent{
				Type:    &file,
				Name:    &fileName,
				Content: &jsonString,
			},
			mockError:    &goGH.ErrorResponse{Response: &http.Response{StatusCode: 404}},
			currentRepo:  &Repository{},
			expectedRepo: &Repository{},
			expectedPath: "Google/.github/" + sloConfigFileName,
			wantErr:      ghErrorResponse,
		},
	}

	for _, test := range tests {
		mock := new(MockGithubRepositoryService)

		mock.Org = test.orgName
		mock.Repo = test.repoName
		mock.Content = test.mockContent
		mock.Error = test.mockError
		client := NewGitHubClient(nil, mock)

		got := test.currentRepo
		ctx := context.Background()
		gotPath, gotErr := got.findSLODoc(ctx, test.orgName, test.repoName, &client)

		if !reflect.DeepEqual(got, test.expectedRepo) {
			t.Errorf("%v did not pass.\n\tWant:\t%v\n\tGot:\t%v", test.name, test.expectedRepo, got)
		}
		if !reflect.DeepEqual(gotPath, test.expectedPath) {
			t.Errorf("%v did not pass.\n\tWant path:\t%v\n\tGot path:\t%v", test.name, test.expectedPath, gotPath)
		}
		if (test.wantErr == nil && gotErr != nil) || (test.wantErr != nil && reflect.TypeOf(gotErr) != reflect.TypeOf(test.wantErr)) {
			t.Errorf("%v did not pass.\n\tWant Err: %v \n\tGot Err: %v", test.name, reflect.TypeOf(test.wantErr), reflect.TypeOf(gotErr))
		}
	}
}

func TestParseSLOs(t *testing.T) {
	tests := []struct {
		name         string
		currentRepo  *Repository
		expectedRepo *Repository
		wantErr      error
	}{
		{
			name: "Parses empty file",
			currentRepo: &Repository{
				SLOFileContent: "",
			},
			expectedRepo: &Repository{
				SLOFileContent: "",
			},
			wantErr: nil,
		},
		{
			name: "Malformed json returns error",
			currentRepo: &Repository{
				SLOFileContent: jsonString,
			},
			expectedRepo: &Repository{
				SLOFileContent: jsonString,
			},
			wantErr: syntaxError,
		},
		{
			name: "json with one rule is parsed",
			currentRepo: &Repository{
				SLOFileContent: `[
					{
						"appliesTo": {},
						"complianceSettings": {
							"responseTime": 0,
							"resolutionTime": 0
						}
					}
				]`,
			},
			expectedRepo: &Repository{
				SLOFileContent: `[
					{
						"appliesTo": {},
						"complianceSettings": {
							"responseTime": 0,
							"resolutionTime": 0
						}
					}
				]`,
				SLORules: []*SLORule{&defaultSLO},
			},
			wantErr: nil,
		},
		{
			name: "json with several rules is parsed",
			currentRepo: &Repository{
				SLOFileContent: `[
					{
						"appliesTo": {},
						"complianceSettings": {
							"responseTime": 0,
							"resolutionTime": 0
						}
					},
					{
						"appliesTo": {},
						"complianceSettings": {
							"responseTime": 0,
							"resolutionTime": 0
						}
					}
				 ]`,
			},
			expectedRepo: &Repository{
				SLOFileContent: `[
					{
						"appliesTo": {},
						"complianceSettings": {
							"responseTime": 0,
							"resolutionTime": 0
						}
					},
					{
						"appliesTo": {},
						"complianceSettings": {
							"responseTime": 0,
							"resolutionTime": 0
						}
					}
				 ]`,
				SLORules: []*SLORule{&defaultSLO, &defaultSLO},
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {

		got := test.currentRepo

		gotErr := got.ParseSLOs()

		if !reflect.DeepEqual(got, test.expectedRepo) {
			t.Errorf("%v did not pass.\n\tWant:\t%v\n\tGot:\t%v", test.name, test.expectedRepo, got)
		}
		if (test.wantErr == nil && gotErr != nil) || (test.wantErr != nil && reflect.TypeOf(gotErr) != reflect.TypeOf(test.wantErr)) {
			t.Errorf("%v did not pass.\n\tWant Err: %v \n\tGot Err: %v", test.name, reflect.TypeOf(test.wantErr), reflect.TypeOf(gotErr))
		}
	}
}
