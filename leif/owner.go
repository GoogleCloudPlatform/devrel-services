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

// This package was modeled after the mocking strategy outlined at:
// https://github.com/google/go-github/issues/113#issuecomment-46023864

package leif

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/GoogleCloudPlatform/devrel-services/leif/githubservices"
)

// ErrRepoAlreadyTracked is a sentinel error which indicates the repository is already tracked by the owner
var ErrRepoAlreadyTracked = errors.New("Already tracked")

type repoAlreadyTrakedError struct {
	owner string
	repo  string
	err   error
}

func (e *repoAlreadyTrakedError) Error() string {
	return fmt.Sprintf("The repository %v/%v was already tracked", e.owner, e.repo)
}

func (e *repoAlreadyTrakedError) Unwrap() error {
	return e.err
}

// Owner represents a GitHub owner and their tracked repositories
// Owners can specify default SLO rules that will apply to all tracked repos
// unless the repository overrides them with its own SLO rules config
type Owner struct {
	name     string
	Repos    []*Repository
	SLORules []*SLORule
}

// Name returns the name of the owner
func (o *Owner) Name() string {
	return o.name
}

// UpdateLoop updates the owner and their tracked repositories every given amount of minutes
func (o *Owner) UpdateLoop(ctx context.Context, minutes int, ghClient *githubservices.Client) error {
	log.Printf("Beginning sync loop for owner %s", o.name)

	ticker := time.NewTicker(time.Duration(minutes) * time.Minute)
	for {
		select {
		case <-ctx.Done():
			err := fmt.Errorf("Context cancelled")
			log.Printf("Ended sync loop for owner %s: %v", o.name, err)
			return err
		case <-ticker.C:
			err := o.Update(ctx, ghClient)
			if err != nil {
				log.Printf("Ended sync loop for owner %s: %v", o.name, err)
				return err
			}
		}
	}
}

// Update reaches out to GitHub to update the SLO rules for the owner,
// then updates the SLO rules for each tracked repository under the owner
func (o *Owner) Update(ctx context.Context, ghClient *githubservices.Client) error {
	// update overarching config
	rules, err := findSLODoc(ctx, *o, "", ghClient)
	if err != nil {
		var ghErrorResponse *goGitHubErr

		if !(errors.As(err, &ghErrorResponse) && ghErrorResponse.Response.StatusCode == 404) {
			// any error other than config not found returns
			return err
		}
	}

	o.SLORules = rules

	// update the repos under it
	for _, r := range o.Repos {
		err = r.Update(ctx, *o, ghClient)
		if err != nil {
			log.Errorf("Error updating repository %s/%s: %s", o.name, r.name, err)
		}
	}

	return nil
}

func (o *Owner) trackRepo(ctx context.Context, repoName string, ghClient *githubservices.Client) error {
	repoIndex := sort.Search(len(o.Repos), func(i int) bool { return o.Repos[i].name >= repoName })

	if repoIndex < len(o.Repos) && o.Repos[repoIndex].name == repoName {
		return &repoAlreadyTrakedError{repo: repoName, owner: o.name, err: ErrRepoAlreadyTracked}
	}

	// check that repo exists:
	_, _, err := ghClient.Repositories.Get(ctx, o.name, repoName)
	if err != nil {
		return err
	}

	addRepo := Repository{name: repoName, ownerName: o.name}
	o.Repos = append(o.Repos, &addRepo)
	copy(o.Repos[repoIndex+1:], o.Repos[repoIndex:])
	o.Repos[repoIndex] = &addRepo

	return nil
}
