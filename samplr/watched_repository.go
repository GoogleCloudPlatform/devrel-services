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

import "context"

// WatchedRepository represents a repository being watched by the Corpus
type WatchedRepository interface {
	// Unique Identifier of the Repository
	// TODO(colnnelson): Work out the details of this
	ID() string
	// Allows iterating over a WatchedRepository's snippets
	ForEachSnippet(func(snippet *Snippet) error) error
	// Allows iterating over a WatchedRepository's snippets that match the given filter
	ForEachSnippetF(func(snippet *Snippet) error, func(snippet *Snippet) bool) error
	// Allows iterating over a WatchedRepository's git commits
	ForEachGitCommit(func(commit *GitCommit) error) error
	// Allows iterating over a WatchedRepository's git commits that match the given filter
	ForEachGitCommitF(func(commit *GitCommit) error, func(commit *GitCommit) bool) error
	// The owner of the repository
	Owner() string
	// The name of the repository
	RepositoryName() string
	// Instructs the Repository to Update
	Update(ctx context.Context) error
}
