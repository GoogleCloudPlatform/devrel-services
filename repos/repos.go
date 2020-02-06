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
	"crypto/sha256"
	"fmt"
)

type RepoList interface {
	// UpdateTrackedRepos, updates the list of trakced repositories, returns
	// true if the list changed, false otherwise
	UpdateTrackedRepos(context.Context) (bool, error)
	GetTrackedRepos() []TrackedRepository
}

type TrackedRepository struct {
	Owner              string `json:"owner"`
	Name               string `json:"name"`
	IsTrackingIssues   bool   `json:"isTrackingIssues"`
	IsTrackingSnippets bool   `json:"isTrackingSnippets"`
}

// RepoSha Creates a Sum224 of the TrackedRepository's name
func (t TrackedRepository) RepoSha() string {
	sh := sha256.Sum224([]byte(fmt.Sprintf("%v/%v", t.Owner, t.Name)))
	return fmt.Sprintf("%x", sh)
}

// String returns the string representation of the TrackedRepository
func (t TrackedRepository) String() string {
	return fmt.Sprintf("%v/%v", t.Owner, t.Name)
}
