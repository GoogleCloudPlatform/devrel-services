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
	"net/http"
	"reflect"
	"testing"

	"github.com/GoogleCloudPlatform/devrel-services/leif/githubservices"
	"github.com/google/go-github/github"
)

func TestOwnerUpdate(t *testing.T) {
	tests := []struct {
		name        string
		owner       Owner
		mockContent *github.RepositoryContent
		mockError   error
		expected    Owner
		wantErr     bool
	}{
		{
			name: "Correctly updates an owner with no repos under it",
			owner: Owner{
				name: "MyOwner",
			},
			mockContent: &github.RepositoryContent{
				Type:    &file,
				Name:    &fileName,
				Content: &sloRulesSample,
			},
			mockError: nil,
			expected: Owner{
				name:     "MyOwner",
				SLORules: sloRulesSampleParsed,
			},
			wantErr: false,
		},
		{
			name: "Correctly updates an owner and a repo under it",
			owner: Owner{
				name:  "MyOwner",
				Repos: []*Repository{&Repository{name: "aRepo"}},
			},
			mockContent: &github.RepositoryContent{
				Type:    &file,
				Name:    &fileName,
				Content: &sloRulesSample,
			},
			mockError: nil,
			expected: Owner{
				name: "MyOwner",
				Repos: []*Repository{&Repository{
					name:     "aRepo",
					SLORules: sloRulesSampleParsed,
				}},
				SLORules: sloRulesSampleParsed,
			},
			wantErr: false,
		},
		{
			name: "Correctly updates an owner and several repos under it",
			owner: Owner{
				name:  "MyOwner",
				Repos: []*Repository{&Repository{name: "aRepo"}, &Repository{name: "aSecondRepo"}},
			},
			mockContent: &github.RepositoryContent{
				Type:    &file,
				Name:    &fileName,
				Content: &sloRulesSample,
			},
			mockError: nil,
			expected: Owner{
				name: "MyOwner",
				Repos: []*Repository{&Repository{
					name:     "aRepo",
					SLORules: sloRulesSampleParsed,
				}, &Repository{
					name:     "aSecondRepo",
					SLORules: sloRulesSampleParsed,
				}},
				SLORules: sloRulesSampleParsed,
			},
			wantErr: false,
		},
		{
			name: "Errors on malformed config",
			owner: Owner{
				name:  "MyOwner",
				Repos: []*Repository{&Repository{name: "aRepo"}},
			},
			mockContent: &github.RepositoryContent{
				Type:    &file,
				Name:    &fileName,
				Content: &malformedSLORules,
			},
			mockError: nil,
			expected: Owner{
				name:  "MyOwner",
				Repos: []*Repository{&Repository{name: "aRepo"}},
			},
			wantErr: true,
		},
		{
			name: "Updates repos even if owner does not have slo doc",
			owner: Owner{
				name:  "MyOwner",
				Repos: []*Repository{&Repository{name: "aRepo"}},
			},
			mockContent: nil,
			mockError: &github.ErrorResponse{
				Response: &http.Response{
					StatusCode: 404,
					Status:     "404 Not Found",
					Request:    &http.Request{},
				},
				Message: "GH error",
			},
			expected: Owner{
				name:  "MyOwner",
				Repos: []*Repository{&Repository{name: "aRepo"}},
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		mock := new(githubservices.MockGithubRepositoryService)

		mock.Owner = test.owner.name
		mock.Content = test.mockContent
		mock.Error = test.mockError
		client := githubservices.NewClient(nil, mock, nil)

		gotOwner := test.owner
		gotErr := gotOwner.Update(context.Background(), &client)

		if !reflect.DeepEqual(gotOwner, test.expected) {
			t.Errorf("%v did not pass.\n\tWant:\t%v\n\tGot:\t%v", test.name, test.expected, gotOwner)
		}

		if (gotErr != nil && !test.wantErr) || (gotErr == nil && test.wantErr) {
			t.Errorf("%v did not pass.\n\tWant Err: %v \n\tGot Err: %v", test.name, test.wantErr, gotErr)
		}
	}
}

func TestOwnerTrackRepo(t *testing.T) {
	tests := []struct {
		name      string
		owner     Owner
		repoName  string
		mockError error
		expected  Owner
		wantErr   bool
	}{
		{
			name: "Correctly tracks a repo",
			owner: Owner{
				name: "MyOwner",
			},
			repoName:  "someRepo",
			mockError: nil,
			expected: Owner{
				name: "MyOwner",
				Repos: []*Repository{
					&Repository{name: "someRepo"},
				},
			},
			wantErr: false,
		},
		{
			name: "Does not track a repo that does not exist",
			owner: Owner{
				name: "MyOwner",
			},
			repoName: "someRepo",
			mockError: &github.ErrorResponse{
				Response: &http.Response{
					StatusCode: 404,
					Status:     "404 Not Found",
					Request:    &http.Request{},
				},
				Message: "GH error",
			},
			expected: Owner{
				name: "MyOwner",
			},
			wantErr: true,
		},
		{
			name: "Does not re-track a duplicate repo",
			owner: Owner{
				name: "MyOwner",
				Repos: []*Repository{
					&Repository{name: "someRepo"},
				},
			},
			repoName:  "someRepo",
			mockError: nil,
			expected: Owner{
				name: "MyOwner",
				Repos: []*Repository{
					&Repository{name: "someRepo"},
				},
			},
			wantErr: true,
		},
		{
			name: "Tracks a repo when tracking other repos",
			owner: Owner{
				name: "MyOwner",
				Repos: []*Repository{
					&Repository{name: "someRepo"},
				},
			},
			repoName:  "aDifferentRepo",
			mockError: nil,
			expected: Owner{
				name: "MyOwner",
				Repos: []*Repository{
					&Repository{name: "aDifferentRepo"},
					&Repository{name: "someRepo"},
				},
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		mock := new(githubservices.MockGithubRepositoryService)

		mock.Owner = test.owner.name
		mock.Error = test.mockError
		client := githubservices.NewClient(nil, mock, nil)

		gotOwner := test.owner
		gotErr := gotOwner.trackRepo(context.Background(), test.repoName, &client)

		if !reflect.DeepEqual(gotOwner, test.expected) {
			t.Errorf("%v did not pass.\n\tWant:\t%v\n\tGot:\t%v", test.name, test.expected, gotOwner)
		}

		if (gotErr != nil && !test.wantErr) || (gotErr == nil && test.wantErr) {
			t.Errorf("%v did not pass.\n\tWant Err: %v \n\tGot Err: %v", test.name, test.wantErr, gotErr)
		}
	}
}
