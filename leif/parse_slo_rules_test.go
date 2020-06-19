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
	"reflect"
	"testing"
	"time"
)

var defaultSLO = SLORule{AppliesTo: AppliesTo{Issues: true, PRs: false}, ComplianceSettings: ComplianceSettings{RequiresAssignee: false, Responders: Responders{Contributors: "WRITE"}}}
var oneMin, _ = time.ParseDuration("1m")
var oneHour, _ = time.ParseDuration("1h")
var oneDay, _ = time.ParseDuration("24h")

var syntaxError *json.SyntaxError
var unmarshalTypeError *json.UnmarshalTypeError

func TestParsesSLORules(t *testing.T) {
	tests := []struct {
		name       string
		jsonString string
		expected   []*SLORule
		wanterr    bool
	}{
		{
			name:       "Empty array returns no rules",
			jsonString: `[]`,
			expected:   nil,
			wanterr:    false,
		},
		{
			name:       "Empty json returns no rules",
			jsonString: ``,
			expected:   nil,
			wanterr:    false,
		},
		{
			name: "Minimum default rule returns a rule with defaults applied",
			jsonString: `[
				{
					"appliesTo": {},
					"complianceSettings": {
						"responseTime": 0,
						"resolutionTime": 0
					}
				}
		 	]`,
			expected: []*SLORule{&defaultSLO},
			wanterr:  false,
		},
		{
			name: "More than one rule is parsed correctly",
			jsonString: `[
				{
					"appliesTo": {},
					"complianceSettings": {
						"responseTime": 0,
						"resolutionTime": 0
					}
				},
				{
					"appliesTo": {},
					"complianceSettings": {
						"responseTime": 0,
						"resolutionTime": 0
					}
				}
		 	]`,
			expected: []*SLORule{&defaultSLO, &defaultSLO},
			wanterr:  false,
		},
		{
			name: "Basic time strings are parsed correctly",
			jsonString: `[
				{
					"appliesTo": {},
					"complianceSettings": {
						"responseTime": "1h",
						"resolutionTime": "1s"
					}
				}
		 	]`,
			expected: []*SLORule{&SLORule{
				AppliesTo: AppliesTo{
					Issues: true,
					PRs:    false,
				},
				ComplianceSettings: ComplianceSettings{
					ResponseTime:   duration(time.Duration(1 * 60 * 60 * oneSec)),
					ResolutionTime: duration(time.Duration(oneSec)),
					Responders:     Responders{Contributors: "WRITE"},
				},
			}},
			wanterr: false,
		},
		{
			name: "Time strings with multiple values are parsed correctly",
			jsonString: `[
				{
					"appliesTo": {},
					"complianceSettings": {
						"responseTime": "1h1m1s",
						"resolutionTime": "1s1h1m"
					}
				}
		 	]`,
			expected: []*SLORule{&SLORule{
				AppliesTo: AppliesTo{
					Issues: true,
					PRs:    false,
				},
				ComplianceSettings: ComplianceSettings{
					ResponseTime:   duration(time.Duration(oneHour + oneMin + oneSec)),
					ResolutionTime: duration(time.Duration(oneHour + oneMin + oneSec)),
					Responders:     Responders{Contributors: "WRITE"},
				},
			}},
			wanterr: false,
		},
		{
			name: "Time strings with day values are parsed correctly",
			jsonString: `[
				{
					"appliesTo": {},
					"complianceSettings": {
						"responseTime": "1d1h1m1s",
						"resolutionTime": "30d"
					}
				}
		 	]`,
			expected: []*SLORule{&SLORule{
				AppliesTo: AppliesTo{
					Issues: true,
					PRs:    false,
				},
				ComplianceSettings: ComplianceSettings{
					ResponseTime:   duration(time.Duration((24 * oneHour) + oneHour + oneMin + oneSec)),
					ResolutionTime: duration(time.Duration(30 * 24 * oneHour)),
					Responders:     Responders{Contributors: "WRITE"},
				},
			}},
			wanterr: false,
		},
		{
			name: "Time defined as a number is parsed correctly",
			jsonString: `[
				{
					"appliesTo": {},
					"complianceSettings": {
						"responseTime": 0,
						"resolutionTime": 43200
					}
				}
		 	]`,
			expected: []*SLORule{&SLORule{
				AppliesTo: AppliesTo{
					Issues: true,
					PRs:    false,
				},
				ComplianceSettings: ComplianceSettings{
					ResponseTime:   duration(time.Duration(0)),
					ResolutionTime: duration(time.Duration(43200 * oneSec)),
					Responders:     Responders{Contributors: "WRITE"},
				},
			}},
			wanterr: false,
		},
		{
			name: "Incorrect time string fails with error",
			jsonString: `[
				{
					"appliesTo": {},
					"complianceSettings": {
						"responseTime": "1w",
						"resolutionTime": 0
					}
				}
		 	]`,
			expected: nil,
			wanterr:  true,
		},
		{
			name: "Priority gets converted to a GitHub label",
			jsonString: `[
				{
					"appliesTo": {
						"priority": "P1"
					},
					"complianceSettings": {
						"responseTime": 0,
						"resolutionTime": 0
					}
				}
		 	]`,
			expected: []*SLORule{&SLORule{
				AppliesTo: AppliesTo{
					GitHubLabels: []string{"priority: P1"},
					Issues:       true,
					PRs:          false,
				},
				ComplianceSettings: ComplianceSettings{
					ResponseTime:   duration(time.Duration(0)),
					ResolutionTime: duration(time.Duration(0)),
					Responders:     Responders{Contributors: "WRITE"},
				},
			}},
			wanterr: false,
		},
		{
			name: "Issue type gets converted to a GitHub label",
			jsonString: `[
				{
					"appliesTo": {
						"issueType": "bug"
					},
					"complianceSettings": {
						"responseTime": 0,
						"resolutionTime": 0
					}
				}
		 	]`,
			expected: []*SLORule{&SLORule{
				AppliesTo: AppliesTo{
					GitHubLabels: []string{"type: bug"},
					Issues:       true,
					PRs:          false,
				},
				ComplianceSettings: ComplianceSettings{
					ResponseTime:   duration(time.Duration(0)),
					ResolutionTime: duration(time.Duration(0)),
					Responders:     Responders{Contributors: "WRITE"},
				},
			}},
			wanterr: false,
		},
		{
			name: "Contributors is set to none if another field is defined in Responders",
			jsonString: `[
				{
					"appliesTo": {},
					"complianceSettings": {
						"responseTime": 0,
						"resolutionTime": 0,
						"responders": {
							"users": ["@jeff"]
						}
					}
				}
		 	]`,
			expected: []*SLORule{&SLORule{
				AppliesTo: AppliesTo{
					Issues: true,
					PRs:    false,
				},
				ComplianceSettings: ComplianceSettings{
					ResponseTime:   duration(time.Duration(0)),
					ResolutionTime: duration(time.Duration(0)),
					Responders:     Responders{Users: []string{"@jeff"}},
				},
			}},
			wanterr: false,
		},
		{
			name: "Can set GitHubLabels as a string",
			jsonString: `[
				{
					"appliesTo": {
						"gitHubLabels": "a label"
					},
					"complianceSettings": {
						"responseTime": 0,
						"resolutionTime": 0
					}
				}
		 	]`,
			expected: []*SLORule{&SLORule{
				AppliesTo: AppliesTo{
					GitHubLabels: []string{"a label"},
					Issues:       true,
					PRs:          false,
				},
				ComplianceSettings: ComplianceSettings{
					ResponseTime:   duration(time.Duration(0)),
					ResolutionTime: duration(time.Duration(0)),
					Responders:     Responders{Contributors: "WRITE"},
				},
			}},
			wanterr: false,
		},
		{
			name: "Can set GitHubLabels as an array",
			jsonString: `[
				{
					"appliesTo": {
						"gitHubLabels": ["label 1", "label 2"]
					},
					"complianceSettings": {
						"responseTime": 0,
						"resolutionTime": 0
					}
				}
		 	]`,
			expected: []*SLORule{&SLORule{
				AppliesTo: AppliesTo{
					GitHubLabels: []string{"label 1", "label 2"},
					Issues:       true,
					PRs:          false,
				},
				ComplianceSettings: ComplianceSettings{
					ResponseTime:   duration(time.Duration(0)),
					ResolutionTime: duration(time.Duration(0)),
					Responders:     Responders{Contributors: "WRITE"},
				},
			}},
			wanterr: false,
		},
		{
			name: "No responders can be specified",
			jsonString: `[
				{
					"appliesTo": {},
					"complianceSettings": {
						"responseTime": 0,
						"resolutionTime": 0,
						"responders": {
							"users": []
						}
					}
				}
		 	]`,
			expected: []*SLORule{&SLORule{
				AppliesTo: AppliesTo{
					Issues: true,
					PRs:    false,
				},
				ComplianceSettings: ComplianceSettings{
					ResponseTime:   duration(time.Duration(0)),
					ResolutionTime: duration(time.Duration(0)),
					Responders:     Responders{Users: []string{}},
				},
			}},
			wanterr: false,
		},
	}
	for _, test := range tests {
		got, err := unmarshalSLOs([]byte(test.jsonString))
		if !reflect.DeepEqual(got, test.expected) {
			t.Errorf("unmarshalSLOs: %v did not pass.\n\tWant:\t%v\n\tGot:\t%v", test.name, test.expected, got)
		}
		if (test.wanterr && err == nil) || (!test.wanterr && err != nil) {
			t.Errorf("unmarshalSLOs: %v did not pass.\n\tWant Err: %v \n\tGot Err: %v", test.name, test.wanterr, err)
		}
	}
}

