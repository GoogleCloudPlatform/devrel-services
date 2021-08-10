// Copyright 2020 Google LLC
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
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"
	"github.com/GoogleCloudPlatform/devrel-services/leif"
	"github.com/GoogleCloudPlatform/devrel-services/leif/githubservices"
	"github.com/GoogleCloudPlatform/devrel-services/leif/leifd/leifapi"
	"github.com/GoogleCloudPlatform/devrel-services/repos"

	"github.com/gregjones/httpcache"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

const gitHubEnvVar = "GITHUB_TOKEN"

// if bucket is specified, repos file is to be found in that bucket, elsewise, it's local
var (
	bucket       = flag.String("bucket", "", "Google Cloud Storage bucket to use for settings storage")
	listen       = flag.String("listen", "0.0.0.0:3009", "gRPC listen address")
	reposFile    = flag.String("repos", "", "File that contains the list of repositories")
	syncInterval = flag.Int("sync", 10, "Update interval in minutes")
	verbose      = flag.Bool("verbose", false, "Verbose logs")
)

var log *logrus.Logger
var repoList repos.RepoList
var ghClient githubservices.Client

func init() {
	log = logrus.New()
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}
	leif.FormatLog(log.Formatter)

	log.Out = os.Stdout
}

func parseFlags() {
	flag.Parse()

	if *verbose == true {
		log.Level = logrus.DebugLevel
		leif.VerboseLog()
	}

	if *reposFile == "" {
		err := fmt.Errorf("must provide --repos")
		log.Fatal(err)
	}
}

func initGHClient(ctx context.Context) {
	if os.Getenv(gitHubEnvVar) == "" {
		log.Fatalf("env var %v is empty", gitHubEnvVar)
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: strings.Trim(os.Getenv(gitHubEnvVar), "\n")},
	)
	hc := oauth2.NewClient(ctx, src)

	cachedTransport := httpcache.Transport{Transport: hc.Transport, Cache: httpcache.NewMemoryCache()}

	ghClient = githubservices.NewClient(cachedTransport.Client(), nil, nil)
}

func main() {
	log = logrus.New()
	parseFlags()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		sig := <-signalCh
		fmt.Printf("termination signal received: %s", sig)
		cancel()
	}()

	initGHClient(ctx)

	if *bucket != "" {
		repoList = repos.NewBucketRepo(*bucket, *reposFile)
	} else {
		repoList = leif.NewDiskRepo(*reposFile)
	}

	_, err := repoList.UpdateTrackedRepos(ctx)
	if err != nil {
		log.Fatal(err)
		return
	}

	// init corpus with repos from list
	corpus := &leif.Corpus{}
	for _, r := range repoList.GetTrackedRepos() {
		if r.IsTrackingIssues {
			err = corpus.TrackRepo(ctx, r.Owner, r.Name, &ghClient)
			if err != nil {
				log.Warningf("Could not track repository %s/%s:%v", r.Owner, r.Name, err)
			}
		}
	}

	err = corpus.Initialize(ctx, &ghClient)
	if err != nil {
		log.Fatalf("Could not initialize corpus: %v", err)
	}

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return corpus.SyncLoop(ctx, *syncInterval, &ghClient)
	})

	group.Go(func() error {
		lis, err := net.Listen("tcp", *listen)
		if err != nil {
			log.Fatalf("failed to listen %v", err)
		}

		grpcServer := grpc.NewServer()
		drghs_v1.RegisterSLOServiceServer(grpcServer, leifapi.NewSLOServiceServer(corpus))

		go func() {
			select {
			case <-ctx.Done():
				grpcServer.GracefulStop()
			}
		}()

		log.Printf("gRPC server listening on: %s", *listen)
		return grpcServer.Serve(lis)
	})

	err = group.Wait()
	log.Fatal(err)
}
