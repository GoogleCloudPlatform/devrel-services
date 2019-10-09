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

package utils

import (
	"encoding/json"
	"strings"
)

// UnmarshalBool takes a json.RawMessage and deserializes it
// to a bool. If bytes is nil or, equal to "null" then
// will return nil
func UnmarshalBool(bytes json.RawMessage) *bool {
	if len(bytes) > 0 {
		s := string(bytes)
		s = strings.ToLower(s)
		if s != "null" {
			b := s == "true"
			return &b
		}
	}

	return nil
}
