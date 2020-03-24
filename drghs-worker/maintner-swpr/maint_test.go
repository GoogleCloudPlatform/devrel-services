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

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"
	"github.com/GoogleCloudPlatform/devrel-services/repos"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/time/rate"
)

func TestRepoToTrackedRepo(t *testing.T) {
	cs := []struct {
		Repo drghs_v1.Repository
		Want *repos.TrackedRepository
		Name string
	}{
		{
			Name: "Expected",
			Repo: drghs_v1.Repository{
				Name: "GoogleCloudPlatform/devrel-services",
			},
			Want: &repos.TrackedRepository{
				Owner: "GoogleCloudPlatform",
				Name:  "devrel-services",
			},
		},
		{
			Name: "Failure",
			Repo: drghs_v1.Repository{
				Name: "owners/GoogleCloudPlatform/repos/devrel-services",
			},
			Want: nil,
		},
	}
	for _, c := range cs {
		got := repoToTrackedRepo(&c.Repo)
		if diff := cmp.Diff(c.Want, got); diff != "" {
			t.Errorf("Test: %v Repositories differ (-want +got)\n%s", c.Name, diff)
		}
	}
}

func TestBuildLimiter(t *testing.T) {
	cases := []struct {
		Name      string
		NIssues   int32
		WantLimit rate.Limit
	}{
		{
			Name:      "LowFrequency",
			NIssues:   864000,
			WantLimit: 0.1,
		},
		{
			Name:      "HighFrequency",
			NIssues:   86400000,
			WantLimit: 10,
		},
		{
			Name:      "EvenFrequency",
			NIssues:   8640000,
			WantLimit: 1,
		},
	}
	for _, c := range cases {

		limiter := buildLimiter(c.NIssues)
		gotLimit := limiter.Limit()
		if gotLimit != c.WantLimit {
			t.Errorf("test: %v failed. limiter improperly set. got: %v want: %v", c.Name, gotLimit, c.WantLimit)
		}
	}
}
