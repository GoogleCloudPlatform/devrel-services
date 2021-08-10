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
	"regexp"
	"strconv"
	"strings"
	"time"

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"
	"github.com/golang/protobuf/ptypes"

	"github.com/GoogleCloudPlatform/devrel-services/drghs-worker/maintnerd/api/filters"
	"github.com/GoogleCloudPlatform/devrel-services/drghs-worker/pkg/googlers"

	"golang.org/x/build/maintner"

	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

var _ drghs_v1.IssueServiceServer = &IssueServiceV1{}

const defaultFilter = "true"

var issueNumReg = regexp.MustCompile(`^[\w.-]+\/[\w.-]+\/issues/(\d+)$`)

// IssueServiceV1 is an implementation of the gRPC service drghs_v1.IssueServiceServer
type IssueServiceV1 struct {
	corpus          *maintner.Corpus
	rp              *repoPaginator
	ip              *issuePaginator
	googlerResolver googlers.Resolver
}

// NewIssueServiceV1 returns a service that implements
// drghs_v1.IssueServiceServer
func NewIssueServiceV1(corpus *maintner.Corpus, resolver googlers.Resolver) *IssueServiceV1 {
	return &IssueServiceV1{
		corpus: corpus,
		rp: &repoPaginator{
			set: make(map[time.Time]repoPage),
		},
		ip: &issuePaginator{
			set: make(map[time.Time]issuePage),
		},
	}
}

// ListRepositories lists the set of repositories tracked by this maintner instance
func (s *IssueServiceV1) ListRepositories(ctx context.Context, r *drghs_v1.ListRepositoriesRequest) (*drghs_v1.ListRepositoriesResponse, error) {

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

// ListIssues lists the issues for the repo in the ListIssuesRequest
func (s *IssueServiceV1) ListIssues(ctx context.Context, r *drghs_v1.ListIssuesRequest) (*drghs_v1.ListIssuesResponse, error) {
	var pg []*drghs_v1.Issue
	var idx int
	var err error
	nextToken := ""

	if r.PageToken != "" {
		//Handle pagination
		pageToken, err := decodePageToken(r.PageToken)
		if err != nil {
			return nil, err
		}

		ftime, err := ptypes.Timestamp(pageToken.FirstRequestTimeUsec)
		if err != nil {

			return nil, err
		}

		pagesize := getPageSize(int(r.PageSize))

		pg, idx, err = s.ip.GetPage(ftime, pagesize)
		if err != nil {
			return nil, err
		}
		nextToken, err = makeNextPageToken(pageToken, idx)
		if err != nil {
			return nil, err
		}
	} else {
		issues := make([]*drghs_v1.Issue, 0)

		err := s.corpus.GitHub().ForeachRepo(func(repo *maintner.GitHubRepo) error {
			repoID := getRepoPath(repo)
			if repoID != r.Parent {
				// Not our repository... ignore
				fmt.Printf("Repo: %v not equal to parent: %v\n", repoID, r.Parent)
				return nil
			}

			return repo.ForeachIssue(func(issue *maintner.GitHubIssue) error {
				i, err := handleIssue(issue, repo.ID(), r, issues)
				issues = i
				return err
			})
		})
		if err != nil {
			return nil, err
		}

		t, err := s.ip.CreatePage(issues)
		if err != nil {
			return nil, err
		}

		pagesize := getPageSize(int(r.PageSize))

		pg, idx, err = s.ip.GetPage(t, pagesize)
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

	return &drghs_v1.ListIssuesResponse{
		Issues:        pg,
		NextPageToken: nextToken,
	}, err
}

// GetIssue returns the issue specified in the GetIssueRequest
func (s *IssueServiceV1) GetIssue(ctx context.Context, r *drghs_v1.GetIssueRequest) (*drghs_v1.GetIssueResponse, error) {
	resp := &drghs_v1.GetIssueResponse{}
	issueID := int32(getIssueID(r.Name))
	var issueResp *drghs_v1.Issue = nil
	err := s.corpus.GitHub().ForeachRepo(func(repo *maintner.GitHubRepo) error {
		repoID := getRepoPath(repo)
		if !strings.HasPrefix(r.Name, repoID) {
			// Not our repository... ignore
			fmt.Printf("Repo: %v not equal to parent: %v\n", repoID, r.Name)
			return nil
		}

		issue := repo.GetIssue(issueID)
		if issue == nil || issue.NotExist {
			return nil
		}

		re, err := makeIssuePB(issue, repo.ID(), r.Comments, r.Reviews, r.FieldMask)
		if err != nil {
			return err
		}
		issueResp = re
		return nil
	})

	if err != nil {
		return resp, err
	}

	if issueResp == nil {
		return nil, status.Errorf(codes.NotFound, "issue: %v not found", r.Name)
	}

	resp.Issue = issueResp

	return resp, err
}

// Check is for health checking.
func (s *IssueServiceV1) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

// Watch is used for Health Checking, but is not supported.
func (s *IssueServiceV1) Watch(req *healthpb.HealthCheckRequest, ws healthpb.Health_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "health check via Watch not implemented")
}

func shouldAddIssue(issue *maintner.GitHubIssue, r *drghs_v1.ListIssuesRequest) (bool, error) {
	if issue.NotExist {
		return false, nil
	}

	switch x := r.PullRequestNullable.(type) {
	case *drghs_v1.ListIssuesRequest_PullRequest:
		if issue.PullRequest != x.PullRequest {
			return false, nil
		}
	case nil:
		// Do nothing
	default:
		// Do nothing
	}

	switch x := r.ClosedNullable.(type) {
	case *drghs_v1.ListIssuesRequest_Closed:
		if issue.Closed != x.Closed {
			return false, nil
		}
	case nil:
		// Do nothing
	default:
		// Do nothing
	}

	return true, nil
}

func getRepoPath(ta *maintner.GitHubRepo) string {
	return fmt.Sprintf("%v/%v", ta.ID().Owner, ta.ID().Repo)
}

func getIssueName(ta *maintner.GitHubRepo, iss *maintner.GitHubIssue) string {
	return fmt.Sprintf("%v/%v/issues/%v", ta.ID().Owner, ta.ID().Repo, iss.Number)
}

func handleIssue(issue *maintner.GitHubIssue, rid maintner.GitHubRepoID, r *drghs_v1.ListIssuesRequest, issues []*drghs_v1.Issue) ([]*drghs_v1.Issue, error) {
	if issue.NotExist {
		return issues, nil
	}

	issClean, err := makeIssuePB(issue, rid, r.Comments, r.Reviews, nil)

	should, err := filters.FilterIssue(issClean, r)
	if err != nil {
		return issues, err
	}
	if should {
		// Add
		iss, err := makeIssuePB(issue, rid, r.Comments, r.Reviews, r.FieldMask)
		if err != nil {
			return issues, err
		}
		return append(issues, iss), nil
	}
	return issues, nil
}

func getIssueID(issueName string) int {
	sm := issueNumReg.FindAllStringSubmatch(issueName, -1)
	if sm == nil {
		return -1
	}
	id := sm[0][1]
	i, err := strconv.Atoi(id)
	if err != nil {
		return -1
	}
	return i
}