func TestParseDurationWithDays(t *testing.T) {
	cases := []struct {
		name     string
		duration string
		want     time.Duration
		wantErr  bool
	}{
		{
			name:     "Standard hours passes",
			duration: "1h",
			want:     time.Duration(oneHour),
			wantErr:  false,
		},
		{
			name:     "One day passes",
			duration: "1d",
			want:     time.Duration(oneDay),
			wantErr:  false,
		},
		{
			name:     "Multiple digits acceptable for days",
			duration: "10d",
			want:     time.Duration(10 * oneDay),
			wantErr:  false,
		},
		{
			name:     "Days may be at any position in the string",
			duration: "1s10d",
			want:     time.Duration(10*oneDay + oneSec),
			wantErr:  false,
		},
		{
			name:     "Fractional input returns error",
			duration: "1.0d",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "Errors on invalid day format",
			duration: "This is so cool 1d",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "Errors on invalid duration format",
			duration: "This is so cool 1h",
			want:     0,
			wantErr:  true,
		},
	}
	for _, c := range cases {
		got, gotErr := parseDurationWithDays(c.duration)
		if got != c.want {
			t.Errorf("Test %v did not pass. Got: %v, Want: %v", c.name, got, c.want)
		}
		if (gotErr == nil && c.wantErr) || (gotErr != nil && !c.wantErr) {
			t.Errorf("Test %v did not pass. GotErr: %v, WantErr: %v", c.name, gotErr, c.wantErr)
		}
	}
}

