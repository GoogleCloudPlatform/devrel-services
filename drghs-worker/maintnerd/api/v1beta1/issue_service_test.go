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

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/build/maintner"
	"google.golang.org/genproto/protobuf/field_mask"
)

func TestIssueFilters(t *testing.T) {
	tests := []struct {
		Name    string
		Issue   maintner.GitHubIssue
		Req     *drghs_v1.ListIssuesRequest
		Want    bool
		WantErr bool
	}{
		{
			Name: "Empty Filter Passes",
			Issue: maintner.GitHubIssue{
				PullRequest: true,
				Closed:      false,
			},
			Req:     &drghs_v1.ListIssuesRequest{},
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Pull Request Filter Passes",
			Issue: maintner.GitHubIssue{
				PullRequest: true,
				Closed:      false,
			},
			Req: &drghs_v1.ListIssuesRequest{
				PullRequestNullable: &drghs_v1.ListIssuesRequest_PullRequest{
					PullRequest: true,
				},
			},
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Pull Request False Filter Passes",
			Issue: maintner.GitHubIssue{
				PullRequest: true,
				Closed:      false,
			},
			Req: &drghs_v1.ListIssuesRequest{
				PullRequestNullable: &drghs_v1.ListIssuesRequest_PullRequest{

					PullRequest: false,
				},
			},
			Want:    false,
			WantErr: false,
		},
		{
			Name: "Closed Filter Passes",
			Issue: maintner.GitHubIssue{
				PullRequest: true,
				Closed:      true,
			},
			Req: &drghs_v1.ListIssuesRequest{
				ClosedNullable: &drghs_v1.ListIssuesRequest_Closed{
					Closed: true,
				},
			},
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Closed False Filter Skips Closed",
			Issue: maintner.GitHubIssue{
				PullRequest: true,
				Closed:      true,
			},
			Req: &drghs_v1.ListIssuesRequest{
				ClosedNullable: &drghs_v1.ListIssuesRequest_Closed{
					Closed: false,
				},
			},
			Want:    false,
			WantErr: false,
		},
		{
			Name: "Closed False Filter Passes",
			Issue: maintner.GitHubIssue{
				PullRequest: true,
				Closed:      false,
			},
			Req: &drghs_v1.ListIssuesRequest{
				ClosedNullable: &drghs_v1.ListIssuesRequest_Closed{
					Closed: false,
				},
			},
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Compound Filter Supported",
			Issue: maintner.GitHubIssue{
				PullRequest: true,
				Closed:      false,
			},
			Req: &drghs_v1.ListIssuesRequest{
				PullRequestNullable: &drghs_v1.ListIssuesRequest_PullRequest{

					PullRequest: true,
				},
				ClosedNullable: &drghs_v1.ListIssuesRequest_Closed{

					Closed: true,
				},
			},
			Want:    false,
			WantErr: false,
		},
		{
			Name: "Compound Filter Passes on PR",
			Issue: maintner.GitHubIssue{
				PullRequest: true,
				Closed:      false,
			},
			Req: &drghs_v1.ListIssuesRequest{
				PullRequestNullable: &drghs_v1.ListIssuesRequest_PullRequest{
					PullRequest: true,
				},
				ClosedNullable: &drghs_v1.ListIssuesRequest_Closed{

					Closed: false,
				},
			},
			Want:    true,
			WantErr: false,
		},
	}

	for _, test := range tests {
		got, goterr := shouldAddIssue(&test.Issue, test.Req)
		if (test.WantErr && goterr == nil) || (!test.WantErr && goterr != nil) {
			t.Errorf("test: %v, errors diff. WantErr: %v, GotErr: %v.", test.Name, test.WantErr, goterr)
		}
		if diff := cmp.Diff(test.Want, got); diff != "" {
			t.Errorf("test: %v, values diff. match (-want +got)\n%s", test.Name, diff)
		}
	}
}

func TestHandlesIssue(t *testing.T) {
	cases := []struct {
		Name    string
		Issue   *maintner.GitHubIssue
		RepoID  maintner.GitHubRepoID
		Request *drghs_v1.ListIssuesRequest
		Issues  []*drghs_v1.Issue
		Want    []*drghs_v1.Issue
		WantErr bool
	}{
		{
			Name:  "Skips Not Exist",
			Issue: &maintner.GitHubIssue{NotExist: true},
			RepoID: maintner.GitHubRepoID{
				Owner: "foo",
				Repo:  "bar",
			},
			Request: &drghs_v1.ListIssuesRequest{},
			Issues:  []*drghs_v1.Issue{},
			Want:    []*drghs_v1.Issue{},
			WantErr: false,
		},
		{
			Name: "Adds with field mask",
			Issue: &maintner.GitHubIssue{
				Title:    "Foobar",
				NotExist: false,
				Closed:   true,
			},
			RepoID: maintner.GitHubRepoID{
				Owner: "foo",
				Repo:  "bar",
			},
			Issues: []*drghs_v1.Issue{},
			Request: &drghs_v1.ListIssuesRequest{
				Comments: false,
				Reviews:  false,
				ClosedNullable: &drghs_v1.ListIssuesRequest_Closed{
					Closed: true,
				},
				FieldMask: &field_mask.FieldMask{Paths: []string{"title"}},
			},
			Want: []*drghs_v1.Issue{
				&drghs_v1.Issue{
					Title: "Foobar",
				},
			},
			WantErr: false,
		},
		{
			Name: "Adds with field mask on false",
			Issue: &maintner.GitHubIssue{
				Title:    "Foobar",
				NotExist: false,
				Closed:   false,
			},
			RepoID: maintner.GitHubRepoID{
				Owner: "foo",
				Repo:  "bar",
			},
			Issues: []*drghs_v1.Issue{},
			Request: &drghs_v1.ListIssuesRequest{
				Comments: false,
				Reviews:  false,
				ClosedNullable: &drghs_v1.ListIssuesRequest_Closed{
					Closed: false,
				},
				FieldMask: &field_mask.FieldMask{Paths: []string{"title"}},
			},
			Want: []*drghs_v1.Issue{
				&drghs_v1.Issue{
					Title: "Foobar",
				},
			},
			WantErr: false,
		},
	}

	for _, c := range cases {
		got, goterr := handleIssue(c.Issue, c.RepoID, c.Request, c.Issues)
		if (c.WantErr && goterr == nil) || (!c.WantErr && goterr != nil) {
			t.Errorf("test: %v, errors diff. WantErr: %v, GotErr: %v.", c.Name, c.WantErr, goterr)
		}
		if diff := cmp.Diff(c.Want, got); diff != "" {
			t.Errorf("test: %v, values diff. match (-want +got)\n%s", c.Name, diff)
		}
	}
}
