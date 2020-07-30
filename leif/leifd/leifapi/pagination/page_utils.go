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
	b64 "encoding/base64"
	"errors"
	"time"

	drghs_v1 "github.com/GoogleCloudPlatform/devrel-services/drghs/v1"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

var (
	// ErrNilPageToken is returned when a PageToken is nil
	ErrNilPageToken = errors.New("nil PageToken")
)

// DecodePageToken translates a token from a drghs_v1 api call to Go
func DecodePageToken(req string) (*drghs_v1.PageToken, error) {
	pageToken := &drghs_v1.PageToken{}
	decstr, err := b64.StdEncoding.DecodeString(req)
	err = pageToken.XXX_Unmarshal(decstr)
	if err != nil {
		return nil, err
	}
	return pageToken, nil
}

// MakeFirstPageToken creates a new string page token for the given key/time
func MakeFirstPageToken(t time.Time, idx int) (string, error) {
	tsp, err := ptypes.TimestampProto(t)
	if err != nil {
		return "", err
	}
	return MakeNextPageToken(&drghs_v1.PageToken{
		FirstRequestTimeUsec: tsp,
		Offset:               int32(idx),
	}, idx)
}

// MakeNextPageToken creates a new string page token at the given index, based on prev
func MakeNextPageToken(prev *drghs_v1.PageToken, idx int) (string, error) {
	nextPageTokenStr := ""
	if prev == nil {
		return "", ErrNilPageToken
	}
	if idx > 0 {
		prev.Offset = int32(idx)
		pagetokenbytes, err := proto.Marshal(prev)
		if err != nil {
			return "", err
		}
		nextPageTokenStr = b64.StdEncoding.EncodeToString(pagetokenbytes)
	}
	return nextPageTokenStr, nil
}

// GetPageSize returns the page size, setting a maximum page size of 100
func GetPageSize(reqPageSize int) int {
	pagesize := 100
	if 0 < reqPageSize && reqPageSize < 100 {
		pagesize = reqPageSize
	}
	return pagesize
}
