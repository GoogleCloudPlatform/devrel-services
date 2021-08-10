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

package samplr

import (
	"reflect"
	"testing"

	git "github.com/GoogleCloudPlatform/devrel-services/git-go"

	"github.com/google/go-cmp/cmp"
)

func TestValidatesFiles(t *testing.T) {
	tests := []struct {
		file     *git.File
		expected bool
	}{
		{
			file: &git.File{
				Name: ".github/workflows/ci.yaml",
			},
			expected: false,
		},
		{
			file: &git.File{
				Name: ".gitignore",
			},
			expected: false,
		},
		{
			file: &git.File{
				Name: "codecov.yaml",
			},
			expected: false,
		},
		{
			file: &git.File{
				Name: "license-checks.xml",
			},
			expected: false,
		},
		{
			file: &git.File{
				Name: "renovate.json",
			},
			expected: false,
		},
		{
			file: &git.File{
				Name: "synth.py",
			},
			expected: false,
		},
		{
			file: &git.File{
				Name: "bar/.kokoro/config.yaml",
			},
			expected: false,
		},
		{
			file: &git.File{
				Name: "samples/.gitignore",
			},
			expected: false,
		},
		{
			file: &git.File{
				Name: "foo.c",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.cpp",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.cc",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.cs",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.css",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "Dockerfile",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo/Dockerfile",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "/bar/foo/dockerfile",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "DOCKERFILE",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "about_dockerfiles.md",
			},
			expected: false,
		},
		{
			file: &git.File{
				Name: "foo.go",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.gs",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.hcl",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.html",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.jade",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.java",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.js",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.js",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.json",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.kt",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.kts",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.m",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.nomad",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.php",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.pug",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.py",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.rb",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.ru",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.sh",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.swift",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.tf",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.tfvars",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.ts",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.workflow",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.xml",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "foo.yaml",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.c",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.cpp",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.cc",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.cs",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.go",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.gs",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.hcl",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.html",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.jade",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.java",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.js",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.js",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.json",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.kt",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.kts",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.m",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.nomad",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.php",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.pug",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.py",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.rb",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.ru",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.sh",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.swift",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.tf",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.tfvars",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.ts",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.workflow",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.xml",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.yaml",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "baz/foo.bazl",
			},
			expected: false,
		},
		{
			file: &git.File{
				Name: "bar/foo.scala",
			},
			expected: true,
		},
		{
			file: &git.File{
				Name: "bar/foo.groovy",
			},
			expected: true,
		},
	}
	for _, test := range tests {
		if got := isValidFile(test.file); got != test.expected {
			t.Errorf("isValidFile(%v) Expected %v. Got %v", test.file.Name, test.expected, got)
		}
	}
}

func TestDetectsRegionTags(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "Empty is Empty",
			content:  "",
			expected: []string{},
		},
		{
			name:     "Start tags are found",
			content:  "[START asdf]",
			expected: []string{"asdf"},
		},
		{
			name:     "Case sensititve",
			content:  "[Start asdf]",
			expected: []string{},
		},
		{
			name:     "Trailing space matters",
			content:  "[START asdf ]",
			expected: []string{},
		},
		{
			name:     "Leading space matters",
			content:  "[ START asdf]",
			expected: []string{},
		},
		{
			name:     "End is ignored",
			content:  "[END asdf]",
			expected: []string{},
		},
		{
			name:     "Only start matters",
			content:  "[START asdf][END asdf]",
			expected: []string{"asdf"},
		},
		{
			name:     "Unicode is not supported",
			content:  `[START 😊]`,
			expected: []string{},
		},
		{
			name: "Multiple returns are supported",
			content: `[START one]
						[START two]
						[START three]`,
			expected: []string{"one", "two", "three"},
		},
		{
			name: "Tags can be returned more than once",
			content: `[START one]
						[START one]
						[START one]`,
			expected: []string{"one", "one", "one"},
		},
	}
	for _, test := range tests {
		if got := detectRegionTags(test.content); !reflect.DeepEqual(test.expected, got) {
			t.Errorf("detectRegionTags: %v. Expected %v Got %v", test.name, test.expected, got)
		}
	}
}

func TestExtractsSnippetsFromFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		fmt     string
		want    map[string]SnippetVersion
		wanterr error
	}{
		{
			name:    "Finds no snippets on empty",
			content: "",
			fmt:     "",
			want:    make(map[string]SnippetVersion, 0),
			wanterr: nil,
		},
		{
			name:    "Finds Snippets",
			content: "// [START foo]\nimport foo\ndef foo:\n  bar\n// [END foo]",
			fmt:     "foo/%s",
			want: map[string]SnippetVersion{"foo": SnippetVersion{
				Name:    "foo/foo",
				File:    nil,
				Lines:   []string{"L1-L5"},
				Content: "// [START foo]\nimport foo\ndef foo:\n  bar\n// [END foo]",
			}},
			wanterr: nil,
		},
		{
			name:    "Finds Multiple Snippets",
			content: "// [START foo]\nimport foo\ndef foo:\n  bar\n// [END foo]\n\n// [START bar]\nimport bar\ndef bar:\n  baz\n// [END bar]",
			fmt:     "baz/%s",
			want: map[string]SnippetVersion{"foo": SnippetVersion{
				Name:    "baz/foo",
				File:    nil,
				Lines:   []string{"L1-L5"},
				Content: "// [START foo]\nimport foo\ndef foo:\n  bar\n// [END foo]",
			}, "bar": SnippetVersion{
				Name:    "baz/bar",
				File:    nil,
				Lines:   []string{"L7-L11"},
				Content: "// [START bar]\nimport bar\ndef bar:\n  baz\n// [END bar]",
			}},
			wanterr: nil,
		},
		{
			name:    "Concatinates disjoint Snippets",
			content: "// [START foo]\nimport foo\ndef foo:\n  bar\n// [END foo]\n\n// [START bar]\nimport bar\ndef bar:\n  baz\n// [END bar]\n\n// [START foo]\ndef biz:\n  fiz\n// [END foo]\n",
			fmt:     "baz/%s",
			want: map[string]SnippetVersion{"foo": SnippetVersion{
				Name:    "baz/foo",
				File:    nil,
				Lines:   []string{"L1-L5", "L13-L16"},
				Content: "// [START foo]\nimport foo\ndef foo:\n  bar\n// [END foo]\n// [START foo]\ndef biz:\n  fiz\n// [END foo]",
			}, "bar": SnippetVersion{
				Name:    "baz/bar",
				File:    nil,
				Lines:   []string{"L7-L11"},
				Content: "// [START bar]\nimport bar\ndef bar:\n  baz\n// [END bar]",
			}},
			wanterr: nil,
		},
		{
			name:    "Ignores lone start tag",
			content: "// [START foo]\nimport foo\ndef foo:\n  bar\n",
			fmt:     "",
			want:    make(map[string]SnippetVersion, 0),
			wanterr: nil,
		},
		{
			name:    "Ignores mismatched tags",
			content: "// [START foo]\nimport foo\ndef foo:\n  bar\n// [END zfoo]\n",
			fmt:     "",
			want:    make(map[string]SnippetVersion, 0),
			wanterr: nil,
		},
		{
			name:    "Ignores lone end tag",
			content: "import foo\ndef foo:\n  bar\n// [END bar]\n",
			fmt:     "",
			want:    make(map[string]SnippetVersion, 0),
			wanterr: nil,
		},
		{
			name:    "Ignores pydoc tags",
			content: "// [START foo]\nimport foo\ndef foo:\n  \"\"\"foo\n\n    :start-after: [START bigtable_create_table]\n    :end-before: [START bigtable_create_table]\n  bar\n// [END foo]\n",
			fmt:     "foo/%s",
			want: map[string]SnippetVersion{"foo": SnippetVersion{
				Name:    "foo/foo",
				File:    nil,
				Lines:   []string{"L1-L9"},
				Content: "// [START foo]\nimport foo\ndef foo:\n  \"\"\"foo\n\n    :start-after: [START bigtable_create_table]\n    :end-before: [START bigtable_create_table]\n  bar\n// [END foo]",
			}},
			wanterr: nil,
		},
	}
	for _, test := range tests {
		if got, err := extractSnippetVersionsFromFile(test.content, test.fmt); !reflect.DeepEqual(got, test.want) || err != test.wanterr {
			t.Errorf("extractSnippetVersionsFromFile: %v.\n\tWant Value: %v. \n\tGot Value: %v. \n\tWant Err: %v \n\tGot Err: %v", test.name, test.want, got, test.wanterr, err)
		}
	}
}

