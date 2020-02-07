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

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/build/maintner"
)

var bugLabels = []string{
	"bug",
	"type: bug",
	"type:bug",
	"kind/bug",
	"end-to-end bugs",
	"type:bug/performance",
}

// Status of an issue.
type Status struct {
	Issue *maintner.GitHubIssue `json:"-"`

	RuleID int `json:"slo_rule_id,omitempty"`
	SLOID  int `json:"slo_id,omitempty"`

	Repo              string
	Priority          Priority
	Type              Type
	PriorityUnknown   bool
	Labels            []string
	LastGooglerUpdate time.Time
	LastUserUpdate    time.Time
	Created           time.Time
	UpdatedAt         time.Time
	PullRequest       bool
	Approved          bool
	ClosedAt          *time.Time `json:",omitempty"`
	Closed            bool
	ClosedBy          *maintner.GitHubUser
	Blocked           bool
	ReleaseBlocking   bool
	Body              string
	Commit            string //sha

	// Filled with FillWithSLO()
	UpdateCompliance     *ComplianceResponse `json:",omitempty"`
	ResolutionCompliance *ComplianceResponse `json:",omitempty"`

	// Filled with Fill()
	CompliantUpdates    *Compliance `json:",omitempty"`
	CompliantResolution *Compliance `json:",omitempty"`
	IssueID             int32
	URL                 string
	Assignees           []*maintner.GitHubUser
	Reporter            *maintner.GitHubUser
	Title               string
	Comments            []*maintner.GitHubComment
	Reviews             []*maintner.GitHubReview
}

// FillWithSLO updates the Status with information from the SLO
func (s *Status) FillWithSLO(slo *RequestConfig) {
	s.FillLabels()

	s.SLOID = slo.ID
	rule := slo.Rule(s.Labels)
	if rule != nil {
		s.RuleID = rule.ID
		s.UpdateCompliance = rule.compliantUpdates(s.LastGooglerUpdate)
		// TODO(cbro): determine compliance for closed issues.
		s.ResolutionCompliance = rule.compliantResolution(s.Issue.Created)
	}

	s.IssueID = s.Issue.Number
	s.URL = s.MakeURL()
	s.Assignees = s.Issue.Assignees
	s.Title = s.Issue.Title
}

// FillLabels updates the Status with information from its Labels
func (s *Status) FillLabels() {
	for _, l := range s.Issue.Labels {
		s.Labels = append(s.Labels, l.Name)
		lowercaseName := strings.ToLower(l.Name)
		switch {
		case strings.Contains(lowercaseName, "p0"):
			s.Priority = P0
			s.PriorityUnknown = false
		case strings.Contains(lowercaseName, "p1"):
			s.Priority = P1
			s.PriorityUnknown = false
		case strings.Contains(lowercaseName, "p2"):
			s.Priority = P2
			s.PriorityUnknown = false
		case strings.Contains(lowercaseName, "p3"):
			s.Priority = P3
			s.PriorityUnknown = false
		case strings.Contains(lowercaseName, "p4"):
			s.Priority = P4
			s.PriorityUnknown = false
		case matchesAny(lowercaseName, bugLabels):
			s.Type = TypeBug
		case strings.Contains(lowercaseName, "enhanc"):
			s.Type = TypeFeature
		case strings.Contains(lowercaseName, "feat"):
			s.Type = TypeFeature
		case strings.Contains(lowercaseName, "addition"):
			s.Type = TypeFeature
		case strings.Contains(lowercaseName, "question"):
			s.Type = TypeCustomer
		case strings.Contains(lowercaseName, "cleanup"):
			s.Type = TypeCleanup
		case strings.Contains(lowercaseName, "process"):
			s.Type = TypeProcess
		case strings.Contains(lowercaseName, "blocked"):
			s.Blocked = true
		case strings.Contains(lowercaseName, "blocking"):
			s.ReleaseBlocking = true
		}
	}
}

func matchesAny(item string, valuesToMatch []string) bool {
	for _, valueToMatch := range valuesToMatch {
		if item == valueToMatch {
			return true
		}
	}
	return false
}

// Fill populates the data of the Status
func (s *Status) Fill() {
	s.FillLabels()
	s.CompliantUpdates = s.compliantUpdates()
	s.CompliantResolution = s.compliantResolution()
	s.IssueID = s.Issue.Number
	s.URL = s.MakeURL()
	s.Assignees = s.Issue.Assignees
	s.Reporter = s.Issue.User
	s.Title = s.Issue.Title
}

func (s *Status) assignees() string {
	var as []string
	for _, u := range s.Issue.Assignees {
		as = append(as, u.Login)
	}
	return strings.Join(as, " ")
}

// MakeURL gets the GitHub url for the Status
func (s *Status) MakeURL() string {
	return fmt.Sprintf("https://github.com/%s/issues/%d", s.Repo, s.Issue.Number)
}

func (s *Status) compliantUpdates() *Compliance {
	target := updateObjectives[s.Priority]
	actual := time.Now().Sub(s.LastGooglerUpdate)
	return &Compliance{actual < target, actual}
}

func (s *Status) compliantResolution() *Compliance {
	closedAt := time.Now()
	if s.Issue.Closed {
		closedAt = s.Issue.ClosedAt
	}
	actual := closedAt.Sub(s.Issue.Created)
	target := resolutionObjectives[s.Priority]
	return &Compliance{actual < target, actual}
}

var updateObjectives = map[Priority]time.Duration{
	P0: 30 * time.Minute,
	P1: 24 * time.Hour,
	P2: 5 * 24 * time.Hour,
	P3: 180 * 24 * time.Hour,
	P4: 365 * 24 * time.Hour,
}

var resolutionObjectives = map[Priority]time.Duration{
	P0: 7 * 24 * time.Hour, // "ASAP"
	P1: 7 * 24 * time.Hour,
}
