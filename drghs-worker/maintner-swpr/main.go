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

	"cloud.google.com/go/errorreporting"
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
	flRtrAddr   *string
	flProjectID *string
	flBucket    = flag.String("settings-bucket", "", "bucket to get repo list")
	flFile      = flag.String("repos-file", "", "file in bucket to read repos from")
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
	flProjectID = flag.String("project-id", "", "the GCP Project ID this is running in.")
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

	if *flBucket == "" {
		log.Fatal("error: must specify --settings-bucket")
	}
	if *flFile == "" {
		log.Fatal("error: must specify --repos-file")
	}
	if os.Getenv(GitHubEnvVar) == "" {
		log.Fatalf("env var %v is empty", GitHubEnvVar)
	}

	if *flProjectID == "" {
		log.Fatal("--project-id is empty")
	}

	var errorClient, err = errorreporting.NewClient(ctx, *flProjectID, errorreporting.Config{
		ServiceName: "maintner-sweeper",
		OnError: func(err error) {
			log.Printf("Could not report error: %v", err)
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer errorClient.Close()

	rlist := repos.NewBucketRepo(*flBucket, *flFile)
	_, err = rlist.UpdateTrackedRepos(ctx)
	if err != nil {
		log.Fatalf("got error updating repos: %v", err)
	}

	errs := make([]error, 0)
	repos := rlist.GetTrackedRepos()
	var nipr int32
	for _, repo := range repos {
		if !repo.IsTrackingIssues {
			log.Infof("skipping repo: %v as it is not tracking issues", repo.String())
		}
		rnipr, err := getIssueAndPRData(ctx, &repo)
		if err != nil {
			errorClient.Report(errorreporting.Entry{
				Error: err,
			})
			continue
		}
		nipr = nipr + rnipr
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
		err := processRepo(ctx, &repo, gqlc)
		if err != nil {
			// Append and report
			errs = append(errs, err)
			errorClient.Report(errorreporting.Entry{
				Error: err,
			})
		}
	}

	log.Infof("finished with %v errors: %v", len(errs), errs)
}

func processRepo(ctx context.Context, tr *repos.TrackedRepository, gqlc *githubv4.Client) error {
	repo := &drghs_v1.Repository{
		Name: fmt.Sprintf("%v/%v", tr.Owner, tr.Name),
	}
	log.Debugf("processing repo: %v", repo.String())

	ghIssues, err := getGitHubIssuesForRepo(ctx, gqlc, repo)
	if err != nil {
		log.Errorf("processing repo %v. hit an error getting GitHub Issues: %v", repo.String(), err)
		return err
	}
	log.Debugf("repo: %v number of issues: %v\n", repo.String(), len(ghIssues))

	ghPrs, err := getGitHubPullRequestsForRepo(ctx, gqlc, repo)
	if err != nil {
		log.Errorf("processing repo %v. hit an error getting GitHub Pull Requests: %v", repo.String(), err)
		return err
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
	if err != nil {
		log.Errorf("processing repo %v. hit an error getting Maintner Issues and PRs: %v", repo.String(), err)
		return err
	}

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
			log.Errorf("processing repo %v. hit an error getting Tombstoning Issues: %v", repo.String(), err)
			return err
		}
	}
	return nil
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
		log.Error(err)
		return nil, err
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
			log.Warnf("got error while listing repositories: %v", err)
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

func getIssueAndPRData(ctx context.Context, tr *repos.TrackedRepository) (int32, error) {
	maddr, err := serviceName(tr)
	if err != nil {
		return 0, err
	}

	conn, err := grpc.Dial(
		maddr+":80",
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(buildRetryInterceptor()),
	)
	if err != nil {
		log.Warnf("got error dialing to repository: %v %v", tr, err)
		return 0, err
	}
	defer conn.Close()

	c := drghs_v1.NewIssueServiceClient(conn)

	repos, err := getTrackedRepositories(ctx, c)
	if err != nil {
		log.Warnf("got error getting tracked repositories for repo: %v, %v", tr, err)
		return 0, err
	}

	var nipr int32
	for _, r := range repos {
		nipr = nipr + r.IssueCount
		nipr = nipr + r.PullRequestCount
	}

	return nipr, nil
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
		return err
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

func buildRetryInterceptor() grpc.UnaryClientInterceptor {
	opts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(500 * time.Millisecond)),
		grpc_retry.WithCodes(codes.NotFound, codes.Aborted, codes.Unavailable, codes.DeadlineExceeded),
		grpc_retry.WithMax(5),
	}
	return grpc_retry.UnaryClientInterceptor(opts...)
}