func TestProcessDeletedFiles(t *testing.T) {

	cases := []struct {
		Name                     string
		Commit                   *GitCommit
		Snippets                 map[string]*Snippet
		DeletedFilesInThisCommit map[string]bool
		SeenSnippets             map[string]map[string]bool
		WantSnippets             map[string]*Snippet
		WantSeenSnippets         map[string]map[string]bool
	}{
		{
			Name:                     "Does nothing on no deletes",
			Commit:                   &GitCommit{Hash: "foo"},
			Snippets:                 map[string]*Snippet{"foo": &Snippet{Name: "Bar"}},
			DeletedFilesInThisCommit: map[string]bool{},
			SeenSnippets:             map[string]map[string]bool{},
			WantSnippets:             map[string]*Snippet{"foo": &Snippet{Name: "Bar"}},
			WantSeenSnippets:         map[string]map[string]bool{},
		},
		{
			Name:                     "Deleted files with no snippets do not affect",
			Commit:                   &GitCommit{Hash: "foo"},
			Snippets:                 map[string]*Snippet{"foo": &Snippet{Name: "Bar"}},
			DeletedFilesInThisCommit: map[string]bool{"foo": true},
			SeenSnippets:             map[string]map[string]bool{"baz": map[string]bool{"bar": true}},
			WantSnippets:             map[string]*Snippet{"foo": &Snippet{Name: "Bar"}},
			WantSeenSnippets:         map[string]map[string]bool{"baz": map[string]bool{"bar": true}},
		},
		{
			Name:                     "Snippets previously deleted have no effect",
			Commit:                   &GitCommit{Hash: "foo"},
			Snippets:                 map[string]*Snippet{"foo": &Snippet{Name: "Bar"}},
			DeletedFilesInThisCommit: map[string]bool{"foo": true},
			SeenSnippets:             map[string]map[string]bool{"baz": map[string]bool{"bar": false, "foo": false}},
			WantSnippets:             map[string]*Snippet{"foo": &Snippet{Name: "Bar"}},
			WantSeenSnippets:         map[string]map[string]bool{"baz": map[string]bool{"bar": false, "foo": false}},
		},
		{
			Name:   "Skips snippets in invalid state",
			Commit: &GitCommit{Hash: "foo"},
			Snippets: map[string]*Snippet{"foo": &Snippet{Name: "Bar",
				Versions: []SnippetVersion{
					SnippetVersion{
						Name: "Foo",
						File: &File{
							FilePath: "foofile",
						},
					},
				}}},
			DeletedFilesInThisCommit: map[string]bool{"foo": true},
			SeenSnippets:             map[string]map[string]bool{"foo": map[string]bool{"bar": true}},
			WantSnippets: map[string]*Snippet{"foo": &Snippet{Name: "Bar",
				Versions: []SnippetVersion{
					SnippetVersion{
						Name: "Foo",
						File: &File{
							FilePath: "foofile",
						},
					},
				}}},
			WantSeenSnippets: map[string]map[string]bool{"foo": map[string]bool{"bar": true}},
		},
		{
			Name:   "Deletes snippets that have been seen",
			Commit: &GitCommit{Hash: "foo"},
			Snippets: map[string]*Snippet{"foo": &Snippet{Name: "Bar",
				Versions: []SnippetVersion{
					SnippetVersion{
						Name: "Foo",
						File: &File{
							FilePath: "foofile",
						},
					},
				}}},
			DeletedFilesInThisCommit: map[string]bool{"foo": true},
			SeenSnippets:             map[string]map[string]bool{"foo": map[string]bool{"foo": true}},
			WantSnippets: map[string]*Snippet{"foo": &Snippet{Name: "Bar",
				Versions: []SnippetVersion{
					SnippetVersion{
						Name: "Foo",
						File: &File{
							FilePath: "foofile",
						},
					},
					SnippetVersion{
						Name: "foo/1",
						File: &File{
							FilePath: "foo",
							GitCommit: &GitCommit{
								Hash: "foo",
							},
							Size: 0,
						},
						Content: "",
						Lines:   make([]string, 0),
					},
				},
				Primary: SnippetVersion{
					Name: "foo/1",
					File: &File{
						FilePath: "foo",
						GitCommit: &GitCommit{
							Hash: "foo",
						},
						Size: 0,
					},
					Content: "",
					Lines:   make([]string, 0),
				},
			}},
			WantSeenSnippets: map[string]map[string]bool{"foo": map[string]bool{"foo": false}},
		},
	}

	for _, c := range cases {
		got := processDeletedFiles(c.Commit, c.Snippets, c.DeletedFilesInThisCommit, c.SeenSnippets)
		if diff := cmp.Diff(c.WantSnippets, got); diff != "" {
			t.Errorf("processDeletedFiles: %v snippets diff. match (-want +got):\n%s", c.Name, diff)
		}

		if diff := cmp.Diff(c.WantSeenSnippets, c.SeenSnippets); diff != "" {
			t.Errorf("processDeletedFiles: %v.seenSnipepts diff. match (-want +got):\n%s", c.Name, diff)
		}
	}
}

