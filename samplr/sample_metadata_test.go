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

	"gopkg.in/yaml.v2"
)

func TestUnMarshals(t *testing.T) {
	tests := []struct {
		Name     string
		Input    string
		WantErr  error
		WantMeta SampleMetadata
	}{
		{
			Name: "Successful Unmarshal",
			Input: `
sample-metadata:
  title: Activate HMAC SA Key.
  description: Activate HMAC SA Key.
  usage: node hmacKeyActivate.js <hmacKeyAccessId> [projectId]
`,
			WantErr: nil,
			WantMeta: SampleMetadata{
				Meta: SampleMeta{
					Title:       "Activate HMAC SA Key.",
					Description: "Activate HMAC SA Key.",
					Usage:       "node hmacKeyActivate.js <hmacKeyAccessId> [projectId]",
				},
			},
		},
		{
			Name: "Complex Unmarshal",
			Input: `
sample-metadata:
  title: Activate HMAC SA Key.
  description: Activate HMAC SA Key.
  usage: node hmacKeyActivate.js <hmacKeyAccessId> [projectId]
  api_version: 1.2.3.5-foo_bar_baz
  snippets:
  - region_tag: foo
    description: bar baz
    usage: biz!
`,
			WantErr: nil,
			WantMeta: SampleMetadata{
				Meta: SampleMeta{
					Title:       "Activate HMAC SA Key.",
					Description: "Activate HMAC SA Key.",
					Usage:       "node hmacKeyActivate.js <hmacKeyAccessId> [projectId]",
					APIVersion:  "1.2.3.5-foo_bar_baz",
					Snippets: []SnippetMetaRef{
						SnippetMetaRef{
							RegionTag:   "foo",
							Description: "bar baz",
							Usage:       "biz!",
						},
					},
				},
			},
		},
		{
			Name: "Misspelled Root",
			Input: `
sample_metadata:
  title: Activate HMAC SA Key.
  description: Activate HMAC SA Key.
  usage: node hmacKeyActivate.js <hmacKeyAccessId> [projectId]
`,
			WantErr:  nil,
			WantMeta: SampleMetadata{},
		},
	}

	for _, c := range tests {
		var d SampleMetadata
		gotErr := yaml.Unmarshal([]byte(c.Input), &d)
		if gotErr != c.WantErr {
			t.Errorf("%v Errors Differ. Want: %v, Got: %v", c.Name, c.WantErr, gotErr)
		}
		if !reflect.DeepEqual(c.WantMeta, d) {
			t.Errorf("%v Metas Differ. Want %v, Got: %v", c.Name, c.WantMeta, d)
		}
	}
}

func TestParsesCommentedLines(t *testing.T) {
	cases := []struct {
		Name     string
		Input    string
		WantMeta *SampleMetadata
		WantErr  error
	}{
		{
			Name: "Successful With Sharp",
			Input: `
# sample-metadata:
#   title: Foo
`,
			WantMeta: &SampleMetadata{
				Meta: SampleMeta{
					Title: "Foo",
				},
			},
			WantErr: nil,
		},
		{
			Name: "Successful With Slash",
			Input: `
// sample-metadata:
//   title: Foo
`,
			WantMeta: &SampleMetadata{
				Meta: SampleMeta{
					Title: "Foo",
				},
			},
			WantErr: nil,
		},
		{
			Name: "Mixed Comment Styles Fail",
			Input: `
# sample-metadata:
//   title: Foo
`,
			WantMeta: &SampleMetadata{
				Meta: SampleMeta{},
			},
			WantErr: nil,
		},
		{
			Name: "Comment Deviations Terminate",
			Input: `
# sample-metadata:
//   title: Foo
#   title: Foo
`,
			WantMeta: &SampleMetadata{
				Meta: SampleMeta{},
			},
			WantErr: nil,
		},
		{
			Name:     "Empty returns nil",
			Input:    ``,
			WantMeta: nil,
			WantErr:  nil,
		},
		{
			Name: "Nonsense returns nil",
			Input: `foo
			bar
			baz biz`,
			WantMeta: nil,
			WantErr:  nil,
		},
	}

	for _, c := range cases {
		gotMeta, gotErr := parseSampleMetadata(c.Input)
		if gotErr != c.WantErr {
			t.Errorf("%v Errors Differ. Want: %v, Got: %v", c.Name, c.WantErr, gotErr)
		}
		if !reflect.DeepEqual(c.WantMeta, gotMeta) {
			t.Errorf("%v Metas Differ. Want %v, Got: %v", c.Name, c.WantMeta, gotMeta)
		}
	}
}
