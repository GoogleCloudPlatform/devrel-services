// Copyright 2020 Google LLC
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

package leif

type SLORule struct {
	AppliesTo          AppliesTo          `json:"appliesTo"`
	ComplianceSettings ComplianceSettings `json:"complianceSettings"`
}

type AppliesTo struct {
	GitHubLabels         []string `json:"gitHubLabels"`
	ExcludedGitHubLabels []string `json:"excludedGitHubLabels"`
	Issues               bool     `json:"issues"`
	PRs                  bool     `json:"prs"`
}

type ComplianceSettings struct {
	ResponseTime     duration   `json:"responseTime"`
	ResolutionTime   duration   `json:"resolutionTime"`
	RequiresAssignee bool       `json:"requiresAssignee"`
	Responders       Responders `json:"responders"`
}

type Responders struct {
	Owners       []string `json:"owners"`
	Contributors string   `json:"contributors"`
	Users        []string `json:"users"`
}
