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
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// Require a minimum of 2 spaces between columns.
const padding = 2

// PrintMap outputs a formatted object to the console.
func PrintMap(w io.Writer, fields []string, m map[string]string) {
	if sw, ok := structured(w); ok {
		printStructured(sw, keysToCamelCase(m))
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, padding, ' ', 0)
	for _, field := range fields {
		fmt.Fprintf(tw, "%v:\t%v\n", field, m[field])
	}
	tw.Flush()
}

// PrintAllMap outputs a formatted object to the console.
func PrintAllMap(w io.Writer, m map[string]string) {
	if sw, ok := structured(w); ok {
		printStructured(sw, keysToCamelCase(m))
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, padding, ' ', 0)
	for field, val := range m {
		fmt.Fprintf(tw, "%v:\t%v\n", field, val)
	}
	tw.Flush()
}

// PrintList outputs a formatted list to the console.
func PrintList(w io.Writer, fields []string, items []map[string]string) {
	if sw, ok := structured(w); ok {
		converted := make([]map[string]string, len(items))
		for i, item := range items {
			converted[i] = keysToCamelCase(item)
		}
		printStructured(sw, converted)
		return
	}

	if len(items) == 0 {
		fmt.Fprintln(w, "no results")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, padding, ' ', 0)

	for _, field := range fields {
		fmt.Fprintf(tw, "%v\t", field)
	}
	fmt.Fprintln(tw, "")

	for _, field := range fields {
		fmt.Fprintf(tw, "%v\t", strings.Repeat("=", len(field)))
	}
	fmt.Fprintln(tw, "")

	for _, m := range items {
		for _, f := range fields {
			fmt.Fprintf(tw, "%v\t", m[f])
		}
		fmt.Fprintln(tw, "")
	}

	tw.Flush()
}
