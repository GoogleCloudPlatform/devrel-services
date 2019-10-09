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

package git

import (
	"io"
	"time"
)

// Commit stores data about a Git Commit
type Commit struct {
	// Hash of the commit object.
	Hash Hash
	// Author is the original author of the commit.
	Author Signature
	// Committer is the one performing the commit, might be different from
	// Author.
	Committer Signature
	// Message is the commit message, contains arbitrary text.
	Message string
	files   []File
}

// Files returns a FilesIter representing the Commit's files
func (c Commit) Files() (FilesIter, error) {
	iter := &sliceFileIter{
		pos:    0,
		series: c.files,
	}
	return iter, nil
}

// Signature is used to identify who and when created a commit or tag.
type Signature struct {
	// Name represents a person name. It is an arbitrary string.
	Name string
	// Email is an email, but it cannot be assumed to be well-formed.
	Email string
	// When is the timestamp of the signature.
	When time.Time
}

// CommitIter is a generic closable interface for iterating over commits.
type CommitIter interface {
	Next() (*Commit, error)
	ForEach(func(*Commit) error) error
	Close()
}

// sliceCommitIter iterates over an internaly held slice
type sliceCommitIter struct {
	pos    int
	series []*Commit
}

func (iter *sliceCommitIter) Next() (*Commit, error) {
	if iter.pos >= len(iter.series) {
		return nil, io.EOF
	}

	obj := iter.series[iter.pos]
	iter.pos++
	return obj, nil
}

func (iter *sliceCommitIter) ForEach(fn func(*Commit) error) error {
	defer iter.Close()
	for _, r := range iter.series {
		if err := fn(r); err != nil {
			if err == ErrStop {
				return nil
			}

			return err
		}
	}
	return nil
}

func (iter *sliceCommitIter) Close() {
	iter.pos = len(iter.series)
}
