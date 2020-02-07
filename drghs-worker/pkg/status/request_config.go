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

package status

import "strings"

// RequestConfig is a grouping if requests and rules
type RequestConfig struct {
	ID    int
	Rules []*RequestRule
	Repos []string
}

// Rule returns a RequestRule based on the labels in the RequestConfig
func (slo *RequestConfig) Rule(labels []string) *RequestRule {
	labelMap := make(map[string]bool)
	for _, l := range labels {
		labelMap[strings.ToLower(l)] = true
	}

	for _, rule := range slo.Rules {
		match := true
		for _, label := range rule.Labels {
			label = strings.ToLower(label)
			if !labelMap[strings.ToLower(label)] {
				match = false
				break
			}
		}
		// TODO: use labels like `break outer`
		if match {
			return rule
		}
	}
	return nil
}
