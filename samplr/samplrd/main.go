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
	"time"

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"
	"github.com/GoogleCloudPlatform/devrel-services/samplr"
	"github.com/GoogleCloudPlatform/devrel-services/samplr/samplrd/samplrapi"

	"cloud.google.com/go/profiler"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

var (
	listen  = flag.String("listen", "0.0.0.0:3009", "gRPC listen address")
	owner   = flag.String("owner", "", "Google Cloud Storage bucket to use for settings storage")
	repo    = flag.String("repo", "", "File that contains the list of repositories")
	verbose = flag.Bool("verbose", false, "Verbose logs")
)

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
	samplr.FormatLog(log.Formatter)

	log.Out = os.Stdout
}

func main() {

	flag.Parse()

	if *verbose == true {
		log.Level = logrus.DebugLevel
		samplr.VerboseLog()
	}

	if *owner == "" {
		err := fmt.Errorf("must provide --owner")
		log.Fatal(err)
	}

	if *repo == "" {
		err := fmt.Errorf("must provide --repos")
		log.Fatal(err)
	}

	// Profiler initialization, best done as early as possible.
	if err := profiler.Start(profiler.Config{
		Service:        fmt.Sprintf("samplrd-%v-%v", *owner, *repo),
		ServiceVersion: "0.0.1",
		MutexProfiling: true,
	}); err != nil {
		log.Error(fmt.Errorf("error initializing profiler: %v", err))
	}

	corpus := &samplr.Corpus{}

	loadGroup, _ := errgroup.WithContext(context.Background())

	repoPath := fmt.Sprintf("https://github.com/%v/%v", *owner, *repo)
	loadGroup.Go(func() error {
		log.Printf("Tracking repo: %s", repoPath)
		return corpus.TrackGit(repoPath)
	})

	if err := loadGroup.Wait(); err != nil {
		log.Fatal(err)
		return
	}

	err := corpus.Initialize(context.Background())
	if err != nil {
		log.Error(err)
		return
	}

	group, ctx := errgroup.WithContext(context.Background())

	group.Go(func() error {
		return corpus.Sync(ctx)
	})

	group.Go(func() error {
		lis, err := net.Listen("tcp", *listen)
		if err != nil {
			log.Fatalf("failed to listen %v", err)
		}

		grpcServer := grpc.NewServer()
		drghs_v1.RegisterSampleServiceServer(grpcServer, samplrapi.NewSampleServiceServer(corpus))

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
