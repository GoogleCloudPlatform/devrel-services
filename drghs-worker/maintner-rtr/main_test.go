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
	"testing"

	"github.com/GoogleCloudPlatform/devrel-services/repos"
	"github.com/google/go-cmp/cmp"
)

func TestCalculateHost(t *testing.T) {
	tests := []struct {
		Name  string
		Input string
		Want  *repos.TrackedRepository
	}{
		{
			Name:  "Success",
			Input: "foo/bar",
			Want: &repos.TrackedRepository{
				Owner: "foo",
				Name:  "bar",
			},
		},
		{
			Name:  "Handles '.' in repo",
			Input: "foo/bar.io",
			Want: &repos.TrackedRepository{
				Owner: "foo",
				Name:  "bar.io",
			},
		},
		{
			Name:  "Handles '.' in owner",
			Input: "foo.com/bar",
			Want: &repos.TrackedRepository{
				Owner: "foo.com",
				Name:  "bar",
			},
		},
		{
			Name:  "Handles '.' in both",
			Input: "foo.com/bar.io",
			Want: &repos.TrackedRepository{
				Owner: "foo.com",
				Name:  "bar.io",
			},
		},
		{
			Name:  "Handles ':'",
			Input: "foo/foo:bar",
			Want: &repos.TrackedRepository{
				Owner: "foo",
				Name:  "foo:bar",
			},
		},
		{
			Name:  "Handles deep names",
			Input: "foo/foo:bar/baz/biz",
			Want: &repos.TrackedRepository{
				Owner: "foo",
				Name:  "foo:bar",
			},
		},
		{
			Name:  "Handles invalid",
			Input: "/foo/bar",
			Want:  nil,
		},
	}
	for _, c := range tests {
		got := buildTR(c.Input)
		if diff := cmp.Diff(c.Want, got); diff != "" {
			t.Errorf("test: %v, values diff. match (-want +got)\n%s", c.Name, diff)
		}
	}
}
