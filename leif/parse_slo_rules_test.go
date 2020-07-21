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

var oneDay = 24 * time.Hour

var syntaxError *json.SyntaxError
var unmarshalTypeError *json.UnmarshalTypeError

func TestParseSLORules(t *testing.T) {
	tests := []struct {
		name      string
		jsonInput string
		expected  []*SLORule
		wantErr   bool
	}{
		{
			name:      "Empty array returns no rules",
			jsonInput: `[]`,
			expected:  nil,
			wantErr:   false,
		},
		{
			name:      "Empty json returns no rules",
			jsonInput: ``,
			expected:  nil,
			wantErr:   false,
		},
		{
			name: "More than one rule is parsed correctly",
			jsonInput: `[
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
			wantErr:  false,
		},
		{
			name:      "Malformed input errors",
			jsonInput: `["no end bracket`,
			expected:  nil,
			wantErr:   true,
		},
		{
			name: "Malformed rule errors",
			jsonInput: `[
				{
					"appliesTo": 1,
					"complianceSettings": {
						"responseTime": 0,
						"resolutionTime": 0
					}
				}
		 	]`,
			expected: nil,
			wantErr:  true,
		},
	}
	for _, test := range tests {
		got, err := unmarshalSLOs([]byte(test.jsonInput))
		if !reflect.DeepEqual(got, test.expected) {
			t.Errorf("%v did not pass.\n\tWant:\t%v\n\tGot:\t%v", test.name, test.expected, got)
		}
		if (test.wantErr && err == nil) || (!test.wantErr && err != nil) {
			t.Errorf("%v did not pass.\n\tWant Err: %v \n\tGot Err: %v", test.name, test.wantErr, err)
		}
	}
}

func TestParseSLORule(t *testing.T) {
	tests := []struct {
		name      string
		jsonInput string
		expected  *SLORule
		wantErr   bool
	}{
		{
			name: "Minimum default rule returns a rule with defaults applied",
			jsonInput: `{
				"appliesTo": {},
				"complianceSettings": {
					"responseTime": 0,
					"resolutionTime": 0
				}
			}`,
			expected: &defaultSLO,
			wantErr:  false,
		},
		{
			name: "Time strings are parsed correctly",
			jsonInput: `{
				"appliesTo": {},
				"complianceSettings": {
					"responseTime": "1h",
					"resolutionTime": "1s"
				}
			}`,
			expected: &SLORule{
				AppliesTo: AppliesTo{
					Issues: true,
					PRs:    false,
				},
				ComplianceSettings: ComplianceSettings{
					ResponseTime:   time.Hour,
					ResolutionTime: time.Second,
					Responders:     Responders{Contributors: "WRITE"},
				},
			},
			wantErr: false,
		},
		{
			name: "Time strings with multiple values are parsed correctly",
			jsonInput: `{
				"appliesTo": {},
				"complianceSettings": {
					"responseTime": "1h1m1s",
					"resolutionTime": "1s1h1d"
				}
			}`,
			expected: &SLORule{
				AppliesTo: AppliesTo{
					Issues: true,
					PRs:    false,
				},
				ComplianceSettings: ComplianceSettings{
					ResponseTime:   (time.Duration(time.Hour + time.Minute + time.Second)),
					ResolutionTime: (time.Duration(time.Second + time.Hour + oneDay)),
					Responders:     Responders{Contributors: "WRITE"},
				},
			},
			wantErr: false,
		},
		{
			name: "Time strings with day values are parsed correctly",
			jsonInput: `{
				"appliesTo": {},
				"complianceSettings": {
					"responseTime": "1d1h1m1s",
					"resolutionTime": "30d"
				}
			}`,
			expected: &SLORule{
				AppliesTo: AppliesTo{
					Issues: true,
					PRs:    false,
				},
				ComplianceSettings: ComplianceSettings{
					ResponseTime:   (time.Duration((24 * time.Hour) + time.Hour + time.Minute + time.Second)),
					ResolutionTime: (time.Duration(30 * 24 * time.Hour)),
					Responders:     Responders{Contributors: "WRITE"},
				},
			},
			wantErr: false,
		},
		{
			name: "Time defined as a number is parsed correctly",
			jsonInput: `{
					"appliesTo": {},
					"complianceSettings": {
						"responseTime": 0,
						"resolutionTime": 43200
					}
				}`,
			expected: &SLORule{
				AppliesTo: AppliesTo{
					Issues: true,
					PRs:    false,
				},
				ComplianceSettings: ComplianceSettings{
					ResponseTime:   0,
					ResolutionTime: time.Duration(43200 * time.Second),
					Responders:     Responders{Contributors: "WRITE"},
				},
			},
			wantErr: false,
		},
		{
			name: "Incorrect time string fails with error",
			jsonInput: `{
					"appliesTo": {},
					"complianceSettings": {
						"responseTime": "1w",
						"resolutionTime": 0
					}
				}`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Priority gets converted to a GitHub label",
			jsonInput: `{
				"appliesTo": {
					"priority": "P1"
				},
				"complianceSettings": {
					"responseTime": 0,
					"resolutionTime": 0
				}
			}`,
			expected: &SLORule{
				AppliesTo: AppliesTo{
					GitHubLabels: []string{"priority: P1"},
					Issues:       true,
					PRs:          false,
				},
				ComplianceSettings: ComplianceSettings{
					ResponseTime:   0,
					ResolutionTime: 0,
					Responders:     Responders{Contributors: "WRITE"},
				},
			},
			wantErr: false,
		},
		{
			name: "Issue type gets converted to a GitHub label",
			jsonInput: `{
				"appliesTo": {
					"issueType": "bug"
				},
				"complianceSettings": {
					"responseTime": 0,
					"resolutionTime": 0
				}
			}`,
			expected: &SLORule{
				AppliesTo: AppliesTo{
					GitHubLabels: []string{"type: bug"},
					Issues:       true,
					PRs:          false,
				},
				ComplianceSettings: ComplianceSettings{
					ResponseTime:   0,
					ResolutionTime: 0,
					Responders:     Responders{Contributors: "WRITE"},
				},
			},
			wantErr: false,
		},
		{
			name: "Contributors is set to none if another field is defined in Responders",
			jsonInput: `{
				"appliesTo": {},
				"complianceSettings": {
					"responseTime": 0,
					"resolutionTime": 0,
					"responders": {
						"users": ["@jeff"]
					}
				}
			}`,
			expected: &SLORule{
				AppliesTo: AppliesTo{
					Issues: true,
					PRs:    false,
				},
				ComplianceSettings: ComplianceSettings{
					ResponseTime:   0,
					ResolutionTime: 0,
					Responders:     Responders{Users: []string{"@jeff"}},
				},
			},
			wantErr: false,
		},
		{
			name: "Can set GitHubLabels as a string",
			jsonInput: `{
				"appliesTo": {
					"gitHubLabels": "a label"
				},
				"complianceSettings": {
					"responseTime": 0,
					"resolutionTime": 0
				}
			}`,
			expected: &SLORule{
				AppliesTo: AppliesTo{
					GitHubLabels: []string{"a label"},
					Issues:       true,
					PRs:          false,
				},
				ComplianceSettings: ComplianceSettings{
					ResponseTime:   0,
					ResolutionTime: 0,
					Responders:     Responders{Contributors: "WRITE"},
				},
			},
			wantErr: false,
		},
		{
			name: "Can set GitHubLabels as an array",
			jsonInput: `{
				"appliesTo": {
					"gitHubLabels": ["label 1", "label 2"]
				},
				"complianceSettings": {
					"responseTime": 0,
					"resolutionTime": 0
				}
			}`,
			expected: &SLORule{
				AppliesTo: AppliesTo{
					GitHubLabels: []string{"label 1", "label 2"},
					Issues:       true,
					PRs:          false,
				},
				ComplianceSettings: ComplianceSettings{
					ResponseTime:   0,
					ResolutionTime: 0,
					Responders:     Responders{Contributors: "WRITE"},
				},
			},
			wantErr: false,
		},
		{
			name: "No responders can be specified",
			jsonInput: `{
				"appliesTo": {},
				"complianceSettings": {
					"responseTime": 0,
					"resolutionTime": 0,
					"responders": {
						"users": []
					}
				}
			}`,
			expected: &SLORule{
				AppliesTo: AppliesTo{
					Issues: true,
					PRs:    false,
				},
				ComplianceSettings: ComplianceSettings{
					ResponseTime:   0,
					ResolutionTime: 0,
					Responders:     Responders{Users: []string{}},
				},
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		e := json.RawMessage(test.jsonInput)
		got, err := parseSLORule(&e)
		if !reflect.DeepEqual(got, test.expected) {
			t.Errorf("%v did not pass.\n\tWant:\t%v\n\tGot:\t%v", test.name, test.expected, got)
		}
		if (test.wantErr && err == nil) || (!test.wantErr && err != nil) {
			t.Errorf("%v did not pass.\n\tWant Err: %v \n\tGot Err: %v", test.name, test.wantErr, err)
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
			want:     time.Hour,
			wantErr:  false,
		},
		{
			name:     "Can parse a day",
			duration: "1d",
			want:     oneDay,
			wantErr:  false,
		},
		{
			name:     "Can parse several days",
			duration: "2d",
			want:     time.Duration(2 * oneDay),
			wantErr:  false,
		},
		{
			name:     "Multiple digits acceptable for days",
			duration: "10d",
			want:     time.Duration(10 * oneDay),
			wantErr:  false,
		},
		{
			name:     "Can parse hours and days",
			duration: "1d1h",
			want:     time.Duration(oneDay + time.Hour),
			wantErr:  false,
		},
		{
			name:     "Days may be at any position in the string",
			duration: "1s10d",
			want:     time.Duration(10*oneDay + time.Second),
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
		if !reflect.DeepEqual(got, c.expected) {
			t.Errorf("%v did not pass.\n\tGot:\t%v\n\tWant:\t%v", c.name, got, c.expected)
		}
		if (c.wantErr == nil && gotErr != nil) || (c.wantErr != nil && reflect.TypeOf(gotErr) != reflect.TypeOf(c.wantErr)) {
			t.Errorf("%v did not pass.\n\tGot Err:\t%v\n\tWant Err:\t%v", c.name, gotErr, c.wantErr)
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
			expected: `0`,
			wantErr:  nil,
		},
		{
			name:     "Marshal duration as time.duration",
			dur:      duration(time.Second),
			expected: `1000000000`,
			wantErr:  nil,
		},
		{
			name:     "Can marshal a day correctly",
			dur:      duration(oneDay),
			expected: `86400000000000`,
			wantErr:  nil,
		},
		{
			name:     "Can marshal several days correctly",
			dur:      duration(2 * oneDay),
			expected: `172800000000000`,
			wantErr:  nil,
		},
		{
			name:     "Can marshal days with seconds correctly",
			dur:      duration(2*oneDay + time.Second),
			expected: `172801000000000`,
			wantErr:  nil,
		},
	}
	for _, c := range cases {
		got, gotErr := json.Marshal(c.dur)

		if !reflect.DeepEqual(got, []byte(c.expected)) {
			t.Errorf("%v did not pass.\n\tGot:\t%v\n\tWant:\t%v", c.name, string(got), c.expected)
		}
		if (c.wantErr == nil && gotErr != nil) || (c.wantErr != nil && reflect.TypeOf(gotErr) != reflect.TypeOf(c.wantErr)) {
			t.Errorf("%v did not pass.\n\tGot Err:\t%v\n\tWant Err:\t%v", c.name, gotErr, c.wantErr)
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
			expected:  duration(time.Second),
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
			name:      "Unmarshals 1 day correctly",
			jsonInput: `"1d"`,
			expected:  duration(oneDay),
			wantErr:   false,
		},
		{
			name:      "Unmarshals several days correctly",
			jsonInput: `"2d"`,
			expected:  duration(2 * oneDay),
			wantErr:   false,
		},
		{
			name:      "Unmarshals days, hours and minutes correctly",
			jsonInput: `"1d1h1m"`,
			expected:  duration(oneDay + time.Hour + time.Minute),
			wantErr:   false,
		},
		{
			name:      "Unmarshals multi-digit days correctly",
			jsonInput: `"10d1m"`,
			expected:  duration(oneDay*10 + time.Minute),
			wantErr:   false,
		},
		{
			name:      "Incorrect time format fails",
			jsonInput: `"1w"`,
			expected:  0,
			wantErr:   true,
		},
		{
			name:      "Partially incorrect time format fails",
			jsonInput: `"1d1e"`,
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
			t.Errorf("%v did not pass.\n\tGot:\t%v\n\tWant:\t%v", c.name, *got, c.expected)
		}
		if (c.wantErr && gotErr == nil) || (!c.wantErr && gotErr != nil) {
			t.Errorf("%v did not pass.\n\tGot Err:\t%v\n\tWant Err:\t%v", c.name, gotErr, c.wantErr)
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
			t.Errorf("%v did not pass. \n\tGot:\t\t%v\n\tExpected:\t%v", c.name, got, c.expected)
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
			t.Errorf("%v did not pass. \n\tGot:\t\t%v\n\tExpected:\t%v", c.name, c.current, c.expected)
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
			t.Errorf("%v did not pass. \n\tGot:\t%v\n\tWant:\t%v", c.name, c.current, c.expected)
		}

	}
}