func TestStringOrArrayUnmarshalling(t *testing.T) {
	cases := []struct {
		name      string
		jsonInput string
		expected  stringOrArray
		wantErr   error
	}{
		{
			name:      "Parses a string into an array",
			jsonInput: `"this is a string"`,
			expected:  []string{"this is a string"},
			wantErr:   nil,
		},
		{
			name:      "Parses an array",
			jsonInput: `["this is an array"]`,
			expected:  []string{"this is an array"},
			wantErr:   nil,
		},
		{
			name:      "Parses an array with several elements",
			jsonInput: `["el 1", "el 2"]`,
			expected:  []string{"el 1", "el 2"},
			wantErr:   nil,
		},
		{
			name:      "Incorrect input returns error",
			jsonInput: `this doesn't work`,
			expected:  nil,
			wantErr:   syntaxError,
		},
		{
			name:      "Numeric input returns error",
			jsonInput: `1`,
			expected:  nil,
			wantErr:   unmarshalTypeError,
		},
		{
			name:      "Malformed json input returns error",
			jsonInput: ``,
			expected:  nil,
			wantErr:   syntaxError,
		},
	}
	for _, c := range cases {
		var got stringOrArray
		gotErr := json.Unmarshal([]byte(c.jsonInput), &got)
		if !reflect.DeepEqual(got, c.expected) || (c.wantErr == nil && gotErr != nil) || (c.wantErr != nil && reflect.TypeOf(gotErr) != reflect.TypeOf(c.wantErr)) {
			t.Errorf("stringOrArrayUmarshal: %v did not pass.\n\tGot:\t%v\n\tWant:\t%v", c.name, got, c.expected)
		}
		if (c.wantErr == nil && gotErr != nil) || (c.wantErr != nil && reflect.TypeOf(gotErr) != reflect.TypeOf(c.wantErr)) {
			t.Errorf("stringOrArrayUmarshal: %v did not pass.\n\tGot Err:\t%v\n\tWant Err:\t%v", c.name, gotErr, c.wantErr)
		}
	}
}

