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
	"reflect"
	"sync"

	"cloud.google.com/go/storage"
)

func NewBucketRepo(bucketName string, fileName string) *bucketRepoList {
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

	 var dat map[string]interface{}

	if err := json.Unmarshal(data, &dat); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal repos")
	}

	// TODO(colnnelson): Update this to unmarshal the struct directly into
	// a useable object.
	repos := dat["repos"].([]interface{})

	reps := make([]TrackedRepository, len(repos))

	for idx, repoDat := range repos {
	    if err := json.Unmarshal(repoDat.([]byte), &reps[idx]); err != nil {
			return nil, fmt.Errorf("Failed to unmarshal repo data")
		}
	}

	return reps, nil
}
