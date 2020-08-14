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

package sloutils

import (
	"testing"
	"time"

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"
	"github.com/golang/protobuf/ptypes"
	"golang.org/x/build/maintner"
)

func TestDoesSloApply(t *testing.T) {
	tests := []struct {
		name  string
		slo   *drghs_v1.SLO
		issue *maintner.GitHubIssue
		want  bool
	}{
		{
			name:  "Test nils",
			slo:   nil,
			issue: nil,
			want:  false,
		},
		{
			name: "Applies to an empty issue",
			slo: &drghs_v1.SLO{
				AppliesToIssues: true,
				AppliesToPrs:    true,
			},
			issue: &maintner.GitHubIssue{},
			want:  true,
		},
		{
			name: "Should not apply if issue does not exist",
			slo: &drghs_v1.SLO{
				AppliesToIssues: true,
				AppliesToPrs:    true,
			},
			issue: &maintner.GitHubIssue{
				NotExist: true,
			},
			want: false,
		},
		{
			name: "Applies to a PR",
			slo: &drghs_v1.SLO{
				AppliesToIssues: true,
				AppliesToPrs:    true,
			},
			issue: &maintner.GitHubIssue{
				PullRequest: true,
			},
			want: true,
		},
		{
			name: "Does not apply to an issue",
			slo: &drghs_v1.SLO{
				AppliesToIssues: false,
				AppliesToPrs:    true,
			},
			issue: &maintner.GitHubIssue{
				PullRequest: false,
			},
			want: false,
		},
		{
			name: "Does not apply to PRs",
			slo: &drghs_v1.SLO{
				AppliesToIssues: true,
				AppliesToPrs:    false,
			},
			issue: &maintner.GitHubIssue{
				PullRequest: true,
			},
			want: false,
		},
		{
			name: "Does not apply if issue doesn't have label",
			slo: &drghs_v1.SLO{
				AppliesToIssues: true,
				GithubLabels:    []string{"labelle"},
			},
			issue: &maintner.GitHubIssue{},
			want:  false,
		},
		{
			name: "Applies with label",
			slo: &drghs_v1.SLO{
				AppliesToIssues: true,
				GithubLabels:    []string{"labelle"},
			},
			issue: &maintner.GitHubIssue{
				Labels: map[int64]*maintner.GitHubLabel{
					1: &maintner.GitHubLabel{
						Name: "labelle",
					},
				},
			},
			want: true,
		},
		{
			name: "Applies with several labels",
			slo: &drghs_v1.SLO{
				AppliesToIssues: true,
				GithubLabels:    []string{"labelle"},
			},
			issue: &maintner.GitHubIssue{
				Labels: map[int64]*maintner.GitHubLabel{
					1: &maintner.GitHubLabel{
						Name: "labelle",
					},
					2: &maintner.GitHubLabel{
						Name: "extra label",
					},
				},
			},
			want: true,
		},
		{
			name: "Does not apply if issue does not have all labels",
			slo: &drghs_v1.SLO{
				AppliesToIssues: true,
				GithubLabels:    []string{"labelle", "this too"},
			},
			issue: &maintner.GitHubIssue{
				Labels: map[int64]*maintner.GitHubLabel{
					1: &maintner.GitHubLabel{
						Name: "labelle",
					},
				},
			},
			want: false,
		},
		{
			name: "Label order is unimportant",
			slo: &drghs_v1.SLO{
				AppliesToIssues: true,
				GithubLabels:    []string{"labelle", "l2"},
			},
			issue: &maintner.GitHubIssue{
				Labels: map[int64]*maintner.GitHubLabel{
					1: &maintner.GitHubLabel{
						Name: "extra",
					},
					2: &maintner.GitHubLabel{
						Name: "l2",
					},
					3: &maintner.GitHubLabel{
						Name: "labelle",
					},
				},
			},
			want: true,
		},
		{
			name: "Applies if issue doesn't have excluded label",
			slo: &drghs_v1.SLO{
				AppliesToIssues:      true,
				ExcludedGithubLabels: []string{"labelle"},
			},
			issue: &maintner.GitHubIssue{},
			want:  true,
		},
		{
			name: "Does not apply if issue has excluded label",
			slo: &drghs_v1.SLO{
				AppliesToIssues:      true,
				ExcludedGithubLabels: []string{"labelle"},
			},
			issue: &maintner.GitHubIssue{
				Labels: map[int64]*maintner.GitHubLabel{
					1: &maintner.GitHubLabel{
						Name: "labelle",
					},
				},
			},
			want: false,
		},
		{
			name: "Does not apply if issue has excluded label and others",
			slo: &drghs_v1.SLO{
				AppliesToIssues:      true,
				GithubLabels:         []string{"labelle"},
				ExcludedGithubLabels: []string{"extra label"},
			},
			issue: &maintner.GitHubIssue{
				Labels: map[int64]*maintner.GitHubLabel{
					1: &maintner.GitHubLabel{
						Name: "labelle",
					},
					2: &maintner.GitHubLabel{
						Name: "extra label",
					},
				},
			},
			want: false,
		},
		{
			name: "Excluded label order is unimportant",
			slo: &drghs_v1.SLO{
				AppliesToIssues:      true,
				ExcludedGithubLabels: []string{"labelle"},
			},
			issue: &maintner.GitHubIssue{
				Labels: map[int64]*maintner.GitHubLabel{
					1: &maintner.GitHubLabel{
						Name: "extra",
					},
					2: &maintner.GitHubLabel{
						Name: "l2",
					},
					3: &maintner.GitHubLabel{
						Name: "labelle",
					},
				},
			},
			want: false,
		},
	}

	for _, test := range tests {

		got := DoesSloApply(test.issue, test.slo)

		if got != test.want {
			t.Errorf("%v did not pass.\n\tWant:\t%v\n\tGot:\t%v", test.name, test.want, got)
		}
	}
}

