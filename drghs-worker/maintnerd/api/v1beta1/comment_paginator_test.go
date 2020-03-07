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

package v1beta1

import (
	"reflect"
	"testing"
	"time"

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"
)

func TestCommentPaginatorPurgesOldRecords(t *testing.T) {
	now := time.Now()
	tests := []struct {
		init map[time.Time]commentPage
		want map[time.Time]commentPage
	}{
		{
			init: map[time.Time]commentPage{
				now.Add(time.Hour * -2).Truncate(0): commentPage{},
			},
			want: map[time.Time]commentPage{},
		},
		{
			init: map[time.Time]commentPage{
				now.Add(time.Hour * -2).Truncate(0): commentPage{},
				now.Add(time.Hour * -1).Truncate(0): commentPage{},
			},
			want: map[time.Time]commentPage{
				now.Add(time.Hour * -1).Truncate(0): commentPage{},
			},
		},
	}
	for _, tst := range tests {
		sp := &commentPaginator{
			set: tst.init,
		}
		sp.PurgeOldRecords()
		if !reflect.DeepEqual(sp.set, tst.want) {
			t.Errorf("PurgeOldRecords. Want %v  Got %v", tst.want, sp.set)
		}
	}
}

func TestCommentPaginatorCreatesPage(t *testing.T) {
	sp := &commentPaginator{
		set: make(map[time.Time]commentPage),
	}
	dt, err := sp.CreatePage([]*drghs_v1.GitHubComment{
		&drghs_v1.GitHubComment{},
	})
	if dt.After(time.Now()) {
		t.Error("Time was created in the future")
	}
	if err != nil {
		t.Errorf("Unexpected error from CreatePage. Wanted nil, Got %v", err)
	}
}

func TestCommentPaginatorGetsPage(t *testing.T) {
	tests := []struct {
		iss    []*drghs_v1.GitHubComment
		cerror error
		gps    int
		garray []*drghs_v1.GitHubComment
		gidx   int
		gerror error
	}{
		{
			iss:    []*drghs_v1.GitHubComment{},
			cerror: nil,
			gps:    100,
			garray: []*drghs_v1.GitHubComment{},
			gidx:   -1,
			gerror: nil,
		},
		{
			iss: []*drghs_v1.GitHubComment{
				&drghs_v1.GitHubComment{},
			},
			cerror: nil,
			gps:    1,
			garray: []*drghs_v1.GitHubComment{
				&drghs_v1.GitHubComment{},
			},
			gidx:   -1,
			gerror: nil,
		},
		{
			iss: []*drghs_v1.GitHubComment{
				&drghs_v1.GitHubComment{},
				&drghs_v1.GitHubComment{},
				&drghs_v1.GitHubComment{},
			},
			cerror: nil,
			gps:    2,
			garray: []*drghs_v1.GitHubComment{
				&drghs_v1.GitHubComment{},
				&drghs_v1.GitHubComment{},
			},
			gidx:   2,
			gerror: nil,
		},
	}

	for _, test := range tests {

		sp := &commentPaginator{
			set: make(map[time.Time]commentPage),
		}
		ct, cerr := sp.CreatePage(test.iss)
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
