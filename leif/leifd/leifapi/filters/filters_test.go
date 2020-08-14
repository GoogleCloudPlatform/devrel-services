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

package filter

import (
	"testing"

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"

	"github.com/google/go-cmp/cmp"
)

func TestOwner(t *testing.T) {
	tests := []struct {
		name         string
		owner        *drghs_v1.Owner
		filter       string
		want         bool
		wantErr      bool
		wantBuildErr bool
	}{
		{
			name:         "Empty filter passes",
			owner:        &drghs_v1.Owner{},
			filter:       "",
			want:         true,
			wantErr:      false,
			wantBuildErr: false,
		},
		{
			name:         "Nil Owner is filtered out",
			owner:        nil,
			filter:       "",
			want:         false,
			wantErr:      false,
			wantBuildErr: false,
		},
		{
			name: "Filter name passes",
			owner: &drghs_v1.Owner{
				Name: "foo",
			},
			filter:       "owner.name == 'foo' ",
			want:         true,
			wantErr:      false,
			wantBuildErr: false,
		},
		{
			name: "Unsupported field fails",
			owner: &drghs_v1.Owner{
				Name: "foo",
			},
			filter:       "field == 'foo'",
			want:         false,
			wantErr:      false,
			wantBuildErr: true,
		},
		{
			name: "Incorrect format fails",
			owner: &drghs_v1.Owner{
				Name: "foo",
			},
			filter:       "baz: foo",
			want:         false,
			wantErr:      false,
			wantBuildErr: true,
		},
	}
	for _, test := range tests {
		prgm, gotBuildErr := BuildOwnerFilter(test.filter)

		got, gotErr := Owner(test.owner, prgm)
		if (test.wantErr && gotErr == nil) || (!test.wantErr && gotErr != nil) {
			t.Errorf("test: %v, errors diff. WantErr: %v, GotErr: %v.", test.name, test.wantErr, gotErr)
		}
		if (test.wantBuildErr && gotBuildErr == nil) || (!test.wantBuildErr && gotBuildErr != nil) {
			t.Errorf("test: %v, errors diff. WantErr: %v, GotErr: %v.", test.name, test.wantBuildErr, gotBuildErr)
		}
		if diff := cmp.Diff(test.want, got); diff != "" {
			t.Errorf("test: %v, values diff. match (-want +got)\n%s", test.name, diff)
		}
	}
}

func TestRepository(t *testing.T) {
	tests := []struct {
		name         string
		repo         *drghs_v1.Repository
		filter       string
		want         bool
		wantErr      bool
		wantBuildErr bool
	}{
		{
			name:         "Empty filter passes",
			repo:         &drghs_v1.Repository{},
			filter:       "",
			want:         true,
			wantErr:      false,
			wantBuildErr: false,
		},
		{
			name:         "Nil Repository is filtered out",
			repo:         nil,
			filter:       "",
			want:         false,
			wantErr:      false,
			wantBuildErr: false,
		},
		{
			name: "Filter name passes",
			repo: &drghs_v1.Repository{
				Name: "foo",
			},
			filter:       "repository.name == 'foo' ",
			want:         true,
			wantErr:      false,
			wantBuildErr: false,
		},
		{
			name: "Unsupported field fails",
			repo: &drghs_v1.Repository{
				Name: "foo",
			},
			filter:       "baz == 'foo'",
			want:         false,
			wantErr:      false,
			wantBuildErr: true,
		},
		{
			name: "Incorrect format fails",
			repo: &drghs_v1.Repository{
				Name: "foo",
			},
			filter:       "baz: foo",
			want:         false,
			wantErr:      false,
			wantBuildErr: true,
		},
	}
	for _, test := range tests {
		p, gotBuildErr := BuildRepositoryFilter(test.filter)
		got, gotErr := Repository(test.repo, p)

		if (test.wantErr && gotErr == nil) || (!test.wantErr && gotErr != nil) {
			t.Errorf("test: %v, errors diff. WantErr: %v, GotErr: %v.", test.name, test.wantErr, gotErr)
		}
		if (test.wantBuildErr && gotBuildErr == nil) || (!test.wantBuildErr && gotBuildErr != nil) {
			t.Errorf("test: %v, errors diff. WantErr: %v, GotErr: %v.", test.name, test.wantBuildErr, gotBuildErr)
		}
		if diff := cmp.Diff(test.want, got); diff != "" {
			t.Errorf("test: %v, values diff. match (-want +got)\n%s", test.name, diff)
		}
	}
}

