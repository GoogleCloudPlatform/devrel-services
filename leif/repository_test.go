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

func TestRepoUpdate(t *testing.T) {
	tests := []struct {
		name        string
		owner       Owner
		repo        Repository
		mockContent *github.RepositoryContent
		mockError   error
		expected    Repository
		wantErr     bool
	}{
		{
			name: "Correctly updates a repo",
			owner: Owner{
				name: "MyOwner",
			},
			repo: Repository{
				name: "MyRepo",
			},
			mockContent: &github.RepositoryContent{
				Type:    &file,
				Name:    &fileName,
				Content: &sloRulesSample,
			},
			mockError: nil,
			expected: Repository{
				name:     "MyRepo",
				SLORules: sloRulesSampleParsed,
			},
			wantErr: false,
		},
		{
			name: "Takes owner rules if file is not found",
			owner: Owner{
				name:     "MyOwner",
				SLORules: sloRulesSampleParsed,
			},
			repo: Repository{
				name: "MyRepo",
			},
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
				Message: "GH error",
			},
			expected: Repository{
				name:     "MyRepo",
				SLORules: sloRulesSampleParsed,
			},
			wantErr: false,
		},
		{
			name: "SLO file with malformed content fails",
			owner: Owner{
				name: "MyOwner",
			},
			repo: Repository{
				name:     "MyRepo",
				SLORules: sloRulesSampleParsed,
			},
			mockContent: &github.RepositoryContent{
				Type:    &file,
				Name:    &fileName,
				Content: &malformedSLORules,
			},
			mockError: nil,
			expected: Repository{
				name:     "MyRepo",
				SLORules: sloRulesSampleParsed,
			},
			wantErr: true,
		},
		{
			name: "Unauthorized errors",
			owner: Owner{
				name: "MyOwner",
			},
			repo: Repository{
				name: "MyRepo",
			},
			mockContent: &github.RepositoryContent{
				Type:    &file,
				Name:    &fileName,
				Content: &sloRulesSample,
			},
			mockError: &github.ErrorResponse{
				Response: &http.Response{
					StatusCode: 401,
					Status:     "401 Unauthorized",
					Request:    &http.Request{},
				},
				Message: "GH error",
			},
			expected: Repository{
				name: "MyRepo",
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		mock := new(githubservices.MockGithubRepositoryService)

		mock.Owner = test.owner.name
		mock.Repo = test.repo.name
		mock.Content = test.mockContent
		mock.Error = test.mockError
		client := githubservices.NewClient(nil, mock, nil)

		gotRepo := test.repo
		gotErr := gotRepo.Update(context.Background(), test.owner, &client)

		if !reflect.DeepEqual(gotRepo, test.expected) {
			t.Errorf("%v did not pass.\n\tWant:\t%v\n\tGot:\t%v", test.name, test.expected, gotRepo)
		}

		if (gotErr != nil && !test.wantErr) || (gotErr == nil && test.wantErr) {
			t.Errorf("%v did not pass.\n\tWant Err: %v \n\tGot Err: %v", test.name, test.wantErr, gotErr)
		}
	}
}
