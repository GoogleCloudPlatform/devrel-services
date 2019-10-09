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

package testutil

import (
	"testing"
)

func TestContainsAll(t *testing.T) {
	tests := []struct {
		searchString string
		substrings   []string
		result       bool
	}{
		{"abba", []string{"a", "b", "ab", "ba"}, true},
		{"abba", []string{"a", "b", "ab", "ba", "baab"}, false},
		{"abba", []string{"z"}, false},
	}
	for _, tc := range tests {
		if got := ContainsAll(tc.searchString, tc.substrings...); got != tc.result {
			t.Errorf("ContainsAll(%q, %q) = %v; want %v", tc.searchString, tc.substrings, got, tc.result)
		}
	}
}

func TestContainsAny(t *testing.T) {
	tests := []struct {
		searchString string
		substrings   []string
		result       bool
	}{
		{"abba", []string{"a", "b", "ab", "ba"}, true},
		{"abba", []string{"a", "b", "ab", "ba", "baab"}, true},
		{"abba", []string{"z"}, false},
	}
	for _, tc := range tests {
		if got := ContainsAny(tc.searchString, tc.substrings...); got != tc.result {
			t.Errorf("ContainsAny(%q, %q) = %v; want %v", tc.searchString, tc.substrings, got, tc.result)
		}
	}
}

func TestContainsRow(t *testing.T) {
	tests := []struct {
		searchString string
		substrings   []string
		result       bool
	}{
		{"[0]\tvalue0\n[1]\tvalue1\n[2]\tvalue2", []string{"[0]", "value0"}, true},
		{"[0]\tvalue0\n[1]\tvalue1\n[2]\tvalue2", []string{"[1]", "value1"}, true},
		{"[0]\tvalue0\n[1]\tvalue1\n[2]\tvalue2", []string{"[2]", "value2"}, true},
		{"[0]\twrong\n[1]\tvalue1\n[2]\tvalue2", []string{"[0]", "value0"}, false},
		{"[0]\tvalue0\n[1]\twrong\n[2]\tvalue2", []string{"[1]", "value1"}, false},
		{"[0]\tvalue0\n[1]\tvalue1\n[2]\twrong", []string{"[2]", "value2"}, false},
		{"value0\t[0]\n[1]\tvalue1\n[2]\tvalue2", []string{"[0]", "value0"}, false},
		{"[0]\tvalue0\nvalu1\t[1]\n[2]\tvalue2", []string{"[1]", "value2"}, false},
		{"[0]\tvalue0\n[1]\tvalue1\nvalue2\t[2]", []string{"[2]", "value2"}, false},
	}
	for _, tc := range tests {
		got, err := ContainsRow(tc.searchString, tc.substrings...)
		if err != nil {
			t.Errorf("Got error for %v: %v", tc.searchString, tc.substrings)
		} else if got != tc.result {
			t.Errorf("ContainsRow(%q, %q) = %v; want %v", tc.searchString, tc.substrings, got, tc.result)
		}
	}
}

func TestAnyLineContainsAll(t *testing.T) {
	tests := []struct {
		searchString string
		substrings   []string
		result       bool
	}{
		{"foo\nabba", []string{"a", "b", "ab", "ba"}, true},
		{"foo\nabba\nbar", []string{"a", "b", "ab", "ba"}, true},
		{"abba\nbar", []string{"a", "b", "ab", "ba"}, true},
		{"foo\nabba", []string{"a", "b", "ab", "ba", "baab"}, false},
		{"foo\nabba\nbar", []string{"a", "b", "ab", "ba", "baab"}, false},
		{"abba\nbar", []string{"a", "b", "ab", "ba", "baab"}, false},
		{"foo\nabba", []string{"z"}, false},
		{"foo\nabba\nbar", []string{"z"}, false},
		{"abba\nbar", []string{"z"}, false},
	}
	for _, tc := range tests {
		if got := AnyLineContainsAll(tc.searchString, tc.substrings...); got != tc.result {
			t.Errorf("AnyLineContainsAll(%q, %q) = %v; want %v", tc.searchString, tc.substrings, got, tc.result)
		}
	}
}
