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
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/GoogleCloudPlatform/devrel-services/leif/githubservices"
	"github.com/google/go-github/github"
)

func TestTrackOwner(t *testing.T) {
	tests := []struct {
		name       string
		currOwners []*Owner
		ownerName  string
		mockError  error
		wantOwners []*Owner
		wantOwner  *Owner
		wantErr    bool
	}{
		{
			name:       "Correctly tracks an owner",
			currOwners: nil,
			ownerName:  "someOwner",
			mockError:  nil,
			wantOwners: []*Owner{&Owner{name: "someOwner"}},
			wantOwner:  &Owner{name: "someOwner"},
			wantErr:    false,
		},
		{
			name:       "Does not track an owner that does not exist",
			currOwners: nil,
			ownerName:  "someOwner",
			mockError: &github.ErrorResponse{
				Response: &http.Response{
					StatusCode: 404,
					Status:     "404 Not Found",
					Request:    &http.Request{},
				},
				Message: "GH error",
			},
			wantOwners: nil,
			wantOwner:  nil,
			wantErr:    true,
		},
		{
			name:       "Does not re-track an already tracked owner",
			currOwners: []*Owner{&Owner{name: "someOwner"}},
			ownerName:  "someOwner",
			mockError:  nil,
			wantOwners: []*Owner{&Owner{name: "someOwner"}},
			wantOwner:  &Owner{name: "someOwner"},
			wantErr:    false,
		},
		{
			name:       "Tracks an owner when tracking other owners",
			currOwners: []*Owner{&Owner{name: "someOwner"}},
			ownerName:  "anotherOwner",
			mockError:  nil,
			wantOwners: []*Owner{
				&Owner{name: "anotherOwner"},
				&Owner{name: "someOwner"},
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

		gotCorpus := Corpus{watchedOwners: test.currOwners}
		gotOwner, gotErr := gotCorpus.trackOwner(context.Background(), test.ownerName, &client)

		if !reflect.DeepEqual(gotCorpus.watchedOwners, test.wantOwners) {
			t.Errorf("%v did not pass.\n\tWant corpus:\t%v\n\tGot corpus:\t%v", test.name, test.wantOwners, gotCorpus.watchedOwners)
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
		currOwners    []*Owner
		ownerName     string
		repoName      string
		mockUserError error
		mockRepoError error
		wantOwners    []*Owner
		wantErr       bool
	}{
		{
			name:          "Correctly tracks a repo",
			currOwners:    nil,
			ownerName:     "someOwner",
			repoName:      "someRepo",
			mockUserError: nil,
			mockRepoError: nil,
			wantOwners: []*Owner{
				&Owner{
					name:  "someOwner",
					Repos: []*Repository{&Repository{name: "someRepo", ownerName: "someOwner"}},
				},
			},
			wantErr: false,
		},
		{
			name:          "Correctly tracks a repo on existing owner",
			currOwners:    []*Owner{&Owner{name: "someOwner"}},
			ownerName:     "someOwner",
			repoName:      "someRepo",
			mockUserError: nil,
			mockRepoError: nil,
			wantOwners: []*Owner{
				&Owner{
					name:  "someOwner",
					Repos: []*Repository{&Repository{name: "someRepo", ownerName: "someOwner"}},
				},
			},
			wantErr: false,
		},
		{
			name: "Correctly tracks a repo on existing owner with repos",
			currOwners: []*Owner{&Owner{
				name:  "someOwner",
				Repos: []*Repository{&Repository{name: "aRepo", ownerName: "someOwner"}},
			}},
			ownerName:     "someOwner",
			repoName:      "someRepo",
			mockUserError: nil,
			mockRepoError: nil,
			wantOwners: []*Owner{
				&Owner{
					name: "someOwner",
					Repos: []*Repository{
						&Repository{name: "aRepo", ownerName: "someOwner"},
						&Repository{name: "someRepo", ownerName: "someOwner"},
					},
				},
			},
			wantErr: false,
		},
		{
			name:       "Errors if owner does not exist",
			currOwners: nil,
			ownerName:  "someOwner",
			repoName:   "someRepo",
			mockUserError: &github.ErrorResponse{
				Response: &http.Response{
					StatusCode: 404,
					Status:     "404 Not Found",
					Request:    &http.Request{},
				},
				Message: "GH error",
			},
			mockRepoError: nil,
			wantOwners:    nil,
			wantErr:       true,
		},
		{
			name:          "Errors if repo does not exist",
			currOwners:    []*Owner{&Owner{name: "someOwner"}},
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
			wantOwners: []*Owner{&Owner{name: "someOwner"}},
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

		gotCorpus := Corpus{watchedOwners: test.currOwners}
		gotErr := gotCorpus.TrackRepo(context.Background(), test.ownerName, test.repoName, &client)

		if !reflect.DeepEqual(gotCorpus.watchedOwners, test.wantOwners) {
			t.Errorf("%v did not pass.\n\tWant owners:\t%v\n\tGot owners:\t%v", test.name, test.wantOwners, gotCorpus.watchedOwners)
		}

		if (gotErr != nil && !test.wantErr) || (gotErr == nil && test.wantErr) {
			t.Errorf("%v did not pass.\n\tWant Err: %v \n\tGot Err: %v", test.name, test.wantErr, gotErr)
		}
	}
}

func TestForEachRepo(t *testing.T) {

	var workingSet map[string]int

	tests := []struct {
		name    string
		corpus  *Corpus
		fn      func(r Repository) error
		wantSet map[string]int
		wantErr bool
	}{
		{
			name: "Iterates through a single repo",
			corpus: &Corpus{watchedOwners: []*Owner{&Owner{
				name: "someOwner",
				Repos: []*Repository{
					&Repository{name: "aRepo", ownerName: "someOwner"},
				},
			}}},
			fn:      func(r Repository) error { workingSet[r.RepoName()]++; return nil },
			wantSet: map[string]int{"aRepo": 1},
			wantErr: false,
		},
		{
			name:    "Iterates through empty corpus",
			corpus:  &Corpus{},
			fn:      func(r Repository) error { workingSet[r.RepoName()]++; return nil },
			wantSet: map[string]int{},
			wantErr: false,
		},
		{
			name: "Iterates through several repos",
			corpus: &Corpus{watchedOwners: []*Owner{&Owner{
				name: "someOwner",
				Repos: []*Repository{
					&Repository{name: "aRepo", ownerName: "someOwner"},
					&Repository{name: "repo2", ownerName: "someOwner"},
					&Repository{name: "repo3", ownerName: "someOwner"},
				},
			}}},
			fn:      func(r Repository) error { workingSet[r.RepoName()]++; return nil },
			wantSet: map[string]int{"aRepo": 1, "repo2": 1, "repo3": 1},
			wantErr: false,
		},
		{
			name: "Iterates through repos with different owners",
			corpus: &Corpus{watchedOwners: []*Owner{
				&Owner{
					name: "someOwner",
					Repos: []*Repository{
						&Repository{name: "aRepo", ownerName: "someOwner"},
						&Repository{name: "repo2", ownerName: "someOwner"},
						&Repository{name: "repo3", ownerName: "someOwner"},
					},
				},
				&Owner{
					name: "anotherOwner",
					Repos: []*Repository{
						&Repository{name: "aRepo", ownerName: "anotherOwner"},
						&Repository{name: "otherRepo", ownerName: "anotherOwner"},
					},
				},
			}},
			fn:      func(r Repository) error { workingSet[r.RepoName()]++; return nil },
			wantSet: map[string]int{"aRepo": 2, "repo2": 1, "repo3": 1, "otherRepo": 1},
			wantErr: false,
		},
		{
			name: "Returns first error from func",
			corpus: &Corpus{watchedOwners: []*Owner{
				&Owner{
					name: "someOwner",
					Repos: []*Repository{
						&Repository{name: "aRepo", ownerName: "someOwner"},
						&Repository{name: "repo2", ownerName: "someOwner"},
					},
				},
			}},
			fn:      func(r Repository) error { workingSet["runCount"]++; return fmt.Errorf("an error") },
			wantSet: map[string]int{"runCount": 1},
			wantErr: true,
		},
	}
	for _, test := range tests {
		workingSet = map[string]int{}

		gotErr := test.corpus.ForEachRepo(test.fn)

		if !reflect.DeepEqual(workingSet, test.wantSet) {
			t.Errorf("%v did not pass.\n\tWant:\t%v\n\tGot:\t%v", test.name, test.wantSet, workingSet)
		}

		if (gotErr != nil && !test.wantErr) || (gotErr == nil && test.wantErr) {
			t.Errorf("%v did not pass.\n\tWant Err: %v \n\tGot Err: %v", test.name, test.wantErr, gotErr)
		}
	}
}

func TestForEachRepoF(t *testing.T) {
	var workingSet map[string]int

	tests := []struct {
		name     string
		corpus   *Corpus
		fn       func(r Repository) error
		filterfn func(r Repository) bool
		wantSet  map[string]int
		wantErr  bool
	}{
		{
			name:     "Iterates through empty corpus",
			corpus:   &Corpus{},
			fn:       func(r Repository) error { workingSet[r.RepoName()]++; return nil },
			filterfn: func(r Repository) bool { return true },
			wantSet:  map[string]int{},
			wantErr:  false,
		},
		{
			name: "Iterates through a single repo",
			corpus: &Corpus{watchedOwners: []*Owner{&Owner{
				name: "someOwner",
				Repos: []*Repository{
					&Repository{name: "aRepo", ownerName: "someOwner"},
				},
			}}},
			fn:       func(r Repository) error { workingSet[r.RepoName()]++; return nil },
			filterfn: func(r Repository) bool { return true },
			wantSet:  map[string]int{"aRepo": 1},
			wantErr:  false,
		},
		{
			name: "Iterates and filters out a single repo",
			corpus: &Corpus{watchedOwners: []*Owner{&Owner{
				name: "someOwner",
				Repos: []*Repository{
					&Repository{name: "aRepo", ownerName: "someOwner"},
				},
			}}},
			fn:       func(r Repository) error { workingSet[r.RepoName()]++; return nil },
			filterfn: func(r Repository) bool { return false },
			wantSet:  map[string]int{},
			wantErr:  false,
		},
		{
			name: "Iterates through several repos",
			corpus: &Corpus{watchedOwners: []*Owner{&Owner{
				name: "someOwner",
				Repos: []*Repository{
					&Repository{name: "aRepo", ownerName: "someOwner"},
					&Repository{name: "repo2", ownerName: "someOwner"},
					&Repository{name: "repo3", ownerName: "someOwner"},
				},
			}}},
			fn:       func(r Repository) error { workingSet[r.RepoName()]++; return nil },
			filterfn: func(r Repository) bool { return true },
			wantSet:  map[string]int{"aRepo": 1, "repo2": 1, "repo3": 1},
			wantErr:  false,
		},
		{
			name: "Iterates through and filters out several repos",
			corpus: &Corpus{watchedOwners: []*Owner{&Owner{
				name: "someOwner",
				Repos: []*Repository{
					&Repository{name: "aRepo", ownerName: "someOwner"},
					&Repository{name: "repo2", ownerName: "someOwner"},
					&Repository{name: "repo3", ownerName: "someOwner"},
				},
			}}},
			fn:       func(r Repository) error { workingSet[r.RepoName()]++; return nil },
			filterfn: func(r Repository) bool { return false },
			wantSet:  map[string]int{},
			wantErr:  false,
		},
		{
			name: "Iterates through several repos and filters one out",
			corpus: &Corpus{watchedOwners: []*Owner{&Owner{
				name: "someOwner",
				Repos: []*Repository{
					&Repository{name: "aRepo", ownerName: "someOwner"},
					&Repository{name: "repo2", ownerName: "someOwner"},
					&Repository{name: "repo3", ownerName: "someOwner"},
				},
			}}},
			fn:       func(r Repository) error { workingSet[r.RepoName()]++; return nil },
			filterfn: func(r Repository) bool { return !(r.RepoName() == "repo3") },
			wantSet:  map[string]int{"aRepo": 1, "repo2": 1},
			wantErr:  false,
		},
		{
			name: "Iterates through several repos and filters several out",
			corpus: &Corpus{watchedOwners: []*Owner{&Owner{
				name: "someOwner",
				Repos: []*Repository{
					&Repository{name: "aRepo", ownerName: "someOwner"},
					&Repository{name: "repo2", ownerName: "someOwner"},
					&Repository{name: "repo3", ownerName: "someOwner"},
				},
			}}},
			fn:       func(r Repository) error { workingSet[r.RepoName()]++; return nil },
			filterfn: func(r Repository) bool { return (r.RepoName() == "repo3") },
			wantSet:  map[string]int{"repo3": 1},
			wantErr:  false,
		},
		{
			name: "Iterates through repos with different owners",
			corpus: &Corpus{watchedOwners: []*Owner{
				&Owner{
					name: "someOwner",
					Repos: []*Repository{
						&Repository{name: "aRepo", ownerName: "someOwner"},
						&Repository{name: "repo2", ownerName: "someOwner"},
						&Repository{name: "repo3", ownerName: "someOwner"},
					},
				},
				&Owner{
					name: "anotherOwner",
					Repos: []*Repository{
						&Repository{name: "aRepo", ownerName: "anotherOwner"},
						&Repository{name: "otherRepo", ownerName: "anotherOwner"},
					},
				},
			}},
			fn:       func(r Repository) error { workingSet[r.RepoName()]++; return nil },
			filterfn: func(r Repository) bool { return true },
			wantSet:  map[string]int{"aRepo": 2, "repo2": 1, "repo3": 1, "otherRepo": 1},
			wantErr:  false,
		},
		{
			name: "Iterates and filters through repos with different owners",
			corpus: &Corpus{watchedOwners: []*Owner{
				&Owner{
					name: "someOwner",
					Repos: []*Repository{
						&Repository{name: "aRepo", ownerName: "someOwner"},
						&Repository{name: "repo2", ownerName: "someOwner"},
						&Repository{name: "repo3", ownerName: "someOwner"},
					},
				},
				&Owner{
					name: "anotherOwner",
					Repos: []*Repository{
						&Repository{name: "aRepo", ownerName: "anotherOwner"},
						&Repository{name: "otherRepo", ownerName: "anotherOwner"},
					},
				},
			}},
			fn:       func(r Repository) error { workingSet[r.RepoName()]++; return nil },
			filterfn: func(r Repository) bool { return (r.RepoName() == "aRepo") },
			wantSet:  map[string]int{"aRepo": 2},
			wantErr:  false,
		},
		{
			name: "Returns first error from func",
			corpus: &Corpus{watchedOwners: []*Owner{
				&Owner{
					name: "someOwner",
					Repos: []*Repository{
						&Repository{name: "aRepo", ownerName: "someOwner"},
						&Repository{name: "repo2", ownerName: "someOwner"},
					},
				},
			}},
			fn:       func(r Repository) error { workingSet["runCount"]++; return fmt.Errorf("an error") },
			filterfn: func(r Repository) bool { return true },
			wantSet:  map[string]int{"runCount": 1},
			wantErr:  true,
		},
	}
	for _, test := range tests {
		workingSet = map[string]int{}

		gotErr := test.corpus.ForEachRepoF(test.fn, test.filterfn)

		if !reflect.DeepEqual(workingSet, test.wantSet) {
			t.Errorf("%v did not pass.\n\tWant:\t%v\n\tGot:\t%v", test.name, test.wantSet, workingSet)
		}

		if (gotErr != nil && !test.wantErr) || (gotErr == nil && test.wantErr) {
			t.Errorf("%v did not pass.\n\tWant Err: %v \n\tGot Err: %v", test.name, test.wantErr, gotErr)
		}
	}
}

func TestForEachRepoFSort(t *testing.T) {

	var workingSlice []string

	tests := []struct {
		name      string
		corpus    *Corpus
		fn        func(r Repository) error
		filterfn  func(r Repository) bool
		sortfn    func([]*Repository) func(i, j int) bool
		wantSlice []string
		wantErr   bool
	}{
		{
			name:     "Iterates through empty corpus",
			corpus:   &Corpus{},
			fn:       func(r Repository) error { workingSlice = append(workingSlice, r.RepoName()); return nil },
			filterfn: func(r Repository) bool { return true },
			sortfn: func(repos []*Repository) func(i, j int) bool {
				return func(i, j int) bool { return repos[i].RepoName() < repos[j].RepoName() }
			},
			wantSlice: []string{},
			wantErr:   false,
		},
		{
			name: "Iterates through a single repo",
			corpus: &Corpus{watchedOwners: []*Owner{&Owner{
				name: "someOwner",
				Repos: []*Repository{
					&Repository{name: "aRepo", ownerName: "someOwner"},
				},
			}}},
			fn:       func(r Repository) error { workingSlice = append(workingSlice, r.RepoName()); return nil },
			filterfn: func(r Repository) bool { return true },
			sortfn: func(repos []*Repository) func(i, j int) bool {
				return func(i, j int) bool { return repos[i].RepoName() < repos[j].RepoName() }
			},
			wantSlice: []string{"aRepo"},
			wantErr:   false,
		},
		{
			name: "Iterates through several repos in order",
			corpus: &Corpus{watchedOwners: []*Owner{&Owner{
				name: "someOwner",
				Repos: []*Repository{
					&Repository{name: "aRepo", ownerName: "someOwner"},
					&Repository{name: "repo2", ownerName: "someOwner"},
					&Repository{name: "repo3", ownerName: "someOwner"},
				},
			}}},
			fn:       func(r Repository) error { workingSlice = append(workingSlice, r.RepoName()); return nil },
			filterfn: func(r Repository) bool { return true },
			sortfn: func(repos []*Repository) func(i, j int) bool {
				return func(i, j int) bool { return repos[i].RepoName() < repos[j].RepoName() }
			},
			wantSlice: []string{"aRepo", "repo2", "repo3"},
			wantErr:   false,
		},
		{
			name: "Iterates through several repos in reverse order",
			corpus: &Corpus{watchedOwners: []*Owner{&Owner{
				name: "someOwner",
				Repos: []*Repository{
					&Repository{name: "aRepo", ownerName: "someOwner"},
					&Repository{name: "repo2", ownerName: "someOwner"},
					&Repository{name: "repo3", ownerName: "someOwner"},
				},
			}}},
			fn:       func(r Repository) error { workingSlice = append(workingSlice, r.RepoName()); return nil },
			filterfn: func(r Repository) bool { return true },
			sortfn: func(repos []*Repository) func(i, j int) bool {
				return func(i, j int) bool { return repos[i].RepoName() > repos[j].RepoName() }
			},
			wantSlice: []string{"repo3", "repo2", "aRepo"},
			wantErr:   false,
		},
		{
			name: "Iterates through several repos in order with filtering",
			corpus: &Corpus{watchedOwners: []*Owner{&Owner{
				name: "someOwner",
				Repos: []*Repository{
					&Repository{name: "aRepo", ownerName: "someOwner"},
					&Repository{name: "repo2", ownerName: "someOwner"},
					&Repository{name: "repo3", ownerName: "someOwner"},
				},
			}}},
			fn:       func(r Repository) error { workingSlice = append(workingSlice, r.RepoName()); return nil },
			filterfn: func(r Repository) bool { return !(r.RepoName() == "repo3") },
			sortfn: func(repos []*Repository) func(i, j int) bool {
				return func(i, j int) bool { return repos[i].RepoName() > repos[j].RepoName() }
			},
			wantSlice: []string{"repo2", "aRepo"},
			wantErr:   false,
		},
		{
			name: "Iterates in order through repos with different owners",
			corpus: &Corpus{watchedOwners: []*Owner{
				&Owner{
					name: "someOwner",
					Repos: []*Repository{
						&Repository{name: "aRepo", ownerName: "someOwner"},
						&Repository{name: "repo2", ownerName: "someOwner"},
						&Repository{name: "repo3", ownerName: "someOwner"},
					},
				},
				&Owner{
					name: "anotherOwner",
					Repos: []*Repository{
						&Repository{name: "aARepo", ownerName: "anotherOwner"},
						&Repository{name: "otherRepo", ownerName: "anotherOwner"},
					},
				},
			}},
			fn:       func(r Repository) error { workingSlice = append(workingSlice, r.RepoName()); return nil },
			filterfn: func(r Repository) bool { return true },
			sortfn: func(repos []*Repository) func(i, j int) bool {
				return func(i, j int) bool { return repos[i].RepoName() < repos[j].RepoName() }
			},
			wantSlice: []string{"aARepo", "aRepo", "otherRepo", "repo2", "repo3"},
			wantErr:   false,
		},
		{
			name: "Returns first error from func",
			corpus: &Corpus{watchedOwners: []*Owner{
				&Owner{
					name: "someOwner",
					Repos: []*Repository{
						&Repository{name: "aRepo", ownerName: "someOwner"},
						&Repository{name: "repo2", ownerName: "someOwner"},
					},
				},
			}},
			fn: func(r Repository) error {
				workingSlice = append(workingSlice, r.RepoName())
				return fmt.Errorf("an error")
			},
			filterfn: func(r Repository) bool { return true },
			sortfn: func(repos []*Repository) func(i, j int) bool {
				return func(i, j int) bool { return repos[i].RepoName() < repos[j].RepoName() }
			},
			wantSlice: []string{"aRepo"},
			wantErr:   true,
		},
	}
	for _, test := range tests {
		workingSlice = []string{}

		gotErr := test.corpus.ForEachRepoFSort(test.fn, test.filterfn, test.sortfn)

		if !reflect.DeepEqual(workingSlice, test.wantSlice) {
			t.Errorf("%v did not pass.\n\tWant:\t%v\n\tGot:\t%v", test.name, test.wantSlice, workingSlice)
		}

		if (gotErr != nil && !test.wantErr) || (gotErr == nil && test.wantErr) {
			t.Errorf("%v did not pass.\n\tWant Err: %v \n\tGot Err: %v", test.name, test.wantErr, gotErr)
		}
	}
}

func TestForEachOwner(t *testing.T) {
	var workingSet map[string]int

	tests := []struct {
		name    string
		corpus  *Corpus
		fn      func(o Owner) error
		wantSet map[string]int
		wantErr bool
	}{
		{
			name:    "Iterates through a single owner",
			corpus:  &Corpus{watchedOwners: []*Owner{&Owner{name: "anOwner"}}},
			fn:      func(o Owner) error { workingSet[o.Name()]++; return nil },
			wantSet: map[string]int{"anOwner": 1},
			wantErr: false,
		},
		{
			name:    "Iterates through empty corpus",
			corpus:  &Corpus{},
			fn:      func(o Owner) error { workingSet[o.Name()]++; return nil },
			wantSet: map[string]int{},
			wantErr: false,
		},
		{
			name: "Iterates through several owners",
			corpus: &Corpus{watchedOwners: []*Owner{
				&Owner{name: "anOwner"},
				&Owner{
					name: "someOwner",
					Repos: []*Repository{
						&Repository{name: "aRepo", ownerName: "someOwner"},
						&Repository{name: "repo2", ownerName: "someOwner"},
						&Repository{name: "repo3", ownerName: "someOwner"},
					},
				},
				&Owner{
					name: "anotherOwner",
					Repos: []*Repository{
						&Repository{name: "aRepo", ownerName: "anotherOwner"},
						&Repository{name: "otherRepo", ownerName: "anotherOwner"},
					},
				},
			}},
			fn:      func(o Owner) error { workingSet[o.Name()]++; return nil },
			wantSet: map[string]int{"someOwner": 1, "anOwner": 1, "anotherOwner": 1},
			wantErr: false,
		},
		{
			name: "Returns first error from func",
			corpus: &Corpus{watchedOwners: []*Owner{
				&Owner{name: "anOwner"},
				&Owner{name: "someOwner"},
			}},
			fn:      func(o Owner) error { workingSet["runCount"]++; return fmt.Errorf("an error") },
			wantSet: map[string]int{"runCount": 1},
			wantErr: true,
		},
	}
	for _, test := range tests {
		workingSet = map[string]int{}

		gotErr := test.corpus.ForEachOwner(test.fn)

		if !reflect.DeepEqual(workingSet, test.wantSet) {
			t.Errorf("%v did not pass.\n\tWant:\t%v\n\tGot:\t%v", test.name, test.wantSet, workingSet)
		}

		if (gotErr != nil && !test.wantErr) || (gotErr == nil && test.wantErr) {
			t.Errorf("%v did not pass.\n\tWant Err: %v \n\tGot Err: %v", test.name, test.wantErr, gotErr)
		}
	}
}

func TestForEachOwnerF(t *testing.T) {
	var workingSet map[string]int

	tests := []struct {
		name     string
		corpus   *Corpus
		fn       func(o Owner) error
		filterfn func(o Owner) bool
		wantSet  map[string]int
		wantErr  bool
	}{
		{
			name:     "Iterates through empty corpus",
			corpus:   &Corpus{},
			fn:       func(o Owner) error { workingSet[o.Name()]++; return nil },
			filterfn: func(o Owner) bool { return true },
			wantSet:  map[string]int{},
			wantErr:  false,
		},
		{
			name:     "Iterates through a single owner",
			corpus:   &Corpus{watchedOwners: []*Owner{&Owner{name: "anOwner"}}},
			fn:       func(o Owner) error { workingSet[o.Name()]++; return nil },
			filterfn: func(o Owner) bool { return true },
			wantSet:  map[string]int{"anOwner": 1},
			wantErr:  false,
		},
		{
			name:     "Iterates and filters out through a single owner",
			corpus:   &Corpus{watchedOwners: []*Owner{&Owner{name: "anOwner"}}},
			fn:       func(o Owner) error { workingSet[o.Name()]++; return nil },
			filterfn: func(o Owner) bool { return false },
			wantSet:  map[string]int{},
			wantErr:  false,
		},
		{
			name: "Iterates through several owners and filters several out",
			corpus: &Corpus{watchedOwners: []*Owner{
				&Owner{name: "anOwner"},
				&Owner{name: "someOwner"},
				&Owner{name: "anotherOwner"},
			}},
			fn:       func(o Owner) error { workingSet[o.Name()]++; return nil },
			filterfn: func(o Owner) bool { return o.Name() == "someOwner" },
			wantSet:  map[string]int{"someOwner": 1},
			wantErr:  false,
		},
		{
			name: "Iterates through several owners and filters all out",
			corpus: &Corpus{watchedOwners: []*Owner{
				&Owner{name: "anOwner"},
				&Owner{name: "someOwner"},
				&Owner{name: "anotherOwner"},
			}},
			fn:       func(o Owner) error { workingSet[o.Name()]++; return nil },
			filterfn: func(o Owner) bool { return false },
			wantSet:  map[string]int{},
			wantErr:  false,
		},
		{
			name: "Returns first error from func",
			corpus: &Corpus{watchedOwners: []*Owner{
				&Owner{name: "anOwner"},
				&Owner{name: "someOwner"},
			}},
			fn:       func(o Owner) error { workingSet["runCount"]++; return fmt.Errorf("an error") },
			filterfn: func(o Owner) bool { return true },
			wantSet:  map[string]int{"runCount": 1},
			wantErr:  true,
		},
	}
	for _, test := range tests {
		workingSet = map[string]int{}

		gotErr := test.corpus.ForEachOwnerF(test.fn, test.filterfn)

		if !reflect.DeepEqual(workingSet, test.wantSet) {
			t.Errorf("%v did not pass.\n\tWant:\t%v\n\tGot:\t%v", test.name, test.wantSet, workingSet)
		}

		if (gotErr != nil && !test.wantErr) || (gotErr == nil && test.wantErr) {
			t.Errorf("%v did not pass.\n\tWant Err: %v \n\tGot Err: %v", test.name, test.wantErr, gotErr)
		}
	}
}

func TestForEachOwnerFSort(t *testing.T) {
	var workingSlice []string

	tests := []struct {
		name      string
		corpus    *Corpus
		fn        func(o Owner) error
		filterfn  func(o Owner) bool
		sortfn    func([]*Owner) func(i, j int) bool
		wantSlice []string
		wantErr   bool
	}{
		{
			name:     "Iterates through a single owner",
			corpus:   &Corpus{watchedOwners: []*Owner{&Owner{name: "anOwner"}}},
			fn:       func(o Owner) error { workingSlice = append(workingSlice, o.Name()); return nil },
			filterfn: func(o Owner) bool { return true },
			sortfn: func(owners []*Owner) func(i, j int) bool {
				return func(i, j int) bool { return owners[i].Name() < owners[j].Name() }
			},
			wantSlice: []string{"anOwner"},
			wantErr:   false,
		},
		{
			name:     "Iterates and filters out through a single owner",
			corpus:   &Corpus{watchedOwners: []*Owner{&Owner{name: "anOwner"}}},
			fn:       func(o Owner) error { workingSlice = append(workingSlice, o.Name()); return nil },
			filterfn: func(o Owner) bool { return false },
			sortfn: func(owners []*Owner) func(i, j int) bool {
				return func(i, j int) bool { return owners[i].Name() < owners[j].Name() }
			},
			wantSlice: []string{},
			wantErr:   false,
		},
		{
			name: "Iterates through several owners in order",
			corpus: &Corpus{watchedOwners: []*Owner{
				&Owner{name: "o2"},
				&Owner{name: "o3"},
				&Owner{name: "o1"},
			}},
			fn:       func(o Owner) error { workingSlice = append(workingSlice, o.Name()); return nil },
			filterfn: func(o Owner) bool { return true },
			sortfn: func(owners []*Owner) func(i, j int) bool {
				return func(i, j int) bool { return owners[i].Name() < owners[j].Name() }
			},
			wantSlice: []string{"o1", "o2", "o3"},
			wantErr:   false,
		},
		{
			name: "Iterates through several owners in reverse order",
			corpus: &Corpus{watchedOwners: []*Owner{
				&Owner{name: "o1"},
				&Owner{name: "o2"},
				&Owner{name: "o3"},
			}},
			fn:       func(o Owner) error { workingSlice = append(workingSlice, o.Name()); return nil },
			filterfn: func(o Owner) bool { return true },
			sortfn: func(owners []*Owner) func(i, j int) bool {
				return func(i, j int) bool { return owners[i].Name() > owners[j].Name() }
			},
			wantSlice: []string{"o3", "o2", "o1"},
			wantErr:   false,
		},
		{
			name: "Iterates through several owners in order with filtering",
			corpus: &Corpus{watchedOwners: []*Owner{
				&Owner{name: "o3"},
				&Owner{name: "o2"},
				&Owner{name: "o1"},
			}},
			fn:       func(o Owner) error { workingSlice = append(workingSlice, o.Name()); return nil },
			filterfn: func(o Owner) bool { return o.Name() != "o3" },
			sortfn: func(owners []*Owner) func(i, j int) bool {
				return func(i, j int) bool { return owners[i].Name() < owners[j].Name() }
			},
			wantSlice: []string{"o1", "o2"},
			wantErr:   false,
		},
		{
			name: "Returns first error from func",
			corpus: &Corpus{watchedOwners: []*Owner{
				&Owner{name: "someOwner"},
				&Owner{name: "anOwner"},
			}},
			fn:       func(o Owner) error { workingSlice = append(workingSlice, o.Name()); return fmt.Errorf("an error") },
			filterfn: func(o Owner) bool { return true },
			sortfn: func(owners []*Owner) func(i, j int) bool {
				return func(i, j int) bool { return owners[i].Name() < owners[j].Name() }
			},
			wantSlice: []string{"anOwner"},
			wantErr:   true,
		},
	}
	for _, test := range tests {
		workingSlice = []string{}

		gotErr := test.corpus.ForEachOwnerFSort(test.fn, test.filterfn, test.sortfn)

		if !reflect.DeepEqual(workingSlice, test.wantSlice) {
			t.Errorf("%v did not pass.\n\tWant:\t%v\n\tGot:\t%v", test.name, test.wantSlice, workingSlice)
		}

		if (gotErr != nil && !test.wantErr) || (gotErr == nil && test.wantErr) {
			t.Errorf("%v did not pass.\n\tWant Err: %v \n\tGot Err: %v", test.name, test.wantErr, gotErr)
		}
	}
}
