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
)

func TestCalculateHost(t *testing.T) {
	tests := []struct {
		Name    string
		Input   string
		Want    string
		WantErr error
	}{
		{
			Name:    "Success",
			Input:   "foo/bar",
			Want:    "mtr-s-9bec779ac30c3c91e5e8055c0f5cb25f39ed7ce0ecb5f9a8f64fdab0:80",
			WantErr: nil,
		},
		{
			Name:    "Handles '.' in repo",
			Input:   "foo/bar.io",
			Want:    "mtr-s-cf8092de84b0aecbc785168376cc27fe92962329a7ce274b99e6b867:80",
			WantErr: nil,
		},
		{
			Name:    "Handles '.' in owner",
			Input:   "foo.com/bar",
			Want:    "mtr-s-44610f96da7a230fc7b41ee569ca88d47ff9cf2c723c936ca7a84014:80",
			WantErr: nil,
		},
		{
			Name:    "Handles '.' in both",
			Input:   "foo.com/bar.io",
			Want:    "mtr-s-643d239d7e9b88b103a918eb67af7f276a1020538b101cc4bcbd0c1b:80",
			WantErr: nil,
		},
		{
			Name:    "Handles ':'",
			Input:   "foo/foo:bar",
			Want:    "mtr-s-98260423ee830fd047446b610c3ea2a5e0246de54e0c3ddc49e5e9f7:80",
			WantErr: nil,
		},
		{
			Name:    "Handles deep names",
			Input:   "foo/foo:bar/baz/biz",
			Want:    "mtr-s-98260423ee830fd047446b610c3ea2a5e0246de54e0c3ddc49e5e9f7:80",
			WantErr: nil,
		},
		{
			Name:    "Handles invalid",
			Input:   "/foo/bar",
			Want:    devnull,
			WantErr: nil,
		},
	}
	for _, c := range tests {
		got, gotErr := calculateHost(c.Input)
		if gotErr != c.WantErr {
			t.Errorf("%v Errors Differ. Want %v. Got %v", c.Name, c.WantErr, gotErr)
		}
		if c.Want != got {
			t.Errorf("%v Outputs Differ. Want %v. Got %v", c.Name, c.Want, got)
		}
	}
}
