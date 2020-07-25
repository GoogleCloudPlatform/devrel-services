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

	"github.com/GoogleCloudPlatform/devrel-services/leif"
)

type sloPage struct {
	items []*leif.SLORule
	index int
}

type sloPaginator struct {
	set map[time.Time]sloPage
	mu  sync.Mutex
}

func (p *sloPaginator) PurgeOldRecords() {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	for t := range p.set {
		if now.Sub(t).Hours() > 2 {
			delete(p.set, t)
		}
	}
}

func (p *sloPaginator) CreatePage(withItems []*leif.SLORule) (time.Time, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	key := time.Now().UTC().Truncate(0)
	if _, ok := p.set[key]; ok {
		return time.Unix(0, 0), errors.New("Key already exists")
	}

	p.set[key] = sloPage{
		items: withItems,
		index: 0,
	}
	return key, nil
}

func (p *sloPaginator) GetPage(key time.Time, numItems int) ([]*leif.SLORule, int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key = key.UTC()
	if _, ok := p.set[key]; !ok {
		return nil, 0, fmt.Errorf("Page key: %v not found", key)
	}
	val := p.set[key]

	numItemsRemaining := len(val.items) - val.index
	log.Debugf("There are %v records remaining, requested %v", numItemsRemaining, numItems)

	if numItems > numItemsRemaining {
		numItems = numItemsRemaining
	}

	if numItems == 0 {
		return nil, 0, errors.New("Get 0 from page")
	}

	retSet := val.items[val.index:(val.index + numItems)]
	val.index = val.index + numItems

	retIndex := val.index
	if val.index == len(val.items) {
		delete(p.set, key)
		retIndex = -1
	} else {
		p.set[key] = val
	}

	return retSet, retIndex, nil
}
