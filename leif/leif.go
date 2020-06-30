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

import (
	"time"
)

type Owner struct {
	name     string
	Repos    []*Repository
	SLORules []*SLORule
}

// SLORule represents a service level objective (SLO) rule
type SLORule struct {
	AppliesTo          AppliesTo          `json:"appliesTo"`
	ComplianceSettings ComplianceSettings `json:"complianceSettings"`
}

// AppliesTo stores structured data on which issues and/or pull requests a SLO applies to
type AppliesTo struct {
	GitHubLabels         []string `json:"gitHubLabels"`
	ExcludedGitHubLabels []string `json:"excludedGitHubLabels"`
	Issues               bool     `json:"issues"`
	PRs                  bool     `json:"prs"`
}

// ComplianceSetting stores data on the requirements for an issue or pull request to be considered compliant with the SLO
type ComplianceSettings struct {
	ResponseTime     time.Duration `json:"responseTime"`
	ResolutionTime   time.Duration `json:"resolutionTime"`
	RequiresAssignee bool          `json:"requiresAssignee"`
	Responders       Responders    `json:"responders"`
}

// Responders stores structured data on the responders to the issue or pull request the SLO applies to
type Responders struct {
	Owners       []string `json:"owners"`
	Contributors string   `json:"contributors"`
	Users        []string `json:"users"`
}
