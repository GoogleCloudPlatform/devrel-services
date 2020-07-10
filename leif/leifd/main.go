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
	"strings"
	"sync"
	"time"

	"github.com/GoogleCloudPlatform/devrel-services/leif"
	"github.com/GoogleCloudPlatform/devrel-services/leif/githubreposervice"
	"github.com/GoogleCloudPlatform/devrel-services/repos"

	"github.com/gregjones/httpcache"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const gitHubEnvVar = "GITHUB_TOKEN"

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

var arr []int
var arrc chan int
var mu sync.RWMutex

func rando() {
	fmt.Println("Hello ra!")

	go func() {

		fmt.Println("in arrc")
		for e := range arrc {
			fmt.Println("updating c", e)
			time.Sleep(100 * time.Millisecond)
		}
		fmt.Println("go func done")
	}()

	for e := range arr {
		fmt.Println("updating", e)
		fmt.Println(arr)
		time.Sleep(100 * time.Millisecond)

	}
	fmt.Println("done")
	return
}

func addTo() {

	mu.Lock()
	defer mu.Unlock()
	// fmt.Println("adding to rrc")
	// fmt.Println(arrc)
	for i := 5; i < 10; i++ {
		arr = append(arr, i)
		arrc <- i
	}
	// // fmt.Println(arrc)
	// fmt.Println("added to rrc")
	return
}

var ghClient githubreposervice.Client

func initGHClient() {
	if os.Getenv(gitHubEnvVar) == "" {
		log.Fatalf("env var %v is empty", gitHubEnvVar)
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: strings.Trim(os.Getenv(gitHubEnvVar), "\n")},
	)
	hc := oauth2.NewClient(context.Background(), src)

	cachedTransport := httpcache.Transport{Transport: hc.Transport, Cache: httpcache.NewMemoryCache()}

	ghClient = githubreposervice.NewClient(cachedTransport.Client(), nil, nil)
}

func main() {
	fmt.Println("Hello World!")

	log = logrus.New()
	parseFlags()
	initGHClient()

	corpus := &leif.Corpus{}

	if *bucket != "" {
		repoList = repos.NewBucketRepo(*bucket, *reposFile)
	} else {
		repoList = leif.NewDiskRepo(*reposFile)
	}

	_, err := repoList.UpdateTrackedRepos(context.Background())
	if err != nil {
		log.Fatal(err)
		return
	}

	for _, r := range repoList.GetTrackedRepos() {
		if r.IsTrackingIssues {
			corpus.TrackRepo(context.Background(), r.Owner, r.Name, &ghClient)
		}
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
