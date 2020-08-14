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
	"reflect"
	"testing"
	"time"

	"github.com/GoogleCloudPlatform/devrel-services/leif/githubservices"
	"github.com/google/go-github/github"
)

var file = "file"
var base64 = "base64"
var fileName = sloConfigFileName
var malformedSLORules = `{"some sort of json"}`

var sloRulesSample = `[{
	"appliesTo": {
		"gitHubLabels": ["priority: P0", "bug"]
	},
	"complianceSettings": {
		"responseTime": 0,
		"resolutionTime": 0,
		"requiresAssignee": true
	}
}]`

var sloRulesSampleParsed = []*SLORule{&SLORule{
	AppliesTo: AppliesTo{
		GitHubLabels: []string{"priority: P0", "bug"},
		Issues:       true,
		PRs:          false,
	},
	ComplianceSettings: ComplianceSettings{
		ResponseTime:     0,
		ResolutionTime:   0,
		RequiresAssignee: true,
		Responders:       []string{"MyOwner"},
	},
}}

func TestFetchFile(t *testing.T) {
	tests := []struct {
		name        string
		ownerName   string
		repoName    string
		filePath    string
		mockContent *github.RepositoryContent
		mockError   error
		expected    string
		wantErr     error
	}{
		{
			name:      "Fetch empty file",
			ownerName: "Google",
			repoName:  "MyRepo",
			filePath:  "file.json",
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
			name:      "Fetches file with content",
			ownerName: "Google",
			repoName:  "MyRepo",
			filePath:  "file.json",
			mockContent: &github.RepositoryContent{
				Type:    &file,
				Name:    &fileName,
				Content: &malformedSLORules,
			},
			mockError: nil,
			expected:  malformedSLORules,
			wantErr:   nil,
		},
		{
			name:      "Fails if content is not a file",
			ownerName: "Google",
			repoName:  "MyRepo",
			filePath:  "directory",
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
			ownerName:   "Google",
			repoName:    "MyRepo",
			filePath:    "no-content",
			mockContent: nil,
			mockError:   nil,
			expected:    "",
			wantErr:     ErrNoContent,
		},
		{
			name:      "Errors if file is not found",
			ownerName: "Google",
			repoName:  "MyRepo",
			filePath:  "dna",
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
			wantErr:  ErrGoGitHub,
		},
	}
	for _, test := range tests {
		mock := new(githubservices.MockGithubRepositoryService)
		mock.Owner = test.ownerName
		mock.Repo = test.repoName
		mock.Content = test.mockContent
		mock.Error = test.mockError

		client := githubservices.NewClient(nil, mock, nil)

		ctx := context.Background()
		got, gotErr := fetchFile(ctx, test.ownerName, test.repoName, test.filePath, &client)

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
		{
			name:     "Find empty file returns empty rules",
			owner:    Owner{name: "Google"},
			repoName: "MyRepo",
			mockContent: &github.RepositoryContent{
				Type:     &file,
				Encoding: &base64,
				Name:     &fileName,
				Content:  new(string),
			},
			mockError: nil,
			expected:  nil,
			wantErr:   false,
		},
		{
			name:     "File with SLO rules returns them",
			owner:    Owner{name: "MyOwner"},
			repoName: "MyRepo",
			mockContent: &github.RepositoryContent{
				Type:    &file,
				Name:    &fileName,
				Content: &sloRulesSample,
			},
			mockError: nil,
			expected:  sloRulesSampleParsed,
			wantErr:   false,
		},
		{
			name:     "Find file with malformed content fails",
			owner:    Owner{name: "Google"},
			repoName: "MyRepo",
			mockContent: &github.RepositoryContent{
				Type:    &file,
				Name:    &fileName,
				Content: &malformedSLORules,
			},
			mockError: nil,
			expected:  nil,
			wantErr:   true,
		},
		{
			name: "Empty but found file does not take owner rules",
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
					},
				}},
			},
			repoName: "MyRepo",
			mockContent: &github.RepositoryContent{
				Type:    &file,
				Name:    &fileName,
				Content: new(string),
			},
			mockError: nil,
			expected:  nil,
			wantErr:   false,
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
				Message: "Not Found",
			},
			expected: []*SLORule{&SLORule{
				AppliesTo: AppliesTo{
					Issues: true,
					PRs:    false,
				},
				ComplianceSettings: ComplianceSettings{
					ResponseTime:   time.Hour,
					ResolutionTime: time.Second,
				},
			}},
			wantErr: false,
		},

		{
			name:        "File not found passes even if owner has no rules",
			owner:       Owner{name: "Google"},
			repoName:    "MyRepo",
			mockContent: nil,
			mockError: &github.ErrorResponse{
				Response: &http.Response{
					StatusCode: 404,
					Status:     "404 Not Found",
					Request:    &http.Request{},
				},
				Message: "Not Found",
			},
			expected: nil,
			wantErr:  false,
		},
		{
			name: "Error other than 404 returns an error with no SLO rules",
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
					},
				}},
			},
			repoName:    "MyRepo",
			mockContent: nil,
			mockError: &github.ErrorResponse{
				Response: &http.Response{
					StatusCode: 401,
					Status:     "401 Unauthorized",
					Request:    &http.Request{},
				},
				Message: "Unauthorized",
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Error other than GH Err Response",
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
					},
				}},
			},
			repoName: "MyRepo",
			mockContent: &github.RepositoryContent{
				Type:    &file,
				Name:    &fileName,
				Content: new(string),
			},
			mockError: fmt.Errorf("other error"),
			expected:  nil,
			wantErr:   true,
		},
	}

	for _, test := range tests {
		mock := new(githubservices.MockGithubRepositoryService)

		mock.Owner = test.owner.name
		mock.Repo = test.repoName
		mock.Content = test.mockContent
		mock.Error = test.mockError
		client := githubservices.NewClient(nil, mock, nil)

		got, gotErr := findSLODoc(context.Background(), test.owner, test.repoName, &client)

		if !reflect.DeepEqual(got, test.expected) {
			t.Errorf("%v did not pass.\n\tWant:\t%v\n\tGot:\t%v", test.name, test.expected, got)
		}

		if (gotErr != nil && !test.wantErr) || (gotErr == nil && test.wantErr) {
			t.Errorf("%v did not pass.\n\tWant Err: %v \n\tGot Err: %v", test.name, test.wantErr, gotErr)
		}
	}
}
