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

package filters

import (
	"testing"

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"

	"github.com/google/go-cmp/cmp"
)

func TestFilterIssue(t *testing.T) {
	tests := []struct {
		Name    string
		Issue   *drghs_v1.Issue
		Request *drghs_v1.ListIssuesRequest
		Want    bool
		WantErr bool
	}{
		{
			Name:    "Empty Filter Passes",
			Issue:   &drghs_v1.Issue{},
			Request: &drghs_v1.ListIssuesRequest{},
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Null Pull Request Passes",
			Issue: &drghs_v1.Issue{
				Name: "foo",
				IsPr: true,
			},
			Request: &drghs_v1.ListIssuesRequest{},
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Null Not Pull Request Passes",
			Issue: &drghs_v1.Issue{
				Name: "foo",
				IsPr: false,
			},
			Request: &drghs_v1.ListIssuesRequest{},
			Want:    true,
			WantErr: false,
		},
		{
			Name: "True Pull Request Passes",
			Issue: &drghs_v1.Issue{
				IsPr: true,
			},
			Request: &drghs_v1.ListIssuesRequest{
				PullRequestNullable: &drghs_v1.ListIssuesRequest_PullRequest{
					PullRequest: true,
				},
			},
			Want:    true,
			WantErr: false,
		},
		{
			Name: "False Pull Request Passes",
			Issue: &drghs_v1.Issue{
				IsPr: true,
			},
			Request: &drghs_v1.ListIssuesRequest{
				PullRequestNullable: &drghs_v1.ListIssuesRequest_PullRequest{
					PullRequest: false,
				},
			},
			Want:    false,
			WantErr: false,
		},
		{
			Name: "True Not Pull Request Passes",
			Issue: &drghs_v1.Issue{
				IsPr: false,
			},
			Request: &drghs_v1.ListIssuesRequest{
				PullRequestNullable: &drghs_v1.ListIssuesRequest_PullRequest{
					PullRequest: true,
				},
			},
			Want:    false,
			WantErr: false,
		},
		{
			Name: "False Not Pull Request Passes",
			Issue: &drghs_v1.Issue{
				IsPr: false,
			},
			Request: &drghs_v1.ListIssuesRequest{
				PullRequestNullable: &drghs_v1.ListIssuesRequest_PullRequest{
					PullRequest: false,
				},
			},
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Null Closed Passes",
			Issue: &drghs_v1.Issue{
				Name:   "foo",
				Closed: true,
			},
			Request: &drghs_v1.ListIssuesRequest{},
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Null Not Closed Passes",
			Issue: &drghs_v1.Issue{
				Name:   "foo",
				Closed: false,
			},
			Request: &drghs_v1.ListIssuesRequest{},
			Want:    true,
			WantErr: false,
		},
		{
			Name: "True Closed Passes",
			Issue: &drghs_v1.Issue{
				Closed: true,
			},
			Request: &drghs_v1.ListIssuesRequest{
				ClosedNullable: &drghs_v1.ListIssuesRequest_Closed{
					Closed: true,
				},
			},
			Want:    true,
			WantErr: false,
		},
		{
			Name: "False Closed Passes",
			Issue: &drghs_v1.Issue{
				Closed: true,
			},
			Request: &drghs_v1.ListIssuesRequest{
				ClosedNullable: &drghs_v1.ListIssuesRequest_Closed{
					Closed: false,
				},
			},
			Want:    false,
			WantErr: false,
		},
		{
			Name: "True Not Closed Passes",
			Issue: &drghs_v1.Issue{
				Closed: false,
			},
			Request: &drghs_v1.ListIssuesRequest{
				ClosedNullable: &drghs_v1.ListIssuesRequest_Closed{
					Closed: true,
				},
			},
			Want:    false,
			WantErr: false,
		},
		{
			Name: "False Not Closed Passes",
			Issue: &drghs_v1.Issue{
				Closed: false,
			},
			Request: &drghs_v1.ListIssuesRequest{
				ClosedNullable: &drghs_v1.ListIssuesRequest_Closed{
					Closed: false,
				},
			},
			Want:    true,
			WantErr: false,
		},
	}

	for _, test := range tests {
		got, goterr := FilterIssue(test.Issue, test.Request)
		if (test.WantErr && goterr == nil) || (!test.WantErr && goterr != nil) {
			t.Errorf("test: %v, errors diff. WantErr: %v, GotErr: %v.", test.Name, test.WantErr, goterr)
		}
		if diff := cmp.Diff(test.Want, got); diff != "" {
			t.Errorf("test: %v, values diff. match (-want +got)\n%s", test.Name, diff)
		}
	}
}

func TestFilterRepo(t *testing.T) {
	tests := []struct {
		Name    string
		Repo    *drghs_v1.Repository
		Filter  string
		Want    bool
		WantErr bool
	}{
		{
			Name:    "Empty Filter Passes",
			Repo:    &drghs_v1.Repository{},
			Filter:  "",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Filter Name Passes",
			Repo: &drghs_v1.Repository{
				Name: "foo",
			},
			Filter:  "repository.name == 'foo'",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Wrong Name Fails",
			Repo: &drghs_v1.Repository{
				Name: "foo",
			},
			Filter:  "repository.name == 'bar'",
			Want:    false,
			WantErr: false,
		},
	}

	for _, test := range tests {
		got, goterr := FilterRepository(test.Repo, test.Filter)
		if (test.WantErr && goterr == nil) || (!test.WantErr && goterr != nil) {
			t.Errorf("test: %v, errors diff. WantErr: %v, GotErr: %v.", test.Name, test.WantErr, goterr)
		}
		if diff := cmp.Diff(test.Want, got); diff != "" {
			t.Errorf("test: %v, values diff. match (-want +got)\n%s", test.Name, diff)
		}
	}
}

func TestFilterComment(t *testing.T) {
	tests := []struct {
		Name    string
		Comment *drghs_v1.GitHubComment
		Filter  string
		Want    bool
		WantErr bool
	}{
		{
			Name:    "Empty Filter Passes",
			Comment: &drghs_v1.GitHubComment{},
			Filter:  "",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Filter Body Passes",
			Comment: &drghs_v1.GitHubComment{
				Body: "foo",
			},
			Filter:  "comment.body == 'foo'",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Wrong Body Fails",
			Comment: &drghs_v1.GitHubComment{
				Body: "foo",
			},
			Filter:  "comment.body == 'bar'",
			Want:    false,
			WantErr: false,
		},
		{
			Name: "Deep check passes",
			Comment: &drghs_v1.GitHubComment{
				User: &drghs_v1.GitHubUser{
					Login: "foo",
				},
			},
			Filter:  "comment.user.login == 'foo'",
			Want:    true,
			WantErr: false,
		},
	}

	for _, test := range tests {
		got, goterr := FilterComment(test.Comment, test.Filter)
		if (test.WantErr && goterr == nil) || (!test.WantErr && goterr != nil) {
			t.Errorf("test: %v, errors diff. WantErr: %v, GotErr: %v.", test.Name, test.WantErr, goterr)
		}
		if diff := cmp.Diff(test.Want, got); diff != "" {
			t.Errorf("test: %v, values diff. match (-want +got)\n%s", test.Name, diff)
		}
	}
}
