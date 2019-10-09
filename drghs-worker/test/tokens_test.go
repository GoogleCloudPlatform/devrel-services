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

package main

import (
	"devrel/cloud/devrel-github-service/drghs-worker/pkg/tokens"
	"fmt"
	"testing"
)

func TestGetstokens(t *testing.T) {
	tests := []struct {
		tokens []string
		err    error
	}{
		{
			tokens: []string{"One", "Two", "Three"},
			err:    nil,
		},
		{
			tokens: []string{"One"},
			err:    nil,
		},
		{
			tokens: []string{},
			err:    fmt.Errorf("no tokens"),
		},
	}
	for _, test := range tests {
		store := tokens.NewRotatingVendor(test.tokens)

		for i := 0; i < len(test.tokens)*2; i++ {
			token, err := store.GetToken()
			if err != test.err {
				t.Errorf("Get token returned an unexpected error. Got %v, want %v", err, test.err)
			}
			if token != test.tokens[i%len(test.tokens)] {
				t.Errorf("GetKy returned an unexpected token. Got %v, want %v", token, test.tokens[i%len(test.tokens)])
			}
		}

		if len(test.tokens) == 0 {
			_, err := store.GetToken()
			if err.Error() != test.err.Error() {
				t.Errorf("Get token returned an unexpected error. Got %v, want %v", err, test.err)
			}
		}
	}
}
