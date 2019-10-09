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
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	drghs_v1 "devrel/cloud/devrel-github-service/drghs/v1"
	"devrel/cloud/devrel-github-service/repos"

	"cloud.google.com/go/errorreporting"
	"cloud.google.com/go/profiler"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Flags
var (
	listen     = flag.String("listen", ":6343", "listen address")
	verbose    = flag.Bool("verbose", false, "enable verbose debug output")
	supervisor = flag.String("supervisor", "", "the name of the service that is hosting the supervisor")
)

var (
	errorClient *errorreporting.Client
	pathRegex   = regexp.MustCompile(`^owners\/([\w-]+)\/repositories\/([\w-]+)[\w\/-]*$`)
)

const (
	// DEVNULL is used to forward requests to an address that
	// is "guarenteed" to fail
	// Using a reserved TLD https://tools.ietf.org/html/rfc2606
	DEVNULL = "devnul.invalid"
)

// Log
var log *logrus.Logger

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

	log.Out = os.Stdout
}

func main() {
	flag.Parse()

	if *verbose {
		log.Level = logrus.TraceLevel
	}

	if *supervisor == "" {
		log.Fatal("error: must specify --supervisor")
	}

	if *listen == "" {
		log.Fatal("error: must specify --listen")
	}

	if err := profiler.Start(profiler.Config{
		Service:        "samplr-rtr",
		ServiceVersion: "0.0.1",
	}); err != nil {
		log.Errorf("error staring profiler: %v", err)
	}

	lis, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	drghs_v1.RegisterSampleServiceServer(grpcServer, &reverseProxyServer{})
	log.Printf("gRPC server listening on: %s", *listen)
	grpcServer.Serve(lis)
}

type reverseProxyServer struct{}

func (s *reverseProxyServer) ListRepositories(ctx context.Context, req *drghs_v1.ListRepositoriesRequest) (*drghs_v1.ListRepositoriesResponse, error) {
	// TODO(colnnelson): This will need to reach out to the k8s api server
	// get all services with "owner" tag == request owner && then read the "repo"
	// tag from them

	return &drghs_v1.ListRepositoriesResponse{}, nil
}

func (s *reverseProxyServer) ListGitCommits(ctx context.Context, req *drghs_v1.ListGitCommitsRequest) (*drghs_v1.ListGitCommitsResponse, error) {
	pth, err := calculateHost(req.Parent)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(pth, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := drghs_v1.NewSampleServiceClient(conn)
	return client.ListGitCommits(ctx, req)
}

func (s *reverseProxyServer) GetGitCommit(ctx context.Context, req *drghs_v1.GetGitCommitRequest) (*drghs_v1.GitCommit, error) {
	pth, err := calculateHost(req.Name)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(pth, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := drghs_v1.NewSampleServiceClient(conn)
	return client.GetGitCommit(ctx, req)
}

func (s *reverseProxyServer) ListFiles(ctx context.Context, req *drghs_v1.ListFilesRequest) (*drghs_v1.ListFilesResponse, error) {
	pth, err := calculateHost(req.Parent)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(pth, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := drghs_v1.NewSampleServiceClient(conn)
	return client.ListFiles(ctx, req)
}

func (s *reverseProxyServer) ListSnippets(ctx context.Context, req *drghs_v1.ListSnippetsRequest) (*drghs_v1.ListSnippetsResponse, error) {
	pth, err := calculateHost(req.Parent)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(pth, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := drghs_v1.NewSampleServiceClient(conn)
	return client.ListSnippets(ctx, req)
}

func (s *reverseProxyServer) ListSnippetVersions(ctx context.Context, req *drghs_v1.ListSnippetVersionsRequest) (*drghs_v1.ListSnippetVersionsResponse, error) {
	pth, err := calculateHost(req.Parent)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(pth, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := drghs_v1.NewSampleServiceClient(conn)
	return client.ListSnippetVersions(ctx, req)
}

func calculateHost(path string) (string, error) {
	// We might need to put some more real "smarts" to this logic
	// in the event we need to handle the /v1/owners/*/repositories
	// call, which asks for a list of all repositories in a given org.
	// We might need to call out to a different API, but for now we can
	// forward to "null"?

	// As of right now, this function assumes all calls into the
	// proxy are of form /v1/owners/OWNERNAME/repositories/REPOSITORYNAME/snippets/*
	log.Tracef("Matching path againtst regex: %v", path)
	mtches := pathRegex.FindAllStringSubmatch(path, -1)
	if mtches != nil {
		log.Trace("Got match!")
		// Got Match
		// for now we are going to hardcode the format to owner.repository
		// long term this might extracted into an argument

		// This match will be of form:
		// [["/v1/owners/foo/repositories/bar1/snippets" "foo" "bar1"]]
		// Therefore slice the array
		ta := repos.TrackedRepository{
			Owner: mtches[0][1],
			Name:  mtches[0][2],
		}

		sn, err := serviceName(ta)
		if err != nil {
			return "", err
		}

		log.Tracef("New Host: %v", sn)
		return sn + ":80", nil
	}
	log.Tracef("No match... returning null: %v", DEVNULL)
	return DEVNULL, nil
}

func serviceName(t repos.TrackedRepository) (string, error) {
	return strings.ToLower(fmt.Sprintf("smp-s-%s", t.RepoSha())), nil
}
