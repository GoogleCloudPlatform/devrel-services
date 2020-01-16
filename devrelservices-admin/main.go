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
	"os/signal"
	"time"

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"

	"cloud.google.com/go/errorreporting"
	"cloud.google.com/go/profiler"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// Flags
var (
	listen  = flag.String("listen", ":6343", "listen address")
	verbose = flag.Bool("verbose", false, "enable verbose debug output")
	smpSpr  = flag.String("smp-spr", "samplrd-sprvsr", "address of samplrd-sprvsr")
	mtrSpr  = flag.String("mtr-spr", "maintnerd-sprvsr", "address of maintnerd-sprvsr")
)

var (
	errorClient *errorreporting.Client
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

	if *smpSpr == "" {
		log.Fatal("error: must specify --smp-spr")
	}

	if *mtrSpr == "" {
		log.Fatal("error: must specify --mnt-spr")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		sig := <-signalCh
		log.Printf("termination signal received: %s", sig)
		cancel()
	}()

	if err := profiler.Start(profiler.Config{
		Service:        "devrelservices-admin",
		ServiceVersion: "0.0.1",
	}); err != nil {
		log.Errorf("error staring profiler: %v", err)
	}

	lis, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	reverseProxy := &adminServer{}
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(unaryInterceptorLog))

	go func() {
		<-ctx.Done()
		log.Warn("shutting down server")
		grpcServer.Stop()
	}()

	drghs_v1.RegisterDevRelServicesAdminServer(grpcServer, reverseProxy)
	healthpb.RegisterHealthServer(grpcServer, reverseProxy)
	log.Printf("gRPC server listening on: %s", *listen)
	grpcServer.Serve(lis)

}

type adminServer struct{}

// Check is for health checking.
func (s *adminServer) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (s *adminServer) Watch(req *healthpb.HealthCheckRequest, ws healthpb.Health_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "health check via Watch not implemented")
}

func (s *adminServer) UpdateTrackedRepos(ctx context.Context, r *drghs_v1.UpdateTrackedReposRequest) (*drghs_v1.UpdateTrackedReposResponse, error) {
	log.Debug("updating repository list for maintner")

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://%s/update", *mtrSpr), nil)
	if err != nil {
		return &drghs_v1.UpdateTrackedReposResponse{}, err
	}
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return &drghs_v1.UpdateTrackedReposResponse{}, err
	}

	log.Debug("updating repository list for samplr")
	req, err = http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://%s/update", *smpSpr), nil)
	if err != nil {
		return &drghs_v1.UpdateTrackedReposResponse{}, err
	}
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return &drghs_v1.UpdateTrackedReposResponse{}, err
	}

	return &drghs_v1.UpdateTrackedReposResponse{}, nil
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
