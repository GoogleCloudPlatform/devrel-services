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

package tokens

import (
	"fmt"
	"sync"
)

type RotatingTokens struct {
	keys         []string
	mux          sync.Mutex
	lastAccessed int
}

func NewRotatingVendor(tokens []string) *RotatingTokens {
	if tokens == nil {
		tokens = make([]string, 0)
	}
	return &RotatingTokens{keys: tokens}
}

func (r *RotatingTokens) GetToken() (string, error) {
	r.mux.Lock()
	defer r.mux.Unlock()
	if r.lastAccessed >= len(r.keys) {
		r.lastAccessed = 0
	}
	if len(r.keys) == 0 {
		return "", fmt.Errorf("no tokens")
	}
	key := r.keys[r.lastAccessed]
	r.lastAccessed++
	return key, nil
}
