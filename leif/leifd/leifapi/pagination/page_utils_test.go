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

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"

	"github.com/golang/protobuf/ptypes/timestamp"
)

func TestGetPageSize(t *testing.T) {
	tests := []struct {
		n    int
		want int
	}{
		{
			n:    1,
			want: 1,
		},
		{
			n:    0,
			want: 100,
		},
		{
			n:    100,
			want: 100,
		},
		{
			n:    50,
			want: 50,
		},
		{
			n:    -1,
			want: 100,
		},
		{
			n:    101,
			want: 100,
		},
	}
	for _, tst := range tests {
		got := GetPageSize(tst.n)
		if got != tst.want {
			t.Errorf("Error in getPageSize. Want %v, got %v", tst.want, got)
		}
	}
}

func TestMakeNextPageToken(t *testing.T) {
	tests := []struct {
		token   *drghs_v1.PageToken
		idx     int
		wantstr string
		wantErr error
	}{
		{
			token:   &drghs_v1.PageToken{},
			idx:     1,
			wantstr: "CAE=",
			wantErr: nil,
		},
		{
			token:   &drghs_v1.PageToken{},
			idx:     10,
			wantstr: "CAo=",
			wantErr: nil,
		},
		{
			token:   nil,
			idx:     10,
			wantstr: "",
			wantErr: ErrNilPageToken,
		},
		{
			token: &drghs_v1.PageToken{
				FirstRequestTimeUsec: &timestamp.Timestamp{Seconds: 1500},
			},
			idx:     10,
			wantstr: "CAoSAwjcCw==",
			wantErr: nil,
		},
	}
	for _, tst := range tests {
		gotstr, gotErr := MakeNextPageToken(tst.token, tst.idx)
		if gotstr != tst.wantstr {
			t.Errorf("makeNextPageToken. Want %v, got %v", tst.wantstr, gotstr)
		}
		if gotErr != tst.wantErr {
			t.Errorf("makeNextPageToken Error. Want %v, got %v", tst.wantErr, gotErr)
		}
	}
}

func TestMakeFirstPageToken(t *testing.T) {
	tests := []struct {
		ti      time.Time
		idx     int
		wantstr string
		wantErr error
	}{
		{
			ti:      time.Unix(0, 0),
			idx:     1,
			wantstr: "CAESAA==",
			wantErr: nil,
		},
		{
			ti:      time.Unix(0, 0),
			idx:     10,
			wantstr: "CAoSAA==",
			wantErr: nil,
		},
		{
			ti: time.Unix(1500, 0),

			idx:     10,
			wantstr: "CAoSAwjcCw==",
			wantErr: nil,
		},
	}
	for _, tst := range tests {
		gotstr, gotErr := MakeFirstPageToken(tst.ti, tst.idx)
		if gotstr != tst.wantstr {
			t.Errorf("MakeFirstPageToken. Want %v, got %v", tst.wantstr, gotstr)
		}
		if gotErr != tst.wantErr {
			t.Errorf("MakeFirstPageToken did not pass. WantErr: %v, gotErr: %v", tst.wantErr, gotErr)
		}
	}
}

func TestDecodePageToken(t *testing.T) {
	tests := []struct {
		str     string
		want    *drghs_v1.PageToken
		wantErr error
	}{
		{
			str: "CAoSAA==",
			want: &drghs_v1.PageToken{
				Offset:               10,
				FirstRequestTimeUsec: &timestamp.Timestamp{Seconds: 0},
			},
			wantErr: nil,
		},
		{
			str: "CAoSAwjcCw==",
			want: &drghs_v1.PageToken{
				Offset:               10,
				FirstRequestTimeUsec: &timestamp.Timestamp{Seconds: 1500},
			},
			wantErr: nil,
		},
	}
	for _, tst := range tests {
		got, gotErr := DecodePageToken(tst.str)
		if tst.wantErr != gotErr {
			t.Errorf("Want %v Got %v", tst.wantErr, gotErr)
		}
		if !reflect.DeepEqual(got, tst.want) {
			t.Errorf("Want %v Got %v", tst.want, got)
		}
	}
}
