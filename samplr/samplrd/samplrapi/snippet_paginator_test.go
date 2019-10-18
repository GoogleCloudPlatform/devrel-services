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

package samplrapi

import (
	"github.com/GoogleCloudPlatform/devrel-services/samplr"
	"reflect"
	"testing"
	"time"
)

func TestSnippetPaginatorPurgesOldRecords(t *testing.T) {
	now := time.Now()
	tests := []struct {
		init map[time.Time]snippetPage
		want map[time.Time]snippetPage
	}{
		{
			init: map[time.Time]snippetPage{
				now.Add(time.Hour * -2).Truncate(0): snippetPage{},
			},
			want: map[time.Time]snippetPage{},
		},
		{
			init: map[time.Time]snippetPage{
				now.Add(time.Hour * -2).Truncate(0): snippetPage{},
				now.Add(time.Hour * -1).Truncate(0): snippetPage{},
			},
			want: map[time.Time]snippetPage{
				now.Add(time.Hour * -1).Truncate(0): snippetPage{},
			},
		},
	}
	for _, tst := range tests {
		sp := &snippetPaginator{
			set: tst.init,
		}
		sp.PurgeOldRecords()
		if !reflect.DeepEqual(sp.set, tst.want) {
			t.Errorf("PurgeOldRecords. Want %v  Got %v", tst.want, sp.set)
		}
	}
}

func TestSnippetPaginatorCreatesPage(t *testing.T) {
	sp := &snippetPaginator{
		set: make(map[time.Time]snippetPage),
	}
	dt, err := sp.CreatePage([]*samplr.Snippet{
		&samplr.Snippet{},
	})
	if dt.After(time.Now()) {
		t.Error("Time was created in the future")
	}
	if err != nil {
		t.Errorf("Unexpected error from CreatePage. Wanted nil, Got %v", err)
	}
}

func TestSnippetPaginatorGetsPage(t *testing.T) {
	tests := []struct {
		snps   []*samplr.Snippet
		cerror error
		gps    int
		garray []*samplr.Snippet
		gidx   int
		gerror error
	}{
		{
			snps:   []*samplr.Snippet{},
			cerror: nil,
			gps:    100,
			garray: []*samplr.Snippet{},
			gidx:   -1,
			gerror: nil,
		},
		{
			snps: []*samplr.Snippet{
				&samplr.Snippet{},
			},
			cerror: nil,
			gps:    1,
			garray: []*samplr.Snippet{
				&samplr.Snippet{},
			},
			gidx:   -1,
			gerror: nil,
		},
		{
			snps: []*samplr.Snippet{
				&samplr.Snippet{},
				&samplr.Snippet{},
				&samplr.Snippet{},
			},
			cerror: nil,
			gps:    2,
			garray: []*samplr.Snippet{
				&samplr.Snippet{},
				&samplr.Snippet{},
			},
			gidx:   2,
			gerror: nil,
		},
	}

	for _, test := range tests {

		sp := &snippetPaginator{
			set: make(map[time.Time]snippetPage),
		}
		ct, cerr := sp.CreatePage(test.snps)
		if cerr != test.cerror {
			t.Errorf("Error in CreatePage. Expected %v, Got %v", test.cerror, cerr)
		}
		gv, gidx, gerr := sp.GetPage(ct, test.gps)
		if gidx != test.gidx {
			t.Errorf("Error in GetPage. Expected Index %v, Got %v", test.gidx, gidx)
		}
		if gerr != test.gerror {
			t.Errorf("Error in GetPage. Expected Error %v, Got %v", test.gerror, gerr)
		}
		if len(gv) != len(test.garray) {
			t.Errorf("Error in GetPage. Expected values %v, Got %v", len(test.garray), len(gv))
		}
	}
}
