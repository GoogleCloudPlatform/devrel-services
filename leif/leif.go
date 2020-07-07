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
	"errors"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/GoogleCloudPlatform/devrel-services/leif/githubreposervice"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.Out = os.Stdout
}

func VerboseLog() {
	log.Level = logrus.DebugLevel
}

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

	watchedOwners []Owner
	// watchedRepos  []Repository do we want this

	// github-specific

	// gitReposToAdd chan WatchedRepository
}

func (c *Corpus) TrackOwner(ctx context.Context, name string, ghClient *githubreposervice.Client) error {

	i := sort.Search(len(c.watchedOwners), func(i int) bool { return c.watchedOwners[i].name >= name })

	if i < len(c.watchedOwners) && c.watchedOwners[i].name == name {
		// already tracked
		return nil
	}

	owner := Owner{name: name}
	rules, err := findSLODoc(ctx, owner, "", ghClient)
	if err != nil {
		log.Error(err)
		var ghErrorResponse *goGitHubErr

		if !(errors.As(err, &ghErrorResponse) && ghErrorResponse.Response.StatusCode == 404) {
			// SLO config not found
			return err
		}
	}
	owner.SLORules = rules
	c.watchedOwners = append(c.watchedOwners, owner)
	copy(c.watchedOwners[i+1:], c.watchedOwners[i:])
	c.watchedOwners[i] = owner

	return nil
}

func (c *Corpus) TrackRepo(ctx context.Context, owner string, repo string, ghClient *githubreposervice.Client) error {

	err := c.TrackOwner(ctx, owner, ghClient)
	if err != nil {
		return err
	}

	ownerIndex := sort.Search(len(c.watchedOwners), func(i int) bool { return c.watchedOwners[i].name >= owner })

	repoIndex := sort.Search(len(c.watchedOwners[ownerIndex].Repos), func(i int) bool { return c.watchedOwners[ownerIndex].Repos[i].name >= repo })

	watchedOwner := &c.watchedOwners[ownerIndex]

	if repoIndex < len(watchedOwner.Repos) && watchedOwner.Repos[repoIndex].name == repo {
		// repo already tracked
		return nil
	}

	addRepo := Repository{name: repo}
	rules, err := findSLODoc(ctx, *watchedOwner, repo, ghClient)
	if err != nil {
		log.Error(err)
		var ghErrorResponse *goGitHubErr

		if !(errors.As(err, &ghErrorResponse) && ghErrorResponse.Response.StatusCode == 404) {
			// SLO config not found
			return err
		}
	}
	addRepo.SLORules = rules
	watchedOwner.Repos = append(watchedOwner.Repos, addRepo)
	copy(watchedOwner.Repos[repoIndex+1:], watchedOwner.Repos[repoIndex:])
	watchedOwner.Repos[repoIndex] = addRepo

	return nil
}

// Owner represents a GitHub owner and their tracked repositories
// Owners can specify default SLO rules that will apply to all tracked repos
// unless the repository overrides them with its own SLO rules config
type Owner struct {
	name     string
	Repos    []Repository
	SLORules []*SLORule
}

// Repository represents a GitHub repository and stores its SLO rules
type Repository struct {
	name     string
	SLORules []*SLORule
}

// SLORule represents a service level objective (SLO) rule
type SLORule struct {
	AppliesTo          AppliesTo          `json:"appliesTo"`
	ComplianceSettings ComplianceSettings `json:"complianceSettings"`
}

// AppliesTo stores structured data on which issues and/or pull requests a SLO applies to
type AppliesTo struct {
	GitHubLabels         []string `json:"gitHubLabels"`
	ExcludedGitHubLabels []string `json:"excludedGitHubLabels"`
	Issues               bool     `json:"issues"`
	PRs                  bool     `json:"prs"`
}

// ComplianceSetting stores data on the requirements for an issue or pull request to be considered compliant with the SLO
type ComplianceSettings struct {
	ResponseTime     time.Duration `json:"responseTime"`
	ResolutionTime   time.Duration `json:"resolutionTime"`
	RequiresAssignee bool          `json:"requiresAssignee"`
	Responders       Responders    `json:"responders"`
}

// Responders stores structured data on the responders to the issue or pull request the SLO applies to
type Responders struct {
	Owners       []string `json:"owners"`
	Contributors string   `json:"contributors"`
	Users        []string `json:"users"`
}
