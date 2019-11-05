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

package samplrapi

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type trackedRepositoryPage struct {
	trackedrepos []string
	idx          int
}

type trackedRepositoryPaginator struct {
	set map[time.Time]trackedRepositoryPage
	mu  sync.Mutex
}

func (p *trackedRepositoryPaginator) PurgeOldRecords() {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	for t := range p.set {
		if now.Sub(t).Hours() > 2 {
			delete(p.set, t)
		}
	}
}

func (p *trackedRepositoryPaginator) CreatePage(s []string) (time.Time, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	key := time.Now().UTC().Truncate(0)
	if _, ok := p.set[key]; ok {
		return time.Unix(0, 0), errors.New("Key already exists")
	}

	p.set[key] = trackedRepositoryPage{
		trackedrepos: s,
		idx:          0,
	}
	return key, nil
}

func (p *trackedRepositoryPaginator) GetPage(key time.Time, n int) ([]string, int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key = key.UTC()
	if _, ok := p.set[key]; !ok {
		return nil, 0, fmt.Errorf("Page key: %v not found", key)
	}
	val := p.set[key]

	nremain := len(val.trackedrepos) - val.idx
	log.Debugf("There are %v records remaining, requested %v", nremain, n)

	if n > nremain {
		n = nremain
	}

	if n == 0 {
		return nil, 0, errors.New("Get 0 from page")
	}

	retset := val.trackedrepos[val.idx:(val.idx + n)]
	val.idx = val.idx + n

	retidx := val.idx
	if val.idx == len(val.trackedrepos) {
		delete(p.set, key)
		retidx = -1
	} else {
		p.set[key] = val
	}

	return retset, retidx, nil
}
