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

import (
	"testing"
)

func TestDetectsNew(t *testing.T) {
	cases := []struct {
		Name    string
		Str     string
		Want    bool
		WantStr string
	}{
		{Name: "Valid", Str: "A	foo", Want: true, WantStr: "foo"},
		{Name: "Valid With Extension", Str: "A	foo.bar", Want: true, WantStr: "foo.bar"},
		{Name: "Valid With Path", Str: "A	foo/bar.baz", Want: true, WantStr: "foo/bar.baz"},
		{Name: "Invalid", Str: "A	foo%", Want: false, WantStr: ""},
		{Name: "Invalid No File", Str: "A	", Want: false, WantStr: ""},
		{Name: "Invalid Wrong prefix", Str: "V	foo", Want: false, WantStr: ""},
	}

	for _, c := range cases {
		got, gotStr := isNew(c.Str)
		if got != c.Want || gotStr != c.WantStr {
			t.Errorf("Test: %v, Got: %v Want: %v. Got Str: %v Want Str: %v", c.Name, got, c.Want, gotStr, c.WantStr)
		}
	}
}

func TestDetectsGone(t *testing.T) {
	cases := []struct {
		Name    string
		Str     string
		Want    bool
		WantStr string
	}{
		{Name: "Valid", Str: "D	foo", Want: true, WantStr: "foo"},
		{Name: "Valid With Extension", Str: "D	foo.bar", Want: true, WantStr: "foo.bar"},
		{Name: "Valid With Path", Str: "D	foo/bar.baz", Want: true, WantStr: "foo/bar.baz"},
		{Name: "Invalid", Str: "D	foo%", Want: false, WantStr: ""},
		{Name: "Invalid No File", Str: "D	", Want: false, WantStr: ""},
		{Name: "Invalid Wrong prefix", Str: "A	foo", Want: false, WantStr: ""},
	}

	for _, c := range cases {
		got, gotStr := isGone(c.Str)
		if got != c.Want || gotStr != c.WantStr {
			t.Errorf("Test: %v, Got: %v Want: %v. Got Str: %v Want Str: %v", c.Name, got, c.Want, gotStr, c.WantStr)
		}
	}
}

func TestDetectsModified(t *testing.T) {
	cases := []struct {
		Name    string
		Str     string
		Want    bool
		WantStr string
	}{
		{Name: "Valid", Str: "M	foo", Want: true, WantStr: "foo"},
		{Name: "Valid With Extension", Str: "M	foo.bar", Want: true, WantStr: "foo.bar"},
		{Name: "Valid With Path", Str: "M	foo/bar.baz", Want: true, WantStr: "foo/bar.baz"},
		{Name: "Invalid", Str: "M	foo%", Want: false, WantStr: ""},
		{Name: "Invalid No File", Str: "M	", Want: false, WantStr: ""},
		{Name: "Invalid Wrong prefix", Str: "A	foo", Want: false, WantStr: ""},
	}

	for _, c := range cases {
		got, gotStr := isModified(c.Str)
		if got != c.Want || gotStr != c.WantStr {
			t.Errorf("Test: %v, Got: %v Want: %v. Got Str: %v Want Str: %v", c.Name, got, c.Want, gotStr, c.WantStr)
		}
	}
}

func TestDetectsCopy(t *testing.T) {
	cases := []struct {
		Name     string
		Str      string
		Want     bool
		WantFrom string
		WantTo   string
	}{
		{Name: "Valid", Str: "C100	foo	bar", Want: true, WantFrom: "foo", WantTo: "bar"},
		{Name: "Valid With Extension", Str: "C100	foo.bar	foo/biz", Want: true, WantFrom: "foo.bar", WantTo: "foo/biz"},
		{Name: "Valid With Path", Str: "C100	foo/bar.baz	foo/baz.biz", Want: true, WantFrom: "foo/bar.baz", WantTo: "foo/baz.biz"},
		{Name: "Invalid", Str: "C	foo	bar", Want: false, WantFrom: "", WantTo: ""},
		{Name: "Invalid No File", Str: "C	", Want: false, WantFrom: "", WantTo: ""},
		{Name: "Invalid No To", Str: "C	foo	", Want: false, WantFrom: "", WantTo: ""},
		{Name: "Invalid Too few Digits", Str: "C50	foo	bar", Want: false, WantFrom: "", WantTo: ""},
		{Name: "Invalid Only One Digit", Str: "C5	foo	bar", Want: false, WantFrom: "", WantTo: ""},
		{Name: "Invalid Too Many Digits", Str: "C5000	foo	bar", Want: false, WantFrom: "", WantTo: ""},
		{Name: "Invalid Wrong prefix", Str: "A	foo	bar", Want: false, WantFrom: "", WantTo: ""},
	}

	for _, c := range cases {
		got, gotFrom, gotTo := isCopy(c.Str)
		if got != c.Want || gotFrom != c.WantFrom || gotTo != c.WantTo {
			t.Errorf("Test: %v, Got: %v Want: %v. Got From: %v Want From: %v. Got To: %v Want To: %v", c.Name, got, c.Want, gotFrom, c.WantFrom, gotTo, c.WantTo)
		}
	}
}

func TestDetectsRename(t *testing.T) {
	cases := []struct {
		Name     string
		Str      string
		Want     bool
		WantFrom string
		WantTo   string
	}{
		{Name: "Valid", Str: "R100	foo	bar", Want: true, WantFrom: "foo", WantTo: "bar"},
		{Name: "Valid With Extension", Str: "R100	foo.bar	foo/biz", Want: true, WantFrom: "foo.bar", WantTo: "foo/biz"},
		{Name: "Valid With Path", Str: "R100	foo/bar.baz	foo/baz.biz", Want: true, WantFrom: "foo/bar.baz", WantTo: "foo/baz.biz"},
		{Name: "Invalid", Str: "R	foo	bar", Want: false, WantFrom: "", WantTo: ""},
		{Name: "Invalid No File", Str: "R	", Want: false, WantFrom: "", WantTo: ""},
		{Name: "Invalid No To", Str: "R	foo	", Want: false, WantFrom: "", WantTo: ""},
		{Name: "Invalid Too few Digits", Str: "R50	foo	bar", Want: false, WantFrom: "", WantTo: ""},
		{Name: "Invalid Only One Digit", Str: "R5	foo	bar", Want: false, WantFrom: "", WantTo: ""},
		{Name: "Invalid Too Many Digits", Str: "R5000	foo	bar", Want: false, WantFrom: "", WantTo: ""},
		{Name: "Invalid Wrong prefix", Str: "A100	foo	bar", Want: false, WantFrom: "", WantTo: ""},
	}

	for _, c := range cases {
		got, gotFrom, gotTo := isRenamed(c.Str)
		if got != c.Want || gotFrom != c.WantFrom || gotTo != c.WantTo {
			t.Errorf("Test: %v, Got: %v Want: %v. Got From: %v Want From: %v. Got To: %v Want To: %v", c.Name, got, c.Want, gotFrom, c.WantFrom, gotTo, c.WantTo)
		}
	}
}