func TestProcessSeenSnippets(t *testing.T) {
	cases := []struct {
		Name                        string
		Commit                      *GitCommit
		Snippets                    map[string]*Snippet
		SnippetVersionsInThisCommit map[*File]map[string]SnippetVersion
		SeenSnippets                map[string]map[string]bool
		WantSnippets                map[string]*Snippet
		WantSeenSnippets            map[string]map[string]bool
	}{
		{
			Name:     "Skips snippets when they are marked as 'seen'",
			Commit:   &GitCommit{Hash: "aHash"},
			Snippets: map[string]*Snippet{},
			SnippetVersionsInThisCommit: map[*File]map[string]SnippetVersion{
				&File{FilePath: "foofile"}: map[string]SnippetVersion{
					"foosnp": SnippetVersion{
						Name: "foosnp",
					},
				},
			},
			SeenSnippets: map[string]map[string]bool{
				"foofile": map[string]bool{
					"foosnp": true,
				},
			},
			WantSnippets: map[string]*Snippet{},
			WantSeenSnippets: map[string]map[string]bool{
				"foofile": map[string]bool{
					"foosnp": true,
				},
			},
		},
		{
			Name:     "Skips snippets whose file is not 'seen' in this commit",
			Commit:   &GitCommit{Hash: "aHash"},
			Snippets: map[string]*Snippet{},
			SnippetVersionsInThisCommit: map[*File]map[string]SnippetVersion{
				&File{FilePath: "barfile"}: map[string]SnippetVersion{
					"foosnp": SnippetVersion{
						Name: "foosnp",
					},
				},
			},
			SeenSnippets: map[string]map[string]bool{
				"foofile": map[string]bool{
					"foosnp": true,
				},
			},
			WantSnippets: map[string]*Snippet{},
			WantSeenSnippets: map[string]map[string]bool{
				"foofile": map[string]bool{
					"foosnp": true,
				},
			},
		},
		{
			Name:   "Adds delete record for snippets that are not 'seen'",
			Commit: &GitCommit{Hash: "aHash"},
			Snippets: map[string]*Snippet{
				"foosnp": &Snippet{
					Name: "foosnp",
					Versions: []SnippetVersion{
						SnippetVersion{
							File: &File{
								FilePath: "foofile",
								Size:     1234,
							},
						},
					},
				},
			},
			SnippetVersionsInThisCommit: map[*File]map[string]SnippetVersion{
				&File{FilePath: "foofile"}: map[string]SnippetVersion{
					"bar": SnippetVersion{
						Name: "snpV",
					},
				},
			},
			SeenSnippets: map[string]map[string]bool{
				"foofile": map[string]bool{
					"foosnp": true,
				},
			},
			WantSnippets: map[string]*Snippet{
				"foosnp": &Snippet{
					Name: "foosnp",
					Versions: []SnippetVersion{
						SnippetVersion{
							File: &File{
								FilePath: "foofile",
								Size:     1234,
							},
						},
						SnippetVersion{
							Name: "foosnp/1",
							File: &File{
								FilePath: "foofile",
								Size:     0,
								GitCommit: &GitCommit{
									Hash: "aHash",
								},
							},
							Lines:   []string{},
							Content: "",
						},
					},
					Primary: SnippetVersion{
						Name: "foosnp/1",
						File: &File{
							FilePath: "foofile",
							Size:     0,
							GitCommit: &GitCommit{
								Hash: "aHash",
							},
						},
						Lines:   []string{},
						Content: "",
					},
				},
			},
			WantSeenSnippets: map[string]map[string]bool{
				"foofile": map[string]bool{
					"foosnp": false,
				},
			},
		},
	}

	for _, c := range cases {
		processPreviouslySeenSnippets(c.Commit, c.Snippets, c.SnippetVersionsInThisCommit, c.SeenSnippets)
		if diff := cmp.Diff(c.WantSeenSnippets, c.SeenSnippets); diff != "" {
			t.Errorf("%v: processPreviouslySeenSnippets: seenSnippets diff. match (-want +got):\n%s", c.Name, diff)
		}
		if diff := cmp.Diff(c.WantSnippets, c.Snippets); diff != "" {
			t.Errorf("%v: processPreviouslySeenSnippets: snippets diff. match (-want +got):\n%s", c.Name, diff)
		}
	}
}

