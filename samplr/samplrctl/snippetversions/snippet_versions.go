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

package snippetversions

import (
	"io"
	"strconv"
	"strings"

	"github.com/GoogleCloudPlatform/devrel-services/samplr"
	"github.com/GoogleCloudPlatform/devrel-services/samplr/samplrctl/output"
)

const (
	nameName    = "Name"
	fileName    = "File"
	linesName   = "Lines"
	contentName = "Content"
	sizeName    = "Size"
	shaName     = "SHA"
)

func snippetVersionToMap(s samplr.SnippetVersion) map[string]string {
	return map[string]string{
		nameName:    s.Name,
		fileName:    s.File.FilePath,
		linesName:   strings.Join(s.Lines, ","),
		contentName: s.Content,
		sizeName:    strconv.FormatInt(int64(len(s.Content)), 10),
		shaName:     s.File.GitCommit.Hash,
	}
}

func OutputSnippetVersion(w io.Writer, s samplr.SnippetVersion) {
	output.PrintAllMap(w, snippetVersionToMap(s))
}

func OutputSnippetVersions(w io.Writer, l []samplr.SnippetVersion) {
	mapList := make([]map[string]string, len(l))
	for i, p := range l {
		mapList[i] = snippetVersionToMap(p)
	}
	fields := []string{
		nameName,
		fileName,
		linesName,
		sizeName,
		shaName,
	}
	output.PrintList(w, fields, mapList)
}
