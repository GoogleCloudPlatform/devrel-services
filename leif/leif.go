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

package leif

import (
	"context"
	// "fmt"
		// "os"
	"sync"
	// "time"

	// "golang.org/x/sync/errgroup"
)



// Corpus holds all of a project's metadata.
type Corpus struct {
	verbose bool

	mu sync.RWMutex // guards all following fields
	// corpus state:
	didInit bool
	debug   bool
	syncing bool

	watchedOrgs []Org

	// watchedGitRepos []WatchedRepository

	// github-specific

	// gitReposToAdd chan WatchedRepository
}

// Initialize should be the first call to the corpus to
// do the initial clone and synchronizing of the corpus's
// repository set.
func (c *Corpus) Init(ctx context.Context) error {
	return nil
}

func (c *Corpus) Sync(ctx context.Context) error {
	return nil
}


func (c *Corpus) ForEachRepo(ctx context.Context) error {
	return nil
}

// func (c *Corpus) ForEachRepoF(ctx context.Context, f func(w WatchedRepository) bool) error {
// 	return nil
// }

func (c *Corpus) TrackOrg(ctx context.Context, orgname string) error {
	// Make a few api calls to GitHub :)
	return nil
}

func (c *Corpus) TrackRepository(ctx context.Context, orgname string, reponame string) error {
	return nil
}


type Org struct {
	Name string

	Repos []*Repository
	SLORules []*SLORule
}

type Repository struct {
	SLORules []*SLORule
}

type SLORule struct {	
	name string
}