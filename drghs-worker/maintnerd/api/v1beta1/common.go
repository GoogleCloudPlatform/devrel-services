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

package v1beta1

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/devrel-services/drghs-worker/pkg/sloutils"
	"github.com/GoogleCloudPlatform/devrel-services/drghs-worker/pkg/utils"
	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/golang/protobuf/ptypes"
	"golang.org/x/build/maintner"
)

var bugLabels = []string{
	"bug",
	"type: bug",
	"type:bug",
	"kind/bug",
	"end-to-end bugs",
	"type:bug/performance",
}

func makeRepoPB(repo *maintner.GitHubRepo) (*drghs_v1.Repository, error) {
	rID := repo.ID()
	nIss := 0
	nPr := 0

	err := repo.ForeachIssue(func(i *maintner.GitHubIssue) error {
		if i.PullRequest {
			nPr = nPr + 1
		} else {
			nIss = nIss + 1
		}
		return nil

	})
	if err != nil {
		return nil, err
	}

	return &drghs_v1.Repository{
		Name:             fmt.Sprintf("%v/%v", rID.Owner, rID.Repo),
		IssueCount:       int32(nIss),
		PullRequestCount: int32(nPr),
	}, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func makeIssuePB(issue *maintner.GitHubIssue, rID maintner.GitHubRepoID, includeComments bool, includeReviews bool, fm *field_mask.FieldMask, slos []*drghs_v1.SLO) (*drghs_v1.Issue, error) {
	paths := fm.GetPaths()
	riss := &drghs_v1.Issue{}

	if paths == nil || contains(paths, "created_at") {
		createdAt, err := ptypes.TimestampProto(issue.Created)
		if err != nil {
			return nil, err
		}
		riss.CreatedAt = createdAt
	}

	if paths == nil || contains(paths, "updated_at") {
		updatedAt, err := ptypes.TimestampProto(issue.Updated)
		if err != nil {
			return nil, err
		}
		riss.UpdatedAt = updatedAt
	}

	if paths == nil || contains(paths, "closed_at") {
		closedAt, err := ptypes.TimestampProto(issue.ClosedAt)
		if err != nil {
			return nil, err
		}
		riss.ClosedAt = closedAt
	}

	if paths == nil || contains(paths, "closed_by") {
		closedBy, err := makeUserPB(issue.ClosedBy)
		if err != nil {
			return nil, err
		}
		riss.ClosedBy = closedBy
	}

	if paths == nil || contains(paths, "reporter") {
		reporter, err := makeUserPB(issue.User)
		if err != nil {
			return nil, err
		}
		riss.Reporter = reporter
	}

	if paths == nil || contains(paths, "assignees") {
		assignees := make([]*drghs_v1.GitHubUser, len(issue.Assignees))
		for i, assign := range issue.Assignees {
			u, err := makeUserPB(assign)
			if err != nil {
				return nil, err
			}
			assignees[i] = u
		}
		riss.Assignees = assignees
	}

	if paths == nil || contains(paths, "commit") {
		commitID := ""
		issue.ForeachEvent(func(event *maintner.GitHubIssueEvent) error {
			// ForeachEvent processes events in chronological order
			if event.CommitID != "" {
				commitID = event.CommitID
			}
			return nil
		})
		riss.Commit = commitID
	}

	if paths == nil || contains(paths, "closed") {
		riss.Closed = issue.Closed
	}

	if paths == nil || contains(paths, "is_pr") {
		riss.IsPr = issue.PullRequest
	}

	if paths == nil || contains(paths, "title") {
		riss.Title = issue.Title
	}

	if paths == nil || contains(paths, "body") {
		riss.Body = issue.Body
	}

	if paths == nil || contains(paths, "issue_id") {
		riss.IssueId = issue.Number
	}

	if paths == nil || contains(paths, "approved") {
		riss.Approved = utils.IsApproved(issue)
	}

	if paths == nil || contains(paths, "url") {
		riss.Url = fmt.Sprintf("https://github.com/%v/%v/issues/%d", rID.Owner, rID.Repo, issue.Number)
	}

	if paths == nil || contains(paths, "repo") {
		riss.Repo = fmt.Sprintf("%v/%v", rID.Owner, rID.Repo)
	}

	labels := make([]string, len(issue.Labels))
	{
		i := 0
		for _, l := range issue.Labels {
			labels[i] = l.Name
			i++
		}
		sort.Slice(labels, func(i, j int) bool { return strings.ToLower(labels[i]) < strings.ToLower(labels[j]) })
	}

	if paths == nil || contains(paths, "labels") {
		riss.Labels = labels
	}

	fillFromLabels(riss, labels, fm)

	if includeComments || (paths != nil && contains(paths, "comments")) {
		riss.Comments = make([]*drghs_v1.GitHubComment, 0)
		err := issue.ForeachComment(func(co *maintner.GitHubComment) error {
			cpb, err := makeCommentPB(co)
			if err != nil {
				return err
			}

			riss.Comments = append(riss.Comments, cpb)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	if includeReviews || (paths != nil && contains(paths, "reviews")) {
		riss.Reviews = make([]*drghs_v1.GitHubReview, 0)
		err := issue.ForeachReview(func(rev *maintner.GitHubReview) error {
			rpb, err := makeReviewPB(rev)
			if err != nil {
				return err
			}

			riss.Reviews = append(riss.Reviews, rpb)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	var sloBudget int64
	sloBudget = math.MaxInt64
	for _, slo := range slos {
		if sloutils.DoesSloApply(issue, slo) {
			compliantUntil := sloutils.CompliantUntil(issue, slo, time.Now())
			if compliantUntil < sloBudget {
				sloBudget = compliantUntil
			}
		}
	}
	if sloBudget == math.MaxInt64 {
		sloBudget = 0
	}

	riss.SloBudget = sloBudget

	return riss, nil
}

func fillFromLabels(s *drghs_v1.Issue, labels []string, fm *field_mask.FieldMask) {
	priority := drghs_v1.Issue_PRIORITY_UNSPECIFIED
	priorityUnknown := true
	issueType := drghs_v1.Issue_GITHUB_ISSUE_TYPE_UNSPECIFIED
	blocked := false
	releaseBlocking := false

	for _, l := range labels {
		lowercaseName := strings.ToLower(l)
		if priorityUnknown == true {
			switch {
			case strings.Contains(lowercaseName, "p0"):
				priority = drghs_v1.Issue_P0
				priorityUnknown = false
			case strings.Contains(lowercaseName, "p1"):
				priority = drghs_v1.Issue_P1
				priorityUnknown = false
			case strings.Contains(lowercaseName, "p2"):
				priority = drghs_v1.Issue_P2
				priorityUnknown = false
			case strings.Contains(lowercaseName, "p3"):
				priority = drghs_v1.Issue_P3
				priorityUnknown = false
			case strings.Contains(lowercaseName, "p4"):
				priority = drghs_v1.Issue_P4
				priorityUnknown = false
			}
		}

		if issueType == drghs_v1.Issue_GITHUB_ISSUE_TYPE_UNSPECIFIED {
			switch {
			case matchesAny(lowercaseName, bugLabels):
				issueType = drghs_v1.Issue_BUG
			case strings.Contains(lowercaseName, "enhanc"):
				issueType = drghs_v1.Issue_FEATURE
			case strings.Contains(lowercaseName, "feat"):
				issueType = drghs_v1.Issue_FEATURE
			case strings.Contains(lowercaseName, "addition"):
				issueType = drghs_v1.Issue_FEATURE
			case strings.Contains(lowercaseName, "question"):
				issueType = drghs_v1.Issue_QUESTION
			case strings.Contains(lowercaseName, "cleanup"):
				issueType = drghs_v1.Issue_CLEANUP
			case strings.Contains(lowercaseName, "process"):
				issueType = drghs_v1.Issue_PROCESS
			}
		}

		switch {
		case strings.Contains(lowercaseName, "blocked"):
			blocked = true
		case strings.Contains(lowercaseName, "blocking"):
			releaseBlocking = true
		}
	}

	paths := fm.GetPaths()
	if paths == nil || contains(paths, "priority") {
		s.Priority = priority
	}
	if paths == nil || contains(paths, "priority_unknown") {
		s.PriorityUnknown = priorityUnknown
	}
	if paths == nil || contains(paths, "issue_type") {
		s.IssueType = issueType
	}
	if paths == nil || contains(paths, "blocked") {
		s.Blocked = blocked
	}
	if paths == nil || contains(paths, "release_blocking") {
		s.ReleaseBlocking = releaseBlocking
	}
}

func makeCommentPB(comment *maintner.GitHubComment) (*drghs_v1.GitHubComment, error) {
	createdAt, err := ptypes.TimestampProto(comment.Created)
	if err != nil {
		return nil, err
	}
	updatedAt, err := ptypes.TimestampProto(comment.Updated)
	if err != nil {
		return nil, err
	}

	user, err := makeUserPB(comment.User)
	if err != nil {
		return nil, err
	}

	return &drghs_v1.GitHubComment{
		Id:        int32(comment.ID),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		User:      user,
		Body:      comment.Body,
	}, nil
}

func makeReviewPB(review *maintner.GitHubReview) (*drghs_v1.GitHubReview, error) {
	createdAt, err := ptypes.TimestampProto(review.Created)
	if err != nil {
		return nil, err
	}

	actor, err := makeUserPB(review.Actor)
	if err != nil {
		return nil, err
	}

	return &drghs_v1.GitHubReview{
		Id:               int32(review.ID),
		CreatedAt:        createdAt,
		Actor:            actor,
		Body:             review.Body,
		State:            review.State,
		ActorAssociation: review.ActorAssociation,
	}, nil
}

func makeUserPB(user *maintner.GitHubUser) (*drghs_v1.GitHubUser, error) {
	if user == nil {
		return nil, nil
	}
	return &drghs_v1.GitHubUser{
		Id:    int32(user.ID),
		Login: user.Login,
	}, nil
}

func matchesAny(item string, valuesToMatch []string) bool {
	for _, valueToMatch := range valuesToMatch {
		if item == valueToMatch {
			return true
		}
	}
	return false
}