func TestProcessSnippetVersionsInThisCommit(t *testing.T) {
	cases := []struct {
		Name                        string
		Commit                      *GitCommit
		Snippets                    map[string]*Snippet
		SnippetVersionsInThisCommit map[*File]map[string]SnippetVersion
		SeenSnippets                map[string]map[string]bool
		WantSnippets                map[string]*Snippet
		WantSeenSnippets            map[string]map[string]bool
	}{

		{
			Name:   "Adds new snippets",
			Commit: &GitCommit{Hash: "aHash"},
			Snippets: map[string]*Snippet{
				"foosnp": &Snippet{
					Name:     "foosnp",
					Language: "TEST",
					Versions: []SnippetVersion{},
				},
			},
			SnippetVersionsInThisCommit: map[*File]map[string]SnippetVersion{
				&File{FilePath: "foofile"}: map[string]SnippetVersion{
					"foosnp": SnippetVersion{
						Name: "foosnp",
						File: &File{
							FilePath: "foofile",
						},
					},
				},
			},
			SeenSnippets: map[string]map[string]bool{},
			WantSnippets: map[string]*Snippet{
				"foosnp": &Snippet{
					Name:     "foosnp",
					Language: "TEST",
					Versions: []SnippetVersion{
						SnippetVersion{
							Name: "foosnp/0",
							File: &File{
								FilePath: "foofile",
							},
						},
					},
					Primary: SnippetVersion{
						Name: "foosnp/0",
						File: &File{
							FilePath: "foofile",
						},
					},
				},
			},
			WantSeenSnippets: map[string]map[string]bool{
				"foofile": map[string]bool{"foosnp": true},
			},
		},
		{
			Name:   "Adds new snippets to existing set",
			Commit: &GitCommit{Hash: "aHash"},
			Snippets: map[string]*Snippet{
				"foosnp": &Snippet{
					Name:     "foosnp",
					Language: "TEST",
					Versions: []SnippetVersion{
						SnippetVersion{
							Name: "foosnp/0",
							File: &File{
								FilePath: "foofile",
							},
						},
					},
					Primary: SnippetVersion{
						Name: "foosnp/0",
						File: &File{
							FilePath: "foofile",
						},
					},
				},
			},
			SnippetVersionsInThisCommit: map[*File]map[string]SnippetVersion{
				&File{FilePath: "foofile"}: map[string]SnippetVersion{
					"foosnp": SnippetVersion{
						Name: "foosnp",
						File: &File{
							FilePath: "foofile",
						},
						Content: "content",
					},
				},
			},
			SeenSnippets: map[string]map[string]bool{},
			WantSnippets: map[string]*Snippet{
				"foosnp": &Snippet{
					Name:     "foosnp",
					Language: "TEST",
					Versions: []SnippetVersion{
						SnippetVersion{
							Name: "foosnp/0",
							File: &File{
								FilePath: "foofile",
							},
						},
						SnippetVersion{
							Name: "foosnp/1",
							File: &File{
								FilePath: "foofile",
							},
							Content: "content",
						},
					},
					Primary: SnippetVersion{
						Name: "foosnp/1",
						File: &File{
							FilePath: "foofile",
						},
						Content: "content",
					},
				},
			},
			WantSeenSnippets: map[string]map[string]bool{
				"foofile": map[string]bool{"foosnp": true},
			},
		},
		{
			Name:   "Skips unchanged snippets",
			Commit: &GitCommit{Hash: "aHash"},
			Snippets: map[string]*Snippet{
				"foosnp": &Snippet{
					Name:     "foosnp",
					Language: "TEST",
					Versions: []SnippetVersion{
						SnippetVersion{
							Name: "foosnp/0",
							File: &File{
								FilePath: "foofile",
							},
							Content: "content",
						},
					},
					Primary: SnippetVersion{
						Name: "foosnp/0",
						File: &File{
							FilePath: "foofile",
						},
						Content: "content",
					},
				},
			},
			SnippetVersionsInThisCommit: map[*File]map[string]SnippetVersion{
				&File{FilePath: "foofile"}: map[string]SnippetVersion{
					"foosnp": SnippetVersion{
						Name: "foosnp",
						File: &File{
							FilePath: "foofile",
						},
						Content: "content",
					},
				},
			},
			SeenSnippets: map[string]map[string]bool{"foofile": map[string]bool{"foosnp": true}},
			WantSnippets: map[string]*Snippet{
				"foosnp": &Snippet{
					Name:     "foosnp",
					Language: "TEST",
					Versions: []SnippetVersion{
						SnippetVersion{
							Name: "foosnp/0",
							File: &File{
								FilePath: "foofile",
							},
							Content: "content",
						},
					},
					Primary: SnippetVersion{
						Name: "foosnp/0",
						File: &File{
							FilePath: "foofile",
						},
						Content: "content",
					},
				},
			},
			WantSeenSnippets: map[string]map[string]bool{
				"foofile": map[string]bool{"foosnp": true},
			},
		},
	}

	for _, c := range cases {
		processSnippetVersionsInThisCommit(c.Commit, c.Snippets, c.SnippetVersionsInThisCommit, c.SeenSnippets)
		if diff := cmp.Diff(c.WantSeenSnippets, c.SeenSnippets); diff != "" {
			t.Errorf("processPreviouslySeenSnippets: %v. diff. match (-want +got):\n%s",
				c.Name, diff)
		}
		if diff := cmp.Diff(c.WantSnippets, c.Snippets); diff != "" {
			t.Errorf("processPreviouslySeenSnippets: %v. diff. match (-want +got):\n%s",
				c.Name, diff)
		}
	}
}

