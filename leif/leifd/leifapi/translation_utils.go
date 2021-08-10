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

package leifapi

import (
	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"
	"github.com/GoogleCloudPlatform/devrel-services/leif"
	"github.com/golang/protobuf/ptypes"
)

func makeOwnerPB(name string) (*drghs_v1.Owner, error) {
	return &drghs_v1.Owner{
		Name: name,
	}, nil
}

func makeRepositoryPB(rname string) (*drghs_v1.Repository, error) {
	return &drghs_v1.Repository{
		Name: rname,
	}, nil
}

func makeSloPB(slo *leif.SLORule) (*drghs_v1.SLO, error) {
	return &drghs_v1.SLO{
		GithubLabels:         slo.AppliesTo.GitHubLabels,
		ExcludedGithubLabels: slo.AppliesTo.ExcludedGitHubLabels,
		AppliesToIssues:      slo.AppliesTo.Issues,
		AppliesToPrs:         slo.AppliesTo.PRs,
		ResponseTime:         ptypes.DurationProto(slo.ComplianceSettings.ResponseTime),
		ResolutionTime:       ptypes.DurationProto(slo.ComplianceSettings.ResolutionTime),
		RequiresAssignee:     slo.ComplianceSettings.RequiresAssignee,
		Responders:           slo.ComplianceSettings.Responders,
	}, nil
}
