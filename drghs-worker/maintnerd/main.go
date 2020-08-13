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

	maintner_internal "github.com/GoogleCloudPlatform/devrel-services/drghs-worker/internal"
	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"

	"github.com/GoogleCloudPlatform/devrel-services/drghs-worker/maintnerd/api/internalapi"
	"github.com/GoogleCloudPlatform/devrel-services/drghs-worker/maintnerd/api/v1beta1"
	"github.com/GoogleCloudPlatform/devrel-services/drghs-worker/pkg/googlers"

	"cloud.google.com/go/errorreporting"
	"golang.org/x/build/maintner"
	"golang.org/x/build/maintner/maintnerd/gcslog"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/instrumentation/grpctrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"cloud.google.com/go/profiler"
)

var (
	listen     = flag.String("listen", "0.0.0.0:6343", "listen address")
	intListen  = flag.String("intListen", "0.0.0:6344", "listen for internal service")
	sloAddress = flag.String("sloServer", "0.0.0:3009", "address of slo service")
	verbose    = flag.Bool("verbose", false, "enable verbose debug output")
	bucket     = flag.String("bucket", "cdpe-maintner", "Google Cloud Storage bucket to use for log storage")
	token      = flag.String("token", "", "Token to Access GitHub with")
	projectID  = flag.String("gcp-project", "", "The GCP Project this is using")
	owner      = flag.String("owner", "", "The owner of the GitHub repository")
	repo       = flag.String("repo", "", "The repository to track")
)

var (
	corpus          = &maintner.Corpus{}
	googlerResolver googlers.Resolver
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

	cfg := profiler.Config{
		Service:        fmt.Sprintf("maintnerd-%v-%v", strings.ToLower(*owner), strings.ToLower(*repo)),
		ServiceVersion: "0.0.3",
		ProjectID:      *projectID,

		// For OpenCensus users:
		// To see Profiler agent spans in APM backend,
		// set EnableOCTelemetry to true
		//EnableOCTelemetry: true,
	}

	if err := profiler.Start(cfg); err != nil {
		log.Fatal(err)
	}

	exporter, err := texporter.NewExporter(texporter.WithProjectID(*projectID))
	if err != nil {
		log.Fatalf("texporter.NewExporter: %v", err)
	}

	config := sdktrace.Config{DefaultSampler: sdktrace.ProbabilitySampler(0.01)}
	tp, err := sdktrace.NewProvider(sdktrace.WithConfig(config), sdktrace.WithSyncer(exporter))
	if err != nil {
		log.Fatal(err)
	}
	global.SetTraceProvider(tp)

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

	googlerResolver = googlers.NewStatic()

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
		grpcServer := grpc.NewServer(
			grpc.UnaryInterceptor(
				grpc_middleware.ChainUnaryServer(
					grpctrace.UnaryServerInterceptor(global.Tracer("maintnerd")),
					unaryInterceptorLog),
			),
		)
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

	group.Go(func() error {
		// Add gRPC service for internal
		grpcServer := grpc.NewServer()
		s := internalapi.NewTransferProxyServer(corpus)
		maintner_internal.RegisterInternalIssueServiceServer(grpcServer, s)

		lis, err := net.Listen("tcp", *intListen)
		if err != nil {
			log.Fatalf("failed to listen %v", err)
		}

		log.Printf("internal gRPC server listening on: %s", *intListen)
		return grpcServer.Serve(lis)
	})

	group.Go(
		// Get SLO rules for the repo
		func() error {
			parent := fmt.Sprintf("owners/%s/repositories/%s", *owner, *repo)

			ticker := time.NewTicker(10 * time.Minute)

			for t := range ticker.C {
				log.Printf("Slo sync at %v", t)

				_, err := getSlos(ctx, parent)
				if err != nil {
					logAndPrintError(err)
					log.Printf("Slo sync err: %v", err)
				}
			}
			return nil
		})

	err = group.Wait()
	log.Fatal(err)
}

func getSlos(ctx context.Context, parent string) ([]*drghs_v1.SLO, error) {

	conn, err := grpc.Dial(
		*sloAddress,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(
			grpc_middleware.ChainUnaryClient(
				grpctrace.UnaryClientInterceptor(global.Tracer("maintner-leif")),
				buildRetryInterceptor(),
			),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("Error connecting to SLO server: %v", err)
	}
	defer conn.Close()

	sloClient := drghs_v1.NewSLOServiceClient(conn)

	response, err := sloClient.ListSLOs(ctx, &drghs_v1.ListSLOsRequest{Parent: parent})
	if err != nil {
		return nil, fmt.Errorf("Error getting SLOs: %v", err)
	}

	slos := response.GetSlos()
	nextPage := response.GetNextPageToken()

	for nextPage != "" {
		response, err = sloClient.ListSLOs(ctx, &drghs_v1.ListSLOsRequest{Parent: parent, PageToken: nextPage})
		if err != nil {
			logAndPrintError(err)
			log.Printf("Error getting SLOs: %v", err)
			continue
		}

		slos = append(slos, response.GetSlos()...)
		nextPage = response.GetNextPageToken()
	}
	return slos, nil
}

func logAndPrintError(err error) {
	errorClient.Report(errorreporting.Entry{
		Error: err,
	})
	log.Print(err)
}

func unaryInterceptorLog(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	log.Printf("Starting RPC: %v at %v", info.FullMethod, start)

	// Used for Debugging incoming context and metadata issues
	// md, _ := metadata.FromIncomingContext(ctx)
	// log.Printf("RPC: %v. Metadata: %v", info.FullMethod, md)
	m, err := handler(ctx, req)
	if err != nil {
		errorClient.Report(errorreporting.Entry{
			Error: err,
		})
		log.Printf("RPC: %v failed with error %v", info.FullMethod, err)
	}

	log.Printf("Finishing RPC: %v. Took: %v", info.FullMethod, time.Now().Sub(start))
	return m, err
}

func buildRetryInterceptor() grpc.UnaryClientInterceptor {
	opts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(500 * time.Millisecond)),
		grpc_retry.WithCodes(codes.NotFound, codes.Aborted, codes.Unavailable, codes.DeadlineExceeded),
		grpc_retry.WithMax(5),
	}
	return grpc_retry.UnaryClientInterceptor(opts...)
}
