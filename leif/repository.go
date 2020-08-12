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

// This package was modeled after the mocking strategy outlined at:
// https://github.com/google/go-github/issues/113#issuecomment-46023864

package leif

import (
	"context"
	"time"

	"github.com/GoogleCloudPlatform/devrel-services/leif/githubservices"
)

// Repository represents a GitHub repository and stores its SLO rules
type Repository struct {
	name      string
	ownerName string
	SLORules  []*SLORule
}

// OwnerName returns the name of the repository's owner
func (r *Repository) OwnerName() string {
	return r.ownerName
}

// RepoName returns the repository's name
func (r *Repository) RepoName() string {
	return r.name
}

// Update reaches out to GitHub to update the SLO rules for the repository
func (r *Repository) Update(ctx context.Context, owner Owner, ghClient *githubservices.Client) error {

	rules, err := findSLODoc(ctx, owner, r.name, ghClient)
	if err != nil {
		return err
	}

	r.SLORules = rules
	return nil
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

// ComplianceSettings stores data on the requirements for an issue or pull request to be considered compliant with the SLO
type ComplianceSettings struct {
	ResponseTime     time.Duration `json:"responseTime"`
	ResolutionTime   time.Duration `json:"resolutionTime"`
	RequiresAssignee bool          `json:"requiresAssignee"`
	Responders       []string      `json:"responders"`
}