func TestDurationMarshalling(t *testing.T) {
	cases := []struct {
		name     string
		dur      duration
		expected string
		wantErr  error
	}{
		{
			name:     "Basic marshal int as time.duration",
			dur:      0,
			expected: `"0s"`,
			wantErr:  nil,
		},
		{
			name:     "Marshal duration as time.duration",
			dur:      duration(oneSec),
			expected: `"1s"`,
			wantErr:  nil,
		},
	}
	for _, c := range cases {
		got, gotErr := json.Marshal(c.dur)

		if !reflect.DeepEqual(got, []byte(c.expected)) {
			t.Errorf("durationMarshal: %v did not pass.\n\tGot:\t%v\n\tWant:\t%v", c.name, string(got), c.expected)
		}
		if (c.wantErr == nil && gotErr != nil) || (c.wantErr != nil && reflect.TypeOf(gotErr) != reflect.TypeOf(c.wantErr)) {
			t.Errorf("durationMarshal: %v did not pass.\n\tGot Err:\t%v\n\tWant Err:\t%v", c.name, gotErr, c.wantErr)
		}

	}
}

func TestDurationUnmarshalling(t *testing.T) {
	cases := []struct {
		name      string
		jsonInput string
		expected  duration
		wantErr   bool
	}{
		{
			name:      "Unmarshal int as seconds",
			jsonInput: `1`,
			expected:  duration(oneSec),
			wantErr:   false,
		},
		{
			name:      "Unmarshal int 0 to 0",
			jsonInput: `0`,
			expected:  0,
			wantErr:   false,
		},
		{
			name:      "Unmarshal 0 sec string to 0",
			jsonInput: `"0s"`,
			expected:  0,
			wantErr:   false,
		},
		{
			name:      "Unmarshals days correctly",
			jsonInput: `"2d"`,
			expected:  duration(2 * oneDay),
			wantErr:   false,
		},
		{
			name:      "Incorrect time format fails",
			jsonInput: `"1w"`,
			expected:  0,
			wantErr:   true,
		},
		{
			name:      "Incorrect input fails",
			jsonInput: `["arrays don't work"]`,
			expected:  0,
			wantErr:   true,
		},
		{
			name:      "Fractional input fails",
			jsonInput: `"1.5h"`,
			expected:  0,
			wantErr:   true,
		},
	}
	for _, c := range cases {
		var got *duration
		gotErr := json.Unmarshal([]byte(c.jsonInput), &got)

		if *got != c.expected {
			t.Errorf("duration unmarshalling: %v did not pass.\n\tGot:\t%v\n\tWant:\t%v", c.name, *got, c.expected)
		}
		if (c.wantErr && gotErr == nil) || (!c.wantErr && gotErr != nil) {
			t.Errorf("duration unmarshalling: %v did not pass.\n\tGot Err:\t%v\n\tWant Err:\t%v", c.name, gotErr, c.wantErr)
		}
	}
}

func TestSLORuleCreation(t *testing.T) {
	cases := []struct {
		name     string
		expected *sloRuleJSON
	}{
		{
			name: "Creates a SLO with defaults",
			expected: &sloRuleJSON{
				AppliesToJSON: AppliesToJSON{
					Issues: true,
					PRs:    false,
				},
				ComplianceSettingsJSON: ComplianceSettingsJSON{
					ResponseTime:     0,
					ResolutionTime:   0,
					RequiresAssignee: false,
				},
			},
		},
	}
	for _, c := range cases {
		if got := newSLORuleJSON(); !reflect.DeepEqual(got, c.expected) {
			t.Errorf("newSLORuleJSON: %v did not pass. \n\tGot:\t\t%v\n\tExpected:\t%v", c.name, got, c.expected)
		}

	}
}

