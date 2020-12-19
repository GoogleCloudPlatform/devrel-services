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

package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"strings"
	"sync"

	"cloud.google.com/go/storage"
)

// NewBucketRepo returns a RepoList based on GCS Buckets
func NewBucketRepo(bucketName string, fileName string) RepoList {
	return &bucketRepoList{
		bucketName:    bucketName,
		reposFileName: fileName,
		reposList:     make([]TrackedRepository, 0),
	}
}

type bucketRepoList struct {
	bucketName    string
	reposFileName string
	reposList     []TrackedRepository
	mu            sync.RWMutex
}

func (r *bucketRepoList) UpdateTrackedRepos(ctx context.Context) (bool, error) {
	newRepos, err := r.getRepos(ctx)
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

func (r *bucketRepoList) GetTrackedRepos() []TrackedRepository {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.reposList
}

type bucketRepo struct {
	Repo               string `json:"repo"`
	DefaultBranch      string `json:"default_branch"`
	IsTrackingIssues   bool   `json:"is_tracking_issues"`
	IsTrackingSnippets bool   `json:"is_tracking_snippets"`
}

func (r *bucketRepoList) getRepos(ctx context.Context) ([]TrackedRepository, error) {
	// Creates a client.
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to create client: %v", err)
	}

	rc, err := client.Bucket(r.bucketName).Object(r.reposFileName).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get bucket: %v", err)
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("Failed to read repos: %v", err)
	}

	// Process data.

	var dat map[string][]bucketRepo

	if err := json.Unmarshal(data, &dat); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal repos")
	}

	repos := dat["repos"]

	reps := make([]TrackedRepository, len(repos))
	for i, re := range repos {
		parts := strings.Split(re.Repo, "/")
		if len(parts) != 2 {
			log.Printf("Bad format for repo %q", re.Repo)
			continue
		}

		tr := TrackedRepository{
			Owner:              parts[0],
			Name:               parts[1],
			IsTrackingIssues:   re.IsTrackingIssues,
			IsTrackingSnippets: re.IsTrackingSnippets,
			DefaultBranch:      re.DefaultBranch,
		}
		reps[i] = tr
	}

	return reps, nil
}
