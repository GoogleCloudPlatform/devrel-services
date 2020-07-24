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

package leif

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/GoogleCloudPlatform/devrel-services/leif/githubservices"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.Out = os.Stdout
}

// VerboseLog sets the log level to DebugLevel
func VerboseLog() {
	log.Level = logrus.DebugLevel
}

// FormatLog sets the log's formatter to the one provided
func FormatLog(f logrus.Formatter) {
	log.Formatter = f
}

// Corpus holds all of a project's metadata.
type Corpus struct {
	mu      sync.RWMutex // guards the sync state
	syncing bool

	watchedOwners []*Owner
	ownersToAdd   chan *Owner
}

// SyncLoop instructs the Corpus to update all the tracked repos every given amount of minutes
func (c *Corpus) SyncLoop(ctx context.Context, minutes int, ghClient *githubservices.Client) error {
	c.mu.Lock()

	if c.syncing {
		c.mu.Unlock()
		return fmt.Errorf("Sync error; duplicate calls to SyncLoop")
	}

	c.syncing = true
	c.ownersToAdd = make(chan *Owner)

	c.mu.Unlock()

	err := c.syncLoop(ctx, minutes, ghClient)

	c.mu.Lock()
	c.syncing = false
	close(c.ownersToAdd)
	c.ownersToAdd = nil

	c.mu.Unlock()

	return err
}

func (c *Corpus) syncLoop(ctx context.Context, minutes int, ghClient *githubservices.Client) error {
	group, ctx := errgroup.WithContext(ctx)

	go func() {
		for o := range c.ownersToAdd {
			ow := o
			group.Go(func() error {
				return ow.UpdateLoop(ctx, minutes, ghClient)
			})
		}
	}()

	for _, o := range c.watchedOwners {
		ow := o
		group.Go(func() error {
			return ow.UpdateLoop(ctx, minutes, ghClient)
		})
	}

	return group.Wait()
}

// TrackRepo adds the repository to the corpus to be tracked
func (c *Corpus) TrackRepo(ctx context.Context, ownerName string, repoName string, ghClient *githubservices.Client) error {

	owner, err := c.trackOwner(ctx, ownerName, ghClient)
	if err != nil {
		return err
	}

	return owner.trackRepo(ctx, repoName, ghClient)
}

func (c *Corpus) trackOwner(ctx context.Context, name string, ghClient *githubservices.Client) (*Owner, error) {

	i := sort.Search(len(c.watchedOwners), func(i int) bool { return c.watchedOwners[i].name >= name })

	if i < len(c.watchedOwners) && c.watchedOwners[i].name == name {
		return c.watchedOwners[i], nil
	}

	// check that it exists in GH:
	_, _, err := ghClient.Users.Get(ctx, name)
	if err != nil {
		log.Errorf("Unable to get owner %s from GitHub: %s", name, err)
		return nil, err
	}

	owner := Owner{name: name}
	c.watchedOwners = append(c.watchedOwners, &owner)
	copy(c.watchedOwners[i+1:], c.watchedOwners[i:])
	c.watchedOwners[i] = &owner
	if c.ownersToAdd != nil {
		c.ownersToAdd <- &owner
	}
	return c.watchedOwners[i], nil
}

// ForEachRepo iterates over the set of repositories and performs the
// given function on each and returns the first non-nil error it recieves.
func (c *Corpus) ForEachRepo(fn func(repo Repository) error) error {
	return c.ForEachRepoF(fn, func(repo Repository) bool { return true })
}

// ForEachRepoF iterates over the set of repositories that match the given filter
// and performs the given function on them, and returns the first non-nil error
// it recieves.
func (c *Corpus) ForEachRepoF(fn func(repo Repository) error, filter func(repo Repository) bool) error {
	for _, owner := range c.watchedOwners {
		for _, repo := range owner.Repos {
			if filter(*repo) {
				if err := fn(*repo); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
