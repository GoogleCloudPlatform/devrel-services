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
	"context"
	"fmt"
	"strings"
	"time"

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"

	"github.com/GoogleCloudPlatform/devrel-services/drghs-worker/maintnerd/api/filters"
	"github.com/GoogleCloudPlatform/devrel-services/drghs-worker/pkg/googlers"

	"golang.org/x/build/maintner"

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

var _ drghs_v1.IssueServiceServer = &issueServiceV1{}

const defaultFilter = "true"

type issueServiceV1 struct {
	corpus *maintner.Corpus
	rp     *repoPaginator
}

// NewIssueServiceV1 returns a service that implements
// drghs_v1.IssueServiceServer
func NewIssueServiceV1(corpus *maintner.Corpus, resolver googlers.GooglersResolver) drghs_v1.IssueServiceServer {
	return &issueServiceV1{
		corpus: corpus,
		rp: &repoPaginator{
			set: make(map[time.Time]repoPage),
		},
	}
}

func (s *issueServiceV1) ListRepositories(ctx context.Context, r *drghs_v1.ListRepositoriesRequest) (*drghs_v1.ListRepositoriesResponse, error) {

	var pg []*drghs_v1.Repository
	var idx int
	var err error
	nextToken := ""

	if r.PageToken != "" {
		pageToken, err := decodePageToken(r.PageToken)
		if err != nil {
			return nil, err
		}

		ftime, err := ptypes.Timestamp(pageToken.FirstRequestTimeUsec)
		if err != nil {
			return nil, err
		}

		pagesize := getPageSize(int(r.PageSize))

		pg, idx, err = s.rp.GetPage(ftime, pagesize)
		if err != nil {
			return nil, err
		}
		nextToken, err = makeNextPageToken(pageToken, idx)

	} else {
		filteredRepos := make([]*drghs_v1.Repository, 0)
		err = s.corpus.GitHub().ForeachRepo(func(repo *maintner.GitHubRepo) error {
			rpb, err := makeRepoPB(repo)
			if err != nil {
				return err
			}
			should, err := filters.FilterRepository(rpb, r.Filter)
			if err != nil {
				return err
			}
			if should {
				filteredRepos = append(filteredRepos, rpb)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}

		t, err := s.rp.CreatePage(filteredRepos)
		if err != nil {
			return nil, err
		}

		pagesize := getPageSize(int(r.PageSize))
		pg, idx, err = s.rp.GetPage(t, pagesize)
		if err != nil {
			return nil, err
		}
		if idx > 0 {
			nextToken, err = makeFirstPageToken(t, idx)
			if err != nil {
				return nil, err
			}
		}
	}

	resp := drghs_v1.ListRepositoriesResponse{
		Repositories:  pg,
		NextPageToken: nextToken,
	}
	return &resp, err
}

func (s *issueServiceV1) ListIssues(ctx context.Context, r *drghs_v1.ListIssuesRequest) (*drghs_v1.ListIssuesResponse, error) {
	resp := drghs_v1.ListIssuesResponse{}

	filter := fmt.Sprintf("issue.is_pr == %v && issue.closed == %v ", r.PullRequest, r.Closed)
	if r.Filter != "" {
		filter = r.Filter
	}

	err := s.corpus.GitHub().ForeachRepo(func(repo *maintner.GitHubRepo) error {
		repoID := getRepoPath(repo)
		if repoID != r.Parent {
			// Not our repository... ignore
			fmt.Printf("Repo: %v not equal to parent: %v\n", repoID, r.Parent)
			return nil
		}

		return repo.ForeachIssue(func(issue *maintner.GitHubIssue) error {
			iss, err := makeIssuePB(issue, r.Comments, r.Reviews)
			if err != nil {
				return err
			}

			should, err := filters.FilterIssue(iss, filter)
			if err != nil {
				return err
			}
			if should {
				// Add
				resp.Issues = append(resp.Issues, iss)
			}
			return nil
		})
	})

	return &resp, err
}

func (s *issueServiceV1) GetIssue(ctx context.Context, r *drghs_v1.GetIssueRequest) (*drghs_v1.GetIssueResponse, error) {
	resp := &drghs_v1.GetIssueResponse{}

	err := s.corpus.GitHub().ForeachRepo(func(repo *maintner.GitHubRepo) error {
		repoID := getRepoPath(repo)
		if !strings.HasPrefix(r.Name, repoID) {
			// Not our repository... ignore
			fmt.Printf("Repo: %v not equal to parent: %v\n", repoID, r.Name)
			return nil
		}

		return repo.ForeachIssue(func(issue *maintner.GitHubIssue) error {
			if r.Name == getIssueName(repo, issue) {
				re, err := makeIssuePB(issue, r.Comments, r.Reviews)
				if err != nil {
					return err
				}
				resp.Issue = re
			}
			return nil
		})
	})

	return resp, err
}

// Check is for health checking.
func (s *issueServiceV1) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (s *issueServiceV1) Watch(req *healthpb.HealthCheckRequest, ws healthpb.Health_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "health check via Watch not implemented")
}

func getRepoPath(ta *maintner.GitHubRepo) string {
	return fmt.Sprintf("%v/%v", ta.ID().Owner, ta.ID().Repo)
}

func getIssueName(ta *maintner.GitHubRepo, iss *maintner.GitHubIssue) string {
	return fmt.Sprintf("%v/%v/issues/%v", ta.ID().Owner, ta.ID().Repo, iss.Number)
}
