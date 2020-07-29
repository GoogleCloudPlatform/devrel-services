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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"sync"

	"github.com/GoogleCloudPlatform/devrel-services/repos"
)

// NewDiskRepo returns a RepoList based on a local file
func NewDiskRepo(fileName string) repos.RepoList {
	return &diskRepoList{
		fileName:  fileName,
		reposList: make([]repos.TrackedRepository, 0),
	}
}

type diskRepoList struct {
	fileName  string
	reposList []repos.TrackedRepository
	mu        sync.RWMutex
}

func (r *diskRepoList) UpdateTrackedRepos(ctx context.Context) (bool, error) {
	newRepos, err := r.getRepos()
	if err != nil {
		return false, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if reflect.DeepEqual(newRepos, r.reposList) {
		return false, nil
	}

	r.reposList = newRepos
	return true, nil
}

func (r *diskRepoList) GetTrackedRepos() []repos.TrackedRepository {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.reposList
}

type diskRepo struct {
	Repo               string `json:"repo"`
	IsTrackingIssues   bool   `json:"is_tracking_issues"`
	IsTrackingSnippets bool   `json:"is_tracking_snippets"`
}

func (r *diskRepoList) getRepos() ([]repos.TrackedRepository, error) {
	file, err := ioutil.ReadFile(r.fileName)
	if err != nil {
		return nil, err
	}

	// Parse
	var data map[string][]diskRepo

	if err := json.Unmarshal(file, &data); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal repos: %w", err)
	}

	fileRepos := data["repos"]

	convertedRepos := make([]repos.TrackedRepository, len(fileRepos))
	for i, repo := range fileRepos {
		repoPath := strings.Split(repo.Repo, "/")
		if len(repoPath) != 2 {
			log.Printf("Bad format for repo %q", repo.Repo)
			continue
		}

		convertedRepos[i] = repos.TrackedRepository{
			Owner:              repoPath[0],
			Name:               repoPath[1],
			IsTrackingIssues:   repo.IsTrackingIssues,
			IsTrackingSnippets: repo.IsTrackingSnippets,
		}
	}

	return convertedRepos, nil
}
