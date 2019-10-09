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

import "time"

type RequestRule struct {
	ID                 int
	Labels             []string
	ResponseDuration   int // seconds
	ResolutionDuration int // seconds
}

func (slo *RequestRule) compliantUpdates(lastUpdate time.Time) *ComplianceResponse {
	target := slo.ResponseDuration
	actual := int(time.Now().Sub(lastUpdate).Seconds())
	return &ComplianceResponse{actual < target, actual}
}

func (slo *RequestRule) compliantResolution(opened time.Time) *ComplianceResponse {
	target := slo.ResolutionDuration
	actual := int(time.Now().Sub(opened).Seconds())
	return &ComplianceResponse{actual < target, actual}
}
