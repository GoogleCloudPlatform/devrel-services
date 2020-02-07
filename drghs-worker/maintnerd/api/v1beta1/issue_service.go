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

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"

	"github.com/GoogleCloudPlatform/devrel-services/drghs-worker/pkg/googlers"

	"golang.org/x/build/maintner"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"

	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

var _ drghs_v1.IssueServiceServer = &IssueServiceV1{}

const defaultFilter = "true"

// IssueServiceV1 is an implementation of the gRPC service drghs_v1.IssueServiceServer
type IssueServiceV1 struct {
	corpus          *maintner.Corpus
	googlerResolver googlers.Resolver
}

// NewIssueServiceV1 returns a service that implements
// drghs_v1.IssueServiceServer
func NewIssueServiceV1(corpus *maintner.Corpus, resolver googlers.Resolver) *IssueServiceV1 {
	return &IssueServiceV1{
		corpus:          corpus,
		googlerResolver: resolver,
	}
}

// ListRepositories lists the set of repositories tracked by this maintner instance
func (s *IssueServiceV1) ListRepositories(ctx context.Context, r *drghs_v1.ListRepositoriesRequest) (*drghs_v1.ListRepositoriesResponse, error) {
	resp := drghs_v1.ListRepositoriesResponse{}
	err := s.corpus.GitHub().ForeachRepo(func(repo *maintner.GitHubRepo) error {
		should, err := shouldAddRepository(repo.ID(), r.Filter)
		if err != nil {
			return err
		}
		if should {
			rpb, err := makeRepoPB(repo)
			if err != nil {
				return err
			}
			resp.Repositories = append(resp.Repositories, rpb)
		}
		return nil
	})
	return &resp, err
}

// ListIssues lists the issues for the repo in the ListIssuesRequest
func (s *IssueServiceV1) ListIssues(ctx context.Context, r *drghs_v1.ListIssuesRequest) (*drghs_v1.ListIssuesResponse, error) {
	resp := drghs_v1.ListIssuesResponse{}

	err := s.corpus.GitHub().ForeachRepo(func(repo *maintner.GitHubRepo) error {
		repoID := getRepoPath(repo)
		if repoID != r.Parent {
			// Not our repository... ignore
			fmt.Printf("Repo: %v not equal to parent: %v\n", repoID, r.Parent)
			return nil
		}

		return repo.ForeachIssue(func(issue *maintner.GitHubIssue) error {
			should, err := shouldAddIssue(issue, r)
			if err != nil {
				return err
			}
			if should {
				// Add
				iss, err := makeIssuePB(issue, repo.ID(), r.Comments, r.Reviews)
				if err != nil {
					return err
				}
				resp.Issues = append(resp.Issues, iss)
			}
			return nil
		})
	})

	return &resp, err
}

// GetIssue returns the issue specified in the GetIssueRequest
func (s *IssueServiceV1) GetIssue(ctx context.Context, r *drghs_v1.GetIssueRequest) (*drghs_v1.GetIssueResponse, error) {
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
				re, err := makeIssuePB(issue, repo.ID(), r.Comments, r.Reviews)
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

// TODO(orthros) This should default to using *maintner.GitHubRepo, but
// due to how maintner stores values, this is impossible to mock for tests
// If other traits of a repository need to be added (labels, milestones etc)
// in order to support filtering, this funciton signature must be changed
// e.g. `owner == 'foo' && labels.size() > 10` to find repositories whose
// owner is 'foo' and number of labels is > 10. This would need to be expanded
func shouldAddRepository(repoID maintner.GitHubRepoID, filter string) (bool, error) {
	if filter == "" {
		filter = defaultFilter
	}

	env, err := cel.NewEnv(
		cel.Declarations(
			decls.NewIdent("repo", decls.String, nil),
			decls.NewIdent("owner", decls.String, nil)))

	parsed, issues := env.Parse(filter)
	if issues != nil && issues.Err() != nil {
		return false, issues.Err()
	}
	checked, issues := env.Check(parsed)
	if issues != nil && issues.Err() != nil {
		return false, issues.Err()
	}
	prg, err := env.Program(checked)
	if err != nil {
		return false, err
	}

	// The `out` var contains the output of a successful evaluation.
	// The `details' var would contain intermediate evalaution state if enabled as
	// a cel.ProgramOption. This can be useful for visualizing how the `out` value
	// was arrive at.
	out, _, err := prg.Eval(map[string]interface{}{
		"repo":  repoID.Repo,
		"owner": repoID.Owner,
	})

	return out == types.True, nil
}