func TestSlo(t *testing.T) {
	tests := []struct {
		name         string
		slo          *drghs_v1.SLO
		filter       string
		want         bool
		wantErr      bool
		wantBuildErr bool
	}{
		{
			name:         "Empty filter passes",
			slo:          &drghs_v1.SLO{},
			filter:       "",
			want:         true,
			wantErr:      false,
			wantBuildErr: false,
		},
		{
			name:         "Nil slo is filtered out",
			slo:          nil,
			filter:       "",
			want:         false,
			wantErr:      false,
			wantBuildErr: false,
		},
		{
			name: "Filter applies to issues passes",
			slo: &drghs_v1.SLO{
				RequiresAssignee: true,
				AppliesToIssues:  true,
			},
			filter:       "slo.applies_to_issues == true ",
			want:         true,
			wantErr:      false,
			wantBuildErr: false,
		},
		{
			name: "Filter applies to prs passes",
			slo: &drghs_v1.SLO{
				RequiresAssignee: true,
				AppliesToPrs:     true,
			},
			filter:       "slo.applies_to_prs == true ",
			want:         true,
			wantErr:      false,
			wantBuildErr: false,
		},
		{
			name: "Filter github labels passes",
			slo: &drghs_v1.SLO{
				GithubLabels:     []string{"label"},
				RequiresAssignee: true,
				AppliesToPrs:     true,
			},
			filter:       "slo.github_labels == ['label']",
			want:         true,
			wantErr:      false,
			wantBuildErr: false,
		},
		{
			name: "Filter reqs assignee passes",
			slo: &drghs_v1.SLO{
				RequiresAssignee: true,
			},
			filter:       "slo.requires_assignee == true ",
			want:         true,
			wantErr:      false,
			wantBuildErr: false,
		},
		{
			name: "Filters several fields passes",
			slo: &drghs_v1.SLO{
				RequiresAssignee: true,
				AppliesToPrs:     true,
			},
			filter:       "slo.requires_assignee == true && slo.applies_to_prs == true",
			want:         true,
			wantErr:      false,
			wantBuildErr: false,
		},
		{
			name: "Filters several fields filters with both fields",
			slo: &drghs_v1.SLO{
				RequiresAssignee: true,
			},
			filter:       "slo.requires_assignee == true && slo.applies_to_prs == true",
			want:         false,
			wantErr:      false,
			wantBuildErr: false,
		},
		{
			name: "Unsupported field fails",
			slo: &drghs_v1.SLO{
				RequiresAssignee: true,
			},
			filter:       "slo.baz == true ",
			want:         false,
			wantErr:      false,
			wantBuildErr: true,
		},
		{
			name: "Incorrect format fails",
			slo: &drghs_v1.SLO{
				RequiresAssignee: true,
			},
			filter:       "baz: foo",
			want:         false,
			wantErr:      false,
			wantBuildErr: true,
		},
	}
	for _, test := range tests {
		p, gotBuildErr := BuildSloFilter(test.filter)
		got, gotErr := Slo(test.slo, p)

		if (test.wantErr && gotErr == nil) || (!test.wantErr && gotErr != nil) {
			t.Errorf("test: %v, errors diff. WantErr: %v, GotErr: %v.", test.name, test.wantErr, gotErr)
		}
		if (test.wantBuildErr && gotBuildErr == nil) || (!test.wantBuildErr && gotBuildErr != nil) {
			t.Errorf("test: %v, errors diff. WantErr: %v, GotErr: %v.", test.name, test.wantBuildErr, gotBuildErr)
		}
		if diff := cmp.Diff(test.want, got); diff != "" {
			t.Errorf("test: %v, values diff. match (-want +got)\n%s", test.name, diff)
		}
	}
}
