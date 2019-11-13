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
		Issue   drghs_v1.Issue
		Filter  string
		Want    bool
		WantErr bool
	}{
		{
			Name:    "Empty Filter Passes",
			Issue:   drghs_v1.Issue{},
			Filter:  "",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Filter Name Passes",
			Issue: drghs_v1.Issue{
				Name: "foo",
			},
			Filter:  "issue.name == 'foo'",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Wrong Name Fails",
			Issue: drghs_v1.Issue{
				Name: "foo",
			},
			Filter:  "issue.name == 'bar'",
			Want:    false,
			WantErr: false,
		},
		{
			Name: "Bool Field passes",
			Issue: drghs_v1.Issue{
				Name: "foo",
				IsPr: true,
			},
			Filter:  "issue.is_pr == true",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Deep filter passes",
			Issue: drghs_v1.Issue{
				Labels: []string{"foo", "bar"},
			},
			Filter:  "issue.labels.size() > 1",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Contains filter passes",
			Issue: drghs_v1.Issue{
				Labels: []string{"foo", "bar"},
			},
			Filter:  "'bar' in issue.labels",
			Want:    true,
			WantErr: false,
		},
	}

	for _, test := range tests {
		got, goterr := FilterIssue(&test.Issue, test.Filter)
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
		Repo    drghs_v1.Repository
		Filter  string
		Want    bool
		WantErr bool
	}{
		{
			Name:    "Empty Filter Passes",
			Repo:    drghs_v1.Repository{},
			Filter:  "",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Filter Name Passes",
			Repo: drghs_v1.Repository{
				Name: "foo",
			},
			Filter:  "repository.name == 'foo'",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Wrong Name Fails",
			Repo: drghs_v1.Repository{
				Name: "foo",
			},
			Filter:  "repository.name == 'bar'",
			Want:    false,
			WantErr: false,
		},
	}

	for _, test := range tests {
		got, goterr := FilterRepository(&test.Repo, test.Filter)
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
		Comment drghs_v1.GitHubComment
		Filter  string
		Want    bool
		WantErr bool
	}{
		{
			Name:    "Empty Filter Passes",
			Comment: drghs_v1.GitHubComment{},
			Filter:  "",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Filter Body Passes",
			Comment: drghs_v1.GitHubComment{
				Body: "foo",
			},
			Filter:  "comment.body == 'foo'",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Wrong Body Fails",
			Comment: drghs_v1.GitHubComment{
				Body: "foo",
			},
			Filter:  "comment.body == 'bar'",
			Want:    false,
			WantErr: false,
		},
		{
			Name: "Deep check passes",
			Comment: drghs_v1.GitHubComment{
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
		got, goterr := FilterComment(&test.Comment, test.Filter)
		if (test.WantErr && goterr == nil) || (!test.WantErr && goterr != nil) {
			t.Errorf("test: %v, errors diff. WantErr: %v, GotErr: %v.", test.Name, test.WantErr, goterr)
		}
		if diff := cmp.Diff(test.Want, got); diff != "" {
			t.Errorf("test: %v, values diff. match (-want +got)\n%s", test.Name, diff)
		}
	}
}
