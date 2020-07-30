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

package paginator

import (
	"reflect"
	"testing"
	"time"
)

func TestInitStrings(t *testing.T) {

	someTime := time.Now().Add(time.Minute * -1).UTC().Truncate(0)

	p := &Strings{set: map[time.Time]stringPage{
		someTime: stringPage{items: []string{"fluff"}, index: 0},
	}}

	_, err := p.CreatePage([]string{"fluff"})
	if err == nil {
		t.Errorf("Error initializing paginator: creating page without initialization did not error")
	}

	_, _, err = p.GetPage(someTime, 1)
	if err == nil {
		t.Errorf("Error initializing paginator: duplicate initialization did not error")
	}

	err = p.Init()
	if err != nil {
		t.Errorf("Error initializing paginator: %v", err)
	}
	err = p.Init()
	if err == nil {
		t.Errorf("Error initializing paginator: duplicate initialization did not error")
	}
}

func TestPurgesOldRecordsStrings(t *testing.T) {
	now := time.Now()
	p := &Strings{}
	p.Init()

	tests := []struct {
		name string
		init map[time.Time]stringPage
		want map[time.Time]stringPage
	}{
		{
			name: "Purges one 2 hour old record",
			init: map[time.Time]stringPage{
				now.Add(time.Hour * -2).Truncate(0): stringPage{},
			},
			want: map[time.Time]stringPage{},
		},
		{
			name: "Purges one 2 hour old record, does not purge 1 hour old record",
			init: map[time.Time]stringPage{
				now.Add(time.Hour * -2).Truncate(0): stringPage{},
				now.Add(time.Hour * -1).Truncate(0): stringPage{},
			},
			want: map[time.Time]stringPage{
				now.Add(time.Hour * -1).Truncate(0): stringPage{},
			},
		},
		{
			name: "Purges several 2 hour old records, does not purge 1 hour old record",
			init: map[time.Time]stringPage{
				now.Add(time.Hour * -1).Truncate(0): stringPage{},
				now.Add(time.Hour * -2).Truncate(0): stringPage{},
				now.Add(time.Hour * -2).Truncate(0): stringPage{},
			},
			want: map[time.Time]stringPage{
				now.Add(time.Hour * -1).Truncate(0): stringPage{},
			},
		},
		{
			name: "Purges records more than 2 hours old",
			init: map[time.Time]stringPage{
				now.Add(time.Hour * -4).Truncate(0): stringPage{},
				now.Add(time.Hour * -3).Truncate(0): stringPage{},
				now.Add(time.Hour * -1).Truncate(0): stringPage{},
			},
			want: map[time.Time]stringPage{
				now.Add(time.Hour * -1).Truncate(0): stringPage{},
			},
		},
	}
	for _, test := range tests {
		p.set = test.init
		p.PurgeOldRecords()

		if !reflect.DeepEqual(p.set, test.want) {
			t.Errorf("PurgeOldRecords: %v did not pass. Want %v  Got %v", test.name, test.want, p.set)
		}
	}
}

func TestCreatesPageStrings(t *testing.T) {
	p := &Strings{}
	p.Init()

	tests := []struct {
		name    string
		items   []string
		wantErr bool
	}{
		{
			name:    "Creates Page with an item",
			items:   []string{"string"},
			wantErr: false,
		},
		{
			name:    "Creates Page with several items",
			items:   []string{"string", "s2"},
			wantErr: false,
		},
		{
			name:    "Creates Page with >100 items",
			items:   make([]string, 150),
			wantErr: false,
		},
	}
	for _, test := range tests {
		key, gotErr := p.CreatePage(test.items)
		gotItems, ok := p.set[key]

		if key.After(time.Now()) {
			t.Errorf("CreatesPage: %v did not pass. Key is in the future.", test.name)
		}

		if !ok {
			t.Errorf("CreatesPage: %v did not pass. Key is not in set.", test.name)
		}

		if !reflect.DeepEqual(gotItems.items, test.items) || gotItems.index != 0 {
			t.Errorf("CreatesPage: %v did not pass. Items not found at key. Want: %v  Got: %v",
				test.name, test.items, gotItems)
		}

		if (gotErr != nil && !test.wantErr) || (gotErr == nil && test.wantErr) {
			t.Errorf("%v did not pass.\n\tWant Err: %v \n\tGot Err: %v", test.name, test.wantErr, gotErr)
		}
	}
}

func TestGetsPageStrings(t *testing.T) {
	now := time.Now()
	p := &Strings{}
	p.Init()

	someTime := now.Add(time.Hour * -1).UTC().Truncate(0)
	time2 := now.Add(time.Minute * -1).UTC().Truncate(0)

	tests := []struct {
		name      string
		initSet   map[time.Time]stringPage
		key       time.Time
		numItems  int
		wantItems []string
		wantIdx   int
		wantErr   bool
	}{
		{
			name: "Errors on empty page",
			initSet: map[time.Time]stringPage{
				someTime: stringPage{items: []string{}, index: 0},
			},
			key:       someTime,
			numItems:  100,
			wantItems: nil,
			wantIdx:   0,
			wantErr:   true,
		},
		{
			name: "Gets one item in one item page",
			initSet: map[time.Time]stringPage{
				someTime: stringPage{items: []string{"item1"}, index: 0},
			},
			key:       someTime,
			numItems:  1,
			wantItems: []string{"item1"},
			wantIdx:   -1,
			wantErr:   false,
		},
		{
			name: "Gets one item from correct page",
			initSet: map[time.Time]stringPage{
				someTime: stringPage{items: []string{"item1"}, index: 0},
				time2:    stringPage{items: []string{"2"}, index: 0},
			},
			key:       someTime,
			numItems:  1,
			wantItems: []string{"item1"},
			wantIdx:   -1,
			wantErr:   false,
		},
		{
			name: "Gets one item in several item page",
			initSet: map[time.Time]stringPage{
				someTime: stringPage{items: []string{"item1", "2", "3"}, index: 0},
			},
			key:       someTime,
			numItems:  1,
			wantItems: []string{"item1"},
			wantIdx:   1,
			wantErr:   false,
		},
		{
			name: "Gets next items",
			initSet: map[time.Time]stringPage{
				someTime: stringPage{items: []string{"item1", "2", "3"}, index: 1},
			},
			key:       someTime,
			numItems:  100,
			wantItems: []string{"2", "3"},
			wantIdx:   -1,
			wantErr:   false,
		},
		{
			name: "Errors if key is not in set",
			initSet: map[time.Time]stringPage{
				time2: stringPage{items: []string{"item1", "2", "3"}, index: 1},
			},
			key:       someTime,
			numItems:  100,
			wantItems: nil,
			wantIdx:   0,
			wantErr:   true,
		},
	}

	for _, test := range tests {

		p.set = test.initSet
		gotItems, gotIdx, gotErr := p.GetPage(test.key, test.numItems)

		if test.wantIdx != gotIdx {
			t.Errorf("%v did not pass. Expected Index %v, Got %v",
				test.name, test.wantIdx, gotIdx)
		}

		if !reflect.DeepEqual(test.wantItems, gotItems) {
			t.Errorf("%v did not pass.\n\tExpected values %v, Got %v",
				test.name, test.wantItems, gotItems)
		}

		if (gotErr != nil && !test.wantErr) || (gotErr == nil && test.wantErr) {
			t.Errorf("%v did not pass.\n\tWant Err: %v \n\tGot Err: %v", test.name, test.wantErr, gotErr)
		}
	}

}
