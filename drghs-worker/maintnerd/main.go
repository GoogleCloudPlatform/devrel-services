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
	drghs_v1 "devrel/cloud/devrel-github-service/drghs/v1"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"devrel/cloud/devrel-github-service/drghs-worker/maintnerd/api/v1beta1"
	"devrel/cloud/devrel-github-service/drghs-worker/maintnerd/internal/apiroutes"
	"devrel/cloud/devrel-github-service/drghs-worker/pkg/googlers"

	"cloud.google.com/go/errorreporting"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"golang.org/x/build/maintner"
	"golang.org/x/build/maintner/maintnerd/gcslog"
	"golang.org/x/time/rate"
	grpc "google.golang.org/grpc"
)

var (
	listen    = flag.String("listen", ":6343", "listen address")
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

	go func() {
		// In the golang.org/x/build/maintner syncloop the update loops
		// are done every 30 seconds.
		// We will go for a less agressive schedule and only sync once every
		// 1 minutes.
		ticker := time.NewTicker(10 * time.Minute)
		for t := range ticker.C {
			log.Printf("Corpus.SyncLoop at %d", t)
			// Lock it for writes
			// Sync
			if err := corpus.Sync(ctx); err != nil {
				logAndPrintError(err)
				log.Printf("Error during corpus sync %v", err)
			}
			// Unlock
		}
	}()

	// Add gRPC service for v1beta1
	grpcServer := grpc.NewServer()
	drghs_v1.RegisterIssueServiceServer(grpcServer, v1beta1.NewIssueServiceV1(corpus, googlerResolver))

	// Send everything through Mux
	r := mux.NewRouter()
	apiSR := r.PathPrefix("/api").Subrouter()

	// Keep a handle on our Api Routers
	apis := registerApis(apiSR)
	log.Printf("Registered: %v Api Routes", len(apis))

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
			return
		}
	})

	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})

	// Add middleware support
	n := negroni.New()
	l := negroni.NewLogger()
	n.Use(l)
	n.Use(negroni.NewRecovery())
	n.UseHandler(r)

	log.Fatal(http.ListenAndServe(*listen, n))
}

func logAndPrintError(err error) {
	errorClient.Report(errorreporting.Entry{
		Error: err,
	})
	log.Print(err)
}

func registerApis(r *mux.Router) []apiroutes.ApiRoute {
	apis := make([]apiroutes.ApiRoute, 0)
	vzSR := r.PathPrefix("/v0").Subrouter()

	api, err := apiroutes.NewV0Api(corpus, googlerResolver, vzSR)
	if err != nil {
		logAndPrintError(err)
		log.Fatalf("Error registering v0 Api Routes %v", err)
	}
	api.Routes()
	apis = append(apis, api)

	vOSr := r.PathPrefix("/v1").Subrouter()
	api, err = apiroutes.NewV1Api(corpus, googlerResolver, vOSr)
	if err != nil {
		logAndPrintError(err)
		log.Fatalf("Error registering v1 Api Routes %v", err)
	}

	api.Routes()
	apis = append(apis, api)

	return apis
}
