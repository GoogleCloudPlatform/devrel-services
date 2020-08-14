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

// CompliantUntil returns the seconds until the given issue is no longer compliant with the given SLO rule
// If the given issue is OOSLO (not compliant with the given SLO rule), then it will return
// a negative value representing the seconds since it became OOSLO
// If the issue will always be compliant, this returns 0
func CompliantUntil(issue *maintner.GitHubIssue, slo *drghs_v1.SLO, now time.Time) int64 {

	if issue == nil || slo == nil || issue.Closed || issue.NotExist {
		return 0
	}

	validResponders := getValidResponders(issue, slo)

	var inSloUntil time.Time

	// check assignees
	if slo.GetRequiresAssignee() {
		hasValidAssignee := false
		for _, assignee := range issue.Assignees {
			_, isResponder := validResponders[assignee.Login]

			if isResponder {
				hasValidAssignee = true
				break
			}
		}

		if !hasValidAssignee {
			inSloUntil = issue.Created
		}
	}

	// check resolution time
	if slo.GetResolutionTime().AsDuration() > 0 {
		shouldBeResolvedBy := issue.Created.Add(slo.GetResolutionTime().AsDuration())
		inSloUntil = earliest(inSloUntil, shouldBeResolvedBy)
	}

	// check response time
	if slo.GetResponseTime().AsDuration() > 0 {
		shouldBeRepliedBy := issue.Created.Add(slo.GetResponseTime().AsDuration())

		if shouldBeRepliedBy.After(now) {
			//in slo for response time
			inSloUntil = earliest(inSloUntil, shouldBeRepliedBy)
		} else if inSloUntil.IsZero() || inSloUntil.After(shouldBeRepliedBy) {
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
				inSloUntil = shouldBeRepliedBy
			}
		}

	}

	if inSloUntil.IsZero() {
		return 0
	}

	//get seconds between now and inslotime
	sec := inSloUntil.Sub(now)

	return int64(sec.Seconds())
}

func getValidResponders(issue *maintner.GitHubIssue, slo *drghs_v1.SLO) map[string]struct{} {
	responders := make(map[string]struct{})

	for _, r := range slo.GetResponders() {
		responders[r] = struct{}{}
	}

	return responders
}

// returns earliest non-zero time
func earliest(t1 time.Time, t2 time.Time) time.Time {
	if t1.IsZero() || (!t2.IsZero() && t1.After(t2)) {
		return t2
	}
	return t1
}
