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
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var oneSec, _ = time.ParseDuration("1s")
var dayReg = regexp.MustCompile(`[0-9]+d`)

type stringOrArray []string

func (soa *stringOrArray) UnmarshalJSON(data []byte) error {
	var tempInterface interface{}
	var slice []string

	err := json.Unmarshal(data, &tempInterface)
	if err != nil {
		//this never returns an error?
		return err
	}

	str, isString := tempInterface.(string)

	if isString {
		slice = append(slice, str)
	} else {
		err = json.Unmarshal(data, &slice)
	}

	*soa = slice

	return err
}

type duration time.Duration

func (d duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (stringOrInt *duration) UnmarshalJSON(data []byte) error {
	var tempInterface interface{}

	err := json.Unmarshal(data, &tempInterface)
	if err != nil {
		return err
	}

	str, isString := tempInterface.(string)

	if isString {
		d, err := ParseDurationWithDays(str)
		*stringOrInt = duration(d)
		return err
	}

	value, isNumber := tempInterface.(float64)

	if isNumber {
		*stringOrInt = duration(time.Duration(int64(value) * int64(oneSec.Nanoseconds())))
		return err
	}

	return errors.New("Invalid duration format")
}

type priority string
type issueType string

type sloRuleJSON struct {
	AppliesToJSON          AppliesToJSON          `json:"appliesTo"`
	ComplianceSettingsJSON ComplianceSettingsJSON `json:"complianceSettings"`
}

// returns a new SLORuleJSON with the defaults applied
// except for the responders default which requires knowing if it was partially assigned
func newSLORuleJSON() *sloRuleJSON {
	return &sloRuleJSON{
		AppliesToJSON:          AppliesToJSON{Issues: true, PRs: false},
		ComplianceSettingsJSON: ComplianceSettingsJSON{RequiresAssignee: false},
	}
}

func (rule *sloRuleJSON) addPriorityToGitHubLabels() {
	if len(rule.AppliesToJSON.Priority) > 0 {
		priority := "priority: " + rule.AppliesToJSON.Priority
		rule.AppliesToJSON.GitHubLabelsRaw = append(rule.AppliesToJSON.GitHubLabelsRaw, priority)
	}
}

func (rule *sloRuleJSON) addIssueTypeToGitHubLabels() {
	if len(rule.AppliesToJSON.IssueType) > 0 {
		issueType := "type: " + rule.AppliesToJSON.IssueType
		rule.AppliesToJSON.GitHubLabelsRaw = append(rule.AppliesToJSON.GitHubLabelsRaw, issueType)
	}
}

func (rule *sloRuleJSON) applyResponderDefault() { //if not defined
	if rule.ComplianceSettingsJSON.RespondersJSON.OwnersRaw == nil &&
		len(rule.ComplianceSettingsJSON.RespondersJSON.Contributors) < 1 &&
		rule.ComplianceSettingsJSON.RespondersJSON.Users == nil {
		rule.ComplianceSettingsJSON.RespondersJSON.Contributors = "WRITE"
	}
}

type AppliesToJSON struct {
	GitHubLabelsRaw         stringOrArray `json:"gitHubLabels"`
	ExcludedGitHubLabelsRaw stringOrArray `json:"excludedGitHubLabels"`
	Priority                string        `json:"priority"`
	IssueType               string        `json:"issueType"`
	Issues                  bool          `json:"issues"`
	PRs                     bool          `json:"prs"`
}

type ComplianceSettingsJSON struct {
	ResponseTime     duration       `json:"responseTime"`
	ResolutionTime   duration       `json:"resolutionTime"`
	RequiresAssignee bool           `json:"requiresAssignee"`
	RespondersJSON   RespondersJSON `json:"responders"`
}

type RespondersJSON struct {
	OwnersRaw    stringOrArray `json:"owners"`
	Contributors string        `json:"contributors"`
	Users        []string      `json:"users"`
}

func ParseDurationWithDays(duration string) (time.Duration, error) {
	if strings.Contains(duration, ".") {
		return 0, errors.New("Duration should not contain fractions")
	}
	str := dayReg.ReplaceAllStringFunc(duration, func(s string) string {
		days, _ := strconv.Atoi(s[:len(s)-1])
		s = strconv.Itoa(days*24) + "h"
		return s
	})
	return time.ParseDuration(str)
}

func parseSLORule(rawRule *json.RawMessage) (*SLORule, error) {
	jsonRule := newSLORuleJSON() //apply defaults

	err := json.Unmarshal(*rawRule, &jsonRule) //convert possible strings to arrays
	if err != nil {
		return nil, err
	}

	jsonRule.addPriorityToGitHubLabels()
	jsonRule.addIssueTypeToGitHubLabels()
	jsonRule.applyResponderDefault()

	marshaled, err := json.Marshal(jsonRule)
	if err != nil {
		return nil, err
	}

	var parsedRule *SLORule
	err = json.Unmarshal(marshaled, &parsedRule)

	return parsedRule, err
}

func unmarshalSLOs(data []byte) ([]*SLORule, error) {
	var sloRules []*SLORule
	var rawSLORules []*json.RawMessage

	if len(data) == 0 {
		return sloRules, nil
	}

	err := json.Unmarshal(data, &rawSLORules)
	if err != nil {
		return sloRules, err
	}

	for _, rawRule := range rawSLORules {
		rule, err := parseSLORule(rawRule)
		if err != nil {
			return sloRules, err
		}
		sloRules = append(sloRules, rule)
	}

	return sloRules, err
}

// func (rule *SLORule) String() string {
// 	str := string(*rule)
// 	return str
// }
