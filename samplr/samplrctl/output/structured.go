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

package output

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
)

type structuredWriter struct {
	writer io.Writer
}

func (s *structuredWriter) Write(p []byte) (int, error) {
	return ioutil.Discard.Write(p)
}

func structured(w io.Writer) (io.Writer, bool) {
	if sw, ok := w.(*structuredWriter); ok {
		return sw.writer, true
	}
	return nil, false
}

func printStructured(w io.Writer, object interface{}) {
	j, err := json.MarshalIndent(object, "", " ")
	if err != nil {
		log.Fatalf("Failed to marshal output to JSON: %v", err)
	}
	fmt.Fprintln(w, string(j))
}

func keysToCamelCase(m map[string]string) map[string]string {
	c := make(map[string]string, len(m))
	for k, v := range m {
		c[toCamelCase(k)] = v
	}
	return c
}

func toCamelCase(s string) string {
	var b strings.Builder
	for _, w := range strings.Split(strings.ToLower(s), " ") {
		if len(w) == 0 {
			continue
		}
		if b.Len() == 0 {
			fmt.Fprint(&b, w)
			continue
		}
		fmt.Fprint(&b, strings.ToUpper(w[:1])+w[1:])
	}
	return b.String()
}
