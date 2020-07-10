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
	"time"

	"github.com/GoogleCloudPlatform/devrel-services/leif/githubreposervice"
	"github.com/google/go-github/v31/github"
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

func (c *Corpus) String() string {
	var s string
	for _, o := range c.watchedOwners {
		s += o.name + ": "
		for _, r := range o.Repos {
			s += r.name
			s += ","
		}
		s += "\n"
	}

	return s
}

func (c *Corpus) Initialize() error {
	c.mu.Lock()
	if c.didInit {
		c.mu.Unlock()
		return fmt.Errorf("Multiple calls to Initialize")
	}
	defer c.mu.Unlock()

	log.Info("Corpus Initializing")

	c.didInit = true
	log.Info("Corpus finished Initializing")
	return nil
}

func (c *Corpus) TrackOwner(ctx context.Context, name string, ghClient *githubreposervice.Client) error {

	owner, err := c.trackOwner(ctx, name, ghClient)
	if err != nil {
		return err
	}

	owner.trackAllRepos(ctx, ghClient)
	return nil
}

func (c *Corpus) TrackRepo(ctx context.Context, ownerName string, repoName string, ghClient *githubreposervice.Client) error {

	owner, err := c.trackOwner(ctx, ownerName, ghClient)
	if err != nil {
		return err
	}

	return owner.trackRepo(ctx, repoName, ghClient)
}

func (c *Corpus) trackOwner(ctx context.Context, name string, ghClient *githubreposervice.Client) (*Owner, error) {
	i := sort.Search(len(c.watchedOwners), func(i int) bool { return c.watchedOwners[i].name >= name })

	if i < len(c.watchedOwners) && c.watchedOwners[i].name == name {
		// already tracked
		return &c.watchedOwners[i], nil
	}

	// check that it exists in GH:
	_, _, err := ghClient.Organizations.Get(ctx, name)
	if err != nil {
		log.Errorf("Unable to get org %s from GitHub: %s", name, err)
		return nil, err
	}

	owner := Owner{name: name}
	c.watchedOwners = append(c.watchedOwners, owner)
	copy(c.watchedOwners[i+1:], c.watchedOwners[i:])
	c.watchedOwners[i] = owner
	return &c.watchedOwners[i], nil
}

// Owner represents a GitHub owner and their tracked repositories
// Owners can specify default SLO rules that will apply to all tracked repos
// unless the repository overrides them with its own SLO rules config
type Owner struct {
	name     string
	Repos    []Repository
	SLORules []*SLORule
}

func (o *Owner) trackAllRepos(ctx context.Context, ghClient *githubreposervice.Client) error {
	o.Repos = []Repository{}

	var page int = 1
	for page > 0 {
		opt := &github.RepositoryListByOrgOptions{Sort: "full_name", ListOptions: github.ListOptions{Page: page, PerPage: 100}}
		repos, resp, err := ghClient.Repositories.ListByOrg(ctx, o.name, opt)
		if err != nil {
			return err
		}

		for _, r := range repos {
			o.Repos = append(o.Repos, Repository{name: *r.Name})
		}

		page = resp.NextPage
	}
	return nil
}

func (owner *Owner) trackRepo(ctx context.Context, repoName string, ghClient *githubreposervice.Client) error {
	repoIndex := sort.Search(len(owner.Repos), func(i int) bool { return owner.Repos[i].name >= repoName })

	if repoIndex < len(owner.Repos) && owner.Repos[repoIndex].name == repoName {
		log.Warningf("Repository %s/%s already tracked", owner.name, repoName)
		return nil
	}

	// check that repo exists:
	_, _, err := ghClient.Repositories.Get(ctx, owner.name, repoName)
	fmt.Println(err)
	if err != nil {
		log.Errorf("Unable to get repository %s/%s from GitHub: %s", owner.name, repoName, err)
		return err
	}

	addRepo := Repository{name: repoName}
	owner.Repos = append(owner.Repos, addRepo)
	copy(owner.Repos[repoIndex+1:], owner.Repos[repoIndex:])
	owner.Repos[repoIndex] = addRepo
	return nil
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

// ComplianceSettings stores data on the requirements for an issue or pull request to be considered compliant with the SLO
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
