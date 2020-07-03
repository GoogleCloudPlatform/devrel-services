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

	"github.com/GoogleCloudPlatform/devrel-services/leif"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// if owner is specified, file is to be found in that bucket, elsewise, it's local
// kept these as in leifd to maintain consistency but may have to change if confusing/diff in this context
var (
	listen  = flag.String("listen", "0.0.0.0:3009", "gRPC listen address")
	owner   = flag.String("owner", "", "Google Cloud Storage bucket to use for settings storage")
	repos   = flag.String("repo", "", "File that contains the list of repositories")
	verbose = flag.Bool("verbose", false, "Verbose logs")
)

var log *logrus.Logger

func parseFlags() {
	flag.Parse()

	if *verbose == true {
		log.Level = logrus.DebugLevel
		// leif.VerboseLog()
	}

	if *repos == "" {
		err := fmt.Errorf("must provide --repos")
		log.Fatal(err)
	}
}

func loadRepos() {

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

}

func main() {
	fmt.Println("Hello World!")

	log = logrus.New()
	parseFlags()

	corpus := &leif.Corpus{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		sig := <-signalCh
		fmt.Printf("termination signal received: %s", sig)
		cancel()
	}()

	repo := leif.Repository{}
	fmt.Println(repo)
	fmt.Println(ctx)
	// err := repo.FindRepository(ctx, "e")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

}
