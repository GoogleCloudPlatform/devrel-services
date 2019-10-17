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
	"github.com/GoogleCloudPlatform/devrel-services/drghs-worker/pkg/utils"
	"encoding/json"
	"testing"
)

func TestUnmarshalBool(t *testing.T) {
	tests := []struct {
		message  json.RawMessage
		expected bool
		isNil    bool
	}{
		{
			message:  nil,
			expected: false,
			isNil:    true,
		},
		{
			message:  []byte("null"),
			expected: false,
			isNil:    true,
		},
		{
			message:  []byte("false"),
			expected: false,
			isNil:    false,
		},
		{
			message:  []byte("TRUE"),
			expected: true,
			isNil:    false,
		},
		{
			message:  []byte("true"),
			expected: true,
			isNil:    false,
		},
		{
			message:  []byte("Random"),
			expected: false,
			isNil:    false,
		},
	}
	for _, test := range tests {
		val := utils.UnmarshalBool(test.message)

		if test.isNil && val != nil {
			t.Errorf("Unmarshal Bool for %v returned an unexpected value. Got %v, want nil", string(test.message), val)
		}

		if !test.isNil && val == nil {
			t.Errorf("Unmarshal Bool for %v returned an unexpected nil. Got nil, want %v", string(test.message), test.expected)
		}

		if val != nil && test.expected != *val {
			t.Errorf("Unmarshal Bool for %v returned an unexpected value. Got %v, want %v", string(test.message), *val, test.expected)
		}
	}
}
