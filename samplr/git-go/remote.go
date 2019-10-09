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
	"fmt"
	"io"
)

const (
	// RemoteOriginName is the name of the remote which is the "origin" of the repository
	RemoteOriginName = "origin"
)

// RemoteConfig contains the configuration for a given remote repository.
type RemoteConfig struct {
	// Name of the remote
	Name string
	// URLs the URLs of a remote repository. It must be non-empty. Fetch will
	// always use the first URL, while push will use all of them.
	URLs []string
	// Fetch the default set of "refspec" for fetch operation
	// Fetch []RefSpec
}

// Remote represents a connection to a remote repository.
type Remote struct {
	c *RemoteConfig
}

// Config returns the RemoteConfig object used to instantiate this Remote.
func (r *Remote) Config() *RemoteConfig {
	if r == nil {
		return nil
	}
	return r.c
}

func (r *Remote) String() string {
	var fetch, push string
	if len(r.c.URLs) > 0 {
		fetch = r.c.URLs[0]
		push = r.c.URLs[0]
	}

	return fmt.Sprintf("%s\t%s (fetch)\n%[1]s\t%[3]s (push)", r.c.Name, fetch, push)
}

// RemoteIter is a generic closable interface for iterating over Remotes.
type RemoteIter interface {
	Next() (*Remote, error)
	ForEach(func(*Remote) error) error
	Close()
}

type sliceRemoteIter struct {
	pos    int
	series []Remote
}

func (iter *sliceRemoteIter) Next() (*Remote, error) {
	if iter.pos >= len(iter.series) {
		return nil, io.EOF
	}

	obj := iter.series[iter.pos]
	iter.pos++
	return &obj, nil
}

func (iter *sliceRemoteIter) Close() {
	iter.pos = len(iter.series)
}

func (iter *sliceRemoteIter) ForEach(fn func(*Remote) error) error {
	defer iter.Close()
	for _, r := range iter.series {
		if err := fn(&r); err != nil {
			if err == ErrStop {
				return nil
			}
			return err
		}
	}
	return nil
}
