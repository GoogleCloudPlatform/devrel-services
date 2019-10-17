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
	"strings"

	"devrel/cloud/devrel-github-service/repos"

	"cloud.google.com/go/storage"
	"golang.org/x/build/maintner"
	"golang.org/x/build/maintner/maintnerd/gcslog"
)

var (
	source         = flag.String("source", "", "The bucket to read from")
	settingsBucket = flag.String("settings-bucket", "cdpe-maintner-settings", "Google Cloud Storage bucket to use for settings storage")
	reposFileName  = flag.String("file", "public_repos.json", "The list of public repos")
	projectID      = flag.String("gcp-project", "", "The GCP Project this is using")
)

type filterMutationSource struct {
	Log   []maintner.MutationStreamEvent
	Owner string
	Repo  string
}

func (fl *filterMutationSource) GetMutations(ctx context.Context) <-chan maintner.MutationStreamEvent {
	ch := make(chan maintner.MutationStreamEvent, 50) // buffered: overlap gunzip/unmarshal with loading
	go func() {
		done := ctx.Done()
		for _, e := range fl.Log {
			if e.Err != nil || e.End {
				// No-op pass through
			} else if e.Mutation.GithubIssue != nil && (e.Mutation.GithubIssue.Owner != fl.Owner || e.Mutation.GithubIssue.Repo != fl.Repo) {
				continue
			} else if e.Mutation.Github != nil && (e.Mutation.Github.Owner != fl.Owner || e.Mutation.Github.Repo != fl.Repo) {
				continue
			}

			select {
			case <-done:
				ch <- maintner.MutationStreamEvent{
					End: true,
				}
				return
			case ch <- e:
			}
		}
	}()
	return ch
}

func cacheMutations(ctx context.Context, m maintner.MutationSource) ([]maintner.MutationStreamEvent, error) {
	evs := make([]maintner.MutationStreamEvent, 0)
	ch := m.GetMutations(ctx)
	done := ctx.Done()
	for {
		select {
		case <-done:
			break
		case e := <-ch:
			if e.Err != nil {
				return nil, e.Err
			}
			evs = append(evs, e)
			if e.End {
				return evs, nil
			}
		}
	}
	return evs, nil
}

func main() {
	flag.Parse()
	if *source == "" {
		log.Fatalf("--source is required")
	}

	if *settingsBucket == "" {
		log.Fatalf("--settings-bucket is required")
	}
	if *projectID == "" {
		log.Fatalf("--gcp-project is required")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repoList := repos.NewBucketRepo(*settingsBucket, *reposFileName)
	_, err := repoList.UpdateTrackedRepos(ctx)
	if err != nil {
		log.Fatal(err)
	}

	sourcelog, err := gcslog.NewGCSLog(ctx, *source)
	if err != nil {
		log.Fatalf("error initializing source log: %v", err)
	}
	log.Print("Creating cache")
	cch, err := cacheMutations(ctx, sourcelog)
	log.Print("Created cache")
	if err != nil {
		log.Fatal(err)
	}

	for _, ta := range repoList.GetTrackedRepos() {

		fil := &filterMutationSource{
			Log:   cch,
			Owner: ta.Owner,
			Repo:  ta.Name,
		}

		bucketN := bucketName(ta)
		log.Printf("Creating bucket: %v, %v\n", bucketN, *projectID)
		err := createBucket(ctx, ta, *projectID)
		if err != nil {
			log.Fatalf("error creating bucket: %v", err)
		}

		destlog, err := gcslog.NewGCSLog(ctx, bucketN)
		if err != nil {
			log.Fatalf("error initializing dest log: %v", err)
		}

		log.Printf("Beginning copy for repo: %v/%v", ta.Owner, ta.Name)
		err = destlog.CopyFrom(fil)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func createBucket(ctx context.Context, ta repos.TrackedRepository, projectID string) error {
	sc, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	name := bucketName(ta)
	b := sc.Bucket(name)
	err = b.Create(ctx, projectID, nil)
	if err != nil && err.Error() == "googleapi: Error 409: You already own this bucket. Please select another name., conflict" {
		err = nil
	}
	return err
}

func bucketName(t repos.TrackedRepository) string {
	bld := strings.Builder{}
	bld.WriteString("mtr-p-")
	s := t.RepoSha()
	bld.WriteString(s)
	return bld.String()
}
