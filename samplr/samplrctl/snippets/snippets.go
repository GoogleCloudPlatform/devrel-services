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

package snippets

import (
	"io"
	"strconv"

	"github.com/GoogleCloudPlatform/devrel-services/samplr"
	"github.com/GoogleCloudPlatform/devrel-services/samplr/samplrctl/output"
)

const (
	nameName      = "Name"
	languageName  = "Language"
	versionsName  = "Versions"
	isDeletedName = "Is Deleted"
)

func snippetToMap(s *samplr.Snippet) map[string]string {
	return map[string]string{
		nameName:      s.Name,
		languageName:  s.Language,
		versionsName:  strconv.FormatInt(int64(len(s.Versions)), 10),
		isDeletedName: strconv.FormatBool(len(s.Primary.Content) == 0),
	}
}

func OutputSnippet(w io.Writer, s *samplr.Snippet) {
	output.PrintAllMap(w, snippetToMap(s))
}

func OutputSnippets(w io.Writer, l []*samplr.Snippet) {
	mapList := make([]map[string]string, len(l))
	for i, p := range l {
		mapList[i] = snippetToMap(p)
	}
	fields := []string{
		nameName,
		languageName,
		versionsName,
		isDeletedName,
	}
	output.PrintList(w, fields, mapList)
}
