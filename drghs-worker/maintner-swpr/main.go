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
	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/oauth2"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
)

// Log
var log *logrus.Logger

// Flags
var (
	flRtrAddr *string
)

// Constants
const (
	GitHubEnvVar  = "GITHUB_TOKEN"
	SecondsPerDay = 86400
)

// Uses
var (
	rNameRegex = regexp.MustCompile(`^([\w-]+)\/([\w-]+)$`)
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

	// Setup drghs client
	var drghsc drghs_v1.IssueServiceClient

	conn, err := grpc.Dial(*flRtrAddr, grpc.WithInsecure())
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

	// Setup GraphQL Client
	//

	// Queries per second as we retrieve 100 issues at a time from GitHub
	qps := int((nipr / 100) / SecondsPerDay)
	limit := rate.Every(time.Second / time.Duration(qps))
	limiter := rate.NewLimiter(limit, qps)
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv(GitHubEnvVar)},
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
		ghIssues, err := getGitHubIssuesForRepo(ctx, gqlc, repo)
		if err != nil {
			log.Fatal(err)
		}

		log.Debugf("repo: %v number of issues: %v\n", repo.Name, len(ghIssues))

		ghIssuesByID := make(map[int32]struct{})
		for _, iss := range ghIssues {
			ghIssuesByID[iss.Number] = struct{}{}
		}

		mtrIssues, err := getMaintnerIssuesForRepo(ctx, nil, repo)

		log.Debugf("repo: %v number of maintner issues %v\n", repo.Name, len(mtrIssues))

		tmbIssues := make([]int32, 0)
		for _, mtri := range mtrIssues {
			if _, ok := ghIssuesByID[mtri.IssueId]; !ok {
				tmbIssues = append(tmbIssues, mtri.IssueId)
			}
		}

		log.Debugf("repo: %v number of tombstoned issues %v\n", repo.Name, len(tmbIssues))

		if len(tmbIssues) > 0 {
			log.Infof("repo: %v tombstoning: %v issues\n", repo.Name, len(tmbIssues))

			err = flagIssuesTombstoned(ctx, repo, tmbIssues)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

type issue struct {
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

func getMaintnerIssuesForRepo(ctx context.Context, c drghs_v1.IssueServiceClient, repo *drghs_v1.Repository) ([]*drghs_v1.Issue, error) {
	ret := make([]*drghs_v1.Issue, 0)
	npt := ""
	for {
		rep, err := c.ListIssues(ctx, &drghs_v1.ListIssuesRequest{
			Parent:    repo.Name,
			PageToken: npt,
			PageSize:  500,
		})
		if err != nil {
			return nil, err
		}
		ret = append(ret, rep.Issues...)
		if rep.NextPageToken == "" {
			break
		}
		npt = rep.NextPageToken
	}

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

func flagIssuesTombstoned(ctx context.Context, repo *drghs_v1.Repository, issueIds []int32) error {
	tr := repoToTrackedRepo(repo)
	if tr == nil {
		return fmt.Errorf("Bad repository: %v", repo)
	}
	maddr, err := serviceName(*tr)
	if err != nil {
		return err
	}

	conn, err := grpc.Dial(maddr, grpc.WithInsecure())
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
		log.Warn("expected to tombstone %v, actually tombstoned: %v", len(issueIds), resp.TombstonedCount)
	}

	return err
}

func getGitHubIssuesForRepo(ctx context.Context, c *githubv4.Client, repo *drghs_v1.Repository) ([]issue, error) {
	parts := strings.Split(repo.GetName(), "/")

	var q ghIssuesQuery
	variables := map[string]interface{}{
		"repositoryOwner": githubv4.String(parts[0]),
		"repositoryName":  githubv4.String(parts[1]),
		"cursor":          (*githubv4.String)(nil), // Null after argument to get first page.
	}
	// Get issues from all pages.
	var allIssues []issue
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
	}

	return allIssues, nil
}

func serviceName(t repos.TrackedRepository) (string, error) {
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

type limitTransport struct {
	limiter *rate.Limiter
	base    http.RoundTripper
}

func (t limitTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	limiter := t.limiter
	if limiter != nil {
		if err := limiter.Wait(r.Context()); err != nil {
			return nil, err
		}
	}
	return t.base.RoundTrip(r)
}
