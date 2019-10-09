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
	"bufio"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

var (
	// Currently supported are "single line" comments that begin with
	// "#" or "//"
	smre = regexp.MustCompile(`^(#|//) sample-metadata:$`)
)

type SampleMetadata struct {
	Meta SampleMeta `yaml:"sample-metadata"`
}

type SampleMeta struct {
	Title       string           `yaml:"title"`
	Description string           `yaml:"description"`
	Usage       string           `yaml:"usage"`
	ApiVersion  string           `yaml:"api_version"`
	Snippets    []SnippetMetaRef `yaml:"snippets"`
}

type SnippetMetaRef struct {
	RegionTag   string `yaml:"region_tag"`
	Description string `yaml:"description"`
	Usage       string `yaml:"usage"`
}

func parseSampleMetadata(content string) (*SampleMetadata, error) {
	// Detect the metadata region
	scn := bufio.NewScanner(strings.NewReader(content))
	comChr := ""
	ymlcom := strings.Builder{}
	for scn.Scan() {
		txt := scn.Text()
		if comChr == "" && smre.MatchString(txt) {
			m := smre.FindAllStringSubmatch(txt, -1)
			comChr = m[0][1]
		}

		if comChr != "" {
			if txt == "" || !strings.HasPrefix(txt, comChr) {
				break
			}
			// Replace JUST the first occurance of the comment character
			// in the string
			cln := strings.Replace(txt, comChr, "", 1)
			ymlcom.WriteString(cln)
			ymlcom.WriteString("\n")
		}
	}

	if comChr == "" {
		// We didn't find any metadata
		return nil, nil
	}

	// Unmarshal it into struct & return
	var sm SampleMetadata

	err := yaml.Unmarshal([]byte(ymlcom.String()), &sm)
	return &sm, err
}
