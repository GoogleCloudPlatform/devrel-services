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
	"time"

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"
	"golang.org/x/build/maintner"
)

// DoesSloApply determines if the given SLO applies to the given issue
func DoesSloApply(issue *maintner.GitHubIssue, slo *drghs_v1.SLO) bool {

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

// IsCompliant determines if the given issue is compliant with the given SLO rule
func IsCompliant(issue *maintner.GitHubIssue, slo *drghs_v1.SLO) bool {

	if issue.Closed || issue.NotExist {
		return true
	}

	now := time.Now()
	validResponders := getValidResponders(issue, slo)

	// check response time
	shouldBeRepliedBy := issue.Created.Add(slo.GetResponseTime().AsDuration())
	if slo.GetResponseTime().AsDuration() > 0 && now.After(shouldBeRepliedBy) {
		//check if a valid person has replied

		validResponderReplied := false

		issue.ForeachComment(func(comment *maintner.GitHubComment) error {
			if validResponderReplied {
				return nil
			}

			_, isResponder := validResponders[comment.User.Login]
			if isResponder {
				validResponderReplied = true
			}
			return nil
		})

		if !validResponderReplied {
			return false
		}
	}

	// check resolution time
	shouldBeResolvedBy := issue.Created.Add(slo.GetResolutionTime().AsDuration())

	if slo.GetResolutionTime().AsDuration() > 0 && now.After(shouldBeResolvedBy) {
		return false
	}

	if slo.GetRequiresAssignee() {

		for _, assignee := range issue.Assignees {
			_, isResponder := validResponders[assignee.Login]

			if isResponder {
				return true
			}
		}
		// went through all assignees, none are valid responders
		return false
	}

	return true
}

func getValidResponders(issue *maintner.GitHubIssue, slo *drghs_v1.SLO) map[string]struct{} {
	responders := make(map[string]struct{})

	for _, r := range slo.GetResponders() {
		responders[r] = struct{}{}
	}

	return responders
}