func TestCompliantUntil(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		slo   *drghs_v1.SLO
		issue *maintner.GitHubIssue
		want  int64
	}{
		{
			name:  "Test nils",
			slo:   nil,
			issue: nil,
			want:  0,
		},
		{
			name: "Issue does not exist",
			slo: &drghs_v1.SLO{
				ResponseTime: ptypes.DurationProto(time.Second),
			},
			issue: &maintner.GitHubIssue{
				NotExist: true,
				Created:  time.Now().Add(-time.Minute),
			},
			want: 0,
		},
		{
			name: "Issue requires assignee and has valid assignee",
			slo: &drghs_v1.SLO{
				RequiresAssignee: true,
				Responders:       []string{"user1"},
			},
			issue: &maintner.GitHubIssue{
				Created: now.Add(-time.Minute),
				Assignees: []*maintner.GitHubUser{
					&maintner.GitHubUser{Login: "user2"},
					&maintner.GitHubUser{Login: "user1"},
				},
			},
			want: 0,
		},
		{
			name: "Issue requires assignee and does not have a valid assignee",
			slo: &drghs_v1.SLO{
				RequiresAssignee: true,
				Responders:       []string{"user1"},
			},
			issue: &maintner.GitHubIssue{
				Created: now.Add(-time.Minute),
				Assignees: []*maintner.GitHubUser{
					&maintner.GitHubUser{Login: "user2"},
				},
			},
			want: -60,
		},
		{
			name: "Issue has resolution time within SLO",
			slo: &drghs_v1.SLO{
				RequiresAssignee: false,
				ResolutionTime:   ptypes.DurationProto(time.Minute),
			},
			issue: &maintner.GitHubIssue{
				Created: now.Add(-time.Second),
				Closed:  false,
			},
			want: 59,
		},
		{
			name: "Issue has resolution time OOSLO",
			slo: &drghs_v1.SLO{
				RequiresAssignee: false,
				ResolutionTime:   ptypes.DurationProto(time.Minute),
			},
			issue: &maintner.GitHubIssue{
				Created: now.Add(-time.Minute * 2),
				Closed:  false,
			},
			want: -60,
		},
		{
			name: "Issue has resolution time OOSLO but is closed",
			slo: &drghs_v1.SLO{
				RequiresAssignee: false,
				ResolutionTime:   ptypes.DurationProto(time.Minute),
			},
			issue: &maintner.GitHubIssue{
				Created: now.Add(-time.Minute * 2),
				Closed:  true,
			},
			want: 0,
		},
		{
			name: "Issue has response time within SLO",
			slo: &drghs_v1.SLO{
				RequiresAssignee: false,
				ResponseTime:     ptypes.DurationProto(time.Minute),
			},
			issue: &maintner.GitHubIssue{
				Created: now.Add(-time.Second),
				Closed:  false,
			},
			want: 59,
		},
		{
			name: "Issue has response time OOSLO",
			slo: &drghs_v1.SLO{
				RequiresAssignee: false,
				ResponseTime:     ptypes.DurationProto(time.Minute),
			},
			issue: &maintner.GitHubIssue{
				Created: now.Add(-time.Minute * 2),
				Closed:  false,
			},
			want: -60,
		},
		{
			name: "Issue has response time within SLO but resolution time OOSLO",
			slo: &drghs_v1.SLO{
				RequiresAssignee: false,
				ResponseTime:     ptypes.DurationProto(time.Hour),
				ResolutionTime:   ptypes.DurationProto(time.Minute),
			},
			issue: &maintner.GitHubIssue{
				Created: now.Add(-time.Minute * 2),
				Closed:  false,
			},
			want: -60,
		},
		{
			name: "Issue has resolution time within SLO but response time OOSLO",
			slo: &drghs_v1.SLO{
				RequiresAssignee: false,
				ResponseTime:     ptypes.DurationProto(time.Minute),
				ResolutionTime:   ptypes.DurationProto(time.Hour),
			},
			issue: &maintner.GitHubIssue{
				Created: now.Add(-time.Minute * 2),
				Closed:  false,
			},
			want: -60,
		},
		{
			name: "Both resolution and response time are OOSLO, OOSLO from response time first",
			slo: &drghs_v1.SLO{
				RequiresAssignee: false,
				ResponseTime:     ptypes.DurationProto(time.Minute),
				ResolutionTime:   ptypes.DurationProto(time.Minute + time.Second),
			},
			issue: &maintner.GitHubIssue{
				Created: now.Add(-time.Minute * 2),
				Closed:  false,
			},
			want: -60,
		},
		{
			name: "Both resolution and response time are OOSLO, OOSLO from resolution time first",
			slo: &drghs_v1.SLO{
				RequiresAssignee: false,
				ResponseTime:     ptypes.DurationProto(time.Minute + time.Second*2),
				ResolutionTime:   ptypes.DurationProto(time.Minute + time.Second),
			},
			issue: &maintner.GitHubIssue{
				Created: now.Add(-time.Minute * 2),
				Closed:  false,
			},
			want: -59,
		},
		{
			name: "Both resolution and response time are in SLO, OOSLO from resolution time will happen first",
			slo: &drghs_v1.SLO{
				RequiresAssignee: false,
				ResponseTime:     ptypes.DurationProto(time.Minute + time.Second*2),
				ResolutionTime:   ptypes.DurationProto(time.Minute + time.Second),
			},
			issue: &maintner.GitHubIssue{
				Created: now.Add(-time.Minute),
				Closed:  false,
			},
			want: 1,
		},
		{
			name: "Both resolution and response time are in SLO, OOSLO from response time will happen first",
			slo: &drghs_v1.SLO{
				RequiresAssignee: false,
				ResponseTime:     ptypes.DurationProto(time.Minute + time.Second*2),
				ResolutionTime:   ptypes.DurationProto(time.Minute + time.Second*4),
			},
			issue: &maintner.GitHubIssue{
				Created: now.Add(-time.Minute),
				Closed:  false,
			},
			want: 2,
		},
		{
			name: "Requires and has assignee, with response and resolution time in SLO",
			slo: &drghs_v1.SLO{
				RequiresAssignee: true,
				Responders:       []string{"user1"},
				ResponseTime:     ptypes.DurationProto(time.Minute + time.Second*2),
				ResolutionTime:   ptypes.DurationProto(time.Minute + time.Second*4),
			},
			issue: &maintner.GitHubIssue{
				Created: now.Add(-time.Minute),
				Closed:  false,
				Assignees: []*maintner.GitHubUser{
					&maintner.GitHubUser{Login: "user1"},
				},
			},
			want: 2,
		},
		{
			name: "Requires and has assignee, with response time in SLO and resolution time OOSLO",
			slo: &drghs_v1.SLO{
				RequiresAssignee: true,
				Responders:       []string{"user1"},
				ResponseTime:     ptypes.DurationProto(time.Minute + time.Second*2),
				ResolutionTime:   ptypes.DurationProto(time.Minute - time.Second*4),
			},
			issue: &maintner.GitHubIssue{
				Created: now.Add(-time.Minute),
				Closed:  false,
				Assignees: []*maintner.GitHubUser{
					&maintner.GitHubUser{Login: "user1"},
				},
			},
			want: -4,
		},
		{
			name: "Requires and has assignee, with response time OOSLO and resolution time in SLO",
			slo: &drghs_v1.SLO{
				RequiresAssignee: true,
				Responders:       []string{"user1"},
				ResponseTime:     ptypes.DurationProto(time.Minute - time.Second*2),
				ResolutionTime:   ptypes.DurationProto(time.Minute + time.Second*4),
			},
			issue: &maintner.GitHubIssue{
				Created: now.Add(-time.Minute),
				Closed:  false,
				Assignees: []*maintner.GitHubUser{
					&maintner.GitHubUser{Login: "user1"},
				},
			},
			want: -2,
		},
		{
			name: "No valid assignee but both resolution and response time are in SLO, OOSLO from response time will happen first",
			slo: &drghs_v1.SLO{
				RequiresAssignee: true,
				ResponseTime:     ptypes.DurationProto(time.Minute + time.Second*2),
				ResolutionTime:   ptypes.DurationProto(time.Minute + time.Second*4),
			},
			issue: &maintner.GitHubIssue{
				Created: now.Add(-time.Minute),
				Closed:  false,
			},
			want: -60,
		},
		{
			name: "No valid assignee, with response time in SLO and resolution time OOSLO",
			slo: &drghs_v1.SLO{
				RequiresAssignee: true,
				Responders:       []string{"user1"},
				ResponseTime:     ptypes.DurationProto(time.Minute + time.Second*2),
				ResolutionTime:   ptypes.DurationProto(time.Minute - time.Second*4),
			},
			issue: &maintner.GitHubIssue{
				Created: now.Add(-time.Minute),
				Closed:  false,
				Assignees: []*maintner.GitHubUser{
					&maintner.GitHubUser{Login: "user2"},
				},
			},
			want: -60,
		},
	}

	for _, test := range tests {

		got := CompliantUntil(test.issue, test.slo, now)

		if got != test.want {
			t.Errorf("%v did not pass.\n\tWant:\t%v\n\tGot:\t%v", test.name, test.want, got)
		}
	}
}

func TestEarliest(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		t1   time.Time
		t2   time.Time
		want time.Time
	}{
		{
			name: "All zeroes",
		},
		{
			name: "t1 before t2",
			t1:   now,
			t2:   now.Add(time.Second),
			want: now,
		},
		{
			name: "t2 before t1",
			t1:   now.Add(time.Second),
			t2:   now,
			want: now,
		},
		{
			name: "Zero value t1",
			t2:   now,
			want: now,
		},
		{
			name: "Zero value t2",
			t1:   now,
			want: now,
		},
	}
	for _, test := range tests {

		got := earliest(test.t1, test.t2)

		if got != test.want {
			t.Errorf("%v did not pass.\n\tWant:\t%v\n\tGot:\t%v", test.name, test.want, got)
		}
	}
}
