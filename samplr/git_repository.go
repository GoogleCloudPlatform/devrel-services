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
	"io/ioutil"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"

	git "github.com/GoogleCloudPlatform/devrel-services/git-go"
	"golang.org/x/sync/errgroup"
)

var urlReg = regexp.MustCompile("https://github.com/([\\w-_]+)/([\\w-_]+)")

type watchedGitRepo struct {
	id         string
	repository *git.Repository
	c          *Corpus
	snippets   map[string][]*Snippet
	commits    map[string][]*GitCommit
	mu         sync.RWMutex
}

func (w *watchedGitRepo) ID() string {
	return w.id
}
func (w *watchedGitRepo) Update(ctx context.Context) error {
	err := w.repository.FetchContext(ctx, &git.FetchOptions{})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		log.Errorf("got error fetching repository: %v", err)
		return err
	}
	if err == git.NoErrAlreadyUpToDate && len(w.snippets) != 0 {
		// If the repo is already up to date and there
		// are snippets in the repository, return early
		log.Trace("already up to date, and we have snippets, skipping update")
		return nil
	}
	// Need to actually pull the remote in to get the new changes

	err = w.repository.PullContext(ctx, &git.PullOptions{RemoteName: "origin"})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		log.Errorf("got error pulling commits: %v", err)
		return err
	}

	group, ctx := errgroup.WithContext(ctx)
	refIter, err := w.repository.Branches()
	if err != nil {
		log.Printf("got error iterating branches %v", err)
		return err
	}
	refIter.ForEach(func(ref *git.Reference) error {
		if ref.Name() != git.Master {
			return nil
		}
		name := ref.Name()
		hash := ref.Hash()
		log.Debugf("Repo %v... working on reference: %v, %v", w.ID(), name, hash)
		group.Go(func() error {
			cIter, err := w.repository.Log(&git.LogOptions{From: hash})
			if err != nil {
				log.Errorf("Error %v", err)
				return err
			}
			snips, err := CalculateSnippets(w.Owner(), w.RepositoryName(), cIter)
			if err != nil {
				log.Errorf("Error calculating snippets for %s: %v", w.ID(), err)
				return err
			}
			w.mu.Lock()
			defer w.mu.Unlock()
			w.snippets[name.String()] = snips
			cIter = nil
			return nil
		})
		return nil
	})
	refIter = nil
	err = group.Wait()
	debug.FreeOSMemory()
	refIter, err = w.repository.Branches()
	if err != nil {
		log.Printf("Error %v", err)
		return err
	}
	err = refIter.ForEach(func(ref *git.Reference) error {
		if ref.Name() != git.Master {
			return nil
		}
		name := ref.Name()
		hash := ref.Hash()
		log.Infof("Repo %v... getting commits for reference: %v, %v", w.ID(), name, hash)
		cIter, err := w.repository.Log(&git.LogOptions{From: hash})
		if err != nil {
			return err
		}
		cmts := make([]*GitCommit, 0)
		err = cIter.ForEach(func(commit *git.Commit) error {
			cmt := GitCommit{
				Body:           commit.Message,
				Subject:        strings.Split(commit.Message, "\n")[0],
				AuthorEmail:    commit.Author.Email,
				AuthoredTime:   commit.Author.When,
				CommitterEmail: commit.Committer.Email,
				CommittedTime:  commit.Committer.When,
				Hash:           commit.Hash.String(),
				Name:           fmt.Sprintf("owners/%s/repositories/%s/gitCommits/%s", w.Owner(), w.RepositoryName(), commit.Hash.String()),
			}
			cmts = append(cmts, &cmt)
			return nil
		})
		w.mu.Lock()
		defer w.mu.Unlock()
		w.commits[name.String()] = cmts
		return err
	})
	log.Infof("%v commits found in repository %v", len(w.commits), w.ID())
	debug.FreeOSMemory()
	return err
}
func (w *watchedGitRepo) Owner() string {
	return urlReg.FindStringSubmatch(w.id)[1]
}
func (w *watchedGitRepo) RepositoryName() string {
	return urlReg.FindStringSubmatch(w.id)[2]
}
func (w *watchedGitRepo) ForEachSnippetF(fn func(snippet *Snippet) error, filter func(snippet *Snippet) bool) error {
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, snippets := range w.snippets {
		for _, snippet := range snippets {
			if filter(snippet) {
				if err := fn(snippet); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
func (w *watchedGitRepo) ForEachSnippet(fn func(snippet *Snippet) error) error {
	return w.ForEachSnippetF(fn, func(snippet *Snippet) bool {
		return true
	})
}
func (w *watchedGitRepo) ForEachGitCommit(fn func(g *GitCommit) error) error {
	return w.ForEachGitCommitF(fn, func(commit *GitCommit) bool {
		return true
	})
}
func (w *watchedGitRepo) ForEachGitCommitF(fn func(g *GitCommit) error, filter func(g *GitCommit) bool) error {
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, val := range w.commits {
		for _, cmt := range val {
			if filter(cmt) {
				if err := fn(cmt); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
func (c *Corpus) TrackGit(url string) error {
	dirname, err := ioutil.TempDir("", "samplr-")
	if err != nil {
		log.Errorf("Could not get a temp dir when cloning %v. Err: %v", url, err)
		return err
	}
	r, err := git.PlainClone(dirname, false, &git.CloneOptions{
		URL: url,
	})
	if err != nil {
		log.Errorf("Error cloning: %v\n%v", url, err)
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	wgh := &watchedGitRepo{
		repository: r,
		c:          c,
		id:         url,
		snippets:   make(map[string][]*Snippet),
		commits:    make(map[string][]*GitCommit),
	}
	c.watchedGitRepos = append(c.watchedGitRepos, wgh)
	if c.gitReposToAdd != nil {
		c.gitReposToAdd <- wgh
	}
	return nil
}
