// Copyright 2019 Google LLC
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

package v1beta1

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/build/maintner"
)

func TestIssueFilters(t *testing.T) {
	tests := []struct {
		Name    string
		Issue   maintner.GitHubIssue
		Filter  string
		Want    bool
		WantErr bool
	}{
		{
			Name: "Empty Filter Passes",
			Issue: maintner.GitHubIssue{
				PullRequest: true,
				Closed:      false,
			},
			Filter:  "",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "True Filter Passes",
			Issue: maintner.GitHubIssue{
				PullRequest: true,
				Closed:      false,
			},
			Filter:  "true",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Pull Request Filter Passes",
			Issue: maintner.GitHubIssue{
				PullRequest: true,
				Closed:      false,
			},
			Filter:  "pull_request==true",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Compound  Filter Supported",
			Issue: maintner.GitHubIssue{
				PullRequest: true,
				Closed:      false,
			},
			Filter:  "pull_request==true && closed==true",
			Want:    false,
			WantErr: false,
		},
		{
			Name: "Invalid Filter",
			Issue: maintner.GitHubIssue{
				PullRequest: true,
				Closed:      false,
			},
			Filter:  `bar==foo && closed==true`,
			Want:    false,
			WantErr: true,
		},
	}

	for _, test := range tests {
		got, goterr := shouldAddIssue(&test.Issue, test.Filter)
		if (test.WantErr && goterr == nil) || (!test.WantErr && goterr != nil) {
			t.Errorf("test: %v, errors diff. WantErr: %v, GotErr: %v.", test.Name, test.WantErr, goterr)
		}
		if diff := cmp.Diff(test.Want, got); diff != "" {
			t.Errorf("test: %v, values diff. match (-want +got)\n%s", test.Name, diff)
		}
	}
}

func TestRepoFilters(t *testing.T) {
	tests := []struct {
		Name    string
		RepoID  maintner.GitHubRepoID
		Filter  string
		Want    bool
		WantErr bool
	}{
		{
			Name: "Empty Filter Passes",
			RepoID: maintner.GitHubRepoID{
				Owner: "foo",
				Repo:  "bar",
			},
			Filter:  "",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "True Filter Passes",
			RepoID: maintner.GitHubRepoID{
				Owner: "foo",
				Repo:  "bar",
			},
			Filter:  "true",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Owner Filter Passes",
			RepoID: maintner.GitHubRepoID{
				Owner: "foo",
				Repo:  "bar",
			},
			Filter:  `owner=="foo"`,
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Repo Filter Passes",
			RepoID: maintner.GitHubRepoID{
				Owner: "foo",
				Repo:  "bar",
			},
			Filter:  `repo=="bar"`,
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Compound Filter Passes",
			RepoID: maintner.GitHubRepoID{
				Owner: "foo",
				Repo:  "bar",
			},
			Filter:  `repo=="bar" && owner=="foo"`,
			Want:    true,
			WantErr: false,
		},
		{
			Name: "In Filter",
			RepoID: maintner.GitHubRepoID{
				Owner: "foo",
				Repo:  "bar",
			},
			Filter:  `repo in ["bar", "baz"]`,
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Invalid Filter",
			RepoID: maintner.GitHubRepoID{
				Owner: "foo",
				Repo:  "bar",
			},
			Filter:  `bar==foo && closed==true`,
			Want:    false,
			WantErr: true,
		},
	}
	for _, test := range tests {
		got, goterr := shouldAddRepository(test.RepoID, test.Filter)
		if (test.WantErr && goterr == nil) || (!test.WantErr && goterr != nil) {
			t.Errorf("test: %v, errors diff. WantErr: %v, GotErr: %v.", test.Name, test.WantErr, goterr)
		}
		if diff := cmp.Diff(test.Want, got); diff != "" {
			t.Errorf("test: %v, values diff. match (-want +got)\n%s", test.Name, diff)
		}
	}
}
