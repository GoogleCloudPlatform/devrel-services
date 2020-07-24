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

package leifapi

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type repositoryPage struct {
	repos []string
	index int
}

type repositoryPaginator struct {
	set map[time.Time]repositoryPage
	mu  sync.Mutex
}

func (p *repositoryPaginator) PurgeOldRecords() {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	for t := range p.set {
		if now.Sub(t).Hours() > 2 {
			delete(p.set, t)
		}
	}
}

func (p *repositoryPaginator) CreatePage(s []string) (time.Time, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	key := time.Now().UTC().Truncate(0)
	if _, ok := p.set[key]; ok {
		return time.Unix(0, 0), errors.New("Key already exists")
	}

	p.set[key] = repositoryPage{
		repos: s,
		index: 0,
	}
	return key, nil
}

func (p *repositoryPaginator) GetPage(key time.Time, numRepos int) ([]string, int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key = key.UTC()
	if _, ok := p.set[key]; !ok {
		return nil, 0, fmt.Errorf("Page key: %v not found", key)
	}
	val := p.set[key]

	numReposRemaining := len(val.repos) - val.idx
	log.Debugf("There are %v records remaining, requested %v", numReposRemaining, numRepos)

	if numRepos > numReposRemaining {
		numRepos = numReposRemaining
	}

	if numRepos == 0 {
		return nil, 0, errors.New("Get 0 from page")
	}

	retSet := val.repos[val.index:(val.index + numRepos)]
	val.index = val.index + numRepos

	retIndex := val.index
	if val.index == len(val.repos) {
		delete(p.set, key)
		retIndex = -1
	} else {
		p.set[key] = val
	}

	return retSet, retIndex, nil
}
