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

package utils

import (
	"github.com/GoogleCloudPlatform/devrel-services/drghs-worker/pkg/status"

	"golang.org/x/build/maintner"
)

// TranslateIssueToStatus takes the given maintner.GitHubIssue and translates it to a status.Status
func TranslateIssueToStatus(issue *maintner.GitHubIssue, repoID string, includeComments bool, includeReviews bool) *status.Status {

	commitID := ""
	issue.ForeachEvent(func(event *maintner.GitHubIssueEvent) error {
		// ForeachEvent processes events in chronological order
		if event.CommitID != "" {
			commitID = event.CommitID
		}
		return nil
	})

	s := &status.Status{
		Issue:           issue,
		Repo:            repoID,
		Priority:        status.P2,
		PriorityUnknown: true,
		PullRequest:     issue.PullRequest,
		Approved:        IsApproved(issue),
		Title:           issue.Title,
		Body:            issue.Body,
		Created:         issue.Created,
		UpdatedAt:       issue.Updated,
		Closed:          issue.Closed,
		ClosedBy:        issue.ClosedBy,
		Commit:          commitIDcommitID,
		IssueID:         issue.Number,
		Assignees:       issue.Assignees,
		Reporter:        issue.User,
	}

	if issue.Closed {
		s.ClosedAt = &issue.ClosedAt
	}

	if includeComments {
		comments := make([]*maintner.GitHubComment, 0)

		issue.ForeachComment(func(comment *maintner.GitHubComment) error {
			comments = append(comments, comment)
			return nil
		})

		s.Comments = comments
	}

	if includeReviews {
		reviews := make([]*maintner.GitHubReview, 0)
		issue.ForeachReview(func(review *maintner.GitHubReview) error {
			reviews = append(reviews, review)
			return nil
		})
		s.Reviews = reviews
	}

	s.FillLabels()
	s.URL = s.Url()

	return s
}

// IsApproved considers a GitHubIssue, loops through the
// Review Events on the issue and, for each reivewer,
// if the last review event was an "approved" event it returns true
// if there are no reviews, this return false
func IsApproved(issue *maintner.GitHubIssue) bool {
	if issue == nil {
		return false
	}

	reviewers := make(map[*maintner.GitHubUser]bool)

	// ForeachReview processes reviews in chronological
	// order. We can just call this serially and if a
	// reviewer ever requests changes after approving
	// this will still set the final review to 'false'
	issue.ForeachReview(func(review *maintner.GitHubReview) error {
		reviewers[review.Actor] = review.State == "APPROVED"
		return nil
	})

	// If there are no reviewers, we shall state that it is not approved
	if len(reviewers) == 0 {
		return false
	}

	for _, approved := range reviewers {
		if !approved {
			return false
		}
	}

	return true
}