func TestSnippetsEquivalent(t *testing.T) {
	cases := []struct {
		Name string
		A    SnippetVersion
		B    SnippetVersion
		Want bool
	}{
		{
			Name: "Deep Equal",
			A: SnippetVersion{
				Content: "a",
				Lines:   []string{"a", "b"},
				File:    &File{FilePath: "foofile"},
			},
			B: SnippetVersion{
				Content: "a",
				Lines:   []string{"a", "b"},
				File:    &File{FilePath: "foofile"},
			},
			Want: true,
		},
		{
			Name: "Equal on both nil files",
			A: SnippetVersion{
				Content: "a",
				Lines:   []string{"a", "b"},
				File:    nil,
			},
			B: SnippetVersion{
				Content: "a",
				Lines:   []string{"a", "b"},
				File:    nil,
			},
			Want: true,
		},
		{
			Name: "Not Equal if one file is non nil",
			A: SnippetVersion{
				Content: "a",
				Lines:   []string{"a", "b"},
				File:    &File{FilePath: "foofile"},
			},
			B: SnippetVersion{
				Content: "a",
				Lines:   []string{"a", "b"},
				File:    nil,
			},
			Want: false,
		},
		{
			Name: "Not Equal if different content",
			A: SnippetVersion{
				Content: "a",
				Lines:   []string{"a", "b"},
				File:    &File{FilePath: "foofile"},
			},
			B: SnippetVersion{
				Content: "b",
				Lines:   []string{"a", "b"},
				File:    &File{FilePath: "foofile"},
			},
			Want: false,
		},
		{
			Name: "Not Equal if different lines",
			A: SnippetVersion{
				Content: "a",
				Lines:   []string{"a", "b"},
				File:    &File{FilePath: "foofile"},
			},
			B: SnippetVersion{
				Content: "a",
				Lines:   []string{"b", "a"},
				File:    &File{FilePath: "foofile"},
			},
			Want: false,
		},
	}
	for _, c := range cases {
		gotAB := snippetsEquivalent(c.A, c.B)
		if gotAB != c.Want {
			t.Errorf("snippetsEquivalent: %v\n\tWant: %v Got: %v.", c.Name, c.Want, gotAB)
		}

		gotBA := snippetsEquivalent(c.B, c.A)

		if gotBA != gotAB {
			t.Errorf("snippetsEquivalent: %v\n\tIs not isomorphic", c.Name)
		}
	}
}

