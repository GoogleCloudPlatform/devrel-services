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
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"

	"github.com/GoogleCloudPlatform/devrel-services/drghs-worker/maintnerd/api/v1beta1"
	"github.com/GoogleCloudPlatform/devrel-services/drghs-worker/pkg/googlers"

	"cloud.google.com/go/errorreporting"
	"golang.org/x/build/maintner"
	"golang.org/x/build/maintner/maintnerd/gcslog"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

var (
	listen    = flag.String("listen", "0.0.0.0:6343", "listen address")
	verbose   = flag.Bool("verbose", false, "enable verbose debug output")
	bucket    = flag.String("bucket", "cdpe-maintner", "Google Cloud Storage bucket to use for log storage")
	token     = flag.String("token", "", "Token to Access GitHub with")
	projectID = flag.String("gcp-project", "", "The GCP Project this is using")
	owner     = flag.String("owner", "", "The owner of the GitHub repository")
	repo      = flag.String("repo", "", "The repository to track")
)

var (
	corpus          = &maintner.Corpus{}
	googlerResolver googlers.GooglersResolver
	errorClient     *errorreporting.Client
)

func main() {
	// Set log to Stdout. Default for log is Stderr
	log.SetOutput(os.Stdout)
	flag.Parse()

	ctx := context.Background()

	if *projectID == "" {
		log.Fatal("must provide --gcp-project")
	}

	var err error
	errorClient, err = errorreporting.NewClient(ctx, *projectID, errorreporting.Config{
		ServiceName: "devrel-github-services",
		OnError: func(err error) {
			log.Printf("Could not report error: %v", err)
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer errorClient.Close()

	if *token == "" {
		err := fmt.Errorf("must provide --token")
		logAndPrintError(err)
		log.Fatal(err)
	}

	if *owner == "" {
		err := fmt.Errorf("must provide --owner")
		logAndPrintError(err)
		log.Fatal(err)
	}

	if *repo == "" {
		err := fmt.Errorf("must provide --repo")
		logAndPrintError(err)
		log.Fatal(err)
	}

	const qps = 1
	limit := rate.Every(time.Second / qps)
	corpus.SetGitHubLimiter(rate.NewLimiter(limit, qps))

	if *bucket == "" {
		err := fmt.Errorf("must provide --bucket")
		logAndPrintError(err)
		log.Fatal(err)
	}

	gl, err := gcslog.NewGCSLog(ctx, fmt.Sprintf("%v/%v/%v/", *bucket, *owner, *repo))
	if err != nil {
		err := fmt.Errorf("NewGCSLog: %v", err)
		logAndPrintError(err)
		log.Fatal(err)
	}

	dataDir := filepath.Join("/tmp", "maintnr")
	log.Printf("dataDir: %v", dataDir)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatal(err)
	}
	log.Printf("Storing data in implicit directory %s", dataDir)

	corpus.EnableLeaderMode(gl, dataDir)

	if err := corpus.Initialize(ctx, gl); err != nil {
		err := fmt.Errorf("Initalize: %v", err)
		logAndPrintError(err)
		log.Fatal(err)
	}

	tkn := strings.TrimSpace(*token)
	corpus.TrackGitHub(*owner, *repo, tkn)

	googlerResolver = googlers.NewGooglersStatic()

	group, ctx := errgroup.WithContext(context.Background())
	group.Go(
		func() error {
			// In the golang.org/x/build/maintner syncloop the update loops
			// are done every 30 seconds.
			// We will go for a less agressive schedule and only sync once every
			// 10 minutes.
			ticker := time.NewTicker(10 * time.Minute)
			for t := range ticker.C {
				log.Printf("Corpus.SyncLoop at %v", t)
				// Lock it for writes
				// Sync
				if err := corpus.Sync(ctx); err != nil {
					logAndPrintError(err)
					log.Printf("Error during corpus sync %v", err)
				}
				// Unlock
			}
			return nil
		})

	group.Go(func() error {
		// Add gRPC service for v1beta1
		grpcServer := grpc.NewServer()
		s := v1beta1.NewIssueServiceV1(corpus, googlerResolver)
		drghs_v1.RegisterIssueServiceServer(grpcServer, s)
		healthpb.RegisterHealthServer(grpcServer, s)

		lis, err := net.Listen("tcp", *listen)
		if err != nil {
			log.Fatalf("failed to listen %v", err)
		}

		log.Printf("gRPC server listening on: %s", *listen)
		return grpcServer.Serve(lis)
	})
	err = group.Wait()
	log.Fatal(err)
}

func logAndPrintError(err error) {
	errorClient.Report(errorreporting.Entry{
		Error: err,
	})
	log.Print(err)
}
