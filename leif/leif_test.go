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

func TestTrackOwner(t *testing.T) {
	tests := []struct {
		name       string
		corpus     Corpus
		ownerName  string
		mockError  error
		wantCorpus Corpus
		wantOwner  *Owner
		wantErr    bool
	}{
		{
			name:      "Correctly tracks an owner",
			corpus:    Corpus{},
			ownerName: "someOwner",
			mockError: nil,
			wantCorpus: Corpus{
				watchedOwners: []*Owner{&Owner{name: "someOwner"}},
			},
			wantOwner: &Owner{name: "someOwner"},
			wantErr:   false,
		},
		{
			name:      "Does not track an owner that does not exist",
			corpus:    Corpus{},
			ownerName: "someOwner",
			mockError: &github.ErrorResponse{
				Response: &http.Response{
					StatusCode: 404,
					Status:     "404 Not Found",
					Request:    &http.Request{},
				},
				Message: "GH error",
			},
			wantCorpus: Corpus{},
			wantOwner:  nil,
			wantErr:    true,
		},
		{
			name: "Does not re-track an already tracked owner",
			corpus: Corpus{
				watchedOwners: []*Owner{&Owner{name: "someOwner"}},
			},
			ownerName: "someOwner",
			mockError: nil,
			wantCorpus: Corpus{
				watchedOwners: []*Owner{&Owner{name: "someOwner"}},
			},
			wantOwner: &Owner{name: "someOwner"},
			wantErr:   false,
		},
		{
			name: "Tracks an owner when tracking other owners",
			corpus: Corpus{
				watchedOwners: []*Owner{&Owner{name: "someOwner"}},
			},
			ownerName: "anotherOwner",
			mockError: nil,
			wantCorpus: Corpus{
				watchedOwners: []*Owner{
					&Owner{name: "anotherOwner"},
					&Owner{name: "someOwner"},
				},
			},
			wantOwner: &Owner{name: "anotherOwner"},
			wantErr:   false,
		},
	}
	for _, test := range tests {
		mock := new(githubservices.MockGithubUserService)

		mock.User = test.ownerName
		mock.Error = test.mockError
		client := githubservices.NewClient(nil, nil, mock)

		gotCorpus := test.corpus
		gotOwner, gotErr := gotCorpus.trackOwner(context.Background(), test.ownerName, &client)

		if !reflect.DeepEqual(gotCorpus, test.wantCorpus) {
			t.Errorf("%v did not pass.\n\tWant corpus:\t%v\n\tGot corpus:\t%v", test.name, test.wantCorpus, gotCorpus)
		}

		if !reflect.DeepEqual(gotOwner, test.wantOwner) {
			t.Errorf("%v did not pass.\n\tWant owner:\t%v\n\tGot owner:\t%v", test.name, test.wantOwner, gotOwner)
		}

		inCorpus := false
		for _, o := range gotCorpus.watchedOwners {
			if o == gotOwner {
				inCorpus = true
				break
			}
		}

		if gotOwner != nil && !inCorpus {
			t.Errorf("%v did not pass.\n\tReturned owner is not in corpus", test.name)
		}

		if (gotErr != nil && !test.wantErr) || (gotErr == nil && test.wantErr) {
			t.Errorf("%v did not pass.\n\tWant Err: %v \n\tGot Err: %v", test.name, test.wantErr, gotErr)
		}
	}
}

func TestTrackRepo(t *testing.T) {
	tests := []struct {
		name          string
		corpus        Corpus
		ownerName     string
		repoName      string
		mockUserError error
		mockRepoError error
		wantCorpus    Corpus
		wantErr       bool
	}{
		{
			name:          "Correctly tracks a repo",
			corpus:        Corpus{},
			ownerName:     "someOwner",
			repoName:      "someRepo",
			mockUserError: nil,
			mockRepoError: nil,
			wantCorpus: Corpus{
				watchedOwners: []*Owner{
					&Owner{
						name:  "someOwner",
						Repos: []*Repository{&Repository{name: "someRepo", ownerName: "someOwner"}},
					}},
			},
			wantErr: false,
		},
		{
			name:          "Correctly tracks a repo on existing owner",
			corpus:        Corpus{watchedOwners: []*Owner{&Owner{name: "someOwner"}}},
			ownerName:     "someOwner",
			repoName:      "someRepo",
			mockUserError: nil,
			mockRepoError: nil,
			wantCorpus: Corpus{
				watchedOwners: []*Owner{
					&Owner{
						name:  "someOwner",
						Repos: []*Repository{&Repository{name: "someRepo", ownerName: "someOwner"}},
					}},
			},
			wantErr: false,
		},
		{
			name: "Correctly tracks a repo on existing owner with repos",
			corpus: Corpus{watchedOwners: []*Owner{&Owner{
				name:  "someOwner",
				Repos: []*Repository{&Repository{name: "aRepo", ownerName: "someOwner"}},
			}}},
			ownerName:     "someOwner",
			repoName:      "someRepo",
			mockUserError: nil,
			mockRepoError: nil,
			wantCorpus: Corpus{
				watchedOwners: []*Owner{
					&Owner{
						name: "someOwner",
						Repos: []*Repository{
							&Repository{name: "aRepo", ownerName: "someOwner"},
							&Repository{name: "someRepo", ownerName: "someOwner"},
						},
					}},
			},
			wantErr: false,
		},
		{
			name:      "Errors if owner does not exist",
			corpus:    Corpus{},
			ownerName: "someOwner",
			repoName:  "someRepo",
			mockUserError: &github.ErrorResponse{
				Response: &http.Response{
					StatusCode: 404,
					Status:     "404 Not Found",
					Request:    &http.Request{},
				},
				Message: "GH error",
			},
			mockRepoError: nil,
			wantCorpus:    Corpus{},
			wantErr:       true,
		},
		{
			name:          "Errors if repo does not exist",
			corpus:        Corpus{watchedOwners: []*Owner{&Owner{name: "someOwner"}}},
			ownerName:     "someOwner",
			repoName:      "someRepo",
			mockUserError: nil,
			mockRepoError: &github.ErrorResponse{
				Response: &http.Response{
					StatusCode: 404,
					Status:     "404 Not Found",
					Request:    &http.Request{},
				},
				Message: "GH error",
			},
			wantCorpus: Corpus{watchedOwners: []*Owner{&Owner{name: "someOwner"}}},
			wantErr:    true,
		},
	}
	for _, test := range tests {
		repoMock := new(githubservices.MockGithubRepositoryService)
		repoMock.Owner = test.ownerName
		repoMock.Error = test.mockRepoError

		userMock := new(githubservices.MockGithubUserService)
		userMock.User = test.ownerName
		userMock.Error = test.mockUserError

		client := githubservices.NewClient(nil, repoMock, userMock)

		gotCorpus := test.corpus
		gotErr := gotCorpus.TrackRepo(context.Background(), test.ownerName, test.repoName, &client)

		if !reflect.DeepEqual(gotCorpus, test.wantCorpus) {
			t.Errorf("%v did not pass.\n\tWant corpus:\t%v\n\tGot corpus:\t%v", test.name, test.wantCorpus, gotCorpus)
		}

		if (gotErr != nil && !test.wantErr) || (gotErr == nil && test.wantErr) {
			t.Errorf("%v did not pass.\n\tWant Err: %v \n\tGot Err: %v", test.name, test.wantErr, gotErr)
		}
	}
}
