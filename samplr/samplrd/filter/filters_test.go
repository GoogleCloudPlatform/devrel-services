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

package filter

import (
	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/go-cmp/cmp"
)

func TestSnippet(t *testing.T) {
	tests := []struct {
		Name    string
		Snippet drghs_v1.Snippet
		Filter  string
		Want    bool
		WantErr bool
	}{
		{
			Name:    "Empty Filter Passes",
			Snippet: drghs_v1.Snippet{},
			Filter:  "",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Filter name Passes",
			Snippet: drghs_v1.Snippet{
				Name: "foo",
			},
			Filter:  "snippet.name == 'foo' ",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Deep filter Passes",
			Snippet: drghs_v1.Snippet{
				Name: "foo",
				Primary: &drghs_v1.SnippetVersion{
					Lines: []string{
						"foo",
						"bar",
					},
				},
			},
			Filter:  "snippet.primary.lines.size() > 1 ",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Filter language Passes",
			Snippet: drghs_v1.Snippet{
				Name:     "foo",
				Language: "bar",
			},
			Filter:  "snippet.language == 'bar' ",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Unsupported Fields fail",
			Snippet: drghs_v1.Snippet{
				Name: "foo",
			},
			Filter:  "baz == 'foo'",
			Want:    false,
			WantErr: true,
		},
	}
	for _, test := range tests {
		got, goterr := Snippet(&test.Snippet, test.Filter)
		if (test.WantErr && goterr == nil) || (!test.WantErr && goterr != nil) {
			t.Errorf("test: %v, errors diff. WantErr: %v, GotErr: %v.", test.Name, test.WantErr, goterr)
		}
		if diff := cmp.Diff(test.Want, got); diff != "" {
			t.Errorf("test: %v, values diff. match (-want +got)\n%s", test.Name, diff)
		}
	}
}

func TestSnippetVersion(t *testing.T) {
	tests := []struct {
		Name           string
		SnippetVersion drghs_v1.SnippetVersion
		Filter         string
		Want           bool
		WantErr        bool
	}{
		{
			Name:           "Empty Filter Passes",
			SnippetVersion: drghs_v1.SnippetVersion{},
			Filter:         "",
			Want:           true,
			WantErr:        false,
		},
		{
			Name: "Filter name Passes",
			SnippetVersion: drghs_v1.SnippetVersion{
				Name: "foo",
			},
			Filter:  "version.name == 'foo' ",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Filter meta Passes",
			SnippetVersion: drghs_v1.SnippetVersion{
				Name: "foo",
				Meta: &drghs_v1.SnippetVersionMeta{
					Title: "bar",
				},
			},
			Filter:  "version.meta.title == 'bar'",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Filter file Passes",
			SnippetVersion: drghs_v1.SnippetVersion{
				Name: "foo",
				File: &drghs_v1.File{
					Size: 10,
				},
			},
			Filter:  "version.file.size < 100 ",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Unsupported Fields fail",
			SnippetVersion: drghs_v1.SnippetVersion{
				Name: "foo",
			},
			Filter:  "baz == 'foo'",
			Want:    false,
			WantErr: true,
		},
	}
	for _, test := range tests {
		got, goterr := SnippetVersion(&test.SnippetVersion, test.Filter)
		if (test.WantErr && goterr == nil) || (!test.WantErr && goterr != nil) {
			t.Errorf("test: %v, errors diff. WantErr: %v, GotErr: %v.", test.Name, test.WantErr, goterr)
		}
		if diff := cmp.Diff(test.Want, got); diff != "" {
			t.Errorf("test: %v, values diff. match (-want +got)\n%s", test.Name, diff)
		}
	}
}

func TestRepository(t *testing.T) {
	tests := []struct {
		Name    string
		Repo    drghs_v1.Repository
		Filter  string
		Want    bool
		WantErr bool
	}{
		{
			Name:    "Empty Filter Passes",
			Repo:    drghs_v1.Repository{},
			Filter:  "",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Filter name Passes",
			Repo: drghs_v1.Repository{
				Name: "foo",
			},
			Filter:  "repository.name == 'foo' ",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Unsupported Fields fail",
			Repo: drghs_v1.Repository{
				Name: "foo",
			},
			Filter:  "baz == 'foo'",
			Want:    false,
			WantErr: true,
		},
	}
	for _, test := range tests {
		got, goterr := Repository(&test.Repo, test.Filter)
		if (test.WantErr && goterr == nil) || (!test.WantErr && goterr != nil) {
			t.Errorf("test: %v, errors diff. WantErr: %v, GotErr: %v.", test.Name, test.WantErr, goterr)
		}
		if diff := cmp.Diff(test.Want, got); diff != "" {
			t.Errorf("test: %v, values diff. match (-want +got)\n%s", test.Name, diff)
		}
	}
}

func TestGitCommit(t *testing.T) {
	tests := []struct {
		Name    string
		Commit  drghs_v1.GitCommit
		Filter  string
		Want    bool
		WantErr bool
	}{
		{
			Name:    "Empty Filter Passes",
			Commit:  drghs_v1.GitCommit{},
			Filter:  "",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Filter name Passes",
			Commit: drghs_v1.GitCommit{
				Name: "foo",
			},
			Filter:  "commit.name == 'foo' ",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Filter committer_email Passes",
			Commit: drghs_v1.GitCommit{
				CommitterEmail: "foo",
			},
			Filter:  "commit.committer_email == 'foo' ",
			Want:    true,
			WantErr: false,
		},
		{
			Name: "Unsupported Fields fail",
			Commit: drghs_v1.GitCommit{
				Name: "foo",
			},
			Filter:  "baz == 'foo'",
			Want:    false,
			WantErr: true,
		},
	}
	for _, test := range tests {
		got, goterr := GitCommit(&test.Commit, test.Filter)
		if (test.WantErr && goterr == nil) || (!test.WantErr && goterr != nil) {
			t.Errorf("test: %v, errors diff. WantErr: %v, GotErr: %v.", test.Name, test.WantErr, goterr)
		}
		if diff := cmp.Diff(test.Want, got); diff != "" {
			t.Errorf("test: %v, values diff. match (-want +got)\n%s", test.Name, diff)
		}
	}
}

func TestGitCommitTimeFilter(t *testing.T) {
	time, _ := ptypes.TimestampProto(time.Now())
	commit :=
		drghs_v1.GitCommit{
			Name:          "foo",
			CommittedTime: time,
		}
	filter := `commit.committed_time > timestamp("1972-01-01T10:00:20.021-05:00")`
	want := true
	wantErr := false

	got, goterr := GitCommit(&commit, filter)
	if (wantErr && goterr == nil) || (!wantErr && goterr != nil) {
		t.Errorf("Errors diff. WantErr: %v, GotErr: %v.", wantErr, goterr)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Values diff. match (-want +got)\n%s", diff)
	}
}
