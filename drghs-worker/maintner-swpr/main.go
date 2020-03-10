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
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"time"

	maintner_internal "github.com/GoogleCloudPlatform/devrel-services/drghs-worker/internal"
	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"
	"github.com/GoogleCloudPlatform/devrel-services/repos"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/oauth2"
	"golang.org/x/time/rate"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Log
var log *logrus.Logger

// Flags
var (
	flRtrAddr *string
)

// Constants
const (
	GitHubEnvVar        = "GITHUB_TOKEN"
	SecondsPerDay       = 86400.0
	ServicePort         = "80"
	ServicePortInternal = "8080"
)

// Uses
var (
	rNameRegex = regexp.MustCompile(`^([.:\w-]+)\/([.:\w-]+)$`)
)

func init() {
	log = logrus.New()
	if !terminal.IsTerminal(int(os.Stdout.Fd())) {
		log.Formatter = &logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "severity",
				logrus.FieldKeyMsg:   "message",
			},
			TimestampFormat: time.RFC3339Nano,
		}
	}
	log.SetLevel(logrus.TraceLevel)
	log.Out = os.Stdout

	flRtrAddr = flag.String("rtr-address", "", "specifies the address of the router to dial")
}

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		sig := <-signalCh
		log.Printf("termination signal received: %s", sig)
		cancel()
	}()

	if *flRtrAddr == "" {
		log.Fatal("--rtr-address is empty")
	}

	if os.Getenv(GitHubEnvVar) == "" {
		log.Fatalf("env var %v is empty", GitHubEnvVar)
	}

	log.Debugf("Setting up and dialing to maintner-rtr: %v", *flRtrAddr)
	// Setup drghs client
	var drghsc drghs_v1.IssueServiceClient

	conn, err := grpc.Dial(
		*flRtrAddr,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(buildRetryInterceptor()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	drghsc = drghs_v1.NewIssueServiceClient(conn)

	repos, err := getTrackedRepositories(ctx, drghsc)
	if err != nil {
		log.Fatal(err)
	}

	var nipr int32
	for _, repo := range repos {
		nipr = nipr + repo.PullRequestCount
		nipr = nipr + repo.IssueCount
	}

	log.Debugf("have %v repositories to query with a total of %v Issues and PRs", len(repos), nipr)

	// Setup GraphQL Client
	//

	// Queries per second as we retrieve 100 issues at a time from GitHub
	limiter := buildLimiter(nipr)
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: strings.Trim(os.Getenv(GitHubEnvVar), "\n")},
	)
	hc := oauth2.NewClient(ctx, src)
	transport := limitTransport{limiter, hc.Transport}
	httpClient := &http.Client{
		Transport: transport,
	}
	gqlc := githubv4.NewClient(httpClient)

	// For each repo, get all the GitHub Issues for the Repo
	// Then get all the mainter issues for the repo
	// Finally, compare the two, find the ones in maintner that
	// are not in GitHub && Flag them as NotExist
	for _, repo := range repos {
		log.Debugf("processing repo: %v", repo.String())

		tr := repoToTrackedRepo(repo)

		ghIssues, err := getGitHubIssuesForRepo(ctx, gqlc, repo)
		if err != nil {
			log.Fatal(err)
		}
		log.Debugf("repo: %v number of issues: %v\n", repo.String(), len(ghIssues))

		ghPrs, err := getGitHubPullRequestsForRepo(ctx, gqlc, repo)
		if err != nil {
			log.Fatal(err)
		}
		log.Debugf("repo: %v number of pull requests: %v\n", repo.String(), len(ghPrs))

		ghIssuesByID := make(map[int32]struct{})
		for _, iss := range ghIssues {
			ghIssuesByID[iss.Number] = struct{}{}
		}
		for _, pr := range ghPrs {
			ghIssuesByID[pr.Number] = struct{}{}
		}

		mtrIssues, err := getMaintnerIssuesForRepo(ctx, tr, repo)

		log.Debugf("repo: %v number of maintner issues %v\n", repo.Name, len(mtrIssues))

		tmbIssues := make([]int32, 0)
		for _, mtri := range mtrIssues {
			if _, ok := ghIssuesByID[mtri.IssueId]; !ok {
				tmbIssues = append(tmbIssues, mtri.IssueId)
			}
		}

		log.Debugf("repo: %v number of tombstoned issues %v\nissues to tombstone: %v\n", repo.Name, len(tmbIssues), tmbIssues)

		if len(tmbIssues) > 0 {
			log.Infof("repo: %v tombstoning: %v issues\n", repo.Name, len(tmbIssues))

			err = flagIssuesTombstoned(ctx, tr, repo, tmbIssues)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	log.Infof("finished!")
}

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

func getMaintnerIssuesForRepo(ctx context.Context, tr *repos.TrackedRepository, repo *drghs_v1.Repository) ([]*drghs_v1.Issue, error) {
	log.Debugf("getting Issues from maintner for repo: %v", repo.Name)

	maddr, err := serviceName(tr)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(
		maddr+":"+ServicePort,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(buildRetryInterceptor()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	c := drghs_v1.NewIssueServiceClient(conn)

	ret := make([]*drghs_v1.Issue, 0)
	npt := ""
	for {
		rep, err := c.ListIssues(ctx, &drghs_v1.ListIssuesRequest{
			Parent:    repo.Name,
			PageToken: npt,
			PageSize:  500,
			FieldMask: &field_mask.FieldMask{
				Paths: []string{
					"issue_id",
				},
			},
		})
		if err != nil {
			return nil, err
		}
		ret = append(ret, rep.Issues...)
		if rep.NextPageToken == "" {
			break
		}
		npt = rep.NextPageToken
		log.Debugf("getting Issues from maintner for repo: %v. requesting next page. current count: %v", repo.Name, len(ret))
	}
	log.Debugf("finished getting Issues from mainter for repo: %v. returning %v Issues", repo.Name, len(ret))
	return ret, nil
}

func getTrackedRepositories(ctx context.Context, c drghs_v1.IssueServiceClient) ([]*drghs_v1.Repository, error) {
	ret := make([]*drghs_v1.Repository, 0)
	npt := ""
	for {
		rep, err := c.ListRepositories(ctx, &drghs_v1.ListRepositoriesRequest{
			PageToken: npt,
			PageSize:  500,
		})
		if err != nil {
			return nil, err
		}
		ret = append(ret, rep.Repositories...)
		if rep.NextPageToken == "" {
			break
		}
		npt = rep.NextPageToken
	}

	return ret, nil
}

func flagIssuesTombstoned(ctx context.Context, tr *repos.TrackedRepository, repo *drghs_v1.Repository, issueIds []int32) error {
	maddr, err := serviceName(tr)
	if err != nil {
		return err
	}

	conn, err := grpc.Dial(
		maddr+":"+ServicePortInternal,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(buildRetryInterceptor()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	c := maintner_internal.NewInternalIssueServiceClient(conn)

	req := &maintner_internal.TombstoneIssuesRequest{
		Parent:       repo.Name,
		IssueNumbers: issueIds,
	}
	resp, err := c.TombstoneIssues(ctx, req)
	if err != nil {
		return err
	}

	log.Infof("tombstoned: %v issues", resp.TombstonedCount)
	if int(resp.TombstonedCount) != len(issueIds) {
		log.Warnf("expected to tombstone %v, actually tombstoned: %v", len(issueIds), resp.TombstonedCount)
	}

	return err
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

func serviceName(t *repos.TrackedRepository) (string, error) {
	return strings.ToLower(fmt.Sprintf("mtr-s-%s", t.RepoSha())), nil
}

func repoToTrackedRepo(r *drghs_v1.Repository) *repos.TrackedRepository {
	var ta *repos.TrackedRepository
	mtches := rNameRegex.FindAllStringSubmatch(r.Name, -1)
	if mtches != nil {
		log.Tracef("Got a match!")
		// This match will be of form:
		// [["/v1/owners/foo/repositories/bar1/issues" "foo" "bar1"]]
		// Therefore slice the array
		ta = &repos.TrackedRepository{
			Owner: mtches[0][1],
			Name:  mtches[0][2],
		}
	}
	return ta
}

func buildLimiter(nipr int32) *rate.Limiter {
	var qps = (float64(nipr) / 100.0) / SecondsPerDay
	log.Debugf("have a qps of %v", qps)

	dur := time.Duration(float64(time.Second) * (1.0 / qps))
	limit := rate.Every(dur)
	limiter := rate.NewLimiter(limit, int(qps)+1)

	return limiter
}

type limitTransport struct {
	limiter *rate.Limiter
	base    http.RoundTripper
}

func (t limitTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	limiter := t.limiter
	if limiter != nil {
		log.Debug("in limitTransport. Round trip Waiting for limiter")
		if err := limiter.Wait(r.Context()); err != nil {
			return nil, err
		}
	}
	return t.base.RoundTrip(r)
}

func buildRetryInterceptor() grpc.UnaryClientInterceptor {
	opts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(500 * time.Millisecond)),
		grpc_retry.WithCodes(codes.NotFound, codes.Aborted),
		grpc_retry.WithMax(5),
	}
	return grpc_retry.UnaryClientInterceptor(opts...)
}
