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

// Package testutil provides common functionality for tests across the CLI.
//
// NOTE: this package is only for generic utilities that don't have dependencies on other packages
// in the CLI. For utilities specific to another CLI package foo, create a foo/test package.
package testutil

import (
	"regexp"
	"strings"
)

// ContainsAll validates all provided substrs are contained in s.
func ContainsAll(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if !strings.Contains(s, substr) {
			return false
		}
	}
	return true
}

// ContainsAny validates any of the provided substrs is contained in s.
func ContainsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// ContainsRow validates that all provided fields appear, in order, in a tab-separated row in s.
func ContainsRow(s string, fields ...string) (bool, error) {
	pattern := "(?m:^[\t ]*"
	for i, field := range fields {
		if i != 0 {
			pattern += `[\t ]+`
		}
		pattern += regexp.QuoteMeta(field)
	}
	pattern += "[\t ]*$)"
	r, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	return r.FindString(s) != "", nil
}

// AnyLineContainsAll validates that some line of the output contains all the provided substrings.
func AnyLineContainsAll(output string, substrs ...string) bool {
	for _, line := range strings.Split(output, "\n") {
		if ContainsAll(line, substrs...) {
			return true
		}
	}
	return false
}
