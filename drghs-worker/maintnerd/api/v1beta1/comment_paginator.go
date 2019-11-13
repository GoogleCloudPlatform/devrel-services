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

package v1beta1

import (
	"errors"
	"fmt"
	"sync"
	"time"

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"
)

type commentPage struct {
	iss []*drghs_v1.GitHubComment
	idx int
}

type commentPaginator struct {
	set map[time.Time]commentPage
	mu  sync.Mutex
}

func (p *commentPaginator) PurgeOldRecords() {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	for t := range p.set {
		if now.Sub(t).Hours() > nHoursStale {
			delete(p.set, t)
		}
	}
}

func (p *commentPaginator) CreatePage(s []*drghs_v1.GitHubComment) (time.Time, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	key := time.Now().UTC().Truncate(0)
	if _, ok := p.set[key]; ok {
		return time.Unix(0, 0), errors.New("Key already exists")
	}

	p.set[key] = commentPage{
		iss: s,
		idx: 0,
	}
	return key, nil
}

func (p *commentPaginator) GetPage(key time.Time, n int) ([]*drghs_v1.GitHubComment, int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key = key.UTC()
	if _, ok := p.set[key]; !ok {
		return nil, 0, fmt.Errorf("Page key: %v not found", key)
	}
	val := p.set[key]

	nremain := len(val.iss) - val.idx

	if n > nremain {
		n = nremain
	}

	if n == 0 {
		return []*drghs_v1.GitHubComment{}, -1, nil
	}

	retset := val.iss[val.idx:(val.idx + n)]
	val.idx = val.idx + n

	retidx := val.idx
	if val.idx == len(val.iss) {
		delete(p.set, key)
		retidx = -1
	} else {
		p.set[key] = val
	}

	return retset, retidx, nil
}
