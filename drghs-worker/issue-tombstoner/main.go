// Copyright 2022 Google LLC
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
	"strconv"
	"strings"
	"time"

	maintner_internal "github.com/GoogleCloudPlatform/devrel-services/drghs-worker/internal"
	"github.com/GoogleCloudPlatform/devrel-services/repos"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"google.golang.org/grpc"
)

// Log
var log *logrus.Logger

// Flags
var (
	flOwner *string
	flRepo  *string
)

// Constants
const (
	ServicePortInternal = "8080"
)

func init() {
	log = logrus.New()
	if !terminal.IsTerminal(int(os.Stdout.Fd())) {
		log.Formatter = &logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "severity",
				logrus.FieldKeyMsg:   "message",
			},
			TimestampFormat: time.RFC3339Nano,
		}
	}
	log.SetLevel(logrus.TraceLevel)
	log.Out = os.Stdout

	flOwner = flag.String("owner", "", "specifies the owner")
	flRepo = flag.String("repo", "", "specifies the repository")
}

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		sig := <-signalCh
		log.Printf("termination signal received: %s", sig)
		cancel()
	}()

	if *flOwner == "" {
		log.Fatal("--owner is empty")
	}

	if *flRepo == "" {
		log.Fatal("--repo is empty")
	}

	if len(flag.Args()) == 0 {
		log.Fatal("must provide more than one issue id")
	}

	issueIds := make([]int32, len(flag.Args()))
	for i, id := range flag.Args() {
		pid, err := strconv.Atoi(id)
		if err != nil {
			log.Fatalf("bad id: %v", id)
		}
		issueIds[i] = int32(pid)
	}

	tr := repos.TrackedRepository{
		Owner: *flOwner,
		Name:  *flRepo,
	}

	err := flagIssuesTombstoned(ctx, &tr, issueIds)
	if err != nil {
		log.Fatalf("got error tombstoning issues: %v", err)
	}

	log.Infof("Successfully tombstoned issues: %v", issueIds)
}

func flagIssuesTombstoned(ctx context.Context, tr *repos.TrackedRepository, issueIds []int32) error {
	maddr, err := serviceName(tr)
	if err != nil {
		return err
	}

	conn, err := grpc.Dial(
		maddr+":"+ServicePortInternal,
		grpc.WithInsecure(),
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	c := maintner_internal.NewInternalIssueServiceClient(conn)

	req := &maintner_internal.TombstoneIssuesRequest{
		Parent:       fmt.Sprintf("%v/%v", tr.Owner, tr.Name),
		IssueNumbers: issueIds,
	}
	resp, err := c.TombstoneIssues(ctx, req)
	if err != nil {
		return err
	}

	log.Infof("tombstoned: %v issues", resp.TombstonedCount)
	if int(resp.TombstonedCount) != len(issueIds) {
		log.Warnf("expected to tombstone %v, actually tombstoned: %v", len(issueIds), resp.TombstonedCount)
	}

	return err
}

func serviceName(t *repos.TrackedRepository) (string, error) {
	return strings.ToLower(fmt.Sprintf("mtr-s-%s", t.RepoSha())), nil
}
