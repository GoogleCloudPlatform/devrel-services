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

package samplr

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	// log.Level = logrus.DebugLevel
	log.Out = os.Stdout
}

// VerboseLog sets the log level to DebugLevel
func VerboseLog() {
	log.Level = logrus.DebugLevel
}

// FormatLog sets the log's formatter to the given one
func FormatLog(f logrus.Formatter) {
	log.Formatter = f
}

// Corpus holds all of a project's metadata.
type Corpus struct {
	verbose bool

	mu sync.RWMutex // guards all following fields
	// corpus state:
	didInit bool
	debug   bool
	syncing bool

	watchedGitRepos []WatchedRepository

	// github-specific

	gitReposToAdd chan WatchedRepository
}

// RLock grabs the corpus's read lock. Grabbing the read lock prevents
// any concurrent writes from mutating the corpus. This is only
// necessary if the application is querying the corpus and calling its
// Update method concurrently.
func (c *Corpus) RLock() { c.mu.RLock() }

// RUnlock unlocks the corpus's read lock.
func (c *Corpus) RUnlock() { c.mu.RUnlock() }

// SetVerbose enables or disables verbose logging.
func (c *Corpus) SetVerbose(v bool) { c.verbose = v }

// SetDebug instructs the Corpus to run in debug mode
func (c *Corpus) SetDebug() {
	c.debug = true
}

func (c *Corpus) debugf(format string, v ...interface{}) {
	if c.debug {
		log.Printf(format, v...)
	}
}

// Initialize should be the first call to the corpus to
// do the initial clone and synchronizing of the corpus's
// repository set.
func (c *Corpus) Initialize(ctx context.Context) error {
	c.mu.Lock()
	if c.didInit {
		c.mu.Unlock()
		return fmt.Errorf("multiple calls to Initialize")
	}
	defer c.mu.Unlock()

	log.Info("Corpus Initializing")

	for _, wrepo := range c.watchedGitRepos {
		log.Debugf("Starting initial update of repo %s", wrepo.ID())
		wrepo.Update(ctx)
		log.Debugf("Finished initial update of repo %s", wrepo.ID())
		return nil
	}

	c.didInit = true
	log.Info("Corpus finished Initializing")
	return nil
}

// Sync instructs the Corpus to iterate over its tracked repositories
// and update all of them.
func (c *Corpus) Sync(ctx context.Context) error {
	c.mu.Lock()
	if c.syncing {
		c.mu.Unlock()
		return fmt.Errorf("Cant SyncLoop while already sync-ing")
	}
	c.syncing = true

	c.gitReposToAdd = make(chan WatchedRepository)
	c.mu.Unlock()

	err := c.sync(ctx)

	c.mu.Lock()
	c.syncing = false
	close(c.gitReposToAdd)
	c.gitReposToAdd = nil
	c.mu.Unlock()

	return err
}

func (c *Corpus) sync(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	updateGitRepo := func(gr WatchedRepository) error {
		log.Printf("Beginning syncLoop for %v...", gr.ID())
		for {
			log.Printf("polling %v ...", gr.ID())
			err := gr.Update(ctx)
			if err == nil {
				time.Sleep(30 * time.Second)
				continue
			}
			log.Printf("git sync ending for %v: %v", gr.ID(), err)
			return err
		}
	}

	// These goroutines will exit when SyncLoop returns as
	// their respective channels will be closed.
	go func() {
		for w := range c.gitReposToAdd {
			repo := w
			group.Go(func() error {
				return updateGitRepo(repo)
			})
		}
	}()

	for _, gr := range c.watchedGitRepos {
		repo := gr
		group.Go(func() error {
			return updateGitRepo(repo)
		})
	}

	return group.Wait()
}

// ForEachRepo iterates over the set of repositories and performs the
// given function on each and returns the first non-nill error it recieves.
func (c *Corpus) ForEachRepo(fn func(repo WatchedRepository) error) error {
	return c.ForEachRepoF(fn, func(repo WatchedRepository) bool { return true })
}

// ForEachRepoF iterates over the set of repositories that match the given filter
// and performs the given function on them, and returns the first non-nill error
// it recieves.
func (c *Corpus) ForEachRepoF(fn func(repo WatchedRepository) error, filter func(repo WatchedRepository) bool) error {
	for _, repo := range c.watchedGitRepos {
		if filter(repo) {
			if err := fn(repo); err != nil {
				return err
			}
		}
	}
	return nil
}