func TestAddToGitHubLabels(t *testing.T) {
	cases := []struct {
		name     string
		prepend  string
		label    string
		current  *sloRuleJSON
		expected *sloRuleJSON
	}{
		{
			name:    "Adds basic priority to labels",
			prepend: "priority: ",
			label:   "P2",
			current: &sloRuleJSON{
				AppliesToJSON:          AppliesToJSON{},
				ComplianceSettingsJSON: ComplianceSettingsJSON{},
			},
			expected: &sloRuleJSON{
				AppliesToJSON: AppliesToJSON{
					GitHubLabelsRaw: []string{"priority: P2"},
				},
				ComplianceSettingsJSON: ComplianceSettingsJSON{},
			},
		},
		{
			name:    "Adds type to labels",
			prepend: "type: ",
			label:   "bug",
			current: &sloRuleJSON{
				AppliesToJSON:          AppliesToJSON{},
				ComplianceSettingsJSON: ComplianceSettingsJSON{},
			},
			expected: &sloRuleJSON{
				AppliesToJSON: AppliesToJSON{
					GitHubLabelsRaw: []string{"type: bug"},
				},
				ComplianceSettingsJSON: ComplianceSettingsJSON{},
			},
		},
		{
			name:    "Empty label is not added",
			prepend: "type: ",
			label:   "",
			current: &sloRuleJSON{
				AppliesToJSON:          AppliesToJSON{},
				ComplianceSettingsJSON: ComplianceSettingsJSON{},
			},
			expected: &sloRuleJSON{
				AppliesToJSON:          AppliesToJSON{},
				ComplianceSettingsJSON: ComplianceSettingsJSON{},
			},
		},
	}
	for _, c := range cases {
		c.current.addToGitHubLabels(c.prepend, c.label)
		if !reflect.DeepEqual(c.current, c.expected) {
			t.Errorf("add to GH labels: %v did not pass. \n\tGot:\t\t%v\n\tExpected:\t%v", c.name, c.current, c.expected)
		}

	}
}

func TestSetResponderDefault(t *testing.T) {
	cases := []struct {
		name     string
		current  *sloRuleJSON
		expected *sloRuleJSON
	}{
		{
			name: "Sets default correctly on empty rule",
			current: &sloRuleJSON{
				AppliesToJSON:          AppliesToJSON{},
				ComplianceSettingsJSON: ComplianceSettingsJSON{},
			},
			expected: &sloRuleJSON{
				AppliesToJSON: AppliesToJSON{},
				ComplianceSettingsJSON: ComplianceSettingsJSON{
					RespondersJSON: RespondersJSON{Contributors: "WRITE"},
				},
			},
		},
		{
			name: "Does nothing on a rule with contributors defined",
			current: &sloRuleJSON{
				AppliesToJSON: AppliesToJSON{},
				ComplianceSettingsJSON: ComplianceSettingsJSON{
					RespondersJSON: RespondersJSON{Contributors: "OWNER"},
				},
			},
			expected: &sloRuleJSON{
				AppliesToJSON: AppliesToJSON{},
				ComplianceSettingsJSON: ComplianceSettingsJSON{
					RespondersJSON: RespondersJSON{Contributors: "OWNER"},
				},
			},
		},
		{
			name: "Does nothing on a rule with users defined",
			current: &sloRuleJSON{
				AppliesToJSON: AppliesToJSON{},
				ComplianceSettingsJSON: ComplianceSettingsJSON{
					RespondersJSON: RespondersJSON{Users: []string{}},
				},
			},
			expected: &sloRuleJSON{
				AppliesToJSON: AppliesToJSON{},
				ComplianceSettingsJSON: ComplianceSettingsJSON{
					RespondersJSON: RespondersJSON{Users: []string{}},
				},
			},
		},
		{
			name: "Does nothing on a rule with owners defined",
			current: &sloRuleJSON{
				AppliesToJSON: AppliesToJSON{},
				ComplianceSettingsJSON: ComplianceSettingsJSON{
					RespondersJSON: RespondersJSON{OwnersRaw: []string{}},
				},
			},
			expected: &sloRuleJSON{
				AppliesToJSON: AppliesToJSON{},
				ComplianceSettingsJSON: ComplianceSettingsJSON{
					RespondersJSON: RespondersJSON{OwnersRaw: []string{}},
				},
			},
		},
	}
	for _, c := range cases {
		c.current.applyResponderDefault()
		if !reflect.DeepEqual(c.current, c.expected) {
			t.Errorf("set responders default: %v did not pass. \n\tGot:\t%v\n\tWant:\t%v", c.name, c.current, c.expected)
		}

	}
}