func TestProcessPreviouslySeenSnippets(t *testing.T) {
	cases := []struct {
		Name                        string
		cmt                         *GitCommit
		snippets                    map[string]*Snippet
		snippetVersionsInThisCommit map[*File]map[string]SnippetVersion
		seenSnippets                map[string]map[string]bool
		WantSnippets                map[string]*Snippet
	}{
		{
			Name: "Deletes to to correct file",
			cmt: &GitCommit{
				Hash: "foo",
			},
			snippets: map[string]*Snippet{
				"foo": &Snippet{
					Name: "foo",
				},
			},
			snippetVersionsInThisCommit: map[*File]map[string]SnippetVersion{
				&File{
					FilePath: "bar.sh",
				}: map[string]SnippetVersion{
					"bar": SnippetVersion{},
				},
				&File{
					FilePath: "foo.sh",
				}: map[string]SnippetVersion{},
			},
			seenSnippets: map[string]map[string]bool{
				"foo.sh": map[string]bool{
					"foo": true,
				},
			},
			WantSnippets: map[string]*Snippet{
				"foo": &Snippet{
					Name: "foo",
					Versions: []SnippetVersion{
						{
							Name: "foo/0",
							File: &File{
								FilePath: "foo.sh",
								GitCommit: &GitCommit{
									Hash: "foo",
								},
							},
							Lines: []string{},
						},
					},
					Primary: SnippetVersion{
						Name: "foo/0",
						File: &File{
							FilePath: "foo.sh",
							GitCommit: &GitCommit{
								Hash: "foo",
							},
						},
						Lines:   []string{},
						Content: "",
					},
				},
			},
		},
	}

	for _, c := range cases {
		processPreviouslySeenSnippets(c.cmt, c.snippets, c.snippetVersionsInThisCommit, c.seenSnippets)
		if diff := cmp.Diff(c.WantSnippets, c.snippets); diff != "" {
			t.Errorf("%v failed. Diff (-want, +got) \n%v", c.Name, diff)
		}
	}
}

func TestCleanLanguage(t *testing.T) {
	cases := []struct {
		Name     string
		Language string
		Want     string
	}{
		{
			Name:     "Converts C#",
			Language: "C#",
			Want:     "CSHARP",
		},
		{
			Name:     "Converts C++",
			Language: "C++",
			Want:     "CPP",
		},
		{
			Name:     "Puts to uppercase",
			Language: "Javascript",
			Want:     "JAVASCRIPT",
		},
		{
			Name:     "Spaces to underscores",
			Language: "Ma ven Pom",
			Want:     "MA_VEN_POM",
		},
		{
			Name:     "Hyphens to underscores",
			Language: "Ma-ven-Pom",
			Want:     "MA_VEN_POM",
		},
	}
	for _, c := range cases {
		got := cleanLanguage(c.Language)
		if diff := cmp.Diff(c.Want, got); diff != "" {
			t.Errorf("%v failed. Diff (-want, +got)\n%v", c.Name, diff)
		}
	}
}
