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
	"os"
	"os/signal"
	"time"

	"github.com/GoogleCloudPlatform/devrel-services/leif"
	"github.com/GoogleCloudPlatform/devrel-services/leif/githubreposervice"
	"github.com/GoogleCloudPlatform/devrel-services/repos"

	"github.com/sirupsen/logrus"
)

// if bucket is specified, repos file is to be found in that bucket, elsewise, it's local
var (
	bucket       = flag.String("bucket", "", "Google Cloud Storage bucket to use for settings storage")
	listen       = flag.String("listen", "0.0.0.0:3009", "gRPC listen address")
	reposFile    = flag.String("repos", "", "File that contains the list of repositories")
	syncInterval = flag.Int("sync", 10, "The SLO rules update every X minutes")
	verbose      = flag.Bool("verbose", false, "Verbose logs")
)

var log *logrus.Logger
var repoList repos.RepoList

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

func main() {
	fmt.Println("Hello World!")

	log = logrus.New()
	parseFlags()

	corpus := &leif.Corpus{}

	if *bucket != "" {
		repoList = repos.NewBucketRepo(*bucket, *reposFile)
	}

	_, err := repoList.UpdateTrackedRepos(context.Background())
	if err != nil {
		log.Fatal(err)
		return
	}

	ghClient := githubreposervice.NewClient(nil, nil)
	for _, r := range repoList.GetTrackedRepos() {
		corpus.TrackRepo(context.Background(), r.Owner, r.Name, &ghClient)
	}

	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		sig := <-signalCh
		fmt.Printf("termination signal received: %s", sig)
		cancel()
	}()

	fmt.Println(corpus)
}
