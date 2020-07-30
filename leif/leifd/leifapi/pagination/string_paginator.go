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

package paginator

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type stringPage struct {
	items []string
	index int
}

// Strings is a paginator for strings
type Strings struct {
	Log     *logrus.Logger
	mu      sync.Mutex
	set     map[time.Time]stringPage
	didInit bool
}

// Init must be the first call to the paginator
func (p *Strings) Init() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.didInit {
		return fmt.Errorf("Paginator already initialized")
	}
	p.set = make(map[time.Time]stringPage)
	p.didInit = true
	return nil
}

// PurgeOldRecords removes all pages that are more than 2 hours old
func (p *Strings) PurgeOldRecords() {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	for t := range p.set {
		if now.Sub(t).Hours() > 2 {
			delete(p.set, t)
		}
	}
}

// CreatePage makes a new page in the paginator
// It uses the current time as key and add the items to that key
func (p *Strings) CreatePage(withItems []string) (time.Time, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.didInit {
		return time.Unix(0, 0), fmt.Errorf("Paginator not initialized")
	}
	key := time.Now().UTC().Truncate(0)
	if _, ok := p.set[key]; ok {
		return time.Unix(0, 0), errors.New("Key already exists")
	}

	p.set[key] = stringPage{
		items: withItems,
		index: 0,
	}
	return key, nil
}

// GetPage gets the next numItems number of items from the given page/key
// Key should be the key returned by a call to CreatePage
// GetPage returns the items and the current index in the total items in the page
func (p *Strings) GetPage(key time.Time, numItems int) ([]string, int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.didInit {
		return nil, 0, fmt.Errorf("Paginator not initialized")
	}

	key = key.UTC()

	if _, ok := p.set[key]; !ok {
		return nil, 0, fmt.Errorf("Page key: %v not found", key)
	}
	val := p.set[key]

	numItemsRemaining := len(val.items) - val.index
	if p.Log != nil {
		p.Log.Debugf("There are %v records remaining, requested %v", numItemsRemaining, numItems)
	}

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
