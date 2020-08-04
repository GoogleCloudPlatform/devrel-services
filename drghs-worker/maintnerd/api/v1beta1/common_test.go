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

package v1beta1

import (
	"testing"
	"time"

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"
	"google.golang.org/genproto/protobuf/field_mask"
	proto "google.golang.org/protobuf/runtime/protoimpl"

	tspb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/build/maintner"
)

func TestMakeIssuePBFieldMask(t *testing.T) {
	rID := maintner.GitHubRepoID{
		Owner: "foo",
		Repo:  "bar",
	}
	now := time.Now()
	ghIss := &maintner.GitHubIssue{
		Created:  now,
		Updated:  now,
		ClosedAt: now,
		ClosedBy: &maintner.GitHubUser{
			Login: "testuser",
		},
		User: &maintner.GitHubUser{
			Login: "testuser2",
		},
		Assignees: []*maintner.GitHubUser{
			{
				Login: "testuser3",
			},
		},
		Closed:      true,
		PullRequest: true,
		Title:       "title",
		Body:        "body",
		Number:      1234,
		Labels: map[int64]*maintner.GitHubLabel{
			1: &maintner.GitHubLabel{Name: "bug"},
			2: &maintner.GitHubLabel{Name: "foo"},
			3: &maintner.GitHubLabel{Name: "Zebra"},
			4: &maintner.GitHubLabel{Name: "bar"},
			5: &maintner.GitHubLabel{Name: "Feat"},
			6: &maintner.GitHubLabel{Name: "p0"},
			7: &maintner.GitHubLabel{Name: "blocked"},
			8: &maintner.GitHubLabel{Name: "blocking"},
		},
	}

	tests := []struct {
		ghIss *maintner.GitHubIssue
		fm    *field_mask.FieldMask
		want  *drghs_v1.Issue
	}{
		{
			fm: nil,
			want: &drghs_v1.Issue{
				CreatedAt: &tspb.Timestamp{
					Seconds: now.Unix(),
					Nanos:   int32(now.Nanosecond()),
				},
				UpdatedAt: &tspb.Timestamp{
					Seconds: now.Unix(),
					Nanos:   int32(now.Nanosecond()),
				},
				ClosedAt: &tspb.Timestamp{
					Seconds: now.Unix(),
					Nanos:   int32(now.Nanosecond()),
				},
				ClosedBy: &drghs_v1.GitHubUser{
					Login: "testuser",
				},
				Reporter: &drghs_v1.GitHubUser{
					Login: "testuser2",
				},
				Assignees: []*drghs_v1.GitHubUser{
					{
						Login: "testuser3",
					},
				},
				Closed:          true,
				IsPr:            true,
				Title:           "title",
				Body:            "body",
				IssueId:         1234,
				Url:             "https://github.com/foo/bar/issues/1234",
				Repo:            "foo/bar",
				Labels:          []string{"bar", "blocked", "blocking", "bug", "Feat", "foo", "p0", "Zebra"},
				Priority:        drghs_v1.Issue_P0,
				PriorityUnknown: false,
				IssueType:       drghs_v1.Issue_BUG,
				Blocked:         true,
				ReleaseBlocking: true,
			},
		},
		{
			fm: &field_mask.FieldMask{Paths: []string{"created_at"}},
			want: &drghs_v1.Issue{
				CreatedAt: &tspb.Timestamp{
					Seconds: now.Unix(),
					Nanos:   int32(now.Nanosecond()),
				},
			},
		},
		{
			fm: &field_mask.FieldMask{Paths: []string{"updated_at"}},
			want: &drghs_v1.Issue{
				UpdatedAt: &tspb.Timestamp{
					Seconds: now.Unix(),
					Nanos:   int32(now.Nanosecond()),
				},
			},
		},
		{
			fm: &field_mask.FieldMask{Paths: []string{"closed_at"}},
			want: &drghs_v1.Issue{
				ClosedAt: &tspb.Timestamp{
					Seconds: now.Unix(),
					Nanos:   int32(now.Nanosecond()),
				},
			},
		},
		{
			fm: &field_mask.FieldMask{Paths: []string{"closed_by"}},
			want: &drghs_v1.Issue{
				ClosedBy: &drghs_v1.GitHubUser{Login: "testuser"},
			},
		},
		{
			fm: &field_mask.FieldMask{Paths: []string{"reporter"}},
			want: &drghs_v1.Issue{
				Reporter: &drghs_v1.GitHubUser{Login: "testuser2"},
			},
		},
		{
			fm: &field_mask.FieldMask{Paths: []string{"assignees"}},
			want: &drghs_v1.Issue{
				Assignees: []*drghs_v1.GitHubUser{
					{Login: "testuser3"},
				},
			},
		},
		{
			fm:   &field_mask.FieldMask{Paths: []string{"closed"}},
			want: &drghs_v1.Issue{Closed: true},
		},
		{
			fm:   &field_mask.FieldMask{Paths: []string{"is_pr"}},
			want: &drghs_v1.Issue{IsPr: true},
		},
		{
			fm:   &field_mask.FieldMask{Paths: []string{"title"}},
			want: &drghs_v1.Issue{Title: "title"},
		},
		{
			fm:   &field_mask.FieldMask{Paths: []string{"body"}},
			want: &drghs_v1.Issue{Body: "body"},
		},
		{
			fm:   &field_mask.FieldMask{Paths: []string{"issue_id"}},
			want: &drghs_v1.Issue{IssueId: 1234},
		},
		{
			fm:   &field_mask.FieldMask{Paths: []string{"url"}},
			want: &drghs_v1.Issue{Url: "https://github.com/foo/bar/issues/1234"},
		},
		{
			fm:   &field_mask.FieldMask{Paths: []string{"repo"}},
			want: &drghs_v1.Issue{Repo: "foo/bar"},
		},
		{
			fm:   &field_mask.FieldMask{Paths: []string{"labels"}},
			want: &drghs_v1.Issue{Labels: []string{"bar", "blocked", "blocking", "bug", "Feat", "foo", "p0", "Zebra"}},
		},
		{
			fm:   &field_mask.FieldMask{Paths: []string{"priority"}},
			want: &drghs_v1.Issue{Priority: drghs_v1.Issue_P0},
		},
		{
			fm:   &field_mask.FieldMask{Paths: []string{"priority_unknown"}},
			want: &drghs_v1.Issue{PriorityUnknown: false},
		},
		{
			fm:   &field_mask.FieldMask{Paths: []string{"issue_type"}},
			want: &drghs_v1.Issue{IssueType: drghs_v1.Issue_BUG},
		},
		{
			fm:   &field_mask.FieldMask{Paths: []string{"blocked"}},
			want: &drghs_v1.Issue{Blocked: true},
		},
		{
			fm:   &field_mask.FieldMask{Paths: []string{"release_blocking"}},
			want: &drghs_v1.Issue{ReleaseBlocking: true},
		},
		{
			fm: &field_mask.FieldMask{Paths: []string{"issue_id", "issue_type", "updated_at"}},
			want: &drghs_v1.Issue{
				UpdatedAt: &tspb.Timestamp{
					Seconds: now.Unix(),
					Nanos:   int32(now.Nanosecond()),
				},
				IssueId:   1234,
				IssueType: drghs_v1.Issue_BUG,
			},
		},
	}

	for _, test := range tests {
		got, err := makeIssuePB(ghIss, rID, false, false, test.fm)
		if err != nil {
			t.Errorf("Unexpected error from makeIssuePB. Wanted nil, Got %v", err)
		}
		if diff := cmp.Diff(test.want, got, cmpopts.IgnoreUnexported(tspb.Timestamp{}, proto.MessageState{})); diff != "" {
			t.Errorf("makeIssuePB() mismatch (-want +got):\n%s", diff)
		}
	}
}
