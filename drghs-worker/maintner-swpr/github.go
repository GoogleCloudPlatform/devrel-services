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

package main

import (
	"context"
	"strings"

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"
	"github.com/shurcooL/githubv4"
)

type issue struct {
	ID     string
	Number int32
}

type pullRequest struct {
	ID     string
	Number int32
}

type ghIssuesQuery struct {
	Repository struct {
		Issues struct {
			Nodes    []issue
			PageInfo struct {
				EndCursor   githubv4.String
				HasNextPage bool
			}
		} `graphql:"issues(first: 100, after: $cursor)"` // 100 per page.
	} `graphql:"repository(owner: $repositoryOwner, name: $repositoryName)"`
}

type ghPullRequestsQuery struct {
	Repository struct {
		PullRequests struct {
			Nodes    []pullRequest
			PageInfo struct {
				EndCursor   githubv4.String
				HasNextPage bool
			}
		} `graphql:"pullRequests(first: 100, after: $cursor)"` // 100 per page.
	} `graphql:"repository(owner: $repositoryOwner, name: $repositoryName)"`
}

func getGitHubIssuesForRepo(ctx context.Context, c *githubv4.Client, repo *drghs_v1.Repository) ([]issue, error) {
	log.Debugf("getting GitHub issues for: %v", repo.String())

	parts := strings.Split(repo.GetName(), "/")

	var q ghIssuesQuery
	variables := map[string]interface{}{
		"repositoryOwner": githubv4.String(parts[0]),
		"repositoryName":  githubv4.String(parts[1]),
		"cursor":          (*githubv4.String)(nil), // Null after argument to get first page.
	}
	// Get issues from all pages.
	var allIssues []issue
	var pageN int
	for {
		err := c.Query(ctx, &q, variables)
		if err != nil {
			return make([]issue, 0), err
		}
		allIssues = append(allIssues, q.Repository.Issues.Nodes...)
		if !q.Repository.Issues.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(q.Repository.Issues.PageInfo.EndCursor)
		log.Debugf("getting GitHub issues for repo: %v. finished page: %v. current count: %v", repo.Name, pageN, len(allIssues))
		pageN++
	}

	log.Debugf("finished getting GitHub issues for: %v. returning issues count: %v", repo.String(), len(allIssues))

	return allIssues, nil
}

func getGitHubPullRequestsForRepo(ctx context.Context, c *githubv4.Client, repo *drghs_v1.Repository) ([]pullRequest, error) {
	log.Debugf("getting GitHub pull requests for: %v", repo.String())

	parts := strings.Split(repo.GetName(), "/")

	var q ghPullRequestsQuery
	variables := map[string]interface{}{
		"repositoryOwner": githubv4.String(parts[0]),
		"repositoryName":  githubv4.String(parts[1]),
		"cursor":          (*githubv4.String)(nil), // Null after argument to get first page.
	}
	// Get pullRequests from all pages.
	var allPullRequests []pullRequest
	var pageN int
	for {
		err := c.Query(ctx, &q, variables)
		if err != nil {
			return make([]pullRequest, 0), err
		}
		allPullRequests = append(allPullRequests, q.Repository.PullRequests.Nodes...)
		if !q.Repository.PullRequests.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(q.Repository.PullRequests.PageInfo.EndCursor)
		log.Debugf("getting GitHub PullRequests for repo: %v. finished page: %v. current count: %v", repo.Name, pageN, len(allPullRequests))
		pageN++
	}
	log.Debugf("finished getting GitHub PullRequests for: %v. returning PullRequests count: %v", repo.String(), len(allPullRequests))

	return allPullRequests, nil
}
