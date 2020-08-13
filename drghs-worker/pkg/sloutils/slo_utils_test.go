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

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"
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

		got := DoesSloApply(test.slo, test.issue)

		if got != test.want {
			t.Errorf("%v did not pass.\n\tWant:\t%v\n\tGot:\t%v", test.name, test.want, got)
		}
	}
}
