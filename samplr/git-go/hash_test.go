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

import "testing"

func TestHashIsZero(t *testing.T) {
	hash := NewHash("f3f4e32b94c97ffd6b2e6f020d96931a9fe2c3f6")
	if hash.IsZero() {
		t.Errorf("Hash: %v was said to be the Zero Hash, when it is not", hash.String())
	}
	if !ZeroHash.IsZero() {
		t.Errorf("The ZeroHas was said to not be the Zero Hash")
	}
}

func TestHashString(t *testing.T) {
	hash := NewHash("f3f4e32b94c97ffd6b2e6f020d96931a9fe2c3f6")
	hashStr := hash.String()
	if hashStr != "f3f4e32b94c97ffd6b2e6f020d96931a9fe2c3f6" {
		t.Errorf("Hash String. Expected %v, got %v", "f3f4e32b94c97ffd6b2e6f020d96931a9fe2c3f6", hashStr)
	}
}
