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

package output

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/GoogleCloudPlatform/devrel-services/samplr/samplrctl/testutil"
)

const (
	field1 = "Field 1"
	field2 = "Field 2"
	field3 = "Field 3"
)

var (
	object1 = map[string]string{
		field1: "Value 1A",
		field2: "Value 2A",
		field3: "Value 3A",
	}

	object2 = map[string]string{
		field1: "Value 1B",
		field2: "Value 2B",
		field3: "Value 3B",
	}
)

func TestPrintMap(t *testing.T) {
	var b bytes.Buffer
	PrintMap(&b, []string{field1, field3}, object1)
	s := b.String()

	want := []string{field1, object1[field1], field3, object1[field3]}
	if !testutil.ContainsAll(s, want...) {
		t.Errorf("Printed: %q; missing: %v", s, want)
	}

	dontWant := []string{field2, object1[field2]}
	if testutil.ContainsAll(s, dontWant...) {
		t.Errorf("Printed: %q; don't want: %v:", s, dontWant)
	}
}

func TestPrintAllMap(t *testing.T) {
	var b bytes.Buffer
	PrintAllMap(&b, object1)
	s := b.String()

	want := []string{field1, object1[field1], field2, object1[field2], field3, object1[field3]}
	if !testutil.ContainsAll(s, want...) {
		t.Errorf("Printed: %q; missing: %v", s, want)
	}
}

func TestPrintList(t *testing.T) {
	var b bytes.Buffer
	PrintList(&b, []string{field1, field3}, []map[string]string{object1, object2})
	s := b.String()

	wantRows := [][]string{
		[]string{field1, field3},
		[]string{object1[field1], object1[field3]},
		[]string{object2[field1], object2[field3]},
	}

	for _, wantRow := range wantRows {
		if b, _ := testutil.ContainsRow(s, wantRow...); !b {
			t.Errorf("Printed: %q; missing: %v", s, wantRow)
		}
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"string", "string"},
		{"String", "string"},
		{"With Spaces", "withSpaces"},
		{"CAPS", "caps"},
		{"Spaces AND caps", "spacesAndCaps"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s / %s", tt.in, tt.want), func(t *testing.T) {
			if got := toCamelCase(tt.in); got != tt.want {
				t.Errorf("Got %q; want %q", got, tt.want)
			}
		})
	}
}
