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
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"
	"github.com/GoogleCloudPlatform/devrel-services/repos"

	"cloud.google.com/go/errorreporting"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

var (
	listen     = flag.String("listen", ":6343", "listen address")
	verbose    = flag.Bool("verbose", false, "enable verbose debug output")
	sprvsrAddr = flag.String("sprvsr", "maintner-sprvsr", "address for supervisor")
	rbucket    = flag.String("settings-bucket", "", "bucket to get repo list")
	rfile      = flag.String("repos-file", "", "file in bucket to read repos from")
)

var (
	errorClient *errorreporting.Client
	pathRegex   = regexp.MustCompile(`^([.:\w-]+)\/([.:\w-]+)[.:\w\/-]*$`)
)

const (
	// Using a reserved TLD https://tools.ietf.org/html/rfc2606
	devnull = "devnull.invalid"

	// Max size to recieve. 32 MB
	maxMessageSize = 32000000
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

	if *listen == "" {
		log.Fatal("error: must specify --listen")
	}

	if *rbucket == "" {
		log.Fatal("error: must specify --settings-bucket")
	}
	if *rfile == "" {
		log.Fatal("error: must specify --repos-file")
	}

	rlist := repos.NewBucketRepo(*rbucket, *rfile)
	_, err := rlist.UpdateTrackedRepos(context.Background())
	if err != nil {
		log.Fatalf("got error updating repos: %v", err)
	}

	lis, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(unaryInterceptorLog))
	reverseProxy := &reverseProxyServer{
		reps: rlist,
	}
	drghs_v1.RegisterIssueServiceServer(grpcServer, reverseProxy)
	drghs_v1.RegisterIssueServiceAdminServer(grpcServer, reverseProxy)
	healthpb.RegisterHealthServer(grpcServer, reverseProxy)
	log.Printf("gRPC server listening on: %s", *listen)
	grpcServer.Serve(lis)
}

type reverseProxyServer struct {
	reps repos.RepoList
}

// Check is for health checking.
func (s *reverseProxyServer) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (s *reverseProxyServer) Watch(req *healthpb.HealthCheckRequest, ws healthpb.Health_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "health check via Watch not implemented")
}

func (s *reverseProxyServer) ListRepositories(ctx context.Context, r *drghs_v1.ListRepositoriesRequest) (*drghs_v1.ListRepositoriesResponse, error) {
	resp := drghs_v1.ListRepositoriesResponse{}
	for _, tr := range s.reps.GetTrackedRepos() {

		if !tr.IsTrackingIssues {
			log.Debugf("skipping repo: %v", tr.String())
			continue
		}

		pth, err := calculateHost(tr.String())
		if err != nil {
			return nil, err
		}
		// Dial and get the repos
		log.Debugf("getting tracked repos from repo: %v path: %v", tr.String(), pth)
		conn, err := grpc.Dial(pth, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}

		client := drghs_v1.NewIssueServiceClient(conn)
		// Naive right now... every service has exactly one repo
		srepos, err := getTrackedRepositories(ctx, client)
		if err != nil {
			return nil, err
		}

		resp.Repositories = append(resp.Repositories, srepos...)
	}
	return &resp, nil
}

func (s *reverseProxyServer) ListIssues(ctx context.Context, r *drghs_v1.ListIssuesRequest) (*drghs_v1.ListIssuesResponse, error) {
	pth, err := calculateHost(r.Parent)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(pth,
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)),
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := drghs_v1.NewIssueServiceClient(conn)
	return client.ListIssues(ctx, r)
}

func (s *reverseProxyServer) GetIssue(ctx context.Context, r *drghs_v1.GetIssueRequest) (*drghs_v1.GetIssueResponse, error) {
	pth, err := calculateHost(r.Name)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(pth,
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)),
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := drghs_v1.NewIssueServiceClient(conn)
	return client.GetIssue(ctx, r)
}

func (s *reverseProxyServer) UpdateTrackedRepos(ctx context.Context, r *drghs_v1.UpdateTrackedReposRequest) (*drghs_v1.UpdateTrackedReposResponse, error) {
	_, err := http.Get(fmt.Sprintf("http://%s/update", *sprvsrAddr))
	s.reps.UpdateTrackedRepos(ctx)

	return &drghs_v1.UpdateTrackedReposResponse{}, err
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

func calculateHost(path string) (string, error) {
	// We might need to put some more real "smarts" to this logic
	// in the event we need to handle the /v1/owners/*/repositories
	// call, which asks for a list of all repositories in a given org.
	// We might need to call out to a different API, but for now we can
	// forward to "null"?

	// As of right now, this function assumes all calls into the
	// proxy are of form /v1/owners/OWNERNAME/repositories/REPOSITORYNAME/issues/*
	log.Tracef("Matching path againtst regex: %v", path)
	mtches := pathRegex.FindAllStringSubmatch(path, -1)
	if mtches != nil {
		log.Tracef("Got a match!")
		// This match will be of form:
		// [["/v1/owners/foo/repositories/bar1/issues" "foo" "bar1"]]
		// Therefore slice the array
		ta := repos.TrackedRepository{
			Owner: mtches[0][1],
			Name:  mtches[0][2],
		}

		sn, err := serviceName(ta)
		if err != nil {
			return "", err
		}
		return sn + ":80", nil
	}
	log.Tracef("No match... returning null: %v", devnull)
	return devnull, nil
}

func serviceName(t repos.TrackedRepository) (string, error) {
	return strings.ToLower(fmt.Sprintf("mtr-s-%s", t.RepoSha())), nil
}

func unaryInterceptorLog(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	log.Tracef("Starting RPC: %v at %v", info.FullMethod, start)

	m, err := handler(ctx, req)
	if err != nil {
		log.Errorf("RPC: %v failed with error %v", info.FullMethod, err)
	}

	log.Tracef("Finishing RPC: %v. Took: %v", info.FullMethod, time.Now().Sub(start))
	return m, err
}
