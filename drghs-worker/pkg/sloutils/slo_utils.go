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
	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"
	"golang.org/x/build/maintner"
)

// DoesSloApply determines if the given SLO applies to the given issue
func DoesSloApply(slo *drghs_v1.SLO, issue *maintner.GitHubIssue) bool {

	if issue == nil || slo == nil {
		return false
	}

	if issue.NotExist {
		return false
	}

	if (issue.PullRequest && !slo.GetAppliesToPrs()) ||
		(!issue.PullRequest && !slo.GetAppliesToIssues()) {
		return false
	}

	for _, label := range slo.GetGithubLabels() {
		if !issue.HasLabel(label) {
			return false
		}
	}

	for _, label := range slo.GetExcludedGithubLabels() {
		if issue.HasLabel(label) {
			return false
		}
	}

	return true
}
