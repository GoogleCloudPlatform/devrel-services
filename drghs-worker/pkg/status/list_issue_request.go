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

import "encoding/json"

type ListIssuesRequest struct {
	Repo string

	// This defaults to undefined, such that both Issues and PullRequests are
	// returned unless this is explicitly set to true or false.
	PullRequest json.RawMessage `json:"PullRequest"`

	// This defaults to undefined, such that both Open and Closed results are
	// returned unless this is explicitly set to true or false.
	Closed json.RawMessage `json:"Closed"`

	// This defaults to false if not explicitly set to true. Set to true in order
	// for comments to be included in the response.
	IncludeComments bool `json:"IncludeComments"`

	// This defaults to false if not explicitly set to true. Set to true in order
	// for reviewts to be included in the response.
	IncludeReviews bool `json:"IncludeReviews"`
}
