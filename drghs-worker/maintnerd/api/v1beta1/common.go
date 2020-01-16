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

	"github.com/GoogleCloudPlatform/devrel-services/drghs-worker/pkg/utils"
	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"

	"github.com/golang/protobuf/ptypes"
	"golang.org/x/build/maintner"
)

func makeRepoPB(repo *maintner.GitHubRepo) (*drghs_v1.Repository, error) {
	rID := repo.ID()
	return &drghs_v1.Repository{
		Name: fmt.Sprintf("%v/%v", rID.Owner, rID.Repo),
	}, nil
}

func makeIssuePB(issue *maintner.GitHubIssue, rID maintner.GitHubRepoID, includeComments, includeReviews bool) (*drghs_v1.Issue, error) {

	createdAt, err := ptypes.TimestampProto(issue.Created)
	if err != nil {
		return nil, err
	}
	updatedAt, err := ptypes.TimestampProto(issue.Updated)
	if err != nil {
		return nil, err
	}

	closedBy, err := makeUserPB(issue.ClosedBy)
	if err != nil {
		return nil, err
	}
	reporter, err := makeUserPB(issue.User)
	if err != nil {
		return nil, err
	}

	assignees := make([]*drghs_v1.GitHubUser, len(issue.Assignees))
	for i, assign := range issue.Assignees {
		u, err := makeUserPB(assign)
		if err != nil {
			return nil, err
		}
		assignees[i] = u
	}

	riss := &drghs_v1.Issue{
		Priority:  drghs_v1.Issue_P2,
		IsPr:      issue.PullRequest,
		Approved:  utils.IsApproved(issue),
		Title:     issue.Title,
		Body:      issue.Body,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Closed:    issue.Closed,
		ClosedBy:  closedBy,
		GitCommit: nil,
		IssueId:   issue.Number,
		Assignees: assignees,
		Reporter:  reporter,
		Url:       fmt.Sprintf("https://github.com/%v/%v/issues/%d", rID.Owner, rID.Repo, issue.Number),
	}

	labels := make([]string, len(issue.Labels))
	{
		i := 0
		for _, l := range issue.Labels {
			labels[i] = l.Name
			i++
		}
	}

	riss.Labels = labels

	if includeComments {
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

	if includeReviews {
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

	return riss, nil
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
